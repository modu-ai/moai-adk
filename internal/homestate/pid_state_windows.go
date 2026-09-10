//go:build windows

package homestate

import "golang.org/x/sys/windows"

func platformPIDState(pid int) ProcessIdentityState {
	if pid <= 0 {
		return ProcessIdentityDead
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err == windows.ERROR_INVALID_PARAMETER {
		return ProcessIdentityDead
	}
	if err != nil {
		return ProcessIdentityIndeterminate
	}
	_ = windows.CloseHandle(h)
	return ProcessIdentityLive
}
