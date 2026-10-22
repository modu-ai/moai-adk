package factorymsg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The re-executed native test binary stays alive until the parent closes its
// input. That handshake lets the parent capture a real start fingerprint,
// rather than racing a shell's exit and inventing an identity after the fact.
func TestNativeProcessFixtureChild(t *testing.T) {
	if os.Getenv("MOAI_FACTORYMSG_NATIVE_PROCESS_CHILD") != "1" {
		return
	}
	code, err := strconv.Atoi(os.Getenv("MOAI_NATIVE_PROCESS_EXIT"))
	if err != nil {
		os.Exit(91)
	}
	fmt.Fprintln(os.Stdout, "ready")
	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		os.Exit(92)
	}
	os.Exit(code)
}

func nativeProcessFixtureCommand(t *testing.T, exitCode int) *exec.Cmd {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "-test.run=^TestNativeProcessFixtureChild$")
	cmd.Env = append(os.Environ(), "MOAI_FACTORYMSG_NATIVE_PROCESS_CHILD=1", "MOAI_NATIVE_PROCESS_EXIT="+strconv.Itoa(exitCode))
	return cmd
}

type nativeProcessFixture struct {
	pid         int
	fingerprint string
	finish      func() error
}

func startNativeProcessFixture(t *testing.T, exitCode int) nativeProcessFixture {
	t.Helper()
	cmd := nativeProcessFixtureCommand(t, exitCode)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	// CommandContext retains the exact executable/arguments/environment above.
	bounded := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	bounded.Env = cmd.Env
	input, err := bounded.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	output, err := bounded.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	bounded.Stderr = &stderr
	if err := bounded.Start(); err != nil {
		cancel()
		t.Fatalf("start native process: %v", err)
	}
	waited := false
	t.Cleanup(func() {
		cancel()
		_ = input.Close()
		if !waited {
			_ = bounded.Process.Kill()
			_ = bounded.Wait()
		}
	})
	ready := make(chan error, 1)
	go func() {
		line, err := bufio.NewReader(output).ReadString('\n')
		if err == nil && line != "ready\n" {
			err = fmt.Errorf("unexpected readiness %q", line)
		}
		ready <- err
	}()
	select {
	case err := <-ready:
		if err != nil {
			t.Fatalf("native child readiness: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("native child readiness timed out: %v", ctx.Err())
	}
	pid := bounded.Process.Pid
	fingerprint, state := homestate.ProbeProcessIdentity(pid)
	if state != homestate.ProcessIdentityLive || fingerprint == "" {
		t.Fatalf("ready native child identity=(%q,%s), want measurable live owner", fingerprint, state)
	}
	return nativeProcessFixture{pid: pid, fingerprint: fingerprint, finish: func() error {
		if waited {
			t.Fatal("native child waited twice")
		}
		_ = input.Close()
		err := bounded.Wait()
		waited = true
		cancel()
		if got := homestate.ClassifyOwnerWith(homestate.ProbeProcessIdentity, pid, fingerprint); got != homestate.OwnerDead {
			t.Fatalf("waited native child owner=%s, want dead; stderr=%s", got, &stderr)
		}
		return err
	}}
}

func TestNativeProcessFixtureObservesLiveThenExitedOwner(t *testing.T) {
	child := startNativeProcessFixture(t, 0)
	if got := homestate.ClassifyOwnerWith(homestate.ProbeProcessIdentity, child.pid, child.fingerprint); got != homestate.OwnerLive {
		t.Fatalf("held native child owner=%s, want live", got)
	}
	if err := child.finish(); err != nil {
		t.Fatalf("native child exit: %v", err)
	}
}
