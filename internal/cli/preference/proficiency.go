// Package preference — M5 proficiency inference (proficiency.go).
//
// This file implements the REQ-ADM-017 / AC-ADM-017 [S2 Critical] proficiency
// estimator (design.md §A.4). Proficiency modulates recommendation strength
// via gate.go's DecideStrength: experts get weak recommendations (info-centric,
// autonomy-preserving); general users get strong recommendations (reduce
// decision fatigue); cold-start gets neutral (no inference yet).
//
// The initial implementation uses session count alone (design.md §A.4 option
// (a) — "초기 구현이 단순하고 cold-start에서 작동"). Decision-consistency (b)
// and explicit self-assessment (c) are deferred to the "complete" tier.
//
// Thresholds are package constants. An OPTIONAL preference.yaml override path
// is provided (design.md §A.4: "임계값은 .moai/config/sections/preference.yaml
// 신규에서 조정 가능") via LoadProficiencyThresholds. When the config file is
// absent or unparseable, the constants are used (config fallback — a missing
// preference.yaml MUST NOT break proficiency inference).

package preference

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Proficiency classifies how experienced the user is, per design.md §A.4.
// It modulates recommendation strength via DecideStrength (gate.go).
type Proficiency int

const (
	// ProficiencyColdStart: sessionCount < ExpertThresholdFloor (5 by default).
	// Neutral recommendation strength — REQ-ADM-014 cold-start gate.
	ProficiencyColdStart Proficiency = iota
	// ProficiencyGeneral: 5..19 sessions. Strong recommendation strength.
	ProficiencyGeneral
	// ProficiencyExpert: >= 20 sessions. Weak recommendation (info-centric).
	ProficiencyExpert
)

// String renders the proficiency for audit logs (AC-ADM-017 "숙련도 추정 로그").
// These are NOT user-facing labels; they appear in decision logs.
func (p Proficiency) String() string {
	switch p {
	case ProficiencyColdStart:
		return "cold_start"
	case ProficiencyGeneral:
		return "general"
	case ProficiencyExpert:
		return "expert"
	default:
		return fmt.Sprintf("unknown(%d)", int(p))
	}
}

// Default proficiency thresholds (design.md §A.4). These are the package-level
// single source of truth; LoadProficiencyThresholds MAY override them via
// preference.yaml but the defaults are used whenever that file is absent or
// unparseable.
const (
	// ProficiencyThresholdExpert is the session count at/above which the user
	// is classified Expert (design.md §A.4: "세션 카운트 ≥ 20 → 전문가").
	ProficiencyThresholdExpert = 20
	// ProficiencyThresholdFloor is the session count at/above which the user
	// leaves ColdStart and enters General (design.md §A.4: "세션 카운트 < 5
	// (초기) → neutral 강도" — so >= 5 is General).
	ProficiencyThresholdFloor = 5
)

// EstimateProficiency infers the user's proficiency from the session count
// (REQ-ADM-017, AC-ADM-017, design.md §A.4).
//
//   - count >= ProficiencyThresholdExpert (20) → Expert
//   - ProficiencyThresholdFloor (5) <= count <= 19 → General
//   - count < ProficiencyThresholdFloor (5) → ColdStart
//
// A negative count (defensive — fresh install or corrupted counter) is clamped
// to ColdStart rather than panicking.
//
// EstimateProficiency uses the package-level default thresholds. Callers that
// need config-overridden thresholds should call EstimateProficiencyWithThresholds
// after loading overrides via LoadProficiencyThresholds.
//
// @MX:NOTE: [AUTO] EstimateProficiency — REQ-ADM-017 숙련도 추정 (세션 카운트 단일 신호)
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-017, design.md §A.4
func EstimateProficiency(sessionCount int) Proficiency {
	return EstimateProficiencyWithThresholds(sessionCount, ProficiencyThresholdExpert, ProficiencyThresholdFloor)
}

// EstimateProficiencyWithThresholds is the config-injectable form. It exists so
// a preference.yaml override can adjust the thresholds without recompiling.
// Production callers pass the package defaults; tests + config-loaded callers
// pass the resolved thresholds.
//
// expert and floor follow the same semantics as ProficiencyThresholdExpert and
// ProficiencyThresholdFloor: a count >= expert is Expert; floor <= count < expert
// is General; count < floor is ColdStart. If expert <= floor (misconfiguration),
// the function treats the band as empty and falls back to the ColdStart/Expert
// boundary only (count >= expert → Expert, else ColdStart) so a bad config does
// not produce a silent General band that never triggers.
func EstimateProficiencyWithThresholds(sessionCount, expert, floor int) Proficiency {
	if sessionCount < 0 {
		// Defensive clamp — a negative count is treated as 0 (ColdStart).
		sessionCount = 0
	}
	// Normalize inverted/invalid thresholds to the defaults so a corrupt
	// preference.yaml cannot produce undefined behavior.
	if expert <= 0 {
		expert = ProficiencyThresholdExpert
	}
	if floor <= 0 || floor > expert {
		floor = ProficiencyThresholdFloor
	}
	if expert <= floor {
		// Degenerate band — collapse to ColdStart/Expert boundary.
		if sessionCount >= expert {
			return ProficiencyExpert
		}
		return ProficiencyColdStart
	}
	if sessionCount >= expert {
		return ProficiencyExpert
	}
	if sessionCount >= floor {
		return ProficiencyGeneral
	}
	return ProficiencyColdStart
}

// preferenceConfigFileName is the optional config file (design.md §A.4). It is
// read by LoadProficiencyThresholds from the project's .moai/config/sections/
// dir. Absent or unparseable → use defaults.
const preferenceConfigFileName = "preference.yaml"

// ProficiencyThresholds holds optional config-overridden proficiency thresholds.
// Zero values mean "use the package default".
type ProficiencyThresholds struct {
	Expert int
	Floor  int
}

// LoadProficiencyThresholds reads the optional preference.yaml from
// projectRoot/.moai/config/sections/preference.yaml and extracts the proficiency
// threshold overrides. Missing file, missing keys, or parse errors all fall
// back to the zero value (caller passes the package defaults). This is the
// config-fallback contract — a missing preference.yaml MUST NOT break
// proficiency inference.
//
// The YAML schema (minimal, only what M5 needs):
//
//	preference:
//	  proficiency:
//	    expert_sessions: 20   # optional
//	    floor_sessions: 5     # optional
//
// Parsing is intentionally minimal (line-based, not a full YAML parser) so the
// preference package does not import a YAML library — the package currently
// serializes via encoding/json + manual YAML emit, and this reader keeps that
// discipline.
func LoadProficiencyThresholds(projectRoot string) ProficiencyThresholds {
	path := filepath.Join(projectRoot, ".moai", "config", "sections", preferenceConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		// Missing file → zero values → caller uses defaults.
		return ProficiencyThresholds{}
	}
	return parseProficiencyThresholds(string(data))
}

// parseProficiencyThresholds is the pure line-based parser extracted for
// testability. It tolerates arbitrary YAML nesting by scanning for the two
// known leaf keys (`expert_sessions:` and `floor_sessions:`) anywhere in the
// file. A non-integer value is ignored (zero value retained).
func parseProficiencyThresholds(raw string) ProficiencyThresholds {
	out := ProficiencyThresholds{}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Match "expert_sessions: <int>" or "floor_sessions: <int>".
		if v, ok := parseThresholdKV(trimmed, "expert_sessions:"); ok {
			out.Expert = v
			continue
		}
		if v, ok := parseThresholdKV(trimmed, "floor_sessions:"); ok {
			out.Floor = v
		}
	}
	return out
}

// parseThresholdKV extracts the integer after the given key prefix from a
// trimmed YAML leaf line. Returns (value, true) on a clean parse; (0, false)
// otherwise.
func parseThresholdKV(line, key string) (int, bool) {
	if !strings.HasPrefix(line, key) {
		return 0, false
	}
	raw := strings.TrimSpace(strings.TrimPrefix(line, key))
	// Strip inline YAML comments ("...  # note" → "...").
	if hash := strings.Index(raw, "#"); hash >= 0 {
		raw = strings.TrimSpace(raw[:hash])
	}
	raw = strings.Trim(raw, `"'`)
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}
