package testguard

import (
	"errors"
	"os"
	"runtime"
	"testing"
	"unsafe"

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

// MakeDirReadOnly denies creation in a test-owned directory. Windows readonly
// attributes do not deny directory entries; the native DACL must do that.
//
// @MX:ANCHOR: [AUTO] Supply a real creation-denial fixture across operating systems.
// @MX:REASON: Verdict persistence, graph metadata and mission receipt error paths share this fixture.
// @MX:WARN: [AUTO] Native policy mutation requires readback and unconditional restoration.
// @MX:REASON: Retain the original WRITE_DAC handle and restore the original DACL/control, never a replacement grant.
func MakeDirReadOnly(t testing.TB, path string) {
	t.Helper()
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.CreateFile(p, windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := windows.CloseHandle(h); err != nil {
			t.Error(err)
		}
	})
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		t.Fatal(err)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		t.Fatal("read-only fixture must be a regular directory")
	}
	saved, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil || saved == nil {
		t.Fatalf("read original fixture policy: %v", err)
	}
	savedACL, _, err := saved.DACL()
	if err != nil {
		t.Fatal(err)
	}
	control, _, err := saved.Control()
	if err != nil {
		t.Fatal(err)
	}
	restoreFlags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION)
	if control&windows.SE_DACL_PROTECTED != 0 {
		restoreFlags = windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION
	}
	// Register before mutation/readback so even a failed fixture assertion
	// restores the exact original ACL through the already-authorized handle.
	t.Cleanup(func() {
		err := windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, restoreFlags, nil, nil, savedACL, nil)
		runtime.KeepAlive(saved)
		if err != nil {
			t.Errorf("restore original fixture DACL: %v", err)
		}
	})
	before, err := os.CreateTemp(path, ".testguard-writable-")
	if err != nil {
		t.Fatalf("fixture must permit creation before denial: %v", err)
	}
	if err := before.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(before.Name()); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	// FILE_WRITE_DATA's directory alias is FILE_ADD_FILE (0x2). Explicit
	// current-user deny takes precedence over all enabled group grants.
	desired, err := windows.SecurityDescriptorFromString("D:P(D;;0x2;;;" + user.User.Sid.String() + ")(A;;FRFX;;;" + user.User.Sid.String() + ")")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := desired.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	observed, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil || observed == nil {
		t.Fatalf("read back fixture DACL: %v", err)
	}
	observedControl, _, err := observed.Control()
	if err != nil || observedControl&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("fixture DACL is not protected: %v", err)
	}
	observedACL, _, err := observed.DACL()
	if err != nil || observedACL == nil || observedACL.AceCount != 2 {
		t.Fatalf("fixture DACL must contain explicit deny and read/execute allow: %v", err)
	}
	for i := uint32(0); i < 2; i++ {
		var got, want *windows.ACCESS_ALLOWED_ACE // Deny ACEs have the same layout.
		if err := windows.GetAce(observedACL, i, &got); err != nil {
			t.Fatal(err)
		}
		if err := windows.GetAce(acl, i, &want); err != nil {
			t.Fatal(err)
		}
		if got == nil || want == nil || got.Header.AceType != want.Header.AceType || got.Header.AceFlags != want.Header.AceFlags || got.Mask != want.Mask || !(*windows.SID)(unsafe.Pointer(&got.SidStart)).Equals(user.User.Sid) {
			t.Fatalf("fixture ACE %d differs from the requested current-user policy", i)
		}
	}
	probe, err := os.CreateTemp(path, ".testguard-denied-")
	if probe != nil {
		if closeErr := probe.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
		if removeErr := os.Remove(probe.Name()); removeErr != nil {
			t.Fatal(removeErr)
		}
		t.Fatal("fixture allowed file creation after native FILE_ADD_FILE denial")
	}
	if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("fixture creation probe = %v, want native ACCESS_DENIED", err)
	}
}
