package factory

// legacy_state_dir_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1 (card t1399),
// AC-016: the legacy project-local state directory is still read from
// `.moai/state/kanban`. The AC names the factory package for this test
// (`./internal/factory`); through M7 the package path is `./internal/factory`,
// where this file lives and is carried along by the package rename at M8.
//
// The directory name below is a literal on purpose: it is the on-disk name an
// older binary wrote, which no rename of a Go identifier may move.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyStateDirStillRead(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, ".moai", "state", "kanban")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "backlog.json"), []byte(`{"items":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if got := LegacyStateDirForRoot(root); got != legacy {
		t.Errorf("LegacyStateDirForRoot = %q, want the frozen on-disk name %q", got, legacy)
	}
	// The read-only surfaces observe the legacy directory in place.
	if got := filepath.Dir(BacklogPathForRoot(root)); got != legacy {
		t.Errorf("BacklogPathForRoot resolved to %q, want the legacy directory %q (nothing else holds a queue)", got, legacy)
	}
	if got := RuntimeStateDirForRoot(root); got != legacy {
		t.Errorf("RuntimeStateDirForRoot = %q, want the legacy directory %q", got, legacy)
	}
	if got := filepath.Dir(RecordPath(root, "0000000a-1111-2222-3333-444444444444")); got != legacy {
		t.Errorf("RecordPath resolved to %q, want the legacy directory %q", got, legacy)
	}
}
