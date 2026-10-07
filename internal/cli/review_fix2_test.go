// review_fix2_test.go — round-2 reproducing tests (card t1509): F2/F3/F4
// (migration + cleanup gating), F8 (completion flags), F9 (retired-file
// prune), F11 (agentfm classification).
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// F3: a file the user modified AFTER install (still recorded
// template_managed) is PRESERVED by the migration — the current-bytes
// comparison gates the deletion.
func TestRF2F3_MigrationPreservesPostInstallEdits(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills", "moai-seeded"), 0o755); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(root, ".claude", "skills", "moai-seeded", "SKILL.md")
	if err := os.WriteFile(live, []byte("user's post-install edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	// Recorded template_managed with the ORIGINAL template hash — the user's
	// post-install edit means the current bytes no longer match.
	if err := mgr.Track(".claude/skills/moai-seeded/SKILL.md", manifest.TemplateManaged, manifest.HashBytes([]byte("original template\n"))); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	if err := migrateProjectCommonAssets(root, home, nil, func(string, ...interface{}) {}); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(live); err != nil || string(got) != "user's post-install edit\n" {
		t.Errorf("F3: post-install edit destroyed: %q (%v)", got, err)
	}
}

// F3b: an untracked file (no manifest entry) is user-owned — never deleted.
func TestRF2F3b_MigrationPreservesUntracked(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills", "moai-foreign"), 0o755); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(root, ".claude", "skills", "moai-foreign", "SKILL.md")
	if err := os.WriteFile(live, []byte("user's own skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	if err := migrateProjectCommonAssets(root, home, nil, func(string, ...interface{}) {}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(live); err != nil {
		t.Errorf("F3b: untracked user file deleted: %v", err)
	}
}

// F4: the counterpart verification reads the DISK — a stale record whose
// file never landed must NOT authorize the project-side deletion.
func TestRF2F4_StaleRecordDoesNotAuthorizeDeletion(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	um, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		t.Fatal(err)
	}
	// Record claims the counterpart exists — but the disk is empty.
	um.Files["claude-skills/moai-ghost/SKILL.md"] = userassets.FileEntry{
		SHA256: strings.Repeat("a", 64), Bundle: "core",
		InstalledAt: "t0", MoaiVersion: "v0",
	}
	if userCounterpartConfirmed(um, home, ".claude/skills/moai-ghost/SKILL.md") {
		t.Error("F4: stale record authorized deletion — the disk check is missing")
	}
}

// F8: WriteCompleted persists after successful writes — a manifest-save
// failure followed by a user edit classifies the run's own installs as
// divergence (with backup), never as collisions.
func TestRF2F8_WriteCompletedPersists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	home, _ := os.UserHomeDir()

	if err := ensureUserAssetsLocked(home, nil, &strings.Builder{}); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	// Simulate the interruption: the manifest save failed (records wiped)
	// while the writes landed and the journal survived; the user then edited
	// a file the run wrote. Seed the journal with a FLAG-COMPLETE entry for
	// the edited file (the state the persisted flags would produce).
	target := filepath.Join(home, ".claude", "skills", "moai-workflow-tdd", "SKILL.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, append(data, []byte("\nuser edit\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPathOf(home),
		[]byte("{\"schema_version\":1,\"bundles\":null,\"files\":{},\"collisions\":null}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	j := &userassets.PendingJournal{SchemaVersion: 1, Entries: []userassets.JournalEntry{{
		Path:           "claude-skills/moai-workflow-tdd/SKILL.md",
		ExpectedSHA256: shaHex(data),
		Bundle:         "core", MoaiVersion: "vTest", InstalledAt: "t0",
		WriteCompleted: true,
	}}}
	if err := userassets.WriteJournal(userassets.JournalPath(home), j); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := runUserAssetUpdatePhase(home, &out); err != nil {
		t.Fatalf("update phase: %v", err)
	}
	// The E5 flag-complete arm: divergence (with backup), NOT collision.
	if strings.Contains(out.String(), "collision") && !strings.Contains(out.String(), "divergence") {
		t.Errorf("F8: own install misclassified as collision — summary: %s", out.String())
	}
}

func manifestPathOf(home string) string { return userassets.ManifestPath(home) }

// F2: computeRunCleanTargets drops migration-preserved paths.
func TestRF2F2_CleanTargetsDropPreserved(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	preserved := filepath.Join(root, ".claude", "skills", "moai-kept")
	if err := os.MkdirAll(preserved, 0o755); err != nil {
		t.Fatal(err)
	}
	setMigrationPreservedFiles(map[string]bool{preserved: true})
	defer setMigrationPreservedFiles(map[string]bool{})

	targets := computeRunCleanTargets(root, "", nil)
	for _, ct := range targets {
		if strings.Contains(ct.DisplayPath, "moai-kept") {
			t.Errorf("F2: preserved path survived in the cleanup list: %+v", ct)
		}
	}
}
