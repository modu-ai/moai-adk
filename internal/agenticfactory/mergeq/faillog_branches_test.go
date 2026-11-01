package mergeq

// The failure branches of FailureLog.Write and its retention pass, reached with real failures of the file system: a
// log that is not configured, a directory where the log's file goes, a parent that cannot be written, and a directory
// that holds things that are not logs.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFailureLog_ALogThatIsNotConfiguredIsAnError(t *testing.T) {
	var none *FailureLog
	if path, err := none.Write("t1", "c", "gate", []byte("x")); err == nil {
		t.Errorf("a nil log wrote %s", path)
	}
	if path, err := (&FailureLog{}).Write("t1", "c", "gate", []byte("x")); err == nil {
		t.Errorf("a log with no directory wrote %s", path)
	}
}

// A non-empty directory at the log's own path cannot be replaced by the rename: the write fails, says so, and leaves
// neither a temporary file nor a damaged directory.
func TestFailureLog_ADirectoryAtTheLogsPathFailsTheWriteAndLeavesNoTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	l := &FailureLog{Dir: dir}
	obstacle := filepath.Join(dir, failureLogName("t1", "c", "gate"))
	if err := os.MkdirAll(filepath.Join(obstacle, "inside"), 0o700); err != nil {
		t.Fatal(err)
	}
	if path, err := l.Write("t1", "c", "gate", []byte("output\n")); err == nil {
		t.Fatalf("the write replaced a non-empty directory: %s", path)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(obstacle) {
		t.Errorf("the directory holds %v, want only the obstacle (no temporary file)", entries)
	}
	if _, serr := os.Stat(filepath.Join(obstacle, "inside")); serr != nil {
		t.Errorf("the obstacle was damaged: %v", serr)
	}
}

// A log directory that cannot be made (its parent cannot be written) is an error, and nothing is written.
func TestFailureLog_ALogDirectoryThatCannotBeMadeIsAnError(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions that bind the user")
	}
	parent := filepath.Join(t.TempDir(), "readonly")
	if err := os.Mkdir(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })
	if path, err := (&FailureLog{Dir: filepath.Join(parent, "gate-logs")}).Write("t1", "c", "gate", []byte("x")); err == nil {
		t.Errorf("a log was written below a parent that cannot be written: %s", path)
	}
}

// The retention pass looks at regular *.log files and nothing else, and never removes the file just written even when
// that file alone is over the total cap.
func TestFailureLog_RetentionLeavesWhatIsNotALogAndKeepsTheFileJustWritten(t *testing.T) {
	dir := t.TempDir()
	foreignFile := filepath.Join(dir, "notes.txt")
	foreignDir := filepath.Join(dir, "sub.log") // named like a log, but a directory
	if err := os.WriteFile(foreignFile, []byte(strings.Repeat("n", 4096)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(foreignDir, 0o700); err != nil {
		t.Fatal(err)
	}
	l := &FailureLog{Dir: dir, TotalCap: 1} // every log is over the cap by itself
	first, err := l.Write("t1", "c1", "gate", []byte("first output\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, serr := os.Stat(first); serr != nil {
		t.Fatalf("the file just written was removed by the retention pass: %v", serr)
	}
	second, err := l.Write("t1", "c2", "gate", []byte("second output\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, serr := os.Stat(first); serr == nil {
		t.Error("the older log stayed although the directory is over its total cap")
	}
	if _, serr := os.Stat(second); serr != nil {
		t.Errorf("the newest log was removed: %v", serr)
	}
	for _, p := range []string{foreignFile, foreignDir} {
		if _, serr := os.Stat(p); serr != nil {
			t.Errorf("the retention pass removed %s, which is not a log: %v", p, serr)
		}
	}
}
