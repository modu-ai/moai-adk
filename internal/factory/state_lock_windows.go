//go:build windows

// state_lock_windows.go — Windows substrate of the state lock:
// atomic-create (O_CREATE|O_EXCL), mirroring internal/spec/lock_windows.go's
// pattern. Windows lacks fcntl-style advisory flock, so the artifact IS the
// lock: a holder killed mid-mutation leaves an artifact that blocks every
// subsequent mutation permanently — which is why REQ-KB-023 records
// the owner identity and provides the bounded clear.
package factory

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"golang.org/x/sys/windows"
)

const (
	// Transient delete-pending / sharing-violation budget, mirroring
	// internal/spec/lock_windows.go's classification: genuine contention is
	// reported as held immediately and NEVER retried; only short-lived
	// Windows conditions (a just-released holder's delete-pending name, an
	// indexer/scanner sharing conflict) consume this budget.
	boardLockTransientRetries = 10
	boardLockTransientDelay   = 5 * time.Millisecond
)

// atomicFileStateLock holds an exclusively-created lock file; releasing means
// closing and removing it.
type atomicFileStateLock struct {
	lockPath string
	file     *os.File
}

func (f *atomicFileStateLock) release() error {
	if f == nil {
		return nil
	}
	if f.file != nil {
		_ = f.file.Close()
		f.file = nil
	}
	if f.lockPath != "" {
		err := atomicfile.RemoveFile(f.lockPath)
		f.lockPath = ""
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// acquireStateLockImpl creates lockPath with O_CREATE|O_EXCL — atomic on
// NTFS — and records this process's identity IN the artifact (REQ-KB-023).
//
// Before the atomic create it refuses a symlinked ancestor (up to `.moai`) and a
// pre-existing symlink or non-regular artifact, wrapping ErrStateLockUnsafePath.
// A pre-existing regular file is contention and still reports ErrStateLockHeld.
func acquireStateLockImpl(lockPath string) (stateLockImpl, error) {
	for attempt := 0; ; attempt++ {
		err := checkStateLockAncestors(lockPath)
		if err == nil {
			err = checkStateLockArtifact(lockPath)
		}
		if err == nil {
			break
		}
		if (!errors.Is(err, os.ErrPermission) && !errors.Is(err, windows.ERROR_SHARING_VIOLATION)) || attempt+1 >= boardLockTransientRetries {
			return nil, err
		}
		time.Sleep(boardLockTransientDelay)
	}
	transientLeft := boardLockTransientRetries
	for {
		file, err := atomicfile.CreateExclusive(lockPath, 0o644)
		if err == nil {
			if _, werr := file.Write(newLockOwnerRecord()); werr != nil {
				_ = file.Close()
				_ = atomicfile.RemoveFile(lockPath)
				return nil, fmt.Errorf("record board lock owner %s: %w", lockPath, werr)
			}
			return &atomicFileStateLock{lockPath: lockPath, file: file}, nil
		}
		switch {
		case os.IsExist(err):
			return nil, ErrStateLockHeld
		case errors.Is(err, os.ErrPermission):
			if transientLeft--; transientLeft <= 0 {
				return nil, fmt.Errorf("open board lock %s: %w", lockPath, err)
			}
			time.Sleep(boardLockTransientDelay)
		default:
			return nil, fmt.Errorf("open board lock %s: %w", lockPath, err)
		}
	}
}
