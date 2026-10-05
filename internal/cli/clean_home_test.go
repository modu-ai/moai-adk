package cli

// SPEC-V3R6-MOAI-CLEAN-HOME-001 REQ-MCH-003/004/007/008 — behavior of
// `moai clean --home`: dry-run by default, --force deletes only the allowlist,
// retention from the home-tier state.yaml. Fixtures hermetic per REQ-MCH-008.

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
)

// snapshotTree records size+mtime for every regular file under root so tests
// can prove a dry-run mutated nothing (AC-MCH-004).
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return nil
		}
		snap[filepath.ToSlash(rel)] = fmt.Sprintf("%d:%d", info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return snap
}

// newHomeTestPrinter builds a Printer writing into capture buffers.
func newHomeTestPrinter() (printer.Printer, *bytes.Buffer, *bytes.Buffer) {
	out, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	return printer.New(printer.WithWriters(out, errBuf)), out, errBuf
}

// seedHomeCleanFixture plants aged + fresh entries in every allowlisted
// category plus protected neighbors, returning the fixture ~/.moai root.
func seedHomeCleanFixture(t *testing.T) string {
	t.Helper()
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")

	// debug/: aged deleted, fresh survives.
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "p1", "debug", "old.log"), 100, agedTime(t))
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "p1", "debug", "fresh.log"), 50, time.Time{})

	// logs/: aged deleted, fresh survives.
	writeHomeFixtureFile(t, filepath.Join(root, "logs", "old-run.log"), 30, agedTime(t))
	writeHomeFixtureFile(t, filepath.Join(root, "logs", "fresh-run.log"), 30, time.Time{})

	// backups/: aged removed-* deleted; fresh removed-* and non-removed survive.
	// The scanner ages a removed-* dir by the DIR's own mtime, so pin that too.
	writeHomeFixtureFile(t, filepath.Join(root, "backups", "removed-old", "x.yaml"), 20, agedTime(t))
	if err := os.Chtimes(filepath.Join(root, "backups", "removed-old"), agedTime(t), agedTime(t)); err != nil {
		t.Fatalf("chtimes removed-old: %v", err)
	}
	writeHomeFixtureFile(t, filepath.Join(root, "backups", "removed-fresh", "y.yaml"), 20, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "backups", "other", "z.yaml"), 20, agedTime(t))
	if err := os.Chtimes(filepath.Join(root, "backups", "other"), agedTime(t), agedTime(t)); err != nil {
		t.Fatalf("chtimes other: %v", err)
	}

	// releases/: five non-current binaries with distinct mtimes (v0.0.5 is
	// the newest, v0.0.1 the oldest) -> the 3 newest survive, the 2 oldest
	// (+ their .sha256 sidecars) are deleted; version.json survives untouched.
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("moai-v0.0.%d-darwin-arm64", i)
		mtime := time.Now().AddDate(0, 0, -(6 - i))
		writeHomeFixtureFile(t, filepath.Join(root, "releases", name), 40, mtime)
		writeHomeFixtureFile(t, filepath.Join(root, "releases", name+".sha256"), 4, mtime)
	}
	writeHomeFixtureFile(t, filepath.Join(root, "releases", "version.json"), 60, time.Time{})

	// Protected neighbor (full carve-out matrix is the M3 guard test).
	writeHomeFixtureFile(t, filepath.Join(root, "projects", "keep.jsonl"), 10, agedTime(t))
	return root
}

// TestCleanHome_DryRunMutatesNothing (AC-MCH-004): without --force the tree
// is byte-identical (file count + sizes + mtimes) before and after.
func TestCleanHome_DryRunMutatesNothing(t *testing.T) {
	root := seedHomeCleanFixture(t)
	home := filepath.Dir(root)
	before := snapshotTree(t, home)

	p, _, errBuf := newHomeTestPrinter()
	if err := runCleanHome(p, false); err != nil {
		t.Fatalf("runCleanHome(dry-run): %v", err)
	}

	after := snapshotTree(t, home)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("dry-run mutated the fixture tree:\nbefore=%v\nafter=%v", before, after)
	}
	out := errBuf.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("dry-run output should carry the [dry-run] marker, got %q", out)
	}
	if !strings.Contains(out, "Would delete") {
		t.Errorf("dry-run output should list what would be deleted, got %q", out)
	}
}

// TestCleanHome_ForceDeletesOnlyAllowlistedCategories (AC-MCH-005): --force
// removes aged debug/, the 2 oldest releases (+ sidecars), aged logs/, aged
// backups/removed-* — and nothing else.
func TestCleanHome_ForceDeletesOnlyAllowlistedCategories(t *testing.T) {
	root := seedHomeCleanFixture(t)

	p, _, errBuf := newHomeTestPrinter()
	if err := runCleanHome(p, true); err != nil {
		t.Fatalf("runCleanHome(force): %v", err)
	}

	deleted := []string{
		filepath.Join(root, "claude-profiles", "p1", "debug", "old.log"),
		filepath.Join(root, "logs", "old-run.log"),
		filepath.Join(root, "backups", "removed-old"),
		filepath.Join(root, "releases", "moai-v0.0.1-darwin-arm64"),
		filepath.Join(root, "releases", "moai-v0.0.1-darwin-arm64.sha256"),
		filepath.Join(root, "releases", "moai-v0.0.2-darwin-arm64"),
		filepath.Join(root, "releases", "moai-v0.0.2-darwin-arm64.sha256"),
	}
	for _, path := range deleted {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("aged entry should be deleted under --force: %s (stat err=%v)", path, err)
		}
	}

	survive := []string{
		filepath.Join(root, "claude-profiles", "p1", "debug", "fresh.log"),
		filepath.Join(root, "logs", "fresh-run.log"),
		filepath.Join(root, "backups", "removed-fresh", "y.yaml"),
		filepath.Join(root, "backups", "other", "z.yaml"),
		filepath.Join(root, "releases", "moai-v0.0.3-darwin-arm64"),
		filepath.Join(root, "releases", "moai-v0.0.4-darwin-arm64"),
		filepath.Join(root, "releases", "moai-v0.0.5-darwin-arm64"),
		filepath.Join(root, "releases", "version.json"),
		filepath.Join(root, "projects", "keep.jsonl"),
	}
	for _, path := range survive {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("non-eligible entry must survive --force: %s (%v)", path, err)
		}
	}
	if !strings.Contains(errBuf.String(), "Deleted") {
		t.Errorf("force output should report deletions, got %q", errBuf.String())
	}
}

// TestCleanHome_RetentionFromHomeTier (AC-MCH-008): the cutoff comes from
// state.home_retention_days in the HOME tier — explicit value honored, absent
// key falls back to the compiled default, explicit 0 disables cleaning.
func TestCleanHome_RetentionFromHomeTier(t *testing.T) {
	writeTier := func(t *testing.T, home, days string) {
		t.Helper()
		path := filepath.Join(home, ".moai", "config", "sections", "state.yaml")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		content := "state:\n"
		if days != "" {
			content += "  home_retention_days: " + days + "\n"
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write state.yaml: %v", err)
		}
	}

	t.Run("explicit value shortens the window", func(t *testing.T) {
		home := hermeticHomeEnv(t)
		root := filepath.Join(home, ".moai")
		writeTier(t, home, "10")
		writeHomeFixtureFile(t, filepath.Join(root, "logs", "twenty-days.log"), 10, time.Now().AddDate(0, 0, -20))
		writeHomeFixtureFile(t, filepath.Join(root, "logs", "five-days.log"), 10, time.Now().AddDate(0, 0, -5))

		p, _, _ := newHomeTestPrinter()
		if err := runCleanHome(p, true); err != nil {
			t.Fatalf("runCleanHome: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "logs", "twenty-days.log")); !os.IsNotExist(err) {
			t.Errorf("20-day-old entry should be deleted under a 10-day retention (stat err=%v)", err)
		}
		if _, err := os.Stat(filepath.Join(root, "logs", "five-days.log")); err != nil {
			t.Errorf("5-day-old entry should survive a 10-day retention: %v", err)
		}
	})

	t.Run("absent key falls back to the compiled default", func(t *testing.T) {
		home := hermeticHomeEnv(t)
		root := filepath.Join(home, ".moai")
		writeTier(t, home, "")
		writeHomeFixtureFile(t, filepath.Join(root, "logs", "very-old.log"), 10,
			time.Now().AddDate(0, 0, -(config.DefaultHomeCleanRetentionDays+10)))
		writeHomeFixtureFile(t, filepath.Join(root, "logs", "mid-age.log"), 10,
			time.Now().AddDate(0, 0, -20))

		p, _, _ := newHomeTestPrinter()
		if err := runCleanHome(p, true); err != nil {
			t.Fatalf("runCleanHome: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "logs", "very-old.log")); !os.IsNotExist(err) {
			t.Errorf("entry older than the default retention should be deleted (stat err=%v)", err)
		}
		if _, err := os.Stat(filepath.Join(root, "logs", "mid-age.log")); err != nil {
			t.Errorf("20-day-old entry should survive the %d-day default: %v", config.DefaultHomeCleanRetentionDays, err)
		}
	})

	t.Run("explicit zero disables cleaning", func(t *testing.T) {
		home := hermeticHomeEnv(t)
		root := filepath.Join(home, ".moai")
		writeTier(t, home, "0")
		writeHomeFixtureFile(t, filepath.Join(root, "logs", "ancient.log"), 10, time.Now().AddDate(0, 0, -400))
		logsDir := filepath.Join(root, "logs")
		if err := os.Chmod(logsDir, 0o755); err != nil {
			t.Fatalf("chmod logs: %v", err)
		}

		p, _, errBuf := newHomeTestPrinter()
		if err := runCleanHome(p, true); err != nil {
			t.Fatalf("runCleanHome: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "logs", "ancient.log")); err != nil {
			t.Errorf("explicit 0 must disable cleaning entirely: %v", err)
		}
		if !strings.Contains(errBuf.String(), "disabled") {
			t.Errorf("disabled run should say so, got %q", errBuf.String())
		}
		info, err := os.Stat(logsDir)
		if err != nil {
			t.Fatalf("stat repaired logs dir: %v", err)
		}
		if info.Mode().Perm() != 0o700 {
			t.Errorf("force must repair directory permissions even when retention is disabled: mode=%v", info.Mode().Perm())
		}
	})
}

func TestScanHomeCleanable_PreservesActiveProfile(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	profileRoot := filepath.Join(root, "claude-profiles", "active")
	project := filepath.Join(profileRoot, "projects", "old-project")
	debug := filepath.Join(profileRoot, "debug", "old.log")
	writeHomeFixtureFile(t, filepath.Join(project, "session.jsonl"), 10, agedTime(t))
	writeHomeFixtureFile(t, debug, 10, agedTime(t))
	if err := os.Chtimes(project, agedTime(t), agedTime(t)); err != nil {
		t.Fatalf("chtimes project: %v", err)
	}
	t.Setenv("CLAUDE_CONFIG_DIR", profileRoot)

	candidates := scanHomeCleanable(root, 30, config.DefaultReleaseKeep, "v-test", time.Now())
	for _, candidate := range candidates {
		if candidate.AbsPath == project || candidate.AbsPath == debug {
			t.Fatalf("active profile path selected for deletion: %s", candidate.AbsPath)
		}
	}
}

func TestScanHomeCleanable_ProfileProjectsRetentionAndCap(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	projects := filepath.Join(root, "claude-profiles", "p1", "projects")
	if err := os.MkdirAll(projects, 0o700); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(projects, "old.jsonl")
	fresh := filepath.Join(projects, "fresh.jsonl")
	for _, path := range []string{old, fresh} {
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(3 * 1024 * 1024 * 1024); err != nil {
			_ = f.Close()
			t.Fatal(err)
		}
		_ = f.Close()
	}
	now := time.Now()
	oldAt := now.AddDate(0, 0, -200)
	if err := os.Chtimes(old, oldAt, oldAt); err != nil {
		t.Fatal(err)
	}
	protected := filepath.Join(projects, "old-project", "credentials-cache")
	if err := os.MkdirAll(protected, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(protected, "token"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(projects, "old-project"), oldAt, oldAt); err != nil {
		t.Fatal(err)
	}
	candidates := scanHomeCleanable(root, 30, 3, "v-test", now)
	seenOld, seenFresh, seenProtected := false, false, false
	for _, c := range candidates {
		if c.AbsPath == old {
			seenOld = true
		}
		if c.AbsPath == fresh {
			seenFresh = true
		}
		if c.AbsPath == filepath.Join(projects, "old-project") {
			seenProtected = true
		}
	}
	if !seenOld {
		t.Error("180-day retention must select the old projects entry")
	}
	if seenFresh {
		t.Error("after removing the old 3 GiB entry the profile is below 5 GiB; fresh entry must survive")
	}
	if seenProtected {
		t.Error("a project entry containing a credentials* descendant must be protected")
	}
}

func TestSecureHomeDirectoriesRepairsMode0700(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX modes are not represented on Windows")
	}
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	nested := filepath.Join(root, "claude-profiles", "p1", "projects")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	nonProfile := filepath.Join(root, "cache", "search")
	if err := os.MkdirAll(nonProfile, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, filepath.Dir(nested), nonProfile} {
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := secureHomeDirectories(root); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, filepath.Join(root, "claude-profiles"), filepath.Join(root, "claude-profiles", "p1"), nested, filepath.Join(root, "cache"), nonProfile} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o700 {
			t.Errorf("%s mode=%#o, want 0700", path, got)
		}
	}
}

// TestCleanHome_NoHomeIsNoop: an absent ~/.moai home is an informative
// no-op, never an error.
func TestCleanHome_NoHomeIsNoop(t *testing.T) {
	hermeticHomeEnv(t)

	p, _, errBuf := newHomeTestPrinter()
	if err := runCleanHome(p, true); err != nil {
		t.Fatalf("absent home should not error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "nothing to clean") {
		t.Errorf("absent home should report nothing to clean, got %q", errBuf.String())
	}
}

// TestCleanHome_HomeFlagWiring: the cobra surface — `clean --home` parses,
// routes to the home scope, and stays dry-run by default.
func TestCleanHome_HomeFlagWiring(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	writeHomeFixtureFile(t, filepath.Join(root, "logs", "old.log"), 10, agedTime(t))

	cmd := newCleanCmd()
	if cmd.Flags().Lookup("home") == nil {
		t.Fatal("clean command should expose a --home flag")
	}
	cmd.SetArgs([]string{"--home"})
	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("clean --home should execute without error, got %v", err)
	}
	if !strings.Contains(errBuf.String(), "[dry-run]") {
		t.Errorf("clean --home should dry-run by default, got %q", errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(root, "logs", "old.log")); err != nil {
		t.Errorf("default invocation must not delete: %v", err)
	}
}
