package preference

import (
	"fmt"
	"strings"
)

// longFact builds a Fact string long enough that ~20 entries push core.yaml
// past the 4KB cap, forcing demotion in the overflow test.
func longFact(i int) string {
	// ~300 bytes of deterministic padding per entry.
	pad := strings.Repeat("x", 300)
	return fmt.Sprintf("user prefers option %d with long fact padding %s", i, pad)
}

// keyFor produces a deterministic decision_key for the overflow test.
func keyFor(i int) string {
	return fmt.Sprintf("overflow_decision_%d", i)
}
