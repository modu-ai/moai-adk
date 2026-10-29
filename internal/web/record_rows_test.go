package web

// record_rows_test.go — the PR tracking rows filled from the records the claim
// store already persists (Issue #1948, backlog B1). No network, no store
// creation, read-only.
//
// Defect answer: these tests catch a row invented for a record the store does
// not hold, a PR phase or check result shown that no persisted record carries
// (the live checks and review verdict are remote facts), the open-PR count
// diverging from the store's own rule (handed off and not done), a record that
// has no receipt shown as having a PR, a secret in a persisted text reaching
// the page, and a source that writes or needs anything but List/MaxOpenPRs.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func writeRecord(t *testing.T, dir string, r claim.Record) {
	t.Helper()
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "issue-"+itoa(r.Issue)+".json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func recordStore(t *testing.T, limit int) (*claim.Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := claim.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if limit > 0 {
		if err := s.SetMaxOpenPRs(limit); err != nil {
			t.Fatal(err)
		}
	}
	return s, dir
}

func seedRecords(t *testing.T, dir string) {
	t.Helper()
	sha := core.HeadSHA(strings.Repeat("a", 40))
	secret := "ghp_" + strings.Repeat("b", 36)
	// 71: handed off, waiting — has a receipt, a wait record, a gate.
	writeRecord(t, dir, claim.Record{
		Issue: 71, Holder: "11111111-2222-3333-4444-555555555555", AgentName: "agent-3", State: core.StateLanding, Head: sha,
		Receipt: &claim.Receipt{HandoffID: "h1", LeaderAccepted: true, PR: 501, Head: sha},
		Wait: &core.Wait{Owner: "11111111-2222-3333-4444-555555555555", NextAction: "re-run checks " + secret, RecheckAt: time.Date(2026, 10, 30, 8, 0, 0, 0, time.UTC),
			Diagnosis: "ci queue", Retries: 1, MaxRetries: 3, Escalate: "leader"},
	})
	// 72: handed off, merged, not done.
	writeRecord(t, dir, claim.Record{Issue: 72, Holder: "s-72", AgentName: "agent-4", State: core.StateLanding, Merged: true,
		Receipt: &claim.Receipt{HandoffID: "h2", LeaderAccepted: true, PR: 502, Head: sha}})
	// 73: done and closed.
	writeRecord(t, dir, claim.Record{Issue: 73, AgentName: "agent-5", State: core.StateDone, Closed: true, Merged: true,
		Receipt: &claim.Receipt{HandoffID: "h3", LeaderAccepted: true, PR: 503, Head: sha}})
	// 74: working, no PR yet.
	writeRecord(t, dir, claim.Record{Issue: 74, Holder: "s-74", AgentName: "agent-6", State: core.StateImplementing})
	// 75: PR handed off, the owner is repairing it (E06 evidence persisted).
	writeRecord(t, dir, claim.Record{Issue: 75, Holder: "s-75", AgentName: "agent-7", State: core.StateLanding,
		Repair:  &claim.Repair{Decision: "fix ci", Owner: "s-75", ScopeReserved: true, OldWriterEnded: true},
		Receipt: &claim.Receipt{HandoffID: "h5", LeaderAccepted: true, PR: 505, Head: sha}})
}

func recordApp(t *testing.T, src RecordSource) *app {
	t.Helper()
	a := newApp(Config{ProjectRoot: t.TempDir(), MonitorRows: RecordRowsFrom(src)})
	return a
}

func TestRecordRows_FromPersistedRecords(t *testing.T) {
	s, dir := recordStore(t, 4)
	seedRecords(t, dir)
	vm := recordApp(t, s).buildPRTracking(context.Background())
	if !vm.Available || len(vm.Rows) != 5 {
		t.Fatalf("vm = %+v", vm)
	}
	by := map[int]PRTrackRowVM{}
	for _, r := range vm.Rows {
		by[r.Issue] = r
	}
	r := by[71]
	if r.PR != "#501" || r.Agent != "agent-3" || r.SessionID != "11111111-2222-3333-4444-555555555555" || r.Head != "aaaaaaaa" || r.State != "landing" {
		t.Errorf("row 71 = %+v", r)
	}
	if r.Phase != prGap || r.Track != "gap" {
		t.Errorf("no persisted record carries a live PR phase; it must stay a gap: %+v", r)
	}
	if r.Checks != "not read" {
		t.Errorf("live checks are a remote fact and must be 'not read': %q", r.Checks)
	}
	if r.Wait.State != "recorded" || r.Wait.Retries != "1 of 3" || strings.Contains(r.Wait.NextAction, "ghp_") {
		t.Errorf("wait = %+v", r.Wait)
	}
	if by[72].Phase != "merged-issue-open" || by[72].Track != "merge" {
		t.Errorf("row 72 = %+v", by[72])
	}
	if by[73].Phase != "issue-closed" || by[73].Track != "closed" || !by[73].Completed {
		t.Errorf("row 73 = %+v", by[73])
	}
	if by[73].SessionID != prGap {
		t.Errorf("a released record has no holder; the session id must be a gap: %+v", by[73])
	}
	if by[74].PR != prGap || by[74].Phase != prGap {
		t.Errorf("no receipt means no PR: %+v", by[74])
	}
	if by[75].Phase != "repair-dispatched" || by[75].Track != "repair" {
		t.Errorf("row 75 = %+v", by[75])
	}
	// The store's own open-PR rule: handed off and not done → 71, 72, 75.
	if b := vm.Backlog; !b.Known || b.Open != 3 || b.Limit != 4 || b.State != "within-limit" || b.QueueKnown {
		t.Errorf("backlog = %+v", b)
	}
}

func TestRecordRows_NoLimitAndFailures(t *testing.T) {
	s, dir := recordStore(t, 0)
	seedRecords(t, dir)
	if b := recordApp(t, s).buildPRTracking(context.Background()).Backlog; b.State != "no-limit-set" || b.Open != 3 {
		t.Errorf("a store with no limit set must read no-limit-set: %+v", b)
	}
	if vm := recordApp(t, nil).buildPRTracking(context.Background()); vm.Available || !strings.Contains(vm.Note, "not configured") {
		t.Errorf("a nil source is not configured: %+v", vm)
	}
	bad := failingSource{errors.New("disk gone")}
	if vm := recordApp(t, bad).buildPRTracking(context.Background()); vm.Available || !strings.Contains(vm.Note, "disk gone") {
		t.Errorf("a failing store must be reported, not hidden: %+v", vm)
	}
	empty, _ := recordStore(t, 2)
	if vm := recordApp(t, empty).buildPRTracking(context.Background()); !vm.Available || len(vm.Rows) != 0 || vm.Backlog.Open != 0 {
		t.Errorf("an empty store is available with no rows: %+v", vm)
	}
}

type failingSource struct{ err error }

func (f failingSource) List() ([]claim.Record, error) { return nil, f.err }
func (f failingSource) MaxOpenPRs() int               { return 0 }

func TestRecordRows_ReadOnlyAndRendered(t *testing.T) {
	s, dir := recordStore(t, 4)
	seedRecords(t, dir)
	before := dirSnapshot(t, dir)
	a := recordApp(t, s)
	m, _ := a.buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	for _, want := range []string{"#501", "agent-3", "wait recorded", "open PRs 3/4", "merged-issue-open"} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor missing %q", want)
		}
	}
	if strings.Contains(html, "ghp_") {
		t.Errorf("a secret in a persisted text reached the page")
	}
	if after := dirSnapshot(t, dir); after != before {
		t.Errorf("rendering changed the store directory:\nbefore %s\nafter  %s", before, after)
	}
}

func dirSnapshot(t *testing.T, dir string) string {
	t.Helper()
	var sb strings.Builder
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			fi, _ := d.Info()
			sb.WriteString(p + ":" + itoa(int(fi.Size())) + ":" + fi.ModTime().String() + ";")
		}
		return nil
	})
	return sb.String()
}
