package git

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// indexFingerprint identifies the on-disk state of a repository's .git/index:
// a status that refreshes the stat cache rewrites the file, which changes both
// its bytes (the cached stat data lives inside it) and its modification time.
type indexFingerprint struct {
	sum     string
	modTime time.Time
}

func fingerprintIndex(t *testing.T, dir string) indexFingerprint {
	t.Helper()
	path := filepath.Join(dir, ".git", "index")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return indexFingerprint{sum: hex.EncodeToString(sum[:]), modTime: info.ModTime()}
}

// racyGitWindow is longer than the one-second granularity git uses to decide
// whether an index entry is "racily clean"; waiting it out makes the index
// state below depend on stat data alone.
const racyGitWindow = 1100 * time.Millisecond

var staleStatFiles = []string{"a.txt", "b.txt", "c.txt"}

// commitStaleStatFiles commits the files whose stat data bumpStaleStatFiles
// later invalidates without touching their content, so a status that is
// allowed to take optional locks has something to write back.
func commitStaleStatFiles(t *testing.T, dir string) {
	t.Helper()
	for _, name := range staleStatFiles {
		writeFixtureFile(t, filepath.Join(dir, name), name+"\n")
	}
	gitFixture(t, dir, append([]string{"add"}, staleStatFiles...)...)
	gitFixture(t, dir, "commit", "-qm", "files")
}

func bumpStaleStatFiles(t *testing.T, dir string) {
	t.Helper()
	now := time.Now()
	for _, name := range staleStatFiles {
		path := filepath.Join(dir, name)
		if err := os.Chtimes(path, now, now); err != nil {
			t.Fatalf("chtimes %s: %v", path, err)
		}
	}
}

// TestStatusDoesNotRewriteIndex pins the index-write side effect of Status.
// The statusline calls Status on every render in every session, so a status
// that refreshes the stat cache takes the index write lock and makes a
// concurrent add, commit or merge in the same worktree fail on index.lock.
// Status must read the tree without writing .git/index.
func TestStatusDoesNotRewriteIndex(t *testing.T) {
	// The control fixture is prepared in lockstep with the subject: if a
	// plain status does not rewrite the control's index, the stat data never
	// went stale and an unchanged subject index would prove nothing.
	subject := newFixtureRepo(t)
	control := newFixtureRepo(t)
	for _, dir := range []string{subject, control} {
		commitStaleStatFiles(t, dir)
	}

	time.Sleep(racyGitWindow)
	for _, dir := range []string{subject, control} {
		gitFixture(t, dir, "status", "--porcelain") // settle the index once
	}

	time.Sleep(racyGitWindow)
	for _, dir := range []string{subject, control} {
		bumpStaleStatFiles(t, dir)
	}

	controlBefore := fingerprintIndex(t, control)
	gitFixture(t, control, "status", "--porcelain")
	controlAfter := fingerprintIndex(t, control)
	if controlBefore == controlAfter {
		t.Fatalf("control: plain status left .git/index unchanged (sha %s); the stat data never went stale, so this test cannot detect a rewrite", controlBefore.sum)
	}
	t.Logf("control: plain status rewrote .git/index: sha %s -> %s", controlBefore.sum, controlAfter.sum)

	repo, err := NewRepository(subject)
	if err != nil {
		t.Fatalf("NewRepository(%q): %v", subject, err)
	}

	before := fingerprintIndex(t, subject)
	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Status(): %v", err)
	}
	after := fingerprintIndex(t, subject)

	if before != after {
		t.Errorf("Status() rewrote .git/index: sha %s -> %s, mtime %s -> %s",
			before.sum, after.sum, before.modTime.Format(time.RFC3339Nano), after.modTime.Format(time.RFC3339Nano))
	}
	if len(status.Modified) != 0 || len(status.Staged) != 0 || len(status.Untracked) != 0 {
		t.Errorf("Status() on an unchanged tree reported changes: modified=%v staged=%v untracked=%v",
			status.Modified, status.Staged, status.Untracked)
	}
}
