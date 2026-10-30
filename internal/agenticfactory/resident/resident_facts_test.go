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
	f, err := r.Runner.Facts.IssueFacts(issue7(t))
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
			if _, err := r.Runner.Facts.IssueFacts(issue7(t)); err == nil {
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
