//go:build !darwin && !windows

package session

// realpathPlatform is a no-op on case-sensitive Unix systems; symlink
// resolution remains canonicalCWD's EvalSymlinks fallback.
func realpathPlatform(dir string) string {
	_ = dir
	return ""
}

// caseFoldsPlatform keeps Unix containment literal on case-sensitive volumes.
func caseFoldsPlatform() bool {
	return false
}
