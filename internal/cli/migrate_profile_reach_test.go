package cli

// migrate_profile_reach_test.go — D1 (U-01 (c), card t1619; operator decision
// d-20261010T060848Z-dc3b). A profile-mapped project keeps its shipped shared
// copies: the migration keeps and marks them, and every other shape behaves as
// before. The dry-run preview is not covered here; its kept line is a change to
// the preview surface, outside this bundle.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// profileReachSandbox confines HOME and the launch ledger to test-owned
// directories and clears CLAUDE_CONFIG_DIR, so neither the developer's shell
// (often a profile session) nor the real home can change the reach predicate.
// The ledger starts empty: the project is unmapped unless a test maps it.
func profileReachSandbox(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv(config.EnvClaudeConfigDir, "")
	base := t.TempDir()
	prev := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = prev })
	if err := os.MkdirAll(filepath.Join(base, "work"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// mapProjectToWork records root under the profile "work" in the sandbox ledger.
func mapProjectToWork(t *testing.T, root string) {
	t.Helper()
	if err := profile.RecordLastUsedProfileForProject(root, "work"); err != nil {
		t.Fatalf("record profile mapping: %v", err)
	}
}

// TestMigrationKeepsSharedCopyUnderProfileReach: a shipped copy (bytes equal to
// its template hash) in a profile-mapped project survives the migration even
// though its user counterpart is confirmed, and it is marked for the 3.2-2-5
// doctor to reclaim.
func TestMigrationKeepsSharedCopyUnderProfileReach(t *testing.T) {
	profileReachSandbox(t)
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	mapProjectToWork(t, root)

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	before, ok := mgr.GetEntry(migIdenticalSkill)
	if !ok {
		t.Fatalf("fixture did not track %s", migIdenticalSkill)
	}
	identical := readFixtureFile(t, root, migIdenticalSkill)

	var lines []string
	removed, err := migrateProjectCommonAssets(root, home, true, nil, func(format string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(format, args...))
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 0 {
		t.Errorf("profile-mapped migration removed %v; a shipped copy must stay project-side", removed)
	}
	if got := readFixtureFile(t, root, migIdenticalSkill); got != identical {
		t.Errorf("shipped copy changed under profile reach: %q", got)
	}
	if !slices.ContainsFunc(lines, func(l string) bool {
		return strings.Contains(l, "kept (profile-mapped project") && strings.Contains(l, migIdenticalSkill)
	}) {
		t.Errorf("no kept line for %s in the migration output: %q", migIdenticalSkill, lines)
	}

	reloaded := manifest.NewManager()
	if _, err := reloaded.Load(root); err != nil {
		t.Fatal(err)
	}
	after, ok := reloaded.GetEntry(migIdenticalSkill)
	if !ok {
		t.Fatal("the manifest entry for the kept copy was dropped")
	}
	if after.KeptBy != manifest.KeptByProfileReach {
		t.Errorf("KeptBy after reload = %q, want %q", after.KeptBy, manifest.KeptByProfileReach)
	}
	if after.TemplateHash != before.TemplateHash {
		t.Errorf("TemplateHash changed from %q to %q; a kept copy keeps its shipped hash", before.TemplateHash, after.TemplateHash)
	}
}

// TestMigrationReachLeavesUserModifiedCopy: a user-modified copy in a
// profile-mapped project is preserved byte-for-byte and carries no mark.
func TestMigrationReachLeavesUserModifiedCopy(t *testing.T) {
	profileReachSandbox(t)
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	mapProjectToWork(t, root)
	edited := readFixtureFile(t, root, migModifiedSkill)

	var lines []string
	if _, err := migrateProjectCommonAssets(root, home, true, nil, func(format string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(format, args...))
	}); err != nil {
		t.Fatal(err)
	}
	if got := readFixtureFile(t, root, migModifiedSkill); got != edited {
		t.Errorf("user-modified copy changed under profile reach: %q", got)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	if entry, ok := mgr.GetEntry(migModifiedSkill); ok && entry.KeptBy != "" {
		t.Errorf("user-modified copy carries KeptBy=%q; only shipped copies are marked", entry.KeptBy)
	}
	for _, l := range lines {
		if strings.Contains(l, "kept (profile-mapped project") && strings.Contains(l, migModifiedSkill) {
			t.Errorf("user-modified copy reported as kept: %s", l)
		}
	}
}

// TestFileEntryKeptBy: an entry with no KeptBy serializes no kept key, and a
// set value round-trips under the kept_by key.
func TestFileEntryKeptBy(t *testing.T) {
	plain, err := json.Marshal(manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: "h"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(plain)), "kept") {
		t.Errorf("an entry with no KeptBy serialized a kept key: %s", plain)
	}

	marked, err := json.Marshal(manifest.FileEntry{
		Provenance:   manifest.TemplateManaged,
		TemplateHash: "h",
		KeptBy:       manifest.KeptByProfileReach,
	})
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(marked, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["kept_by"] != manifest.KeptByProfileReach {
		t.Errorf("kept_by = %v in %s, want %q", raw["kept_by"], marked, manifest.KeptByProfileReach)
	}
	var back manifest.FileEntry
	if err := json.Unmarshal(marked, &back); err != nil {
		t.Fatal(err)
	}
	if back.KeptBy != manifest.KeptByProfileReach || back.TemplateHash != "h" {
		t.Errorf("round trip lost fields: %+v", back)
	}
}
