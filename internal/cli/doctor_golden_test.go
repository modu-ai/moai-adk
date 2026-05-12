package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateDoctorGolden controls golden snapshot regeneration. Set via UPDATE_GOLDEN=1.
var updateDoctorGolden = os.Getenv("UPDATE_GOLDEN") == "1"

// doctorGoldenPath returns the path to a golden snapshot file under testdata/.
func doctorGoldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

// checkDoctorGolden compares got to a golden file, regenerating it if UPDATE_GOLDEN=1.
func checkDoctorGolden(t *testing.T, name, got string) {
	t.Helper()
	path := doctorGoldenPath(name)
	if updateDoctorGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Logf("updated golden: %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to generate)", path, err)
	}
	if got != string(want) {
		t.Errorf("doctor output mismatch for %s\ngot:\n%s\nwant:\n%s", name, got, string(want))
	}
}

// captureDoctorCmd executes doctorCmd and returns stdout as a string.
func captureDoctorCmd(t *testing.T) string {
	t.Helper()
	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)
	if err := doctorCmd.RunE(doctorCmd, []string{}); err != nil {
		t.Fatalf("doctorCmd.RunE: %v", err)
	}
	// Reset cobra output writers after capture to avoid test pollution.
	doctorCmd.SetOut(nil)
	doctorCmd.SetErr(nil)
	return buf.String()
}

// --- DDD PRESERVE: Characterization tests for doctor command output ---
//
// These tests capture the AFTER state of doctorCmd output (tui.Section + tui.CheckLine + summary Pills).
// They serve as the regression baseline for future DDD cycles.
//
// Deterministic env var pinning:
//   - MOAI_GO_VERSION_OVERRIDE=go1.99.99 → pins Go version in "Go Runtime" check
//   - CLAUDE_CODE_VERSION=test-claude-99 → pins Claude Code version check
//
// lipgloss AdaptiveColor behaviour under cmd.SetOut(buf):
//   - lipgloss detects non-TTY writer and disables ANSI colour output.
//   - Each env combination receives its own golden file for clarity.
//
// To regenerate snapshots:
//   UPDATE_GOLDEN=1 go test ./internal/cli/ -run "TestDoctor_Current" -count=1

// TestDoctor_Current_Light captures doctorCmd output with light-theme env.
// 특징: tui.Section + 19+ tui.CheckLine + tui.Box + tui.Pill 요약.
func TestDoctor_Current_Light(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "light")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	got := captureDoctorCmd(t)
	if len(got) == 0 {
		t.Fatal("doctorCmd produced no output")
	}
	checkDoctorGolden(t, "doctor-light", got)
}

// TestDoctor_Current_Dark captures doctorCmd output with dark-theme env.
// 특징: tui.DarkTheme() 적용, Section 헤더 + CheckLine + Pill 요약.
func TestDoctor_Current_Dark(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "dark")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	got := captureDoctorCmd(t)
	if len(got) == 0 {
		t.Fatal("doctorCmd produced no output")
	}
	checkDoctorGolden(t, "doctor-dark", got)
}

// TestDoctor_NoColor captures doctorCmd output with NO_COLOR=1 (plain text mode).
// tui.MonochromeTheme() 적용: 모든 ANSI 색상 제거, Pill은 [label] 형식으로 degraded.
func TestDoctor_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	got := captureDoctorCmd(t)
	if len(got) == 0 {
		t.Fatal("doctorCmd produced no output")
	}
	checkDoctorGolden(t, "doctor-nocolor", got)
}

// --- DDD PRESERVE: structural invariant tests (no golden snapshots) ---

// TestDoctor_CheckCount verifies at least 19 checks are registered (AC-CLI-TUI-003).
func TestDoctor_CheckCount(t *testing.T) {
	checks := runDiagnosticChecks(false, "")
	if len(checks) < 19 {
		t.Errorf("expected >= 19 diagnostic checks (AC-CLI-TUI-003), got %d", len(checks))
	}
}

// TestDoctor_GroupsPresent verifies System, MoAI-ADK, and Workspace groups are rendered.
func TestDoctor_GroupsPresent(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	got := captureDoctorCmd(t)
	for _, group := range []string{"System", "MoAI-ADK", "Workspace"} {
		if !strings.Contains(got, group) {
			t.Errorf("doctor output should contain group %q", group)
		}
	}
}

// TestDoctor_GlamourCachePlaceholder verifies the D8 Glamour Cache placeholder is present.
func TestDoctor_GlamourCachePlaceholder(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	got := captureDoctorCmd(t)
	if !strings.Contains(got, "Glamour Cache") {
		t.Errorf("doctor output should contain 'Glamour Cache' D8 placeholder")
	}
	if !strings.Contains(got, "glamour 미도입") {
		t.Errorf("doctor output should contain 'glamour 미도입' placeholder message")
	}
}

// TestDoctor_SummaryPillsPresent verifies the summary pills appear in the output.
func TestDoctor_SummaryPillsPresent(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	got := captureDoctorCmd(t)
	// 통과/주의/실패 Pill 중 최소 하나는 출력에 있어야 함.
	if !strings.Contains(got, "통과") {
		t.Errorf("doctor output should contain '통과' pill in summary")
	}
}

// TestDoctor_GoVersionDeterministic verifies MOAI_GO_VERSION_OVERRIDE is honored.
func TestDoctor_GoVersionDeterministic(t *testing.T) {
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")
	t.Setenv("NO_COLOR", "1")

	got := captureDoctorCmd(t)
	if !strings.Contains(got, "1.99.99") {
		t.Errorf("doctor output should contain pinned Go version '1.99.99', got:\n%s", got)
	}
}

// TestDoctor_ClaudeVersionDeterministic verifies CLAUDE_CODE_VERSION is honored.
func TestDoctor_ClaudeVersionDeterministic(t *testing.T) {
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")
	t.Setenv("NO_COLOR", "1")

	got := captureDoctorCmd(t)
	if !strings.Contains(got, "test-claude-99") {
		t.Errorf("doctor output should contain pinned claude version 'test-claude-99', got:\n%s", got)
	}
}

// TestCheckGoRuntime_UsesGoVersionHelper verifies that checkGoRuntime uses goVersion()
// (which reads MOAI_GO_VERSION_OVERRIDE) instead of runtime.Version() directly.
// This is the lesson NEW (CI Go toolchain divergence) compliance check.
func TestCheckGoRuntime_UsesGoVersionHelper(t *testing.T) {
	const override = "1.PINNED.test" // no "go" prefix — goVersion() trims "go" from runtime.Version()
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", override)

	check := checkGoRuntime(false)
	if !strings.Contains(check.Message, override) {
		t.Errorf("checkGoRuntime should use MOAI_GO_VERSION_OVERRIDE=%q via goVersion() helper, got %q",
			override, check.Message)
	}
}

// TestCheckStatusToTUI verifies the status mapping is correct.
func TestCheckStatusToTUI(t *testing.T) {
	tests := []struct {
		status CheckStatus
		want   string
	}{
		{CheckOK, "ok"},
		{CheckWarn, "warn"},
		{CheckFail, "err"},
		{CheckStatus("unknown"), "info"},
	}
	for _, tt := range tests {
		got := checkStatusToTUI(tt.status)
		if got != tt.want {
			t.Errorf("checkStatusToTUI(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

// TestNewChecks_Names verifies new workspace checks return correct names.
func TestNewChecks_Names(t *testing.T) {
	tmpDir := t.TempDir()

	checks := []struct {
		name string
		fn   func() DiagnosticCheck
	}{
		{"Hooks Config", func() DiagnosticCheck { return checkHooksConfig(tmpDir, false) }},
		{"Slash Commands", func() DiagnosticCheck { return checkSlashCommands(tmpDir, false) }},
		{"Skills Allowlist", func() DiagnosticCheck { return checkSkillsAllowlist(tmpDir, false) }},
		{"MX Tag Config", func() DiagnosticCheck { return checkMXTagConfig(tmpDir, false) }},
		{"Worktree State", func() DiagnosticCheck { return checkWorktreeState(tmpDir, false) }},
		{"BODP Config", func() DiagnosticCheck { return checkBODPConfig(tmpDir, false) }},
		{"Telemetry Config", func() DiagnosticCheck { return checkTelemetryConfig(tmpDir, false) }},
		{"Glamour Cache", func() DiagnosticCheck { return checkGlamourCache(false) }},
		{"Claude Code", func() DiagnosticCheck { return checkClaudeCode(false) }},
		{"GitHub CLI", func() DiagnosticCheck { return checkGitHubCLI(false) }},
	}
	for _, tc := range checks {
		c := tc.fn()
		if c.Name != tc.name {
			t.Errorf("check Name = %q, want %q", c.Name, tc.name)
		}
		// status must be one of the valid values
		switch c.Status {
		case CheckOK, CheckWarn, CheckFail:
			// valid
		default:
			t.Errorf("check %q returned invalid status %q", tc.name, c.Status)
		}
	}
}

// TestGlamourCacheIsWarn verifies the D8 placeholder returns CheckWarn (not CheckFail).
func TestGlamourCacheIsWarn(t *testing.T) {
	c := checkGlamourCache(false)
	if c.Status != CheckWarn {
		t.Errorf("Glamour Cache D8 placeholder should return CheckWarn, got %q", c.Status)
	}
}

// TestRunGroupedChecks_Structure verifies the group structure returns expected number of groups.
func TestRunGroupedChecks_Structure(t *testing.T) {
	groups := runGroupedChecks(false, "")
	if len(groups) != 3 {
		t.Errorf("expected 3 check groups (System, MoAI-ADK, Workspace), got %d", len(groups))
	}
	names := []string{"System", "MoAI-ADK", "Workspace"}
	for i, g := range groups {
		if g.title != names[i] {
			t.Errorf("groups[%d].title = %q, want %q", i, g.title, names[i])
		}
	}
}
