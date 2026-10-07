//go:build windows

package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveProgressPathWindowsVolumeRoot (gate round-46 item 6c) — the
// component walk's base must PRESERVE the volume root: a drive path
// (`C:\repo\...`) walks from `C:/` and a UNC path
// (`\\server\share\...`) walks from the server+share — never from a bare
// separator, which turned the first probe into `\C:` and lost the UNC
// root (windows-runtime verification owned by CI windows).
func TestResolveProgressPathWindowsVolumeRoot(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "spec")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	vol := filepath.VolumeName(dir)
	if vol == "" {
		t.Skip("no volume name on the temp path")
	}
	// A dangling FINAL component returns the walk result — volume root
	// included, never a bare-separator-prefixed first component.
	target := filepath.Join(sub, "progress.md")
	resolved, err := resolveProgressPath(target)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != filepath.ToSlash(target) {
		t.Fatalf("resolved %q, want %q (volume root preserved, components applied in order)", resolved, filepath.ToSlash(target))
	}
	if strings.HasPrefix(resolved, "/"+vol[:1]) || strings.HasPrefix(resolved, "/"+filepath.VolumeName(resolved)) {
		t.Fatalf("the volume root was not preserved: %q", resolved)
	}
}
