package runtime

import "strings"

// SPEC-AUDIT-SNAPSHOT-001 (A3, REQ-AUDIT-SNAPSHOT-003): the 4-dimension
// sync-audit workflow verdict becomes BINDING on the happy path (clean sync:
// all 4 dims pass). The cold sync-auditor subagent is spawned ONLY on
// INCOMPLETE / dim-0 / contested-finding fallback. This type + IsBinding()
// codify REQ-003's mechanical predicate so the binding decision is
// machine-evaluable and testable rather than left as orchestrator prose.
//
// The verdict shape mirrors what `.claude/workflows/sync-audit-4dim.js` emits
// (verdict / scores / zero_scored / missing / findings). The orchestrator
// consumes the workflow run output and constructs a FourDimVerdict for the
// IsBinding() decision. This is a pure consumer-side codification — the
// workflow's output schema is unchanged (OQ-1 RESOLVED).

// @MX:NOTE: [AUTO] SPEC-AUDIT-SNAPSHOT-001 A3 의 Go 체현. clean PASS 시 4-dim 워크플로
// verdict가 BINDING이 되는 기계 평가 가능 계약 (cold sync-auditor spawn 억제). 워크플로
// 출력 스키마가 바뀌면 IsBinding의 정규화도 함께 점검.
//
// FourDimVerdict is the Go-shaped mirror of the sync-audit-4dim.js verdict
// output, consumed to decide whether the workflow verdict is BINDING on the
// happy path (no cold sync-auditor spawn) or whether a fallback spawn is
// required.
//
// Fields are intentionally permissive (map[string]any slices) so the
// orchestrator can construct this from parsed JSON without a rigid binding to
// the workflow's exact key set; IsBinding normalizes via string formatting.
type FourDimVerdict struct {
	// Verdict is the top-level workflow verdict: "PASS", "FAIL", or
	// "INCOMPLETE". Only "PASS" can be binding, and only on the happy path.
	Verdict string

	// Scores carries the per-dimension score entries, each shaped
	// {"dimension": <name>, "score": <0..1>}. A score of 0 on any dimension
	// trips the dim-0 fallback.
	Scores []map[string]any

	// ZeroScored lists dimensions that scored 0 (the workflow emits this
	// discrete array in its FAIL branch).
	ZeroScored []string

	// Missing lists dimensions whose judge did not return a finite score (the
	// workflow emits this in its INCOMPLETE branch).
	Missing []string

	// Findings carries per-judge findings, each shaped
	// {"dimension": <name>, "severity": "critical|major|minor", ...}. The
	// contested-finding predicates (i) any critical, (ii) >=2 conflicting
	// severities on the same dimension, are evaluated over this slice.
	Findings []map[string]any
}

// IsBinding evaluates REQ-003's mechanical binding predicate. Returns
// (true, "") on the happy path (the workflow verdict is BINDING; the cold
// sync-auditor is NOT spawned). Returns (false, reason) when any fallback
// trigger fires — the orchestrator spawns the cold sync-auditor and treats ITS
// verdict as binding for that cycle.
//
// Fallback triggers (any one fires; order is detection-order, not precedence):
//
//  1. Verdict != "PASS" (covers INCOMPLETE and FAIL top-level verdicts).
//  2. Any dimension scored 0 (ZeroScored non-empty, OR a Scores entry at 0).
//  3. Any finding at "critical" severity (contested-finding predicate i).
//  4. Two or more findings with conflicting severities on the same dimension
//     (contested-finding predicate ii — e.g. one judge marks Functionality
//     "major" while another marks it "minor").
//
// Triggers 1 and 2 also surface via the workflow's top-level verdict (INCOMPLETE
// / FAIL), but IsBinding re-checks them defensively so a malformed verdict
// object cannot bypass the fallback.
func (v *FourDimVerdict) IsBinding() (bool, string) {
	if v == nil {
		return false, "nil verdict"
	}

	// Trigger 1: top-level verdict must be PASS.
	if !strings.EqualFold(strings.TrimSpace(v.Verdict), "PASS") {
		return false, "workflow verdict is not PASS (INCOMPLETE/FAIL) — cold sync-auditor fallback"
	}

	// Trigger 2a: explicit ZeroScored array (workflow FAIL branch).
	if len(v.ZeroScored) > 0 {
		return false, "must-pass dimension scored 0: " + strings.Join(v.ZeroScored, ", ")
	}
	// Trigger 2b: defensive — scan Scores for any 0 entry.
	for _, s := range v.Scores {
		if score, ok := toFloat(s["score"]); ok && score <= 0 {
			dim := strOr(s["dimension"], "?")
			return false, "must-pass dimension scored 0: " + dim
		}
	}
	// Trigger 2c: INCOMPLETE surfaces a Missing array — treat as fallback even
	// if Verdict were somehow "PASS" with missing judges (defensive).
	if len(v.Missing) > 0 {
		return false, "missing judge dimensions: " + strings.Join(v.Missing, ", ")
	}

	// Triggers 3 + 4: contested-finding predicates over the findings slice.
	byDim := make(map[string]map[string]bool) // dimension -> set of severities
	for _, f := range v.Findings {
		sev := strings.ToLower(strings.TrimSpace(strOr(f["severity"], "")))
		if sev == "" {
			continue
		}
		// Predicate (i): any critical finding.
		if sev == "critical" {
			return false, "contested finding: critical severity reported"
		}
		// Predicate (ii): >=2 conflicting severities on the same dimension.
		dim := strOr(f["dimension"], "")
		if dim == "" {
			continue
		}
		if byDim[dim] == nil {
			byDim[dim] = make(map[string]bool)
		}
		byDim[dim][sev] = true
		if len(byDim[dim]) >= 2 {
			return false, "contested finding: conflicting severities on dimension " + dim
		}
	}

	return true, ""
}

// toFloat extracts a float64 from an any (handles the JSON-number shapes
// float64 and int).
func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	}
	return 0, false
}

// strOr returns s as a string if it is one, else fallback.
func strOr(v any, fallback string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fallback
}
