// Package harness — classifier_simhash.go
// SimHash 64-bit 지문 생성, Hamming 거리, feature 문자열 빌더.
// REQ-HRN-CLS-002: 결정적(deterministic) SimHash로 유사 패턴 클러스터 탐지 지원.
package harness

import (
	"hash/fnv"
	"math/bits"
	"strings"
	"unicode"
)

// SimHash64는 Charikar SimHash 알고리즘으로 feature 슬라이스를 64-bit 지문으로 변환한다.
// FNV-1a 64-bit 해시를 기반으로 각 비트에 가중치를 누적한 후 부호로 지문을 생성한다.
// 빈 슬라이스나 nil 슬라이스는 0을 반환한다.
//
// @MX:ANCHOR: [AUTO] Wave C clusterSingletons와 classifier_cluster_test.go에서 호출됨.
// @MX:REASON: [AUTO] fan_in >= 3: clusterSingletons, classifier_cluster_test.go, classifier_simhash_test.go
// @MX:SPEC: REQ-HRN-CLS-002
func SimHash64(features []string) uint64 {
	if len(features) == 0 {
		return 0
	}

	// v[i]는 i번째 비트의 가중치 합 (양수 → 1, 음수 → 0)
	var v [64]int

	for _, feature := range features {
		if feature == "" {
			continue
		}
		// FNV-1a 64-bit 해시
		h := fnvHash64(feature)
		// 각 비트 위치에 +1(해시 비트=1) 또는 -1(해시 비트=0) 누적
		for i := range 64 {
			if (h>>uint(i))&1 == 1 {
				v[i]++
			} else {
				v[i]--
			}
		}
	}

	// 부호에 따라 지문 비트 결정
	var fingerprint uint64
	for i := range 64 {
		if v[i] > 0 {
			fingerprint |= 1 << uint(i)
		}
	}
	return fingerprint
}

// Hamming은 두 uint64 값의 Hamming 거리(다른 비트 수)를 반환한다.
// bits.OnesCount64(a ^ b)를 사용하여 O(1)로 계산한다.
func Hamming(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// tokenize는 문자열을 소문자 유니코드 단어 토큰 슬라이스로 분할한다.
// 스테밍 없이 단어 경계만 사용한다. 빈 문자열은 빈 슬라이스를 반환한다.
//
// @MX:NOTE: [AUTO] Unicode word-boundary split. 스테밍 미적용 — 단순성 우선 (Karpathy Simplicity First).
func tokenize(s string) []string {
	if s == "" {
		return nil
	}

	lower := strings.ToLower(s)
	var tokens []string
	var cur strings.Builder

	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

// buildFeatureString은 Event에서 SimHash 입력용 feature 슬라이스를 생성한다.
// 허용 필드: subject, prompt_preview, prompt_lang, agent_name, agent_type.
// full prompt 텍스트 필드(prompt_content JSON 키)는 PII 가드(REQ-HRN-CLS-014)에 의해
// 명시적으로 제외된다 — Event.PromptPreview(64바이트 미리보기)만 허용.
// 빈 필드는 토큰을 생성하지 않는다.
//
// @MX:NOTE: [AUTO] PII 가드: full-prompt 필드는 이 빌더에서 의도적으로 제외됨.
// @MX:SPEC: REQ-HRN-CLS-014
func buildFeatureString(evt Event) []string {
	var features []string

	// 허용된 필드에서 토큰화하여 feature 추가
	// 닫힌 switch: 새 필드 추가 시 @MX:NOTE 코멘트와 근거 필수
	allowed := []string{
		evt.Subject,
		evt.PromptPreview,
		evt.PromptLang,
		evt.AgentName,
		evt.AgentType,
	}

	for _, field := range allowed {
		if field == "" {
			continue
		}
		tokens := tokenize(field)
		features = append(features, tokens...)
	}

	return features
}

// fnvHash64는 FNV-1a 64-bit 해시를 계산한다.
// stdlib hash/fnv를 사용하여 stdlib-only 요건을 준수한다.
func fnvHash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
