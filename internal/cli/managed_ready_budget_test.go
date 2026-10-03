package cli

// managed_ready_budget_test.go — card t1410, F9 handshake budget.

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestManagedCodexReadyTimeoutBudget pins F9: the one handshake budget is the
// config constant (an unmeasured relaxation), and both consumers derive from
// it rather than restating a duration.
func TestManagedCodexReadyTimeoutBudget(t *testing.T) {
	if got := config.DefaultManagedCodexReadyTimeout; got != 30*time.Second {
		t.Fatalf("DefaultManagedCodexReadyTimeout = %v, want 30s", got)
	}
	src, err := os.ReadFile("managed_codex_factory.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"websocket.Dialer{HandshakeTimeout: config.DefaultManagedCodexReadyTimeout}",
		"context.WithTimeout(context.Background(), config.DefaultManagedCodexReadyTimeout)",
	} {
		if n := strings.Count(string(src), want); n != 1 {
			t.Fatalf("%q appears %d times, want exactly 1", want, n)
		}
	}
}
