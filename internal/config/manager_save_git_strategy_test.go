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

// TestConfigManagerSaveDirtyFlagReset pins EC-3 of
// SPEC-GITSTRATEGY-SAVE-ISOLATION-001: the highest-risk property of the section
// isolation mechanism. After a Save() that wrote git-strategy.yaml (because the
// section was SetSection'd, so the dirty flag was true), the flag MUST reset to
// false. A subsequent Save() with NO further SetSection("git_strategy") MUST NOT
// rewrite an EXISTING git-strategy.yaml — proving the reset gates the skip path
// that restores cross-section isolation (REQ-GSI-001/002). Without the reset, the
// second Save would re-serialize the compiled-default tree and clobber any on-disk
// content the loader does not model.
func TestConfigManagerSaveDirtyFlagReset(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Mutate the git_strategy section and Save() — this sets dirty=true, so the
	// first Save writes git-strategy.yaml and then resets dirty to false.
	probe := NewDefaultGitStrategyConfig()
	probe.Mode = "personal"
	if err := m.SetSection("git_strategy", probe); err != nil {
		t.Fatalf("SetSection(git_strategy) error: %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("first Save() error: %v", err)
	}

	gitStrategyPath := filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")

	// Overwrite git-strategy.yaml on disk with a sentinel the loader does not model.
	// This stands in for any out-of-band content (e.g. a hand-edited key) that a
	// subsequent isolation-respecting Save() must leave untouched.
	sentinel := []byte("git_strategy:\n  sentinel: DO_NOT_TOUCH\n")
	if err := os.WriteFile(gitStrategyPath, sentinel, 0o644); err != nil {
		t.Fatalf("write sentinel to git-strategy.yaml: %v", err)
	}

	// Second Save() with NO further SetSection("git_strategy"): the dirty flag was
	// reset by the first Save and the file exists, so git-strategy.yaml MUST be
	// left byte-identical to the sentinel we just wrote.
	if err := m.Save(); err != nil {
		t.Fatalf("second Save() error: %v", err)
	}

	after, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatalf("read git-strategy.yaml after second Save(): %v", err)
	}
	if string(after) != string(sentinel) {
		t.Errorf("git-strategy.yaml rewritten by a Save() that did not modify git_strategy "+
			"(dirty flag not reset):\nwant: %q\ngot:  %q", sentinel, after)
	}
}

// TestConfigManagerReloadResetsDirtyFlag pins EC-3/D4 from the Reload() side: a
// whole-config replacement via Reload() MUST clear the git_strategy dirty flag, so
// a Save() after Reload (with no fresh SetSection) does NOT rewrite an existing
// git-strategy.yaml. This guards the isolation invariant across the Reload path,
// complementing the Load/Save path covered above.
func TestConfigManagerReloadResetsDirtyFlag(t *testing.T) {
	t.Parallel()

	root := setupManagerTestDir(t, []string{"user.yaml", "language.yaml", "quality.yaml"})
	m := NewConfigManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	gitStrategyPath := filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")
	// Seed an existing git-strategy.yaml with a sentinel so the file-absent create
	// disjunct does not mask the dirty-reset behavior.
	sentinel := []byte("git_strategy:\n  sentinel: DO_NOT_TOUCH\n")
	if err := os.WriteFile(gitStrategyPath, sentinel, 0o644); err != nil {
		t.Fatalf("seed git-strategy.yaml sentinel: %v", err)
	}

	// Set the section dirty, then Reload() — Reload replaces the whole config and
	// MUST reset the dirty flag.
	probe := NewDefaultGitStrategyConfig()
	probe.Mode = "personal"
	if err := m.SetSection("git_strategy", probe); err != nil {
		t.Fatalf("SetSection(git_strategy) error: %v", err)
	}
	if err := m.Reload(); err != nil {
		t.Fatalf("Reload() error: %v", err)
	}

	// Save() after Reload with no fresh SetSection: git-strategy.yaml exists, dirty
	// was reset by Reload, so the file MUST remain byte-identical to the sentinel.
	if err := m.Save(); err != nil {
		t.Fatalf("Save() after Reload() error: %v", err)
	}

	after, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatalf("read git-strategy.yaml after Reload()+Save(): %v", err)
	}
	if string(after) != string(sentinel) {
		t.Errorf("git-strategy.yaml rewritten after Reload() reset the dirty flag:\n"+
			"want: %q\ngot:  %q", sentinel, after)
	}
}
