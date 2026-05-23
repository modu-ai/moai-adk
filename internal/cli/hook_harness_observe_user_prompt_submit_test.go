// Package cli — full UserPromptSubmit hook handler test matrix (T-C3).
// REQ-HRN-OBS-007: verifies user_prompt events are recorded.
// REQ-HRN-OBS-014: PII strategy matrix (A=hash / B=preview / C=full / None=skip).
// REQ-HRN-FND-009: isHarnessLearningEnabled gate.
package cli

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

// TestRunHarnessObserveUserPromptSubmit_StrategyNone_NoOp verifies that no event
// is recorded when the strategy is "none", even when learning.enabled=true.
// REQ-HRN-OBS-014: Strategy None = full no-op (no recording).
func TestRunHarnessObserveUserPromptSubmit_StrategyNone_NoOp(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: none\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"prompt":"sensitive prompt text"}`, func() {
		if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveUserPromptSubmit 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("Strategy None 시 usage-log.jsonl이 생성되어서는 안 됨; stat err=%v", err)
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyA_HashLenLang verifies that
// Strategy A (default) records only prompt_hash, prompt_len, and prompt_lang
// without the raw prompt.
// REQ-HRN-OBS-014: Strategy A = SHA-256 hash + length + language (minimal PII).
func TestRunHarnessObserveUserPromptSubmit_StrategyA_HashLenLang(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: hash\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	// Use a pure-Korean prompt (detectPromptLang would return "en" if ASCII appeared first)
	const prompt = "테스트 전략 A 검증 프롬프트"
	payloadBytes, _ := json.Marshal(map[string]string{"prompt": prompt})
	cmd := &cobra.Command{}
	withStdin(t, string(payloadBytes), func() {
		if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveUserPromptSubmit 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl 미생성: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// Verify event_type
	if entry["event_type"] != "user_prompt" {
		t.Errorf("event_type: got=%v, want=%q", entry["event_type"], "user_prompt")
	}

	// Verify the SHA-256 hash value (first 16 hex chars only, REQ-HRN-OBS-006 / AC-HRN-OBS-004)
	h := sha256.Sum256([]byte(prompt))
	wantHash := fmt.Sprintf("%x", h)[:16]
	if entry["prompt_hash"] != wantHash {
		t.Errorf("prompt_hash: got=%v, want=%q", entry["prompt_hash"], wantHash)
	}

	// Verify length = byte count (multi-byte UTF-8 preserved, AC-HRN-OBS-004)
	wantLen := float64(len([]byte(prompt)))
	if entry["prompt_len"] != wantLen {
		t.Errorf("prompt_len: got=%v, want=%v", entry["prompt_len"], wantLen)
	}

	// Language detection = "ko" (Korean characters present)
	if entry["prompt_lang"] != "ko" {
		t.Errorf("prompt_lang: got=%v, want=%q", entry["prompt_lang"], "ko")
	}

	// prompt_content and prompt_preview must be absent (PII protection)
	for _, field := range []string{"prompt_content", "prompt_preview"} {
		if _, ok := entry[field]; ok {
			t.Errorf("Strategy A 시 %q가 엔트리에 존재해서는 안 됨 (PII 보호)", field)
		}
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview verifies that
// Strategy B records prompt_hash + prompt_len + prompt_lang + prompt_preview
// (first 64 bytes).
// REQ-HRN-OBS-013: prompt_preview = the first 64 bytes of the prompt.
// AC-HRN-OBS-008.a: prompt_preview is the first 64 bytes (full prompt if shorter).
func TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview(t *testing.T) {
	cases := []struct {
		name        string
		prompt      string
		wantPreview string
		desc        string
	}{
		{
			name:        "short_ascii_under_64",
			prompt:      "hello",
			wantPreview: "hello",
			desc:        "64바이트 미만 짧은 프롬프트: preview == 전체",
		},
		{
			name:        "long_ascii_over_64",
			prompt:      strings.Repeat("a", 100),
			wantPreview: strings.Repeat("a", 64),
			desc:        "64바이트 초과 ASCII 프롬프트: preview = 첫 64바이트",
		},
		{
			name: "multibyte_korean_over_64",
			// A Korean syllable like the test fixture equals 3 UTF-8 bytes. 22 chars = 66 bytes > 64.
			// 64-byte boundary: 21 chars = 63 bytes (complete rune boundary).
			prompt:      strings.Repeat("가", 22),
			wantPreview: strings.Repeat("가", 21),
			desc:        "멀티바이트(한글): 64바이트 경계가 룬 중간 → 룬 경계로 후퇴",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: preview\n")
			writeSystemYAMLHookOptIn(t, dir, true)
			t.Chdir(dir)

			payloadBytes, _ := json.Marshal(map[string]string{"prompt": tc.prompt})
			cmd := &cobra.Command{}
			withStdin(t, string(payloadBytes), func() {
				if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
					t.Fatalf("runHarnessObserveUserPromptSubmit 에러 반환: %v", err)
				}
			})

			logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("usage-log.jsonl 미생성: %v", err)
			}

			var entry map[string]any
			if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &entry); err != nil {
				t.Fatalf("JSONL 파싱 실패: %v", err)
			}

			// Verify prompt_preview presence
			preview, ok := entry["prompt_preview"]
			if !ok {
				t.Fatalf("[%s] Strategy B 시 prompt_preview 누락", tc.desc)
			}
			previewStr, _ := preview.(string)

			// Verify the 64-byte cap (AC-HRN-OBS-008.a)
			if len([]byte(previewStr)) > 64 {
				t.Errorf("[%s] prompt_preview 바이트 길이 초과: got=%d, want<=64",
					tc.desc, len([]byte(previewStr)))
			}

			// Verify UTF-8 validity
			if !utf8.ValidString(previewStr) {
				t.Errorf("[%s] prompt_preview가 유효한 UTF-8이 아님", tc.desc)
			}

			// Verify expected value
			if previewStr != tc.wantPreview {
				t.Errorf("[%s] prompt_preview: got=%q, want=%q", tc.desc, previewStr, tc.wantPreview)
			}

			// Short prompts: preview should equal the entire raw prompt
			if len([]byte(tc.prompt)) <= 64 && previewStr != tc.prompt {
				t.Errorf("[%s] 64바이트 이하 프롬프트: preview가 원문과 달라야 하지 않음; got=%q", tc.desc, previewStr)
			}

			// prompt_content must be absent (Strategy B does not include full text)
			if _, ok := entry["prompt_content"]; ok {
				t.Errorf("[%s] Strategy B 시 prompt_content이 존재해서는 안 됨", tc.desc)
			}

			// Verify prompt_hash, prompt_len, prompt_lang are present
			for _, field := range []string{"prompt_hash", "prompt_len", "prompt_lang"} {
				if _, ok := entry[field]; !ok {
					t.Errorf("[%s] Strategy B 시 %q 누락", tc.desc, field)
				}
			}
		})
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyC_Full verifies that Strategy C
// records the full raw prompt in prompt_content.
// REQ-HRN-OBS-014: Strategy C = Strategy A + full-text recording (explicit opt-in).
func TestRunHarnessObserveUserPromptSubmit_StrategyC_Full(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: full\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	const prompt = "SPEC-V3R4-HARNESS-002 전문 기록 테스트 - Strategy C"
	payloadBytes, _ := json.Marshal(map[string]string{"prompt": prompt})
	cmd := &cobra.Command{}
	withStdin(t, string(payloadBytes), func() {
		if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveUserPromptSubmit 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl 미생성: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// Verify prompt_content contains the full raw prompt (JSON key: "prompt_content")
	if entry["prompt_content"] != prompt {
		t.Errorf("prompt_content: got=%v, want=%q", entry["prompt_content"], prompt)
	}

	// prompt_hash, prompt_len, prompt_lang must all be present simultaneously
	for _, field := range []string{"prompt_hash", "prompt_len", "prompt_lang"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("Strategy C 시 %q 누락", field)
		}
	}
}

// TestRunHarnessObserveUserPromptSubmit_SpecIDExtraction verifies that a SPEC-ID
// inside a prompt is extracted as the subject.
// REQ-HRN-OBS-007: detect SPEC-ID patterns (SPEC-[A-Z][A-Z0-9]+-[0-9]+).
func TestRunHarnessObserveUserPromptSubmit_SpecIDExtraction(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	// specIDRegexp = SPEC-[A-Z][A-Z0-9]+-[0-9]+
	// SPEC-WF-002 matches: SPEC- + W + F + - + 002 (single-segment ID per the regex)
	const specID = "SPEC-WF-002"
	const prompt = "ultrathink. " + specID + " run 진입"
	payloadBytes, _ := json.Marshal(map[string]string{"prompt": prompt})
	cmd := &cobra.Command{}
	withStdin(t, string(payloadBytes), func() {
		if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveUserPromptSubmit 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl 미생성: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// subject is the SPEC-ID extracted from the prompt (REQ-HRN-OBS-007)
	if entry["subject"] != specID {
		t.Errorf("subject: got=%v, want=%q", entry["subject"], specID)
	}
}

// TestRunHarnessObserveUserPromptSubmit_FailOpenOnInvalidStrategy verifies that
// an invalid strategy value falls back to Strategy A (hash) and the event is recorded.
// REQ-HRN-OBS-014: fail-open to Strategy A.
func TestRunHarnessObserveUserPromptSubmit_FailOpenOnInvalidStrategy(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: INVALID_VALUE\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"prompt":"fallback test prompt"}`, func() {
		if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveUserPromptSubmit 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("잘못된 전략 시 fail-open으로 이벤트가 기록되어야 함: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// Strategy A fallback = prompt_hash present, prompt_content absent
	if _, ok := entry["prompt_hash"]; !ok {
		t.Errorf("잘못된 전략 폴백 시 prompt_hash 누락 (Strategy A 폴백 미작동)")
	}
	if _, ok := entry["prompt_content"]; ok {
		t.Errorf("잘못된 전략 폴백 시 prompt_content이 존재해서는 안 됨 (Strategy A 폴백)")
	}
}

// TestRunHarnessObserveUserPromptSubmit_LangHeuristic verifies that the
// language-detection result from detectPromptLang is reflected correctly in the
// JSONL entry (integration check).
func TestRunHarnessObserveUserPromptSubmit_LangHeuristic(t *testing.T) {
	cases := []struct {
		name     string
		prompt   string
		wantLang string
	}{
		{"korean", "안녕하세요 SPEC 테스트", "ko"},
		{"english", "hello world test prompt", "en"},
		{"japanese", "こんにちは世界", "ja"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
			writeSystemYAMLHookOptIn(t, dir, true)
			t.Chdir(dir)

			cmd := &cobra.Command{}
			withStdin(t, `{"prompt":"`+tc.prompt+`"}`, func() {
				if err := runHarnessObserveUserPromptSubmit(cmd, nil); err != nil {
					t.Fatalf("runHarnessObserveUserPromptSubmit 에러 반환: %v", err)
				}
			})

			logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("usage-log.jsonl 미생성: %v", err)
			}

			var entry map[string]any
			if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &entry); err != nil {
				t.Fatalf("JSONL 파싱 실패: %v", err)
			}

			if entry["prompt_lang"] != tc.wantLang {
				t.Errorf("prompt_lang: got=%v, want=%q", entry["prompt_lang"], tc.wantLang)
			}
		})
	}
}
