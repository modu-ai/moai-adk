package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// fakeGH records every invocation and replays scripted results in order.
type fakeGH struct {
	calls   [][]string
	results []result
}

type result struct {
	out, errOut string
	exit        int
	err         error
}

func (f *fakeGH) Run(_ context.Context, args ...string) ([]byte, []byte, int, error) {
	f.calls = append(f.calls, args)
	if len(f.results) == 0 {
		return nil, nil, 0, nil
	}
	r := f.results[0]
	f.results = f.results[1:]
	return []byte(r.out), []byte(r.errOut), r.exit, r.err
}

// httpFail is what `gh api -i` leaves on stdout for a failed call: the status
// line, the headers, a blank line and the body. kv pairs are header names and
// values; the key "body" sets the body instead. stderr carries wording that the
// code must not decide on.
func httpFail(status int, kv ...string) result {
	var hdr, body string
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i] == "body" {
			body = kv[i+1]
			continue
		}
		hdr += kv[i] + ": " + kv[i+1] + "\r\n"
	}
	return result{
		out:    fmt.Sprintf("HTTP/2.0 %d status\r\n%s\r\n%s", status, hdr, body),
		errOut: fmt.Sprintf("gh: request failed (HTTP %d)", status),
		exit:   1,
	}
}

func ledger() core.Ledger {
	ref, _ := core.NewIssueRef(1948)
	return core.Ledger{Issue: ref, State: core.StateVerifying, Owner: "sess-a", StoreSeq: 3, NextAction: "wait for audit"}
}

func writable(f *fakeGH) *Mirror { return &Mirror{Repo: "o/r", Exec: f, AllowWrite: true} }

func TestWriteIsDisabledByDefault(t *testing.T) {
	f := &fakeGH{}
	m := &Mirror{Repo: "o/r", Exec: f}
	if err := m.WriteLedger(context.Background(), ledger()); !errors.Is(err, ErrWriteDisabled) {
		t.Errorf("want ErrWriteDisabled, got %v", err)
	}
	if len(f.calls) != 0 {
		t.Errorf("gh must not be called when writes are disabled: %v", f.calls)
	}
}

func TestWriteCreatesLedgerCommentWhenAbsent(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	if err := writable(f).WriteLedger(context.Background(), ledger()); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("want list + create, got %v", f.calls)
	}
	post := strings.Join(f.calls[1], " ")
	if !strings.Contains(post, "-X POST") || !strings.Contains(post, "repos/o/r/issues/1948/comments") || !strings.Contains(post, Marker) {
		t.Errorf("create call = %s", post)
	}
}

func TestWriteUpdatesExistingLedgerComment(t *testing.T) {
	existing := `{"id":777,"body":"` + Marker + `\n` + "```json\\n{}\\n```" + `"}` + "\n" + `{"id":5,"body":"human note"}`
	f := &fakeGH{results: []result{{out: existing}, {out: "{}"}}}
	if err := writable(f).WriteLedger(context.Background(), ledger()); err != nil {
		t.Fatal(err)
	}
	patch := strings.Join(f.calls[1], " ")
	if !strings.Contains(patch, "-X PATCH") || !strings.Contains(patch, "repos/o/r/issues/comments/777") {
		t.Errorf("update call = %s", patch)
	}
}

func TestOnlyLedgerEndpointsAreTouched(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	_ = writable(f).WriteLedger(context.Background(), ledger())
	for _, c := range f.calls {
		j := strings.Join(c, " ")
		if c[0] != "api" || strings.Contains(j, "close") || strings.Contains(j, "DELETE") || strings.Contains(j, "/pulls") {
			t.Errorf("unexpected gh call: %s", j)
		}
	}
}

func TestReadRoundTrip(t *testing.T) {
	var written string
	f := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	m := writable(f)
	if err := m.WriteLedger(context.Background(), ledger()); err != nil {
		t.Fatal(err)
	}
	for i, a := range f.calls[1] {
		if a == "-f" || a == "--raw-field" {
			written = strings.TrimPrefix(f.calls[1][i+1], "body=")
		}
	}
	line, _ := json.Marshal(map[string]any{"id": 9, "body": written})
	m2 := &Mirror{Repo: "o/r", Exec: &fakeGH{results: []result{{out: string(line)}}}}
	got, err := m2.ReadLedger(context.Background(), 1948)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateVerifying || got.Owner != "sess-a" || got.StoreSeq != 3 {
		t.Errorf("round trip: %+v", got)
	}
}

func TestReadWithoutLedgerComment(t *testing.T) {
	m := &Mirror{Repo: "o/r", Exec: &fakeGH{results: []result{{out: `{"id":1,"body":"hello"}`}}}}
	if _, err := m.ReadLedger(context.Background(), 1948); !errors.Is(err, ErrLedgerAbsent) {
		t.Errorf("want ErrLedgerAbsent, got %v", err)
	}
}

func TestSecretsAreRedactedAndCounted(t *testing.T) {
	l := ledger()
	l.NextAction = "use ghp_abcdefghijklmnopqrstuvwxyz0123456789 and password=hunter2 then continue"
	f := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	if err := writable(f).WriteLedger(context.Background(), l); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(f.calls[1], " ")
	if strings.Contains(body, "ghp_abcdef") || strings.Contains(body, "hunter2") {
		t.Errorf("secret leaked into body: %s", body)
	}
	if !strings.Contains(body, "redacted") {
		t.Errorf("redaction must be disclosed: %s", body)
	}
}

func TestFailuresBecomeBlocksWithoutRetry(t *testing.T) {
	cases := []struct {
		name string
		res  result
		want core.BlockReason
	}{
		{"401", httpFail(401), core.ReasonAuthFailed},
		{"403 resource not accessible", httpFail(403, "body", `{"message":"Resource not accessible by integration"}`), core.ReasonPermissionDenied},
		{"403 with no signal", httpFail(403), core.ReasonUnclassified},
		{"429", httpFail(429), core.ReasonRateLimited},
		{"not installed", result{err: exec.ErrNotFound}, core.ReasonNotInstalled},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeGH{results: []result{c.res}}
			err := writable(f).WriteLedger(context.Background(), ledger())
			var be *BlockedError
			if !errors.As(err, &be) || be.Reason != c.want {
				t.Fatalf("want block %q, got %v", c.want, err)
			}
			if len(f.calls) != 1 {
				t.Errorf("must not retry: %d calls", len(f.calls))
			}
		})
	}
}

func TestUnclassifiedFailureIsNotABlockReason(t *testing.T) {
	f := &fakeGH{results: []result{{errOut: "boom", exit: 1}}}
	err := writable(f).WriteLedger(context.Background(), ledger())
	var be *BlockedError
	if err == nil || errors.As(err, &be) {
		t.Errorf("unknown failure must be a plain error, got %v", err)
	}
}

func TestPublishRecordsMirrorStateAndNeverChangesIssueState(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 100)
	ref, _ := core.NewIssueRef(1948)
	rec, err := st.Acquire(ref, "sess-a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	// 429 -> mirror_state blocked(rate-limited), execution state untouched.
	f := &fakeGH{results: []result{httpFail(429)}}
	out, err := Publish(context.Background(), writable(f), st, ref, "sess-a", rec.Epoch)
	var be *BlockedError
	if !errors.As(err, &be) {
		t.Fatalf("want BlockedError, got %v", err)
	}
	if out.Mirror != "blocked(rate-limited)" || out.State != core.StateClaimed {
		t.Errorf("record after 429: %+v", out)
	}
	// Success -> current.
	f2 := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	out, err = Publish(context.Background(), writable(f2), st, ref, "sess-a", rec.Epoch)
	if err != nil || out.Mirror != "current" {
		t.Errorf("record after success: %+v %v", out, err)
	}
}

func TestPublishRefusedWithoutOwnership(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 100)
	ref, _ := core.NewIssueRef(1948)
	rec, _ := st.Acquire(ref, "sess-a", "d1", "run-1")
	f := &fakeGH{}
	_, err := Publish(context.Background(), writable(f), st, ref, "intruder", rec.Epoch)
	if !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("want ownership refusal, got %v", err)
	}
	if len(f.calls) != 0 {
		t.Errorf("gh called despite refused ownership: %v", f.calls)
	}
}

func TestBlockedErrorText(t *testing.T) {
	e := &BlockedError{Reason: core.ReasonRateLimited, Detail: "HTTP 429"}
	if !strings.Contains(e.Error(), "rate-limited") {
		t.Errorf("text = %q", e.Error())
	}
}

// GH is the real runner; these cases use only local no-op binaries and never
// reach GitHub.
func TestGHRunnerExitAndMissingBinary(t *testing.T) {
	ctx := context.Background()
	if _, _, _, err := (GH{Bin: "moai-no-such-gh-binary"}).Run(ctx, "x"); !errors.Is(err, exec.ErrNotFound) {
		t.Errorf("missing binary: %v", err)
	}
	for bin, want := range map[string]int{"true": 0, "false": 1} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not available", bin)
		}
		if _, _, exit, err := (GH{Bin: bin}).Run(ctx); err != nil || exit != want {
			t.Errorf("%s: exit=%d err=%v", bin, exit, err)
		}
	}
}

func TestReadLedgerMalformedComments(t *testing.T) {
	cases := map[string]string{
		"no json block": `{"id":1,"body":"` + Marker + `"}`,
		"bad json":      `{"id":1,"body":"` + Marker + "\\n```json\\n{nope\\n```" + `"}`,
	}
	for name, line := range cases {
		m := &Mirror{Repo: "o/r", Exec: &fakeGH{results: []result{{out: line}}}}
		if _, err := m.ReadLedger(context.Background(), 1948); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	m := &Mirror{Repo: "o/r", Exec: &fakeGH{results: []result{httpFail(401)}}}
	var be *BlockedError
	if _, err := m.ReadLedger(context.Background(), 1948); !errors.As(err, &be) {
		t.Errorf("read must classify 401: %v", err)
	}
}

func TestPublishPlainFailureLeavesMirrorStateAlone(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 100)
	ref, _ := core.NewIssueRef(1948)
	rec, _ := st.Acquire(ref, "sess-a", "d1", "run-1")
	f := &fakeGH{results: []result{{errOut: "boom", exit: 1}}}
	if _, err := Publish(context.Background(), writable(f), st, ref, "sess-a", rec.Epoch); err == nil {
		t.Fatal("want error")
	}
	if cur, _ := st.Load(ref); cur.Mirror != "" || cur.Seq != rec.Seq {
		t.Errorf("record changed on unclassified failure: %+v", cur)
	}
}

func TestLedgerFromRecordCarriesTheS46Position(t *testing.T) {
	park := &core.Park{Reason: core.ReasonRateLimited, Evidence: "429", Resume: "after 15:00"}
	gate := &core.Gate{Reason: core.ReasonDebtPresent, Detail: "debt", Resume: "rerun"}
	ref := &core.Refusal{Reason: core.ReasonForeignClosing, Detail: "would close #3000"}
	wait := &core.Wait{Owner: "h", NextAction: "re-run", RecheckAt: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC), Diagnosis: "ci red", MaxRetries: 2, Escalate: "leader"}
	l := LedgerFromRecord(claim.Record{Issue: 7, Holder: "h", State: core.StateVerifying, Stage: core.StageBuilding, Sub: core.SubImplement,
		Park: park, Gate: gate, Refusal: ref, Wait: wait, Seq: 4})
	if l.Wait != wait {
		t.Errorf("the wait record must reach the ledger: %+v", l.Wait)
	}
	if l.Stage != core.StageBuilding || l.Sub != core.SubImplement || l.Park != park || l.Gate != gate || l.Refusal != ref || l.State != core.StateVerifying {
		t.Errorf("projection: %+v", l)
	}
}

func TestLedgerFromRecord(t *testing.T) {
	l := LedgerFromRecord(claim.Record{Issue: 7, Holder: "h", State: core.StateBlocked, Return: core.StateVerifying, Reason: core.ReasonDebtPresent, Seq: 4})
	if l.Issue.Number != 7 || l.Owner != "h" || l.ReturnStage != core.StateVerifying || l.BlockReason != core.ReasonDebtPresent || l.StoreSeq != 4 {
		t.Errorf("projection: %+v", l)
	}
}
