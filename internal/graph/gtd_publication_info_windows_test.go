package graph

import (
	"errors"
	"io"
	"os"

	"golang.org/x/sys/windows"
)

func privateGTDTestFileInfo(path string) (os.FileInfo, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(h), path)
	info, statErr := f.Stat()
	return info, errors.Join(statErr, f.Close())
}

// The observer shares DELETE with the retained cleanup handle. It still reads
// the complete requested path, and both read and close failures are reported.
func privateGTDTestReadFile(path string) ([]byte, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(h), path)
	bytes, readErr := io.ReadAll(f)
	return bytes, errors.Join(readErr, f.Close())
}
