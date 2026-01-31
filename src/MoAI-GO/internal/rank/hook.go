package rank

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// HookFileName is the name of the rank session tracking hook script
	HookFileName = "session_end__rank_submit.py"
)

// hookEntry represents a single hook configuration entry
type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// globalSettingsPath returns ~/.claude/settings.json
func globalSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// globalHookDir returns ~/.claude/hooks/moai/
func globalHookDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "hooks", "moai"), nil
}

// IsHookInstalled checks if the rank hook is installed in global settings
func IsHookInstalled() bool {
	settingsPath, err := globalSettingsPath()
	if err != nil {
		return false
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return false
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return false
	}

	hooksRaw, ok := raw["hooks"]
	if !ok {
		return false
	}

	var hooks map[string]json.RawMessage
	if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
		return false
	}

	sessionEndRaw, ok := hooks["SessionEnd"]
	if !ok {
		return false
	}

	var entries []hookEntry
	if err := json.Unmarshal(sessionEndRaw, &entries); err != nil {
		return false
	}

	hookName := HookFileName
	for _, entry := range entries {
		if containsHookRef(entry.Command, hookName) {
			return true
		}
	}
	return false
}

// containsHookRef checks if a command string references the hook file
func containsHookRef(command, hookName string) bool {
	return len(command) > 0 && (contains(command, hookName))
}

// contains is a simple substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// InstallHook installs the rank tracking hook into global Claude Code settings
func InstallHook() error {
	// Ensure hook directory exists
	hookDir, err := globalHookDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return fmt.Errorf("failed to create hook directory: %w", err)
	}

	// Write hook script
	hookPath := filepath.Join(hookDir, HookFileName)
	hookScript := generateHookScript()
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("failed to write hook script: %w", err)
	}

	// Update settings.json
	settingsPath, err := globalSettingsPath()
	if err != nil {
		return err
	}

	settings, err := readSettings(settingsPath)
	if err != nil {
		return err
	}

	hookCommand := buildHookCommand(hookPath)
	addSessionEndHook(settings, hookCommand)

	return writeSettings(settingsPath, settings)
}

// UninstallHook removes the rank tracking hook from global Claude Code settings
func UninstallHook() error {
	// Remove from settings.json
	settingsPath, err := globalSettingsPath()
	if err != nil {
		return err
	}

	settings, err := readSettings(settingsPath)
	if err != nil {
		// Settings file doesn't exist, nothing to uninstall
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	removeSessionEndHook(settings)

	if err := writeSettings(settingsPath, settings); err != nil {
		return err
	}

	// Remove hook script
	hookDir, dirErr := globalHookDir()
	if dirErr != nil {
		return nil
	}
	hookPath := filepath.Join(hookDir, HookFileName)
	if removeErr := os.Remove(hookPath); removeErr != nil && !os.IsNotExist(removeErr) {
		return fmt.Errorf("failed to remove hook script: %w", removeErr)
	}

	return nil
}

// readSettings reads and parses settings.json, returns empty map if not exists
func readSettings(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("failed to read settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}
	return settings, nil
}

// writeSettings writes settings.json with proper formatting
func writeSettings(path string, settings map[string]any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// buildHookCommand creates the platform-appropriate hook command
func buildHookCommand(hookPath string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("python \"%s\"", hookPath)
	}
	return fmt.Sprintf("${SHELL:-/bin/bash} -l -c 'python \"%s\"'", hookPath)
}

// addSessionEndHook adds the rank hook to SessionEnd hooks
func addSessionEndHook(settings map[string]any, command string) {
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}

	var entries []any
	if existing, ok := hooks["SessionEnd"].([]any); ok {
		entries = existing
	}

	// Check if already exists
	for _, e := range entries {
		if entry, ok := e.(map[string]any); ok {
			if cmd, ok := entry["command"].(string); ok {
				if containsHookRef(cmd, HookFileName) {
					return // Already installed
				}
			}
		}
	}

	entries = append(entries, map[string]any{
		"type":    "command",
		"command": command,
	})
	hooks["SessionEnd"] = entries
}

// removeSessionEndHook removes the rank hook from SessionEnd hooks
func removeSessionEndHook(settings map[string]any) {
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		return
	}

	entries, ok := hooks["SessionEnd"].([]any)
	if !ok {
		return
	}

	var filtered []any
	for _, e := range entries {
		if entry, ok := e.(map[string]any); ok {
			if cmd, ok := entry["command"].(string); ok {
				if containsHookRef(cmd, HookFileName) {
					continue // Remove this entry
				}
			}
		}
		filtered = append(filtered, e)
	}

	if len(filtered) == 0 {
		delete(hooks, "SessionEnd")
	} else {
		hooks["SessionEnd"] = filtered
	}
}

// generateHookScript creates a minimal Python hook script for session tracking
func generateHookScript() string {
	return `#!/usr/bin/env python3
"""MoAI Rank session tracking hook.
Called by Claude Code on SessionEnd to submit token usage.
"""
import json
import sys

def main():
    try:
        session_data = json.loads(sys.stdin.read())
    except (json.JSONDecodeError, Exception):
        return

    try:
        from moai_adk.rank.hook import submit_session_hook
        submit_session_hook(session_data)
    except ImportError:
        # moai_adk not installed, skip silently
        pass
    except Exception as e:
        print(f"rank hook error: {e}", file=sys.stderr)

if __name__ == "__main__":
    main()
`
}
