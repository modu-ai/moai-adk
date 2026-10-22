package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// writeNativeToolFixture builds a real host executable before the caller
// replaces PATH. Neither Windows PATHEXT lookup nor execution relies on a
// POSIX shebang, a command shell, or an installed copy of the fake tool.
func writeNativeToolFixture(t *testing.T, name, source string) string {
	t.Helper()
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("native tool fixture needs Go: %v", err)
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	if err := os.WriteFile(src, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(dir, name)
	build := exec.Command(goTool, "build", "-p", "1", "-o", binary, src)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build native %s: %v\n%s", name, err, out)
	}
	return binary
}
