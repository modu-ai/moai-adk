// SPEC-V3R6-CG-MODE-HARDENING-001 — M3 detector SSOT realignment tests.
//
// Tests REQ-CGH-006: IsCGMode detects CG mode from the deterministic, design-clean
// source of truth (llm.yaml team_mode == "cg" + tmux SESSION env GLM markers) so a
// teammate spawn from a CLEAN leader process env is correctly classified. The new
// path is ADDITIVE; the existing teammateMode=tmux + process-env path is preserved
// (sibling tests in cg_detect_test.go stay green — verified separately).

package tmux

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeProjectWithTeamMode builds a temp project layout with a .claude dir holding
// settings.local.json (with the given teammateMode) and a .moai/config/sections
// dir holding llm.yaml (with the given team_mode). It returns the settings path so
// IsCGMode(settingsPath, ...) can derive the project root from it.
func writeProjectWithTeamMode(t *testing.T, teammateMode, llmTeamMode string) string {
	t.Helper()
	root := t.TempDir()

	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	var content string
	if teammateMode == "" {
		content = `{"otherField":"value"}`
	} else {
		content = `{"teammateMode":"` + teammateMode + `"}`
	}
	if err := os.WriteFile(settingsPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if llmTeamMode != "" {
		llmContent := "llm:\n  team_mode: " + llmTeamMode + "\n"
		if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(llmContent), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return settingsPath
}

// TestIsCGMode_AuthoritativeOnTeamMode verifies AC-CGH-006 Scenario 6a:
// with llm.yaml team_mode == "cg" AND the tmux session-env carrying GLM markers,
// IsCGMode returns true EVEN WHEN the leader PROCESS env is clean (no
// ANTHROPIC_AUTH_TOKEN, no z.ai base URL). This is the headline fix: CG mode is
// detectable from a clean leader pane.
func TestIsCGMode_AuthoritativeOnTeamMode(t *testing.T) {
	settingsPath := writeProjectWithTeamMode(t, "tmux", "cg")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origReader := sessionEnvReaderFn
	defer func() { sessionEnvReaderFn = origReader }()
	sessionEnvReaderFn = func() (map[string]string, error) {
		return map[string]string{"ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic"}, nil
	}

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("IsCGMode = false, want true (team_mode=cg + session-env GLM, clean process env)")
	}
	if buf.Len() != 0 {
		t.Errorf("expected no drift warning, got: %q", buf.String())
	}
}

// TestIsCGMode_TeamModeCG_SessionEnvMissing_DriftWarning verifies AC-CGH-006
// Scenario 6b on the NEW path: team_mode=cg but the tmux session env lacks GLM
// markers → the REQ-WTL-009 drift warning fires (reconciled, relocated to the
// session-env layer) and IsCGMode returns false.
func TestIsCGMode_TeamModeCG_SessionEnvMissing_DriftWarning(t *testing.T) {
	settingsPath := writeProjectWithTeamMode(t, "tmux", "cg")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origReader := sessionEnvReaderFn
	defer func() { sessionEnvReaderFn = origReader }()
	sessionEnvReaderFn = func() (map[string]string, error) {
		return map[string]string{}, nil
	}

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false (team_mode=cg but session env lacks GLM)")
	}
	if !strings.Contains(buf.String(), "GLM env vars are absent") {
		t.Errorf("expected REQ-WTL-009 drift warning containing 'GLM env vars are absent', got: %q", buf.String())
	}
}

// TestIsCGMode_NoTeamMode_FallsBackToProcessEnvPath verifies the PRESERVED existing
// path still works: with NO llm.yaml team_mode but teammateMode=tmux + a process-env
// GLM token, IsCGMode returns true via the existing fallback disjunct.
func TestIsCGMode_NoTeamMode_FallsBackToProcessEnvPath(t *testing.T) {
	settingsPath := writeProjectWithTeamMode(t, "tmux", "")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "glm-token-abc")
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origReader := sessionEnvReaderFn
	defer func() { sessionEnvReaderFn = origReader }()
	sessionEnvReaderFn = func() (map[string]string, error) { return map[string]string{}, nil }

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("IsCGMode = false, want true (preserved teammateMode=tmux + process-env token path)")
	}
}

// TestIsCGMode_NoTeamMode_NoProcessEnv_DriftWarning verifies the PRESERVED drift
// warning on the existing path: teammateMode=tmux, no team_mode, no process-env GLM
// → the original REQ-WTL-009 warning still fires and IsCGMode returns false.
func TestIsCGMode_NoTeamMode_NoProcessEnv_DriftWarning(t *testing.T) {
	settingsPath := writeProjectWithTeamMode(t, "tmux", "")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origReader := sessionEnvReaderFn
	defer func() { sessionEnvReaderFn = origReader }()
	sessionEnvReaderFn = func() (map[string]string, error) { return map[string]string{}, nil }

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false (no team_mode, teammateMode=tmux, no process-env GLM)")
	}
	if !strings.Contains(buf.String(), "GLM env vars are absent") {
		t.Errorf("expected preserved drift warning, got: %q", buf.String())
	}
}

// TestIsCGMode_SessionEnvReaderError_GracefulFalse verifies EC-4: when the session
// env reader errors (e.g. not in a real tmux), the new path degrades gracefully and
// IsCGMode falls through to the preserved process-env path (here: no process env →
// false), without crashing.
func TestIsCGMode_SessionEnvReaderError_GracefulFalse(t *testing.T) {
	settingsPath := writeProjectWithTeamMode(t, "tmux", "cg")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origReader := sessionEnvReaderFn
	defer func() { sessionEnvReaderFn = origReader }()
	sessionEnvReaderFn = func() (map[string]string, error) {
		return nil, os.ErrPermission
	}

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error (reader error must degrade gracefully): %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false (session reader errored, no process-env GLM)")
	}
}
