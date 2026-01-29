package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Test 1: Load config from non-existent directory returns defaults
	t.Run("NonExistentDirectory", func(t *testing.T) {
		cfg, err := LoadConfig(tempDir)
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		if cfg == nil {
			t.Fatal("LoadConfig() returned nil config")
		}

		// Check default values
		if cfg.Language.ConversationLanguage != "en" {
			t.Errorf("default conversation_language = %s, want %s", cfg.Language.ConversationLanguage, "en")
		}

		if cfg.Constitution.TestCoverageTarget != 85 {
			t.Errorf("default test_coverage_target = %d, want %d", cfg.Constitution.TestCoverageTarget, 85)
		}
	})

	// Test 2: Load config from existing directory with valid YAML
	t.Run("ValidYAML", func(t *testing.T) {
		// Create .moai/config/sections directory
		configPath := filepath.Join(tempDir, ".moai", "config", "sections")
		if err := os.MkdirAll(configPath, 0755); err != nil {
			t.Fatalf("failed to create config directory: %v", err)
		}

		// Create user.yaml
		userYAML := `user:
  name: "Test User"
`
		if err := os.WriteFile(filepath.Join(configPath, "user.yaml"), []byte(userYAML), 0644); err != nil {
			t.Fatalf("failed to write user.yaml: %v", err)
		}

		// Create language.yaml
		languageYAML := `language:
  conversation_language: ko
  code_comments: en
`
		if err := os.WriteFile(filepath.Join(configPath, "language.yaml"), []byte(languageYAML), 0644); err != nil {
			t.Fatalf("failed to write language.yaml: %v", err)
		}

		// Load config
		cfg, err := LoadConfig(tempDir)
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		// Check loaded values
		if cfg.User.Name != "Test User" {
			t.Errorf("user.name = %s, want 'Test User'", cfg.User.Name)
		}

		if cfg.Language.ConversationLanguage != "ko" {
			t.Errorf("language.conversation_language = %s, want 'ko'", cfg.Language.ConversationLanguage)
		}
	})
}

func TestEnvironmentVariableOverrides(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create config directory
	configPath := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	// Create user.yaml with default value
	userYAML := `user:
  name: "File User"
`
	if err := os.WriteFile(filepath.Join(configPath, "user.yaml"), []byte(userYAML), 0644); err != nil {
		t.Fatalf("failed to write user.yaml: %v", err)
	}

	// Test 1: Environment variable overrides file value
	t.Run("EnvVarOverridesFile", func(t *testing.T) {
		// Set environment variable
		oldVal := os.Getenv("MOAI_USER_NAME")
		_ = os.Setenv("MOAI_USER_NAME", "Env User")
		defer func() {
			if oldVal == "" {
				_ = os.Unsetenv("MOAI_USER_NAME")
			} else {
				_ = os.Setenv("MOAI_USER_NAME", oldVal)
			}
		}()

		cfg, err := LoadConfig(tempDir)
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		if cfg.User.Name != "Env User" {
			t.Errorf("user.name = %s, want 'Env User' (from env var)", cfg.User.Name)
		}
	})

	// Test 2: Language environment variables
	t.Run("LanguageEnvVars", func(t *testing.T) {
		oldConv := os.Getenv("MOAI_CONVERSATION_LANG")
		oldComments := os.Getenv("MOAI_CODE_COMMENTS_LANG")
		_ = os.Setenv("MOAI_CONVERSATION_LANG", "ko")
		_ = os.Setenv("MOAI_CODE_COMMENTS_LANG", "en")
		defer func() {
			if oldConv == "" {
				_ = os.Unsetenv("MOAI_CONVERSATION_LANG")
			} else {
				_ = os.Setenv("MOAI_CONVERSATION_LANG", oldConv)
			}
			if oldComments == "" {
				_ = os.Unsetenv("MOAI_CODE_COMMENTS_LANG")
			} else {
				_ = os.Setenv("MOAI_CODE_COMMENTS_LANG", oldComments)
			}
		}()

		cfg, err := LoadConfig(tempDir)
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		if cfg.Language.ConversationLanguage != "ko" {
			t.Errorf("language.conversation_language = %s, want 'ko'", cfg.Language.ConversationLanguage)
		}

		if cfg.Language.CodeComments != "en" {
			t.Errorf("language.code_comments = %s, want 'en'", cfg.Language.CodeComments)
		}
	})
}

func TestValidateConfig(t *testing.T) {
	// Test 1: Valid config passes validation
	t.Run("ValidConfig", func(t *testing.T) {
		cfg := &Config{
			User:     UserConfig{Name: "Test"},
			Language: LanguageConfig{ConversationLanguage: "en"},
			Constitution: ConstitutionConfig{
				DevelopmentMode:    "ddd",
				TestCoverageTarget: 85,
				DDDDSettings: DDDSettings{
					MaxTransformationSize: "small",
				},
			},
			System: SystemConfig{
				MoAI: MoAIConfig{
					UpdateCheckFrequency: "daily",
				},
				GitHub: GitHubConfig{
					SpecGitWorkflow: "main_direct",
				},
			},
			GitStrategy: GitStrategyConfig{
				Mode: "manual",
				Manual: GitModeConfig{
					Workflow:       "github-flow",
					CommitStyle:    CommitStyleConfig{Format: "conventional"},
					BranchCreation: BranchCreationConfig{},
					Automation:     AutomationConfig{},
					Hooks:          HooksConfig{},
				},
				Personal: GitModeConfig{
					Workflow:       "github-flow",
					CommitStyle:    CommitStyleConfig{Format: "conventional"},
					BranchCreation: BranchCreationConfig{},
					Automation:     AutomationConfig{},
					Hooks:          HooksConfig{},
				},
				Team: GitTeamConfig{
					Workflow:       "github-flow",
					BranchCreation: BranchCreationConfig{},
					Automation:     AutomationConfig{},
					Hooks:          HooksConfig{},
					CommitStyle:    CommitStyleConfig{Format: "conventional"},
				},
			},
			LLM: LLMConfig{
				Mode: "claude-only",
				GLM: GLMConfig{
					Models: map[string]string{
						"haiku": "glm-4.7-flashx",
					},
				},
				Routing: RoutingConfig{},
			},
			Service: ServiceConfig{
				Type:        "claude_subscription",
				PricingPlan: "pro",
				ModelAllocation: ModelAllocationConfig{
					Strategy: "auto",
				},
			},
			Ralph: RalphConfig{
				LSP: RalphLSPConfig{
					TimeoutSeconds: 15,
					PollIntervalMs: 1000,
				},
				Loop: RalphLoopConfig{
					MaxIterations: 10,
					Completion: RalphCompletionConfig{
						CoverageThreshold: 85,
					},
				},
				ASTGrep: RalphASTGrepConfig{},
				Hooks:   RalphHooksConfig{},
			},
			Workflow: WorkflowConfig{
				LoopPrevention: LoopPreventionConfig{
					MaxIterations:       100,
					NoProgressThreshold: 5,
				},
				ExecutionMode:     WorkflowExecutionConfig{},
				CompletionMarkers: CompletionMarkersConfig{},
			},
		}

		err := ValidateConfig(cfg)
		if err != nil {
			t.Errorf("ValidateConfig() failed: %v", err)
		}
	})

	// Test 2: Invalid language code fails validation
	t.Run("InvalidLanguageCode", func(t *testing.T) {
		cfg := &Config{
			Language: LanguageConfig{ConversationLanguage: "invalid"},
		}

		err := ValidateConfig(cfg)
		if err == nil {
			t.Error("ValidateConfig() expected error for invalid language code, got nil")
		}
	})

	// Test 3: Invalid test coverage target fails validation
	t.Run("InvalidTestCoverageTarget", func(t *testing.T) {
		cfg := &Config{
			Constitution: ConstitutionConfig{
				TestCoverageTarget: 150,
			},
		}

		err := ValidateConfig(cfg)
		if err == nil {
			t.Error("ValidateConfig() expected error for invalid test coverage target, got nil")
		}
	})

	// Test 4: Invalid LLM mode fails validation
	t.Run("InvalidLLMMode", func(t *testing.T) {
		cfg := &Config{
			LLM: LLMConfig{Mode: "invalid"},
		}

		err := ValidateConfig(cfg)
		if err == nil {
			t.Error("ValidateConfig() expected error for invalid LLM mode, got nil")
		}
	})
}

func TestGetConfig(t *testing.T) {
	// Note: These tests must run sequentially due to global state
	// Test 1: GetConfig returns config when loaded
	t.Run("ReturnsConfigAfterLoad", func(t *testing.T) {
		tempDir := t.TempDir()

		// Load config
		cfg, err := LoadConfig(tempDir)
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		// GetConfig should return the same config
		gotCfg := GetConfig()
		if gotCfg == nil {
			t.Error("GetConfig() returned nil after LoadConfig()")
		}
		if gotCfg != nil && gotCfg.User.Name != cfg.User.Name {
			t.Errorf("GetConfig() = %v, want %v", gotCfg, cfg)
		}
	})
}

func TestIsConfigLoaded(t *testing.T) {
	tempDir := t.TempDir()

	// Before loading, check initial state
	// Note: Due to parallel test execution, we can't guarantee initial state
	// So we just verify that it becomes true after loading

	_, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if !IsConfigLoaded() {
		t.Error("IsConfigLoaded() = false, want true after LoadConfig()")
	}
}

func TestNormalizeLanguageCode(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want string
	}{
		{"Korean", "ko", "ko_KR"},
		{"English", "en", "en_US"},
		{"Japanese", "ja", "ja_JP"},
		{"Chinese", "zh", "zh_CN"},
		{"Unknown", "fr", "fr_FR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLanguageCode(tt.lang)
			if got != tt.want {
				t.Errorf("normalizeLanguageCode(%s) = %s, want %s", tt.lang, got, tt.want)
			}
		})
	}
}

func TestConfigFileExists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".moai", "config", "sections")

	// Test 1: Non-existent file
	t.Run("NonExistent", func(t *testing.T) {
		if ConfigFileExists(tempDir, "user") {
			t.Error("ConfigFileExists() = true for non-existent file, want false")
		}
	})

	// Test 2: Existing file
	t.Run("Existing", func(t *testing.T) {
		if err := os.MkdirAll(configPath, 0755); err != nil {
			t.Fatalf("failed to create config directory: %v", err)
		}

		userYAML := `user:
  name: "Test"
`
		if err := os.WriteFile(filepath.Join(configPath, "user.yaml"), []byte(userYAML), 0644); err != nil {
			t.Fatalf("failed to write user.yaml: %v", err)
		}

		if !ConfigFileExists(tempDir, "user") {
			t.Error("ConfigFileExists() = false for existing file, want true")
		}
	})
}
