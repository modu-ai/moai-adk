package privatefile

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestCheckRejectsBroadAndInheritedWindowsACLs(t *testing.T) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	for name, dacl := range map[string]string{
		"single-foreign":  "D:P(A;;FA;;;WD)",
		"owner-read-only": "D:P(A;;FR;;;" + user.User.Sid.String() + ")",
		"broad":           "D:P(A;;FA;;;" + user.User.Sid.String() + ")(A;;FR;;;WD)",
		"null":            "D:NO_ACCESS_CONTROL",
		"inherited":       "D:(A;;FA;;;" + user.User.Sid.String() + ")",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "receipt")
			if err := WriteFile(path, []byte("evidence")); err != nil {
				t.Fatal(err)
			}
			sd, err := windows.SecurityDescriptorFromString(dacl)
			if err != nil {
				t.Fatal(err)
			}
			acl, _, err := sd.DACL()
			if err != nil {
				t.Fatal(err)
			}
			flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION)
			if name == "inherited" {
				flags = windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION
			}
			if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, flags, nil, nil, acl, nil); err != nil {
				t.Fatal(err)
			}
			if err := Check(path); err == nil {
				t.Fatalf("%s access policy accepted", name)
			}
		})
	}
}

func TestWindowsFileHasProtectedOwnerOnlyDACL(t *testing.T) {
	f, err := CreateTemp(t.TempDir(), "before-write-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sd, err := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := sd.Owner()
	if err != nil || !owner.Equals(user.User.Sid) {
		t.Fatalf("owner mismatch: %v", err)
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("inheritance is not protected: %v", err)
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil || acl.AceCount != 1 {
		t.Fatalf("expected one owner ACE: %v", err)
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(acl, 0, &ace); err != nil {
		t.Fatal(err)
	}
	sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Mask != fileAllAccess || !sid.Equals(user.User.Sid) {
		t.Fatal("unexpected owner ACE")
	}
	if info, err := os.Stat(f.Name()); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("not a regular file: %v", err)
	}
}
