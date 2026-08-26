package kanban

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoBareJoinBacklogPathSurvives is the guard that would have caught the file
// the first pass missed.
//
// The queue's path contract lives in BacklogPathForRoot, but it had been spelled
// by hand as `filepath.Join(<root>, "backlog.json")` in several places — the
// implementation, and test files that therefore agreed with the implementation's
// error. Correcting the implementation alone left one test still asserting the
// wrong convention, and CI found it rather than this package's own suite did.
// A convention held in prose across files drifts; this holds it mechanically.
//
// What is forbidden is a bare join against a QUEUE ROOT. Joining against a
// directory that is already `.moai/state/kanban` is a different thing and stays
// legal, so the pattern keys on the receiver's name rather than on the filename.
func TestNoBareJoinBacklogPathSurvives(t *testing.T) {
	// Package directories relative to this one, so the test is independent of
	// where the module is checked out.
	pkgs := []string{".", "../cli", "../web", "../statusline"}
	bad := regexp.MustCompile(`Join\((?:root|fallbackRoot|queueRoot|base|projectRoot)\s*,\s*"backlog\.json"\)`)

	for _, pkg := range pkgs {
		if _, err := os.Stat(pkg); err != nil {
			continue // a sibling package that does not exist is not this test's failure
		}
		err := filepath.WalkDir(pkg, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
				return nil
			}
			b, readErr := os.ReadFile(p)
			if readErr != nil {
				return nil
			}
			for i, line := range strings.Split(string(b), "\n") {
				if strings.Contains(line, "bare-join-intentional") {
					continue // a line that deliberately names the old path to assert it is empty
				}
				if bad.MatchString(line) {
					t.Errorf("%s:%d builds the queue path by hand: %s\n\tuse BacklogPathForRoot(root) — one contract, one spelling",
						p, i+1, strings.TrimSpace(line))
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", pkg, err)
		}
	}
}
