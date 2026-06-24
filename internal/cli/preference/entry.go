// Package preference implements the AskUserQuestion user-decision memory layer
// for SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 (component C1, REQ-ADM-001~004).
//
// The store is keyed by (domain, decision_key) and upserts REPLACE rather than
// append (REQ-ADM-001). It lives under
//   ~/.claude/projects/{slug}/memory/user_decisions/
// physically separate from the technical-lesson feedback namespace
// (memory/feedback_*.md + MEMORY.md) so user-decision facts and engineering
// lessons remain independently queryable (REQ-ADM-002).
//
// Three tiers cascade on retrieval (REQ-ADM-004):
//   - core (core.yaml): always-loaded user profile, capped at MaxCoreBytes
//     (NFR-ADM-002). Frequent stable facts live here.
//   - recall (recall.jsonl): recent-session facts.
//   - archival (archival/): full-search targets.
// On a core hit the cascade MUST NOT access recall or archival.
//
// All writes are atomic: recall/core writes go through a temp file in the same
// directory followed by os.Rename, so a SIGKILL mid-write cannot leave a
// partially-upserted JSON record (AC-ADM-001 edge case, design.md §C/KI-3).
package preference

import (
	"errors"
	"fmt"
	"time"
)

// Scope classifies how stable a preference is. REQ-ADM-003 schema field.
//
// - ScopeStable: durable user trait (e.g. "prefers Go backend"). Exempt from
//   pure time-decay in the M4 decay policy (REQ-ADM-011).
// - ScopeTransient: situational preference (e.g. "verbose logs this session").
//   Subject to power-law decay + 28-day TTL (REQ-ADM-011, REQ-ADM-012).
type Scope string

const (
	ScopeStable    Scope = "stable"
	ScopeTransient Scope = "transient"
)

// validScope is the closed set of acceptable Scope values.
var validScope = map[Scope]bool{
	ScopeStable:    true,
	ScopeTransient: true,
}

// Confidence records whether a fact was directly observed or inferred.
// REQ-ADM-003 schema field; REQ-ADM-018 (verification-claim-integrity) requires
// recommendations that claim to reflect a past user decision to map to an
// observed (or disclosed-basis inferred) entry.
type Confidence string

const (
	// ConfidenceObserved: the fact was directly captured from an explicit user
	// action (AskUserQuestion tool_result, explicit correction, etc.).
	ConfidenceObserved Confidence = "observed"
	// ConfidenceInferred: the fact was inferred by the orchestrator. Inferences
	// MUST be disclosed with a basis (REQ-ADM-018) and are subject to the
	// correction loop (REQ-ADM-016).
	ConfidenceInferred Confidence = "inferred"
)

// validConfidence is the closed set of acceptable Confidence values.
var validConfidence = map[Confidence]bool{
	ConfidenceObserved: true,
	ConfidenceInferred: true,
}

// Entry is the canonical 7-field preference memory record (REQ-ADM-003).
//
// The lookup key is the (Domain, DecisionKey) pair; DecisionKey is part of the
// record so a deserialized entry is self-describing. The remaining fields are:
//
//   - Fact: the preference statement itself (e.g. "prefers Go backend").
//   - SourceCitation: provenance — file:line, session_id, or tool_result id.
//   - ValidTime: when the fact was first established.
//   - LastUsed: when the fact was last referenced; refreshed on reuse.
//   - Scope: ScopeStable | ScopeTransient.
//   - Confidence: ConfidenceObserved | ConfidenceInferred.
type Entry struct {
	Fact           string      `yaml:"fact"           json:"fact"`
	SourceCitation string      `yaml:"source_citation" json:"source_citation"`
	ValidTime      time.Time   `yaml:"valid_time"     json:"valid_time"`
	LastUsed       time.Time   `yaml:"last_used"      json:"last_used"`
	Scope          Scope       `yaml:"scope"          json:"scope"`
	Domain         string      `yaml:"domain"         json:"domain"`
	DecisionKey    string      `yaml:"decision_key"   json:"decision_key"`
	Confidence     Confidence  `yaml:"confidence"     json:"confidence"`

	// Weight is an internal ranking score derived from recency + confidence.
	// It is not part of the REQ-ADM-003 schema surfaced to callers, but it
	// drives core-tier eviction (lowest-weight demotes first) and is persisted
	// so demotion decisions are reproducible across sessions.
	Weight float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}

// ErrInvalidEntry is returned by Entry.Validate / Store.Upsert when an entry is
// missing a required field or carries an invalid enum value.
var ErrInvalidEntry = errors.New("preference: invalid entry")

// Validate verifies the entry carries all 7 required fields with valid enum
// values. It is the AC-ADM-003 schema guard.
func (e Entry) Validate() error {
	if e.Fact == "" {
		return fmt.Errorf("%w: fact is empty", ErrInvalidEntry)
	}
	if e.SourceCitation == "" {
		return fmt.Errorf("%w: source_citation is empty", ErrInvalidEntry)
	}
	if e.ValidTime.IsZero() {
		return fmt.Errorf("%w: valid_time is zero", ErrInvalidEntry)
	}
	if e.LastUsed.IsZero() {
		return fmt.Errorf("%w: last_used is zero", ErrInvalidEntry)
	}
	if e.Domain == "" {
		return fmt.Errorf("%w: domain is empty", ErrInvalidEntry)
	}
	if e.DecisionKey == "" {
		return fmt.Errorf("%w: decision_key is empty", ErrInvalidEntry)
	}
	if !validScope[e.Scope] {
		return fmt.Errorf("%w: scope %q is not stable|transient", ErrInvalidEntry, e.Scope)
	}
	if !validConfidence[e.Confidence] {
		return fmt.Errorf("%w: confidence %q is not observed|inferred", ErrInvalidEntry, e.Confidence)
	}
	return nil
}
