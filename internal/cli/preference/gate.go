// Package preference — M5 sensitive-domain gate + recommendation strength
// (gate.go).
//
// This file implements REQ-ADM-014 / AC-ADM-014 [S2 Critical]: the
// sensitive-domain strength-reduction gate. When the orchestrator is operating
// in a sensitive domain (security review, one-off exploration query,
// cold-start, vulnerability triage, incident response, exploit analysis), the
// recommendation strength is reduced to Neutral — NO inferred-preference-based
// `(권장)` label is placed — AND the reduction is disclosed to the user.
//
// The gate also encodes the cold-start neutral branch (design.md §A.4: "세션
// 카운트 < 5 → neutral 강도") and the proficiency-modulated strength for
// non-sensitive, non-cold-start contexts (REQ-ADM-017):
//
//   - Expert   → Weak   (info-centric, autonomy-preserving)
//   - General  → Strong (reduce decision fatigue)
//   - ColdStart → Neutral (no inference yet)
//
// DecideStrength is PURE: it depends only on its three inputs. Callers pass
// the resolved proficiency (M5.2), the cold-start flag, and the domain.

package preference

import (
	"fmt"
	"strings"
)

// RecommendationStrength is the 3-value strength enum that governs how the
// orchestrator places the `(권장)` label on the user-facing question channel's
// options.
type RecommendationStrength int

const (
	// StrengthStrong: full `(권장)` label + transparent reason. Used for general
	// users in non-sensitive, non-cold-start domains (REQ-ADM-017).
	StrengthStrong RecommendationStrength = iota
	// StrengthWeak: info-centric — inferred preference is disclosed but the
	// `(권장)` label override is NOT placed. Used for experts in non-sensitive
	// domains (REQ-ADM-017 "약 추천 강도").
	StrengthWeak
	// StrengthNeutral: no preference-based recommendation at all. Used for
	// sensitive domains (REQ-ADM-014) and cold-start (design.md §A.4 gate).
	StrengthNeutral
)

// String renders the strength for audit logs.
func (s RecommendationStrength) String() string {
	switch s {
	case StrengthStrong:
		return "strong"
	case StrengthWeak:
		return "weak"
	case StrengthNeutral:
		return "neutral"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// sensitiveDomainTokens is the closed set of substrings that mark a domain as
// sensitive (REQ-ADM-014, design.md §B.1 "민감 도메인"). The match is
// case-insensitive and token-substring based so "Security-Review",
// "security_review", "vulnerability-scan" all match.
//
// The set is intentionally over-inclusive (per REQ-ADM-014 rationale: better
// to force neutral on a false-positive than to place a recommendation in a
// genuine security-review context).
var sensitiveDomainTokens = []string{
	"security",
	"cold-start",
	"cold_start",
	"one-off",
	"one_off",
	"vulnerab", // matches vulnerability / vulnerable
	"cve",
	"incident",
	"exploit",
	"pen-test",
	"pen_test",
	"breach",
}

// IsSensitiveDomain reports whether the given decision domain is sensitive
// (REQ-ADM-014, AC-ADM-014). A sensitive domain forces Neutral recommendation
// strength and a disclosure.
//
// The empty string is NOT sensitive (it represents an unclassified domain; the
// caller should treat it as non-sensitive and let the proficiency/cold-start
// branches decide strength).
//
// @MX:NOTE: [AUTO] IsSensitiveDomain — REQ-ADM-014 민감 도메인 강도 저하 게이트
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-014, AC-ADM-014
func IsSensitiveDomain(domain string) bool {
	if domain == "" {
		return false
	}
	lower := strings.ToLower(domain)
	for _, tok := range sensitiveDomainTokens {
		if strings.Contains(lower, tok) {
			return true
		}
	}
	return false
}

// sensitiveDomainDisclosure is the canonical disclosure string the gate emits
// when it reduces strength due to a sensitive domain (AC-ADM-014 evidence:
// '"personalization reduced for sensitive domain" 공개 로그'). It is returned
// as the disclosure value from DecideStrength and the orchestrator appends it
// to the recommendation option description.
const sensitiveDomainDisclosure = "personalization reduced for sensitive domain"

// coldStartDisclosure is the disclosure for the cold-start neutral branch. It
// is shorter than the sensitive-domain one because the cold-start condition is
// not a domain property — it is a data-availability condition ("we don't have
// enough observations yet"). Honesty: the orchestrator has no inferred
// preference to disclose, so the disclosure says so.
const coldStartDisclosure = "based on static default, N observations needed for personalization"

// DecideStrength resolves the recommendation strength for a given domain +
// proficiency + cold-start flag (REQ-ADM-014 + REQ-ADM-017 + design.md §A.4).
//
// Returns (strength, disclosure). The disclosure is a fragment the orchestrator
// appends to the recommendation option description. It is "" (empty) when no
// disclosure is warranted (the Strong/Weak non-sensitive paths).
//
// Decision order:
//
//  1. Sensitive domain → Neutral + sensitiveDomainDisclosure (overrides all).
//  2. Cold-start flag OR ProficiencyColdStart → Neutral + coldStartDisclosure.
//  3. ProficiencyExpert → Weak (info-centric).
//  4. ProficiencyGeneral → Strong.
//  5. Unknown proficiency → Neutral (defensive — treat as cold-start).
//
// The sensitive-domain check is FIRST so it overrides a cold-start flag or an
// expert proficiency — a security-review context is neutral regardless of who
// the user is (AC-ADM-014 edge case: "민감 도메인 + 전문가 사용자 동시 → 민감
// 도메인 게이트가 우선").
//
// @MX:ANCHOR: [AUTO] DecideStrength — 추천 강도 결정 진입점 (sensitive/cold-start/proficiency 분기)
// @MX:REASON: fan_in >= 3 예상 (orchestrator 발화 직전 인라인 결정, AC-ADM-003 p95 ≤ 10ms); 3-축 분기가 핵심 invariant
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-014, REQ-ADM-017, design.md §A.4
func DecideStrength(domain string, prof Proficiency, coldStart bool) (RecommendationStrength, string) {
	// (1) Sensitive domain gate — overrides everything (AC-ADM-014 edge case).
	if IsSensitiveDomain(domain) {
		return StrengthNeutral, sensitiveDomainDisclosure
	}
	// (2) Cold-start branch — explicit flag OR inferred ColdStart proficiency.
	if coldStart || prof == ProficiencyColdStart {
		return StrengthNeutral, coldStartDisclosure
	}
	// (3) Expert → Weak (info-centric).
	if prof == ProficiencyExpert {
		return StrengthWeak, ""
	}
	// (4) General → Strong.
	if prof == ProficiencyGeneral {
		return StrengthStrong, ""
	}
	// (5) Unknown proficiency — defensive neutral (should not happen with the
	// closed Proficiency enum, but keeps the function total).
	return StrengthNeutral, coldStartDisclosure
}
