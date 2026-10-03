package cli

// managed_codex_tui_repair_test.go — sync-audit repair round of card t1408
// (SPEC-FACTORY-MANAGED-TUI-001): the connection-loss order, the session log
// file's safety, and the operator-turn window.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestManagedCodexConnectionLostBeforeAttachStopsTUI covers the order the
// server-death monitor used to miss: the connection reader ends BEFORE the TUI
// is attached. The loss must still stop the TUI and reach the driver.
func TestManagedCodexConnectionLostBeforeAttachStopsTUI(t *testing.T) {
	tuiSetDuration(t, &managedTUIStopGrace, 400*time.Millisecond)
	f := newTUIFake(t, "tui-lost-before-attach")
	sess := f.newSession()
	if err := sess.DeliverTurn(managedPrimingPrompt); err != nil {
		t.Fatalf("priming turn: %v", err)
	}
	f.control("close-launcher")
	ended := waitUntil(func() bool {
		select {
		case _, ok := <-sess.client.events:
			return !ok
		default:
			return false
		}
	}, tuiAttachWait)
	if !ended {
		t.Fatalf("the connection reader did not end after the server closed the connection")
	}
	done, attached := sess.AttachOperator()
	if !attached {
		t.Fatalf("the session did not attach the TUI")
	}
	ok, err := tuiAwait(done, tuiAttachWait)
	if !ok {
		t.Fatalf("BUG: the connection was lost before attach; the TUI stayed alive and no result reached the driver within %s", tuiAttachWait)
	}
	if !errors.Is(err, errManagedCodexConnectionClosed) {
		t.Errorf("result = %v, want the connection-closed error", err)
	}
	sess.tui.mu.Lock()
	pid := sess.tui.cmd.Process.Pid
	sess.tui.mu.Unlock()
	if pid == 0 || tuiPIDAlive(pid) {
		t.Errorf("the TUI child (pid %d) outlived the lost-connection result", pid)
	}
}

func TestManagedTUILogFileIsSafe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes and symlinks")
	}
	t.Run("symlink_not_followed", func(t *testing.T) {
		root := t.TempDir()
		outside := filepath.Join(t.TempDir(), "outside.txt")
		if err := os.WriteFile(outside, []byte("keep\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		logs := filepath.Join(root, ".moai", "logs")
		if err := os.MkdirAll(logs, 0o755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(logs, "factory-managed-run1-lane-1.log")
		if err := os.Symlink(outside, link); err != nil {
			t.Fatal(err)
		}
		f, path, err := openManagedTUILog(root, "run1", "lane-1")
		if err == nil {
			_, _ = f.WriteString("probe\n")
			_ = f.Close()
		}
		got, _ := os.ReadFile(outside)
		if string(got) != "keep\n" {
			t.Errorf("the log write followed the planted symlink out of the project: outside file = %q", got)
		}
		if err == nil {
			info, lerr := os.Lstat(path)
			if lerr != nil || info.Mode()&os.ModeSymlink != 0 {
				t.Errorf("the log path is still a symlink after open (lstat err %v)", lerr)
			}
		}
	})
	t.Run("existing_file_mode_tightened", func(t *testing.T) {
		root := t.TempDir()
		logs := filepath.Join(root, ".moai", "logs")
		if err := os.MkdirAll(logs, 0o755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(logs, "factory-managed-run2-lane-1.log")
		if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o644); err != nil {
			t.Fatal(err)
		}
		f, _, err := openManagedTUILog(root, "run2", "lane-1")
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("pre-existing log mode = %o, want 600", got)
		}
		if b, _ := os.ReadFile(path); string(b) != "old\n" {
			t.Errorf("the existing content must be kept (append), got %q", b)
		}
	})
	t.Run("new_file_mode", func(t *testing.T) {
		root := t.TempDir()
		f, path, err := openManagedTUILog(root, "run3", "lane-1")
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("new log mode = %o, want 600", got)
		}
	})
}

// TestManagedOperatorTurnInsideArmedWindowKnownDebt pins audit F4 (t1408) as a
// known limitation: an operator turn that starts AND completes inside the
// owner's armed window (before the owner's turn/start response) is marked
// owned, the window closes, and the owner's own lifecycle frames are then
// filtered, so the owner would wait for its turn until the turn timeout. A
// small safe fix does not exist (attributing the first turn/started needs the
// turn/start response id, i.e. holding frames until the response, which touches
// the REQ-MT-008 exception). When the debt is fixed this test flips: change the
// two assertions below.
func TestManagedOperatorTurnInsideArmedWindowKnownDebt(t *testing.T) {
	c := &managedCodexAppClient{}
	c.enableOperatorScoping(time.Minute)
	c.armTurn()
	frame := func(id string) json.RawMessage {
		return json.RawMessage(`{"turn":{"id":"` + id + `","status":"x"}}`)
	}
	// Operator turn op-1 starts and completes inside the armed window.
	c.noteTurnStarted(frame("op-1"))
	c.trackTurnCompleted(frame("op-1"))
	c.noteTurnCompleted(frame("op-1"))
	// The owner's turn/start response now arrives, then its lifecycle frames.
	c.noteTurnStartResponse(json.RawMessage(`{"turn":{"id":"own-1"}}`))
	if c.noteTurnStarted(frame("own-1")) {
		t.Errorf("known debt F4 appears fixed: the owner's turn/started is now forwarded; update this test to assert the fix")
	}
	c.mu.Lock()
	_, ownedOp := c.owned["op-1"]
	c.mu.Unlock()
	if !ownedOp {
		t.Errorf("known debt F4 appears fixed: op-1 is no longer marked owned; update this test to assert the fix")
	}
}
