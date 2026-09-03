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
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &gitStrategyFileWrapper{}
	loaded, err := loadYAMLFile(dir, "git-strategy.yaml", wrapper)
	if err != nil || !loaded {
		return ""
	}
	profile, ok := wrapper.GitStrategy.ActiveModeProfile()
	// The mode check is not redundant with the profile check: develop_branch is
	// a manual-mode key by contract, so a personal/team profile that happens to
	// carry a git-flow workflow and a develop_branch does not qualify.
	if !ok || wrapper.GitStrategy.Mode != "manual" || profile.Workflow != gitFlowWorkflow {
		return ""
	}
	return strings.TrimSpace(profile.DevelopBranch)
}
