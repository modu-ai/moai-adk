// todo_jev_finding_test.go — SPEC-JEV-CONSUMERS-001 M4 (Consumer C): the
// admission-path write, the render form, the admission-only scope, the
// write-path `agent` prohibition, and the display-only queue-hash guard.
//
// Every absence asserted here carries a positive control, because the search
// that matches nothing and the search that never ran produce the same output.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// installJevProbe replaces the Consumer C seam for one test and restores it
// afterwards. The count is the positive control every absence assertion in
// this file leans on.
func installJevProbe(t *testing.T, judge func(rec *kanban.BacklogRecord, item kanban.BacklogItem) (jevNearDuplicateJudgment, bool)) *int {
	t.Helper()
	calls := 0
	prev := jevNearDuplicateProbe
	jevNearDuplicateProbe = func(rec *kanban.BacklogRecord, item kanban.BacklogItem) (jevNearDuplicateJudgment, bool) {
		calls++
		return judge(rec, item)
	}
	t.Cleanup(func() { jevNearDuplicateProbe = prev })
	return &calls
}

// alwaysNearDuplicateOf answers with a fixed related id and probability.
func alwaysNearDuplicateOf(relatedID string, p float64) func(*kanban.BacklogRecord, kanban.BacklogItem) (jevNearDuplicateJudgment, bool) {
	return func(_ *kanban.BacklogRecord, _ kanban.BacklogItem) (jevNearDuplicateJudgment, bool) {
		return jevNearDuplicateJudgment{RelatedID: relatedID, Probability: p}, true
	}
}

// TestJevFinding_WrittenAtAdmission — REQ-JEVN-001: the admission path records
// the judgment as a finding whose Source is the third constant, naming the NEW
// card as the subject (the orientation the mechanical write path already uses).
func TestJevFinding_WrittenAtAdmission(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "alpha one"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	calls := installJevProbe(t, alwaysNearDuplicateOf("t1", 0.87))

	if _, _, err := runTodo(t, "add", "beta two"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if *calls == 0 {
		t.Fatal("positive control failed: the admission path never consulted the Jev seam")
	}

	var got []kanban.BacklogFinding
	for _, f := range loadFindings(t, store) {
		if f.Source == kanban.BacklogSourceJev {
			got = append(got, f)
		}
	}
	if len(got) != 1 {
		t.Fatalf("jev-sourced findings = %d, want exactly 1: %+v", len(got), loadFindings(t, store))
	}
	if got[0].SubjectID != "t2" || got[0].RelatedID != "t1" {
		t.Errorf("finding = %s->%s, want t2->t1 (the new card is the subject)", got[0].SubjectID, got[0].RelatedID)
	}
	if got[0].Relation != kanban.BacklogRelationNearDuplicate {
		t.Errorf("relation = %q, want %q", got[0].Relation, kanban.BacklogRelationNearDuplicate)
	}
	if got[0].Score != 0.87 {
		t.Errorf("probability carried through as %v, want 0.87", got[0].Score)
	}
}

// TestJevFinding_PrecedenceHalfA_Suppressed — AC-JEVN-003 sub-cases 1 and 2
// (REQ-JEVN-006 half (a)) at the WRITE PATH: a Jev finding is not appended
// when a finding of ANY source already names that unordered pair with that
// relation. Both existing sources are exercised, and the reversed-pair
// orientation is used for the agent case so an ordered comparison fails here.
func TestJevFinding_PrecedenceHalfA_Suppressed(t *testing.T) {
	for _, existing := range []string{kanban.BacklogSourceMechanical, kanban.BacklogSourceAgent} {
		t.Run(existing, func(t *testing.T) {
			_, store := todoFixture(t)
			if _, _, err := runTodo(t, "add", "alpha one"); err != nil {
				t.Fatalf("seed add: %v", err)
			}
			// Seeded ahead of the arrival, naming the pair {t1, t2} — in the
			// reversed orientation for the agent case.
			subject, related := "t2", "t1"
			if existing == kanban.BacklogSourceAgent {
				subject, related = "t1", "t2"
			}
			seedFindings(t, store, kanban.BacklogFinding{
				SubjectID: subject, RelatedID: related,
				Relation: kanban.BacklogRelationNearDuplicate, Source: existing,
			})
			before := len(loadFindings(t, store))

			calls := installJevProbe(t, alwaysNearDuplicateOf("t1", 0.87))
			if _, _, err := runTodo(t, "add", "beta two"); err != nil {
				t.Fatalf("add: %v", err)
			}
			if *calls == 0 {
				t.Fatal("positive control failed: the seam was never consulted, so the absence below is unattributable")
			}

			after := loadFindings(t, store)
			if len(after) != before {
				t.Errorf("findings = %d, want %d — the arriving Jev finding was appended over an existing %s finding",
					len(after), before, existing)
			}
			for _, f := range after {
				if f.Source == kanban.BacklogSourceJev {
					t.Errorf("a jev-sourced finding landed despite an existing %s finding for the pair: %+v", existing, f)
				}
			}
		})
	}
}

// TestJevFindingLine_DistinctFromMechanicalScore — AC-JEVN-004 (REQ-JEVN-005):
// a Jev probability renders in a form a reader cannot mistake for the text
// analyser's similarity score, and the pre-existing mechanical and agent forms
// are unchanged (including the `machine-only` mark's meaning).
func TestJevFindingLine_DistinctFromMechanicalScore(t *testing.T) {
	rec := &kanban.BacklogRecord{}
	mk := func(source string) kanban.BacklogFinding {
		return kanban.BacklogFinding{
			SubjectID: "t2", RelatedID: "t1",
			Relation: kanban.BacklogRelationNearDuplicate, Source: source, Score: 0.87,
		}
	}

	jevLine := todoFindingLine(rec, "t2", mk(kanban.BacklogSourceJev))
	mechLine := todoFindingLine(rec, "t2", mk(kanban.BacklogSourceMechanical))
	agentLine := todoFindingLine(rec, "t2", mk(kanban.BacklogSourceAgent))

	// Regression control on the two pre-existing forms first: if these drift,
	// the distinctness assertion below is measuring the wrong baseline.
	if !strings.Contains(mechLine, "score 0.87") {
		t.Fatalf("positive control failed: the mechanical line lost its score form: %q", mechLine)
	}
	if !strings.Contains(mechLine, "machine-only") {
		t.Fatalf("positive control failed: the mechanical line lost the machine-only mark: %q", mechLine)
	}
	if strings.Contains(agentLine, "score") {
		t.Errorf("the agent line gained a score: %q", agentLine)
	}
	if strings.Contains(agentLine, "machine-only") {
		t.Errorf("the agent line gained the machine-only mark: %q", agentLine)
	}

	if strings.Contains(jevLine, "score") {
		t.Errorf("the jev line renders the word %q — a model probability must not read as a measured similarity: %q",
			"score", jevLine)
	}
	if !strings.Contains(jevLine, jev.SignalLabel) {
		t.Errorf("the jev line carries no %q marker: %q", jev.SignalLabel, jevLine)
	}
	if !strings.Contains(jevLine, "0.87") {
		t.Errorf("the jev line dropped the probability entirely: %q", jevLine)
	}
	if !strings.Contains(jevLine, kanban.BacklogSourceJev) {
		t.Errorf("the jev line does not name its source: %q", jevLine)
	}
	// A Jev finding never carries the machine-only mark: the mark is about
	// agent-sourced records and is rendered only for mechanical findings.
	if strings.Contains(jevLine, "machine-only") {
		t.Errorf("the jev line carries the machine-only mark: %q", jevLine)
	}
}

// TestJevFinding_AdmissionOnly_ReSweepNeverCallsIt — AC-JEVN-013 (the
// admission-only scope clause of REQ-JEVN-001): the `moai todo analyze`
// re-sweep records zero Jev requests. The positive control runs the admission
// path against the SAME counting stub, so a stub that never counted could not
// produce the zero above.
func TestJevFinding_AdmissionOnly_ReSweepNeverCallsIt(t *testing.T) {
	_, _ = todoFixture(t)
	calls := installJevProbe(t, alwaysNearDuplicateOf("t1", 0.87))

	for _, text := range []string{"queue drain stalls on lock", "queue drain stalls on locks"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add %q: %v", text, err)
		}
	}
	admissionCalls := *calls
	if admissionCalls == 0 {
		t.Fatal("positive control failed: the admission path recorded zero Jev requests, so the " +
			"re-sweep zero below would be unattributable")
	}

	if _, _, err := runTodo(t, "analyze"); err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if got := *calls - admissionCalls; got != 0 {
		t.Errorf("the `todo analyze` re-sweep issued %d Jev requests, want 0 — Consumer C is admission-only", got)
	}
}

// TestJevProducers_NeverWriteSourceAgent — AC-JEVN-012 (the write-path half of
// REQ-JEVN-003): no Jev producer in the diff sets BacklogSourceAgent. The
// positive control is the same search over internal/cli/todo_relate.go, which
// does set it — without the control, a search with a typo'd symbol produces
// the same empty output as a clean one.
func TestJevProducers_NeverWriteSourceAgent(t *testing.T) {
	const symbol = "BacklogSourceAgent"
	producers := []string{
		filepath.Join("todo_jev_finding.go"),
		filepath.Join("todo_analysis.go"),
	}

	read := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path) // #nosec G304 -- fixed in-repository source path
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(data)
	}

	// Positive control FIRST: the search must be shown to fire before its
	// silence anywhere else is read as an absence.
	control := read("todo_relate.go")
	if !strings.Contains(control, symbol) {
		t.Fatalf("positive control failed: %q not found in todo_relate.go, so the absences below are unmeasured", symbol)
	}

	for _, path := range producers {
		body := read(path)
		if !strings.Contains(body, "BacklogSourceJev") {
			t.Fatalf("%s does not name BacklogSourceJev — it is not a Jev producer, so scanning it "+
				"asserts nothing about the Jev write path", path)
		}
		if strings.Contains(body, symbol) {
			t.Errorf("%s names %s on a Jev production path — a Jev finding written as `agent` would make "+
				"the queue claim a review that never happened", path, symbol)
		}
	}
}

// TestJevFinding_WritesNoCardField — AC-JEVN-005 (REQ-JEVN-007): recording a
// Jev finding writes no card field. Measured as a SHA-256 over the queue's
// card array with the newly admitted card excluded — the card admission itself
// adds is not a consequence of the finding — plus an assertion that the only
// delta in the record is the one appended finding.
func TestJevFinding_WritesNoCardField(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "alpha one"); err != nil {
		t.Fatalf("seed add: %v", err)
	}

	load := func() *kanban.BacklogRecord {
		t.Helper()
		rec, err := store.Load()
		if err != nil {
			t.Fatalf("load backlog: %v", err)
		}
		return rec
	}
	// Every field of every card enters the digest, so a field written as a
	// consequence of the finding cannot hide behind a field-by-field
	// comparison someone forgot to extend when the card gained a field.
	digestItems := func(items []kanban.BacklogItem) string {
		h := sha256.New()
		for _, it := range items {
			raw, err := json.Marshal(it)
			if err != nil {
				t.Fatalf("marshal item: %v", err)
			}
			h.Write(raw)
		}
		return hex.EncodeToString(h.Sum(nil))
	}
	engineDigest := func() string {
		t.Helper()
		data, err := os.ReadFile(store.EnginePath()) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatalf("read backlog engine: %v", err)
		}
		sum := sha256.Sum256(data)
		return hex.EncodeToString(sum[:])
	}

	before := load()
	beforeDigest := digestItems(before.Items)
	beforeEngine := engineDigest()

	calls := installJevProbe(t, alwaysNearDuplicateOf("t1", 0.87))
	if _, _, err := runTodo(t, "add", "beta two"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if *calls == 0 {
		t.Fatal("positive control failed: no Jev finding was produced, so this test measures nothing")
	}

	after := load()
	if len(after.Items) != len(before.Items)+1 {
		t.Fatalf("items = %d, want %d (the admitted card and nothing else)", len(after.Items), len(before.Items)+1)
	}
	if got := digestItems(after.Items[:len(before.Items)]); got != beforeDigest {
		t.Errorf("pre-existing card bytes changed: %s -> %s — a card field was written as a consequence of a finding",
			beforeDigest, got)
	}
	if len(after.Findings) != len(before.Findings)+1 {
		t.Fatalf("findings = %d, want %d (exactly the appended Jev finding)",
			len(after.Findings), len(before.Findings)+1)
	}
	if appended := after.Findings[len(after.Findings)-1]; appended.Source != kanban.BacklogSourceJev {
		t.Errorf("appended finding source = %q, want %q", appended.Source, kanban.BacklogSourceJev)
	}
	// Control on the digest itself: the engine bytes DID move, so the
	// item-digest equality above is a preserved value rather than a
	// measurement taken on a file nothing wrote to.
	if engineDigest() == beforeEngine {
		t.Fatal("positive control failed: the queue engine bytes are unchanged, so this test " +
			"compared two reads of the same untouched state")
	}
}

// TestJevProbe_DisabledCapabilityProducesNothing — REQ-JEVC-017 at this
// consumer: the shipped default is off, and while off the consumer records
// nothing and constructs no request. The live probe is exercised directly
// against a fixture whose config carries no jev block.
func TestJevProbe_DisabledCapabilityProducesNothing(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "alpha one"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	if _, _, err := runTodo(t, "add", "alpha two"); err != nil {
		t.Fatalf("add: %v", err)
	}
	for _, f := range loadFindings(t, store) {
		if f.Source == kanban.BacklogSourceJev {
			t.Errorf("a jev-sourced finding was recorded while the capability is off: %+v", f)
		}
	}
	// Positive control: the seam IS reached on the admission path, so the
	// absence above is the gate's doing rather than an unwired call site.
	calls := installJevProbe(t, func(*kanban.BacklogRecord, kanban.BacklogItem) (jevNearDuplicateJudgment, bool) {
		return jevNearDuplicateJudgment{}, false
	})
	if _, _, err := runTodo(t, "add", "alpha three"); err != nil {
		t.Fatalf("control add: %v", err)
	}
	if *calls == 0 {
		t.Fatal("positive control failed: the admission path does not reach the Jev seam at all")
	}
}
