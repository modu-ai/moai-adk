//go:build windows

package session

// Preserve the caller's spelling for registry storage; only the anchor's
// containment comparison folds case on Windows.
func realpathPlatform(string) string { return "" }

// Windows default volumes fold case. Treating a casing alias conservatively
// as an anchor also protects live sessions if their descendant has vanished.
func caseFoldsPlatform() bool { return true }
