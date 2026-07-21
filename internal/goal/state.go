package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

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

// ClearGoal marks the goal cleared by deleting its state file (or, if the
// caller prefers a tombstone, setting StatusCleared). The hook's `clear` verb
// uses the delete form so a subsequent evaluation sees no armed goal.
func ClearGoal(projectRoot, sessionID string) error {
	path := StatePath(projectRoot, sessionID)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // already absent — idempotent
		}
		return fmt.Errorf("goal clear %s: %w", path, err)
	}
	return nil
}
