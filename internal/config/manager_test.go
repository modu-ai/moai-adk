package config

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// setupManagerTestDir creates a project root with .moai/config/sections and
// copies testdata files into it. Returns the project root path.
func setupManagerTestDir(t *testing.T, files []string) string {
	t.Helper()
	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	for _, f := range files {
		src := filepath.Join("testdata", "valid", f)
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("failed to read testdata file %s: %v", f, err)
		}
		dst := filepath.Join(sectionsDir, f)
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatalf("failed to write test file %s: %v", dst, err)
		}
	}
	return tempDir
}

func TestNewConfigManager(t *testing.T) {
	t.Parallel()

	m := NewConfigManager()
	if m == nil {
		t.Fatal("NewConfigManager() returned nil")
	}
	if m.loader == nil {
		t.Error("NewConfigManager() should initialize loader")
	}
	if m.state != stateUninitialized {
		t.Errorf("expected state %d (uninitialized), got %d", stateUninitialized, m.state)
	}
}

func TestConfigManagerLoadValid(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()

	cfg, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.User.Name != "TestUser" {
		t.Errorf("User.Name: got %q, want %q", cfg.User.Name, "TestUser")
	}
	if cfg.Language.ConversationLanguage != "ko" {
		t.Errorf("Language.ConversationLanguage: got %q, want %q",
			cfg.Language.ConversationLanguage, "ko")
	}
	if cfg.Quality.DevelopmentMode != "ddd" {
		t.Errorf("Quality.DevelopmentMode: got %q, want %q",
			cfg.Quality.DevelopmentMode, "ddd")
	}
}

func TestConfigManagerLoadDefaults(t *testing.T) {
	t.Parallel()

	// Empty project root with no .moai directory
	root := t.TempDir()
	m := NewConfigManager()

	cfg, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Should get all defaults
	if cfg.Language.ConversationLanguage != DefaultConversationLanguage {
		t.Errorf("Language.ConversationLanguage: got %q, want default %q",
			cfg.Language.ConversationLanguage, DefaultConversationLanguage)
	}
	if cfg.Quality.DevelopmentMode != models.ModeTDD {
		t.Errorf("Quality.DevelopmentMode: got %q, want default %q",
			cfg.Quality.DevelopmentMode, models.ModeTDD)
	}
}

func TestConfigManagerLoadValidationError(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml"})

	// Overwrite user.yaml with empty name
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	emptyUser := []byte("user:\n  name: \"\"\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), emptyUser, 0o644); err != nil {
		t.Fatalf("failed to write user.yaml: %v", err)
	}

	m := NewConfigManager()
	_, err := m.Load(root)
	if err == nil {
		t.Fatal("Load() expected validation error for empty user.name")
	}
}

func TestConfigManagerGet(t *testing.T) {
	t.Parallel()

	t.Run("before load returns nil", func(t *testing.T) {
		t.Parallel()
		m := NewConfigManager()
		if got := m.Get(); got != nil {
			t.Errorf("Get() before Load() should return nil, got: %v", got)
		}
	})

	t.Run("after load returns config", func(t *testing.T) {
		t.Parallel()
		root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
		m := NewConfigManager()
		_, err := m.Load(root)
		if err != nil {
			t.Fatalf("Load() error: %v", err)
		}

		cfg := m.Get()
		if cfg == nil {
			t.Fatal("Get() returned nil after Load()")
		}
		if cfg.User.Name != "TestUser" {
			t.Errorf("User.Name: got %q, want %q", cfg.User.Name, "TestUser")
		}
	})
}

func TestConfigManagerGetSection(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name    string
		section string
		check   func(any) bool
	}{
		{"user", "user", func(v any) bool {
			u, ok := v.(models.UserConfig)
			return ok && u.Name == "TestUser"
		}},
		{"language", "language", func(v any) bool {
			l, ok := v.(models.LanguageConfig)
			return ok && l.ConversationLanguage == "ko"
		}},
		{"quality", "quality", func(v any) bool {
			q, ok := v.(models.QualityConfig)
			return ok && q.DevelopmentMode == "ddd"
		}},
		{"project", "project", func(v any) bool {
			_, ok := v.(models.ProjectConfig)
			return ok
		}},
		{"git_strategy", "git_strategy", func(v any) bool {
			_, ok := v.(GitStrategyConfig)
			return ok
		}},
		{"git_convention", "git_convention", func(v any) bool {
			_, ok := v.(models.GitConventionConfig)
			return ok
		}},
		{"system", "system", func(v any) bool {
			_, ok := v.(SystemConfig)
			return ok
		}},
		{"llm", "llm", func(v any) bool {
			_, ok := v.(LLMConfig)
			return ok
		}},
		{"pricing", "pricing", func(v any) bool {
			_, ok := v.(PricingConfig)
			return ok
		}},
		{"ralph", "ralph", func(v any) bool {
			_, ok := v.(RalphConfig)
			return ok
		}},
		{"workflow", "workflow", func(v any) bool {
			_, ok := v.(WorkflowConfig)
			return ok
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v, err := m.GetSection(tt.section)
			if err != nil {
				t.Fatalf("GetSection(%q) error: %v", tt.section, err)
			}
			if !tt.check(v) {
				t.Errorf("GetSection(%q) returned unexpected value: %v", tt.section, v)
			}
		})
	}
}

func TestConfigManagerGetSectionNotInitialized(t *testing.T) {
	t.Parallel()

	m := NewConfigManager()
	_, err := m.GetSection("user")
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got: %v", err)
	}
}

func TestConfigManagerGetSectionNotFound(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	_, err = m.GetSection("nonexistent")
	if !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound, got: %v", err)
	}
}

func TestConfigManagerSetSection(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Set user section
	newUser := models.UserConfig{Name: "NewName"}
	if err := m.SetSection("user", newUser); err != nil {
		t.Fatalf("SetSection(user) error: %v", err)
	}

	// Verify it was updated
	v, err := m.GetSection("user")
	if err != nil {
		t.Fatalf("GetSection(user) error: %v", err)
	}
	u := v.(models.UserConfig)
	if u.Name != "NewName" {
		t.Errorf("User.Name: got %q, want %q", u.Name, "NewName")
	}
}

func TestConfigManagerSetSectionAllTypes(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name  string
		value any
	}{
		{"user", models.UserConfig{Name: "X"}},
		{"language", models.LanguageConfig{ConversationLanguage: "en"}},
		{"quality", models.QualityConfig{DevelopmentMode: models.ModeTDD}},
		{"project", models.ProjectConfig{}},
		{"git_strategy", GitStrategyConfig{BranchPrefix: "test/"}},
		{"git_convention", models.GitConventionConfig{Convention: "angular"}},
		{"system", SystemConfig{LogLevel: "debug"}},
		{"llm", LLMConfig{DefaultModel: "opus"}},
		{"pricing", PricingConfig{TokenBudget: 100}},
		{"ralph", RalphConfig{MaxIterations: 1}},
		{"workflow", WorkflowConfig{PlanTokens: 1000}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.SetSection(tt.name, tt.value); err != nil {
				t.Errorf("SetSection(%q) error: %v", tt.name, err)
			}
		})
	}
}

func TestConfigManagerSetSectionNotInitialized(t *testing.T) {
	t.Parallel()

	m := NewConfigManager()
	err := m.SetSection("user", models.UserConfig{Name: "X"})
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got: %v", err)
	}
}

func TestConfigManagerSetSectionNotFound(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	err = m.SetSection("nonexistent", "value")
	if !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound, got: %v", err)
	}
}

func TestConfigManagerSetSectionTypeMismatch(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name  string
		value any
	}{
		{"user", "wrong type"},
		{"language", 42},
		{"quality", true},
		{"project", "string"},
		{"git_strategy", 123},
		{"git_convention", "wrong"},
		{"system", false},
		{"llm", []string{"wrong"}},
		{"pricing", "wrong"},
		{"ralph", 99},
		{"workflow", map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.SetSection(tt.name, tt.value)
			if !errors.Is(err, ErrSectionTypeMismatch) {
				t.Errorf("SetSection(%q, %T) expected ErrSectionTypeMismatch, got: %v",
					tt.name, tt.value, err)
			}
		})
	}
}

func TestConfigManagerSave(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Modify config in memory
	if err := m.SetSection("user", models.UserConfig{Name: "SavedUser"}); err != nil {
		t.Fatalf("SetSection() error: %v", err)
	}

	// Save to disk
	if err := m.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload and verify
	m2 := NewConfigManager()
	cfg, err := m2.Load(root)
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if cfg.User.Name != "SavedUser" {
		t.Errorf("User.Name after Save/Load: got %q, want %q", cfg.User.Name, "SavedUser")
	}
}

func TestConfigManagerSaveNotInitialized(t *testing.T) {
	t.Parallel()

	m := NewConfigManager()
	err := m.Save()
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got: %v", err)
	}
}

func TestConfigManagerReload(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Modify file on disk
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	newContent := []byte("user:\n  name: ReloadedUser\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), newContent, 0o644); err != nil {
		t.Fatalf("failed to write updated user.yaml: %v", err)
	}

	// Reload
	if err := m.Reload(); err != nil {
		t.Fatalf("Reload() error: %v", err)
	}

	cfg := m.Get()
	if cfg.User.Name != "ReloadedUser" {
		t.Errorf("User.Name after Reload: got %q, want %q", cfg.User.Name, "ReloadedUser")
	}
}

func TestConfigManagerReloadNotInitialized(t *testing.T) {
	t.Parallel()

	m := NewConfigManager()
	err := m.Reload()
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got: %v", err)
	}
}

func TestConfigManagerWatch(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	var callbackCfg Config
	called := false
	err = m.Watch(func(cfg Config) {
		called = true
		callbackCfg = cfg
	})
	if err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// Modify file and reload to trigger callback
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	newContent := []byte("user:\n  name: WatchedUser\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), newContent, 0o644); err != nil {
		t.Fatalf("failed to write updated user.yaml: %v", err)
	}

	if err := m.Reload(); err != nil {
		t.Fatalf("Reload() error: %v", err)
	}

	if !called {
		t.Error("Watch callback was not called after Reload")
	}
	if callbackCfg.User.Name != "WatchedUser" {
		t.Errorf("callback config User.Name: got %q, want %q",
			callbackCfg.User.Name, "WatchedUser")
	}
}

func TestConfigManagerWatchNotInitialized(t *testing.T) {
	t.Parallel()

	m := NewConfigManager()
	err := m.Watch(func(cfg Config) {})
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got: %v", err)
	}
}

func TestConfigManagerEnvOverrideDevelopmentMode(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})

	t.Setenv("MOAI_DEVELOPMENT_MODE", "tdd")

	m := NewConfigManager()
	cfg, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// quality.yaml has "ddd" but env var should override to "tdd"
	if cfg.Quality.DevelopmentMode != models.ModeTDD {
		t.Errorf("DevelopmentMode: got %q, want %q (env override)",
			cfg.Quality.DevelopmentMode, models.ModeTDD)
	}
}

func TestConfigManagerEnvOverrideLogLevel(t *testing.T) {
	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})

	t.Setenv("MOAI_LOG_LEVEL", "debug")

	m := NewConfigManager()
	cfg, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.System.LogLevel != "debug" {
		t.Errorf("System.LogLevel: got %q, want %q (env override)",
			cfg.System.LogLevel, "debug")
	}
}

func TestConfigManagerEnvOverrideLogFormat(t *testing.T) {
	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})

	t.Setenv("MOAI_LOG_FORMAT", "json")

	m := NewConfigManager()
	cfg, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.System.LogFormat != "json" {
		t.Errorf("System.LogFormat: got %q, want %q (env override)",
			cfg.System.LogFormat, "json")
	}
}

func TestConfigManagerEnvOverrideNoColor(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"true sets NoColor", "true", true},
		{"1 sets NoColor", "1", true},
		{"false does not set NoColor", "false", false},
		{"0 does not set NoColor", "0", false},
		{"empty does not set NoColor", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
			if tt.envValue != "" {
				t.Setenv("MOAI_NO_COLOR", tt.envValue)
			}

			m := NewConfigManager()
			cfg, err := m.Load(root)
			if err != nil {
				t.Fatalf("Load() error: %v", err)
			}

			if cfg.System.NoColor != tt.want {
				t.Errorf("System.NoColor: got %v, want %v", cfg.System.NoColor, tt.want)
			}
		})
	}
}

func TestConfigManagerEnvOverrideConfigDir(t *testing.T) {
	// Create a custom config directory
	customDir := t.TempDir()
	sectionsDir := filepath.Join(customDir, "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create custom config dir: %v", err)
	}

	// Write user.yaml with custom name
	content := []byte("user:\n  name: CustomDirUser\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), content, 0o644); err != nil {
		t.Fatalf("failed to write user.yaml: %v", err)
	}

	t.Setenv("MOAI_CONFIG_DIR", customDir)

	m := NewConfigManager()
	// Use a different project root that does NOT have .moai dir
	projectRoot := t.TempDir()
	cfg, err := m.Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.User.Name != "CustomDirUser" {
		t.Errorf("User.Name: got %q, want %q (from MOAI_CONFIG_DIR)",
			cfg.User.Name, "CustomDirUser")
	}
}

func TestConfigManagerEnvOverrideInvalidMode(t *testing.T) {
	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})

	t.Setenv("MOAI_DEVELOPMENT_MODE", "waterfall")

	m := NewConfigManager()
	_, err := m.Load(root)
	if err == nil {
		t.Fatal("Load() expected validation error for invalid mode from env var")
	}
	if !errors.Is(err, ErrInvalidDevelopmentMode) {
		t.Errorf("expected ErrInvalidDevelopmentMode, got: %v", err)
	}
}

func TestConfigManagerConcurrentReads(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent Get() calls
	for range goroutines {
		wg.Go(func() {
			cfg := m.Get()
			if cfg == nil {
				t.Error("Get() returned nil during concurrent access")
				return
			}
			if cfg.User.Name != "TestUser" {
				t.Errorf("concurrent Get() User.Name: got %q, want %q",
					cfg.User.Name, "TestUser")
			}
		})
	}

	// Concurrent GetSection() calls
	for range goroutines {
		wg.Go(func() {
			v, err := m.GetSection("user")
			if err != nil {
				t.Errorf("concurrent GetSection() error: %v", err)
				return
			}
			u := v.(models.UserConfig)
			if u.Name != "TestUser" {
				t.Errorf("concurrent GetSection() User.Name: got %q, want %q",
					u.Name, "TestUser")
			}
		})
	}

	wg.Wait()
}

func TestConfigManagerConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	var wg sync.WaitGroup
	const goroutines = 30

	// Concurrent readers
	for range goroutines {
		wg.Go(func() {
			_ = m.Get()
			_, _ = m.GetSection("language")
		})
	}

	// Concurrent writers
	for i := range goroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = m.SetSection("system", SystemConfig{
				LogLevel: "info",
			})
		}(i)
	}

	wg.Wait()

	// Verify config is still accessible after concurrent operations
	cfg := m.Get()
	if cfg == nil {
		t.Fatal("Get() returned nil after concurrent read/write")
	}
}

func TestConfigManagerSaveCreatesDirectory(t *testing.T) {
	t.Parallel()

	// Create a minimal project root without .moai directory
	root := t.TempDir()
	m := NewConfigManager()

	// Load with defaults (no files)
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Save should create the directory structure
	if err := m.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify directory was created
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if _, err := os.Stat(sectionsDir); os.IsNotExist(err) {
		t.Errorf("Save() did not create sections directory at %s", sectionsDir)
	}

	// Verify files were created
	for _, f := range []string{"user.yaml", "language.yaml", "quality.yaml"} {
		path := filepath.Join(sectionsDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Save() did not create file: %s", f)
		}
	}
}

func TestConfigManagerSaveAndReloadRoundTrip(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	_, err := m.Load(root)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Modify all three saveable sections
	_ = m.SetSection("user", models.UserConfig{Name: "RoundTripUser"})
	_ = m.SetSection("language", models.LanguageConfig{
		ConversationLanguage:     "ja",
		ConversationLanguageName: "Japanese",
		AgentPromptLanguage:      "en",
		GitCommitMessages:        "en",
		CodeComments:             "en",
		Documentation:            "en",
		ErrorMessages:            "en",
	})
	_ = m.SetSection("quality", models.QualityConfig{
		DevelopmentMode:    models.ModeTDD,
		EnforceQuality:     true,
		TestCoverageTarget: 90,
		DDDSettings:        NewDefaultDDDSettings(),
		TDDSettings:        NewDefaultTDDSettings(),
		CoverageExemptions: NewDefaultCoverageExemptions(),
	})

	// Save
	if err := m.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload into new manager
	m2 := NewConfigManager()
	cfg, err := m2.Load(root)
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}

	if cfg.User.Name != "RoundTripUser" {
		t.Errorf("User.Name round-trip: got %q, want %q", cfg.User.Name, "RoundTripUser")
	}
	if cfg.Language.ConversationLanguage != "ja" {
		t.Errorf("Language.ConversationLanguage round-trip: got %q, want %q",
			cfg.Language.ConversationLanguage, "ja")
	}
	if cfg.Quality.DevelopmentMode != models.ModeTDD {
		t.Errorf("Quality.DevelopmentMode round-trip: got %q, want %q",
			cfg.Quality.DevelopmentMode, models.ModeTDD)
	}
	if cfg.Quality.TestCoverageTarget != 90 {
		t.Errorf("Quality.TestCoverageTarget round-trip: got %d, want 90",
			cfg.Quality.TestCoverageTarget)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()

	// Test with no env vars set (parallel safe since we only read defaults)
	applyEnvOverrides(cfg)

	// Values should remain as defaults since no env vars are set in this test
	if cfg.Quality.DevelopmentMode != models.ModeTDD {
		t.Errorf("DevelopmentMode: got %q, want default %q",
			cfg.Quality.DevelopmentMode, models.ModeTDD)
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrConfigNotFound", ErrConfigNotFound, "config: configuration directory not found"},
		{"ErrInvalidConfig", ErrInvalidConfig, "config: invalid configuration"},
		{"ErrSectionNotFound", ErrSectionNotFound, "config: section not found"},
		{"ErrInvalidDevelopmentMode", ErrInvalidDevelopmentMode, "config: invalid development_mode, must be one of: ddd, tdd"},
		{"ErrNotInitialized", ErrNotInitialized, "config: manager not initialized, call Load() first"},
		{"ErrSectionTypeMismatch", ErrSectionTypeMismatch, "config: section type mismatch"},
		{"ErrDynamicToken", ErrDynamicToken, "config: unexpanded dynamic token detected"},
		{"ErrInvalidYAML", ErrInvalidYAML, "config: invalid YAML syntax"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err.Error() != tt.msg {
				t.Errorf("Error(): got %q, want %q", tt.err.Error(), tt.msg)
			}
		})
	}
}
