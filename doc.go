// Package agentcore is the shared core of a terminal coding agent: the parts
// of an agent that do not depend on one application's screen, commands or
// storage. It builds on github.com/rfizzle/llm for the model providers and
// holds the agent loop's reusable pieces, such as metering, secrets and
// skills, as packages beneath it. The root package carries only this
// description; the code lives in the subpackages.
package agentcore
