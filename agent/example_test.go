package agent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rfizzle/agentcore/agent"
	"github.com/rfizzle/llm"
)

// script is a StreamFunc that serves one scripted reply per request, which is
// all a model is to the loop: a channel of events and a cancel. An app's
// closure opens its provider here instead.
func script(replies ...[]llm.StreamEvent) agent.StreamFunc {
	i := 0
	return func([]llm.Message, string) (<-chan llm.StreamEvent, context.CancelFunc, error) {
		evs := replies[i]
		i++
		ch := make(chan llm.StreamEvent, len(evs))
		for _, ev := range evs {
			ch <- ev
		}
		close(ch)
		return ch, func() {}, nil
	}
}

func say(text string) []llm.StreamEvent {
	return []llm.StreamEvent{{Token: text}, {Done: true}}
}

// An Agent is a passive state machine: the app starts a turn, opens the
// stream the agent was given, and reads the events itself.
func ExampleNew() {
	a := agent.New(
		[]llm.Message{{Role: llm.RoleSystem, Content: "You are terse."}},
		script(say("Hello there.")),
	)
	a.StartTurn("Say hello.")

	events, cancel, err := a.Stream(a.RequestMessages())
	if err != nil {
		panic(err)
	}
	defer cancel()
	var reply strings.Builder
	for ev := range events {
		reply.WriteString(ev.Token)
	}
	a.Append(llm.Message{Role: llm.RoleAssistant, Content: reply.String()})

	for _, m := range a.Messages() {
		fmt.Printf("%s: %s\n", m.Role, m.Content)
	}
	// Output:
	// system: You are terse.
	// user: Say hello.
	// assistant: Hello there.
}

// A tool round is split by the approval gate: the calls the gate lets go run
// at once through the executor, and the ones it holds wait in a queue for the
// app to decide. Both kinds of result land in the conversation under the id
// of the call they answer.
func ExampleAgent_BeginToolRound() {
	a := agent.New(nil, nil)
	a.SetExecutor(func(_ context.Context, name string, _ json.RawMessage) (string, error) {
		return "ran " + name, nil
	})
	gate := func(tc llm.ToolCall) bool { return tc.Name == "write_file" }

	calls := []llm.ToolCall{
		{ID: "c1", Name: "read_file", Arguments: `{"path":"a.txt"}`},
		{ID: "c2", Name: "write_file", Arguments: `{"path":"a.txt"}`},
	}
	auto, gated := a.BeginToolRound("Let me look, then change it.", calls, gate)
	fmt.Println("auto:", auto[0].Name, "gated:", gated[0].Name)

	a.RecordAutoResults(a.ExecuteCalls(context.Background(), auto))

	call, _ := a.NextApproval()
	a.ResolveApproval("error: the user declined " + call.Name)

	for _, m := range a.Messages()[1:] {
		fmt.Printf("%s %s: %s\n", m.Role, m.ToolCallID, m.Content)
	}
	// Output:
	// auto: read_file gated: write_file
	// tool c1: ran read_file
	// tool c2: error: the user declined write_file
}

// Headless drives a whole turn without a screen. Gate says which calls need a
// decision and Resolve makes it: in an unattended run approval is a policy,
// not a prompt.
func ExampleHeadless_Run() {
	a := agent.New(nil, script(
		[]llm.StreamEvent{{ToolCalls: []llm.ToolCall{
			{ID: "c1", Name: "run_command", Arguments: `{"command":"rm -rf build"}`},
		}}},
		say("I left the build directory alone."),
	))
	h := &agent.Headless{
		Agent: a,
		Gate:  func(tc llm.ToolCall) bool { return tc.Name == "run_command" },
		Resolve: func(tc llm.ToolCall) string {
			return "error: policy refuses " + tc.Name
		},
		OnToolResult: func(r agent.ToolResult) { fmt.Println(r.Call.Name, "->", r.Result) },
	}

	final, err := h.Run("Clean the build directory.")
	fmt.Println(final, err)
	// Output:
	// run_command -> error: policy refuses run_command
	// I left the build directory alone. <nil>
}

// A Compactor is the window-recovery step a driver runs at a round boundary.
// Under the line it does nothing; over it, old tool results are elided first
// and, where that is not enough, the conversation is replaced by a summary of
// itself. The driver supplies the ask: how a summary request is made and
// waited on.
func ExampleCompactor() {
	filler := strings.Repeat("word ", 1600)
	a := agent.New([]llm.Message{
		{Role: llm.RoleSystem, Content: "You are terse."},
		{Role: llm.RoleUser, Content: filler},
		{Role: llm.RoleAssistant, Content: filler},
		{Role: llm.RoleUser, Content: "Carry on."},
		{Role: llm.RoleAssistant, Content: "Yes."},
	}, nil)
	c := &agent.Compactor{Model: "demo-model", Window: 4000}

	n := c.Recover(a, func(_ []llm.Message, choice string) (string, error) {
		fmt.Println("summary asked with tools:", choice)
		return "The user asked for nothing yet.", nil
	})
	fmt.Println("compacted:", n.Compacted, "kept turns:", n.Kept)
	for _, m := range a.Messages() {
		fmt.Println(m.Role)
	}
	// Output:
	// summary asked with tools: none
	// compacted: true kept turns: 1
	// system
	// user
	// user
	// assistant
}
