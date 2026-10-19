package secret_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rfizzle/agentcore/secret"
)

// A Vault holds the values the model may use but never read. Declared by
// name, a secret is handed to commands as an environment variable of that
// name and replaced by a placeholder in everything that comes back, so a
// tool that prints one leaks the name and not the value.
func ExampleVault_WrapExecutor() {
	v := secret.New()
	if err := v.Add("DEPLOY_TOKEN", "tok-4f9a1c7e2b8d"); err != nil {
		panic(err)
	}

	// An executor that surfaces the value, as a command printing its
	// environment would.
	leaky := func(_ context.Context, _ string, _ json.RawMessage) (string, error) {
		return "Authorization: Bearer tok-4f9a1c7e2b8d", nil
	}
	safe := v.WrapExecutor(leaky)

	out, _ := safe(context.Background(), "run_command", nil)
	fmt.Println(out)
	fmt.Println(v.Scrub("only the first half leaks: tok-4f9a1c7e"))
	// Output:
	// Authorization: Bearer [secret:DEPLOY_TOKEN]
	// only the first half leaks: [secret:DEPLOY_TOKEN]
}
