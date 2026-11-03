//go:build !windows

package cli

// codex_close_parent_exit_test.go — what happens to the codex session's process
// group when the PARENT ends (Issue #1948).
//
// The session runs in a process group of its own, so a terminal Ctrl-C (which is
// delivered to the terminal's foreground group) no longer reaches the codex child
// directly. Only close ends the group, so the question is which ways of ending
// the parent still reach close:
//
//   - the MCP server's path, a SIGTERM/SIGINT handled by signal.NotifyContext
//     that cancels the context and lets the deferred close run — guarded here;
//   - a parent that dies without running its deferred calls (SIGKILL, or a signal
//     with no handler) — measured here, and logged as a limit, not asserted away.
//
// The parent is a re-run of this test binary (TestCodexCloseParentExitHelper)
// with the fake child of codex_close_descendant_repro_test.go. Everything is torn
// down by the stop file the fake's descendant watches, and by a signal to the
// group this test's helper created, never by name.

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

const closeHelperEnv = "MOAI_TEST_CODEX_CLOSE_HELPER"

// TestCodexCloseParentExitHelper is the parent process. It does nothing unless
// the environment names a mode; the outer tests re-run the binary with it set.
func TestCodexCloseParentExitHelper(t *testing.T) {
	mode := os.Getenv(closeHelperEnv)
	if mode == "" {
		return
	}
	bin := os.Getenv(closeHelperEnv + "_BIN")
	conn, err := (realCodexSessionRunner{}).start(context.Background(), bin,
		[]string{"app-server", os.Getenv(closeHelperEnv + "_WORK"), os.Getenv(closeHelperEnv + "_STOP"), "descendant", "eof"},
		filepath.Dir(bin))
	if err != nil {
		t.Fatalf("start the fake child: %v", err)
	}
	pid := conn.(*realCodexConn).pid()
	if err := os.WriteFile(os.Getenv(closeHelperEnv+"_READY"), []byte(strconv.Itoa(pid)), 0o600); err != nil {
		t.Fatal(err)
	}
	if mode == "server" {
		// The MCP server's shape: a handled signal ends the wait, the deferred
		// close runs, the process exits normally.
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
		defer stop()
		defer func() { _ = conn.close() }()
		<-ctx.Done()
		return
	}
	select {} // plain: no handler, no close — the outer test ends this process
}

type closeParentRun struct {
	cmd   *exec.Cmd
	work  string
	stop  string
	group int
	// plugins is the directory the fake's descendant writes into.
	plugins string
}

func startCloseParent(t *testing.T, mode string) *closeParentRun {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "codex")
	if err := os.WriteFile(bin, []byte(reproChildScript), 0o755); err != nil {
		t.Fatal(err)
	}
	work, err := os.MkdirTemp("", "fakecodexparent-")
	if err != nil {
		t.Fatal(err)
	}
	r := &closeParentRun{work: work, stop: filepath.Join(dir, "stop"), plugins: filepath.Join(work, ".tmp", "plugins-clone-X", "plugins")}
	ready := filepath.Join(dir, "ready")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCodexCloseParentExitHelper$")
	cmd.Env = append(os.Environ(),
		closeHelperEnv+"="+mode, closeHelperEnv+"_BIN="+bin, closeHelperEnv+"_WORK="+work,
		closeHelperEnv+"_STOP="+r.stop, closeHelperEnv+"_READY="+ready)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start the parent: %v", err)
	}
	r.cmd = cmd
	t.Cleanup(func() {
		// Cleanup guarantee, in order: the descendant ends on the stop file; the
		// parent (ours) is killed if still running; the group this test's parent
		// created is signalled only if it still has a member.
		_ = os.WriteFile(r.stop, nil, 0o600)
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
		time.Sleep(100 * time.Millisecond)
		if codexGroupAlive(r.group) {
			signalCodexGroup(r.group, false)
			waitCodexGroupGone(r.group, time.Second)
			signalCodexGroup(r.group, true)
		}
		_ = os.RemoveAll(work)
	})
	deadline := time.Now().Add(15 * time.Second)
	for {
		b, rerr := os.ReadFile(ready)
		_, serr := os.Stat(filepath.Join(r.plugins, "f0"))
		if rerr == nil && serr == nil {
			r.group, _ = strconv.Atoi(string(b))
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the parent never reported a running session (ready err %v, writer err %v)", rerr, serr)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return r
}

// grows reports whether the descendant added entries during a short window.
func (r *closeParentRun) grows() bool {
	count := func() int {
		es, _ := os.ReadDir(r.plugins)
		return len(es)
	}
	n := count()
	time.Sleep(300 * time.Millisecond)
	return count() > n
}

func (r *closeParentRun) waitExit(t *testing.T) {
	t.Helper()
	done := make(chan struct{})
	go func() { _ = r.cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("the parent did not exit")
	}
}

// A SIGTERM handled the way the MCP server handles it ends the session's group:
// the process exits normally and nothing is left writing.
func TestCodexClose_HandledSignalEndsTheSessionGroup(t *testing.T) {
	r := startCloseParent(t, "server")
	if !r.grows() {
		t.Fatal("control: the descendant was not writing before the signal, so the test measures nothing")
	}
	if err := r.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	r.waitExit(t)
	if r.grows() {
		t.Errorf("a handled SIGTERM left the session's descendant writing")
	}
	if codexGroupAlive(r.group) {
		t.Errorf("the session's process group %d still has a member after the parent exited", r.group)
	}
	if err := os.RemoveAll(r.work); err != nil {
		t.Errorf("RemoveAll after the parent exited: %v", err)
	}
}

// A parent that dies without running its deferred calls cannot end the group: no
// code runs in it. This is a limit of the repair, measured rather than asserted
// away; the descendant is ended by the test's own cleanup, not by the product.
func TestCodexClose_ParentKilledLeavesTheGroup_Limit(t *testing.T) {
	r := startCloseParent(t, "plain")
	if !r.grows() {
		t.Fatal("control: the descendant was not writing before the parent was killed")
	}
	if err := r.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	r.waitExit(t)
	t.Logf("MEASURED LIMIT: after the parent was SIGKILLed, the session group %d alive=%v, descendant still writing=%v",
		r.group, codexGroupAlive(r.group), r.grows())
}
