//go:build windows

package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestWindowsLockProcess(t *testing.T) {
	path := os.Getenv("MOAI_WINDOWS_LOCK_FIXTURE")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := lockFile(ctx, f); err != nil {
		t.Fatal(err)
	}
	defer unlockFile(f)
	fmt.Println("LOCKED")
	// The parent kills this process; this bound also prevents an orphan fixture.
	time.Sleep(10 * time.Second)
}

func TestWindowsLockCancelledBeforeAcquisition(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := lockFile(ctx, f); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled acquisition: %v", err)
	}
	if err := lockFile(context.Background(), f); err != nil {
		t.Fatalf("cancelled attempt retained ownership: %v", err)
	}
	unlockFile(f)
}

func TestWindowsLockContentionCancellationAndCrashRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestWindowsLockProcess$")
	cmd.Env = append(os.Environ(), "MOAI_WINDOWS_LOCK_FIXTURE="+path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		if !waited {
			_ = cmd.Wait()
		}
	})
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || line != "LOCKED\n" {
		t.Fatalf("bounded child handshake %q: %v", line, err)
	}
	blocked, stop := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer stop()
	if err := lockFile(blocked, f); !errors.Is(err, context.DeadlineExceeded) {
		if err == nil {
			unlockFile(f)
		}
		t.Fatalf("cross-process exclusion/cancellation: %v", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	waited = true
	if err == nil || cmd.ProcessState == nil || cmd.ProcessState.Success() {
		t.Fatalf("fixture did not terminate abnormally: %v", err)
	}
	retry, stopRetry := context.WithTimeout(context.Background(), time.Second)
	defer stopRetry()
	if err := lockFile(retry, f); err != nil {
		t.Fatalf("dead process retained lock: %v", err)
	}
	unlockFile(f)
}
