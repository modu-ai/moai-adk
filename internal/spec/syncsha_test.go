package spec

// syncsha_test.go — M1 grammar + shared-predicate unit tests
// (SPEC-SYNC-SHA-SLOT-FORMAT-001, card t299).
//
// These tests pin the §D.1 value grammar itself: the token split, the quote
// handling, the SHA length band, and the canonical placeholder family. The two
// KNOWN LIMITATIONS recorded in spec.md §D.1 (L1 hex-word false negative, L2 no
// reachability) are asserted here as known behavior rather than left implicit —
// a limitation nobody wrote down is indistinguishable from a bug nobody found.

import "testing"

// TestSyncSHAValueToken_FirstWhitespaceRun pins the §D.1 token rule: the value
// token is the FIRST whitespace-delimited run, and everything after it is a
// free-form annotation that neither decision may see.
//
// Mutation that must turn it red: return the whole trimmed value instead of its
// first run — the annotated rows below stop yielding a bare token.
func TestSyncSHAValueToken_FirstWhitespaceRun(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"bare", "a6bbbf82b", "a6bbbf82b"},
		{"hash annotation", "a6bbbf82b   # backfilled in the following commit", "a6bbbf82b"},
		{"em-dash annotation", "a6bbbf82b — landed on develop", "a6bbbf82b"},
		{"paren annotation", "a6bbbf82b (backfilled post-commit)", "a6bbbf82b"},
		{"quoted", `"a6bbbf82b"`, "a6bbbf82b"},
		{"quoted with annotation", `"pending-backfill-sync"   # D3 self-reference exemption`, "pending-backfill-sync"},
		{"single quoted", "'a6bbbf82b'", "a6bbbf82b"},
		{"backtick quoted", "`a6bbbf82b`", "a6bbbf82b"},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"space-bearing placeholder", "(this commit)", "(this"},
		{"leading whitespace", "   a6bbbf82b", "a6bbbf82b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := syncSHAValueToken(tc.raw); got != tc.want {
				t.Errorf("syncSHAValueToken(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

// TestIsCommitSHAToken_LengthBand pins the 7-40 hex band of §D.1.
//
// Mutation that must turn it red: narrow the pattern to {40} — the 7-char and
// 9-char rows flip to false.
func TestIsCommitSHAToken_LengthBand(t *testing.T) {
	cases := []struct {
		token string
		want  bool
	}{
		{"a6bbbf8", true},                                  // 7 — lower bound
		{"a6bbbf82b", true},                                // 9 — the corpus's usual short form
		{"a6bbbf82b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6", true}, // 40 — upper bound
		{"A6BBBF82B", true},                                // uppercase hex is hex
		{"a6bbbf", false},                                  // 6 — below the band
		{"a6bbbf82b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f", false}, // 41 — above the band
		{"", false}, // the empty slot
		{"null", false},
		{"pending", false},
		{"<pending>", false},
		{"TBD", false},
		{"(this", false},
		{"pending-backfill", false},
		{"a6bbbf82b#annot", false}, // a token is a token, not a prefix match
	}
	for _, tc := range cases {
		if got := isCommitSHAToken(tc.token); got != tc.want {
			t.Errorf("isCommitSHAToken(%q) = %v, want %v", tc.token, got, tc.want)
		}
	}
}

// TestIsSyncSHAPlaceholder_CanonicalFamily pins the placeholder family of §D.1:
// the bare canonical spelling plus any `-`-suffixed member.
//
// CASE HANDLING, stated explicitly because the corpus forces the decision: the
// PREFIX `pending-backfill` is matched case-SENSITIVELY (the grammar writes it
// as a lowercase literal), while the SUFFIX admits `[A-Za-z0-9-]+` and is
// therefore case-insensitive by construction. This is what admits the corpus's
// `pending-backfill-SYNC` (SPEC-AUDIT-SNAPSHOT-001) while still refusing
// `Pending-Backfill`. Loosening the prefix to case-insensitive would widen the
// exemption with no corpus occurrence to justify it.
//
// Mutation that must turn it red: drop the optional-suffix group — every
// suffixed row flips to false, including the 24 corpus occurrences of
// `pending-backfill-sync`.
func TestIsSyncSHAPlaceholder_CanonicalFamily(t *testing.T) {
	cases := []struct {
		token string
		want  bool
	}{
		{"pending-backfill", true},
		{"pending-backfill-sync", true},
		{"pending-backfill-SYNC", true}, // corpus: SPEC-AUDIT-SNAPSHOT-001
		{"pending-backfill-after-merge", true},
		{"pending-backfill-m6-m7", true},
		{"pending-backfill-a382666ce", true},
		{"Pending-Backfill", false}, // prefix is case-sensitive — see doc comment
		{"pending-backfill-", false},
		{"pending", false},
		{"<pending>", false},
		{"backfill-pending", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isSyncSHAPlaceholder(tc.token); got != tc.want {
			t.Errorf("isSyncSHAPlaceholder(%q) = %v, want %v", tc.token, got, tc.want)
		}
	}
}

// TestSyncSHAGrammar_KnownLimitationL1 asserts spec.md §D.1's limitation L1 as
// KNOWN BEHAVIOR: a 7-40 character all-hex English word parses as a SHA, so a
// slot holding one is a false negative the guard does not catch.
//
// The class is accepted rather than defended against: the alternative is a
// dictionary, and the false negative costs one unflagged slot while a dictionary
// would cost a maintenance surface. This test exists so that a later change
// SILENTLY closing (or widening) the class is visible in the diff rather than
// discovered by a confused reader.
//
// Mutation that must turn it red: add a dictionary exclusion for `defaced` —
// the assertion below flips.
func TestSyncSHAGrammar_KnownLimitationL1(t *testing.T) {
	for _, word := range []string{"defaced", "deadbeef", "facade", "accede"} {
		token := syncSHAValueToken(word)
		isHexWord := len(word) >= 7 && len(word) <= 40
		if got := isCommitSHAToken(token); got != isHexWord {
			t.Errorf("L1: isCommitSHAToken(%q) = %v, want %v (all-hex English words of 7-40 chars parse as SHAs)", word, got, isHexWord)
		}
	}
}

// TestSyncSHAGrammar_KnownLimitationL2 asserts spec.md §D.1's limitation L2 as
// KNOWN BEHAVIOR: the predicate reads SHAPE, never reachability. Seven plausible
// hex characters that name no commit in any repository are accepted exactly as a
// real commit SHA is.
//
// Reachability verification is a separate concern and is not claimed here; this
// test is what stops a reader inferring it from the predicate's name.
//
// Mutation that must turn it red: make isCommitSHAToken consult git — the
// unreachable rows flip to false.
func TestSyncSHAGrammar_KnownLimitationL2(t *testing.T) {
	for _, token := range []string{"0000000", "fffffff", "0123456789abcdef0123456789abcdef01234567"} {
		if !isCommitSHAToken(token) {
			t.Errorf("L2: isCommitSHAToken(%q) = false, want true (shape-only — reachability is out of scope)", token)
		}
	}
}

// TestSyncSHASlotFormat_AbsentFromEraDemotableCodes decides AC-SSF-010, and so
// decides REQ-SSF-007 mechanically rather than by a grep a reader has to
// remember to run.
//
// `eraDemotableCodes` demotes ERRORS (lint.go:284 gates on
// `f.Severity == SeverityError`); a warning is already made advisory by the
// branch immediately below it. An entry for a warning-severity code would
// therefore be INERT, and an inert entry in a policy map reads to a later
// maintainer as intent that was never meant. MovingRefUnpinnedRule records the
// same choice for the same reason.
//
// [HARD] This criterion decides the requirement AS WRITTEN; it does not decide
// the contingency. REQ-SSF-007's justification holds only while this rule's
// severity is `warning` at the Finding level. If a later change promotes at that
// level rather than at `Report.HasErrors`, the map becomes the ONLY remaining
// shelter for the five findings on closed history, the requirement inverts, and
// this test would be enforcing the wrong thing. The instruction in that case is
// to stop and report — not to add the entry, and not to weaken this test.
//
// Mutation that must turn it red: add `"SyncSHASlotFormat": true` to the map.
func TestSyncSHASlotFormat_AbsentFromEraDemotableCodes(t *testing.T) {
	if eraDemotableCodes["SyncSHASlotFormat"] {
		t.Error("SyncSHASlotFormat is present in eraDemotableCodes; the entry is inert for a warning and reads as intent (REQ-SSF-007, AC-SSF-010)")
	}
	want := map[string]bool{"MissingExclusions": true, "FrontmatterInvalid": true}
	if len(eraDemotableCodes) != len(want) {
		t.Fatalf("eraDemotableCodes has %d entries, want exactly %d (%v); got %v", len(eraDemotableCodes), len(want), want, eraDemotableCodes)
	}
	for code := range want {
		if !eraDemotableCodes[code] {
			t.Errorf("eraDemotableCodes is missing the expected entry %q", code)
		}
	}
}
