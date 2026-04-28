//go:build integration
// +build integration

// Package harness_integration — SPEC-V3R3-HARNESS-LEARNING-001 통합 테스트.
// T-P5-01: 100-이벤트 세션 재생 + tier 분포 검증 (REQ-HL-001, REQ-HL-002).
package harness_integration

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// TestIT01_ReplaySession은 100개 이벤트를 JSONL 로그에 기록한 뒤
// AggregatePatterns와 ClassifyTier를 통해 tier 분포를 검증한다.
// REQ-HL-001: 이벤트 스키마 준수.
// REQ-HL-002: 집계 후 Tier 분류 정확성.
//
// 결정론적 보장: 고정 시드(seed=42)와 고정 타임스탬프를 사용한다.
func TestIT01_ReplaySession(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")

	// ── 고정 시드 난수 생성기 (결정론적 재현 보장) ──────────────────────
	// rand.New + rand.NewSource로 seed=42 고정
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // 테스트용 비암호화 난수

	// 고정 기준 타임스탬프 (clock 의존 없음)
	baseTime := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	// ── 100개 이벤트 시나리오 정의 ────────────────────────────────────────
	// 시나리오:
	// - "/moai plan" 이벤트 × 15회 (TierRule 기대: count>=5)
	// - "expert-backend" 이벤트 × 10회 (TierRule 기대: count>=5)
	// - "/moai run" 이벤트 × 3회 (TierHeuristic 기대: count==3)
	// - "SPEC-001" 이벤트 × 2회 (TierObservation 기대: count==2)
	// - 나머지 70개: 각각 1회 고유 subject (TierObservation)
	//
	// 총합 = 15 + 10 + 3 + 2 + 70 = 100개

	type scenario struct {
		eventType harness.EventType
		subject   string
		count     int
	}

	scenarios := []scenario{
		{harness.EventTypeMoaiSubcommand, "/moai plan", 15},
		{harness.EventTypeAgentInvocation, "expert-backend", 10},
		{harness.EventTypeMoaiSubcommand, "/moai run", 3},
		{harness.EventTypeSpecReference, "SPEC-001", 2},
	}

	// 시나리오 이벤트 목록 구성
	var events []harness.Event
	for _, sc := range scenarios {
		for i := range sc.count {
			events = append(events, harness.Event{
				Timestamp:     baseTime.Add(time.Duration(len(events)) * time.Minute),
				EventType:     sc.eventType,
				Subject:       sc.subject,
				ContextHash:   "ctx_fixed",
				TierIncrement: 0,
				SchemaVersion: harness.LogSchemaVersion,
			})
			_ = i
		}
	}

	// 고유 subject 70개 추가 (각 1회)
	usedCount := len(events) // 현재 30개
	uniqueCount := 100 - usedCount
	for i := range uniqueCount {
		subject := randomSubject(rng, i)
		events = append(events, harness.Event{
			Timestamp:     baseTime.Add(time.Duration(len(events)) * time.Minute),
			EventType:     harness.EventTypeMoaiSubcommand,
			Subject:       subject,
			ContextHash:   "ctx_unique",
			TierIncrement: 0,
			SchemaVersion: harness.LogSchemaVersion,
		})
	}

	if len(events) != 100 {
		t.Fatalf("이벤트 수 = %d, want 100", len(events))
	}

	// ── JSONL 파일 기록 ────────────────────────────────────────────────────
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("로그 파일 생성 실패: %v", err)
	}
	enc := json.NewEncoder(f)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			_ = f.Close()
			t.Fatalf("이벤트 인코딩 실패: %v", err)
		}
	}
	_ = f.Close()

	// ── AggregatePatterns ─────────────────────────────────────────────────
	patterns, err := harness.AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 실패: %v", err)
	}

	// 패턴 수 검증: 고유 key 개수
	// - "moai_subcommand:/moai plan:ctx_fixed" → 1
	// - "agent_invocation:expert-backend:ctx_fixed" → 1
	// - "moai_subcommand:/moai run:ctx_fixed" → 1
	// - "spec_reference:SPEC-001:ctx_fixed" → 1
	// - 고유 subject 70개 (ctx_unique) → 70
	// 총 74개 패턴 기대
	expectedPatternCount := 4 + uniqueCount
	if len(patterns) != expectedPatternCount {
		t.Errorf("패턴 수 = %d, want %d", len(patterns), expectedPatternCount)
	}

	// ── ClassifyTier 검증 ─────────────────────────────────────────────────
	thresholds := []int{1, 3, 5, 10}

	// "/moai plan": count=15 → TierAutoUpdate (>=10)
	assertTier(t, patterns, "moai_subcommand:/moai plan:ctx_fixed", thresholds, harness.TierAutoUpdate)

	// "expert-backend": count=10 → TierAutoUpdate (>=10)
	assertTier(t, patterns, "agent_invocation:expert-backend:ctx_fixed", thresholds, harness.TierAutoUpdate)

	// "/moai run": count=3 → TierHeuristic (>=3, <5)
	assertTier(t, patterns, "moai_subcommand:/moai run:ctx_fixed", thresholds, harness.TierHeuristic)

	// "SPEC-001": count=2 → TierObservation (<3)
	assertTier(t, patterns, "spec_reference:SPEC-001:ctx_fixed", thresholds, harness.TierObservation)

	// tier 분포 집계: Observation >= 70개, Heuristic >= 1개, Rule/AutoUpdate >= 2개
	tierCounts := make(map[harness.Tier]int)
	for _, p := range patterns {
		tier := harness.ClassifyTier(p, thresholds)
		tierCounts[tier]++
	}

	if tierCounts[harness.TierObservation] < 70 {
		t.Errorf("TierObservation 수 = %d, 최소 70 기대", tierCounts[harness.TierObservation])
	}
	if tierCounts[harness.TierHeuristic] < 1 {
		t.Errorf("TierHeuristic 수 = %d, 최소 1 기대", tierCounts[harness.TierHeuristic])
	}
	if tierCounts[harness.TierAutoUpdate] < 2 {
		t.Errorf("TierAutoUpdate 수 = %d, 최소 2 기대", tierCounts[harness.TierAutoUpdate])
	}
}

// assertTier는 패턴 맵에서 key에 해당하는 패턴의 tier가 expected와 일치하는지 검증한다.
func assertTier(t *testing.T, patterns map[string]*harness.Pattern, key string, thresholds []int, expected harness.Tier) {
	t.Helper()
	p, ok := patterns[key]
	if !ok {
		t.Errorf("패턴 키 %q 없음", key)
		return
	}
	got := harness.ClassifyTier(p, thresholds)
	if got != expected {
		t.Errorf("패턴 %q tier = %s, want %s (count=%d)", key, got, expected, p.Count)
	}
}

// randomSubject는 고정 시드 rng로 고유 subject 문자열을 생성한다.
// 시드가 고정되어 있으므로 결정론적으로 동일한 값이 생성된다.
func randomSubject(rng *rand.Rand, idx int) string {
	// 결정론적: idx 기반으로 고유 subject 생성
	// rng는 호출 순서가 고정되므로 결과도 고정된다.
	n := rng.Intn(1000) + 1000 //nolint:gosec // 테스트용 비암호화 난수
	return "/moai_unique_" + itoa(n) + "_" + itoa(idx)
}

// itoa는 정수를 문자열로 변환하는 단순 헬퍼이다 (strconv 임포트 없이).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	// 역순
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
