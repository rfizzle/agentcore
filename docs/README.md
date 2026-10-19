# agentcore · documents

Each file answers one question. Facts live here once; `AGENTS.md` says where
the code is; skills say how a change is made.

| File | Question it answers | Changes when |
|---|---|---|
| `architecture.md` | What are the big shapes, and why those? | a shape changes |
| `decisions.md` | What was decided, when and why? | a decision is made |

- The API is documented in each package's `doc.go` and doc comments; nothing
  here restates it.
- Documents name no code symbol; code cites documents by path and heading.
  `make docs-check` fails on a citation that does not resolve and lists
  headings nothing cites.
- Planning is not in this repository. Nothing here names a story, ticket or
  sprint.
