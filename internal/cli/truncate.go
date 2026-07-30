package cli

import "unicode/utf8"

// truncate.go provides the shared rune-boundary truncation helpers introduced
// by SPEC-CLIFIX-HYGIENE-001 AC-HYG-001-006. Every user-facing string
// truncation site in the CLI tree that previously byte-sliced (s[:N]) — which
// splits multi-byte (CJK) runes and yields invalid UTF-8 — is routed through
// these helpers so the truncated output is always utf8.ValidString == true.
//
// Sites routed (complete set, derived by grepping the CLI tree for byte-slice
// truncation on user-content strings; ASCII-only slices such as hex hashes,
// API-key masking, session IDs, and SHAs are excluded because they cannot
// carry multi-byte runes):
//
//  1. constitution.go renderConstitutionTable — clause (prefix)
//  2. constitution.go renderConstitutionTable — file path (suffix)
//  3. tool_policy.go   renderList            — audit column (prefix)
//  4. tool_policy.go   truncateArg           — args_pattern (prefix)
//  5. github.go        runParseIssue         — issue body (prefix)

// truncateRunes returns the first max runes of s without splitting a
// multi-byte UTF-8 rune. If s contains max runes or fewer, s is returned
// unchanged. max <= 0 yields the empty string.
//
// The cut always lands on a rune boundary (the byte offset where the
// (max+1)-th rune begins), so the result satisfies utf8.ValidString == true
// for any UTF-8 input — including CJK content where each rune is 3 bytes.
func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	count := 0
	for i := range s {
		if count == max {
			return s[:i]
		}
		count++
	}
	return s // s has fewer than max runes
}

// truncateRunesSuffix returns the last max runes of s without splitting a
// multi-byte UTF-8 rune — the suffix counterpart of truncateRunes, used for
// right-aligned truncation (e.g. file paths where the filename, not the
// directory prefix, is the meaningful part). If s contains max runes or
// fewer, s is returned unchanged. max <= 0 yields the empty string.
func truncateRunesSuffix(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	total := utf8.RuneCountInString(s)
	if total <= max {
		return s
	}
	skip := total - max
	count := 0
	for i := range s {
		if count == skip {
			return s[i:]
		}
		count++
	}
	return s // unreachable when total > max; defensive
}
