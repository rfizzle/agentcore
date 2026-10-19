package repeat

import (
	"context"
	"strings"
	"sync"

	"github.com/rfizzle/agentcore/agent"
	"github.com/rfizzle/llm"
	llmattachment "github.com/rfizzle/llm/attachment"
)

// The loop and the app words these tests drive the detector with, copied
// from the agent package's tests: an Agent that never streams, digest for the
// outcome word a reader takes off a result's "error:" head, and resultParts
// for the side channel a tool attaches an image through.

func noStream([]llm.Message, string) (<-chan llm.StreamEvent, context.CancelFunc, error) {
	ch := make(chan llm.StreamEvent)
	close(ch)
	_, cancel := context.WithCancel(context.Background())
	return ch, cancel, nil
}

func newTestAgent() *agent.Agent {
	return agent.New([]llm.Message{{Role: llm.RoleSystem, Content: "sys"}}, noStream)
}

// digest reads a result's outcome the way every reader of one does: a
// failure by its "error:" head.
var digest = fakeDigest{OutcomeError: "error"}

type fakeDigest struct {
	OutcomeError string
}

func (d fakeDigest) Outcome(result string) string {
	if strings.HasPrefix(result, "error:") {
		return d.OutcomeError
	}
	return "ok"
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
