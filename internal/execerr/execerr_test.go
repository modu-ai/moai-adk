package execerr

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// realExitError runs a genuinely failing subprocess (the test binary itself)
// so the error carries a populated ProcessState, matching what wrap sites see.
func realExitError(t *testing.T) *exec.ExitError {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	err := cmd.Run()
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	return ee
}

// TestHelperProcess is the child target realExitError executes.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(3)
}

func TestStatusDetailExitError(t *testing.T) {
	ee := realExitError(t)
	got := StatusDetail(ee)
	// Run()-produced ExitError has no captured stderr → status line only.
	if !strings.Contains(got, "exited with status 3") {
		t.Fatalf("StatusDetail(exit error) = %q; want it to name the exit status", got)
	}
	// The result is plain text: no error chain is created to carry the raw
	// type upward (the seam-side guard is internal/cli ResolveExitCode).
	var raw *exec.ExitError
	if errors.As(errors.New(got), &raw) {
		t.Fatal("status detail text must not be an error chain")
	}
}

func TestStatusDetailWithCapturedStderr(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	// Output() captures the child's stderr into ExitError.Stderr.
	_, err := cmd.Output()
	if err == nil {
		t.Fatal("expected error from failing child")
	}
	if got := StatusDetail(err); !strings.Contains(got, "exited with status 3") {
		t.Fatalf("StatusDetail(output error) = %q; want it to name the exit status", got)
	}
}

func TestStatusDetailNonExitError(t *testing.T) {
	plain := errors.New("binary not found")
	if got := StatusDetail(plain); got != plain.Error() {
		t.Fatalf("StatusDetail(non-exit error) = %q; want the error text unchanged", got)
	}
	wrapped := fmt.Errorf("outer: %w", plain)
	if got := StatusDetail(wrapped); got != wrapped.Error() {
		t.Fatalf("StatusDetail(wrapped non-exit error) = %q; want the error text unchanged", got)
	}
}
