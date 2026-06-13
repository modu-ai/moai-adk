package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigManagerSaveGitStrategyRoundTrip verifies the git-strategy WRITE leg
// added by SPEC-PREPUSH-SAVE-WIRING-001 (the inverse of LOADER-WIRING-001's READ
// leg). It seeds GitStrategyConfig from NewDefaultGitStrategyConfig(), mutates
// two non-default values (a top-level field and a nested mode-profile field),
// Save()s, and confirms a fresh Load() recovers both — proving set→save→reload
// fidelity (AC-PSW-002, REQ-PSW-004).
func TestConfigManagerSaveGitStrategyRoundTrip(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Seed from defaults (Mode defaults to "team", Team.Hooks.PrePush to "warn"),
	// then mutate ONE top-level field and ONE nested field to non-default values.
	probe := NewDefaultGitStrategyConfig()
	probe.Mode = "personal"              // default is "team"
	probe.Team.Hooks.PrePush = "enforce" // default is "warn"

	if err := m.SetSection("git_strategy", probe); err != nil {
		t.Fatalf("SetSection(git_strategy) error: %v", err)
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload into a fresh manager and confirm both mutated values survived.
	m2 := NewConfigManager()
	cfg, err := m2.Load(root)
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if cfg.GitStrategy.Mode != "personal" {
		t.Errorf("GitStrategy.Mode round-trip: got %q, want %q", cfg.GitStrategy.Mode, "personal")
	}
	if cfg.GitStrategy.Team.Hooks.PrePush != "enforce" {
		t.Errorf("GitStrategy.Team.Hooks.PrePush round-trip: got %q, want %q",
			cfg.GitStrategy.Team.Hooks.PrePush, "enforce")
	}
}

// TestConfigManagerSaveCreatesGitStrategyFile verifies Save() creates
// git-strategy.yaml on disk alongside the other section files (AC-PSW-001,
// REQ-PSW-001 / REQ-PSW-002) and that the written file carries the top-level
// git_strategy: key produced by the reused gitStrategyFileWrapper (AC-PSW-003,
// REQ-PSW-003).
func TestConfigManagerSaveCreatesGitStrategyFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	m := NewConfigManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if _, err := os.Stat(gitStrategyPath); os.IsNotExist(err) {
		t.Fatalf("Save() did not create git-strategy.yaml at %s", gitStrategyPath)
	}

	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatalf("read git-strategy.yaml error: %v", err)
	}
	if !strings.Contains(string(data), "git_strategy:") {
		t.Errorf("git-strategy.yaml missing top-level git_strategy: key; got:\n%s", data)
	}
}
