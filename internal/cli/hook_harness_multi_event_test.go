// Package cli — multi-event harness observer tests.
// T-A3: Stop 훅 핸들러 (REQ-HRN-OBS-003, REQ-HRN-OBS-008)
// T-A4: SubagentStop 훅 핸들러 (REQ-HRN-OBS-005, REQ-HRN-OBS-008)
// T-A5: UserPromptSubmit 훅 핸들러 (REQ-HRN-OBS-007, REQ-HRN-OBS-014)
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
// T-A3: runHarnessObserveStop 테스트
// ─────────────────────────────────────────────

// TestRunHarnessObserveStop_NoOpWhenDisabled는 learning.enabled=false일 때
// Stop 핸들러가 usage-log.jsonl을 생성/수정하지 않음을 검증한다.
// REQ-HRN-FND-009: isHarnessLearningEnabled 게이트 재사용.
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

// TestRunHarnessObserveStop_RecordsBaseline는 learning.enabled=true일 때
// Stop 핸들러가 session_stop 이벤트를 기록하는지 검증한다.
// REQ-HRN-OBS-003, REQ-HRN-FND-010.
func TestRunHarnessObserveStop_RecordsBaseline(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
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

	// 기존 4-필드 스키마 보존 (REQ-HRN-FND-010)
	for _, field := range []string{"timestamp", "event_type", "subject", "schema_version"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("required field %q missing from Stop event entry", field)
		}
	}
}

// ─────────────────────────────────────────────
// T-A4: runHarnessObserveSubagentStop 테스트
// ─────────────────────────────────────────────

// TestRunHarnessObserveSubagentStop_NoOpWhenDisabled는 learning.enabled=false일 때
// SubagentStop 핸들러가 usage-log.jsonl을 생성하지 않음을 검증한다.
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

// TestRunHarnessObserveSubagentStop_RecordsSubagentStop는 SubagentStop 핸들러가
// subagent_stop 이벤트를 올바르게 기록하는지 검증한다.
// REQ-HRN-OBS-005, REQ-HRN-FND-010.
func TestRunHarnessObserveSubagentStop_RecordsSubagentStop(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
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
// T-A5: runHarnessObserveUserPromptSubmit 테스트
// ─────────────────────────────────────────────

// TestRunHarnessObserveUserPromptSubmit_NoOpWhenDisabled는 learning.enabled=false일 때
// UserPromptSubmit 핸들러가 usage-log.jsonl을 생성하지 않음을 검증한다.
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

// TestRunHarnessObserveUserPromptSubmit_StrategyA_DefaultHash는 Strategy A(기본값)로
// 프롬프트가 SHA-256 해시 + 길이만 기록되는지 검증한다.
// REQ-HRN-OBS-007, REQ-HRN-OBS-014: 기본값은 Strategy A (PII 최소화).
func TestRunHarnessObserveUserPromptSubmit_StrategyA_DefaultHash(t *testing.T) {
	dir := t.TempDir()
	// Strategy A는 user_prompt_content 설정이 없거나 "hash"일 때
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
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

	// Strategy A: prompt_hash 존재, prompt_full 없음
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

// TestResolveUserPromptStrategy_FailOpen은 잘못된 strategy 설정 시
// Strategy A로 폴백(fail-open)되는지 검증한다.
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

// TestDetectPromptLang는 Unicode 블록 기반 언어 감지 헬퍼를 검증한다.
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
