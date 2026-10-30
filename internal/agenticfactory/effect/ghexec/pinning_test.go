package ghexec

import (
	"errors"
	"strings"
	"testing"
)

// F12: gh flags have spellings the first allowlist did not read. `--web=true` opens a browser just like
// `--web`; `-R` selects a repository just like `--repo`; and a write is pinned to the configured repository
// whichever flag names it.
func TestEveryWebFlagSpellingIsRefused(t *testing.T) {
	e := ex()
	for _, a := range [][]string{
		{"pr", "view", "9", "--web"}, {"pr", "view", "9", "-w"},
		{"pr", "view", "9", "--web=true"}, {"issue", "view", "7", "--web=false"},
		{"issue", "list", "--web=1"},
	} {
		if e.allowed(t.Context(), a) {
			t.Errorf("%v was allowed", a)
		}
	}
}

func TestAnIssueCreateIsPinnedToTheConfiguredRepository(t *testing.T) {
	e := ex()
	e.AllowWrite = true
	create := func(extra ...string) []string {
		return append([]string{"issue", "create", "--title", "t", "--body", "b"}, extra...)
	}
	for name, tc := range map[string]struct {
		args []string
		want bool
	}{
		"--repo configured":    {create("--repo", "o/r"), true},
		"--repo= configured":   {create("--repo=o/r"), true},
		"-R configured":        {create("-R", "o/r"), true},
		"-R= configured":       {create("-R=o/r"), true},
		"case differs":         {create("--repo", "O/R"), true},
		"--repo other":         {create("--repo", "x/y"), false},
		"-R other":             {create("-R", "x/y"), false},
		"-R= other":            {create("-R=x/y"), false},
		"both, one other":      {create("--repo", "o/r", "-R", "x/y"), false},
		"both, same":           {create("--repo", "o/r", "-R", "o/r"), true},
		"no repo flag at all":  {create(), false},
		"repo with a subpath":  {create("--repo", "o/r/branches/main"), false},
		"repo with whitespace": {create("--repo", "o/r "), false},
	} {
		if got := e.allowed(t.Context(), tc.args); got != tc.want {
			t.Errorf("%s: allowed = %v, want %v (%v)", name, got, tc.want, tc.args)
		}
	}
	none := e
	none.Repo = ""
	if none.allowed(t.Context(), create("--repo", "o/r")) {
		t.Error("a write with no configured repository was allowed")
	}
}

// F8: the text of a failing gh call is stored and logged; it passes the one sanitizer first.
func TestAnExitErrorNeverEchoesACredentialOrPersonalData(t *testing.T) {
	secret := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	err := &ExitError{Verb: "pr view", Code: 1, Stderr: "fatal: https://x:" + secret + "@github.com denied for someone@example.com at /Users/someone/work"}
	var target *ExitError
	if !errors.As(err, &target) {
		t.Fatal("not an ExitError")
	}
	for _, bad := range []string{secret, "someone@example.com", "/Users/someone"} {
		if strings.Contains(err.Error(), bad) {
			t.Errorf("Error() holds %q: %s", bad, err.Error())
		}
	}
	if !strings.Contains(err.Error(), "pr view") || !strings.Contains(err.Error(), "exited 1") {
		t.Errorf("the verb and code are lost: %s", err.Error())
	}
}

// A bundle of short options can hide -w: `-cw` is `-c -w`.
func TestBundledShortOptionsCannotHideTheWebFlag(t *testing.T) {
	e := ex()
	for _, a := range []string{"-cw", "-wc", "-aw"} {
		if e.allowed(t.Context(), []string{"pr", "view", "9", a}) {
			t.Errorf("%s was allowed", a)
		}
	}
	if !e.allowed(t.Context(), []string{"pr", "view", "9", "-R", "o/r"}) {
		t.Error("a plain -R was refused")
	}
}
