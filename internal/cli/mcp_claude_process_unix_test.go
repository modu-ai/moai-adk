//go:build !windows

package cli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestClaudeAuditProcessHelper(t *testing.T) {
	if os.Getenv("MOAI_CLAUDE_PROCESS_HELPER") != "1" {
		return
	}
	child := exec.Command("sleep", "30")
	if err := child.Start(); err != nil {
		os.Exit(2)
	}
	pidFile := os.Getenv("MOAI_CLAUDE_CHILD_PID_FILE")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
		_ = child.Process.Kill()
		os.Exit(3)
	}
	_ = child.Wait()
	os.Exit(0)
}

func TestClaudeAuditProcessCancellationKillsDescendants_AC_CLA_014(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestClaudeAuditProcessHelper$")
	cmd.Env = append(os.Environ(),
		"MOAI_CLAUDE_PROCESS_HELPER=1",
		"MOAI_CLAUDE_CHILD_PID_FILE="+pidFile,
	)
	configureClaudeAuditProcess(cmd)
	done := make(chan error, 1)
	go func() { done <- runClaudeAuditProcess(cmd) }()

	var childPID int
	for childPID == 0 {
		data, err := os.ReadFile(pidFile)
		if err == nil {
			childPID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
		}
		select {
		case err := <-done:
			t.Fatalf("helper exited before cancellation: %v", err)
		case <-ctx.Done():
			t.Fatal("timed out waiting for helper child pid")
		case <-time.After(10 * time.Millisecond):
		}
	}
	t.Cleanup(func() { _ = syscall.Kill(childPID, syscall.SIGKILL) })

	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled process returned nil error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled Claude audit process did not exit")
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		err := syscall.Kill(childPID, 0)
		if errors.Is(err, syscall.ESRCH) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("descendant pid %d survived process-group cancellation: %v", childPID, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
