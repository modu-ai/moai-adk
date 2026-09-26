package escalation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// resolveAll sets every open record of the class to resolved, as a decider would.
func resolveAll(t *testing.T, w *escalationtest.Worktree, class string) {
	t.Helper()
	dir := escalation.RecordDir(w.Root, w.Card)
	for _, name := range records(t, w) {
		if !strings.HasPrefix(name, class+"-") {
			continue
		}
		p := filepath.Join(dir, name)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		out := strings.Replace(strings.Replace(string(data), "status: open", "status: resolved", 1),
			"decider: \"\"", "decider: fixture-decider", 1)
		if err := os.WriteFile(p, []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// openCount counts the class's records with status open.
func openCount(t *testing.T, w *escalationtest.Worktree, class string) int {
	t.Helper()
	rs, _ := recordsOfClass(t, w, class)
	n := 0
	for _, r := range rs {
		if r.Status == escalation.StatusOpen {
			n++
		}
	}
	return n
}

// Sync-audit F1: classes 5 and 9 re-read persistent evidence files at every
// commit. A resolved record re-opens only when the evidence changed; the same
// unchanged file at the next commit is not a new observation. Genuinely new
// evidence still re-trips as <class>-<fingerprint>-2.md (AC-AE-023).
func TestResolvedEvidenceRecordReopensOnlyOnNewEvidence(t *testing.T) {
	t.Run("class9", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", budgetEdit("40", "0"), "")
		s := contractSettings(t, w)
		w.Write(".moai/reports/t9001/plan-audit-iter1.md", "Verdict: FAIL\n")
		commitCheckpoint(s, w)
		if n := openCount(t, w, escalation.ClassAuditFailAtRetryCap); n != 1 {
			t.Fatalf("class9_open=%d after the first checkpoint, want 1", n)
		}
		resolveAll(t, w, escalation.ClassAuditFailAtRetryCap)
		commitCheckpoint(s, w)
		if n := openCount(t, w, escalation.ClassAuditFailAtRetryCap); n != 0 {
			t.Errorf("open_after_resolve_and_recommit=%d, want 0 (records %v)", n, records(t, w))
		}
		// New evidence: the same file now records a different FAIL.
		w.Write(".moai/reports/t9001/plan-audit-iter1.md", "Verdict: FAIL\n\nNew finding F9.\n")
		commitCheckpoint(s, w)
		if !anySuffix(records(t, w), "-2.md") || openCount(t, w, escalation.ClassAuditFailAtRetryCap) != 1 {
			t.Errorf("new evidence did not re-trip as -2: %v", records(t, w))
		}
	})
	t.Run("class5", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", nil, "")
		s := contractSettings(t, w)
		body := `{"per_backend_verdicts":[{"backend":"claude","verdict":"pass"},{"backend":"glm","verdict":"pass"}],"disagreement_flag":true}`
		w.Write(".moai/state/audit-multi/session-1.json", body)
		commitCheckpoint(s, w)
		if n := openCount(t, w, escalation.ClassContradictoryEvidence); n != 1 {
			t.Fatalf("class5_open=%d after the first checkpoint, want 1", n)
		}
		resolveAll(t, w, escalation.ClassContradictoryEvidence)
		commitCheckpoint(s, w)
		if n := openCount(t, w, escalation.ClassContradictoryEvidence); n != 0 {
			t.Errorf("open_after_resolve_and_recommit=%d, want 0 (records %v)", n, records(t, w))
		}
		w.Write(".moai/state/audit-multi/session-1.json", strings.Replace(body, `"glm"`, `"codex"`, 1))
		commitCheckpoint(s, w)
		if !anySuffix(records(t, w), "-2.md") || openCount(t, w, escalation.ClassContradictoryEvidence) != 1 {
			t.Errorf("new evidence did not re-trip as -2: %v", records(t, w))
		}
	})
}

// Re-audit N1: the class 7 audit_retries count is re-derived from the audit
// files at every commit. A resolved budget record re-opens only when the
// per-kind count actually rises; an unchanged count at the next commit is not
// a new observation. A higher count still re-trips as -2.
func TestResolvedAuditRetriesBudgetReopensOnlyOnHigherCount(t *testing.T) {
	isolateStore(t)
	w := armedWith(t, "t9001", budgetEdit("40", "0"), "")
	s := contractSettings(t, w)
	w.Write(".moai/reports/t9001/plan-audit-iter1.md", "Verdict: PASS\n")
	w.Write(".moai/reports/t9001/plan-audit-iter2.md", "Verdict: PASS\n")
	commitCheckpoint(s, w)
	if n := openCount(t, w, escalation.ClassBudgetExceeded); n != 1 {
		t.Fatalf("budget_open=%d after the first checkpoint, want 1 (records %v)", n, records(t, w))
	}
	resolveAll(t, w, escalation.ClassBudgetExceeded)
	commitCheckpoint(s, w)
	if n := openCount(t, w, escalation.ClassBudgetExceeded); n != 0 {
		t.Errorf("open_after_resolve_and_recommit=%d, want 0 (records %v)", n, records(t, w))
	}
	// A genuinely higher count: a third iteration of the same kind.
	w.Write(".moai/reports/t9001/plan-audit-iter3.md", "Verdict: PASS\n")
	commitCheckpoint(s, w)
	if !anySuffix(records(t, w), "-2.md") || openCount(t, w, escalation.ClassBudgetExceeded) != 1 {
		t.Errorf("higher count did not re-trip as -2: %v", records(t, w))
	}
}

// anySuffix reports whether any name ends with suffix.
func anySuffix(names []string, suffix string) bool {
	for _, n := range names {
		if strings.HasSuffix(n, suffix) {
			return true
		}
	}
	return false
}
