package config

import (
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

func TestConfigStructCreation(t *testing.T) {
	t.Parallel()

	cfg := Config{
		User:     models.UserConfig{Name: "TestUser"},
		Language: models.LanguageConfig{ConversationLanguage: "ko"},
		Quality:  models.QualityConfig{DevelopmentMode: models.ModeDDD},
		Project:  models.ProjectConfig{},
		GitStrategy: GitStrategyConfig{
			AutoBranch:   true,
			BranchPrefix: "feature/",
			CommitStyle:  "conventional",
		},
		System: SystemConfig{
			Version:  "1.0.0",
			LogLevel: "debug",
		},
		LLM: LLMConfig{
			DefaultModel: "sonnet",
		},
		Pricing: PricingConfig{
			TokenBudget: 100000,
		},
		Ralph: RalphConfig{
			MaxIterations: 3,
		},
		Workflow: WorkflowConfig{
			PlanTokens: 30000,
		},
	}

	if cfg.User.Name != "TestUser" {
		t.Errorf("User.Name: got %q, want %q", cfg.User.Name, "TestUser")
	}
	if cfg.Language.ConversationLanguage != "ko" {
		t.Errorf("Language.ConversationLanguage: got %q, want %q", cfg.Language.ConversationLanguage, "ko")
	}
	if cfg.Quality.DevelopmentMode != models.ModeDDD {
		t.Errorf("Quality.DevelopmentMode: got %q, want %q", cfg.Quality.DevelopmentMode, models.ModeDDD)
	}
	if cfg.GitStrategy.AutoBranch != true {
		t.Error("GitStrategy.AutoBranch: expected true")
	}
	if cfg.GitStrategy.BranchPrefix != "feature/" {
		t.Errorf("GitStrategy.BranchPrefix: got %q, want %q", cfg.GitStrategy.BranchPrefix, "feature/")
	}
	if cfg.System.LogLevel != "debug" {
		t.Errorf("System.LogLevel: got %q, want %q", cfg.System.LogLevel, "debug")
	}
	if cfg.LLM.DefaultModel != "sonnet" {
		t.Errorf("LLM.DefaultModel: got %q, want %q", cfg.LLM.DefaultModel, "sonnet")
	}
	if cfg.Pricing.TokenBudget != 100000 {
		t.Errorf("Pricing.TokenBudget: got %d, want %d", cfg.Pricing.TokenBudget, 100000)
	}
	if cfg.Ralph.MaxIterations != 3 {
		t.Errorf("Ralph.MaxIterations: got %d, want %d", cfg.Ralph.MaxIterations, 3)
	}
	if cfg.Workflow.PlanTokens != 30000 {
		t.Errorf("Workflow.PlanTokens: got %d, want %d", cfg.Workflow.PlanTokens, 30000)
	}
}

func TestConfigZeroValue(t *testing.T) {
	t.Parallel()

	var cfg Config
	if cfg.User.Name != "" {
		t.Errorf("zero-value User.Name: got %q, want empty", cfg.User.Name)
	}
	if cfg.Quality.DevelopmentMode != "" {
		t.Errorf("zero-value Quality.DevelopmentMode: got %q, want empty", cfg.Quality.DevelopmentMode)
	}
	if cfg.GitStrategy.AutoBranch != false {
		t.Error("zero-value GitStrategy.AutoBranch: expected false")
	}
	if cfg.Pricing.TokenBudget != 0 {
		t.Errorf("zero-value Pricing.TokenBudget: got %d, want 0", cfg.Pricing.TokenBudget)
	}
}

func TestIsValidSectionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"user is valid", "user", true},
		{"language is valid", "language", true},
		{"quality is valid", "quality", true},
		{"project is valid", "project", true},
		{"git_strategy is valid", "git_strategy", true},
		{"git_convention is valid", "git_convention", true},
		{"system is valid", "system", true},
		{"llm is valid", "llm", true},
		{"pricing is valid", "pricing", true},
		{"ralph is valid", "ralph", true},
		{"workflow is valid", "workflow", true},
		{"empty string is invalid", "", false},
		{"unknown section is invalid", "unknown", false},
		{"User uppercase is invalid", "User", false},
		{"QUALITY uppercase is invalid", "QUALITY", false},
		{"git-strategy with hyphen is invalid", "git-strategy", false},
		{"space-padded is invalid", " user ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidSectionName(tt.input); got != tt.want {
				t.Errorf("IsValidSectionName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidSectionNames(t *testing.T) {
	t.Parallel()

	names := ValidSectionNames()

	// Verify count
	if len(names) != 11 {
		t.Fatalf("expected 11 section names, got %d", len(names))
	}

	// Verify all expected names are present
	expected := map[string]bool{
		"user": true, "language": true, "quality": true, "project": true,
		"git_strategy": true, "git_convention": true, "system": true, "llm": true,
		"pricing": true, "ralph": true, "workflow": true,
	}
	for _, name := range names {
		if !expected[name] {
			t.Errorf("unexpected section name: %q", name)
		}
	}
}

func TestValidSectionNamesReturnsCopy(t *testing.T) {
	t.Parallel()

	names1 := ValidSectionNames()
	names2 := ValidSectionNames()

	// Mutating one slice must not affect the other
	names1[0] = "MUTATED"
	if names2[0] == "MUTATED" {
		t.Error("ValidSectionNames() returned the same underlying slice, expected a copy")
	}
}

func TestGitStrategyConfigFields(t *testing.T) {
	t.Parallel()

	cfg := GitStrategyConfig{
		AutoBranch:        true,
		BranchPrefix:      "moai/",
		CommitStyle:       "conventional",
		WorktreeRoot:      "/tmp/worktree",
		Provider:          "gitlab",
		GitLabInstanceURL: "https://gitlab.company.com",
	}
	if !cfg.AutoBranch {
		t.Error("AutoBranch: expected true")
	}
	if cfg.BranchPrefix != "moai/" {
		t.Errorf("BranchPrefix: got %q, want %q", cfg.BranchPrefix, "moai/")
	}
	if cfg.CommitStyle != "conventional" {
		t.Errorf("CommitStyle: got %q, want %q", cfg.CommitStyle, "conventional")
	}
	if cfg.WorktreeRoot != "/tmp/worktree" {
		t.Errorf("WorktreeRoot: got %q, want %q", cfg.WorktreeRoot, "/tmp/worktree")
	}
	if cfg.Provider != "gitlab" {
		t.Errorf("Provider: got %q, want %q", cfg.Provider, "gitlab")
	}
	if cfg.GitLabInstanceURL != "https://gitlab.company.com" {
		t.Errorf("GitLabInstanceURL: got %q, want %q", cfg.GitLabInstanceURL, "https://gitlab.company.com")
	}
}

func TestSystemConfigFields(t *testing.T) {
	t.Parallel()

	cfg := SystemConfig{
		Version:        "2.0.0",
		LogLevel:       "warn",
		LogFormat:      "json",
		NoColor:        true,
		NonInteractive: true,
	}
	if cfg.Version != "2.0.0" {
		t.Errorf("Version: got %q, want %q", cfg.Version, "2.0.0")
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel: got %q, want %q", cfg.LogLevel, "warn")
	}
	if cfg.LogFormat != "json" {
		t.Errorf("LogFormat: got %q, want %q", cfg.LogFormat, "json")
	}
	if !cfg.NoColor {
		t.Error("NoColor: expected true")
	}
	if !cfg.NonInteractive {
		t.Error("NonInteractive: expected true")
	}
}

func TestLLMConfigFields(t *testing.T) {
	t.Parallel()

	cfg := LLMConfig{
		DefaultModel: "opus",
		QualityModel: "opus",
		SpeedModel:   "haiku",
	}
	if cfg.DefaultModel != "opus" {
		t.Errorf("DefaultModel: got %q, want %q", cfg.DefaultModel, "opus")
	}
	if cfg.QualityModel != "opus" {
		t.Errorf("QualityModel: got %q, want %q", cfg.QualityModel, "opus")
	}
	if cfg.SpeedModel != "haiku" {
		t.Errorf("SpeedModel: got %q, want %q", cfg.SpeedModel, "haiku")
	}
}

func TestPricingConfigFields(t *testing.T) {
	t.Parallel()

	cfg := PricingConfig{
		TokenBudget:  500000,
		CostTracking: true,
	}
	if cfg.TokenBudget != 500000 {
		t.Errorf("TokenBudget: got %d, want %d", cfg.TokenBudget, 500000)
	}
	if !cfg.CostTracking {
		t.Error("CostTracking: expected true")
	}
}

func TestRalphConfigFields(t *testing.T) {
	t.Parallel()

	cfg := RalphConfig{
		MaxIterations: 10,
		AutoConverge:  true,
		HumanReview:   false,
	}
	if cfg.MaxIterations != 10 {
		t.Errorf("MaxIterations: got %d, want %d", cfg.MaxIterations, 10)
	}
	if !cfg.AutoConverge {
		t.Error("AutoConverge: expected true")
	}
	if cfg.HumanReview {
		t.Error("HumanReview: expected false")
	}
}

func TestWorkflowConfigFields(t *testing.T) {
	t.Parallel()

	cfg := WorkflowConfig{
		AutoClear:  true,
		PlanTokens: 30000,
		RunTokens:  180000,
		SyncTokens: 40000,
	}
	if !cfg.AutoClear {
		t.Error("AutoClear: expected true")
	}
	if cfg.PlanTokens != 30000 {
		t.Errorf("PlanTokens: got %d, want %d", cfg.PlanTokens, 30000)
	}
	if cfg.RunTokens != 180000 {
		t.Errorf("RunTokens: got %d, want %d", cfg.RunTokens, 180000)
	}
	if cfg.SyncTokens != 40000 {
		t.Errorf("SyncTokens: got %d, want %d", cfg.SyncTokens, 40000)
	}
}

func TestLSPQualityGatesFields(t *testing.T) {
	t.Parallel()

	gates := LSPQualityGates{
		Enabled: true,
		Plan:    PlanGate{RequireBaseline: true},
		Run: RunGate{
			MaxErrors:       0,
			MaxTypeErrors:   0,
			MaxLintErrors:   0,
			AllowRegression: false,
		},
		Sync: SyncGate{
			MaxErrors:       0,
			MaxWarnings:     10,
			RequireCleanLSP: true,
		},
		CacheTTLSeconds: 5,
		TimeoutSeconds:  3,
	}

	if !gates.Enabled {
		t.Error("Enabled: expected true")
	}
	if !gates.Plan.RequireBaseline {
		t.Error("Plan.RequireBaseline: expected true")
	}
	if gates.Run.MaxErrors != 0 {
		t.Errorf("Run.MaxErrors: got %d, want 0", gates.Run.MaxErrors)
	}
	if gates.Run.AllowRegression {
		t.Error("Run.AllowRegression: expected false")
	}
	if gates.Sync.MaxWarnings != 10 {
		t.Errorf("Sync.MaxWarnings: got %d, want 10", gates.Sync.MaxWarnings)
	}
	if !gates.Sync.RequireCleanLSP {
		t.Error("Sync.RequireCleanLSP: expected true")
	}
	if gates.CacheTTLSeconds != 5 {
		t.Errorf("CacheTTLSeconds: got %d, want 5", gates.CacheTTLSeconds)
	}
	if gates.TimeoutSeconds != 3 {
		t.Errorf("TimeoutSeconds: got %d, want 3", gates.TimeoutSeconds)
	}
}
