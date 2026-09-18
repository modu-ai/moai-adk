package migrations_test

// m001_runner_integration_test.go — card t895.
//
// This file exists because of a gap the sibling tests cannot close from either
// side.
//
// The runner's own tests (internal/migration/runner_test.go) all call
// withTestRegistry and replace the registry with touchMigration fakes — fakes
// NAMED "m001" and "m002", which makes the substitution invisible to a reader
// skimming them. So no runner test ever runs a real registered migration. The
// migration's own tests, in turn, call Apply directly and never reach the
// runner. A migration whose Apply returned a non-nil error on success therefore
// satisfied both suites while reporting every successful run to the user as
// "마이그레이션 1 적용 실패: m001 적용 완료" — a sentence that contradicts itself.
//
// The test below is the seam: the REAL registry, through the REAL runner,
// against a project that actually needs migrating. It is deliberately not in
// package migration, which cannot import the migrations package without an
// import cycle.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/migration"
	_ "github.com/modu-ai/moai-adk/internal/migration/migrations"
)

// TestRunner_AppliesRealM001AsSuccess runs the registered migrations over a
// project carrying the hardcoded literal and asserts the runner reports
// success — both in its return and in the version it advances to.
func TestRunner_AppliesRealM001AsSuccess(t *testing.T) {
	root := t.TempDir()

	dir := filepath.Join(root, ".claude", "hooks", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	wrapper := filepath.Join(dir, "handle-stop.sh")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nexec "+hardcodedLiteral+" hook stop\n"), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}

	// Assert the premise: migration 1 must actually be pending here, or the
	// run below would prove nothing about it.
	if m := migration.FindByVersion(1); m == nil {
		t.Fatal("registry carries no migration at version 1; this test would assert nothing")
	}

	applied, err := migration.NewRunner(root).Apply(context.Background())
	if err != nil {
		t.Fatalf("runner reported failure on a project it migrated correctly: %v", err)
	}

	var sawOne bool
	for _, v := range applied {
		if v == 1 {
			sawOne = true
		}
	}
	if !sawOne {
		t.Errorf("applied = %v, want it to include migration 1", applied)
	}

	// The work the runner claims to have done must actually be on disk.
	got, err := os.ReadFile(wrapper)
	if err != nil {
		t.Fatalf("read wrapper: %v", err)
	}
	if strings.Contains(string(got), hardcodedLiteral) {
		t.Errorf("runner reported success but the wrapper still carries the literal: %q", string(got))
	}
	if !strings.Contains(string(got), portableForm) {
		t.Errorf("runner reported success but the wrapper lacks the portable form: %q", string(got))
	}
}

// TestRunner_AppliesRealM001OnCleanProjectAsSuccess covers the other success
// path through the same seam: a project with nothing to rewrite is already
// migrated, and the runner must not call that a failure either.
func TestRunner_AppliesRealM001OnCleanProjectAsSuccess(t *testing.T) {
	root := t.TempDir()

	dir := filepath.Join(root, ".claude", "hooks", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "handle-stop.sh"),
		[]byte("#!/bin/sh\nexec "+portableForm+" hook stop\n"), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}

	if _, err := migration.NewRunner(root).Apply(context.Background()); err != nil {
		t.Fatalf("runner reported failure on an already-clean project: %v", err)
	}
}
