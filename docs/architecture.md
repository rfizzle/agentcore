# agentcore · architecture

## A library of seams

The module is the reusable core of a terminal coding agent: the parts of an
agent that do not depend on one application's screens, commands or storage.
It holds no application. An app assembles the packages it needs and shapes
them through a short list of named extension points, the seams, and through
nothing else. Where the loop needs a word or a decision only an app can
supply, there is a seam, and a seam left empty keeps the module's own neutral
behaviour. The root of the module holds no code; it carries the list of seams
as its documentation, and a test holds that list to the code in both
directions.

## The agent loop is one bot's turn

The loop is a passive state machine over one conversation: the messages, the
request for the next model stream, the dispatch of tool calls, the queue of
calls that wait for an approval, and a cap on the rounds of one turn. It does
not run itself. A surface drives it step by step, or a headless driver runs a
turn to completion, with approval decided by policy rather than by a prompt.
The module has no tools of its own; the app supplies the executor. Keeping the
loop passive is what lets an interactive screen and an unattended run share
one implementation.

## The meter prices a usage

Every request an app makes, whether the agent's own rounds, a summary or a
sub-agent's turns, is a call to a provider, so the provider is the one thing a
feature cannot route around. The meter wraps the provider, prices what passes
through it and attributes it to the requester that spent it. The alternative is
each feature remembering to report what it spent, which works until a feature
added later counts nothing and nothing fails: the session under-reports, the
one kind of wrong a spend meter must never be. What an app does with the
total, such as the limit it sets, is the app's.

## Four independent packages

The secret, skill, reading and retry packages each do one job and stand apart
from the loop.

- Secret holds values the model may use but never read, and scrubs them from
  everything that comes back. It knows nothing of the loop, so any surface
  that shows the model text can put it in front.
- Skill discovers and loads Agent Skills in the published format, so a skill
  written for another harness loads unchanged, and one written here loads
  there.
- Reading holds the bounded calls outside the main turn, chiefly the summary
  that judges whether a run still serves its instruction. It is a leaf under
  the loop: the loop imports it and it imports neither the loop nor a surface.
- Retry decides how long to wait and how many times when the provider never
  answered. The waiting is the driver's, so a surface can show a countdown or
  a person can press past it.

## Two rules keep the list short

The module never learns which app is calling it: no seam, name, default or
comment refers to an application. And a new seam needs a second consumer that
wants it: a hook one app asks for is built in that app until another app asks
for the same thing. Together they keep the list of seams short.
