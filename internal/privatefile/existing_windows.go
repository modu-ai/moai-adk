package privatefile

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// @MX:ANCHOR: [AUTO] Construct the same protected current-owner policy for native file and directory operations.
// @MX:REASON: Temporary creation, append creation, directory creation and existing-object narrowing share this policy.
func ownerDescriptor(directory bool) (*windows.SECURITY_DESCRIPTOR, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	sid := user.User.Sid.String()
	flags := ""
	if directory {
		flags = "OICI"
	}
	return windows.SecurityDescriptorFromString("O:" + sid + "D:P(A;" + flags + ";FA;;;" + sid + ")")
}

func openExisting(path string, access uint32, directory bool) (*os.File, error) {
	return openExistingWithShare(path, access, directory, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
}

func openExistingWithShare(path string, access uint32, directory bool, share uint32) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	flags := uint32(windows.FILE_FLAG_OPEN_REPARSE_POINT)
	if directory {
		flags |= windows.FILE_FLAG_BACKUP_SEMANTICS
	}
	h, err := windows.CreateFile(p, access, share,
		nil, windows.OPEN_EXISTING, flags, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(h), path), nil
}

func checkDirectoryNative(path string, before os.FileInfo) error {
	f, err := openExisting(path, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES, true)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(before, info) {
		return fmt.Errorf("private directory: identity changed")
	}
	return checkObjectHandle(windows.Handle(f.Fd()), true)
}

func secureExistingNative(path string, before os.FileInfo, directory, propagate bool) error {
	share := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE)
	if directory && !propagate {
		// SetSecurityInfo's documented exclusive-directory behavior suppresses
		// automatic propagation to existing children. Future inheritance and
		// the strict descriptor readback remain unchanged.
		share = 0
	}
	f, err := openExistingWithShare(path, windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES, directory, share)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(before, info) {
		return fmt.Errorf("private path: identity changed")
	}
	return restrictHandleNative(f, directory)
}

// @MX:WARN: [AUTO] Narrow an existing object only after its identity and current owner are known.
// @MX:REASON: Metadata, ownership, descriptor creation, DACL installation and native readback errors must all stop the operation.
func restrictHandleNative(f *os.File, directory bool) error {
	h := windows.Handle(f.Fd())
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory {
		return fmt.Errorf("private path: unexpected object type")
	}
	if !directory && info.NumberOfLinks != 1 {
		return fmt.Errorf("private path: refusing to change a file with %d links", info.NumberOfLinks)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return err
	}
	if owner == nil {
		return fmt.Errorf("private path: owner is unknown")
	}
	if !owner.Equals(user.User.Sid) {
		return fmt.Errorf("private path: owner %s differs from current user %s", owner.String(), user.User.Sid.String())
	}
	desired, err := ownerDescriptor(directory)
	if err != nil {
		return err
	}
	acl, _, err := desired.DACL()
	if err != nil {
		return err
	}
	// Only the existing owner's DACL is narrowed. Owner/group and token
	// privileges are never changed; an unsupported operation fails closed.
	if err := windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		return err
	}
	return checkObjectHandle(h, directory)
}

func openAppendNative(path string) (*os.File, error) {
	sd, err := ownerDescriptor(false)
	if err != nil {
		return nil, err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.FILE_APPEND_DATA|windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, &sa,
		windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(h), path), nil
}

func mkdirNative(path string) error {
	sd, err := ownerDescriptor(true)
	if err != nil {
		return err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	if err := windows.CreateDirectory(p, &sa); err != nil {
		return &os.PathError{Op: "mkdir", Path: path, Err: err}
	}
	return nil
}
