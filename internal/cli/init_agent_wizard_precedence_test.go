package cli

// SPEC-INIT-HARNESS-PROMPT-001 — the precedence RULE, asserted directly.
//
// The end-to-end rows in init_agent_wizard_test.go cannot carry the --agent
// claude case on their own: its observable outcome is identical before and
// after this change, because the wizard answer used to be discarded
// unconditionally. Nothing in a file assertion distinguishes "the flag won"
// from "the wizard was never read". This file asserts the helper's return.
//
// @MX:SPEC: SPEC-INIT-HARNESS-PROMPT-001

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// TestResolveAgentWiringWithWizard_PrecedenceTable asserts AC-IHP-004 and the
// AC-IHP-003 resolution half: the flag wins when explicitly set to a non-empty
// value, the wizard answer is used otherwise, and claude is the fallback when
// neither speaks.
//
// The flag branch requires BOTH conjuncts — flagChanged AND a non-empty value —
// matching applyAutonomyTierFromWizard and consistent with validateInitFlags,
// which short-circuits on agent != "". The cobra default is the empty string,
// not claude, so `--agent ""` is explicitly set and empty: it must fall through
// to the wizard rather than pin claude, or the fallback stops being
// attributable to absence.
func TestResolveAgentWiringWithWizard_PrecedenceTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		flagChanged bool
		flagValue   string
		wizard      string
		want        agentWiring
	}{
		// Flag absent: the wizard decides.
		{"flag absent, wizard codex", false, "", "codex", agentWiringCodex},
		{"flag absent, wizard both", false, "", "both", agentWiringBoth},
		{"flag absent, wizard claude", false, "", "claude", agentWiringClaude},
		{"flag absent, wizard silent", false, "", "", agentWiringClaude},

		// Flag set: it wins and the wizard answer is discarded. The claude row
		// is the one whose OUTCOME is vacuous end-to-end; here the rule is
		// visible because the wizard said something different.
		{"flag claude beats wizard codex", true, "claude", "codex", agentWiringClaude},
		{"flag codex beats wizard claude", true, "codex", "claude", agentWiringCodex},
		{"flag both beats wizard codex", true, "both", "codex", agentWiringBoth},
		{"flag claude beats wizard both", true, "claude", "both", agentWiringClaude},

		// Explicitly set and empty: not a flag win — fall through to the wizard.
		{"flag set empty, wizard codex", true, "", "codex", agentWiringCodex},
		{"flag set empty, wizard silent", true, "", "", agentWiringClaude},

		// An unrecognized value falls back to claude, exactly as the flag-only
		// primitive always did (invalid values are rejected earlier, fail-loud,
		// by validateInitFlags — REQ-IHP-011).
		{"unrecognized flag value falls back", true, "gemini", "codex", agentWiringClaude},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res := &wizard.WizardResult{AgentWiring: tc.wizard}
			got := resolveAgentWiringWithWizard(tc.flagChanged, tc.flagValue, res)
			if got != tc.want {
				t.Errorf("resolveAgentWiringWithWizard(changed=%v, value=%q, wizard=%q) = %q, want %q",
					tc.flagChanged, tc.flagValue, tc.wizard, got, tc.want)
			}
		})
	}
}

// TestResolveAgentWiringWithWizard_NilResultIsClaude pins the defensive case:
// runInit holds a non-nil empty WizardResult when the wizard did not run, but
// the helper must not panic if that ever changes.
func TestResolveAgentWiringWithWizard_NilResultIsClaude(t *testing.T) {
	t.Parallel()
	if got := resolveAgentWiringWithWizard(false, "", nil); got != agentWiringClaude {
		t.Errorf("resolveAgentWiringWithWizard(_, _, nil) = %q, want %q", got, agentWiringClaude)
	}
}

// TestMCPPrecedenceComment_StatesHarnessRule asserts AC-IHP-010: the
// justification "the flag beats the wizard answer" is falsified once harness is
// itself a wizard axis — TestRunInit_WizardCodexDeclinesMCPProvisioning is a
// wizard answer overriding a wizard answer, with no flag present — so it must
// not survive verbatim, and the true rule must be stated in its place.
//
// This is a COMPANION to the behavioural test, never a substitute: a comment
// cannot gate behaviour. The behavioural test is the binding gate.
func TestMCPPrecedenceComment_StatesHarnessRule(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	body := string(src)

	if strings.Contains(body, "the flag beats the wizard answer") {
		t.Error(`init.go still carries the falsified justification "the flag beats the wizard answer" — a wizard harness answer now overrides a wizard mcp_provision answer with no flag present (REQ-IHP-010)`)
	}
	if !strings.Contains(body, "harness selection") {
		t.Error(`init.go must state the REQ-IHP-009 rule: the harness selection — from flag or wizard — decides provisioning`)
	}
}
