//go:build windows

package harness

// entryOwnedByCurrentUser reports whether the entry at path is owned by the current user. Per-user
// file ownership is not read on Windows, so the owner is never determinable and every entry counts as
// not owned: a faulty state-path entry is left in place and retention stays off with a warning.
func entryOwnedByCurrentUser(string) bool { return false }
