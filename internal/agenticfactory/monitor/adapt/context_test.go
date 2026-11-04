package adapt

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// AC-AF-018 (design: the context ledger and where evidence lives). The cases (a)..(g) of the acceptance
// criterion, read through the monitor's context items, which is where a reader sees the ledger.

const (
	cSHA1 = "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
	cSHA2 = "b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2"
	cHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	base  = "https://github.com/o/r/blob"
	cPath = ".moai/specs/SPEC-X/spec.md"
)

func ledgerItem() ctxmodel.Item {
	return ctxmodel.Item{Section: ctxmodel.SectionSpec, Title: "SPEC", Summary: "spec", Path: cPath,
		Commit: cSHA1, TargetHead: cSHA1, SHA256: cHash, Owner: "manager-spec", Revision: "r1"}
}

// ledgerAt returns a one-item ledger advanced to the given state (RemoteVerified goes through a lookup).
func ledgerAt(t *testing.T, to ctxmodel.LinkState) *ctxmodel.Ledger {
	t.Helper()
	l := &ctxmodel.Ledger{Issue: 7}
	if err := l.Add(ledgerItem()); err != nil {
		t.Fatal(err)
	}
	for _, s := range []ctxmodel.LinkState{ctxmodel.StateMasked, ctxmodel.StateCommitted, ctxmodel.StatePushed} {
		if l.Items[0].State == to {
			return l
		}
		if err := l.Advance(0, s); err != nil {
			t.Fatal(err)
		}
	}
	if to == ctxmodel.StateRemoteVerified {
		if err := l.VerifyRemote(0, okVerifier{}); err != nil {
			t.Fatal(err)
		}
	}
	return l
}

type okVerifier struct{}

func (okVerifier) Lookup(string, string) (string, bool, error) { return cHash, true, nil }

type failVerifier struct{}

func (failVerifier) Lookup(string, string) (string, bool, error) {
	return "", false, errors.New("gh 502")
}

// collected runs the items through the monitor's collector, which derives availability.
type ctxReader struct{ items []monitor.ContextItem }

func (c ctxReader) ReadContext(context.Context, int) ([]monitor.ContextItem, error) {
	return c.items, nil
}

type nopClaims struct{}

func (nopClaims) ReadClaim(context.Context, int) (monitor.ClaimRow, error) {
	return monitor.ClaimRow{Issue: 7}, nil
}

type nopRemote struct{}

func (nopRemote) RequiredChecks(context.Context, string) ([]monitor.CheckRun, error) { return nil, nil }

type nopLedger struct{}

func (nopLedger) NextAction(context.Context, int) (string, error) { return "", nil }

func collect(items []monitor.ContextItem) []monitor.ContextItem {
	c := &monitor.Collector{Claims: nopClaims{}, Remote: nopRemote{}, Ledger: nopLedger{}, Context: ctxReader{items},
		Now: func() time.Time { return time.Unix(1_700_000_000, 0) }}
	return c.Collect(context.Background(), []int{7})[0].Context.Value
}

func TestEvidenceStage_LocalOnlyEvidence_IsInaccessibleLocal_AndPostsNoLink(t *testing.T) {
	for _, s := range []ctxmodel.LinkState{ctxmodel.StatePrepared, ctxmodel.StateMasked} {
		items := ContextItems(ledgerAt(t, s), base)
		if items[0].State != monitor.CtxInaccessibleLocal || items[0].Link != "" {
			t.Errorf("%s: %+v, want inaccessible-local with no link", s, items[0])
		}
		if got := collect(items)[0].Availability; got != monitor.LinkInaccessible {
			t.Errorf("%s: availability = %s, want inaccessible", s, got)
		}
	}
}

func TestEvidenceStage_LocalCommitOnly_IsInaccessibleUnpushed_NeverAccessibleOffline(t *testing.T) {
	items := ContextItems(ledgerAt(t, ctxmodel.StateCommitted), base)
	if items[0].State != monitor.CtxInaccessibleUnpushed || items[0].Link != "" {
		t.Fatalf("%+v, want inaccessible-unpushed with no link", items[0])
	}
	if got := collect(items)[0].Availability; got != monitor.LinkInaccessible {
		t.Errorf("availability = %s, want inaccessible: a local commit is not reachable", got)
	}
}

func TestLedger_HeadChange_IsStale_CountUnchanged_VerdictPreserved(t *testing.T) {
	l := ledgerAt(t, ctxmodel.StateRemoteVerified)
	before := ContextItems(l, base)[0]
	if n := l.Refresh(map[string]ctxmodel.Current{cPath: {Head: cSHA2, SHA256: cHash}}); n != 1 {
		t.Fatalf("stale count = %d", n)
	}
	items := ContextItems(l, base)
	if len(items) != 1 {
		t.Fatalf("items = %d, want the stale item kept", len(items))
	}
	if items[0].State != monitor.CtxStale || items[0].Stage != before.Stage || items[0].Hash != before.Hash {
		t.Errorf("%+v, want stale with the stage and hash it was judged at (%s / %s)", items[0], before.Stage, before.Hash)
	}
	if collect(items)[0].Availability == monitor.LinkAvailable {
		t.Error("a stale item was offered as available")
	}
}

func TestLedger_BlobChange_IsStale(t *testing.T) {
	l := ledgerAt(t, ctxmodel.StateRemoteVerified)
	other := strings.Repeat("f", 64)
	if n := l.Refresh(map[string]ctxmodel.Current{cPath: {Head: cSHA1, SHA256: other}}); n != 1 {
		t.Fatalf("stale count = %d", n)
	}
	if got := ContextItems(l, base)[0].State; got != monitor.CtxStale {
		t.Errorf("state = %s, want stale", got)
	}
}

// (e) A failing remote lookup does not make anything available. (The ledger keeps no memory of the failed
// lookup, so the item stays "pushed"; a distinct UNKNOWN label is not produced - recorded as a gap.)
func TestEvidenceStage_LookupFailure_IsNeverAvailable(t *testing.T) {
	l := ledgerAt(t, ctxmodel.StatePushed)
	if err := l.VerifyRemote(0, failVerifier{}); err == nil {
		t.Fatal("a failing lookup reported success")
	}
	items := ContextItems(l, base)
	if items[0].RemoteVerified {
		t.Error("RemoteVerified set after a failed lookup")
	}
	if got := collect(items)[0].Availability; got == monitor.LinkAvailable {
		t.Errorf("availability = %s after a failed lookup", got)
	}
}

func TestEvidenceStage_Pushed_IsNotAvailableBeforeTheRemoteLookup(t *testing.T) {
	items := ContextItems(ledgerAt(t, ctxmodel.StatePushed), base)
	if got := collect(items)[0].Availability; got != monitor.LinkUnverified {
		t.Errorf("pushed: availability = %s, want unverified until a lookup confirms it", got)
	}
	verified := collect(ContextItems(ledgerAt(t, ctxmodel.StateRemoteVerified), base))[0]
	if verified.Availability != monitor.LinkAvailable {
		t.Errorf("remote-verified: availability = %s, want available", verified.Availability)
	}
}

// (g) A link is built from a full commit SHA only; a branch name is refused at the ledger, and the adapter
// never composes a link from anything else.
func TestLedger_BranchNameLink_IsRejected_AndLinksUseTheFullSHA(t *testing.T) {
	for _, commit := range []string{"main", "origin/main", "a1a1a1a"} {
		l := &ctxmodel.Ledger{Issue: 7}
		it := ledgerItem()
		it.Commit = commit
		if err := l.Add(it); !errors.Is(err, ctxmodel.ErrLinkBlocked) {
			t.Errorf("commit %q: %v, want ErrLinkBlocked", commit, err)
		}
	}
	link := ContextItems(ledgerAt(t, ctxmodel.StateRemoteVerified), base+"/")[0].Link
	if link != base+"/"+cSHA1+"/"+cPath {
		t.Errorf("link = %q", link)
	}
}
