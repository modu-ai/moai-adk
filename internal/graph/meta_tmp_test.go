package graph

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// D20 (REQ-GR-008 sibling hardening, SPEC-GRAPH-REPORT-001 M3): the meta
// sidecar's temp file carries a per-refresh UNIQUE suffix — NOT a pid-based
// or fixed name — because two racing refreshes can be SAME-PROCESS (the
// SessionStart deferred goroutine and a query-time refresh inside one
// binary). N concurrent meta writes to the same destination must all
// succeed, leave a parseable sidecar, and leave no residual temp files.
func TestWriteEdgesMeta_ConcurrentSameProcessNoCollision(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "project", "graph")
	metaPath := filepath.Join(dir, MetaFileName)

	const writers = 8
	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			errs <- WriteEdgesMeta(metaPath, root, map[string]string{"codemaps": "x"}, n)
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent meta write must not collide (unique temp suffix required): %v", err)
		}
	}

	if _, ok := ReadEdgesMeta(metaPath); !ok {
		t.Fatal("the surviving meta sidecar must be parseable")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("residual temp file after concurrent writes: %s", e.Name())
		}
	}
}

// The hardened write's failure branches report wrapped errors rather than
// falling through: an unwritable destination directory and a parent that is
// a regular file each fail loudly at their own step.
func TestWriteEdgesMeta_FailureBranches(t *testing.T) {
	root := t.TempDir()

	// Parent path is a regular FILE: MkdirAll fails.
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := WriteEdgesMeta(filepath.Join(blocker, "sub", MetaFileName), root, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "mkdir") {
		t.Errorf("file-in-place-of-parent must fail at mkdir, got: %v", err)
	}

	// Destination directory exists but is read-only: CreateTemp fails.
	ro := filepath.Join(root, "ro")
	if err := os.MkdirAll(ro, 0o500); err != nil {
		t.Fatal(err)
	}
	err = WriteEdgesMeta(filepath.Join(ro, MetaFileName), root, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "create edges meta temp") {
		t.Errorf("read-only directory must fail at temp creation, got: %v", err)
	}
}

// An empty report's Describe states the clean verdict (the zero-value
// surface callers print when wiring diagnostics).
func TestShrinkReport_DescribeEmpty(t *testing.T) {
	var r ShrinkReport
	if !r.Empty() {
		t.Fatal("zero report must be empty")
	}
	if got := r.Describe(); got != "no unexplained shrink" {
		t.Errorf("empty report Describe = %q", got)
	}
}

// fileExistsUnderRoot's edge branches: an unresolvable ROOT classifies as
// NOT-existing (the permissive direction — a broken root never refuses an
// overwrite; every joined path under it is ENOENT), while a symlink LOOP (a
// non-NotExist resolution error) is INDETERMINATE — ok=false, the caller
// skips the edge rather than guessing.
func TestFileExistsUnderRoot_IndeterminatePaths(t *testing.T) {
	root := t.TempDir()
	exists, ok := fileExistsUnderRoot(filepath.Join(root, "gone"), "x.go")
	if ok && exists {
		t.Error("unresolvable root must never classify as existing")
	}

	loop := filepath.Join(root, "loop.go")
	if err := os.Symlink(loop, loop); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, ok := fileExistsUnderRoot(root, "loop.go"); ok {
		t.Error("symlink loop must be indeterminate (ok=false), never treated as existing or deleted")
	}
}
