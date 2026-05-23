package cli

// @MX:NOTE: [AUTO] layout v2.1 helpers for moai init: compact header and
// Next-steps block. Replaces the full-screen PrintBanner + PrintWelcomeMessage
// pair so repeat invocations are not noisy.

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// renderInitHeader returns a compact accent box that introduces the init run.
// The header lists the target directory, project name (when provided), and
// the active profile summary so the user can confirm context at a glance
// before the wizard starts.
// nolint:unused // SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (init UX redesign helper)
func renderInitHeader(rootPath string, opts project.InitOptions, profileName string, prefs profile.ProfilePreferences) string {
	title := "moai init"
	if opts.ProjectName != "" {
		title += " · " + opts.ProjectName
	}

	pairs := []kvPair{
		{"Target", rootPath},
	}
	if profileName != "" {
		profileDesc := profileName
		extras := []string{}
		if prefs.UserName != "" {
			extras = append(extras, prefs.UserName)
		}
		if prefs.ConversationLang != "" {
			extras = append(extras, prefs.ConversationLang)
		}
		if len(extras) > 0 {
			profileDesc += " (" + strings.Join(extras, ", ") + ")"
		}
		pairs = append(pairs, kvPair{"Profile", profileDesc})
	}

	th := resolveTheme()
	body := "Initialize a new MoAI project in this directory."
	if details := renderKeyValueLines(pairs); details != "" {
		body += "\n\n" + details
	}
	return tui.Box(tui.BoxOpts{
		Title:  title,
		Body:   body,
		Width:  80,
		Theme:  &th,
		Accent: true,
	})
}

// renderInitNextSteps returns the "Next steps" block appended to the success
// card. The block guides the user toward the natural follow-up commands.
// nolint:unused // SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (init UX redesign helper)
func renderInitNextSteps(projectRoot, projectName string) string {
	cdTarget := projectName
	if cdTarget == "" {
		cdTarget = filepath.Base(projectRoot)
	}
	if cdTarget == "" || cdTarget == "." || cdTarget == "/" {
		cdTarget = ""
	}

	lines := []string{"", "Next steps", cliMuted.Render("──────────")}
	step := 1
	if cdTarget != "" {
		lines = append(lines, fmt.Sprintf("  %d. cd %s", step, cdTarget))
		step++
	}
	lines = append(lines,
		fmt.Sprintf("  %d. moai doctor", step),
		fmt.Sprintf("  %d. moai plan \"first feature\"", step+1),
	)
	return strings.Join(lines, "\n")
}
