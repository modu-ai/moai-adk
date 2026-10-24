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

// LoadGitFlowDevelopBranch returns the configured integration branch for the
// project rooted at projectRoot, or the empty string when the key is absent,
// the active mode profile is not a local-integration workflow, or the file
// cannot be read. Under manual git-flow it is the develop_branch key; under
// manual agentic-trunk it is the projected main (REQ-AGT-025, card t1831) —
// the key is ignored there. The returned value is trimmed.
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
	// AgenticTrunkWorkflow reports that the ACTIVE mode profile's workflow is
	// agentic-trunk (card t1831, REQ-AGT-001). It takes the github-flow class
	// (valid, not git-flow) and the git-flow local-integration delivery route
	// with a fixed main target.
	AgenticTrunkWorkflow bool
	// DevelopBranch is the trimmed develop branch under manual git-flow, or
	// the projected main under manual agentic-trunk (REQ-AGT-025: the legacy
	// develop-branch seam, with the develop_branch key ignored). Empty in
	// every other state.
	DevelopBranch string

	// Workflow is the ACTIVE mode profile's raw workflow value (card t656,
	// SPEC-GITSTRAT-WORKFLOW-READER-001). Empty when the file is unreadable
	// or the mode has no active profile — which is "unknown", not invalid.
	Workflow string
	// Disposition is the three-way classification of Workflow
	// (loader_workflow_disposition.go). The zero value means unknown.
	Disposition WorkflowDisposition
	// Mode is the raw git_strategy.mode value (card t1453 M1b), set whenever
	// the file loaded; it names the profile a refusal tells the reader to fix
	// (EmptyTargetGuidance). Empty only when the file is unreadable or sets none.
	Mode string
	// IntegrationTarget is the flow-scoped integration target projected from
	// the D2 interpretation table: github-flow → "main", agentic-trunk → the
	// fixed "main" (REQ-AGT-003), git-flow → the gated DevelopBranch,
	// gitlab-flow → environment, release-flow → release_branch_prefix; empty
	// on invalid/unknown. git-flow keeps the Manual && GitFlowWorkflow gate
	// (REQ-GWS-005) — a non-manual git-flow profile resolves no target,
	// exactly as DevelopBranch stays empty today.
	IntegrationTarget string
	// MergeMethod is the ACTIVE mode profile's raw merge_method, trimmed (card
	// t1453 M2-B): the github-flow delivery edge maps it to the `gh pr merge`
	// method flag. Empty when the profile sets none or the file is unreadable;
	// the caller applies the squash default.
	MergeMethod string
}

// IsGitFlow reports whether the project's git strategy is git-flow: manual
// mode AND an active profile whose workflow is git-flow. An absent or
// unreadable file is not git-flow — neither half holds.
func (c GitFlowIntegrationConfig) IsGitFlow() bool {
	return c.Manual && c.GitFlowWorkflow
}

// UsesLocalIntegration reports whether the project lands cards through the
// local integration-window route (the dedicated integration worktree and
// its --no-ff merge): manual mode AND an active profile whose workflow is
// git-flow or agentic-trunk (REQ-AGT-005). It is the widened predicate the
// session-exit auto-merge and the acquire target read; IsGitFlow keeps its
// present answers for its own consumers.
func (c GitFlowIntegrationConfig) UsesLocalIntegration() bool {
	return c.Manual && (c.GitFlowWorkflow || c.AgenticTrunkWorkflow)
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
		Manual:               wrapper.GitStrategy.Mode == "manual",
		GitFlowWorkflow:      ok && profile.Workflow == WorkflowGitFlow,
		AgenticTrunkWorkflow: ok && profile.Workflow == WorkflowAgenticTrunk,
		Mode:                 wrapper.GitStrategy.Mode,
	}
	if ok {
		// Card t656: carry the raw value and its three-way disposition.
		// Classified only when a profile exists, so the zero disposition
		// stays reserved for "unknown" (unreadable file / no active profile).
		cfg.Workflow = profile.Workflow
		cfg.Disposition = ClassifyWorkflowDisposition(profile.Workflow)
		cfg.MergeMethod = strings.TrimSpace(profile.MergeMethod)
	}
	if cfg.Manual && cfg.GitFlowWorkflow {
		cfg.DevelopBranch = strings.TrimSpace(profile.DevelopBranch)
	}
	if cfg.Manual && cfg.AgenticTrunkWorkflow {
		// Card t1831 REQ-AGT-025: the legacy develop-branch seam projects to
		// main under manual agentic-trunk, whether or not the develop_branch
		// key is present (the key is ignored there — REQ-AGT-010).
		cfg.DevelopBranch = agenticTrunkFixedTarget
	}
	if ok {
		// Project the D2 interpretation table onto the seam. git-flow goes
		// through the gated DevelopBranch (empty when the manual gate fails
		// or the key is empty), preserving today's adoption contract.
		if profile.Workflow == WorkflowGitFlow {
			cfg.IntegrationTarget = cfg.DevelopBranch
		} else {
			cfg.IntegrationTarget = WorkflowIntegrationTarget(profile.Workflow, *profile)
		}
	}
	return cfg
}
