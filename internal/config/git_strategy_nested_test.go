package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// gitStrategyNestedFixture is a synthetic, self-contained nested yaml fixture
// that mirrors the production yaml STRUCTURE (3-mode nested hierarchy) and is
// populated with the TEMPLATE-CANONICAL DEFAULT VALUES (the SSOT per spec.md
// §1.5 + REQ-GSS-006, sourced from
// internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl).
//
// Per acceptance.md AC-GSS-002: the production
// .moai/config/sections/git-strategy.yaml intentionally diverges as a 1-person
// OSS Hybrid Trunk customization (CLAUDE.local.md §22) and is NOT the fixture
// oracle. The values below MUST stay aligned with NewDefaultGitStrategyConfig()
// (M3) and the AC-GSS-005 defaults assertion so fixture/defaults/constructor
// form ONE consistent oracle.
const gitStrategyNestedFixture = `
git_strategy:
  mode: "team"
  provider: "github"
  github_username: ""
  gitlab:
    instance_url: ""

  manual:
    workflow: github-flow
    environment: local
    github_integration: false
    auto_checkpoint: disabled
    push_to_remote: false
    branch_creation:
      prompt_always: true
      auto_enabled: false
    automation:
      auto_branch: false
      auto_commit: true
      auto_pr: false
      auto_push: false
    hooks:
      pre_commit: enforce
      pre_push: warn
      commit_msg: warn
    commit_style:
      format: conventional
      scope_required: false

  personal:
    workflow: github-flow
    environment: github
    github_integration: true
    push_to_remote: true
    branch_prefix: "feature/SPEC-"
    main_branch: main
    branch_creation:
      prompt_always: true
      auto_enabled: false
    automation:
      auto_branch: false
      auto_commit: true
      auto_pr: false
      auto_push: false
    hooks:
      pre_commit: enforce
      pre_push: warn
      commit_msg: warn
    commit_style:
      format: conventional
      scope_required: false

  team:
    workflow: github-flow
    environment: github
    github_integration: true
    push_to_remote: true
    branch_prefix: "feature/SPEC-"
    main_branch: main
    draft_pr: true
    required_reviews: 1
    branch_protection: true
    branch_creation:
      prompt_always: true
      auto_enabled: false
    automation:
      auto_branch: false
      auto_commit: true
      auto_pr: false
      auto_push: true
    hooks:
      pre_commit: enforce
      pre_push: warn
      commit_msg: warn
    commit_style:
      format: conventional
      scope_required: true
`

// gitStrategyFixtureWrapper unmarshals the top-level git_strategy: key from the
// fixture into a GitStrategyConfig, mirroring the Config.GitStrategy yaml binding.
type gitStrategyFixtureWrapper struct {
	GitStrategy GitStrategyConfig `yaml:"git_strategy"`
}

// TestGitStrategyConfig_NestedRoundTrip verifies that yaml.Unmarshal populates
// every nested field of GitStrategyConfig from a nested yaml document matching
// the production structure (REQ-GSS-003, AC-GSS-002).
//
// The expected values are the template-canonical defaults — NOT the production
// yaml (which intentionally diverges per CLAUDE.local.md §22).
func TestGitStrategyConfig_NestedRoundTrip(t *testing.T) {
	t.Parallel()

	var wrapper gitStrategyFixtureWrapper
	if err := yaml.Unmarshal([]byte(gitStrategyNestedFixture), &wrapper); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	cfg := wrapper.GitStrategy

	// Top-level wire-through fields.
	if cfg.Mode != "team" {
		t.Errorf("Mode: got %q, want %q", cfg.Mode, "team")
	}
	if cfg.Provider != "github" {
		t.Errorf("Provider: got %q, want %q", cfg.Provider, "github")
	}
	if cfg.GitLab.InstanceURL != "" {
		t.Errorf("GitLab.InstanceURL: got %q, want empty", cfg.GitLab.InstanceURL)
	}

	// Team mode nested fields (Late-Branch defaults + automation + commit_style).
	if cfg.Team.BranchCreation.AutoEnabled != false {
		t.Errorf("Team.BranchCreation.AutoEnabled: got %v, want false", cfg.Team.BranchCreation.AutoEnabled)
	}
	if cfg.Team.BranchCreation.PromptAlways != true {
		t.Errorf("Team.BranchCreation.PromptAlways: got %v, want true", cfg.Team.BranchCreation.PromptAlways)
	}
	if cfg.Team.Automation.AutoBranch != false {
		t.Errorf("Team.Automation.AutoBranch: got %v, want false", cfg.Team.Automation.AutoBranch)
	}
	if cfg.Team.Automation.AutoPush != true {
		t.Errorf("Team.Automation.AutoPush: got %v, want true", cfg.Team.Automation.AutoPush)
	}
	if cfg.Team.CommitStyle.ScopeRequired != true {
		t.Errorf("Team.CommitStyle.ScopeRequired: got %v, want true", cfg.Team.CommitStyle.ScopeRequired)
	}
	if cfg.Team.DraftPR != true {
		t.Errorf("Team.DraftPR: got %v, want true", cfg.Team.DraftPR)
	}
	if cfg.Team.RequiredReviews != 1 {
		t.Errorf("Team.RequiredReviews: got %d, want 1", cfg.Team.RequiredReviews)
	}

	// Manual + Personal mode-conditional scalars.
	if cfg.Manual.Environment != "local" {
		t.Errorf("Manual.Environment: got %q, want %q", cfg.Manual.Environment, "local")
	}
	if cfg.Personal.BranchPrefix != "feature/SPEC-" {
		t.Errorf("Personal.BranchPrefix: got %q, want %q", cfg.Personal.BranchPrefix, "feature/SPEC-")
	}
}

// TestGitStrategyConfig_ZeroValueSafety verifies that a zero-value
// GitStrategyConfig does not panic when ActiveModeProfile() is called and
// returns the (nil, false) sentinel pair (R-GSS-002 mitigation, Edge-GSS-001).
func TestGitStrategyConfig_ZeroValueSafety(t *testing.T) {
	t.Parallel()

	cfg := GitStrategyConfig{}
	profile, ok := cfg.ActiveModeProfile()
	if ok {
		t.Errorf("zero-value ActiveModeProfile: ok = true, want false")
	}
	if profile != nil {
		t.Errorf("zero-value ActiveModeProfile: profile = %v, want nil", profile)
	}
}

// TestGitStrategyConfig_ActiveModeProfile verifies the ActiveModeProfile()
// accessor across all five mode cases (REQ-GSS-004, AC-GSS-003).
func TestGitStrategyConfig_ActiveModeProfile(t *testing.T) {
	t.Parallel()

	cfg := GitStrategyConfig{
		Manual:   ModeProfile{Environment: "local"},
		Personal: ModeProfile{Environment: "github", BranchPrefix: "feature/SPEC-"},
		Team:     ModeProfile{Environment: "github", DraftPR: true},
	}

	tests := []struct {
		name       string
		mode       string
		wantOK     bool
		wantTarget *ModeProfile
	}{
		{"manual", "manual", true, &cfg.Manual},
		{"personal", "personal", true, &cfg.Personal},
		{"team", "team", true, &cfg.Team},
		{"empty", "", false, nil},
		{"unknown", "unknown", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.Mode = tt.mode
			profile, ok := cfg.ActiveModeProfile()
			if ok != tt.wantOK {
				t.Errorf("ActiveModeProfile(%q): ok = %v, want %v", tt.mode, ok, tt.wantOK)
			}
			if profile != tt.wantTarget {
				t.Errorf("ActiveModeProfile(%q): profile pointer mismatch (got %p, want %p)", tt.mode, profile, tt.wantTarget)
			}
		})
	}
}
