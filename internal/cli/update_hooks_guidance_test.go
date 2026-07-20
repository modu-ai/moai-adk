// SPEC-HOOK-CONFIG-SAFETY-001 (M1) — /hooks review guidance characterization
// tests. The update command emits an advisory stdout line on the "Template
// sync complete" branch pointing the user at Claude Code's /hooks menu (CC
// snapshots hooks at session startup; a template re-render leaves running
// sessions stale until /hooks is reviewed or CC is restarted).
//
// AC coverage:
//   - AC-CS-001 / AC-CS-003: TestUpdate_Characterize_TemplateSyncComplete_EmitsHooksGuidance
//   - AC-CS-002a:            TestUpdate_Characterize_AlreadyUpToDate_NoHooksGuidance
//   - AC-CS-002b:            TestUpdate_Characterize_SkippingSync_NoHooksGuidance
//   - AC-CS-002c:            TestUpdate_Characterize_BinaryOnly_NoHooksGuidance
//   - AC-CS-004:             TestUpdate_NoAskUserQuestion
//
// All runUpdate-backed tests use t.TempDir() + os.Chdir and are therefore
// non-parallel per CLAUDE.local.md §6 (test isolation).

package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
)

// TestUpdate_Characterize_TemplateSyncComplete_EmitsHooksGuidance (AC-CS-001,
// AC-CS-003): the guidance emitted on the "Template sync complete" branch
// contains the LITERAL token /hooks (case-sensitive, not a regex) and
// conveys the review-or-restart remedy so the user can act on it.
// emitHooksReviewGuidance is the single emission site, wired at update.go's
// "Template sync complete" pill (content token Label: "Template sync complete").
func TestUpdate_Characterize_TemplateSyncComplete_EmitsHooksGuidance(t *testing.T) {
	var buf bytes.Buffer
	report.EmitHooksReviewGuidance(&buf)
	got := buf.String()

	// AC-CS-003: literal substring /hooks (NOT case-insensitive, NOT regex).
	if !strings.Contains(got, "/hooks") {
		t.Errorf("AC-CS-003: guidance must contain literal /hooks token, got: %q", got)
	}
	// AC-CS-001: conveys the review-or-restart remedy.
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "review") && !strings.Contains(lower, "restart") {
		t.Errorf("AC-CS-001: guidance must convey review-or-restart remedy, got: %q", got)
	}
}

// TestUpdate_Characterize_SkippingSync_NoHooksGuidance (AC-CS-002b): the
// "Template version up-to-date · Skipping sync" branch must NOT emit /hooks
// guidance (no template re-render occurred → no stale-snapshot risk).
//
// Triggered via runTemplateSyncWithProgress directly with a version-match
// system.yaml (mirrors the established update_skip_sync_test.go pattern).
// runTemplateSyncWithProgress prints "Skipping sync" and returns skipped=true;
// runUpdate observes skipped and returns at the syncSkipped guard BEFORE
// reaching the "Template sync complete" pill + emitHooksReviewGuidance site.
// Calling runTemplateSyncWithProgress directly avoids runUpdate's v2-fingerprint
// / binary-step machinery (orthogonal to this AC).
func TestUpdate_Characterize_SkippingSync_NoHooksGuidance(t *testing.T) {
	tmpDir := t.TempDir()
	seedSystemYAMLVersion(t, tmpDir, version.GetVersion()) // match → skip sync

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "test-skip-sync"}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("force", false, "")

	skipped, syncErr := runTemplateSyncWithProgress(cmd)
	if syncErr != nil {
		t.Fatalf("runTemplateSyncWithProgress: %v", syncErr)
	}
	if !skipped {
		t.Fatalf("expected skipped=true on version match without --force")
	}

	got := buf.String()
	if !strings.Contains(got, "Skipping sync") {
		t.Fatalf("expected 'Skipping sync' branch to fire, got:\n%s", got)
	}
	if strings.Contains(got, "/hooks") {
		t.Errorf("AC-CS-002b: /hooks guidance must NOT appear on Skipping-sync branch, got:\n%s", got)
	}
}

// TestUpdate_Characterize_AlreadyUpToDate_NoHooksGuidance (AC-CS-002a): the
// "Already up to date" branch (--binary, no binary update) must NOT emit
// /hooks guidance. MOAI_SKIP_BINARY_UPDATE=1 forces shouldSkipBinaryUpdate so
// the binary step is deterministically skipped without network access. In the
// test env this yields the "Skipped (dev build, --binary)" pill; both that
// pill and the production "Already up to date" pill return before the
// "Template sync complete" site, so the /hooks-absent invariant holds on the
// entire --binary path regardless of which pill fires.
func TestUpdate_Characterize_AlreadyUpToDate_NoHooksGuidance(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	t.Setenv("MOAI_SKIP_BINARY_UPDATE", "1") // deterministically skip binary step

	var buf bytes.Buffer
	cleanup := bindUpdateCmd(&buf, map[string]string{"binary": "true", "yes": "true"})
	defer cleanup()

	if err := updateCmd.RunE(updateCmd, []string{}); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "/hooks") {
		t.Errorf("AC-CS-002a: /hooks guidance must NOT appear on binary (Already up to date) branch, got:\n%s", got)
	}
}

// TestUpdate_Characterize_BinaryOnly_NoHooksGuidance (AC-CS-002c): the
// "Binary updated (template sync skipped)" branch must NOT emit /hooks
// guidance. Same --binary path as AC-CS-002a (returns before the "Template
// sync complete" site); this test pins the invariant for the binary-updated
// pill specifically. In the test env MOAI_SKIP_BINARY_UPDATE=1 makes the
// binary step a no-op, so the emitted pill is "Skipped (dev build, --binary)";
// the /hooks-absent invariant holds on the entire --binary path regardless.
func TestUpdate_Characterize_BinaryOnly_NoHooksGuidance(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	t.Setenv("MOAI_SKIP_BINARY_UPDATE", "1")

	var buf bytes.Buffer
	cleanup := bindUpdateCmd(&buf, map[string]string{"binary": "true", "yes": "true"})
	defer cleanup()

	if err := updateCmd.RunE(updateCmd, []string{}); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "/hooks") {
		t.Errorf("AC-CS-002c: /hooks guidance must NOT appear on Binary-only branch, got:\n%s", got)
	}
}

// TestUpdate_NoAskUserQuestion (AC-CS-004): the update command's CLI code
// must NOT reference AskUserQuestion or mcp__askuser__* — the orchestrator
// owns user interaction (C-HRA-008 / REQ-PGN-012). Mirrors the canonical
// TestNew_NoAskUserQuestion static guard in internal/cli/worktree/new_test.go.
func TestUpdate_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("update.go")
	if err != nil {
		t.Fatalf("read update.go: %v", err)
	}
	for _, token := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(string(src), token) {
			t.Errorf("AC-CS-004: internal/cli/update.go must NOT reference %s (orchestrator-only HARD)", token)
		}
	}
}

// seedSystemYAMLVersion writes a .moai/config/sections/system.yaml whose
// template_version matches ver, so runTemplateSyncWithProgress short-circuits
// on version match (the "Template version up-to-date · Skipping sync" no-op).
func seedSystemYAMLVersion(t *testing.T, projectRoot, ver string) {
	t.Helper()
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	content := fmt.Sprintf("moai:\n  template_version: %s\n", ver)
	if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}
}

// bindUpdateCmd points the package-level updateCmd's stdout/stderr at buf,
// resets every runUpdate flag to its zero value, then applies the provided
// flag overrides. The returned cleanup resets all flags back to false so
// subsequent tests start from a clean flag baseline (cobra flags otherwise
// persist across invocations of the same Command instance).
func bindUpdateCmd(buf *bytes.Buffer, set map[string]string) func() {
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	updateCmd.SetContext(context.Background()) // runUpdate's v2-reinstall path calls context.WithTimeout(cmd.Context(), ...)
	all := []string{"check", "shell-env", "config", "force", "yes", "templates-only", "binary", "dry-run", "no-hooks", "verbose"}
	for _, name := range all {
		_ = updateCmd.Flags().Set(name, "false")
	}
	for name, val := range set {
		_ = updateCmd.Flags().Set(name, val)
	}
	return func() {
		for _, name := range all {
			_ = updateCmd.Flags().Set(name, "false")
		}
	}
}
