package spec

// syncsha.go — the `sync_commit_sha` slot's value grammar and the ONE shared
// predicate both guards derive from (SPEC-SYNC-SHA-SLOT-FORMAT-001 M1,
// REQ-SSF-001/REQ-SSF-002, card t299).
//
// WHY ONE PREDICATE. Four readers of this field already exist and they hold
// three different notions of "a value" (spec.md §B.4): era classification and
// the drift audit share `cleanFieldValue`; the closer normalizes with that and
// then applies its OWN four-value allowlist; the Epic status renderer carries a
// fifth independent regex. Every one of those was a fresh test written at a
// fresh site. A sixth would be the same defect again, so the closer's backfill
// decision (closer.go `needsSHABackfill`) and the lint rule's finding decision
// (lint_syncsha.go) both call `isCommitSHAToken` here rather than re-deriving
// the hex test locally. AC-SSF-007 is what holds that property.
//
// THE GRAMMAR (spec.md §D.1), verbatim:
//
//	line        := "sync_commit_sha:" WS* value
//	value       := QUOTE? token QUOTE? ( WS+ annotation )?
//	token       := SHA | PLACEHOLDER
//	SHA         := [0-9a-fA-F]{7,40}
//	PLACEHOLDER := "pending-backfill" ( "-" [A-Za-z0-9-]+ )?
//
// THE TOKEN IS THE FIRST WHITESPACE-DELIMITED RUN. This is chosen over
// enumerating annotation separators (`#`, ` — `, ` (…)`) because all three forms
// occur in the corpus — 105 of 346 values carry an annotation, 99 of them on
// otherwise-conforming SHA values — and a separator enumeration would have to
// grow every time an author picks a fourth. A first-token rule admits all of
// them without naming any.
//
// QUOTE HANDLING IS ON THE TOKEN, NOT THE LINE. This is the narrow repair of the
// artifact spec.md §B.4 measured live: `cleanFieldValue` trims quotes from the
// ends of the WHOLE string, so `"pending-backfill-sync"   # D3 …` yields the
// token `pending-backfill-sync"   # D3 …` — a stray quote plus a prose sentence
// published by the drift report as if it were a SHA. Splitting first and
// stripping second yields `pending-backfill-sync`.
//
// KNOWN LIMITATIONS, accepted rather than defended against (asserted as known
// behavior in syncsha_test.go):
//
//	L1 — a 7-40 character all-hex English word (`defaced`) parses as a SHA. The
//	     alternative is a dictionary: the false negative costs one unflagged
//	     slot, a dictionary would cost a maintenance surface.
//	L2 — the predicate reads SHAPE, never reachability. It cannot tell a SHA
//	     naming a real commit from seven plausible hex characters.
//
// This file deliberately does NOT touch `cleanFieldValue` (REQ-SSF-008,
// AC-SSF-009): the era heuristics H-3/H-4 depend on its current
// `(this commit)`-as-value behavior, and changing the normalizer would move an
// unmeasured set of SPECs between eras.

import (
	"regexp"
	"strings"
	"unicode"
)

// commitSHATokenPattern is SHA of §D.1. Anchored at both ends: the subject is a
// TOKEN already split off the value, so a prefix match would accept
// `a6bbbf82b#annot` as a SHA and quietly re-open the annotation hole the token
// split exists to close.
var commitSHATokenPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

// syncSHAPlaceholderPattern is PLACEHOLDER of §D.1 — the canonical backfill
// spelling the schema doctrine already prescribes
// (`spec-frontmatter-schema.md` § SHA placeholder backfill exemption, "D3").
//
// Case handling, decided explicitly because the corpus forces it: the PREFIX is
// a lowercase literal and is matched case-SENSITIVELY, while the suffix class
// `[A-Za-z0-9-]+` is case-insensitive by construction. That is what admits the
// corpus's `pending-backfill-SYNC` (SPEC-AUDIT-SNAPSHOT-001) without widening the
// exemption to a `Pending-Backfill` spelling no corpus slot actually uses.
//
// The suffixed family is admitted deliberately: 28 distinct spellings exist in
// the corpus today (spec.md §B.3), and refusing them would turn the 24
// `pending-backfill-sync` occurrences alone into findings on the first run.
var syncSHAPlaceholderPattern = regexp.MustCompile(`^pending-backfill(-[A-Za-z0-9-]+)?$`)

// syncSHAValueToken extracts the value token from a raw `sync_commit_sha` value
// per §D.1: trim, take the first whitespace-delimited run, then strip a SINGLE
// leading and a SINGLE trailing quote from that run.
//
// Backtick is included in the quote set alongside `"` and `'` so that the
// closer's pre-inversion normalization (`strings.Trim(v, "\"'`+"`"+`")`) stays
// covered — a value the old predicate handled must not stop being handled.
func syncSHAValueToken(raw string) string {
	v := strings.TrimSpace(raw)
	if i := strings.IndexFunc(v, unicode.IsSpace); i >= 0 {
		v = v[:i]
	}
	v = strings.TrimPrefix(v, `"`)
	v = strings.TrimPrefix(v, `'`)
	v = strings.TrimPrefix(v, "`")
	v = strings.TrimSuffix(v, `"`)
	v = strings.TrimSuffix(v, `'`)
	v = strings.TrimSuffix(v, "`")
	return v
}

// isCommitSHAToken reports whether a value token is a commit SHA (REQ-SSF-002).
//
// This is THE shared predicate. Both `needsSHABackfill` (the closer's write-side
// decision) and `SyncSHASlotFormatRule` (the lint's read-side decision) derive
// from it; a second copy of the hex test anywhere in this package is the defect
// AC-SSF-007 exists to catch.
func isCommitSHAToken(token string) bool {
	return commitSHATokenPattern.MatchString(token)
}

// isSyncSHAPlaceholder reports whether a value token is a recognized backfill
// placeholder (REQ-SSF-005).
//
// This is a READ-SIDE exemption only. The closer deliberately does NOT consult
// it: a placeholder is precisely a slot still owed a real SHA, so the write side
// must treat it as requiring backfill (REQ-SSF-003) while the read side stays
// silent about it (REQ-SSF-005). The two guards answer different questions, and
// wiring this predicate into `needsSHABackfill` would re-create the exact blind
// spot the card was opened to close.
func isSyncSHAPlaceholder(token string) bool {
	return syncSHAPlaceholderPattern.MatchString(token)
}
