package cli

// t1619_update_migrate_coverage_test.go — card t1619, bundle B2: the unexecuted changed lines
// of update_migrate.go (the deployment_mode write-failure warnings of stampMigrationDeployMode).
// Family: TestMigrat.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// t1619BlockLLMRecord makes the deployment-mode record unwritable: a directory sits where the
// llm.yaml file belongs, so the record can be neither read as a file nor rewritten.
func t1619BlockLLMRecord(t *testing.T, projectRoot string) {
	t.Helper()
	blocked := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	if err := os.MkdirAll(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestMigrateDeployModeWriteFailureIsWarnedOnConfirmed — a confirmed migration whose
// deployment_mode record cannot be written warns instead of stamping silently.
func TestMigrateDeployModeWriteFailureIsWarnedOnConfirmed(t *testing.T) {
	projectRoot := t.TempDir()
	t1619BlockLLMRecord(t, projectRoot)
	var errOut strings.Builder
	stampMigrationDeployMode(projectRoot, migrateConfirmed, &errOut)
	if !strings.Contains(errOut.String(), "deployment_mode write warning") {
		t.Fatalf("a failed plugin-mode record write must be warned, got %q", errOut.String())
	}
}

// TestMigrateDeployModeWriteFailureIsWarnedOnOptedOut — the local stamp of an opted-out
// migration that cannot be written warns the same way.
func TestMigrateDeployModeWriteFailureIsWarnedOnOptedOut(t *testing.T) {
	projectRoot := t.TempDir()
	t1619BlockLLMRecord(t, projectRoot)
	var errOut strings.Builder
	stampMigrationDeployMode(projectRoot, migrateOptedOut, &errOut)
	if !strings.Contains(errOut.String(), "deployment_mode write warning") {
		t.Fatalf("a failed local record write must be warned, got %q", errOut.String())
	}
}
