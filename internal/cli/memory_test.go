// memory_test.go — `moai memory` doctor/archive.
//
// The archive path is the one that moves files, so it carries the heavier
// tests: all-or-nothing validation, never-delete, and index maintenance.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedMemoryDir lays out a store: an index linking each name in linked, and a
// topic file per name in present.
func seedMemoryDir(t *testing.T, linked, present []string) string {
	t.Helper()
	dir := t.TempDir()

	var idx strings.Builder
	idx.WriteString("# Memory Index\n\n")
	for _, n := range linked {
		idx.WriteString("- [T](" + n + ") — hook\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(idx.String()), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	for _, n := range present {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("---\n---\nbody\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", n, err)
		}
	}
	return dir
}

func TestArchiveMovesFileAndDropsIndexLine(t *testing.T) {
	t.Parallel()
	dir := seedMemoryDir(t,
		[]string{"feedback_a.md", "project_b.md"},
		[]string{"feedback_a.md", "project_b.md"})

	var out bytes.Buffer
	if err := archiveMemoryFiles(&out, dir, []string{"project_b.md"}); err != nil {
		t.Fatalf("archive: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "project_b.md")); !os.IsNotExist(err) {
		t.Error("archived file still sits in the store root")
	}
	if _, err := os.Stat(filepath.Join(dir, "_archive", "project_b.md")); err != nil {
		t.Errorf("archived file is not under _archive: %v", err)
	}

	idx, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if strings.Contains(string(idx), "project_b.md") {
		t.Errorf("index still links the archived file:\n%s", idx)
	}
	if !strings.Contains(string(idx), "feedback_a.md") {
		t.Errorf("index lost an unrelated entry:\n%s", idx)
	}
}

// TestArchiveNeverDeletes is the constitutional guarantee: archiving keeps
// the audit trail, so the bytes must survive the move.
func TestArchiveNeverDeletes(t *testing.T) {
	t.Parallel()
	dir := seedMemoryDir(t, []string{"feedback_a.md"}, []string{"feedback_a.md"})
	original, err := os.ReadFile(filepath.Join(dir, "feedback_a.md"))
	if err != nil {
		t.Fatalf("read original: %v", err)
	}

	var out bytes.Buffer
	if err := archiveMemoryFiles(&out, dir, []string{"feedback_a.md"}); err != nil {
		t.Fatalf("archive: %v", err)
	}

	moved, err := os.ReadFile(filepath.Join(dir, "_archive", "feedback_a.md"))
	if err != nil {
		t.Fatalf("read archived: %v", err)
	}
	if !bytes.Equal(original, moved) {
		t.Error("archived content differs from the original")
	}
}

// TestArchiveValidatesEveryNameBeforeMoving pins the all-or-nothing contract:
// a typo in the last name must not leave the earlier ones half-archived.
func TestArchiveValidatesEveryNameBeforeMoving(t *testing.T) {
	t.Parallel()
	dir := seedMemoryDir(t,
		[]string{"feedback_a.md", "project_b.md"},
		[]string{"feedback_a.md", "project_b.md"})

	var out bytes.Buffer
	err := archiveMemoryFiles(&out, dir, []string{"feedback_a.md", "does_not_exist.md"})
	if err == nil {
		t.Fatal("archive accepted a missing name")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "feedback_a.md")); statErr != nil {
		t.Error("a valid name was moved despite the batch failing")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "_archive")); statErr == nil {
		t.Error("_archive was created for a batch that never ran")
	}
}

// TestArchiveRefusesTheIndexItself guards the obvious foot-gun.
func TestArchiveRefusesTheIndexItself(t *testing.T) {
	t.Parallel()
	dir := seedMemoryDir(t, []string{"feedback_a.md"}, []string{"feedback_a.md"})

	var out bytes.Buffer
	if err := archiveMemoryFiles(&out, dir, []string{"MEMORY.md"}); err == nil {
		t.Error("archive accepted MEMORY.md")
	}
	if _, err := os.Stat(filepath.Join(dir, "MEMORY.md")); err != nil {
		t.Errorf("index was moved: %v", err)
	}
}

// TestArchiveOrphanNeedsNoIndexLine covers the common case: the orphans
// doctor reports have no index line to drop, which is not an error.
func TestArchiveOrphanNeedsNoIndexLine(t *testing.T) {
	t.Parallel()
	dir := seedMemoryDir(t, []string{"feedback_a.md"},
		[]string{"feedback_a.md", "project_orphan.md"})

	var out bytes.Buffer
	if err := archiveMemoryFiles(&out, dir, []string{"project_orphan.md"}); err != nil {
		t.Fatalf("archiving an orphan failed: %v", err)
	}
	if !strings.Contains(out.String(), "0 index line(s) dropped") {
		t.Errorf("expected a zero-index-line report, got: %q", out.String())
	}
}

// TestDoctorReportsBothStores pins the two-store reality: a profile session
// and a plain session write to different directories, and a report that
// showed only one would hide half the store.
func TestDoctorReportsBothStores(t *testing.T) {
	home := t.TempDir()
	profile := filepath.Join(home, "profile")
	project := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_PROJECT_DIR", project)
	t.Setenv("CLAUDE_CONFIG_DIR", profile)

	slug := memoryProjectSlug(project)
	for _, root := range []string{
		filepath.Join(profile, "projects", slug, "memory"),
		filepath.Join(home, ".claude", "projects", slug, "memory"),
	} {
		if err := os.MkdirAll(root, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", root, err)
		}
		if err := os.WriteFile(filepath.Join(root, "feedback_x.md"), []byte("---\n---\n"), 0o644); err != nil {
			t.Fatalf("seed %s: %v", root, err)
		}
	}

	reports, err := collectMemoryReports("", 0)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("resolved %d store(s), want 2: %+v", len(reports), reports)
	}
	for _, r := range reports {
		if !r.Exists || r.TopicFiles != 1 {
			t.Errorf("store %s: exists=%v files=%d, want true/1", r.Store.Dir, r.Exists, r.TopicFiles)
		}
		// The seeded file is unindexed, so each store must report it.
		if len(r.Findings) == 0 {
			t.Errorf("store %s reported no orphan for an unindexed file", r.Store.Dir)
		}
	}
}
