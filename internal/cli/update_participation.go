// Package cli — moai update participation ask (SPEC-FEEDBACK-PARTICIPATION-001
// REQ-ANON-004).
//
// update_participation.go owns runParticipationStep: the ONE prompt `moai
// update` may raise on its own. It runs only on a plain template-sync run, in
// an interactive terminal, with CI empty, and only when the user-scoped
// participation.asked value is still false — asked once, defaulting to no,
// persisting both values through the same consent writer the init wizard and
// the console drive.
//
// The prompt text is the init question's locale text (single source of truth);
// the mode-flag gate below classifies every flag of the update command so a
// new flag forces a decision (TestUpdateFlagInventoryClassified) instead of
// silently asking under an operation that is not a plain template sync.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/spf13/cobra"
)

// updateParticipationModeFlags are the update flags after which the ask never
// runs: each selects another operation than a plain template sync (or a
// non-interactive/auto-confirm mode, --yes included — consent is never
// inferred or auto-confirmed).
var updateParticipationModeFlags = []string{
	"check", "shell-env", "config", "binary", "dry-run",
	"restore", "version", "templates-only", "yes",
}

// updateParticipationModifierFlags are the flags that modify a plain
// template-sync run without changing what it is; the ask still runs under
// each.
var updateParticipationModifierFlags = []string{
	"force", "no-hooks", "no-plugin", "verbose", "profile",
}

// participationLocaleFn resolves the locale the prompt text is rendered in.
// It is a package-level indirection so tests pin it; production reads the
// profile's conversation language, the same source the init wizard pre-fills.
var participationLocaleFn = func() string {
	prefs, err := profile.ReadPreferences(profile.GetCurrentName())
	if err != nil || prefs.ConversationLang == "" {
		return "en"
	}
	return prefs.ConversationLang
}

// updateModeFlagActive reports whether any mode flag of the update command is
// set: a run carrying one is not a plain template sync, and nothing may ask.
func updateModeFlagActive(cmd *cobra.Command) bool {
	for _, name := range updateParticipationModeFlags {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			continue
		}
		switch f.Value.Type() {
		case "bool":
			if getBoolFlag(cmd, name) {
				return true
			}
		default:
			if getStringFlag(cmd, name) != "" {
				return true
			}
		}
	}
	return false
}

// runParticipationStep asks the participation question once on a plain
// template-sync update run that finishes in an interactive terminal with CI
// empty and no stored asked value. An empty answer persists enabled=false and
// asked=true, exactly like a decline: consent is never inferred, and the
// question is never re-asked.
//
// The error is reported to the caller but the update flow treats a failure
// here as a warning: an update must not fail because the ask could not be
// recorded (the capability stays off).
func runParticipationStep(cmd *cobra.Command, out io.Writer) error {
	if updateModeFlagActive(cmd) {
		return nil
	}
	if os.Getenv("CI") != "" {
		return nil
	}
	if !isInteractiveStdin() {
		return nil
	}
	if config.ReadUserParticipation().Asked {
		return nil
	}

	locale := participationLocaleFn()
	enQ := wizard.QuestionByID(wizard.InitQuestions("."), wizard.ParticipationQuestionID)
	if enQ == nil {
		return fmt.Errorf("participation question not found in the init set")
	}
	question := wizard.GetLocalizedQuestion(enQ, locale)

	// Default no: promptBool returns the default on empty input, and "n"/"no"
	// decline — every path lands on enabled=false; only an explicit yes
	// enables.
	answer := promptBool(out, question.Title+"\n"+question.Description, false)
	return settings.WriteUserParticipation(config.UserParticipation{
		Enabled: answer,
		Asked:   true,
	})
}
