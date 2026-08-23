package cli

// SPEC-CODEX-WIRING-001 M2 — the update-path wiring refresh (REQ-CW-009):
// `moai update` refreshes the Codex wiring ONLY in projects that already
// carry a wiring file; file existence is the user's standing opt-in. A
// failure warns and the update continues (best-effort, spec §F).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// TestRefreshCodexWiringBestEffort_AtPath covers the path-explicit seam the
// runUpdate call site uses for testability: a wired project is refreshed and
// the re-trust guidance is printed; an unwired project is a silent no-op.
func TestRefreshCodexWiringBestEffort_AtPath(t *testing.T) {
	// Wired project: refresh restores a removed handler and prints guidance.
	root := t.TempDir()
	var wireOut, wireWarn bytes.Buffer
	if _, err := codexwiring.Wire(root, &wireOut, &wireWarn); err != nil {
		t.Fatalf("Wire: %v", err)
	}
	hooksPath := filepath.Join(root, codexwiring.HooksRelPath)
	raw, _ := os.ReadFile(hooksPath)
	tampered := strings.Replace(string(raw), "moai hook stop --harness codex", "user-removed", 1)
	if err := os.WriteFile(hooksPath, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, warn bytes.Buffer
	refreshCodexWiringBestEffortAt(root, &out, &warn)
	if !strings.Contains(out.String(), "/hooks to re-trust") {
		t.Errorf("content-changing refresh missing the re-trust guidance (REQ-CW-009 → REQ-CW-008):\n%s", out.String())
	}
	restored, _ := os.ReadFile(hooksPath)
	if !strings.Contains(string(restored), "moai hook stop --harness codex") {
		t.Errorf("refresh did not restore the removed MoAI handler")
	}

	// Unwired project: silent no-op — update must not create wiring.
	bare := t.TempDir()
	var out2, warn2 bytes.Buffer
	refreshCodexWiringBestEffortAt(bare, &out2, &warn2)
	if _, err := os.Stat(filepath.Join(bare, ".codex")); err == nil {
		t.Errorf("update-path refresh created wiring in an opt-out project")
	}
	if out2.Len() != 0 {
		t.Errorf("update-path refresh printed guidance in an opt-out project: %q", out2.String())
	}
}

// TestRunUpdate_CallsCodexWiringRefresh is the reachability guard (source
// inspection, TestRunInit_CallsMCPProvisioning pattern): runUpdate must call
// the codex wiring refresh in its post-sync phase.
func TestRunUpdate_CallsCodexWiringRefresh(t *testing.T) {
	src, err := os.ReadFile("update.go")
	if err != nil {
		t.Fatalf("read update.go: %v", err)
	}
	if !strings.Contains(string(src), "refreshCodexWiringBestEffort(") {
		t.Error("runUpdate must call refreshCodexWiringBestEffort in its post-sync phase — without it a wired project's hooks.json goes stale (REQ-CW-009)")
	}
}
