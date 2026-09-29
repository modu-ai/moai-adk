// todo_stale_store.go — SPEC-TODO-STALE-STORE-001 (card t1307): the ghost
// queue store a home-DB project can carry without any signal.
//
// After the home-database cutover (SPEC-TODO-QUEUE-HOME-CANON-001) the
// canonical queue is ~/.moai/db/<project-key>/todo/backlog.db, but a
// rollback snapshot can remain project-local under .moai/state/todo/ or
// .moai/state/kanban/. Nothing reads it, yet nothing says it is there —
// and on 2026-09-29 a reader took such a snapshot's queued count for the
// live queue (home last_seq 1305 vs the snapshot's 661). This file is THE
// single detector for that divergence (REQ-TSS-004): the stderr disclosure
// and the doctor check both consume the one fact structure built here, and
// no second probe of the same question exists anywhere.
//
// The detection path is READ-ONLY on every branch (REQ-TSS-003): it opens
// neither store's engine, takes no lock, writes no marker, and changes no
// byte. The ghost store is a rollback snapshot — mutating it destroys the
// very evidence it preserves.
package kanban

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

// StaleLocalStore is one project-local legacy queue store the probe found
// on disk. A store whose meta.last_seq cannot be read (a zero-byte file, a
// non-SQLite image, a corrupt meta row) is reported with Readable false —
// an unreadable store is a separate fact from a divergent one, and is
// never reported as divergence (acceptance §D.2).
type StaleLocalStore struct {
	Path     string
	LastSeq  int
	Readable bool
}

// StaleStoreFact is the divergence fact one project root yields. Divergent
// is true only when the home database is present and readable AND at least
// one legacy store is present, readable, and carries a different
// meta.last_seq — every other shape must stay silent on the disclosure
// surface (REQ-TSS-005).
type StaleStoreFact struct {
	HomePath     string
	HomePresent  bool
	HomeReadable bool
	HomeLastSeq  int
	Stores       []StaleLocalStore
	Divergent    bool
}

// @MX:ANCHOR: [AUTO] InspectStaleLocalStores — the single divergence detector both disclosure surfaces and the doctor check consume
// @MX:REASON: a second divergence probe would let the stderr warning and the doctor check disagree about whether a ghost store exists (REQ-TSS-004, the "no second inspector" rule)
//
// InspectStaleLocalStores reports the stale-local-store fact for root: the
// home database's presence and meta.last_seq, every project-local legacy
// store found (both the todo-named and kanban-named directories), and
// whether any readable pair diverges. Read-only on every branch: no
// migration, no DDL, no lock, no marker file.
func InspectStaleLocalStores(root string) StaleStoreFact {
	fact := StaleStoreFact{
		HomePath: backlogSQLitePath(filepath.Join(StateDirForRoot(root), backlogFileName)),
	}
	if _, err := os.Stat(fact.HomePath); err == nil {
		fact.HomePresent = true
		if seq, ok := readStaleStoreLastSeq(fact.HomePath); ok {
			fact.HomeReadable = true
			fact.HomeLastSeq = seq
		}
	}
	for _, dir := range []string{projectStateDirForRoot(root), LegacyStateDirForRoot(root)} {
		path := backlogSQLitePath(filepath.Join(dir, backlogFileName))
		if path == fact.HomePath {
			// A temporary-root project resolves its home path onto the
			// project-local directory; the two names are one file, and a
			// store cannot diverge from itself.
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}
		seq, readable := readStaleStoreLastSeq(path)
		fact.Stores = append(fact.Stores, StaleLocalStore{Path: path, LastSeq: seq, Readable: readable})
	}
	for _, st := range fact.Stores {
		if fact.HomePresent && fact.HomeReadable && st.Readable && st.LastSeq != fact.HomeLastSeq {
			fact.Divergent = true
			break
		}
	}
	return fact
}

// readStaleStoreLastSeq reads meta.last_seq from one queue database
// without adopting its schema: the open is the same read-only reader the
// archive vouch probe uses (query-only, no DDL, no lock). A file that
// cannot yield an integer last_seq — absent, zero-byte, not SQLite, a
// corrupt meta row — reads as unreadable, which callers report as its own
// fact rather than as divergence.
func readStaleStoreLastSeq(dbPath string) (int, bool) {
	info, err := os.Stat(dbPath)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		return 0, false
	}
	eng, err := openBacklogReader(dbPath)
	if err != nil {
		return 0, false
	}
	defer func() { _ = eng.close() }()
	ctx, cancel := context.WithTimeout(context.Background(), backlogOpenTimeout)
	defer cancel()
	var raw string
	err = eng.db.QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = ?`, backlogMetaKeyLastSeq).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		// An unstamped database reads as 0, matching readLastSeq's
		// convention for a database that never issued.
		return 0, true
	}
	if err != nil {
		return 0, false
	}
	seq, convErr := strconv.Atoi(raw)
	if convErr != nil {
		return 0, false
	}
	return seq, true
}
