# agentcore

The reusable core of a terminal coding agent, shared by more than one
application. It holds the agent loop and the parts that go with it, and none of
any app's screens, commands or storage.

```
go get github.com/rfizzle/agentcore
```

It builds on [`github.com/rfizzle/llm`](https://github.com/rfizzle/llm), which
supplies the message, stream, usage and pricing types and the model providers.
An app that uses `agentcore` also requires `llm`, and imports the backends it
ships (see the `llm` README).

## What it holds, and what it doesn't

Each package does one job:

- `agent`: the loop. An `Agent` is a passive state machine over the
  conversation (messages, the stream it asks a model for, tool dispatch, the
  approval queue, the per-turn round cap, check-ins and steers). `Headless`
  drives it to completion without a screen; an interactive surface drives it
  step by step. `Compactor` recovers the context window, `Recorder` keeps the
  rows a summary is read from, and `SummaryRun` schedules the readings.
- `repeat`: notices a session going in circles and says so in the result the
  model reads (`RepeatDetector`).
- `reading`: the bounded calls outside the main turn, chiefly the summarizer
  that judges whether a run still serves its instruction.
- `retry`: the bounded wait for a request the provider never answered
  (`Backoff`). It decides; the driver does the waiting.
- `meter`: spend accounting. A `Ledger` wraps a provider so every request is
  billed to an origin, prices it, and enforces a budget.
- `secret`: a `Vault` of values the model may use but never read, scrubbed from
  every tool result.
- `skill`: Agent Skills discovery and loading (`SKILL.md` directories, plus
  built-ins from an `fs.FS`).

It does not hold the app's policy. These stay in the app:

- the tools: which exist, what each does, and which are gated;
- the screens and commands a person reads and types, and the wording the person
  sees;
- where anything is stored: sessions, evidence, settings and credentials;
- the choice of provider, model and prices, which are `llm`'s to resolve and
  the app's to choose.

Nothing in the module names an app. Where the loop needs a word or a decision
only an app can supply, there is a seam, and the default is neutral.

## The seams an app fills

Each seam is a field, a setter or a constructor argument, and a nil one keeps
the module's own behaviour.

| Seam | Package | What it is for |
| --- | --- | --- |
| `SetLogger` | `agent`, `reading`, `retry` | Where the package writes down a retry, a trim or a compaction. Each discards until it is given a `*slog.Logger`. |
| `Agent.SetExecutor`, `Agent.SetScrub` | `agent` | The executor runs the calls the gate lets through. The scrub rewrites every message on the way in and out; build it on `secret.Vault.Scrub`. |
| `Agent.SetResultParts` | `agent` | Collects the non-text parts (an image) a tool attached to a result. They are lost if it is not set. |
| `Agent.SetElidedWording` | `agent` | The sentence a trim leaves in place of an archived result. It must carry the id after `full output stored as evidence `. |
| `Agent.SetCallObserver` | `agent` | Told the calls of each round as the round begins. |
| `Headless.Boundary` | `agent` | Asked at a turn's start and between rounds whether the run is owed a message. |
| `SetSessionSteerWording` | `agent` | The sentence that frames a line another session sent. An app names itself in it. |
| `NewRecorder(max, Digest)` | `agent` | The words a recorded row is made of: `Arg`, `Outcome` and `FirstLine`. |
| `NewRepeatDetector(Shapes)` | `repeat` | Which tools are a question, a search of a place, or a write. |
| `skill.Load(roots, fsys)` | `skill` | The built-in skills, read from `builtin/*/SKILL.md` in the `fs.FS` the app passes (usually an `embed.FS`). A nil one carries none. |
| `SummaryConfig.Prompt` | `reading` | Replaces the summarizer's instruction. Empty keeps the built-in. |

## Build an agent

An `Agent` is given the conversation it starts from and a `StreamFunc`, the
closure that opens a completion stream. The provider, the model and the request
options live inside the closure.

```go
a := agent.New(
	[]llm.Message{{Role: llm.RoleSystem, Content: "You are terse."}},
	func(msgs []llm.Message, choice string) (<-chan llm.StreamEvent, context.CancelFunc, error) {
		// Open the app's provider here.
	},
)
a.StartTurn("Say hello.")
events, cancel, err := a.Stream(a.RequestMessages())
```

This is `ExampleNew` in `agent/example_test.go`, run against a scripted stream
so it needs no network.

## Split a tool round with a gate

`BeginToolRound` records the assistant's message and splits its calls by the
`ApprovalGate`: the calls the gate lets go run through the executor, and the
rest wait in a queue for the app to decide.

```go
auto, gated := a.BeginToolRound(text, calls, gate)
a.RecordAutoResults(a.ExecuteCalls(ctx, auto))
call, _ := a.NextApproval()
a.ResolveApproval("error: the user declined " + call.Name)
```

See `ExampleAgent_BeginToolRound`. An interactive surface asks a person at
`NextApproval`; a scripted one decides by policy.

## Run a turn without a screen

`Headless.Run` drives a whole turn: stream, dispatch, record, repeat until the
model answers without a tool or the round cap is hit. `Gate` says which calls
need a decision and `Resolve` makes it, so in an unattended run approval is
policy and not a prompt.

```go
h := &agent.Headless{
	Agent:   a,
	Gate:    func(tc llm.ToolCall) bool { return tc.Name == "run_command" },
	Resolve: func(tc llm.ToolCall) string { return "error: policy refuses " + tc.Name },
}
final, err := h.Run("Clean the build directory.")
```

See `ExampleHeadless_Run`. The `On*` hooks, `Steer`, `Hold`, `OnClose` and
`Compact` are the other fields a driver fills; each is documented where it is
declared and each may be nil. `agent.UnlimitedToolRounds` passed to
`SetMaxRounds` removes the per-turn cap for a run a person is watching.

## Keep a long run inside its window

A `Compactor` is installed as `Headless.Compact`, or called by an interactive
driver at a round boundary with `Recover`. Under the line it does nothing; over
it, old tool results are elided first and, where that is not enough, the
conversation is replaced by a summary of itself. The driver owns the request
and the wait, so it hands `Recover` a `CompactAsk`. See `ExampleCompactor`.

## Retry a stalled request

`retry.Backoff` counts the attempts one stall has used and says how long to wait
before each. `Next(err)` is false for a failure waiting cannot fix and once the
bound is used up. The waiting itself is the driver's, so a surface can show a
countdown or a person can press past it. `Reset` ends the stall when the
provider answers. See `ExampleBackoff`.

## Meter the spend

`meter.New(prices)` takes an `llm/pricing` table. `Ledger.For` and
`Ledger.ForOrigin` wrap a provider so everything it streams is billed to a
source, or to one named requester such as a single sub-agent, and the totals
read back by source, origin or model. `SetBudget` sets a warning and a cap in
cents. See `ExampleLedger_ForOrigin`.

## Hide secrets from the model

A `secret.Vault` holds values declared by name. They reach commands as
environment variables of that name and are replaced by `[secret:NAME]` in
everything that comes back, including partial matches of eight bytes or more.
`WrapExecutor` scrubs every tool result; `Scrub` takes any string. See
`ExampleVault_WrapExecutor`.

## Load skills

`skill.Load(roots, fsys)` discovers `SKILL.md` directories under the roots the
app names, then adds the built-ins from `fsys`. `skill.PromptBlock` is the list
of names and descriptions the model reads, and a skill's body is read only when
it is activated. See `ExampleLoad`, which loads from a `testing/fstest.MapFS`.

## Versions and pinning

`agentcore` is versioned by tags, `v0.N.0`. A change to the module is committed
and tagged in this repository, and an app requires the tag it was built
against, together with the `llm` tag this module requires. A tag, once an app
has recorded it, is never moved: a fix is a new patch version, because module
caches and `go.sum` already hold the old content. Publishing a tag is the
maintainer's step, and it comes before any app commit that requires it is
pushed.

Before `v1.0.0` the surface is allowed to change between minor versions; read
the commit messages between two tags before moving a pin.

## Versioning

Every tag message lists the incompatible API changes the tag carries, or says
`api: compatible`. `make apicheck` produces the list: it compares the exported
API of every package with the previous tag and prints the version bump the
difference requires, and the gate runs it, so a tag message is written from
that output and not from memory.

The module stays at v0 until both applications that use it have shipped a
milestone on one tag without needing a change to it. Freezing the API earlier
would lock in the assumptions of whichever app wrote it first. From then on it
is v1, with Go's compatibility promise. Because the exported API carries `llm`
types, an incompatible `llm` change bumps this module in lockstep, and the
module and both apps sit on one `llm` tag.

## Checks

`make help` lists them. The gate is `make ci`, which is `fmt-check vet lint test apicheck`; the
examples above are ordinary tests and run under `test` with no network.

## License

MIT. See `LICENSE`.
