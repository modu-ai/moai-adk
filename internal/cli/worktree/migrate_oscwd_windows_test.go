//go:build windows

// migrate_oscwd_windows_test.go — card t2075: the premise behind the Windows
// probe. The migration no longer skips trees for want of a cwd scan because the
// operating system refuses to rename a directory a process holds as its working
// directory. This test measures that premise on the CI Windows runner; it was
// not run on the authoring host.

package worktree

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestWindowsProbeReportsOSGuard(t *testing.T) {
	dirs, err := defaultActiveProcessCWDs()
	if !errors.Is(err, errProcessCWDOSGuarded) {
		t.Fatalf("err = %v, want errProcessCWDOSGuarded", err)
	}
	if len(dirs) != 0 {
		t.Fatalf("dirs = %v, want none: an empty list would read as no process inside", dirs)
	}
}

func TestWindowsRefusesRenamingAProcessWorkingDirectory(t *testing.T) {
	tree := filepath.Join(t.TempDir(), "tree")
	if err := os.MkdirAll(tree, 0o700); err != nil {
		t.Fatal(err)
	}
	child := exec.Command("ping", "-n", "30", "127.0.0.1")
	child.Dir = tree
	if err := child.Start(); err != nil {
		t.Fatalf("start a process inside the tree: %v", err)
	}
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	})
	time.Sleep(500 * time.Millisecond) // let the child settle in its directory

	moved := tree + "-moved"
	if err := os.Rename(tree, moved); err == nil {
		t.Fatalf("renaming a directory that a process holds as its working directory succeeded; the OS guard the migration relies on does not hold")
	}
	if _, err := os.Stat(tree); err != nil {
		t.Fatalf("the tree must stay in place after the refused rename: %v", err)
	}
}
