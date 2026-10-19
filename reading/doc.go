// Package reading is the bounded calls outside the main turn, and the session
// summarizer among them. Each is one request to a cheap model through the
// llm.Provider it is handed, no retries here, and a failed reading changes
// nothing.
//
// It is a leaf under the agent: the agent's own readers and the front-ends
// import it, and it imports neither the agent nor a front-end, so what the
// readers share (the model a reader asks with, the small bounds on what it is
// shown) lives in shared.go.
package reading
