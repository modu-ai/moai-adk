//go:build !windows

package hook

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// readRuleFileBytes reads the rule file at path with the POSIX TOCTOU guard:
// the path is opened NON-BLOCKING first, the OPENED HANDLE is fstat-ed for
// regularity, and only a regular file is read (O_NONBLOCK cleared so the read
// itself blocks on data, not on the open). The Stat→ReadFile two-call shape
// has a replacement window — a path swapped to a FIFO (or a symlink turned to
// one) between the check and the read blocks os.ReadFile forever, hanging
// SessionStart with no rules, warning, or directive. Opening the final object
// and judging the handle closes that window: whatever the path became, the
// open either succeeds non-blocking on a regular file or is rejected.
func readRuleFileBytes(path string) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("opened-handle stat failed: %w", err)
	}
	if !fi.Mode().IsRegular() {
		return nil, errNotRegularFile
	}
	// Clear O_NONBLOCK: regular files ignore it, but leaving it set would
	// surface as EAGAIN on a read that legitimately blocks for data.
	if err := unix.SetNonblock(int(f.Fd()), false); err != nil {
		return nil, fmt.Errorf("clear non-blocking mode: %w", err)
	}
	return io.ReadAll(f)
}
