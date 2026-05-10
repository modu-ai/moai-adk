package statusline

import (
	"os"
	"path/filepath"
	"strings"
)

// ReadModelCache reads the last used model name from cache file.
//
// @MX:ANCHOR: [AUTO] Model cache reader for fallback chain (AC-SF-004)
// @MX:REASON: [AUTO] Called by CollectMetrics on every empty-stdin invocation — high fan_in
// @MX:SPEC: SPEC-V3R3-STATUSLINE-FALLBACK-001
// Returns empty string if file doesn't exist or is corrupted.
// Cache file location: <homeDir>/.moai/state/last-model.txt
func ReadModelCache(homeDir string) (string, error) {
	if homeDir == "" {
		return "", nil
	}

	cachePath := filepath.Join(homeDir, ".moai", "state", "last-model.txt")

	content, err := os.ReadFile(cachePath)
	if err != nil {
		// File doesn't exist or cannot be read - return empty string (no error)
		return "", nil
	}

	// Trim whitespace and newlines
	modelName := strings.TrimSpace(string(content))
	if modelName == "" {
		return "", nil
	}

	return modelName, nil
}

// WriteModelCache writes the model name to cache file atomically.
//
// @MX:ANCHOR: [AUTO] Model cache writer — atomic temp+rename (AC-SF-005)
// @MX:REASON: [AUTO] Called on every successful model extraction — high fan_in
// @MX:SPEC: SPEC-V3R3-STATUSLINE-FALLBACK-001
// AC-SF-005: Creates directory if needed, uses atomic write (temp + rename).
// EC-SF-003: Write failures are silently ignored (no error returned).
func WriteModelCache(homeDir, modelName string) error {
	if homeDir == "" || modelName == "" {
		return nil
	}

	stateDir := filepath.Join(homeDir, ".moai", "state")

	// Create directory if it doesn't exist (AC-SF-005)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		// EC-SF-003: Silent ignore on write failure
		return nil
	}

	cachePath := filepath.Join(stateDir, "last-model.txt")

	// Atomic write: write to temp file, then rename
	tempPath := cachePath + ".tmp"

	// Write to temp file
	if err := os.WriteFile(tempPath, []byte(modelName), 0o644); err != nil {
		// EC-SF-003: Silent ignore on write failure
		return nil
	}

	// Atomic rename (AC-SF-005)
	if err := os.Rename(tempPath, cachePath); err != nil {
		// EC-SF-003: Silent ignore on write failure
		// Clean up temp file
		_ = os.Remove(tempPath) //nolint:errcheck // cleanup best-effort
		return nil
	}

	return nil
}
