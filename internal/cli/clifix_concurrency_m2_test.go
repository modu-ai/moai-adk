package cli

// SPEC-CLIFIX-CONCURRENCY-001 M2 — Reproduction tests (RED → GREEN).
//
// These tests verify AC-CONC-001-003: the ~/.claude.json read-modify-write at
// glm_tools.go (runEnableMCPServerForTool / enableMCPServerIdempotentForTool)
// must not lose a concurrent writer's update.
//
// TestClaudeJSONGuard_ConcurrentForceNoLostUpdate:
//   Exercises the --force path (runEnableMCPServerForTool). Three goroutines,
//   one per tool (vision/websearch/webreader), race the RMW simultaneously.
//   Pre-M2: each goroutine reads the same empty state and overwrites the
//   others' mcpServers entries — keys are LOST. This test FAILS.
//   Post-M2: the RMW is guarded by mutateClaudeJSONAtomic (flock +
//   compare-retry); the flock serializes cooperating writers and the
//   compare-retry re-applies the mutation to the fresh in-lock state. All
//   three keys survive. This test PASSES.
//
// TestClaudeJSONGuard_ConcurrentIdempotentNoLostUpdate:
//   Same race shape but via the idempotent path (enableMCPServerIdempotent-
//   ForTool). Identical RED→GREEN semantics.
//
// Per SPEC §D.5, each test MUST demonstrably fail against the pre-fix commit.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// claudeJSONConcurrentTools lists one distinct tool per goroutine so each
// writes a DIFFERENT mcpServers key — the lost-update manifests as a missing
// key when two goroutines read the same state and the later rename overwrites
// the earlier one.
var claudeJSONConcurrentTools = []string{"vision", "websearch", "webreader"}

// expectedClaudeJSONKeys is the set of mcpServers keys the three concurrent
// tools must collectively register.
var expectedClaudeJSONKeys = []string{zaiMCPServerKey, zaiWebSearchPrimeKey, zaiWebReaderKey}

// runClaudeJSONConcurrentEnable races one goroutine per tool (released
// simultaneously via a start barrier to maximize the interleaving window),
// then asserts every tool's mcpServers key survived the concurrent RMW.
func runClaudeJSONConcurrentEnable(t *testing.T, configPath string, enable func(configPath, tool, token string) error) {
	t.Helper()
	const iterations = 200
	const token = "shared-concurrency-token"
	for iter := 0; iter < iterations; iter++ {
		// Re-seed an empty mcpServers map each iteration so every goroutine
		// starts from a known-empty baseline (maximizes the lost-update window).
		if err := os.WriteFile(configPath, []byte(`{"mcpServers":{}}`), 0o600); err != nil {
			t.Fatalf("iter %d: seed write: %v", iter, err)
		}

		var wg sync.WaitGroup
		start := make(chan struct{})
		var errCount int64
		for _, tool := range claudeJSONConcurrentTools {
			wg.Add(1)
			go func(tool string) {
				defer wg.Done()
				<-start // release all goroutines simultaneously
				if err := enable(configPath, tool, token); err != nil {
					atomic.AddInt64(&errCount, 1)
				}
			}(tool)
		}
		close(start)
		wg.Wait()

		if errCount > 0 {
			t.Fatalf("iter %d: concurrent enable returned %d errors (RMW corruption)", iter, errCount)
		}

		// Assert ALL three MCP server keys survived the concurrent RMW.
		root, err := readClaudeJSON(configPath)
		if err != nil {
			t.Fatalf("iter %d: readClaudeJSON: %v", iter, err)
		}
		mcpServers := getMCPServers(root)
		for _, key := range expectedClaudeJSONKeys {
			if _, ok := mcpServers[key]; !ok {
				t.Fatalf("iter %d: MCP server %q lost (concurrent RMW lost update — guard missing or ineffective)", iter, key)
			}
		}
	}
}

// TestClaudeJSONGuard_ConcurrentForceNoLostUpdate — AC-CONC-001-003 (--force path).
func TestClaudeJSONGuard_ConcurrentForceNoLostUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude.json")
	runClaudeJSONConcurrentEnable(t, configPath, func(cp, tool, token string) error {
		return runEnableMCPServerForTool(cp, tool, token)
	})
}

// TestClaudeJSONGuard_ConcurrentIdempotentNoLostUpdate — AC-CONC-001-003 (idempotent path).
func TestClaudeJSONGuard_ConcurrentIdempotentNoLostUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude.json")
	runClaudeJSONConcurrentEnable(t, configPath, func(cp, tool, token string) error {
		_, err := enableMCPServerIdempotentForTool(cp, tool, token)
		return err
	})
}

// TestClaudeJSONGuard_ExternalWriteDetectedByCompareRetry — AC-CONC-001-003
// deterministic variant. The concurrent tests above exercise the flock
// serialization (cooperating writers). This test exercises the compare-and-
// retry path against a NON-cooperating writer (simulating a live Claude Code
// process that does not respect the .lock file) by injecting an external raw
// write between the guard's prep-read and its in-lock compare via the
// claudeJSONGuardPreLockHook. The guard MUST detect the changed bytes,
// re-apply the mutation on the fresh state, and preserve BOTH the external
// write and the MCP entry. This is 100% deterministic (no goroutine scheduling
// dependency).
//
// This test is GREEN-only (it references claudeJSONGuardPreLockHook which does
// not exist pre-M2). The concurrent tests above serve as the RED evidence.
//
// NOTE: deliberately NOT t.Parallel() — this test writes the package-global
// claudeJSONGuardPreLockHook, which mutateClaudeJSONAtomic reads on every call.
// Running it in parallel with any other test that exercises the guarded RMW
// (including the concurrent tests above AND pre-existing GLMTools tests that
// call enableMCPServerIdempotent) would be a data race on the global. Serial
// execution guarantees the hook is set, exercised, and restored before any
// parallel test reads it.
func TestClaudeJSONGuard_ExternalWriteDetectedByCompareRetry(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude.json")
	if err := os.WriteFile(configPath, []byte(`{"mcpServers":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// Inject an external raw write between the prep-read and the in-lock compare,
	// simulating a live Claude Code process writing a non-MCP top-level key.
	origHook := claudeJSONGuardPreLockHook
	claudeJSONGuardPreLockHook = func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var root map[string]any
		if json.Unmarshal(data, &root) != nil {
			return
		}
		root["externalSetting"] = "from-cc"
		out, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			return
		}
		_ = os.WriteFile(path, out, 0o600)
	}
	t.Cleanup(func() { claudeJSONGuardPreLockHook = origHook })

	if err := runEnableMCPServerForTool(configPath, "websearch", "tok"); err != nil {
		t.Fatalf("runEnableMCPServerForTool: %v", err)
	}

	root, err := readClaudeJSON(configPath)
	if err != nil {
		t.Fatalf("readClaudeJSON: %v", err)
	}
	// The external write MUST survive — compare-retry detected the change and
	// re-applied the MCP mutation on the externally-modified state.
	if root["externalSetting"] != "from-cc" {
		t.Fatalf("external concurrent write was lost (compare-retry did not detect it): got %v", root["externalSetting"])
	}
	// The MCP entry MUST also be present.
	mcpServers := getMCPServers(root)
	if _, ok := mcpServers[zaiWebSearchPrimeKey]; !ok {
		t.Fatalf("MCP entry %s lost after guarded RMW", zaiWebSearchPrimeKey)
	}
}
