package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/testguard"
	"golang.org/x/sys/windows"
)

func TestSecureHomeDirectoriesRestrictsDescendantFilesWithoutFollowingLinks(t *testing.T) {
	root := filepath.Join(t.TempDir(), "home")
	if err := privatefile.Mkdir(root); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(root, "private-payload")
	outside := filepath.Join(t.TempDir(), "user-asset")
	for _, path := range []string{inside, outside} {
		if err := privatefile.WriteFile(path, []byte("synthetic content")); err != nil {
			t.Fatal(err)
		}
		testguard.MakeFilePublic(t, path)
		if err := privatefile.Check(path); err == nil {
			t.Fatalf("fixture %s must be public", path)
		}
	}
	link := filepath.Join(root, "external-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if err := secureHomeDirectories(root); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.CheckDirectory(root); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Check(inside); err != nil {
		t.Fatalf("descendant file is still broad: %v", err)
	}
	if err := privatefile.Check(outside); err == nil {
		t.Fatal("external linked user asset policy was changed")
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("user link was changed: %v", err)
	}
	raw, err := os.ReadFile(outside)
	if err != nil || string(raw) != "synthetic content" {
		t.Fatalf("user asset changed: %q err=%v", raw, err)
	}
}

func TestSecureHomeDirectoriesRefusesHardlinkAndPreservesExternalPolicy(t *testing.T) {
	base := t.TempDir()
	root, outside := filepath.Join(base, "home"), filepath.Join(base, "external-asset")
	if err := privatefile.Mkdir(root); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.WriteFile(outside, []byte("preserved external asset")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, outside} {
		testguard.MakeFilePublic(t, path)
	}
	before, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "external-hardlink")
	if err := os.Link(outside, alias); err != nil {
		t.Fatal(err)
	}
	if err := secureHomeDirectories(root); err == nil {
		t.Error("home privacy repair admitted an external hardlink")
	}
	if err := privatefile.CheckDirectory(root); err != nil {
		t.Errorf("managed home directory was not kept private: %v", err)
	}
	after, err := windows.GetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if before.String() != after.String() {
		t.Errorf("external policy changed: before=%s after=%s", before.String(), after.String())
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != "preserved external asset" {
		t.Fatalf("external payload=%q err=%v", data, err)
	}
}
