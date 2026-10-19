package meter_test

import (
	"context"
	"fmt"

	"github.com/rfizzle/agentcore/meter"
	"github.com/rfizzle/llm"
	"github.com/rfizzle/llm/pricing"
)

// replier is a provider that answers every request with the same usage, the
// way a real one reports it on its terminal event.
type replier struct{ usage llm.Usage }

func (replier) Name() string { return "demo" }

func (r replier) StreamCompletion(_ context.Context, _ []llm.Message, _ llm.CompletionOpts) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 2)
	u := r.usage
	ch <- llm.StreamEvent{Token: "ok"}
	ch <- llm.StreamEvent{Usage: &u, Done: true}
	close(ch)
	return ch, nil
}

func use(p llm.Provider, model string) {
	events, err := p.StreamCompletion(context.Background(), nil, llm.CompletionOpts{Model: model})
	if err != nil {
		panic(err)
	}
	for range events {
	}
}

// A Ledger prices each request against the model that answered it, and
// ForOrigin wraps a provider so everything it streams is billed to one named
// requester. The spend then reads back by origin, not only as a total.
func ExampleLedger_ForOrigin() {
	prices := pricing.NewTable(map[string]pricing.ModelPricing{
		"demo-model": {InputCostPerToken: 0.000001, OutputCostPerToken: 0.000002},
	})
	ledger := meter.New(prices)
	base := replier{usage: llm.Usage{PromptTokens: 1000, CompletionTokens: 100}}

	use(ledger.For(base, meter.SourceAgent), "demo-model")
	use(ledger.ForOrigin(base, meter.Origin{Source: meter.SourceSubagent, Label: "scout"}), "demo-model")
	use(ledger.ForOrigin(base, meter.Origin{Source: meter.SourceSubagent, Label: "scout"}), "demo-model")

	for _, e := range ledger.ByOrigin() {
		fmt.Printf("%s: %d requests, %d in, %d out, $%.4f\n", e.Origin, e.Requests, e.In, e.Out, e.Cost)
	}
	total := ledger.Total()
	fmt.Printf("total: %d requests, $%.4f\n", total.Requests, total.Cost)
	// Output:
	// agent: 1 requests, 1000 in, 100 out, $0.0012
	// sub-agent:scout: 2 requests, 2000 in, 200 out, $0.0024
	// total: 3 requests, $0.0036
}
