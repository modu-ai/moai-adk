package escalation_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// Under a non-contract mode Observe reads and writes nothing: no store file,
// no record (REQ-AE-001).
func TestObserveInertUnderGuided(t *testing.T) {
	home := isolateStore(t)
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteMode("guided", false)
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	s := contractSettings(t, w)
	preWrite(s, w, "internal/bar/x.go")
	escalation.Observe(s, escalation.Event{Hook: escalation.HookPostToolUse, CWD: w.Root, ToolName: "Bash", Command: "ls"})
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Errorf("guided mode wrote under MOAI_HOME: %v", entries)
	}
	if got := records(t, w); len(got) != 0 {
		t.Errorf("guided mode wrote records: %v", got)
	}
}

// The M2 class 3 predicate: a write inside the worktree root that no
// ownership.write glob covers, or that an effective_never glob matches, trips
// ownership-move with contract_ref pointing at the tripped contract line; the
// card's report directory, .moai/state, and ownership.scratch are exempt; a
// repeat of the same write increments one record.
func TestOwnershipMoveInsideRoot(t *testing.T) {
	isolateStore(t)
	w, s := armedWorktree(t, "t9001")
	cdata := w.Read(".moai/specs/SPEC-A-001/contract.yaml")

	for _, rel := range []string{
		"internal/fixture/ok.go",             // ownership.write
		".moai/reports/t9001/verdict.md",     // card report dir
		".moai/state/x.json",                 // state dir
		".moai/state/verify/out.txt",         // scratch (and state)
		".moai/specs/SPEC-A-001/progress.md", // ownership.write (.moai/specs/<ID>/**)
	} {
		preWrite(s, w, rel)
	}
	if got := records(t, w); len(got) != 0 {
		t.Fatalf("exempt or owned writes tripped: %v", got)
	}

	preWrite(s, w, "internal/bar/x.go")
	preWrite(s, w, "internal/bar/x.go")
	escalation.Observe(s, escalation.Event{Hook: escalation.HookPreToolUse, CWD: w.Path("internal"),
		ToolName: "Edit", FilePath: "x/secret.go"}) // relative to CWD: internal/x/secret.go (never)
	rs, raw := recordsOfClass(t, w, escalation.ClassOwnershipMove)
	if len(rs) != 2 {
		t.Fatalf("ownership-move records = %d (%v), want 2", len(rs), records(t, w))
	}
	var sawWrite, sawNever bool
	for i, r := range rs {
		line, err := strconv.Atoi(strings.TrimPrefix(r.ContractRef, "contract.yaml:"))
		if err != nil {
			t.Fatalf("contract_ref %q", r.ContractRef)
		}
		text := lineAt(t, cdata, line)
		switch {
		case strings.Contains(raw[i], "internal/bar/x.go"):
			sawWrite = true
			if r.Occurrences != 2 || !strings.Contains(text, "write:") {
				t.Errorf("write-miss record occurrences=%d line %q", r.Occurrences, text)
			}
		case strings.Contains(raw[i], "internal/x/secret.go"):
			sawNever = true
			if !strings.Contains(text, "internal/x/**") {
				t.Errorf("never record points at %q", text)
			}
		}
		if r.EscalateOn != "ownership-move" || r.Spec != "SPEC-A-001" || r.HeadSHA != escalationtest.HeadSHA {
			t.Errorf("record frontmatter = %+v", r)
		}
	}
	if !sawWrite || !sawNever {
		t.Errorf("records did not cover both the write miss and the never match")
	}

	// A write outside the worktree root is not judged in this milestone.
	escalation.Observe(s, escalation.Event{Hook: escalation.HookPreToolUse, CWD: w.Root, ToolName: "Write",
		FilePath: filepath.Join(filepath.Dir(w.Root), "elsewhere.txt")})
	if n := len(records(t, w)); n != 2 {
		t.Errorf("outside-root write changed records: %v", records(t, w))
	}
}

// Class 7 on an armed card counts against the contract budget and points at
// the contract's budget.operations line; after a disarm it keeps counting
// against the budget recorded at arming (REQ-AE-014).
func TestBudgetOperationsArmedAndAfterDisarm(t *testing.T) {
	isolateStore(t)
	w, s := armedWorktree(t, "t9001")
	s.BudgetDefault.Operations = 1000 // must not be the limit used
	cdata := w.Read(".moai/specs/SPEC-A-001/contract.yaml")
	post := func() {
		escalation.Observe(s, escalation.Event{Hook: escalation.HookPostToolUse, CWD: w.Root, ToolName: "Bash", Command: "ls"})
	}
	for i := 0; i < 41; i++ {
		post()
	}
	rs, raw := recordsOfClass(t, w, escalation.ClassBudgetExceeded)
	if len(rs) != 1 || !strings.Contains(raw[0], "operations observed 41, limit 40") {
		t.Fatalf("budget records = %+v", rs)
	}
	if want := "contract.yaml:"; !strings.HasPrefix(rs[0].ContractRef, want) {
		t.Fatalf("contract_ref = %q", rs[0].ContractRef)
	}
	line, _ := strconv.Atoi(strings.TrimPrefix(rs[0].ContractRef, "contract.yaml:"))
	if l := lineAt(t, cdata, line); !strings.Contains(l, "operations") {
		t.Errorf("budget contract_ref points at %q", l)
	}

	// Disarm by deleting the contract: contract-absent, one record, one entry.
	if err := os.Remove(w.Path(".moai/specs/SPEC-A-001/contract.yaml")); err != nil {
		t.Fatal(err)
	}
	post()
	dis, _ := recordsOfClass(t, w, escalation.ClassDetectionDisarmed)
	if len(dis) != 1 || dis[0].ContractRef != "disarm:contract-absent" {
		t.Fatalf("disarm records = %+v", dis)
	}
	if cardLog(t, w).Armed() {
		t.Fatal("card armed after its contract was deleted")
	}
	rs, _ = recordsOfClass(t, w, escalation.ClassBudgetExceeded)
	if len(rs) != 1 || rs[0].Occurrences != 2 {
		t.Errorf("budget after disarm = %+v, want the same record incremented", rs)
	}
}

// A contract edited after signing is seen at the next PreToolUse because its
// bytes no longer match the cached digest: signature-invalid, one record.
func TestPreToolUseSeesChangedContractBytes(t *testing.T) {
	isolateStore(t)
	w, s := armedWorktree(t, "t9001")
	w.Replace(".moai/specs/SPEC-A-001/contract.yaml", "Implement the fixture feature", "Implement another feature")
	preWrite(s, w, "internal/fixture/a.go")
	dis, _ := recordsOfClass(t, w, escalation.ClassDetectionDisarmed)
	if len(dis) != 1 || dis[0].ContractRef != "disarm:signature-invalid" {
		t.Fatalf("disarm records = %+v", dis)
	}
	lg := cardLog(t, w)
	if lg.Armed() || !lg.ChainIntact {
		t.Errorf("armed=%v chain=%v after disarm", lg.Armed(), lg.ChainIntact)
	}
	// The resolver's not-armed line follows the disarmed entry.
	var kinds []string
	for _, e := range lg.Entries {
		kinds = append(kinds, e.Kind)
	}
	joined := strings.Join(kinds, ",")
	if !strings.Contains(joined, "disarmed,not-armed") {
		t.Errorf("log kinds %s: not-armed does not follow disarmed", joined)
	}
}
