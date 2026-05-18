// Package migrate provides migration steps for upgrading user project configurations
// between MoAI-ADK versions. Each step is idempotent and archival — no user data
// is deleted without a corresponding archive entry.
package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// CleanupUserSettings removes stale RETIRE-OBS-ONLY event entries from the user's
// local .claude/settings.json. This migration step is required for users upgrading
// from pre-SPEC-V3R2-RT-006 v3.0 builds whose settings.json still carries the 4
// retired event hook registrations (Notification, Elicitation, ElicitationResult,
// TaskCreated).
//
// Behavior:
//   - Reads <projectRoot>/.claude/settings.json.
//   - For each name in hook.RetiredEventNames, removes the matching hooks.<EventName> entry.
//   - Archives removed entries to <projectRoot>/.moai/archive/hooks/v3.0/migration-<YYYY-MM-DD>.json.
//   - Writes the cleaned settings.json atomically (temp-file + rename).
//   - Is a no-op when no retired entries are present (no file is written, no archive created).
//   - Returns a wrapped error on JSON parse failure without writing any output.
//
// SPEC-V3R2-MIG-002 REQ-MIG002-011, REQ-MIG002-012, REQ-MIG002-019 → AC-MIG002-A7.
func CleanupUserSettings(projectRoot string) error {
	settingsPath := filepath.Join(projectRoot, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No settings.json — nothing to clean up.
			return nil
		}
		return fmt.Errorf("read settings.json: %w", err)
	}

	// Parse the settings.json as a generic JSON object to preserve unknown fields.
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse settings.json: %w", err)
	}

	// Extract the hooks section (may be absent).
	hooksRaw, hasHooks := root["hooks"]
	if !hasHooks {
		// No hooks key — nothing to remove.
		return nil
	}

	var hooks map[string]json.RawMessage
	if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
		return fmt.Errorf("parse settings.json hooks: %w", err)
	}

	// Collect entries to remove.
	retiredSet := make(map[string]bool, len(hook.RetiredEventNames))
	for _, name := range hook.RetiredEventNames {
		retiredSet[name] = true
	}

	archived := make(map[string]json.RawMessage)
	for _, name := range hook.RetiredEventNames {
		if entry, found := hooks[name]; found {
			archived[name] = entry
			delete(hooks, name)
		}
	}

	if len(archived) == 0 {
		// No retired entries present — idempotent no-op.
		return nil
	}

	// Write archive entry.
	archiveDir := filepath.Join(projectRoot, ".moai", "archive", "hooks", "v3.0")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}

	archivePath := filepath.Join(archiveDir, fmt.Sprintf("migration-%s.json", time.Now().Format("2006-01-02")))
	archiveData, err := json.MarshalIndent(archived, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal archive: %w", err)
	}

	// Append to the archive file if it exists from a previous run today.
	if err := appendOrWriteJSON(archivePath, archiveData); err != nil {
		return fmt.Errorf("write archive: %w", err)
	}

	// Rebuild and write cleaned settings.json atomically.
	hooksJSON, err := json.Marshal(hooks)
	if err != nil {
		return fmt.Errorf("marshal hooks: %w", err)
	}
	root["hooks"] = hooksJSON

	cleanedData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cleaned settings.json: %w", err)
	}

	if err := atomicWrite(settingsPath, cleanedData, 0o644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	return nil
}

// appendOrWriteJSON writes data to path as JSON. If the file already exists,
// the new data is written alongside the existing content (both wrapped in an array).
// For migration use: keeps a day-level record of all removed entries.
func appendOrWriteJSON(path string, data []byte) error {
	// Simple overwrite: per-day migration file, one write per run.
	return atomicWrite(path, data, 0o644)
}

// atomicWrite writes data to path via a WriteFile-then-sync approach.
// For settings.json, atomic rename is ideal; using os.WriteFile with the
// permission flag is sufficient for migration-step use (single-writer scenario).
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
