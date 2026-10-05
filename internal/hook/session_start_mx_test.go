package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// writeFile is a tiny helper to reduce repetition.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestSessionStartMXColdStartIntegration is the INVERTED form of the test that
// previously asserted SessionStart builds the MX sidecar index. The build it
// asserted could only ever land on the inline test path: in production the
// scan was dispatched from a goroutine in a short-lived CLI process that exits
// as soon as Handle returns, so the index was measured present in 2 of 153
// real worktrees. The capability now lives in the consumer that needs it —
// 'moai mx query' builds the index in-process (see internal/cli/mx_index.go) —
// and SessionStart must not build it at all.
//
// The test is inverted rather than deleted so the removal is asserted
// behaviourally: relocating and renaming the scan would defeat any grep for
// its identifiers, but not this.
func TestSessionStartMXColdStartIntegration(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "svc", "worker.go"), "package svc\n"+
		"// @MX:NOTE: [AUTO] integration marker\nfunc Bar() {}\n")

	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:  "test-session-mx",
		ProjectDir: dir,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	idxPath := filepath.Join(dir, ".moai", "state", mx.SidecarFileName)
	if _, statErr := os.Stat(idxPath); statErr == nil {
		t.Fatalf("SessionStart built an MX index at %s; the cold-start scan must be gone from this path", idxPath)
	} else if !os.IsNotExist(statErr) {
		t.Fatalf("unexpected stat error on %s: %v", idxPath, statErr)
	}
}
