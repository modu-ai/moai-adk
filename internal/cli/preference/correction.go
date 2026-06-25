// Package preference — M5 correction loop (correction.go).
//
// This file implements REQ-ADM-016 / AC-ADM-016 [S2 Critical]: the correction
// loop. When the user corrects an inferred preference ("이 추론이 틀리다"), the
// corrected fact is immediately upserted with Confidence: ConfidenceObserved
// AND the prior inferred entry's weight is reduced (demoted to archival as an
// audit record).
//
// The correction loop is the "reactive autonomy buffer" named in design.md
// §B.1: it lets the user retroactively fix a misfired inference without a full
// preference-reset. It is distinct from the session-level recovery toggle
// (REQ-ADM-013, toggle.go): the toggle disables personalization for the whole
// session; the correction loop fixes a single entry while leaving the rest of
// the personalization layer active.

package preference

import (
	"fmt"
	"time"
)

// correctionWeightReduction is the multiplicative factor applied to a prior
// inferred entry's weight when it is superseded by a correction. The reduced
// weight is recorded in the archival audit copy so post-hoc analysis can see
// "this inference was corrected; its effective weight at correction time was
// X". The factor is conservative (0.25 — a corrected inference retains a
// quarter of its prior weight) so a re-promotion path (if ever added) has a
// signal to work with, but the entry is clearly de-prioritized relative to
// the new observed entry (weight 1.0).
const correctionWeightReduction = 0.25

// correctionSourceCitationPrefix is the SourceCitation prefix for entries
// produced by the correction loop. It records the correction provenance so
// REQ-ADM-018 (verification-claim-integrity) is satisfied: a recommendation
// grounded in a corrected entry can cite its provenance back to the correction
// event.
const correctionSourceCitationPrefix = "correction"

// CorrectInferred applies the user's correction to an inferred preference
// (REQ-ADM-016, AC-ADM-016).
//
// Behavior:
//
//  1. Validate the arguments (non-empty domain/key/fact).
//  2. Look up the prior entry for (domain, decisionKey) via the cascade.
//  3. If the prior entry exists AND has Confidence == ConfidenceInferred:
//     reduce its weight (Weight *= correctionWeightReduction) and write the
//     reduced-weight copy to archival as a correction audit record. The prior
//     inferred entry is then superseded by the upsert in step 4.
//  4. Upsert the corrected fact as a new ConfidenceObserved entry with
//     SourceCitation recording the correction provenance (REQ-ADM-018) and
//     LastUsed/ValidTime = now.
//
// If the prior entry does NOT exist, or has Confidence == ConfidenceObserved,
// step 3 is skipped — there is no inference to demote, and the upsert in step 4
// is a clean observed write (idempotent for observed→observed transitions).
//
// `now` is injected (not time.Now()) so the correction is deterministic in
// tests and the audit record carries a reproducible timestamp.
//
// @MX:NOTE: [AUTO] CorrectInferred — REQ-ADM-016 정정 루프 (inferred→observed 전환 + weight 감소)
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-016, AC-ADM-016, REQ-ADM-018
func CorrectInferred(store Store, domain, decisionKey, correctedFact string, now time.Time) error {
	if domain == "" || decisionKey == "" {
		return fmt.Errorf("%w: CorrectInferred requires non-empty domain and decision_key", ErrInvalidEntry)
	}
	if correctedFact == "" {
		return fmt.Errorf("%w: CorrectInferred requires non-empty correctedFact", ErrInvalidEntry)
	}
	if store == nil {
		return fmt.Errorf("preference: CorrectInferred received nil store")
	}

	// (2) Look up the prior entry.
	prior, found, _, err := store.Get(domain, decisionKey)
	if err != nil {
		return fmt.Errorf("preference: CorrectInferred lookup prior: %w", err)
	}

	// (3) If the prior was inferred, demote it: reduce weight, archive as audit.
	if found && prior.Confidence == ConfidenceInferred {
		fs, ok := store.(*fileStore)
		if ok {
			demoted := prior
			demoted.Weight *= correctionWeightReduction
			if err := fs.writeArchivalEntry(domain, decisionKey, demoted); err != nil {
				// Advisory: the audit-record write is best-effort. If archival
				// is unwritable, the correction still proceeds — the observed
				// upsert is the load-bearing half. We surface the error because
				// a silent archival failure would lose the audit trail; the
				// caller decides whether to retry or accept.
				return fmt.Errorf("preference: CorrectInferred archive prior inferred: %w", err)
			}
		}
		// If the store is NOT a *fileStore (e.g. a future in-memory test fake),
		// the demotion step is skipped — the observed upsert below still
		// supersedes the inferred entry at the Store.Upsert level.
	}

	// (4) Upsert the corrected observed entry. The new entry REPLACES the
	// prior entry at the (domain, decisionKey) key (Upsert semantics,
	// REQ-ADM-001). SourceCitation records the correction provenance.
	corrected := Entry{
		Fact:           correctedFact,
		Domain:         domain,
		DecisionKey:    decisionKey,
		Scope:          prior.Scope, // inherit scope; default to stable if no prior
		Confidence:     ConfidenceObserved,
		SourceCitation: fmt.Sprintf("%s at=%s", correctionSourceCitationPrefix, now.UTC().Format(time.RFC3339)),
		ValidTime:      now,
		LastUsed:       now,
		Weight:         1.0, // fresh-entry baseline; the corrected fact is the strongest signal
	}
	if !found {
		// No prior entry: default the scope to stable (corrections usually
		// reflect a durable preference the user bothered to state explicitly).
		corrected.Scope = ScopeStable
	}
	if corrected.Scope == "" {
		corrected.Scope = ScopeStable
	}
	if err := store.Upsert(domain, decisionKey, corrected); err != nil {
		return fmt.Errorf("preference: CorrectInferred upsert corrected: %w", err)
	}
	return nil
}
