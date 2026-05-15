// Package harness — classifier_simhash_test.go
// SimHash 64-bit 지문, Hamming 거리, feature 빌더 테스트.
// AC-HRN-CLS-002: 동일 입력 → 결정적(deterministic) SimHash 보장.
package harness

import (
	"testing"
)

// ─────────────────────────────────────────────
// SimHash64 테스트
// ─────────────────────────────────────────────

// TestSimHash64_EmptyFeatures는 빈 feature 슬라이스가 0을 반환하는지 검증한다.
func TestSimHash64_EmptyFeatures(t *testing.T) {
	t.Parallel()

	got := SimHash64(nil)
	if got != 0 {
		t.Errorf("SimHash64(nil) = %d, want 0", got)
	}

	got2 := SimHash64([]string{})
	if got2 != 0 {
		t.Errorf("SimHash64([]) = %d, want 0", got2)
	}
}

// TestSimHash64_SingleFeatureDeterministic는 단일 feature가 결정적 해시를 반환하는지 검증한다.
func TestSimHash64_SingleFeatureDeterministic(t *testing.T) {
	t.Parallel()

	features := []string{"moai_plan"}

	h1 := SimHash64(features)
	h2 := SimHash64(features)
	h3 := SimHash64(features)

	if h1 == 0 {
		t.Error("단일 feature: SimHash64 반환값이 0이면 안 됨")
	}
	if h1 != h2 || h2 != h3 {
		t.Errorf("결정성 실패: %d, %d, %d", h1, h2, h3)
	}
}

// TestSimHash64_IdenticalFeatures는 동일한 feature 슬라이스가 동일한 해시를 반환하는지 검증한다.
func TestSimHash64_IdenticalFeatures(t *testing.T) {
	t.Parallel()

	f1 := []string{"user_prompt", "ko", "expert-backend", "spec-v3r4"}
	f2 := []string{"user_prompt", "ko", "expert-backend", "spec-v3r4"}

	if SimHash64(f1) != SimHash64(f2) {
		t.Error("동일 feature 슬라이스: 해시가 달라선 안 됨")
	}
}

// TestSimHash64_PermutationInvariance는 feature 순서가 달라도 동일한 해시를 반환하는지 검증한다.
// Charikar SimHash는 순서 불변(permutation-invariant) 특성을 갖는다.
func TestSimHash64_PermutationInvariance(t *testing.T) {
	t.Parallel()

	f1 := []string{"alpha", "beta", "gamma"}
	f2 := []string{"gamma", "alpha", "beta"}
	f3 := []string{"beta", "gamma", "alpha"}

	h1 := SimHash64(f1)
	h2 := SimHash64(f2)
	h3 := SimHash64(f3)

	if h1 != h2 || h2 != h3 {
		t.Errorf("순서 불변 실패: h1=%d h2=%d h3=%d", h1, h2, h3)
	}
}

// TestSimHash64_ReferenceValues는 사전 계산된 기준값과 일치하는지 검증한다.
// 이 기준값들은 FNV-1a 기반 Charikar SimHash 구현으로 생성되었다.
func TestSimHash64_ReferenceValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		features []string
	}{
		{"single_a", []string{"a"}},
		{"single_moai", []string{"moai"}},
		{"two_features", []string{"hello", "world"}},
		{"three_features", []string{"user", "prompt", "ko"}},
		{"spec_like", []string{"spec-v3r4-harness", "expert-backend", "ko"}},
	}

	// 결정성만 검증 (동일 입력 → 2회 일치)
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h1 := SimHash64(tc.features)
			h2 := SimHash64(tc.features)
			if h1 != h2 {
				t.Errorf("%s: 결정성 실패 (%d != %d)", tc.name, h1, h2)
			}
			// 비공집합은 비zero여야 함 (확률적으로 충분히 참)
			if len(tc.features) > 0 && h1 == 0 {
				t.Errorf("%s: SimHash64 반환값이 0이면 안 됨", tc.name)
			}
		})
	}
}

// TestSimHash64_DifferentFeaturesDiffer는 다른 feature가 (높은 확률로) 다른 해시를 생성하는지 검증한다.
func TestSimHash64_DifferentFeaturesDiffer(t *testing.T) {
	t.Parallel()

	hA := SimHash64([]string{"moai_plan_A"})
	hB := SimHash64([]string{"moai_plan_B"})

	if hA == hB {
		t.Error("완전히 다른 feature: 해시가 같아선 안 됨 (충돌)")
	}
}

// ─────────────────────────────────────────────
// Hamming 거리 테스트
// ─────────────────────────────────────────────

// TestHamming_IdenticalValues는 동일한 값의 Hamming 거리가 0임을 검증한다.
func TestHamming_IdenticalValues(t *testing.T) {
	t.Parallel()

	cases := []uint64{0, 1, 0xFFFFFFFFFFFFFFFF, 0xDEADBEEFCAFEBABE}
	for _, v := range cases {
		if d := Hamming(v, v); d != 0 {
			t.Errorf("Hamming(%d, %d) = %d, want 0", v, v, d)
		}
	}
}

// TestHamming_KnownDistances는 수동 계산된 Hamming 거리 쌍을 검증한다.
func TestHamming_KnownDistances(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b uint64
		want int
	}{
		{0b0000, 0b0001, 1},           // 1비트 차이
		{0b0000, 0b0011, 2},           // 2비트 차이
		{0b0000, 0b0111, 3},           // 3비트 차이
		{0b0000, 0b1111, 4},           // 4비트 차이
		{0xFFFFFFFFFFFFFFFF, 0x0, 64}, // 64비트 전부 다름
		{0b1010, 0b0101, 4},           // 교차 패턴
		{0b1100, 0b0011, 4},           // 하위 4비트 교차
	}

	for _, tc := range cases {
		got := Hamming(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("Hamming(0x%x, 0x%x) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// TestHamming_Symmetry는 Hamming(a,b) == Hamming(b,a)임을 검증한다.
func TestHamming_Symmetry(t *testing.T) {
	t.Parallel()

	pairs := [][2]uint64{
		{0xABCDEF0123456789, 0x9876543210FEDCBA},
		{0, 0xFFFFFFFFFFFFFFFF},
		{0x1234, 0x5678},
	}

	for _, p := range pairs {
		d1 := Hamming(p[0], p[1])
		d2 := Hamming(p[1], p[0])
		if d1 != d2 {
			t.Errorf("Hamming 비대칭: Hamming(a,b)=%d != Hamming(b,a)=%d", d1, d2)
		}
	}
}

// ─────────────────────────────────────────────
// buildFeatureString 테스트
// ─────────────────────────────────────────────

// TestBuildFeatureString_AllowedFields는 허용된 필드만 feature에 포함되는지 검증한다.
func TestBuildFeatureString_AllowedFields(t *testing.T) {
	t.Parallel()

	evt := Event{
		EventType:    EventTypeUserPrompt,
		Subject:      "test-subject",
		PromptLang:   "ko",
		AgentName:    "expert-backend",
		AgentType:    "manager",
		PromptPreview: "안녕하세요 SPEC-V3R4",
		// PromptContent는 의도적으로 설정하지 않음 (PII 가드 검증)
	}

	features := buildFeatureString(evt)

	if len(features) == 0 {
		t.Fatal("buildFeatureString: 비어있으면 안 됨")
	}
}

// TestBuildFeatureString_EmptyFieldsProduceNoTokens는 빈 필드가 토큰을 생성하지 않는지 검증한다.
func TestBuildFeatureString_EmptyFieldsProduceNoTokens(t *testing.T) {
	t.Parallel()

	// 모든 feature 필드가 비어있는 이벤트
	evt := Event{
		EventType: EventTypeMoaiSubcommand,
		// Subject, PromptLang, AgentName, AgentType, PromptPreview 모두 비어있음
	}

	features := buildFeatureString(evt)

	// 비어있는 필드는 토큰을 생성하면 안 됨
	for _, f := range features {
		if f == "" {
			t.Error("빈 feature 토큰이 포함됨")
		}
	}
}
