//go:build windows

package skillresidue

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Both names are resolved by the kernel against held directory handles. A
// moved directory remains the same destination, including during a put-back.
const renameIsDescriptorBound = true

// openMoveEntry opens one entry relative to a held directory, without following
// a terminal reparse point. Sharing delete lets another actor move the held
// directories; containment must come from their identities, not denying moves.
func openMoveEntry(dir *os.File, name string, options uint32) (windows.Handle, error) {
	objectName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, err
	}
	attrs := windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(dir.Fd()), ObjectName: objectName, Attributes: windows.OBJ_CASE_INSENSITIVE}
	attrs.Length = uint32(unsafe.Sizeof(attrs))
	var handle windows.Handle
	err = windows.NtCreateFile(&handle, windows.DELETE|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE, &attrs, &windows.IO_STATUS_BLOCK{}, nil, 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN,
		windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT|options, 0, 0)
	runtime.KeepAlive(dir)
	return handle, moveError(err)
}

func moveError(err error) error {
	if status, ok := err.(windows.NTStatus); ok {
		return status.Errno()
	}
	return err
}

// noReplaceRenameAt uses FILE_RENAME_INFORMATION with ReplaceIfExists false.
// RootDirectory is the held destination handle; no absolute path is re-resolved
// after the source/destination folders have been opened through os.Root.
func noReplaceRenameAt(srcDir *os.File, src string, dstDir *os.File, dst string) error {
	h, err := openMoveEntry(srcDir, src, windows.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	buf, err := renameInformationBuffer(windows.Handle(dstDir.Fd()), dst)
	if err != nil {
		return err
	}
	err = windows.NtSetInformationFile(h, &windows.IO_STATUS_BLOCK{}, &buf[0], uint32(len(buf)), windows.FileRenameInformation)
	runtime.KeepAlive(dstDir)
	err = moveError(err)
	if errors.Is(err, windows.ERROR_NOT_SUPPORTED) || errors.Is(err, windows.ERROR_INVALID_FUNCTION) {
		return fmt.Errorf("%w: %v", errNoReplaceUnsupported, err)
	}
	return err
}

type renameInformation struct {
	ReplaceIfExists uint32
	RootDirectory   windows.Handle
	FileNameLength  uint32
	FileName        [1]uint16
}

// The NT buffer contract includes sizeof(FILE_RENAME_INFORMATION) plus the
// filename bytes. Offsetof(FileName)+bytes alone drops native trailing padding
// and is too short for one-character destinations (22 < 24 on Windows/amd64).
func renameInformationBuffer(root windows.Handle, dst string) ([]byte, error) {
	name, err := windows.UTF16FromString(dst)
	if err != nil {
		return nil, err
	}
	nameBytes := (len(name) - 1) * 2
	buf := make([]byte, int(unsafe.Sizeof(renameInformation{}))+nameBytes)
	info := (*renameInformation)(unsafe.Pointer(&buf[0]))
	info.RootDirectory = root
	info.FileNameLength = uint32(nameBytes)
	copy(unsafe.Slice(&info.FileName[0], len(name)-1), name[:len(name)-1])
	return buf, nil
}

// removeDirAt marks only the opened, empty directory for deletion. The handle
// is relative to dir and a reparse-point replacement is refused, never followed.
func removeDirAt(dir *os.File, name string) error {
	h, err := openMoveEntry(dir, name, windows.FILE_DIRECTORY_FILE)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("refusing reparse-point staging directory %s", name)
	}
	disposition := byte(1)
	return moveError(windows.NtSetInformationFile(h, &windows.IO_STATUS_BLOCK{}, &disposition, 1, windows.FileDispositionInformation))
}
