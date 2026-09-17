package cli

// update_outcome_roots_test.go — SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-004
// (F13 summary half) and REQ-ICU-006 (F15 manual-recovery advisory): the
// update outcome summary names EVERY backup root the run created, and states
// that customizations outside .moai/config/sections are not merge-restored.
// The three-root structure itself is record-only (update_namespace_protect.go
// package doc — no consolidation).

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestRenderUpdateOutcome_NamesEveryBackupRoot pins REQ-ICU-004: with config,
// namespace, and archive-drift roots all created, the render carries each
// root's recoverable path.
func TestRenderUpdateOutcome_NamesEveryBackupRoot(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	detail := updateOutcomeDetail{
		NamespaceBackupPath: ".moai/backups/update-2026-09-14T00-00-00Z",
		ArchiveDriftRoots:   []string{".moai/archive/skills/v2.16-drift-20260914T000000Z"},
	}
	renderUpdateOutcome(&buf, 10, detail, ".moai-backups/20260914_020747", tui.LightTheme())
	out := stripSGR(buf.String())

	for _, want := range []string{
		"Backup: .moai-backups/20260914_020747",
		"Namespace backup: .moai/backups/update-2026-09-14T00-00-00Z",
		"Archive drift backup: .moai/archive/skills/v2.16-drift-20260914T000000Z",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("outcome must name every created backup root; missing %q in:\n%s", want, out)
		}
	}
}

// TestRenderUpdateOutcome_NoRootsNoNewRows pins the zero-root boundary
// (acceptance.md §E): a render with no namespace and no drift roots adds no
// new backup rows beyond the pre-existing config-backup note.
func TestRenderUpdateOutcome_NoRootsNoNewRows(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderUpdateOutcome(&buf, 24, updateOutcomeDetail{}, "", tui.LightTheme())
	out := stripSGR(buf.String())
	if strings.Contains(out, "Namespace backup:") || strings.Contains(out, "Archive drift backup:") {
		t.Errorf("zero-root render must not add backup rows, got:\n%s", out)
	}
	if strings.Contains(out, "not merge-restored") {
		t.Errorf("no config backup → no manual-recovery advisory, got:\n%s", out)
	}
}

// TestRenderUpdateOutcome_ConfigOnlyLegacyShape pins the single-root
// regression guard (AC-004): the pre-existing rows for a config-only run —
// pill, Backup line, Recover line — stay byte-identical; the only addition is
// the REQ-ICU-006 advisory row, and the new root rows stay absent.
func TestRenderUpdateOutcome_ConfigOnlyLegacyShape(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderUpdateOutcome(&buf, 24, updateOutcomeDetail{}, ".moai-backups/x", tui.LightTheme())
	out := stripSGR(buf.String())

	if !strings.Contains(out, "Backup: .moai-backups/x") ||
		!strings.Contains(out, "Recover: moai update --restore .moai-backups/x") {
		t.Errorf("legacy config-backup rows must survive unchanged, got:\n%s", out)
	}
	if strings.Contains(out, "Namespace backup:") || strings.Contains(out, "Archive drift backup:") {
		t.Errorf("single-root render must not add namespace/drift rows, got:\n%s", out)
	}
	if !strings.Contains(out, "not merge-restored") {
		t.Errorf("REQ-ICU-006 advisory must accompany a config backup, got:\n%s", out)
	}
	if !strings.Contains(out, "evaluator-profiles/") || !strings.Contains(out, "astgrep-rules/") {
		t.Errorf("advisory must name the non-section customization directories, got:\n%s", out)
	}
	if !strings.Contains(out, ".moai-backups/x") {
		t.Errorf("advisory context must keep the backup path visible, got:\n%s", out)
	}
}
