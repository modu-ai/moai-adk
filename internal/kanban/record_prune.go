// record_prune.go — card t1312. Session records accumulate one file per
// factory session launch with no bound (398 entries measured in a
// long-lived checkout, 2026-09-29). Every reader needs liveness only — the
// doctor Factory Run check (doctor_factory_run.go), the web ops console
// (viewmodel_ops.go loadFactoryRecords), and the stale-run hook
// (session_stale_run.go, which reads only the current session's own record) —
// and none needs forensics: a record carries launch facts alone. Age-based
// retention is therefore the whole contract, and the sweep is confined to
// files the writer actually wrote.
package kanban

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PruneExpiredRecords removes session records older than retentionDays and
// returns how many it removed.
//
// The sweep deletes a file only when ALL of the following hold, which is what
// structurally spares every non-record sibling sharing the directory
// (companions.json, leads.json, backlog.json/backlog.db, the auto-done log —
// and any future tenant of the same directory):
//
//   - the name ends in .json;
//   - the body parses as a Record;
//   - the body's session_id is non-empty AND equals the file's stem — the
//     invariant the writer establishes, so a wrong-body file is never
//     deleted on its name alone;
//   - the record's age is datable and predates the cutoff. entered_at is the
//     datum; the file mtime is the fallback for records that predate the
//     field. An undatable record is spared — the prune deletes only what it
//     can date.
//
// retentionDays <= 0 is a no-op (the explicit-off config value). An absent
// state directory is a state, not a failure — the same contract as ReadAll.
// Per-file failures (unreadable, unparsable, unremovable) skip that file
// silently: the caller is the SessionStart path, which must never gate a
// launch on cleanup, so this returns an error only for a failure that makes
// the whole sweep meaningless.
func PruneExpiredRecords(root string, retentionDays int, now time.Time) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	dir := RuntimeStateDirForRoot(root)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("prune kanban records: %w", err)
	}

	cutoff := now.AddDate(0, 0, -retentionDays)
	removed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		stale, datable := recordFileIsStale(path, e.Name(), cutoff)
		if !datable || !stale {
			continue
		}
		if rmErr := os.Remove(path); rmErr == nil {
			removed++
		}
	}
	return removed, nil
}

// recordFileIsStale reports whether the file at path is a stem-matched session
// record predating cutoff. datable is false for anything the sweep must spare
// or cannot date: unreadable files, non-record bodies, stem mismatches, and
// records whose age neither entered_at nor the mtime can establish.
func recordFileIsStale(path, name string, cutoff time.Time) (stale, datable bool) {
	raw, err := os.ReadFile(path) // #nosec G304 -- path is dir-joined from an os.ReadDir entry
	if err != nil {
		return false, false
	}
	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return false, false
	}
	if rec.SessionID == "" || rec.SessionID != strings.TrimSuffix(name, ".json") {
		return false, false
	}

	entered, err := time.Parse(time.RFC3339, rec.EnteredAt)
	if err != nil {
		info, statErr := os.Stat(path)
		if statErr != nil {
			return false, false
		}
		entered = info.ModTime()
	}
	return entered.Before(cutoff), true
}
