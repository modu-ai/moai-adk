package ctxledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

const (
	mc  = "0123456789abcdef0123456789abcdef01234567"
	mh  = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	mh2 = "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
)

func modelStore(t *testing.T) *claim.Store {
	t.Helper()
	s, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func addItem(t *testing.T, s *claim.Store, it ctxmodel.Item) {
	t.Helper()
	if _, err := ctxmodel.Update(s, 7, func(l *ctxmodel.Ledger) error { return l.Add(it) }); err != nil {
		t.Fatal(err)
	}
}

func specItem(title string) ctxmodel.Item {
	return ctxmodel.Item{Section: ctxmodel.SectionSpec, Title: title, Summary: "the adopted contract", Path: ".moai/specs/SPEC-X/spec.md", Commit: mc, SHA256: mh, Owner: "agent-1"}
}

type verifier struct{ hash string }

func (v verifier) Lookup(string, string) (string, bool, error) { return v.hash, true, nil }

// advance walks item idx to REMOTE_VERIFIED the way the model allows: step by step, the last one by lookup.
func advance(t *testing.T, s *claim.Store, idx int) {
	t.Helper()
	if _, err := ctxmodel.Update(s, 7, func(l *ctxmodel.Ledger) error {
		for _, to := range []ctxmodel.LinkState{ctxmodel.StateMasked, ctxmodel.StateCommitted, ctxmodel.StatePushed} {
			if err := l.Advance(idx, to); err != nil {
				return err
			}
		}
		return l.VerifyRemote(idx, verifier{hash: l.Items[idx].SHA256})
	}); err != nil {
		t.Fatal(err)
	}
}

func TestThePostedBodyIsExactlyWhatTheModelRenders(t *testing.T) {
	l, _, r := counted()
	s := modelStore(t)
	addItem(t, s, specItem("Contract"))
	led, _ := ctxmodel.Load(s, 7)
	want, err := led.Render()
	if err != nil {
		t.Fatal(err)
	}
	res := l.PublishModel(t.Context(), s, 7)
	if res.Status != StepPublished {
		t.Fatalf("res = %+v", res)
	}
	if len(r.comments[7]) != 1 || !strings.HasSuffix(r.comments[7][0].Body, want) {
		t.Errorf("comment = %v, want it to end with the rendered model %q", r.comments[7], want)
	}
	if strings.Contains(r.comments[7][0].Body, ".moai/specs/SPEC-X/spec.md") {
		t.Errorf("an item that is not remote-verified must not be shown as a link: %q", r.comments[7][0].Body)
	}
}

func TestOnlyARemoteVerifiedItemAppearsAsAnAvailableLink(t *testing.T) {
	l, _, r := counted()
	s := modelStore(t)
	addItem(t, s, specItem("Contract"))
	addItem(t, s, ctxmodel.Item{Section: ctxmodel.SectionAudit, Title: "Audit", Summary: "round 1", Path: ".moai/evidence/issue-7/audit.md", Commit: mc, SHA256: mh2, Owner: "agent-2"})
	advance(t, s, 0)
	if res := l.PublishModel(t.Context(), s, 7); res.Status != StepPublished {
		t.Fatalf("res = %+v", res)
	}
	body := r.comments[7][0].Body
	if !strings.Contains(body, "AVAILABLE `.moai/specs/SPEC-X/spec.md`") {
		t.Errorf("the verified item is missing its link: %q", body)
	}
	if strings.Contains(body, ".moai/evidence/issue-7/audit.md") {
		t.Errorf("the unverified item leaked its path: %q", body)
	}
}

func TestAnUnchangedModelIsANoOpAndAChangedOneUpdatesTheSameComment(t *testing.T) {
	l, c, r := counted()
	s := modelStore(t)
	addItem(t, s, specItem("Contract"))
	l.PublishModel(t.Context(), s, 7)
	before := c.Snapshot()
	l.PublishModel(t.Context(), s, 7)
	if after := c.Snapshot(); after.Create != before.Create || after.Update != before.Update {
		t.Errorf("an unchanged model wrote: %+v -> %+v", before, after)
	}
	advance(t, s, 0)
	if res := l.PublishModel(t.Context(), s, 7); res.Status != StepPublished {
		t.Fatalf("res = %+v", res)
	}
	if len(r.comments[7]) != 1 || !strings.Contains(r.comments[7][0].Body, "AVAILABLE") {
		t.Errorf("comments = %+v", r.comments[7])
	}
}

// Render refuses instead of masking. The refusal must reach the caller as a block and nothing may be sent.
func TestAModelThatRenderRefusesIsBlockedAndNothingIsSent(t *testing.T) {
	l, c, r := counted()
	s := modelStore(t)
	it := specItem("Contract")
	it.Summary = "token ghp_abcdefghijklmnopqrstuvwxyz0123456789 leaked into a summary"
	addItem(t, s, it)
	res := l.PublishModel(t.Context(), s, 7)
	if res.Status != StepBlocked || !errors.Is(res.Err, ctxmodel.ErrUnmasked) {
		t.Fatalf("res = %+v", res)
	}
	if c.Snapshot().Total() != 0 || len(r.comments[7]) != 0 {
		t.Error("a body that Render refused was sent")
	}
	raw := specItem("Raw log")
	raw.Summary = strings.Repeat("line of a raw log\n", 40)
	s2 := modelStore(t)
	addItem(t, s2, raw)
	if res := l.PublishModel(t.Context(), s2, 7); res.Status != StepBlocked {
		t.Errorf("a raw-log-shaped summary: %+v", res)
	}
}

// E-mail addresses and home paths never reach the comment: Render refuses them (core's redaction covers
// them), and this package's sender check is a second layer behind it. Whichever layer stops them, the step
// is blocked and nothing is sent.
func TestPersonalDataIsBlockedWhicheverLayerCatchesIt(t *testing.T) {
	email, home := "person"+"@"+"example.org", "/Us"+"ers/someone/project"
	cases := map[string]ctxmodel.Item{
		"email in a summary":                   func() ctxmodel.Item { it := specItem("Contract"); it.Summary = "owner " + email; return it }(),
		"home path in summary":                 func() ctxmodel.Item { it := specItem("Contract"); it.Summary = "worktree " + home; return it }(),
		"email as the owner of a pending item": {Section: ctxmodel.SectionPlan, Title: "Plan", Owner: email, Pending: true},
	}
	for name, it := range cases {
		l, c, r := counted()
		s := modelStore(t)
		addItem(t, s, it)
		res := l.PublishModel(t.Context(), s, 7)
		if res.Status != StepBlocked || res.Err == nil {
			t.Errorf("%s: res = %+v", name, res)
		}
		if c.Snapshot().Total() != 0 || len(r.comments[7]) != 0 {
			t.Errorf("%s: personal data was sent", name)
		}
	}
}

func TestAnEmptyModelPublishesNothing(t *testing.T) {
	l, c, _ := counted()
	if res := l.PublishModel(t.Context(), modelStore(t), 7); res.Status != StepSkipped || c.Snapshot().Total() != 0 {
		t.Errorf("res = %+v calls = %+v", res, c.Snapshot())
	}
}

func TestAHandEditedCommentIsRestoredFromTheModel(t *testing.T) {
	l, _, r := counted()
	s := modelStore(t)
	addItem(t, s, specItem("Contract"))
	l.PublishModel(t.Context(), s, 7)
	r.comments[7][0].Body += "\nsomeone's note"
	l.PublishModel(t.Context(), s, 7)
	if strings.Contains(r.comments[7][0].Body, "someone's note") || len(r.comments[7]) != 1 {
		t.Errorf("comment = %+v", r.comments[7])
	}
}
