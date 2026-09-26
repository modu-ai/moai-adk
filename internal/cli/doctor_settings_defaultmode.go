package cli

// doctor_settings_defaultmode.go — card t1247.
//
// Advisory detection of a dead permissions.defaultMode override. Claude Code
// ignores a "bypassPermissions" defaultMode set in PROJECT- or LOCAL-scope
// settings (scope rule since v2.1.142; warned explicitly since v2.1.283 —
// `[WARN] settings defaultMode "bypassPermissions" ignored — only
// policy/user/flag settings may grant bypass mode (projectSettings and
// localSettings are repo-controllable)`), because both scopes are
// repo-controllable and must not grant bypass. The moai launcher delivers
// its effective permission mode through the --permission-mode flag, so a
// project/local defaultMode of "bypassPermissions" is dead configuration: it
// warns on every session start and grants nothing. The check reports it and
// points at the two sources that do work — the launcher flag or USER-scope
// settings (the same split internal/config/toolpolicy renders tiers with,
// tier_render.go AP-6).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/defs"
)

const settingsDefaultModeCheckName = "Settings DefaultMode"

// ignoredDefaultMode is the only mode this check flags. "auto" is subject to
// the same scope rule, but card t1247 scopes the detection to
// "bypassPermissions"; widening the set is a separate decision that would
// also have to update the launcher-flag guidance in the message.
const ignoredDefaultMode = "bypassPermissions"

// checkSettingsDefaultMode reports permissions.defaultMode="bypassPermissions"
// in the project (.claude/settings.json) or local (.claude/settings.local.json)
// settings — a value Claude Code silently ignores in those scopes.
func checkSettingsDefaultMode(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: settingsDefaultModeCheckName}

	var offenders []string
	for _, name := range []string{defs.SettingsJSON, defs.SettingsLocalJSON} {
		path := filepath.Join(projectRoot, defs.ClaudeDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			// Absent file: nothing this check owns. On a codex-only project
			// the claude surface is never deployed (REQ-IH-011 downgrade).
			continue
		}
		var doc struct {
			Permissions struct {
				DefaultMode string `json:"defaultMode"`
			} `json:"permissions"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			// Malformed JSON is another check's surface; skipping keeps this
			// check advisory-only about the one dead value it owns.
			continue
		}
		if doc.Permissions.DefaultMode == ignoredDefaultMode {
			offenders = append(offenders, name)
		}
	}

	if len(offenders) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "no ignored defaultMode override in project/local settings"
		return check
	}

	check.Status = uikit.CheckWarn
	check.Message = fmt.Sprintf(
		"%s set(s) permissions.defaultMode=%q, which Claude Code ignores in project/local scope — use the launcher flag (--permission-mode) or user-scope settings",
		strings.Join(offenders, ", "), ignoredDefaultMode)
	if verbose {
		check.Detail = "scope rule: only policy/user/flag settings may grant bypass mode; project/local are repo-controllable (Claude Code v2.1.142+)"
	}
	return check
}
