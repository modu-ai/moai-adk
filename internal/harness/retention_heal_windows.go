//go:build windows

package harness

// acquireHealLock is a no-op on Windows: it creates no file and returns a release that does nothing.
// The heal lock gives no exclusion on Windows. It is never reached there, because the Windows owner
// check refuses every entry, so a faulty state-path entry is left in place with a warning and no heal
// runs. The Windows runtime is not observed.
func acquireHealLock(_ string, _ func(string) bool) (func(), error) {
	return func() {}, nil
}
