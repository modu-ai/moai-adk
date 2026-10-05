//go:build windows

package cli

// memory_fold_wiring_fifo_windows_test.go — the windows twin of the blocked
// -read fixture: named pipes are not created here (plan B1), and the FIFO
// cells skip at runtime before ever reaching blockOnRead. This file only
// keeps the package compiling on windows.

import "testing"

// wireFIFO is the windows twin; no fixture is ever built on windows.
type wireFIFO struct{}

// release is never reached on windows (the cells skip first).
func (f *wireFIFO) release() {}

// blockOnRead is never reached on windows (the cells skip first).
func blockOnRead(t *testing.T, path string) *wireFIFO {
	t.Fatal("named-pipe fixtures are unix-only (AC-MFB-008 vii/viii/ix, plan B1)")
	return nil
}
