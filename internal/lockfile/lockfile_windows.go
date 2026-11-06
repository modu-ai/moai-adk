//go:build windows

package lockfile

import (
	"errors"
	"os"
	"sync"
)

// fileLocksMu guards concurrent access to the fileLocks map.
var fileLocksMu sync.Mutex

// @MX:NOTE: [AUTO] Windows in-process-mutex limitation preserved verbatim from the
// pre-migration internal/cli/team_spawn_lock_windows.go (SPEC-AGENT-TEAM-RETIRE-001
// REQ-ATR-001 — behavior preservation is the contract; do NOT silently "upgrade"
// this to LockFileEx).
//
// fileLocks holds in-process mutexes keyed by absolute file path.
// Windows lacks portable advisory file locks (no fcntl/flock equivalent in stdlib),
// so we fall back to process-local mutexes. This means:
//   - Concurrent writes within the SAME process are serialized (safe for tests and
//     tmux teammates that run as separate processes within the same OS user session)
//   - Concurrent writes across DIFFERENT OS processes are NOT protected
//     (acceptable limitation: ClaimTask is primarily exercised by tmux-based team
//     workflows, which are macOS/Linux-only; Windows users run solo mode)
var fileLocks = map[string]*sync.Mutex{}

// Lock acquires a process-local mutex for the given file on Windows.
// Multi-process locking is not supported; see package comment above.
func Lock(f *os.File) error {
	fileLocksMu.Lock()
	path := f.Name()
	mu, ok := fileLocks[path]
	if !ok {
		mu = &sync.Mutex{}
		fileLocks[path] = mu
	}
	fileLocksMu.Unlock()
	mu.Lock()
	return nil
}

// TryLock acquires the process-local mutex for the given file without
// blocking, returning an error when another holder owns it
// (SPEC-WORKTREE-HOME-ROOT-001 M5, design.md D-12: the state-file lock is
// try-and-bounded-wait). Nothing else about the lockfile surface changes.
func TryLock(f *os.File) error {
	fileLocksMu.Lock()
	path := f.Name()
	mu, ok := fileLocks[path]
	if !ok {
		mu = &sync.Mutex{}
		fileLocks[path] = mu
	}
	fileLocksMu.Unlock()
	if mu.TryLock() {
		return nil
	}
	return errLockHeld
}

// errLockHeld reports a try-acquisition refused because another holder owns
// the lock (Windows TryLock arm).
var errLockHeld = errors.New("lock already held")

// Unlock releases the process-local mutex for the given file on Windows.
func Unlock(f *os.File) error {
	fileLocksMu.Lock()
	mu, ok := fileLocks[f.Name()]
	fileLocksMu.Unlock()
	if !ok {
		return nil // never locked or already released
	}
	mu.Unlock()
	return nil
}
