package testguard

import (
	"testing"

	"golang.org/x/sys/windows"
)

// MakeFilePublic grants Everyone read access to a test-owned fixture. Windows
// chmod does not modify DACLs, so this supplies the native weak-policy mutant.
func MakeFilePublic(t testing.TB, path string) {
	t.Helper()
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;" + user.User.Sid.String() + ")(A;;FR;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
}

// MakeDirReadOnly supplies the native directory-write-denial fixture; unlike
// chmod(0500), its DACL really refuses creation of children on Windows.
func MakeDirReadOnly(t testing.TB, path string) {
	t.Helper()
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	set := func(rights string) error {
		sd, err := windows.SecurityDescriptorFromString("D:P(A;;" + rights + ";;;" + user.User.Sid.String() + ")")
		if err != nil {
			return err
		}
		acl, _, err := sd.DACL()
		if err != nil {
			return err
		}
		return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
	}
	if err := set("FRFX"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := set("FA"); err != nil {
			t.Error(err)
		}
	})
}
