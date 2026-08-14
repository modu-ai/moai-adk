//go:build windows

// board_lock_windows.go — Windows substrate of the board-wide lock:
// atomic-create (O_CREATE|O_EXCL), mirroring internal/spec/lock_windows.go's
// pattern. Windows lacks fcntl-style advisory flock, so the artifact IS the
// lock: a holder killed mid-mutation leaves an artifact that blocks every
// subsequent board mutation permanently — which is why REQ-KB-023 records
// the owner identity and provides the bounded clear.
package kanban

import (
	"errors"
	"fmt"
	"os"
	"time"
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

// atomicFileBoardLock holds an exclusively-created lock file; releasing means
// closing and removing it.
type atomicFileBoardLock struct {
	lockPath string
	file     *os.File
}

func (f *atomicFileBoardLock) release() error {
	if f == nil {
		return nil
	}
	if f.file != nil {
		_ = f.file.Close()
		f.file = nil
	}
	if f.lockPath != "" {
		err := os.Remove(f.lockPath)
		f.lockPath = ""
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// acquireBoardLockImpl creates lockPath with O_CREATE|O_EXCL — atomic on
// NTFS — and records this process's identity IN the artifact (REQ-KB-023).
func acquireBoardLockImpl(lockPath string) (boardLockImpl, error) {
	transientLeft := boardLockTransientRetries
	for {
		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
		if err == nil {
			if _, werr := file.Write(newLockOwnerRecord()); werr != nil {
				_ = file.Close()
				_ = os.Remove(lockPath)
				return nil, fmt.Errorf("record board lock owner %s: %w", lockPath, werr)
			}
			return &atomicFileBoardLock{lockPath: lockPath, file: file}, nil
		}
		switch {
		case os.IsExist(err):
			return nil, ErrBoardLockHeld
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
