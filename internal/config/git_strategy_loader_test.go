package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeGitStrategyFixture creates a .moai/config/sections/ layout under a
// t.TempDir() root and writes the given git-strategy.yaml body into it. It
// returns the .moai dir to pass to Loader.Load(). When body is empty, no
// git-strategy.yaml file is written (the absent-file case).
//
// The fixtures are routed through Loader.Load() — NOT a direct yaml.Unmarshal —
// so the loader wiring path (loadGitStrategySection) is exercised end-to-end
// (see git_strategy_nested_test.go for the direct-unmarshal counterpart).
func writeGitStrategyFixture(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}
	if body != "" {
		dst := filepath.Join(sectionsDir, "git-strategy.yaml")
		if err := os.WriteFile(dst, []byte(body), 0o644); err != nil {
			t.Fatalf("failed to write git-strategy.yaml: %v", err)
		}
	}
	return filepath.Join(root, ".moai")
}

// AC-PLW-001 — git-strategy.yaml present → values loaded + section flagged
// (REQ-PLW-001, REQ-PLW-002).
func TestLoader_GitStrategy_Present_LoadsValuesAndFlagsSection(t *testing.T) {
	t.Parallel()

	const fixture = `git_strategy:
  mode: team
  team:
    hooks:
      pre_push: enforce
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// The file's non-default mode-profile hook value is observed (not the
	// compiled default "warn"). See AC-PLW-007 for the ActiveModeProfile assertion.
	if got := cfg.GitStrategy.Team.Hooks.PrePush; got != "enforce" {
		t.Errorf("GitStrategy.Team.Hooks.PrePush: got %q, want %q (file value, not compiled default)", got, "enforce")
	}

	loaded := loader.LoadedSections()
	if !loaded["git_strategy"] {
		t.Errorf("LoadedSections()[\"git_strategy\"]: got false, want true")
	}
}

// AC-PLW-002 — git-strategy.yaml absent → defaults kept + flag unset
// (REQ-PLW-003).
func TestLoader_GitStrategy_Absent_KeepsDefaultsAndFlagUnset(t *testing.T) {
	t.Parallel()

	// No git-strategy.yaml written; the sections dir exists and is otherwise valid.
	moaiDir := writeGitStrategyFixture(t, "")

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Compiled defaults preserved: default team profile pre_push is "warn".
	wantDefault := NewDefaultGitStrategyConfig()
	if cfg.GitStrategy.Team.Hooks.PrePush != wantDefault.Team.Hooks.PrePush {
		t.Errorf("GitStrategy.Team.Hooks.PrePush: got %q, want compiled default %q",
			cfg.GitStrategy.Team.Hooks.PrePush, wantDefault.Team.Hooks.PrePush)
	}
	if cfg.GitStrategy.Team.Hooks.PrePush != "warn" {
		t.Errorf("sanity: compiled default Team.Hooks.PrePush expected %q, got %q", "warn", cfg.GitStrategy.Team.Hooks.PrePush)
	}

	loaded := loader.LoadedSections()
	if loaded["git_strategy"] {
		t.Errorf("LoadedSections()[\"git_strategy\"]: got true, want false (file absent)")
	}
}

// AC-PLW-003 — partial git-strategy.yaml → specified keys override, unspecified
// keep defaults (REQ-PLW-004).
func TestLoader_GitStrategy_Partial_OverridesSpecifiedKeepsDefaults(t *testing.T) {
	t.Parallel()

	// Only the top-level mode is set; the personal profile's hook blocks are omitted.
	const fixture = `git_strategy:
  mode: personal
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Specified key overridden.
	if cfg.GitStrategy.Mode != "personal" {
		t.Errorf("GitStrategy.Mode: got %q, want %q (file override)", cfg.GitStrategy.Mode, "personal")
	}

	// Unspecified key keeps its compiled default (personal profile pre_push stays "warn").
	wantDefault := NewDefaultGitStrategyConfig()
	if cfg.GitStrategy.Personal.Hooks.PrePush != wantDefault.Personal.Hooks.PrePush {
		t.Errorf("GitStrategy.Personal.Hooks.PrePush: got %q, want compiled default %q (not overridden by file)",
			cfg.GitStrategy.Personal.Hooks.PrePush, wantDefault.Personal.Hooks.PrePush)
	}
	if cfg.GitStrategy.Personal.Hooks.PrePush != "warn" {
		t.Errorf("sanity: default Personal.Hooks.PrePush expected %q, got %q", "warn", cfg.GitStrategy.Personal.Hooks.PrePush)
	}
}

// AC-PLW-004 — malformed / unknown-key git-strategy.yaml → Load() does not fail
// (REQ-PLW-005).
func TestLoader_GitStrategy_UnknownKeys_LoadDoesNotFail(t *testing.T) {
	t.Parallel()

	const fixture = `git_strategy:
  mode: team
  bogus_key: 123
  team:
    hooks:
      pre_push: enforce
    bogus_nested: "ignored"
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() returned error on unknown keys (non-strict yaml.Unmarshal expected): %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Recognized keys still loaded normally despite the unknown keys.
	if cfg.GitStrategy.Mode != "team" {
		t.Errorf("GitStrategy.Mode: got %q, want %q", cfg.GitStrategy.Mode, "team")
	}
	if cfg.GitStrategy.Team.Hooks.PrePush != "enforce" {
		t.Errorf("GitStrategy.Team.Hooks.PrePush: got %q, want %q", cfg.GitStrategy.Team.Hooks.PrePush, "enforce")
	}
}

// AC-PLW-007 — end-to-end chain completion: ActiveModeProfile reads YAML hook
// value (REQ-PLW-008). This is the headline AC proving SPEC-PREPUSH-MODE-WIRING-001's
// resolvePrePushAction() now sees the real file value end-to-end.
func TestLoader_GitStrategy_EndToEnd_ActiveModeProfileReadsYAMLHook(t *testing.T) {
	t.Parallel()

	const fixture = `git_strategy:
  mode: team
  team:
    hooks:
      pre_push: enforce
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	profile, ok := cfg.GitStrategy.ActiveModeProfile()
	if !ok {
		t.Fatal("ActiveModeProfile() returned ok=false; expected true for mode=team")
	}
	if profile.Hooks.PrePush != "enforce" {
		t.Errorf("ActiveModeProfile().Hooks.PrePush: got %q, want %q (YAML value flowed end-to-end; "+
			"without the loader wiring this would be the compiled default %q)",
			profile.Hooks.PrePush, "enforce", "warn")
	}
}

// Edge case — empty git_strategy block (file present, no children) → defaults
// preserved, section flag still true (the flag keys off file-existed-and-parsed,
// not off whether values changed). Mirrors sibling-loader behavior.
func TestLoader_GitStrategy_EmptyBlock_DefaultsKeptFlagTrue(t *testing.T) {
	t.Parallel()

	const fixture = `git_strategy:
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Empty block overlays nothing → result equals compiled defaults.
	wantDefault := NewDefaultGitStrategyConfig()
	if cfg.GitStrategy.Team.Hooks.PrePush != wantDefault.Team.Hooks.PrePush {
		t.Errorf("GitStrategy.Team.Hooks.PrePush: got %q, want compiled default %q",
			cfg.GitStrategy.Team.Hooks.PrePush, wantDefault.Team.Hooks.PrePush)
	}

	// File existed and parsed → flag is true even though no value changed.
	loaded := loader.LoadedSections()
	if !loaded["git_strategy"] {
		t.Errorf("LoadedSections()[\"git_strategy\"]: got false, want true (file existed and parsed)")
	}
}

// Edge case — genuinely invalid YAML syntax (not just unknown keys) → loadYAMLFile
// errors, loadGitStrategySection logs slog.Warn + keeps defaults; Load() does NOT
// propagate the error and the section flag stays unset (mirror sibling resilience).
func TestLoader_GitStrategy_InvalidYAML_LoadResilientFlagUnset(t *testing.T) {
	t.Parallel()

	// Unbalanced brackets / broken mapping — unparseable.
	const fixture = "git_strategy: [unterminated\n  mode: team\n"
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() propagated a parse error; expected sibling resilience (slog.Warn + keep defaults): %v", err)
	}

	// Defaults preserved.
	wantDefault := NewDefaultGitStrategyConfig()
	if cfg.GitStrategy.Team.Hooks.PrePush != wantDefault.Team.Hooks.PrePush {
		t.Errorf("GitStrategy.Team.Hooks.PrePush: got %q, want compiled default %q after invalid-YAML resilience",
			cfg.GitStrategy.Team.Hooks.PrePush, wantDefault.Team.Hooks.PrePush)
	}

	// Section flag stays unset on parse failure.
	loaded := loader.LoadedSections()
	if loaded["git_strategy"] {
		t.Errorf("LoadedSections()[\"git_strategy\"]: got true, want false (parse failed)")
	}
}

// AC-MMC-002 — git-strategy.yaml present but omits merge_method → compiled default
// "squash" retained (REQ-MMC-003). Mirrors the partial-override contract proved by
// TestLoader_GitStrategy_Absent_KeepsDefaultsAndFlagUnset for the new field.
func TestLoader_GitStrategy_MergeMethod_Absent_KeepsDefault(t *testing.T) {
	t.Parallel()

	// team: block present, but merge_method omitted.
	const fixture = `git_strategy:
  mode: team
  team:
    hooks:
      pre_push: enforce
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.GitStrategy.Team.MergeMethod; got != "squash" {
		t.Errorf("GitStrategy.Team.MergeMethod: got %q, want compiled default %q (not empty)", got, "squash")
	}
}

// AC-MMC-003 — git-strategy.yaml sets merge_method → file value populated, NOT the
// compiled default (REQ-MMC-004). Mirrors TestLoader_GitStrategy_Present_LoadsValuesAndFlagsSection.
func TestLoader_GitStrategy_MergeMethod_Present_LoadsFileValue(t *testing.T) {
	t.Parallel()

	const fixture = `git_strategy:
  mode: team
  team:
    merge_method: merge
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.GitStrategy.Team.MergeMethod; got != "merge" {
		t.Errorf("GitStrategy.Team.MergeMethod: got %q, want %q (file value, not compiled default)", got, "merge")
	}
}

// AC-MMC-004 — partial merge_method override: one mode set, siblings keep default
// (REQ-MMC-003/004). Mirrors TestLoader_GitStrategy_Partial_OverridesSpecifiedKeepsDefaults.
func TestLoader_GitStrategy_MergeMethod_Partial_KeepsSiblingDefaults(t *testing.T) {
	t.Parallel()

	const fixture = `git_strategy:
  mode: team
  team:
    merge_method: rebase
`
	moaiDir := writeGitStrategyFixture(t, fixture)

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.GitStrategy.Team.MergeMethod; got != "rebase" {
		t.Errorf("GitStrategy.Team.MergeMethod: got %q, want %q (file override)", got, "rebase")
	}
	if got := cfg.GitStrategy.Manual.MergeMethod; got != "squash" {
		t.Errorf("GitStrategy.Manual.MergeMethod: got %q, want compiled default %q (not overridden)", got, "squash")
	}
	if got := cfg.GitStrategy.Personal.MergeMethod; got != "squash" {
		t.Errorf("GitStrategy.Personal.MergeMethod: got %q, want compiled default %q (not overridden)", got, "squash")
	}
}
