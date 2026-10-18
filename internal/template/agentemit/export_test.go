// export_test.go — test-only access to the unexported write seam of sets.go.
package agentemit

// SetWriteHook installs f as the seam called at the named stages of the atomic
// write ("before-temp", "before-commit") and returns the function restoring
// the previous seam.
func SetWriteHook(f func(stage string)) (restore func()) {
	old := testHook
	testHook = f
	return func() { testHook = old }
}
