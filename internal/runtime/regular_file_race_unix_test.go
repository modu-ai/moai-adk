//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

// regular_file_race_unix_test.go — card t1632, audit N6/N9: every read of a
// path a caller supplied (or a stored caller path) checks the OPENED file, not
// the path. The swap tests replace the path between the check and the open
// through the beforeRegularOpen seam; the FIFO tests prove the prompt refusal.
package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// regularFileDeadline bounds every call that would block forever on a FIFO, so
// a regression fails the test instead of hanging the suite.
const regularFileDeadline = 3 * time.Second

// withinDeadline runs fn in a goroutine and reports whether it returned in
// time. fifo, when non-empty, is opened non-blocking for write at cleanup so a
// reader still blocked in open(2) on it (the regression case) is released.
func withinDeadline(t *testing.T, fifo string, fn func()) bool {
	t.Helper()
	if fifo != "" {
		t.Cleanup(func() {
			if w, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
				_ = w.Close()
			}
		})
	}
	done := make(chan struct{})
	go func() { fn(); close(done) }()
	select {
	case <-done:
		return true
	case <-time.After(regularFileDeadline):
		return false
	}
}

// setBeforeRegularOpen installs hook for the test and restores the no-op.
func setBeforeRegularOpen(t *testing.T, hook func(path string)) {
	t.Helper()
	prev := beforeRegularOpen
	beforeRegularOpen = hook
	t.Cleanup(func() { beforeRegularOpen = prev })
}

func TestStateCache_StoreRefusesSymlinkSwappedBetweenCheckAndOpen(t *testing.T) {
	root := t.TempDir()
	report := writeVerdictFile(t, root, "review.md", "# plan review: pass")
	target := writeVerdictFile(t, root, "other.md", "# some other file's content")
	setBeforeRegularOpen(t, func(p string) {
		if p != report {
			return
		}
		_ = os.Remove(report)
		_ = os.Symlink(target, report)
	})
	c := NewStateAuditCache(root)
	var err error
	if !withinDeadline(t, "", func() { err = c.StoreVerdict("SPEC-X-001", "hash-1", passResult(report)) }) {
		t.Fatal("StoreVerdict did not return after a symlink swap")
	}
	if err == nil || !strings.Contains(err.Error(), "verdict file evidence") {
		t.Fatalf("StoreVerdict(path swapped to a symlink after the check) err = %v, want a `verdict file evidence: ...` refusal — the link target must never be hashed as the verdict", err)
	}
	if _, statErr := os.Stat(c.entryPath("SPEC-X-001")); statErr == nil {
		t.Fatal("a symlink-swapped report produced a cache entry")
	}
}

func TestFileSHA256_RefusesFIFOSwappedBetweenCheckAndOpen(t *testing.T) {
	dir := t.TempDir()
	path := writeVerdictFile(t, dir, "review.md", "# plan review: pass")
	setBeforeRegularOpen(t, func(p string) {
		if p != path {
			return
		}
		_ = os.Remove(path)
		_ = syscall.Mkfifo(path, 0o644)
	})
	var err error
	if !withinDeadline(t, path, func() { _, err = fileSHA256(path) }) {
		t.Fatal("fileSHA256 blocked on a path swapped to a FIFO after the check — the open must not block, and a non-regular file must be refused")
	}
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("fileSHA256(path swapped to a FIFO) err = %v, want a `not a regular file` refusal", err)
	}
}

func TestComputeHash_RefusesFIFOPlanArtifactPromptly(t *testing.T) {
	specDir := t.TempDir()
	fifo := filepath.Join(specDir, "spec.md")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	var err error
	if !withinDeadline(t, fifo, func() { _, err = (&InMemoryCache{}).ComputeHash(specDir) }) {
		t.Fatal("ComputeHash blocked on a FIFO named like a plan artifact — a non-regular artifact must be refused promptly")
	}
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("ComputeHash(FIFO spec.md) err = %v, want a `not a regular file` refusal", err)
	}
}

func TestComputeHash_RefusesSymlinkedPlanArtifactAndSkipsMissingOnes(t *testing.T) {
	specDir := t.TempDir()
	real := writeVerdictFile(t, specDir, "real.txt", "# real")
	if err := os.Symlink(real, filepath.Join(specDir, "spec.md")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := (&InMemoryCache{}).ComputeHash(specDir); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("ComputeHash(symlinked spec.md) err = %v, want a `not a regular file` refusal", err)
	}

	// A missing optional artifact is still skipped exactly as before.
	okDir := t.TempDir()
	writeVerdictFile(t, okDir, "spec.md", "# spec")
	if h, err := (&InMemoryCache{}).ComputeHash(okDir); err != nil || h == "" {
		t.Fatalf("ComputeHash(spec.md only) = %q, %v — a missing optional artifact must be skipped", h, err)
	}
}

func TestAttachCeiling_DoesNotBlockOnFIFOReportPath(t *testing.T) {
	root := t.TempDir()
	fifo := filepath.Join(root, "report.md")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	c := &GateConfig{SpecID: "SPEC-X-001", SpecDir: root, ProjectDir: root, Cache: NewInMemoryCache()}
	res := &AuditResult{Verdict: VerdictFail, ReportPath: fifo}
	if !withinDeadline(t, fifo, func() { c.attachCeiling(res) }) {
		t.Fatal("attachCeiling blocked on a FIFO ReportPath — the best-effort ceiling read must refuse a non-regular file, not wait on it")
	}
	if res.Ceiling != nil {
		t.Fatalf("a FIFO report produced a ceiling outcome: %+v", res.Ceiling)
	}
}
