package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// StateDir is the per-session goal-state directory under a project root.
const StateDir = ".moai/state/goal"

// ConsumedDir is where orphaned state files are moved by prune.
const ConsumedDir = ".moai/state/goal/consumed"

// StatePath returns the per-session goal-state file path. One file per session
// (never a single shared filename — REQ-GLE-004 multi-session write-race
// avoidance). When sessionID is empty, the caller MUST supply a writer_pid-based
// discriminator via WriterPidKey.
func StatePath(projectRoot, sessionID string) string {
	id := sessionID
	if id == "" {
		id = WriterPidKey()
	}
	return filepath.Join(projectRoot, StateDir, id+".json")
}

// HTMLPath returns the per-session goal-dashboard HTML sibling path
// (.moai/state/goal/<session>.html). It derives the .html path from the StatePath
// .json path via suffix swap (SPEC-GOAL-HTML-FLOW-001 REQ-GHF-004/005, K-4). The
// derivation is defensive: if StatePath ever returns a path without the .json
// suffix, the fallback produces <path>.html (ugly but safe).
func HTMLPath(projectRoot, sessionID string) string {
	jsonPath := StatePath(projectRoot, sessionID)
	if strings.HasSuffix(jsonPath, ".json") {
		return strings.TrimSuffix(jsonPath, ".json") + ".html"
	}
	return jsonPath + ".html"
}

// WriterPidKey returns the writer_pid-based fallback discriminator used when the
// runtime does not expose a session id (REQ-GLE-008). The key is stable for a
// single process so concurrent processes do not clobber one another.
func WriterPidKey() string {
	return "pid-" + strconv.Itoa(os.Getpid())
}

// LoadGoal reads the goal state for the given session from projectRoot. A
// missing file (no armed goal) returns (nil, nil) — not an error — so the hook
// can exit 0 with no block (REQ-GLE-012 preamble: no goal → no block).
func LoadGoal(projectRoot, sessionID string) (*Goal, error) {
	path := StatePath(projectRoot, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("goal load %s: %w", path, err)
	}
	var g Goal
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("goal parse %s: %w", path, err)
	}
	return &g, nil
}

// SaveGoal writes the goal state atomically: a temp file is written in the
// target directory, then renamed into place (REQ-GLE-006 — never a partial
// in-place write that a concurrent reader could observe mid-write).
func SaveGoal(projectRoot string, g *Goal) error {
	dir := filepath.Join(projectRoot, StateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("goal mkdir %s: %w", dir, err)
	}
	finalPath := StatePath(projectRoot, g.SessionID)
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("goal marshal: %w", err)
	}
	// Atomic write: temp file in the SAME directory (so rename is atomic on the
	// same filesystem), then rename over the final path.
	tmp, err := os.CreateTemp(dir, ".goal-*.tmp")
	if err != nil {
		return fmt.Errorf("goal tmp create: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("goal tmp write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("goal tmp close: %w", err)
	}
	if err := atomicfile.Replace(tmpName, finalPath); err != nil {
		return fmt.Errorf("goal rename %s: %w", finalPath, err)
	}
	cleanup = false
	return nil
}

// ClearGoal marks the goal cleared by deleting its state file AND the .html
// dashboard sibling written by `moai goal render` (REQ-GHF-005 / AC-GHF-003).
// Both removes use the fs.ErrNotExist-is-idempotent contract: a file already
// absent is not an error. The hook's `clear` verb uses this so a subsequent
// evaluation sees no armed goal and no orphan dashboard lingers.
func ClearGoal(projectRoot, sessionID string) error {
	paths := []string{
		StatePath(projectRoot, sessionID),
		HTMLPath(projectRoot, sessionID),
	}
	for _, p := range paths {
		if err := os.Remove(p); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // already absent — idempotent
			}
			return fmt.Errorf("goal clear %s: %w", p, err)
		}
	}
	return nil
}
