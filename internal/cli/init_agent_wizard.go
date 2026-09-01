// Package cli — wizard→harness agent-wiring resolution
// (SPEC-INIT-HARNESS-PROMPT-001, spec.md §4 D1).
//
// init_agent_wizard.go holds resolveAgentWiringWithWizard: the SINGLE point at
// which the agent-harness selection is resolved, from either the --agent flag
// or the wizard's answer. It mirrors applyAutonomyTierFromWizard
// (init_autonomy_wizard.go, SPEC-INIT-WIZARD-REPAIR-001 REQ-005) minus the opts
// write — the harness selection persists nothing, so no InitOptions field is
// added. (Reversal condition: if the selection ever must be persisted to
// project config, add the field then and record the persistence consumer that
// justifies it.)
//
// Why a resolution point rather than a second reader: resolveAgentWiring takes
// *cobra.Command and reads only the flag, so before this change no wizard
// answer could reach EITHER of its two consumers — the Codex wiring call and
// the .mcp.json provisioning precedence switch. runInit now computes the value
// once into a local that both consumers read, which makes "the answer reaches
// both" structural rather than a convention (REQ-IHP-004).
//
// @MX:SPEC: SPEC-INIT-HARNESS-PROMPT-001
package cli

import (
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// resolveAgentWiringWithWizard resolves the harness selection, honouring
// flag-over-wizard precedence (REQ-IHP-002/003).
//
// Precedence:
//  1. When the --agent flag was explicitly set to a non-empty value
//     (flagChanged && flagValue != ""), it wins and the wizard selection is
//     discarded. BOTH conjuncts are required: the flag's cobra default is the
//     empty string, not claude, so `--agent ""` is "explicitly set and empty"
//     and must fall through rather than pin claude — otherwise the claude
//     fallback stops being attributable to the flag's absence. This matches
//     validateInitFlags, which short-circuits on agent != "".
//  2. Otherwise the wizard's selection decides. It is empty when the wizard did
//     not run (--non-interactive or no TTY), which resolves to claude.
//  3. An unrecognized value on either input falls back to claude rather than
//     erroring here — invalid flag values are rejected earlier, fail-loud, by
//     validateInitFlags (REQ-IHP-011), and the wizard's options are a closed set
//     by construction.
func resolveAgentWiringWithWizard(flagChanged bool, flagValue string, res *wizard.WizardResult) agentWiring {
	if flagChanged && flagValue != "" {
		return normalizeAgentWiring(flagValue)
	}
	if res == nil {
		return agentWiringClaude
	}
	return normalizeAgentWiring(res.AgentWiring)
}
