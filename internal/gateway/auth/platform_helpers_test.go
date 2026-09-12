package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func privateBrokerTestHome(t *testing.T) (string, error) {
	t.Helper()
	parent, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		return "", e
	}
	return privateScratch(parent, "broker-")
}
func testPrivatePath(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return privateDirectory(path) == nil
	}
	return privateRegular(path, info)
}
