// main_test.go — Unit tests for docs-version-snapshot
// SPEC-DOCS-SITE-001 Phase 5, AC-G3-04
//
// Tests:
//   TestMinorRelease        v2.12.0 -> v2.13.0 creates content/{locale}/v2.12/
//   TestPatchRelease        v2.12.1 -> v2.12.2 no snapshot created
//   TestMajorRelease        v2.12.0 -> v3.0.0 creates content/{locale}/v2/
//   TestSkipExistingVersions v2.11/ already exists is not overwritten
//   TestCopyFidelity        source file count == snapshot file count

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// setupFakeContent creates a temporary content directory with locale sub-dirs
// and the specified number of .md files per locale.
// Returns the content dir path and a cleanup function.
func setupFakeContent(t *testing.T, localeFiles map[string]int) string {
	t.Helper()
	contentDir := t.TempDir()
	for _, locale := range locales {
		count, ok := localeFiles[locale]
		if !ok {
			count = 3
		}
		dir := filepath.Join(contentDir, locale)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		for i := 0; i < count; i++ {
			path := filepath.Join(dir, fmt.Sprintf("page-%02d.md", i))
			if err := os.WriteFile(path, []byte(fmt.Sprintf("# Page %d\ncontent\n", i)), 0o644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
		}
	}
	return contentDir
}

// countFiles counts all .md files under dir (non-recursive skips sub-dirs).
func countFiles(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir %s: %v", dir, err)
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			count++
		}
	}
	return count
}

// runWithLocalCopy runs the snapshot using local filesystem copy (no git).
// This replaces the git-based copy for unit tests.
func runLocalSnapshot(cfg Config) error {
	prev, err := parseSemver(cfg.PreviousVersion)
	if err != nil {
		return fmt.Errorf("parse previous-version: %w", err)
	}
	curr, err := parseSemver(cfg.CurrentVersion)
	if err != nil {
		return fmt.Errorf("parse current-version: %w", err)
	}

	rt, err := classifyRelease(prev, curr)
	if err != nil {
		return err
	}

	if rt == patchRelease {
		return nil // no snapshot
	}

	snapName := snapshotDirName(prev, rt)

	var createdDirs []string

	for _, locale := range locales {
		srcDir := filepath.Join(cfg.ContentDir, locale)
		destDir := filepath.Join(cfg.ContentDir, locale, snapName)

		// Guard: skip if already exists and no --force
		if _, err := os.Stat(destDir); err == nil {
			if !cfg.Force {
				return fmt.Errorf("snapshot directory already exists: %s (use --force to overwrite)", destDir)
			}
			_ = os.RemoveAll(destDir)
		}

		if cfg.DryRun {
			fmt.Printf("[dry-run] would create %s\n", destDir)
			continue
		}

		if err := os.MkdirAll(destDir, 0o755); err != nil {
			rollback(createdDirs)
			return fmt.Errorf("mkdir %s: %w", destDir, err)
		}
		createdDirs = append(createdDirs, destDir)

		// Copy all .md files from srcDir into destDir (skip nested v* dirs)
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			rollback(createdDirs)
			return fmt.Errorf("readdir %s: %w", srcDir, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue // skip versioned sub-dirs
			}
			src := filepath.Join(srcDir, e.Name())
			dst := filepath.Join(destDir, e.Name())
			if err := copyFile(src, dst); err != nil {
				rollback(createdDirs)
				return fmt.Errorf("copy %s: %w", src, err)
			}
		}
	}
	return nil
}

// TestMinorRelease verifies that v2.12.0 -> v2.13.0 creates content/{locale}/v2.12/
func TestMinorRelease(t *testing.T) {
	contentDir := setupFakeContent(t, map[string]int{"ko": 3, "en": 3, "ja": 3, "zh": 3})

	cfg := Config{
		CurrentVersion:  "v2.13.0",
		PreviousVersion: "v2.12.0",
		ContentDir:      contentDir,
		DryRun:          false,
	}

	if err := runLocalSnapshot(cfg); err != nil {
		t.Fatalf("runLocalSnapshot: %v", err)
	}

	// Verify content/{locale}/v2.12/ directories were created
	for _, locale := range locales {
		snapDir := filepath.Join(contentDir, locale, "v2.12")
		if _, err := os.Stat(snapDir); os.IsNotExist(err) {
			t.Errorf("expected snapshot dir %s to exist", snapDir)
		}
		got := countFiles(t, snapDir)
		if got == 0 {
			t.Errorf("locale %s: snapshot dir is empty", locale)
		}
	}
}

// TestPatchRelease verifies that v2.12.1 -> v2.12.2 produces no snapshot.
func TestPatchRelease(t *testing.T) {
	contentDir := setupFakeContent(t, nil)

	cfg := Config{
		CurrentVersion:  "v2.12.2",
		PreviousVersion: "v2.12.1",
		ContentDir:      contentDir,
		DryRun:          false,
	}

	if err := runLocalSnapshot(cfg); err != nil {
		t.Fatalf("runLocalSnapshot: %v", err)
	}

	// No v2.12 or v2.12.1 directories should be created
	for _, locale := range locales {
		snapDir := filepath.Join(contentDir, locale, "v2.12")
		if _, err := os.Stat(snapDir); err == nil {
			t.Errorf("expected no snapshot dir for patch release, but %s exists", snapDir)
		}
	}
}

// TestMajorRelease verifies that v2.12.0 -> v3.0.0 creates content/{locale}/v2.12/.
// (Major bump: snapshot the previous minor, dir name is "v2.12" not "v2")
func TestMajorRelease(t *testing.T) {
	contentDir := setupFakeContent(t, nil)

	cfg := Config{
		CurrentVersion:  "v3.0.0",
		PreviousVersion: "v2.12.0",
		ContentDir:      contentDir,
		DryRun:          false,
	}

	if err := runLocalSnapshot(cfg); err != nil {
		t.Fatalf("runLocalSnapshot: %v", err)
	}

	// For major release, previous v2.12.0 -> snap dir is "v2.12"
	expectedSnap := "v2.12"
	for _, locale := range locales {
		snapDir := filepath.Join(contentDir, locale, expectedSnap)
		if _, err := os.Stat(snapDir); os.IsNotExist(err) {
			t.Errorf("expected snapshot dir %s to exist for major release", snapDir)
		}
	}
}

// TestSkipExistingVersions verifies that an existing v2.11/ directory is not overwritten.
func TestSkipExistingVersions(t *testing.T) {
	contentDir := setupFakeContent(t, nil)

	// Pre-create v2.11/ with a sentinel file
	for _, locale := range locales {
		dir := filepath.Join(contentDir, locale, "v2.11")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		sentinel := filepath.Join(dir, "sentinel.md")
		if err := os.WriteFile(sentinel, []byte("original\n"), 0o644); err != nil {
			t.Fatalf("write sentinel: %v", err)
		}
	}

	// v2.12.0 -> v2.13.0 should create v2.12, not touch v2.11
	cfg := Config{
		CurrentVersion:  "v2.13.0",
		PreviousVersion: "v2.12.0",
		ContentDir:      contentDir,
		DryRun:          false,
	}

	if err := runLocalSnapshot(cfg); err != nil {
		t.Fatalf("runLocalSnapshot: %v", err)
	}

	// v2.11/ sentinel must still exist unmodified
	for _, locale := range locales {
		sentinel := filepath.Join(contentDir, locale, "v2.11", "sentinel.md")
		data, err := os.ReadFile(sentinel)
		if err != nil {
			t.Errorf("sentinel %s missing: %v", sentinel, err)
			continue
		}
		if string(data) != "original\n" {
			t.Errorf("sentinel %s was modified (got %q)", sentinel, string(data))
		}
	}

	// v2.12/ should also have been created
	for _, locale := range locales {
		snapDir := filepath.Join(contentDir, locale, "v2.12")
		if _, err := os.Stat(snapDir); os.IsNotExist(err) {
			t.Errorf("expected v2.12 snapshot dir for locale %s", locale)
		}
	}
}

// TestCopyFidelity verifies that the number of files in the snapshot equals
// the number of files in the original locale directory.
func TestCopyFidelity(t *testing.T) {
	// Use 5 files per locale
	contentDir := setupFakeContent(t, map[string]int{"ko": 5, "en": 5, "ja": 5, "zh": 5})

	cfg := Config{
		CurrentVersion:  "v2.13.0",
		PreviousVersion: "v2.12.0",
		ContentDir:      contentDir,
		DryRun:          false,
	}

	if err := runLocalSnapshot(cfg); err != nil {
		t.Fatalf("runLocalSnapshot: %v", err)
	}

	for _, locale := range locales {
		srcDir := filepath.Join(contentDir, locale)
		snapDir := filepath.Join(contentDir, locale, "v2.12")

		srcCount := countFiles(t, srcDir)
		snapCount := countFiles(t, snapDir)

		if srcCount != snapCount {
			t.Errorf("locale %s: source has %d files but snapshot has %d files",
				locale, srcCount, snapCount)
		}
	}
}

// TestClassifyRelease verifies the version classification logic.
func TestClassifyRelease(t *testing.T) {
	tests := []struct {
		prev    string
		curr    string
		want    releaseType
		wantErr bool
	}{
		{"v2.12.0", "v2.13.0", minorRelease, false},
		{"v2.12.1", "v2.12.2", patchRelease, false},
		{"v2.12.0", "v3.0.0", majorRelease, false},
		{"v2.0.0", "v2.1.0", minorRelease, false},
		{"v1.5.0", "v2.0.0", majorRelease, false},
		{"v2.12.0", "v2.12.0", patchRelease, false},   // same version -> patch
		{"v2.13.0", "v2.12.0", patchRelease, true},    // downgrade -> error
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%s", tt.prev, tt.curr), func(t *testing.T) {
			prev, _ := parseSemver(tt.prev)
			curr, _ := parseSemver(tt.curr)
			got, err := classifyRelease(prev, curr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("classifyRelease(%s, %s) = %d, want %d", tt.prev, tt.curr, got, tt.want)
			}
		})
	}
}

// TestSnapshotDirName verifies directory name calculation.
func TestSnapshotDirName(t *testing.T) {
	tests := []struct {
		prev string
		rt   releaseType
		want string
	}{
		{"v2.12.0", minorRelease, "v2.12"},
		{"v2.12.4", minorRelease, "v2.12"},
		// Major release: snapshot dir is still MAJOR.MINOR of previous version (AC-G3-04)
		{"v2.12.0", majorRelease, "v2.12"},
		{"v1.5.3", majorRelease, "v1.5"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", tt.prev, tt.rt), func(t *testing.T) {
			prev, _ := parseSemver(tt.prev)
			got := snapshotDirName(prev, tt.rt)
			if got != tt.want {
				t.Errorf("snapshotDirName(%s, %d) = %q, want %q", tt.prev, tt.rt, got, tt.want)
			}
		})
	}
}
