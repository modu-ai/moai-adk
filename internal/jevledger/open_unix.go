//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package jevledger

import (
	"os"
	"syscall"
)

// openLedger opens the ledger for appending, and only when the OPENED handle
// is a regular file.
//
// The flags carry the whole safety argument. O_NOFOLLOW makes a symbolic link
// at the path fail the open instead of appending to its target. O_NONBLOCK
// makes a FIFO with no reader fail at once (ENXIO) instead of blocking the
// hook process forever. A FIFO WITH a reader does open, so neither flag
// protects it: the handle check below does, and it is the one thing standing
// between the row and that pipe. It runs on the opened descriptor, never on
// the path, so no later step resolves the path again.
//
// @MX:WARN: [AUTO] Opens a path the project controls, which may have been replaced by a link, a pipe or a directory
// @MX:REASON: Dropping O_NOFOLLOW, O_NONBLOCK or the handle check makes a hook process block forever or write outside the project; each is covered by its own mutant in fifo_unix_test.go
func openLedger(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, fileMode)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, errNotRegular
	}
	return f, nil
}
