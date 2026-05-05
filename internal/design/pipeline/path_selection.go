// Package pipeline: /moai design workflow path selection persistence layer.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 2 (T2-02) implementation.
//
// PathSelection provides functionality to save and load user-selected design path (A/B1/B2)
// and related metadata to `.moai/design/path-selection.json`.
// JSON field order is fixed by struct definition order, guaranteeing same input → same bytes.
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PathSelectionFile: path-selection.json filename (subpath under `.moai/design/`).
const PathSelectionFile = "path-selection.json"

// PathSelection: /moai design workflow path selection data.
// JSON tag order determines serialization field order (no map usage → stable ordering).
type PathSelection struct {
	// Path: Selected design path ("A" | "B1" | "B2")
	Path string `json:"path"`
	// BrandContextLoaded: Whether brand context (.moai/project/brand/) was loaded
	BrandContextLoaded bool `json:"brand_context_loaded"`
	// SpecID: Associated SPEC identifier (e.g., "SPEC-V3R3-DESIGN-001")
	SpecID string `json:"spec_id"`
	// Timestamp: Path selection timestamp (RFC3339 UTC)
	Timestamp time.Time `json:"ts"`
	// SessionID: Session identifier where selection was made
	SessionID string `json:"session_id"`
}

// MissingFieldError: Structured error returned when required field is missing.
type MissingFieldError struct {
	Field string
}

// Error: String representation of MissingFieldError.
func (e *MissingFieldError) Error() string {
	return fmt.Sprintf("missing required field: %s", e.Field)
}

// WritePathSelection: Saves PathSelection as JSON to dir/PathSelectionFile.
// Uses json.MarshalIndent for stable serialization in struct field order.
// Creates directory with MkdirAll if it doesn't exist.
func WritePathSelection(dir string, ps PathSelection) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("directory creation failed (%s): %w", dir, err)
	}

	// Normalize Timestamp to UTC for deterministic serialization
	ps.Timestamp = ps.Timestamp.UTC()

	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return fmt.Errorf("PathSelection JSON serialization failed: %w", err)
	}

	dst := filepath.Join(dir, PathSelectionFile)
	if err := os.WriteFile(dst, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("path-selection.json write failed (%s): %w", dst, err)
	}

	return nil
}

// ReadPathSelection: Reads dir/PathSelectionFile and returns PathSelection.
//
// Error conditions:
//   - File doesn't exist: error including os.ErrNotExist
//   - Corrupted JSON: json parsing error (with context)
//   - "path" field missing or empty: *MissingFieldError{Field: "path"}
func ReadPathSelection(dir string) (PathSelection, error) {
	src := filepath.Join(dir, PathSelectionFile)

	data, err := os.ReadFile(src)
	if err != nil {
		return PathSelection{}, fmt.Errorf("path-selection.json read failed (%s): %w", src, err)
	}

	var ps PathSelection
	if err := json.Unmarshal(data, &ps); err != nil {
		return PathSelection{}, fmt.Errorf("path-selection.json parse failed: %w", err)
	}

	// Required field validation: path
	if ps.Path == "" {
		return PathSelection{}, &MissingFieldError{Field: "path"}
	}

	return ps, nil
}
