// Package secret holds the values a session may use but the model may never
// see: API keys, tokens, passphrases. A secret is declared by name, handed to
// every command the model runs as an environment variable of that name, and
// scrubbed from everything that comes back — tool results, command output, the
// user's own typing — before any of it reaches a provider or a file that
// outlives the session.
package secret
