//go:build windows

package cli

import (
	"os"
	"sync"
)

// fileLocksMu guards concurrent access to the fileLocks map.
var fileLocksMu sync.Mutex

// fileLocks holds in-process mutexes keyed by absolute file path.
// Windows lacks portable advisory file locks (no fcntl/flock equivalent in stdlib),
// so we fall back to process-local mutexes. This means:
//   - Concurrent writes within the SAME process are serialized (safe for tests and
//     tmux teammates that run as separate processes within the same OS user session)
//   - Concurrent writes across DIFFERENT OS processes are NOT protected
//     (acceptable limitation: ClaimTask is primarily exercised by tmux-based team
//     workflows, which are macOS/Linux-only; Windows users run solo mode)
var fileLocks = map[string]*sync.Mutex{}

// lockFile acquires a process-local mutex for the given file on Windows.
// Multi-process locking is not supported; see package comment above.
func lockFile(f *os.File) error {
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

// unlockFile releases the process-local mutex for the given file on Windows.
func unlockFile(f *os.File) error {
	fileLocksMu.Lock()
	mu, ok := fileLocks[f.Name()]
	fileLocksMu.Unlock()
	if !ok {
		return nil // never locked or already released
	}
	mu.Unlock()
	return nil
}
