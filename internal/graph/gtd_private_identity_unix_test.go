//go:build !windows

package graph

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestPrivateGTDUnixRealUIDAndModes(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "todo")
	projection, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(dir, "backlog.db"), 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	checker := SystemPrivateGTDPermissionChecker()
	current, err := checker.CurrentUID()
	if err != nil || current != strconv.Itoa(os.Getuid()) {
		t.Fatalf("current identity %q differs from real UID: %v", current, err)
	}
	for _, item := range []struct {
		path          string
		private, weak os.FileMode
	}{{dir, 0o700, 0o755}, {projection.DataPath, 0o600, 0o644}, {projection.MetaPath, 0o600, 0o644}} {
		owner, err := checker.OwnerUID(item.path)
		if err != nil || owner != current {
			t.Fatalf("owner %q differs from current UID: %v", owner, err)
		}
		mode, err := checker.Mode(item.path)
		if err != nil || mode != item.private {
			t.Fatalf("native mode %v want %v: %v", mode, item.private, err)
		}
		if err := AuthorizePrivateGTDProjection(projection, current, checker); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(item.path, item.weak); err != nil {
			t.Fatal(err)
		}
		if err := AuthorizePrivateGTDProjection(projection, current, checker); err == nil {
			t.Errorf("weak native mode %v authorized for %s", item.weak, item.path)
		}
		if err := os.Chmod(item.path, item.private); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := checker.OwnerUID(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing native owner accepted")
	}
}
