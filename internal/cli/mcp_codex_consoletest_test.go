package cli

// SPEC-MCP-CONSOLE-001 M3 — AC-C-009 test.
//
// Verifies that the codex opt-in toggles (review_gate.enabled,
// task.allow_write), when written through the schema seam
// (settings.ApplySchemaEdits), are read back by the existing fail-closed
// readers (readCodexReviewGateEnabled / readCodexTaskAllowWrite). One source of
// truth — no parallel key.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestAC_C_009_TogglesWriteTheSeamThatGatesRead writes the two codex opt-in
// toggles through the schema seam and reads them back through the fail-closed
// readers the gates consume — proving the console write and the gate read share
// one workflow.yaml path (no parallel key).
func TestAC_C_009_TogglesWriteTheSeamThatGatesRead(t *testing.T) {
	projectDir := t.TempDir()
	// Create the config directory and an empty workflow.yaml so yamlpatch can
	// upsert the nested codex keys (the seam requires the file to exist).
	sectionsDir := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "workflow.yaml"), []byte("workflow: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write both toggles ON through the schema seam.
	edits := map[string]string{
		"workflow.codex.review_gate.enabled": "true",
		"workflow.codex.task.allow_write":    "true",
	}
	if err := settings.ApplySchemaEdits(projectDir, edits); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}

	// Read back through the fail-closed readers — the SAME functions the
	// review-gate and task gates call at runtime.
	if got := readCodexReviewGateEnabled(projectDir); !got {
		t.Error("AC-C-009: readCodexReviewGateEnabled = false after writing true through the schema seam (one source of truth violated)")
	}
	if got := readCodexTaskAllowWrite(projectDir); !got {
		t.Error("AC-C-009: readCodexTaskAllowWrite = false after writing true through the schema seam (one source of truth violated)")
	}
}

// TestAC_C_009_ToggleOffRoundTrips verifies the OFF state also round-trips
// correctly through the same seam → reader path.
func TestAC_C_009_ToggleOffRoundTrips(t *testing.T) {
	projectDir := t.TempDir()
	sectionsDir := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "workflow.yaml"), []byte("workflow: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Seed with ON, then write OFF.
	if err := settings.ApplySchemaEdits(projectDir, map[string]string{
		"workflow.codex.review_gate.enabled": "true",
	}); err != nil {
		t.Fatal(err)
	}
	if err := settings.ApplySchemaEdits(projectDir, map[string]string{
		"workflow.codex.review_gate.enabled": "false",
	}); err != nil {
		t.Fatal(err)
	}
	if got := readCodexReviewGateEnabled(projectDir); got {
		t.Error("AC-C-009: readCodexReviewGateEnabled = true after writing false through the seam")
	}
}
