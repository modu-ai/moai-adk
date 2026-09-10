// prlink_test.go — SPEC-KANBAN-QUEUE-PR-SYNC-001 AC-001..AC-003, AC-006..AC-008,
// AC-012 (the AC-011 controls live in prlink_landed_test.go).
//
// FIXTURE PROVENANCE — READ BEFORE EDITING.
//
// The pull-request fixtures below are TRANSCRIBED from the pinned block in
// acceptance.md, which was measured with:
//
//	gh pr list --state open --limit 40 --json number,title,body \
//	  -q '.[] | "\(.number)|TITLE:...|BODY:..."'
//
//	1614|TITLE:t203|BODY:t1 t151 t203 t69 t9
//	1612|TITLE:t200|BODY:t200 t201
//	1611|TITLE:|BODY:t201
//	1601|TITLE:|BODY:t188
//	1600|TITLE:|BODY:t184
//
// They are NOT re-fetched at test time. The live open set changes
// continuously — it has already grown since the measurement — and a test that
// queries GitHub is not a test: it is a non-deterministic assertion about
// somebody else's repository state.
package kanban

import (
	"errors"
	"reflect"
	"testing"
)

// pinnedOpenPRs is the measured open set, transcribed. Titles carry exactly
// the tokens the measurement recorded; the surrounding prose carries none.
func pinnedOpenPRs() []PRRecord {
	return []PRRecord{
		{Number: 1614, Title: "fix(kanban): dispatch ordering (t203)", Body: "closes t1 t151 t203 t69 t9", State: "OPEN"},
		{Number: 1612, Title: "feat(cli): queue render (t200)", Body: "t200 supersedes t201", State: "OPEN"},
		{Number: 1611, Title: "chore: tidy the lock sweep", Body: "part of t201", State: "OPEN"},
		{Number: 1601, Title: "docs: worktree lifecycle", Body: "card t188", State: "OPEN"},
		{Number: 1600, Title: "release: roll up the batch", Body: "delivering t184", State: "OPEN"},
	}
}

// pinned1600CommitTokens is the commit-token carriage of PR #1600, recorded
// so AC-006's premise is visible rather than assumed. The delivering card
// (t184) is ABSENT from it — the branch merged the release branch and
// inherited other cards' commits. That asymmetry is the whole reason commit
// messages are not a Q1 carrier.
var pinned1600CommitTokens = []string{"t165", "t82", "t81", "t173", "t175", "t167", "t176", "t174", "t83"}

// stubLanded answers Q2 from a fixed set, and records what it was asked.
type stubLanded struct {
	landed map[string]bool
	err    error
	asked  []string
}

func (s *stubLanded) Landed(cardID string) (LandingAnswer, error) {
	s.asked = append(s.asked, cardID)
	if s.err != nil {
		return LandingUnknown, s.err
	}
	if s.landed[cardID] {
		return LandingLanded, nil
	}
	return LandingNotLanded, nil
}

// AC-001 — an id in the PR title resolves `exact` (REQ-1.2).
func TestResolve_ExactFromTitle(t *testing.T) {
	got, err := ResolveCardPRLink("t200", pinnedOpenPRs(), &stubLanded{})
	if err != nil {
		t.Fatalf("resolve t200: %v", err)
	}
	if got.Kind != PRLinkLinked {
		t.Errorf("kind = %q, want %q", got.Kind, PRLinkLinked)
	}
	if !reflect.DeepEqual(got.PRs, []int{1612}) {
		t.Errorf("PRs = %v, want [1612]", got.PRs)
	}
	if got.Confidence != PRLinkExact {
		t.Errorf("confidence = %q, want %q", got.Confidence, PRLinkExact)
	}
	if got.PRState != "OPEN" {
		t.Errorf("pr_state = %q, want OPEN", got.PRState)
	}
}

// AC-002 — no title token, exactly one body token resolves `inferred`
// (REQ-1.3). t188 appears in no title and in exactly one body (#1601).
func TestResolve_InferredFromSingleBody(t *testing.T) {
	got, err := ResolveCardPRLink("t188", pinnedOpenPRs(), &stubLanded{})
	if err != nil {
		t.Fatalf("resolve t188: %v", err)
	}
	if got.Kind != PRLinkLinked {
		t.Errorf("kind = %q, want %q", got.Kind, PRLinkLinked)
	}
	if !reflect.DeepEqual(got.PRs, []int{1601}) {
		t.Errorf("PRs = %v, want [1601]", got.PRs)
	}
	if got.Confidence != PRLinkInferred {
		t.Errorf("confidence = %q, want %q", got.Confidence, PRLinkInferred)
	}
}

// AC-003 — several body tokens resolve `ambiguous`, never a best guess
// (REQ-1.4, REQ-1.6). t201 appears in no title and in TWO bodies.
func TestResolve_AmbiguousNotCollapsed(t *testing.T) {
	got, err := ResolveCardPRLink("t201", pinnedOpenPRs(), &stubLanded{})
	if err != nil {
		t.Fatalf("resolve t201: %v", err)
	}
	if got.Kind != PRLinkAmbiguous {
		t.Fatalf("kind = %q, want %q", got.Kind, PRLinkAmbiguous)
	}
	if !reflect.DeepEqual(got.PRs, []int{1611, 1612}) {
		t.Errorf("candidates = %v, want [1611 1612]", got.PRs)
	}
	if got.Confidence != "" {
		t.Errorf("ambiguous outcome carries confidence %q; it must carry none", got.Confidence)
	}
}

// AC-006 — commit tokens are not an attribution carrier (REQ-1.5).
//
// The structural half is that PRRecord has no commit field at all, so no
// resolver path can reach one. The behavioural half is below: a card present
// ONLY in #1600's commit tokens links to nothing.
func TestResolve_IgnoresCommitTokensForAttribution(t *testing.T) {
	for _, card := range pinned1600CommitTokens {
		if card == "t184" {
			t.Fatalf("fixture premise broken: t184 must NOT be among #1600's commit tokens")
		}
		got, err := ResolveCardPRLink(card, pinnedOpenPRs(), &stubLanded{})
		if err != nil {
			t.Fatalf("resolve %s: %v", card, err)
		}
		if got.Kind == PRLinkLinked || got.Kind == PRLinkAmbiguous {
			t.Errorf("%s resolved %q against PRs %v; commit tokens must not attribute", card, got.Kind, got.PRs)
		}
		for _, n := range got.PRs {
			if n == 1600 {
				t.Errorf("%s was attributed to #1600 via its commit tokens", card)
			}
		}
	}
}

// AC-007 — all four outcome kinds are mutually distinguishable (REQ-1.7).
func TestResolve_FourOutcomeKindsDistinct(t *testing.T) {
	landed := &stubLanded{landed: map[string]bool{"t199": true}}
	cases := map[string]PRLinkKind{
		"t200": PRLinkLinked,    // title hit
		"t201": PRLinkAmbiguous, // two body hits
		"t199": PRLinkLanded,    // no PR, in origin/main
		"t777": PRLinkNoLink,    // no PR, not landed
	}
	seen := map[PRLinkKind]string{}
	for card, want := range cases {
		got, err := ResolveCardPRLink(card, pinnedOpenPRs(), landed)
		if err != nil {
			t.Fatalf("resolve %s: %v", card, err)
		}
		if got.Kind != want {
			t.Errorf("%s kind = %q, want %q", card, got.Kind, want)
		}
		if prev, dup := seen[got.Kind]; dup {
			t.Errorf("%s and %s produced the same kind %q", prev, card, got.Kind)
		}
		seen[got.Kind] = card
	}
	if len(seen) != 4 {
		t.Errorf("distinct kinds = %d, want 4", len(seen))
	}
	// landed vs no-link is the distinction that cost a lane when it was
	// missed, so it is asserted on kind alone, not on any other field.
	if seen[PRLinkLanded] == seen[PRLinkNoLink] {
		t.Error("landed is indistinguishable from no-link")
	}
	// no-link is distinguishable from ambiguous BY KIND, not by an empty
	// candidate list on an ambiguous record.
	amb, _ := ResolveCardPRLink("t201", pinnedOpenPRs(), landed)
	none, _ := ResolveCardPRLink("t777", pinnedOpenPRs(), landed)
	if amb.Kind == none.Kind {
		t.Error("ambiguous and no-link share a kind")
	}
}

// AC-008 — whole-token matching (REQ-1.8). #1612's title carries t200; the
// card t20 must not match it.
func TestResolve_WholeTokenOnly(t *testing.T) {
	got, err := ResolveCardPRLink("t20", pinnedOpenPRs(), &stubLanded{})
	if err != nil {
		t.Fatalf("resolve t20: %v", err)
	}
	if got.Kind == PRLinkLinked || got.Kind == PRLinkAmbiguous {
		t.Errorf("t20 resolved %q against %v; t200 is not t20", got.Kind, got.PRs)
	}
	if len(got.PRs) != 0 {
		t.Errorf("t20 candidates = %v, want none", got.PRs)
	}
}

// AC-012 (structural half) — the resolver exposes no accessor that could
// return a delivering commit. The behavioural half runs against a real
// repository in prlink_landed_test.go.
func TestResolve_LandedCarriesNoCommit(t *testing.T) {
	landed := &stubLanded{landed: map[string]bool{"t199": true}}
	got, err := ResolveCardPRLink("t199", nil, landed)
	if err != nil {
		t.Fatalf("resolve t199: %v", err)
	}
	if got.Kind != PRLinkLanded {
		t.Fatalf("kind = %q, want %q", got.Kind, PRLinkLanded)
	}
	// Every field of the outcome record, enumerated: none may hold a SHA or a
	// subject. Adding one fails here rather than in review.
	typ := reflect.TypeOf(got)
	allowed := map[string]bool{"CardID": true, "Kind": true, "PRs": true, "PRState": true, "Confidence": true}
	for i := 0; i < typ.NumField(); i++ {
		if !allowed[typ.Field(i).Name] {
			t.Errorf("PRLinkOutcome gained field %q — the landed answer is a boolean (REQ-1.10)", typ.Field(i).Name)
		}
	}
}

// AC-TLS-004 / AC-TLS-005 — a landed-query failure degrades fail-open rather
// than aborting the render, and it degrades to `unknown`, NOT to `no-link`.
//
// This assertion was deliberately REVERSED by SPEC-TODO-LANDING-STATE-001: the
// former expectation (`no-link`) is the defect, because it renders a card that
// was never checked identically to one that was checked and found untouched.
func TestResolve_LandedErrorDegradesToUnknown(t *testing.T) {
	boom := errors.New("git unavailable")
	got, err := ResolveCardPRLink("t777", pinnedOpenPRs(), &stubLanded{err: boom})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want the querier's error", err)
	}
	if got.Kind != PRLinkUnknown {
		t.Errorf("kind = %q, want %q on a degraded landed query", got.Kind, PRLinkUnknown)
	}
	// The control that makes the above non-vacuous: a card whose query
	// SUCCEEDS and finds nothing still renders no-link, so the two outcomes
	// are genuinely distinguished rather than uniformly renamed.
	clean, cleanErr := ResolveCardPRLink("t777", pinnedOpenPRs(), &stubLanded{})
	if cleanErr != nil {
		t.Fatalf("resolve (answerable query): %v", cleanErr)
	}
	if clean.Kind != PRLinkNoLink {
		t.Errorf("kind = %q, want %q for an answered query that found nothing", clean.Kind, PRLinkNoLink)
	}
	if clean.Kind == got.Kind {
		t.Error("an unanswerable query and an answered-empty one render the same kind — the SPEC's whole delta")
	}
}

// The landed querier is consulted ONLY when no pull request carries the token
// (REQ-1.9) — a linked card costs no git process.
func TestResolve_LandedNotConsultedWhenLinked(t *testing.T) {
	landed := &stubLanded{}
	if _, err := ResolveCardPRLink("t200", pinnedOpenPRs(), landed); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(landed.asked) != 0 {
		t.Errorf("landed querier consulted %v for a title-linked card", landed.asked)
	}
}

// A card id that is not a bare token is refused rather than compiled into a
// pattern — the same guard the landed check applies to its --grep argument.
func TestResolve_RefusesNonToken(t *testing.T) {
	for _, bad := range []string{"", "t1 t2", "t1|t2", "../t1"} {
		out, err := ResolveCardPRLink(bad, pinnedOpenPRs(), &stubLanded{})
		if err == nil {
			t.Errorf("ResolveCardPRLink(%q) = nil error, want a refusal", bad)
		}
		if out.Kind != PRLinkNoLink {
			t.Errorf("%q kind = %q on refusal, want %q", bad, out.Kind, PRLinkNoLink)
		}
	}
}

// Two pull requests titling the SAME card is unmeasured and, on the measured
// set, absent — but it is genuinely ambiguous rather than exact, and REQ-1.6
// forbids picking a winner.
func TestResolve_TwoTitlesAreAmbiguousNotExact(t *testing.T) {
	prs := []PRRecord{
		{Number: 10, Title: "feat: half of t500", Body: "", State: "OPEN"},
		{Number: 11, Title: "feat: the rest of t500", Body: "", State: "OPEN"},
	}
	got, err := ResolveCardPRLink("t500", prs, &stubLanded{})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.Kind != PRLinkAmbiguous {
		t.Errorf("kind = %q, want %q — two titles is ambiguous, not exact", got.Kind, PRLinkAmbiguous)
	}
	if !reflect.DeepEqual(got.PRs, []int{10, 11}) {
		t.Errorf("candidates = %v, want [10 11]", got.PRs)
	}
}

// With no landed querier supplied the resolver reports no-link rather than
// panicking — the Q1-only caller is a legitimate shape.
func TestResolve_NilLandedQuerier(t *testing.T) {
	got, err := ResolveCardPRLink("t777", pinnedOpenPRs(), nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.Kind != PRLinkNoLink {
		t.Errorf("kind = %q, want %q", got.Kind, PRLinkNoLink)
	}
}
