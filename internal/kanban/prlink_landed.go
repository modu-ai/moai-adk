// prlink_landed.go — the landed check and its silent-failure guards
// (SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.9, REQ-1.10, M2;
// SPEC-TODO-LANDING-STATE-001 REQ-TLS-001..REQ-TLS-004, M1/M2;
// SPEC-TODO-LANDING-ATTRIBUTION-001 REQ-TLA-001..006, REQ-TLA-013, M1).
//
// This file is separated from the resolver because its failure mode is
// SILENT. The original guard was for the regex engine: `\b` is not POSIX ERE,
// git does not error on it, and an `-E` regression returns an empty commit set
// byte-for-byte indistinguishable from "this card has not landed".
//
// SPEC-TODO-LANDING-ATTRIBUTION-001 closed a second, larger silence in the
// same surface: the old query matched the card token against the WHOLE commit
// message, so a commit whose body merely mentioned a card read as that card
// having landed (measured precision 0/2 on the population the check marked
// landed). The check now feeds `git log <ref> --format=%s` — subjects ONLY —
// and decides on ATTRIBUTION POSITION within a subject: the six positional
// shapes of spec.md §A.4 and the non-attribution rule, in one named place
// (landedSubjectForms + subjectAttribution), never on the presence of the
// token anywhere in a message.
//
// The SAME silence has a third entrance, and it is what the landing-state
// SPEC closed: asking the question about the WRONG ref. A project that
// integrates on a branch other than the default answers "not landed" for
// every card that shipped, with no error and no empty-output warning. The ref
// is therefore resolved from configuration and threaded through as an input,
// and an unanswerable query is reported as `unknown` rather than collapsed
// into `not-landed`.
package kanban

import (
	"fmt"
	"regexp"
	"strings"
)

// LandedSubjectFormatFlag is the query shape the landed check is built on: a
// SUBJECT stream, never a whole-message stream. It is a constant rather than
// a literal so the tripwire test can assert the built argv against the same
// symbol the implementation uses — a `%B` regression would re-open the
// body-mention defect this file exists to prevent.
const LandedSubjectFormatFlag = "--format=%s"

// DefaultLandedRef is the ref the landed question falls back to when neither
// the project configuration nor the repository's own recorded default names
// an integration branch. It is the historical constant, so an unconfigured
// project without a symref behaves exactly as it did before the ref became
// resolvable.
const DefaultLandedRef = "origin/main"

// LandingAnswer is the landed check's THREE-valued answer.
//
// The third value is the point. Collapsing an unanswerable query into
// `not-landed` makes "the check ran and found nothing" and "the check could
// not run" the same bytes, so a guard that never executed reads exactly like
// a guard that passed. They are different facts and they are reported
// differently.
type LandingAnswer string

const (
	// LandingLanded means a commit whose subject ATTRIBUTES the card landed
	// on the resolved ref.
	LandingLanded LandingAnswer = "landed"
	// LandingNotLanded means the query ran against the resolved ref and its
	// subject stream carries no attribution of the card.
	LandingNotLanded LandingAnswer = "not-landed"
	// LandingUnknown means the question could not be asked — no git, no such
	// ref, a broken remote. It is NOT evidence of not-landed.
	LandingUnknown LandingAnswer = "unknown"
)

// LandedRefFor returns the ref the landed question should be asked about for
// the project rooted at projectRoot, resolved through the three-level chain
// (REQ-TLA-007): the configured integration branch, then the repository's own
// recorded default, then DefaultLandedRef. The chain lives in
// prlink_landedref.go; this wrapper preserves the historical one-argument
// signature for callers that need only the ref.
//
// "Has this landed?" means "is it in the branch this project integrates on".
// For a project that integrates on `develop`, asking `origin/main` answers a
// question nobody posed and answers it wrongly.
func LandedRefFor(projectRoot string) string {
	ref, _ := LandedRefForWithLevel(projectRoot)
	return ref
}

// CommandRunner runs one subprocess and returns its combined stdout. It is
// the injection seam the subprocess-census tests count invocations through.
type CommandRunner func(name string, args ...string) (string, error)

// landedSubjectForms holds the pattern of every positional shape the landed
// predicate recognises, with its form number from SPEC-TODO-LANDING-ATTRIBUTION-001
// §A.4. This table plus subjectAttribution directly below are THE one named
// place the enumeration lives in (REQ-TLA-002): a further form is a
// reviewable one-place diff, and removing an entry makes that form's named
// criterion go red (acceptance.md AC-TLA-005).
//
// None of these is an occurrence test. "The subject contains the token" is
// not a position, and admitting it reintroduces the defect REQ-TLA-001
// closes.
var landedSubjectForms = struct {
	// form 1 — conventional-commit scope at subject start.
	scope *regexp.Regexp
	// form 2 — trailing parenthetical closing the subject, the group
	// carrying nothing else.
	trailing *regexp.Regexp
	// form 2b — a card-bearing parenthetical group followed by exactly one
	// pull-request reference group that closes the subject.
	trailingRef *regexp.Regexp
	// form 3a — merge, card-led: `Merge card <card>`.
	mergeCard *regexp.Regexp
	// form 3c — merge, card-led, the local-merge spelling: `merge: <card>`.
	mergeColon *regexp.Regexp
	// the non-attribution rule and form 3b's target test — a merge subject's
	// NAMED TARGET. Rule (contraposed): a merge-commit subject that names a
	// target attributes no card unless that target is the branch the
	// resolved landed ref names.
	mergeTarget *regexp.Regexp
	// the merge-commit subject spelling the rule keys on.
	mergeCommit *regexp.Regexp
	// form 3b's trailing parenthetical group.
	trailingGroup *regexp.Regexp
}{
	scope:         regexp.MustCompile(`^[a-z]+\((t[0-9]+)\)!?:`),
	trailing:      regexp.MustCompile(`\((?:card )?(t[0-9]+)\)$`),
	trailingRef:   regexp.MustCompile(`\(([^()]*t[0-9]+[^()]*)\) \(#[0-9]+\)$`),
	mergeCard:     regexp.MustCompile(`^Merge card (t[0-9]+)`),
	mergeColon:    regexp.MustCompile(`^merge: (t[0-9]+)`),
	mergeTarget:   regexp.MustCompile(`^[Mm]erge\b.* into ([A-Za-z0-9/_.-]+)`),
	mergeCommit:   regexp.MustCompile(`^Merge\b`),
	trailingGroup: regexp.MustCompile(`\(([^()]*)\)$`),
}

// subjectCardToken is the whole-token extractor for subject attribution. The
// word boundaries are what stop `t44` from matching a `t443` token, and the
// literal comparison afterwards is what keeps the queried card id out of any
// regular expression.
var subjectCardToken = regexp.MustCompile(`\bt[0-9]+\b`)

// distinctCardTokens returns the distinct card tokens carried in s.
func distinctCardTokens(s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, tok := range subjectCardToken.FindAllString(s, -1) {
		if !seen[tok] {
			seen[tok] = true
			out = append(out, tok)
		}
	}
	return out
}

// subjectAttribution is the single point every landed verdict flows through.
// It returns the card the SUBJECT attributes, or "" when it attributes
// nothing. Together with landedSubjectForms it is the enumeration's one named
// place (REQ-TLA-002).
//
// Evaluation order encodes acceptance.md §D's deterministic tiebreak: the
// trailing parenthetical is consulted before the conventional-commit scope,
// so `docs(t263): … (t461)` attributes t461. A group carrying two or more
// distinct card tokens attributes NOTHING — that is a shape that delivered
// neither card, not an ambiguity to tiebreak (§D).
//
// The non-attribution rule, contraposed (§A.4): a merge-COMMIT subject that
// names an integration target attributes no card unless that target is the
// branch the resolved landed ref names (landedBranch — derived from the
// resolved ref, never a compiled-in name, REQ-TLA-013). A merge into anything
// else absorbs work rather than landing it. The rule keys on git's
// merge-commit spelling (`Merge …`); the lowercase `merge:` prefix is the
// card-led work-log spelling (form 3c family) whose `into` phrasing is prose
// — measured over the pinned corpus, keying the rule on lowercase subjects
// too would lose t78, the only id whose sole attribution is such a subject.
func subjectAttribution(subject, landedBranch string) string {
	// The non-attribution rule. Checked first: an absorb-direction merge
	// attributes nothing through ANY shape its subject carries.
	if m := landedSubjectForms.mergeTarget.FindStringSubmatch(subject); m != nil &&
		landedSubjectForms.mergeCommit.MatchString(subject) && m[1] != landedBranch {
		return ""
	}

	// Form 2 — trailing parenthetical, group carries nothing else.
	if m := landedSubjectForms.trailing.FindStringSubmatch(subject); m != nil {
		return m[1]
	}
	// Form 2b — the same position with exactly one reference group after it;
	// the card-bearing group must carry exactly one DISTINCT card token. A
	// rejected group falls through to the remaining forms.
	if m := landedSubjectForms.trailingRef.FindStringSubmatch(subject); m != nil {
		if toks := distinctCardTokens(m[1]); len(toks) == 1 {
			return toks[0]
		}
	}
	// Form 1 — conventional-commit scope at subject start.
	if m := landedSubjectForms.scope.FindStringSubmatch(subject); m != nil {
		return m[1]
	}
	// Forms 3a and 3c — one shape in two spellings: the card is the first
	// token after the merge verb.
	if m := landedSubjectForms.mergeCard.FindStringSubmatch(subject); m != nil {
		return m[1]
	}
	if m := landedSubjectForms.mergeColon.FindStringSubmatch(subject); m != nil {
		return m[1]
	}
	// Form 3b — an integration-targeted merge whose trailing group carries
	// exactly one card token. The target comparison is against the branch
	// the resolved landed ref names.
	if m := landedSubjectForms.mergeTarget.FindStringSubmatch(subject); m != nil && m[1] == landedBranch {
		if g := landedSubjectForms.trailingGroup.FindStringSubmatch(subject); g != nil {
			if toks := distinctCardTokens(g[1]); len(toks) == 1 {
				return toks[0]
			}
		}
	}
	return ""
}

// landedBranchFromRef derives the branch name the resolved landed ref names
// (REQ-TLA-013): `origin/develop` → `develop`, `origin/release/v9` →
// `release/v9`. The branch is never spelled anywhere in the predicate.
func landedBranchFromRef(ref string) string {
	r := strings.TrimSpace(ref)
	if r == "" {
		r = DefaultLandedRef
	}
	return strings.TrimPrefix(r, "origin/")
}

// LandedSubjectArgs builds the exact argv the landed check runs against ref.
//
// The query is a SUBJECT stream (`--format=%s`), never a whole-message
// stream: the predicate decides on attribution position within a subject, and
// a `%B` stream would feed commit bodies to it as if they were subjects. The
// ref is an INPUT rather than a package constant so the package stays a pure
// function of what it is given (prlink.go's NFR-4 ruling). An empty ref falls
// back to DefaultLandedRef rather than emitting an empty argv element, which
// git would read as the working tree.
//
// Exported so the tripwire can assert against the implementation's own
// construction rather than against a transcription of it — a test that
// rebuilds the argv itself proves nothing about the code (REQ-TLA-005,
// AC-TLA-006).
//
// @MX:NOTE: the single exported query builder for the landed check —
// REQ-TLA-005 requires exactly one; the tripwire test asserts this symbol's
// output, so a shape change here is what the tripwire sees.
func LandedSubjectArgs(ref string) ([]string, error) {
	if strings.TrimSpace(ref) == "" {
		ref = DefaultLandedRef
	}
	return []string{
		"log", ref,
		LandedSubjectFormatFlag,
	}, nil
}

// GitLandedQuerier answers Q2 against local git history. It performs no
// network I/O — the landed ref is an existing local ref — so it sits outside
// the one-`gh`-query budget entirely (NFR-2) and keeps working when `gh` does
// not (REQ-2.3).
type GitLandedQuerier struct {
	// Run executes git. Nil means the querier is unusable rather than
	// silently answering "not landed".
	Run CommandRunner
	// Ref is the ref the question is asked about. Empty means
	// DefaultLandedRef, which keeps an unconfigured project byte-identical to
	// the pre-resolution behaviour. Callers resolve it with LandedRefFor.
	Ref string
}

// LandedRef reports the ref this querier actually asks about, so a caller can
// name it in a refusal or a note without re-deriving it.
func (q GitLandedQuerier) LandedRef() string {
	if strings.TrimSpace(q.Ref) == "" {
		return DefaultLandedRef
	}
	return q.Ref
}

// Landed reports whether the resolved ref's history carries a subject that
// ATTRIBUTES the card — one of spec.md §A.4's six positional shapes, with the
// non-attribution rule applied. It does NOT report whether the card token
// appears anywhere in a message: a body mention, a mid-sentence mention, a
// branch name, and a dependency note attribute nothing (REQ-TLA-001..004).
//
// It returns a three-valued answer and an error, and NOTHING else (REQ-1.10).
// The subject stream is consumed here and discarded: no SHA and no subject
// escapes this function.
//
// An unanswerable query yields LandingUnknown alongside the error, so a caller
// that drops the error still cannot mistake it for not-landed (REQ-TLA-006).
func (q GitLandedQuerier) Landed(cardID string) (LandingAnswer, error) {
	ref := q.LandedRef()
	if q.Run == nil {
		return LandingUnknown, fmt.Errorf("prlink: landed querier has no command runner")
	}
	if !validCardToken.MatchString(cardID) {
		return LandingUnknown, fmt.Errorf("prlink: card id %q is not a bare token", cardID)
	}
	args, err := LandedSubjectArgs(ref)
	if err != nil {
		return LandingUnknown, err
	}
	out, err := q.Run("git", args...)
	if err != nil {
		return LandingUnknown, fmt.Errorf("prlink: git log %s: %w", ref, err)
	}
	branch := landedBranchFromRef(ref)
	for _, subject := range strings.Split(out, "\n") {
		if subjectAttribution(subject, branch) == cardID {
			return LandingLanded, nil
		}
	}
	return LandingNotLanded, nil
}
