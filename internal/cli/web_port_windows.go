//go:build windows

// web_port_windows.go — Windows port-holder lookup + process termination stubs.
//
// lsof/ps are POSIX-only and syscall.Kill (SIGTERM) does not compile on Windows.
// Both functions return an "unsupported" error, and ensurePortFree treats a
// finder error not as a hard failure but as "cannot reclaim → proceed". So the
// build still succeeds under GOOS=windows and web.Run surfaces the normal bind
// error exactly as it does today.

package cli

import (
	"errors"

	"golang.org/x/sys/windows"
)

// findPortHolderImpl is unsupported on Windows. It returns an error so
// ensurePortFree skips reclamation and delegates to web.Run.
func findPortHolderImpl(_ int) (int, bool, error) {
	return 0, false, errors.New("port holder lookup not supported on windows")
}

// killProcessImpl is unsupported on Windows. ensurePortFree cannot identify a
// moai holder (findPortHolderImpl error), so this path is never reached; an
// explicit unsupported error is returned for interface completeness.
func killProcessImpl(_ int) error {
	return errors.New("process termination not supported on windows")
}

// isAddrInUse reports whether a bind failure is "address already in use".
//
// Winsock reports WSAEADDRINUSE (10048), NOT the POSIX syscall.EADDRINUSE (48)
// constant that also exists on Windows with an unrelated value — matching only
// the POSIX constant makes every Windows port conflict read as "port free".
// The Windows error message ("Only one usage of each socket address ... is
// normally permitted") does not contain the unix "address already in use"
// string either, so the caller's string fallback cannot cover this.
func isAddrInUse(err error) bool {
	return errors.Is(err, windows.WSAEADDRINUSE)
}
