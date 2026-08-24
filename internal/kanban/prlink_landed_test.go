// prlink_landed_test.go — SPEC-KANBAN-QUEUE-PR-SYNC-001 AC-011, AC-012 (M2).
//
// AC-011 carries controls in BOTH directions because the failure mode is a
// silent empty result: `\b` is not POSIX ERE, git does not error on it, and an
// `-E` regression reports every card as not-landed. A suite with only a
// negative control passes that regression cleanly, which is exactly the shape
// of an unobserved verification claim.
//
// Three controls, all required:
//
//	positive  — t199 returns landed, and the underlying query is non-empty
//	negative  — t205 returns no-link, and the query is empty
//	tripwire  — the same query under -E returns empty for t199, and the
//	            implementation's own result must NOT match that -E result
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
// t199 (twice, once through a report commit that merely mentions it) and
// never names t205.
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
	// commit. AC-012 is why the querier returns a boolean instead.
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
	// half, an -E regression makes every card report no-link and the whole
	// criterion passes on a result that observed nothing.
	args, err := LandedGrepArgs("t199")
	if err != nil {
		t.Fatalf("argv: %v", err)
	}
	raw, err := run("git", args...)
	if err != nil {
		t.Fatalf("positive control query: %v", err)
	}
	if strings.TrimSpace(raw) == "" {
		t.Fatalf("positive control: query returned an EMPTY commit set; the fixture or the engine flag is broken")
	}
	landed, err := q.Landed("t199")
	if err != nil {
		t.Fatalf("landed t199: %v", err)
	}
	if !landed {
		t.Errorf("t199 landed = false, want true (query returned %q)", strings.TrimSpace(raw))
	}

	// --- negative control ------------------------------------------------
	notLanded, err := q.Landed("t205")
	if err != nil {
		t.Fatalf("landed t205: %v", err)
	}
	if notLanded {
		t.Error("t205 landed = true; the fixture names no such commit")
	}

	// --- regex-engine tripwire -------------------------------------------
	// The implementation's own argv must carry --perl-regexp and must not
	// carry -E. Asserted against the built argv, not against a transcription
	// of it, so a change to the builder is what the test sees.
	if !slices.Contains(args, LandedRegexpEngineFlag) {
		t.Errorf("argv %v does not carry %s", args, LandedRegexpEngineFlag)
	}
	if slices.Contains(args, "-E") || slices.Contains(args, "--extended-regexp") {
		t.Errorf("argv %v uses POSIX ERE; \\b is unsupported there and the empty result is silent", args)
	}
	// And the behavioural half: the same query under -E returns empty, and
	// the implementation's result must NOT match that empty result on the
	// positive control.
	ere := slices.Clone(args)
	for i, a := range ere {
		if a == LandedRegexpEngineFlag {
			ere[i] = "-E"
		}
	}
	eraw, err := run("git", ere...)
	if err != nil {
		t.Fatalf("-E control query: %v", err)
	}
	if strings.TrimSpace(eraw) != "" {
		t.Skipf("this git build accepts \\b under -E (output %q); the tripwire cannot discriminate here", strings.TrimSpace(eraw))
	}
	if strings.TrimSpace(raw) == strings.TrimSpace(eraw) {
		t.Error("the implementation's result matches the -E result on the positive control — the engine flag is not taking effect")
	}
}

// AC-012 — the landed answer is a boolean and names no delivering commit
// (REQ-1.10), against a repository whose FIRST matching commit is a report
// commit that merely mentions the card.
func TestLandedCheck_BooleanOnly(t *testing.T) {
	dir := landedRepo(t)
	run := gitIn(dir)

	// Establish the premise: the newest match IS the report commit, so any
	// "first match is the delivering commit" reading attributes wrongly.
	args, _ := LandedGrepArgs("t199")
	raw, err := run("git", args...)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	first := strings.SplitN(strings.TrimSpace(raw), "\n", 2)[0]
	if !strings.Contains(first, "CHANGELOG") {
		t.Fatalf("fixture premise broken: newest match = %q, want the report commit", first)
	}

	// The querier's whole return surface is (bool, error). There is nothing
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

// A card id that is not a bare token is refused rather than interpolated into
// the git --grep pattern.
func TestLandedGrepArgs_RefusesNonToken(t *testing.T) {
	for _, bad := range []string{"", "t1 t2", `t1"`, "t1;rm -rf /", "t1\n"} {
		if _, err := LandedGrepArgs(bad); err == nil {
			t.Errorf("LandedGrepArgs(%q) = nil error, want a refusal", bad)
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
// engine-flag guard exists to prevent, arriving by a different route.
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
	if landed {
		t.Error("landed = true on a failed query")
	}
}

// The refusal reaches the querier too, not only the argv builder.
func TestLandedQuerier_RefusesNonToken(t *testing.T) {
	q := GitLandedQuerier{Run: func(string, ...string) (string, error) {
		t.Fatal("git was spawned for a refused card id")
		return "", nil
	}}
	if _, err := q.Landed("t1 t2"); err == nil {
		t.Error("Landed with a non-token card id = nil error, want a refusal")
	}
}
