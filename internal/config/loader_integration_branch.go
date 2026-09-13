package config

// loader_integration_branch.go — card t449.
//
// A single-key reader for the git-flow integration branch, modelled on
// LoadWorktreeBaseBranch (loader_worktree_base.go): the consumer — the
// `moai integration acquire` record — runs outside the Loader's lifecycle and
// holds no loaded *Config at the point it needs the value.
//
// The value is gated on the workflow discriminator rather than read raw:
// develop_branch is a manual-mode, git-flow-only key (types.go), so a
// github-flow project carrying a stale develop_branch must not have it
// adopted as its integration branch. Every failure path — missing file,
// unparseable file, non-git-flow workflow, no active profile, empty value —
// yields the empty string, the neutral "no configured integration branch"
// value that sends the caller to its fallback.

import (
	"path/filepath"
	"strings"
)

// gitFlowWorkflow is the git-strategy workflow discriminator for the
// manual-mode git-flow profile. develop_branch is meaningful only under it.
const gitFlowWorkflow = "git-flow"

// LoadGitFlowDevelopBranch returns the configured git-flow integration branch
// for the project rooted at projectRoot, or the empty string when the key is
// absent, the active mode profile is not git-flow, or the file cannot be
// read. The returned value is trimmed.
//
// Callers that already hold a loaded *Config should read
// cfg.GitStrategy.ActiveModeProfile() directly instead of re-reading the file.
func LoadGitFlowDevelopBranch(projectRoot string) string {
	return LoadGitFlowIntegrationConfig(projectRoot).DevelopBranch
}

// GitFlowIntegrationConfig separates "is this project git-flow" from "what is
// its develop branch" (card t637). LoadGitFlowDevelopBranch answers "" alike
// for a non-git-flow project, a git-flow project whose develop_branch is
// empty, and an absent or unreadable file; the acquire warning must tell the
// second case from the other two.
type GitFlowIntegrationConfig struct {
	// Manual reports that the git strategy mode is manual.
	Manual bool
	// GitFlowWorkflow reports that the ACTIVE mode profile's workflow is
	// git-flow.
	GitFlowWorkflow bool
	// DevelopBranch is the trimmed develop branch, non-empty only when the
	// project is git-flow (both halves above hold).
	DevelopBranch string

	// Workflow is the ACTIVE mode profile's raw workflow value (card t656,
	// SPEC-GITSTRAT-WORKFLOW-READER-001). Empty when the file is unreadable
	// or the mode has no active profile — which is "unknown", not invalid.
	Workflow string
	// Disposition is the three-way classification of Workflow
	// (loader_workflow_disposition.go). The zero value means unknown.
	Disposition WorkflowDisposition
	// IntegrationTarget is the flow-scoped integration target projected from
	// the D2 interpretation table: github-flow → "main", git-flow → the
	// gated DevelopBranch, gitlab-flow → environment, release-flow →
	// release_branch_prefix; empty on invalid/unknown. git-flow keeps the
	// Manual && GitFlowWorkflow gate (REQ-GWS-005) — a non-manual git-flow
	// profile resolves no target, exactly as DevelopBranch stays empty today.
	IntegrationTarget string
}

// IsGitFlow reports whether the project's git strategy is git-flow: manual
// mode AND an active profile whose workflow is git-flow. An absent or
// unreadable file is not git-flow — neither half holds.
func (c GitFlowIntegrationConfig) IsGitFlow() bool {
	return c.Manual && c.GitFlowWorkflow
}

// LoadGitFlowIntegrationConfig reads the git strategy once and reports both
// halves of the git-flow predicate plus the develop branch it gates. Every
// failure path — missing file, unparseable file — yields the zero value.
//
// @MX:ANCHOR: [AUTO] single read of git-strategy.yaml feeding three production consumers — integration acquire target, session-exit auto-merge gate, doctor Git Strategy Workflow check
// @MX:REASON: fan_in reached 3 when t656 added the doctor consumer; a signature or semantics change here moves all three callers (SPEC-GITSTRAT-WORKFLOW-READER-001)
func LoadGitFlowIntegrationConfig(projectRoot string) GitFlowIntegrationConfig {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &gitStrategyFileWrapper{}
	loaded, err := loadYAMLFile(dir, "git-strategy.yaml", wrapper)
	if err != nil || !loaded {
		return GitFlowIntegrationConfig{}
	}
	profile, ok := wrapper.GitStrategy.ActiveModeProfile()
	// The mode check is not redundant with the profile check: develop_branch is
	// a manual-mode key by contract, so a personal/team profile that happens to
	// carry a git-flow workflow and a develop_branch does not qualify.
	cfg := GitFlowIntegrationConfig{
		Manual:          wrapper.GitStrategy.Mode == "manual",
		GitFlowWorkflow: ok && profile.Workflow == gitFlowWorkflow,
	}
	if ok {
		// Card t656: carry the raw value and its three-way disposition.
		// Classified only when a profile exists, so the zero disposition
		// stays reserved for "unknown" (unreadable file / no active profile).
		cfg.Workflow = profile.Workflow
		cfg.Disposition = ClassifyWorkflowDisposition(profile.Workflow)
	}
	if cfg.Manual && cfg.GitFlowWorkflow {
		cfg.DevelopBranch = strings.TrimSpace(profile.DevelopBranch)
	}
	if ok {
		// Project the D2 interpretation table onto the seam. git-flow goes
		// through the gated DevelopBranch (empty when the manual gate fails
		// or the key is empty), preserving today's adoption contract.
		if profile.Workflow == gitFlowWorkflow {
			cfg.IntegrationTarget = cfg.DevelopBranch
		} else {
			cfg.IntegrationTarget = WorkflowIntegrationTarget(profile.Workflow, *profile)
		}
	}
	return cfg
}
