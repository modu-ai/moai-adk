//go:build unix

package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/modu-ai/moai-adk/internal/config"
)

// auditHygieneUnixCells are the FIFO cells of TestLauncherAnchorAuditHygiene.
func auditHygieneUnixCells(t *testing.T, verdict *leaderAnchorVerdict) {
	t.Helper()

	// A FIFO at the audit path blocks a plain open until a reader appears, which
	// would hang the leader launch: the Lstat refusal comes before any open.
	t.Run("fifo_refused_without_blocking", func(t *testing.T) {
		root := t.TempDir()
		auditPath := filepath.Join(root, config.DefaultLeaderAnchorAuditRelPath)
		if err := os.MkdirAll(filepath.Dir(auditPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := unix.Mkfifo(auditPath, 0o600); err != nil {
			t.Skipf("create the FIFO: %v", err)
		}
		done := make(chan error, 1)
		go func() { done <- appendLeaderAnchorAuditRow(root, "claude", verdict) }()
		select {
		case err := <-done:
			if err == nil {
				t.Errorf("an audit row into a FIFO returned nil, want a refusal")
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("the append did not return within 5s: the open is blocked on the FIFO")
		}
	})

	// A FIFO with a reader attached, swapped in after the Lstat check, survives
	// O_NOFOLLOW|O_NONBLOCK; the descriptor's own file type is re-read and refused.
	t.Run("descriptor_recheck_refuses_a_swapped_fifo", func(t *testing.T) {
		root := t.TempDir()
		auditPath := filepath.Join(root, config.DefaultLeaderAnchorAuditRelPath)
		if err := os.MkdirAll(filepath.Dir(auditPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(auditPath, []byte("regular\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		var reader *os.File
		swapped := false
		prev := leaderAnchorAuditOpenRace
		t.Cleanup(func() {
			leaderAnchorAuditOpenRace = prev
			if reader != nil {
				_ = reader.Close()
			}
		})
		leaderAnchorAuditOpenRace = func(path string) {
			if swapped || path != auditPath {
				return
			}
			swapped = true
			if err := os.Remove(path); err != nil {
				t.Errorf("remove the regular audit log in the race window: %v", err)
				return
			}
			if err := unix.Mkfifo(path, 0o600); err != nil {
				t.Errorf("create the FIFO in the race window: %v", err)
				return
			}
			r, err := os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK, 0)
			if err != nil {
				t.Errorf("open the FIFO reader in the race window: %v", err)
				return
			}
			reader = r
		}
		f, err := openLeaderAnchorAuditLog(auditPath)
		if f != nil {
			_ = f.Close()
		}
		if !swapped || reader == nil {
			t.Fatal("the race seam never staged the swap: the open does not reach the window between the Lstat check and the open")
		}
		if err == nil || f != nil {
			t.Errorf("a FIFO swapped in after the Lstat check yielded handle=%v err=%v, want a refusal and no handle", f, err)
		}
		read := make(chan int, 1)
		go func() {
			buf := make([]byte, 256)
			n, _ := reader.Read(buf)
			read <- n
		}()
		select {
		case n := <-read:
			if n != 0 {
				t.Errorf("the FIFO reader read %d byte(s): the audit row reached the reader", n)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("reading the FIFO blocked: a writer is still attached to it")
		}
	})
}
