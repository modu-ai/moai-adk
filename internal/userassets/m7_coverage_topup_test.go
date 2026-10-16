// m7_coverage_topup_test.go — M7.1 coverage top-up (gate round 42
// follow-up): closes the DETERMINISTIC error and refusal arms of the lock
// family and the confined-write boundary that the milestone suites left
// uncovered. Every injection here is platform-independent (path
// occupation, path kind, symlinks) per the M0 standing rule — the arms
// that need unix permission semantics or flock contention live in
// m7_coverage_topup_unix_test.go, and the race arms (mid-write I/O
// failure, rename-window races, exhausted-retry) are declared irreducible
// in progress.md §E.2.
package userassets

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
)

// resolvedSkillsRoot resolves the claude-skills root for a temp home, the
// boundary every confinedWrite case below writes through.
func resolvedSkillsRoot(t *testing.T, home string) resolvedRoot {
	t.Helper()
	root, err := resolveRoot(home, Root{Slug: RootClaudeSkills, Dir: filepath.Join(home, ".claude", "skills")})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestLockFamilyCoverageArms(t *testing.T) {
	t.Run("release on a nil receiver is a no-op", func(t *testing.T) {
		var l *UserLock
		if err := l.Release(); err != nil {
			t.Fatalf("nil release returned %v", err)
		}
	})

	t.Run("release skips a lock reclaimed by another owner", func(t *testing.T) {
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("pid=1 token=other acquired=2026-10-09T00:00:00Z\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		l := &UserLock{path: path, token: "mine"}
		err := l.Release()
		if err == nil || !strings.Contains(err.Error(), "reclaimed") {
			t.Fatalf("release of a reclaimed lock returned %v, want reclaimed-owner error", err)
		}
		if _, statErr := os.Stat(path); statErr != nil {
			t.Fatalf("the new owner's lock was deleted: %v", statErr)
		}
	})

	t.Run("state string has an unknown fallback", func(t *testing.T) {
		if got := GuardMarkerState(99).String(); got != "unknown" {
			t.Fatalf("GuardMarkerState(99) = %q, want unknown", got)
		}
	})

	t.Run("classify lock file arms", func(t *testing.T) {
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if state, _ := ClassifyLockFile(path); state != GuardMarkerAbsent {
			t.Fatalf("absent lock classified %s, want absent", state)
		}
		if err := os.WriteFile(path, []byte("no ownership record here\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if state, _ := ClassifyLockFile(path); state != GuardMarkerOwnerless {
			t.Fatalf("pid-less lock classified %s, want ownerless", state)
		}
		dir := LockPath(home) + "-as-dir"
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if state, _ := ClassifyLockFile(dir); state != GuardMarkerIrregular {
			t.Fatalf("directory at lock path classified %s, want irregular", state)
		}
		deadPID := deadProcessPID(t)
		dead := "pid=" + itoaTest(deadPID) + " token=x acquired=2026-10-09T00:00:00Z\n"
		if err := os.WriteFile(path, []byte(dead), 0o644); err != nil {
			t.Fatal(err)
		}
		state, pid := ClassifyLockFile(path)
		if state != GuardMarkerOwnerDead {
			t.Fatalf("dead-owner lock classified %s, want owner-dead", state)
		}
		if pid != deadPID {
			t.Fatalf("dead-owner pid = %d, want %d", pid, deadPID)
		}
		alive := "pid=" + itoaTest(os.Getpid()) + " token=self acquired=2026-10-09T00:00:00Z\n"
		if err := os.WriteFile(path, []byte(alive), 0o644); err != nil {
			t.Fatal(err)
		}
		if state, _ := ClassifyLockFile(path); state != GuardMarkerOwnerAlive {
			t.Fatalf("live-owner lock classified %s, want owner-alive", state)
		}
	})
}

func TestAcquireLockErrorArms(t *testing.T) {
	t.Run("uncreatable lock home is an error", func(t *testing.T) {
		home := t.TempDir()
		// A FILE occupies the lock's parent directory name — MkdirAll
		// cannot create it on any platform.
		if err := os.WriteFile(MoaiHome(home), []byte("not a dir\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := AcquireUserLock(home, 50*time.Millisecond); err == nil || !strings.Contains(err.Error(), "mkdir lock home") {
			t.Fatalf("acquire over file-occupied lock home returned %v, want mkdir error", err)
		}
	})

	t.Run("lock path occupied by a directory is a non-EEXIST error", func(t *testing.T) {
		home := t.TempDir()
		if err := os.MkdirAll(LockPath(home), 0o755); err != nil {
			t.Fatal(err)
		}
		_, err := AcquireUserLock(home, 50*time.Millisecond)
		if err == nil || strings.Contains(err.Error(), "ErrLocked") {
			t.Fatalf("acquire over directory-occupied lock path returned %v, want a hard (non-EEXIST) error", err)
		}
	})
}

func TestConfinedWriteRefusalArms(t *testing.T) {
	f := newFixture(t)
	inst := f.installer(t)
	home := f.home

	t.Run("invalid relative path", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		err := inst.confinedWrite(root, "../escape.txt", []byte("x"), false)
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("confinedWrite(../escape.txt) = %v, want ErrPathInvalid", err)
		}
	})

	t.Run("missing parent without mkdirs", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		err := inst.confinedWrite(root, "no/such/leaf.txt", []byte("x"), false)
		if err == nil || !strings.Contains(err.Error(), "resolve destination parent") {
			t.Fatalf("confinedWrite into an absent parent = %v, want resolve error", err)
		}
	})

	t.Run("parent symlink escaping the root is refused", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		outside := filepath.Join(home, "outside")
		if err := os.MkdirAll(outside, 0o755); err != nil {
			t.Fatal(err)
		}
		symlinkOrSkip(t, outside, filepath.Join(root.dir, "link-out"))
		err := inst.confinedWrite(root, "link-out/leaf.txt", []byte("x"), false)
		if err == nil || !strings.Contains(err.Error(), "resolves outside root") {
			t.Fatalf("escape-symlink parent = %v, want C2 refusal", err)
		}
	})

	t.Run("leaf symlink is never written through", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		victim := filepath.Join(home, "innocent-leaf.txt")
		if err := os.WriteFile(victim, []byte("original\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		symlinkOrSkip(t, victim, filepath.Join(root.dir, "leaf-link.txt"))
		err := inst.confinedWrite(root, "leaf-link.txt", []byte("payload\n"), false)
		if err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("leaf-symlink write = %v, want C2 leaf refusal", err)
		}
		if got, readErr := os.ReadFile(victim); readErr != nil || string(got) != "original\n" {
			t.Fatalf("symlink target was modified: %q (%v)", got, readErr)
		}
	})

	t.Run("destination occupied by a directory fails the rename", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		if err := os.MkdirAll(filepath.Join(root.dir, "occupied.txt"), 0o755); err != nil {
			t.Fatal(err)
		}
		err := inst.confinedWrite(root, "occupied.txt", []byte("x"), false)
		if err == nil {
			t.Fatal("write over a directory-occupied destination succeeded, want a rename error")
		}
	})

	t.Run("path segment occupied by a file is not a directory", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		if err := os.WriteFile(filepath.Join(root.dir, "blocker"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		err := inst.confinedWrite(root, "blocker/leaf.txt", []byte("x"), true)
		if err == nil || !strings.Contains(err.Error(), "is not a directory") {
			t.Fatalf("write through file-occupied segment = %v, want not-a-directory error", err)
		}
	})

	t.Run("segment symlink outside the root is refused", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		outside := filepath.Join(home, "outside-two")
		if err := os.MkdirAll(outside, 0o755); err != nil {
			t.Fatal(err)
		}
		symlinkOrSkip(t, outside, filepath.Join(root.dir, "link-mkdir"))
		err := inst.confinedWrite(root, "link-mkdir/leaf.txt", []byte("x"), true)
		if err == nil || !strings.Contains(err.Error(), "resolves outside root") {
			t.Fatalf("escape-symlink segment = %v, want C2 refusal", err)
		}
	})

	t.Run("segment symlink inside the root is a legal boundary", func(t *testing.T) {
		root := resolvedSkillsRoot(t, home)
		realDir := filepath.Join(root.dir, "real-dir")
		if err := os.MkdirAll(realDir, 0o755); err != nil {
			t.Fatal(err)
		}
		symlinkOrSkip(t, realDir, filepath.Join(root.dir, "link-in"))
		if err := inst.confinedWrite(root, "link-in/leaf.txt", []byte("payload\n"), true); err != nil {
			t.Fatalf("in-root symlink segment refused: %v", err)
		}
		if _, statErr := os.Stat(filepath.Join(realDir, "leaf.txt")); statErr != nil {
			t.Fatalf("payload did not land at the resolved target: %v", statErr)
		}
	})
}

func TestInstallEntryArms(t *testing.T) {
	t.Run("installer without catalog or source is refused", func(t *testing.T) {
		if _, err := (&Installer{Home: t.TempDir()}).Install(nil); err == nil || !strings.Contains(err.Error(), "needs a catalog and a source tree") {
			t.Fatalf("bare installer Install = %v, want construction error", err)
		}
	})

	t.Run("unknown bundle is refused before anything saves", func(t *testing.T) {
		f := newFixture(t)
		if _, err := f.installer(t).Install([]string{"no-such-pack"}); err == nil || !strings.Contains(err.Error(), "unknown bundle") {
			t.Fatalf("unknown bundle Install = %v, want unknown-bundle error", err)
		}
	})

	t.Run("unresolvable root refuses the run", func(t *testing.T) {
		f := newFixture(t)
		// A FILE at ~/.claude makes MkdirAll of every root under it fail.
		if err := os.WriteFile(filepath.Join(f.home, ".claude"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := f.installer(t).Install(nil); err == nil || !strings.Contains(err.Error(), "resolve root") {
			t.Fatalf("Install over file-occupied .claude = %v, want root-resolve error", err)
		}
	})

	t.Run("unsupported-schema journal refuses the run in place", func(t *testing.T) {
		f := newFixture(t)
		raw := "{\n  \"schema_version\": 99,\n  \"bundles_selection\": [\"extras\"],\n  \"started_at\": \"2026-10-09T00:00:00Z\"\n}\n"
		if err := os.MkdirAll(MoaiHome(f.home), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(JournalPath(f.home), []byte(raw), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := f.installer(t).Install(nil)
		if err == nil || !strings.Contains(err.Error(), "schema_version") {
			t.Fatalf("Install over schema-99 journal = %v, want schema refusal", err)
		}
		if _, statErr := os.Stat(JournalPath(f.home)); statErr != nil {
			t.Fatalf("the journal must stay at its original path: %v", statErr)
		}
	})

	t.Run("journal selection is restored for a bare retry", func(t *testing.T) {
		f := newFixture(t)
		inst := f.installer(t)
		j := &PendingJournal{BundlesSelection: []string{"extras"}}
		if err := WriteJournal(JournalPath(f.home), j); err != nil {
			t.Fatal(err)
		}
		res, err := inst.Install(nil)
		if err != nil {
			t.Fatalf("retry with a pending journal selection: %v", err)
		}
		if res.Installed == 0 {
			t.Fatal("the journal's interrupted request was not restored — nothing installed")
		}
		manifest, loadErr := Load(ManifestPath(f.home))
		if loadErr != nil {
			t.Fatal(loadErr)
		}
		if len(manifest.Bundles) != 1 || manifest.Bundles[0] != "extras" {
			t.Fatalf("restored bundles = %v, want [extras]", manifest.Bundles)
		}
	})
}

func TestInstallerNowFallsBackToWallClock(t *testing.T) {
	in := &Installer{}
	if in.now().IsZero() {
		t.Fatal("now() with a nil Now hook returned the zero time")
	}
}

// TestMalformedSourceArms drives the installTargets error propagation with
// a catalog/source pair that violates the deploy-shape contract — each
// case names the failure the installer must refuse (fail-closed before any
// write).
func TestMalformedSourceArms(t *testing.T) {
	brokenFixture := func(t *testing.T, mutate func(cat *template.Catalog, src fstest.MapFS)) *Installer {
		t.Helper()
		f := newFixture(t)
		mutate(f.cat, f.src)
		return f.installer(t)
	}

	t.Run("agent entry without a codex TOML is refused", func(t *testing.T) {
		inst := brokenFixture(t, func(_ *template.Catalog, src fstest.MapFS) {
			delete(src, ".codex/agents/moai/manager-x.toml")
		})
		_, err := inst.Install(nil)
		if err == nil || !strings.Contains(err.Error(), "codex agent TOML") {
			t.Fatalf("missing codex TOML = %v, want a refusal naming it", err)
		}
	})

	t.Run("source file absent from the tree is refused", func(t *testing.T) {
		inst := brokenFixture(t, func(_ *template.Catalog, src fstest.MapFS) {
			delete(src, ".claude/agents/moai/manager-x.md")
		})
		_, err := inst.Install(nil)
		if err == nil || !strings.Contains(err.Error(), "read source") {
			t.Fatalf("missing source bytes = %v, want a read-source refusal", err)
		}
	})

	t.Run("entry with an invalid install name is refused", func(t *testing.T) {
		inst := brokenFixture(t, func(cat *template.Catalog, _ fstest.MapFS) {
			skills := cat.Catalog.Core.Skills
			skills[0].Name = "../escape"
			cat.Catalog.Core.Skills = skills
		})
		_, err := inst.Install(nil)
		if err == nil || !strings.Contains(err.Error(), "target ") {
			t.Fatalf("invalid entry name = %v, want a target refusal", err)
		}
	})
}
