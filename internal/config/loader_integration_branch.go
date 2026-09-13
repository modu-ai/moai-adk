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
	if cfg.Manual && cfg.GitFlowWorkflow {
		cfg.DevelopBranch = strings.TrimSpace(profile.DevelopBranch)
	}
	return cfg
}
