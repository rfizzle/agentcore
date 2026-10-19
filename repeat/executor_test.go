package repeat

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/rfizzle/agentcore/agent"
	"github.com/rfizzle/llm"
	llmattachment "github.com/rfizzle/llm/attachment"
)

func TestExecuteCalls_ImagesSurviveAWholeRoundOfReaders(t *testing.T) {
	a := newTestAgent()
	// The repeat detector rewrites a result it has seen before, which is the
	// wrapper an image has to survive: reading the same picture twice is what
	// a person asks for when they did not believe the first answer.
	a.SetResultParts(resultParts.TakeResult)
	a.SetExecutor(NewRepeatDetector(Shapes{}).WrapExecutor(func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		notice := string(args) + " is an image"
		resultParts.NoteResult(notice, llmattachment.Attachment{
			Kind: llmattachment.Image, Name: "x.png", MediaType: "image/png", Data: []byte{1},
		})
		return notice, nil
	}))

	var calls []llm.ToolCall
	for i := range agent.MaxParallelToolCalls {
		calls = append(calls, llm.ToolCall{
			ID:        strconv.Itoa(i),
			Name:      "read_file",
			Arguments: `{"path":"` + strconv.Itoa(i) + `"}`,
		})
	}
	// Twice: the second round is where every result carries the detector's
	// notice above the text the tool wrote.
	for round := range 2 {
		for _, r := range a.ExecuteCalls(context.Background(), calls) {
			if len(r.Attachments) != 1 {
				t.Fatalf("round %d, call %s: expected the image on the result, got %d",
					round, r.Call.ID, len(r.Attachments))
			}
		}
	}
}
