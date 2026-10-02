package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// Launch model precedence on the Claude backend (card t1441, operator decision):
//
//  1. explicit --model argument
//  2. profile model (preferences.yaml, or settings.local.json DO_CLAUDE_MODEL)
//  3. the model the user saved with /model in the user-scope settings.json of
//     the config dir this launch uses (CLAUDE_CONFIG_DIR, else ~/.claude)
//  4. nothing — Claude Code then applies the project's .claude/settings.json
//     model pin, which the launcher announces rather than hides
//
// Steps 1 and 2 are resolved by the existing argument and preference merge in
// runLaunchClaude; this file owns steps 3 and 4. Step 3 is skipped under a GLM
// backend: there the main-session model must be a slot alias that routes
// through the ANTHROPIC_DEFAULT_*_MODEL env configured by setGLMEnv, and a
// value picked with /model on an Anthropic account names a different model
// family, so passing it would route the session to the wrong backend slot.

// readSettingsModel returns the top-level "model" string of a settings JSON
// file, or "" when the file is absent, unreadable, malformed, or carries none.
func readSettingsModel(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var s struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return ""
	}
	return s.Model
}

// userScopeSettingsPath is the settings.json /model writes to for this launch:
// $CLAUDE_CONFIG_DIR/settings.json, or ~/.claude/settings.json when unset.
func userScopeSettingsPath() string {
	if dir := os.Getenv(config.EnvClaudeConfigDir); dir != "" {
		return filepath.Join(dir, defs.SettingsJSON)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", defs.SettingsJSON)
}

// readUserScopeModel returns the model saved with /model for this launch.
func readUserScopeModel() string {
	path := userScopeSettingsPath()
	if path == "" {
		return ""
	}
	return readSettingsModel(path)
}

// readProjectPinnedModel returns the model Claude Code will take from the
// project when nothing else decides: settings.local.json outranks
// settings.json, so a local value means no project pin is in play.
func readProjectPinnedModel(root string) string {
	if root == "" {
		return ""
	}
	claudeDir := filepath.Join(root, defs.ClaudeDir)
	if readSettingsModel(filepath.Join(claudeDir, defs.SettingsLocalJSON)) != "" {
		return ""
	}
	return readSettingsModel(filepath.Join(claudeDir, defs.SettingsJSON))
}

// noteUserScopeModel tells the operator the launch took its model from /model.
func noteUserScopeModel(w io.Writer, model string) {
	_, _ = fmt.Fprintf(w, "model: %s (user /model)\n", model)
}

// warnProjectModelPin tells the operator that no model was resolved and the
// project pin decides the session model, and how to change that.
func warnProjectModelPin(w io.Writer, pinned string) {
	_, _ = fmt.Fprintf(w,
		"model: none resolved from --model, the profile, or user /model — Claude Code will use the project pin \"model\": %q in .claude/settings.json.\n"+
			"  Change it with: moai profile setup (profile model), /model (saved for this profile), or \"model\" in .claude/settings.local.json\n", pinned)
}
