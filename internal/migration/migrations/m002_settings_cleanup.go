package migrations

// @MX:NOTE - m002 settings-cleanup migration step. Relocated from the former
// internal/migrate.CleanupUserSettings by SPEC-DEADPKG-INVESTIGATE-001 (the
// internal/migrate package had a single function with zero callers). Folded here
// as a proper registry-registered migration following the m001 pattern. Removes
// stale RETIRE-OBS-ONLY event entries from the user's local .claude/settings.json
// (SPEC-V3R2-MIG-002). Idempotent + archival: no user data is deleted without a
// corresponding archive entry.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/migration"
)

// init registers m002 in the migration registry.
func init() {
	migration.Register(migration.Migration{
		Version:  2,
		Name:     "cleanup_retired_hook_events",
		Apply:    m002Apply,
		Rollback: nil, // archival cleanup; not rollback-able
	})
}

// m002Apply removes stale RETIRE-OBS-ONLY event entries from the user's local
// .claude/settings.json. This migration step is required for users upgrading from
// pre-SPEC-V3R2-RT-006 v3.0 builds whose settings.json still carries the 4 retired
// event hook registrations (Notification, Elicitation, ElicitationResult, TaskCreated).
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
func m002Apply(projectRoot string) error {
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

// appendOrWriteJSON writes data to path as JSON. For migration use it keeps a
// day-level record of removed entries (one write per run).
func appendOrWriteJSON(path string, data []byte) error {
	// Simple overwrite: per-day migration file, one write per run.
	return atomicWrite(path, data, 0o644)
}

// atomicWrite writes data to path atomically using a temp file + rename sequence.
//
// @MX:WARN: writes to user .claude/settings.json — partial writes corrupt configuration.
// @MX:REASON: both callers (settings.json write + archive write) depend on all-or-nothing
//             semantics; a plain os.WriteFile body would break the atomicity contract implied
//             by the function name. Pattern matches internal/runtime/persist.go (canonical
//             reference) and the other atomic writers in this codebase (config/manager.go,
//             manifest/manifest.go, template/deployer.go, harness/applier.go, harness/tier/tier.go).
//
// SPEC-V3R5-ATOMIC-WRITE-001 REQ-AWR-001..006 → AC-AWR-001..008.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".settings_cleanup_tmp_*")
	if err != nil {
		return fmt.Errorf("atomicWrite: create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomicWrite: write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomicWrite: sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomicWrite: close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomicWrite: rename: %w", err)
	}
	if err := os.Chmod(path, perm); err != nil {
		return fmt.Errorf("atomicWrite: chmod: %w", err)
	}
	return nil
}
