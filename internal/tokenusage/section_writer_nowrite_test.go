package tokenusage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// card t2013: re-writing the same §I section performs ZERO writes — the file's
// mtime and inode stay as they were. A changed attribution still writes.
func TestWriteSectionI_IdenticalRewritePerformsNoWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.md")
	if err := os.WriteFile(path, []byte("# Progress\n\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	attr := sampleAttribution()
	if err := WriteSectionI(path, attr); err != nil {
		t.Fatal(err)
	}
	past := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Stat(path)
	bytesBefore, _ := os.ReadFile(path)

	if err := WriteSectionI(path, attr); err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(path)
	bytesAfter, _ := os.ReadFile(path)
	if !after.ModTime().Equal(past) {
		t.Errorf("identical §I rewrite moved mtime: %v -> %v (a write happened)", past, after.ModTime())
	}
	if !os.SameFile(before, after) {
		t.Errorf("identical §I rewrite replaced the inode")
	}
	if string(bytesBefore) != string(bytesAfter) {
		t.Errorf("bytes changed on an identical rewrite")
	}

	attr.SessionCount = 3
	if err := WriteSectionI(path, attr); err != nil {
		t.Fatal(err)
	}
	changed, _ := os.ReadFile(path)
	if string(changed) == string(bytesAfter) {
		t.Errorf("a changed attribution must still write")
	}
	if st, _ := os.Stat(path); st.ModTime().Equal(past) {
		t.Errorf("a changed attribution must move mtime")
	}
}
