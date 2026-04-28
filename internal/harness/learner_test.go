// Package harness — learner.go 테스트.
// REQ-HL-002: 패턴 집계, tier 분류, promotion 기록 검증.
package harness

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ─────────────────────────────────────────────
// AggregatePatterns 테스트 (T-P2-02)
// ─────────────────────────────────────────────

// TestAggregatePatterns_EmptyFile은 빈 JSONL 파일에서 빈 map을 반환하는지 검증한다.
func TestAggregatePatterns_EmptyFile(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "usage-log.jsonl")
	if err := os.WriteFile(logPath, []byte{}, 0o644); err != nil {
		t.Fatalf("로그 파일 생성 실패: %v", err)
	}

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("빈 파일: len(patterns) = %d, want 0", len(patterns))
	}
}

// TestAggregatePatterns_FileNotExist은 파일이 없으면 빈 map을 반환하는지 검증한다.
func TestAggregatePatterns_FileNotExist(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "no-such-file.jsonl")

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("없는 파일: len(patterns) = %d, want 0", len(patterns))
	}
}

// TestAggregatePatterns_Groups은 1,000개 이벤트를 (event_type,subject,context_hash)로 그룹핑하는지 검증한다.
func TestAggregatePatterns_Groups(t *testing.T) {
	t.Parallel()

	logPath := writeSyntheticEvents(t, 1000)

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}

	// 10가지 (event_type, subject, context_hash) 조합 * 100 = 1000 이벤트
	// 각 패턴의 count = 100
	if len(patterns) != 10 {
		t.Errorf("패턴 수 = %d, want 10", len(patterns))
	}
	for key, p := range patterns {
		if p.Count != 100 {
			t.Errorf("패턴[%s].Count = %d, want 100", key, p.Count)
		}
	}
}

// TestAggregatePatterns_CountAccumulation은 동일 키가 count를 누적하는지 검증한다.
func TestAggregatePatterns_CountAccumulation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")

	events := []Event{
		makeEvent(EventTypeMoaiSubcommand, "/moai plan", "hash1"),
		makeEvent(EventTypeMoaiSubcommand, "/moai plan", "hash1"),
		makeEvent(EventTypeMoaiSubcommand, "/moai plan", "hash1"),
		makeEvent(EventTypeAgentInvocation, "expert-backend", "hash2"),
		makeEvent(EventTypeAgentInvocation, "expert-backend", "hash2"),
	}
	writeEvents(t, logPath, events)

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}
	if len(patterns) != 2 {
		t.Fatalf("패턴 수 = %d, want 2", len(patterns))
	}

	key1 := patternKey(EventTypeMoaiSubcommand, "/moai plan", "hash1")
	if patterns[key1].Count != 3 {
		t.Errorf("패턴[key1].Count = %d, want 3", patterns[key1].Count)
	}

	key2 := patternKey(EventTypeAgentInvocation, "expert-backend", "hash2")
	if patterns[key2].Count != 2 {
		t.Errorf("패턴[key2].Count = %d, want 2", patterns[key2].Count)
	}
}

// TestAggregatePatterns_MalformedLinesSkipped은 파싱 실패 줄을 건너뛰는지 검증한다.
func TestAggregatePatterns_MalformedLinesSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")

	good := makeEvent(EventTypeFeedback, "/moai feedback", "h1")
	data, _ := json.Marshal(good)
	content := string(data) + "\n" + "not-valid-json\n" + string(data) + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}
	key := patternKey(EventTypeFeedback, "/moai feedback", "h1")
	if patterns[key].Count != 2 {
		t.Errorf("유효 이벤트 count = %d, want 2", patterns[key].Count)
	}
}

// ─────────────────────────────────────────────
// ClassifyTier 테스트 (T-P2-03)
// ─────────────────────────────────────────────

// TestClassifyTier_BoundaryValues는 {0,1,2,3,4,5,9,10,11}에서 올바른 tier를 반환하는지 검증한다.
func TestClassifyTier_BoundaryValues(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}
	cases := []struct {
		count      int
		confidence float64
		wantTier   Tier
	}{
		{0, 0.90, TierObservation},  // count=0: 아직 미관찰
		{1, 0.90, TierObservation},  // count=1: Observation
		{2, 0.90, TierObservation},  // count=2: 아직 Observation
		{3, 0.90, TierHeuristic},    // count=3: Heuristic
		{4, 0.90, TierHeuristic},    // count=4: 아직 Heuristic
		{5, 0.90, TierRule},         // count=5: Rule
		{9, 0.90, TierRule},         // count=9: 아직 Rule
		{10, 0.90, TierAutoUpdate},  // count=10: AutoUpdate
		{11, 0.90, TierAutoUpdate},  // count=11: 여전히 AutoUpdate
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("count=%d_conf=%.2f", tc.count, tc.confidence), func(t *testing.T) {
			t.Parallel()
			p := &Pattern{Count: tc.count, Confidence: tc.confidence}
			got := ClassifyTier(p, thresholds)
			if got != tc.wantTier {
				t.Errorf("count=%d confidence=%.2f: got %s, want %s",
					tc.count, tc.confidence, got, tc.wantTier)
			}
		})
	}
}

// TestClassifyTier_LowConfidenceForceObservation은 신뢰도 < 0.70이면 count에 관계없이 TierObservation을 반환하는지 검증한다.
func TestClassifyTier_LowConfidenceForceObservation(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}
	counts := []int{1, 3, 5, 10, 100}

	for _, count := range counts {
		count := count
		t.Run(fmt.Sprintf("count=%d_lowconf", count), func(t *testing.T) {
			t.Parallel()
			p := &Pattern{Count: count, Confidence: 0.69} // 0.70 미만
			got := ClassifyTier(p, thresholds)
			if got != TierObservation {
				t.Errorf("count=%d confidence=0.69: got %s, want observation", count, got)
			}
		})
	}
}

// TestClassifyTier_ExactBoundaryConfidence는 0.70 경계에서 올바르게 분류하는지 검증한다.
func TestClassifyTier_ExactBoundaryConfidence(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}

	// 정확히 0.70: count 충분하면 Observation 탈출 가능
	p := &Pattern{Count: 3, Confidence: 0.70}
	got := ClassifyTier(p, thresholds)
	if got != TierHeuristic {
		t.Errorf("count=3 confidence=0.70: got %s, want heuristic", got)
	}

	// 0.699: 여전히 Observation 강제
	p2 := &Pattern{Count: 3, Confidence: 0.699}
	got2 := ClassifyTier(p2, thresholds)
	if got2 != TierObservation {
		t.Errorf("count=3 confidence=0.699: got %s, want observation", got2)
	}
}

// TestClassifyTier_EmptyThresholds는 thresholds가 없으면 TierObservation을 반환하는지 검증한다.
func TestClassifyTier_EmptyThresholds(t *testing.T) {
	t.Parallel()

	p := &Pattern{Count: 100, Confidence: 0.99}
	got := ClassifyTier(p, nil)
	if got != TierObservation {
		t.Errorf("empty thresholds: got %s, want observation", got)
	}
}

// ─────────────────────────────────────────────
// WritePromotion 테스트 (T-P2-06)
// ─────────────────────────────────────────────

// TestWritePromotion_AppendsLine은 Promotion이 JSONL 파일에 올바르게 기록되는지 검증한다.
func TestWritePromotion_AppendsLine(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	promoPath := filepath.Join(dir, "tier-promotions.jsonl")

	now := time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)
	promo := Promotion{
		Ts:               now,
		PatternKey:       "moai_subcommand:/moai plan",
		FromTier:         TierObservation.String(),
		ToTier:           TierHeuristic.String(),
		ObservationCount: 3,
		Confidence:       0.82,
	}

	l := NewLearner(promoPath)
	if err := l.WritePromotion(promo); err != nil {
		t.Fatalf("WritePromotion 오류: %v", err)
	}

	// 파일이 존재하고 유효한 JSON인지 검증
	data, err := os.ReadFile(promoPath)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("라인 수 = %d, want 1", len(lines))
	}

	var got Promotion
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}
	if got.PatternKey != promo.PatternKey {
		t.Errorf("PatternKey = %q, want %q", got.PatternKey, promo.PatternKey)
	}
	if got.FromTier != promo.FromTier {
		t.Errorf("FromTier = %q, want %q", got.FromTier, promo.FromTier)
	}
	if got.ToTier != promo.ToTier {
		t.Errorf("ToTier = %q, want %q", got.ToTier, promo.ToTier)
	}
	if got.ObservationCount != promo.ObservationCount {
		t.Errorf("ObservationCount = %d, want %d", got.ObservationCount, promo.ObservationCount)
	}
}

// TestWritePromotion_Appends은 여러 번 호출 시 누적 append되는지 검증한다.
func TestWritePromotion_Appends(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	promoPath := filepath.Join(dir, "tier-promotions.jsonl")
	l := NewLearner(promoPath)

	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		promo := Promotion{
			Ts:               now,
			PatternKey:       fmt.Sprintf("moai_subcommand:/moai cmd%d", i),
			FromTier:         TierObservation.String(),
			ToTier:           TierHeuristic.String(),
			ObservationCount: 3,
			Confidence:       0.80,
		}
		if err := l.WritePromotion(promo); err != nil {
			t.Fatalf("WritePromotion[%d] 오류: %v", i, err)
		}
	}

	f, err := os.Open(promoPath)
	if err != nil {
		t.Fatalf("파일 열기 실패: %v", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			lineCount++
		}
	}
	if lineCount != 3 {
		t.Errorf("라인 수 = %d, want 3", lineCount)
	}
}

// TestWritePromotion_DirectoryAutoCreate은 부모 디렉토리가 없어도 자동 생성하는지 검증한다.
func TestWritePromotion_DirectoryAutoCreate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	promoPath := filepath.Join(dir, "learning-history", "tier-promotions.jsonl")
	l := NewLearner(promoPath)

	promo := Promotion{
		Ts:               time.Now().UTC(),
		PatternKey:       "feedback:/moai feedback",
		FromTier:         TierHeuristic.String(),
		ToTier:           TierRule.String(),
		ObservationCount: 5,
		Confidence:       0.85,
	}
	if err := l.WritePromotion(promo); err != nil {
		t.Fatalf("WritePromotion 오류: %v", err)
	}
	if _, err := os.Stat(promoPath); os.IsNotExist(err) {
		t.Error("프로모션 파일이 생성되지 않음")
	}
}

// ─────────────────────────────────────────────
// Tier.String 테스트
// ─────────────────────────────────────────────

// TestTierString은 Tier 열거형의 String() 결과를 검증한다.
func TestTierString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		tier Tier
		want string
	}{
		{TierObservation, "observation"},
		{TierHeuristic, "heuristic"},
		{TierRule, "rule"},
		{TierAutoUpdate, "auto_update"},
		{Tier(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.tier.String(); got != tc.want {
			t.Errorf("Tier(%d).String() = %q, want %q", tc.tier, got, tc.want)
		}
	}
}

// ─────────────────────────────────────────────
// 테스트 헬퍼
// ─────────────────────────────────────────────

// makeEvent는 테스트용 Event를 생성한다.
func makeEvent(et EventType, subject, contextHash string) Event {
	return Event{
		Timestamp:     time.Now().UTC(),
		EventType:     et,
		Subject:       subject,
		ContextHash:   contextHash,
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
}

// writeEvents는 이벤트 슬라이스를 JSONL 파일로 기록한다.
func writeEvents(t *testing.T, logPath string, events []Event) {
	t.Helper()

	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("인코딩 실패: %v", err)
		}
	}
}

// writeSyntheticEvents는 10가지 패턴 * 100 반복 = 1,000개 이벤트를 기록한다.
func writeSyntheticEvents(t *testing.T, total int) string {
	t.Helper()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")

	// 10가지 (event_type, subject, context_hash) 조합
	combos := []struct {
		et      EventType
		subject string
		hash    string
	}{
		{EventTypeMoaiSubcommand, "/moai plan", "h1"},
		{EventTypeMoaiSubcommand, "/moai run", "h2"},
		{EventTypeMoaiSubcommand, "/moai sync", "h3"},
		{EventTypeAgentInvocation, "expert-backend", "h4"},
		{EventTypeAgentInvocation, "expert-frontend", "h5"},
		{EventTypeAgentInvocation, "manager-spec", "h6"},
		{EventTypeSpecReference, "SPEC-001", "h7"},
		{EventTypeSpecReference, "SPEC-002", "h8"},
		{EventTypeFeedback, "/moai feedback", "h9"},
		{EventTypeMoaiSubcommand, "/moai loop", "h10"},
	}

	perPattern := total / len(combos)
	var events []Event
	for _, c := range combos {
		for i := 0; i < perPattern; i++ {
			events = append(events, makeEvent(c.et, c.subject, c.hash))
		}
	}

	writeEvents(t, logPath, events)
	return logPath
}

// patternKey는 AggregatePatterns가 반환하는 map의 키를 생성한다.
func patternKey(et EventType, subject, contextHash string) string {
	return fmt.Sprintf("%s:%s:%s", et, subject, contextHash)
}
