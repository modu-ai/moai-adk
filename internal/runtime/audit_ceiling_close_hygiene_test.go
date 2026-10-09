//go:build darwin || linux

package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAppendProgressRecordSeedCloseHygiene (gate round-46 item 4; F16
// close-hygiene guard, SPEC-PROGRESS-RECORD-IO-001 M3 step 4) — the held
// descriptor is closed by the time appendProgressRecord returns on BOTH
// paths: the success path closes it before the rename
// (audit_ceiling.go close-before-rename) and the deferred close covers
// every abort path. A seedFileMetadataFn wrapper captures the descriptor
// held after os.CreateTemp; after the call returns, a second Close must
// report the file already closed (os.ErrClosed) — a refactor dropping the
// closes leaks the descriptor and this probe observes it.
func TestAppendProgressRecordSeedCloseHygiene(t *testing.T) {
	t.Run("success path closes before rename", func(t *testing.T) {
		specDir := t.TempDir()
		path := filepath.Join(specDir, "progress.md")
		pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
		if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
			t.Fatal(err)
		}
		var held *os.File
		orig := seedFileMetadataFn
		seedFileMetadataFn = func(tmp *os.File, tmpPath, original string) error {
			held = tmp
			return orig(tmp, tmpPath, original)
		}
		t.Cleanup(func() { seedFileMetadataFn = orig })

		if err := appendProgressRecord(specDir, "- new record"); err != nil {
			t.Fatalf("the replace failed before the close contract could be observed: %v", err)
		}
		if held == nil {
			t.Fatal("the seeder was never invoked — the descriptor was not captured")
		}
		if err := held.Close(); !errors.Is(err, os.ErrClosed) {
			t.Fatalf("the held descriptor survived the successful replace (second Close: %v) — a close was dropped on the success path", err)
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatal(rerr)
		}
		if !strings.Contains(string(raw), "- new record") {
			t.Fatalf("the record did not land through the completed replace:\n%s", raw)
		}
	})

	t.Run("abort path closes through the defer", func(t *testing.T) {
		specDir := t.TempDir()
		path := filepath.Join(specDir, "progress.md")
		pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
		if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
			t.Fatal(err)
		}
		var held *os.File
		orig := seedFileMetadataFn
		seedFileMetadataFn = func(tmp *os.File, tmpPath, original string) error {
			held = tmp
			return errSeedInjected
		}
		t.Cleanup(func() { seedFileMetadataFn = orig })

		if err := appendProgressRecord(specDir, "- new record"); err == nil {
			t.Fatal("the injected seed failure did not abort the replace")
		}
		if held == nil {
			t.Fatal("the seeder was never invoked — the descriptor was not captured")
		}
		if err := held.Close(); !errors.Is(err, os.ErrClosed) {
			t.Fatalf("the held descriptor survived the aborted replace (second Close: %v) — the deferred close was dropped on the abort path", err)
		}
	})
}
