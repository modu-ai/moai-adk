// Package cli — multi-event harness observer tests.
// T-A3: Stop hook handler (REQ-HRN-OBS-003, REQ-HRN-OBS-008)
// T-A4: SubagentStop hook handler (REQ-HRN-OBS-005, REQ-HRN-OBS-008)
// T-A5: UserPromptSubmit hook handler (REQ-HRN-OBS-007, REQ-HRN-OBS-014)
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────
// T-A3: runHarnessObserveStop tests
// ─────────────────────────────────────────────

// TestRunHarnessObserveStop_NoOpWhenDisabled verifies that the Stop handler
// neither creates nor modifies usage-log.jsonl when learning.enabled=false.
// REQ-HRN-FND-009: reuses the isHarnessLearningEnabled gate.
func TestRunHarnessObserveStop_NoOpWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	cmd := &cobra.Command{}
	// T-A3 spec: nested stdin JSON — last_assistant_message + session.id
	withStdin(t, `{"last_assistant_message":"","session":{"id":"sess-abc"},"hook_event_name":"Stop"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop returned error: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("usage-log.jsonl must not exist when learning.enabled is false; stat err=%v", err)
	}
}

// TestRunHarnessObserveStop_RecordsBaseline verifies that the Stop handler
// records a session_stop event when learning.enabled=true.
// REQ-HRN-OBS-003, REQ-HRN-FND-010.
func TestRunHarnessObserveStop_RecordsBaseline(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	// T-A3 spec: nested stdin JSON — last_assistant_message + session.id
	withStdin(t, `{"last_assistant_message":"test msg","session":{"id":"sess-xyz"},"hook_event_name":"Stop"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop error: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl was not created: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSONL entry, got %d", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("entry is not valid JSON: %v", err)
	}

	// event_type must be session_stop
	if entry["event_type"] != "session_stop" {
		t.Errorf("event_type = %v, want \"session_stop\"", entry["event_type"])
	}

	// Preserve the existing 4-field schema (REQ-HRN-FND-010)
	for _, field := range []string{"timestamp", "event_type", "subject", "schema_version"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("required field %q missing from Stop event entry", field)
		}
	}
}

// ─────────────────────────────────────────────
// T-A4: runHarnessObserveSubagentStop tests
// ─────────────────────────────────────────────

// TestRunHarnessObserveSubagentStop_NoOpWhenDisabled verifies that the SubagentStop
// handler does not create usage-log.jsonl when learning.enabled=false.
func TestRunHarnessObserveSubagentStop_NoOpWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	// T-A4 spec: camelCase agentName + nested session.id
	cmd := &cobra.Command{}
	withStdin(t, `{"agentName":"expert-backend","hook_event_name":"SubagentStop"}`, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop returned error: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("usage-log.jsonl must not exist when learning.enabled is false")
	}
}

// TestRunHarnessObserveSubagentStop_RecordsSubagentStop verifies that the SubagentStop
// handler records a subagent_stop event correctly.
// REQ-HRN-OBS-005, REQ-HRN-FND-010.
func TestRunHarnessObserveSubagentStop_RecordsSubagentStop(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	// T-A4 spec: camelCase agentName + nested session.id
	cmd := &cobra.Command{}
	withStdin(t, `{"agentName":"expert-backend","hook_event_name":"SubagentStop"}`, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop error: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl was not created: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSONL entry, got %d", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("entry is not valid JSON: %v", err)
	}

	// event_type must be subagent_stop
	if entry["event_type"] != "subagent_stop" {
		t.Errorf("event_type = %v, want \"subagent_stop\"", entry["event_type"])
	}

	// subject must reflect agent name from stdin
	if entry["subject"] != "expert-backend" {
		t.Errorf("subject = %v, want \"expert-backend\"", entry["subject"])
	}
}

// ─────────────────────────────────────────────
// T-A5: runHarnessObserveUserPromptSubmit tests
// ─────────────────────────────────────────────

// TestRunHarnessObserveUserPromptSubmit_NoOpWhenDisabled verifies that the
// UserPromptSubmit handler does not create usage-log.jsonl when learning.enabled=false.
func TestRunHarnessObserveUserPromptSubmit_NoOpWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"prompt":"hello world"}`, func() {
		if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveUserPromptSubmit returned error: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("usage-log.jsonl must not exist when learning.enabled is false")
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyA_DefaultHash verifies that
// Strategy A (default) records only the prompt's SHA-256 hash and length.
// REQ-HRN-OBS-007, REQ-HRN-OBS-014: the default is Strategy A (minimal PII).
func TestRunHarnessObserveUserPromptSubmit_StrategyA_DefaultHash(t *testing.T) {
	dir := t.TempDir()
	// Strategy A applies when user_prompt_content is unset or set to "hash"
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"prompt":"SPEC-V3R4-HARNESS-002 계획을 세워줘"}`, func() {
		if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveUserPromptSubmit error: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl was not created: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSONL entry, got %d", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("entry is not valid JSON: %v", err)
	}

	// event_type must be user_prompt
	if entry["event_type"] != "user_prompt" {
		t.Errorf("event_type = %v, want \"user_prompt\"", entry["event_type"])
	}

	// Strategy A: prompt_hash present, prompt_full absent
	if _, ok := entry["prompt_hash"]; !ok {
		t.Errorf("prompt_hash missing from Strategy A entry")
	}
	if _, ok := entry["prompt_full"]; ok {
		t.Errorf("prompt_full must NOT appear in Strategy A entry (PII protection)")
	}
	if _, ok := entry["prompt_preview"]; ok {
		t.Errorf("prompt_preview must NOT appear in Strategy A entry (PII protection)")
	}

	// prompt_len must be set
	if _, ok := entry["prompt_len"]; !ok {
		t.Errorf("prompt_len missing from user_prompt entry")
	}
}

// TestResolveUserPromptStrategy_FailOpen verifies fall-back to Strategy A
// (fail-open) when the strategy setting is invalid.
// REQ-HRN-OBS-014: fail-open to Strategy A.
func TestResolveUserPromptStrategy_FailOpen(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		wantEnum UserPromptStrategy
	}{
		{"empty", "", UserPromptStrategyHash},
		{"hash_explicit", "hash", UserPromptStrategyHash},
		{"preview_explicit", "preview", UserPromptStrategyPreview},
		{"full_explicit", "full", UserPromptStrategyFull},
		{"none_explicit", "none", UserPromptStrategyNone},
		{"invalid_fallback", "banana", UserPromptStrategyHash},
		{"invalid_case", "HASH", UserPromptStrategyHash},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveUserPromptStrategy(tc.input)
			if got != tc.wantEnum {
				t.Errorf("resolveUserPromptStrategy(%q) = %v, want %v", tc.input, got, tc.wantEnum)
			}
		})
	}
}

// TestDetectPromptLang verifies the Unicode-block-based language-detection helper.
func TestDetectPromptLang(t *testing.T) {
	t.Parallel()

	cases := []struct {
		prompt string
		want   string
	}{
		{"안녕하세요", "ko"},       // Hangul
		{"こんにちは", "ja"},       // Hiragana
		{"カタカナ", "ja"},        // Katakana
		{"你好世界", "zh"},        // CJK
		{"hello world", "en"}, // ASCII
		{"", ""},              // empty
		{"12345!@#$%", ""},    // no letters
	}

	for _, tc := range cases {
		t.Run(tc.prompt, func(t *testing.T) {
			t.Parallel()
			got := detectPromptLang(tc.prompt)
			if got != tc.want {
				t.Errorf("detectPromptLang(%q) = %q, want %q", tc.prompt, got, tc.want)
			}
		})
	}
}
