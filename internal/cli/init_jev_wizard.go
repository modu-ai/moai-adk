// Package cli — wizard -> settings Jev opt-in apply wiring
// (SPEC-JEV-OPTIN-MEASURE-001 REQ-JEVO-002).
//
// init_jev_wizard.go holds applyJevFromWizard: it routes the init wizard's
// Jev answer into the SHARED persistence seam that the `moai web` settings
// screen also drives. It is a sibling of init_autonomy_wizard.go and follows
// the same shape — a small, testable bridge rather than an inline block inside
// the init command.
//
// @MX:SPEC: SPEC-JEV-OPTIN-MEASURE-001
package cli

import (
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// applyJevFromWizard persists the wizard's Jev opt-in for the project at
// projectRoot.
//
// wizardRan is required rather than inferred. The wizard's zero value and its
// declined answer are the same bool, so without it a non-interactive `moai
// init --force` over an existing project would write `false` over a `true` the
// user set earlier — silently turning a capability off that nobody asked to
// turn off. When the wizard did not run, this writes nothing at all and the
// persisted value (or the shipped default) stands.
//
// When the wizard DID run the answer is applied in both directions: an opt-in
// enables, and a decline disables. A decline that matches the persisted value
// reaches the file as nothing — settings.ApplySchemaEdits' value-invariant
// write gate absorbs it — so the common case touches no bytes.
func applyJevFromWizard(wizardRan bool, res *wizard.WizardResult, projectRoot string) error {
	if !wizardRan || res == nil {
		return nil
	}
	return settings.SetJevEnabled(projectRoot, res.JevEnabled)
}
