//go:build darwin || freebsd || netbsd || openbsd || dragonfly

// proc_info_bsd.go — ancestry lookup via the BSD/darwin process sysctl.
package session

import "golang.org/x/sys/unix"

// platformProcInfo reads kern.proc.pid for the given PID. P_comm is a
// fixed-width, NUL-padded field holding the truncated executable name.
func platformProcInfo(pid int) (int, string, bool) {
	if pid <= 0 {
		return 0, "", false
	}
	kp, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil || kp == nil {
		return 0, "", false
	}
	raw := kp.Proc.P_comm
	name := make([]byte, 0, len(raw))
	for _, c := range raw {
		if c == 0 {
			break
		}
		name = append(name, byte(c))
	}
	return int(kp.Eproc.Ppid), string(name), true
}
