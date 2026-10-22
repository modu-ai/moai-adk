package migrations_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/testguard"
)

func TestM001PreservesWindowsPrivatePolicyAndForeignStagingName(t *testing.T) {
	root := t.TempDir()
	path := writeWrapper(t, root, "stop", "exec "+hardcodedLiteral+" hook stop\n", 0o600)
	foreign := path + ".tmp"
	if err := os.WriteFile(foreign, []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := m001(t).Apply(root); err != nil {
		t.Fatal(err)
	}
	if err := testguard.CheckFileModePolicy(path, 0o600); err != nil {
		t.Fatalf("private wrapper policy changed: %v", err)
	}
	if data, err := os.ReadFile(foreign); err != nil || string(data) != "foreign" {
		t.Errorf("foreign fixed staging name changed: %q, %v", data, err)
	}
	matches, err := filepath.Glob(path + ".*.tmp")
	if err != nil || len(matches) != 0 {
		t.Errorf("rewrite left random stages: %v, %v", matches, err)
	}
}
