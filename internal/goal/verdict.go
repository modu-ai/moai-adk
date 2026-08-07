package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// VerdictSidecarSuffix is the file suffix used for the per-session verdict
// sidecar (.moai/state/goal/<session-id>.verdict.json). The sidecar is the
// at-ceiling-only persistence artifact (SPEC-GOAL-HTML-WIRING-001 REQ-WIRE-002)
// written by the stop-goal evaluator and read by `moai goal render`.
const VerdictSidecarSuffix = ".verdict.json"

// VerdictPath returns the per-session verdict sidecar path
// (.moai/state/goal/<session-id>.verdict.json). It derives from StatePath via
// suffix replacement, mirroring HTMLPath's derivation (K-4).
func VerdictPath(projectRoot, sessionID string) string {
	jsonPath := StatePath(projectRoot, sessionID)
	if strings.HasSuffix(jsonPath, ".json") {
		return strings.TrimSuffix(jsonPath, ".json") + VerdictSidecarSuffix
	}
	return jsonPath + VerdictSidecarSuffix
}

// SaveVerdict persists v to the per-session sidecar atomically (temp-file +
// rename in the same directory). Single-writer: the stop-goal evaluator calls
// this at-ceiling-ONLY (AC-WIRE-013). Single-reader: `moai goal render` loads
// via LoadVerdict. A persistence error is surfaced to the caller; fail-open
// discipline is the caller's responsibility (the evaluator must not block its
// stdout-JSON duty on a sidecar write failure).
func SaveVerdict(projectRoot, sessionID string, v *Verdict) error {
	dir := filepath.Join(projectRoot, StateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("verdict mkdir %s: %w", dir, err)
	}
	finalPath := VerdictPath(projectRoot, sessionID)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("verdict marshal: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".verdict-*.tmp")
	if err != nil {
		return fmt.Errorf("verdict tmp create: %w", err)
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
		return fmt.Errorf("verdict tmp write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("verdict tmp close: %w", err)
	}
	if err := atomicfile.Replace(tmpName, finalPath); err != nil {
		return fmt.Errorf("verdict rename %s: %w", finalPath, err)
	}
	cleanup = false
	return nil
}

// LoadVerdict reads the per-session sidecar. A missing file (no verdict yet for
// this session) returns (nil, nil) — not an error — so the render path degrades
// cleanly to the "no verdict yet" placeholder (AC-GHF-011 preserved). A corrupt
// sidecar returns (nil, err); callers treating the verdict as best-effort
// ignore the error and render the placeholder (REQ-WIRE-003 fail-open).
func LoadVerdict(projectRoot, sessionID string) (*Verdict, error) {
	path := VerdictPath(projectRoot, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("verdict load %s: %w", path, err)
	}
	var v Verdict
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("verdict parse %s: %w", path, err)
	}
	return &v, nil
}
