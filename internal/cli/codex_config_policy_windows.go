package cli

import (
	"io/fs"
	"os"

	"golang.org/x/sys/windows"
)

// The config retains the user's native access rules. POSIX mode bits are
// synthetic on Windows, so chmod cannot preserve this policy. The staged
// payload starts private and receives the original DACL before replacement.
func restoreCodexConfigPolicy(staged, original string, mode fs.FileMode) error {
	// Keep the existing read-only attribute carry as well as the real ACL.
	// Apply it while the staged file still grants its owner metadata access.
	if err := os.Chmod(staged, mode); err != nil {
		return err
	}
	sd, err := windows.GetNamedSecurityInfo(original, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
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
	return windows.SetNamedSecurityInfo(staged, windows.SE_FILE_OBJECT, flags, nil, nil, acl, nil)
}
