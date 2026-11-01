package mergeq

// N43 (audit 46dbcce0d4, a P2 of the required Codex gate): a write that died between creating its temp file and
// renaming it left a ".tmp-*" file that the retention pass never looked at, so it was never removed and never counted
// in the directory's total cap. The retention pass now removes a temp file that is stale and counts one that is not,
// and never touches a temp file young enough to belong to a write in flight. Real files, real modification times.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// leftoverTemp writes a temp file the way a dead write leaves one, with the modification time given.
func leftoverTemp(t *testing.T, dir, name string, size int, age time.Duration) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, failureLogTmpPrefix+name)
	if err := os.WriteFile(p, bytes.Repeat([]byte("v"), size), 0o600); err != nil {
		t.Fatal(err)
	}
	mt := time.Now().Add(-age)
	if err := os.Chtimes(p, mt, mt); err != nil {
		t.Fatal(err)
	}
	return p
}

func dirBytes(t *testing.T, dir string) (sum int64) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		fi, ierr := e.Info()
		if ierr != nil {
			t.Fatal(ierr)
		}
		sum += fi.Size()
	}
	return sum
}

const tempTestLine = "a line of output\n" // 17 bytes

// A stale leftover is removed and the directory ends within its total cap, whatever the leftover weighed.
func TestFailureLog_AStaleLeftoverTempIsRemovedAndTheDirectoryEndsWithinTheCap(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "gate-logs")
	l := &FailureLog{Dir: dir, TotalCap: 2 << 10}
	if _, err := l.Write("c0", "s0", "gate", []byte(strings.Repeat(tempTestLine, 30))); err != nil {
		t.Fatal(err)
	}
	left := leftoverTemp(t, dir, "12345", 300<<10, 1000*time.Hour)
	for i := 0; i < 5; i++ {
		if _, err := l.Write("c"+string(rune('1'+i)), "s", "gate", []byte(strings.Repeat(tempTestLine, 30))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(left); err == nil {
		t.Error("the stale leftover temp is still there")
	}
	if sum := dirBytes(t, dir); sum > l.TotalCap {
		t.Errorf("the directory holds %d bytes, over its total cap %d", sum, l.TotalCap)
	}
}

// A stale leftover goes at the next write even when the directory is far below its cap: it is not a log, nothing reads
// it, and it holds output.
func TestFailureLog_AStaleLeftoverTempIsRemovedEvenBelowTheCap(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "gate-logs")
	left := leftoverTemp(t, dir, "dead", 10, 1000*time.Hour)
	if _, err := (&FailureLog{Dir: dir}).Write("c0", "s0", "gate", []byte("output\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(left); err == nil {
		t.Error("the stale leftover temp is still there")
	}
}

// A temp file young enough to be a write in flight is not touched, and its bytes count toward the cap: the oldest log
// goes in its place.
func TestFailureLog_AFreshTempIsLeftAloneAndCountedInTheTotal(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "gate-logs")
	l := &FailureLog{Dir: dir, TotalCap: 2 << 10}
	older, err := l.Write("c0", "s0", "gate", []byte(strings.Repeat(tempTestLine, 60))) // 1020 bytes
	if err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(older, old, old); err != nil {
		t.Fatal(err)
	}
	fresh := leftoverTemp(t, dir, "inflight", 1500, 5*time.Second)
	// the logs alone (2040 bytes) fit the cap of 2048; with the temp (3540) they do not
	newer, err := l.Write("c1", "s1", "gate", []byte(strings.Repeat(tempTestLine, 60)))
	if err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(fresh); err != nil || len(b) != 1500 {
		t.Errorf("the fresh temp file was touched: %d bytes, %v", len(b), err)
	}
	if _, err := os.Stat(older); err == nil {
		t.Error("the older log stayed although the fresh temp's bytes put the directory over its cap")
	}
	if _, err := os.Stat(newer); err != nil {
		t.Errorf("the log just written is gone: %v", err)
	}
}
