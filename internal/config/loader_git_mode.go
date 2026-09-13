package config

// loader_git_mode.go — a single-key reader for git_strategy.mode, modelled on
// LoadWorktreeBaseBranch (loader_worktree_base.go). Its consumer is the update
// render path, which builds a template context outside the Loader's lifecycle
// and, on the template-sync path, must read the value before the managed
// cleanup removes .moai/config.

import (
	"path/filepath"
	"strings"
)

// LoadGitMode returns git_strategy.mode for the project rooted at projectRoot,
// or the empty string when the key is absent, the file is missing, or the file
// cannot be parsed.
//
// The value is trimmed and returned as written; validation belongs to the
// consumer. template.WithGitMode ignores anything other than manual, personal,
// or team, so an unreadable or unknown mode leaves the template default in
// place rather than rendering an invalid mode.
func LoadGitMode(projectRoot string) string {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &gitStrategyFileWrapper{}
	loaded, err := loadYAMLFile(dir, "git-strategy.yaml", wrapper)
	if err != nil || !loaded {
		return ""
	}
	return strings.TrimSpace(wrapper.GitStrategy.Mode)
}
