//go:build !windows

package factory

import (
	"os"
	"path/filepath"
	"testing"
)

func factoryRootAlias(t *testing.T, target string) string {
	t.Helper()
	alias := filepath.Join(t.TempDir(), "project-alias")
	if err := os.Symlink(target, alias); err != nil {
		t.Fatal(err)
	}
	return alias
}
