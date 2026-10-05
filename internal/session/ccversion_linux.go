//go:build linux

// ccversion_linux.go — the running-version platform probe: the
// /proc/<pid>/exe link target (SPEC-SESSION-CC-VERSION-001 REQ-SCV-001).
// The parse of the returned path is shared (ccversion.go), where the anchor
// duty lives: the path must name the claude binary itself, and a version
// segment elsewhere must not satisfy the read.
package session

import (
	"os"
	"strconv"
)

// platformReadProcessMapping reads the process's executable path. Failure —
// the pid dead, the link unreadable, a permission boundary — is ok=false: the
// caller renders unknown, never an error.
func platformReadProcessMapping(pid int) (string, bool) {
	if pid <= 0 {
		return "", false
	}
	target, err := os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe")
	if err != nil {
		return "", false
	}
	return target, true
}
