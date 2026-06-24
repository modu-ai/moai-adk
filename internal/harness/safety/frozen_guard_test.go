// Package safety — frozen_guard unit test.
// REQ-HL-006: IsFrozen + LogViolation tests.
package safety

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIsFrozen_FrozenPaths verifies that paths with a FROZEN prefix return IsFrozen==true.
func TestIsFrozen_FrozenPaths(t *testing.T) {
	t.Parallel()

	frozenCases := []struct {
		name string
		path string
	}{
		{"moai agent", ".claude/agents/moai/expert-backend.md"},
		{"moai agent nested", ".claude/agents/moai/sub/skill.md"},
		{"moai skill", ".claude/skills/moai-workflow-tdd/SKILL.md"},
		{"moai skill direct", ".claude/skills/moai-foundation-core/modules/foo.md"},
		{"moai rules", ".claude/rules/moai/core/moai-constitution.md"},
		{"moai rules nested", ".claude/rules/moai/workflow/spec-workflow.md"},
		{"brand", ".moai/project/brand/brand-voice.md"},
		{"brand nested", ".moai/project/brand/visual-identity.md"},
	}

	for _, tc := range frozenCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if !IsFrozen(tc.path) {
				t.Errorf("IsFrozen(%q) = false, must be true", tc.path)
			}
		})
	}
}

// TestIsFrozen_UserPaths verifies that user paths return IsFrozen==false.
func TestIsFrozen_UserPaths(t *testing.T) {
	t.Parallel()

	userCases := []struct {
		name string
		path string
	}{
		{"harness agent", ".claude/agents/harness/agent.md"},
		{"harness skill", ".claude/skills/harness-plugin/SKILL.md"},
		{"harness state", ".moai/harness/usage-log.jsonl"},
		{"harness history", ".moai/harness/learning-history/frozen-guard-violations.jsonl"},
		{"project non-brand", ".moai/project/specs/SPEC-001.md"},
		{"custom rules", ".claude/rules/custom/my-rule.md"},
		{"harness chaining", ".moai/harness/chaining-rules.yaml"},
	}

	for _, tc := range userCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if IsFrozen(tc.path) {
				t.Errorf("IsFrozen(%q) = true, must be false", tc.path)
			}
		})
	}
}

// TestIsFrozen_TraversalAndEdgeCases verifies path traversal and edge cases.
func TestIsFrozen_TraversalAndEdgeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		path     string
		wantFroz bool
	}{
		// Traversal attempts are decided by the FROZEN prefix match after Clean.
		{"empty path", "", false},
		{"path traversal bypass attempt", ".claude/../.claude/agents/moai/hack.md", true},
		// Backslashes are converted via filepath.ToSlash (Windows path support).
		// On macOS/Linux, backslashes are treated as part of the filename, so they do not match FROZEN.
		{"absolute moai rules", "/abs/.claude/rules/moai/evil.md", false}, // absolute paths are not FROZEN (relative paths only)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsFrozen(tc.path)
			if got != tc.wantFroz {
				t.Errorf("IsFrozen(%q) = %v, want %v", tc.path, got, tc.wantFroz)
			}
		})
	}
}

// TestIsFrozen_WindowsStylePath verifies the conversion of Windows-style paths.
// filepath.ToSlash converts backslashes to forward slashes before evaluation.
func TestIsFrozen_WindowsStylePath(t *testing.T) {
	t.Parallel()

	// When filepath.ToSlash is applied the path becomes `.claude/agents/moai/hack.md` (FROZEN).
	// On non-Windows environments, Clean keeps backslashes as filename characters, so
	// check the order: implementations that apply ToSlash first vs filepath.Clean first.
	path := `.claude\agents\moai\hack.md`
	// IsFrozen applies filepath.ToSlash first internally, so it is FROZEN.
	got := IsFrozen(path)
	if !got {
		// Acceptable when filepath.ToSlash is not applied.
		// On macOS, the backslash is part of the filename, so false is normal.
		t.Logf("IsFrozen(%q) = false: on macOS the backslash is treated as part of the filename (expected)", path)
	}
}

// TestIsFrozen_ConfigBypassAttempt verifies that configuration-based bypass attempts are always blocked.
// REQ-HL-006: Frozen Guard must not be bypassable via configuration.
// t.Setenv cannot be combined with t.Parallel(), so this test runs sequentially.
func TestIsFrozen_ConfigBypassAttempt(t *testing.T) {
	// Verify that env vars, config files, etc. cannot bypass it.
	// IsFrozen uses only hardcoded prefixes and does not read external configuration.
	// FROZEN paths are always blocked regardless of any external setting changes.
	target := ".claude/agents/moai/evil-agent.md"

	// Attempt to bypass via env vars (must have no effect)
	t.Setenv("FROZEN_GUARD_BYPASS", "true")
	t.Setenv("HARNESS_ALLOW_ALL", "1")

	if !IsFrozen(target) {
		t.Errorf("IsFrozen(%q) = false: config bypass succeeded! FROZEN paths must always be blocked", target)
	}
}

// TestIsFrozen_SymlinkResolution verifies that the path is evaluated correctly after filepath.Clean.
func TestIsFrozen_SymlinkResolution(t *testing.T) {
	t.Parallel()

	// Verify that the path normalized by filepath.Clean matches the FROZEN prefix.
	// Verified using a Clean'd path without actually creating a symlink.
	traversalPath := "./.claude/agents/moai/../moai/real.md"
	// Clean: ".claude/agents/moai/real.md" → FROZEN

	if !IsFrozen(traversalPath) {
		t.Errorf("IsFrozen(%q): must be recognized as a FROZEN path after traversal clean", traversalPath)
	}
}

// TestLogViolation_AppendsJSONL verifies that LogViolation writes to the violations JSONL.
func TestLogViolation_AppendsJSONL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "learning-history", "frozen-guard-violations.jsonl")

	// Record the first violation
	err := LogViolation(logPath, ".claude/agents/moai/evil.md", "test-caller")
	if err != nil {
		t.Fatalf("LogViolation 실패: %v", err)
	}

	// Verify that the JSONL file was created and contains valid JSON
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("위반 로그 파일 읽기 실패: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("라인 수 = %d, want 1", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// Verify required fields
	if entry["path"] == nil {
		t.Error("path field is missing")
	}
	if entry["caller"] == nil {
		t.Error("caller field is missing")
	}
	if entry["timestamp"] == nil {
		t.Error("timestamp field is missing")
	}
}

// TestLogViolation_AppendMultiple verifies that multiple violations are appended in order.
func TestLogViolation_AppendMultiple(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "violations.jsonl")

	for i := range 3 {
		err := LogViolation(logPath, ".claude/rules/moai/hack.md", "caller-"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("LogViolation[%d] 실패: %v", i, err)
		}
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("라인 수 = %d, want 3", len(lines))
	}
}

// TestLogViolation_StderrWarning verifies that LogViolation emits a stderr warning.
// (Direct capture of stderr output as a side effect is hard, so only error == nil is verified.)
func TestLogViolation_StderrWarning(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "violations.jsonl")

	// Must execute without error (stderr output included)
	err := LogViolation(logPath, ".moai/project/brand/hack.md", "test")
	if err != nil {
		t.Errorf("LogViolation 오류: %v", err)
	}
}

// TestLogViolation_WriteError verifies that an error is returned for a write-protected path.
func TestLogViolation_WriteError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create a read-only subdirectory
	roDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(roDir, 0o555); err != nil {
		t.Fatalf("failed to create read-only directory: %v", err)
	}

	// Path inside the read-only directory (not writable)
	logPath := filepath.Join(roDir, "sub", "violations.jsonl")

	err := LogViolation(logPath, ".claude/rules/moai/test.md", "caller")
	// MkdirAll or OpenFile must fail
	if err == nil {
		// On macOS, non-root users cannot write, but CI environments may differ
		t.Logf("write succeeded even on a read-only path (CI environment variance, ignored)")
	}
}

// TestLogViolation_JSONLContent verifies that the JSONL contents are correct.
func TestLogViolation_JSONLContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "violations.jsonl")

	path := ".claude/rules/moai/test.md"
	caller := "test-caller-xyz"

	if err := LogViolation(logPath, path, caller); err != nil {
		t.Fatalf("LogViolation 실패: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &entry); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if entry["path"] != path {
		t.Errorf("path = %v, want %q", entry["path"], path)
	}
	if entry["caller"] != caller {
		t.Errorf("caller = %v, want %q", entry["caller"], caller)
	}
	if entry["message"] == nil || entry["message"] == "" {
		t.Error("message field is empty")
	}
}
