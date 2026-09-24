package config

// loader_update_values.go — the user-owned section values, beyond the names in
// loader_identity.go, that init renders into language.yaml, quality.yaml, and
// git-strategy.yaml (card t1147). Their consumer is the update render path,
// which reads them before the managed cleanup removes .moai/config; without
// them every render falls back to the template default.
//
// Like the identity readers, the values are returned as stored; whether the
// render can carry one is decided by the caller against the real renderer.

import (
	"path/filepath"
)

// UpdateRenderValues holds the stored values of the user-owned keys the update
// render carries. An empty field means the key, its file, or a parseable file
// was absent.
type UpdateRenderValues struct {
	ConversationLanguage string
	GitCommitMessages    string
	CodeComments         string
	Documentation        string
	DevelopmentMode      string
	GitProvider          string
	GitHubUsername       string
	GitLabInstanceURL    string
}

// LoadUpdateRenderValues reads UpdateRenderValues for the project rooted at
// projectRoot. Each file is read on its own, so a missing or unparseable file
// empties only its own fields.
func LoadUpdateRenderValues(projectRoot string) UpdateRenderValues {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	var v UpdateRenderValues

	var lang struct {
		Language struct {
			ConversationLanguage string `yaml:"conversation_language"`
			GitCommitMessages    string `yaml:"git_commit_messages"`
			CodeComments         string `yaml:"code_comments"`
			Documentation        string `yaml:"documentation"`
		} `yaml:"language"`
	}
	if ok, err := loadYAMLFile(dir, "language.yaml", &lang); ok && err == nil {
		v.ConversationLanguage = lang.Language.ConversationLanguage
		v.GitCommitMessages = lang.Language.GitCommitMessages
		v.CodeComments = lang.Language.CodeComments
		v.Documentation = lang.Language.Documentation
	}

	var quality struct {
		Constitution struct {
			DevelopmentMode string `yaml:"development_mode"`
		} `yaml:"constitution"`
	}
	if ok, err := loadYAMLFile(dir, "quality.yaml", &quality); ok && err == nil {
		v.DevelopmentMode = quality.Constitution.DevelopmentMode
	}

	var git struct {
		GitStrategy struct {
			Provider       string `yaml:"provider"`
			GitHubUsername string `yaml:"github_username"`
			GitLab         struct {
				InstanceURL string `yaml:"instance_url"`
			} `yaml:"gitlab"`
		} `yaml:"git_strategy"`
	}
	if ok, err := loadYAMLFile(dir, "git-strategy.yaml", &git); ok && err == nil {
		v.GitProvider = git.GitStrategy.Provider
		v.GitHubUsername = git.GitStrategy.GitHubUsername
		v.GitLabInstanceURL = git.GitStrategy.GitLab.InstanceURL
	}
	return v
}
