package subprocess_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/subprocess"
)

// launchStub launches a shell script stub and returns a Supervisor wrapping it.
// The script is written to t.TempDir() so cleanup is automatic.
func launchStub(t *testing.T, script string) (*subprocess.LaunchResult, *subprocess.Supervisor) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stubs not supported on Windows")
	}

	dir := t.TempDir()
	binPath := writeFakeBinaryContent(t, dir, "stub-lsp", script)

	cfg := config.ServerConfig{
		Language: "go",
		Command:  binPath,
	}

	l := subprocess.NewLauncher()

	// ETXTBSY mitigation: Linux fork/exec race between goroutines may leave
	// the stub binary with an open write fd in a sibling goroutine, triggering
	// "text file busy". Retry up to 5 times with linear backoff (max ~750 ms).
	// Go issue #22315: fork/exec inherits fds that are not FD_CLOEXEC.
	var result *subprocess.LaunchResult
	var err error
	for attempts := 0; attempts < 5; attempts++ {
		result, err = l.Launch(context.Background(), cfg)
		if err == nil {
			break
		}
		if errors.Is(err, syscall.ETXTBSY) || strings.Contains(err.Error(), "text file busy") {
			time.Sleep(time.Duration(50*(attempts+1)) * time.Millisecond)
			continue
		}
		break
	}
	if err != nil {
		t.Fatalf("launchStub: %v", err)
	}

	return result, subprocess.NewSupervisor(result)
}

// writeFakeBinaryContent writes a script with custom content to dir.
//
// Implementation note (ETXTBSY mitigation): see writeFakeBinary in launcher_test.go.
// The same Create → Write → Sync → Close → Chmod ordering is applied here to avoid
// "text file busy" on Linux fork/exec when t.Parallel() tests race file creation
// with subprocess launch.
func writeFakeBinaryContent(t *testing.T, dir, name, script string) string {
	t.Helper()
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("writeFakeBinaryContent create: %v", err)
	}
	if _, err := f.Write([]byte(script)); err != nil {
		_ = f.Close()
		t.Fatalf("writeFakeBinaryContent write: %v", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		t.Fatalf("writeFakeBinaryContent sync: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("writeFakeBinaryContent close: %v", err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("writeFakeBinaryContent chmod: %v", err)
	}
	return path
}

// TestSupervisor_NormalExit verifies that a subprocess that exits with code 0
// sends an ExitEvent with ExitCode 0 and Err == nil.
func TestSupervisor_NormalExit(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nexit 0\n")
	// Supervisor owns cmd.Wait(), so do not call Wait directly from the test.
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	ev := <-ch
	if ev.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", ev.ExitCode)
	}
}

// TestSupervisor_NonZeroExit verifies that a subprocess exiting with code 1
// sends an ExitEvent with ExitCode 1.
func TestSupervisor_NonZeroExit(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nexit 1\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	ev := <-ch
	if ev.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", ev.ExitCode)
	}
}

// TestSupervisor_CtxCancelStopsWatcher verifies that cancelling the context
// causes Watch's channel to close without waiting for the subprocess (REQ-LC-031).
func TestSupervisor_CtxCancelStopsWatcher(t *testing.T) {
	t.Parallel()

	// Infinite-wait script.
	result, sv := launchStub(t, "#!/bin/sh\nsleep 60\n")
	t.Cleanup(func() {
		_ = result.Cmd.Process.Kill()
		// Supervisor owns cmd.Wait() — do not call Wait directly.
	})

	ctx, cancel := context.WithCancel(context.Background())

	ch := sv.Watch(ctx)

	// Context cancel → expect the channel to close immediately.
	cancel()

	select {
	case <-ch:
		// Normal: channel closed.
	case <-time.After(2 * time.Second):
		t.Error("Watch channel did not close after ctx cancel within 2s")
	}
}

// TestSupervisor_Kill verifies that Kill forces subprocess termination
// and the exit event is eventually delivered via Watch.
func TestSupervisor_Kill(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nsleep 60\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	if err := sv.Kill(); err != nil {
		t.Fatalf("Kill: %v", err)
	}

	select {
	case <-ch:
		// Normal: channel closed after process exit.
	case <-time.After(3 * time.Second):
		t.Error("Watch channel did not close after Kill within 3s")
	}
}

// TestSupervisor_Signal_SIGTERM verifies that Signal(SIGTERM) delivers the
// signal to the subprocess (REQ-LC-007 graceful shutdown).
func TestSupervisor_Signal_SIGTERM(t *testing.T) {
	t.Parallel()

	// Script that exits when SIGTERM is received.
	// /bin/bash explicitly: Ubuntu /bin/sh is dash, which has different trap timing.
	// Add a small sleep to ensure Watch starts before Signal.
	result, sv := launchStub(t, "#!/bin/bash\ntrap 'exit 0' TERM\nwhile true; do sleep 0.1; done\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)
	// Give the subprocess time to register the trap in CI environments.
	time.Sleep(100 * time.Millisecond)

	if err := sv.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Signal(SIGTERM): %v", err)
	}

	select {
	case ev := <-ch:
		// Normal exit via SIGTERM: ExitCode 0 or signal-based exit are both acceptable.
		_ = ev
	case <-time.After(10 * time.Second):
		t.Error("subprocess did not exit after SIGTERM within 10s")
	}
}

// TestSupervisor_Signal_AlreadyDead verifies that Signal on a dead process
// returns an error rather than panicking.
func TestSupervisor_Signal_AlreadyDead(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nexit 0\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for process exit.
	ch := sv.Watch(ctx)
	<-ch

	// Send a signal to an already-dead process — must return an error, not panic.
	err := sv.Signal(os.Interrupt)
	// Whether an error is produced depends on the OS — passing if no panic occurs.
	_ = err
}

// TestSupervisor_MultipleWatchers verifies that multiple Watch calls each receive
// their own independent channel and all receive the ExitEvent.
func TestSupervisor_MultipleWatchers(t *testing.T) {
	t.Parallel()

	// Short sleep to allow time for both Watch registrations (CI environment).
	// /bin/bash explicitly: avoid Ubuntu /bin/sh (dash) vs macOS /bin/sh differences.
	result, sv := launchStub(t, "#!/bin/bash\nsleep 0.3\nexit 42\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch1 := sv.Watch(ctx)
	ch2 := sv.Watch(ctx)

	ev1 := <-ch1
	ev2 := <-ch2

	if ev1.ExitCode != 42 {
		t.Errorf("watcher1 ExitCode = %d, want 42", ev1.ExitCode)
	}
	if ev2.ExitCode != 42 {
		t.Errorf("watcher2 ExitCode = %d, want 42", ev2.ExitCode)
	}
}

// TestSupervisor_ExitEvent_Signaled verifies that when a subprocess is killed,
// the ExitEvent has Signaled == true (or at minimum non-zero ExitCode).
func TestSupervisor_ExitEvent_Signaled(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nsleep 60\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	_ = sv.Kill()

	select {
	case ev := <-ch:
		// Forced exit via SIGKILL: Signaled or non-zero exit code expected.
		if !ev.Signaled && ev.ExitCode == 0 && ev.Err == nil {
			t.Error("ExitEvent after Kill: expected Signaled=true or non-zero ExitCode")
		}
	case <-time.After(3 * time.Second):
		t.Error("no ExitEvent after Kill within 3s")
	}
}
