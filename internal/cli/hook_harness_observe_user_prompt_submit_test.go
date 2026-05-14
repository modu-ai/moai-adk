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

	// SHA-256 해시 값 검증
	h := sha256.Sum256([]byte(prompt))
	wantHash := fmt.Sprintf("%x", h)
	if entry["prompt_hash"] != wantHash {
		t.Errorf("prompt_hash: got=%v, want=%q", entry["prompt_hash"], wantHash)
	}

	// 길이 = rune count 검증
	wantLen := float64(utf8.RuneCountInString(prompt))
	if entry["prompt_len"] != wantLen {
		t.Errorf("prompt_len: got=%v, want=%v", entry["prompt_len"], wantLen)
	}

	// 언어 감지 = "ko" (한글 포함)
	if entry["prompt_lang"] != "ko" {
		t.Errorf("prompt_lang: got=%v, want=%q", entry["prompt_lang"], "ko")
	}

	// prompt_full, prompt_preview 미존재 (PII 보호)
	for _, field := range []string{"prompt_full", "prompt_preview"} {
		if _, ok := entry[field]; ok {
			t.Errorf("Strategy A 시 %q가 엔트리에 존재해서는 안 됨 (PII 보호)", field)
		}
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview는 Strategy B로
// prompt_hash + prompt_len + prompt_lang + prompt_preview(앞 200자)가 기록됨을 검증한다.
// REQ-HRN-OBS-014: Strategy B = Strategy A 내용 + 앞 200자 미리보기.
func TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n  user_prompt_content: preview\n")
	t.Chdir(dir)

	// 정확히 250자 프롬프트 생성 (미리보기는 앞 200자만)
	prompt := strings.Repeat("가", 250) // 한글 250자
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

	// prompt_preview 존재 및 200자 제한 검증
	preview, ok := entry["prompt_preview"]
	if !ok {
		t.Fatalf("Strategy B 시 prompt_preview 누락")
	}
	previewStr, _ := preview.(string)
	previewRunes := []rune(previewStr)
	if len(previewRunes) != 200 {
		t.Errorf("prompt_preview 길이: got=%d, want=200 runes", len(previewRunes))
	}
	if previewStr != strings.Repeat("가", 200) {
		t.Errorf("prompt_preview 내용이 앞 200자와 다름")
	}

	// prompt_full 미존재 (Strategy B는 full text 미포함)
	if _, ok := entry["prompt_full"]; ok {
		t.Errorf("Strategy B 시 prompt_full이 존재해서는 안 됨")
	}

	// prompt_hash, prompt_len, prompt_lang 존재 검증
	for _, field := range []string{"prompt_hash", "prompt_len", "prompt_lang"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("Strategy B 시 %q 누락", field)
		}
	}
}

// TestRunHarnessObserveUserPromptSubmit_StrategyC_Full은 Strategy C로
// prompt_full에 원문 전체가 기록됨을 검증한다.
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

	// prompt_full에 원문 전체 검증 (JSON키: "prompt_full")
	if entry["prompt_full"] != prompt {
		t.Errorf("prompt_full: got=%v, want=%q", entry["prompt_full"], prompt)
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

	// Strategy A로 폴백 = prompt_hash 존재, prompt_full 없음
	if _, ok := entry["prompt_hash"]; !ok {
		t.Errorf("잘못된 전략 폴백 시 prompt_hash 누락 (Strategy A 폴백 미작동)")
	}
	if _, ok := entry["prompt_full"]; ok {
		t.Errorf("잘못된 전략 폴백 시 prompt_full이 존재해서는 안 됨 (Strategy A 폴백)")
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
		tc := tc
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
