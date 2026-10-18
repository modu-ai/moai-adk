//go:build windows

// process_probe_windows.go — card t1779: the Windows half of the
// sandbox-signature probe. There is no ps on Windows, so no sandbox signature
// is available there; the launch proceeds as before.

package homestate

// ProcessListProbeError reports nil on Windows: no ps probe exists, so the
// sandbox signature can never match and the gate stays open.
func ProcessListProbeError(_ int) error { return nil }
