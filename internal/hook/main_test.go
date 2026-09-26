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
	"fmt"
	"os"
	"testing"

	"go.uber.org/goleak"

	"github.com/modu-ai/moai-adk/internal/config"
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
//
// Finally it clears CLAUDE_PROJECT_DIR for the whole binary. The write-side
// resolver (resolveProjectRoot) prefers that variable over input.CWD, so a
// `go test` launched from inside a Claude Code session — which exports it —
// would otherwise send every parallel test's write to the real project tree
// instead of its temp dir (card t1165). Tests that need a value set it with
// t.Setenv, which restores this cleared state afterwards.
//
// The same reasoning covers the profile-lease store (card t1229). SessionStart
// enriches the row named by MOAI_PROFILE_LEASE_TOKEN (or inserts one when
// CLAUDE_CONFIG_DIR is set), and SessionEnd releases every row carrying the
// input's session id — "" included. Both resolve the database under MOAI_HOME,
// so a run launched from a lane session rewrote that lane's live lease with a
// test session id and a pid that exits with `go test`. MOAI_HOME therefore
// points at a directory owned by this binary, and the two lane variables are
// cleared. A helper process re-executed from this binary inherits the marker
// and keeps the MOAI_HOME its parent test chose rather than minting its own.
func TestMain(m *testing.M) {
	_ = os.Unsetenv(config.EnvClaudeProjectDir)
	_ = os.Unsetenv("MOAI_PROFILE_LEASE_TOKEN")
	_ = os.Unsetenv(config.EnvClaudeConfigDir)
	moaiHome := ""
	if os.Getenv(moaiHomeSandboxEnv) == "" {
		dir, err := os.MkdirTemp("", "moai-hook-test-home-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "TestMain: create MOAI_HOME sandbox: %v\n", err)
			os.Exit(1)
		}
		moaiHome = dir
		_ = os.Setenv(config.EnvHome, dir)
		_ = os.Setenv(moaiHomeSandboxEnv, dir)
	}
	deferredScanSeamMu.Lock()
	deferredScansAsync = false
	deferredScanSeamMu.Unlock()
	goleak.VerifyTestMain(m, goleak.Cleanup(func(code int) {
		if moaiHome != "" {
			_ = os.RemoveAll(moaiHome)
		}
		os.Exit(code)
	}))
}

// moaiHomeSandboxEnv marks a process whose MOAI_HOME was already sandboxed by
// this package's TestMain (card t1229).
const moaiHomeSandboxEnv = "MOAI_HOOK_TEST_HOME_SANDBOX"
