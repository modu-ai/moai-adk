package config

import (
	"os"
	"path/filepath"
	"strings"
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

// --- Additional tests for coverage improvement ---

func TestLoadConfigFromEnv(t *testing.T) {
	tempDir := t.TempDir()

	cfg, err := LoadConfigFromEnv(tempDir)
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfigFromEnv() returned nil config")
	}
	if cfg.Language.ConversationLanguage != "en" {
		t.Errorf("default conversation_language = %s, want en", cfg.Language.ConversationLanguage)
	}
}

func TestGetConfigPath(t *testing.T) {
	got := GetConfigPath("/project")
	want := filepath.Join("/project", ".moai", "config", "sections")
	if got != want {
		t.Errorf("GetConfigPath() = %s, want %s", got, want)
	}
}

func TestGetMoaiDir(t *testing.T) {
	got := GetMoaiDir("/project")
	want := filepath.Join("/project", ".moai")
	if got != want {
		t.Errorf("GetMoaiDir() = %s, want %s", got, want)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	// Write invalid YAML to user.yaml
	invalidYAML := `user:
  name: [invalid yaml
    broken: {{{
`
	if err := os.WriteFile(filepath.Join(configPath, "user.yaml"), []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write user.yaml: %v", err)
	}

	_, err := LoadConfig(tempDir)
	if err == nil {
		t.Error("LoadConfig() expected error for invalid YAML, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to parse user.yaml") {
		t.Errorf("LoadConfig() error = %v, want error containing 'failed to parse user.yaml'", err)
	}
}

func TestLoadConfig_UnreadableFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	// Create a user.yaml that is a directory (to cause read error)
	userYAMLPath := filepath.Join(configPath, "user.yaml")
	if err := os.MkdirAll(userYAMLPath, 0755); err != nil {
		t.Fatalf("failed to create directory as user.yaml: %v", err)
	}

	_, err := LoadConfig(tempDir)
	if err == nil {
		t.Error("LoadConfig() expected error for unreadable file, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to read user.yaml") {
		t.Errorf("LoadConfig() error = %v, want error containing 'failed to read user.yaml'", err)
	}
}

func TestNormalizeLanguageCode_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want string
	}{
		{"Spanish", "es", "es_ES"},
		{"German", "de", "de_DE"},
		{"UpperCase_ko", "KO", "ko_KR"},
		{"UpperCase_EN", "EN", "en_US"},
		{"Unknown2Char", "pt", "pt_PT"},
		{"Unknown2Char_it", "it", "it_IT"},
		{"LongerCode", "por", "por"},
		{"SingleChar", "x", "x"},
		{"EmptyString", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLanguageCode(tt.lang)
			if got != tt.want {
				t.Errorf("normalizeLanguageCode(%q) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}

func TestApplyEnvVarOverrides_AllVars(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	envVars := map[string]string{
		"MOAI_USER_NAME":                "EnvUser",
		"MOAI_CONVERSATION_LANG":        "ja",
		"MOAI_CODE_COMMENTS_LANG":       "ko",
		"MOAI_GIT_COMMIT_MESSAGES_LANG": "fr",
		"MOAI_DOCUMENTATION_LANG":       "de",
		"MOAI_ERROR_MESSAGES_LANG":      "es",
	}

	// Save and set all env vars
	saved := make(map[string]string)
	for k, v := range envVars {
		saved[k] = os.Getenv(k)
		_ = os.Setenv(k, v)
	}
	defer func() {
		for k, v := range saved {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}()

	cfg, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.User.Name != "EnvUser" {
		t.Errorf("user.name = %s, want EnvUser", cfg.User.Name)
	}
	if cfg.Language.ConversationLanguage != "ja" {
		t.Errorf("conversation_language = %s, want ja", cfg.Language.ConversationLanguage)
	}
	if cfg.Language.CodeComments != "ko" {
		t.Errorf("code_comments = %s, want ko", cfg.Language.CodeComments)
	}
	if cfg.Language.GitCommitMessages != "fr" {
		t.Errorf("git_commit_messages = %s, want fr", cfg.Language.GitCommitMessages)
	}
	if cfg.Language.Documentation != "de" {
		t.Errorf("documentation = %s, want de", cfg.Language.Documentation)
	}
	if cfg.Language.ErrorMessages != "es" {
		t.Errorf("error_messages = %s, want es", cfg.Language.ErrorMessages)
	}
}

// --- Validate() method tests on types ---

func TestUserConfig_Validate(t *testing.T) {
	cfg := &UserConfig{Name: "test"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("UserConfig.Validate() = %v, want nil", err)
	}

	cfg2 := &UserConfig{Name: ""}
	if err := cfg2.Validate(); err != nil {
		t.Errorf("UserConfig.Validate() with empty name = %v, want nil", err)
	}
}

func TestLanguageConfig_Validate(t *testing.T) {
	cfg := &LanguageConfig{ConversationLanguage: "en"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("LanguageConfig.Validate() = %v, want nil", err)
	}
}

func TestConstitutionConfig_Validate(t *testing.T) {
	t.Run("DDD_mode", func(t *testing.T) {
		cfg := &ConstitutionConfig{DevelopmentMode: "ddd"}
		if err := cfg.Validate(); err != nil {
			t.Errorf("ConstitutionConfig.Validate() = %v, want nil", err)
		}
	})

	t.Run("Other_mode", func(t *testing.T) {
		cfg := &ConstitutionConfig{DevelopmentMode: "other"}
		if err := cfg.Validate(); err != nil {
			t.Errorf("ConstitutionConfig.Validate() = %v, want nil", err)
		}
	})
}

func TestSystemConfig_Validate(t *testing.T) {
	cfg := &SystemConfig{}
	if err := cfg.Validate(); err != nil {
		t.Errorf("SystemConfig.Validate() = %v, want nil", err)
	}
}

func TestGitStrategyConfig_Validate(t *testing.T) {
	t.Run("ValidMode", func(t *testing.T) {
		cfg := &GitStrategyConfig{Mode: "manual"}
		if err := cfg.Validate(); err != nil {
			t.Errorf("GitStrategyConfig.Validate() = %v, want nil", err)
		}
	})

	t.Run("OtherMode", func(t *testing.T) {
		cfg := &GitStrategyConfig{Mode: "other"}
		if err := cfg.Validate(); err != nil {
			t.Errorf("GitStrategyConfig.Validate() = %v, want nil", err)
		}
	})
}

func TestProjectConfig_Validate(t *testing.T) {
	cfg := &ProjectConfig{Name: "test"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("ProjectConfig.Validate() = %v, want nil", err)
	}
}

func TestLLMConfig_Validate(t *testing.T) {
	cfg := &LLMConfig{Mode: "claude-only"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("LLMConfig.Validate() = %v, want nil", err)
	}
}

func TestServiceConfig_Validate(t *testing.T) {
	cfg := &ServiceConfig{Type: "claude_subscription"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("ServiceConfig.Validate() = %v, want nil", err)
	}
}

func TestRalphConfig_Validate(t *testing.T) {
	cfg := &RalphConfig{}
	if err := cfg.Validate(); err != nil {
		t.Errorf("RalphConfig.Validate() = %v, want nil", err)
	}
}

func TestWorkflowConfig_Validate(t *testing.T) {
	cfg := &WorkflowConfig{}
	if err := cfg.Validate(); err != nil {
		t.Errorf("WorkflowConfig.Validate() = %v, want nil", err)
	}
}

// --- ValidationErrors.Error() test ---

func TestValidationErrors_Error(t *testing.T) {
	ve := &ValidationErrors{
		Errors: []error{
			&ValidationError{Field: "field1", Message: "error1"},
			&ValidationError{Field: "field2", Message: "error2"},
		},
	}

	got := ve.Error()
	if !strings.Contains(got, "field1") || !strings.Contains(got, "field2") {
		t.Errorf("ValidationErrors.Error() = %q, want both field1 and field2 mentioned", got)
	}
	if !strings.Contains(got, "; ") {
		t.Errorf("ValidationErrors.Error() = %q, want errors separated by '; '", got)
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{Field: "test_field", Message: "test message"}
	got := ve.Error()
	want := "validation error in test_field: test message"
	if got != want {
		t.Errorf("ValidationError.Error() = %q, want %q", got, want)
	}
}

// --- Comprehensive validator tests ---

func TestValidateConfig_MultipleErrors(t *testing.T) {
	// Config with multiple validation errors to trigger multiple error paths
	cfg := &Config{
		Language: LanguageConfig{ConversationLanguage: "invalid"},
		Constitution: ConstitutionConfig{
			TestCoverageTarget: 150,
		},
		Ralph: RalphConfig{
			LSP: RalphLSPConfig{
				TimeoutSeconds: 15,
				PollIntervalMs: 1000,
			},
			Loop: RalphLoopConfig{
				MaxIterations: 10,
				Completion:    RalphCompletionConfig{CoverageThreshold: 85},
			},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{
				MaxIterations:       100,
				NoProgressThreshold: 5,
			},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("ValidateConfig() expected error for multiple invalid fields, got nil")
	}

	ve, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("ValidateConfig() error type = %T, want *ValidationErrors", err)
	}
	if len(ve.Errors) < 2 {
		t.Errorf("ValidateConfig() returned %d errors, want at least 2", len(ve.Errors))
	}
}

func TestValidateConstitutionConfig_InvalidDevMode(t *testing.T) {
	cfg := &Config{
		Constitution: ConstitutionConfig{
			DevelopmentMode:    "invalid_mode",
			TestCoverageTarget: 85,
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid development_mode, got nil")
	}
}

func TestValidateConstitutionConfig_InvalidTransformSize(t *testing.T) {
	cfg := &Config{
		Constitution: ConstitutionConfig{
			DevelopmentMode:    "ddd",
			TestCoverageTarget: 85,
			DDDDSettings: DDDSettings{
				MaxTransformationSize: "huge",
			},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid max_transformation_size, got nil")
	}
}

func TestValidateConstitutionConfig_NegativeCoverage(t *testing.T) {
	cfg := &Config{
		Constitution: ConstitutionConfig{
			TestCoverageTarget: -5,
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for negative test_coverage_target, got nil")
	}
}

func TestValidateSystemConfig_InvalidUpdateFrequency(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			MoAI: MoAIConfig{UpdateCheckFrequency: "hourly"},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid update_check_frequency, got nil")
	}
}

func TestValidateSystemConfig_InvalidSpecGitWorkflow(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			GitHub: GitHubConfig{SpecGitWorkflow: "invalid_workflow"},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid spec_git_workflow, got nil")
	}
}

func TestValidateGitStrategyConfig_InvalidMode(t *testing.T) {
	cfg := &Config{
		GitStrategy: GitStrategyConfig{Mode: "invalid_mode"},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid git_strategy mode, got nil")
	}
}

func TestValidateGitStrategyConfig_InvalidPersonalWorkflow(t *testing.T) {
	cfg := &Config{
		GitStrategy: GitStrategyConfig{
			Mode:     "personal",
			Personal: GitModeConfig{Workflow: "invalid-flow"},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid personal workflow, got nil")
	}
}

func TestValidateGitStrategyConfig_InvalidTeamWorkflow(t *testing.T) {
	cfg := &Config{
		GitStrategy: GitStrategyConfig{
			Mode: "team",
			Team: GitTeamConfig{Workflow: "invalid-flow"},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid team workflow, got nil")
	}
}

func TestValidateGitStrategyConfig_InvalidCommitStyle(t *testing.T) {
	cfg := &Config{
		GitStrategy: GitStrategyConfig{
			Mode: "manual",
			Manual: GitModeConfig{
				Workflow:    "github-flow",
				CommitStyle: CommitStyleConfig{Format: "invalid_style"},
			},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid commit style, got nil")
	}
}

func TestValidateLLMConfig_InvalidGLMModel(t *testing.T) {
	cfg := &Config{
		LLM: LLMConfig{
			Mode: "claude-only",
			GLM: GLMConfig{
				Models: map[string]string{
					"invalid_model": "some-value",
				},
			},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid GLM model key, got nil")
	}
}

func TestValidateServiceConfig_InvalidType(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{Type: "invalid_type"},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid service type, got nil")
	}
}

func TestValidateServiceConfig_InvalidPricingPlan(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Type:        "claude_subscription",
			PricingPlan: "invalid_plan",
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid pricing plan, got nil")
	}
}

func TestValidateServiceConfig_InvalidAllocationStrategy(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Type:        "claude_subscription",
			PricingPlan: "pro",
			ModelAllocation: ModelAllocationConfig{
				Strategy: "invalid_strategy",
			},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid allocation strategy, got nil")
	}
}

func TestValidateRalphConfig_InvalidTimeout(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 0, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid timeout_seconds (0), got nil")
	}
}

func TestValidateRalphConfig_TimeoutTooHigh(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 500, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for timeout_seconds (500), got nil")
	}
}

func TestValidateRalphConfig_InvalidPollInterval(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 50},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid poll_interval_ms (50), got nil")
	}
}

func TestValidateRalphConfig_PollIntervalTooHigh(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 70000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for poll_interval_ms (70000), got nil")
	}
}

func TestValidateRalphConfig_InvalidMaxIterations(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 0, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid loop.max_iterations (0), got nil")
	}
}

func TestValidateRalphConfig_MaxIterationsTooHigh(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 2000, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for loop.max_iterations (2000), got nil")
	}
}

func TestValidateRalphConfig_InvalidCoverageThreshold(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: -1}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid coverage_threshold (-1), got nil")
	}
}

func TestValidateRalphConfig_CoverageThresholdTooHigh(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 150}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for coverage_threshold (150), got nil")
	}
}

func TestValidateRalphConfig_InvalidSeverityThreshold(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
			Hooks: RalphHooksConfig{
				PostToolLSP: RalphPostToolLSPConfig{
					SeverityThreshold: "critical",
				},
			},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid severity_threshold, got nil")
	}
}

func TestValidateWorkflowConfig_InvalidMaxIterations(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{
				MaxIterations:       0,
				NoProgressThreshold: 5,
			},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid loop_prevention.max_iterations (0), got nil")
	}
}

func TestValidateWorkflowConfig_MaxIterationsTooHigh(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{
				MaxIterations:       2000,
				NoProgressThreshold: 5,
			},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for loop_prevention.max_iterations (2000), got nil")
	}
}

func TestValidateWorkflowConfig_InvalidNoProgressThreshold(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{
				MaxIterations:       100,
				NoProgressThreshold: 0,
			},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid no_progress_threshold (0), got nil")
	}
}

func TestValidateWorkflowConfig_NoProgressThresholdTooHigh(t *testing.T) {
	cfg := &Config{
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{
				MaxIterations:       100,
				NoProgressThreshold: 200,
			},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for no_progress_threshold (200), got nil")
	}
}

func TestValidateConfig_AllSectionsValid(t *testing.T) {
	cfg := getDefaultConfig()
	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig() on default config failed: %v", err)
	}
}

func TestLoadConfig_AllSectionFiles(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	// Write multiple section files to exercise more of the loading loop
	sections := map[string]string{
		"user.yaml": `user:
  name: "AllSections"
`,
		"language.yaml": `language:
  conversation_language: ja
  conversation_language_name: Japanese
`,
		"quality.yaml": `constitution:
  development_mode: ddd
  enforce_quality: true
  test_coverage_target: 90
`,
		"system.yaml": `system:
  moai:
    version: "1.0.0"
    update_check_frequency: weekly
`,
		"git-strategy.yaml": `git_strategy:
  mode: team
`,
		"project.yaml": `project:
  name: "test-project"
  type: "web"
`,
		"llm.yaml": `llm:
  mode: mashup
`,
		"pricing.yaml": `service:
  type: claude_api
  pricing_plan: max5
`,
		"ralph.yaml": `ralph:
  enabled: false
`,
		"workflow.yaml": `workflow:
  loop_prevention:
    max_iterations: 50
`,
	}

	for name, content := range sections {
		if err := os.WriteFile(filepath.Join(configPath, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	cfg, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig() with all sections failed: %v", err)
	}

	if cfg.User.Name != "AllSections" {
		t.Errorf("user.name = %s, want AllSections", cfg.User.Name)
	}
	if cfg.Language.ConversationLanguage != "ja" {
		t.Errorf("conversation_language = %s, want ja", cfg.Language.ConversationLanguage)
	}
	if cfg.Constitution.TestCoverageTarget != 90 {
		t.Errorf("test_coverage_target = %d, want 90", cfg.Constitution.TestCoverageTarget)
	}
	if cfg.System.MoAI.Version != "1.0.0" {
		t.Errorf("moai.version = %s, want 1.0.0", cfg.System.MoAI.Version)
	}
	if cfg.System.MoAI.UpdateCheckFrequency != "weekly" {
		t.Errorf("update_check_frequency = %s, want weekly", cfg.System.MoAI.UpdateCheckFrequency)
	}
	if cfg.GitStrategy.Mode != "team" {
		t.Errorf("git_strategy.mode = %s, want team", cfg.GitStrategy.Mode)
	}
	if cfg.LLM.Mode != "mashup" {
		t.Errorf("llm.mode = %s, want mashup", cfg.LLM.Mode)
	}
	if cfg.Service.Type != "claude_api" {
		t.Errorf("service.type = %s, want claude_api", cfg.Service.Type)
	}
	if cfg.Ralph.Enabled != false {
		t.Errorf("ralph.enabled = %v, want false", cfg.Ralph.Enabled)
	}
	if cfg.Workflow.LoopPrevention.MaxIterations != 50 {
		t.Errorf("loop_prevention.max_iterations = %d, want 50", cfg.Workflow.LoopPrevention.MaxIterations)
	}
}

func TestValidateRalphConfig_ValidSeverities(t *testing.T) {
	validSeverities := []string{"error", "warning", "info"}
	for _, sev := range validSeverities {
		t.Run(sev, func(t *testing.T) {
			cfg := &Config{
				Ralph: RalphConfig{
					LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
					Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
					Hooks: RalphHooksConfig{
						PostToolLSP: RalphPostToolLSPConfig{
							SeverityThreshold: sev,
						},
					},
				},
				Workflow: WorkflowConfig{
					LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
				},
			}

			err := ValidateConfig(cfg)
			if err != nil {
				t.Errorf("ValidateConfig() with severity %q failed: %v", sev, err)
			}
		})
	}
}

func TestValidateGitStrategyConfig_InvalidManualWorkflow(t *testing.T) {
	cfg := &Config{
		GitStrategy: GitStrategyConfig{
			Mode:   "manual",
			Manual: GitModeConfig{Workflow: "trunk-based"},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid manual workflow, got nil")
	}
}

func TestValidateLLMConfig_NilModels(t *testing.T) {
	cfg := &Config{
		LLM: LLMConfig{
			Mode: "claude-only",
			GLM:  GLMConfig{Models: nil},
		},
		Ralph: RalphConfig{
			LSP:  RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop: RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig() with nil GLM models should pass: %v", err)
	}
}

func TestValidateConfig_EmptyStringsPassValidation(t *testing.T) {
	// All empty string fields should pass validation (optional fields)
	cfg := &Config{
		Language:     LanguageConfig{ConversationLanguage: ""},
		Constitution: ConstitutionConfig{DevelopmentMode: "", TestCoverageTarget: 85},
		System: SystemConfig{
			MoAI:   MoAIConfig{UpdateCheckFrequency: ""},
			GitHub: GitHubConfig{SpecGitWorkflow: ""},
		},
		GitStrategy: GitStrategyConfig{
			Mode:     "",
			Manual:   GitModeConfig{Workflow: "", CommitStyle: CommitStyleConfig{Format: ""}},
			Personal: GitModeConfig{Workflow: ""},
			Team:     GitTeamConfig{Workflow: ""},
		},
		LLM:     LLMConfig{Mode: ""},
		Service: ServiceConfig{Type: "", PricingPlan: "", ModelAllocation: ModelAllocationConfig{Strategy: ""}},
		Ralph: RalphConfig{
			LSP:   RalphLSPConfig{TimeoutSeconds: 15, PollIntervalMs: 1000},
			Loop:  RalphLoopConfig{MaxIterations: 10, Completion: RalphCompletionConfig{CoverageThreshold: 85}},
			Hooks: RalphHooksConfig{PostToolLSP: RalphPostToolLSPConfig{SeverityThreshold: ""}},
		},
		Workflow: WorkflowConfig{
			LoopPrevention: LoopPreventionConfig{MaxIterations: 100, NoProgressThreshold: 5},
		},
	}

	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig() with empty optional fields should pass: %v", err)
	}
}
