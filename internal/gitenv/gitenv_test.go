package gitenv

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// The scrub list this test asserts against is the test's OWN literal. Asserting
// against RepoScopingVars would pass by construction whatever that variable
// contained — including an empty slice, which is the exact defect.
var wantRemoved = []string{
	"GIT_DIR",
	"GIT_COMMON_DIR",
	"GIT_WORK_TREE",
	"GIT_INDEX_FILE",
	"GIT_OBJECT_DIRECTORY",
	"GIT_ALTERNATE_OBJECT_DIRECTORIES",
	"GIT_NAMESPACE",
	"GIT_PREFIX",
	"GIT_SUPER_PREFIX",
	"GIT_CEILING_DIRECTORIES",
	"GIT_QUARANTINE_PATH",
}

// names returns the NAME half of each NAME=VALUE entry.
func names(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		name, _, _ := strings.Cut(kv, "=")
		out = append(out, name)
	}
	return out
}

// os/exec reads a nil Env as "inherit the parent's environment" — the very
// behaviour this package exists to prevent. A caller that scrubs an empty
// environment and assigns the result must not thereby restore full inheritance,
// so the empty case is a correctness requirement, not a style preference.
func TestScrub_EmptyInputIsNonNil(t *testing.T) {
	if got := Scrub(nil); got == nil {
		t.Error("Scrub(nil) returned nil; os/exec reads a nil Env as full inheritance")
	}
	if got := Scrub([]string{}); got == nil {
		t.Error("Scrub(empty) returned nil; os/exec reads a nil Env as full inheritance")
	}
}

// Every repository-location variable is removed. Each is checked individually
// so a partial list fails naming the survivor, rather than failing as one
// opaque set mismatch.
func TestScrub_RemovesEveryRepoScopingVar(t *testing.T) {
	in := make([]string, 0, len(wantRemoved))
	for _, name := range wantRemoved {
		in = append(in, name+"=/somewhere")
	}

	got := names(Scrub(in))
	for _, name := range wantRemoved {
		if slices.Contains(got, name) {
			t.Errorf("%s survived the scrub", name)
		}
	}
	if len(got) != 0 {
		t.Errorf("scrubbing only repo-scoping vars should leave nothing; got %v", got)
	}
}

// Identity and behaviour variables reach the child, and a variable is not
// removed merely for starting with GIT_.
//
// This is the boundary the package draws, and it is load-bearing in both
// directions: a child stripped of its identity fails for an unrelated reason
// (turning a clean isolation failure into a confusing one), and a child
// stripped of its behaviour settings does something different rather than
// doing it somewhere different.
func TestScrub_KeepsIdentityAndBehaviour(t *testing.T) {
	keep := []string{
		"GIT_AUTHOR_NAME=A",
		"GIT_AUTHOR_EMAIL=a@e.invalid",
		"GIT_COMMITTER_NAME=C",
		"GIT_COMMITTER_EMAIL=c@e.invalid",
		"GIT_EDITOR=true",
		"GIT_PAGER=cat",
		"GIT_SSH_COMMAND=ssh",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_TERMINAL_PROMPT=0",
		"PATH=/usr/bin",
		"HOME=/home/x",
	}
	in := append([]string{"GIT_DIR=/leak"}, keep...)

	got := Scrub(in)
	for _, kv := range keep {
		if !slices.Contains(got, kv) {
			t.Errorf("%q was removed; only repository-location variables should be", kv)
		}
	}
	if slices.Contains(names(got), "GIT_DIR") {
		t.Error("GIT_DIR survived the scrub")
	}
}

// Order is preserved, so a later duplicate still wins the way it would have in
// the unscrubbed environment. A scrub that reordered entries would silently
// change which value a child resolves for a repeated name.
func TestScrub_PreservesOrder(t *testing.T) {
	in := []string{"A=1", "GIT_DIR=/leak", "B=2", "A=3"}
	want := []string{"A=1", "B=2", "A=3"}

	got := Scrub(in)
	if !slices.Equal(got, want) {
		t.Errorf("Scrub reordered or dropped entries: got %v, want %v", got, want)
	}
}

// An entry without "=" is passed through rather than guessed at. Dropping it
// would silently discard something the caller's environment carried; treating
// it as a name could remove an unrelated entry.
func TestScrub_PassesThroughNonPairEntries(t *testing.T) {
	in := []string{"NOT_A_PAIR", "GIT_DIR=/leak", "OK=1"}
	want := []string{"NOT_A_PAIR", "OK=1"}

	got := Scrub(in)
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// Env is Scrub over the live environment. Asserting on a variable this test
// sets itself keeps the check independent of whatever the ambient environment
// happens to hold.
func TestEnv_ScrubsLiveEnvironment(t *testing.T) {
	t.Setenv("GIT_DIR", "/leak")
	t.Setenv("GIT_AUTHOR_NAME", "t560")

	got := Env()
	if got == nil {
		t.Fatal("Env() returned nil; os/exec reads a nil Env as full inheritance")
	}
	if slices.Contains(names(got), "GIT_DIR") {
		t.Error("Env() carried GIT_DIR through")
	}
	if !slices.Contains(got, "GIT_AUTHOR_NAME=t560") {
		t.Error("Env() dropped GIT_AUTHOR_NAME; identity must survive")
	}

	// Nothing else was lost: the scrubbed environment differs from the live one
	// only by repo-scoping variables. Without this, an Env() that returned an
	// almost-empty slice would satisfy both assertions above.
	live := names(os.Environ())
	wantLen := 0
	for _, name := range live {
		if !slices.Contains(wantRemoved, name) {
			wantLen++
		}
	}
	if len(got) != wantLen {
		t.Errorf("Env() returned %d entries, want %d (live %d minus repo-scoping)",
			len(got), wantLen, len(live))
	}
}
