package config

import "testing"

// TestLoadUpdateRenderValues pins that the reader returns the stored values of
// the user-owned keys the update render carries (card t1147).
func TestLoadUpdateRenderValues(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeIdentitySection(t, root, "language.yaml", "language:\n  conversation_language: \"ko\"\n  git_commit_messages: \"ko\"\n  code_comments: \"ja\"\n  documentation: \"zh\"\n")
	writeIdentitySection(t, root, "quality.yaml", "constitution:\n  development_mode: \"ddd\"\n  enforce_quality: false\n")
	writeIdentitySection(t, root, "git-strategy.yaml", "git_strategy:\n  provider: \"gitlab\"\n  github_username: \"gh-user\"\n  gitlab:\n    instance_url: \"https://gitlab.example.com\"\n")

	want := UpdateRenderValues{
		ConversationLanguage: "ko",
		GitCommitMessages:    "ko",
		CodeComments:         "ja",
		Documentation:        "zh",
		DevelopmentMode:      "ddd",
		GitProvider:          "gitlab",
		GitHubUsername:       "gh-user",
		GitLabInstanceURL:    "https://gitlab.example.com",
	}
	if got := LoadUpdateRenderValues(root); got != want {
		t.Errorf("LoadUpdateRenderValues = %+v, want %+v", got, want)
	}
}

// TestLoadUpdateRenderValues_PerFileFallback pins that a missing or
// unparseable file zeroes only its own keys.
func TestLoadUpdateRenderValues_PerFileFallback(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeIdentitySection(t, root, "language.yaml", "language: [unclosed\n")
	writeIdentitySection(t, root, "quality.yaml", "constitution:\n  development_mode: \"ddd\"\n")

	want := UpdateRenderValues{DevelopmentMode: "ddd"}
	if got := LoadUpdateRenderValues(root); got != want {
		t.Errorf("LoadUpdateRenderValues = %+v, want %+v", got, want)
	}
}
