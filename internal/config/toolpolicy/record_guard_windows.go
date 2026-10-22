//go:build windows

package toolpolicy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const recordGuardOwnerRights = windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE | windows.DELETE | windows.WRITE_DAC

func recordNativePath(path string) (*uint16, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(abs, `\\?\`) {
		if strings.HasPrefix(abs, `\\`) {
			abs = `\\?\UNC\` + strings.TrimPrefix(abs, `\\`)
		} else {
			abs = `\\?\` + abs
		}
	}
	return windows.UTF16PtrFromString(abs)
}

func openRecordGuard(path string) (*recordGuard, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	sid := user.User.Sid.String()
	sd, err := windows.SecurityDescriptorFromString(fmt.Sprintf("O:%sD:P(A;;0x%x;;;%s)", sid, recordGuardOwnerRights, sid))
	if err != nil {
		return nil, err
	}
	name, err := recordNativePath(path)
	if err != nil {
		return nil, err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	// Denying DELETE sharing pins the guard pathname while any caller holds it.
	h, err := windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, &sa,
		windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if errors.Is(err, windows.ERROR_FILE_EXISTS) || errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		h, err = windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil,
			windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	}
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(h), path)
	id, err := inspectRecordGuard(f, path)
	if err != nil {
		return nil, errors.Join(err, f.Close())
	}
	return &recordGuard{file: f, path: path, id: id}, nil
}

func inspectRecordGuard(f *os.File, path string) (id string, result error) {
	h := windows.Handle(f.Fd())
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return "", err
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_READONLY) != 0 || info.NumberOfLinks != 1 {
		return "", errors.New("record guard must be a writable non-reparse regular single-link file")
	}
	if info.FileIndexHigh == 0 && info.FileIndexLow == 0 {
		return "", errors.New("record guard file identity is unavailable")
	}
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return "", err
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil {
		return "", errors.Join(errors.New("record guard owner is unavailable"), err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return "", err
	}
	if !owner.Equals(user.User.Sid) {
		return "", errors.New("record guard is not owned by the current token user")
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return "", errors.Join(errors.New("record guard DACL must be protected"), err)
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil || acl.AceCount != 1 {
		return "", errors.Join(errors.New("record guard requires exactly one owner permission ACE"), err)
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(acl, 0, &ace); err != nil {
		return "", err
	}
	if ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != 0 || ace.Header.AceSize < uint16(unsafe.Sizeof(windows.ACCESS_ALLOWED_ACE{})) || uint32(ace.Mask) != recordGuardOwnerRights {
		return "", errors.New("unknown or non-private record guard permission ACE")
	}
	principal := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	if !principal.IsValid() || int(ace.Header.AceSize) < int(unsafe.Offsetof(ace.SidStart))+principal.Len() || !principal.Equals(owner) {
		return "", errors.New("record guard ACE has an invalid or different principal")
	}
	// Read identity through another native handle; path-based FileInfo on
	// Windows may lazily resolve a name that has since disappeared or changed.
	current, err := openRecordObservation(path)
	if err != nil {
		return "", err
	}
	defer func() { result = errors.Join(result, current.Close()) }()
	var currentInfo windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(current.Fd()), &currentInfo); err != nil {
		return "", err
	}
	id = fmt.Sprintf("%x:%x:%x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow)
	if id != fmt.Sprintf("%x:%x:%x", currentInfo.VolumeSerialNumber, currentInfo.FileIndexHigh, currentInfo.FileIndexLow) {
		return "", errors.New("persistent record guard path was replaced")
	}
	return id, nil
}

func tryRecordGuardLock(f *os.File) error {
	var overlapped windows.Overlapped
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return errRecordGuardBusy
	}
	return err
}

func unlockRecordGuard(f *os.File) error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &overlapped)
}

func openRecordObservation(path string) (*os.File, error) {
	name, err := recordNativePath(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(name, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(h), path)
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return nil, errors.Join(err, f.Close())
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT) != 0 || info.NumberOfLinks != 1 {
		return nil, errors.Join(errors.New("record lock is not a non-reparse regular single-link file"), f.Close())
	}
	return f, nil
}

// The primary descriptor must allow DELETE sharing while retained for identity:
// both normal release and the existing aged-live-holder takeover remove it.
// Its inherited creation policy remains the original Windows OpenFile policy;
// the persistent guard has a distinct, explicit private security descriptor.
func createRecordPrimary(path string) (*os.File, error) {
	name, err := recordNativePath(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(h), path), nil
}
