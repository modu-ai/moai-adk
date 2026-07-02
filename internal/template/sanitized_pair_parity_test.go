// sanitized_pair_parity_test.go: CI guard for "sanitized-pair" rule files — the
// rule files that exist in BOTH `.claude/rules/moai/**` (working copy, retains
// internal-development provenance) AND
// `internal/template/templates/.claude/rules/moai/**` (distribution mirror,
// held sanitized per CLAUDE.local.md §25). These pairs are DELIBERATELY excluded
// from the byte-exact mirror parity test (rule_template_mirror_test.go) because
// byte-identity cannot hold once the template copy is sanitized.
//
// The hazard this guard closes: because sanitized pairs are outside byte-parity,
// a real (non-token) doctrine change to the LOCAL copy can silently fail to
// propagate to the template mirror — users receive stale doctrine, and nothing
// catches it. This test normalizes away the intentionally-divergent internal
// tokens (SPEC-IDs, REQ-/AC- tokens, commit SHAs, ISO dates) from BOTH copies,
// then checks that no DOCTRINE CONTENT is present in one copy but absent from the
// other beyond the reword tolerance that §25 sanitization legitimately produces.
//
// Sentinel on failure: SANITIZED_PAIR_PARITY_DRIFT
//
// IMPORTANT design note — why not "normalize then assert byte-identical":
// §25 sanitization in this repo is implemented as prose REWORDING/REMOVAL, not
// mechanical token substitution. Example (ci-watch-protocol.md):
//   local:    "...added by SPEC-XXX Layer C in response to the W3 YYY meta-analysis..."
//   template: "...added in response to a meta-analysis..."
// A token-only normalization replaces SPEC-XXX with a placeholder but CANNOT
// equalize the reworded surrounding prose. So a normalize-then-exact-match guard
// would report EVERY current sanitized pair as drift (an always-red guard). This
// test instead measures STRUCTURAL drift: after token-normalization it compares
// the multiset of content lines and tolerates balanced reword (a line changed in
// place = one local-only + one template-only line), while flagging one-sided
// blocks of doctrine lines (a whole section present in one copy, absent in the
// other) that exceed the reword tolerance. That is the shape of real,
// non-propagated doctrine drift.
//
// Because every normalizer is applied IDENTICALLY to both copies, an over-broad
// normalizer can only ever MASK a real difference (false negative); it can never
// manufacture a false difference (false positive). This property lets the SHA
// normalizer stay a simple hex-run regex without risk of a spurious failure.
package template_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// sanitizedPairPaths is the sanitized-pair registry. Each entry is a rule/agent
// file that (a) exists in both the working copy and the template mirror, and
// (b) differs byte-wise because the mirror is held sanitized per §25.
//
// Registry provenance (built from evidence, not memory): the first five entries
// are the files documented as removed from the byte-parity allowlist in
// rule_template_mirror_test.go (workflowOptMirroredPaths comment block). The last
// two (askuser-protocol.md, verification-claim-integrity.md) are sanitized-
// divergent pairs found in NEITHER the byte-mirror allowlist nor that documented
// exclusion — they are the gap this guard was written to close. Every entry was
// confirmed at authoring time to exist in both trees and to differ by `diff -q`.
//
// Maintenance: add a file here only after confirming both copies exist and the
// divergence is genuine §25 sanitization (not accidental drift). Removing a file
// here should coincide with promoting it back into the byte-parity allowlist.
var sanitizedPairPaths = []string{
	".claude/rules/moai/development/manager-develop-prompt-template.md",
	".claude/rules/moai/workflow/ci-watch-protocol.md",
	".claude/rules/moai/core/agent-common-protocol.md",
	".claude/rules/moai/workflow/verification-batch-pattern.md",
	".claude/agents/moai/plan-auditor.md",
	".claude/rules/moai/core/askuser-protocol.md",
	".claude/rules/moai/core/verification-claim-integrity.md",
}

// tokenNormalizer pairs a regex matching an intentionally-divergent internal
// token class with the placeholder that replaces it. Kept as a package-level
// slice so the token set is a single, reviewable, extensible list.
type tokenNormalizer struct {
	name        string
	re          *regexp.Regexp
	placeholder string
}

// internalTokenNormalizers are applied IDENTICALLY to both the local and template
// copy before comparison. Order is fixed but effectively independent: the token
// classes do not overlap (SPEC-/REQ-/AC- prefixes are disjoint; a SHA run is
// lowercase-hex only; a date is digit-hyphen only). Applied to both sides, an
// over-broad match can only mask a real diff, never create a false one.
var internalTokenNormalizers = []tokenNormalizer{
	// Internal SPEC identifiers, e.g. SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001,
	// SPEC-EVIDENCE-CLAIM-INVARIANT-001. Stripped from the distribution mirror.
	{"spec-id", regexp.MustCompile(`\bSPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`), "<SPEC-ID>"},
	// Requirement tokens, e.g. REQ-LEDGER-001, REQ-WO-004.
	{"req-token", regexp.MustCompile(`\bREQ-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`), "<REQ>"},
	// Acceptance-criterion tokens, e.g. AC-LEDGER-006, AC-HV4-005b (trailing
	// lowercase sub-letter allowed).
	{"ac-token", regexp.MustCompile(`\bAC-[A-Z0-9]+(?:-[0-9A-Za-z]+)*\b`), "<AC>"},
	// Commit SHAs: any lowercase-hex run of 7..40 chars. Symmetric application
	// makes an occasional all-hex-letter word match harmless (see file header).
	{"commit-sha", regexp.MustCompile(`\b[0-9a-f]{7,40}\b`), "<SHA>"},
	// ISO dates, e.g. 2026-06-17.
	{"iso-date", regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`), "<DATE>"},
}

// structuralDriftToleranceLines is the reword tolerance: the number of net
// one-sided content lines permitted before a pair is flagged as structural
// drift. §25 sanitization rewords lines IN PLACE (a changed line contributes one
// local-only + one template-only line, netting zero), and occasionally compresses
// an N-line clause to N-1 lines. A whole non-propagated doctrine section is tens
// of one-sided lines — far above this tolerance. Empirically the current reword
// pairs net <= 1 one-sided line; a whole missing section nets ~24. The gap is
// wide, so this small tolerance separates the two cleanly without masking a real
// section-sized drift.
const structuralDriftToleranceLines = 4

// TestSanitizedPairParity verifies that each sanitized-pair rule file carries the
// same doctrine in both trees, tolerating only the internal-token divergence and
// in-place reword that §25 sanitization legitimately produces.
//
// Reports the literal sentinel SANITIZED_PAIR_PARITY_DRIFT on real drift so CI
// log parsers can match the failure pattern.
func TestSanitizedPairParity(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	for _, rel := range sanitizedPairPaths {
		t.Run(filepath.Base(rel), func(t *testing.T) {
			t.Parallel()

			srcPath := filepath.Join(projectRoot, rel)
			mirrorPath := filepath.Join(projectRoot, "internal", "template", "templates", rel)

			srcContent, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatalf("registry entry %s: source unreadable %s: %v", rel, srcPath, err)
			}
			mirrorContent, merr := os.ReadFile(mirrorPath)
			if merr != nil {
				if os.IsNotExist(merr) {
					// A registry entry MUST have a mirror; a missing mirror is a
					// harder failure than drift (the pair invariant is broken).
					t.Fatalf(
						"SANITIZED_PAIR_PARITY_DRIFT: registry entry %s has no mirror at %s; "+
							"either the file is not a sanitized pair (remove it from sanitizedPairPaths) "+
							"or the mirror was deleted (restore it)",
						rel, mirrorPath,
					)
				}
				t.Fatalf("registry entry %s: mirror unreadable %s: %v", rel, mirrorPath, merr)
			}

			localLines := normalizedContentLines(srcContent)
			tmplLines := normalizedContentLines(mirrorContent)

			// Strongest pass: after token-normalization the doctrine content is
			// byte-identical (divergence was purely internal tokens). No current
			// pair reaches this, but it is the ideal end-state for a fully
			// token-only-sanitized file.
			if strings.Join(localLines, "\n") == strings.Join(tmplLines, "\n") {
				t.Logf("%s: token-only divergence — normalized doctrine identical (%d content lines)", rel, len(localLines))
				return
			}

			localOnly := multisetSubtract(localLines, tmplLines)
			tmplOnly := multisetSubtract(tmplLines, localLines)
			rewordPairs := min(len(localOnly), len(tmplOnly))
			localExcess := len(localOnly) - len(tmplOnly) // >0: local doctrine not in template

			t.Logf(
				"%s: normalized diff — %d local-only, %d template-only, ~%d reword pairs, net one-sided=%d (tolerance %d)",
				rel, len(localOnly), len(tmplOnly), rewordPairs, localExcess, structuralDriftToleranceLines,
			)

			switch {
			case localExcess > structuralDriftToleranceLines:
				t.Errorf(
					"SANITIZED_PAIR_PARITY_DRIFT: %s carries %d content line(s) of doctrine present in the "+
						"LOCAL copy but ABSENT from the template mirror (beyond the %d-line reword tolerance). "+
						"Users receive stale doctrine. Propagate the change to %s (sanitized per §25), or if the "+
						"omission is intentional, remove this file from sanitizedPairPaths with a documented reason.\n"+
						"First local-only normalized lines:\n%s",
					rel, localExcess, structuralDriftToleranceLines, mirrorPath,
					sampleLines(localOnly, 24),
				)
			case -localExcess > structuralDriftToleranceLines:
				t.Errorf(
					"SANITIZED_PAIR_PARITY_DRIFT: %s carries %d content line(s) present in the TEMPLATE mirror "+
						"but ABSENT from the local copy (template ahead of source — the working copy lost doctrine "+
						"the mirror still ships).\nFirst template-only normalized lines:\n%s",
					rel, -localExcess, sampleLines(tmplOnly, 24),
				)
			default:
				// Within tolerance: divergence is internal-token + in-place reword
				// only. This is the expected, tolerated §25 sanitization shape.
				t.Logf("%s: within reword tolerance — intentional §25 sanitization, no structural drift", rel)
			}
		})
	}
}

// normalizeInternalTokens applies every internalTokenNormalizers regex to s.
func normalizeInternalTokens(s string) string {
	for _, n := range internalTokenNormalizers {
		s = n.re.ReplaceAllString(s, n.placeholder)
	}
	return s
}

// normalizedContentLines returns the token-normalized, blank-stripped content
// lines of a file. Blank lines and trailing whitespace are dropped so pure
// formatting differences never register as doctrine drift.
func normalizedContentLines(content []byte) []string {
	norm := normalizeInternalTokens(string(content))
	raw := strings.Split(norm, "\n")
	out := make([]string, 0, len(raw))
	for _, ln := range raw {
		trimmed := strings.TrimRight(ln, " \t\r")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

// multisetSubtract returns the lines of a not covered by an equal-or-greater
// count of the same line in b (multiset difference a - b). A line reworded in
// place appears in both a-b and b-a; a line only in a (a one-sided addition)
// appears only in a-b.
func multisetSubtract(a, b []string) []string {
	remaining := make(map[string]int, len(b))
	for _, l := range b {
		remaining[l]++
	}
	var only []string
	for _, l := range a {
		if remaining[l] > 0 {
			remaining[l]--
			continue
		}
		only = append(only, l)
	}
	return only
}

// sampleLines renders up to n lines for a failure message, one per line, with a
// truncation marker when the slice is longer.
func sampleLines(lines []string, n int) string {
	if len(lines) > n {
		return "    " + strings.Join(lines[:n], "\n    ") +
			"\n    ... (" + itoa(len(lines)-n) + " more)"
	}
	return "    " + strings.Join(lines, "\n    ")
}

// itoa is a tiny local int->string to avoid importing strconv for one call site.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
