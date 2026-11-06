//go:build windows

package worktree

import (
	"os"

	"golang.org/x/sys/windows"
)

// tryMigrationLock takes the migration lock without blocking, across
// processes (card t2053, C11). internal/lockfile's Windows arm is a
// process-local mutex by contract, so two `moai` processes would both "hold"
// the migration lock; LockFileEx on the open lock file is the cross-process
// equivalent of the flock the other platforms use. The lock is released when
// the file handle is closed or the process dies.
func tryMigrationLock(f *os.File) error {
	var overlapped windows.Overlapped
	return windows.LockFileEx(windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0, &overlapped)
}
