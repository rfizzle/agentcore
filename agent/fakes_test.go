package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	llmattachment "github.com/rfizzle/llm/attachment"
)

// The app's words and stores these tests drive the loop with. The loop owns
// none of them, so the tests carry small stand-ins under the names an app's
// packages would have: digest for the words a recorder row is made of,
// evidence for the store a trim hands an elided result to, and resultParts
// for the side channel a tool attaches an image through.

// digest is the vocabulary a Recorder is handed: the call's key argument, the
// outcome word read off the "error:" head, and a command's first line.
var digest = fakeDigest{OutcomeError: "error"}

type fakeDigest struct {
	OutcomeError string
}

// Arg is the first of the arguments a row is about, or the raw text's first
// line when it is not an object.
func (fakeDigest) Arg(_, rawArgs string) string {
	var args map[string]any
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return digest.FirstLine(rawArgs)
	}
	for _, k := range []string{"path", "pattern", "command", "query", "url"} {
		if v, ok := args[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func (d fakeDigest) Outcome(result string) string {
	if strings.HasPrefix(result, "error:") {
		return d.OutcomeError
	}
	return "ok"
}

func (fakeDigest) FirstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i] + " …"
	}
	return s
}

// evidence opens an in-memory store whose Keep files a result verbatim under a
// fresh id beside the tool that produced it, and whose Read pages it back.
var evidence fakeEvidence

type fakeEvidence struct{}

func (fakeEvidence) Open(_, _ string) (*fakeStore, error) {
	return &fakeStore{entries: map[string]fakeEntry{}}, nil
}

// NewReducer is the store itself: Keep is all a trim asks of it.
func (fakeEvidence) NewReducer(s *fakeStore) *fakeStore { return s }

type fakeStore struct {
	mu      sync.Mutex
	entries map[string]fakeEntry
}

type fakeEntry struct {
	data []byte
	meta fakeMeta
}

type fakeMeta struct {
	Tool string
}

type fakeStats struct {
	Entries int
}

func (s *fakeStore) Keep(tool, content string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := fmt.Sprintf("ev-%016x", len(s.entries)+1)
	s.entries[id] = fakeEntry{data: []byte(content), meta: fakeMeta{Tool: tool}}
	return id, true
}

func (s *fakeStore) Read(id string, offset, limit int) ([]byte, fakeMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[id]
	if !ok {
		return nil, fakeMeta{}, fmt.Errorf("no evidence %s", id)
	}
	data := e.data[min(offset, len(e.data)):]
	return data[:min(limit, len(data))], e.meta, nil
}

func (s *fakeStore) Stats() fakeStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fakeStats{Entries: len(s.entries)}
}

// resultParts holds what a tool attached to a result until the loop takes it:
// NoteResult files the parts under the result's text, and TakeResult hands
// them to the first result that contains that text, so a wrapper that adds a
// notice to a result does not lose them.
var resultParts = &fakeParts{pending: map[string][]llmattachment.Attachment{}}

type fakeParts struct {
	mu      sync.Mutex
	order   []string
	pending map[string][]llmattachment.Attachment
}

func (p *fakeParts) NoteResult(result string, atts ...llmattachment.Attachment) {
	if result == "" || len(atts) == 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, seen := p.pending[result]; !seen {
		p.order = append(p.order, result)
	}
	p.pending[result] = atts
}

func (p *fakeParts) TakeResult(result string) []llmattachment.Attachment {
	if result == "" {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, key := range p.order {
		if !strings.Contains(result, key) {
			continue
		}
		atts := p.pending[key]
		delete(p.pending, key)
		p.order = append(p.order[:i], p.order[i+1:]...)
		return atts
	}
	return nil
}
