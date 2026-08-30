package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// Regression guard for SPEC-TEMPDIR-CLEANUP-RACE-001 (card t352).
//
// hook.Handle dispatches a durable write into the CALLER's ProjectDir from a
// goroutine that is joined with a bounded deadline, so the write can land after
// Handle has returned. In the internal/hook test binary that goroutine is
// removed by TestMain flipping a package-private seam; that protection is
// package-private and does not cross a test-binary boundary, so a cross-package
// caller such as internal/cli gets the production async path. When that caller
// owns the directory and deletes it on return — every t.TempDir does — the late
// write races the cleanup RemoveAll and surfaces as
// "TempDir RemoveAll cleanup: unlinkat .../.moai/state: directory not empty".
//
// The guard asserts REQ-TCR-001 directly: under the synchronous option, no new
// entry appears anywhere under the caller's directory after Handle returns.
//
// Two properties of the fixture are load-bearing and must not be "simplified":
//
//   - The tree is PADDED. The reproduction measured the index landing 11ms late
//     at 200 files, 43ms at 2000 and 223ms at 8000, while the empty-directory
//     case sat on the join-budget boundary and flipped between bases. A guard
//     with an unpadded fixture is itself a flake.
//   - The directory is created with os.MkdirTemp and removed only AFTER the
//     settle window, not with t.TempDir. t.TempDir would begin its RemoveAll at
//     the instant the guard is measuring, which is the collision this guard
//     exists to prevent — it would make the guard's own failure mode the thing
//     under test.
const deferredWriteGuardPadFiles = 8000

// deferredWriteGuardSettle is how long the guard waits after Handle returns
// before taking the second snapshot. The largest measured lateness was 223ms on
// an 8000-file tree; 1s leaves a wide margin without making the guard slow.
const deferredWriteGuardSettle = time.Second

// entrySet walks dir and returns every path below it, relative to dir, sorted.
// The comparison is over the WHOLE entry set rather than the single writer the
// reproduction established, so the guard matches REQ-TCR-001's "every durable
// write" scope and catches a future writer this card never saw.
func entrySet(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk entry %s: %w", path, err)
		}
		if path == dir {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return fmt.Errorf("relativize %s: %w", path, relErr)
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	sort.Strings(out)
	return out
}

// diffEntries returns the entries present in after but not in before.
func diffEntries(before, after []string) []string {
	seen := make(map[string]struct{}, len(before))
	for _, e := range before {
		seen[e] = struct{}{}
	}
	var added []string
	for _, e := range after {
		if _, ok := seen[e]; !ok {
			added = append(added, e)
		}
	}
	return added
}

func TestSessionStartDeferredWriteDoesNotOutliveHandle(t *testing.T) {
	dir, err := os.MkdirTemp("", "t352-guard")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	// Removed only after the settle window below, so the guard never races its
	// own cleanup (see the const block above).
	t.Cleanup(func() {
		if rmErr := os.RemoveAll(dir); rmErr != nil {
			t.Logf("guard fixture teardown: %v", rmErr)
		}
	})

	pad := filepath.Join(dir, "internal", "pad")
	if err := os.MkdirAll(pad, 0o755); err != nil {
		t.Fatalf("mkdir pad: %v", err)
	}
	for i := 0; i < deferredWriteGuardPadFiles; i++ {
		body := fmt.Sprintf("package pad\n\n// @MX:NOTE: pad %d\nfunc F%d() {}\n", i, i)
		if err := os.WriteFile(filepath.Join(pad, fmt.Sprintf("p%05d.go", i)), []byte(body), 0o644); err != nil {
			t.Fatalf("write pad file: %v", err)
		}
	}

	if _, err := hook.NewSessionStartHandler(nil, hook.WithSynchronousDeferredScans()).Handle(context.Background(), &hook.HookInput{
		SessionID:     "t352-guard",
		CWD:           dir,
		ProjectDir:    dir,
		HookEventName: "SessionStart",
	}); err != nil {
		t.Fatalf("session start Handle: %v", err)
	}

	// The instant the test body would return and a t.TempDir cleanup would begin.
	atReturn := entrySet(t, dir)

	time.Sleep(deferredWriteGuardSettle)
	afterSettle := entrySet(t, dir)

	if added := diffEntries(atReturn, afterSettle); len(added) > 0 {
		t.Fatalf("a durable write outlived Handle: %d entr(ies) appeared under the caller's directory "+
			"during the %v settle window after Handle returned: %v\n"+
			"REQ-TCR-001 requires every durable write into the caller's ProjectDir to complete before Handle returns.",
			len(added), deferredWriteGuardSettle, added)
	}
	if removed := diffEntries(afterSettle, atReturn); len(removed) > 0 {
		t.Fatalf("entr(ies) disappeared from the caller's directory after Handle returned: %v", removed)
	}
}
