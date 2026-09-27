package homestate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// restoreDir re-grants write permission so t.TempDir cleanup can remove the
// tree; permission-revoking subtests register this before revoking.
func restoreDir(t *testing.T, dir string) {
	t.Helper()
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
}

// The unavailable-record log reports why an append failed instead of writing
// elsewhere: the factory directory being a file fails at MkdirAll, a
// read-only directory fails at the admission lock, and a log path that is a
// directory fails at open.
func TestAppendRecordUnavailableFailureModes(t *testing.T) {
	t.Run("factory dir is a file", func(t *testing.T) {
		root := factorySandbox(t)
		path, err := RecordUnavailablePath(root)
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Dir(path)
		if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dir, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := AppendRecordUnavailable(root, RecordUnavailableEntry{RunID: frRun, CardID: "c", Lane: "w", Error: "e"}); err == nil {
			t.Fatal("append with file-as-factory-dir: err = nil, want MkdirAll failure")
		}
	})
	t.Run("read-only factory dir fails at the lock", func(t *testing.T) {
		root := factorySandbox(t)
		path, err := RecordUnavailablePath(root)
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		restoreDir(t, dir)
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		err = AppendRecordUnavailable(root, RecordUnavailableEntry{RunID: frRun, CardID: "c", Lane: "w", Error: "e"})
		if err == nil || !strings.Contains(err.Error(), "lock record-unavailable") {
			t.Fatalf("append into read-only dir: err = %v, want admission-lock failure", err)
		}
	})
	t.Run("log path is a directory", func(t *testing.T) {
		root := factorySandbox(t)
		path, err := RecordUnavailablePath(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := AppendRecordUnavailable(root, RecordUnavailableEntry{RunID: frRun, CardID: "c", Lane: "w", Error: "e"}); err == nil {
			t.Fatal("append onto a directory: err = nil, want open failure")
		}
	})
}

// The reconciliation rewrite fails loudly and leaves the log intact when it
// cannot take the admission lock (.lock path is a directory) or cannot write
// the replacement (factory directory read-only, so the temp file cannot be
// created).
func TestMarkRecordUnavailableReconciledFailureModes(t *testing.T) {
	t.Run("lock path is a directory", func(t *testing.T) {
		root := factorySandbox(t)
		if err := AppendRecordUnavailable(root, RecordUnavailableEntry{ID: "one", RunID: frRun, CardID: "lost", Lane: "w", Error: "boom"}); err != nil {
			t.Fatal(err)
		}
		path, err := RecordUnavailablePath(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(path + ".lock"); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(path+".lock", 0o700); err != nil {
			t.Fatal(err)
		}
		if err := markRecordUnavailableReconciled(path, map[string]bool{"one": true}); err == nil {
			t.Fatal("reconcile with directory-as-lock: err = nil, want lock failure")
		}
	})
	t.Run("read-only factory dir fails at the temp file", func(t *testing.T) {
		root := factorySandbox(t)
		if err := AppendRecordUnavailable(root, RecordUnavailableEntry{ID: "one", RunID: frRun, CardID: "lost", Lane: "w", Error: "boom"}); err != nil {
			t.Fatal(err)
		}
		path, err := RecordUnavailablePath(root)
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Dir(path)
		restoreDir(t, dir)
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		if err := markRecordUnavailableReconciled(path, map[string]bool{"one": true}); err == nil {
			t.Fatal("reconcile rewrite in read-only dir: err = nil, want temp-file failure")
		}
		// The entries are intact — nothing was rewritten or truncated.
		entries, skipped, err := readRecordUnavailableFile(path)
		if err != nil || skipped != 0 || len(entries) != 1 || entries[0].ID != "one" || entries[0].Reconciled {
			t.Fatalf("log after failed rewrite = %v skipped=%d err=%v, want the single unreconciled entry intact", entries, skipped, err)
		}
	})
}
