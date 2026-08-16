package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
)

// plantedMoaiManagedPaths is the literal, non-empty set of moai-managed paths
// the recovery-manifest fixture plants before the destructive step runs. Every
// entry is a path deploy.CleanMoaiManagedPaths actually removes.
//
// SPEC-UPDATE-DATA-SURVIVAL-001 AC-UDS-001 command (b) reads this declaration
// from the test source and requires it to be non-empty. The count MUST come
// from this literal and NOT from observing CleanMoaiManagedPaths' own output —
// a count derived from the function under test compares the function against
// itself (acceptance.md §A.3 shape (a)), a self-comparison that can never fail.
var plantedMoaiManagedPaths = []string{
	".claude/settings.json",
	".claude/commands/moai/plan.md",
	".claude/agents/moai/manager-develop.md",
	".claude/rules/moai/core/moai-constitution.md",
	".claude/output-styles/moai/moai.md",
	".claude/hooks/moai/handle-session-start.sh",
}

// writeFixtureFile creates parent directories and writes content at
// projectRoot/rel.
func writeFixtureFile(t *testing.T, projectRoot, rel, content string) {
	t.Helper()
	abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func fixturePathExists(projectRoot, rel string) bool {
	_, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(rel)))
	return err == nil
}

// TestUpdateFailure_WritesRecoveryManifest verifies AC-UDS-001 (REQ-UDS-019,
// REQ-UDS-020): a stage failure after the first destructive step writes and
// prints a recovery manifest, and does NOT roll the tree back.
func TestUpdateFailure_WritesRecoveryManifest(t *testing.T) {
	if len(plantedMoaiManagedPaths) == 0 {
		t.Fatal("plantedMoaiManagedPaths must be non-empty: an empty planted set " +
			"makes the no-rollback assertion vacuous (AC-UDS-001 clause 3)")
	}

	projectRoot := t.TempDir()

	// The config directory is the backup subject; plant it so BackupMoaiConfig
	// produces a real run-scoped backup directory to host the manifest.
	writeFixtureFile(t, projectRoot, ".moai/config/sections/system.yaml",
		"moai:\n  version: 3.0.0\n")

	for _, rel := range plantedMoaiManagedPaths {
		writeFixtureFile(t, projectRoot, rel, "planted-"+rel+"\n")
	}

	backupDir, err := backup.BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	if backupDir == "" {
		t.Fatal("BackupMoaiConfig returned an empty backup dir; fixture is wrong")
	}

	var printed bytes.Buffer
	guard := newRecoveryGuard(projectRoot, backupDir, &printed)

	injected := errors.New("injected deploy failure")

	stages := []updateStage{
		{
			name:        cleanManagedPathsStage,
			destructive: true,
			run: func() error {
				// Clause 3 (pre): every planted path exists immediately before
				// the destructive step runs.
				for _, rel := range plantedMoaiManagedPaths {
					if !fixturePathExists(projectRoot, rel) {
						t.Errorf("clause 3 pre: planted path %s absent before CleanMoaiManagedPaths", rel)
					}
				}
				if cleanErr := deploy.CleanMoaiManagedPaths(projectRoot, io.Discard, fstest.MapFS{}); cleanErr != nil {
					return cleanErr
				}
				// Clause 3 (post): every planted path is gone immediately after
				// the destructive step returns, so the removed set is non-empty
				// by construction.
				for _, rel := range plantedMoaiManagedPaths {
					if fixturePathExists(projectRoot, rel) {
						t.Errorf("clause 3 post: planted path %s survived CleanMoaiManagedPaths", rel)
					}
				}
				return nil
			},
		},
		{
			name: "Deploy Templates",
			run: func() error {
				return injected
			},
		},
	}

	err = runUpdateStages(guard, stages)
	if err == nil {
		t.Fatal("runUpdateStages returned nil; the injected stage failure must propagate")
	}
	if !errors.Is(err, injected) {
		t.Errorf("returned error does not wrap the injected cause: %v", err)
	}

	// Clause 4 (REQ-UDS-020): no automatic rollback — every destroyed path is
	// still absent when the outer call returns.
	for _, rel := range plantedMoaiManagedPaths {
		if fixturePathExists(projectRoot, rel) {
			t.Errorf("clause 4: %s reappeared after the failing call returned "+
				"(automatic rollback is prohibited by REQ-UDS-020)", rel)
		}
	}

	// Clause 1: the manifest exists inside the run-scoped backup directory and
	// names the failed step and the restore command.
	manifestPath := filepath.Join(backupDir, recoveryManifestFileName)
	raw, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		t.Fatalf("clause 1: recovery manifest not found at %s: %v", manifestPath, readErr)
	}
	manifest := string(raw)
	for _, want := range []string{"Deploy Templates", injected.Error(), backupDir, "--restore"} {
		if !strings.Contains(manifest, want) {
			t.Errorf("clause 1: manifest does not mention %q\n--- manifest ---\n%s", want, manifest)
		}
	}

	// Clause 2: the same manifest text was printed to the captured writer.
	if !strings.Contains(printed.String(), manifest) {
		t.Errorf("clause 2: manifest text was not printed\n--- printed ---\n%s\n--- manifest ---\n%s",
			printed.String(), manifest)
	}
}

// TestRecoveryGuard_SilentBeforeDestructiveRegion verifies that a failure
// BEFORE the first destructive step writes no manifest: the tree is still
// intact, so there is nothing to recover.
func TestRecoveryGuard_SilentBeforeDestructiveRegion(t *testing.T) {
	projectRoot := t.TempDir()
	backupDir := filepath.Join(projectRoot, "backup")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("mkdir backup: %v", err)
	}

	var printed bytes.Buffer
	guard := newRecoveryGuard(projectRoot, backupDir, &printed)

	injected := errors.New("validation failed")
	err := runUpdateStages(guard, []updateStage{
		{name: "Validate Templates", run: func() error { return injected }},
	})
	if !errors.Is(err, injected) {
		t.Fatalf("expected the injected error, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(backupDir, recoveryManifestFileName)); statErr == nil {
		t.Error("a recovery manifest was written for a pre-destructive failure")
	}
	if printed.Len() != 0 {
		t.Errorf("pre-destructive failure printed output: %q", printed.String())
	}
}
