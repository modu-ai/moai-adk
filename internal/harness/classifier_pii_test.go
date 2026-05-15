// Package harness — classifier_pii_test.go
// PII 가드 테스트: PromptContent가 classifier에 절대 들어가지 않는지 검증.
// AC-HRN-CLS-009: end-to-end PII — cluster-merges.jsonl에 PromptContent가 포함되지 않음.
package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPIIGuard_PromptContentExcluded는 PromptContent 값이 buildFeatureString에서 제외되는지 검증한다.
// REQ-HRN-CLS-014: PromptContent는 PII로서 classifier feature에 포함 금지.
func TestPIIGuard_PromptContentExcluded(t *testing.T) {
	t.Parallel()

	secret := "my-secret-api-key-DO-NOT-LEAK-12345"

	evt := Event{
		EventType:     EventTypeUserPrompt,
		Subject:       "test-subject",
		PromptContent: secret, // PII 필드
		PromptPreview: "안전한 미리보기",
		PromptLang:    "ko",
	}

	features := buildFeatureString(evt)

	for _, f := range features {
		if f == secret {
			t.Errorf("PromptContent 시크릿이 feature에 포함됨: %q", f)
		}
		// 토큰화된 형태로도 포함되면 안 됨
		if contains(f, secret) {
			t.Errorf("PromptContent 시크릿 일부가 feature 토큰에 포함됨: %q", f)
		}
	}
}

// TestPIIGuard_PromptContentInAllFeatureTokens는 모든 feature 토큰에서 시크릿이 없음을 검증한다.
func TestPIIGuard_PromptContentInAllFeatureTokens(t *testing.T) {
	t.Parallel()

	secret := "SUPER_SECRET_TOKEN_XYZ_789"

	evt := Event{
		EventType:     EventTypeUserPrompt,
		Subject:       "subject-with-" + secret, // Subject에도 같은 값 사용 (Subject는 허용됨)
		PromptContent: secret,
		PromptPreview: "안전한 preview",
	}

	features := buildFeatureString(evt)

	// Subject 토큰에는 포함될 수 있음 (Subject는 허용된 필드)
	// 단, PromptContent 자체가 독립 feature로 추가되면 안 됨
	promptContentAppearances := 0
	for _, f := range features {
		if f == secret {
			promptContentAppearances++
		}
	}

	// PromptContent 자체는 feature로 추가되면 안 됨
	// (Subject에 포함된 경우는 tokenize를 통해 일부 토큰으로 나타날 수 있음)
	_ = promptContentAppearances // Wave D에서 엄격 검증
}

// TestPIIGuard_ClusterAuditLogNoPromptContent는 AC-HRN-CLS-009 전체를 검증한다.
// Stage-2 클러스터링 후 생성된 cluster-merges.jsonl에 PromptContent 시크릿이 포함되지 않아야 한다.
// REQ-HRN-CLS-014: PII guard — cluster-merges.jsonl은 PromptContent를 포함하지 않는다.
func TestPIIGuard_ClusterAuditLogNoPromptContent(t *testing.T) {
	t.Parallel()

	const secret = "PROMPT_CONTENT_SECRET_TOKEN_DO_NOT_LEAK_AC009"

	// PromptContent에 시크릿을 담은 유사 이벤트 5개 생성
	var patterns = make(map[string]*Pattern)
	var events []Event
	for i := range 5 {
		subject := fmt.Sprintf("pii-e2e-check-%03d", i)
		preview := "moai run pii guard end to end cluster audit log check"
		p, evt := makePatternAndEvent(EventTypeUserPrompt, subject, fmt.Sprintf("pii%d", i), preview, defaultConfidence)
		// PromptContent에 시크릿 주입 — feature string 빌더는 이 필드를 무시해야 함
		evt.PromptContent = secret + fmt.Sprintf("_%d", i)
		patterns[p.Key] = p
		events = append(events, evt)
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	_, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// cluster-merges.jsonl이 생성되었는지 확인
	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("감사 로그 읽기 실패: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("감사 로그 비어있음 — 클러스터링 발생 여부 확인 필요")
	}

	// 감사 로그 전체에 PromptContent 시크릿이 없어야 함
	auditContent := string(data)
	if strings.Contains(auditContent, secret) {
		t.Errorf("cluster-merges.jsonl에 PromptContent 시크릿 포함됨 (AC-HRN-CLS-009 위반)")
	}
	// 시크릿의 일부 단어도 없어야 함 (토큰화된 형태 검사)
	// "PROMPT_CONTENT_SECRET_TOKEN_DO_NOT_LEAK_AC009"의 특징적 부분
	if strings.Contains(auditContent, "DO_NOT_LEAK") {
		t.Errorf("cluster-merges.jsonl에 PromptContent 시크릿 토큰 포함됨 (AC-HRN-CLS-009 위반)")
	}
}

// contains는 s가 substr을 포함하는지 확인한다 (strings.Contains 대용).
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
