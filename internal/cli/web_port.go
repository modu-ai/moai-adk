package cli

// @MX:NOTE: [AUTO] web_port.go owns the safety logic that reclaims the target
// port just before `moai web` starts. Core invariant: only when the process
// holding the port is moai does it reclaim via SIGTERM; an external (non-moai)
// process is never killed and an explicit error is returned instead.
// findPortHolder / killProcess / checkPortInUse follow the same package-var
// indirection pattern as findProjectRootFn so tests can substitute fakes
// (platform-syscall-free testing).

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// moaiProcessName is the token searched for in the ps command name when deciding
// whether the port-holding process is a reclamation target (our own instance).
// No inline strings (§14 hardcoding rule).
const moaiProcessName = "moai"

// Port-reclamation polling parameters. After SIGTERM, waits up to
// portPollAttempts × portPollInterval (default 30 × 100ms ≈ 3s) for the port
// to be released. Kept as package-vars so tests can shorten the polling.
var (
	portPollAttempts = 30
	portPollInterval = 100 * time.Millisecond
)

// Test injection points (mirrors findProjectRootFn). The real implementations
// live in the platform-specific files (web_port_posix.go /
// web_port_windows.go) and in isPortInUse below.
var (
	findPortHolder = findPortHolderImpl
	killProcess    = killProcessImpl
	checkPortInUse = isPortInUse
)

// isPortInUse probes whether the port is occupied by attempting a 127.0.0.1:port
// bind. On a successful bind it closes immediately and returns false (free); on
// "address already in use" it returns true (held). Any other bind failure
// (permissions, etc.) is NOT treated as a reclamation signal and returns false,
// delegating the actual bind error to web.Run so it surfaces normally.
func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		// isAddrInUse is platform-split: Windows reports WSAEADDRINUSE (10048),
		// which is a DIFFERENT errno value from the POSIX syscall.EADDRINUSE
		// (48) constant, so a single errors.Is against the POSIX constant
		// silently misses every Windows conflict. The string fallback keeps the
		// historical unix behavior for wrapped errors that lose the errno — and
		// it only ever matches the unix message.
		return isAddrInUse(err) ||
			strings.Contains(err.Error(), "address already in use")
	}
	_ = ln.Close()
	return false
}

// ensurePortFree secures the target port just before delegating to web.Run.
//
// Order:
//  1. If reuse=false, skip reclamation and return nil (preserve the existing
//     fail-on-conflict behavior).
//  2. If the port is free, return nil.
//  3. Look up the holding process. On lookup failure (windows / no lsof /
//     TOCTOU) return nil without a hard failure so web.Run surfaces the normal
//     bind error.
//  4. If it is an external (non-moai) process, do not kill it; return an error
//     (pointing the user at --port).
//  5. If it is a moai process, log to stderr, send SIGTERM, and poll until the
//     port is released.
func ensurePortFree(errOut io.Writer, port int, reuse bool) error {
	if !reuse {
		return nil
	}
	if !checkPortInUse(port) {
		return nil
	}

	pid, isMoai, err := findPortHolder(port)
	if err != nil {
		// Holder could not be identified → cannot reclaim. Proceed and let
		// web.Run handle it.
		return nil
	}
	if !isMoai {
		return fmt.Errorf("port %d is held by a non-moai process (PID %d); specify a different port via --port", port, pid)
	}

	_, _ = fmt.Fprintf(errOut, "Stopping existing moai web instance (PID %d) and reusing port %d\n", pid, port)
	if err := killProcess(pid); err != nil {
		return fmt.Errorf("failed to stop existing moai instance (PID %d): %w", pid, err)
	}

	for i := 0; i < portPollAttempts; i++ {
		if !checkPortInUse(port) {
			return nil
		}
		time.Sleep(portPollInterval)
	}
	return fmt.Errorf("port %d still in use after SIGTERM (PID %d)", port, pid)
}
