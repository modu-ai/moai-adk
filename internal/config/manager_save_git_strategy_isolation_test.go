package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// TestSaveGitStrategyIsolation_PreservesUnknownFields is the cross-package
// regression coverage for issue #1064 / the SPEC-WEB-CONSOLE-003 M5 isolation
// contract (AC-WC3-007). It complements the internal/web seam-level test
// TestWriteProjectConfigSectionIsolation by exercising the same invariant at
// the config-manager layer:
//
//	A Save() call that did NOT explicitly SetSection("git_strategy", ...) MUST
//	leave an existing git-strategy.yaml file BYTE-IDENTICAL on disk. Unknown
//	YAML keys (sentinels, forward-compatible template additions) are silently
//	dropped by yaml.Unmarshal, so a naive round-trip would expand the file
//	into the full default tree and destroy the original content.
//
// This guards the boundary where a single-section wiring change in
// internal/config can silently break a sibling section's isolation — the
// failure mode that produced the bug in commit 33215af27.
func TestSaveGitStrategyIsolation_PreservesUnknownFields(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}

	// Seed git-strategy.yaml with a sentinel field that is NOT part of the
	// GitStrategyConfig schema. yaml.Unmarshal will silently drop it on load;
	// our gate must prevent Save() from re-serializing the section and
	// erasing the sentinel from disk.
	sentinelBytes := []byte("git_strategy:\n  sentinel: DO_NOT_TOUCH\n")
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, sentinelBytes, 0o644); err != nil {
		t.Fatalf("seed git-strategy.yaml: %v", err)
	}

	// Drive a quality-only write through the manager — the same pattern
	// internal/web.writeProjectConfig follows when persisting a
	// project-config change scoped to quality + git_convention.
	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	quality := m.Get().Quality
	quality.DevelopmentMode = models.DevelopmentMode("ddd")
	if err := m.SetSection("quality", quality); err != nil {
		t.Fatalf("SetSection(quality): %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// The git-strategy.yaml file must be byte-identical to the seeded
	// sentinel content — Save() must not have touched it.
	got, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatalf("read git-strategy.yaml after Save: %v", err)
	}
	if string(got) != string(sentinelBytes) {
		t.Errorf("git-strategy.yaml was rewritten by a scoped Save() (issue #1064 regression):\n  before: %q\n  after:  %q",
			sentinelBytes, got)
	}
}

// TestSaveGitStrategyIsolation_WritesWhenDirty proves the gate still permits
// the SPEC-PREPUSH-SAVE-WIRING-001 round-trip: when git_strategy was modified
// via SetSection, Save() persists the change even if a (now stale) file
// exists on disk.
func TestSaveGitStrategyIsolation_WritesWhenDirty(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}

	// Seed an existing git-strategy.yaml so loadedSections["git_strategy"]=true;
	// the SetSection below provides the dirty signal that authorizes the write.
	staleBytes := []byte("git_strategy:\n  mode: team\n")
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, staleBytes, 0o644); err != nil {
		t.Fatalf("seed git-strategy.yaml: %v", err)
	}

	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	probe := NewDefaultGitStrategyConfig()
	probe.Mode = "personal"
	if err := m.SetSection("git_strategy", probe); err != nil {
		t.Fatalf("SetSection(git_strategy): %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// A fresh load must observe the dirty value, proving Save() rewrote the file.
	m2 := NewConfigManager()
	cfg, err := m2.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw after Save: %v", err)
	}
	if cfg.GitStrategy.Mode != "personal" {
		t.Errorf("git_strategy mode did not round-trip through Save(): got %q, want %q",
			cfg.GitStrategy.Mode, "personal")
	}
}
