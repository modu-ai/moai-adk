package mirror

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func ledgerWith(next string) core.Ledger {
	l := ledger()
	l.NextAction = next
	return l
}

func calls(f *fakeGH) []string {
	var out []string
	for _, c := range f.calls {
		out = append(out, strings.Join(c, " "))
	}
	return out
}

// First write lists and creates; later writes go straight to the remembered
// comment, and an unchanged body sends nothing.
func TestRepeatedWritesSkipListAndUnchangedBodies(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":777}`}, {out: "{}"}}}
	m := writable(f)
	c := context.Background()
	if err := m.WriteLedger(c, ledgerWith("a")); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 2 || !strings.Contains(calls(f)[1], "-X POST") {
		t.Fatalf("first write: %v", calls(f))
	}
	if err := m.WriteLedger(c, ledgerWith("b")); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 3 || !strings.Contains(calls(f)[2], "-X PATCH") || !strings.Contains(calls(f)[2], "issues/comments/777") {
		t.Fatalf("second write must PATCH the remembered comment without listing: %v", calls(f))
	}
	if err := m.WriteLedger(c, ledgerWith("b")); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 3 {
		t.Errorf("an unchanged body must send nothing: %v", calls(f))
	}
}

func TestGoneCommentFallsBackToListAndKeepsOneComment(t *testing.T) {
	f := &fakeGH{results: []result{
		{out: ""}, {out: `{"id":777}`}, // first write
		httpFail(404),                  // PATCH of the remembered id
		{out: ""}, {out: `{"id":901}`}, // list finds nothing, then create
		{out: "{}"}, // next write patches the new id
	}}
	m := writable(f)
	c := context.Background()
	_ = m.WriteLedger(c, ledgerWith("a"))
	if err := m.WriteLedger(c, ledgerWith("b")); err != nil {
		t.Fatalf("404 must fall back, got %v", err)
	}
	cs := calls(f)
	if len(cs) != 5 || !strings.Contains(cs[2], "comments/777") || !strings.Contains(cs[3], "issues/1948/comments") || !strings.Contains(cs[4], "-X POST") {
		t.Fatalf("fallback sequence: %v", cs)
	}
	if err := m.WriteLedger(c, ledgerWith("c")); err != nil {
		t.Fatal(err)
	}
	if last := calls(f)[len(f.calls)-1]; !strings.Contains(last, "comments/901") || !strings.Contains(last, "-X PATCH") {
		t.Errorf("the new id must be remembered: %s", last)
	}
}

func TestFailedWriteIsNotRememberedAsWritten(t *testing.T) {
	f := &fakeGH{results: []result{
		{out: ""}, {out: `{"id":777}`}, // first write
		httpFail(429), // changed body, rate limited
		{out: "{}"},   // same changed body again must be sent
	}}
	m := writable(f)
	c := context.Background()
	_ = m.WriteLedger(c, ledgerWith("a"))
	var be *BlockedError
	if err := m.WriteLedger(c, ledgerWith("b")); !errors.As(err, &be) || be.Reason != core.ReasonRateLimited {
		t.Fatalf("want a rate-limit block, got %v", err)
	}
	if err := m.WriteLedger(c, ledgerWith("b")); err != nil {
		t.Fatal(err)
	}
	if n := len(f.calls); n != 4 {
		t.Errorf("the failed body must be sent again, calls=%v", calls(f))
	}
}

func TestInvalidateForcesTheFullPath(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":777}`}, {out: `{"id":777,"body":"` + Marker + `"}`}, {out: "{}"}}}
	m := writable(f)
	c := context.Background()
	_ = m.WriteLedger(c, ledgerWith("a"))
	m.Invalidate(1948)
	if err := m.WriteLedger(c, ledgerWith("a")); err != nil {
		t.Fatal(err)
	}
	if cs := calls(f); len(cs) != 4 || !strings.Contains(cs[2], "issues/1948/comments") || !strings.Contains(cs[3], "-X PATCH") {
		t.Errorf("after Invalidate the list must be read and the body re-sent: %v", cs)
	}
}

func TestReadNeverUsesTheMemo(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":777}`}, {out: `{"id":777,"body":"` + Marker + "\\n```json\\n{\\\"issue\\\":{\\\"Number\\\":1948},\\\"state\\\":\\\"verifying\\\",\\\"store_seq\\\":3}\\n```" + `"}`}}}
	m := writable(f)
	c := context.Background()
	_ = m.WriteLedger(c, ledgerWith("a"))
	before := len(f.calls)
	if _, err := m.ReadLedger(c, 1948); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != before+1 || !strings.Contains(calls(f)[before], "issues/1948/comments") {
		t.Errorf("a read must always ask the remote: %v", calls(f))
	}
}

func TestMemoIsPerIssueAndASensitiveLedgerNeverReachesGh(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":10}`}, {out: ""}, {out: `{"id":20}`}}}
	m := writable(f)
	c := context.Background()
	a := ledgerWith("password=hunter2 and ghp_abcdefghijklmnopqrstuvwxyz0123")
	b := ledgerWith("x")
	b.Issue = core.IssueRef{Number: 2000}
	d := ledgerWith("y")
	if err := m.WriteLedger(c, a); !errors.Is(err, core.ErrUnsanitized) {
		t.Fatalf("a sensitive ledger must be refused: %v", err)
	}
	if len(f.calls) != 0 {
		t.Fatalf("a refused ledger reached gh: %v", calls(f))
	}
	if err := m.WriteLedger(c, b); err != nil {
		t.Fatal(err)
	}
	if err := m.WriteLedger(c, d); err != nil {
		t.Fatal(err)
	}
	all := strings.Join(calls(f), "\n")
	if strings.Contains(all, "hunter2") || strings.Contains(all, "ghp_abcdef") {
		t.Errorf("secret reached a gh call:\n%s", all)
	}
	if !strings.Contains(calls(f)[1], "issues/2000/comments") || !strings.Contains(calls(f)[3], "issues/1948/comments") {
		t.Errorf("each issue needs its own comment: %v", calls(f))
	}
}

// A person deleted the ledger comment. The next changed update must recover in
// the same call (not report success and leave nothing), keeping exactly one comment.
func TestHumanDeletedTheCommentChangedBodyRecoversInTheSameCall(t *testing.T) {
	f := &fakeGH{results: []result{
		{out: ""}, {out: `{"id":777}`}, // first write creates the comment
		httpFail(404),                  // the human deleted it; PATCH finds nothing
		{out: ""}, {out: `{"id":888}`}, // list is empty, so create again
	}}
	m := writable(f)
	c := context.Background()
	if err := m.WriteLedger(c, ledgerWith("before")); err != nil {
		t.Fatal(err)
	}
	if err := m.WriteLedger(c, ledgerWith("after")); err != nil {
		t.Fatalf("the update after a deletion must succeed by recreating: %v", err)
	}
	cs := calls(f)
	posts := 0
	for _, c := range cs {
		if strings.Contains(c, "-X POST") {
			posts++
		}
	}
	if len(cs) != 5 || posts != 2 || !strings.Contains(cs[4], "-X POST") {
		t.Fatalf("recovery sequence: %v", cs)
	}
	if !strings.Contains(cs[4], "after") {
		t.Errorf("the recreated comment must carry the new body: %s", cs[4])
	}
}

// Honest limit: with an unchanged body nothing is sent, so a deleted comment
// stays deleted until the memo is invalidated. The test pins that fact so the
// before/after comparison cannot mistake it for a silent success.
func TestHumanDeletedTheCommentUnchangedBodyIsNotRepairedUntilInvalidated(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":777}`}, {out: ""}, {out: `{"id":888}`}}}
	m := writable(f)
	c := context.Background()
	_ = m.WriteLedger(c, ledgerWith("same"))
	_ = m.WriteLedger(c, ledgerWith("same"))
	if len(f.calls) != 2 {
		t.Fatalf("an unchanged body sends nothing, calls=%v", calls(f))
	}
	m.Invalidate(1948)
	if err := m.WriteLedger(c, ledgerWith("same")); err != nil {
		t.Fatal(err)
	}
	if last := calls(f)[len(f.calls)-1]; !strings.Contains(last, "-X POST") {
		t.Errorf("after Invalidate the missing comment must be recreated: %s", last)
	}
}

func TestEmailAndHomePathAreRefusedAndNeverReachGh(t *testing.T) {
	f := &fakeGH{}
	m := writable(f)
	l := ledgerWith("mail dev.person@example.co.kr, log /Users/goos/MoAI/x.log")
	err := m.WriteLedger(context.Background(), l)
	if !errors.Is(err, core.ErrUnsanitized) {
		t.Fatalf("want ErrUnsanitized, got %v", err)
	}
	if len(f.calls) != 0 {
		t.Fatalf("a refused ledger reached gh: %v", calls(f))
	}
	for _, leak := range []string{"example.co.kr", "dev.person", "/Users/goos"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("the refusal repeats %q: %v", leak, err)
		}
	}
	for _, name := range []string{"email", "home-path"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("the refusal must name %q: %v", name, err)
		}
	}
}
