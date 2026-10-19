// Package agentcore is the shared core of a terminal coding agent: the parts
// of an agent that do not depend on one application's screen, commands or
// storage. It builds on github.com/rfizzle/llm for the model providers and
// holds the agent loop's reusable pieces, such as metering, secrets and
// skills, as packages beneath it. The root package carries only this
// description; the code lives in the subpackages.
//
// # Seams
//
// An app shapes the module through the extension points below and nowhere
// else. Each line names one, as package.Name or package.Type.Name, and says
// what the app supplies. The tests in the root package hold this list to the
// code both ways: every exported Set function or method and every exported
// function-typed struct field is listed, and every name listed is declared.
//
//   - `agent.Agent.SetExecutor`: runs a tool call; the module has no tools of its own
//   - `agent.Agent.SetCallCancel`: cancels one call of a round without ending the round
//   - `agent.Agent.SetCallObserver`: sees each round's calls before they run
//   - `agent.Agent.SetTurnStart`: runs as each turn begins
//   - `agent.Agent.SetElidedWording`: the text that stands in for a tool result left out of the window
//   - `agent.Agent.SetResultParts`: turns a tool result into attachments the model also reads
//   - `agent.Agent.SetScrub`: rewrites a message before it is sent
//   - `agent.Agent.SetSteering`: the wordings of the steers and notices the loop adds
//   - `agent.Agent.SetFinished`: the line that marks a turn's work done
//   - `agent.Agent.SetCheckInInterval`: how many rounds pass between check-ins
//   - `agent.Agent.SetCheckInBudget`: the spend reading a check-in reports
//   - `agent.Agent.SetCheckInSteps`: the step reading a check-in reports
//   - `agent.Agent.SetInterveneBounds`: how soon and how often the loop may intervene
//   - `agent.Agent.SetProgressIntervals`: how many calls and how much time between status requests
//   - `agent.Agent.SetMaxRounds`: the per-turn round cap
//   - `agent.Agent.SetMessages`: replaces the conversation, to resume one
//   - `agent.Agent.SetFolded`: the messages a compaction has already folded
//   - `agent.Agent.SetTurn`: the turn counter, to resume one
//   - `agent.SetSessionSteerWording`: the wording of the steer a session's instruction gets
//   - `agent.SetLogger`: where the agent package logs
//   - `agent.Headless.SetRetryLimit`: the cap on retries of an unanswered request
//   - `agent.Headless.Gate`: decides which calls need approval
//   - `agent.Headless.Resolve`: decides and executes each gated call
//   - `agent.Compactor.Stream`: the stream a compaction asks its summary on
//   - `agent.Headless.Now`: the clock a run reads
//   - `agent.Headless.Steer`: messages that join the conversation between rounds
//   - `agent.Headless.Hold`: parks the run at a round tail until released
//   - `agent.Headless.Boundary`: a message the driver owes the conversation at a boundary
//   - `agent.Headless.OnClose`: the executable definition of done, asked when a turn ends
//   - `agent.Headless.OnText`: streamed assistant text
//   - `agent.Headless.OnProgress`: public status a checkpoint asked for
//   - `agent.Headless.OnToolCall`: a call about to run
//   - `agent.Headless.OnToolResult`: a result as it is recorded
//   - `agent.Headless.OnUsage`: per-request usage
//   - `agent.Headless.OnCompact`: a compaction's notice
//   - `agent.Headless.OnContinue`: a turn carried on after a stop
//   - `agent.Headless.OnIntervene`: an intervention the loop made
//   - `agent.Headless.OnRetry`: a retry of an unanswered request
//   - `agent.Headless.OnSummary`: a summary verdict
//   - `agent.Headless.OnWithheld`: a reading held back, with the reason
//   - `agent.RoundTail.Hold`: parks the turn at the boundary until released
//   - `agent.RoundTail.Steer`: whether a person's words joined at the boundary
//   - `agent.RoundTail.Target`: the instruction a steer quotes back
//   - `agent.RoundTail.Recover`: the window-recovery step at the tail
//   - `agent.RoundTail.HoldProgress`: defers the status request to a later boundary
//   - `agent.RoundTail.Boundary`: a message the driver owes the conversation at the boundary
//   - `agent.RoundTail.OnIntervene`: an intervention the tail made
//   - `agent.RoundTail.OnSummary`: a summary verdict
//   - `agent.RoundTail.OnWithheld`: a reading held back, with the reason
//   - `agent.Compactor.Workspace`: rewrites the system prompt after a rebuild
//   - `agent.Compactor.Before`: fires as a compaction is about to be asked for
//   - `agent.Compactor.After`: fires once the conversation is rebuilt
//   - `agent.Compactor.Carry`: state kept outside the conversation, appended under the summary
//   - `agent.NewRecorder`: the recorder of rows a summary is read from
//   - `agent.Digest.Arg`: how a call's argument is shown in a row
//   - `agent.Digest.Outcome`: how a result is shown in a row
//   - `agent.Digest.FirstLine`: how a line is cut for a row
//   - `reading.SetLogger`: where the reading package logs
//   - `reading.SummaryConfig.ModelAt`: the model a reading is asked of
//   - `reading.SummaryConfig.CadenceAt`: how often readings are taken
//   - `reading.SummaryConfig.Prompt`: replaces the built-in reading instruction
//   - `repeat.Shapes.Question`: which calls ask the person a question
//   - `repeat.Shapes.SearchScope`: the scope a search call covers
//   - `repeat.Shapes.Wrote`: which calls write
//   - `retry.SetLogger`: where the retry package logs
//   - `retry.Backoff.SetLimit`: the cap on waits
//   - `meter.Source`: the app's own kinds of spend, as constants of its own
//   - `meter.Ledger.SetBudget`: the spend limit a ledger enforces
//   - `meter.Ledger.SetCacheLease`: the lease a cache entry is held under
//   - `meter.Ledger.SetClock`: the clock a ledger reads
//   - `secret.Vault.SetEnvMask`: whether values are masked in the environment
//   - `skill.Load`: discovers skills from roots and an fs.FS of built-ins
//
// Two rules keep the list short. The module never learns which app is calling
// it: no seam, name, default or comment refers to an application. And a new
// seam needs a second consumer that wants it: a hook one app asks for is built
// in that app until another app asks for the same thing.
package agentcore
