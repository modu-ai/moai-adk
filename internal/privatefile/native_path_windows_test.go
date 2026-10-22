package privatefile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsNativePathPreservesNamespace(t *testing.T) {
	longPart := strings.Repeat("segment-", 40)
	for _, tc := range []struct{ path, want string }{
		{`C:\short\file`, `C:\short\file`},
		{`\\server\share\short`, `\\server\share\short`},
		{`C:\` + longPart, `\\?\C:\` + longPart},
		{`\\server\share\` + longPart, `\\?\UNC\server\share\` + longPart},
		{`\\?\C:\` + longPart, `\\?\C:\` + longPart},
		{`\\?\UNC\server\share\` + longPart, `\\?\UNC\server\share\` + longPart},
		{`\\.\` + longPart, `\\.\` + longPart},
	} {
		if got := nativePath(tc.path); got != tc.want {
			t.Errorf("nativePath(%q)=%q, want%q", tc.path, got, tc.want)
		}
	}
}

func TestWindowsNativePrivateOperationsAcceptLongPath(t *testing.T) {
	dir := t.TempDir()
	for len(dir) < 320 {
		dir = filepath.Join(dir, strings.Repeat("long-path-", 8))
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := SecureDirectory(dir); err != nil {
		t.Fatal(err)
	}
	if err := CheckDirectory(dir); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(dir, "private-child")
	if err := Mkdir(child); err != nil {
		t.Fatal(err)
	}
	f, err := CreateTemp(child, "payload-")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	if _, err := f.WriteString("preserved"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Check(name); err != nil {
		t.Fatal(err)
	}
	if err := SecureFile(name); err != nil {
		t.Fatal(err)
	}
	f, err = OpenAppend(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("-append"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(name); err != nil || string(raw) != "preserved-append" {
		t.Fatalf("payload=%q err=%v", raw, err)
	}
}
