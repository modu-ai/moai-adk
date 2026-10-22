package testguard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func fixtureModeRights(perm os.FileMode) (owner, everyone uint32, err error) {
	owner = windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE | windows.DELETE | windows.WRITE_DAC
	switch perm {
	case 0o600:
	case 0o644:
		everyone = windows.FILE_GENERIC_READ
	case 0o755:
		owner |= windows.FILE_GENERIC_EXECUTE
		everyone = windows.FILE_GENERIC_READ | windows.FILE_GENERIC_EXECUTE
	default:
		err = fmt.Errorf("unsupported native fixture permission %04o", perm)
	}
	return
}

func fixtureNativePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(abs, `\\?\`) {
		return abs, nil
	}
	if strings.HasPrefix(abs, `\\`) {
		return `\\?\UNC\` + strings.TrimPrefix(abs, `\\`), nil
	}
	return `\\?\` + abs, nil
}

// CheckFileModePolicy inspects a real regular file through a native handle.
// A protected exact owner/Everyone DACL establishes non-owner read access
// without inventing Unix mode bits or opening an impersonation token.
func CheckFileModePolicy(path string, perm os.FileMode) (retErr error) {
	ownerRights, publicRights, err := fixtureModeRights(perm)
	if err != nil {
		return err
	}
	native, err := fixtureNativePath(path)
	if err != nil {
		return err
	}
	name, err := windows.UTF16PtrFromString(native)
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer func() { retErr = errors.Join(retErr, windows.CloseHandle(h)) }()
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_READONLY) != 0 {
		return errors.New("native policy requires a writable regular non-reparse file")
	}
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return fmt.Errorf("read native owner: %w", err)
	}
	if owner == nil {
		return errors.New("native owner is missing")
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	if !owner.Equals(user.User.Sid) {
		return errors.New("new fixture policy does not name its current-user owner")
	}
	control, _, err := sd.Control()
	if err != nil {
		return fmt.Errorf("read native DACL control: %w", err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("native DACL must be protected from inherited grants")
	}
	acl, _, err := sd.DACL()
	count := uint16(1)
	if publicRights != 0 {
		count++
	}
	if err != nil {
		return fmt.Errorf("read native DACL: %w", err)
	}
	if acl == nil || acl.AceCount != count {
		return fmt.Errorf("native DACL must contain exactly %d permission ACEs", count)
	}
	world, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	if err != nil {
		return err
	}
	ownerSeen, worldSeen := false, false
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			return err
		}
		if ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != 0 || ace.Header.AceSize < uint16(unsafe.Sizeof(windows.ACCESS_ALLOWED_ACE{})) {
			return errors.New("unknown, inherited or malformed native permission ACE")
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.IsValid() || int(ace.Header.AceSize) < int(unsafe.Offsetof(ace.SidStart))+sid.Len() {
			return errors.New("invalid or truncated native permission SID")
		}
		switch {
		case sid.Equals(owner) && !ownerSeen && uint32(ace.Mask) == ownerRights:
			ownerSeen = true
		case publicRights != 0 && sid.Equals(world) && !worldSeen && uint32(ace.Mask) == publicRights:
			worldSeen = true
		default:
			return errors.New("unexpected native principal, duplicate ACE or permission grant")
		}
	}
	if !ownerSeen || (publicRights != 0 && !worldSeen) {
		return errors.New("required native permission principal is missing")
	}
	rw, err := windows.CreateFile(name, windows.FILE_READ_DATA|windows.FILE_WRITE_DATA,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return fmt.Errorf("native current-user read/write access check: %w", err)
	}
	var rwInfo windows.ByHandleFileInformation
	infoErr := windows.GetFileInformationByHandle(rw, &rwInfo)
	closeErr := windows.CloseHandle(rw)
	if err := errors.Join(infoErr, closeErr); err != nil {
		return err
	}
	if rwInfo.VolumeSerialNumber != info.VolumeSerialNumber || rwInfo.FileIndexHigh != info.FileIndexHigh || rwInfo.FileIndexLow != info.FileIndexLow {
		return errors.New("native access check reached a different file identity")
	}
	// Also ask the kernel about current-token execute access. An exact 0644
	// or 0600 DACL must not accidentally grant execution; 0755 must grant it.
	execHandle, execErr := windows.CreateFile(name, windows.FILE_EXECUTE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if execErr == nil {
		if closeErr := windows.CloseHandle(execHandle); closeErr != nil {
			return closeErr
		}
		if perm != 0o755 {
			return errors.New("non-executable native file unexpectedly grants execute access")
		}
	} else if perm == 0o755 || !errors.Is(execErr, windows.ERROR_ACCESS_DENIED) {
		return fmt.Errorf("native execute access check: %w", execErr)
	}
	return nil
}

// SeedFileModePolicy creates only a new fixture, assigning its intended DACL
// atomically before data bytes. All operations stay inside caller-owned temp
// fixtures; no account, token privilege or environment policy is changed.
func SeedFileModePolicy(path string, data []byte, perm os.FileMode) error {
	ownerRights, publicRights, err := fixtureModeRights(perm)
	if err != nil {
		return err
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	sid := user.User.Sid.String()
	sddl := fmt.Sprintf("O:%sD:P(A;;0x%x;;;%s)", sid, ownerRights, sid)
	if publicRights != 0 {
		sddl += fmt.Sprintf("(A;;0x%x;;;WD)", publicRights)
	}
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	native, err := fixtureNativePath(path)
	if err != nil {
		return err
	}
	name, err := windows.UTF16PtrFromString(native)
	if err != nil {
		return err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	h, err := windows.CreateFile(name, windows.GENERIC_WRITE|windows.READ_CONTROL|windows.DELETE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, &sa,
		windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(h), path)
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return CheckFileModePolicy(path, perm)
}
