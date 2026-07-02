// Package proposalgen mapper unit tests.
//
// 어휘/스키마 정렬 계약 (SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-001/002).
// 본 SPEC은 SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-004/005의 어휘·스키마 조항을
// 부분 supersede한다:
//   - actionable tier 어휘 = {rule, auto_update} (Tier.String() SSOT 파생)
//     — 구 {recommendation, approval_required}는 classifier가 방출하지 않는 유령 어휘.
//   - pattern_key 스키마 = <event_type>:<subject>:<context_hash>
//     (event_type ∈ EventType enum SSOT, subject/context_hash 빈 문자열 허용)
//     — 구 수기 prefix 목록 {code_change,error_pattern,tool_failure,repeated_edit}는
//     방출 observer가 존재하지 않는 유령 목록.
package proposalgen

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// TestMapper_CurrentDataProducesCandidates 는 구조적 0 결함의 해소를 실증한다
// (SPEC-HARNESS-EVO-PIPE-REPAIR-001 DoD-2 단위 레벨).
// 실측 baseline fixture(8 레코드, 4 unique pattern_key)에서 auto_update tier의
// 시스템 이벤트 3종(subagent_stop/session_stop/user_prompt)이 이제 후보로 채택된다.
// agent_invocation:Bash: 는 observation tier이므로 pre-actionable 배제 유지.
func TestMapper_CurrentDataProducesCandidates(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "tier-promotions-current-baseline.jsonl")
	promotions, _, err := ReadPromotions(path)
	if err != nil {
		t.Fatalf("ReadPromotions: %v", err)
	}
	if len(promotions) != 8 {
		t.Fatalf("setup: expected 8 records, got %d", len(promotions))
	}

	candidates := MapPromotions(promotions)
	// 4 unique key 중 observation(agent_invocation:Bash:) 1종 배제 → 3 candidate.
	if got, want := len(candidates), 3; got != want {
		t.Errorf("len(candidates) = %d, want %d (auto_update system events now actionable)", got, want)
		for _, c := range candidates {
			t.Logf("candidate: pattern=%q tier=%q conf=%v", c.PatternKey, c.Tier, c.Confidence)
		}
	}
	// agent_invocation:Bash: (observation) 는 배제되어야 한다.
	for _, c := range candidates {
		if c.PatternKey == "agent_invocation:Bash:" {
			t.Errorf("observation-tier pattern %q must not be admitted", c.PatternKey)
		}
	}
}

// TestMapper_RealDataAutoUpdate 는 AC-HEP-001a: 실데이터 형태의 auto_update 및 rule
// promotion이 채택됨을 검증한다 (구 impl에서는 0 — RED).
func TestMapper_RealDataAutoUpdate(t *testing.T) {
	t.Parallel()

	in := []harness.Promotion{
		{
			Ts:               time.Date(2026, 6, 17, 10, 53, 2, 0, time.UTC),
			PatternKey:       "user_prompt::",
			FromTier:         "",
			ToTier:           "auto_update",
			ObservationCount: 196,
			Confidence:       1,
		},
		{
			Ts:               time.Date(2026, 6, 17, 10, 53, 3, 0, time.UTC),
			PatternKey:       "moai_subcommand:/moai run:abc123",
			FromTier:         "heuristic",
			ToTier:           "rule",
			ObservationCount: 6,
			Confidence:       0.9,
		},
	}
	candidates := MapPromotions(in)
	if got, want := len(candidates), 2; got != want {
		t.Fatalf("len(candidates) = %d, want %d (auto_update + rule both actionable)", got, want)
	}
	tiers := map[string]bool{}
	for _, c := range candidates {
		tiers[c.Tier] = true
		if c.DraftID == "" || !strings.HasPrefix(c.DraftID, "PROPOSAL-") {
			t.Errorf("DraftID = %q, want PROPOSAL-<date>-<hash>", c.DraftID)
		}
	}
	if !tiers["auto_update"] || !tiers["rule"] {
		t.Errorf("expected both auto_update and rule tiers admitted, got %v", tiers)
	}
}

// TestMapper_PreActionableExcluded 는 AC-HEP-001b: observation/heuristic tier는
// confidence가 충분해도 pre-actionable로 배제됨을 검증한다.
func TestMapper_PreActionableExcluded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tier      string
		wantCount int
	}{
		{tier: "observation", wantCount: 0},
		{tier: "heuristic", wantCount: 0},
		{tier: "rule", wantCount: 1},
		{tier: "auto_update", wantCount: 1},
		{tier: "", wantCount: 0},
		{tier: "unknown_tier", wantCount: 0},
		// 구 유령 어휘는 더 이상 채택되지 않는다.
		{tier: "recommendation", wantCount: 0},
		{tier: "approval_required", wantCount: 0},
	}
	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			in := []harness.Promotion{
				{
					Ts:               time.Now().UTC(),
					PatternKey:       "session_stop::",
					ToTier:           tt.tier,
					ObservationCount: 12,
					Confidence:       0.9,
				},
			}
			got := MapPromotions(in)
			if len(got) != tt.wantCount {
				t.Errorf("tier=%q len(candidates) = %d, want %d", tt.tier, len(got), tt.wantCount)
			}
		})
	}
}

// TestMapper_RealDataSchemaPass 는 AC-HEP-002a: 실측 pattern_key 3형(빈 subject/hash 포함)이
// prefix/빈-segment 사유로 거부되지 않음을 검증한다 (tier+confidence 충족 시 채택).
func TestMapper_RealDataSchemaPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		patternKey string
		wantAdmit  bool
	}{
		// 실측 pattern_key 3형 — 모두 유효 EventType prefix + 빈 segment 허용.
		{"user_prompt::", true},          // 빈 subject + 빈 context_hash
		{"agent_invocation:Bash:", true}, // 빈 context_hash
		{"session_stop::", true},         // 빈 subject + 빈 context_hash
		{"subagent_stop:unknown:", true}, // 비어있지 않은 subject
		// 나머지 pattern-bearing EventType prefix.
		{"moai_subcommand:/moai plan:h1", true},
		{"spec_reference:SPEC-001:h2", true},
		{"feedback:issue:h3", true},

		// apply_outcome 는 pattern_key를 갖지 않으므로 파생 집합에서 제외 → 거부.
		{"apply_outcome::", false},
		// 비 EventType prefix (구 유령 목록) 는 거부.
		{"code_change:func_extract:auth_module", false},
		{"error_pattern:nil_deref:payment_handler", false},

		// 형식 위반 (segment 수 오류) 는 거부.
		{"no_colons_at_all", false},
		{"only:one_colon", false},
		{"session_stop:a:b:c", false}, // colon 4개 (segment 4)
	}
	for _, tt := range tests {
		t.Run(tt.patternKey, func(t *testing.T) {
			in := []harness.Promotion{
				{
					Ts:               time.Now().UTC(),
					PatternKey:       tt.patternKey,
					ToTier:           "auto_update",
					ObservationCount: 12,
					Confidence:       1,
				},
			}
			got := MapPromotions(in)
			gotAdmit := len(got) == 1
			if gotAdmit != tt.wantAdmit {
				t.Errorf("pattern=%q admitted=%v, want %v", tt.patternKey, gotAdmit, tt.wantAdmit)
			}
		})
	}
}

// TestMapper_ConfidenceThresholdBoundary 는 EC-3: confidence >= 0.70 경계가 tier 유효와
// 무관하게 유지됨을 검증한다 (0.70 inclusive, 미만 배제).
func TestMapper_ConfidenceThresholdBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		confidence float64
		wantCount  int
	}{
		{name: "below threshold (0.69)", confidence: 0.69, wantCount: 0},
		{name: "at threshold (0.70)", confidence: 0.70, wantCount: 1},
		{name: "above threshold (0.95)", confidence: 0.95, wantCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := []harness.Promotion{
				{
					Ts:               time.Now().UTC(),
					PatternKey:       "subagent_stop:unknown:",
					ToTier:           "rule",
					ObservationCount: 5,
					Confidence:       tt.confidence,
				},
			}
			got := MapPromotions(in)
			if len(got) != tt.wantCount {
				t.Errorf("len(candidates) = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// TestMapper_EmptyInput 는 EC-1: 빈/부재 입력이 빈(non-nil) slice + 무오류를 반환함을 검증한다.
func TestMapper_EmptyInput(t *testing.T) {
	t.Parallel()

	got := MapPromotions(nil)
	if got == nil {
		t.Fatal("MapPromotions(nil) = nil, want empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("len(candidates) = %d, want 0", len(got))
	}
}

// TestMapper_DedupByPatternKey 는 동일 pattern_key 반복이 최대 1 후보(최신 Ts 유지)를
// 산출함을 검증한다.
func TestMapper_DedupByPatternKey(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 5, 24, 11, 0, 0, 0, time.UTC)
	in := []harness.Promotion{
		{
			Ts: older, PatternKey: "user_prompt::",
			ToTier: "auto_update", ObservationCount: 33, Confidence: 1,
		},
		{
			Ts: newer, PatternKey: "user_prompt::",
			ToTier: "auto_update", ObservationCount: 196, Confidence: 1,
		},
	}
	got := MapPromotions(in)
	if len(got) != 1 {
		t.Fatalf("len(candidates) = %d, want 1 (dedup expected)", len(got))
	}
	if got[0].ObservationCount != 196 {
		t.Errorf("dedup kept stale record: count=%d, want count=196", got[0].ObservationCount)
	}
	if !got[0].SourceTs.Equal(newer) {
		t.Errorf("dedup kept stale timestamp: got=%v want=%v", got[0].SourceTs, newer)
	}
}

// TestMapper_DraftIDFormat 는 draft ID가 PROPOSAL-<YYYYMMDD>-<hash> 형식이며 동일 입력에
// 대해 안정적임을 검증한다.
func TestMapper_DraftIDFormat(t *testing.T) {
	t.Parallel()

	in := []harness.Promotion{
		{
			Ts:         time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
			PatternKey: "subagent_stop:unknown:",
			ToTier:     "auto_update", ObservationCount: 41, Confidence: 1,
		},
	}
	first := MapPromotions(in)
	second := MapPromotions(in)
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("setup: want 1 candidate per call")
	}
	if first[0].DraftID != second[0].DraftID {
		t.Errorf("DraftID not stable across calls: %q vs %q", first[0].DraftID, second[0].DraftID)
	}
	id := first[0].DraftID
	if !strings.HasPrefix(id, "PROPOSAL-20260524-") {
		t.Errorf("DraftID = %q, want prefix PROPOSAL-20260524-", id)
	}
	// Hash segment must be exactly 8 hex characters.
	if got := len(id) - len("PROPOSAL-20260524-"); got != 8 {
		t.Errorf("DraftID hash segment length = %d, want 8 (id=%q)", got, id)
	}
}

// TestSortCandidatesByConfidenceDesc 는 CLI가 호출하는 결정적 정렬(confidence 내림차순)을
// 검증한다.
func TestSortCandidatesByConfidenceDesc(t *testing.T) {
	t.Parallel()

	candidates := []ProposalCandidate{
		{PatternKey: "a", Confidence: 0.75},
		{PatternKey: "b", Confidence: 0.95},
		{PatternKey: "c", Confidence: 0.80},
	}
	SortCandidatesByConfidenceDesc(candidates)
	got := []string{candidates[0].PatternKey, candidates[1].PatternKey, candidates[2].PatternKey}
	want := []string{"b", "c", "a"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("sorted order = %v, want %v", got, want)
	}
}
