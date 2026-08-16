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
func TestRenderUpdateOutcome_ManagedBreakdown(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	detail := updateOutcomeDetail{
		ManagedRedeployed: 143,
		RemovedManaged:    155,
		RemovedLocalOnly:  12,
	}
	renderUpdateOutcome(&buf, 32, detail, ".moai-backups/x", tui.LightTheme())
	got := buf.String()

	if !strings.Contains(got, "175 files") {
		t.Errorf("pill total must include managed re-deployments (32+143=175), got:\n%s", got)
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

// --- Defect 3: --dry-run must preview the managed-cleanup deletion list ---

// TestPreviewManagedCleanup seeds a managed tree with one file the templates
// re-deploy and one local-only file, then asserts the preview names the
// local-only loss, counts the re-deployment, and mutates nothing.
func TestPreviewManagedCleanup(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Re-deployed: exists in the embedded template tree.
	restored := filepath.Join(root, ".claude", "rules", "moai", "core", "zone-registry.md")
	if err := os.MkdirAll(filepath.Dir(restored), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(restored, []byte("stale copy"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Local-only: a legacy skill gone from the templates.
	localOnly := filepath.Join(root, ".claude", "skills", legacySkillIDs[0], "SKILL.md")
	makeSkillDir(t, root, legacySkillIDs[0], "# local-only customization")

	// .moai/config presence (removed wholesale, restored from backup).
	cfg := filepath.Join(root, ".moai", "config", "sections", "user.yaml")
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, []byte("user: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := previewManagedCleanup(root, &out); err != nil {
		t.Fatalf("previewManagedCleanup: %v", err)
	}
	got := out.String()

	if !strings.Contains(got, "not restored") {
		t.Errorf("preview must classify the local-only file as not restored, got:\n%s", got)
	}
	if !strings.Contains(got, "skills/"+legacySkillIDs[0]) {
		t.Errorf("preview must name the local-only skill path, got:\n%s", got)
	}
	if !strings.Contains(got, "re-deployed") {
		t.Errorf("preview must count template re-deployments, got:\n%s", got)
	}
	if !strings.Contains(got, ".moai/config") {
		t.Errorf("preview must mention the .moai/config wholesale removal, got:\n%s", got)
	}

	// Read-only contract: nothing deleted, nothing created.
	for _, path := range []string{restored, localOnly, cfg} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("preview must not remove %s: %v", path, err)
		}
	}
}

// TestPreviewManagedCleanup_EmptyProject asserts a project without managed
// paths prints nothing (no noise on non-moai or fresh directories).
func TestPreviewManagedCleanup_EmptyProject(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := previewManagedCleanup(t.TempDir(), &out); err != nil {
		t.Fatalf("previewManagedCleanup on empty root: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("empty project must produce no preview output, got:\n%s", out.String())
	}
}
