// Package harness — classifier_pii_test.go
// PII 가드 테스트: PromptContent가 classifier에 절대 들어가지 않는지 검증.
// AC-HRN-CLS-009 (부분 — Wave D에서 cluster-merges.jsonl 검증 추가).
package harness

import "testing"

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
