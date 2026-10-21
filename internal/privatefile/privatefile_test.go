package privatefile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemporaryFileIsPrivateBeforeWriting(t *testing.T) {
	f, err := CreateTemp(t.TempDir(), "probe-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := Check(f.Name()); err != nil {
		t.Fatalf("empty temporary file is not private: %v", err)
	}
	if _, err := f.WriteString("private payload"); err != nil {
		t.Fatal(err)
	}
	if err := Check(f.Name()); err != nil {
		t.Fatal(err)
	}
}

func TestWriteFileNarrowsExistingAndPreservesContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(path, []byte("new payload")); err != nil {
		t.Fatal(err)
	}
	if err := Check(path); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "new payload" {
		t.Fatalf("payload = %q", raw)
	}
}

func TestWriteFileRejectsSymlinkWithoutChangingTarget(t *testing.T) {
	dir := t.TempDir()
	target, link := filepath.Join(dir, "target"), filepath.Join(dir, "link")
	if err := WriteFile(target, []byte("unchanged")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(link, []byte("must not escape")); err == nil {
		t.Fatal("symlink destination accepted")
	}
	if err := Check(link); err == nil {
		t.Fatal("symlink accepted as private file")
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "unchanged" {
		t.Fatalf("target changed: %q", raw)
	}
}
