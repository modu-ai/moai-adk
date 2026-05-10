// Package golden provides the cumulative golden snapshot index for internal/tui.
// This file enumerates all expected golden files from M1-M7 milestones and
// verifies that the count and MANIFEST.txt are kept in sync.
//
// Usage:
//
//	go test ./internal/tui/golden/... -v                    # verify manifest
//	UPDATE_GOLDEN=1 go test ./internal/tui/golden/...       # regenerate manifest
package golden_test

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const (
	testdataDir  = "../testdata"
	manifestFile = "../testdata/MANIFEST.txt"
	// minGoldenFiles is the floor count from plan.md §8.3 DOD: 100+ files.
	minGoldenFiles = 100
)

// TestGoldenCount verifies the number of golden snapshot files meets the DOD floor.
func TestGoldenCount(t *testing.T) {
	files, err := goldenFiles()
	if err != nil {
		t.Fatalf("list golden files: %v", err)
	}
	if len(files) < minGoldenFiles {
		t.Errorf("golden snapshot count = %d, want >= %d (plan.md §8.3 DOD)", len(files), minGoldenFiles)
	}
	t.Logf("golden snapshot count: %d", len(files))
}

// TestManifest verifies that MANIFEST.txt matches the current golden files on disk.
// Run with UPDATE_GOLDEN=1 to regenerate the manifest.
func TestManifest(t *testing.T) {
	files, err := goldenFiles()
	if err != nil {
		t.Fatalf("list golden files: %v", err)
	}

	// Build current checksum map.
	current, err := buildChecksumMap(files)
	if err != nil {
		t.Fatalf("build checksum map: %v", err)
	}

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := writeManifest(current); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
		t.Logf("manifest regenerated: %s (%d entries)", manifestFile, len(current))
		return
	}

	// Read existing manifest.
	recorded, err := readManifest()
	if err != nil {
		t.Fatalf("read manifest: %v (run with UPDATE_GOLDEN=1 to generate)", err)
	}

	// Compare.
	for path, checksum := range current {
		if rec, ok := recorded[path]; !ok {
			t.Errorf("golden file %s is new (not in manifest); run UPDATE_GOLDEN=1", path)
		} else if rec != checksum {
			t.Errorf("golden file %s checksum changed: manifest=%s current=%s", path, rec[:8], checksum[:8])
		}
	}
	for path := range recorded {
		if _, ok := current[path]; !ok {
			t.Errorf("manifest entry %s is missing on disk; run UPDATE_GOLDEN=1", path)
		}
	}
}

// goldenFiles returns sorted list of all *.golden files in testdata/.
func goldenFiles() ([]string, error) {
	var files []string
	err := filepath.WalkDir(testdataDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".golden") {
			rel, _ := filepath.Rel(testdataDir, p)
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// buildChecksumMap computes sha256 for each file relative to testdataDir.
func buildChecksumMap(files []string) (map[string]string, error) {
	out := make(map[string]string, len(files))
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(testdataDir, rel))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		out[rel] = fmt.Sprintf("%x", sum)
	}
	return out, nil
}

// writeManifest serialises the checksum map to MANIFEST.txt.
func writeManifest(m map[string]string) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	f, err := os.Create(manifestFile)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	for _, k := range keys {
		_, _ = fmt.Fprintf(w, "%s  %s\n", m[k], k)
	}
	return w.Flush()
}

// readManifest parses MANIFEST.txt into a checksum map.
func readManifest() (map[string]string, error) {
	f, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	out := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			continue
		}
		out[strings.TrimSpace(parts[1])] = strings.TrimSpace(parts[0])
	}
	return out, s.Err()
}
