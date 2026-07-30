//go:build !windows

// web_port_posix.go — POSIX (linux/darwin/...) port-holder lookup + process
// termination implementation.
//
// findPortHolderImpl obtains the PID of the LISTEN socket via lsof and looks up
// the command name via ps to decide whether that process is moai.
// killProcessImpl sends SIGTERM. syscall.Kill does not compile on Windows, so
// it is split out via a build tag (same separation reason as
// update_cleanup_unix.go).

package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// findPortHolderImpl returns the PID of the process LISTENing on port and
// whether it is moai. lsof -nP -iTCP:<port> -sTCP:LISTEN -t → PIDs (one per
// line) → ps -o comm= on the first PID yields the command name, which is
// checked for the moai token. If nobody is holding it or the lookup fails, an
// error is returned so the caller (ensurePortFree) skips reclamation and
// proceeds.
func findPortHolderImpl(port int) (int, bool, error) {
	out, err := exec.Command("lsof", "-nP", fmt.Sprintf("-iTCP:%d", port), "-sTCP:LISTEN", "-t").Output()
	if err != nil {
		// lsof exits 1 on no match → lands here.
		return 0, false, fmt.Errorf("failed to look up holder for port %d (lsof): %w", port, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0, false, fmt.Errorf("no process listening on port %d", port)
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse lsof PID %q: %w", fields[0], err)
	}

	comm, err := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return pid, false, fmt.Errorf("failed to look up command name for PID %d (ps): %w", pid, err)
	}
	isMoai := strings.Contains(string(comm), moaiProcessName)
	return pid, isMoai, nil
}

// @MX:WARN: [AUTO] killProcessImpl sends SIGTERM to another process, terminating it.
// @MX:REASON: [AUTO] This function is only called after ensurePortFree has verified
// via findPortHolder that the target is moai — never reaching an external process
// is the safety contract.
func killProcessImpl(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

// isAddrInUse reports whether a bind failure is "address already in use".
//
// syscall.EADDRINUSE is the POSIX errno the kernel returns here.
func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
}
