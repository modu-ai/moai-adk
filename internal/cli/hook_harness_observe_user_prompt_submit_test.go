// Package cli — UserPromptSubmit 훅 핸들러 전체 테스트 매트릭스 (T-C3).
// REQ-HRN-OBS-007: user_prompt 이벤트 기록 검증.
// REQ-HRN-OBS-014: PII 전략 매트릭스 (A=hash/B=preview/C=full/None=skip).
// REQ-HRN-FND-009: isHarnessLearningEnabled 게이트.
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

// TestRunHarnessObserveUserPromptSubmit_StrategyNone_NoOp은 전략이 "none"일 때
// learning.enabled=true이더라도 이벤트가 기록되지 않음을 검증한다.
// REQ-HRN-OBS-014: Strategy None = 완전 no-op (기록 없음).
func TestRunHarnessObserveUserPromptSubmit_StrategyNone_NoOp(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: none\n")
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

// TestRunHarnessObserveUserPromptSubmit_StrategyA_HashLenLang은 Strategy A (기본값)로
// prompt_hash, prompt_len, prompt_lang만 기록되고 원문은 없음을 검증한다.
// REQ-HRN-OBS-014: Strategy A = SHA-256 해시 + 길이 + 언어 (PII 최소화).
func TestRunHarnessObserveUserPromptSubmit_StrategyA_HashLenLang(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: hash\n")
	t.Chdir(dir)

	// 순수 한글 프롬프트 사용 (ASCII 문자가 앞에 오면 detectPromptLang이 "en"을 먼저 반환)
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

	// event_type 검증
	if entry["event_type"] != "user_prompt" {
		t.Errorf("event_type: got=%v, want=%q", entry["event_type"], "user_prompt")
	}

	// SHA-256 해시 값 검증 (앞 16 hex chars만, REQ-HRN-OBS-006 / AC-HRN-OBS-004)
	h := sha256.Sum256([]byte(prompt))
	wantHash := fmt.Sprintf("%x", h)[:16]
	if entry["prompt_hash"] != wantHash {
		t.Errorf("prompt_hash: got=%v, want=%q", entry["prompt_hash"], wantHash)
	}

	// 길이 = byte count 검증 (multi-byte UTF-8 보존, AC-HRN-OBS-004)
	wantLen := float64(len([]byte(prompt)))
	if entry["prompt_len"] != wantLen {
		t.Errorf("prompt_len: got=%v, want=%v", entry["prompt_len"], wantLen)
	}

	// 언어 감지 = "ko" (한글 포함)
	if entry["prompt_lang"] != "ko" {
		t.Errorf("prompt_lang: got=%v, want=%q", entry["prompt_lang"], "ko")
	}

	// prompt_content, prompt_preview 미존재 (PII 보호)
	for _, field := range []string{"prompt_content", "prompt_preview"} {
		if _, ok := entry[field]; ok {
			t.Errorf("Strategy A 시 %q가 엔트리에 존재해서는 안 됨 (PII 보호)", field)
		}
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview는 Strategy B로
// prompt_hash + prompt_len + prompt_lang + prompt_preview(앞 64바이트)가 기록됨을 검증한다.
// REQ-HRN-OBS-013: prompt_preview = 프롬프트의 첫 64바이트.
// AC-HRN-OBS-008.a: prompt_preview는 첫 64바이트 (프롬프트가 더 짧으면 전체).
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
			// 한글 '가' = 3바이트 UTF-8. 22자 = 66바이트 > 64.
			// 64바이트 경계: 21자 = 63바이트 (완전한 룬 경계).
			prompt:      strings.Repeat("가", 22),
			wantPreview: strings.Repeat("가", 21),
			desc:        "멀티바이트(한글): 64바이트 경계가 룬 중간 → 룬 경계로 후퇴",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: preview\n")
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

			// prompt_preview 존재 검증
			preview, ok := entry["prompt_preview"]
			if !ok {
				t.Fatalf("[%s] Strategy B 시 prompt_preview 누락", tc.desc)
			}
			previewStr, _ := preview.(string)

			// 바이트 상한 64 검증 (AC-HRN-OBS-008.a)
			if len([]byte(previewStr)) > 64 {
				t.Errorf("[%s] prompt_preview 바이트 길이 초과: got=%d, want<=64",
					tc.desc, len([]byte(previewStr)))
			}

			// UTF-8 유효성 검증
			if !utf8.ValidString(previewStr) {
				t.Errorf("[%s] prompt_preview가 유효한 UTF-8이 아님", tc.desc)
			}

			// 기댓값 일치 검증
			if previewStr != tc.wantPreview {
				t.Errorf("[%s] prompt_preview: got=%q, want=%q", tc.desc, previewStr, tc.wantPreview)
			}

			// 짧은 프롬프트: preview == 원문 전체
			if len([]byte(tc.prompt)) <= 64 && previewStr != tc.prompt {
				t.Errorf("[%s] 64바이트 이하 프롬프트: preview가 원문과 달라야 하지 않음; got=%q", tc.desc, previewStr)
			}

			// prompt_content 미존재 (Strategy B는 full text 미포함)
			if _, ok := entry["prompt_content"]; ok {
				t.Errorf("[%s] Strategy B 시 prompt_content이 존재해서는 안 됨", tc.desc)
			}

			// prompt_hash, prompt_len, prompt_lang 존재 검증
			for _, field := range []string{"prompt_hash", "prompt_len", "prompt_lang"} {
				if _, ok := entry[field]; !ok {
					t.Errorf("[%s] Strategy B 시 %q 누락", tc.desc, field)
				}
			}
		})
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyC_Full은 Strategy C로
// prompt_content에 원문 전체가 기록됨을 검증한다.
// REQ-HRN-OBS-014: Strategy C = Strategy A + 전문 기록 (명시적 opt-in).
func TestRunHarnessObserveUserPromptSubmit_StrategyC_Full(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: full\n")
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

	// prompt_content에 원문 전체 검증 (JSON키: "prompt_content")
	if entry["prompt_content"] != prompt {
		t.Errorf("prompt_content: got=%v, want=%q", entry["prompt_content"], prompt)
	}

	// prompt_hash, prompt_len, prompt_lang 동시 존재
	for _, field := range []string{"prompt_hash", "prompt_len", "prompt_lang"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("Strategy C 시 %q 누락", field)
		}
	}
}

// TestRunHarnessObserveUserPromptSubmit_SpecIDExtraction은 프롬프트 내 SPEC-ID를
// subject로 추출하는지 검증한다.
// REQ-HRN-OBS-007: SPEC-ID 패턴 (SPEC-[A-Z][A-Z0-9]+-[0-9]+) 탐지.
func TestRunHarnessObserveUserPromptSubmit_SpecIDExtraction(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	// specIDRegexp = SPEC-[A-Z][A-Z0-9]+-[0-9]+
	// SPEC-WF-002 매칭: SPEC- + W + F + - + 002 (정규식 기준 단일 세그먼트 ID)
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

	// subject는 프롬프트에서 추출한 SPEC-ID (REQ-HRN-OBS-007)
	if entry["subject"] != specID {
		t.Errorf("subject: got=%v, want=%q", entry["subject"], specID)
	}
}

// TestRunHarnessObserveUserPromptSubmit_FailOpenOnInvalidStrategy는 잘못된 전략 값이
// Strategy A(hash)로 폴백되어 이벤트가 기록됨을 검증한다.
// REQ-HRN-OBS-014: fail-open to Strategy A.
func TestRunHarnessObserveUserPromptSubmit_FailOpenOnInvalidStrategy(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: INVALID_VALUE\n")
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

	// Strategy A로 폴백 = prompt_hash 존재, prompt_content 없음
	if _, ok := entry["prompt_hash"]; !ok {
		t.Errorf("잘못된 전략 폴백 시 prompt_hash 누락 (Strategy A 폴백 미작동)")
	}
	if _, ok := entry["prompt_content"]; ok {
		t.Errorf("잘못된 전략 폴백 시 prompt_content이 존재해서는 안 됨 (Strategy A 폴백)")
	}
}

// TestRunHarnessObserveUserPromptSubmit_LangHeuristic은 detectPromptLang 헬퍼의
// 언어 감지 결과가 JSONL 엔트리에 올바르게 반영되는지 검증한다 (통합 검증).
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
