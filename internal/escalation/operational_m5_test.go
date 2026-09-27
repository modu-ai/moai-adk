package escalation_test

import (
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/civerdict"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// fixtureHead is the fixture worktree's HEAD SHA; otherHead is any other.
const fixtureHead = escalationtest.HeadSHA

const otherHead = "ffffffffffffffffffffffffffffffffffffffff"

// budgetEdit sets the contract's budget before signing.
func budgetEdit(ops, retries string) func(string) string {
	return func(d string) string {
		return strings.Replace(d, "\nescalate_on:",
			"\nbudget:\n  turns: 60\n  operations: "+ops+"\n  audit_retries: "+retries+"\n\nescalate_on:", 1)
	}
}

// armedWith arms a contract-mode worktree whose contract carries the draft edit.
func armedWith(t *testing.T, card string, edit func(string) string, autonomyExtra string) *escalationtest.Worktree {
	t.Helper()
	w := escalationtest.NewWorktree(t, card)
	w.WriteContractMode(autonomyExtra)
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Edit: edit})
	preWrite(contractSettings(t, w), w, "internal/fixture/a.go")
	if !cardLog(t, w).Armed() {
		t.Fatalf("card did not arm: %+v", cardLog(t, w).Entries)
	}
	return w
}

func postBash(t *testing.T, w *escalationtest.Worktree, cmd string, failed bool, diag string) {
	t.Helper()
	escalation.Observe(contractSettings(t, w), escalation.Event{Hook: escalation.HookPostToolUse, CWD: w.Root,
		ToolName: "Bash", Command: cmd, Failed: failed, Diagnostic: diag})
}

// AC-AE-014 (REQ-AE-014): a fourth operation against budget.operations 3
// writes budget-exceeded naming operations 4/3; after a disarm, operations
// still count against the budget recorded at arming; a worktree with no
// resolved contract counts against budget_default.
func TestBudgetExceededTrips(t *testing.T) {
	isolateStore(t)
	w := armedWith(t, "t9001", budgetEdit("3", "2"), "")
	for i := 0; i < 3; i++ {
		postBash(t, w, "ls", false, "")
	}
	if rs, _ := recordsOfClass(t, w, escalation.ClassBudgetExceeded); len(rs) != 0 {
		t.Fatalf("budget tripped at 3 operations")
	}
	postBash(t, w, "ls", false, "")
	rs, raw := recordsOfClass(t, w, escalation.ClassBudgetExceeded)
	if len(rs) != 1 || !strings.Contains(raw[0], "operations observed 4, limit 3") ||
		!strings.Contains(refLine(t, w, rs[0].ContractRef), "operations: 3") {
		t.Fatalf("budget record = %+v\n%s", rs, strings.Join(raw, "\n"))
	}

	// Disarm (contract deleted); further operations still count against 3.
	if err := os.Remove(w.Path(".moai/specs/SPEC-A-001/contract.yaml")); err != nil {
		t.Fatal(err)
	}
	postBash(t, w, "ls", false, "")
	if cardLog(t, w).Armed() {
		t.Fatal("card still armed without its contract")
	}
	rs, _ = recordsOfClass(t, w, escalation.ClassBudgetExceeded)
	if len(rs) != 1 || rs[0].Occurrences != 2 {
		t.Errorf("after disarm: budget record occurrences = %+v", rs)
	}

	// No resolved contract at all: budget_default.
	w2 := escalationtest.NewWorktree(t, "t9002")
	w2.WriteContractMode("escalation:\n  budget_default:\n    operations: 1")
	postBash(t, w2, "ls", false, "")
	postBash(t, w2, "ls", false, "")
	rs2, _ := recordsOfClass(t, w2, escalation.ClassBudgetExceeded)
	if len(rs2) != 1 || rs2[0].ContractRef != "config:workflow.autonomy.escalation.budget_default.operations" {
		t.Errorf("budget_default record = %+v", rs2)
	}
}

// Class 7 turns: Stop events count against budget.turns.
func TestBudgetTurnsCountedAtStop(t *testing.T) {
	isolateStore(t)
	w := armedWith(t, "t9001", func(d string) string {
		return strings.Replace(d, "\nescalate_on:", "\nbudget:\n  turns: 2\n  operations: 40\n  audit_retries: 2\n\nescalate_on:", 1)
	}, "")
	s := contractSettings(t, w)
	for i := 0; i < 3; i++ {
		escalation.Observe(s, escalation.Event{Hook: escalation.HookStop, CWD: w.Root})
	}
	rs, raw := recordsOfClass(t, w, escalation.ClassBudgetExceeded)
	if len(rs) != 1 || !strings.Contains(raw[0], "turns observed 3, limit 2") {
		t.Errorf("turns record = %+v\n%s", rs, strings.Join(raw, "\n"))
	}
}

// AC-AE-015 part 1 (REQ-AE-015): three consecutive failures with one
// fingerprint write one same-diagnostic-repeat record; fail, fail, pass,
// fail writes none.
func TestSameDiagnosticRepeatTrips(t *testing.T) {
	isolateStore(t)
	w := armedWith(t, "t9001", nil, "")
	for _, failed := range []bool{true, true, false, true} {
		postBash(t, w, "go test ./x", failed, "--- FAIL: TestX")
	}
	if rs, _ := recordsOfClass(t, w, escalation.ClassSameDiagnosticRepeat); len(rs) != 0 {
		t.Fatalf("interrupted sequence tripped: %+v", rs)
	}
	postBash(t, w, "go test ./x", true, "--- FAIL: TestX")
	postBash(t, w, "go test ./x", true, "--- FAIL: TestX")
	rs, _ := recordsOfClass(t, w, escalation.ClassSameDiagnosticRepeat)
	if len(rs) != 1 || rs[0].ContractRef != "rule:same-diagnostic-3" || rs[0].Kind != escalation.KindOperational {
		t.Fatalf("same-diagnostic records = %+v", rs)
	}
	// A fourth consecutive failure with the same fingerprint
	// increments the one record rather than writing a second.
	postBash(t, w, "go test ./x", true, "--- FAIL: TestX")
	if rs, _ := recordsOfClass(t, w, escalation.ClassSameDiagnosticRepeat); len(rs) != 1 || rs[0].Occurrences != 2 {
		t.Errorf("fourth failure: %+v", rs)
	}
}

// AC-AE-015 part 2 (REQ-AE-016): FAIL at iteration audit_retries+1 trips
// audit-fail-at-retry-cap; FAIL at an earlier iteration does not.
func TestAuditFailAtRetryCapTrips(t *testing.T) {
	t.Run("retries-0", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", budgetEdit("40", "0"), "")
		w.Write(".moai/reports/t9001/plan-audit-iter1.md", "# plan audit\n\nIteration: 1\nVerdict: **FAIL**\n")
		commitCheckpoint(contractSettings(t, w), w)
		rs, _ := recordsOfClass(t, w, escalation.ClassAuditFailAtRetryCap)
		if len(rs) != 1 || !strings.Contains(refLine(t, w, rs[0].ContractRef), "audit_retries") {
			t.Fatalf("records = %+v", rs)
		}
	})
	t.Run("retries-2", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", budgetEdit("40", "2"), "")
		s := contractSettings(t, w)
		w.Write(".moai/reports/t9001/sync-audit-iter1.md", "Verdict: FAIL\n")
		commitCheckpoint(s, w)
		if rs, _ := recordsOfClass(t, w, escalation.ClassAuditFailAtRetryCap); len(rs) != 0 {
			t.Fatalf("FAIL at iteration 1 of 3 tripped: %+v", rs)
		}
		w.Write(".moai/reports/t9001/sync-audit-iter3.md", "Verdict: FAIL\n")
		commitCheckpoint(s, w)
		if rs, _ := recordsOfClass(t, w, escalation.ClassAuditFailAtRetryCap); len(rs) != 1 {
			t.Fatalf("FAIL at iteration 3 did not trip: %+v", rs)
		}
	})
}

// AC-AE-012 (REQ-AE-010, REQ-AE-022). (a) disagreement_flag true and (b) a
// second_model verdict opposite the first each trip contradictory-evidence;
// (d) a null flag and (e) a local pass without a CI verdict record write
// none and are listed not-observed. Clause (c) — a recorded CI failure — is
// judged by the CI limb (SPEC-CI-VERDICT-PRODUCER-001): see
// TestContradictoryEvidenceCISameHeadTrips below.
func TestContradictoryEvidenceTrips(t *testing.T) {
	cases := []struct {
		name, body string
		trips      bool
		listed     string
	}{
		{"a-flag-true", `{"per_backend_verdicts":[{"backend":"claude","verdict":"pass"},{"backend":"glm","verdict":"pass"}],"overall_verdict":"pass","disagreement_flag":true}`, true, ""},
		{"b-second-model", `{"per_backend_verdicts":[{"backend":"claude","verdict":"pass"},{"backend":"codex","verdict":"fail"}],"overall_verdict":"fail","disagreement_flag":false}`, true, ""},
		{"d-flag-null", `{"per_backend_verdicts":[{"backend":"claude","verdict":"pass"}],"overall_verdict":"pass","disagreement_flag":null}`, false, "disagreement_flag"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateStore(t)
			w := armedWith(t, "t9001", nil, "")
			w.Write(".moai/state/audit-multi/session-1.json", tc.body)
			commitCheckpoint(contractSettings(t, w), w)
			rs, _ := recordsOfClass(t, w, escalation.ClassContradictoryEvidence)
			if tc.trips != (len(rs) == 1) || len(rs) > 1 {
				t.Fatalf("contradictory-evidence records = %d, want trip=%v", len(rs), tc.trips)
			}
			no := notObservedEntries(t, w)
			all := ""
			for _, e := range no {
				all += strings.Join(e.NotObserved, "|")
			}
			if tc.listed != "" && !strings.Contains(all, tc.listed) {
				t.Errorf("not-observed = %q, want %q listed", all, tc.listed)
			}
			if !strings.Contains(all, "ci verdict") {
				t.Errorf("the CI verdict is not listed not-observed: %q", all)
			}
		})
	}
	t.Run("e-no-ci-verdict", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", nil, "")
		postBash(t, w, "go test ./...", false, "")
		commitCheckpoint(contractSettings(t, w), w)
		if rs, _ := recordsOfClass(t, w, escalation.ClassContradictoryEvidence); len(rs) != 0 {
			t.Fatalf("records without any recorded verdict: %+v", rs)
		}
		no := notObservedEntries(t, w)
		if len(no) == 0 || !slices.ContainsFunc(no[len(no)-1].NotObserved, func(s string) bool { return strings.Contains(s, "ci verdict") }) {
			t.Errorf("checkpoint does not list the CI verdict: %+v", no)
		}
	})
}

// writeCIVerdict records a CI verdict for head via the civerdict store (the
// producer's write path, REQ-CV-004).
func writeCIVerdict(t *testing.T, w *escalationtest.Worktree, head, conclusion, runID string) {
	t.Helper()
	rec := civerdict.Record{
		HeadSHA:    head,
		Conclusion: conclusion,
		RunID:      runID,
		ObservedAt: time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Producer:   "test",
	}
	if err := civerdict.Save(w.Root, rec); err != nil {
		t.Fatalf("civerdict.Save: %v", err)
	}
}

// writeLocalPass records a verify snapshot at head with one exit-0 check
// entry (the REQ-CV-006 local-pass source).
func writeLocalPass(t *testing.T, w *escalationtest.Worktree, head string) {
	t.Helper()
	snap := &verify.Snapshot{
		Key: head + ":testdigest",
		Checks: []verify.CheckEntry{{
			CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: time.Now(),
		}},
	}
	if err := verify.Save(w.Root, snap); err != nil {
		t.Fatalf("verify.Save: %v", err)
	}
}

// lastNotObserved joins the not-observed items of the worktree's last
// checkpoint entry.
func lastNotObserved(t *testing.T, w *escalationtest.Worktree) string {
	t.Helper()
	no := notObservedEntries(t, w)
	if len(no) == 0 {
		return ""
	}
	return strings.Join(no[len(no)-1].NotObserved, "|")
}

// AC-CV-005 (REQ-CV-006, REQ-CV-007) — clause (c), the previously dead limb:
// a recorded local pass and a recorded CI failure at the SAME head trip class
// contradictory-evidence once, through the existing contradiction path, and
// the completed observation is not listed under not_observed.
func TestContradictoryEvidenceCISameHeadTrips(t *testing.T) {
	isolateStore(t)
	w := armedWith(t, "t9001", nil, "")
	writeLocalPass(t, w, fixtureHead)
	writeCIVerdict(t, w, fixtureHead, civerdict.ConclusionFailure, "run-1")
	commitCheckpoint(contractSettings(t, w), w)
	rs, raw := recordsOfClass(t, w, escalation.ClassContradictoryEvidence)
	if len(rs) != 1 {
		t.Fatalf("contradictory-evidence records = %d, want 1\n%s", len(rs), strings.Join(raw, "\n"))
	}
	if !strings.Contains(raw[0], "passed locally") || !strings.Contains(raw[0], "CI failed") ||
		!strings.Contains(raw[0], fixtureHead) {
		t.Errorf("observation does not name the local pass and the CI failure at the head:\n%s", raw[0])
	}
	if all := lastNotObserved(t, w); strings.Contains(all, "ci verdict") {
		t.Errorf("completed CI observation listed not-observed: %q", all)
	}
}

// AC-CV-006 (REQ-CV-008) — no-trip dispositions of the CI limb.
func TestContradictoryEvidenceCINoTrip(t *testing.T) {
	cases := []struct {
		name       string
		head       string
		conclusion string
		localPass  bool
		listed     string // substring required in not_observed; "" = the limb must list nothing ci-related
	}{
		// (a) CI failure recorded for a different head.
		{"different-head", otherHead, civerdict.ConclusionFailure, true, "ci verdict"},
		// (b) success at the head completes the observation: nothing listed.
		{"success-with-pass", fixtureHead, civerdict.ConclusionSuccess, true, ""},
		// (c) CI failure at the head without a recorded local pass.
		{"failure-no-local-pass", fixtureHead, civerdict.ConclusionFailure, false, "local"},
		// (e) neutral at the head is a completed observation: nothing listed.
		{"neutral-with-pass", fixtureHead, civerdict.ConclusionNeutral, true, ""},
		// (f) success at the head without a local pass: only the missing pass listed.
		{"success-no-local-pass", fixtureHead, civerdict.ConclusionSuccess, false, "local"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateStore(t)
			w := armedWith(t, "t9001", nil, "")
			if tc.localPass {
				writeLocalPass(t, w, fixtureHead)
			}
			writeCIVerdict(t, w, tc.head, tc.conclusion, "run-1")
			commitCheckpoint(contractSettings(t, w), w)
			if rs, raw := recordsOfClass(t, w, escalation.ClassContradictoryEvidence); len(rs) != 0 {
				t.Fatalf("contradictory-evidence records written: %+v\n%s", rs, strings.Join(raw, "\n"))
			}
			all := lastNotObserved(t, w)
			switch {
			case tc.listed != "" && !strings.Contains(all, tc.listed):
				t.Errorf("not-observed = %q, want %q listed", all, tc.listed)
			case tc.listed == "" && (strings.Contains(all, "ci verdict") || strings.Contains(all, "local verification pass")):
				t.Errorf("not-observed = %q, want nothing ci-related listed", all)
			}
			// The success or neutral conclusion at the head itself must never
			// surface under not_observed (REQ-CV-008): cases (b), (e), and (f)
			// list at most the missing local pass.
			if tc.conclusion != civerdict.ConclusionFailure && strings.Contains(all, "conclusion "+tc.conclusion) {
				t.Errorf("completed CI observation listed not-observed: %q", all)
			}
		})
	}
}

// AC-CV-007 (REQ-CV-009, REQ-CV-004) — the idempotent producer re-run cannot
// re-trip a resolved record without new evidence; a changed CI record re-trips
// once (freshEvidence content-hash gate).
func TestContradictoryEvidenceCIFreshness(t *testing.T) {
	isolateStore(t)
	w := armedWith(t, "t9001", nil, "")
	s := contractSettings(t, w)
	writeLocalPass(t, w, fixtureHead)
	writeCIVerdict(t, w, fixtureHead, civerdict.ConclusionFailure, "run-1")
	commitCheckpoint(s, w)
	if n := openCount(t, w, escalation.ClassContradictoryEvidence); n != 1 {
		t.Fatalf("open=%d after the first trip, want 1", n)
	}
	resolveAll(t, w, escalation.ClassContradictoryEvidence)
	// Idempotent producer re-run: byte-identical record rewrite.
	writeCIVerdict(t, w, fixtureHead, civerdict.ConclusionFailure, "run-1")
	commitCheckpoint(s, w)
	if rs, _ := recordsOfClass(t, w, escalation.ClassContradictoryEvidence); len(rs) != 1 {
		t.Errorf("records after identical re-record = %d, want 1 (no duplicate)", len(rs))
	}
	if n := openCount(t, w, escalation.ClassContradictoryEvidence); n != 0 {
		t.Errorf("resolved record re-opened without new evidence: open=%d", n)
	}
	// Genuinely new evidence: a changed run_id re-trips once as -2.
	writeCIVerdict(t, w, fixtureHead, civerdict.ConclusionFailure, "run-2")
	commitCheckpoint(s, w)
	if !anySuffix(records(t, w), "-2.md") || openCount(t, w, escalation.ClassContradictoryEvidence) != 1 {
		t.Errorf("changed CI record did not re-trip as -2: %v", records(t, w))
	}
}
