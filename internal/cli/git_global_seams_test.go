package cli

import "testing"

// isolateGitGlobalSeams replaces the four seams that run `git config --global`
// (safe.directory add, unset and get-all, and the global value read) with
// neutral stubs for the duration of the test. A test that reaches the session
// worktree materializer or cleanup, directly or through a command, calls it so
// no path from the test can read or write the global git config. Compose it
// with swapSessionWorktreeSeams, which does not cover these seams; tests that
// assert the safe.directory behavior pass their own stubs to swapM7Seams
// instead. Like swapM7Seams it mutates package state, so the test must not call
// t.Parallel.
func isolateGitGlobalSeams(t *testing.T) {
	t.Helper()
	swapM7Seams(t, m7Seams{
		safeDirAdd:    func(string) error { return nil },
		safeDirUnset:  func(string) error { return nil },
		safeDirGetAll: func() ([]string, error) { return nil, nil },
		globalGet:     func(string) string { return "" },
	})
}
