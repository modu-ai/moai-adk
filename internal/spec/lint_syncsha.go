package spec

// lint_syncsha.go — SyncSHASlotFormatRule (SPEC-SYNC-SHA-SLOT-FORMAT-001 M3,
// REQ-SSF-004 / REQ-SSF-005 / REQ-SSF-006, card t299). A `warning`-severity
// SPEC-artifact lint rule that flags a `sync_commit_sha` slot occupied by a
// value token that is neither a commit SHA nor a recognized backfill
// placeholder.
//
// WHAT THIS RULE IS FOR, AND WHAT IT IS NOT FOR. The motivating instance —
// SPEC-BACKLOG-LOCK-BUDGET-001 holding `pending-backfill-sync` forever — is
// repaired by the CLOSER-side inversion (closer.go `needsSHABackfill`), and this
// rule will NEVER flag that slot. It cannot: a recognized placeholder is a
// sanctioned intermediate state and REQ-SSF-005 requires silence on it. The two
// guards answer different questions:
//
//	closer: is this slot still owed a real SHA?                    → yes, repair it
//	lint:   is this slot occupied by something that is neither a
//	        SHA nor a sanctioned placeholder?                      → silent here
//
// This rule is the BACKSTOP for the class the closer cannot help — a value that
// is neither, so no close will ever be triggered to repair it and no reader can
// tell what the slot was supposed to say. spec.md §B.6 measures that class at
// five slots, all of them in `completed` SPECs.
//
// SEVERITY IS `warning`, NEVER `error`, and the code is DELIBERATELY ABSENT from
// eraDemotableCodes (spec.md §D.3, REQ-SSF-007, AC-SSF-010). The two decisions
// are one argument:
//
//   - `applyEraDemotion` (lint.go:288) marks every warning of a demoted document
//     advisory, and `Report.HasErrors` (lint.go:61) escalates under --strict only
//     for a NON-advisory warning. All five corpus findings sit in `completed`
//     SPECs, which `terminalStatusEnum` shelters, so this rule contributes
//     nothing to the strict exit status on the corpus as it stands.
//   - `eraDemotableCodes` demotes ERRORS (lint.go:284 gates on
//     `f.Severity == SeverityError`). For a warning the entry would be INERT, and
//     an inert entry in a policy map reads to a later maintainer as intent that
//     was never meant. MovingRefUnpinnedRule records the same choice for the same
//     reason.
//
// An `error` severity would invert both: with the code absent from
// eraDemotableCodes an error has NO shelter, so five closed SPECs' history would
// enter the strict path and `lint.skip` becomes the rational response — the
// outcome the rule exists to prevent.
//
// [HARD] REQ-SSF-007's justification is CONDITIONAL on the severity staying
// `warning` at the Finding level. If a later change promotes at that level
// rather than at `Report.HasErrors`, eraDemotableCodes becomes the only
// remaining shelter and the requirement inverts. That is an operator decision,
// not an implementer's: stop and report rather than adding the entry.
//
// The rule reads the SPEC's own SIBLING `progress.md`. `SPECDoc.Body` carries
// spec.md alone, so a body-only rule would see none of the corpus — the field
// lives exclusively in progress.md §E.4. MovingRefUnpinnedRule is the precedent
// for a rule reaching past the single parsed document.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// syncSHALinePrefix is the line shape of §D.1: the field at column 0. Anchoring
// at the line start (rather than allowing indentation or a list marker) matches
// the classifier the §B.6 prediction was derived with, so AC-SSF-006's measured
// count is comparable to the predicted one rather than merely similar to it.
const syncSHALinePrefix = "sync_commit_sha:"

// canonicalSyncSHAPlaceholder is the spelling the finding message names. It is
// the schema doctrine's own prescription (`spec-frontmatter-schema.md` §D3
// `pending-backfill-*`) written down precisely, and the single most-used
// spelling in the corpus; the suffixed forms stay admitted.
const canonicalSyncSHAPlaceholder = "pending-backfill"

// SyncSHASlotFormatRule implements the Rule interface for REQ-SSF-004.
type SyncSHASlotFormatRule struct{}

// Code returns the stable lint finding code.
func (r *SyncSHASlotFormatRule) Code() string { return "SyncSHASlotFormat" }

// Check scans the SPEC's sibling progress.md for malformed `sync_commit_sha`
// slots.
func (r *SyncSHASlotFormatRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if doc == nil || doc.Path == "" {
		return nil
	}
	path := filepath.Join(filepath.Dir(doc.Path), "progress.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // no progress.md — nothing to check, not a finding
	}

	var findings []Finding
	for i, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, syncSHALinePrefix) {
			continue
		}
		// REQ-SSF-006: the token is the first whitespace-delimited run, so a
		// trailing annotation affects neither decision. 105 of 346 corpus values
		// carry one, 99 of them on otherwise-conforming SHAs — a grammar
		// rejecting annotations would flag nearly a third of a healthy corpus.
		token := syncSHAValueToken(strings.TrimPrefix(line, syncSHALinePrefix))
		if isCommitSHAToken(token) {
			continue
		}
		// REQ-SSF-005: the D3 backfill window is a sanctioned intermediate
		// state, not a defect. A commit cannot cite its own hash, so a
		// placeholder in the phase commit followed by a backfill in a later
		// commit is the only workable procedure; a rule that flagged it would
		// forbid the only way the field can ever be populated.
		if isSyncSHAPlaceholder(token) {
			continue
		}
		findings = append(findings, Finding{
			File:     path,
			Line:     i + 1,
			Severity: SeverityWarning,
			Code:     r.Code(),
			Message: fmt.Sprintf(
				"`sync_commit_sha` holds %q, which is neither a commit SHA nor a recognized backfill placeholder, "+
					"so no close will be triggered to repair it and no reader can tell what the slot was meant to say. "+
					"Write the commit SHA that carried the sync phase, or — while the backfill is genuinely still owed — "+
					"the canonical placeholder `%s` (a `-`-suffixed member of that family, e.g. `%s-sync`, is equally admitted). "+
					"A trailing annotation after the value is ignored and needs no removal.",
				token, canonicalSyncSHAPlaceholder, canonicalSyncSHAPlaceholder),
		})
	}
	return findings
}
