package kanban

import "time"

// AcquireFactoryStepLock takes the cross-process lock that serializes the
// worktree step of a factory lease (SPEC-FACTORY-ATOMIC-LEASE-001 plan D3),
// waiting at most wait, and returns the function that releases it.
//
// Compile-only stub (WM1): it takes no lock, returns a non-nil no-op release
// (a test that calls it must never meet a nil function) and the sentinel error
// until milestone WM5 replaces it.
func AcquireFactoryStepLock(root string, wait time.Duration) (release func() error, err error) {
	return func() error { return nil }, errWithLockNotImplemented
}
