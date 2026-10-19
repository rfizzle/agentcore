// Package agent owns the front-end-agnostic agentic loop state: the
// conversation message list, stream requests, tool dispatch, the queue of
// approval-gated tool calls, and the per-turn iteration guard. Front-ends (an
// interactive surface, or the headless runner in this package) drive an Agent
// step by step and surface approvals and progress to the user themselves.
package agent
