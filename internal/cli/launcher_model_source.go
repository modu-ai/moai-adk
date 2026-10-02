package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// Launch model precedence on the Claude backend (card t1441, operator decision):
//
//  1. explicit --model argument
//  2. profile model (preferences.yaml, or settings.local.json DO_CLAUDE_MODEL)
//  3. the ANTHROPIC_MODEL environment variable (this launch only)
//  4. the project's .claude/settings.local.json "model" (this project only)
//  5. the model the user saved with /model in the user-scope settings.json of
//     the config dir this launch uses (CLAUDE_CONFIG_DIR, else ~/.claude)
//  6. nothing — Claude Code then applies the project's .claude/settings.json
//     model pin, which the launcher announces rather than hides
//
// Steps 1 and 2 are resolved by the argument and preference merge in
// runLaunchClaude. Levels 3 and 4 are applied by Claude Code itself and sit
// above a saved /model value, so the launcher passes nothing and stays out of
// their way. Level 5 is the gap this card closes: Claude Code ranks the project
// pin above user-scope settings, so without it the user's /model choice was
// hidden. The more specific and more recent the user's own selection, the
// higher it ranks.
//
// Levels 3 to 5 are skipped under a GLM backend: there the main-session model
// must be a slot alias that routes through the ANTHROPIC_DEFAULT_*_MODEL env
// configured by setGLMEnv, and a value picked with /model on an Anthropic
// account names a different model family, so passing it would route the
// session to the wrong backend slot.

// readSettingsModel returns the top-level "model" string of a settings JSON
// file, trimmed, or "" when the file is absent, unreadable, malformed, carries
// no string "model", or carries a blank one. Every failure falls through to the
// next precedence level; none surfaces as an error.
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
	return strings.TrimSpace(s.Model)
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

// readProjectModel returns the "model" of one of the project's settings files
// (defs.SettingsJSON for the committed pin, defs.SettingsLocalJSON for the
// project-local override), or "" when the root is unknown.
func readProjectModel(root, file string) string {
	if root == "" {
		return ""
	}
	return readSettingsModel(filepath.Join(root, defs.ClaudeDir, file))
}

// launchModelSetElsewhere reports whether the ANTHROPIC_MODEL environment
// variable or the project's settings.local.json already names a model. Claude
// Code ranks both above a saved /model value, so the launcher must neither
// override them nor announce a pin that does not decide the session.
func launchModelSetElsewhere(root string) bool {
	if strings.TrimSpace(os.Getenv(config.EnvAnthropicModel)) != "" {
		return true
	}
	return readProjectModel(root, defs.SettingsLocalJSON) != ""
}

// noteUserScopeModel tells the operator the launch took its model from /model.
func noteUserScopeModel(w io.Writer, model string) {
	_, _ = fmt.Fprintf(w, "model: %s (user /model)\n", model)
}

// warnProjectModelPin tells the operator that no model was resolved and the
// project pin decides the session model, and how to change that.
func warnProjectModelPin(w io.Writer, pinned string) {
	_, _ = fmt.Fprintf(w,
		"model: none resolved from --model, the profile, ANTHROPIC_MODEL, settings.local.json, or user /model — Claude Code will use the project pin \"model\": %q in .claude/settings.json.\n"+
			"  Change it with: moai profile setup (profile model), /model (saved for this profile), or \"model\" in .claude/settings.local.json\n", pinned)
}

// resolveLaunchModelFallback applies precedence levels 3 to 6 for a launch whose
// --model argument and profile model were both empty, and returns the model to
// pass to claude ("" means pass none). Notices go to w.
func resolveLaunchModelFallback(model string, glmBackend bool, root, profileName string, w io.Writer) string {
	if model != "" {
		return model
	}
	if glmBackend {
		if isNamedProfile(profileName) {
			warnNoModelResolved(w, profileName)
		}
		return ""
	}
	if launchModelSetElsewhere(root) {
		return ""
	}
	if userModel := readUserScopeModel(); userModel != "" {
		noteUserScopeModel(w, userModel)
		return userModel
	}
	if pinned := readProjectModel(root, defs.SettingsJSON); pinned != "" {
		warnProjectModelPin(w, pinned)
	} else if isNamedProfile(profileName) {
		warnNoModelResolved(w, profileName)
	}
	return ""
}
