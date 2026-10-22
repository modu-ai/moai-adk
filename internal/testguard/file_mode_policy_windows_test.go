package testguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestNativeFileModePolicyRejectsAlteredGrants(t *testing.T) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sid := user.User.Sid.String()
	for _, tc := range []struct {
		name, dacl  string
		unprotected bool
	}{
		{"non-owner write", "D:P(A;;0x17019f;;;" + sid + ")(A;;FRFW;;;WD)", false},
		{"non-owner execute", "D:P(A;;0x17019f;;;" + sid + ")(A;;FRFX;;;WD)", false},
		{"extra ACE", "D:P(A;;0x17019f;;;" + sid + ")(A;;FR;;;WD)(A;;FR;;;SY)", false},
		{"missing non-owner read", "D:P(A;;0x17019f;;;" + sid + ")", false},
		{"null DACL", "D:NO_ACCESS_CONTROL", false},
		{"unprotected inherited grants", "D:(A;;0x17019f;;;" + sid + ")(A;;FR;;;WD)", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "fixture")
			if err := SeedFileModePolicy(path, []byte("fixture"), 0o644); err != nil {
				t.Fatal(err)
			}
			sd, err := windows.SecurityDescriptorFromString(tc.dacl)
			if err != nil {
				t.Fatal(err)
			}
			acl, _, err := sd.DACL()
			if err != nil {
				t.Fatal(err)
			}
			flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION)
			if tc.unprotected {
				flags = windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION
			}
			if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, flags, nil, nil, acl, nil); err != nil {
				t.Fatal(err)
			}
			if err := CheckFileModePolicy(path, 0o644); err == nil {
				t.Fatal("changed native DACL was accepted as exact public-read mode")
			}
		})
	}
	t.Run("read-only metadata", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "fixture")
		if err := SeedFileModePolicy(path, []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o444); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Chmod(path, 0o644); err != nil {
				t.Errorf("restore owned fixture readonly attribute: %v", err)
			}
		})
		if err := CheckFileModePolicy(path, 0o644); err == nil {
			t.Fatal("read-only metadata was accepted as writable mode")
		}
	})
}

func TestNativeFileModePolicyHandlesLongFixturePath(t *testing.T) {
	dir := t.TempDir()
	for len(dir) < 300 {
		dir = filepath.Join(dir, strings.Repeat("segment", 4))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "fixture")
	if err := SeedFileModePolicy(path, []byte("long native fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckFileModePolicy(path, 0o644); err != nil {
		t.Fatal(err)
	}
}
