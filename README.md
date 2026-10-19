# agentcore

The reusable core of a terminal coding agent: the agent loop and the parts
that go with it, with none of any app's screens, commands or storage, built on
[`github.com/rfizzle/llm`](https://github.com/rfizzle/llm).

Status: v0, filling in package by package. The API may change between minor
versions; read the commit messages between two tags before moving a pin.

## Install

    go get github.com/rfizzle/agentcore

An app that uses `agentcore` also requires `llm` and imports the provider
backends it ships (see the `llm` README).

## Use

Import the packages you need: `agent`, `repeat`, `reading`, `retry`, `meter`,
`secret` and `skill`. Each is independent, and the root package's
documentation (`doc.go`) lists the seams an app fills to shape them. The
runnable examples in each package's `example_test.go` show the common uses and
run under `go test` with no network; the API is on
[pkg.go.dev](https://pkg.go.dev/github.com/rfizzle/agentcore).

The design is in `docs/`; start at `docs/README.md`. MIT licence; see
`LICENSE`.
