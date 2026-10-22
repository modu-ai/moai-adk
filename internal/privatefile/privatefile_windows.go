package privatefile

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// fileAllAccess is FILE_ALL_ACCESS from winnt.h, the native mapping of FA in
// SDDL. Only the current user receives this right; inheritance is disabled.
const fileAllAccess = 0x001f01ff

// CreateTemp installs the protected DACL atomically at CREATE_NEW, before the
// handle can receive secret bytes. No privilege is enabled or requested.
// @MX:ANCHOR: [AUTO] Create private staging files before receiving payload bytes.
// @MX:REASON: Atomic private replacement, mission receipts and mission state share this creation boundary.
func CreateTemp(dir, prefix string) (*os.File, error) {
	sd, err := ownerDescriptor(false)
	if err != nil {
		return nil, err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	for range 100 {
		var nonce [16]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return nil, err
		}
		name := filepath.Join(dir, prefix+hex.EncodeToString(nonce[:]))
		p, err := windows.UTF16PtrFromString(nativePath(name))
		if err != nil {
			return nil, err
		}
		h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE|windows.READ_CONTROL,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, &sa,
			windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
		if err == windows.ERROR_FILE_EXISTS || err == windows.ERROR_ALREADY_EXISTS {
			continue
		}
		if err != nil {
			return nil, err
		}
		f := os.NewFile(uintptr(h), name)
		if err := checkHandle(h); err != nil {
			_ = f.Close()
			_ = os.Remove(name)
			return nil, err
		}
		return f, nil
	}
	return nil, fmt.Errorf("private file: unable to allocate temporary name")
}

func checkNative(path string, _ os.FileInfo) error {
	p, err := windows.UTF16PtrFromString(nativePath(path))
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(p, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return checkHandle(h)
}

func checkHandle(h windows.Handle) error { return checkObjectHandle(h, false) }

// @MX:ANCHOR: [AUTO] Verify exact current-owner access on the opened Windows object.
// @MX:REASON: File validation, directory validation and policy narrowing share this native readback boundary.
// @MX:WARN: [AUTO] Native access-policy validation has multiple independent refusal checks.
// @MX:REASON: Object type, owner, protected inheritance and the exact access rule must all agree before acceptance.
func checkObjectHandle(h windows.Handle, directory bool) error {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory {
		return fmt.Errorf("private file: handle is not a regular file")
	}

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return err
	}
	if owner == nil || !owner.Equals(user.User.Sid) {
		return fmt.Errorf("private file: owner differs from current user")
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		return fmt.Errorf("private file: DACL permits inheritance")
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	if acl == nil || acl.AceCount != 1 {
		return fmt.Errorf("private file: DACL is not owner-only")
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(acl, 0, &ace); err != nil {
		return err
	}
	expectedFlags := uint8(0)
	if directory {
		expectedFlags = windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE
	}
	if ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != expectedFlags || ace.Mask != fileAllAccess {
		return fmt.Errorf("private file: unexpected owner access rule")
	}
	sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	if !sid.Equals(user.User.Sid) {
		return fmt.Errorf("private file: access granted to another identity")
	}
	return nil
}
