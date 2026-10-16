//go:build unix

package backup

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// runRestoreBounded runs RestoreUnmanagedConfigDirs in a goroutine and fails the test
// if it does not return in time. A FIFO planted where the restore reads a file blocks
// os.ReadFile forever, so the bound turns that hang into a clear failure.
func runRestoreBounded(t *testing.T, project, backupDir string) []UnmanagedRestoreRef {
	t.Helper()
	type result struct {
		refs []UnmanagedRestoreRef
		err  error
	}
	done := make(chan result, 1)
	go func() {
		refs, err := RestoreUnmanagedConfigDirs(project, backupDir)
		done <- result{refs, err}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("RestoreUnmanagedConfigDirs: %v", r.err)
		}
		return r.refs
	case <-time.After(5 * time.Second):
		t.Fatal("restore did not return: a FIFO was read as a regular file and blocked")
		return nil
	}
}

// A FIFO at the working-file target must be refused and reported, never read.
func TestRestoreUnmanaged_FIFOTargetIsRefusedNotRead(t *testing.T) {
	project := t.TempDir()
	backupDir := t.TempDir()
	const name = "retired.md"
	src := filepath.Join(backupDir, "evaluator-profiles")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, name), []byte("backup profile\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(project, ".moai", "config", "evaluator-profiles")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(target, name), 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	refs := runRestoreBounded(t, project, backupDir)
	const wantPath = ".moai/config/evaluator-profiles/retired.md"
	for _, r := range refs {
		if r.Path == wantPath && r.Action == "target-refused" {
			return
		}
	}
	t.Errorf("refs = %+v, want a target-refused entry for %s", refs, wantPath)
}

// A FIFO at the side-copy name must be refused, never written over or read.
func TestRestoreUnmanaged_FIFOSideCopyIsRefusedNotRead(t *testing.T) {
	project := t.TempDir()
	backupDir := t.TempDir()
	const name = "retired.md"
	src := filepath.Join(backupDir, "evaluator-profiles")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, name), []byte("backup profile\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(project, ".moai", "config", "evaluator-profiles")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	// The working file differs from the backup, so the restore takes the side-copy path.
	if err := os.WriteFile(filepath.Join(target, name), []byte("new template profile\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stamp := filepath.Base(filepath.Clean(backupDir))
	if err := syscall.Mkfifo(filepath.Join(target, name+".from-"+stamp), 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	refs := runRestoreBounded(t, project, backupDir)
	const wantPath = ".moai/config/evaluator-profiles/retired.md"
	for _, r := range refs {
		if r.Path == wantPath && r.Action == "side-copy-refused" {
			return
		}
	}
	t.Errorf("refs = %+v, want a side-copy-refused entry for %s", refs, wantPath)
}
