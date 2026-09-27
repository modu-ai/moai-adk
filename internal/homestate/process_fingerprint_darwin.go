//go:build darwin

package homestate

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func platformProcessFingerprint(pid int) (string, bool) {
	if pid <= 0 {
		return "", false
	}
	info, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil || info == nil || int(info.Proc.P_pid) != pid {
		return "", false
	}
	started := info.Proc.P_starttime
	return fmt.Sprintf("%d.%06d", started.Sec, started.Usec), true
}
