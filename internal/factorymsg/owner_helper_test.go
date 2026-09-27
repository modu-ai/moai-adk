package factorymsg

import (
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

const ownerHelperEnv = "MOAI_T1082_FACTORYMSG_OWNER_HELPER"

// TestFactorymsgOwnerHelperProcess is not a test. Re-executed as a child, it
// is a live process distinct from the test process, so an endpoint it owns is
// current under the t1074 live-owner rule until the test stops it. It lives
// until its stdin closes.
func TestFactorymsgOwnerHelperProcess(t *testing.T) {
	if os.Getenv(ownerHelperEnv) != "1" {
		return
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}

// stoppableOwner is a live helper process the test can make not current.
type stoppableOwner struct {
	PID   int
	Start string
	stop  func()
}

// Stop kills AND reaps the helper, then waits until the live-owner probe no
// longer reports it current. A killed but unreaped child is a zombie that the
// probe still reads as live, so a kill alone does not make the owner stale.
func (o stoppableOwner) Stop(t *testing.T) {
	t.Helper()
	o.stop()
	deadline := time.Now().Add(5 * time.Second)
	for {
		start, state := homestate.ProbeProcessIdentity(o.PID)
		if state != homestate.ProcessIdentityLive || start != o.Start {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("owner %d still current after kill and wait", o.PID)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// startStoppableOwner starts the helper child. Cleanup stops it on every path.
func startStoppableOwner(t *testing.T) stoppableOwner {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestFactorymsgOwnerHelperProcess$", "-test.timeout=300s")
	cmd.Env = append(os.Environ(), ownerHelperEnv+"=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start owner helper: %v", err)
	}
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
	t.Cleanup(stop)
	deadline := time.Now().Add(5 * time.Second)
	for {
		start, state := homestate.ProbeProcessIdentity(cmd.Process.Pid)
		if state == homestate.ProcessIdentityLive && start != "" {
			return stoppableOwner{PID: cmd.Process.Pid, Start: start, stop: stop}
		}
		if time.Now().After(deadline) {
			t.Fatalf("owner helper %d never became probeable (state=%v)", cmd.Process.Pid, state)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
