//go:build unix

package hook

// zone_user_read_unix.go — the unix half of the zone's non-blocking
// regular-file read (SPEC-USERASSET-DEPLOY-GUARD-001, gate round 30): a
// FIFO at the user-assets manifest path must not hang the hook. The open
// is NON-BLOCKING (an O_NONBLOCK open of a FIFO returns immediately), the
// OPEN HANDLE is fstat'd — regular files only — and the bytes are read
// from the same handle (the gate-23 handle-binding pattern).

import (
	"io"
	"os"
	"syscall"
)

// readRegularFile reads path as a regular file without ever blocking on a
// non-regular object. ok is false for absent paths, non-regular files, and
// read errors — callers treat !ok as "no evidence" (the over-protection
// guard's fail direction).
func readRegularFile(path string) (data []byte, ok bool) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	info, statErr := f.Stat()
	if statErr != nil || !info.Mode().IsRegular() {
		return nil, false
	}
	data, readErr := io.ReadAll(f)
	if readErr != nil {
		return nil, false
	}
	return data, true
}
