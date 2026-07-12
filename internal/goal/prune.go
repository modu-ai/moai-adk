package goal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OrphanTTL is the age beyond which a state file with no matching active
// session is treated as orphaned even if the active-sessions registry cannot be
// read. This bounds stale-state growth on hosts without the session-registry
// layer.
const OrphanTTL = 7 * 24 * time.Hour

// PruneOrphans moves state files that are no longer active into the consumed/
// directory. A state file is orphaned when EITHER (a) its session id is absent
// from activeSessionIDs, OR (b) its mtime is older than OrphanTTL. The move is
// best-effort: a failure to move one file does not abort the rest. Returns the
// list of moved filenames.
//
// REQ-GLE-007: "When a session starts, the goal engine shall prune orphan state
// files (session absent from active-sessions.json OR TTL expired) by moving them
// to .moai/state/goal/consumed/."
func PruneOrphans(projectRoot string, activeSessionIDs []string, now time.Time) ([]string, error) {
	srcDir := filepath.Join(projectRoot, StateDir)
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil // no goal state dir yet — nothing to prune
		}
		return nil, fmt.Errorf("goal prune read %s: %w", srcDir, err)
	}
	active := make(map[string]bool, len(activeSessionIDs))
	for _, id := range activeSessionIDs {
		active[id] = true
	}
	consumedDir := filepath.Join(projectRoot, ConsumedDir)
	var moved []string
	for _, e := range entries {
		if e.IsDir() {
			continue // skip consumed/ and any other subdirectory
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if active[id] {
			continue // session still active — keep
		}
		info, err := e.Info()
		if err != nil {
			continue // unreadable metadata — skip rather than abort
		}
		if now.Sub(info.ModTime()) < OrphanTTL {
			// Not in the active registry but not yet TTL-expired either. Keep
			// it; the active registry may simply be incomplete on this host.
			continue
		}
		if err := os.MkdirAll(consumedDir, 0o755); err != nil {
			return moved, fmt.Errorf("goal prune mkdir %s: %w", consumedDir, err)
		}
		src := filepath.Join(srcDir, name)
		dst := filepath.Join(consumedDir, name)
		if err := os.Rename(src, dst); err != nil {
			continue // best-effort: a failed move does not abort the sweep
		}
		moved = append(moved, name)
	}
	return moved, nil
}
