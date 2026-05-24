// Package proposalgen mapper unit tests.
// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-004..005.
package proposalgen

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// TestMapper_CurrentDataNoOp verifies AC-PGN-002: the mapper produces 0
// actionable proposals from the current system-event-only fixture (4 unique
// pattern_keys, all matching system-event prefixes excluded by REQ-PGN-004).
func TestMapper_CurrentDataNoOp(t *testing.T) {
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
	if got, want := len(candidates), 0; got != want {
		t.Errorf("len(candidates) = %d, want %d (current data is system-event only)", got, want)
		for _, c := range candidates {
			t.Logf("unexpected candidate: pattern=%q tier=%q conf=%v", c.PatternKey, c.Tier, c.Confidence)
		}
	}
}

// TestMapper_ActionablePattern verifies REQ-PGN-004 happy path: a synthetic
// promotion with an actionable pattern_key (code_change prefix) and
// sufficient confidence + non-pre-actionable tier produces 1 candidate.
func TestMapper_ActionablePattern(t *testing.T) {
	t.Parallel()

	in := []harness.Promotion{
		{
			Ts:               time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
			PatternKey:       "code_change:func_extract:auth_module",
			FromTier:         "observation",
			ToTier:           "recommendation",
			ObservationCount: 7,
			Confidence:       0.85,
		},
	}
	candidates := MapPromotions(in)
	if got, want := len(candidates), 1; got != want {
		t.Fatalf("len(candidates) = %d, want %d", got, want)
	}
	c := candidates[0]
	if c.PatternKey != "code_change:func_extract:auth_module" {
		t.Errorf("PatternKey = %q, want code_change:func_extract:auth_module", c.PatternKey)
	}
	if c.Tier != "recommendation" {
		t.Errorf("Tier = %q, want recommendation", c.Tier)
	}
	if c.Confidence != 0.85 {
		t.Errorf("Confidence = %v, want 0.85", c.Confidence)
	}
	if c.ObservationCount != 7 {
		t.Errorf("ObservationCount = %d, want 7", c.ObservationCount)
	}
	if c.DraftID == "" || !strings.HasPrefix(c.DraftID, "PROPOSAL-") {
		t.Errorf("DraftID = %q, want PROPOSAL-<date>-<hash>", c.DraftID)
	}
}

// TestMapper_ConfidenceThresholdBoundary verifies REQ-PGN-004: the >= 0.70
// boundary is inclusive at 0.70 and exclusive below it.
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
					PatternKey:       "code_change:rename:user_service",
					ToTier:           "approval_required",
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

// TestMapper_TierFilter verifies REQ-PGN-004: only "recommendation" and
// "approval_required" tiers map to candidates; "observation" and
// "auto_update" are pre-actionable and excluded.
func TestMapper_TierFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tier      string
		wantCount int
	}{
		{tier: "observation", wantCount: 0},
		{tier: "auto_update", wantCount: 0},
		{tier: "recommendation", wantCount: 1},
		{tier: "approval_required", wantCount: 1},
		{tier: "", wantCount: 0},
		{tier: "unknown_tier", wantCount: 0},
	}
	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			in := []harness.Promotion{
				{
					Ts:               time.Now().UTC(),
					PatternKey:       "error_pattern:nil_deref:payment_handler",
					ToTier:           tt.tier,
					ObservationCount: 4,
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

// TestMapper_RegexPrefixFilter verifies REQ-PGN-004: only pattern_keys with
// allowed actionable prefixes (code_change, error_pattern, tool_failure,
// repeated_edit) and the canonical `<prefix>:<subject>:<scope>` shape are
// admitted; system-event prefixes (session_stop, user_prompt, etc.) are
// excluded.
func TestMapper_RegexPrefixFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		patternKey string
		wantAdmit  bool
	}{
		// Actionable prefixes (in scope).
		{"code_change:func_extract:auth_module", true},
		{"error_pattern:nil_deref:payment_handler", true},
		{"tool_failure:bash_timeout:db_migrate", true},
		{"repeated_edit:format_fix:user_test", true},

		// System-event prefixes (out of scope per §1.2).
		{"agent_invocation:Bash:", false},
		{"subagent_stop:unknown:", false},
		{"session_stop::", false},
		{"user_prompt::", false},

		// Malformed shape (must reject without panic).
		{"no_colons_at_all", false},
		{"only:one_colon", false},
		{"code_change:bad-segment:scope", false}, // hyphen not in [a-z_]
		{"CODE_CHANGE:func_extract:scope", false}, // uppercase not allowed in subject prefix
	}
	for _, tt := range tests {
		t.Run(tt.patternKey, func(t *testing.T) {
			in := []harness.Promotion{
				{
					Ts:               time.Now().UTC(),
					PatternKey:       tt.patternKey,
					ToTier:           "recommendation",
					ObservationCount: 6,
					Confidence:       0.85,
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

// TestMapper_DedupByPatternKey verifies that repeated promotions for the same
// pattern_key produce at most one candidate (the latest by Ts).
func TestMapper_DedupByPatternKey(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 5, 24, 11, 0, 0, 0, time.UTC)
	in := []harness.Promotion{
		{
			Ts: older, PatternKey: "code_change:func_extract:auth_module",
			ToTier: "recommendation", ObservationCount: 3, Confidence: 0.80,
		},
		{
			Ts: newer, PatternKey: "code_change:func_extract:auth_module",
			ToTier: "recommendation", ObservationCount: 7, Confidence: 0.95,
		},
	}
	got := MapPromotions(in)
	if len(got) != 1 {
		t.Fatalf("len(candidates) = %d, want 1 (dedup expected)", len(got))
	}
	if got[0].ObservationCount != 7 || got[0].Confidence != 0.95 {
		t.Errorf("dedup kept stale record: count=%d conf=%v, want count=7 conf=0.95",
			got[0].ObservationCount, got[0].Confidence)
	}
	if !got[0].SourceTs.Equal(newer) {
		t.Errorf("dedup kept stale timestamp: got=%v want=%v", got[0].SourceTs, newer)
	}
}

// TestMapper_DraftIDFormat verifies REQ-PGN-006 partial: draft IDs follow the
// PROPOSAL-<YYYYMMDD>-<hash> format and are stable across mapper calls for
// the same pattern_key + date.
func TestMapper_DraftIDFormat(t *testing.T) {
	t.Parallel()

	in := []harness.Promotion{
		{
			Ts: time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
			PatternKey: "code_change:func_extract:auth_module",
			ToTier: "recommendation", ObservationCount: 5, Confidence: 0.9,
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

// TestSortCandidatesByConfidenceDesc verifies the deterministic ordering rule
// invoked by the CLI: candidates are sorted by confidence descending so the
// orchestrator's --limit truncation favors high-confidence proposals.
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
