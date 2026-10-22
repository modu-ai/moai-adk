//go:build windows

package cli

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// FILE_EXECUTE checks the current token's execute access without loading or
// executing the pin. OPEN_EXISTING never creates/truncates it; omitting
// OPEN_REPARSE_POINT keeps the existing Stat/launch symlink-following policy.
// See https://learn.microsoft.com/en-us/windows/win32/fileio/file-access-rights-constants.
func checkClaudeBinaryExecuteAccess(pin string, _ os.FileInfo) error {
	path, err := windows.UTF16PtrFromString(pin)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(path, windows.FILE_EXECUTE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return fmt.Errorf("grant execute permission to the current user or point the pin at an accessible executable file: %w", err)
	}
	if err := windows.CloseHandle(handle); err != nil {
		return fmt.Errorf("close execute permission check: %w", err)
	}
	return nil
}
