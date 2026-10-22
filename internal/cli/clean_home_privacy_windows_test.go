package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/testguard"
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
