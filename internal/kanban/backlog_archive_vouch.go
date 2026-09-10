// backlog_archive_vouch.go — SPEC-TODO-ARCHIVE-QUERY-001 (card t394):
// which store is answering a queue read, and whether that store can vouch
// for an archive (REQ-TAQ-013).
//
// A store that cannot carry an archive must not be allowed to answer
// `absent` as though it could. Two reachable degraded shapes exist: a
// database whose archive tables are missing (a pre-archive binary's
// database), and a legacy backlog.json with no backlog.db, which a
// pre-archive binary writes with no `archived` field at all. This probe
// renders that fact for read-only surfaces; it changes nothing itself.
//
// The ORDERING is the load-bearing part: openBacklogEngine runs the DDL on
// every open, and every statement in it is IF NOT EXISTS — so the first
// engine open of a dropped-tables database RECREATES the archive tables and
// erases exactly the fact a caller needs. The probe therefore reads the
// schema from its own connection, before any engine open, and never runs
// the DDL.
package kanban

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
)

// Store names InspectBacklogArchiveVouch reports.
const (
	// BacklogStoreSQLite is the SQLite database engine.
	BacklogStoreSQLite = "the SQLite backlog store"
	// BacklogStoreLegacyJSON is a pre-SQLite backlog.json answered directly.
	BacklogStoreLegacyJSON = "the legacy backlog.json store"
	// BacklogStoreNone is a project with neither artifact.
	BacklogStoreNone = "no backlog store"
)

// BacklogArchiveVouch is the archive-availability fact about one queue
// layout. HasArchive means the store can vouch for an archive — not that
// the archive is non-empty: an empty-but-present archive is vouched,
// because the store can answer "empty" authoritatively.
type BacklogArchiveVouch struct {
	Store      string
	HasArchive bool
	// NonAuthoritativeJSON reports that a backlog.json sits at the
	// canonical queue path while the SQLite store is the one answering
	// reads — State D (SPEC-BACKLOG-JSON-DISCLOSURE-001, REQ-BJD-001).
	// The file answers a direct read silently and is not the queue.
	//
	// Meaningful ONLY on the SQLite branch: on BacklogStoreLegacyJSON the
	// JSON *is* the answering store, and on BacklogStoreNone there is
	// nothing at all — both keep it false.
	NonAuthoritativeJSON bool
}

// InspectBacklogArchiveVouch reports which store answers reads for the
// queue at queuePath and whether that store can vouch for an archive.
// Read-only on every branch: no migration, no DDL, no lock.
func InspectBacklogArchiveVouch(queuePath string) BacklogArchiveVouch {
	layout := inspectBacklogLayout(queuePath)
	if !layout.dbExists {
		if !layout.jsonExists {
			return BacklogArchiveVouch{Store: BacklogStoreNone}
		}
		return BacklogArchiveVouch{Store: BacklogStoreLegacyJSON}
	}
	// layout.jsonExists is measured on every call and was previously
	// discarded on this branch; it is the State D fact, not a new probe.
	return BacklogArchiveVouch{
		Store:                BacklogStoreSQLite,
		HasArchive:           archiveTablesPresent(backlogSQLitePath(queuePath)),
		NonAuthoritativeJSON: layout.jsonExists,
	}
}

// archiveTablesPresent answers whether BOTH archive tables exist in the
// database image right now — from a dedicated connection that runs the
// schema query and nothing else, never the DDL. A probe error reports
// not-vouched: the honest direction, since an unreadable database cannot
// answer archive questions authoritatively either.
func archiveTablesPresent(dbPath string) bool {
	v := url.Values{}
	v.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", backlogBusyTimeoutMS))
	u := url.URL{Scheme: "file", Path: dbPath, RawQuery: v.Encode()}
	db, err := sql.Open(sqliteDriverName, u.String())
	if err != nil {
		return false
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), backlogOpenTimeout)
	defer cancel()
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN ('archived_items','archived_findings')`).Scan(&n); err != nil {
		return false
	}
	return n == 2
}
