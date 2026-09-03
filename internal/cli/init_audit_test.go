package cli

// init_audit_test.go — SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 / AC-MCP-020).
//
// Verifies the audit + MCP opt-in wizard selection flows through
// applyWizardPage3ToOpts into project.InitOptions. The selection reuses the M3
// typed-config vocabulary (internal/config AuditModel* / AuditGate* constants).

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
)

// TestApplyWizardPage3ToOpts_AuditSelection pins AC-MCP-020: the six audit/MCP
// wizard answers map verbatim onto opts, and AuditConfigSet is asserted so the
// workflow.yaml writer persists the selection. The values reuse the M3 enum
// constants (no fork).
func TestApplyWizardPage3ToOpts_AuditSelection(t *testing.T) {
	cmd := newInitTestCmd()
	opts := seedOptsFromFlags(cmd)

	applyWizardPage3ToOpts(cmd, &wizard.WizardResult{
		// Page-3 Quality & Workflow (existing) — seeded so the helper runs its
		// full body.
		ProjectMode:        "team",
		WorktreeAutoCreate: true,
		// M4 audit selection (the surface under test).
		AuditModel:        config.AuditModelGLM,
		AuditGateClaude:   config.AuditGateAdvisory,
		AuditGateCodex:    config.AuditGateOff,
		AuditGateGLM:      config.AuditGateRequired,
		CodexAuditEnabled: true,
		MCPProvision:      true,
	}, &opts)

	if opts.AuditModel != config.AuditModelGLM {
		t.Errorf("AuditModel = %q, want %q", opts.AuditModel, config.AuditModelGLM)
	}
	if opts.AuditGateClaude != config.AuditGateAdvisory {
		t.Errorf("AuditGateClaude = %q, want %q", opts.AuditGateClaude, config.AuditGateAdvisory)
	}
	if opts.AuditGateCodex != config.AuditGateOff {
		t.Errorf("AuditGateCodex = %q, want %q", opts.AuditGateCodex, config.AuditGateOff)
	}
	if opts.AuditGateGLM != config.AuditGateRequired {
		t.Errorf("AuditGateGLM = %q, want %q", opts.AuditGateGLM, config.AuditGateRequired)
	}
	if !opts.CodexAuditEnabled {
		t.Error("CodexAuditEnabled = false, want true")
	}
	if !opts.MCPProvision {
		t.Error("MCPProvision = false, want true")
	}
	if !opts.AuditConfigSet {
		t.Error("AuditConfigSet = false, want true (wizard ran → writer must persist the audit block)")
	}
}

// TestApplyWizardPage3ToOpts_AuditConfigSetFalseByDefault pins the C6
// opt-in-default-off companion invariant: when the wizard did NOT run
// (--non-interactive, result absent), AuditConfigSet stays false so the
// workflow.yaml writer leaves the deployed template (no audit block) untouched.
// The helper itself never clears AuditConfigSet — the zero-value InitOptions
// already has it false; this test documents that the only path that sets it
// true is the wizard-collected apply (the test above).
func TestApplyWizardPage3ToOpts_AuditConfigSetFalseByDefault(t *testing.T) {
	cmd := newInitTestCmd()
	opts := seedOptsFromFlags(cmd)
	// No wizard result applied — mirrors the --non-interactive path.
	if opts.AuditConfigSet {
		t.Error("AuditConfigSet = true on a fresh opts, want false (wizard must opt it in)")
	}
}
