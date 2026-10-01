//go:build !darwin

package session

// realpathPlatform is a no-op off darwin: case-sensitive filesystems have no
// case axis to canonicalize, symlink resolution is canonicalCWD's EvalSymlinks
// fallback, and Windows case handling is unmeasured (t1293 does not guess).
func realpathPlatform(dir string) string {
	_ = dir
	return ""
}
