// Package hook — main_test.go
//
// SPEC-V3R6-HOOK-ASYNC-EXPAND-001 M5 / AC-HAE-007: goroutine leak detection
// for the entire internal/hook test binary.
//
// goleak.VerifyTestMain runs after every Test* function in the package and
// reports any goroutine that did not terminate. The four async handlers
// (FileChanged, ConfigChange, TaskCreated, Notification) MUST self-cancel
// via context.WithTimeout(context.Background(), asyncDeadline) before the
// test binary exits — failure to do so indicates a deadline-enforcement bug
// per REQ-HAE-005.
//
// IgnoreTopFunction entries below cover pre-existing background goroutines
// that are NOT in scope of this SPEC (REQ-HAE scope is the 4 target hook
// handlers only). These ignores were verified at M5 baseline capture:
//
//   - internal/hook/trace.(*TraceWriter).run — long-lived background writer
//     goroutine that exits when its channel is closed. Some tests construct
//     a TraceWriter without explicitly closing it. Owned by SPEC-V3R2-RT-006
//     observability subsystem.
package hook

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain enables goroutine leak detection across all internal/hook tests.
// AC-HAE-007 verifies zero leaks reported beyond the documented ignore list.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// Pre-existing background goroutine from observability subsystem.
		// Out of scope for SPEC-V3R6-HOOK-ASYNC-EXPAND-001.
		goleak.IgnoreTopFunction("github.com/modu-ai/moai-adk/internal/hook/trace.(*TraceWriter).run"),
	)
}
