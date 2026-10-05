package rosterguard

import (
	"os"
	"path/filepath"
	"testing"
)

// repoRoot resolves the repository root from this package's directory
// (internal/harness/rosterguard).
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Join(wd, "..", "..", "..")
	// Assert the resolved root is the tree this guard means to measure. Without
	// this the whole package could run green against the wrong directory and
	// report nothing, which is the shape of a vacuous pass.
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("resolved repo root %s does not carry go.mod: %v", root, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "agents", "moai")); err != nil {
		t.Fatalf("resolved repo root %s does not carry .claude/agents/moai: %v", root, err)
	}
	return root
}
