//go:build windows

package atomicfile

import (
	"errors"

	"golang.org/x/sys/windows"
)

var openProcessForLock = windows.OpenProcess

// Only ERROR_INVALID_PARAMETER or a successfully observed exit proves absence.
// Access denied and failed status reads retain the possibly-live owner's lock.
func PidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if uint64(pid) > uint64(^uint32(0)) {
		return true
	} // not a representable Windows identity
	h, err := openProcessForLock(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return !errors.Is(err, windows.ERROR_INVALID_PARAMETER)
	}
	defer windows.CloseHandle(h)
	var status uint32
	return windows.GetExitCodeProcess(h, &status) != nil || status == 259
}
