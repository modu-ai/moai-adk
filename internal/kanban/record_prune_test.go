package kanban

// record_prune_test.go — card t1312. PruneExpiredRecords is the retention
// mechanism behind the SessionStart prune-on-write path. The sibling-protection
// rules are pinned HERE, in the package that owns the record format: a prune
// that could touch companions.json, leads.json, the auto-done log, or the
// backlog pair would be a data-loss defect, so the negatives are first-class
// tests rather than implementation details.
//
// The consumer measurement this policy encodes (card t1312 verdict): every
// reader of <session>.json records needs liveness only — the doctor Factory
// Run check, the web ops console, and the stale-run hook each lose accuracy
// from stale records and none needs forensics — so age-based retention with a
// stem-matched record identity is the whole contract.

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const pruneTestRetentionDays = 30

// seedStateFile writes a file into the project's session-record directory and
// backdates its mtime. mtime is the fallback age datum for records whose
// entered_at is absent or unparsable.
func seedStateFile(t *testing.T, root, name, body string, modTime time.Time) {
	t.Helper()
	path := filepath.Join(RuntimeStateDirForRoot(root), name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if !modTime.IsZero() {
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
	}
}

func recordBody(sessionID string, enteredAt time.Time) string {
	return `{"session_id":"` + sessionID + `","backend":"claude","entered_at":"` +
		enteredAt.UTC().Format(time.RFC3339) + `"}` + "\n"
}

// The core sweep: stale session records go, everything else stays. The kept
// set exercises each structural protection in turn — recency, a body whose
// session_id does not match the file stem, and every non-record sibling the
// directory shares with the queue and the registries.
func TestPruneExpiredRecordsRemovesOnlyStaleSessionRecords(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	old := now.AddDate(0, 0, -40)

	seedStateFile(t, root, "fresh.json", recordBody("fresh", now), time.Time{})
	seedStateFile(t, root, "old.json", recordBody("old", old), time.Time{})
	// No entered_at: the mtime fallback decides, and the mtime is stale.
	seedStateFile(t, root, "oldmtime.json", `{"session_id":"oldmtime","backend":"claude"}`, old)
	// A record body whose session_id names another file: spared — the stem
	// match is what confines the prune to files the writer actually wrote.
	seedStateFile(t, root, "mismatch.json", recordBody("other", old), time.Time{})
	// Non-record siblings sharing the directory. None parses as a Record with
	// a stem-matching session_id, so every one of them survives any cutoff.
	seedStateFile(t, root, "companions.json", `{"run":[]}`, time.Time{})
	seedStateFile(t, root, "leads.json", `[]`, time.Time{})
	seedStateFile(t, root, "backlog.json", `{"meta":{"last_seq":7}}`, time.Time{})
	seedStateFile(t, root, "auto-done-log.jsonl", `{"card":"t1"}`, time.Time{})
	seedStateFile(t, root, "notes.txt", "not a record", time.Time{})

	removed, err := PruneExpiredRecords(root, pruneTestRetentionDays, now)
	if err != nil {
		t.Fatalf("PruneExpiredRecords: %v", err)
	}
	if removed != 2 {
		t.Fatalf("removed = %d, want 2 (old.json and oldmtime.json)", removed)
	}
	for _, gone := range []string{"old.json", "oldmtime.json"} {
		if _, err := os.Stat(filepath.Join(RuntimeStateDirForRoot(root), gone)); !os.IsNotExist(err) {
			t.Errorf("%s survived the prune (stat err = %v)", gone, err)
		}
	}
	for _, kept := range []string{"fresh.json", "mismatch.json", "companions.json", "leads.json", "backlog.json", "auto-done-log.jsonl", "notes.txt"} {
		if _, err := os.Stat(filepath.Join(RuntimeStateDirForRoot(root), kept)); err != nil {
			t.Errorf("kept file %s is missing after the prune: %v", kept, err)
		}
	}
}

// A record whose age cannot be established — no parsable entered_at and an
// mtime that does not predate the cutoff — is spared. The prune deletes only
// what it can date.
func TestPruneExpiredRecordsSparesUndatableRecord(t *testing.T) {
	root := t.TempDir()
	now := time.Now()

	seedStateFile(t, root, "undated.json", `{"session_id":"undated","backend":"claude"}`, time.Time{})

	removed, err := PruneExpiredRecords(root, pruneTestRetentionDays, now)
	if err != nil {
		t.Fatalf("PruneExpiredRecords: %v", err)
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0 — an undatable record is spared", removed)
	}
	if _, err := os.Stat(filepath.Join(RuntimeStateDirForRoot(root), "undated.json")); err != nil {
		t.Errorf("undated.json is missing after the prune: %v", err)
	}
}

// A future-dated record (clock skew) is never stale.
func TestPruneExpiredRecordsSparesFutureDatedRecord(t *testing.T) {
	root := t.TempDir()
	now := time.Now()

	seedStateFile(t, root, "future.json", recordBody("future", now.AddDate(0, 0, 3)), time.Time{})

	removed, err := PruneExpiredRecords(root, pruneTestRetentionDays, now)
	if err != nil {
		t.Fatalf("PruneExpiredRecords: %v", err)
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0 — a future-dated record is never stale", removed)
	}
}

// retentionDays <= 0 disables the sweep entirely — the explicit-off config
// value must be observable as a no-op, not as a full wipe.
func TestPruneExpiredRecordsZeroRetentionDisables(t *testing.T) {
	root := t.TempDir()
	now := time.Now()

	seedStateFile(t, root, "old.json", recordBody("old", now.AddDate(0, 0, -400)), time.Time{})

	for _, days := range []int{0, -1} {
		removed, err := PruneExpiredRecords(root, days, now)
		if err != nil {
			t.Fatalf("PruneExpiredRecords(%d): %v", days, err)
		}
		if removed != 0 {
			t.Fatalf("retention %d removed %d records, want a no-op", days, removed)
		}
		if _, err := os.Stat(filepath.Join(RuntimeStateDirForRoot(root), "old.json")); err != nil {
			t.Fatalf("retention %d removed old.json — disable must be a no-op", days)
		}
	}
}

// An absent state directory is a state, not a failure — the same contract as
// ReadAll, and the shape a fresh project's first launch hits.
func TestPruneExpiredRecordsAbsentDirIsNoOp(t *testing.T) {
	removed, err := PruneExpiredRecords(t.TempDir(), pruneTestRetentionDays, time.Now())
	if err != nil {
		t.Fatalf("PruneExpiredRecords on an absent state dir: %v", err)
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0", removed)
	}
}
