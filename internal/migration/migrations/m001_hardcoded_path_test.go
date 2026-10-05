package migrations_test

// The six tests below were authored as RED placeholders and left skipped with
// the reason "waiting for m001 implementation". The implementation landed, the
// reason went stale, and the skips kept the coverage gap open silently: m001
// returned a non-nil error on BOTH of its success paths, and nothing ran to
// notice. Card t895 revives them against the shipped implementation.
//
// They exercise m001 through the registry rather than by calling the unexported
// m001Apply directly, so they cover the same entry point the runner uses. All
// filesystem work happens under t.TempDir(); the real home is never touched.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/migration"
	_ "github.com/modu-ai/moai-adk/internal/migration/migrations"
)

const (
	hardcodedLiteral = "/Users/goos/go/bin/moai"
	portableForm     = "$HOME/go/bin/moai"
)

// m001 returns the registered version-1 migration. It fails loudly when the
// registry does not carry it, because an absent registration would make every
// assertion in this file vacuous.
func m001(t *testing.T) migration.Migration {
	t.Helper()
	m := migration.FindByVersion(1)
	if m == nil {
		t.Fatal("registry carries no migration at version 1; these tests would assert nothing")
	}
	if m.Apply == nil {
		t.Fatal("migration 1 registers a nil Apply; these tests would assert nothing")
	}
	return *m
}

// writeWrapper creates .claude/hooks/moai/handle-<name>.sh under root with the
// given body and mode, and returns its path.
func writeWrapper(t *testing.T, root, name, body string, mode os.FileMode) string {
	t.Helper()
	dir := filepath.Join(root, ".claude", "hooks", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	p := filepath.Join(dir, "handle-"+name+".sh")
	if err := os.WriteFile(p, []byte(body), mode); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	// os.WriteFile honours the mode only on create; set it explicitly so the
	// executable-bit test measures what it means to.
	if err := os.Chmod(p, mode); err != nil {
		t.Fatalf("chmod %s: %v", p, err)
	}
	return p
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

// TestM001_RewritesHardcodedLiteral verifies that m001 rewrites the hardcoded literal.
// REQ-V3R2-RT-007-022: migration 1 substitutes /Users/goos/go/bin/moai with $HOME/go/bin/moai.
func TestM001_RewritesHardcodedLiteral(t *testing.T) {
	root := t.TempDir()
	p := writeWrapper(t, root, "stop", "#!/bin/sh\nexec "+hardcodedLiteral+" hook stop\n", 0o755)

	if err := m001(t).Apply(root); err != nil {
		t.Fatalf("Apply on a rewritable project returned %v, want nil — a completed rewrite is not a failure", err)
	}

	got := readFile(t, p)
	if strings.Contains(got, hardcodedLiteral) {
		t.Errorf("wrapper still carries the hardcoded literal: %q", got)
	}
	if !strings.Contains(got, portableForm) {
		t.Errorf("wrapper does not carry the portable form: %q", got)
	}
}

// TestM001_NoOpWhenAlreadyClean verifies the no-op path for an already clean project.
// REQ-V3R2-RT-007-023: with no hardcoded literal present, the project is treated as already migrated.
func TestM001_NoOpWhenAlreadyClean(t *testing.T) {
	root := t.TempDir()
	body := "#!/bin/sh\nexec " + portableForm + " hook stop\n"
	p := writeWrapper(t, root, "stop", body, 0o755)

	if err := m001(t).Apply(root); err != nil {
		t.Fatalf("Apply on an already-clean project returned %v, want nil — a no-op is not a failure", err)
	}

	if got := readFile(t, p); got != body {
		t.Errorf("no-op path modified the wrapper:\n got %q\nwant %q", got, body)
	}
}

// TestM001_PreservesExecutableBit verifies that execute permissions are preserved.
// REQ-V3R2-RT-007-022: file rewrites preserve the executable bit (0o755).
func TestM001_PreservesExecutableBit(t *testing.T) {
	root := t.TempDir()
	p := writeWrapper(t, root, "stop", "#!/bin/sh\nexec "+hardcodedLiteral+" hook stop\n", 0o755)

	before, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat before: %v", err)
	}
	// Assert the precondition rather than assume it: if the fixture is not
	// executable to begin with, the test proves nothing about preservation.
	if before.Mode().Perm()&0o111 == 0 {
		t.Fatalf("fixture is not executable (%v); the preservation assertion would be vacuous", before.Mode().Perm())
	}

	if err := m001(t).Apply(root); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	after, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat after: %v", err)
	}
	if after.Mode().Perm() != before.Mode().Perm() {
		t.Errorf("mode changed across the rewrite: before %v, after %v", before.Mode().Perm(), after.Mode().Perm())
	}
}

// TestM001_PreservesOtherContent verifies that other content is preserved.
// REQ-V3R2-RT-007-022: only the hardcoded literal is substituted; all other content is preserved.
func TestM001_PreservesOtherContent(t *testing.T) {
	root := t.TempDir()
	body := "#!/bin/sh\n" +
		"# a comment that mentions no path at all\n" +
		"export SOME_VAR=keep-me\n" +
		"exec " + hardcodedLiteral + " hook stop\n" +
		"# trailing line\n"
	p := writeWrapper(t, root, "stop", body, 0o755)

	if err := m001(t).Apply(root); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	want := strings.Replace(body, hardcodedLiteral, portableForm, 1)
	if got := readFile(t, p); got != want {
		t.Errorf("content outside the literal changed:\n got %q\nwant %q", got, want)
	}
}

// TestM001_RollbackNotImplemented verifies that m001 does not support rollback.
// REQ-V3R2-RT-007-024: m001 declares Rollback: nil and does not support rollback.
func TestM001_RollbackNotImplemented(t *testing.T) {
	if got := m001(t).Rollback; got != nil {
		t.Errorf("migration 1 registers a non-nil Rollback; REQ-V3R2-RT-007-024 declares it non-rollback-able")
	}

	root := t.TempDir()
	err := migration.NewRunner(root).Rollback(1)
	if !errors.Is(err, migration.ErrMigrationNotRollbackable) {
		t.Errorf("Rollback(1) returned %v, want ErrMigrationNotRollbackable", err)
	}
}

// TestM001_WindowsGitBash verifies behavior in the Windows Git Bash environment.
// REQ-V3R2-RT-007-060: $HOME is shell-expanded, so the migration is platform-independent.
//
// The platform-independence claim is testable on any OS: what makes it hold is
// that the substitution writes the LITERAL "$HOME/..." for the shell to expand
// at run time, rather than expanding it at migration time into whatever this
// machine's home happens to be. A migration that baked in the current home
// would pass a naive "no longer hardcoded" check and still be wrong.
func TestM001_WindowsGitBash(t *testing.T) {
	root := t.TempDir()
	p := writeWrapper(t, root, "stop", "#!/bin/sh\nexec "+hardcodedLiteral+" hook stop\n", 0o755)

	if err := m001(t).Apply(root); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got := readFile(t, p)
	if !strings.Contains(got, "$HOME/go/bin/moai") {
		t.Errorf("wrapper does not carry the unexpanded $HOME form: %q", got)
	}
	if home, ok := os.LookupEnv("HOME"); ok && home != "" && strings.Contains(got, filepath.Join(home, "go", "bin", "moai")) {
		t.Errorf("wrapper carries this machine's expanded home rather than the literal $HOME: %q", got)
	}
}
