# AGENTS.md

`agentcore` is the reusable core of a terminal coding agent, built on
`github.com/rfizzle/llm`: the agent loop and the parts that go with it, as
independent Go packages that applications assemble through named seams. It is
consumed by shhh and hearth, and the module never learns which app calls it.
The design is in `docs/`; start at `docs/README.md`.

## Scope and restraint

Make the smallest change that is correct, safe and verifiable, following the
local pattern. Don't refactor, rename, reformat, update dependencies, rewrite
tests or expand documentation unless the requested behaviour needs it.

## Documentation

`project-docs` is the standard for the document set, comments and the
same-commit rule; `sound-patterns` settles style.

- Each document answers one question; `docs/README.md` is the table.
- Documents name no code symbol; code cites documents:
  `// See docs/<file>.md#<heading>`. Headings are anchors; renaming one
  means fixing its citations in the same commit.
- The package contract is the `doc.go` of each package; nothing in `docs/`
  restates it.
- A change that moves a command, path, behaviour or decision updates the
  document in the same commit. `make docs-check` verifies the mechanical
  part.
- Never a story, ticket or plan id in a document, comment or test name.

## Commands

| Task | Command |
|------|---------|
| Hermetic tests | `make test` |
| Static checks | `make fmt-check vet lint` |
| Rewrite formatting | `make fmt` |
| Exported API against the previous tag | `make apicheck` |
| Docs citations, paths, targets and ids | `make docs-check` |
| Shared skills match their source | `make skills-check` |
| Refresh shared skills from `../skills` | `make skills-sync` |
| Tests behind the integration tag | `make integration` |
| Everything CI runs | `make ci` |
| List every target | `make help` |

`make apicheck` compares the exported API of every package with the latest
tag reachable from `HEAD` (a checkout of it under `.cache/`), prints the diff
and the version bump it requires. Override the base with
`make apicheck API_BASE=v0.1.0`. It needs the tags locally; CI fetches them.

## Testing

- `make test` is hermetic: no listener, clipboard, daemon or network. The
  other tiers live behind the build tags in `TIER_TAGS` in the `Makefile`
  (`contract`, `integration`); `make vet` reads them, `make test` does not.
  No file carries a tag today; `make integration` is where one would run.
- Fakes sit at the seams: a scripted stream function and tool executor for the
  agent loop (`agent/fakes_test.go`), a `testing/fstest.MapFS` for skills.
- The examples (`example_test.go` in each package) are tests with
  `// Output:`; they run under `make test`.
- `seams_test.go` holds the Seams list in `doc.go` to the code both ways.
- **A test must not chdir, and `make test` takes no `-count=1`**: either makes
  the package uncacheable.

## Version control

- Never `git add -A`, `git add .` or `git commit -a`. Name the paths.
- Subject: `type(scope): subject`, imperative, lower case, under 72
  characters. A subject that needs "and" is two commits.
- A tag, once an app has recorded it, is never moved; a fix is a new patch
  version. Publishing a tag is the maintainer's step.
- The push is the person's. Never push.

## Architecture

Package doc comments (`doc.go`) are the detailed map.

    doc.go            the root package: the Seams list, no code
    agent/            the loop: conversation, tool dispatch, approvals, compaction, headless run
    repeat/           notices a session going in circles
    reading/          bounded calls outside the main turn (the summariser)
    retry/            the bounded wait for a request the provider never answered
    meter/            prices a usage and attributes it to an origin
    secret/           values the model may use but never read, scrubbed from results
    skill/            Agent Skills discovery and loading
    scripts/          docscheck and skillscheck, vendored from ../skills

Dependencies point down: `repeat` imports `agent` and `reading`; `agent`
imports `reading` and `retry`; `meter`, `secret`, `skill`, `reading` and
`retry` import no sibling.

## Invariants

- **No application is named.** No seam, name, default or comment refers to an
  app (`doc.go`).
- **A seam is listed.** Every exported `Set` function or method and every
  exported function-typed struct field is in the Seams list in `doc.go`, and
  every name listed is declared. `seams_test.go` fails otherwise.
- **A new seam needs a second consumer.** A hook one app asks for is built in
  that app until another asks for the same thing.
- **A nil seam keeps the module's own behaviour,** and the default is neutral.
- **A tag message names the incompatible API changes** from `make apicheck`,
  or says `api: compatible`.

## Gotchas

- **`apicheck: no tag to compare with`.** The clone has no tags; fetch them.
- **`goimports` must be on the `PATH`** for `make fmt-check`; it refuses to run
  without it. Install: `go install golang.org/x/tools/cmd/goimports@latest`.
- **Skills live in `.agents/skills`**; `.claude/skills` is a symlink. Shared
  ones are edited in `../skills`, then `make skills-sync`.
