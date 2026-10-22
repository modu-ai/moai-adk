//go:build windows

package toolpolicy

import (
	"errors"
	"log/slog"

	"golang.org/x/sys/windows"
)

// pidProvablyDead uses a native process query; signal 0 is not supported on
// Windows. Only a nonexistent process or a read terminal exit code proves
// death. Access-denied, unrepresentable IDs and unknown observations retain
// the holder. The caller's existing retry and stale-age bounds are unchanged.
func pidProvablyDead(pid int) (dead bool) {
	if pid <= 0 || uint64(pid) > uint64(^uint32(0)) {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return errors.Is(err, windows.ERROR_INVALID_PARAMETER)
	}
	defer func() {
		if closeErr := windows.CloseHandle(h); closeErr != nil {
			slog.Warn("failed to close tool-policy process probe handle", "error", closeErr)
			dead = false
		}
	}()
	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	const stillActive = 259
	return code != stillActive
}
