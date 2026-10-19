package retry_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/rfizzle/agentcore/retry"
	"github.com/rfizzle/llm/failure"
)

// A Backoff is the decision and not the wait: Next says whether the failure
// earns another attempt and how long to wait first, and the driver does the
// waiting. The bound counts across the whole stall, and a failure waiting
// cannot fix ends it at once.
func ExampleBackoff() {
	var b retry.Backoff
	busy := &failure.Failure{Class: failure.ClassOverloaded, Status: 529, RetryAfter: 2 * time.Second}

	for {
		n, ok := b.Next(busy)
		if !ok {
			fmt.Println("giving up after", b.Attempt()-1, "retries")
			break
		}
		// A real driver sleeps n.Wait here, or lets a person press past it.
		fmt.Printf("attempt %d of %d, waits at least the provider's %s: %t\n",
			n.Attempt, n.Max, busy.RetryAfter, n.Wait >= busy.RetryAfter)
	}

	b.Reset()
	_, ok := b.Next(&failure.Failure{Class: failure.ClassAuth})
	fmt.Println("a rejected key is retried:", ok)

	_, ok = b.Next(errors.New("not a provider failure"))
	fmt.Println("an unclassified error is retried:", ok)
	// Output:
	// attempt 1 of 3, waits at least the provider's 2s: true
	// attempt 2 of 3, waits at least the provider's 2s: true
	// attempt 3 of 3, waits at least the provider's 2s: true
	// giving up after 3 retries
	// a rejected key is retried: false
	// an unclassified error is retried: false
}
