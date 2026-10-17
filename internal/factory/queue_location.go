// queue_location.go — SPEC-MOAI-DB-PATHS-001 (card t1827): the read-only
// location report of the backlog queue that `moai paths` renders.
package factory

// Queue layout names reported by QueueLocation.Layout.
const (
	// QueueLayoutHome is the home layout: <moai-home>/db/<project-key>/todo.
	QueueLayoutHome = "home"
	// QueueLayoutProjectLocal is the project-local layout: <root>/.moai/state/todo.
	QueueLayoutProjectLocal = "project-local"
)

// QueueLocation is where the backlog queue of a project root lives.
type QueueLocation struct {
	// Dir is the canonical queue directory.
	Dir string
	// DBPath is the queue database file inside Dir.
	DBPath string
	// Layout is QueueLayoutHome or QueueLayoutProjectLocal.
	Layout string
	// LegacyDir is the directory that holds the queue while Dir holds none; empty otherwise.
	LegacyDir string
}

// QueueLocationForRoot reports the location of root's queue without touching it.
func QueueLocationForRoot(root string) QueueLocation {
	return QueueLocation{}
}
