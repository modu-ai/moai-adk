// t1619_install_coverage_test.go — card t1619 bundle A: changed-line coverage
// closure for the install, journal, remove, and lock-state paths of
// internal/userassets. Each test drives a real input through the production
// path and asserts the observed outcome. No production code is changed.
package userassets

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template"
)

var errT1619InjectedRead = errors.New("injected source read failure")

// t1619ReadFailingSource serves the catalog tree from a MapFS but fails the
// read of one path, so the stat that precedes that read still succeeds.
type t1619ReadFailingSource struct {
	fstest.MapFS
	failPath string
}

func (s t1619ReadFailingSource) ReadFile(name string) ([]byte, error) {
	if name == s.failPath {
		return nil, errT1619InjectedRead
	}
	return s.MapFS.ReadFile(name)
}

// t1619FailureReason reports whether res records a failure against path whose
// reason starts with prefix.
func t1619FailureReason(res *Result, path, prefix string) bool {
	if res == nil {
		return false
	}
	for _, f := range res.Failures {
		if f.Path == path && strings.HasPrefix(f.Reason, prefix) {
			return true
		}
	}
	return false
}

// t1619SabotageJournalAfterFirstTarget replaces the pending journal with an
// empty directory once the first installable target has been handled. Each
// later journal persist then fails at the rename onto a directory. The run's
// own ClearJournal removes the empty directory at the end.
func t1619SabotageJournalAfterFirstTarget(t *testing.T, inst *Installer, journal string) {
	t.Helper()
	fired := false
	inst.afterTargetPersist = func(installTarget) {
		if fired {
			return
		}
		fired = true
		if err := os.Remove(journal); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove journal before sabotage: %v", err)
		}
		if err := os.Mkdir(journal, 0o755); err != nil {
			t.Errorf("replace journal with a directory: %v", err)
		}
	}
}

// t1619SeedJournal writes a pending journal with the given entries and an
// empty bundle selection, so the selection-restore branch stays out of the way.
func t1619SeedJournal(t *testing.T, home string, entries ...JournalEntry) {
	t.Helper()
	j := &PendingJournal{
		SchemaVersion: SchemaVersion,
		StartedAt:     "2000-01-01T00:00:00Z",
		Entries:       entries,
	}
	if err := WriteJournal(JournalPath(home), j); err != nil {
		t.Fatalf("seed journal: %v", err)
	}
}

// t1619Contains reports whether list holds s.
func t1619Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// t1619EntryNames returns the sorted names of entries.
func t1619EntryNames(entries []template.Entry) []string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name)
	}
	sort.Strings(names)
	return names
}

// t1619CyclicCatalog adds two optional packs that depend on each other to the
// fixture catalog. Each pack carries one skill of its own.
func t1619CyclicCatalog(f *fixture) {
	f.cat.Catalog.OptionalPacks["cyc-a"] = &template.Pack{
		DependsOn: []string{"cyc-b"},
		Skills: []template.Entry{
			{Name: "cyc-skill-a", Tier: "optional-pack:cyc-a", Path: "templates/.claude/skills/cyc-skill-a/"},
		},
	}
	f.cat.Catalog.OptionalPacks["cyc-b"] = &template.Pack{
		DependsOn: []string{"cyc-a"},
		Skills: []template.Entry{
			{Name: "cyc-skill-b", Tier: "optional-pack:cyc-b", Path: "templates/.claude/skills/cyc-skill-b/"},
		},
	}
}

// TestInstallReportsCompletionFlagPersistFailure covers the per-file
// completion-flag persist. The asset write for each file lands, the journal
// persist after the first write fails, and the failure is reported against the
// journal path. The run itself does not fail.
func TestInstallReportsCompletionFlagPersistFailure(t *testing.T) {
	f := newFixture(t)
	inst := f.installer(t)
	journal := JournalPath(f.home)
	t1619SabotageJournalAfterFirstTarget(t, inst, journal)

	res, err := inst.Install(nil)
	if err != nil {
		t.Fatalf("Install returned %v; a failed per-file persist must not abort the run", err)
	}
	if !t1619FailureReason(res, journal, "persist completion flag: ") {
		t.Fatalf("no completion-flag persist failure against %s; failures = %+v", journal, res.Failures)
	}
	if got := readBytes(t, filepath.Join(f.home, ".claude", "skills", "moai-alpha", "SKILL.md")); string(got) != string(f.alphaV1) {
		t.Fatalf("the asset write did not land: %q", got)
	}
	if _, statErr := os.Lstat(journal); !os.IsNotExist(statErr) {
		t.Fatalf("the journal was not cleared with the manifest save: %v", statErr)
	}
}

// TestInstallReportsRefreshedHashPersistFailure covers the post-loop hash
// re-record. A journal-claimed file is refreshed, the carried journal entry's
// hash changes, and the journal persist that records the new hash fails.
func TestInstallReportsRefreshedHashPersistFailure(t *testing.T) {
	f := newFixture(t)
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("first install: %v", err)
	}
	agent := filepath.Join(f.home, ".claude", "agents", "manager-x.md")
	// An interrupted run journaled this file's bytes, and they are still on disk.
	// The retry claims them, and the refresh arm then rewrites them to the new
	// shipped bytes.
	t1619SeedJournal(t, f.home, JournalEntry{
		Path:           "claude-agents/manager-x.md",
		ExpectedSHA256: shaHex(f.agentV1),
		Bundle:         "core",
		MoaiVersion:    "v3.2.0-test",
		InstalledAt:    "2026-10-08T00:00:00Z",
	})
	agentV2 := []byte("---\nname: manager-x\n---\nagent body v2\n")
	f.src[".claude/agents/moai/manager-x.md"] = &fstest.MapFile{Data: agentV2}

	inst := f.installer(t)
	journal := JournalPath(f.home)
	t1619SabotageJournalAfterFirstTarget(t, inst, journal)
	res, err := inst.Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !t1619FailureReason(res, journal, "persist refreshed hashes: ") {
		t.Fatalf("no refreshed-hash persist failure against %s; failures = %+v", journal, res.Failures)
	}
	if got := readBytes(t, agent); string(got) != string(agentV2) {
		t.Fatalf("the refresh did not write the shipped bytes: %q", got)
	}
	m, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if got := m.Files["claude-agents/manager-x.md"].SHA256; got != shaHex(agentV2) {
		t.Fatalf("manifest records %s, want the refreshed bytes' hash %s", got, shaHex(agentV2))
	}
}

// TestInstallRefusesUnreadableCodexTOML covers the target-build failure for an
// agent's Codex TOML. The stat succeeds and the read fails, so the run refuses
// before any asset is written.
func TestInstallRefusesUnreadableCodexTOML(t *testing.T) {
	f := newFixture(t)
	inst := f.installer(t)
	inst.Source = t1619ReadFailingSource{MapFS: f.src, failPath: ".codex/agents/moai/manager-x.toml"}

	res, err := inst.Install(nil)
	if err == nil || !strings.Contains(err.Error(), "read source .codex/agents/moai/manager-x.toml") {
		t.Fatalf("Install = (%v, %v), want the Codex TOML read refusal", res, err)
	}
	if _, statErr := os.Lstat(filepath.Join(f.home, ".claude", "agents", "manager-x.md")); !os.IsNotExist(statErr) {
		t.Fatalf("an asset was written before the refusal: %v", statErr)
	}
}

// TestInstallCollidesJournalTargetThatIsNotARegularFile covers reconciliation
// of a journaled path that a directory now occupies. The path is never read or
// claimed. It is reported as a collision, and the directory is left in place.
func TestInstallCollidesJournalTargetThatIsNotARegularFile(t *testing.T) {
	f := newFixture(t)
	const key = "claude-skills/moai-alpha/SKILL.md"
	occupied := filepath.Join(f.home, ".claude", "skills", "moai-alpha", "SKILL.md")
	if err := os.MkdirAll(occupied, 0o755); err != nil {
		t.Fatal(err)
	}
	t1619SeedJournal(t, f.home, JournalEntry{
		Path:           key,
		ExpectedSHA256: shaHex(f.alphaV1),
		Bundle:         "core",
		MoaiVersion:    "v3.2.0-test",
		InstalledAt:    "2026-10-08T00:00:00Z",
	})

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !t1619Contains(res.Collisions, key) {
		t.Fatalf("collisions = %v, want %s", res.Collisions, key)
	}
	if res.CollisionSkipped != 1 {
		t.Fatalf("CollisionSkipped = %d, want 1 (the journaled directory only)", res.CollisionSkipped)
	}
	if fi, statErr := os.Lstat(occupied); statErr != nil || !fi.IsDir() {
		t.Fatalf("the directory at the journaled path was disturbed: %v", statErr)
	}
}

// TestApplyTargetRefusesSymlinkLeafOnManifestMatchRefresh covers the refresh
// arm's write refusal. The leaf is a symlink to bytes the manifest records, so
// the target classifies as manifest-match, and only the leaf policy stops the
// rewrite.
func TestApplyTargetRefusesSymlinkLeafOnManifestMatchRefresh(t *testing.T) {
	f := newFixture(t)
	root := resolvedSkillsRoot(t, f.home)
	leaf := filepath.Join(root.dir, "moai-alpha", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(leaf), 0o755); err != nil {
		t.Fatal(err)
	}
	linkTarget := filepath.Join(f.home, "elsewhere", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(linkTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(linkTarget, f.alphaV1, 0o644); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, linkTarget, leaf)

	const key = "claude-skills/moai-alpha/SKILL.md"
	manifest := &Manifest{
		SchemaVersion: SchemaVersion,
		Files: map[string]FileEntry{
			key: {SHA256: shaHex(f.alphaV1), Bundle: "core", MoaiVersion: "v3.2.0-test"},
		},
	}
	alphaV2 := []byte("alpha skill body v2\n")
	tgt := installTarget{
		manifestKey: key,
		root:        RootClaudeSkills,
		rel:         "moai-alpha/SKILL.md",
		sourcePath:  ".claude/skills/moai-alpha/SKILL.md",
		bundle:      "core",
		sha:         shaHex(alphaV2),
		data:        alphaV2,
	}
	res := &Result{}
	_, wrote, err := f.installer(t).applyTarget(tgt, manifest, map[RootSlug]resolvedRoot{RootClaudeSkills: root}, res)
	if err == nil || !strings.Contains(err.Error(), "C2 leaf policy") {
		t.Fatalf("applyTarget error = %v, want the leaf-symlink refusal", err)
	}
	if wrote {
		t.Fatal("the refused write was reported as landed")
	}
	if got := readBytes(t, linkTarget); string(got) != string(f.alphaV1) {
		t.Fatalf("the symlink target was rewritten through the leaf: %q", got)
	}
	if got := manifest.Files[key].SHA256; got != shaHex(f.alphaV1) {
		t.Fatalf("the manifest record moved on a refused write: %s", got)
	}
}

// TestApplyTargetReadsLazyTargetFromSource covers a target that carries no
// pre-read bytes. The source read supplies the shipped bytes, and the write
// lands with the recorded hash.
func TestApplyTargetReadsLazyTargetFromSource(t *testing.T) {
	f := newFixture(t)
	root := resolvedSkillsRoot(t, f.home)
	const key = "claude-skills/moai-alpha/SKILL.md"
	tgt := installTarget{
		manifestKey: key,
		root:        RootClaudeSkills,
		rel:         "moai-alpha/SKILL.md",
		sourcePath:  ".claude/skills/moai-alpha/SKILL.md",
		bundle:      "core",
		sha:         shaHex(f.alphaV1),
	}
	manifest := &Manifest{SchemaVersion: SchemaVersion, Files: map[string]FileEntry{}}
	res := &Result{}
	recorded, wrote, err := f.installer(t).applyTarget(tgt, manifest, map[RootSlug]resolvedRoot{RootClaudeSkills: root}, res)
	if err != nil {
		t.Fatalf("applyTarget: %v", err)
	}
	if !wrote || recorded != tgt.sha {
		t.Fatalf("applyTarget = (recorded %q, wrote %v), want (%q, true)", recorded, wrote, tgt.sha)
	}
	if got := readBytes(t, filepath.Join(root.dir, "moai-alpha", "SKILL.md")); string(got) != string(f.alphaV1) {
		t.Fatalf("the source bytes did not land: %q", got)
	}
}

// TestApplyTargetRefusesLazyTargetWithAbsentSource covers the same lazy target
// when its source is absent. The read failure refuses the target before any
// write, and nothing is recorded.
func TestApplyTargetRefusesLazyTargetWithAbsentSource(t *testing.T) {
	f := newFixture(t)
	root := resolvedSkillsRoot(t, f.home)
	const key = "claude-skills/moai-missing/SKILL.md"
	tgt := installTarget{
		manifestKey: key,
		root:        RootClaudeSkills,
		rel:         "moai-missing/SKILL.md",
		sourcePath:  ".claude/skills/moai-missing/SKILL.md",
		bundle:      "core",
		sha:         shaHex(f.alphaV1),
	}
	manifest := &Manifest{SchemaVersion: SchemaVersion, Files: map[string]FileEntry{}}
	res := &Result{}
	recorded, wrote, err := f.installer(t).applyTarget(tgt, manifest, map[RootSlug]resolvedRoot{RootClaudeSkills: root}, res)
	if err == nil || !strings.Contains(err.Error(), "read shipped bytes") {
		t.Fatalf("applyTarget error = %v, want the shipped-bytes read refusal", err)
	}
	if wrote || recorded != "" {
		t.Fatalf("applyTarget = (recorded %q, wrote %v) after a refused read", recorded, wrote)
	}
	if _, statErr := os.Lstat(filepath.Join(root.dir, "moai-missing", "SKILL.md")); !os.IsNotExist(statErr) {
		t.Fatalf("a file was written for an unreadable source: %v", statErr)
	}
}

// TestInstallReportsDivergentBackupFailure covers the divergent-target arm's
// backup step. A user edit on a tracked file is kept on disk. The backup write
// fails because the backup home is a regular file, so the target is reported
// as a failure rather than skipped silently.
func TestInstallReportsDivergentBackupFailure(t *testing.T) {
	f := newFixture(t)
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("first install: %v", err)
	}
	const key = "claude-skills/moai-alpha/SKILL.md"
	skill := filepath.Join(f.home, ".claude", "skills", "moai-alpha", "SKILL.md")
	userEdit := []byte("user edit, neither recorded nor shipped\n")
	if err := os.WriteFile(skill, userEdit, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(BackupHome(f.home), []byte("occupied\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !t1619FailureReason(res, key, "backup shipped bytes: ") {
		t.Fatalf("no backup failure against %s; failures = %+v", key, res.Failures)
	}
	if got := readBytes(t, skill); string(got) != string(userEdit) {
		t.Fatalf("the user's edit was overwritten: %q", got)
	}
}

// TestCollectEntriesFollowsDependencyCycleOnce covers the cycle guard of the
// install-side closure. Each pack's skill appears exactly once, and the walk
// terminates.
func TestCollectEntriesFollowsDependencyCycleOnce(t *testing.T) {
	f := newFixture(t)
	t1619CyclicCatalog(f)
	got := strings.Join(t1619EntryNames(f.installer(t).collectEntries([]string{"cyc-a"})), ",")
	const want = "cyc-skill-a,cyc-skill-b,manager-x,moai-alpha"
	if got != want {
		t.Fatalf("collectEntries names = %s, want %s", got, want)
	}
}

// TestPreservedEntriesFollowsDependencyCycleOnce covers the cycle guard and the
// dependency walk of the removal-side retention set.
func TestPreservedEntriesFollowsDependencyCycleOnce(t *testing.T) {
	f := newFixture(t)
	t1619CyclicCatalog(f)
	got := strings.Join(t1619EntryNames(f.installer(t).preservedEntries([]string{"cyc-a"})), ",")
	const want = "cyc-skill-a,cyc-skill-b,manager-x,moai-alpha"
	if got != want {
		t.Fatalf("preservedEntries names = %s, want %s", got, want)
	}
}

// TestPreservedEntriesIgnoresUnknownSelectionName covers the unknown-pack arm
// of the retention walk. An unknown name adds nothing beyond the core entries.
func TestPreservedEntriesIgnoresUnknownSelectionName(t *testing.T) {
	f := newFixture(t)
	got := strings.Join(t1619EntryNames(f.installer(t).preservedEntries([]string{"ghost-pack"})), ",")
	const want = "manager-x,moai-alpha"
	if got != want {
		t.Fatalf("preservedEntries names = %s, want %s", got, want)
	}
}

// TestRemoveBundleKeepsRemainingSelectionAndSkipsBareManifestKeys covers the
// manifest-attribution pass of RemoveBundle. A non-empty remaining selection
// keeps the files its packs own, and a manifest key with no root slug is
// passed over without being removed.
func TestRemoveBundleKeepsRemainingSelectionAndSkipsBareManifestKeys(t *testing.T) {
	f := newFixture(t)
	f.cat.Catalog.OptionalPacks["keep"] = &template.Pack{Description: "remaining selection"}
	inst := f.installer(t)
	if _, err := inst.Install([]string{"extras"}); err != nil {
		t.Fatalf("install extras: %v", err)
	}
	manifest, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	const bare = "bare-legacy-record"
	manifest.Files[bare] = FileEntry{SHA256: shaHex(f.alphaV1), Bundle: "legacy-bundle", MoaiVersion: "v3.1.0"}

	res, err := inst.RemoveBundle(manifest, "extras", []string{"keep"})
	if err != nil {
		t.Fatalf("RemoveBundle: %v", err)
	}
	if res.Removed != 2 {
		t.Fatalf("Removed = %d, want 2 (the extras skill under both roots)", res.Removed)
	}
	for _, p := range []string{
		filepath.Join(f.home, ".claude", "skills", "moai-beta", "SKILL.md"),
		filepath.Join(f.home, ".agents", "skills", "moai-beta", "SKILL.md"),
	} {
		if _, statErr := os.Lstat(p); !os.IsNotExist(statErr) {
			t.Fatalf("the unselected extras file survived removal: %s (%v)", p, statErr)
		}
	}
	if got := readBytes(t, filepath.Join(f.home, ".claude", "skills", "moai-alpha", "SKILL.md")); string(got) != string(f.alphaV1) {
		t.Fatalf("the core skill was touched by the removal: %q", got)
	}
	if _, ok := manifest.Files[bare]; !ok {
		t.Fatal("the bare manifest record was dropped by the attribution pass")
	}
}

// TestGuardMarkerStateStringNamesIrregular covers the irregular arm of the
// doctor-visible state names.
func TestGuardMarkerStateStringNamesIrregular(t *testing.T) {
	if got := GuardMarkerIrregular.String(); got != "irregular" {
		t.Fatalf("GuardMarkerIrregular.String() = %q, want %q", got, "irregular")
	}
}
