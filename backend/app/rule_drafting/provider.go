package rule_drafting

import "context"

// Provider sends one system prompt and one user message to a language model
// and returns the raw text of its reply. Keeping it this small is what lets
// the tests and the offline checks run with a fake instead of a network call.
type Provider interface {
	Complete(ctx context.Context, system, user string) (string, error)
}
