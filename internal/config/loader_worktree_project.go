package config

// loader_worktree_project_key.go — SPEC-WORKTREE-HOME-ROOT-001 M1.
//
// A single-key reader for git_strategy.worktree_project_key, modelled on
// LoadWorktreeBaseBranch (loader_worktree_base.go): the consumers of this
// key that run outside the Loader's lifecycle read it straight from the
// project's git-strategy.yaml. The resolver keeps its own import-light read
// (internal/worktreeroot); this helper is for config-package consumers that
// already import this package.

import (
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadWorktreeProjectKey returns the configured worktree project-key pin for
// the project rooted at projectRoot, or the empty string when the key is
// absent, the file is missing, or the file cannot be parsed. Every failure
// path yields the empty string — the neutral "derive the key from the
// primary checkout's basename" value (REQ-WHR-004). The returned value is
// trimmed; validation of the segment shape belongs to the resolver, which
// rejects a pin that is not one safe path segment.
func LoadWorktreeProjectKey(projectRoot string) string {
	data, err := readNonBlockingRegularFile(filepath.Join(projectRoot, ".moai", "config", "sections", "git-strategy.yaml"))
	if err != nil {
		return ""
	}
	var wrapper gitStrategyFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return ""
	}
	return strings.TrimSpace(wrapper.GitStrategy.WorktreeProjectKey)
}
