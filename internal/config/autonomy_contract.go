package config

import (
	"errors"
	"fmt"
	"strconv"
)

// autonomy_contract.go — the single read point for workflow.autonomy.*
// (SPEC-AUTONOMY-CONTRACT-001 REQ-CONTRACT-015). Distinct from autonomy.go,
// which reads the MOAI_AUTONOMY_TIER environment key; the two are unrelated.

// Closed value sets for workflow.autonomy.*.
const (
	AutonomyModeGuided   = "guided"
	AutonomyModeContract = "contract"

	AutonomySecondReviewRequired = "required"
	AutonomySecondReviewAdvisory = "advisory"
	AutonomySecondReviewOff      = "off"

	AutonomyDeciderHuman  = "human"
	AutonomyDeciderLLM    = "llm"
	AutonomyDeciderLLMJev = "llm+jev"
	// AutonomyDeciderJev is not a member of the decider set. It is named so
	// the resolver can report it as a configuration error rather than as an
	// ordinary out-of-set value.
	AutonomyDeciderJev = "jev"
)

// Full dotted key names used in warnings and in the decider error.
const (
	autonomyKeyMode             = "workflow.autonomy.mode"
	autonomyKeySecondReview     = "workflow.autonomy.contract.second_review"
	autonomyKeyDecider          = "workflow.autonomy.kickoff.decider"
	autonomyKeyJevMinConfidence = "workflow.autonomy.kickoff.jev_min_confidence"
)

// ErrKickoffDeciderJevSole is reported (never returned as a load failure) when
// workflow.autonomy.kickoff.decider is "jev": Jev is never a sole decider. The
// resolver applies no fallback for it — the effective decider stays "jev" — so
// only the receipt signing path, which reads this error, refuses.
var ErrKickoffDeciderJevSole = errors.New(autonomyKeyDecider + ": jev is never a sole decider (use llm+jev)")

// AutonomyBudget is the effective escalation budget default.
type AutonomyBudget struct {
	Turns        int
	Operations   int
	AuditRetries int
}

// AutonomySettings is the effective workflow.autonomy configuration.
type AutonomySettings struct {
	Mode         string
	BatchSign    bool
	SecondReview string
	PushDevelop  bool
	// Decider is the effective kickoff decider. DeciderDerived is true when
	// the key was absent or empty and the value came from the effective mode.
	Decider          string
	DeciderDerived   bool
	JevMinConfidence float64
	BudgetDefault    AutonomyBudget
	// JevEnabled mirrors workflow.jev.enabled for the signer; an llm+jev
	// decider with Jev disabled is not an error here (the fallback is the
	// receipt's to record).
	JevEnabled bool
	// Warnings names each key whose out-of-set value was replaced.
	Warnings []string
	// DeciderError is ErrKickoffDeciderJevSole when decider is "jev", else nil.
	DeciderError error
}

// ResolveAutonomy resolves workflow.autonomy.* to its effective values. It
// never fails: absent keys take the DefaultAutonomy* values, out-of-set values
// fall back toward the stricter value with a warning naming the key, and a
// "jev" decider is reported through DeciderError without a fallback.
//
// @MX:ANCHOR: [AUTO] single resolver for workflow.autonomy — every consumer reads effective values here
// @MX:REASON: callers are the contract CLI (sign / show / verify) and the A2/A3 escalation and kickoff paths; a second reader would let defaults, fallbacks, or the decider derivation drift apart
func ResolveAutonomy(wf WorkflowConfig) AutonomySettings {
	a := wf.Autonomy
	s := AutonomySettings{
		BatchSign:   a.Contract.BatchSign,
		PushDevelop: a.Contract.PushDevelop,
		JevEnabled:  wf.Jev.Enabled,
	}

	switch a.Mode {
	case "":
		s.Mode = DefaultAutonomyMode
	case AutonomyModeGuided, AutonomyModeContract:
		s.Mode = a.Mode
	default:
		s.Mode = DefaultAutonomyMode
		s.warn(autonomyKeyMode, a.Mode, DefaultAutonomyMode)
	}

	switch a.Contract.SecondReview {
	case "":
		s.SecondReview = DefaultAutonomySecondReview
	case AutonomySecondReviewRequired, AutonomySecondReviewAdvisory, AutonomySecondReviewOff:
		s.SecondReview = a.Contract.SecondReview
	default:
		s.SecondReview = DefaultAutonomySecondReview
		s.warn(autonomyKeySecondReview, a.Contract.SecondReview, DefaultAutonomySecondReview)
	}

	switch a.Kickoff.Decider {
	case "":
		s.DeciderDerived = true
		s.Decider = AutonomyDeciderHuman
		if s.Mode == AutonomyModeContract {
			s.Decider = AutonomyDeciderLLM
		}
	case AutonomyDeciderHuman, AutonomyDeciderLLM, AutonomyDeciderLLMJev:
		s.Decider = a.Kickoff.Decider
	case AutonomyDeciderJev:
		s.Decider = AutonomyDeciderJev
		s.DeciderError = ErrKickoffDeciderJevSole
	default:
		s.Decider = AutonomyDeciderHuman
		s.warn(autonomyKeyDecider, a.Kickoff.Decider, AutonomyDeciderHuman)
	}

	s.JevMinConfidence = DefaultAutonomyJevMinConfidence
	if c := a.Kickoff.JevMinConfidence; c != nil {
		if *c >= 0 && *c <= 1 {
			s.JevMinConfidence = *c
		} else {
			s.warn(autonomyKeyJevMinConfidence, strconv.FormatFloat(*c, 'g', -1, 64),
				strconv.FormatFloat(DefaultAutonomyJevMinConfidence, 'f', 2, 64))
		}
	}

	b := a.Escalation.BudgetDefault
	s.BudgetDefault = AutonomyBudget{
		Turns:        intOrDefault(b.Turns, DefaultAutonomyBudgetTurns),
		Operations:   intOrDefault(b.Operations, DefaultAutonomyBudgetOperations),
		AuditRetries: intOrDefault(b.AuditRetries, DefaultAutonomyBudgetAuditRetries),
	}
	return s
}

// warn records that key held got, which was replaced by used.
func (s *AutonomySettings) warn(key, got, used string) {
	s.Warnings = append(s.Warnings, fmt.Sprintf("%s: unrecognized value %q; using %q", key, got, used))
}

// intOrDefault returns *p, or def when the key was absent (nil).
func intOrDefault(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}
