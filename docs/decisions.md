# agentcore · decisions

These five were decided on 2026-10-19 in hearth's design, where hearth's
decision log records them against this module, and are recorded here as the
module's own. A decision is recorded once, here; other documents state the
result. Ids are never reused; a superseded row keeps its text and changes only
its status.

| ID | Question | Decision | Why | Date | Status |
|---|---|---|---|---|---|
| D1 | What is the module called, and what does its root hold? | `github.com/rfizzle/agentcore`; the root holds no code. | The name never appears in a selector. | 2026-10-19 | decided |
| D2 | Does shhh's memory package move into the module? | No. Memory stays in each app. | Its kinds and wordings are shhh's, and hearth's per-bot notes (visible remember calls, consolidation, pins) are a different model. | 2026-10-19 | decided |
| D3 | The agent package exceeds astimate's ceilings (about 2,450 sloc and 230 exports): exemption or split first? | A standing exemption for the agent package in `astimate.yaml`; the two follow-up splits wait until a second consumer imports the package, which decides the cut. | The brief ruled out rewriting the agent's surface during the move. | 2026-10-19 | decided |
| D4 | How stable is the API, and when does it earn v1? | v0 until shhh and hearth both ship a milestone on one tag without needing a change to the module; then v1 with Go's compatibility promise. Until then the gate reports the exported-API diff against the previous tag, and every tag message names its incompatible changes or says it has none. | At v0 both consumers are the same person and the API is still shhh-shaped; freezing it before hearth consumes it would lock in shhh's assumptions. An incompatible llm change bumps this module in lockstep, and the module and both apps sit on one llm tag. | 2026-10-19 | decided |
| D5 | What may the module know about its callers, and when does a seam earn a place? | The module never learns which app calls it. A new seam needs a second consumer that wants it. | A hook one app asks for is built in that app until another app asks for the same thing. Why the module must not know its caller: reason not recorded. | 2026-10-19 | decided |
