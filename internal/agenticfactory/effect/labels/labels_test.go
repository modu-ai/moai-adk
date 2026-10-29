package labels

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// fakeHost is the only "remote" the tests see. It counts every call and never leaves the process.
type fakeHost struct {
	labels                  map[string]Label // keyed by lower-case name
	lists, creates, updates int
	createErr, updateErr    error
	listErr                 error
	loseNextCreate          bool // the create lands but the response is lost
	ignoreUpdates           bool // update "succeeds" and changes nothing
	onList                  func()
}

func newHost(ls ...Label) *fakeHost {
	h := &fakeHost{labels: map[string]Label{}}
	for _, l := range ls {
		h.labels[strings.ToLower(l.Name)] = l
	}
	return h
}

func (h *fakeHost) List(context.Context) ([]Label, error) {
	h.lists++
	if h.onList != nil {
		h.onList()
	}
	if h.listErr != nil {
		return nil, h.listErr
	}
	var out []Label
	for _, l := range h.labels {
		out = append(out, l)
	}
	return out, nil
}

func (h *fakeHost) Create(_ context.Context, l Label) error {
	h.creates++
	if h.createErr != nil {
		return h.createErr
	}
	h.labels[strings.ToLower(l.Name)] = l
	if h.loseNextCreate {
		h.loseNextCreate = false
		return errors.New("response lost")
	}
	return nil
}

func (h *fakeHost) Update(_ context.Context, name string, l Label) error {
	h.updates++
	if h.updateErr != nil {
		return h.updateErr
	}
	if !h.ignoreUpdates {
		h.labels[strings.ToLower(name)] = l
	}
	return nil
}

func (h *fakeHost) writes() int { return h.creates + h.updates }

var want = Manifest{Labels: []Label{
	{Name: "type:bug", Color: "d73a4a", Description: "Something is broken"},
	{Name: "P0", Color: "b60205", Description: "Top priority"},
	{Name: "agent:active", Color: "0e8a16", Description: ""},
}}

func applier(h Host, allow bool) *Applier {
	return &Applier{
		Host: h, AllowWrite: allow, Repo: "o/r", Store: effect.NewMemIntentStore(),
		Timeouts: effect.Timeouts{LabelApply: time.Second},
	}
}

func approve(t *testing.T, a *Applier, m Manifest) Approval {
	t.Helper()
	p, err := a.Plan(t.Context(), m)
	if err != nil {
		t.Fatal(err)
	}
	return Approval{PlanDigest: p.Digest(), By: "operator"}
}

func TestCheckRejectsBadManifests(t *testing.T) {
	bad := map[string]Manifest{
		"empty name":      {Labels: []Label{{Name: " ", Color: "ffffff"}}},
		"duplicate":       {Labels: []Label{{Name: "Bug", Color: "ffffff"}, {Name: "bug", Color: "000000"}}},
		"bad color":       {Labels: []Label{{Name: "x", Color: "zzzzzz"}}},
		"short color":     {Labels: []Label{{Name: "x", Color: "fff"}}},
		"no labels":       {},
		"newline in name": {Labels: []Label{{Name: "a\nb", Color: "ffffff"}}},
	}
	for name, m := range bad {
		if _, err := Check(m); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	got, err := Check(Manifest{Labels: []Label{{Name: "x", Color: "#D73A4A"}}})
	if err != nil || got.Labels[0].Color != "d73a4a" {
		t.Errorf("color not normalised: %+v err=%v", got, err)
	}
}

func TestPlanIsReadOnlyAndNeverDeletes(t *testing.T) {
	h := newHost(Label{Name: "p0", Color: "000000", Description: "old"}, Label{Name: "stale-extra", Color: "111111"}, Label{Name: "type:bug", Color: "d73a4a", Description: "Something is broken"})
	p, err := applier(h, false).Plan(t.Context(), want)
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[string]OpKind{}
	for _, o := range p.Ops {
		kinds[o.Name] = o.Kind
	}
	if kinds["P0"] != OpUpdate || kinds["agent:active"] != OpCreate || len(p.Ops) != 2 {
		t.Errorf("ops = %+v", p.Ops)
	}
	if len(p.Unmanaged) != 1 || p.Unmanaged[0] != "stale-extra" {
		t.Errorf("unmanaged = %v", p.Unmanaged)
	}
	if h.writes() != 0 || h.lists != 1 {
		t.Errorf("plan wrote or listed wrongly: writes=%d lists=%d", h.writes(), h.lists)
	}
	for _, k := range []OpKind{OpCreate, OpUpdate} {
		if string(k) == "delete" {
			t.Fatal("delete exists")
		}
	}
}

func TestDiffTextIsDeterministic(t *testing.T) {
	h := newHost(Label{Name: "p0", Color: "000000", Description: "old"})
	p1, _ := applier(h, false).Plan(t.Context(), want)
	p2, _ := applier(h, false).Plan(t.Context(), want)
	if p1.Diff() != p2.Diff() || p1.Digest() != p2.Digest() || !strings.Contains(p1.Diff(), "P0") {
		t.Errorf("diff differs:\n%s\n%s", p1.Diff(), p2.Diff())
	}
}

func TestApplyIsOffByDefaultAndMakesNoCall(t *testing.T) {
	h := newHost()
	a := applier(h, false)
	_, err := a.Apply(t.Context(), want, Approval{PlanDigest: "x", By: "operator"})
	if !errors.Is(err, ErrWriteDisabled) || h.lists+h.writes() != 0 {
		t.Errorf("err=%v calls=%d", err, h.lists+h.writes())
	}
}

func TestApplyNeedsAnApprovalForThisExactPlan(t *testing.T) {
	h := newHost()
	a := applier(h, true)
	if _, err := a.Apply(t.Context(), want, Approval{}); !errors.Is(err, ErrNotApproved) || h.lists+h.writes() != 0 {
		t.Errorf("empty approval: err=%v calls=%d", err, h.lists+h.writes())
	}
	if _, err := a.Apply(t.Context(), want, Approval{PlanDigest: "stale", By: "operator"}); !errors.Is(err, ErrPlanChanged) {
		t.Errorf("wrong digest: err=%v", err)
	}
	if h.writes() != 0 {
		t.Errorf("wrote %d times without a matching approval", h.writes())
	}
	if _, err := a.Apply(t.Context(), want, Approval{PlanDigest: "x"}); !errors.Is(err, ErrNotApproved) {
		t.Errorf("approval without an approver accepted: %v", err)
	}
}

func TestRemoteChangedAfterApprovalRefusesToWrite(t *testing.T) {
	h := newHost()
	a := applier(h, true)
	ap := approve(t, a, want)
	h.labels["p0"] = Label{Name: "P0", Color: "123456"} // someone edited the remote after the diff was shown
	if _, err := a.Apply(t.Context(), want, ap); !errors.Is(err, ErrPlanChanged) || h.writes() != 0 {
		t.Errorf("err=%v writes=%d", err, h.writes())
	}
}

func TestApplyWritesExactlyThePlanThenVerifiesByReadback(t *testing.T) {
	h := newHost(Label{Name: "type:bug", Color: "d73a4a", Description: "Something is broken"})
	a := applier(h, true)
	ap := approve(t, a, want)
	res, err := a.Apply(t.Context(), want, ap)
	if err != nil {
		t.Fatal(err)
	}
	if h.creates != 2 || h.updates != 0 || !res.Verified || len(res.Mismatches) != 0 {
		t.Errorf("creates=%d updates=%d res=%+v", h.creates, h.updates, res)
	}
	if res.ReadbackLists < 1 {
		t.Error("success was claimed without reading the remote back")
	}
}

func TestApplyTwiceWritesOnce(t *testing.T) {
	h := newHost()
	a := applier(h, true)
	ap := approve(t, a, want)
	if _, err := a.Apply(t.Context(), want, ap); err != nil {
		t.Fatal(err)
	}
	w := h.writes()
	ap2 := approve(t, a, want) // empty plan now
	res, err := a.Apply(t.Context(), want, ap2)
	if err != nil || h.writes() != w || !res.Verified {
		t.Errorf("second apply wrote: writes %d -> %d res=%+v err=%v", w, h.writes(), res, err)
	}
}

func TestLostCreateResponseIsReconciledNotRepeated(t *testing.T) {
	h := newHost()
	h.loseNextCreate = true
	a := applier(h, true)
	res, err := a.Apply(t.Context(), want, approve(t, a, want))
	if err != nil || !res.Verified {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if h.creates != 3 {
		t.Errorf("creates = %d, want exactly 3 (one per missing label, the lost one not repeated)", h.creates)
	}
}

func TestSilentlyIgnoredUpdateIsCaughtByReadback(t *testing.T) {
	h := newHost(Label{Name: "p0", Color: "000000", Description: "old"})
	h.ignoreUpdates = true
	a := applier(h, true)
	res, _ := a.Apply(t.Context(), want, approve(t, a, want))
	if res.Verified || len(res.Mismatches) == 0 {
		t.Errorf("a no-op update passed verification: %+v", res)
	}
}

func TestDenialStopsAndParksWithoutRetry(t *testing.T) {
	h := newHost()
	h.createErr = &forge.StatusError{Status: 403, Header: map[string]string{"x-ratelimit-remaining": "4000"}}
	a := applier(h, true)
	res, err := a.Apply(t.Context(), want, approve(t, a, want))
	if err == nil || res.Verified || !res.Stopped {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if h.creates != 1 {
		t.Errorf("creates after a denial = %d, want 1 (stop, no retry, no skipping ahead)", h.creates)
	}
	calls := h.creates
	_, _ = a.Apply(t.Context(), want, approve(t, a, want))
	if h.creates != calls {
		t.Errorf("a parked operation was sent again: %d -> %d", calls, h.creates)
	}
}

func TestRateLimitStopsTheRunAndWaitsForNotBefore(t *testing.T) {
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	h := newHost()
	h.createErr = &forge.StatusError{Status: 429, Header: map[string]string{"retry-after": "30"}}
	a := applier(h, true)
	a.Backoff = forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}
	a.Now = func() time.Time { return now }
	res, err := a.Apply(t.Context(), want, approve(t, a, want))
	if err == nil || !res.Stopped || h.creates != 1 {
		t.Fatalf("res=%+v err=%v creates=%d", res, err, h.creates)
	}
	a.Now = func() time.Time { return now.Add(5 * time.Second) }
	_, _ = a.Apply(t.Context(), want, approve(t, a, want))
	if h.creates != 1 {
		t.Errorf("sent before NotBefore: creates=%d", h.creates)
	}
	h.createErr = nil
	a.Now = func() time.Time { return now.Add(31 * time.Second) }
	res, err = a.Apply(t.Context(), want, approve(t, a, want))
	if err != nil || !res.Verified {
		t.Errorf("after NotBefore: res=%+v err=%v", res, err)
	}
}

func TestWritesArePacedBetweenCalls(t *testing.T) {
	h := newHost()
	a := applier(h, true)
	paced := 0
	a.Pace = func(context.Context) { paced++ }
	if _, err := a.Apply(t.Context(), want, approve(t, a, want)); err != nil {
		t.Fatal(err)
	}
	if paced < 2 {
		t.Errorf("paced %d times for 3 writes, want at least 2 gaps", paced)
	}
}
