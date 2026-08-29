package spec

// closer_syncsha_test.go — M2 closer write-side inversion
// (SPEC-SYNC-SHA-SLOT-FORMAT-001, AC-SSF-004 / AC-SSF-005, card t299).
//
// The inversion is the substance of the fix: `needsSHABackfill` stops being an
// OPEN enumeration of what a placeholder might be spelled as and becomes a
// CLOSED positive test for what a SHA is. The set of non-SHAs needs no
// maintenance; the set of placeholder spellings provably does — 28 distinct
// spellings exist in the corpus today (spec.md §B.3), and every new suffix an
// author invents silently opted that SPEC out of repair.

import "testing"

// TestNeedsSHABackfill_OutOfAllowlistPlaceholder decides AC-SSF-004: card t354's
// failure, reproduced as a test rather than described.
//
// `pending-backfill-sync` is the value SPEC-BACKLOG-LOCK-BUDGET-001 has carried
// permanently, and it is the spelling `spec-frontmatter-schema.md` §D3 itself
// prescribes — yet it was absent from the retired four-value allowlist, so the
// closer mistook it for a populated SHA and never repaired it. The blind spot
// was aimed at the sanctioned pattern.
//
// Mutation that must turn it red: restore the four-value switch
// (`case "", "(this commit)", "(pending)", "<pending>": return true`) — every
// row below flips to false.
func TestNeedsSHABackfill_OutOfAllowlistPlaceholder(t *testing.T) {
	outOfAllowlist := []string{
		"pending-backfill-sync", // SPEC-BACKLOG-LOCK-BUDGET-001, the motivating slot
		"pending-backfill",
		"pending-backfill-SYNC",
		"pending-backfill-after-merge",
		"null",
		"TBD (filled post-commit)",
		"pending  # backfilled via L60 atomic chore",
		"<pending>  # backfilled post-commit",
	}
	for _, v := range outOfAllowlist {
		if !needsSHABackfill(v) {
			t.Errorf("needsSHABackfill(%q) = false, want true (not a commit SHA, so the slot is still owed one)", v)
		}
	}
}

// TestNeedsSHABackfill_LegacySetPreserved decides AC-SSF-005: the four values of
// the retired allowlist still return true.
//
// A widening that silently drops a previously-caught case is a regression
// wearing a fix's clothes. This criterion is what makes "strictly a widening"
// (spec.md §D.2) a measurement rather than a claim. The sharp one is
// `(this commit)`: it contains a space, so its first token is `(this` — a token
// split that mishandled it would drop the single most common legacy placeholder.
//
// Mutation that must turn it red: make isCommitSHAToken return true for the
// empty string — the `""` row fails.
func TestNeedsSHABackfill_LegacySetPreserved(t *testing.T) {
	for _, v := range []string{"", "(this commit)", "(pending)", "<pending>"} {
		if !needsSHABackfill(v) {
			t.Errorf("needsSHABackfill(%q) = false, want true (retired-allowlist value must stay caught)", v)
		}
	}
}

// TestNeedsSHABackfill_SHAValuesNotRepaired is the negative control the two
// criteria above do not supply between them: a predicate returning true
// unconditionally satisfies both.
//
// Mutation that must turn it red: `return true` — every row below fails.
func TestNeedsSHABackfill_SHAValuesNotRepaired(t *testing.T) {
	populated := []string{
		"a6bbbf82b",
		`"a6bbbf82b"`,
		"a6bbbf82b   # backfilled in the following commit",
		"a6bbbf82b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6",
	}
	for _, v := range populated {
		if needsSHABackfill(v) {
			t.Errorf("needsSHABackfill(%q) = true, want false (slot already holds a commit SHA)", v)
		}
	}
}

// TestNeedsSHABackfill_MxDeliberateDeclarationsInScope RECORDS the inheritance
// spec.md §E accepts explicitly, rather than leaving "mx_commit_sha inherits the
// widening" as the whole account of a write path.
//
// `needsSHABackfill` is shared with the `mx_commit_sha` backfill at
// closer.go:332, which on a true writes the literal `(this commit)` over
// whatever the slot held. Three corpus values are DELIBERATE declarations rather
// than un-populated slots, and the inversion newly brings them into the
// predicate's scope:
//
//	<NA>                                     SPEC-CLIFIX-CONCURRENCY-001
//	_<pending Mx-phase>_                     SPEC-V3R6-BASH-RISK-GOVERNANCE-001
//	_(not applicable — this SPEC removes …)_  SPEC-V3R6-LIFECYCLE-REDESIGN-001
//
// The exposure is PROSPECTIVE, not live: all three owning SPECs are already
// `status: completed`, so no close is expected against them and the blast radius
// measured today is zero. The Mx phase itself is retired by the third SPEC in
// that list, so the class shrinks rather than grows.
//
// This test asserts the accepted behavior, it does not guard against it. No
// `mx_commit_sha` guard is written by this card (spec.md §E) — a durable
// "not applicable" declaration in that field is an mx-path requirement for its
// own card to raise.
//
// Mutation that must turn it red: add an mx exemption branch to
// `needsSHABackfill` — the rows below flip to false, and the reader learns that
// scope was taken without the SPEC being changed.
func TestNeedsSHABackfill_MxDeliberateDeclarationsInScope(t *testing.T) {
	declarations := []string{
		"<NA>",
		"_<pending Mx-phase>_",
		"_(not applicable — this SPEC removes the Mx-phase concept; REQ-LR-004/007)_",
	}
	for _, v := range declarations {
		if !needsSHABackfill(v) {
			t.Errorf("needsSHABackfill(%q) = false; the accepted behavior recorded in spec.md §E is true (in scope, blast radius zero)", v)
		}
	}
}
