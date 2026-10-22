package testguard

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestMakeDirReadOnlyDeniesCreationAndRestoresWindowsACL(t *testing.T) {
	for _, protected := range []bool{false, true} {
		name := "inherited"
		if protected {
			name = "protected"
		}
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "records")
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			seed := filepath.Join(dir, "read-control")
			if err := os.WriteFile(seed, []byte("fixture"), 0o600); err != nil {
				t.Fatal(err)
			}
			policy := func(t testing.TB) *windows.SECURITY_DESCRIPTOR {
				t.Helper()
				sd, err := windows.GetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
				if err != nil || sd == nil {
					t.Fatalf("read fixture policy: sd=%v err=%v", sd, err)
				}
				return sd
			}
			original := policy(t)
			acl, _, err := original.DACL()
			if err != nil {
				t.Fatal(err)
			}
			flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION)
			if protected {
				flags = windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION
			}
			if err := windows.SetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT, flags, nil, nil, acl, nil); err != nil {
				t.Fatal(err)
			}
			original = policy(t)
			originalText := original.String()
			originalControl, _, err := original.Control()
			if err != nil {
				t.Fatal(err)
			}
			t.Run("deny", func(t *testing.T) {
				MakeDirReadOnly(t, dir)
				sd := policy(t)
				acl, _, err := sd.DACL()
				if err != nil || acl == nil || acl.AceCount != 2 {
					t.Fatalf("expected explicit deny plus read/execute allow: acl=%v err=%v", acl, err)
				}
				var deny *windows.ACCESS_ALLOWED_ACE // allow/deny ACEs share this layout.
				if err := windows.GetAce(acl, 0, &deny); err != nil {
					t.Fatal(err)
				}
				user, err := windows.GetCurrentProcessToken().GetTokenUser()
				if err != nil {
					t.Fatal(err)
				}
				if deny == nil || deny.Header.AceType != windows.ACCESS_DENIED_ACE_TYPE || deny.Mask != windows.FILE_WRITE_DATA || !(*windows.SID)(unsafe.Pointer(&deny.SidStart)).Equals(user.User.Sid) {
					t.Fatal("fixture did not install a current-user FILE_ADD_FILE deny")
				}
				f, err := os.CreateTemp(dir, "must-not-create-")
				if f != nil {
					if err := f.Close(); err != nil {
						t.Fatal(err)
					}
					t.Fatal("read-only fixture allowed actual file creation")
				}
				if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
					t.Fatalf("creation error=%v, want native ACCESS_DENIED", err)
				}
				if data, err := os.ReadFile(seed); err != nil || string(data) != "fixture" {
					t.Fatalf("denial lost read access: %q, %v", data, err)
				}
			})
			restored := policy(t)
			control, _, err := restored.Control()
			if err != nil {
				t.Fatal(err)
			}
			if restored.String() != originalText || control != originalControl {
				t.Fatalf("cleanup changed original owner/DACL/control: before=%s/%#x after=%s/%#x", originalText, originalControl, restored.String(), control)
			}
			f, err := os.CreateTemp(dir, "restored-create-")
			if err != nil {
				t.Fatalf("original creation access was not restored: %v", err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(f.Name()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
