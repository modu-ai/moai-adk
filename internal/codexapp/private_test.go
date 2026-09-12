package codexapp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePrivatePathRejectsSymlinkAndWrongKind(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePrivatePath(dir, true); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePrivatePath(dir, false); err == nil {
		t.Fatal("directory accepted as file")
	}
	file := filepath.Join(dir, "state")
	if err := os.WriteFile(file, []byte("state"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePrivatePath(file, false); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(file, link); err == nil {
		if ValidatePrivatePath(link, false) == nil {
			t.Fatal("symlink accepted")
		}
	}
}

func TestPrivatePathRejectsRelativeMissingAndLinkedParents(t *testing.T) {
	if ValidatePrivatePath(".", true) == nil {
		t.Fatal("relative private path accepted")
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if ValidatePrivatePath(filepath.Join(dir, "missing"), false) == nil {
		t.Fatal("missing private file accepted")
	}
	child := filepath.Join(dir, "child")
	if err = os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(child, "state")
	if err = os.WriteFile(file, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "alias")
	if err = os.Symlink(child, link); err != nil {
		t.Skip("symlink fixture unavailable")
	}
	if ValidatePrivatePath(filepath.Join(link, "state"), false) == nil {
		t.Fatal("symlink parent traversal accepted")
	}
}
