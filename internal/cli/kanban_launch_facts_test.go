package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// AC-KRS-003(b): a launcher invocation creates no record. The listing of the
// record directory is identical before and after — same names, no addition.
//
// The fixture root is created and populated by the test, so this listing is
// reproducible in a way the live directory is not.
func TestLauncherWritesNoKanbanRecord(t *testing.T) {
	root := t.TempDir()
	recordDir := filepath.Join(root, ".moai", "state", "kanban")
	if err := os.MkdirAll(recordDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// A pre-existing record, so "no addition" is distinguishable from "no
	// directory at all". Seeded as bytes rather than through the package
	// writer: AC-KRS-003(a) greps internal/cli for a record write and requires
	// zero hits, and a test that called the writer would be one.
	if err := os.WriteFile(filepath.Join(recordDir, "pre-existing.json"),
		[]byte(`{"session_id":"pre-existing","spec_id":"","role":"lead","backend":"claude",`+
			`"entered_at":"2026-08-23T17:47:22Z","deepscan_dir":"","verify_reentries":0}`+"\n"),
		0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// The single slot holds an identifier the pre-change writer would have
	// keyed on — the exact input that produced the defect.
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.WriteFile(filepath.Join(stateDir, "current-session-id.txt"),
		[]byte("P-00000000-1111-2222-3333-444444444444"), 0o600); err != nil {
		t.Fatalf("seed sidecar: %v", err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", root)

	before := listNames(t, recordDir)

	restore := exportKanbanLaunchFacts("SPEC-EXAMPLE-001", kanban.BackendGLM)
	defer restore()

	after := listNames(t, recordDir)
	if len(before) != len(after) {
		t.Fatalf("record directory changed: before %v, after %v", before, after)
	}
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("record directory changed: before %v, after %v", before, after)
		}
	}
}

// AC-KRS-006(a): the backend travels through the launch environment rather
// than as a literal argument to a record write, and the SPEC identifier is
// exported on every launch path (not only the kanban lead's).
func TestLaunchFactsAreExportedForTheSessionToRead(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	t.Setenv(config.EnvMoaiKanbanBackend, "")
	t.Setenv(config.EnvMoaiKanbanSpec, "")

	restore := exportKanbanLaunchFacts("SPEC-EXAMPLE-001", kanban.BackendGLM)

	if got := os.Getenv(config.EnvMoaiKanbanBackend); got != kanban.BackendGLM {
		t.Fatalf("%s = %q, want %q", config.EnvMoaiKanbanBackend, got, kanban.BackendGLM)
	}
	if got := os.Getenv(config.EnvMoaiKanbanSpec); got != "SPEC-EXAMPLE-001" {
		t.Fatalf("%s = %q, want SPEC-EXAMPLE-001", config.EnvMoaiKanbanSpec, got)
	}

	restore()

	if got := os.Getenv(config.EnvMoaiKanbanBackend); got != "" {
		t.Fatalf("after restore %s = %q, want empty", config.EnvMoaiKanbanBackend, got)
	}
	if got := os.Getenv(config.EnvMoaiKanbanSpec); got != "" {
		t.Fatalf("after restore %s = %q, want empty", config.EnvMoaiKanbanSpec, got)
	}
}

// An absent SPEC identifier is not exported as an empty value — the session
// reads presence, and an empty export would be a SPEC that is not there.
func TestEmptySpecIsNotExported(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	if err := os.Unsetenv(config.EnvMoaiKanbanSpec); err != nil {
		t.Fatalf("Unsetenv: %v", err)
	}

	restore := exportKanbanLaunchFacts("", kanban.BackendClaude)
	defer restore()

	if _, present := os.LookupEnv(config.EnvMoaiKanbanSpec); present {
		t.Fatalf("%s was exported though no SPEC was supplied", config.EnvMoaiKanbanSpec)
	}
}

func listNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
