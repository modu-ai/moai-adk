package mirror

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// bodyOf is the comment body a gh call sends.
func bodyOf(t *testing.T, call []string) string {
	t.Helper()
	for _, a := range call {
		if strings.HasPrefix(a, "body=") {
			return strings.TrimPrefix(a, "body=")
		}
	}
	t.Fatalf("no body in %v", call)
	return ""
}

// A ledger that holds a secret or personal data is not posted (design 4.4): what is
// posted in its place is a RESTRICTED record with three fields, status, a short
// summary and a SHA-256, and WriteLedger still reports the refusal. The summary
// names the categories, never the values, and the hash is of the masked form so it
// reveals nothing about the original.
func TestASensitiveLedgerIsPostedAsARestrictedRecordAndReportedRefused(t *testing.T) {
	l := ledger()
	l.NextAction = "use ghp_abcdefghijklmnopqrstuvwxyz0123456789 and password=hunter2 then continue"
	f := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	err := writable(f).WriteLedger(context.Background(), l)
	if !errors.Is(err, core.ErrUnsanitized) {
		t.Fatalf("want ErrUnsanitized, got %v", err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("want list + create of the RESTRICTED record, got %v", f.calls)
	}
	body := bodyOf(t, f.calls[1])
	for _, leak := range []string{"ghp_abcdef", "hunter2", "wait for audit", "sess-a"} {
		if strings.Contains(body, leak) {
			t.Errorf("the RESTRICTED record carries %q: %s", leak, body)
		}
	}
	start := strings.Index(body, "{")
	end := strings.LastIndex(body, "}")
	var rec map[string]string
	if err := json.Unmarshal([]byte(body[start:end+1]), &rec); err != nil {
		t.Fatalf("the record is JSON: %v\n%s", err, body)
	}
	if len(rec) != 3 || rec["status"] != "RESTRICTED" || rec["summary"] == "" || len(rec["sha256"]) != 64 {
		t.Errorf("the record must be exactly status, summary and a SHA-256: %v", rec)
	}
	// the hash is of the masked form: it matches a hash computed here from
	// Sanitize, and differs from the hash of the original ledger text
	rawLedger, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	masked, _ := core.Sanitize(string(rawLedger))
	wantSum := sha256.Sum256([]byte(masked))
	origSum := sha256.Sum256(rawLedger)
	if rec["sha256"] != hex.EncodeToString(wantSum[:]) || rec["sha256"] == hex.EncodeToString(origSum[:]) {
		t.Errorf("the SHA-256 must be of the masked ledger: %s", rec["sha256"])
	}
	// DEC-AF-28: the hash of the original text appears nowhere in what gh was sent
	for _, c := range f.calls {
		if strings.Contains(strings.Join(c, " "), hex.EncodeToString(origSum[:])) {
			t.Errorf("the hash of the original ledger reached gh: %v", c)
		}
	}
	for _, name := range []string{"token", "credential-assignment"} {
		if !strings.Contains(rec["summary"], name) {
			t.Errorf("the summary must name %q: %q", name, rec["summary"])
		}
	}
	// the hash is of the masked form: the same ledger again changes nothing sent
	f2 := &fakeGH{}
	m := writable(&fakeGH{results: []result{{out: ""}, {out: `{"id":5}`}}})
	if err := m.WriteLedger(context.Background(), l); !errors.Is(err, core.ErrUnsanitized) {
		t.Fatal(err)
	}
	m.Exec = f2
	if err := m.WriteLedger(context.Background(), l); !errors.Is(err, core.ErrUnsanitized) || len(f2.calls) != 0 {
		t.Errorf("the same refused ledger twice must send nothing the second time: %v %v", err, f2.calls)
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

// A record whose position could not be derived reaches the ledger marked, not as a
// confirmed "no park".
func TestLedgerFromRecordCarriesAnUnknownPosition(t *testing.T) {
	l := LedgerFromRecord(claim.Record{Issue: 7, Holder: "h", State: core.StateBlocked, PositionUnknown: "blocked needs an in-flight return stage"})
	if l.PositionUnknown == "" || l.Stage != "" || l.Park != nil {
		t.Errorf("the marker must reach the ledger: %+v", l)
	}
	if h := LedgerFromRecord(claim.Record{Issue: 7, Holder: "h", State: core.StateClaimed}); h.PositionUnknown != "" {
		t.Errorf("a known position carries no marker: %+v", h)
	}
}

// Publish records that the ledger was refused, leaves the execution state alone and
// returns the refusal; only the RESTRICTED record reaches gh.
func TestPublishRecordsARefusedLedger(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 100)
	ref, _ := core.NewIssueRef(1948)
	rec, err := st.Acquire(ref, "sess-a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.Update(ref, "sess-a", rec.Epoch, func(r *claim.Record) error { r.Resume = "ask dev@example.com"; return nil }); err != nil {
		t.Fatal(err)
	}
	f := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	out, err := Publish(context.Background(), writable(f), st, ref, "sess-a", rec.Epoch)
	if !errors.Is(err, core.ErrUnsanitized) {
		t.Fatalf("want ErrUnsanitized, got %v", err)
	}
	for _, c := range f.calls {
		if strings.Contains(strings.Join(c, " "), "dev@example.com") {
			t.Errorf("a refused ledger leaked into gh: %v", c)
		}
	}
	if out.Mirror != "refused" || out.State != core.StateClaimed {
		t.Errorf("record after the refusal: mirror=%q state=%s", out.Mirror, out.State)
	}
}
