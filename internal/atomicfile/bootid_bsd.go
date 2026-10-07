//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package atomicfile

// The BSD family (darwin included) exposes the boot time as a sysctl. The
// raw bytes are stable for the boot's lifetime, which is all the identity
// needs.

import "syscall"

// currentBootID returns a boot-unique identity for this machine.
func currentBootID() string {
	raw, err := syscall.Sysctl("kern.boottime")
	if err != nil {
		return ""
	}
	return raw
}
