// Package harness — classifier_schema_regression_test.go
// Wave D T-D4: 스키마 버전 회귀 방지 테스트.
// AC-HRN-CLS-005: Event schema_version 필드가 JSONL 라운드트립에서 보존된다.
// 이 파일은 Wave A-C 구현이 깨지지 않는지 지속 감시하는 가드다.
package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSchemaRegression_SchemaVersionPreservedInEvent는 Event 직렬화/역직렬화 시
// schema_version 필드가 보존되는지 검증한다 (AC-HRN-CLS-005).
func TestSchemaRegression_SchemaVersionPreservedInEvent(t *testing.T) {
	t.Parallel()

	evt := Event{
		EventType:     EventTypeUserPrompt,
		Subject:       "schema-regression-subject",
		ContextHash:   "schctx1",
		SchemaVersion: "v1",
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Event 직렬화 실패: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Event 역직렬화 실패: %v", err)
	}

	if decoded.SchemaVersion != "v1" {
		t.Errorf("schema_version 라운드트립 실패: got %q, want %q", decoded.SchemaVersion, "v1")
	}
}

// TestSchemaRegression_AggregateHandlesV1Events는 schema_version=v1 이벤트가 있는 JSONL을
// AggregatePatterns가 오류 없이 읽는지 검증한다.
func TestSchemaRegression_AggregateHandlesV1Events(t *testing.T) {
	t.Parallel()

	// stage2_similar_10.jsonl (schema_version=v1 포함) 읽기
	logPath := filepath.Join("testdata", "stage2_similar_10.jsonl")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Skip("stage2_similar_10.jsonl 없음 — Wave C fixture 확인")
	}

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}
	if len(patterns) == 0 {
		t.Error("패턴 맵 비어있음 — v1 이벤트 파싱 실패 가능성")
	}
}

// TestSchemaRegression_PatternKeyFormat은 buildPatternKey 출력이
// "event_type:subject:context_hash" 형식임을 검증한다 (Wave A 계약).
func TestSchemaRegression_PatternKeyFormat(t *testing.T) {
	t.Parallel()

	key := buildPatternKey(EventTypeUserPrompt, "my-subject", "hash-abc")
	want := "user_prompt:my-subject:hash-abc"
	if key != want {
		t.Errorf("buildPatternKey = %q, want %q", key, want)
	}
}

// TestSchemaRegression_MergedKeyFormat은 clusterSingletons의 merged_key 형식이
// "event_type:lex-min-subject" (2-field)임을 검증한다 (Wave C 계약).
func TestSchemaRegression_MergedKeyFormat(t *testing.T) {
	t.Parallel()

	var patterns = make(map[string]*Pattern)
	var events []Event
	for i := range 4 {
		subject := "merged-key-format-check-" + string(rune('a'+i))
		preview := "moai run merged key format regression check test"
		p, evt := makePatternAndEvent(EventTypeUserPrompt, subject, "ctx-mf", preview, defaultConfidence)
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

	result, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// merged_key는 "event_type:lex-min-subject" 2-field 형식이어야 함
	for key := range result {
		// Stage-2 병합된 키: 2개 필드 (event_type:subject)
		// Stage-1 원본 키: 3개 필드 (event_type:subject:context_hash) — count=1이면 병합 안 됨
		_ = key // 형식 검증은 audit log에서 수행
	}

	// audit log의 merged_key 형식 검증
	data, readErr := os.ReadFile(auditLogPath)
	if readErr != nil {
		t.Fatalf("감사 로그 읽기 실패: %v", readErr)
	}

	for _, line := range splitNonEmpty(string(data)) {
		var entry struct {
			MergedKey string `json:"merged_key"`
		}
		if jsonErr := json.Unmarshal([]byte(line), &entry); jsonErr != nil {
			t.Fatalf("감사 로그 파싱 실패: %v", jsonErr)
		}
		// "event_type:subject" 형식: 정확히 1개의 콜론
		colonCount := 0
		for _, c := range entry.MergedKey {
			if c == ':' {
				colonCount++
			}
		}
		if colonCount != 1 {
			t.Errorf("merged_key %q: 콜론 수 = %d, want 1 (event_type:subject 형식)", entry.MergedKey, colonCount)
		}
	}
}
