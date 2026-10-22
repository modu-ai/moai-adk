package cli

import (
	"fmt"
	"io/fs"
	"os"

	"golang.org/x/sys/windows"
)

var setCodexConfigSecurityInfo = windows.SetSecurityInfo

// The config retains the user's native access rules. POSIX mode bits are
// synthetic on Windows, so chmod cannot preserve this policy. The staged
// payload starts private and receives the original owner and DACL before replacement.
func restoreCodexConfigPolicy(staged, original string, mode fs.FileMode) error {
	// Keep the existing read-only attribute carry as well as the real ACL.
	// Apply it while the staged file still grants its owner metadata access.
	if err := os.Chmod(staged, mode); err != nil {
		return err
	}
	sd, err := windows.GetNamedSecurityInfo(original, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return err
	}
	if owner == nil {
		return fmt.Errorf("original Codex config owner is unknown")
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION)
	if control&windows.SE_DACL_PROTECTED != 0 {
		flags = windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION
	}
	name, err := windows.UTF16PtrFromString(staged)
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(name, windows.READ_CONTROL|windows.WRITE_DAC|windows.WRITE_OWNER,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	// Do not enable privileges or take ownership. A caller unable to restore
	// the original owner must refuse publication of the staged replacement.
	if err := setCodexConfigSecurityInfo(h, windows.SE_FILE_OBJECT, flags|windows.OWNER_SECURITY_INFORMATION, owner, nil, acl, nil); err != nil {
		return err
	}
	restored, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	actualOwner, _, err := restored.Owner()
	if err != nil {
		return err
	}
	if actualOwner == nil || !actualOwner.Equals(owner) {
		return fmt.Errorf("staged Codex config owner was not restored")
	}
	return nil
}
