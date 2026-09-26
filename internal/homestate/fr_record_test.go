package homestate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Store-level coverage for the T1 record, the contract pointer, the `after`
// guard, and the unavailable-record log. The command-level acceptance tests
// in internal/cli exercise the same code through `moai factory` and the
// dispatch mirror; these pin the store contract on its own.

func strp(s string) *string { return &s }

func TestFR_RecordPickedCreateUpdateRefuse(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	c, err := db.RecordPicked(ctx, frRun, "p1", CardFields{HintPrefer: strp("backend=codex")}, "assign", frNow)
	if err != nil || c.State != CardPicked || c.Version != 1 || c.HintPrefer != "backend=codex" {
		t.Fatalf("create = %+v err=%v", c, err)
	}
	same, err := db.RecordPicked(ctx, frRun, "p1", CardFields{HintPrefer: strp("backend=codex")}, "assign", frNow)
	if err != nil || same.Version != 1 {
		t.Fatalf("no-op update = v%d err=%v, want v1", same.Version, err)
	}
	upd, err := db.RecordPicked(ctx, frRun, "p1", CardFields{HintAfter: strp("p0"), SpecID: strp("SPEC-DEMO-001"), WorktreePath: strp(t.TempDir())}, "assign", frNow)
	if err != nil || upd.Version != 2 || upd.HintAfter != "p0" || upd.SpecID != "SPEC-DEMO-001" {
		t.Fatalf("update = %+v err=%v", upd, err)
	}
	if n := len(frEvents(t, db, "card.fields")); n != 1 {
		t.Fatalf("card.fields events = %d, want 1", n)
	}
	cleared, err := db.RecordPicked(ctx, frRun, "p1", CardFields{HintAfter: strp("")}, "assign", frNow)
	if err != nil || cleared.HintAfter != "" {
		t.Fatalf("clear after = %+v err=%v", cleared, err)
	}
	if got, err := db.RecordPicked(ctx, frRun, "p1", CardFields{}, "assign", frNow); err != nil || got.Version != cleared.Version {
		t.Fatalf("empty fields = v%d err=%v, want unchanged", got.Version, err)
	}
	for name, f := range map[string]CardFields{
		"prefer without =":  {HintPrefer: strp("codex")},
		"after not an id":   {HintAfter: strp("../x")},
		"bad spec id":       {SpecID: strp("spec-lower")},
		"relative worktree": {WorktreePath: strp("rel/path")},
	} {
		if _, err := db.RecordPicked(ctx, frRun, "p2", f, "assign", frNow); !errors.Is(err, ErrInvalidCardInput) {
			t.Fatalf("%s: err = %v, want ErrInvalidCardInput", name, err)
		}
	}
	if _, err := db.RecordPicked(ctx, frRun, "../bad", CardFields{}, "assign", frNow); !errors.Is(err, ErrInvalidCardInput) {
		t.Fatalf("bad card id: err = %v", err)
	}
	frPlace(t, db, Card{RunID: frRun, CardID: "past", State: CardRun, Version: 1})
	if _, err := db.RecordPicked(ctx, frRun, "past", CardFields{HintPrefer: strp("a=b")}, "assign", frNow); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("fields on a card past picked: err = %v, want ErrIllegalTransition", err)
	}
	all, err := db.ListCards(ctx, "")
	if err != nil || len(all) != 2 {
		t.Fatalf("ListCards(all) = %d err=%v, want 2", len(all), err)
	}
}

func TestFR_ContractRefFormat(t *testing.T) {
	d, e := strings.Repeat("a", 64), strings.Repeat("b", 64)
	good := []string{
		"SPEC-EXAMPLE-001," + d + ",2026-09-26T09:00:00Z," + e,
		"SPEC-EXAMPLE-001," + d + ",2026-09-26T09:00:00Z",
	}
	for _, s := range good {
		if _, err := ParseContractRef(s); err != nil {
			t.Fatalf("ParseContractRef(%q) = %v", s, err)
		}
	}
	bad := []string{
		"SPEC-EXAMPLE-001," + d,
		"not-a-spec," + d + ",2026-09-26T09:00:00Z",
		"SPEC-EXAMPLE-001," + d[:63] + ",2026-09-26T09:00:00Z",
		"SPEC-EXAMPLE-001," + d + ",yesterday",
		"SPEC-EXAMPLE-001," + d + ",2026-09-26T09:00:00Z," + strings.Repeat("Z", 64),
	}
	for _, s := range bad {
		if _, err := ParseContractRef(s); !errors.Is(err, ErrInvalidCardInput) {
			t.Fatalf("ParseContractRef(%q) = %v, want ErrInvalidCardInput", s, err)
		}
	}
}

func TestFR_AfterGuardAcrossRuns(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	frPlace(t, db, Card{RunID: frRun, CardID: "succ", State: CardPicked, Version: 1, HintAfter: "pred"})
	req := TransitionRequest{RunID: frRun, CardID: "succ", To: CardAssigned, ExpectedVersion: 1, Actor: "assign", Owner: "worker-1", Now: frNow}
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrUnknownPredecessor) {
		t.Fatalf("no predecessor: err = %v", err)
	}
	frPlace(t, db, Card{RunID: "other-run", CardID: "pred", State: CardDone, Version: 1})
	if _, err := db.Transition(ctx, req); err != nil {
		t.Fatalf("predecessor done in another run: %v", err)
	}
}

func TestFR_StateHelpersAndResumeTarget(t *testing.T) {
	if len(CardStates()) != 19 {
		t.Fatalf("CardStates = %d, want 19", len(CardStates()))
	}
	for resume, want := range map[string]string{
		CardAssigned: CardAssigned, CardRun: CardAssigned, CardMerging: CardAssigned,
		CardPicked: CardPicked, CardMergedLocal: CardMergedLocal, CardPushed: CardPushed, CardCIGreen: CardCIGreen,
		CardDone: "", "": "",
	} {
		if got := ResumeTarget(resume); got != want {
			t.Fatalf("ResumeTarget(%q) = %q, want %q", resume, got, want)
		}
	}
	c := Card{State: CardRun, LeaseExpiresAt: "not-a-time"}
	if !c.LeaseExpired(frNow) {
		t.Fatal("an unparseable lease expiry must read as expired")
	}
}

func TestFR_RecordUnavailableLogAndReconcile(t *testing.T) {
	root := factorySandbox(t)
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	if entries, err := ReadRecordUnavailable(root, ""); err != nil || len(entries) != 0 {
		t.Fatalf("empty log = %v err=%v", entries, err)
	}
	for _, e := range []RecordUnavailableEntry{
		{RunID: frRun, CardID: "lost", Lane: "worker-2", Error: "boom"},
		{RunID: "other", CardID: "elsewhere", Lane: "worker-9", Error: "boom"},
	} {
		if err := AppendRecordUnavailable(root, e); err != nil {
			t.Fatal(err)
		}
	}
	if entries, _ := ReadRecordUnavailable(root, frRun); len(entries) != 1 || entries[0].CardID != "lost" || entries[0].ID == "" {
		t.Fatalf("run-filtered read = %+v", entries)
	}
	// A successful write for the run reconciles its entry, and only its entry.
	if _, err := db.RecordPicked(ctx, frRun, "next", CardFields{}, "assign", frNow); err != nil {
		t.Fatal(err)
	}
	drift := frEvents(t, db, "record.drift")
	if len(drift) != 1 || !strings.Contains(drift[0], `"card_id":"lost"`) || !strings.Contains(drift[0], `"factory_state":"absent"`) {
		t.Fatalf("record.drift = %v", drift)
	}
	if entries, _ := ReadRecordUnavailable(root, frRun); len(entries) != 0 {
		t.Fatalf("run entry still unreconciled: %+v", entries)
	}
	if entries, _ := ReadRecordUnavailable(root, "other"); len(entries) != 1 {
		t.Fatalf("other run's entry was touched: %+v", entries)
	}
	if _, err := db.RecordPicked(ctx, frRun, "next2", CardFields{}, "assign", frNow.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if n := len(frEvents(t, db, "record.drift")); n != 1 {
		t.Fatalf("second write appended drift again: %d", n)
	}
	// A corrupt log never blocks a record write.
	path, _ := RecordUnavailablePath(root)
	if err := os.WriteFile(path, []byte("{not json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecordPicked(ctx, frRun, "next3", CardFields{}, "assign", frNow); err != nil {
		t.Fatalf("write with a corrupt log: %v", err)
	}
	if _, err := ReadRecordUnavailable(root, ""); err == nil {
		t.Fatal("a corrupt log read without error")
	}
}

func TestFR_EvidenceReaderRefusals(t *testing.T) {
	repo := frNewRepo(t, true)
	ctx := context.Background()
	if _, err := verifyAuditEntry(ctx, repo.Dir, repo.Commit, "-flag"); !errors.Is(err, ErrEvidence) {
		t.Fatalf("option-shaped artifact: %v", err)
	}
	if _, err := resolveCommit(ctx, repo.Dir, "HEAD"); !errors.Is(err, ErrEvidence) {
		t.Fatalf("non-hex sha: %v", err)
	}
	if _, err := readAuditVerdict(repo.Dir, "../x", "plan-audit", repo.Commit); !errors.Is(err, ErrEvidence) {
		t.Fatalf("unsafe card id: %v", err)
	}
	if _, err := readAuditVerdict("", "c", "plan-audit", repo.Commit); !errors.Is(err, ErrEvidence) {
		t.Fatalf("no worktree: %v", err)
	}
	frWriteVerdict(t, repo.Dir, "two-sha", "plan-audit.md", "PASS", repo.Commit)
	frWrite(t, filepath.Join(repo.Dir, ".moai", "reports", "two-sha", "plan-audit.md"), "verdict: PASS\naudited_sha: "+repo.Commit+"\naudited_sha: "+repo.Merge+"\n")
	if _, err := readAuditVerdict(repo.Dir, "two-sha", "plan-audit", repo.Commit); !errors.Is(err, ErrEvidence) {
		t.Fatalf("conflicting audited_sha: %v", err)
	}
	frWriteVerdict(t, repo.Dir, "prefix", "plan-audit.md", "PASS", repo.Commit[:12])
	if _, err := readAuditVerdict(repo.Dir, "prefix", "plan-audit", repo.Commit); err != nil {
		t.Fatalf("12-char audited_sha prefix: %v", err)
	}
	if _, err := verifyMerge(ctx, repo.Dir, repo.Merge, repo.Remeasure, "-bad"); !errors.Is(err, ErrEvidence) {
		t.Fatalf("bad integration branch: %v", err)
	}
	if _, err := verifyMerge(ctx, repo.Dir, repo.Merge, "", repo.Integration); !errors.Is(err, ErrEvidence) {
		t.Fatalf("missing remeasure: %v", err)
	}
	if _, err := verifyPushed(ctx, repo.Dir, repo.Merge, "no-such-branch"); !errors.Is(err, ErrEvidence) {
		t.Fatalf("missing remote-tracking ref: %v", err)
	}
	if _, err := gitRead(ctx, "", "status"); err == nil {
		t.Fatal("gitRead without a worktree path succeeded")
	}
	if ok, err := hasRemote(ctx, filepath.Join(t.TempDir(), "missing")); err == nil || ok {
		t.Fatalf("hasRemote on a missing dir = %v, %v", ok, err)
	}
}
