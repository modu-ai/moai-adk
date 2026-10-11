// Package cli — wizard -> user-scoped consent participation apply wiring
// (SPEC-FEEDBACK-PARTICIPATION-001 REQ-ANON-003).
//
// init_participation_wizard.go holds applyParticipationFromWizard: it routes
// the init wizard's participation answer into the user-scoped consent writer
// that the `moai update` ask and the `moai web` console also drive. It is a
// sibling of init_jev_wizard.go and follows the same shape — a small, testable
// bridge rather than an inline block inside the init command.
//
// Two gates sit in front of the write, beyond the Jev bridge's one: the
// wizardRan gate (a non-interactive init must never write over a consent it
// never asked about) and the CI gate (REQ-ANON-003 — "the answer shall be
// persisted only when the wizard actually ran and the CI environment variable
// is empty"). A consent recorded by an unattended environment would not be a
// person's consent.
package cli

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// applyParticipationFromWizard persists the wizard's participation answer to
// the user-scoped consent file. It writes nothing when the wizard did not run,
// when the result is nil, or when CI is set; a decline is applied in both
// directions like the Jev bridge — an existing true becomes false.
func applyParticipationFromWizard(wizardRan bool, res *wizard.WizardResult, projectRoot string) error {
	if !wizardRan || res == nil {
		return nil
	}
	if os.Getenv("CI") != "" {
		return nil
	}
	_ = projectRoot // the consent file is user-scoped; no project path reaches it
	return settings.WriteUserParticipation(config.UserParticipation{
		Enabled: res.ParticipationEnabled,
		Asked:   true,
	})
}
