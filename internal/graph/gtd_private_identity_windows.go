package graph

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func privateGTDCurrentIdentity() (string, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return "", err
	}
	return privateGTDSIDIdentity(user.User.Sid)
}

func privateGTDSIDIdentity(sid *windows.SID) (string, error) {
	if sid == nil || !sid.IsValid() {
		return "", errors.New("gtd projection: owner_unavailable")
	}
	identity := sid.String()
	if identity == "" {
		return "", errors.New("gtd projection: owner_unavailable")
	}
	return identity, nil
}

// Read the owner on the same non-reparse object whose identity was inspected.
// No security descriptor, owner or token privilege is changed by this lookup.
func privateGTDOwnerIdentity(path string) (identity string, resultErr error) {
	before, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if before.Mode()&os.ModeSymlink != 0 || (!before.IsDir() && !before.Mode().IsRegular()) {
		return "", errors.New("gtd projection: unsafe_owner_object")
	}
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	h, err := windows.CreateFile(p, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", err
	}
	f := os.NewFile(uintptr(h), path)
	defer func() { resultErr = errors.Join(resultErr, f.Close()) }()
	after, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !os.SameFile(before, after) {
		return "", errors.New("gtd projection: owner_object_changed")
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return "", err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != before.IsDir() {
		return "", errors.New("gtd projection: unsafe_owner_object")
	}
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return "", err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return "", err
	}
	return privateGTDSIDIdentity(owner)
}
