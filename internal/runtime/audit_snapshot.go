package runtime

// SPEC-AUDIT-SNAPSHOT-001 (A2): per-tier plan-audit skip-eligible score
// thresholds. These mirror the per-tier PASS thresholds canonically defined in
// .claude/rules/moai/workflow/spec-workflow.md § SPEC Complexity Tier. The
// retired flat `score >= 0.90` skip predicate is replaced by these per-tier
// values: a SPEC whose plan-phase audit verdict legitimately PASSED is
// skip-eligible by default.
//
// The doctrinal SSOT is spec-workflow.md; these constants are the Go codification
// so the predicate is mechanically testable (per AC-AUDIT-SNAPSHOT-002's
// "table-driven unit test" requirement) rather than left as orchestrator-side
// prose judgment.
const (
	// SkipThresholdTierS is the Tier S (Simple) plan-audit PASS threshold.
	SkipThresholdTierS = 0.75
	// SkipThresholdTierM is the Tier M (Medium) plan-audit PASS threshold.
	SkipThresholdTierM = 0.80
	// SkipThresholdTierL is the Tier L (Large) plan-audit PASS threshold. An
	// unrecognized/empty tier falls back to this strict value (per
	// spec-frontmatter-schema.md § Optional Fields, a tier-absent SPEC is
	// treated as Tier L for backward compat).
	SkipThresholdTierL = 0.85
)

// @MX:NOTE: [AUTO] SPEC-AUDIT-SNAPSHOT-001 A2 의 Go 체현. doctrinal SSOT는
// spec-workflow.md § SPEC Complexity Tier (S 0.75 / M 0.80 / L 0.85); 이 상수들이
// 폐기된 flat `score >= 0.90` predicate를 대체한다. 임계값을 변경하려면 두 표면을 함께.
//
// SkipEligibleByScore evaluates plan-audit skip-policy condition 2 (overall
// score threshold) per SPEC-AUDIT-SNAPSHOT-001 A2: the threshold equals the
// SPEC's per-tier PASS threshold (S 0.75 / M 0.80 / L 0.85), NOT the retired
// flat 0.90. Returns true when the score clears the tier's threshold.
//
// An empty or unrecognized tier resolves to Tier L (the strictest), matching
// the spec-frontmatter-schema.md § Optional Fields backward-compat rule
// ("when tier: is absent, the SPEC is treated as Tier L"). This never silently
// permisses an unknown tier — strictness is the safe default.
func SkipEligibleByScore(tier string, score float64) bool {
	threshold := tierSkipThreshold(tier)
	return score >= threshold
}

// tierSkipThreshold resolves the per-tier PASS threshold. Unknown/empty tier →
// Tier L (strictest), per the backward-compat rule.
func tierSkipThreshold(tier string) float64 {
	switch tier {
	case "S":
		return SkipThresholdTierS
	case "M":
		return SkipThresholdTierM
	case "L":
		return SkipThresholdTierL
	default:
		// Empty / unrecognized tier → strict (Tier L). Never permissive.
		return SkipThresholdTierL
	}
}
