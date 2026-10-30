package ghexec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ghidentity"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

var _ forge.Transport = Transport{}

// A fake gh that prints FAKE_GH_STDOUT (verbatim, via printf %b) to standard output, FAKE_GH_STDERR to
// standard error, and exits FAKE_GH_EXIT. Its arguments are recorded. The real gh never runs.
const fakeGHAPI = `#!/bin/sh
dir="$FAKE_GH_DIR"
echo spawned > "$dir/spawned"
: > "$dir/args"
for a in "$@"; do printf '%s\0' "$a" >> "$dir/args"; done
printf '%b' "$FAKE_GH_STDOUT"
printf '%s' "$FAKE_GH_STDERR" >&2
exit "${FAKE_GH_EXIT:-0}"
`

func newAPI(t *testing.T, stdout, stderr string, exit string) *fake {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(fakeGHAPI), 0o755); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_DIR", dir)
	t.Setenv("FAKE_GH_STDOUT", stdout)
	t.Setenv("FAKE_GH_STDERR", stderr)
	t.Setenv("FAKE_GH_EXIT", exit)
	return &fake{t: t, dir: dir}
}

func get(path string, hdr map[string]string) forge.Request {
	return forge.Request{Method: "GET", Path: path, Header: hdr}
}

func TestADocumentIsParsedIntoAResponse(t *testing.T) {
	f := newAPI(t, `HTTP/2.0 200 OK\r\nEtag: "abc"\r\nContent-Type: application/json\r\nX-Ratelimit-Remaining: 4999\r\n\r\n{"id":1}`, "", "0")
	res, err := Transport{Exec: ex()}.Do(t.Context(), get("/repos/o/r/issues/7", map[string]string{"if-none-match": `"abc"`}))
	if err != nil || res.Status != 200 || res.Redirected {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if res.Header["etag"] != `"abc"` || res.Header["x-ratelimit-remaining"] != "4999" || string(res.Body) != `{"id":1}` {
		t.Errorf("res = %+v", res)
	}
	args := f.args()
	if strings.Join(args[:3], " ") != "api -i /repos/o/r/issues/7" || !strings.Contains(strings.Join(args, "\x00"), "-H\x00if-none-match: \"abc\"") {
		t.Errorf("args = %q", args)
	}
}

func TestANotModifiedAnswerIsAResponseNotAnError(t *testing.T) {
	newAPI(t, `HTTP/2.0 304 Not Modified\r\nEtag: "abc"\r\n\r\n`, "gh: HTTP 304", "1")
	res, err := Transport{Exec: ex()}.Do(t.Context(), get("/repos/o/r/issues/7", nil))
	if err != nil || res.Status != 304 || len(res.Body) != 0 {
		t.Errorf("res=%+v err=%v", res, err)
	}
}

func TestAnErrorStatusIsAResponseTheClientClassifies(t *testing.T) {
	for status, line := range map[int]string{404: "404 Not Found", 502: "502 Bad Gateway"} {
		newAPI(t, `HTTP/2.0 `+line+`\r\n\r\n{"message":"x"}`, "gh: x (HTTP "+line[:3]+")", "1")
		res, err := Transport{Exec: ex()}.Do(t.Context(), get("/repos/o/r/issues/7", nil))
		if err != nil || res.Status != status {
			t.Errorf("%d: res=%+v err=%v", status, res, err)
		}
	}
}

func TestPermissionAnswersAreDenied(t *testing.T) {
	for _, line := range []string{"401 Unauthorized", "403 Forbidden"} {
		newAPI(t, `HTTP/2.0 `+line+`\r\n\r\n{"message":"Resource not accessible by integration"}`, "gh: x (HTTP "+line[:3]+")", "1")
		_, err := Transport{Exec: ex()}.Do(t.Context(), get("/repos/o/r/issues/7", nil))
		if !errors.Is(err, forge.ErrDenied) {
			t.Errorf("%s: err = %v, want forge.ErrDenied", line, err)
		}
	}
}

// A 429, and a 403 that carries rate-limit evidence (a retry-after, a zero remaining count, or the
// secondary-limit message), are not permission answers: the headers go to the client unchanged so it
// can wait.
func TestRateLimitAnswersPassThroughWithTheirHeaders(t *testing.T) {
	for name, tc := range map[string]struct {
		stdout string
		status int
		hdr    string
		val    string
	}{
		"429":                   {`HTTP/2.0 429 Too Many Requests\r\nRetry-After: 60\r\n\r\n{}`, 429, "retry-after", "60"},
		"403 retry-after":       {`HTTP/2.0 403 Forbidden\r\nRetry-After: 30\r\n\r\n{}`, 403, "retry-after", "30"},
		"403 remaining zero":    {`HTTP/2.0 403 Forbidden\r\nX-Ratelimit-Remaining: 0\r\nX-Ratelimit-Reset: 1700000000\r\n\r\n{}`, 403, "x-ratelimit-reset", "1700000000"},
		"403 secondary message": {`HTTP/2.0 403 Forbidden\r\n\r\n{"message":"You have exceeded a secondary rate limit."}`, 403, "", ""},
	} {
		newAPI(t, tc.stdout, "gh: limited (HTTP "+strings.Fields(tc.stdout)[1]+")", "1")
		res, err := Transport{Exec: ex()}.Do(t.Context(), get("/repos/o/r/issues/7", nil))
		if err != nil || res.Status != tc.status || (tc.hdr != "" && res.Header[tc.hdr] != tc.val) {
			t.Errorf("%s: res=%+v err=%v", name, res, err)
		}
	}
}

func TestNothingButAReadOnlyGetRuns(t *testing.T) {
	for name, r := range map[string]forge.Request{
		"POST":        {Method: "POST", Path: "/repos/o/r/issues/7/comments", Body: []byte(`{"body":"x"}`)},
		"PATCH":       {Method: "PATCH", Path: "/repos/o/r/issues/comments/5"},
		"GET + body":  {Method: "GET", Path: "/repos/o/r/issues/7", Body: []byte("x")},
		"header lf":   {Method: "GET", Path: "/repos/o/r/issues/7", Header: map[string]string{"x": "a\nb: c"}},
		"web in path": {Method: "GET", Path: "/repos/o/r/issues/7", Header: map[string]string{"x": "y"}, Body: nil},
	} {
		f := newAPI(t, "", "", "0")
		_, err := Transport{Exec: ex()}.Do(t.Context(), r)
		if name == "web in path" { // a plain clean request: control case, must run
			if err == nil && !f.spawned() {
				t.Errorf("%s: a clean GET did not run", name)
			}
			continue
		}
		if err == nil || f.spawned() {
			t.Errorf("%s: err=%v spawned=%v, want a refusal with no process", name, err, f.spawned())
		}
	}
}

func TestUnreadableOutputIsAnErrorNotAGuess(t *testing.T) {
	newAPI(t, `{"id":1}`, "", "0") // no -i preamble: not a response we can classify
	if _, err := (Transport{Exec: ex()}).Do(t.Context(), get("/repos/o/r/issues/7", nil)); err == nil {
		t.Error("output without a status line was accepted")
	}
	newAPI(t, "", "gh: not logged in", "4") // failed before any HTTP answer, no typed status
	if _, err := (Transport{Exec: ex()}).Do(t.Context(), get("/repos/o/r/issues/7", nil)); err == nil {
		t.Error("a failure with no HTTP answer was reported as a response")
	}
}

func TestATypedStatusWithNoPrintedHeadersStillClassifies(t *testing.T) {
	newAPI(t, "", "gh: Bad Gateway (HTTP 502)", "1")
	res, err := Transport{Exec: ex()}.Do(t.Context(), get("/repos/o/r/issues/7", nil))
	if err != nil || res.Status != 502 {
		t.Errorf("res=%+v err=%v", res, err)
	}
}

// B10 point 4, through the real components. Scope, read from core/authz.go: the comparison runs only on
// the branch for another author's Issue with collaboration on (edit, close, take over, label, assign).
// Own-Issue operations and comments on others' Issues return Allow before it, so a moved repository is not
// stopped there — a finding for the core owner, not something this adapter can fix.
// gh follows a redirect and shows only the final answer, so
// Redirected stays false — but the repository the response names differs from the one asked about, and
// the permission guard denies on that comparison.
func TestAMovedRepositoryIsCaughtByTheActualRepoComparison(t *testing.T) {
	newAPI(t, `HTTP/2.0 200 OK\r\nEtag: "a"\r\n\r\n{"user":{"id":2002},"repository_url":"https://api.github.com/repos/o/new-name"}`, "", "0")
	now := time.Unix(1_700_000_000, 0)
	g := &ghidentity.GitHub{Now: func() time.Time { return now }, Client: &forge.Client{
		T: Transport{Exec: ex()}, B: &forge.Budget{}, Now: func() time.Time { return now },
		Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3},
	}}
	facts, err := g.IssueFacts(t.Context(), "o/r", 7)
	if err != nil || facts.Redirected || facts.ActualRepo != "o/new-name" {
		t.Fatalf("facts = %+v err=%v", facts, err)
	}
	c := core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"o/r", "o/new-name"}, Close: true}
	v := c.Authorize(core.AuthzRequest{Op: core.OpClose, RequestedRepo: "o/r", ActualRepo: facts.ActualRepo, Redirected: facts.Redirected, UserID: 1001, IssueAuthorID: facts.AuthorID}) // another author's Issue: the branch that compares repositories
	if v != core.DenyRepoNotAllowed {
		t.Errorf("verdict = %s, want %s: an answer for another repository must not authorize", v, core.DenyRepoNotAllowed)
	}
}
