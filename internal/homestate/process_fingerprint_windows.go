//go:build windows

package homestate

import (
	"fmt"
	"golang.org/x/sys/windows"
)

func platformProcessFingerprint(pid int) (string, bool) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return "", false
	}
	defer windows.CloseHandle(h)
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &created, &exited, &kernel, &user); err != nil {
		return "", false
	}
	return fmt.Sprintf("%d", created.Nanoseconds()), true
}
