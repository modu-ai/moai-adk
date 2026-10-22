package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReviewFactoryRefusesEveryOccupiedLandingLeaf(t *testing.T) {
	for _, shape := range []string{"directory", "file", "symlink", "uninspectable"} {
		t.Run(shape, func(t *testing.T) {
			base := t.TempDir()
			path := filepath.Join(base, "t7")
			switch shape {
			case "directory":
				if err := os.Mkdir(path, 0755); err != nil {
					t.Fatal(err)
				}
			case "file":
				if err := os.WriteFile(path, []byte("operator data"), 0600); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Symlink(filepath.Join(base, "absent-target"), path); err != nil {
					t.Fatal(err)
				}
			case "uninspectable":
				if err := os.WriteFile(path, []byte("not a directory"), 0600); err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(path, "child")
			}
			if err := factoryRefuseForeignWorktreeAt(context.Background(), path, "", "t7"); err == nil {
				t.Errorf("pre-lease check accepted %s at %s", shape, path)
			}
		})
	}
	if err := factoryRefuseForeignWorktreeAt(context.Background(), filepath.Join(t.TempDir(), "t7"), "", "t7"); err != nil {
		t.Fatalf("absent destination refused: %v", err)
	}
}
