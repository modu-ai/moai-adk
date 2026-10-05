package cli

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// SPEC-QUOTA-AWARE-SCHEDULING-001 M5 (REQ-QAS-018..021): the steering lane
// inventory and its one formatter.
//
// With quota pressure on, `moai todo --auto` and `moai factory status` say which
// live registered lanes run a non-Claude backend, so the operator or leader can
// route the next card there. This file only reads and prints: it dispatches,
// assigns, and reassigns nothing, looks at no card, writes no registry byte, and
// reads no session record — the backend comes from the registry row the lane's
// claim wrote (REQ-QAS-023). Like the rest of the quota path it opens no network
// connection and spawns no process.

// factoryQuotaWarningNoLane is the JSON warning marker of a pressure reading
// that has no candidate lane (REQ-QAS-020).
const factoryQuotaWarningNoLane = "no-non-claude-lane"

// factoryQuotaPressurePrefix begins every steering line (the recommendation and
// the no-candidate warning); the hold line has its own prefix (factory_quota.go).
const factoryQuotaPressurePrefix = "quota pressure: "

// factoryQuotaLane is one candidate lane: its registered label and the backend
// its claim recorded.
type factoryQuotaLane struct {
	Label   string `json:"label"`
	Backend string `json:"backend"`
}

// factoryQuotaInventory is the lane inventory: the live registered lanes whose
// recorded backend is glm or gpt, and the count of live lanes whose backend is
// empty or unrecognised (never a candidate). A lane recording claude is neither.
type factoryQuotaInventory struct {
	Candidates []factoryQuotaLane
	Unknown    int
}

// factoryQuotaReadLanes reads the inventory from the project's factory registry
// through a genuinely read-only open (plan.md §D.1): the path is only computed,
// an absent file means no lanes and creates nothing, the connection is opened
// `mode=ro` with `query_only`, and one SELECT runs over workers. Every failure —
// an unresolvable path, a missing or unreadable file, a read-only directory that
// cannot hold the -shm file, a scan error — yields no candidates, never an error
// and never a partial list.
//
// It must not call FactoryRegistryPath or EnsureProjectLayout (they create
// directories), OpenFactoryPath (journal pragma, schema DDL, migrations, chmod),
// or LoadFactoryRegistry (ImportLegacyWorkers can insert the
// legacy_workers_imported meta row). SQLite may create -wal/-shm sidecar files
// for a read-only reader of a WAL database; that is not a registry write.
//
// @MX:NOTE: [AUTO] The one registry reader of the quota steering path; a caller that opened the registry through the factory's own open would write to it from a status or --auto read.
// @MX:SPEC: SPEC-QUOTA-AWARE-SCHEDULING-001
func factoryQuotaReadLanes(root string) factoryQuotaInventory {
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return factoryQuotaInventory{}
	}
	if _, err := os.Stat(path); err != nil {
		return factoryQuotaInventory{}
	}
	db, err := sql.Open("sqlite", factoryQuotaReadOnlyDSN(path))
	if err != nil {
		return factoryQuotaInventory{}
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	rows, err := db.Query(`SELECT label,pid,backend FROM workers ORDER BY label`)
	if err != nil {
		return factoryQuotaInventory{}
	}
	defer func() { _ = rows.Close() }()
	var inv factoryQuotaInventory
	for rows.Next() {
		var label, backend string
		var pid int
		if err := rows.Scan(&label, &pid, &backend); err != nil {
			return factoryQuotaInventory{}
		}
		if !factoryProcessAlive(pid) {
			continue
		}
		switch backend {
		case factory.BackendGLM, factory.BackendGPT:
			inv.Candidates = append(inv.Candidates, factoryQuotaLane{Label: label, Backend: backend})
		case factory.BackendClaude:
			// A Claude lane is exactly what the steering steers away from.
		default:
			inv.Unknown++
		}
	}
	if err := rows.Err(); err != nil {
		return factoryQuotaInventory{}
	}
	return inv
}

// factoryQuotaReadOnlyDSN is the read-only connection string: `mode=ro` (never
// `immutable`, which can miss rows a live writer has only in the WAL), the
// `query_only` pragma, and a short busy timeout so a locked registry fails fast.
// The path spelling follows homestate.OpenFactoryPath.
func factoryQuotaReadOnlyDSN(path string) string {
	v := url.Values{}
	v.Add("mode", "ro")
	v.Add("_pragma", "query_only(ON)")
	v.Add("_pragma", "busy_timeout(100)")
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p, RawQuery: v.Encode()}).String()
}

// factoryQuotaSteeringLine is the one formatter both steering surfaces use
// (REQ-QAS-019, -020): `quota pressure: <window segment(s)>; recommend
// non-Claude lane(s): <label> (<backend>), …` when the inventory has a
// candidate, and a no-candidate warning carrying the unknown-lane count when it
// has none. A window segment is the one the hold line uses (name, used
// percentage, reset instant), one per held window.
func factoryQuotaSteeringLine(held []factoryQuotaWindowState, inv factoryQuotaInventory) string {
	segments := make([]string, 0, len(held))
	for _, w := range held {
		segments = append(segments, factoryQuotaHoldSegment(w))
	}
	line := factoryQuotaPressurePrefix + strings.Join(segments, "; ")
	if len(inv.Candidates) == 0 {
		return fmt.Sprintf("%s; warning: no live non-Claude lane (unknown backend: %d); nothing is re-dispatched", line, inv.Unknown)
	}
	names := make([]string, 0, len(inv.Candidates))
	for _, c := range inv.Candidates {
		names = append(names, fmt.Sprintf("%s (%s)", c.Label, c.Backend))
	}
	return line + "; recommend non-Claude lane(s): " + strings.Join(names, ", ")
}

// factoryQuotaSteering is the `moai todo --auto` seam's live value: the steering
// line for the project root, or "" while quota pressure is off. It evaluates the
// shared pressure afresh on every call and reads the registry only when pressure
// is on, so a gate that is disabled or a pressure that is off costs no read.
func factoryQuotaSteering(root string) string {
	held := factoryQuotaEvaluate(root).HeldWindows()
	if len(held) == 0 {
		return ""
	}
	return factoryQuotaSteeringLine(held, factoryQuotaReadLanes(root))
}
