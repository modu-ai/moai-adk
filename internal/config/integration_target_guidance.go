package config

// integration_target_guidance.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1b (card t1453).
//
// The base-branch consumers (worktree done / sweep, goal approve, factory
// merge ready / complete, the codex card-scope base) refuse an empty
// integration target rather than substituting a branch. A refusal that does
// not say WHAT TO SET is itself a defect, so every one of those sites renders
// its reason and fix through this one builder, driven by the state the loader
// already holds. The cause classification sits next to the interpretation
// table (loader_workflow_disposition.go) and reuses its disposition and
// allowed-set rather than re-deriving them.

import (
	"fmt"
	"path/filepath"
	"strings"
)

// gitStrategyRelPath is the one file the integration target is read from.
const gitStrategyRelPath = ".moai/config/sections/git-strategy.yaml"

// EmptyTargetCause names why a project resolves no integration target. The
// zero value means the target resolved.
type EmptyTargetCause string

const (
	// EmptyTargetUnreadable: the file is absent, not valid YAML, or sets no mode.
	EmptyTargetUnreadable EmptyTargetCause = "unreadable"
	// EmptyTargetUnknownMode: git_strategy.mode selects no profile.
	EmptyTargetUnknownMode EmptyTargetCause = "unknown-mode"
	// EmptyTargetUnknownWorkflow: the workflow value is outside the allowed set.
	EmptyTargetUnknownWorkflow EmptyTargetCause = "unknown-workflow"
	// EmptyTargetGitFlowNotManual: git-flow resolves a target only in manual mode.
	EmptyTargetGitFlowNotManual EmptyTargetCause = "git-flow-not-manual"
	// EmptyTargetDevelopEmpty: git-flow in manual mode with an empty develop_branch.
	EmptyTargetDevelopEmpty EmptyTargetCause = "develop-branch-empty"
	// EmptyTargetEnvironmentEmpty: gitlab-flow with an empty environment.
	EmptyTargetEnvironmentEmpty EmptyTargetCause = "environment-empty"
	// EmptyTargetReleasePrefixEmpty: release-flow with an empty release_branch_prefix.
	EmptyTargetReleasePrefixEmpty EmptyTargetCause = "release-branch-prefix-empty"
)

// EmptyTargetCause classifies why IntegrationTarget is empty; "" when it
// resolved. Mode is set whenever the file loaded, Disposition whenever the
// mode selects a profile — the two together separate every empty cause.
func (c GitFlowIntegrationConfig) EmptyTargetCause() EmptyTargetCause {
	if c.IntegrationTarget != "" {
		return ""
	}
	if c.Disposition == "" {
		if c.Mode == "" {
			return EmptyTargetUnreadable
		}
		return EmptyTargetUnknownMode
	}
	if c.Disposition == DispositionInvalid {
		return EmptyTargetUnknownWorkflow
	}
	switch c.Workflow {
	case WorkflowGitFlow:
		if !c.Manual {
			return EmptyTargetGitFlowNotManual
		}
		return EmptyTargetDevelopEmpty
	case WorkflowGitLabFlow:
		return EmptyTargetEnvironmentEmpty
	case WorkflowReleaseFlow:
		return EmptyTargetReleasePrefixEmpty
	}
	return ""
}

// EmptyTargetGuidance renders the reason an integration target is empty and
// the fix, ending in the calling site's own escape hatch (tail, e.g. "pass
// --base origin/<branch>"; "" when the site has none). It returns "" when the
// target resolved — the caller reaches it only on an empty target. root is the
// project root the config was read from; it appears only in the unreadable
// case, where the file path is the thing to fix.
func (c GitFlowIntegrationConfig) EmptyTargetGuidance(root, tail string) string {
	var msg string
	switch c.EmptyTargetCause() {
	case "":
		return ""
	case EmptyTargetUnreadable:
		msg = fmt.Sprintf("cannot read %s (absent or not valid YAML) or it sets no git_strategy.mode; restore the file",
			filepath.Join(root, filepath.FromSlash(gitStrategyRelPath)))
	case EmptyTargetUnknownMode:
		msg = fmt.Sprintf("git_strategy.mode %q is not one of manual, personal, team; correct it in %s", c.Mode, gitStrategyRelPath)
	case EmptyTargetUnknownWorkflow:
		msg = fmt.Sprintf("git_strategy.%s.workflow %q is not one of %s; correct it in %s",
			c.Mode, c.Workflow, strings.Join(AllowedWorkflows(), ", "), gitStrategyRelPath)
	case EmptyTargetGitFlowNotManual:
		msg = fmt.Sprintf("git-flow resolves a target only when git_strategy.mode is manual (mode is %s); set git_strategy.mode: manual in %s, or set git_strategy.%s.workflow: github-flow (target main)",
			c.Mode, gitStrategyRelPath, c.Mode)
	case EmptyTargetDevelopEmpty:
		msg = fmt.Sprintf("git_strategy.%s.develop_branch is empty; set it (for example `develop`) in %s, or set git_strategy.%s.workflow: github-flow (target main)",
			c.Mode, gitStrategyRelPath, c.Mode)
	case EmptyTargetEnvironmentEmpty:
		msg = fmt.Sprintf("git_strategy.%s.environment is empty; set it to the integration branch in %s", c.Mode, gitStrategyRelPath)
	case EmptyTargetReleasePrefixEmpty:
		msg = fmt.Sprintf("git_strategy.%s.release_branch_prefix is empty; set it in %s", c.Mode, gitStrategyRelPath)
	}
	if tail != "" {
		msg += ", or " + tail
	}
	return msg
}

// TargetProvenance names the config value the resolved target came from, in
// the form `git_strategy.<mode>.workflow=<workflow>`, so a refusal built on the
// target reads as a config fact. "" when no profile was active.
func (c GitFlowIntegrationConfig) TargetProvenance() string {
	if c.Disposition == "" {
		return ""
	}
	return fmt.Sprintf("git_strategy.%s.workflow=%s", c.Mode, c.Workflow)
}
