package resident

// resident_facts_test.go — the Runner's host Facts are wired in Build (Issue #1948, audit F2, B10).
//
// Defect answer: these tests catch a Resident whose Runner has no Facts even though a transport was given
// (every authority check then ends as an unknown identity), a Runner that gets Facts it should not have when
// no transport exists, a lookup that writes, a repository other than the one the settings name, an Issue
// author or acting user that is not what the host said, and an unbounded lookup. The transport is a fake;
// nothing reaches a network, a process or a real repository.

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

type fakeHost struct {
	mu     sync.Mutex
	reqs   []forge.Request
	user   string // body of GET /user
	issue  string // body of GET /repos/o/a/issues/7
	fail   bool
	status int
}

func (f *fakeHost) Do(_ context.Context, r forge.Request) (forge.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reqs = append(f.reqs, r)
	if f.fail {
		return forge.Response{}, errors.New("transport down")
	}
	st := f.status
	if st == 0 {
		st = http.StatusOK
	}
	h := map[string]string{"x-ratelimit-remaining": "4000", "x-ratelimit-resource": "core"}
	switch {
	case r.Path == "/user":
		return forge.Response{Status: st, Header: h, Body: []byte(f.user)}, nil
	case strings.HasPrefix(r.Path, "/repos/o/a/issues/"):
		return forge.Response{Status: st, Header: h, Body: []byte(f.issue)}, nil
	}
	return forge.Response{Status: http.StatusNotFound, Header: h}, nil
}

func (f *fakeHost) paths() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var p []string
	for _, r := range f.reqs {
		p = append(p, r.Method+" "+r.Path)
	}
	return p
}

func buildWithTransport(t *testing.T, tr forge.Transport) *Resident {
	t.Helper()
	c := &clock{now: t0}
	d := deps(t, c)
	d.Transport = tr
	r, err := BuildFromProject(project(t, goodYAML), d)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func issue7(t *testing.T) core.IssueRef {
	t.Helper()
	i, err := core.NewIssueRef(7)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

const (
	userBody  = `{"id": 42}`
	issueBody = `{"user": {"id": 77}, "repository_url": "https://api.github.com/repos/o/a"}`
)

// Without a transport the Runner has no Facts, and the guard answers an unknown identity: the fail-closed state.
func TestRunnerHasNoFactsWithoutATransport(t *testing.T) {
	r := buildWithTransport(t, nil)
	if r.Runner == nil {
		t.Fatal("Build left the Runner nil")
	}
	if r.Runner.Facts != nil {
		t.Fatalf("Facts = %T without a transport, want nil", r.Runner.Facts)
	}
	run, _ := core.ParseRunID("run-1")
	if _, err := r.Runner.Claim(issue7(t), "w1", "d1", run); !errors.Is(err, core.ErrIdentityUnknown) {
		t.Errorf("Claim without Facts: %v, want ErrIdentityUnknown", err)
	}
}

// With a transport the Facts are the host's answers, for the repository the settings name, read-only.
func TestRunnerFactsComeFromTheHostLookups(t *testing.T) {
	host := &fakeHost{user: userBody, issue: issueBody}
	r := buildWithTransport(t, host)
	if r.Runner.Facts == nil {
		t.Fatal("Facts are nil although a transport was given")
	}
	f, err := r.Runner.Facts.IssueFacts(context.Background(), issue7(t))
	if err != nil {
		t.Fatal(err)
	}
	if f.UserID != 42 || f.AuthorID != 77 {
		t.Errorf("user %d author %d, want 42 and 77 (the host's answers)", f.UserID, f.AuthorID)
	}
	if f.RequestedRepo != "o/a" || f.ActualRepo != "o/a" || f.Redirected {
		t.Errorf("repository requested %q actual %q redirected %v, want o/a, o/a, false", f.RequestedRepo, f.ActualRepo, f.Redirected)
	}
	if f.Owner != core.OwnerNone {
		t.Errorf("owner = %v, want none (no claim record exists)", f.Owner)
	}
	// Reads only, and only the two lookups.
	want := []string{"GET /user", "GET /repos/o/a/issues/7"}
	got := host.paths()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("host requests = %v, want %v", got, want)
	}
}

// Every shortfall is an error and so an unknown identity - never a guessed fact.
func TestRunnerFactsRefuseAShortfall(t *testing.T) {
	for name, host := range map[string]*fakeHost{
		"transport down":        {fail: true},
		"user has no id":        {user: `{"login":"x"}`, issue: issueBody},
		"author has no id":      {user: userBody, issue: `{"user":{},"repository_url":"https://api.github.com/repos/o/a"}`},
		"user id not a number":  {user: `{"id":"abc"}`, issue: issueBody},
		"unreadable user body":  {user: `not json`, issue: issueBody},
		"unreadable issue body": {user: userBody, issue: `not json`},
	} {
		t.Run(name, func(t *testing.T) {
			r := buildWithTransport(t, host)
			if _, err := r.Runner.Facts.IssueFacts(context.Background(), issue7(t)); err == nil {
				t.Error("a shortfall was answered as a fact")
			}
		})
	}
}

// The lookup is bounded by the named gh-read timeout, not by an unbounded context.
func TestRunnerFactsAreBounded(t *testing.T) {
	r := buildWithTransport(t, &fakeHost{user: userBody, issue: issueBody})
	hf, ok := r.Runner.Facts.(runner.HostFacts)
	if !ok {
		t.Fatalf("Facts = %T, want runner.HostFacts", r.Runner.Facts)
	}
	if hf.Timeout != checkObserveTimeout || hf.Timeout <= 0 || hf.Timeout > time.Minute {
		t.Errorf("lookup timeout = %v, want the named gh-read bound %v", hf.Timeout, checkObserveTimeout)
	}
	if hf.Repo != "o/a" {
		t.Errorf("repository = %q, want the settings' pr_source o/a", hf.Repo)
	}
}

// A monitor pass makes no host request: the Facts exist but only an authority check (a claim) reads them. A
// Waiter or notify method that starts calling the guard would otherwise begin using the network quietly.
func TestAPassMakesNoHostRequest(t *testing.T) {
	host := &fakeHost{user: userBody, issue: issueBody}
	r := buildWithTransport(t, host)
	if r.Runner.Facts == nil {
		t.Fatal("control: the Facts are wired")
	}
	if _, err := r.Service.Tick(context.Background()); err != nil {
		t.Fatalf("a pass: %v", err)
	}
	if got := host.paths(); len(got) != 0 {
		t.Errorf("a pass made host requests %v, want none", got)
	}
	// positive control: the same fake does see a request when the Facts are read.
	if _, err := r.Runner.Facts.IssueFacts(context.Background(), issue7(t)); err != nil {
		t.Fatal(err)
	}
	if len(host.paths()) == 0 {
		t.Error("control: reading the Facts made no request, so the zero above proves nothing")
	}
}

// The same property on a pass that really walks the reminder path: an overdue wait record is planted, the
// Waiter retries it (the path that calls into the Runner), and still no host request is made. An empty store
// would not reach that path (lane-19 P3).
func TestAPassOverAnOverdueWaitMakesNoHostRequest(t *testing.T) {
	host := &fakeHost{user: `{"id": 42}`, issue: `{"user": {"id": 42}, "repository_url": "https://api.github.com/repos/o/a"}`}
	r := buildWithTransport(t, host)
	run, _ := core.ParseRunID("run-1")
	tk, err := r.Runner.Claim(issue7(t), "w1", "d-1", run) // own Issue: author 42 == user 42
	if err != nil {
		t.Fatalf("claim over the fake host: %v", err)
	}
	claimRequests := len(host.paths())
	if claimRequests == 0 {
		t.Fatal("control: the claim made no host request, so the guard did not run")
	}
	w := r.NewWait("w1", "tell the owner", "waiting on a reply", t0.Add(-time.Hour))
	if _, err := r.Runner.SetWait(tk, w); err != nil {
		t.Fatalf("plant the overdue wait: %v", err)
	}
	acts, err := r.Service.Tick(context.Background())
	if err != nil {
		t.Fatalf("a pass: %v", err)
	}
	walked := false
	for _, a := range acts {
		if a.Issue == tk.Issue.Number {
			walked = true
		}
	}
	if !walked {
		t.Fatalf("control: the pass did nothing for the overdue wait (actions %+v), so the reminder path was not walked", acts)
	}
	if got := len(host.paths()); got != claimRequests {
		t.Errorf("the pass made %d host requests beyond the claim's %d", got-claimRequests, claimRequests)
	}
}

// The acting user is looked up once while the identity cache holds it: a second authority check does not
// repeat GET /user (audit lane-19 point 3: no extra reduction is needed, the cache age does it).
func TestTheActingUserIsLookedUpOnceWithinTheCacheAge(t *testing.T) {
	host := &fakeHost{user: userBody, issue: issueBody}
	r := buildWithTransport(t, host)
	for i := 0; i < 3; i++ {
		if _, err := r.Runner.Facts.IssueFacts(context.Background(), issue7(t)); err != nil {
			t.Fatal(err)
		}
	}
	users := 0
	for _, p := range host.paths() {
		if p == "GET /user" {
			users++
		}
	}
	if users != 1 {
		t.Errorf("GET /user was made %d times over 3 authority checks, want 1 (identity.max_age holds it)", users)
	}
}

// A lookup with no time limit is refused before any request is made (the wired Timeout is non-zero, and a
// zero one cannot reach the host).
func TestAZeroLookupTimeoutIsRefusedBeforeAnyRequest(t *testing.T) {
	host := &fakeHost{user: userBody, issue: issueBody}
	r := buildWithTransport(t, host)
	hf, ok := r.Runner.Facts.(runner.HostFacts)
	if !ok {
		t.Fatalf("Facts = %T, want runner.HostFacts", r.Runner.Facts)
	}
	if hf.Timeout <= 0 {
		t.Fatalf("control: the wired Timeout is %v, want positive", hf.Timeout)
	}
	hf.Timeout = 0
	if _, err := hf.IssueFacts(context.Background(), issue7(t)); err == nil {
		t.Error("a zero Timeout was accepted")
	}
	if got := host.paths(); len(got) != 0 {
		t.Errorf("a refused lookup still made requests %v", got)
	}
}

// The Attributor stays at its lock constant: identity.Book is not connected (REQ-AF-011).
func TestRunnerAttributorStaysAtTheLockConstant(t *testing.T) {
	r := buildWithTransport(t, &fakeHost{user: userBody, issue: issueBody})
	if r.Runner.Attributor != nil {
		t.Errorf("Attributor = %T, want nil (the runner's lock constant applies)", r.Runner.Attributor)
	}
	// the Waits path of the notify service is the same Runner, so one Runner exists per process.
	if r.Service.Waits != any(r.Runner) {
		t.Error("the notify service's Waits is not the Resident's Runner")
	}
}
