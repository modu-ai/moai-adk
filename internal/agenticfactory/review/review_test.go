package review

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const (
	headA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	headB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func lookupOf(m map[string]any) Lookup {
	return func(key string) (any, bool) {
		v, ok := m[key]
		return v, ok
	}
}

// counters records every LookPath and Run call a Codex adapter makes.
type counters struct{ lookPath, run int }

func (c *counters) deps() CodexDeps {
	return CodexDeps{
		LookPath: func(string) (string, error) { c.lookPath++; return "/usr/bin/codex", nil },
		Run: func(context.Context, Request) (Result, error) {
			c.run++
			return Result{HeadSHA: headA, Verdict: VerdictPass}, nil
		},
	}
}

// REQ-AF-013: with the opt-in key unset, zero Codex calls in every case.
func TestCodexUnset_ZeroCalls(t *testing.T) {
	cases := map[string]struct {
		lookup Lookup
		env    map[string]string
	}{
		"key absent":         {lookup: lookupOf(nil)},
		"key false":          {lookup: lookupOf(map[string]any{CodexOptInKey: false})},
		"nil config":         {lookup: nil},
		"key not a bool":     {lookup: lookupOf(map[string]any{CodexOptInKey: "true"})},
		"codex on PATH+env":  {lookup: lookupOf(nil), env: map[string]string{"CODEX_API_KEY": "x", "OPENAI_API_KEY": "x", "CODEX_HOME": t.TempDir()}},
		"legacy codex gates": {lookup: lookupOf(map[string]any{"workflow.audit.gates.codex": "required", "workflow.codex.review_gate.enabled": true})},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			var c counters
			_, err := NewCodex(tc.lookup, c.deps()).Review(context.Background(), Request{HeadSHA: headA, Kind: KindCodeReview})
			if !errors.Is(err, ErrNotConfigured) {
				t.Fatalf("err = %v, want ErrNotConfigured", err)
			}
			if c.lookPath != 0 || c.run != 0 {
				t.Fatalf("calls lookPath=%d run=%d, want 0/0", c.lookPath, c.run)
			}
		})
	}
}

// Positive control: the same wiring does call out once the key is true.
func TestCodexOptedIn_CallsOnce(t *testing.T) {
	var c counters
	res, err := NewCodex(lookupOf(map[string]any{CodexOptInKey: true}), c.deps()).
		Review(context.Background(), Request{HeadSHA: headA, Kind: KindCodeReview})
	if err != nil || res.Verdict != VerdictPass {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if c.lookPath != 1 || c.run != 1 {
		t.Fatalf("calls lookPath=%d run=%d, want 1/1", c.lookPath, c.run)
	}
}

// An opted-in project without the binary blocks instead of running.
func TestCodexOptedIn_BinaryAbsent(t *testing.T) {
	ran := 0
	deps := CodexDeps{
		LookPath: func(string) (string, error) { return "", errors.New("not found") },
		Run:      func(context.Context, Request) (Result, error) { ran++; return Result{}, nil },
	}
	_, err := NewCodex(lookupOf(map[string]any{CodexOptInKey: true}), deps).
		Review(context.Background(), Request{HeadSHA: headA})
	if !errors.Is(err, ErrBinaryAbsent) || ran != 0 {
		t.Fatalf("err=%v ran=%d", err, ran)
	}
}

func TestProviderEnabled(t *testing.T) {
	on := lookupOf(map[string]any{CodexOptInKey: true})
	if !ProviderEnabled(ProviderClaude, nil) {
		t.Error("claude independent audit must not depend on any key")
	}
	if ProviderEnabled(ProviderCodex, nil) || !ProviderEnabled(ProviderCodex, on) {
		t.Error("codex enabled only by the explicit key")
	}
	if ProviderEnabled(ProviderGLM, on) || ProviderEnabled("unknown", on) {
		t.Error("glm and unknown providers are never enabled here")
	}
}

func TestClaudeAuditor_UsesExecutor(t *testing.T) {
	fake := &FakeExecutor{Result: Result{HeadSHA: headA, Verdict: VerdictPass}}
	res, err := NewClaudeAuditor(fake).Review(context.Background(), Request{HeadSHA: headA, Kind: KindAudit})
	if err != nil || res.Verdict != VerdictPass || len(fake.Calls) != 1 || fake.Calls[0].HeadSHA != headA {
		t.Fatalf("res=%+v err=%v calls=%+v", res, err, fake.Calls)
	}
}

func completed(t *testing.T, v Verdict, head, session string) Record {
	t.Helper()
	r := Record{HeadSHA: head, Kind: KindAudit, Provider: ProviderClaude, Verdict: v,
		Producer: Producer{Role: "auditor", Session: session, Attempt: 1, Epoch: 1}}
	for _, st := range []State{StateSent, StateReceived, StateExecuted, StateCompleted} {
		if err := r.Mark(st, sourceFor[st]); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

// REQ-AF-003: a send result or idle notice never sets a later state.
func TestMark_RefusesForeignEvidence(t *testing.T) {
	var r Record
	for _, st := range []State{StateReceived, StateExecuted, StateCompleted} {
		for _, src := range []Source{SourceSendResult, SourceIdleNotice} {
			if err := r.Mark(st, src); err == nil {
				t.Errorf("Mark(%s, %s) accepted", st, src)
			}
		}
	}
	if err := r.Mark(StateSent, SourceSendResult); err != nil {
		t.Fatal(err)
	}
	if r.Reached(StateReceived) {
		t.Error("sent must not imply received")
	}
}

// REQ-AF-018: only a same-HEAD, debt-free PASS from another session is admitted.
func TestAdmit(t *testing.T) {
	cases := []struct {
		name string
		rec  func() Record
		head string
		want Reason
	}{
		{"pass", func() Record { return completed(t, VerdictPass, headA, "s-audit") }, headA, ReasonNone},
		{"stale head", func() Record { return completed(t, VerdictPass, headA, "s-audit") }, headB, ReasonStale},
		{"short sha", func() Record { return completed(t, VerdictPass, "aaaa", "s-audit") }, "aaaa", ReasonBadSHA},
		{"pass-with-debt", func() Record { return completed(t, VerdictPassWithDebt, headA, "s-audit") }, headA, ReasonDebt},
		{"skipped", func() Record { return completed(t, VerdictSkipped, headA, "s-audit") }, headA, ReasonNotPass},
		{"fail", func() Record { return completed(t, VerdictFail, headA, "s-audit") }, headA, ReasonNotPass},
		{"empty verdict", func() Record { return completed(t, "", headA, "s-audit") }, headA, ReasonNotPass},
		{"writer session", func() Record { return completed(t, VerdictPass, headA, "s-writer") }, headA, ReasonProducer},
		{"not completed", func() Record {
			r := completed(t, VerdictPass, headA, "s-audit")
			r.delivery = map[State]Source{StateSent: SourceSendResult}
			return r
		}, headA, ReasonIncomplete},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rec().Admit(tc.head, "s-writer"); got != tc.want {
				t.Fatalf("Admit = %q, want %q", got, tc.want)
			}
		})
	}
}

// Records are evidence: they must not carry secrets from the request.
func TestRecordString_NoSecretFields(t *testing.T) {
	r := completed(t, VerdictPass, headA, "s-audit")
	if s := r.String(); strings.Contains(strings.ToLower(s), "token") || strings.Contains(strings.ToLower(s), "key") {
		t.Fatalf("record string leaks credential-shaped text: %s", s)
	}
}
