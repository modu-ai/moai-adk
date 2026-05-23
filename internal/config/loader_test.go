package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// setupTestdataDir creates a .moai/config/sections structure under tempDir
// and copies the given YAML files from testdata/valid/ into it.
func setupTestdataDir(t *testing.T, tempDir string, files []string) string {
	t.Helper()
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

func TestLoaderLoadAllSections(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	root := setupTestdataDir(t, tempDir, []string{"user.yaml", "language.yaml", "quality.yaml"})

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Verify user section was loaded
	if cfg.User.Name != "TestUser" {
		t.Errorf("User.Name: got %q, want %q", cfg.User.Name, "TestUser")
	}

	// Verify language section was loaded
	if cfg.Language.ConversationLanguage != "ko" {
		t.Errorf("Language.ConversationLanguage: got %q, want %q",
			cfg.Language.ConversationLanguage, "ko")
	}
	if cfg.Language.ConversationLanguageName != "Korean" {
		t.Errorf("Language.ConversationLanguageName: got %q, want %q",
			cfg.Language.ConversationLanguageName, "Korean")
	}

	// Verify quality section was loaded (uses "constitution:" key)
	if cfg.Quality.DevelopmentMode != "ddd" {
		t.Errorf("Quality.DevelopmentMode: got %q, want %q",
			cfg.Quality.DevelopmentMode, "ddd")
	}
	if cfg.Quality.TestCoverageTarget != 85 {
		t.Errorf("Quality.TestCoverageTarget: got %d, want 85",
			cfg.Quality.TestCoverageTarget)
	}
}

func TestLoaderLoadedSections(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	root := setupTestdataDir(t, tempDir, []string{"user.yaml", "language.yaml", "quality.yaml"})

	loader := NewLoader()
	_, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	sections := loader.LoadedSections()
	expectedSections := []string{"user", "language", "quality"}
	for _, name := range expectedSections {
		if !sections[name] {
			t.Errorf("expected section %q to be loaded", name)
		}
	}
}

func TestLoaderLoadedSectionsReturnsCopy(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	root := setupTestdataDir(t, tempDir, []string{"user.yaml"})

	loader := NewLoader()
	_, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	s1 := loader.LoadedSections()
	s2 := loader.LoadedSections()

	s1["user"] = false
	if !s2["user"] {
		t.Error("LoadedSections() returned shared map, expected a copy")
	}
}

func TestLoaderMissingSectionsDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	// Create .moai but not the config/sections subdirectory
	moaiDir := filepath.Join(tempDir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("failed to create moai dir: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Should return default config
	if cfg.Language.ConversationLanguage != DefaultConversationLanguage {
		t.Errorf("expected default ConversationLanguage %q, got %q",
			DefaultConversationLanguage, cfg.Language.ConversationLanguage)
	}
}

func TestLoaderMissingIndividualFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	// Create only user.yaml, missing language.yaml and quality.yaml
	root := setupTestdataDir(t, tempDir, []string{"user.yaml"})

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// User should be loaded
	if cfg.User.Name != "TestUser" {
		t.Errorf("User.Name: got %q, want %q", cfg.User.Name, "TestUser")
	}

	// Language and quality should use defaults
	if cfg.Language.ConversationLanguage != DefaultConversationLanguage {
		t.Errorf("Language.ConversationLanguage: got %q, want default %q",
			cfg.Language.ConversationLanguage, DefaultConversationLanguage)
	}
	if cfg.Quality.TestCoverageTarget != DefaultTestCoverageTarget {
		t.Errorf("Quality.TestCoverageTarget: got %d, want default %d",
			cfg.Quality.TestCoverageTarget, DefaultTestCoverageTarget)
	}

	// Only user should be in loaded sections
	sections := loader.LoadedSections()
	if !sections["user"] {
		t.Error("expected user section to be loaded")
	}
	if sections["language"] {
		t.Error("expected language section to NOT be loaded")
	}
	if sections["quality"] {
		t.Error("expected quality section to NOT be loaded")
	}
}

func TestLoaderInvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// Write invalid YAML
	invalidYAML := []byte("user:\n  name: [invalid yaml here\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), invalidYAML, 0o644); err != nil {
		t.Fatalf("failed to write invalid yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() should not return error for invalid YAML (skips with warning), got: %v", err)
	}

	// User should use defaults since invalid YAML was skipped
	if cfg.User.Name != "" {
		t.Errorf("User.Name should be default (empty), got %q", cfg.User.Name)
	}

	// user section should NOT be marked as loaded
	sections := loader.LoadedSections()
	if sections["user"] {
		t.Error("expected user section to NOT be loaded after invalid YAML")
	}
}

func TestLoaderEmptyYAMLFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// Write empty YAML file
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write empty yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Empty YAML is valid but unmarshals to zero values
	// The file was read and parsed so loadYAMLFile returns true
	if cfg.User.Name != "" {
		t.Errorf("User.Name: got %q, want empty", cfg.User.Name)
	}
}

func TestLoaderQualityConstitutionKey(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	root := setupTestdataDir(t, tempDir, []string{"quality.yaml"})

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify the quality.yaml "constitution:" key is properly parsed
	if cfg.Quality.DevelopmentMode != "ddd" {
		t.Errorf("Quality.DevelopmentMode: got %q, want %q",
			cfg.Quality.DevelopmentMode, "ddd")
	}
	if !cfg.Quality.EnforceQuality {
		t.Error("Quality.EnforceQuality: expected true")
	}
	if !cfg.Quality.DDDSettings.RequireExistingTests {
		t.Error("Quality.DDDSettings.RequireExistingTests: expected true")
	}
	if cfg.Quality.DDDSettings.MaxTransformationSize != "small" {
		t.Errorf("Quality.DDDSettings.MaxTransformationSize: got %q, want %q",
			cfg.Quality.DDDSettings.MaxTransformationSize, "small")
	}
}

func TestLoadYAMLFileNonExistent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	var target userFileWrapper
	loaded, err := loadYAMLFile(tempDir, "nonexistent.yaml", &target)
	if err != nil {
		t.Fatalf("loadYAMLFile() error for missing file: %v", err)
	}
	if loaded {
		t.Error("loadYAMLFile() should return false for missing file")
	}
}

func TestLoadYAMLFileValidContent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	content := []byte("user:\n  name: Alice\n")
	if err := os.WriteFile(filepath.Join(tempDir, "user.yaml"), content, 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var target userFileWrapper
	loaded, err := loadYAMLFile(tempDir, "user.yaml", &target)
	if err != nil {
		t.Fatalf("loadYAMLFile() error: %v", err)
	}
	if !loaded {
		t.Error("loadYAMLFile() should return true for valid file")
	}
	if target.User.Name != "Alice" {
		t.Errorf("User.Name: got %q, want %q", target.User.Name, "Alice")
	}
}

func TestLoadYAMLFileInvalidContent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	content := []byte("user:\n  name: [broken yaml\n")
	if err := os.WriteFile(filepath.Join(tempDir, "test.yaml"), content, 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var target userFileWrapper
	loaded, err := loadYAMLFile(tempDir, "test.yaml", &target)
	if err == nil {
		t.Fatal("loadYAMLFile() expected error for invalid YAML, got nil")
	}
	if loaded {
		t.Error("loadYAMLFile() should return false for invalid YAML")
	}
}

func TestNewLoaderReturnsNonNil(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}
}

func TestLoaderMultipleLoads(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	root := setupTestdataDir(t, tempDir, []string{"user.yaml"})

	loader := NewLoader()

	// First load
	cfg1, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("first Load() error: %v", err)
	}
	if cfg1.User.Name != "TestUser" {
		t.Errorf("first load User.Name: got %q, want %q", cfg1.User.Name, "TestUser")
	}

	// Overwrite with different content
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	newContent := []byte("user:\n  name: NewUser\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), newContent, 0o644); err != nil {
		t.Fatalf("failed to write updated file: %v", err)
	}

	// Second load should pick up new content
	cfg2, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("second Load() error: %v", err)
	}
	if cfg2.User.Name != "NewUser" {
		t.Errorf("second load User.Name: got %q, want %q", cfg2.User.Name, "NewUser")
	}
}

func TestLoaderLoadGitConventionSection(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	root := setupTestdataDir(t, tempDir, []string{"git-convention.yaml"})

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GitConvention.Convention != "conventional-commits" {
		t.Errorf("GitConvention.Convention: got %q, want %q",
			cfg.GitConvention.Convention, "conventional-commits")
	}
	if !cfg.GitConvention.Validation.EnforceOnPush {
		t.Error("GitConvention.Validation.EnforceOnPush: expected true")
	}
	if cfg.GitConvention.AutoDetection.Enabled {
		t.Error("GitConvention.AutoDetection.Enabled: expected false")
	}
	if cfg.GitConvention.Validation.MaxLength != 100 {
		t.Errorf("GitConvention.Validation.MaxLength: got %d, want 100", cfg.GitConvention.Validation.MaxLength)
	}
	if cfg.GitConvention.AutoDetection.SampleSize != 50 {
		t.Errorf("GitConvention.AutoDetection.SampleSize: got %d, want 50", cfg.GitConvention.AutoDetection.SampleSize)
	}

	sections := loader.LoadedSections()
	if !sections["git_convention"] {
		t.Error("expected git_convention section to be loaded")
	}
}

func TestLoaderLoadStateSection(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	stateYAML := []byte(`state:
  state_dir: ".moai/custom-state"
`)
	if err := os.WriteFile(filepath.Join(sectionsDir, "state.yaml"), stateYAML, 0o644); err != nil {
		t.Fatalf("failed to write state.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.State.StateDir != ".moai/custom-state" {
		t.Errorf("State.StateDir: got %q, want %q", cfg.State.StateDir, ".moai/custom-state")
	}

	sections := loader.LoadedSections()
	if !sections["state"] {
		t.Error("expected state section to be loaded")
	}
}

func TestLoaderStateDefaults(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	// Load without a state.yaml - should use defaults
	root := setupTestdataDir(t, tempDir, []string{"user.yaml"})

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.State.StateDir != DefaultStateDir {
		t.Errorf("State.StateDir: got %q, want default %q", cfg.State.StateDir, DefaultStateDir)
	}

	sections := loader.LoadedSections()
	if sections["state"] {
		t.Error("expected state section to NOT be loaded when file is missing")
	}
}

func TestLoaderGitConventionDefaults(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	// Load with only user.yaml - git-convention should use defaults
	root := setupTestdataDir(t, tempDir, []string{"user.yaml"})

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GitConvention.Convention != DefaultGitConvention {
		t.Errorf("GitConvention.Convention: got %q, want default %q",
			cfg.GitConvention.Convention, DefaultGitConvention)
	}
	if cfg.GitConvention.Validation.EnforceOnPush {
		t.Error("GitConvention.Validation.EnforceOnPush: expected default false")
	}
	if !cfg.GitConvention.AutoDetection.Enabled {
		t.Error("GitConvention.AutoDetection.Enabled: expected default true")
	}
	if cfg.GitConvention.Validation.MaxLength != DefaultGitConventionMaxLength {
		t.Errorf("GitConvention.Validation.MaxLength: got %d, want default %d",
			cfg.GitConvention.Validation.MaxLength, DefaultGitConventionMaxLength)
	}

	sections := loader.LoadedSections()
	if sections["git_convention"] {
		t.Error("expected git_convention section to NOT be loaded")
	}
}

// TestLoaderMIG003SectionsLoadedViaLoaderLoad verifies that the 4 new MIG-003
// sections are correctly loaded when valid YAML files are present.
// Covers the happy path of loadConstitutionSection / loadContextSection /
// loadInterviewSection / loadDesignSection (T-MIG003-18).
func TestLoaderMIG003SectionsLoadedViaLoaderLoad(t *testing.T) {
	t.Parallel()
	resetSunsetNoticeOnce() // avoid notice side-effects

	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Copy valid fixture files into the sections dir.
	fixtures := map[string]string{
		"constitution.yaml": filepath.Join("testdata", "constitution-valid", "constitution.yaml"),
		"context.yaml":      filepath.Join("testdata", "context-valid", "context.yaml"),
		"interview.yaml":    filepath.Join("testdata", "interview-valid", "interview.yaml"),
		"design.yaml":       filepath.Join("testdata", "design-valid", "design.yaml"),
	}
	for dst, src := range fixtures {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read fixture %s: %v", src, err)
		}
		if err := os.WriteFile(filepath.Join(sectionsDir, dst), data, 0o644); err != nil {
			t.Fatalf("write %s: %v", dst, err)
		}
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tmpDir, ".moai"))
	if err != nil {
		t.Fatalf("Loader.Load(): %v", err)
	}

	// Verify all 4 sections loaded.
	sections := loader.LoadedSections()
	for _, key := range []string{"constitution", "context_search", "interview", "design"} {
		if !sections[key] {
			t.Errorf("section %q not marked as loaded", key)
		}
	}

	// Spot-check values.
	if len(cfg.Constitution.ForbiddenPatterns) == 0 {
		t.Error("Constitution.ForbiddenPatterns: expected non-empty")
	}
	if cfg.ContextSearch.TokenBudget.MaxInjectionTokens != 5000 {
		t.Errorf("ContextSearch.TokenBudget.MaxInjectionTokens: want 5000, got %d",
			cfg.ContextSearch.TokenBudget.MaxInjectionTokens)
	}
	if cfg.Interview.ClarityThreshold != 4 {
		t.Errorf("Interview.ClarityThreshold: want 4, got %d", cfg.Interview.ClarityThreshold)
	}
	if cfg.Design.GanLoop.PassThreshold != 0.75 {
		t.Errorf("Design.GanLoop.PassThreshold: want 0.75, got %f", cfg.Design.GanLoop.PassThreshold)
	}
}

// TestLoaderMIG003MalformedSectionsUseDefaults verifies that when a MIG-003 section
// YAML is malformed, the loader skips it (with slog.Warn) and uses defaults.
// Covers the slog.Warn branch in loadConstitutionSection / loadContextSection /
// loadInterviewSection / loadDesignSection.
func TestLoaderMIG003MalformedSectionsUseDefaults(t *testing.T) {
	t.Parallel()
	resetSunsetNoticeOnce()

	for _, tc := range []struct {
		file    string
		section string
	}{
		{"constitution.yaml", "constitution"},
		{"context.yaml", "context_search"},
		{"interview.yaml", "interview"},
		{"design.yaml", "design"},
	} {
		tc := tc
		t.Run(tc.file, func(t *testing.T) {
			t.Parallel()
			resetSunsetNoticeOnce()

			tmpDir := t.TempDir()
			sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
			if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			// Write malformed YAML for this section.
			malformed := []byte(tc.file[:len(tc.file)-5] + ":\n  broken: [\n")
			if err := os.WriteFile(filepath.Join(sectionsDir, tc.file), malformed, 0o644); err != nil {
				t.Fatalf("write %s: %v", tc.file, err)
			}

			loader := NewLoader()
			_, err := loader.Load(filepath.Join(tmpDir, ".moai"))
			if err != nil {
				t.Fatalf("Loader.Load() should not error on malformed section: %v", err)
			}

			// Section should NOT be marked loaded (slog.Warn path taken).
			sections := loader.LoadedSections()
			if sections[tc.section] {
				t.Errorf("section %q should NOT be loaded for malformed YAML", tc.section)
			}
		})
	}
}

// TestLoadHarnessConfigValid verifies that a valid harness config with a
// per_iteration value loads without error (AC-HRN-002-04).
func TestLoadHarnessConfigValid(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	src := filepath.Join("testdata", "eval-memory-valid", "harness.yaml")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", src, err)
	}
	dst := filepath.Join(tempDir, "harness.yaml")
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("failed to write harness.yaml: %v", err)
	}

	cfg, err := LoadHarnessConfig(dst)
	if err != nil {
		t.Fatalf("LoadHarnessConfig() error = %v, want nil", err)
	}
	if cfg.Evaluator.MemoryScope != "per_iteration" {
		t.Errorf("MemoryScope = %q, want %q", cfg.Evaluator.MemoryScope, "per_iteration")
	}
}

// TestLoadHarnessConfigFrozenViolation verifies that a cumulative value is
// rejected by loader validation with ErrEvalMemoryFrozen (AC-HRN-002-04).
func TestLoadHarnessConfigFrozenViolation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	src := filepath.Join("testdata", "eval-memory-frozen-violation", "harness.yaml")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", src, err)
	}
	dst := filepath.Join(tempDir, "harness.yaml")
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("failed to write harness.yaml: %v", err)
	}

	_, err = LoadHarnessConfig(dst)
	if err == nil {
		t.Fatal("LoadHarnessConfig() error = nil, want ErrEvalMemoryFrozen")
	}
	if !errors.Is(err, ErrEvalMemoryFrozen) {
		t.Errorf("LoadHarnessConfig() error = %v, want errors.Is(ErrEvalMemoryFrozen)", err)
	}
}

// TestLoadHarnessConfigMissingField verifies that a missing evaluator key
// returns a required-field error (AC-HRN-002-04).
func TestLoadHarnessConfigMissingField(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	src := filepath.Join("testdata", "eval-memory-missing", "harness.yaml")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", src, err)
	}
	dst := filepath.Join(tempDir, "harness.yaml")
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("failed to write harness.yaml: %v", err)
	}

	_, err = LoadHarnessConfig(dst)
	if err == nil {
		t.Fatal("LoadHarnessConfig() error = nil, want required-field error")
	}
	// Missing field must surface as ErrEvalMemoryFrozen or ErrInvalidConfig family.
	if !errors.Is(err, ErrEvalMemoryFrozen) && !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("LoadHarnessConfig() error = %v, want ErrEvalMemoryFrozen or ErrInvalidConfig", err)
	}
}
