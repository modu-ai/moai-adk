// Package harness — learner.go tests.
// REQ-HL-002: pattern aggregation, tier classification, promotion record verification.
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
// AggregatePatterns tests (T-P2-02)
// ─────────────────────────────────────────────

// TestAggregatePatterns_EmptyFile verifies that an empty JSONL file returns an empty map.
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
		t.Errorf("empty file: len(patterns) = %d, want 0", len(patterns))
	}
}

// TestAggregatePatterns_FileNotExist verifies that a missing file returns an empty map.
func TestAggregatePatterns_FileNotExist(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "no-such-file.jsonl")

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("missing file: len(patterns) = %d, want 0", len(patterns))
	}
}

// TestAggregatePatterns_Groups verifies that 1,000 events are grouped by (event_type, subject, context_hash).
func TestAggregatePatterns_Groups(t *testing.T) {
	t.Parallel()

	logPath := writeSyntheticEvents(t, 1000)

	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}

	// 10 (event_type, subject, context_hash) combinations * 100 = 1000 events
	// Each pattern's count = 100
	if len(patterns) != 10 {
		t.Errorf("pattern count = %d, want 10", len(patterns))
	}
	for key, p := range patterns {
		if p.Count != 100 {
			t.Errorf("pattern[%s].Count = %d, want 100", key, p.Count)
		}
	}
}

// TestAggregatePatterns_CountAccumulation verifies that duplicate keys accumulate count.
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
		t.Fatalf("pattern count = %d, want 2", len(patterns))
	}

	key1 := patternKey(EventTypeMoaiSubcommand, "/moai plan", "hash1")
	if patterns[key1].Count != 3 {
		t.Errorf("pattern[key1].Count = %d, want 3", patterns[key1].Count)
	}

	key2 := patternKey(EventTypeAgentInvocation, "expert-backend", "hash2")
	if patterns[key2].Count != 2 {
		t.Errorf("pattern[key2].Count = %d, want 2", patterns[key2].Count)
	}
}

// TestAggregatePatterns_MalformedLinesSkipped verifies that lines failing to parse are skipped.
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
		t.Errorf("valid event count = %d, want 2", patterns[key].Count)
	}
}

// ─────────────────────────────────────────────
// ClassifyTier tests (T-P2-03)
// ─────────────────────────────────────────────

// TestClassifyTier_BoundaryValues verifies the correct tier for {0,1,2,3,4,5,9,10,11}.
func TestClassifyTier_BoundaryValues(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}
	cases := []struct {
		count      int
		confidence float64
		wantTier   Tier
	}{
		{0, 0.90, TierObservation}, // count=0: not yet observed
		{1, 0.90, TierObservation}, // count=1: Observation
		{2, 0.90, TierObservation}, // count=2: still Observation
		{3, 0.90, TierHeuristic},   // count=3: Heuristic
		{4, 0.90, TierHeuristic},   // count=4: still Heuristic
		{5, 0.90, TierRule},        // count=5: Rule
		{9, 0.90, TierRule},        // count=9: still Rule
		{10, 0.90, TierAutoUpdate}, // count=10: AutoUpdate
		{11, 0.90, TierAutoUpdate}, // count=11: still AutoUpdate
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

// TestClassifyTier_LowConfidenceForceObservation verifies that confidence < 0.70 forces TierObservation regardless of count.
func TestClassifyTier_LowConfidenceForceObservation(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}
	counts := []int{1, 3, 5, 10, 100}

	for _, count := range counts {
		count := count
		t.Run(fmt.Sprintf("count=%d_lowconf", count), func(t *testing.T) {
			t.Parallel()
			p := &Pattern{Count: count, Confidence: 0.69} // below 0.70
			got := ClassifyTier(p, thresholds)
			if got != TierObservation {
				t.Errorf("count=%d confidence=0.69: got %s, want observation", count, got)
			}
		})
	}
}

// TestClassifyTier_ExactBoundaryConfidence verifies correct classification at the 0.70 boundary.
func TestClassifyTier_ExactBoundaryConfidence(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}

	// Exactly 0.70: can escape Observation when count is sufficient
	p := &Pattern{Count: 3, Confidence: 0.70}
	got := ClassifyTier(p, thresholds)
	if got != TierHeuristic {
		t.Errorf("count=3 confidence=0.70: got %s, want heuristic", got)
	}

	// 0.699: still forced to Observation
	p2 := &Pattern{Count: 3, Confidence: 0.699}
	got2 := ClassifyTier(p2, thresholds)
	if got2 != TierObservation {
		t.Errorf("count=3 confidence=0.699: got %s, want observation", got2)
	}
}

// TestClassifyTier_EmptyThresholds verifies TierObservation is returned when thresholds are absent.
func TestClassifyTier_EmptyThresholds(t *testing.T) {
	t.Parallel()

	p := &Pattern{Count: 100, Confidence: 0.99}
	got := ClassifyTier(p, nil)
	if got != TierObservation {
		t.Errorf("empty thresholds: got %s, want observation", got)
	}
}

// ─────────────────────────────────────────────
// WritePromotion tests (T-P2-06)
// ─────────────────────────────────────────────

// TestWritePromotion_AppendsLine verifies that Promotion is recorded correctly into the JSONL file.
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

	// Verify the file exists and contains valid JSON
	data, err := os.ReadFile(promoPath)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1", len(lines))
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

// TestWritePromotion_Appends verifies cumulative appends across multiple calls.
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
		t.Errorf("line count = %d, want 3", lineCount)
	}
}

// TestWritePromotion_DirectoryAutoCreate verifies that a missing parent directory is auto-created.
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
		t.Error("promotion file was not created")
	}
}

// ─────────────────────────────────────────────
// Tier.String tests
// ─────────────────────────────────────────────

// TestTierString verifies the String() result of the Tier enum.
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
// Test helpers
// ─────────────────────────────────────────────

// makeEvent builds a test Event.
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

// writeEvents writes a slice of events to a JSONL file.
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

// writeSyntheticEvents writes 10 patterns * 100 repeats = 1,000 events.
func writeSyntheticEvents(t *testing.T, total int) string {
	t.Helper()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")

	// 10 (event_type, subject, context_hash) combinations
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

// patternKey constructs the map key returned by AggregatePatterns.
func patternKey(et EventType, subject, contextHash string) string {
	return fmt.Sprintf("%s:%s:%s", et, subject, contextHash)
}
