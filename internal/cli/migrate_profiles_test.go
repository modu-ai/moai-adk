// migrate_profiles_test.go — `moai migrate profiles`.
//
// The migration moves memory topic files from the per-profile stores
// (~/.moai/claude-profiles/<name>/projects/<slug>/memory) into the default
// store (~/.claude/projects/<slug>/memory) and merges the MEMORY.md indexes.
//
// The hazard it has to avoid is the one this project already lived through:
// a bulk operation over a user's home directory that loses content. So the
// tests here are mostly about what must survive — differing files, existing
// index entries, and a second run.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// profileFixture builds a home with one default store and one profile store.
func profileFixture(t *testing.T) (home, project, defaultMem, profileMem string) {
	t.Helper()
	home = t.TempDir()
	project = t.TempDir()
	slug := memoryProjectSlug(project)

	defaultMem = filepath.Join(home, ".claude", "projects", slug, "memory")
	profileMem = filepath.Join(home, ".moai", "claude-profiles", "p1", "projects", slug, "memory")
	for _, d := range []string{defaultMem, profileMem} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	return home, project, defaultMem, profileMem
}

func writeProfileFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func readProfileFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func TestMigrateProfilesMovesTopicFilesAndMergesIndex(t *testing.T) {
	home, project, defaultMem, profileMem := profileFixture(t)
	t.Setenv("HOME", home)

	writeProfileFile(t, defaultMem, "MEMORY.md", "# Memory Index\n\n- [A](feedback_a.md) — a\n")
	writeProfileFile(t, defaultMem, "feedback_a.md", "A\n")
	writeProfileFile(t, profileMem, "MEMORY.md", "# Memory Index\n\n- [B](project_b.md) — b\n")
	writeProfileFile(t, profileMem, "project_b.md", "B\n")

	res, err := migrateProfileMemory(project, migrateProfileOpts{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if res.Moved != 1 {
		t.Errorf("Moved = %d, want 1: %+v", res.Moved, res)
	}

	if got := readProfileFile(t, filepath.Join(defaultMem, "project_b.md")); got != "B\n" {
		t.Errorf("migrated body = %q, want %q", got, "B\n")
	}
	if _, err := os.Stat(filepath.Join(profileMem, "project_b.md")); !os.IsNotExist(err) {
		t.Error("source file still present after a move")
	}

	idx := readProfileFile(t, filepath.Join(defaultMem, "MEMORY.md"))
	if !strings.Contains(idx, "feedback_a.md") {
		t.Errorf("merge dropped the pre-existing index entry:\n%s", idx)
	}
	if !strings.Contains(idx, "project_b.md") {
		t.Errorf("merge did not carry the profile's index entry:\n%s", idx)
	}
}

// TestMigrateProfilesKeepsBothOnContentCollision is the data-loss guard: two
// stores can hold the same filename with different content, and neither may
// silently win.
func TestMigrateProfilesKeepsBothOnContentCollision(t *testing.T) {
	home, project, defaultMem, profileMem := profileFixture(t)
	t.Setenv("HOME", home)

	writeProfileFile(t, defaultMem, "feedback_x.md", "DEFAULT VERSION\n")
	writeProfileFile(t, profileMem, "feedback_x.md", "PROFILE VERSION\n")

	res, err := migrateProfileMemory(project, migrateProfileOpts{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if res.Renamed != 1 {
		t.Errorf("Renamed = %d, want 1: %+v", res.Renamed, res)
	}

	if got := readProfileFile(t, filepath.Join(defaultMem, "feedback_x.md")); got != "DEFAULT VERSION\n" {
		t.Errorf("the existing file was overwritten: %q", got)
	}
	kept := filepath.Join(defaultMem, "feedback_x__p1.md")
	if got := readProfileFile(t, kept); got != "PROFILE VERSION\n" {
		t.Errorf("colliding content was not preserved at %s: %q", kept, got)
	}
}

// TestMigrateProfilesSkipsIdenticalContent keeps the migration idempotent:
// the same file in both stores is already migrated, not a collision.
func TestMigrateProfilesSkipsIdenticalContent(t *testing.T) {
	home, project, defaultMem, profileMem := profileFixture(t)
	t.Setenv("HOME", home)

	writeProfileFile(t, defaultMem, "feedback_same.md", "SAME\n")
	writeProfileFile(t, profileMem, "feedback_same.md", "SAME\n")

	res, err := migrateProfileMemory(project, migrateProfileOpts{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if res.Skipped != 1 || res.Renamed != 0 {
		t.Errorf("identical content: Skipped=%d Renamed=%d, want 1/0", res.Skipped, res.Renamed)
	}
	if _, err := os.Stat(filepath.Join(defaultMem, "feedback_same__p1.md")); err == nil {
		t.Error("identical content produced a needless duplicate")
	}
}

// TestMigrateProfilesDryRunTouchesNothing pins that the preview is a preview.
func TestMigrateProfilesDryRunTouchesNothing(t *testing.T) {
	home, project, defaultMem, profileMem := profileFixture(t)
	t.Setenv("HOME", home)

	writeProfileFile(t, profileMem, "project_b.md", "B\n")

	res, err := migrateProfileMemory(project, migrateProfileOpts{DryRun: true})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if res.Moved != 1 {
		t.Errorf("dry run should still report 1 planned move, got %d", res.Moved)
	}
	if _, err := os.Stat(filepath.Join(profileMem, "project_b.md")); err != nil {
		t.Error("dry run moved the source file")
	}
	if _, err := os.Stat(filepath.Join(defaultMem, "project_b.md")); err == nil {
		t.Error("dry run wrote into the target store")
	}
}

// TestMigrateProfilesSecondRunIsQuiet — running twice must be safe, because
// the advisory in `moai update` will nag until the user runs it, and some
// will run it more than once.
func TestMigrateProfilesSecondRunIsQuiet(t *testing.T) {
	home, project, _, profileMem := profileFixture(t)
	t.Setenv("HOME", home)

	writeProfileFile(t, profileMem, "project_b.md", "B\n")
	writeProfileFile(t, profileMem, "MEMORY.md", "- [B](project_b.md) — b\n")

	if _, err := migrateProfileMemory(project, migrateProfileOpts{}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	second, err := migrateProfileMemory(project, migrateProfileOpts{})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if second.Moved != 0 || second.Renamed != 0 {
		t.Errorf("second run was not a no-op: %+v", second)
	}
}

// TestMigrateProfilesNoProfilesIsNotAnError covers the common case: a user
// who never used -p has nothing to migrate.
func TestMigrateProfilesNoProfilesIsNotAnError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	res, err := migrateProfileMemory(project, migrateProfileOpts{})
	if err != nil {
		t.Fatalf("no profiles: %v", err)
	}
	if res.Moved != 0 || len(res.Profiles) != 0 {
		t.Errorf("expected an empty result, got %+v", res)
	}
}

// TestDefaultMemoryStore_DoesNotEscapeToRealHome guards the other site pinned
// by TestHomeJoinSiteCountIsPinned (internal/hook). Same shape, same reason:
// the home is read inside the function, so only the environment can isolate
// it, and a test that forgets leaks a permanent directory per run.
//
// Falsifiability: delete the t.Setenv("HOME", tmp) line and this test fails.
func TestDefaultMemoryStore_DoesNotEscapeToRealHome(t *testing.T) {
	// Cannot run in parallel: t.Setenv mutates process-wide state.
	realHome, err := os.UserHomeDir() // captured BEFORE HOME is overridden
	if err != nil {
		t.Fatalf("os.UserHomeDir(): %v", err)
	}

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir, err := defaultMemoryStore(t.TempDir())
	if err != nil {
		t.Fatalf("defaultMemoryStore: %v", err)
	}
	if strings.HasPrefix(dir, realHome+string(os.PathSeparator)) {
		t.Errorf("default store escaped into the real home: %s", dir)
	}
}
