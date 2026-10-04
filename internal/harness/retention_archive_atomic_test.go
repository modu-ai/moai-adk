package harness

// Card t1467 M1: the archive append must be atomic. appendToGzip used to open the live
// archive with O_APPEND, so a kill inside the write left a truncated gzip member and the
// next prune's append made the whole archive unreadable (probe: 0/35802 events reachable).
// The fix writes a temp file inside archiveDir (existing bytes copied, new stream appended)
// and renames it onto the archive path.
//
// These tests compile against the pre-fix code. Two of them are RED there and two are
// preserve-guards:
//   - RED on the pre-fix code: TestPruneSweepsOldArchiveTempsSparesFresh (the old sweep does
//     not know the archive temp namespace) and, in retention_archive_atomic_unix_test.go,
//     TestPruneFailedRenamePreservesPreviousArchive (the old direct append needs no directory
//     write permission, so it succeeds where the tmp+rename write must fail).
//   - preserve-guards (pass before and after, they pin the contract the new mechanism must
//     keep): TestPruneInterruptedArchiveTempNeverReachesArchive and
//     TestPruneAppendPreservesExistingArchiveStreams.

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeArchiveFixture writes a one-event gzip archive at path, as a previous prune of an
// earlier interval would have left it.
func writeArchiveFixture(t *testing.T, path string, subject string, ts time.Time) {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	enc := json.NewEncoder(gw)
	if err := enc.Encode(Event{Timestamp: ts, EventType: EventTypeFeedback, Subject: subject, ContextHash: "h", SchemaVersion: LogSchemaVersion}); err != nil {
		t.Fatalf("encode archive fixture: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("close archive fixture: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write archive fixture: %v", err)
	}
}

// archiveSubjects decodes every event of a (possibly multi-stream) archive and returns
// subject -> occurrence. Unlike readArchive it reports occurrences instead of raw text, so
// a corrupted extra member fails the decode instead of hiding in a string.
func archiveSubjects(t *testing.T, path string) map[string]int {
	t.Helper()
	out := map[string]int{}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	out2, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("decompress archive: %v", err)
	}
	for _, line := range splitNonEmptyLines(string(out2)) {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("archive line does not parse: %v", err)
		}
		out[evt.Subject]++
	}
	return out
}

// TestPruneInterruptedArchiveTempNeverReachesArchive: a partial archive temp left behind by
// a killed copy is never renamed onto the archive path; the prune that follows writes a
// clean archive holding exactly its own stale events, and the fresh temp is spared for the
// sweep.
func TestPruneInterruptedArchiveTempNeverReachesArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}

	// The interrupted state: a temp with a partial gzip member inside, exactly where a
	// killed copy would have left it.
	interrupted := filepath.Join(archiveDir, "archive-999999999.jsonl.gz.tmp")
	if err := os.WriteFile(interrupted, []byte{0x1f, 0x8b, 0x08, 0x99, 0x42}, 0o600); err != nil {
		t.Fatalf("plant interrupted temp: %v", err)
	}

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}

	archivePath := filepath.Join(archiveDir, now.AddDate(0, 0, -40).UTC().Format("2006-01")+".jsonl.gz")
	subjects := archiveSubjects(t, archivePath)
	if subjects["stale-1"] != 1 || len(subjects) != 1 {
		t.Errorf("archive subjects = %v, want exactly stale-1 once (the interrupted temp's bytes must not be inside)", subjects)
	}
	if tempBytes, terr := os.ReadFile(interrupted); terr == nil {
		if arcBytes, aerr := os.ReadFile(archivePath); aerr == nil && bytes.Equal(tempBytes, arcBytes) {
			t.Errorf("the archive is byte-identical to the interrupted temp")
		}
	}
	if !pathExists(interrupted) {
		t.Errorf("the fresh interrupted temp (0 minutes old) was not left for the orphan sweep")
	}
}

// TestPruneAppendPreservesExistingArchiveStreams: a prune that appends to a month archive
// from an earlier interval keeps the old stream readable next to the new one.
func TestPruneAppendPreservesExistingArchiveStreams(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}
	archivePath := filepath.Join(archiveDir, now.AddDate(0, 0, -40).UTC().Format("2006-01")+".jsonl.gz")
	writeArchiveFixture(t, archivePath, "archived-1", now.AddDate(0, 0, -40))

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}

	subjects := archiveSubjects(t, archivePath)
	if subjects["archived-1"] != 1 {
		t.Errorf("the pre-existing stream is gone or unreadable: subjects = %v", subjects)
	}
	if subjects["stale-1"] != 1 {
		t.Errorf("the new stream is missing: subjects = %v", subjects)
	}
}

// TestPruneSweepsOldArchiveTempsSparesFresh: the lock holder's sweep reaps archive temps
// older than the orphan age under archiveDir, the same rule the log rewrite temps get, and
// spares young ones and everything that is not a prune temp.
func TestPruneSweepsOldArchiveTempsSparesFresh(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}

	oldA := filepath.Join(archiveDir, "archive-111111111.jsonl.gz.tmp")
	oldB := filepath.Join(archiveDir, "archive-222222222.jsonl.gz.tmp")
	young := filepath.Join(archiveDir, "archive-333333333.jsonl.gz.tmp")
	writeTmpAged(t, oldA, now, 48*time.Hour)
	writeTmpAged(t, oldB, now, time.Hour)
	writeTmpAged(t, young, now, time.Minute)

	// Entries the sweep must never take: not the temp namespace, and a directory that
	// matches the namespace (only regular files are ours).
	keepers := []string{
		filepath.Join(archiveDir, "notes.txt"),
		filepath.Join(archiveDir, "archive-444444444.jsonl.gz.bak"),
	}
	for _, k := range keepers {
		writeTmpAged(t, k, now, 48*time.Hour)
	}
	dirMatch := filepath.Join(archiveDir, "archive-555555555.jsonl.gz.tmp")
	if err := os.Mkdir(dirMatch, 0o755); err != nil {
		t.Fatalf("mkdir keeper: %v", err)
	}

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}

	for _, gone := range []string{oldA, oldB} {
		if pathExists(gone) {
			t.Errorf("old archive orphan %s survived the prune", filepath.Base(gone))
		}
	}
	if !pathExists(young) {
		t.Errorf("a young archive temp (1 minute old) was swept; it may belong to a live copy")
	}
	for _, k := range append(keepers, dirMatch) {
		if !pathExists(k) {
			t.Errorf("%s was swept but is not a prune archive temp", filepath.Base(k))
		}
	}
}
