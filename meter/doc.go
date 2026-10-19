// Package meter is session spend accounting. Every request an app makes — the
// agent's own rounds, the permission classifier, the session summary, each
// sub-agent's turns — is a call to a provider, so the provider is the one thing
// a feature cannot route around. The gate wraps a provider and attributes what
// passes through it to an origin; the ledger behind the gate adds it up.
//
// The alternative, and what this replaces, is each feature remembering to
// report what it spent. That works until it doesn't: a feature added later
// counts nothing, and nothing fails — the session simply under-reports, which
// is the one kind of wrong a spend meter must never be.
package meter
