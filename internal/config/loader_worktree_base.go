package config

// loader_worktree_base.go — SPEC-WORKTREE-BASEREF-001 M1 (card t313).
//
// A single-key reader for git_strategy.worktree_base_branch, modelled on
// LoadSystemHookOptInEnabled (loader_system.go): both consumers of this SPEC
// run outside the Loader's lifecycle — the SessionStart alignment step
// (internal/hook) and moai's own worktree creation path (internal/cli) — and
// neither holds a loaded *Config at the point it needs the value.

import (
	"path/filepath"
	"strings"
)

// LoadWorktreeBaseBranch returns the configured card-worktree base branch for
// the project rooted at projectRoot, or the empty string when the key is
// absent, the file is missing, or the file cannot be parsed.
//
// Every failure path yields the empty string, which is the REQ-WBR-002 neutral
// "take no action" value — a project that cannot be read behaves exactly as a
// project that never configured the key. The returned value is trimmed
// (acceptance.md §D.2 edge case 1): a value carrying surrounding whitespace is
// never handed to git untrimmed.
//
// Callers that already hold a loaded *Config should read
// cfg.GitStrategy.WorktreeBaseBranch directly instead of re-reading the file.
func LoadWorktreeBaseBranch(projectRoot string) string {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &gitStrategyFileWrapper{}
	loaded, err := loadYAMLFile(dir, "git-strategy.yaml", wrapper)
	if err != nil || !loaded {
		return ""
	}
	return strings.TrimSpace(wrapper.GitStrategy.WorktreeBaseBranch)
}
