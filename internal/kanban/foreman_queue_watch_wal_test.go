// foreman_queue_watch_wal_test.go — SPEC-BACKLOG-JSON-DISCLOSURE-001
// (card t395), AC-BJD-010: a committed write the SQLite store has deferred
// to backlog.db-wal must not hide a queue mutation from the foreman watch.
//
// The premise is real, not hypothetical: backlogDSN sets
// `_pragma journal_mode(WAL)` on every queue connection, so a committed
// write lands in backlog.db-wal and the main database image can stay byte-
// identical until a checkpoint folds it back.
//
// TWO vacuity paths are closed here, and the second is the likelier one.
//
//  1. Widening the observation window until it happens to pass. The window
//     is the same watchWindow every other watch test uses.
//  2. Measuring an already-checkpointed state. r1-repro.md saw backlog.db's
//     cksum move immediately on a `moai todo add`, because that process
//     opens the engine, writes, and CLOSES — and the close checkpoints. A
//     naive run that just mutates and then watches therefore measures a
//     checkpointed database and passes having tested nothing about
//     deferral. The Given is built ON PURPOSE here, from a connection that
//     sets wal_autocheckpoint(0) and STAYS OPEN, and it is EVIDENCED before
//     the watch is armed: backlog.db-wal non-empty AND cksum(backlog.db)
//     unmoved. `wal_autocheckpoint` is per-connection, so this only works
//     because the writing connection is the one that sets it — setting it
//     from a side connection would read as configured while controlling
//     nothing.
//
// If the Given cannot be built, the test FAILS with a Gap message rather
// than passing: an unbuilt Given is not a satisfied criterion.
package kanban

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// cksumOf returns `cksum <path>` output with the path stripped, so the
// value is comparable across calls. This is the literal evidence value
// AC-BJD-010 names.
func cksumOf(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("cksum", path).Output()
	if err != nil {
		t.Fatalf("cksum %s: %v", path, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) < 2 {
		t.Fatalf("cksum %s: unexpected output %q", path, string(out))
	}
	return fields[0] + " " + fields[1]
}

// openDeferringConn opens a queue database on a connection that will NOT
// autocheckpoint. The caller closes it; while it is open, every committed
// write it makes stays in backlog.db-wal.
func openDeferringConn(t *testing.T, dbPath string) *sql.DB {
	t.Helper()
	v := url.Values{}
	v.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", backlogBusyTimeoutMS))
	v.Add("_pragma", "journal_mode(WAL)")
	v.Add("_pragma", "wal_autocheckpoint(0)")
	v.Add("_txlock", "immediate")
	p := filepath.ToSlash(dbPath)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p, RawQuery: v.Encode()}
	db, err := sql.Open(sqliteDriverName, u.String())
	if err != nil {
		t.Fatalf("open deferring connection: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// commitDeferred inserts one queued card and commits it on the deferring
// connection, which keeps the write in the WAL.
func commitDeferred(t *testing.T, db *sql.DB, seq int, id, text string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), backlogOpenTimeout)
	defer cancel()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO items (seq, id, text, added_at, spec_id, state) VALUES (?, ?, ?, ?, NULL, 'queued')`,
		seq, id, text, time.Now().UTC().Format(time.RFC3339)); err != nil {
		t.Fatalf("deferred insert %s: %v", id, err)
	}
}

// dbOnlyWatchScript is the NAIVE repair — the shipped loop with its target
// swapped to backlog.db and nothing else. It is pinned here as the control
// that decides the watch target: if this stays silent under a deferred
// commit while the shipped block's successor fires, then covering the WAL
// is a measured requirement rather than a precaution.
const dbOnlyWatchScript = `f=.moai/state/todo/backlog.db
last=init
while true; do
  if [ -f "$f" ]; then cur=$(cksum "$f"); else cur=missing; fi
  if [ "$cur" != "$last" ]; then
    [ "$last" != init ] && echo "backlog changed"
    last=$cur
  fi
  sleep 5
done`

// TestForemanQueueWatch_DBOnlyTargetMissesWALDeferral — the measurement
// behind the chosen target. A watch that polls backlog.db ALONE cannot see
// a committed-but-uncheckpointed write, so the naive repair would trade one
// silent blind spot for another.
func TestForemanQueueWatch_DBOnlyTargetMissesWALDeferral(t *testing.T) {
	requirePOSIXWatchTools(t)
	root, store := watchFixture(t)
	dbPath := store.EnginePath()
	baseline := cksumOf(t, dbPath)
	conn := openDeferringConn(t, dbPath)
	commitDeferred(t, conn, 900, "t900", "deferred card one")
	walInfo, err := os.Stat(dbPath + "-wal")
	if err != nil || walInfo.Size() == 0 {
		t.Fatalf("GAP — the Given was not built (wal stat %v); this control asserts nothing", err)
	}
	if got := cksumOf(t, dbPath); got != baseline {
		t.Fatalf("GAP — the Given was not built: cksum(backlog.db) moved %s -> %s", baseline, got)
	}

	fired := armWatch(t, root, dbOnlyWatchScript, func() {
		commitDeferred(t, conn, 901, "t901", "deferred card two — mutation under watch")
	})
	if got := cksumOf(t, dbPath); got != baseline {
		t.Fatalf("GAP — cksum(backlog.db) moved %s -> %s during the window; a checkpoint intervened and this control asserts nothing", baseline, got)
	}
	if fired {
		t.Error("the backlog.db-only watch fired under a WAL-deferred commit — then covering backlog.db-wal is not the measured requirement it is documented as, and this control must be re-derived")
	}
}

// TestForemanQueueWatch_SeesWALDeferredCommit — AC-BJD-010.
func TestForemanQueueWatch_SeesWALDeferredCommit(t *testing.T) {
	requirePOSIXWatchTools(t)
	root, store := watchFixture(t)
	dbPath := store.EnginePath()
	walPath := dbPath + "-wal"

	baseline := cksumOf(t, dbPath)
	t.Logf("baseline cksum(backlog.db) = %s", baseline)

	// --- Given: a committed, uncheckpointed write, evidenced. ---
	conn := openDeferringConn(t, dbPath)
	commitDeferred(t, conn, 900, "t900", "deferred card one")

	walInfo, err := os.Stat(walPath)
	if err != nil {
		t.Fatalf("GAP — the Given was not built: backlog.db-wal absent after a committed write (%v); AC-BJD-010 is a Gap, not a pass", err)
	}
	if walInfo.Size() == 0 {
		t.Fatal("GAP — the Given was not built: backlog.db-wal is zero-length after a committed write; AC-BJD-010 is a Gap, not a pass")
	}
	afterCommit := cksumOf(t, dbPath)
	if afterCommit != baseline {
		t.Fatalf("GAP — the Given was not built: cksum(backlog.db) moved %s -> %s, so the write was checkpointed and this run would measure an already-checkpointed state; AC-BJD-010 is a Gap, not a pass", baseline, afterCommit)
	}
	t.Logf("Given established: backlog.db-wal = %d bytes, cksum(backlog.db) unmoved at %s", walInfo.Size(), afterCommit)

	// --- When: the watch is armed and a second deferred commit lands. ---
	script := extractForemanWatchScript(t, foremanSkillPaths["local"])
	fired := armWatch(t, root, script, func() {
		commitDeferred(t, conn, 901, "t901", "deferred card two — mutation under watch")
	})

	// The database image must STILL be unmoved, or the event could have
	// come from a checkpoint rather than from the deferred write.
	afterWindow := cksumOf(t, dbPath)
	if afterWindow != baseline {
		t.Fatalf("GAP — cksum(backlog.db) moved %s -> %s during the window, so a checkpoint intervened and this run does not establish that the watch sees a DEFERRED write; AC-BJD-010 is a Gap, not a pass", baseline, afterWindow)
	}
	if !fired {
		t.Errorf("no change event within %s across a committed-but-uncheckpointed write — the watch target does not cover the WAL deferral", watchWindow)
	}
}
