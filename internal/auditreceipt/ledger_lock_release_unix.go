//go:build !windows

package auditreceipt

import "os"

func openLedgerLockForRelease(path string) (*os.File, error) { return os.Open(path) }

// Preserve token-aware POSIX release while refusing a pathname already replaced
// since the read. POSIX unlink remains path-based; Windows uses its held handle.
func removeOpenedLedgerLock(f *os.File) error {
	opened, err := f.Stat()
	if err != nil {
		return err
	}
	current, err := os.Stat(f.Name())
	if err != nil {
		return err
	}
	if !os.SameFile(opened, current) {
		return nil
	}
	return os.Remove(f.Name())
}
