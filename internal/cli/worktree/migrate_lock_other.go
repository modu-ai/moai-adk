//go:build !windows

package worktree

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// tryMigrationLock takes the migration lock without blocking: an OS flock on
// the open lock file, released when the file is closed or the process dies.
func tryMigrationLock(f *os.File) error {
	return lockfile.TryLock(f)
}
