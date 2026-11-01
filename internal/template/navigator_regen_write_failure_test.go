package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNavigatorRegen_WriteFailureDoesNotAdvanceSentinel pins card t2010 / PJ-10:
// when one of the three navigator outputs cannot be written, the freshness
// sentinel (last-regen-commit.txt, written last) must not advance, the failure
// is logged, and the script stays fail-open (exit 0).
func TestNavigatorRegen_WriteFailureDoesNotAdvanceSentinel(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-FIX-001", "Fixture", "draft", "run", "internal/x")

	// Fault: the navigator output directory is read-only, so none of the three
	// outputs can be written (the state directory stays writable).
	if os.Geteuid() == 0 {
		t.Skip("permission fault cannot be injected as root")
	}
	navDir := filepath.Join(dir, ".moai", "project", "navigator")
	if err := os.MkdirAll(navDir, 0o755); err != nil {
		t.Fatalf("mkdir nav dir: %v", err)
	}
	if err := os.Chmod(navDir, 0o555); err != nil {
		t.Fatalf("chmod nav dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(navDir, 0o755) })

	out, exitErr := runRegen(t, dir)
	if exitErr != nil {
		t.Fatalf("regen must stay fail-open (exit 0), got %v\n%s", exitErr, out)
	}

	sentinel := filepath.Join(dir, ".moai", "state", "navigator", "last-regen-commit.txt")
	if _, err := os.Stat(sentinel); err == nil {
		t.Errorf("freshness sentinel %s was written although capability-map.md could not be written", sentinel)
	}

	warn, err := os.ReadFile(filepath.Join(dir, ".moai", "logs", "navigator-warnings.log"))
	if err != nil {
		t.Fatalf("expected a warning log naming the failed write: %v", err)
	}
	if !strings.Contains(string(warn), "capability-map.md") {
		t.Errorf("warning log does not name the failed output; got:\n%s", warn)
	}
}

// TestNavigatorRegen_SuccessStillWritesSentinel guards the other direction: a
// clean run still writes the sentinel (the instrument can go green).
func TestNavigatorRegen_SuccessStillWritesSentinel(t *testing.T) {
	dir := t.TempDir()
	head := initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-FIX-001", "Fixture", "draft", "run", "internal/x")

	if out, exitErr := runRegen(t, dir); exitErr != nil {
		t.Fatalf("regen failed: %v\n%s", exitErr, out)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".moai", "state", "navigator", "last-regen-commit.txt"))
	if err != nil {
		t.Fatalf("sentinel missing after a clean run: %v", err)
	}
	if string(got) != head+"\n" {
		t.Errorf("sentinel = %q, want %q", got, head+"\n")
	}
}
