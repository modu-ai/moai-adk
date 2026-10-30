package resident

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge/webhook"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// Settings come from a file in a temp directory; the claim store is a temp directory; the clock, the gh
// runner and the transport are fakes. Nothing reaches a network, a process or a real repository.

const goodYAML = `workflow:
  agentic_factory:
    enabled: true
    claim_store:
      dir: ".moai/state/af"
    provider:
      github:
        backoff:
          min_secondary_wait: 61s
          base: 3s
          max: 11m
          max_attempts: 5
    collector:
      poll:
        min: 12s
        max: 6m
        safety: 31m
    budget:
      reserve: 7
      report_resource: core
    identity:
      max_age: 17m
    labels:
      per_page: 53
    intake:
      duplicate_search_limit: 6
      list_limit: 211
    monitor:
      pr_list_limit: 307
    cursor:
      max_doc_bytes: 9000
      max_chunks: 13
    queue:
      webhook:
        max_doc_bytes: 7000
        seen_limit: 5
        max_id_bytes: 40
    notify:
      inbox:
        max_doc_bytes: 8000
        seen_limit: 4
        max_id_bytes: 32
      renotify_after: 90m
    webhook:
      max_body: 4321
      provider: github
      routes:
        - repository: o/a
        - repository: o/b
    wait:
      max_retries: 4
      escalate_to: leader
    pr_source:
      repository: o/a
    context:
      summary_max_bytes: 2048
`

func project(t *testing.T, yaml string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agentic-factory.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

var t0 = time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)

const secretValue = "test-only-secret"

func deps(t *testing.T, c *clock) Deps {
	t.Helper()
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	return Deps{
		Store: st, Now: c.Now,
		GHRead: func(context.Context, ...string) ([]byte, error) { return nil, errors.New("no gh in tests") },
		Secret: func(string) []byte { return []byte(secretValue) },
	}
}

func build(t *testing.T) (*Resident, *clock) {
	t.Helper()
	c := &clock{now: t0}
	r, err := BuildFromProject(project(t, goodYAML), deps(t, c))
	if err != nil {
		t.Fatal(err)
	}
	return r, c
}

// Every setting reaches the component field it is for. The values are all different, so a swapped pair
// of fields fails here.
func TestEverySettingReachesItsComponent(t *testing.T) {
	r, _ := build(t)
	eq := func(what string, got, want any) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %v, want %v", what, got, want)
		}
	}
	eq("backoff.MinSecondaryWait", r.Backoff.MinSecondaryWait, 61*time.Second)
	eq("backoff.Base", r.Backoff.Base, 3*time.Second)
	eq("backoff.Max", r.Backoff.Max, 11*time.Minute)
	eq("backoff.MaxAttempts", r.Backoff.MaxAttempts, 5)
	eq("poll.Min", r.Poll.Min, 12*time.Second)
	eq("poll.Max", r.Poll.Max, 6*time.Minute)
	eq("poll.Safety", r.Poll.Safety, 31*time.Minute)
	eq("budget.ReserveCalls", r.Budget.ReserveCalls, 7)
	eq("budget.ReportResource", r.Budget.ReportResource, "core")
	eq("cursors.MaxDocBytes", r.Cursors.MaxDocBytes, 9000)
	eq("cursors.MaxChunks", r.Cursors.MaxChunks, 13)
	eq("inbox.MaxDocBytes", r.Inbox.MaxDocBytes, 8000)
	eq("inbox.SeenLimit", r.Inbox.SeenLimit, 4)
	eq("inbox.MaxIDBytes", r.Inbox.MaxIDBytes, 32)
	eq("service.ReNotifyAfter", r.Service.ReNotifyAfter, 90*time.Minute)
	eq("webhook.MaxBody", r.Webhook.MaxBody, int64(4321))
	eq("webhook.Provider", r.Webhook.Provider, "github")
	eq("webhook routes", len(r.Webhook.Routes), 2)
	eq("route 0", r.Webhook.Routes[0].Repo, "o/a")
	eq("route 1", r.Webhook.Routes[1].Repo, "o/b")
	for repo, q := range r.queues {
		eq("webhook queue MaxDocBytes of "+repo, q.MaxDocBytes, 7000)
		eq("webhook queue SeenLimit of "+repo, q.SeenLimit, 5)
		eq("webhook queue MaxIDBytes of "+repo, q.MaxIDBytes, 40)
	}
	eq("webhook queues", len(r.queues), 2)

	w := r.NewWait("sess-a", "re-run", "CI red", t0.Add(time.Hour))
	eq("wait.MaxRetries", w.MaxRetries, 4)
	eq("wait.Escalate", w.Escalate, "leader")
	eq("wait.Owner", w.Owner, "sess-a")
	if err := w.Validate(); err != nil {
		t.Errorf("a wait built from the settings is invalid: %v", err)
	}

	gh := r.IntakeGH(func(context.Context, ...string) ([]byte, error) { return nil, nil })
	eq("intake.ListLimit", gh.ListLimit, 211)
	eq("intake.AllowWrite", gh.AllowWrite, false)
	eq("intake.Repo", gh.Repo, "o/a")
	eq("intake duplicate limit", r.IntakeDuplicateLimit(), 6)
	eq("PRReader.PRListLimit", r.PRReader(func(context.Context, ...string) ([]byte, error) { return nil, nil }).PRListLimit, 307)
	eq("identity.MaxAge", r.Identity(nil).MaxAge, 17*time.Minute)
	eq("labels.PerPage", r.Labels(nil).PerPage, 53)
	eq("labels.Repo", r.Labels(nil).Repo, "o/a")
	eq("ledger.MaxBytes", r.Ledger(nil, nil).MaxBytes, 2048)
}

func TestBuildNeedsWhatTheSettingsCannotProvide(t *testing.T) {
	c := &clock{now: t0}
	for name, mut := range map[string]func(*Deps){
		"store":  func(d *Deps) { d.Store = nil },
		"gh":     func(d *Deps) { d.GHRead = nil },
		"clock":  func(d *Deps) { d.Now = nil },
		"secret": func(d *Deps) { d.Secret = nil },
	} {
		d := deps(t, c)
		mut(&d)
		if _, err := BuildFromProject(project(t, goodYAML), d); err == nil {
			t.Errorf("built without %s", name)
		}
	}
}

func TestAProjectThatIsNotConfiguredBuildsNothing(t *testing.T) {
	c := &clock{now: t0}
	if r, err := BuildFromProject(t.TempDir(), deps(t, c)); !errors.Is(err, ErrNotConfigured) || r != nil {
		t.Errorf("no file: %v %v", r, err)
	}
	off := strings.Replace(goodYAML, "enabled: true", "enabled: false", 1)
	if r, err := BuildFromProject(project(t, off), deps(t, c)); !errors.Is(err, ErrNotConfigured) || r != nil {
		t.Errorf("enabled false: %v %v", r, err)
	}
	broken := strings.Replace(goodYAML, "pr_list_limit: 307", "pr_list_limit: 0", 1)
	if r, err := BuildFromProject(project(t, broken), deps(t, c)); err == nil || errors.Is(err, ErrNotConfigured) || r != nil {
		t.Errorf("a rejected value must be a validation error, not 'not configured': %v", err)
	}
	// the secret is injected: the settings file is refused when it carries one
	leaky := strings.Replace(goodYAML, "- repository: o/a", "- repository: o/a\n          secret: hunter2", 1)
	if _, err := BuildFromProject(project(t, leaky), deps(t, c)); err == nil || strings.Contains(err.Error(), "hunter2") {
		t.Errorf("a secret value in the file: %v", err)
	}
}

// The loop: one pass at start, then one per tick. A wait that nobody answers is reminded at each
// recheck and escalated at the limit, through the inbox; cancelling the context ends the loop.
func TestTheLoopRemindsAndEscalatesThroughTheInboxAndStopsWithItsContext(t *testing.T) {
	r, c := build(t)
	ref, _ := core.NewIssueRef(7)
	run, _ := core.ParseRunID("run-1")
	rec, err := r.deps.Store.Acquire(ref, "sess-a", "cmd-1", run)
	if err != nil {
		t.Fatal(err)
	}
	tk := runner.Ticket{Issue: ref, Holder: "sess-a", Epoch: rec.Epoch}
	if _, err := runner.New(r.deps.Store).SetWait(tk, r.NewWait("sess-a", "re-run the check", "CI red", t0.Add(time.Hour))); err != nil {
		t.Fatal(err)
	}

	ticks := make(chan time.Time)
	passes := make(chan Pass)
	ctx, cancel := context.WithCancel(context.Background())
	var runErr error
	finished := make(chan struct{})
	go func() {
		runErr = r.Run(ctx, ticks, func(p Pass) {
			select {
			case passes <- p:
			case <-ctx.Done(): // a failed test must not leave the loop blocked on a report nobody reads
			}
		})
		close(finished)
	}()
	t.Cleanup(func() { cancel(); <-finished })

	next := func() Pass {
		t.Helper()
		select {
		case p := <-passes:
			return p
		case <-time.After(10 * time.Second):
			t.Fatal("the loop did not report a pass")
			return Pass{}
		}
	}
	if p := next(); p.Err != nil || len(p.Actions) != 0 { // the pass at start: nothing is due
		t.Fatalf("first pass: %+v", p)
	}
	// each recheck is ReNotifyAfter (90m) after the previous one; MaxRetries is 4
	for i, at := range []time.Duration{time.Hour, 150 * time.Minute, 240 * time.Minute, 330 * time.Minute, 420 * time.Minute} {
		c.now = t0.Add(at)
		ticks <- c.now
		if p := next(); p.Err != nil {
			t.Fatalf("pass %d: %v", i, p.Err)
		}
	}
	got, err := r.Inbox.Notices("sess-a")
	if err != nil || len(got) != 4 {
		t.Fatalf("holder inbox: %d notices, err=%v", len(got), err)
	}
	esc, err := r.Inbox.Escalations("leader")
	if err != nil || len(esc) != 1 || esc[0].Item.Issue != 7 || esc[0].Item.MaxRetries != 4 {
		t.Fatalf("leader inbox: %+v %v", esc, err)
	}
	cancel()
	select {
	case <-finished:
		if !errors.Is(runErr, context.Canceled) {
			t.Errorf("Run returned %v", runErr)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the loop did not stop")
	}
}

func TestTheLoopEndsWhenTheTicksEnd(t *testing.T) {
	r, _ := build(t)
	ticks := make(chan time.Time)
	close(ticks)
	n := 0
	if err := r.Run(context.Background(), ticks, func(Pass) { n++ }); err != nil || n != 1 {
		t.Errorf("err=%v passes=%d, want one pass at start and a clean end", err, n)
	}
}

type emptyTransport struct{}

func (emptyTransport) Do(context.Context, forge.Request) (forge.Response, error) {
	return forge.Response{Status: 200, Body: []byte("[]")}, nil
}

// From a signed delivery on the wired handler to the wired collector, under the webhook limits from the
// settings (not the inbox's).
func TestAWebhookDeliveryReachesTheWiredCollectorUnderTheConfiguredLimits(t *testing.T) {
	r, c := build(t)
	srv := httptest.NewServer(r.Webhook)
	t.Cleanup(srv.Close)
	cli := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}
	t.Cleanup(cli.CloseIdleConnections)
	post := func(path, id string, body []byte) int {
		t.Helper()
		req, _ := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(body))
		req.Header.Set("X-Hub-Signature-256", webhook.Sign([]byte(secretValue), body))
		req.Header.Set("X-GitHub-Delivery", id)
		req.Header.Set("X-GitHub-Event", "issues")
		resp, err := cli.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		return resp.StatusCode
	}
	good := []byte(`{"action":"edited","issue":{"number":7},"repository":{"full_name":"o/a"}}`)
	if code := post("/hook/o/a", strings.Repeat("d", 36), good); code != http.StatusAccepted { // a UUID-length id
		t.Fatalf("status = %d", code)
	}
	if code := post("/hook/o/a", strings.Repeat("d", 41), good); code != http.StatusBadRequest { // the webhook queue's bound is 40, not the inbox's 32
		t.Errorf("an id over the webhook queue bound: %d", code)
	}
	if code := post("/hook/o/b", strings.Repeat("e", 36), []byte(`{"action":"edited","issue":{"number":8},"repository":{"full_name":"o/b"}}`)); code != http.StatusAccepted {
		t.Errorf("second route: %d", code)
	}
	if code := post("/hook/o/a", "big", bytes.Repeat([]byte("x"), 4322)); code != http.StatusRequestEntityTooLarge {
		t.Errorf("a body over max_body: %d", code)
	}
	col, err := r.Collector("o/a", "/repos/o/a/issues", emptyTransport{}, forge.DiffByKey("number"))
	if err != nil {
		t.Fatal(err)
	}
	rep := col.Tick(t.Context(), c.now)
	if len(rep.Changes) != 1 || rep.Changes[0].Key != "7" || rep.Changes[0].Kind != forge.ChangeModified {
		t.Fatalf("the collector of o/a: %+v", rep.Changes) // o/b's delivery stays in o/b's queue
	}
	if err := col.Ack(rep); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Collector("o/zzz", "/p", emptyTransport{}, forge.DiffByKey("number")); err == nil {
		t.Error("a collector for a repository with no webhook route was built")
	}
}
