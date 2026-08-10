// Package cli — migrate_agency_idempotent_test.go
//
// Regression guards for the `moai migrate agency` re-run failure (issue #1414).
//
// checkTargetsAbsent was a bare existence test, so re-running the command on an
// already-migrated project failed with MIGRATE_TARGET_EXISTS and advised
// `--force`. Following that advice overwrote a v3-schema design.yaml with
// v2-derived content, destroying data. The classifier below distinguishes
// "already migrated" (skip, no error) from "unrecognised content" (still an
// error), and the error message no longer recommends --force as the remedy.

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedMigratedTargets writes the v3-shaped migration outputs into dir, standing
// in for a project that has already been migrated once.
func seedMigratedTargets(t *testing.T, dir string) {
	t.Helper()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	// v3 shape: parses as YAML with a top-level `design:` key.
	if err := os.WriteFile(filepath.Join(sections, "design.yaml"),
		[]byte("design:\n    gan_loop:\n        max_iterations: 5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// v3 shape: the observations target exists as a directory.
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research", "observations"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestMigrateAgency_IdempotentOnAlreadyMigrated is the issue #1414 guard: a
// second `moai migrate agency` on an already-migrated project must succeed
// WITHOUT --force, and must leave the existing v3 design.yaml untouched rather
// than overwriting it with v2-derived content.
func TestMigrateAgency_IdempotentOnAlreadyMigrated(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)
	seedMigratedTargets(t, dir)

	designYAML := filepath.Join(dir, ".moai", "config", "sections", "design.yaml")
	before, err := os.ReadFile(designYAML)
	if err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{projectRoot: dir, homeDir: dir}
	if _, runErr := m.Run(); runErr != nil {
		t.Fatalf("Run() on an already-migrated project returned error: %v\n"+
			"A re-run must be a no-op for already-migrated targets, not a "+
			"MIGRATE_TARGET_EXISTS failure that pushes the user toward --force", runErr)
	}

	after, err := os.ReadFile(designYAML)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("already-migrated design.yaml was rewritten by the re-run\nbefore=%q\nafter=%q",
			string(before), string(after))
	}
}

// TestMigrateAgency_UnrecognisedTargetStillErrors asserts the classifier did not
// become permissive: a target holding content the migrator cannot recognise as
// v3 must still refuse to proceed.
func TestMigrateAgency_UnrecognisedTargetStillErrors(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	// Parses as YAML but carries no top-level `design:` key — not v3.
	if err := os.WriteFile(filepath.Join(sections, "design.yaml"),
		[]byte("agency:\n    gan_loop:\n        max_iterations: 5\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{projectRoot: dir, homeDir: dir}
	_, err := m.Run()
	if err == nil {
		t.Fatal("expected an error when a target holds unrecognised (non-v3) content")
	}
	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T (%v)", err, err)
	}
	if me.Code != ErrMigrateTargetExists {
		t.Errorf("expected code %s, got %s", ErrMigrateTargetExists, me.Code)
	}
}

// TestMigrateAgency_TargetExistsMessageDoesNotPushForce asserts the harmful
// hint is gone: the message must not present `--force` as the remedy, because
// following that advice overwrites a v3 design.yaml with v2-derived content.
func TestMigrateAgency_TargetExistsMessageDoesNotPushForce(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "design.yaml"),
		[]byte(":\n  not: valid: yaml:\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{projectRoot: dir, homeDir: dir}
	_, err := m.Run()
	if err == nil {
		t.Fatal("expected an error when a target holds unparseable content")
	}
	msg := err.Error()
	if strings.Contains(msg, "use --force to overwrite") {
		t.Errorf("error still recommends --force as the remedy: %q\n"+
			"Following that advice overwrites the existing file with v2-derived "+
			"content; the message must recommend inspecting and moving it instead", msg)
	}
	for _, want := range []string{"inspect", "--force"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message missing %q; got %q", want, msg)
		}
	}
}
