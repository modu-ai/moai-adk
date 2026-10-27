package cli

// init_profile_sync.go — the profile synchronization a completed init runs after its
// template deploy, shared by the healthy tail (init.go) and the resume dispatch
// (init_resume.go) so the two paths cannot drift (card t1591, Item D).

import (
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// syncInitProfile writes the profile preferences into the project config
// (.moai/config/sections/user.yaml and language.yaml). The wizard's answers for the
// user name and the conversation language take precedence over the profile values,
// the same precedence the template deploy gives opts. A nil wizardResult, or an empty
// answer, leaves the profile value in place.
func syncInitProfile(projectRoot string, prefs profile.ProfilePreferences, wizardResult *wizard.WizardResult) error {
	if wizardResult != nil {
		if wizardResult.ConversationLang != "" {
			prefs.ConversationLang = wizardResult.ConversationLang
		}
		if wizardResult.UserName != "" {
			prefs.UserName = wizardResult.UserName
		}
	}
	return profile.SyncToProjectConfig(projectRoot, prefs)
}
