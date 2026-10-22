//go:build windows

package auditreceipt

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func openLedgerLockForRelease(path string) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(ledgerLockNativePath(path))
	if err != nil {
		return nil, err
	}
	var h windows.Handle
	err = retryLedgerReleaseSharing(func() error {
		var openErr error
		h, openErr = windows.CreateFile(p, windows.GENERIC_READ|windows.DELETE,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
			nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
		return openErr
	})
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(h), path), nil
}

// Mark only the object whose token was read. FILE_SHARE_DELETE lets stale
// reclamation rename that object, but never redirects this handle to a new
// claim at the old name. Close applies the disposition to this object only.
//
// @MX:WARN: [AUTO] ANCHOR breach for security: remove the opened claim, never a reused name.
// @MX:REASON: A stale ledger holder can finish while its successor publishes a different claim.
func removeOpenedLedgerLock(f *os.File) error {
	remove := byte(1) // FILE_DISPOSITION_INFO.DeleteFile is a BOOLEAN.
	err := retryLedgerReleaseSharing(func() error {
		return windows.SetFileInformationByHandle(windows.Handle(f.Fd()), windows.FileDispositionInfo, &remove, 1)
	})
	if err != nil {
		return &os.PathError{Op: "remove", Path: f.Name(), Err: err}
	}
	return nil
}

// Match Go's extended-path behavior for long runner paths, preserving explicit
// device paths and UNC shares rather than interpreting them as drive paths.
func ledgerLockNativePath(path string) string {
	if len(path) < 248 || strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\\.\`) {
		return path
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if strings.HasPrefix(absolute, `\\`) {
		return `\\?\UNC\` + absolute[2:]
	}
	return `\\?\` + absolute
}

// Preserve the claim substrate's bounded 20-attempt, 5ms sharing-conflict
// budget. Removal retries keep the same opened identity throughout.
func retryLedgerReleaseSharing(operation func() error) error {
	for attempt := 0; ; attempt++ {
		err := operation()
		if err == nil || attempt+1 >= 20 ||
			(!errors.Is(err, windows.ERROR_ACCESS_DENIED) && !errors.Is(err, windows.ERROR_SHARING_VIOLATION)) {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
}
