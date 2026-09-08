// prlink_landed_test.go — SPEC-KANBAN-QUEUE-PR-SYNC-001 AC-011, AC-012 (M2).
//
// AC-011 carries controls in BOTH directions because the failure mode is a
// silent empty result: `\b` is not POSIX ERE, git does not error on it, and a
// regression reports every card as not-landed. A suite with only a negative
// control passes that regression cleanly, which is exactly the shape of an
// unobserved verification claim.
//
// SPEC-TODO-LANDING-ATTRIBUTION-001 replaced the whole-message grep with a
// subject-stream query built by LandedSubjectArgs. The -E engine-flag
// tripwire is REPLACED (plan.md M3 item 1) by the shape tripwire: the failure
// mode this file was written about now lives in the query SHAPE — a %B
// regression would feed commit bodies to the positional matcher as if they
// were subjects, which is the same silent re-admission by another door.
//
// Three controls, all required:
//
//	positive  — t199 returns landed, and the underlying subject stream is
//	            non-empty and attributes t199
//	negative  — t205 returns not-landed, and the stream carries no such card
//	tripwire  — the built argv carries the subject format and no whole-message
//	            format or --grep; and the %B stream provably differs from the
//	            %s stream, so the shape is what the verdict rests on
package kanban

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// landedRepo builds a throwaway repository whose origin/main history names
// t199 (once through a form-2 trailing attribution, once through a report
// commit that merely mentions it) and never names t205.
func landedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Config isolation, portably: GIT_CONFIG_GLOBAL points at a path INSIDE the
	// fixture (an absent file reads as empty config on every platform), and the
	// system layer is switched off by flag rather than by pointing it at a
	// device node. `/dev/null` works on Unix and does not exist on native
	// Windows, where it would leave the developer's real config in play.
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
	commit := func(file, msg string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, file), []byte(msg+"\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
		run("add", file)
		run("commit", "-q", "-m", msg)
	}
	commit("a.txt", "chore: seed the tree")
	commit("b.txt", "fix(web): register signal handling before binding the listener (t199)")
	// A REPORT commit that merely mentions t199. Ordering matters: git log
	// returns newest first, so this is the FIRST match a naive "the first hit
	// is the delivering commit" reading would return — and it is the wrong
	// commit. AC-012 is why the querier returns a verdict instead.
	commit("c.txt", "docs: update CHANGELOG mentioning t199 among others")
	// origin/main is the ref the landed question is asked about, so the
	// fixture provides it rather than relying on a remote.
	run("update-ref", "refs/remotes/origin/main", "HEAD")
	return dir
}

// gitIn returns a CommandRunner bound to one repository.
func gitIn(dir string) CommandRunner {
	return func(name string, args ...string) (string, error) {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		out, err := cmd.Output()
		return string(out), err
	}
}

// AC-011 — the landed check works, and cannot pass vacuously (REQ-1.9).
func TestLandedCheck_Controls(t *testing.T) {
	dir := landedRepo(t)
	run := gitIn(dir)
	q := GitLandedQuerier{Run: run}

	// --- positive control ------------------------------------------------
	// The card is landed AND the underlying query is non-empty. Without this
	// half, a silent-empty regression makes every card report no-link and the
	// whole criterion passes on a result that observed nothing.
	args, err := LandedSubjectArgs(DefaultLandedRef)
	if err != nil {
		t.Fatalf("argv: %v", err)
	}
	raw, err := run("git", args...)
	if err != nil {
		t.Fatalf("positive control query: %v", err)
	}
	if strings.TrimSpace(raw) == "" {
		t.Fatalf("positive control: query returned an EMPTY subject stream — the fixture is broken")
	}
	landed, err := q.Landed("t199")
	if err != nil {
		t.Fatalf("landed t199: %v", err)
	}
	if landed != LandingLanded {
		t.Errorf("t199 landed = %q, want %q (query returned %q)", landed, LandingLanded, strings.TrimSpace(raw))
	}

	// --- negative control ------------------------------------------------
	notLanded, err := q.Landed("t205")
	if err != nil {
		t.Fatalf("landed t205: %v", err)
	}
	if notLanded != LandingNotLanded {
		t.Errorf("t205 landed = %q, want %q; the fixture names no such commit", notLanded, LandingNotLanded)
	}

	// --- query-shape tripwire ---------------------------------------------
	// The implementation's own argv must carry the subject-stream format and
	// must not carry a whole-message format or a --grep element (an occurrence
	// test, not a position). Asserted against the built argv, not against a
	// transcription of it, so a change to the builder is what the test sees.
	if !slices.Contains(args, LandedSubjectFormatFlag) {
		t.Errorf("argv %v does not carry %s", args, LandedSubjectFormatFlag)
	}
	if slices.Contains(args, "--format=%B") || slices.Contains(args, "--grep") {
		t.Errorf("argv %v queries the whole message; commit bodies would be fed to the matcher as subjects", args)
	}
	// And the behavioural half: the same query under --format=%B returns a
	// DIFFERENT stream, so the shape is what the verdict rests on — the
	// streams must not be byte-identical, or the tripwire proves nothing.
	bargs := slices.Clone(args)
	for i, a := range bargs {
		if a == LandedSubjectFormatFlag {
			bargs[i] = "--format=%B"
		}
	}
	braw, err := run("git", bargs...)
	if err != nil {
		t.Fatalf("%%B control query: %v", err)
	}
	if strings.TrimSpace(braw) == "" {
		t.Skipf("this fixture's %%B stream is empty (query %v); the tripwire cannot discriminate here", bargs)
	}
	if strings.TrimSpace(raw) == strings.TrimSpace(braw) {
		t.Error("the subject stream is byte-identical to the whole-message stream — the format flag is not taking effect")
	}
}

// AC-012 — the landed answer carries no delivering commit, against a
// repository whose FIRST matching commit is a report commit that merely
// mentions the card.
func TestLandedCheck_BooleanOnly(t *testing.T) {
	dir := landedRepo(t)
	run := gitIn(dir)

	// Establish the premise: the newest SUBJECT naming t199 is the report
	// commit, so any "first match is the delivering commit" reading
	// attributes wrongly.
	args, argvErr := LandedSubjectArgs(DefaultLandedRef)
	if argvErr != nil {
		t.Fatalf("argv: %v", argvErr)
	}
	raw, err := run("git", args...)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	first := strings.SplitN(strings.TrimSpace(raw), "\n", 2)[0]
	if !strings.Contains(first, "CHANGELOG") {
		t.Fatalf("fixture premise broken: newest subject = %q, want the report commit", first)
	}

	// The querier's whole return surface is (verdict, error). There is nothing
	// to leak a SHA through, and the resolver's outcome record carries none.
	out, err := ResolveCardPRLink("t199", nil, GitLandedQuerier{Run: run})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if out.Kind != PRLinkLanded {
		t.Fatalf("kind = %q, want %q", out.Kind, PRLinkLanded)
	}
	sha := strings.Fields(first)[0]
	for _, field := range []string{out.CardID, string(out.Kind), out.PRState, string(out.Confidence)} {
		if field != "" && strings.Contains(field, sha) {
			t.Errorf("outcome field %q carries the commit sha %s", field, sha)
		}
	}
}

// A card id that is not a bare token is refused by the querier rather than
// reaching the query.
func TestLandedQuerier_RefusesNonToken(t *testing.T) {
	for _, bad := range []string{"", "t1 t2", `t1"`, "t1;rm -rf /", "t1\n"} {
		if _, err := (GitLandedQuerier{Run: gitIn(t.TempDir())}).Landed(bad); err == nil {
			t.Errorf("Landed(%q) = nil error, want a refusal", bad)
		}
	}
}

// A querier with no runner reports an error rather than answering "not
// landed" — the absence of a signal is not evidence of absence.
func TestLandedQuerier_NoRunnerErrors(t *testing.T) {
	if _, err := (GitLandedQuerier{}).Landed("t199"); err == nil {
		t.Error("Landed with no runner = nil error, want a refusal")
	}
}

// A failing git process is an ERROR, never a quiet false. Reporting "not
// landed" for a query that never ran is the same silent-empty failure the
// query-shape guard exists to prevent, arriving by a different route.
func TestLandedQuerier_GitFailureIsNotFalse(t *testing.T) {
	boom := errors.New("exit status 128: not a git repository")
	q := GitLandedQuerier{Run: func(string, ...string) (string, error) { return "", boom }}
	landed, err := q.Landed("t199")
	if err == nil {
		t.Fatal("a failing git process reported nil error")
	}
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap the runner's error", err)
	}
	// A failed query is UNKNOWN, never not-landed: collapsing the two makes
	// "the check could not run" and "the check ran and found nothing" the
	// same value (REQ-TLS-003, REQ-TLS-004, REQ-TLA-006).
	if landed != LandingUnknown {
		t.Errorf("landed = %q on a failed query, want %q", landed, LandingUnknown)
	}
}
