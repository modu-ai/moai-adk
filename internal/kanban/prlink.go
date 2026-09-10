// prlink.go — the card-to-pull-request link resolver
// (SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.1..REQ-1.10, M1).
//
// The resolver answers TWO questions that look alike and behave differently,
// and keeping them apart is the whole design:
//
//	Q1 — attribution: "which card does this OPEN pull request deliver?"
//	     Carriers, in order: the PR title, then the PR body. Commit messages
//	     are NOT a carrier here (REQ-1.5) — a branch that merges the release
//	     branch inherits every other card's commits, so the noise scales with
//	     integration rather than with the card.
//
//	Q2 — landed: "is this card's work already in the branch this project
//	     integrates on?" The ref is RESOLVED from configuration
//	     (LandedRefFor), not hardcoded: asking origin/main in a project that
//	     integrates on develop answers a question nobody posed.
//	     Carrier: landed commit messages, and nothing else. The inherited-commit
//	     noise that ruins the commit carrier for Q1 is HARMLESS here, because a
//	     commit that rode in on another branch is still genuinely landed.
//
// The resolver is a pure function of (card id, PR record set, landed querier):
// no network, no filesystem, no repository (NFR-4). It writes nothing, ever —
// the read-only ruling in spec.md §B is a property of this package, not a
// convention its callers observe.
package kanban

import (
	"fmt"
	"regexp"
	"sort"
)

// PRLinkKind is one of the five mutually distinguishable outcome kinds
// (REQ-1.7; the fifth added by SPEC-TODO-LANDING-STATE-001 REQ-TLS-004). A
// consumer tells "already on the landed ref" from "nobody has started this"
// from "several candidates" from "the question could not be asked" by kind
// ALONE — never by inspecting whether some other field happens to be empty.
type PRLinkKind string

const (
	// PRLinkLinked marks a card attributed to exactly one open pull request.
	PRLinkLinked PRLinkKind = "linked"
	// PRLinkAmbiguous marks a card whose token appears in more than one open
	// pull request body. Every candidate is enumerated; none is chosen.
	PRLinkAmbiguous PRLinkKind = "ambiguous"
	// PRLinkLanded marks a card whose id is named by the resolved landed
	// ref's history. It is a boolean fact and nothing more (REQ-1.10).
	PRLinkLanded PRLinkKind = "landed"
	// PRLinkNoLink marks a card no open pull request and no landed commit
	// names — genuinely untouched work.
	PRLinkNoLink PRLinkKind = "no-link"
	// PRLinkUnknown marks a card whose landed question could NOT be asked —
	// the ref does not resolve, git is absent, the query failed. It is a
	// distinct kind rather than a degraded no-link precisely because the two
	// used to render identically: an unstarted card and an unanswerable one
	// are different facts (REQ-TLS-004).
	PRLinkUnknown PRLinkKind = "unknown"
)

// PRLinkConfidence labels how a linked outcome was reached (REQ-1.1). The set
// is closed: a caller may switch on it exhaustively.
type PRLinkConfidence string

const (
	// PRLinkExact marks a link read off a PR title. Measured precision on the
	// title carrier is 7/7 — every token present is the delivering card.
	PRLinkExact PRLinkConfidence = "exact"
	// PRLinkInferred marks a link read off a single PR body. The body carrier
	// has full recall and poor precision, so the label is honest about which
	// one answered.
	PRLinkInferred PRLinkConfidence = "inferred"
)

// PRRecord is one open pull request as the resolver needs it. It is exactly
// the shape `gh pr list --json number,title,body,state` returns, so the
// fixture form and the live form are the same type.
type PRRecord struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
}

// LandedQuerier answers Q2 and nothing else.
//
// The method returns a THREE-VALUED answer (REQ-1.10, REQ-TLS-003). It
// deliberately exposes no commit SHA and no commit subject, because a card's
// first matching commit may be another card's report commit that merely
// mentions it — so any "first match is the delivering commit" reading
// attributes wrongly. There is no accessor to add one through.
//
// The answer is a return VALUE rather than a flag on some other field, so a
// caller that ignores it does not compile.
type LandedQuerier interface {
	Landed(cardID string) (LandingAnswer, error)
}

// PRLinkOutcome is the single outcome record the resolver returns per card
// (REQ-1.1). Fields absent for a kind stay zero-valued, and a consumer keys
// off Kind rather than off emptiness.
type PRLinkOutcome struct {
	CardID string `json:"card_id"`
	// Kind is the load-bearing field: five values, mutually distinguishable.
	Kind PRLinkKind `json:"outcome"`
	// PRs carries the delivering pull request for a linked outcome and every
	// candidate for an ambiguous one, ascending. Empty for landed / no-link.
	PRs []int `json:"pr,omitempty"`
	// PRState is the state of the pull request(s) in PRs. The resolver is fed
	// the open set, so in practice it is the open state; it is carried rather
	// than assumed so a consumer never has to infer it.
	PRState string `json:"pr_state,omitempty"`
	// Confidence is populated for the linked kind only.
	Confidence PRLinkConfidence `json:"confidence,omitempty"`
}

// cardTokenPattern is the whole-token matcher (REQ-1.8). The word boundaries
// are what stop `t20` from matching a `t200` token: the character after
// `t20` in `t200` is a word character, so no boundary exists there.
func cardTokenPattern(cardID string) (*regexp.Regexp, error) {
	if !validCardToken.MatchString(cardID) {
		return nil, fmt.Errorf("prlink: card id %q is not a bare token", cardID)
	}
	pattern, err := regexp.Compile(`\b` + regexp.QuoteMeta(cardID) + `\b`)
	if err != nil {
		return nil, fmt.Errorf("prlink: card id %q: %w", cardID, err)
	}
	return pattern, nil
}

// validCardToken bounds what may be interpolated into a regular expression —
// here, and (load-bearing) into the git --grep pattern the landed check
// builds. Card ids are issued as `t<n>`; the class is a little wider than
// that so a renamed scheme does not silently fail closed.
var validCardToken = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// ResolveCardPRLink returns exactly one outcome record for one card.
//
// Order of resolution, and it is not interchangeable:
//
//  1. PR TITLE hit            → linked / exact       (REQ-1.2)
//  2. exactly one PR BODY hit → linked / inferred    (REQ-1.3)
//  3. several PR BODY hits    → ambiguous, every candidate enumerated,
//     never collapsed to a best guess (REQ-1.4, REQ-1.6)
//  4. no PR hit at all        → ask the landed querier (REQ-1.9):
//     at least one commit → landed; none → no-link; question unanswerable →
//     unknown (REQ-TLS-004)
//
// Commit messages are consulted at step 4 ONLY, through the querier, and only
// for Q2. Step 1-3 never see them (REQ-1.5).
//
// A landed-query failure degrades fail-open, but NOT into no-link: the
// outcome is `unknown` and the error is returned alongside it, so a caller can
// report the degradation without losing the rest of the render AND without a
// card that was never checked rendering as one that was.
func ResolveCardPRLink(cardID string, prs []PRRecord, landed LandedQuerier) (PRLinkOutcome, error) {
	out := PRLinkOutcome{CardID: cardID, Kind: PRLinkNoLink}

	re, err := cardTokenPattern(cardID)
	if err != nil {
		return out, err
	}

	var titleHits, bodyHits []PRRecord
	for _, pr := range prs {
		switch {
		case re.MatchString(pr.Title):
			titleHits = append(titleHits, pr)
		case re.MatchString(pr.Body):
			bodyHits = append(bodyHits, pr)
		}
	}

	// A title hit wins outright. Two pull requests titling the SAME card is
	// unmeasured and, on the measured set, impossible — but it is genuinely
	// ambiguous rather than exact, and REQ-1.6 forbids picking one, so it
	// falls through to the ambiguous shape rather than inventing a winner.
	if len(titleHits) == 1 {
		return linkedOutcome(out, titleHits[0], PRLinkExact), nil
	}
	if len(titleHits) > 1 {
		return ambiguousOutcome(out, titleHits), nil
	}
	if len(bodyHits) == 1 {
		return linkedOutcome(out, bodyHits[0], PRLinkInferred), nil
	}
	if len(bodyHits) > 1 {
		return ambiguousOutcome(out, bodyHits), nil
	}

	if landed == nil {
		return out, nil
	}
	answer, err := landed.Landed(cardID)
	if err != nil {
		out.Kind = PRLinkUnknown
		return out, err
	}
	switch answer {
	case LandingLanded:
		out.Kind = PRLinkLanded
	case LandingUnknown:
		// An answer of unknown WITHOUT an error still means the question was
		// not answered; it must not fall through to no-link.
		out.Kind = PRLinkUnknown
	}
	return out, nil
}

// linkedOutcome fills the linked shape: one pull request, its state, and the
// confidence label that says which carrier answered.
func linkedOutcome(out PRLinkOutcome, pr PRRecord, conf PRLinkConfidence) PRLinkOutcome {
	out.Kind = PRLinkLinked
	out.PRs = []int{pr.Number}
	out.PRState = pr.State
	out.Confidence = conf
	return out
}

// ambiguousOutcome enumerates every candidate and assigns no confidence —
// an ambiguous outcome is reported as ambiguous (REQ-1.6).
func ambiguousOutcome(out PRLinkOutcome, hits []PRRecord) PRLinkOutcome {
	out.Kind = PRLinkAmbiguous
	for _, pr := range hits {
		out.PRs = append(out.PRs, pr.Number)
	}
	sort.Ints(out.PRs)
	out.PRState = hits[0].State
	return out
}
