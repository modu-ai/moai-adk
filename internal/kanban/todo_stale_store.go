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
	"regexp"
	"sort"
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

// Ghost artifact classes (SPEC-TODO-SURFACE-POLISH-001 REQ-TSP-040). Each
// names one non-SQLite shape the home-DB cutover left behind, reported as
// its own fact — never folded into the SQLite divergence verdict.
const (
	// GhostClassLegacyJSON is a `backlog.json` at a queue path the SQLite
	// engine no longer reads: the downgrade route's file name surviving as
	// a byte document in the home queue directory or a project-local
	// todo/kanban directory.
	GhostClassLegacyJSON = "legacy-json"
	// GhostClassMigratedJSON is a `backlog.json.migrated` — the quarantine
	// rename the lazy migration left behind.
	GhostClassMigratedJSON = "migrated-json"
	// GhostClassSessionRecord is a `<uuid>.json` session-record file in a
	// todo state directory — the pre-factory-database registry entries.
	GhostClassSessionRecord = "session-record"
)

// ghostSessionRecordShaped matches the session-record file name the
// launcher writes (`<uuid>.json`) — the shape that separates registry
// entries from the named artifacts (companions.json, leads.json, the
// autodone log) sharing the same directory.
var ghostSessionRecordShaped = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\.json$`)

// GhostArtifact is one non-SQLite ghost artifact the probe found on disk.
// Bytes is the size the probe measured; a negative Bytes marks an artifact
// that could not be stat'd — its own contradiction fact (REQ-TSP-042's
// FAIL state), never a silent zero.
type GhostArtifact struct {
	Path  string
	Class string
	Bytes int64
}

// StaleStoreFact is the divergence fact one project root yields. Divergent
// is true only when the home database is present and readable AND at least
// one legacy store is present, readable, and carries a different
// meta.last_seq — every other shape must stay silent on the disclosure
// surface (REQ-TSS-005).
//
// Ghosts (SPEC-TODO-SURFACE-POLISH-001 REQ-TSP-040) are the non-SQLite
// ghost artifacts found beside those SQLite facts — each class its own
// entry, sorted by path, and never a second detector: the same probe that
// walks the queue directories reports them.
type StaleStoreFact struct {
	Root         string
	HomePath     string
	HomePresent  bool
	HomeReadable bool
	HomeLastSeq  int
	Stores       []StaleLocalStore
	Divergent    bool
	Ghosts       []GhostArtifact
}

// @MX:ANCHOR: [AUTO] InspectStaleLocalStores — the single divergence detector both disclosure surfaces and the doctor check consume
// @MX:REASON: a second divergence probe would let the stderr warning and the doctor check disagree about whether a ghost store exists (REQ-TSS-004, the "no second inspector" rule)
//
// InspectStaleLocalStores reports the stale-local-store fact for root: the
// home database's presence and meta.last_seq, every project-local legacy
// store found (both the todo-named and kanban-named directories), whether
// any readable pair diverges, and the non-SQLite ghost artifacts in the
// same directories (REQ-TSP-040). Read-only on every branch: no migration,
// no DDL, no lock, no marker file — a ghost is a rollback snapshot and
// every byte of it stays exactly where the probe found it.
func InspectStaleLocalStores(root string) StaleStoreFact {
	fact := StaleStoreFact{
		Root:     root,
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
	fact.Ghosts = inspectGhostArtifacts(root)
	sort.Slice(fact.Ghosts, func(i, j int) bool { return fact.Ghosts[i].Path < fact.Ghosts[j].Path })
	return fact
}

// inspectGhostArtifacts walks the same queue directories the SQLite probe
// walks — the resolved (home or temporary) state directory, the
// project-local todo directory, and the legacy kanban directory — and
// reports the non-SQLite ghost classes found in each: a plain backlog.json,
// its .migrated quarantine sibling, and UUID-shaped session-record files.
//
// One precision rule keeps a LIVE queue from being reported as its own
// ghost: in the CANONICAL state directory, a backlog.json without a
// sibling backlog.db is the live pre-SQLite JSON queue (the lazy
// migration has simply never run for it) — the layout and archive-vouch
// disclosures already speak for that shape, and calling it a ghost would
// name the answering store stale. With the sibling .db present the JSON
// is a leftover, and in every NON-canonical directory it is one
// unconditionally. The .migrated quarantine is always a leftover — the
// migration that produced it wrote the engine's database.
//
// Read-only: every observation is a stat or a directory read.
func inspectGhostArtifacts(root string) []GhostArtifact {
	canonical := StateDirForRoot(root)
	seen := map[string]bool{}
	var ghosts []GhostArtifact
	dirs := []string{canonical, projectStateDirForRoot(root), LegacyStateDirForRoot(root)}
	for _, dir := range dirs {
		if seen[dir] {
			continue
		}
		seen[dir] = true
		entries, err := os.ReadDir(dir)
		if err != nil {
			// An absent or unreadable directory holds no artifacts to name;
			// the SQLite probe reports its own absence facts already.
			continue
		}
		// The engine artifact decides the canonical directory's json class.
		_, dbErr := os.Stat(backlogSQLitePath(filepath.Join(dir, backlogFileName)))
		engineOwns := dbErr == nil
		for _, e := range entries {
			name := e.Name()
			class := ""
			switch {
			case name == backlogFileName:
				if dir == canonical && !engineOwns {
					continue // the live pre-SQLite JSON queue, not a ghost
				}
				class = GhostClassLegacyJSON
			case name == backlogFileName+backlogMigratedSuffix:
				class = GhostClassMigratedJSON
			case ghostSessionRecordShaped.MatchString(name):
				class = GhostClassSessionRecord
			default:
				continue
			}
			info, statErr := e.Info()
			var size int64
			if statErr != nil || !info.Mode().IsRegular() {
				size = -1 // unreadable-or-not-regular is its own fact
			} else {
				size = info.Size()
			}
			ghosts = append(ghosts, GhostArtifact{
				Path:  filepath.Join(dir, name),
				Class: class,
				Bytes: size,
			})
		}
	}
	return ghosts
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
