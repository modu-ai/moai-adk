// queue_location.go — SPEC-MOAI-DB-PATHS-001 (card t1827): the read-only
// location report of the backlog queue that `moai paths` renders.
//
// It exists so the command never joins queue path segments of its own and never
// calls BacklogPathForRoot, which answers with the directory the queue is READ
// from right now — a legacy directory while the canonical one holds no queue —
// and that directory vanishes at the first adopting verb. The report names the
// canonical directory, the place the queue will live, and discloses the legacy
// directory separately.
package factory

import "path/filepath"

// Queue layout names reported by QueueLocation.Layout.
const (
	// QueueLayoutHome is the home layout: <moai-home>/db/<project-key>/todo.
	QueueLayoutHome = "home"
	// QueueLayoutProjectLocal is the project-local layout: <root>/.moai/state/todo.
	QueueLayoutProjectLocal = "project-local"
)

// QueueLocation is where the backlog queue of a project root lives.
type QueueLocation struct {
	// Dir is the canonical queue directory (StateDirForRoot).
	Dir string
	// DBPath is the queue database file inside Dir.
	DBPath string
	// Layout is QueueLayoutHome or QueueLayoutProjectLocal.
	Layout string
	// LegacyDir is the directory that holds the queue while Dir holds none;
	// empty once the queue is canonical.
	LegacyDir string
}

// QueueLocationForRoot reports the location of root's queue. It performs only
// stat and directory reads: it opens no database, relocates nothing, and
// creates nothing. root is the queue root a verb resolves (ResolveTodoQueueRoot).
func QueueLocationForRoot(root string) QueueLocation {
	dir := StateDirForRoot(root)
	loc := QueueLocation{
		Dir:    dir,
		DBPath: backlogSQLitePath(filepath.Join(dir, backlogFileName)),
		Layout: QueueLayoutHome,
	}
	if dir == projectStateDirForRoot(root) {
		loc.Layout = QueueLayoutProjectLocal
	}
	// resolveStateDir(root, false) is the pure form: it observes, moves nothing.
	if legacy, isLegacy := resolveStateDir(root, false); isLegacy {
		loc.LegacyDir = legacy
	}
	return loc
}
