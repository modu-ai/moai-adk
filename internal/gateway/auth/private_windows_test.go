//go:build windows

package auth

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestWindowsPrivateStorageCandidate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "private")
	if err := createWindowsPrivateDirectory(dir); err != nil {
		t.Fatal(err)
	}
	if err := validateWindowsPrivatePath(dir, true); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "candidate")
	if err := writeWindowsPrivateCandidate(file, []byte("synthetic-state")); err != nil {
		t.Fatal(err)
	}
	if err := validateWindowsPrivatePath(file, false); err != nil {
		t.Fatal(err)
	}
	if err := writeWindowsPrivateCandidate(file, []byte("overwrite")); err == nil {
		t.Fatal("candidate must use CREATE_NEW")
	}
	final := filepath.Join(dir, "state.json")
	if err := replaceWindowsWriteThrough(file, final); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(final)
	if err != nil || string(raw) != "synthetic-state" {
		t.Fatalf("readback %q, %v", raw, err)
	}
	if err := validateWindowsPrivatePath(final, false); err != nil {
		t.Fatal(err)
	}
	s, e := OpenStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	s.Close()

}

func TestWindowsPrivateStorageRejectsBroadACL(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "private")
	if err := createWindowsPrivateDirectory(dir); err != nil {
		t.Fatal(err)
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	if err := validateWindowsPrivatePath(dir, true); err == nil {
		t.Fatal("Everyone ACL accepted")
	}
}

func TestWindowsPrivateStorageRejectsWrongTypeAndRelative(t *testing.T) {
	if err := createWindowsPrivateDirectory("relative"); err == nil {
		t.Fatal("relative path accepted")
	}
	dir := filepath.Join(t.TempDir(), "private")
	if err := createWindowsPrivateDirectory(dir); err != nil {
		t.Fatal(err)
	}
	if err := validateWindowsPrivatePath(dir, false); err == nil {
		t.Fatal("directory accepted as file")
	}
	if err := createWindowsPrivateDirectory(dir); err == nil {
		t.Fatal("existing directory adopted")
	}
}
