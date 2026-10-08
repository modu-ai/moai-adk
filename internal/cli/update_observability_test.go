// t40 — observability tests for `moai update` reporting.
//
// Three defects share one root: a quiet failure the user has no way to notice.
//  1. archiveLegacySkills runs AFTER the managed-path cleanup wipes
//     .claude/skills/moai*, so the real run always archives 0 while
//     --dry-run (evaluated before the wipe) announces N archivals.
//  2. "Updated N files" counts only non-managed paths (AnalyzeFiles skips
//     IsMoaiManaged), so the summary undercounts the files the run writes and
//     never mentions the files it removes.
//  3. --dry-run previews no deletion list for CleanMoaiManagedPaths, so
//     local-only files under managed roots vanish without any prior notice.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// --- Defect 1: dry-run must predict removal, not a never-happening archive ---

// TestDryRunArchive_PredictsRemovalNotArchive asserts the dry-run archive plan
// states what the real run actually does: managed cleanup removes the sources
// before the archive step, so nothing is archived.
func TestDryRunArchive_PredictsRemovalNotArchive(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	makeSkillDir(t, root, legacySkillIDs[0], "# legacy")

	var out bytes.Buffer
	if err := dryRunArchiveLegacySkills(root, &out); err != nil {
		t.Fatalf("dryRunArchiveLegacySkills: %v", err)
	}
	got := out.String()

	if !strings.Contains(got, "will NOT be archived") {
		t.Errorf("dry-run must state the skills will NOT be archived, got:\n%s", got)
	}
	if !strings.Contains(got, legacySkillIDs[0]) {
		t.Errorf("output must still name the present skill %s, got:\n%s", legacySkillIDs[0], got)
	}
	if !strings.Contains(got, "total:") || !strings.Contains(got, "[dry-run]") {
		t.Errorf("output must keep the [dry-run] total: summary, got:\n%s", got)
	}
}

// TestDryRunArchive_NoSkills asserts the empty-project wording stays a plain
// "nothing present" plan.
func TestDryRunArchive_NoSkills_HonestZero(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	var out bytes.Buffer
	if err := dryRunArchiveLegacySkills(root, &out); err != nil {
		t.Fatalf("dryRunArchiveLegacySkills on empty project: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "total:") {
		t.Errorf("output missing summary line, got:\n%s", got)
	}
	if !strings.Contains(got, "0 will be archived") {
		t.Errorf("summary must state 0 will be archived, got:\n%s", got)
	}
}

// --- Defect 1: the real run must report the archive shortfall loudly ---

func TestPresentLegacySkillIDs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	makeSkillDir(t, root, legacySkillIDs[0], "# a")
	makeSkillDir(t, root, legacySkillIDs[1], "# b")

	got := presentLegacySkillIDs(root)
	if len(got) != 2 || got[0] != legacySkillIDs[0] || got[1] != legacySkillIDs[1] {
		t.Errorf("presentLegacySkillIDs = %v, want [%s %s]", got, legacySkillIDs[0], legacySkillIDs[1])
	}

	empty := presentLegacySkillIDs(t.TempDir())
	if len(empty) != 0 {
		t.Errorf("presentLegacySkillIDs on empty root = %v, want none", empty)
	}
}

// TestReportArchiveShortfall verifies the loud warning fires exactly when
// skills present before the sync were removed without being archived.
func TestReportArchiveShortfall(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_ = root

	// Shortfall: 3 present before sync, 0 archived → warn naming both numbers.
	var out bytes.Buffer
	reportArchiveShortfall([]string{"a", "b", "c"}, 0, &out)
	if !strings.Contains(out.String(), "0 of 3") {
		t.Errorf("shortfall warning must say 0 of 3, got:\n%s", out.String())
	}
	if !strings.HasPrefix(out.String(), "!") {
		t.Errorf("shortfall must render with the warn marker (!), got:\n%s", out.String())
	}
	// The archive now runs before the managed cleanup, so the only cause left
	// is an entry the archive could not copy; the warning must say so rather
	// than blame the cleanup order.
	if !strings.Contains(out.String(), "could not be archived") {
		t.Errorf("shortfall warning must name the archive failure as the cause, got:\n%s", out.String())
	}

	// No shortfall: everything archived → silent.
	out.Reset()
	reportArchiveShortfall([]string{"a", "b"}, 2, &out)
	if out.Len() != 0 {
		t.Errorf("no shortfall must print nothing, got:\n%s", out.String())
	}

	// Nothing present before sync → silent (nothing was lost).
	out.Reset()
	reportArchiveShortfall(nil, 0, &out)
	if out.Len() != 0 {
		t.Errorf("empty pre-sync inventory must print nothing, got:\n%s", out.String())
	}
}

// --- Defect 2: the outcome summary must count managed re-deployments and removals ---

// TestRenderUpdateOutcome_ManagedBreakdown asserts the pill total includes
// managed re-deployments and the detail note carries the removal accounting.
// SPEC-UPDATE-MIGRATION-001 (card t1547, REQ-UPM-032): the caller's count now
// ALREADY includes the managed files (the plan.AnalyzeFiles exclusion is
// gone), so the renderer no longer adds detail.ManagedRedeployed into the
// pill — that would double-count — and the breakdown states the inclusion.
func TestRenderUpdateOutcome_ManagedBreakdown(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	detail := updateOutcomeDetail{
		ManagedRedeployed: 143,
		RemovedManaged:    155,
		RemovedLocalOnly:  12,
	}
	// 175 = 32 non-managed + 143 managed, composed by the caller from the
	// now-unfiltered analysis.
	renderUpdateOutcome(&buf, 175, detail, ".moai-backups/x", tui.LightTheme())
	got := buf.String()

	if !strings.Contains(got, "175 files") {
		t.Errorf("pill total must be the caller's inclusive count, got:\n%s", got)
	}
	if strings.Contains(got, "350") || strings.Contains(got, "318") {
		t.Errorf("pill must not double-count managed re-deployments, got:\n%s", got)
	}
	for _, want := range []string{"143", "155", "12"} {
		if !strings.Contains(got, want) {
			t.Errorf("breakdown note must carry %s, got:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "not restored") {
		t.Errorf("breakdown must name the not-restored local-only count, got:\n%s", got)
	}
}

// TestRenderUpdateOutcome_ZeroDetailUnchanged asserts the zero-detail render
// keeps the legacy byte shape (no breakdown noise on trivial runs).
func TestRenderUpdateOutcome_ZeroDetailUnchanged(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderUpdateOutcome(&buf, 24, updateOutcomeDetail{}, ".moai-backups/x", tui.LightTheme())
	got := buf.String()
	if !strings.Contains(got, "24 files") {
		t.Errorf("expected legacy 24 files pill, got:\n%s", got)
	}
	if strings.Contains(got, "re-deployed") || strings.Contains(got, "removed") {
		t.Errorf("zero detail must not emit a breakdown note, got:\n%s", got)
	}
}

// TestRenderUpdateOutcome_ArchivedRemovalsNotRedeployed — gate round 21
// (card t1547): stale-only removals whose copies reached the archive must
// NOT be reported as "all re-deployed" — a recovery copy is not a
// redeployment in place. The breakdown names the archived-recovery
// disposition instead.
func TestRenderUpdateOutcome_ArchivedRemovalsNotRedeployed(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	detail := updateOutcomeDetail{
		RemovedManaged:      2,
		ArchivedForRecovery: 2,
	}
	renderUpdateOutcome(&buf, 175, detail, "", tui.LightTheme())
	got := buf.String()

	if !strings.Contains(got, "archived for recovery") {
		t.Errorf("breakdown must name the archived-recovery disposition, got:\n%s", got)
	}
	if !strings.Contains(got, "not redeployed") {
		t.Errorf("breakdown must say the files are NOT redeployed, got:\n%s", got)
	}
	if strings.Contains(got, "all re-deployed") {
		t.Errorf("breakdown must not claim all-re-deployed over archived removals, got:\n%s", got)
	}
}

// --- Defect 3 → SPEC-UPDATE-MIGRATION-001: --dry-run previews the
// reconciliation plan (the cleanup-deletion preview's subject no longer
// exists — the default path preserves instead of deleting) ---

// TestPreviewReconciliation seeds a managed tree with one template-carried
// file and one local-only file, then asserts the preview reports the
// preservation (never a deletion), counts the refresh set, and mutates
// nothing.
func TestPreviewReconciliation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Template-carried: exists in the embedded template tree → refreshed.
	restored := filepath.Join(root, ".claude", "rules", "moai", "core", "zone-registry.md")
	if err := os.MkdirAll(filepath.Dir(restored), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(restored, []byte("stale copy"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Local-only: a legacy skill gone from the templates → PRESERVED.
	localOnly := filepath.Join(root, ".claude", "skills", legacySkillIDs[0], "SKILL.md")
	makeSkillDir(t, root, legacySkillIDs[0], "# local-only customization")

	// .moai/config section: classified, never wiped wholesale.
	cfg := filepath.Join(root, ".moai", "config", "sections", "user.yaml")
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, []byte("user: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := previewReconciliation(root, template.DeployModeLocal, &out); err != nil {
		t.Fatalf("previewReconciliation: %v", err)
	}
	got := out.String()

	if strings.Contains(got, "not restored") || strings.Contains(got, "removed entirely") {
		t.Errorf("preview must not announce the old destructive semantics, got:\n%s", got)
	}
	if !strings.Contains(got, "preserved") {
		t.Errorf("preview must report the preserved disposition, got:\n%s", got)
	}
	if !strings.Contains(got, "skills/"+legacySkillIDs[0]) {
		t.Errorf("preview must name the preserved local-only skill path, got:\n%s", got)
	}
	if !strings.Contains(got, "refreshed") {
		t.Errorf("preview must count the refresh set, got:\n%s", got)
	}

	// Read-only contract: nothing deleted, nothing created.
	for _, path := range []string{restored, localOnly, cfg} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("preview must not remove %s: %v", path, err)
		}
	}
}

// TestPreviewReconciliation_EmptyProject asserts a project without managed
// paths prints nothing (no noise on non-moai or fresh directories).
func TestPreviewReconciliation_EmptyProject(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := previewReconciliation(t.TempDir(), template.DeployModeLocal, &out); err != nil {
		t.Fatalf("previewReconciliation on empty root: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("empty project must produce no preview output, got:\n%s", out.String())
	}
}

// TestPreviewReconciliation_ReadOnlyOnCorruptManifest — gate round 19: a
// CORRUPT manifest must not mutate the project on a dry run. Manager.Load's
// production recovery renames the original to .corrupt; the preview reads
// through the read-only loader instead — the original stays byte-identical,
// no .corrupt appears, and the preview still renders (conservative
// no-record classification).
func TestPreviewReconciliation_ReadOnlyOnCorruptManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	const corrupt = "{ this is not valid json"
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, ".moai", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(corrupt), 0o644); err != nil {
		t.Fatal(err)
	}
	localRule := filepath.Join(root, ".claude", "rules", "moai", "local.md")
	if err := os.MkdirAll(filepath.Dir(localRule), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localRule, []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := previewReconciliation(root, template.DeployModeLocal, &out); err != nil {
		t.Fatalf("previewReconciliation: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil || string(data) != corrupt {
		t.Errorf("corrupt manifest was moved or altered: %q (err=%v)", data, err)
	}
	if _, err := os.Stat(manifestPath + ".corrupt"); !os.IsNotExist(err) {
		t.Errorf(".corrupt file created by a dry run: %v", err)
	}
	if !strings.Contains(out.String(), "local.md") {
		t.Errorf("preview must still classify (conservative no-record route), got:\n%s", out.String())
	}
}

// TestPreviewReconciliation_ModeScoped is card t1438 review finding 4,
// carried onto the reconciliation preview: the preview must share the run's
// exact target computation. On a plugin-mode record the dropped roots are
// out of the run's scope, so a preserved moai-custom file under
// .claude/skills must NOT be listed; the local mode lists it (preserved).
func TestPreviewReconciliation_ModeScoped(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/llm.yaml",
		"llm:\n  harness: claude\n  deployment_mode: plugin\n")
	skill := filepath.Join(root, ".claude", "skills", "moai-custom", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("# the user's own skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// SPEC-USER-ASSET-INSTALL-001 (t1509 M7) retired the plugin payload:
	// DeployMode is Local-only now, so the former plugin-scope subtest is
	// gone with its carrier. The local-mode preview still must list the
	// preserved unknown skill.
	var localOut bytes.Buffer
	if err := previewReconciliation(root, template.DeployModeLocal, &localOut); err != nil {
		t.Fatalf("previewReconciliation (local): %v", err)
	}
	if !strings.Contains(localOut.String(), "moai-custom") {
		t.Errorf("local-mode preview lost the preserved skill listing:\n%s", localOut.String())
	}
}
