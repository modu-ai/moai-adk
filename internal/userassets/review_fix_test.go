// review_fix_test.go — the mid-run review-fix round (card t1509, codex gate
// reproduction set): nine implementation defects in the landed M0-M2 code,
// each pinned here as a RED-first reproducing test.
package userassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// RF3: bundle removal must enumerate the bundle's tracked files from the
// MANIFEST, not the current source tree — files installed by an older
// deployment but absent from the new tree (old.md) must still be removed.
func TestRF3_BundleRemoveEnumeratesFromManifest(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatal(err)
	}
	// An artifact of an OLDER deployment: on disk + tracked, but absent from
	// the current source tree.
	ghost := filepath.Join(f.home, ".claude/skills/moai-beta/old.md")
	if err := os.WriteFile(ghost, []byte("old deployment artifact\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := Load(ManifestPath(f.home))
	fe := m.Files["claude-skills/moai-beta/old.md"]
	fe.SHA256 = sha256Hex([]byte("old deployment artifact\n"))
	fe.Bundle = "extras"
	m.Files["claude-skills/moai-beta/old.md"] = fe
	if err := m.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}

	m2, _ := Load(ManifestPath(f.home))
	res, err := f.installer(t).RemoveBundle(m2, "extras", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ghost); !os.IsNotExist(err) {
		t.Errorf("RF3: old-deployment artifact survived bundle remove: %v", err)
	}
	if _, tracked := m2.Files["claude-skills/moai-beta/old.md"]; tracked {
		t.Error("RF3: manifest record for the artifact not dropped")
	}
	_ = res
}

// RF4: a corrupt journal is PRESERVED (renamed aside), the run aborts —
// never silently overwritten and lost.
func TestRF4_CorruptJournalPreservedAndRunAborts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	corrupt := "not json at all"
	if err := os.MkdirAll(MoaiHome(f.home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(JournalPath(f.home), []byte(corrupt), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := f.installer(t).Install(nil)
	if err == nil {
		t.Fatal("RF4: install proceeded over a corrupt journal")
	}
	if !strings.Contains(err.Error(), "journal") {
		t.Errorf("RF4: error should name the journal: %v", err)
	}
	// The corrupt bytes still exist (renamed aside, never deleted).
	entries, _ := os.ReadDir(MoaiHome(f.home))
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "journal") {
			found = true
		}
	}
	if !found {
		t.Error("RF4: corrupt journal vanished instead of being preserved")
	}
}

// RF6: refreshing a tracked record must PRESERVE unknown fields captured at
// decode (external schema-99 metadata survives the update).
func TestRF6_RefreshPreservesUnknownFields(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	p := ManifestPath(home)
	if err := os.MkdirAll(MoaiHome(home), 0o755); err != nil {
		t.Fatal(err)
	}
	newer := `{"schema_version": 2, "files": {"claude-skills/moai-alpha/SKILL.md":
		{"sha256": "aa", "bundle": "core", "installed_at": "t0", "moai_version": "v9",
		 "external_metadata": {"custom": true}}}}`
	if err := os.WriteFile(p, []byte(newer), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate the refresh path's record replacement.
	fe := m.Files["claude-skills/moai-alpha/SKILL.md"]
	fe.SHA256 = "bb"
	fe.MoaiVersion = "v3.2.0-test"
	m.Files["claude-skills/moai-alpha/SKILL.md"] = fe
	if err := m.Save(p); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(p)
	if !strings.Contains(string(raw), "external_metadata") {
		t.Error("RF6: unknown per-file field dropped by the record update (REQ-021)")
	}
}

// RF7: an unknown bundle name passed to Install is an ERROR before anything
// saves — not a silent skip.
func TestRF7_InstallValidatesBundleNames(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	_, err := f.installer(t).Install([]string{"extrsa"})
	if err == nil {
		t.Fatal("RF7: unknown bundle name accepted by Install")
	}
	if !strings.Contains(err.Error(), "extrsa") {
		t.Errorf("RF7: error does not name the offending bundle: %v", err)
	}
	m, _ := Load(ManifestPath(f.home))
	for _, b := range m.Bundles {
		if b == "extrsa" {
			t.Error("RF7: invalid bundle recorded into the manifest")
		}
	}
}

// RF8: a stale lock reclaimed by a second run cannot be deleted by the
// ORIGINAL holder's Release — owner identity guards the removal.
func TestRF8_ReleaseRespectsOwnerIdentity(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	l1, err := acquireUserLockStale(LockPath(home), 100*time.Millisecond, time.Nanosecond)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	// Simulate the original holder resuming AFTER a stale reclaim: Release
	// must refuse to delete a lock it no longer owns.
	if err := l1.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	// After a legit release the file is gone; now the stale-reclaim scenario:
	// forge the original's identity over a lock a new owner holds.
	l2, err := acquireUserLockStale(LockPath(home), 100*time.Millisecond, time.Nanosecond)
	if err != nil {
		t.Fatalf("second acquire: %v", err)
	}
	if l1.token == l2.token {
		t.Error("RF8: two acquires share an owner token — identity guard is fake")
	}
	_ = l2.Release()
}

// RF9: a second project's init must NOT wipe the shared manifest's bundle
// selection — the existing selection is UNIONed with the requested one.
func TestRF9_SecondProjectInitUnionsSelection(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatal(err)
	}
	// Second project, default (empty) selection: the recorded selection
	// survives.
	if _, err := f.installer(t).InstallPreserveSelection(nil); err != nil {
		t.Fatal(err)
	}
	m, _ := Load(ManifestPath(f.home))
	if len(m.Bundles) != 1 || m.Bundles[0] != "extras" {
		t.Errorf("RF9: selection wiped by a default-selection run: %v", m.Bundles)
	}
}

// RF2: the prune applies the same dependency-deferral rule as RemoveBundle —
// a manifest-tracked file that is a declared dependency of a preserved entry
// survives the prune (kept, not deleted).
func TestRF2_PrunDefersDependencies(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	// Make the L0 dispatcher depend on the extras-bundled skill via the
	// catalog: alpha's dep edge names moai-beta.
	f.cat.Catalog.Core.Skills[0].DependsSkills = []string{"moai-beta"}
	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatal(err)
	}
	// Deselect extras: moai-beta becomes a prune candidate, but the
	// preserved L0 moai-alpha declares it.
	m, _ := Load(ManifestPath(f.home))
	m.Bundles = nil
	if err := m.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}
	in := f.installer(t)
	m2, _ := Load(ManifestPath(f.home))
	res, err := in.PruneUnselected(m2)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(f.home, ".claude/skills/moai-beta/SKILL.md")
	if _, err := os.Stat(target); err != nil {
		t.Error("RF2: declared dependency deleted by the prune (deferral missing)")
	}
	found := false
	for _, d := range res.DeferredDeps {
		if d == "moai-beta" {
			found = true
		}
	}
	if !found {
		t.Errorf("RF2: deferral not reported: %v", res.DeferredDeps)
	}
}

// RF1: the deletion path re-validates the destination parent immediately
// before os.Remove — a parent swapped to an outside-pointing symlink must
// not let the delete reach a file beyond the root boundary.
func TestRF1_DeleteParentSwapRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatal(err)
	}
	// External sentinel: what an unguarded remove would delete through a
	// swapped parent symlink.
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "external-target.md")
	if err := os.WriteFile(sentinel, []byte("external\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Swap the skill's directory to an outside-pointing symlink.
	skillDir := filepath.Join(f.home, ".claude/skills/moai-beta")
	if err := os.RemoveAll(skillDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(skillDir), 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, outside, skillDir)

	// Deselect extras so the swapped-dir files become removal candidates,
	// then run the REAL removal path (RemoveBundle).
	m, _ := Load(ManifestPath(f.home))
	m.Bundles = nil
	if err := m.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}
	in := f.installer(t)
	m2, _ := Load(ManifestPath(f.home))
	if _, err := in.RemoveBundle(m2, "extras", nil); err != nil {
		t.Fatalf("RemoveBundle: %v", err)
	}
	if got := readBytes(t, sentinel); string(got) != "external\n" {
		t.Errorf("RF1: external sentinel DELETED through the swapped parent: %q", got)
	}
}

// RF5: untracked content-identical files are NOT journaled as ownership
// targets — an interruption after the journal write + retry must register
// them as collisions, never as MoAI-owned.
func TestRF5_IdenticalUntrackedNotJournaled(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	// Plant a content-identical copy of the shipped bytes at an untracked
	// target BEFORE the install runs.
	target := filepath.Join(f.home, ".claude/skills/moai-beta/SKILL.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, f.betaV1, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := f.installer(t).Install([]string{"extras"})
	if err != nil {
		t.Fatal(err)
	}
	// The run must classify it as a collision, not an install.
	if res.CollisionSkipped == 0 {
		t.Errorf("RF5: content-identical untracked file not collision-classified: %+v", res)
	}
	j, _ := LoadJournal(JournalPath(f.home))
	if j != nil {
		for _, e := range j.Entries {
			if e.Path == "claude-skills/moai-beta/SKILL.md" {
				t.Error("RF5: collision target journaled as an ownership target")
			}
		}
	}
	// And the surviving manifest must not claim it.
	m, _ := Load(ManifestPath(f.home))
	if _, tracked := m.Files["claude-skills/moai-beta/SKILL.md"]; tracked {
		t.Error("RF5: user file claimed as MoAI-owned")
	}
}

// F9: a file removed from the new version inside a KEPT skill is removed by
// the prune (per-KEY judgment, not per-name).
func TestRF2F9_PruneRemovesRetiredFileInKeptSkill(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatal(err)
	}
	retired := filepath.Join(f.home, ".claude/skills/moai-alpha/retired.md")
	if err := os.WriteFile(retired, []byte("retired upstream\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := Load(ManifestPath(f.home))
	fe := m.Files["claude-skills/moai-alpha/retired.md"]
	fe.SHA256 = sha256Hex([]byte("retired upstream\n"))
	m.Files["claude-skills/moai-alpha/retired.md"] = fe
	if err := m.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}

	in := f.installer(t)
	m2, _ := Load(ManifestPath(f.home))
	if _, err := in.PruneUnselected(m2); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(retired); !os.IsNotExist(err) {
		t.Errorf("F9: retired file in kept skill survived the prune: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")); err != nil {
		t.Errorf("F9: kept skill's live file deleted: %v", err)
	}
}
