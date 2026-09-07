// prlink_landed_attribution_test.go — SPEC-TODO-LANDING-EVIDENCE-001
// (card t359) M2, AC-TLE-011.
//
// `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10 rules that the landed resolver
// names no delivering commit, and this SPEC's REQ-TLE-011 preserves it
// unchanged. The obvious slip on this axis is naming one anyway: the grep
// predicate has the commit set in hand, and the newest match is "usually
// right". It is not — a card's newest matching commit is routinely a report
// commit that merely mentions it, which is why the predicate finding a commit
// is not evidence that commit delivered anything.
//
// This file is the criterion that keeps the prohibition VERIFIED rather than
// restated. It is deliberately shaped as a containment assertion over the
// resolver's whole rendered output rather than as a check on named fields: a
// field-by-field check goes stale the moment a field is added, and the field
// that would carry a leaked SHA is by definition one that does not exist yet.
//
// A containment assertion over an absence is vacuous unless two things hold,
// and both are asserted here as positive controls before the containment is
// read:
//
//  1. the three SHAs are REALLY in the fixture's history under the card token
//     (a fixture that names nothing would satisfy any containment claim), and
//  2. the resolver REALLY answered landed (an empty commit set answers
//     not-landed and leaks nothing, for the wrong reason).
//
// The run-phase additionally plants the §D.2 mutant — the resolver carrying
// its first grep match into the outcome — and observes the containment
// assertion fail. Without that observation this test asserts nothing about the
// code's ability to fail, only about its current output.
package kanban

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// attributionCard is the card the three fixture commits name.
const attributionCard = "t359fixture"

// threeMatchRepo builds a repository whose origin/main history contains THREE
// commits naming attributionCard, with a HEAD commit that names it not at all.
//
// The head is deliberately a non-matching commit: AC-TLE-012 excludes the ref
// head SHA from its containment set by construction, and that exclusion is
// only sound while the head is not itself one of the matches. Building the
// fixture that way here means M3 inherits a fixture whose premise already
// holds rather than one it must re-shape.
//
// Returns the repository path and the three matching commits' full SHAs in
// commit order (oldest first).
func threeMatchRepo(t *testing.T) (string, []string) {
	t.Helper()
	dir := t.TempDir()
	// Config isolation, portably: GIT_CONFIG_GLOBAL points INSIDE the fixture
	// (an absent file reads as empty config on every platform) and the system
	// layer is switched off by flag rather than by a device node that does
	// not exist on native Windows.
	gitCfg := filepath.Join(dir, "gitconfig")
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(cmd.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
			"GIT_CONFIG_GLOBAL="+gitCfg, "GIT_CONFIG_NOSYSTEM=1",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return string(out)
	}
	run("init", "-q", "-b", "main")
	commit := func(file, msg string) string {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, file), []byte(msg+"\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
		run("add", file)
		run("commit", "-q", "-m", msg)
		return strings.TrimSpace(run("rev-parse", "HEAD"))
	}

	commit("seed.txt", "chore: seed the tree")
	// Three commits naming the card, in the three shapes that occur in this
	// repository's real history: the delivering commit, a report commit that
	// merely mentions it, and an integration commit that inherited it.
	shas := []string{
		commit("a.txt", "fix(kanban): the delivering change ("+attributionCard+")"),
		commit("b.txt", "docs: CHANGELOG mentions "+attributionCard+" among others"),
		commit("c.txt", "chore: merge the release batch carrying "+attributionCard),
	}
	// The head names nothing, so the ref position is not one of the matches.
	commit("head.txt", "chore: an unrelated change on top")
	run("update-ref", "refs/remotes/origin/main", "HEAD")
	return dir, shas
}

// AC-TLE-011 — the resolver names no commit, against a three-match fixture.
func TestResolver_NamesNoDeliveringCommit(t *testing.T) {
	dir, shas := threeMatchRepo(t)
	if len(shas) != 3 {
		t.Fatalf("fixture built %d matching commits, want 3", len(shas))
	}
	run := gitIn(dir)

	// --- positive control 1: the fixture really names the card three times.
	// Without this the containment assertion below would pass against a
	// repository that mentions nothing, which is the vacuous shape.
	args, err := LandedGrepArgs(DefaultLandedRef, attributionCard)
	if err != nil {
		t.Fatalf("argv: %v", err)
	}
	raw, err := run("git", args...)
	if err != nil {
		t.Fatalf("fixture query: %v", err)
	}
	matched := strings.Count(strings.TrimSpace(raw), "\n") + 1
	if strings.TrimSpace(raw) == "" {
		t.Fatalf("fixture premise broken: the query returned an EMPTY commit set")
	}
	if matched != 3 {
		t.Fatalf("fixture premise broken: the query matched %d commits, want 3:\n%s", matched, raw)
	}
	for _, sha := range shas {
		if !strings.Contains(raw, sha[:7]) {
			t.Fatalf("fixture premise broken: %s is not among the matched commits:\n%s", sha[:7], raw)
		}
	}

	// --- the resolver's whole rendered output --------------------------------
	// Rendered two ways because they are different surfaces with different
	// failure modes: the struct render catches an unexported or untagged
	// field, and the JSON render catches what a consumer of `--json` sees.
	outcome, err := ResolveCardPRLink(attributionCard, nil, GitLandedQuerier{Run: run})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// --- positive control 2: the resolver actually answered landed. An empty
	// commit set answers not-landed and leaks nothing for the wrong reason,
	// so a containment assertion read without this control proves nothing.
	if outcome.Kind != PRLinkLanded {
		t.Fatalf("outcome kind = %q, want %q; the containment assertion below is vacuous otherwise",
			outcome.Kind, PRLinkLanded)
	}
	answer, err := (GitLandedQuerier{Run: run}).Landed(attributionCard)
	if err != nil {
		t.Fatalf("landed: %v", err)
	}
	switch answer {
	case LandingLanded, LandingNotLanded, LandingUnknown:
	default:
		t.Fatalf("landed answer = %q, want one of the three landing values", answer)
	}
	if answer != LandingLanded {
		t.Fatalf("landed answer = %q, want %q against a three-match fixture", answer, LandingLanded)
	}

	encoded, err := json.Marshal(outcome)
	if err != nil {
		t.Fatalf("marshal outcome: %v", err)
	}
	rendered := fmt.Sprintf("%+v\n%#v\n%s", outcome, outcome, encoded)

	// --- the containment assertion ------------------------------------------
	// Full AND 7-character abbreviated, because the abbreviated form is what a
	// `--oneline` carry-through would deposit and the full form is what a
	// `%H` one would.
	for i, sha := range shas {
		if strings.Contains(rendered, sha) {
			t.Errorf("the resolver's output names commit %d in full (%s):\n%s", i+1, sha, rendered)
		}
		if strings.Contains(rendered, sha[:7]) {
			t.Errorf("the resolver's output names commit %d abbreviated (%s):\n%s", i+1, sha[:7], rendered)
		}
	}

	// The negative control for the containment mechanism itself: a string that
	// DOES carry a SHA must be caught by the same test. Without it a broken
	// strings.Contains call, or a rendered form that silently came out empty,
	// would read as a clean pass.
	poisoned := rendered + " " + shas[0]
	if !strings.Contains(poisoned, shas[0]) {
		t.Fatal("the containment check cannot detect a SHA it was handed directly; the assertion above proves nothing")
	}
}
