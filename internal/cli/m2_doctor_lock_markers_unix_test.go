//go:build unix

package cli

// m2_doctor_lock_markers_unix_test.go — the unix-execution arm of the
// doctor's User Lock row (gate round 21: the standing compile-vs-runtime
// rule — a unix-only syscall lives in a `//go:build unix` file from the
// first write, never under a runtime guard).

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// TestDoctorUserLockMarkersIrregularIsSurfaced — gate round 19: a FIFO at
// a marker path is warned, never read (this test returning is the proof
// the row did not hang).
func TestDoctorUserLockMarkersIrregularIsSurfaced(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(userassets.GuardMarkerPath(home)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(userassets.GuardMarkerPath(home), 0o644); err != nil {
		t.Fatal(err)
	}
	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("an irregular marker produced %s (want warn): %q", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "irregular object") {
		t.Errorf("the row does not name the irregular object: %q", check.Message)
	}
}
