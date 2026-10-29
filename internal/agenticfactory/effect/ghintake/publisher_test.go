package ghintake

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/intake"
)

var _ intake.Publisher = (*Publisher)(nil)

// hub is the only GitHub the tests see: it records every call and holds the Issues it "created".
type hub struct {
	calls   [][]string
	issues  []map[string]any // number, title, body, url
	failNew error
	lose    bool // the create lands but its answer is lost
	listErr error
}

func (h *hub) run(_ context.Context, args ...string) ([]byte, error) {
	h.calls = append(h.calls, args)
	switch args[0] + " " + args[1] {
	case "issue create":
		if h.failNew != nil {
			return nil, h.failNew
		}
		var title, body string
		for i, a := range args {
			if a == "--title" {
				title = args[i+1]
			}
			if a == "--body" {
				body = args[i+1]
			}
		}
		n := 100 + len(h.issues)
		h.issues = append(h.issues, map[string]any{"number": n, "title": title, "body": body, "url": "https://example.org/o/r/issues/" + itoa(n)})
		if h.lose {
			return nil, errors.New("connection reset")
		}
		return []byte("https://example.org/o/r/issues/" + itoa(n) + "\n"), nil
	case "issue list":
		if h.listErr != nil {
			return nil, h.listErr
		}
		search := ""
		for i, a := range args {
			if a == "--search" {
				search = args[i+1]
			}
		}
		term := strings.TrimSuffix(strings.TrimSuffix(search, " in:body"), " in:title")
		term = strings.Trim(term, `"`)
		var out []map[string]any
		for _, is := range h.issues {
			if strings.Contains(is["body"].(string), term) || strings.Contains(is["title"].(string), term) {
				out = append(out, is)
			}
		}
		raw, _ := json.Marshal(out)
		return raw, nil
	}
	return nil, errors.New("unexpected call " + strings.Join(args, " "))
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

type fixedIdentity struct {
	user    int64
	userErr error
	authors map[int]int64
}

func (f fixedIdentity) CurrentUserID(context.Context) (int64, error) { return f.user, f.userErr }
func (f fixedIdentity) IssueFacts(_ context.Context, _ string, n int) (effect.IssueFacts, error) {
	if id, ok := f.authors[n]; ok {
		return effect.IssueFacts{AuthorID: id, ActualRepo: "o/r"}, nil
	}
	return effect.IssueFacts{}, errors.New("unknown")
}
func (fixedIdentity) CommentFacts(context.Context, string, int64) (effect.CommentFacts, error) {
	return effect.CommentFacts{}, errors.New("unused")
}

func newPub(h *hub, allow bool) *Publisher {
	return &Publisher{
		GH:       GH{Repo: "o/r", Run: h.run, AllowWrite: allow, ListLimit: 100},
		Identity: fixedIdentity{user: 1001, authors: map[int]int64{}},
		Content:  func(context.Context, claim.Request) (string, string, error) { return "Add the thing", "Details.", nil },
		Limit:    5,
	}
}

func request() claim.Request { return claim.Request{Key: "req-abc-123", PayloadHash: "h1"} }

func TestWritesOffByDefaultAndProvablyNotSent(t *testing.T) {
	h := &hub{}
	_, _, err := newPub(h, false).PublishIssue(t.Context(), request())
	if !errors.Is(err, intake.ErrNotSent) {
		t.Fatalf("err = %v, want ErrNotSent so intake may retry", err)
	}
	if len(h.calls) != 0 {
		t.Errorf("gh ran with writes disabled: %v", h.calls)
	}
}

func TestPublishCreatesOneIssueWithTheRequestMarkerAndFindLocatesIt(t *testing.T) {
	h := &hub{}
	p := newPub(h, true)
	n, ref, err := p.PublishIssue(t.Context(), request())
	if err != nil || n != 100 || !strings.HasSuffix(ref, "/issues/100") {
		t.Fatalf("n=%d ref=%q err=%v", n, ref, err)
	}
	if len(h.issues) != 1 || !strings.Contains(h.issues[0]["body"].(string), RequestMarker("req-abc-123")) {
		t.Fatalf("issues = %+v", h.issues)
	}
	fn, fref, found, err := p.FindIssue(t.Context(), request())
	if err != nil || !found || fn != 100 || fref != ref {
		t.Errorf("find = %d %q %v %v", fn, fref, found, err)
	}
	if _, _, found, _ := p.FindIssue(t.Context(), claim.Request{Key: "other-request"}); found {
		t.Error("another request's marker was matched")
	}
}

func TestALostAnswerIsFoundByTheMarkerAndNeverCreatedTwice(t *testing.T) {
	h := &hub{lose: true}
	p := newPub(h, true)
	_, _, err := p.PublishIssue(t.Context(), request())
	if err == nil || errors.Is(err, intake.ErrNotSent) {
		t.Fatalf("a lost answer must be an unknown effect, not 'not sent': %v", err)
	}
	n, _, found, ferr := p.FindIssue(t.Context(), request())
	if ferr != nil || !found || n != 100 || len(h.issues) != 1 {
		t.Errorf("find = %d %v %v issues=%d", n, found, ferr, len(h.issues))
	}
}

func TestOnlyClientErrorsAreProvablyNotSent(t *testing.T) {
	for _, tc := range []struct {
		err     error
		notSent bool
	}{
		{&forge.StatusError{Status: 403}, true},
		{&forge.StatusError{Status: 422}, true},
		{&forge.StatusError{Status: 429}, true},
		{&forge.StatusError{Status: 502}, false},
		{errors.New("connection reset"), false},
	} {
		h := &hub{failNew: tc.err}
		_, _, err := newPub(h, true).PublishIssue(t.Context(), request())
		if err == nil || errors.Is(err, intake.ErrNotSent) != tc.notSent {
			t.Errorf("%v: err=%v notSent=%v, want %v", tc.err, err, errors.Is(err, intake.ErrNotSent), tc.notSent)
		}
	}
}

func TestTheGuardRefusalIsProvablyNotSentToo(t *testing.T) {
	h := &hub{}
	p := newPub(h, true)
	p.GH.Repo = "o/r/branches/main/protection/x"
	_, _, err := p.PublishIssue(t.Context(), request())
	if !errors.Is(err, intake.ErrNotSent) || !errors.Is(err, effect.ErrRefused) || len(h.calls) != 0 {
		t.Errorf("err=%v calls=%v", err, h.calls)
	}
}

func TestAuthenticatedUserReusesTheIdentityLookup(t *testing.T) {
	p := newPub(&hub{}, true)
	if id, err := p.AuthenticatedUser(t.Context()); err != nil || id != 1001 {
		t.Errorf("id=%d err=%v", id, err)
	}
	p.Identity = fixedIdentity{userErr: errors.New("boom")}
	if id, err := p.AuthenticatedUser(t.Context()); err == nil || id != 0 {
		t.Errorf("a failed lookup must give 0 and an error: id=%d err=%v", id, err)
	}
	p.Identity = nil
	if id, err := p.AuthenticatedUser(t.Context()); err == nil || id != 0 {
		t.Errorf("no identity wired: id=%d err=%v", id, err)
	}
}

func TestFindDuplicatesIsReadOnlyAndReportsNumericAuthors(t *testing.T) {
	h := &hub{issues: []map[string]any{
		{"number": 41, "title": "Add the thing", "body": "x", "url": "https://example.org/o/r/issues/41"},
		{"number": 42, "title": "Add the thing", "body": "y", "url": "https://example.org/o/r/issues/42"},
		{"number": 43, "title": "Something else", "body": "z", "url": "https://example.org/o/r/issues/43"},
	}}
	p := newPub(h, true)
	p.Identity = fixedIdentity{user: 1001, authors: map[int]int64{41: 2002}} // 42's author cannot be established
	got, err := p.FindDuplicates(t.Context(), request())
	if err != nil || len(got) != 2 {
		t.Fatalf("got = %+v err=%v", got, err)
	}
	byIssue := map[int]intake.Candidate{}
	for _, c := range got {
		byIssue[c.Issue] = c
	}
	if byIssue[41].AuthorID != 2002 || byIssue[42].AuthorID != 0 || !strings.HasSuffix(byIssue[41].Ref, "/issues/41") {
		t.Errorf("candidates = %+v", byIssue)
	}
	for _, c := range h.calls {
		if !readOnly(c) {
			t.Errorf("a duplicate search made a non-read call: %v", c)
		}
	}
	if len(h.issues) != 3 {
		t.Error("a duplicate search created or changed an Issue")
	}
}

func TestFindDuplicatesNeedsAConfiguredLimitAndFailsLoudly(t *testing.T) {
	h := &hub{listErr: errors.New("gh 502")}
	p := newPub(h, true)
	if _, err := p.FindDuplicates(t.Context(), request()); err == nil {
		t.Error("a failed search was reported as 'no duplicates'")
	}
	p.Limit = 0
	if _, err := p.FindDuplicates(t.Context(), request()); err == nil {
		t.Error("an unconfigured candidate limit was accepted")
	}
}

// A search that reaches the list limit may have missed the Issue or a duplicate: both are errors, never a
// confident "not found" that would create a second Issue.
func TestASearchAtTheListLimitIsAnErrorForFindIssueAndFindDuplicates(t *testing.T) {
	run := func(_ context.Context, args ...string) ([]byte, error) {
		return []byte(`[{"number":1,"title":"other","body":"x","url":"u1"},{"number":2,"title":"other","body":"y","url":"u2"}]`), nil
	}
	p := &Publisher{GH: GH{Repo: "o/r", Run: run, ListLimit: 2}, Limit: 5, Content: func(context.Context, claim.Request) (string, string, error) {
		return "wanted title", "body", nil
	}}
	req := claim.Request{Key: "k"}
	if _, _, found, err := p.FindIssue(t.Context(), req); err == nil || found {
		t.Errorf("FindIssue: found=%v err=%v", found, err)
	}
	if _, err := p.FindDuplicates(t.Context(), req); err == nil {
		t.Error("FindDuplicates answered from a list that reached the limit")
	}
	p.GH.ListLimit = 0
	if _, _, _, err := p.FindIssue(t.Context(), req); err == nil {
		t.Error("FindIssue ran without a configured ListLimit")
	}
}
