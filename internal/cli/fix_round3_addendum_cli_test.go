// fix_round3_addendum_cli_test.go — addendum item 1 (card t1509): the
// migration's deletion path is anchored to the project root — a .claude
// swapped to an outside-pointing symlink cannot delete an external sentinel.
package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestFR3A1_MigrationDeletionAnchoredToProjectRoot(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "external.md")
	if err := os.WriteFile(sentinel, []byte("external\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A project whose .claude dir is an outside-pointing symlink.
	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, "seed"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "seed", "f.txt"), []byte("s"), 0o644); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, outside, filepath.Join(proj, ".claude"))
	if err := os.MkdirAll(filepath.Join(proj, ".claude", "skills", "moai-x"), 0o755); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(proj, ".claude", "skills", "moai-x", "SKILL.md")
	if err := os.WriteFile(live, []byte("template bytes\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(proj); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Track(".claude/skills/moai-x/SKILL.md", manifest.TemplateManaged, manifest.HashBytes([]byte("template bytes\n"))); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	if err := migrateProjectCommonAssets(proj, home, true, nil, func(string, ...interface{}) {}); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(sentinel); err != nil || string(got) != "external\n" {
		t.Errorf("Item 1: external sentinel affected: %q (%v)", got, err)
	}
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
