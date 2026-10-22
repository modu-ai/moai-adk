package privatefile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/testguard"
	"golang.org/x/sys/windows"
)

func TestWindowsPrivateNarrowingRefusesHardlink(t *testing.T) {
	for _, operation := range []string{"secure-file", "append"} {
		t.Run(operation, func(t *testing.T) {
			dir := t.TempDir()
			outside, alias := filepath.Join(dir, "external-asset"), filepath.Join(dir, "home-link")
			if err := WriteFile(outside, []byte("external bytes remain unchanged")); err != nil {
				t.Fatal(err)
			}
			testguard.MakeFilePublic(t, outside)
			before, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Link(outside, alias); err != nil {
				t.Fatal(err)
			}
			if operation == "secure-file" {
				err = SecureFile(alias)
			} else {
				var f *os.File
				f, err = OpenAppend(alias)
				if f != nil {
					_ = f.Close()
				}
			}
			if err == nil {
				t.Error("multiply linked external file was admitted for policy mutation")
			}
			after, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
			if err != nil {
				t.Fatal(err)
			}
			if before.String() != after.String() {
				t.Errorf("external owner's security descriptor changed: before=%s after=%s", before.String(), after.String())
			}
			data, err := os.ReadFile(outside)
			if err != nil || string(data) != "external bytes remain unchanged" {
				t.Fatalf("external payload=%q err=%v", data, err)
			}
		})
	}
}

func TestWindowsDirectoryNarrowingDoesNotPropagateToHardlink(t *testing.T) {
	base := t.TempDir()
	root, outside := filepath.Join(base, "managed-home"), filepath.Join(base, "external-asset")
	if err := Mkdir(root); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(outside, []byte("external bytes remain unchanged")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, outside} {
		testguard.MakeFilePublic(t, path)
	}
	sd, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	before, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Link(outside, filepath.Join(root, "external-hardlink")); err != nil {
		t.Fatal(err)
	}
	if err := SecureDirectoryNoPropagation(root); err != nil {
		t.Fatal(err)
	}
	if err := CheckDirectory(root); err != nil {
		t.Fatalf("managed directory is not private: %v", err)
	}
	after, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if before.String() != after.String() {
		t.Errorf("directory policy propagated onto external hardlink: before=%s after=%s", before.String(), after.String())
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != "external bytes remain unchanged" {
		t.Fatalf("external payload=%q err=%v", data, err)
	}
}
