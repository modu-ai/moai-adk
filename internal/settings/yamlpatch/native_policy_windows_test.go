package yamlpatch

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

// Preserve the owner Windows actually assigned, including a normal group
// owner, rather than assuming every existing file belongs to TokenUser.
// The original object gets a custom fixture DACL; production may only read it.
func TestAtomicWritePreservesObservedWindowsOwnerAndDACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "section.yaml")
	if err := os.WriteFile(path, []byte("a: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	initial, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := initial.Owner()
	if err != nil || owner == nil {
		t.Fatalf("observe original fixture owner: %v", err)
	}
	custom, err := windows.SecurityDescriptorFromString("D:P(A;;0x17019f;;;" + owner.String() + ")(A;;FR;;;SY)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := custom.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	before, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	// Confirm a useful readable/writable fixture before attributing refusal
	// to the writer. This never enables a privilege or changes the owner.
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("original native policy does not allow this fixture's caller to read/write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := atomicWrite(path, []byte("a: 2\n")); err != nil {
		t.Fatal(err)
	}
	after, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if before.String() != after.String() {
		t.Errorf("existing native owner/DACL changed: before=%q after=%q", before.String(), after.String())
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "a: 2\n" {
		t.Errorf("updated bytes = %q, err=%v", data, err)
	}
}
