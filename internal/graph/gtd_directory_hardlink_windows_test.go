package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/testguard"
	"golang.org/x/sys/windows"
)

func TestPrivateGTDDirectoryPolicyPreservesExternalHardlink(t *testing.T) {
	base := t.TempDir()
	root, outside := filepath.Join(base, "todo"), filepath.Join(base, "external-asset")
	publisher := osPrivateGTDPublisher{}
	if err := publisher.PrepareDir(root); err != nil {
		t.Fatal(err)
	}
	assertDirectoryOperationPreservesExternalHardlink(t, root, outside, func() error {
		return publisher.PrepareDir(root)
	})
}

func assertDirectoryOperationPreservesExternalHardlink(t *testing.T, root, outside string, operation func() error) {
	t.Helper()
	const payload = "external bytes remain unchanged"
	if err := privatefile.WriteFile(outside, []byte(payload)); err != nil {
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
	if err := windows.SetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	before, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Link(outside, filepath.Join(root, "external-hardlink")); err != nil {
		t.Fatal(err)
	}
	if err := operation(); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.CheckDirectory(root); err != nil {
		t.Fatalf("managed directory is not private: %v", err)
	}
	after, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if before.String() != after.String() {
		t.Errorf("external descriptor changed: before=%s after=%s", before.String(), after.String())
	}
	got, err := os.ReadFile(outside)
	if err != nil || string(got) != payload {
		t.Fatalf("external payload changed: got=%q err=%v", got, err)
	}
}
