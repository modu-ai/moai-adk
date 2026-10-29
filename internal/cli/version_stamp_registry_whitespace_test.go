package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

// TestVersionStampPopulationKeepsWhitespaceNames pins card t1987 audit P2: the
// tracked-file listing must hand the sweep the raw names git reports. A name
// with a leading or trailing space is a different file from its trimmed form,
// so trimming it both lets "go.mod " slip into the go.mod exclusion and makes
// "keep.txt " read a path that does not exist. A synthetic token (v9.9.9) is
// used so the test does not depend on the real authoritative version.
func TestVersionStampPopulationKeepsWhitespaceNames(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file names with leading or trailing spaces cannot exist on windows")
	}

	root := t.TempDir()
	initGitRepo(t, root)

	const token = "v9.9.9"
	whitespaceNames := []string{"go.mod ", " go.mod", "go.sum ", " go.sum", "keep.txt "}
	excludedNames := []string{"go.mod", "go.sum"}

	names := append(append([]string{}, excludedNames...), whitespaceNames...)
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(root, name), []byte("carries "+token+"\n"), 0o644); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}
	addArgs := append([]string{"add", "--"}, names...)
	add := exec.Command("git", addArgs...)
	add.Dir = root
	if out, err := add.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	population := trackedPopulation(t, root)
	contents := readPopulationContents(t, root, population)
	swept, judged := versionStampSweep(token, population, contents)

	for _, name := range whitespaceNames {
		if !slices.Contains(population, name) {
			t.Errorf("population dropped or altered %q (got %q)", name, population)
		}
		if !slices.Contains(swept, name) {
			t.Errorf("sweep did not report %q carrying the token (got %q)", name, swept)
		}
	}
	for _, name := range excludedNames {
		if slices.Contains(population, name) {
			t.Errorf("population kept excluded name %q", name)
		}
	}
	if judged != len(population) {
		t.Errorf("judged %d of %d population paths; some were handed but unexamined (population %q)",
			judged, len(population), population)
	}
}
