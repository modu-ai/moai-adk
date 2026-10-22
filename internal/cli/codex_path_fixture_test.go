package cli

import (
	"path/filepath"
	"testing"
)

// codexAbsolutePathFixture keeps slash-form declarations fully qualified on
// both drive-letter and POSIX hosts without cleaning their dot-dot segments.
func codexAbsolutePathFixture(t *testing.T, path string) string {
	t.Helper()
	path = filepath.ToSlash(filepath.VolumeName(t.TempDir())) + path
	if !filepath.IsAbs(path) {
		t.Fatalf("fixture %q is not a host-absolute declaration", path)
	}
	return path
}
