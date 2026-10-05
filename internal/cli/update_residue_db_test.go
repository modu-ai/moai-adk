package cli

// Card t1139 defect 2 — the v3 residue sweep must not remove .moai/db.
//
// homestate.ProjectDir places a project's live home-state database INSIDE the
// project at .moai/db/<key> whenever the project sits under a temp root and
// MOAI_HOME is not explicitly set. .moai/db is also a defs.DeprecatedPaths
// entry, so a non-force `moai update` backed it up and deleted it, destroying
// the live project.json. The sweep now leaves .moai/db alone entirely.
//
// t.Setenv forbids t.Parallel here.

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestResidueCleanup_KeepsLiveHomestateDB(t *testing.T) {
	// A non-explicit MOAI_HOME is the precondition for the in-tree placement.
	t.Setenv("MOAI_HOME", "")

	root := newV3ConfirmedProject(t, "3.0.1")

	liveDir, err := homestate.ProjectDir(root)
	if err != nil {
		t.Fatalf("homestate.ProjectDir: %v", err)
	}
	// Premise: t.TempDir() is under a temp root, so the live db is in-tree.
	// Without this the test would pass without exercising the defect.
	canonicalRoot := homestate.CanonicalProjectRoot(root)
	if !strings.HasPrefix(liveDir, filepath.Join(canonicalRoot, ".moai", "db")+string(filepath.Separator)) {
		t.Fatalf("premise: live homestate dir %s is not under %s/.moai/db", liveDir, canonicalRoot)
	}

	liveFile := filepath.Join(liveDir, "project.json")
	if err := os.MkdirAll(liveDir, 0o755); err != nil {
		t.Fatalf("mkdir live dir: %v", err)
	}
	if err := os.WriteFile(liveFile, []byte(`{"root":"x"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write project.json: %v", err)
	}
	// Positive control for the sweep itself: another deprecated path in the
	// same run must still be removed.
	residueWriteFile(t, root, residueDeprecatedRel, "legacy planner\n")

	var out bytes.Buffer
	result, err := runV3ResidueCleanup(root, false, false, &out)
	if err != nil {
		t.Fatalf("runV3ResidueCleanup: %v\noutput: %s", err, out.String())
	}

	if _, statErr := os.Stat(liveFile); statErr != nil {
		t.Errorf("live homestate project.json was removed by the residue sweep: %v\noutput: %s", statErr, out.String())
	}
	if slices.Contains(result.Removed, ".moai/db") {
		t.Errorf("result.Removed = %v, must not contain .moai/db", result.Removed)
	}
	if !slices.Contains(result.Removed, residueDeprecatedRel) {
		t.Errorf("result.Removed = %v, want it to contain %s (the sweep must still run)", result.Removed, residueDeprecatedRel)
	}
}

// TestResidueCleanup_LeavesDotMoaiDBEntirely pins the scope decision: the
// residue sweep excludes .moai/db as a whole, including content that is not
// the live homestate directory, and a tree whose only residue is .moai/db
// reports nothing removed.
func TestResidueCleanup_LeavesDotMoaiDBEntirely(t *testing.T) {
	root := newV3ConfirmedProject(t, "3.0.1")
	legacy := residueWriteFile(t, root, ".moai/db/schema.md", "legacy\n")

	var out bytes.Buffer
	result, err := runV3ResidueCleanup(root, false, false, &out)
	if err != nil {
		t.Fatalf("runV3ResidueCleanup: %v", err)
	}
	if _, statErr := os.Stat(legacy); statErr != nil {
		t.Errorf(".moai/db content removed: %v", statErr)
	}
	if len(result.Removed) != 0 || result.BackupDir != "" {
		t.Errorf("result = %+v, want nothing removed and no backup", result)
	}
}
