// todo_relate_test.go — acceptance tests for the agent-layer relation verbs
// and the operator visibility surface: relate, unrelate, why, and the
// finding lines `list` renders.
//
// Every test runs on a t.TempDir() fixture; the operator's live queue is
// never a subject.
package cli

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// findingLines returns the indented finding lines of a `todo list` render.
func findingLines(out string) []string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "\t↳ ") {
			lines = append(lines, line)
		}
	}
	return lines
}

// TestTodoRelateAndUnrelateTouchNoCard — AC-TA-004 (REQ-TA-008, REQ-TA-009):
// `relate` adds exactly one finding and `unrelate` removes exactly the one
// addressed, with the items array byte-identical at both points.
//
// The lengths are asserted exactly in both directions, so an implementation
// that reclaims too little (the addressed finding survives) and one that
// reclaims too much (the unrelated t3↔t4 finding is swept away) both fail —
// and naming the survivor catches the implementation that removed the wrong
// one while keeping the count right.
func TestTodoRelateAndUnrelateTouchNoCard(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta card", "Gamma card", "Delta card")
	seedFindings(t, store, kanban.BacklogFinding{
		SubjectID: "t3", RelatedID: "t4",
		Relation: kanban.BacklogRelationContains, Source: kanban.BacklogSourceAgent,
	})
	snapshot := snapshotItems(t, store)

	if _, _, err := runTodo(t, "relate", "t2", "t1", "--relation", "absorbs",
		"--note", "t2 covers t1"); err != nil {
		t.Fatalf("relate: %v", err)
	}
	afterRelate := loadFindings(t, store)
	if len(afterRelate) != 2 {
		t.Fatalf("findings after relate = %d, want 2: %+v", len(afterRelate), afterRelate)
	}
	recorded := -1
	for i, f := range afterRelate {
		if f.SubjectID == "t2" && f.RelatedID == "t1" &&
			f.Relation == kanban.BacklogRelationAbsorbs && f.Source == kanban.BacklogSourceAgent {
			recorded = i + 1
		}
	}
	if recorded < 0 {
		t.Fatalf("no {t2, t1, absorbs, agent} finding was recorded: %+v", afterRelate)
	}
	assertItemsUnchanged(t, store, snapshot, "after relate")

	if _, _, err := runTodo(t, "unrelate", strconv.Itoa(recorded)); err != nil {
		t.Fatalf("unrelate: %v", err)
	}
	afterUnrelate := loadFindings(t, store)
	if len(afterUnrelate) != 1 {
		t.Fatalf("findings after unrelate = %d, want exactly 1: %+v",
			len(afterUnrelate), afterUnrelate)
	}
	survivor := afterUnrelate[0]
	if !survivor.Names("t3") || !survivor.Names("t4") {
		t.Errorf("the surviving finding = %+v, want the untouched t3↔t4 one", survivor)
	}
	for _, f := range afterUnrelate {
		if f.Names("t2") && f.Names("t1") && f.Relation == kanban.BacklogRelationAbsorbs {
			t.Errorf("the addressed finding survived unrelate: %+v", f)
		}
	}
	assertItemsUnchanged(t, store, snapshot, "after unrelate")
}

// TestSemanticRelationsChangeNothing — AC-TA-005 (REQ-TA-009, REQ-TA-010):
// all four semantic relations are recorded, and then the read verbs leave
// the file byte-identical.
//
// The Given cannot be built without `relate`, so this test cannot pass by
// the feature being absent — the structural answer to an absence claim that
// would otherwise be satisfied by a queue with no analysis layer at all.
func TestSemanticRelationsChangeNothing(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta card", "Gamma card")

	for _, rel := range []struct{ a, b, relation string }{
		{"t1", "t2", kanban.BacklogRelationContains},
		{"t2", "t3", kanban.BacklogRelationAbsorbs},
		{"t1", "t3", kanban.BacklogRelationReplaces},
		{"t3", "t1", kanban.BacklogRelationConflicts},
	} {
		if _, _, err := runTodo(t, "relate", rel.a, rel.b, "--relation", rel.relation); err != nil {
			t.Fatalf("relate %s %s %s: %v", rel.a, rel.relation, rel.b, err)
		}
	}
	if got := len(loadFindings(t, store)); got != 4 {
		t.Fatalf("findings = %d, want 4 — one per relation", got)
	}
	before := queueDigest(t, store)

	for _, args := range [][]string{
		{"list"}, {"next"}, {"why", "t1"}, {"list", "--json"},
	} {
		if _, _, err := runTodo(t, args...); err != nil {
			t.Errorf("%v: %v", args, err)
		}
	}
	if after := queueDigest(t, store); after != before {
		t.Errorf("a read command wrote to the queue: %s -> %s", before, after)
	}
}

// TestTodoListJSONIsIdempotent — AC-TA-007 (REQ-TA-010): two consecutive
// `list --json` runs emit identical bytes over an identical file.
//
// This is the assertion that forbids analysing on read: an implementation
// re-analysing (or re-stamping finding timestamps) each time renders a queue
// that reorganizes itself between two looks.
func TestTodoListJSONIsIdempotent(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta card")
	seedFindings(t, store, kanban.BacklogFinding{
		SubjectID: "t2", RelatedID: "t1",
		Relation: kanban.BacklogRelationNearDuplicate, Source: kanban.BacklogSourceMechanical,
		Score: 0.9, At: "2026-01-01T00:00:00Z",
	})
	before := queueDigest(t, store)

	first, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("first list: %v", err)
	}
	middle := queueDigest(t, store)
	second, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	if first != second {
		t.Errorf("two list --json runs diverged:\nfirst=%q\nsecond=%q", first, second)
	}
	if !strings.Contains(first, "\"findings\"") {
		t.Errorf("list --json omits the findings array:\n%s", first)
	}
	after := queueDigest(t, store)
	if middle != before || after != before {
		t.Errorf("list --json wrote to the queue: %s -> %s -> %s", before, middle, after)
	}
}

// TestTodoListShowsFindingAndNextStep — AC-TA-009 (REQ-TA-011): the finding
// is visible on the screen the operator actually looks at, and it names the
// command that would act on it.
//
// A finding carried only in --json is invisible to the operator, and an
// analysis nobody sees cannot inform the decision this feature exists to
// inform.
func TestTodoListShowsFindingAndNextStep(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta", "Gamma", "Delta", "Epsilon")
	seedFindings(t, store, kanban.BacklogFinding{
		SubjectID: "t5", RelatedID: "t1",
		Relation: kanban.BacklogRelationNearDuplicate, Source: kanban.BacklogSourceMechanical,
		Score: 0.87, At: "2026-01-01T00:00:00Z",
	})

	out, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var line string
	for _, candidate := range findingLines(out) {
		if strings.Contains(candidate, "t1") {
			line = candidate
			break
		}
	}
	if line == "" {
		t.Fatalf("no indented finding line under t5:\n%s", out)
	}
	for _, want := range []string{"near-duplicate", "t1", "mechanical"} {
		if !strings.Contains(line, want) {
			t.Errorf("finding line %q omits %q", line, want)
		}
	}
	if !strings.Contains(line, "moai todo drop") && !strings.Contains(line, "moai todo edit") {
		t.Errorf("finding line %q names no command the operator could run", line)
	}
}

// TestTodoWhySaysNothingFound — AC-TA-010 (REQ-TA-012): a card with no
// findings gets an explicit answer.
//
// Silence is the failure this catches: an empty stdout reads exactly like a
// crash, and an operator who cannot tell "nothing found" from "the command
// broke" stops trusting the verb.
func TestTodoWhySaysNothingFound(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card")

	out, _, err := runTodo(t, "why", "t1")
	if err != nil {
		t.Fatalf("why: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("why printed nothing for a card with no findings")
	}
	if !strings.Contains(out, "no findings") {
		t.Errorf("why output carries no explicit no-findings statement:\n%s", out)
	}
}

// TestMachineOnlyMarkAppearsAndClears — AC-TA-011 (REQ-TA-013): the mark is
// present while only the analyser has spoken, and gone once an agent
// finding names the same pair.
//
// The agent finding is written in the OPPOSITE direction (t1→t5 against the
// analyser's t5→t1) on purpose: an implementation comparing pairs as ordered
// tuples leaves the mark in place, and only the reverse direction catches
// it. Asserting one half alone would pass an implementation with no mark at
// all, or one whose mark can never clear.
func TestMachineOnlyMarkAppearsAndClears(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta", "Gamma", "Delta", "Epsilon")
	seedFindings(t, store, kanban.BacklogFinding{
		SubjectID: "t5", RelatedID: "t1",
		Relation: kanban.BacklogRelationNearDuplicate, Source: kanban.BacklogSourceMechanical,
		Score: 0.87, At: "2026-01-01T00:00:00Z",
	})

	before, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("first list: %v", err)
	}
	if !marksMachineOnly(before) {
		t.Errorf("no near-duplicate line carries machine-only:\n%s", before)
	}

	if _, _, err := runTodo(t, "relate", "t1", "t5", "--relation", "replaces"); err != nil {
		t.Fatalf("relate: %v", err)
	}
	after, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	if marksMachineOnly(after) {
		t.Errorf("machine-only survived an agent finding on the same unordered pair:\n%s", after)
	}
}

// marksMachineOnly reports whether any near-duplicate line in a list render
// still carries the machine-only mark.
func marksMachineOnly(out string) bool {
	for _, line := range findingLines(out) {
		if strings.Contains(line, "near-duplicate") && strings.Contains(line, "machine-only") {
			return true
		}
	}
	return false
}

// TestTodoNewVerbsAreHeadless — AC-TA-013 (REQ-TA-015): every new verb runs
// to a definite exit with stdin closed, and `--help` lists all four.
//
// The --help assertion is the positive control. A test asserting only "the
// prompt guard reports nothing" would pass on a surface where the verbs were
// never added — the guard finds no prompt in a file that does not exist.
func TestTodoNewVerbsAreHeadless(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta card")

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"add", []string{"add", "Gamma card"}},
		{"analyze", []string{"analyze"}},
		{"relate", []string{"relate", "t1", "t2", "--relation", "conflicts"}},
		{"why", []string{"why", "t1"}},
		{"list", []string{"list"}},
		{"unrelate", []string{"unrelate", "1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := runTodoWithClosedStdin(t, tc.args...); err != nil {
				t.Errorf("%v did not complete cleanly with stdin closed: %v", tc.args, err)
			}
		})
	}

	help, _, err := runTodo(t, "--help")
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	for _, verb := range []string{"analyze", "relate", "unrelate", "why"} {
		if !strings.Contains(help, verb) {
			t.Errorf("--help does not list %q:\n%s", verb, help)
		}
	}
}

// runTodoWithClosedStdin runs a todo verb with a closed stdin, so a verb
// that tried to read from the terminal would error rather than block.
func runTodoWithClosedStdin(t *testing.T, args ...string) error {
	t.Helper()
	closed, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	if err := closed.Close(); err != nil {
		t.Fatalf("close stdin stand-in: %v", err)
	}
	cmd := newTodoCmd()
	cmd.SetIn(closed)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	return cmd.Execute()
}

// TestRelateAndUnrelateRefusals pins the refusal paths, each of which is a
// way the verbs could otherwise write a record that means nothing: a
// relation outside the four, a card that is not in the queue, a card
// related to itself, a duplicate of a record already present, and an index
// addressing no finding. The queue is asserted untouched after all of them.
func TestRelateAndUnrelateRefusals(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta card")
	if _, _, err := runTodo(t, "relate", "t1", "t2", "--relation", "contains"); err != nil {
		t.Fatalf("seed relate: %v", err)
	}
	before := queueDigest(t, store)

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"unknown relation", []string{"relate", "t1", "t2", "--relation", "near-duplicate"}},
		{"missing relation flag", []string{"relate", "t1", "t2"}},
		{"absent card", []string{"relate", "t1", "t9", "--relation", "conflicts"}},
		{"self relation", []string{"relate", "t1", "t1", "--relation", "conflicts"}},
		{"already recorded", []string{"relate", "t1", "t2", "--relation", "contains"}},
		{"index past the end", []string{"unrelate", "9"}},
		{"index zero", []string{"unrelate", "0"}},
		{"index not a number", []string{"unrelate", "second"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := runTodo(t, tc.args...); err == nil {
				t.Errorf("%v was accepted, want a refusal", tc.args)
			}
		})
	}

	if after := queueDigest(t, store); after != before {
		t.Errorf("a refused verb wrote to the queue: %s -> %s", before, after)
	}
	if got := len(loadFindings(t, store)); got != 1 {
		t.Errorf("findings = %d, want the single seeded one", got)
	}
}
