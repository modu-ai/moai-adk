//go:build unix

package cli

// init_resume_read_unix.go — the unix half of the resume path's manifest
// read (gate round 36): a FIFO at the user-assets manifest path must not
// hang the resume. The open is NON-BLOCKING, the OPEN HANDLE is fstat'd —
// regular files only — and the bytes are read from the same handle (the
// gate-23 handle-binding pattern).

import (
	"io"
	"os"
	"syscall"
)

// readManifestRecord reads the user manifest without ever blocking on a
// non-regular object. ok is false for absent paths, non-regular files, and
// read errors — callers treat !ok as "no record to quarantine".
func readManifestRecord(path string) (data []byte, ok bool) {
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
