// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M1 CG mode detector tests.
//
// Tests REQ-WTL-006 (CG mode detection definition) and REQ-WTL-009 (drift
// warning when teammateMode=tmux but GLM env is absent). Covers AC-WTL-005.

package tmux

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSettingsFile is a test helper that writes a settings.local.json with
// the given teammateMode value (or omits the field entirely when empty).
func writeSettingsFile(t *testing.T, dir string, teammateMode string) string {
	t.Helper()
	settingsPath := filepath.Join(dir, "settings.local.json")
	var content string
	if teammateMode == "" {
		content = `{"otherField": "value"}`
	} else {
		content = `{"teammateMode": "` + teammateMode + `"}`
	}
	if err := os.WriteFile(settingsPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}
	return settingsPath
}

// clearGLMEnv removes GLM-related env vars for the duration of the test.
func clearGLMEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")
}

// TestIsCGMode_InTmux_TeammateMode_GLMToken_True verifies AC-WTL-005 scenario 1:
// in tmux + teammateMode=tmux + ANTHROPIC_AUTH_TOKEN set → expect true.
func TestIsCGMode_InTmux_TeammateMode_GLMToken_True(t *testing.T) {
	dir := t.TempDir()
	settingsPath := writeSettingsFile(t, dir, "tmux")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "glm-token-abc")
	t.Setenv("ANTHROPIC_BASE_URL", "")

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("IsCGMode = false, want true (in tmux + teammateMode=tmux + GLM token)")
	}
	if buf.Len() != 0 {
		t.Errorf("expected no warning output, got: %q", buf.String())
	}
}

// TestIsCGMode_InTmux_TeammateMode_NoGLM_False_WithWarning verifies AC-WTL-005
// scenario 2 and REQ-WTL-009: in tmux + teammateMode=tmux + no GLM env → expect
// false AND stderr contains "GLM env vars are absent".
func TestIsCGMode_InTmux_TeammateMode_NoGLM_False_WithWarning(t *testing.T) {
	dir := t.TempDir()
	settingsPath := writeSettingsFile(t, dir, "tmux")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	clearGLMEnv(t)

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false (drift case: teammateMode=tmux but no GLM env)")
	}
	if !strings.Contains(buf.String(), "GLM env vars are absent") {
		t.Errorf("expected stderr warning containing 'GLM env vars are absent', got: %q", buf.String())
	}
}

// TestIsCGMode_NoTmux_False verifies AC-WTL-005 scenario 3: not in tmux → false.
func TestIsCGMode_NoTmux_False(t *testing.T) {
	dir := t.TempDir()
	settingsPath := writeSettingsFile(t, dir, "tmux")
	t.Setenv("TMUX", "")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "glm-token-abc")

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false (not in tmux)")
	}
}

// TestIsCGMode_InTmux_NoTeammateMode_False verifies AC-WTL-005 scenario 4:
// in tmux + no teammateMode + no GLM env → false.
func TestIsCGMode_InTmux_NoTeammateMode_False(t *testing.T) {
	dir := t.TempDir()
	settingsPath := writeSettingsFile(t, dir, "")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	clearGLMEnv(t)

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false (teammateMode not set)")
	}
}

// TestIsCGMode_NoSettingsFile_GracefulFalse verifies graceful handling when
// settings.local.json does not exist: expect (false, nil), not an error.
func TestIsCGMode_NoSettingsFile_GracefulFalse(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.local.json") // not created
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("expected nil error for missing settings file, got: %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false (no settings file)")
	}
}

// TestIsCGMode_CorruptJSON_Error verifies that a corrupt settings.local.json
// returns (false, error). REQ-WTL-006 implementation contract.
func TestIsCGMode_CorruptJSON_Error(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.local.json")
	if err := os.WriteFile(settingsPath, []byte("{invalid json"), 0o600); err != nil {
		t.Fatalf("write corrupt settings: %v", err)
	}
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err == nil {
		t.Errorf("expected non-nil error for corrupt JSON, got nil")
	}
	if got {
		t.Errorf("IsCGMode = true, want false on parse error")
	}
}

// TestIsCGMode_BaseURL_AlsoCounts verifies that ANTHROPIC_BASE_URL containing
// "z.ai" is sufficient evidence of GLM env (REQ-WTL-006: "GLM token OR base URL").
func TestIsCGMode_BaseURL_AlsoCounts(t *testing.T) {
	dir := t.TempDir()
	settingsPath := writeSettingsFile(t, dir, "tmux")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_BASE_URL", "https://api.z.ai/v1")

	var buf bytes.Buffer
	got, err := IsCGMode(settingsPath, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("IsCGMode = false, want true (BASE_URL contains z.ai)")
	}
}

// TestIsCGMode_NilStderrSink verifies nil stderrSink is tolerated without panic
// when drift warning would have been emitted.
func TestIsCGMode_NilStderrSink(t *testing.T) {
	dir := t.TempDir()
	settingsPath := writeSettingsFile(t, dir, "tmux")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	clearGLMEnv(t)

	got, err := IsCGMode(settingsPath, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("IsCGMode = true, want false")
	}
}

// TestIsCGMode_DriftWarning is a dedicated test for the drift warning behavior
// (AC-WTL-005 verification command 2). Keeps assertions narrow on the stderr
// substring for the grep-based AC verification.
func TestIsCGMode_DriftWarning(t *testing.T) {
	dir := t.TempDir()
	settingsPath := writeSettingsFile(t, dir, "tmux")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	clearGLMEnv(t)

	var buf bytes.Buffer
	_, _ = IsCGMode(settingsPath, &buf)
	out := buf.String()
	if !strings.Contains(out, "GLM env vars are absent") {
		t.Errorf("expected drift warning containing 'GLM env vars are absent', got: %q", out)
	}
}
