package config

// U-05 (AC-RC30-06, REQ-RC30-007): a Save of a dirty llm section over an
// existing llm.yaml changes only the field the caller set. deployment_mode,
// an unmodelled key, and comments survive byte-for-byte.
//
// Gap G-06 is resolved here as llm.team_mode: the field the caller sets is
// the one the audit saw lost in a re-marshal (U-24: team_mode reset from glm
// to ""), and it is the field the launcher reads (cli/launcher.go TeamMode).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSaveRoundTripLLM(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const fixture = "# llm header comment\n" +
		"llm:\n" +
		"  harness: claude\n" +
		"  team_mode: claude  # team note\n" +
		"  deployment_mode: plugin\n" +
		"  future_flag: keep-me\n" +
		"  glm:\n" +
		"    models:\n" +
		"      high: glm-5.2\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	llm := m.Get().LLM
	llm.TeamMode = "glm" // the single field this round-trip sets (Gap G-06)
	if err := m.SetSection("llm", llm); err != nil {
		t.Fatalf("SetSection(llm): %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	want := strings.Replace(fixture, "team_mode: claude", "team_mode: glm", 1)
	got := readFile(t, filepath.Join(sectionsDir, "llm.yaml"))
	if got != want {
		t.Errorf("llm.yaml after a team_mode-only Save changed more than the field:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
