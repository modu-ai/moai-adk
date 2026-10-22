package privatefile

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// OpenPinned atomically opens/creates a private regular file and holds its
// deletion-sharing fence on that same handle. It does not truncate or write.
// FILE_READ_DATA participates in share checks; metadata-only access did not
// prevent SQLite sidecar deletion in the native diagnostic. WRITE_DAC permits
// narrowing an already-owned object, never changing its owner or privileges.
// acquired registers ownership immediately after CreateFile, before validation
// can fail or panic. The owner closes all acquired handles, including refused
// ones, and retains them if native cleanup cannot be confirmed.
// @MX:WARN: [AUTO] Keep atomic acquisition and every native policy refusal in one lifetime.
// @MX:REASON: Registration, identity, single-link, owner and DACL checks must precede SQLite writes.
func OpenPinned(path string, acquired func(io.Closer)) error {
	if acquired == nil {
		return fmt.Errorf("private pin: lifetime owner is required")
	}
	before, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if before != nil && !before.Mode().IsRegular() {
		return fmt.Errorf("private pin: not a regular file")
	}
	sd, err := ownerDescriptor(false)
	if err != nil {
		return err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	p, err := windows.UTF16PtrFromString(nativePath(path))
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(p, windows.FILE_READ_DATA|windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, &sa, windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(h), path)
	acquired(f)
	opened, err := f.Stat()
	if err == nil && (!opened.Mode().IsRegular() || (before != nil && !os.SameFile(before, opened))) {
		err = fmt.Errorf("private pin: identity changed")
	}
	var info windows.ByHandleFileInformation
	if err == nil {
		err = windows.GetFileInformationByHandle(h, &info)
	}
	if err == nil && info.NumberOfLinks != 1 {
		err = fmt.Errorf("private pin: object has %d links", info.NumberOfLinks)
	}
	if err == nil {
		err = restrictHandleNative(f, false)
	}
	return err
}
