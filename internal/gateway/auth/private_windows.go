//go:build windows

package auth

// Native Windows ownership checks used by Store; permission bits alone cannot establish a DACL.
import (
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

func windowsPrivateDescriptor(directory bool) (*windows.SECURITY_DESCRIPTOR, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, ErrAuthState
	}
	sid := user.User.Sid.String()
	flags := ""
	if directory {
		flags = "OICI"
	}
	// Protected DACL: a single explicit current-user full-access ACE. Children
	// inherit only this user; an explicitly created file uses its own protected ACL.
	sd, err := windows.SecurityDescriptorFromString("O:" + sid + "D:P(A;" + flags + ";FA;;;" + sid + ")")
	if err != nil {
		return nil, ErrAuthState
	}
	return sd, nil
}

func createWindowsPrivateDirectory(path string) error {
	if !filepath.IsAbs(path) {
		return ErrAuthState
	}
	sd, err := windowsPrivateDescriptor(true)
	if err != nil {
		return err
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return ErrAuthState
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	// Existing directories are never adopted or chmod-repaired by this primitive.
	if err := windows.CreateDirectory(name, &sa); err != nil {
		return ErrAuthState
	}
	runtime.KeepAlive(sd)
	return validateWindowsPrivatePath(path, true)
}

func validateWindowsPrivatePath(path string, directory bool) error {
	if !filepath.IsAbs(path) {
		return ErrAuthState
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return ErrAuthState
	}
	h, err := windows.CreateFile(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return ErrAuthState
	}
	defer windows.CloseHandle(h)
	return validateWindowsPrivateHandle(h, directory)
}

func validateWindowsPrivateHandle(h windows.Handle, directory bool) error {
	var info windows.ByHandleFileInformation
	if windows.GetFileInformationByHandle(h, &info) != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory {
		return ErrAuthState
	}
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return ErrAuthState
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return ErrAuthState
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil || !windows.EqualSid(owner, user.User.Sid) {
		return ErrAuthState
	}
	control, _, err := sd.Control()
	if err != nil || directory && control&windows.SE_DACL_PROTECTED == 0 {
		return ErrAuthState
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil || acl.AceCount != 1 {
		return ErrAuthState
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if windows.GetAce(acl, 0, &ace) != nil || ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
		return ErrAuthState
	}
	flags := uint8(0)
	if directory {
		flags = windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE
	}
	// GetAce returns a pointer into the OS-validated descriptor; SidStart is the
	// start of its variable-length SID, not a standalone uint32 identifier.
	sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	const fileAllAccess = windows.STANDARD_RIGHTS_REQUIRED | windows.SYNCHRONIZE | 0x1ff
	// A broker-created file can inherit the sole current-user ACE from its
	// already verified private parent. It must still have exactly one effective
	// full-access user ACE, the current owner and no propagation flags.
	if !directory && control&windows.SE_DACL_PROTECTED == 0 {
		flags = windows.INHERITED_ACE
	}
	if ace.Header.AceFlags != flags || ace.Mask != fileAllAccess || !sid.IsValid() || !windows.EqualSid(sid, user.User.Sid) {
		return ErrAuthState
	}
	runtime.KeepAlive(sd)
	return nil
}

func writeWindowsPrivateCandidate(path string, raw []byte) error {
	if !filepath.IsAbs(path) || len(raw) > 1<<20 || validateWindowsPrivatePath(filepath.Dir(path), true) != nil {
		return ErrAuthState
	}
	sd, err := windowsPrivateDescriptor(false)
	if err != nil {
		return err
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return ErrAuthState
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	h, err := windows.CreateFile(name, windows.GENERIC_WRITE|windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES, 0, &sa, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_WRITE_THROUGH, 0)
	runtime.KeepAlive(sd)
	if err != nil {
		return ErrAuthState
	}
	f := os.NewFile(uintptr(h), path)
	if err = validateWindowsPrivateHandle(h, false); err == nil {
		_, err = f.Write(raw)
	}
	if err == nil {
		err = windows.FlushFileBuffers(h)
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return ErrAuthState
	}
	return nil
}

// replaceWindowsWriteThrough is a rename candidate, NOT a complete durable
// Store transaction. The caller must hold state.lock and prove parent identity,
// the approved API completion contract and readback before publishing success. No COPY_ALLOWED
// flag is used: credential state must never move across volumes.
func replaceWindowsWriteThrough(candidate, target string) error {
	if filepath.Dir(candidate) != filepath.Dir(target) || validateWindowsPrivatePath(candidate, false) != nil {
		return ErrAuthState
	}
	if _, err := os.Lstat(target); err == nil {
		if validateWindowsPrivatePath(target, false) != nil {
			return ErrAuthState
		}
	} else if !os.IsNotExist(err) {
		return ErrAuthState
	}
	from, err := windows.UTF16PtrFromString(candidate)
	if err != nil {
		return ErrAuthState
	}
	to, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return ErrAuthState
	}
	if windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH) != nil {
		return ErrAuthState
	}
	return validateWindowsPrivatePath(target, false)
}
