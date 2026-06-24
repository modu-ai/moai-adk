package config

// === SPEC-CC2178-MODEL-POLICY-REPAIR-001 M3 (ResolveCycleType) ===
//
// cycle_type.go authors the ResolveCycleType symbol per plan.md §F.3.1 contract.
// This file was NEW at M3 — prior to this SPEC there was no symbol anywhere in
// internal/ mapping harness level to cycle_type (verified by grep at plan-phase).

// harness level constants for the 3-level harness system (minimal/standard/thorough).
// These mirror the harness Complexity Estimator levels resolved at
// internal/harness/router/router.go and the harness.yaml config.
const (
	harnessLevelMinimal  = "minimal"
	harnessLevelStandard = "standard"
	harnessLevelThorough = "thorough"
)

// cycle_type constants returned by ResolveCycleType. These mirror the existing
// DevelopmentMode type (type DevelopmentMode string in internal/models/) — the
// string values "tdd" / "ddd" / "autofix" are the canonical cycle_type tokens
// used across quality.yaml, manager-develop delegation, and the run-phase
// dispatch table.
const (
	cycleTypeTDD = "tdd"
	cycleTypeDDD = "ddd"
)

// ResolveCycleType determines the run-phase cycle_type (ddd | tdd | autofix)
// from the harness level, honoring an explicit quality.yaml
// constitution.development_mode pin.
//
// Precedence (highest first):
//  1. explicitPin (non-empty) — AG-01 backward-compat: an explicit
//     constitution.development_mode pin always wins, regardless of harness
//     level. This preserves existing pinned projects so that introducing
//     harness-based routing does not break them.
//  2. harnessLevel dispatch table (minimal→ddd, standard→tdd, thorough→tdd).
//  3. fallback "tdd" — the current global default; the function never returns
//     an empty string.
//
// The returned string is one of "tdd", "ddd", or "autofix" (string-typed, not
// a custom enum, to match the existing DevelopmentMode type in internal/models/
// which is already `type DevelopmentMode string`).
//
// AC bindings: AC-MPR-004 (minimal→ddd), AC-MPR-005 (thorough→tdd),
// AC-MPR-006 (explicit pin preserved), AC-MPR-014 (standard→tdd).
func ResolveCycleType(harnessLevel string, explicitPin string) string {
	// Precedence 1: explicit pin wins (AG-01 backward-compat).
	if explicitPin != "" {
		return explicitPin
	}

	// Precedence 2: harness-level dispatch table.
	switch harnessLevel {
	case harnessLevelMinimal:
		// REQ-MPR-004: minimal maps to lightweight ddd (characterization-test-first
		// DDD without full TDD RED-GREEN overhead). minimal does NOT mean "no
		// methodology" — it means DDD-lite.
		return cycleTypeDDD
	case harnessLevelStandard:
		// REQ-MPR-007 / AC-MPR-014: standard is the current default — tdd unchanged.
		return cycleTypeTDD
	case harnessLevelThorough:
		// REQ-MPR-005: critical features retain full TDD discipline.
		return cycleTypeTDD
	default:
		// Precedence 3: unknown or empty harness level — safe fallback to the
		// current global default. Never returns empty.
		return cycleTypeTDD
	}
}
