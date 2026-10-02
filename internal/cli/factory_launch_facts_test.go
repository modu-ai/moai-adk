package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// AC-KRS-003(b): a launcher invocation creates no record. The listing of the
// record directory is identical before and after — same names, no addition.
//
// The fixture root is created and populated by the test, so this listing is
// reproducible in a way the live directory is not.
func TestLauncherWritesNoFactoryRecord(t *testing.T) {
	// Queue state is staged under a temp project root; drop the TestMain
	// MOAI_HOME sandbox so it resolves project-locally (card t1229).
	t.Setenv(config.EnvHome, "")
	root := t.TempDir()
	recordDir := factory.StateDirForRoot(root)
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

	restore := exportFactoryLaunchFacts("SPEC-EXAMPLE-001", factory.BackendGLM)
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
// than as a literal argument to a record write. The backend marker is the one
// launch fact this function publishes (SPEC-LAUNCHER-ENTRY-FLAGS-001 M5a
// dropped the SPEC identifier with its last launcher).
func TestLaunchFactsAreExportedForTheSessionToRead(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	t.Setenv(config.EnvFactoryBackend, "")

	restore := exportFactoryLaunchFacts("", factory.BackendGLM)

	if got := os.Getenv(config.EnvFactoryBackend); got != factory.BackendGLM {
		t.Fatalf("%s = %q, want %q", config.EnvFactoryBackend, got, factory.BackendGLM)
	}

	restore()

	if got := os.Getenv(config.EnvFactoryBackend); got != "" {
		t.Fatalf("after restore %s = %q, want empty", config.EnvFactoryBackend, got)
	}
}

// No launcher publishes the retired SPEC marker (REQ-011): the function the
// factory entries share exports the backend and nothing else, whatever
// identifier the caller's entry parse carries.
func TestLaunchFactsPublishNoSpecMarker(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	t.Setenv(retiredSpecMarker, "")
	if err := os.Unsetenv(retiredSpecMarker); err != nil {
		t.Fatalf("Unsetenv: %v", err)
	}

	restore := exportFactoryLaunchFacts("SPEC-EXAMPLE-001", factory.BackendClaude)
	defer restore()

	if _, present := os.LookupEnv(retiredSpecMarker); present {
		t.Fatalf("%s was exported by the launch facts", retiredSpecMarker)
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
