// todo_ghost_notice.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// REQ-TSP-041: the once-per-class ghost artifact notice.
//
// A ghost artifact (a legacy backlog.json, a .migrated quarantine, a
// session-record JSON — internal/kanban's StaleStoreFact.Ghosts) is a
// leftover the home-database cutover left behind. The FIRST read or write
// verb that discovers a class says so on stderr, once, and records the
// fact in a notice marker inside the ALIVE state directory; every later
// read stays silent. stdout is byte-identical either way — the notice is a
// stderr-only fact (REQ-TSS-002 lineage).
//
// Marker placement (REQ-TSP-041 + REQ-TSS-003): the marker lives in the
// alive state directory — the directory the launcher already fills with
// live runtime artifacts (companions.json, leads.json, the autodone log),
// whose contents are working state, not rollback evidence. The ghost
// evidence itself is never touched: no byte of a ghost file changes, and
// the pure ghost directories (the legacy kanban dir, the ghost positions
// in the home queue directory) gain nothing — the sha-verified read-only
// judgment is TestGhostNoticeOnce's.
//
// Fail-open on every marker error: a marker that cannot be read or
// written costs at most a repeated notice on a later read, never a failed
// verb — the notice is an advisory, and stderr-only.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// ghostNoticeMarkerName is the marker file's base name inside the alive
// state directory.
const ghostNoticeMarkerName = "ghost-notices.json"

// ghostNoticeMarkerPath returns the marker's location for a project root.
func ghostNoticeMarkerPath(root string) string {
	return filepath.Join(kanban.RuntimeStateDirForRoot(root), ghostNoticeMarkerName)
}

// readGhostNoticeMarker loads the noticed-class map the marker records. A
// missing or unreadable marker reads as "nothing noticed yet" — fail-open,
// never an error surfaced to the verb.
func readGhostNoticeMarker(root string) map[string]string {
	raw, err := os.ReadFile(ghostNoticeMarkerPath(root))
	if err != nil {
		return map[string]string{}
	}
	noticed := map[string]string{}
	if err := json.Unmarshal(raw, &noticed); err != nil {
		return map[string]string{}
	}
	return noticed
}

// writeGhostNoticeMarker persists the noticed-class map. Best-effort: a
// write failure costs a repeated notice later, never a failed verb.
func writeGhostNoticeMarker(root string, noticed map[string]string) {
	path := ghostNoticeMarkerPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	encoded, err := json.Marshal(noticed)
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(encoded, '\n'), 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// ghostClassTotals sums one ghost class into a per-class count and byte
// total over the fact's artifacts.
type ghostClassTotals struct {
	count int
	bytes int64
}

// discloseGhostStoresOnce writes one stderr line per ghost class this run
// discovers for the first time — classes the marker already records are
// silent. The marker write lands after the lines, so a failed write costs
// a repeated notice rather than a lost one.
func discloseGhostStoresOnce(w io.Writer, verb string, fact kanban.StaleStoreFact) error {
	if len(fact.Ghosts) == 0 {
		return nil
	}
	noticed := readGhostNoticeMarker(fact.Root)

	totals := map[string]*ghostClassTotals{}
	for _, g := range fact.Ghosts {
		t := totals[g.Class]
		if t == nil {
			t = &ghostClassTotals{}
			totals[g.Class] = t
		}
		t.count++
		if g.Bytes > 0 {
			t.bytes += g.Bytes
		}
	}

	classes := make([]string, 0, len(totals))
	for class := range totals {
		classes = append(classes, class)
	}
	sort.Strings(classes)

	fresh := 0
	for _, class := range classes {
		if _, seen := noticed[class]; seen {
			continue
		}
		t := totals[class]
		fresh++
		if _, err := fmt.Fprintf(w,
			"%s: %d ghost queue artifact(s) of class %q (%d bytes) exist outside the queue — leftovers of the storage cutover, NOT the queue; reported once per class (notice marker: %s)\n",
			verb, t.count, class, t.bytes, ghostNoticeMarkerPath(fact.Root)); err != nil {
			return err
		}
		noticed[class] = time.Now().UTC().Format(time.RFC3339)
	}
	if fresh > 0 {
		writeGhostNoticeMarker(fact.Root, noticed)
	}
	return nil
}
