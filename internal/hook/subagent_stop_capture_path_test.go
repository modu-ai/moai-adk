// Reproduction test for the `internal/hook/.moai/` regression artifact bug.
//
// Root cause: subagent_stop.go:dispatchCapture() falls back to os.Getwd()
// when input.CWD is empty. When tests run from internal/hook/ as cwd, the
// captured observations.yaml lands at internal/hook/.moai/harness/, polluting
// the source tree. Fix: use resolveProjectRoot(input) which prefers
// CLAUDE_PROJECT_DIR, falls back to input.CWD, and guards on .moai/ existence.

package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestDispatchCapture_NoCwdLeak verifies that calling subagent stop handler
// with an empty CWD does NOT create a .moai/ subtree relative to the test's
// working directory. The capture must skip when the resolved project root
// is not a valid MoAI project (no pre-existing .moai/).
func TestDispatchCapture_NoCwdLeak(t *testing.T) {
	// Force a clean state: ensure no stale .moai/ exists relative to the
	// package's cwd (which is internal/hook/ under `go test`).
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	leakPath := filepath.Join(cwd, ".moai")
	_ = os.RemoveAll(leakPath)
	t.Cleanup(func() { _ = os.RemoveAll(leakPath) })

	// Clear CLAUDE_PROJECT_DIR so the resolver cannot escape to it.
	t.Setenv("CLAUDE_PROJECT_DIR", "")

	h := &subagentStopHandler{}
	input := &HookInput{
		SessionID:    "sess-leak-guard",
		AgentName:    "expert-backend",
		TeammateName: "leak-guard-mate",
		// CWD intentionally empty to trigger the fallback path.
	}

	// dispatchCapture is best-effort (errors swallowed) but the path-resolution
	// guard MUST prevent any filesystem write under cwd.
	h.dispatchCapture(input)

	if _, err := os.Stat(leakPath); err == nil {
		t.Fatalf("regression: dispatchCapture created %s relative to package cwd", leakPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("unexpected stat error on %s: %v", leakPath, err)
	}
}

// TestDispatchCapture_HonorsExplicitCwd verifies that when input.CWD points
// at a valid MoAI project root (has .moai/ already), the capture proceeds
// and writes to the correct location.
func TestDispatchCapture_HonorsExplicitCwd(t *testing.T) {
	dir := t.TempDir()
	// Pre-create .moai/ to mark this as a valid MoAI project root.
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "harness"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", "")

	// Use Handle to exercise the real dispatch path end-to-end.
	h := NewSubagentStopHandler()
	input := &HookInput{
		CWD:          dir,
		SessionID:    "sess-explicit",
		AgentName:    "expert-backend",
		TeammateName: "explicit-mate",
	}
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	obsPath := filepath.Join(dir, ".moai", "harness", "observations.yaml")
	if _, err := os.Stat(obsPath); err != nil {
		t.Fatalf("expected observations.yaml at %s, stat err=%v", obsPath, err)
	}
}
