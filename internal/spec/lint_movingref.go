package spec

// lint_movingref.go — MovingRefUnpinnedRule (SPEC-MOVING-REF-GUARD-001 M3,
// REQ-MRG-001..REQ-MRG-009). A `warning`-severity SPEC-artifact lint rule that
// flags a line whose invariant claim is decided against a MOVING git ref
// (origin/main, origin/develop, origin/HEAD) with nothing pinning the value.
//
// The defect: a measurement taken at `origin/main` is valid at the instant it
// is taken and silently expires the moment mainline advances. Re-serving that
// expired measurement as current is a baseline-attribution violation
// (`.claude/rules/moai/core/verification-claim-integrity.md` §2, and its §2.1
// anchor-or-subject specialization, whose detectable half this rule enforces).
//
// REQ-MRG-001 is a FOUR-way conjunction. A finding fires only when a line:
//  1. names a moving ref,
//  2. inside a git-command context,
//  3. carries an invariant-claim marker, and
//  4. carries NEITHER a 7-40 char hex SHA nor a frozen-baseline variable
//     reference (REQ-MRG-008).
//
// Conjunct 3 is load-bearing and is the one with no obvious negative control:
// dropping it yields the SIMPLER rule, which passes every other criterion in
// this SPEC while over-firing ~12x on the live corpus (495 findings against the
// 42 the `warning` severity is sized on — spec.md §D.5, acceptance.md
// AC-MRG-014). AC-MRG-014 is the negative control that catches that mutant.
//
// Severity is `warning`, NEVER `error` (spec.md §D.5): 42 candidate lines exist
// in the corpus today, and reddening them on the first run makes bulk
// suppression the rational response — the outcome this rule exists to prevent.
//
// Every finding is ALSO emitted `Advisory: true` (REQ-MRG-009 as amended at
// spec.md v0.8.0, §D.7), so `--strict` does not escalate it. `warning` alone
// delivered half of §D.5's intent and `--strict` took it back: the run-phase
// merge reddened `develop`'s SPEC Lint job, which invokes the linter with
// `--strict`, on 110 findings the corpus already carried. Advisory emission is
// the completion of §D.5's intent, not a retreat from it — the finding stays a
// `warning`, stays in the report under every invocation, and gates nothing.
//
// The mechanism is deliberately the EMISSION SITE, not the era map. The code is
// absent from `eraDemotableCodes` and must stay absent: that map is consulted
// only for `SeverityError` findings, so a warning never reaches it, and the
// gating findings sit on modern-era (V3R6) SPECs that no era path can demote.
// `StatusGitConsistency` (lint.go) is the existing precedent for emission-site
// advisory.
//
// The promotion condition is prose, in spec.md §D.7, and prose does not fire:
// once the corpus findings are remediated or R3-exempted, the retracted
// `--strict` clause returns verbatim. Until then this guard reports without
// gating, which is the shape this SPEC itself exists to detect.
//
// Three shapes this rule deliberately does NOT treat as exemptions:
//   - The three-dot form `A...B` (REQ-MRG-007). Merge-base is not stable under
//     upstream advance: spec.md §B.2 measured the IDENTICAL wrong answer from
//     both forms in this tree. An implementer adding a `...` early return is
//     the anticipated mistake; AC-MRG-004 is the guard.
//   - A bare `<!-- moving-ref-ok: -->` marker (REQ-MRG-003). A reason-less
//     marker would make silencing cheaper than fixing, inverting the incentive.
//   - The R4 form (REQ-MRG-010). DEFERRED out of M3 by operator decision (Q0
//     option C): spec.md §B.7 measured R4's reachable class as 0 of 42
//     candidate lines on two independent probes, so an exclusion for it can
//     only over-exempt today, never under-exempt. Early R4-form lines are
//     silenced with the R3 marker instead. R4 remains in the finding MESSAGE —
//     the doctrine carries the remedy; only its lint exclusion is deferred.
//
// Detection limits are stated in spec.md §F (L1-L7) and are NOT claimed covered
// here. In particular L2: the detector reads SHAPE, never SUBJECT. It cannot
// apply the anchor-or-subject predicate, so every finding is a question put to
// a human, never a verdict — and the message says so.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// movingRefGitContextPattern is filter 2 of spec.md §B.3, recorded verbatim in
// progress.md §E.1 so the corpus figures (527 lines) reproduce. The `[^`]*`
// span deliberately cannot cross a backtick: it keeps the git verb and the ref
// inside one inline-code span rather than pairing a verb in one span with a ref
// in another.
//
// The trailing `\b` is a deliberate TIGHTENING of the recorded filter, added at
// M3 and not present in the §B.3 grep. Without it `origin/mainx` matches on the
// `origin/main` prefix, and AC-MRG-001's stated mutation (rename the fixture's
// ref to `origin/mainx`, expect the count to drop to 0) cannot turn the
// criterion red — the criterion would be vacuous, which is the exact hazard
// acceptance.md §A exists to prevent. The boundary does not lose any real form:
// `origin/main...HEAD`, `origin/main -- path`, and `origin/develop` at
// end-of-line all keep a non-word character (or the line end) after the ref.
var movingRefGitContextPattern = regexp.MustCompile("git [a-z-]+[^`]*origin/(main|develop|HEAD)\\b")

// movingRefTokenPattern extracts the moving ref named on the line, for the
// finding message. Filter 1 of §B.3, with the same `\b` tightening.
var movingRefTokenPattern = regexp.MustCompile(`origin/(main|develop|HEAD)\b`)

// invariantClaimPattern is filter 3 of §B.3 — the claim-marker alternation,
// recorded verbatim in progress.md §E.1 (the piece the iter-1 auditor could not
// reconstruct). Applied case-insensitively. This is conjunct 3 of REQ-MRG-001;
// AC-MRG-014 is its negative control.
var invariantClaimPattern = regexp.MustCompile(
	`(?i)byte-unchanged|byte unchanged|unchanged|preserv|보존|no diff|empty|0 files|부재|absent|변경 ?없|그대로`)

// shaPinPattern is filter 4 of §B.3 (REQ-MRG-008): a 7-40 character hex token
// on the line means the claim carries its own anchor.
var shaPinPattern = regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)

// frozenBaselinePattern is the R2 remedy's signature (REQ-MRG-008):
// `$BASELINE_SHA` or an equivalent frozen-baseline variable reference.
var frozenBaselinePattern = regexp.MustCompile(`\$\{?[A-Z_]*BASELINE[A-Z_]*\}?`)

// movingRefMarkerPattern is the M2 exemption surface (REQ-MRG-002/003):
// `<!-- moving-ref-ok: <reason> -->`, on the flagged line or the line
// immediately above it. An HTML comment is used so it is invisible in rendered
// markdown. Capture group 1 is the reason, which must be non-empty.
var movingRefMarkerPattern = regexp.MustCompile(`<!--\s*moving-ref-ok:(.*?)-->`)

// divergenceCommandPattern matches the branch-divergence read (REQ-MRG-006).
var divergenceCommandPattern = regexp.MustCompile(`rev-list\s+--count\s+--left-right`)

// divergenceFigurePattern matches a CITED divergence figure — a pair of
// integers presented as a result (`→ 0 0`, `=> 3 1`, “ `0 0` “). The citation
// is what separates REQ-MRG-006's shape from a plain instruction to measure: a
// line that merely names the command asserts nothing, and is AC-MRG-014's
// negative control rather than a finding.
var divergenceFigurePattern = regexp.MustCompile(
	"(?:(?:→|->|=>|:|=)\\s*`?\\s*\\d+\\s+\\d+|`\\s*\\d+\\s+\\d+\\s*`)")

// datedReferencePattern is REQ-MRG-006's second escape hatch: a divergence
// figure carrying a date is already demoted to a dated reference (R4 shape).
var datedReferencePattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// MovingRefUnpinnedRule implements the Rule interface for REQ-MRG-001.
//
// It reads SIBLING artifacts, not just `doc.Body`: `SPECDoc.Body` carries
// spec.md alone, while most real occurrences live in progress.md and
// acceptance.md (spec.md §B.3). A body-only rule would miss the majority of the
// corpus while appearing to work — AC-MRG-009 is that criterion. The precedent
// for a rule reaching past the single parsed document is HaikuResidualRule.
type MovingRefUnpinnedRule struct{}

// Code returns the stable lint finding code.
func (r *MovingRefUnpinnedRule) Code() string { return "MovingRefUnpinned" }

// Check scans every markdown artifact in the SPEC's own directory.
func (r *MovingRefUnpinnedRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if doc == nil || doc.Path == "" {
		return nil
	}
	dir := filepath.Dir(doc.Path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	var findings []Finding
	for _, name := range names {
		path := filepath.Join(dir, name)
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			continue
		}
		findings = append(findings, r.scanLines(path, strings.Split(string(data), "\n"))...)
	}
	return findings
}

// scanLines applies REQ-MRG-001 (and its REQ-MRG-006 divergence variant) to one
// artifact's lines.
func (r *MovingRefUnpinnedRule) scanLines(path string, lines []string) []Finding {
	var findings []Finding
	for i, line := range lines {
		// Conjuncts 1+2: a moving ref inside a git-command context.
		if !movingRefGitContextPattern.MatchString(line) {
			continue
		}
		// Conjunct 4 (REQ-MRG-008): an already-anchored claim is not a finding.
		// NOTE: no `...` escape hatch here — REQ-MRG-007, AC-MRG-004.
		if shaPinPattern.MatchString(line) || frozenBaselinePattern.MatchString(line) {
			continue
		}

		// Conjunct 3 (REQ-MRG-001), or the REQ-MRG-006 divergence variant.
		claim := invariantClaimPattern.MatchString(line)
		divergence := divergenceCommandPattern.MatchString(line) &&
			divergenceFigurePattern.MatchString(line) &&
			!datedReferencePattern.MatchString(line)
		if !claim && !divergence {
			continue
		}

		ref := "a moving ref"
		if m := movingRefTokenPattern.FindString(line); m != "" {
			ref = "`" + m + "`"
		}

		// REQ-MRG-002 / REQ-MRG-003: the author-declared exemption.
		reason, present := markerReason(lines, i)
		switch {
		case present && reason != "":
			continue // suppressed — the author applied the predicate and declared R3
		case present:
			findings = append(findings, Finding{
				File:     path,
				Line:     i + 1,
				Severity: SeverityWarning,
				Advisory: true, // REQ-MRG-009 (v0.8.0): reports, never gates — see file header
				Code:     r.Code(),
				Message: fmt.Sprintf(
					"incomplete `moving-ref-ok` marker: the exemption for %s carries no reason, so it does not suppress. "+
						"The reason is mandatory — a bare marker would make silencing cheaper than fixing. "+
						"Write `<!-- moving-ref-ok: <why this ref is the subject, not an anchor> -->`, or apply one of the four remedies. %s",
					ref, remedyClause()),
			})
		default:
			findings = append(findings, Finding{
				File:     path,
				Line:     i + 1,
				Severity: SeverityWarning,
				Advisory: true, // REQ-MRG-009 (v0.8.0): reports, never gates — see file header
				Code:     r.Code(),
				Message: fmt.Sprintf(
					"%s decides an invariant claim on this line with no SHA pin and no frozen-baseline variable, "+
						"so the measurement expires the moment mainline advances. %s",
					ref, remedyClause()),
			})
		}
	}
	return findings
}

// remedyClause names ALL FOUR remediation branches of spec.md §D.2
// (REQ-MRG-004, AC-MRG-008). The COUNT is what does the work: with one remedy
// on offer the reader pins everything — this SPEC's dominant failure mode — and
// with four, none of them free, choosing requires applying the predicate.
// Pinning is deliberately not presented as the sole or default remedy, and R4
// stays here even though its lint exclusion is deferred, because it is the
// branch a reader cannot reach by intuition.
//
// The closing sentence discharges limit L2: the detector reads shape, never
// subject, so a finding is a question put to a human and never a verdict.
func remedyClause() string {
	return "Is this ref the ANCHOR of the measurement or its SUBJECT? Apply the anchor-or-subject predicate " +
		"(`.claude/rules/moai/core/verification-claim-integrity.md` §2.1), then choose among four remedies — " +
		"R1 pin the resolved SHA; " +
		"R2 freeze at pre-flight (`BASELINE_SHA=$(git rev-parse origin/main)`) and decide the criterion against the frozen value; " +
		"R3 keep the moving ref and declare the exemption with a non-empty reason, `<!-- moving-ref-ok: <reason> -->`; " +
		"R4 state the measuring command as the criterion and demote any value to a dated reference. " +
		"Pinning is one branch of four, not the default. This finding is a question, not a verdict: the detector reads shape, never subject."
}

// markerReason looks for a `moving-ref-ok` marker on the flagged line or the
// line immediately above it (the M2 line-scope rule). It reports the trimmed
// reason and whether a marker was present at all — the two are distinct, since
// a present-but-empty reason is REQ-MRG-003's incomplete-marker finding rather
// than a suppression.
func markerReason(lines []string, idx int) (reason string, present bool) {
	for _, j := range []int{idx, idx - 1} {
		if j < 0 || j >= len(lines) {
			continue
		}
		m := movingRefMarkerPattern.FindStringSubmatch(lines[j])
		if m == nil {
			continue
		}
		if trimmed := strings.TrimSpace(m[1]); trimmed != "" {
			return trimmed, true
		}
		present = true
	}
	return "", present
}
