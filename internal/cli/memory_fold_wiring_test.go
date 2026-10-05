package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// pathBeneath reports whether path lies at or under root, comparing cleaned
// absolute paths segment by segment (a prefix string match would treat
// "/home/a" as the parent of "/home/ab").
func pathBeneath(root, path string) bool {
	if root == "" {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// TestMemoryFoldOnDone_ExistingClosePathsContained is the containment cell of
// AC-MFB-008 (xi), plan M0: the memory store resolution that every close path
// will reach must stay inside the TestMain home sandbox and never under the
// developer's real home.
//
// It runs in the TestMain-sandboxed environment exactly as the pre-existing
// close-path tests do: no per-test HOME, USERPROFILE, CLAUDE_CONFIG_DIR or
// MOAI_HOME override. M0 asserts on the candidate list the shared resolver
// (memoryCandidateStores) returns. The gate constant (config.EnvMemoryFoldOnDone,
// M1) and the recorder seam of foldClosedCardMemory (M4) do not exist yet; M4
// extends this cell to set the gate and drive the three close paths.
func TestMemoryFoldOnDone_ExistingClosePathsContained(t *testing.T) {
	if capturedRealHome == "" || homeSandboxDir == "" {
		t.Fatalf("TestMain home sandbox not initialised (real=%q sandbox=%q): the cell is unmeasured", capturedRealHome, homeSandboxDir)
	}

	stores, err := memoryCandidateStores(t.TempDir())
	if err != nil {
		t.Fatalf("memoryCandidateStores: %v", err)
	}
	if len(stores) == 0 {
		t.Fatalf("no candidate store recorded: an empty recording reads as unmeasured, not as contained")
	}

	for _, s := range stores {
		if pathBeneath(capturedRealHome, s.Dir) {
			t.Errorf("candidate store %q (origin %q) lies beneath the real home %q", s.Dir, s.Origin, capturedRealHome)
		}
		if !pathBeneath(homeSandboxDir, s.Dir) {
			t.Errorf("candidate store %q (origin %q) is not beneath the sandbox root %q", s.Dir, s.Origin, homeSandboxDir)
		}
	}
}
