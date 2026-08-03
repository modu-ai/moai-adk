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
// There is no ignore list. An IgnoreTopFunction entry for
// internal/hook/trace.(*TraceWriter).run used to sit here, masking a writer
// goroutine that outlived its test. That goroutine leaked because nothing
// crossed the writer's flush barrier — the same missing teardown that lost
// trace entries in production. SPEC-HOOK-TRACE-FLUSH-001 gave the registry a
// Shutdown path, the one test that detached a writer without closing it now
// calls it, and the suppression was removed rather than re-justified.
//
// Restoring a suppression here re-hides that class of defect: a leaked writer
// goroutine means entries are still queued, so a green package would once
// again say nothing about whether traces reach disk.
package hook

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain enables goroutine leak detection across all internal/hook tests.
// AC-HAE-007 verifies zero leaks reported; the package runs with no
// TraceWriter exemptions at all (SPEC-HOOK-TRACE-FLUSH-001 gave the
// registry a Shutdown path, so a leaked writer goroutine is no longer
// masked).
//
// It also flips deferredScansAsync to false for the test binary: dozens of
// Handle-calling tests do not install the deferred-scan join seam, so the
// async path would leak the advisory goroutine past their test boundary and
// its slog writes would race unrelated parallel tests that mutate os.Stderr
// / the slog handler. The inline-sync path eliminates the goroutine entirely.
// session_start_parallel_test.go opts back into async=true per-test to keep
// the production async path covered.
func TestMain(m *testing.M) {
	deferredScanSeamMu.Lock()
	deferredScansAsync = false
	deferredScanSeamMu.Unlock()
	goleak.VerifyTestMain(m)
}
