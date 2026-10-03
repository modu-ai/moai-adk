package kanban

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// factoryStepLockRelPath is the worktree-step lock's artifact beneath a
// project root. It is a lock of its own: not the queue's lock, not the
// leader's board.lock, not a `moai slot` lease (SPEC-FACTORY-ATOMIC-LEASE-001
// plan D3).
var factoryStepLockRelPath = filepath.Join(".moai", "state", "factory-worktree-step.lock")

// AcquireFactoryStepLock takes the cross-process lock that serializes the
// worktree step of a factory lease (SPEC-FACTORY-ATOMIC-LEASE-001 plan D3),
// waiting at most wait, and returns the function that releases it.
//
// The substrate is the board lock's path-parameterized implementation (flock
// on Unix, atomic-create on Windows) polled with the board lock's retry
// policy, against a separate lock file. On timeout the error wraps
// ErrBoardLockHeld and names the artifact; the returned release function is
// never nil, and releasing twice is safe.
//
// @MX:WARN: [AUTO] Cross-process lock serializing worktree creation and branch rename for every factory lane.
// @MX:REASON: a caller that holds it across a queue write or a record write inverts the lock order (queue, record, log); the step takes it around the creator and `git branch -m` only and releases before the record write (plan D3).
// @MX:SPEC: SPEC-FACTORY-ATOMIC-LEASE-001
func AcquireFactoryStepLock(root string, wait time.Duration) (release func() error, err error) {
	noop := func() error { return nil }
	path := filepath.Join(root, factoryStepLockRelPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return noop, fmt.Errorf("worktree step lock: creating state dir: %w", err)
	}
	deadline := time.Now().Add(wait)
	for attempt := 0; ; attempt++ {
		impl, err := acquireBoardLockImpl(path)
		if err == nil {
			held := &BoardLock{path: path, impl: impl}
			return held.Release, nil
		}
		if !IsBoardLockHeld(err) {
			return noop, fmt.Errorf("worktree step lock %s: %w", path, err)
		}
		if !time.Now().Before(deadline) {
			return noop, fmt.Errorf("worktree step lock %s not obtained within %s: %w", path, wait, err)
		}
		time.Sleep(boardLockRetryWait(attempt))
	}
}
