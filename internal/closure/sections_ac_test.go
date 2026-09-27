package closure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// mustBuild runs Build and fails the test on error.
func mustBuild(t *testing.T, in BuildInput) *Report {
	t.Helper()
	r, err := Build(in)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return r
}

// jsonKeyOrder decodes the top-level key order of a marshaled report.
func jsonKeyOrder(t *testing.T, r *Report) []string {
	t.Helper()
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return rawJSONKeyOrder(t, data)
}

func rawJSONKeyOrder(t *testing.T, data []byte) []string {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(string(data)))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		t.Fatalf("expected object start, got %v %v", tok, err)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("token: %v", err)
		}
		keys = append(keys, tok.(string))
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("decode value: %v", err)
		}
	}
	return keys
}

// TestAC_CLOSURE_002 — section order, form parity, determinism.
func TestAC_CLOSURE_002(t *testing.T) {
	f := newACFixture(t)
	in := f.input()
	r := mustBuild(t, in)

	want := []string{
		"schema_version", "card", "spec_id", "head_sha", "generated_at",
		"mode", "second_review_policy",
		"summary", "kickoff", "reconciliation", "invariants", "ownership",
		"new_apis", "escalations", "first_verdict", "second_verdict",
		"plan_audit_binding", "not_performed", "residual_risk",
		"human_verdict", "sources",
	}
	got := jsonKeyOrder(t, r)
	if len(got) != len(want) {
		t.Fatalf("key count = %d, want %d", len(got), len(want))
	}
	for i, k := range want {
		if got[i] != k {
			t.Fatalf("key[%d] = %q, want %q", i, got[i], k)
		}
	}
	if r.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", r.SchemaVersion)
	}

	// Markdown headings in the same order with the same values.
	md := RenderMarkdown(r)
	headings := []string{
		"## Summary", "## Kickoff", "## Contract Reconciliation", "## Invariants",
		"## Ownership", "## New APIs", "## Escalations", "## First Verdict",
		"## Second Verdict", "## Plan-Audit Binding", "## Not Performed",
		"## Residual Risk", "## Human Verdict",
	}
	pos := -1
	for _, h := range headings {
		next := strings.Index(md[pos+1:], h)
		if next < 0 {
			t.Fatalf("markdown heading %q missing or out of order", h)
		}
		pos += next + 1
	}

	// Determinism: two builds byte-identical apart from generated_at.
	r2 := mustBuild(t, in)
	savedGen, savedGen2 := r.GeneratedAt, r2.GeneratedAt
	r.GeneratedAt, r2.GeneratedAt = "", ""
	first, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	second, err := json.Marshal(r2)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	r.GeneratedAt, r2.GeneratedAt = savedGen, savedGen2
	if string(first) != string(second) {
		t.Fatalf("two builds differ apart from generated_at:\n%s\n---\n%s", first, second)
	}
}

// TestAC_CLOSURE_003 — missing inputs are listed, never passed.
func TestAC_CLOSURE_003(t *testing.T) {
	f := newACFixture(t)
	// Remove progress.md; leave no records/state/second-review/verdict files.
	if err := os.Remove(filepath.Join(f.CardDir, ".moai", "specs", fixSpecID, "progress.md")); err != nil {
		t.Fatalf("remove progress.md: %v", err)
	}
	in := f.input()
	in.Records, in.UnreadableRecords = nil, nil
	in.Verdicts = nil
	in.DetectorArmed = false
	r := mustBuild(t, in)

	tokens := map[string]bool{}
	for _, np := range r.NotPerformed {
		tokens[np.Item] = true
	}
	for _, want := range []string{
		NotPerformedProgressMissing,
		NotPerformedSecondReviewNotPerformed,
		NotPerformedEscalationDetectorNotArmed,
		NotPerformedHumanVerdictNone,
	} {
		if !tokens[want] {
			t.Fatalf("not_performed missing %q: %v", want, r.NotPerformed)
		}
	}
	// No section renders a pass state.
	if r.HumanVerdict.State != HumanVerdictNoneRecorded {
		t.Fatalf("human verdict state = %q, want none recorded", r.HumanVerdict.State)
	}
	if r.SecondVerdict.State != SecondReviewNotPerformed {
		t.Fatalf("second verdict state = %q, want not-performed", r.SecondVerdict.State)
	}
	if r.FirstVerdict.RunStatus != NotRecorded {
		t.Fatalf("run_status = %q, want not recorded", r.FirstVerdict.RunStatus)
	}
	if r.Ownership.Count.Observed {
		t.Fatalf("ownership count observed without an armed detector")
	}
}

// TestAC_CLOSURE_004 — contract reconciliation per AC.
func TestAC_CLOSURE_004(t *testing.T) {
	f := newACFixture(t)
	// Keep only the first ID's row; replace the other two with one unknown
	// row (AC-004: rows for the first ID plus a fourth ID absent from
	// acceptance.md).
	progress := strings.Replace(fixProgress,
		"| AC-FIXTURE-002 | PASS | `go test -run TestAC ok` |\n| AC-FIXTURE-003 | FAIL | `go test -run TestAC missing` |",
		"| AC-EXTRA-009 | PASS | not in acceptance.md |", 1)
	if progress == fixProgress {
		t.Fatalf("fixture replacement did not apply")
	}
	f.write(filepath.Join(f.CardDir, ".moai", "specs", fixSpecID, "progress.md"), progress)

	in := f.input()
	rep := in.Verify
	r := mustBuild(t, in)

	rows := map[string]ReconciliationRow{}
	for _, row := range r.Reconciliation.Rows {
		rows[row.ID] = row
	}
	first, ok := rows["AC-FIXTURE-001"]
	if !ok || first.Status != "PASS" || first.Evidence != "`go test -run TestAC ok`" {
		t.Fatalf("first row = %+v, want PASS with evidence verbatim", first)
	}
	for _, id := range []string{"AC-FIXTURE-002", "AC-FIXTURE-003"} {
		row, ok := rows[id]
		if !ok || row.Status != "not reported" {
			t.Fatalf("%s row = %+v, want not reported", id, row)
		}
	}
	extra, ok := rows["AC-EXTRA-009"]
	if !ok || extra.Status != "unknown" {
		t.Fatalf("extra row = %+v, want unknown", extra)
	}
	if rep.Acceptance.SHA256 == nil || r.Reconciliation.RecordedAcceptanceSHA != *rep.Acceptance.SHA256 {
		t.Fatalf("recorded acceptance hash = %q", r.Reconciliation.RecordedAcceptanceSHA)
	}
	if r.Reconciliation.MeasuredACCount != rep.Acceptance.MeasuredACCount || r.Reconciliation.MeasuredACCount != 3 {
		t.Fatalf("measured AC count = %d", r.Reconciliation.MeasuredACCount)
	}
	if r.Reconciliation.VerifyState != contract.StateSignedValid {
		t.Fatalf("reconciliation verify state = %q", r.Reconciliation.VerifyState)
	}
}

// TestAC_CLOSURE_005 — invariant results never claim an unobserved pass.
func TestAC_CLOSURE_005(t *testing.T) {
	f := newACFixture(t)
	in := f.input()
	// Swap the contract view for one carrying the four invariants of the AC
	// (pure verify over synthetic inputs; the fixture's own contract keeps
	// its two plain invariants for signed-validity).
	draft := draftContract(fixCard, fixSpecID,
		[]string{"constitution:X-1", "frozen-files", "go test ./pkg/...", "make check"},
		[]string{"src/**", ".moai/specs/" + fixSpecID + "/**"})
	rep := contract.Verify(contract.Inputs{
		SpecID: fixSpecID, Contract: []byte(draft),
		Policy: contract.Policy{Mode: "contract", SecondReview: "required"},
	})
	in.Verify = rep

	open := escalation.Record{
		SchemaVersion: 1, Card: fixCard, Spec: fixSpecID,
		Kind: escalation.KindContract, Class: escalation.ClassInvariantViolation,
		Fingerprint: "f1", EscalateOn: "go test ./pkg/...",
		Status: escalation.StatusOpen, Occurrences: 1,
	}
	in.Records = []escalation.Record{open}

	t.Run("armed", func(t *testing.T) {
		in.DetectorArmed = true
		r := mustBuild(t, in)
		want := map[string]string{
			"constitution:X-1":  InvariantNotObserved,
			"frozen-files":      InvariantNoViolation,
			"go test ./pkg/...": InvariantViolationOpen,
			"make check":        InvariantNotObserved,
		}
		for _, row := range r.Invariants.Rows {
			if row.Result != want[row.Invariant] {
				t.Fatalf("invariant %q result = %q, want %q", row.Invariant, row.Result, want[row.Invariant])
			}
		}
	})

	t.Run("disarmed", func(t *testing.T) {
		in.DetectorArmed = false
		r := mustBuild(t, in)
		for _, row := range r.Invariants.Rows {
			if row.Result == InvariantNoViolation {
				t.Fatalf("invariant %q shows %q while disarmed", row.Invariant, row.Result)
			}
		}
		if r.Invariants.Rows[1].Result != InvariantNotObserved {
			t.Fatalf("frozen-files result = %q, want not observed while disarmed", r.Invariants.Rows[1].Result)
		}
	})
}

// TestAC_CLOSURE_006 — ownership count.
func TestAC_CLOSURE_006(t *testing.T) {
	f := newACFixture(t)
	in := f.input()

	open := escalation.Record{
		SchemaVersion: 1, Card: fixCard, Kind: escalation.KindContract,
		Class: escalation.ClassOwnershipMove, Fingerprint: "o1",
		Status: escalation.StatusOpen, Occurrences: 1,
	}
	resolved := open
	resolved.Fingerprint = "o2"
	resolved.Status = escalation.StatusResolved
	resolved.Occurrences = 2

	t.Run("two records armed count 2", func(t *testing.T) {
		in.Records = []escalation.Record{open, resolved}
		in.DetectorArmed = true
		r := mustBuild(t, in)
		if len(r.Ownership.Records) != 2 {
			t.Fatalf("records = %d, want 2", len(r.Ownership.Records))
		}
		if !r.Ownership.Count.Observed || r.Ownership.Count.N != 2 {
			t.Fatalf("count = %+v, want 2", r.Ownership.Count)
		}
	})
	t.Run("no records armed count 0", func(t *testing.T) {
		in.Records = nil
		in.DetectorArmed = true
		r := mustBuild(t, in)
		if !r.Ownership.Count.Observed || r.Ownership.Count.N != 0 {
			t.Fatalf("count = %+v, want observed 0", r.Ownership.Count)
		}
	})
	t.Run("no records disarmed not observed", func(t *testing.T) {
		in.Records = nil
		in.DetectorArmed = false
		r := mustBuild(t, in)
		if r.Ownership.Count.Observed {
			t.Fatalf("count = %+v, want not observed", r.Ownership.Count)
		}
	})
}

// TestAC_CLOSURE_007 — New APIs via the comparison seam.
func TestAC_CLOSURE_007(t *testing.T) {
	f := newACFixture(t)
	in := f.input()

	before := dirSnapshot(t, f.EscDir)
	in.NewAPICompare = func() ([]escalation.Addition, error) {
		return []escalation.Addition{
			{Kind: "new-package", Name: "fixture", Path: "internal/fixture"},
			{Kind: "cli-verb", Name: "fixture-cmd", Path: "internal/cli/fixture.go"},
		}, nil
	}
	r := mustBuild(t, in)
	if !r.NewAPIs.Observed || len(r.NewAPIs.Additions) != 2 {
		t.Fatalf("new apis = %+v, want two additions", r.NewAPIs)
	}
	if r.NewAPIs.Additions[0].Kind != "new-package" || r.NewAPIs.Additions[0].Path != "internal/fixture" {
		t.Fatalf("addition[0] = %+v", r.NewAPIs.Additions[0])
	}
	after := dirSnapshot(t, f.EscDir)
	if before != after {
		t.Fatalf("escalation directory changed: %q -> %q", before, after)
	}

	t.Run("comparison error", func(t *testing.T) {
		in.NewAPICompare = func() ([]escalation.Addition, error) {
			return nil, errCompare
		}
		r := mustBuild(t, in)
		if r.NewAPIs.Observed {
			t.Fatalf("observed = true, want not observed")
		}
		found := false
		for _, np := range r.NotPerformed {
			if np.Item == NotPerformedNewAPIComparisonUnavailable {
				found = true
			}
		}
		if !found {
			t.Fatalf("not_performed missing new-api-comparison-unavailable: %v", r.NotPerformed)
		}
	})
}

var errCompare = errors.New("comparison unavailable")

// TestAC_CLOSURE_008 — escalations and needs-decision.
func TestAC_CLOSURE_008(t *testing.T) {
	f := newACFixture(t)
	in := f.input()
	in.DetectorArmed = true

	f.writeA2Record(escalation.Record{
		SchemaVersion: 1, Card: fixCard, Kind: escalation.KindContract,
		Class: escalation.ClassContradictoryEvidence, Fingerprint: "e1",
		Status: escalation.StatusOpen, Occurrences: 3,
		ContractRef: "contract.yaml:4", Decider: "llm",
	})
	f.writeA2Record(escalation.Record{
		SchemaVersion: 1, Card: fixCard, Kind: escalation.KindOperational,
		Class: escalation.ClassBudgetExceeded, Fingerprint: "e2",
		Status: escalation.StatusResolved, Occurrences: 1,
		ContractRef: "budget", Decider: "human",
	})
	f.writeA2Record(escalation.Record{
		SchemaVersion: 1, Card: fixCard, Kind: escalation.KindRevoke,
		Class: escalation.ClassRevokeOperator, Fingerprint: "e3",
		Status: escalation.StatusResolved, Occurrences: 1,
	})
	f.writeA2Raw("broken.md", "not a record\n")
	// Records are loaded at input() time, which ran before these writes:
	// reload them now that the record files exist.
	in.Records, in.UnreadableRecords = mustLoadA2Records(t, f.CardDir, fixCard)

	r := mustBuild(t, in)
	if len(r.Escalations.Records) != 3 {
		t.Fatalf("records = %d, want 3 parsed", len(r.Escalations.Records))
	}
	if r.Escalations.NeedsDecision != "yes" {
		t.Fatalf("needs_decision = %q, want yes", r.Escalations.NeedsDecision)
	}
	if len(r.Escalations.Unreadable) != 1 || !strings.Contains(r.Escalations.Unreadable[0], "broken.md") {
		t.Fatalf("unreadable = %v, want broken.md", r.Escalations.Unreadable)
	}
	found := false
	for _, np := range r.NotPerformed {
		if np.Item == NotPerformedEscalationUnreadable {
			found = true
		}
	}
	if !found {
		t.Fatalf("not_performed missing escalation-unreadable")
	}

	t.Run("only resolved operational and revoke", func(t *testing.T) {
		f2 := newACFixture(t)
		in2 := f2.input()
		in2.DetectorArmed = true
		f2.writeA2Record(escalation.Record{
			SchemaVersion: 1, Card: fixCard, Kind: escalation.KindOperational,
			Class: escalation.ClassBudgetExceeded, Fingerprint: "f1",
			Status: escalation.StatusResolved, Occurrences: 1,
		})
		f2.writeA2Record(escalation.Record{
			SchemaVersion: 1, Card: fixCard, Kind: escalation.KindRevoke,
			Class: escalation.ClassRevokeOperator, Fingerprint: "f2",
			Status: escalation.StatusResolved, Occurrences: 1,
		})
		in2.Records, _ = mustLoadA2Records(f2.t, f2.CardDir, fixCard)
		r2 := mustBuild(f2.t, in2)
		if r2.Escalations.NeedsDecision != "no" {
			f2.t.Fatalf("needs_decision = %q, want no", r2.Escalations.NeedsDecision)
		}
	})
}

// TestAC_CLOSURE_009 — first verdict display and count mismatch.
func TestAC_CLOSURE_009(t *testing.T) {
	f := newACFixture(t)
	in := f.input() // §E.3: 2 pass, 1 fail; measured live ACs = 3 → 3 == 3, no mismatch
	r := mustBuild(t, in)
	if r.FirstVerdict.RunStatus != "complete" || r.FirstVerdict.ACPassCount != "2" ||
		r.FirstVerdict.ACFailCount != "1" || r.FirstVerdict.RunCommitSHA != "abc1234" {
		t.Fatalf("first verdict = %+v, want verbatim §E.3 values", r.FirstVerdict)
	}
	if r.FirstVerdict.CountMismatch {
		t.Fatalf("count mismatch = true, want false (2+1 == 3 measured)")
	}

	// Without run_commit_sha, the field shows not recorded.
	progress := strings.Replace(fixProgress, "run_commit_sha: abc1234\n", "", 1)
	f.write(filepath.Join(f.CardDir, ".moai", "specs", fixSpecID, "progress.md"), progress)
	in = f.input()
	r = mustBuild(t, in)
	if r.FirstVerdict.RunCommitSHA != NotRecorded {
		t.Fatalf("run_commit_sha = %q, want %q", r.FirstVerdict.RunCommitSHA, NotRecorded)
	}

	// pass+fail that does not sum to the measured count flags a mismatch.
	progress = strings.Replace(string(fixProgress), "ac_pass_count: 2", "ac_pass_count: 9", 1)
	f.write(filepath.Join(f.CardDir, ".moai", "specs", fixSpecID, "progress.md"), progress)
	in = f.input()
	r = mustBuild(t, in)
	if !r.FirstVerdict.CountMismatch {
		t.Fatalf("count mismatch = false, want true (9+1 != 3)")
	}
}

// dirSnapshot is a stable text summary of a directory tree (empty when
// absent), used to assert byte-identical directory state.
func dirSnapshot(t *testing.T, dir string) string {
	t.Helper()
	if _, err := os.Stat(dir); err != nil {
		return ""
	}
	var b strings.Builder
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		fmt.Fprintf(&b, "%s %d %s\n", e.Name(), info.Size(), info.ModTime().UTC().Format(time.RFC3339Nano))
	}
	return b.String()
}
