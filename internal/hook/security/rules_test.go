package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// RED PHASE: Tests for RuleManager implementation
// These tests define expected behavior per SPEC-HOOK-003 REQ-HOOK-110~112
// =============================================================================

// TestNewRuleManager verifies rule manager creation.
func TestNewRuleManager(t *testing.T) {
	t.Run("creates rule manager instance", func(t *testing.T) {
		rm := NewRuleManager()
		if rm == nil {
			t.Fatal("expected non-nil rule manager")
		}
	})
}

// TestRuleManager_FindRulesConfig verifies config file discovery.
// REQ-HOOK-110: System must detect and load project-specific security rules.
func TestRuleManager_FindRulesConfig(t *testing.T) {
	rm := NewRuleManager()

	t.Run("returns empty string when no config found", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := rm.FindRulesConfig(tmpDir)

		if configPath != "" {
			t.Errorf("expected empty string for no config, got %q", configPath)
		}
	})

	t.Run("finds sgconfig.yml in project root", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "sgconfig.yml")
		if err := os.WriteFile(configFile, []byte("rules: []"), 0644); err != nil {
			t.Fatalf("failed to create config file: %v", err)
		}

		configPath := rm.FindRulesConfig(tmpDir)
		if configPath != configFile {
			t.Errorf("expected %q, got %q", configFile, configPath)
		}
	})

	t.Run("finds sgconfig.yaml in project root", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "sgconfig.yaml")
		if err := os.WriteFile(configFile, []byte("rules: []"), 0644); err != nil {
			t.Fatalf("failed to create config file: %v", err)
		}

		configPath := rm.FindRulesConfig(tmpDir)
		if configPath != configFile {
			t.Errorf("expected %q, got %q", configFile, configPath)
		}
	})

	t.Run("finds config in .ast-grep directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		astGrepDir := filepath.Join(tmpDir, ".ast-grep")
		if err := os.MkdirAll(astGrepDir, 0755); err != nil {
			t.Fatalf("failed to create .ast-grep dir: %v", err)
		}
		configFile := filepath.Join(astGrepDir, "sgconfig.yml")
		if err := os.WriteFile(configFile, []byte("rules: []"), 0644); err != nil {
			t.Fatalf("failed to create config file: %v", err)
		}

		configPath := rm.FindRulesConfig(tmpDir)
		if configPath != configFile {
			t.Errorf("expected %q, got %q", configFile, configPath)
		}
	})

	t.Run("finds config in .claude/skills/moai-tool-ast-grep/rules", func(t *testing.T) {
		tmpDir := t.TempDir()
		rulesDir := filepath.Join(tmpDir, ".claude", "skills", "moai-tool-ast-grep", "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			t.Fatalf("failed to create rules dir: %v", err)
		}
		configFile := filepath.Join(rulesDir, "sgconfig.yml")
		if err := os.WriteFile(configFile, []byte("rules: []"), 0644); err != nil {
			t.Fatalf("failed to create config file: %v", err)
		}

		configPath := rm.FindRulesConfig(tmpDir)
		if configPath != configFile {
			t.Errorf("expected %q, got %q", configFile, configPath)
		}
	})

	t.Run("prioritizes project root over .ast-grep", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create both configs
		rootConfig := filepath.Join(tmpDir, "sgconfig.yml")
		_ = os.WriteFile(rootConfig, []byte("rules: []"), 0644)

		astGrepDir := filepath.Join(tmpDir, ".ast-grep")
		_ = os.MkdirAll(astGrepDir, 0755)
		_ = os.WriteFile(filepath.Join(astGrepDir, "sgconfig.yml"), []byte("rules: []"), 0644)

		configPath := rm.FindRulesConfig(tmpDir)
		if configPath != rootConfig {
			t.Errorf("expected root config %q, got %q", rootConfig, configPath)
		}
	})
}

// TestRuleManager_LoadRules verifies rule loading.
// REQ-HOOK-110, REQ-HOOK-112.
func TestRuleManager_LoadRules(t *testing.T) {
	rm := NewRuleManager()

	t.Run("returns error for non-existent config", func(t *testing.T) {
		_, err := rm.LoadRules("/nonexistent/sgconfig.yml")
		if err == nil {
			t.Error("expected error for non-existent config")
		}
	})

	t.Run("loads rules from valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "sgconfig.yml")
		content := `
ruleDirs:
  - ./rules
rules:
  sql-injection:
    severity: error
    message: Potential SQL injection
  hardcoded-secret:
    severity: warning
    message: Hardcoded secret detected
`
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create config file: %v", err)
		}

		rules, err := rm.LoadRules(configFile)
		if err != nil {
			t.Fatalf("failed to load rules: %v", err)
		}
		if len(rules) == 0 {
			t.Error("expected rules to be loaded")
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "sgconfig.yml")
		content := `invalid: yaml: content: [[[`
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create config file: %v", err)
		}

		_, err := rm.LoadRules(configFile)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})
}

// TestRuleManager_GetDefaultRules verifies default OWASP rules.
// REQ-HOOK-111: System must use built-in OWASP rules when no config found.
func TestRuleManager_GetDefaultRules(t *testing.T) {
	rm := NewRuleManager()

	t.Run("returns non-empty default rules", func(t *testing.T) {
		rules := rm.GetDefaultRules()
		if len(rules) == 0 {
			t.Error("expected non-empty default rules")
		}
	})

	t.Run("includes OWASP rules", func(t *testing.T) {
		rules := rm.GetDefaultRules()

		// Should contain common security patterns
		var rulesStr strings.Builder
		for _, r := range rules {
			rulesStr.WriteString(r)
		}

		// Check for common security rule indicators
		expectedPatterns := []string{
			"sql", "injection", "xss", "secret", "password", "eval",
		}

		foundCount := 0
		for _, pattern := range expectedPatterns {
			if containsIgnoreCase(rulesStr.String(), pattern) {
				foundCount++
			}
		}

		// At least 3 of these patterns should be present
		if foundCount < 3 {
			t.Errorf("expected at least 3 OWASP-related patterns, found %d", foundCount)
		}
	})
}

// TestRuleManager_FallbackBehavior verifies fallback to default rules.
// REQ-HOOK-112: System must fall back to default rules on invalid config.
func TestRuleManager_FallbackBehavior(t *testing.T) {
	rm := NewRuleManager()

	t.Run("GetEffectiveRules returns default rules when no config", func(t *testing.T) {
		tmpDir := t.TempDir() // Empty directory
		rules := rm.GetEffectiveRules(tmpDir)

		if len(rules) == 0 {
			t.Error("expected default rules to be returned")
		}

		// Should be same as GetDefaultRules
		defaultRules := rm.GetDefaultRules()
		if len(rules) != len(defaultRules) {
			t.Errorf("expected %d rules, got %d", len(defaultRules), len(rules))
		}
	})

	t.Run("GetEffectiveRules returns project rules when config exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "sgconfig.yml")
		content := `
rules:
  custom-rule:
    severity: warning
    message: Custom rule
`
		_ = os.WriteFile(configFile, []byte(content), 0644)

		rules := rm.GetEffectiveRules(tmpDir)
		if len(rules) == 0 {
			t.Error("expected project rules to be returned")
		}
	})
}

// containsIgnoreCase checks if s contains substr (case insensitive).
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > 0 && len(substr) > 0 &&
				(s[0] == substr[0] || s[0]+32 == substr[0] || s[0]-32 == substr[0]) &&
				containsIgnoreCase(s[1:], substr[1:]) ||
			len(s) > 0 && containsIgnoreCase(s[1:], substr))
}
