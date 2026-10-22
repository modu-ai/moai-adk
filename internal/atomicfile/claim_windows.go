//go:build windows

package atomicfile

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

const claimTransientAttempts = 20
const claimTransientDelay = 5 * time.Millisecond

// CREATE_NEW remains the arbiter on every attempt. Delete-pending names may
// report access denied briefly after the prior holder removed the name; they
// are not evidence that its process is dead and never authorize breaking it.
func createExclusive(path string, perm os.FileMode) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(claimNativePath(path))
	if err != nil {
		return nil, err
	}
	attrs := uint32(windows.FILE_ATTRIBUTE_NORMAL)
	if perm&0o200 == 0 {
		attrs = windows.FILE_ATTRIBUTE_READONLY
	}
	for attempt := 0; ; attempt++ {
		h, err := windows.CreateFile(p, windows.GENERIC_WRITE,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
			nil, windows.CREATE_NEW, attrs, 0)
		if err == nil {
			return os.NewFile(uintptr(h), path), nil
		}
		if !isTransientReplaceError(err) || attempt+1 >= claimTransientAttempts {
			return nil, &os.PathError{Op: "open", Path: path, Err: err}
		}
		time.Sleep(claimTransientDelay)
	}
}

func removeFile(path string) error {
	for attempt := 0; ; attempt++ {
		err := os.Remove(path)
		if err == nil || !isTransientReplaceError(err) || attempt+1 >= claimTransientAttempts {
			return err
		}
		time.Sleep(claimTransientDelay)
	}
}

// Sharing delete prevents a verdict reader from blocking a concurrent owner's
// removal or replacement. The returned handle still names the opened file.
func openRead(path string) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(claimNativePath(path))
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(h), path), nil
}

// Match os.OpenFile's extended-path handling without changing device or short
// path semantics. Long state paths occur under Windows runner temp directories.
func claimNativePath(path string) string {
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
