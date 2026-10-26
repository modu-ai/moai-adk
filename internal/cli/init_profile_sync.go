package cli

// init_profile_sync.go — the profile synchronization a completed init runs after its
// template deploy, shared by the healthy tail (init.go) and the resume dispatch
// (init_resume.go) so the two paths cannot drift (card t1591, Item D).

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// syncInitProfile writes the profile preferences into the project config
// (.moai/config/sections/user.yaml and language.yaml). The wizard's answers for the
// user name and the conversation language take precedence over the profile values,
// the same precedence the template deploy gives opts. A nil wizardResult, or an empty
// answer that is not an explicit clear, leaves the profile value in place.
func syncInitProfile(projectRoot string, prefs profile.ProfilePreferences, wizardResult *wizard.WizardResult) error {
	clearName := false
	if wizardResult != nil {
		if wizardResult.ConversationLang != "" {
			prefs.ConversationLang = wizardResult.ConversationLang
		}
		if wizardResult.UserName != "" {
			prefs.UserName = wizardResult.UserName
		} else if wizardResult.UserNameCleared {
			// Card t1966: an explicit clear must not re-save the profile name.
			prefs.UserName = ""
			clearName = true
		}
	}
	if err := profile.SyncToProjectConfig(projectRoot, prefs); err != nil {
		return err
	}
	if clearName {
		return clearProjectUserName(projectRoot)
	}
	return nil
}

// clearProjectUserName blanks user.name in the project's user.yaml. SyncToProjectConfig
// only writes a non-empty name, and a resumed init redeploys no template, so a name
// the interrupted run left behind would survive an explicit clear (card t1966, F1).
// A missing user.yaml has nothing to clear.
func clearProjectUserName(projectRoot string) error {
	userPath := filepath.Join(projectRoot, ".moai", "config", "sections", "user.yaml")
	if _, err := os.Stat(userPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("clear user name: %w", err)
	}
	if err := yamlpatch.PatchFile(userPath, []yamlpatch.KeyEdit{{Path: []string{"user", "name"}, Value: ""}}); err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	return nil
}
