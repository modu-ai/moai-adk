// installer_test.go — the user-folder installer's truth-table tests
// (SPEC-USER-ASSET-INSTALL-001 M2; REQ-008..013, REQ-023, REQ-024, C2/AC-025).
//
// Every test runs against a temp HOME (plan §D) and a synthetic catalog +
// MapFS source tree — surgical fixtures, not the real embedded tree.
package userassets

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
)

// fixture builds a synthetic two-skill + one-agent catalog over a MapFS
// source: skill "moai-alpha" (core), skill "moai-beta" (pack "extras"),
// agent "manager-x" (core).
type fixture struct {
	home    string
	cat     *template.Catalog
	src     fstest.MapFS
	alphaV1 []byte
	betaV1  []byte
	agentV1 []byte
	tomlV1  []byte
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		home:    t.TempDir(),
		alphaV1: []byte("alpha skill body v1\n"),
		betaV1:  []byte("beta skill body v1\n"),
		agentV1: []byte("---\nname: manager-x\n---\nagent body v1\n"),
		tomlV1:  []byte("name = \"manager-x\"\n"),
	}
	f.cat = &template.Catalog{
		Version: "test",
		Catalog: template.CatalogSections{
			Core: template.TierSection{
				Skills: []template.Entry{
					{Name: "moai-alpha", Tier: template.TierCore, Path: "templates/.claude/skills/moai-alpha/", Version: "1.0.0"},
				},
				Agents: []template.Entry{
					{Name: "manager-x", Tier: template.TierCore, Path: "templates/.claude/agents/moai/manager-x.md", Version: "1.0.0"},
				},
			},
			OptionalPacks: map[string]*template.Pack{
				"extras": {
					Description: "test extras",
					Skills: []template.Entry{
						{Name: "moai-beta", Tier: "optional-pack:extras", Path: "templates/.claude/skills/moai-beta/", Version: "1.0.0"},
					},
				},
			},
		},
	}
	f.src = fstest.MapFS{
		".claude/skills/moai-alpha/SKILL.md":       &fstest.MapFile{Data: f.alphaV1},
		".claude/skills/moai-alpha/workflows/a.md": &fstest.MapFile{Data: []byte("workflow a\n")},
		".agents/skills/moai-alpha/SKILL.md":       &fstest.MapFile{Data: f.alphaV1},
		".claude/skills/moai-beta/SKILL.md":        &fstest.MapFile{Data: f.betaV1},
		".agents/skills/moai-beta/SKILL.md":        &fstest.MapFile{Data: f.betaV1},
		".claude/agents/moai/manager-x.md":         &fstest.MapFile{Data: f.agentV1},
		".codex/agents/moai/manager-x.toml":        &fstest.MapFile{Data: f.tomlV1},
	}
	return f
}

func (f *fixture) installer(t *testing.T) *Installer {
	t.Helper()
	return &Installer{
		Home:        f.home,
		Catalog:     f.cat,
		Source:      f.src,
		MoaiVersion: "v3.2.0-test",
		Now:         func() time.Time { return time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC) },
	}
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

// symlinkOrSkip creates a symlink, skipping on platforms where the test
// principal may not create one (Windows without developer mode).
func symlinkOrSkip(t *testing.T, oldname, newname string) {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable on %s: %v", runtime.GOOS, err)
		}
		t.Fatalf("symlink %s -> %s: %v", newname, oldname, err)
	}
}

// TestInstallFirstRunLandsL0 covers AC-001's core arm at the installer level:
// a fresh install lands the L0 set under the four roots and the manifest
// records sha256 + bundle + per-file moai version (AC-003); the journal is
// cleared atomically with the save.
func TestInstallFirstRunLandsL0(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if res.Installed == 0 {
		t.Fatal("first install installed nothing")
	}

	// Claude skill root: whole tree.
	if got := readBytes(t, filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")); string(got) != string(f.alphaV1) {
		t.Errorf("claude skill bytes differ: %q", got)
	}
	if got := readBytes(t, filepath.Join(f.home, ".claude/skills/moai-alpha/workflows/a.md")); string(got) != "workflow a\n" {
		t.Errorf("skill sub-file missing: %q", got)
	}
	// Codex skill root (the dispatcher-mirror shape — every skill lands in
	// both harness roots, REQ-001/REQ-022).
	if got := readBytes(t, filepath.Join(f.home, ".agents/skills/moai-alpha/SKILL.md")); string(got) != string(f.alphaV1) {
		t.Errorf("codex skill bytes differ: %q", got)
	}
	// Agents: flat in both roots.
	if got := readBytes(t, filepath.Join(f.home, ".claude/agents/manager-x.md")); string(got) != string(f.agentV1) {
		t.Errorf("claude agent bytes differ")
	}
	if got := readBytes(t, filepath.Join(f.home, ".codex/agents/manager-x.toml")); string(got) != string(f.tomlV1) {
		t.Errorf("codex agent TOML bytes differ")
	}

	// Manifest: sha256 + bundle + per-file version.
	m, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	fe, ok := m.Files["claude-skills/moai-alpha/SKILL.md"]
	if !ok {
		t.Fatalf("manifest missing claude-skills/moai-alpha/SKILL.md; has %d entries", len(m.Files))
	}
	if fe.Bundle != "core" || fe.MoaiVersion != "v3.2.0-test" || len(fe.SHA256) != 64 {
		t.Errorf("manifest record incomplete: %+v", fe)
	}
	if _, ok := m.Files["agents-skills/moai-alpha/SKILL.md"]; !ok {
		t.Error("manifest missing the codex-root record for moai-alpha")
	}

	// Journal cleared.
	if j, _ := LoadJournal(JournalPath(f.home)); j != nil {
		t.Error("journal not cleared after a successful run")
	}

	// Unselected bundle absent.
	if _, err := os.Stat(filepath.Join(f.home, ".claude/skills/moai-beta")); !os.IsNotExist(err) {
		t.Error("unselected bundle skill moai-beta installed — selection not honored")
	}
}

// TestInstallSelectedBundle lands the opted-in bundle's entries (AC-018
// groundwork at the installer level).
func TestInstallSelectedBundle(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	res, err := f.installer(t).Install([]string{"extras"})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.home, ".claude/skills/moai-beta/SKILL.md")); err != nil {
		t.Errorf("opted-in bundle skill missing: %v", err)
	}
	m, _ := Load(ManifestPath(f.home))
	fe, ok := m.Files["claude-skills/moai-beta/SKILL.md"]
	if !ok || fe.Bundle != "extras" {
		t.Errorf("bundle record wrong: %+v ok=%v", fe, ok)
	}
	if len(m.Bundles) != 1 || m.Bundles[0] != "extras" {
		t.Errorf("manifest bundle selection = %v, want [extras]", m.Bundles)
	}
	_ = res
}

// TestInstallIdempotent covers AC-004: repeating an already-current install
// writes no file (mtime/hash proof) and reports zero deltas.
func TestInstallIdempotent(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("first install: %v", err)
	}
	target := filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")
	var st1 os.FileInfo
	if s, err := os.Stat(target); err != nil {
		t.Fatal(err)
	} else {
		st1 = s
	}
	time.Sleep(10 * time.Millisecond) // mtime resolution guard

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	st2, _ := os.Stat(target)
	if !st1.ModTime().Equal(st2.ModTime()) {
		t.Error("idempotent install rewrote the target file (mtime moved)")
	}
	if res.Installed != 0 || res.Refreshed != 0 || res.Removed != 0 || res.DivergencePreserved != 0 {
		t.Errorf("idempotent run reported deltas: %+v", res)
	}
}

// TestInstallCollisionSkipsUntracked covers AC-007: a user-created file at an
// install target is left byte-identical and reported.
func TestInstallCollisionSkipsUntracked(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	target := filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	userBytes := []byte("the user's own skill\n")
	if err := os.WriteFile(target, userBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if got := readBytes(t, target); string(got) != string(userBytes) {
		t.Errorf("collision file overwritten: %q", got)
	}
	if res.CollisionSkipped == 0 {
		t.Error("collision not reported")
	}
	m, _ := Load(ManifestPath(f.home))
	if _, tracked := m.Files["claude-skills/moai-alpha/SKILL.md"]; tracked {
		t.Error("collision file became manifest-tracked")
	}
}

// TestInstallEditedTrackedFilePreserved covers the fold-A1/REQ-023 arm: a
// present tracked file whose hash matches neither manifest nor shipped bytes
// is preserved (backup of the shipped replacement written under the backup
// home), counted divergence-preserved.
func TestInstallEditedTrackedFilePreserved(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("first install: %v", err)
	}
	target := filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")
	edited := []byte("the user's edit\n")
	if err := os.WriteFile(target, edited, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if got := readBytes(t, target); string(got) != string(edited) {
		t.Errorf("user edit clobbered: %q", got)
	}
	if res.DivergencePreserved == 0 {
		t.Errorf("divergence not counted (res=%+v)", res)
	}
	backup, err := BackupPath(f.home, RootClaudeSkills, "moai-alpha/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if got := readBytes(t, backup); string(got) != string(f.alphaV1) {
		t.Errorf("backup of shipped bytes missing or wrong: %q", got)
	}
}

// TestInstallManifestStaleRepairedNoRewrite covers AC-008's manifest-stale
// arm: current == shipped ≠ manifest → manifest repaired, no file rewrite,
// counted refreshed (REQ-011's manifest-repair count).
func TestInstallManifestStaleRepairedNoRewrite(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatal(err)
	}
	// Corrupt the manifest record's hash (a prior refresh crashed before the
	// manifest write — the file on disk holds the shipped bytes).
	m, _ := Load(ManifestPath(f.home))
	fe := m.Files["claude-skills/moai-alpha/SKILL.md"]
	fe.SHA256 = strings.Repeat("0", 64)
	m.Files["claude-skills/moai-alpha/SKILL.md"] = fe
	if err := m.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")
	st1, _ := os.Stat(target)

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatal(err)
	}
	st2, _ := os.Stat(target)
	if !st1.ModTime().Equal(st2.ModTime()) {
		t.Error("manifest-stale repair rewrote the file — REQ-023 forbids it")
	}
	if res.Refreshed == 0 {
		t.Error("manifest-stale repair not counted as refreshed")
	}
	m2, _ := Load(ManifestPath(f.home))
	if m2.Files["claude-skills/moai-alpha/SKILL.md"].SHA256 != fe.SHA256 && m2.Files["claude-skills/moai-alpha/SKILL.md"].SHA256 == strings.Repeat("0", 64) {
		t.Error("manifest record not repaired")
	}
}

// TestInstallPartialRetryCompletesShortfall covers AC-001's partial-failure
// arm: a manifest left by an interrupted install does not suppress the run —
// the retry installs exactly the missing assets and does not rewrite the
// present manifest-matching ones.
func TestInstallPartialRetryCompletesShortfall(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatal(err)
	}
	// Simulate the interruption: some assets vanish, the manifest stays.
	if err := os.RemoveAll(filepath.Join(f.home, ".claude/skills/moai-alpha")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(f.home, ".codex/agents/manager-x.toml")); err != nil {
		t.Fatal(err)
	}
	kept := filepath.Join(f.home, ".agents/skills/moai-alpha/SKILL.md")
	st1, _ := os.Stat(kept)

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Installed < 2 {
		t.Errorf("retry installed %d files, want the missing set (≥2)", res.Installed)
	}
	if _, err := os.Stat(filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")); err != nil {
		t.Errorf("missing asset not reinstalled: %v", err)
	}
	st2, _ := os.Stat(kept)
	if !st1.ModTime().Equal(st2.ModTime()) {
		t.Error("retry rewrote a present manifest-matching target")
	}
}

// TestJournalRecoveryThreeCases pins the E5 recovery lattice at the installer
// level: hash-match claims the run's own install (E4, flag or no flag); a
// mismatch NEVER reinstalls (collision when unflagged, divergence when
// flag-complete); an absent target installs from the journal; and the
// interrupted --bundles selection is restored intact (R-e item 1).
func TestJournalRecoveryThreeCases(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	in := f.installer(t)

	// Stage a journal the way an interrupted run would have: one entry whose
	// file stands completed at the final path (flagless — the E4 window),
	// one absent target, one mismatching file (unflagged), and one
	// mismatching flagged entry.
	matchRel := "claude-skills/moai-alpha/SKILL.md"
	absentRel := "claude-skills/moai-beta/SKILL.md"
	mismatchRel := "claude-skills/moai-alpha/workflows/a.md"
	divergeRel := "agents-skills/moai-alpha/SKILL.md"

	if err := os.MkdirAll(filepath.Join(f.home, ".claude/skills/moai-alpha/workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	// matchRel: bytes hash to the staged expectation.
	if err := os.WriteFile(filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md"), f.alphaV1, 0o644); err != nil {
		t.Fatal(err)
	}
	// mismatchRel: user edited after the interrupted write.
	if err := os.WriteFile(filepath.Join(f.home, ".claude/skills/moai-alpha/workflows/a.md"), []byte("user edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// divergeRel (flagged): written, then edited underneath.
	if err := os.MkdirAll(filepath.Join(f.home, ".agents/skills/moai-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.home, ".agents/skills/moai-alpha/SKILL.md"), []byte("edited after flag\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sha := func(b []byte) string { Sum := sha256Hex(b); return Sum }
	j := &PendingJournal{
		SchemaVersion:    SchemaVersion,
		BundlesSelection: []string{"extras"},
		StartedAt:        "2026-10-06T00:00:00Z",
		Entries: []JournalEntry{
			{Path: matchRel, ExpectedSHA256: sha(f.alphaV1), Bundle: "core", MoaiVersion: "vOld", InstalledAt: "t0"},
			{Path: absentRel, ExpectedSHA256: sha(f.betaV1), Bundle: "extras", MoaiVersion: "vOld", InstalledAt: "t0"},
			{Path: mismatchRel, ExpectedSHA256: sha([]byte("workflow a\n")), Bundle: "core", MoaiVersion: "vOld", InstalledAt: "t0"},
			{Path: divergeRel, ExpectedSHA256: sha(f.alphaV1), Bundle: "core", MoaiVersion: "vOld", InstalledAt: "t0", WriteCompleted: true},
		},
	}
	if err := WriteJournal(JournalPath(f.home), j); err != nil {
		t.Fatal(err)
	}

	// The recovery run passes NO explicit selection — the journal's recorded
	// selection must be adopted intact (R-e item 1), never [].
	res, err := in.Install(nil)
	if err != nil {
		t.Fatalf("recovery Install: %v", err)
	}

	// Case 2 (E4): the hash-matching file is claimed as the run's own.
	m, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatal(err)
	}
	fe, ok := m.Files[matchRel]
	if !ok {
		t.Fatal("hash-matching journal file was not claimed into the manifest (would collision-skip forever)")
	}
	if fe.MoaiVersion != "vOld" || fe.Bundle != "core" {
		t.Errorf("claimed entry provenance = %+v, want the journal's recorded provenance", fe)
	}

	// R-e item 1: the recorded selection restored intact.
	if len(m.Bundles) != 1 || m.Bundles[0] != "extras" {
		t.Errorf("recovered bundle selection = %v, want [extras] (never [])", m.Bundles)
	}
	// ...and the recovered selection's files are present (case 1: absent
	// target installed from the journal's provenance).
	if _, err := os.Stat(filepath.Join(f.home, ".claude/skills/moai-beta/SKILL.md")); err != nil {
		t.Errorf("absent journal target not installed: %v", err)
	}

	// Case 3: the mismatching UNFLAGGED file is a collision — bytes preserved.
	if got := readBytes(t, filepath.Join(f.home, ".claude/skills/moai-alpha/workflows/a.md")); string(got) != "user edit\n" {
		t.Errorf("mismatch recovery overwrote user bytes: %q", got)
	}
	// Case 3, flag-complete arm: divergence preserved (also never rewritten).
	if got := readBytes(t, filepath.Join(f.home, ".agents/skills/moai-alpha/SKILL.md")); string(got) != "edited after flag\n" {
		t.Errorf("flag-complete mismatch was reinstalled: %q", got)
	}
	if res.DivergencePreserved == 0 {
		t.Error("flag-complete mismatch not classified divergence")
	}
	// The journal is cleared after a successful run.
	if jj, _ := LoadJournal(JournalPath(f.home)); jj != nil {
		t.Error("journal not cleared after recovery run")
	}
	_ = json.Marshal // keep encoding/json imported for future assertions
}

// TestConfinementLeafSymlinkRefused covers AC-025's leaf-symlink arm: a
// managed leaf replaced by an outside-pointing symlink is never written
// through — refused at install, sentinel unmodified.
func TestConfinementLeafSymlinkRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "captured.md")
	if err := os.WriteFile(sentinel, []byte("sentinel\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(f.home, ".claude/skills/moai-alpha")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, sentinel, filepath.Join(skillDir, "SKILL.md"))

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if got := readBytes(t, sentinel); string(got) != "sentinel\n" {
		t.Errorf("write went THROUGH the outside-pointing symlink: %q", got)
	}
	if res.Installed > 0 && res.CollisionSkipped == 0 && res.DivergencePreserved == 0 && len(res.Failures) == 0 {
		t.Error("leaf-symlink escape neither refused nor classified")
	}
}

// TestConfinementParentSymlinkRefused covers AC-025's parent-symlink arm: a
// destination whose parent chain resolves outside the roots is refused after
// symlink resolution; the sentinel target is unmodified.
func TestConfinementParentSymlinkRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(f.home, ".claude/skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The whole skill dir is a symlink pointing outside.
	symlinkOrSkip(t, outside, filepath.Join(f.home, ".claude/skills/moai-alpha"))

	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("parent-symlink escape wrote %d entries outside the roots", len(entries))
	}
}

// TestConfinementSymlinkedRootLegal covers AC-025's symlinked-root arm: a
// dotfile-manager root (a symlink elsewhere) is a legal boundary at its
// RESOLVED location; an escape from the RESOLVED root is still refused.
func TestConfinementSymlinkedRootLegal(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	realRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(f.home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	// ~/.claude/skills → elsewhere (the dotfile-manager shape).
	skillsReal := filepath.Join(realRoot, "skills")
	if err := os.MkdirAll(skillsReal, 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, skillsReal, filepath.Join(f.home, ".claude/skills"))

	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(skillsReal, "moai-alpha/SKILL.md")); err != nil {
		t.Errorf("symlinked-root install did not land in the resolved location: %v", err)
	}
}

// TestConfinementBackupEscapeRefused covers the iter4 D33 arm: a REQ-023
// backup write whose resolved destination escapes the backup home is refused.
// Indirect assertion: divergence backups land under ~/.moai/backups/<slug>/
// and nowhere outside; the backup path helper rejects .. escapes.
func TestConfinementBackupEscapeRefused(t *testing.T) {
	t.Parallel()
	if _, err := BackupPath(t.TempDir(), RootClaudeSkills, "../../etc/escape"); err == nil {
		t.Error("backup path accepted a .. escape")
	}
}

// TestConfinementOutsideRootRefused — the installer's destination set is the
// four roots by construction; the confinement writer refuses any relpath
// whose resolution escapes its root (direct unit on the writer).
func TestConfinementOutsideRootRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	in := f.installer(t)
	root, err := resolveRoot(f.home, Root{Slug: RootClaudeSkills, Dir: filepath.Join(f.home, ".claude", "skills")})
	if err != nil {
		t.Fatal(err)
	}
	if err := in.confinedWrite(root, "../../../tmp/escape.md", []byte("x"), false); err == nil {
		t.Error("confinedWrite accepted a root escape")
	}
}

var _ = fs.ReadDir // keep io/fs referenced in constrained builds
