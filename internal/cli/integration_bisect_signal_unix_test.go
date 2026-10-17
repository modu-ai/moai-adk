//go:build !windows

package cli

// card t1833 fix round 2 (N1): an interrupt must not leave the card worktree
// detached mid-bisect. The verb runs in a helper subprocess (this test binary
// re-executed) so a real signal can be delivered to a real process.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const bisectHelperEnv = "MOAI_BISECT_SIGNAL_HELPER"

// TestBisectSignalHelper is the child side: it runs the verb and exits with a
// distinct status, so the parent can tell "verb returned" from "killed by the signal".
func TestBisectSignalHelper(t *testing.T) {
	spec := os.Getenv(bisectHelperEnv)
	if spec == "" {
		t.Skip("helper process for TestBisectRestoresTheTreeOnSignal")
	}
	parts := strings.Split(spec, "\x1f")
	err := runIntegrationBisect(os.Stdout, integrationBisectOpts{Dir: parts[0], Good: parts[1], Bad: parts[2], Command: parts[3]})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	os.Exit(0)
}

func TestBisectRestoresTheTreeOnSignal(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			_, wt, base, tip := bisectRepo(t)
			scratch := t.TempDir()
			marker, pidFile := filepath.Join(scratch, "started"), filepath.Join(scratch, "judged.pid")
			command := fmt.Sprintf("echo $$ > %s && touch %s && exec sleep 120", pidFile, marker)

			cmd := exec.Command(os.Args[0], "-test.run=^TestBisectSignalHelper$", "-test.v")
			cmd.Env = append(os.Environ(), bisectHelperEnv+"="+strings.Join([]string{wt, base, tip, command}, "\x1f"))
			var out bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &out
			cmd.WaitDelay = 3 * time.Second // an orphaned judged command must not hold the pipe open
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			t.Cleanup(func() { _ = cmd.Process.Kill() })

			deadline := time.Now().Add(60 * time.Second)
			for {
				if _, err := os.Stat(marker); err == nil {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("the judged command never started:\n%s", out.String())
				}
				time.Sleep(50 * time.Millisecond)
			}
			// Mid-bisect now: HEAD is on an intermediate commit.
			if head := fcGit(t, wt, "rev-parse", "HEAD"); head == tip {
				t.Fatalf("test setup: the bisect has not moved HEAD, so the signal would prove nothing")
			}
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 3 {
					t.Fatalf("the verb must return an error (exit 3) after %s, got %v\n%s", sig, err, out.String())
				}
			case <-time.After(30 * time.Second):
				t.Fatalf("the verb did not exit after %s:\n%s", sig, out.String())
			}

			if !strings.Contains(out.String(), "interrupted") || !strings.Contains(out.String(), "restored") {
				t.Errorf("the error must say the search was interrupted and the tree restored:\n%s", out.String())
			}
			if head := fcGit(t, wt, "rev-parse", "HEAD"); head != tip {
				t.Errorf("HEAD must be back at %s after %s, is %s", tip, sig, head)
			}
			if branch, _ := runScrubbedGit(wt, "symbolic-ref", "--quiet", "--short", "HEAD"); strings.TrimSpace(branch) != "" {
				t.Errorf("the starting HEAD was detached, now on branch %q", branch)
			}
			if bisectInProgress(wt) {
				t.Errorf("BISECT_LOG is still present after %s", sig)
			}
			raw, err := os.ReadFile(pidFile)
			if err != nil {
				t.Fatal(err)
			}
			pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
			gone := false
			for i := 0; i < 100 && !gone; i++ {
				gone = syscall.Kill(pid, 0) != nil
				time.Sleep(50 * time.Millisecond)
			}
			if !gone {
				_ = syscall.Kill(pid, syscall.SIGKILL)
				t.Errorf("the judged command (pid %d) outlived the interrupted search", pid)
			}
		})
	}
}
