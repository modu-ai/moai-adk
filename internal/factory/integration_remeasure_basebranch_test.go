package factory

// integration_remeasure_basebranch_test.go — card t1790, GH #1817: an empty
// integration branch never reaches git. Before the guard the empty base made
// `git merge-base refs/heads/ HEAD` die with exit status 128 AFTER the
// measured command had already run.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRemeasureRefusesEmptyBaseBeforeRunningTheCommand(t *testing.T) {
	tree := t.TempDir()
	marker := filepath.Join(t.TempDir(), "ran")
	for _, base := range []string{"", "   ", "\t"} {
		_, err := RunRemeasure(t.TempDir(), tree, base, "touch "+marker)
		if err == nil {
			t.Fatalf("base %q: an empty integration branch was accepted", base)
		}
		if !strings.Contains(err.Error(), "no integration branch") {
			t.Fatalf("base %q: the refusal must name the missing branch, got %v", base, err)
		}
		if strings.Contains(err.Error(), "merge-base") {
			t.Fatalf("base %q: git was called with the empty ref: %v", base, err)
		}
		if _, statErr := os.Stat(marker); statErr == nil {
			t.Fatalf("base %q: the measured command ran although the base was empty", base)
		}
	}
}
