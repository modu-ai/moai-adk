package cli

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/graph"
)

// Card t2011 — the pre-creation lookup of `todo add`. Every test runs on a
// t.TempDir() fixture store; the operator's live queue is never a subject.

func setCardState(t *testing.T, store *factory.BacklogStore, id string, state factory.BacklogState) {
	t.Helper()
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				rec.Items[i].State = state
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("set state: %v", err)
	}
}

func activeCards(t *testing.T, store *factory.BacklogStore, needle string) int {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, it := range rec.Items {
		if todoActiveState(it.State) && strings.Contains(it.Text, needle) {
			n++
		}
	}
	return n
}

func TestTodoAddRefGuard_RefusesASharedRefNamingTheActiveCard(t *testing.T) {
	for _, state := range []factory.BacklogState{factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStateHold} {
		t.Run(string(state), func(t *testing.T) {
			_, store := todoFixture(t)
			seedItems(t, store, "stabilise the flaky restore test from PR #1943")
			setCardState(t, store, "t1", state)
			before := queueDigest(t, store)

			for _, args := range [][]string{
				{"add", "a different wording about #1943 and its review comments"},
				{"add", "--pick", "yet another phrasing that cites #1943"},
				{"add", "--force", "forced wording that also cites #1943"},
			} {
				_, _, err := runTodo(t, args...)
				if err == nil {
					t.Fatalf("%v was admitted, want a refusal naming t1", args)
				}
				if !strings.Contains(err.Error(), "t1 ("+string(state)+")") {
					t.Fatalf("%v: refusal %q does not name t1 and its state %q", args, err, state)
				}
			}
			if after := queueDigest(t, store); after != before {
				t.Fatalf("a refusal changed the queue: %s -> %s", before, after)
			}
			// No id was consumed by the three refusals.
			if _, _, err := runTodo(t, "add", "an unrelated card with no reference"); err != nil {
				t.Fatal(err)
			}
			if ids := itemIDs(t, store); len(ids) != 2 || ids[1] != "t2" {
				t.Fatalf("ids = %v, want the next card to be t2 (no id consumed by refusals)", ids)
			}
		})
	}
}

func TestTodoAddRefGuard_PassesWhatItShould(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "stabilise the flaky restore test from PR #1943", "a dropped card about #2000", "a done-like card about #2001")
	setCardState(t, store, "t2", factory.BacklogStateDropped)
	for _, text := range []string{
		"a card about another number #1944",
		"a card that cites no number at all",
		"a card about #2000 whose only holder is dropped",
		"a card mentioning #19430 which is a different number",
	} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
	}
	if n := activeCards(t, store, "#2000"); n != 1 {
		t.Fatalf("active cards on #2000 = %d, want the new one only (dropped is excluded from the lookup)", n)
	}
}

func TestTodoAddRefGuard_IndependentScopeAdmitsAndRecordsTheReason(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "stabilise the flaky restore test from PR #1943")

	for _, reason := range []string{"", "   "} {
		before := queueDigest(t, store)
		if _, _, err := runTodo(t, "add", "--independent-scope", reason, "docs follow-up for #1943"); err == nil {
			t.Fatalf("empty reason %q admitted", reason)
		}
		if queueDigest(t, store) != before {
			t.Fatal("an empty reason changed the queue")
		}
	}
	if _, _, err := runTodo(t, "add", "--independent-scope", "docs only, no code overlap with t1", "docs follow-up for #1943"); err != nil {
		t.Fatalf("independent scope add: %v", err)
	}
	var found *factory.BacklogFinding
	fs := loadFindings(t, store)
	for i := range fs {
		if fs[i].Relation == factory.BacklogRelationIndependentScope {
			found = &fs[i]
		}
	}
	if found == nil || found.SubjectID != "t2" || found.RelatedID != "t1" || found.Note != "docs only, no code overlap with t1" {
		t.Fatalf("independent-scope finding = %+v, want {subject t2, related t1, the reason as note}", found)
	}
}

func TestTodoAddAttach_RecordsOneFindingAndChangesNoCard(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "stabilise the flaky restore test from PR #1943")
	items := snapshotItems(t, store)

	out, _, err := runTodo(t, "add", "--attach", "t1", "review comment about #1943: the retry is unbounded")
	if err != nil || !strings.HasPrefix(strings.TrimSpace(out), "attached t1") {
		t.Fatalf("attach: out=%q err=%v", out, err)
	}
	assertItemsUnchanged(t, store, items, "after attach")
	if ids := itemIDs(t, store); len(ids) != 1 {
		t.Fatalf("attach created a card: %v", ids)
	}
	out2, _, err := runTodo(t, "add", "--attach", "t1", "a second, different note")
	if err != nil || !strings.HasPrefix(strings.TrimSpace(out2), "already-attached t1") {
		t.Fatalf("repeat attach: out=%q err=%v", out2, err)
	}
	fs := loadFindings(t, store)
	if len(fs) != 1 || fs[0].Relation != factory.BacklogRelationAttachedText || fs[0].SubjectID != "t1" ||
		fs[0].Note != "review comment about #1943: the retry is unbounded" {
		t.Fatalf("findings = %+v, want exactly the first attached note", fs)
	}
	assertItemsUnchanged(t, store, items, "after the repeat")
}

func TestTodoAddAttach_Refusals(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "active card about #1943", "dropped card about #1944")
	setCardState(t, store, "t2", factory.BacklogStateDropped)
	before := queueDigest(t, store)
	for _, args := range [][]string{
		{"add", "--attach", "t9", "note"},
		{"add", "--attach", "t2", "note to a dropped card"},
		{"add", "--attach", "", "note"},
		{"add", "--attach", "t1", "--pick", "note"},
		{"add", "--attach", "t1", "--force", "note"},
		{"add", "--attach", "t1", "--independent-scope", "x", "note"},
	} {
		if _, _, err := runTodo(t, args...); err == nil {
			t.Errorf("%v was accepted, want a refusal", args)
		}
	}
	if queueDigest(t, store) != before {
		t.Fatal("a refused attach changed the queue")
	}
}

func TestTodoAddDryRun_ReportsTheLookupAndWritesNothing(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "stabilise the flaky restore test from PR #1943")
	before := queueDigest(t, store)
	_, stderr, err := runTodo(t, "add", "--dry-run", "wording about #1943")
	if err != nil || !strings.Contains(stderr, "a real add would refuse") || !strings.Contains(stderr, "t1 (queued)") {
		t.Fatalf("dry-run: err=%v stderr=%q", err, stderr)
	}
	_, stderr, err = runTodo(t, "add", "--dry-run", "--attach", "t1", "a note")
	if err != nil || !strings.Contains(stderr, "a real --attach would record one finding on t1") {
		t.Fatalf("dry-run attach: err=%v stderr=%q", err, stderr)
	}
	_, stderr, err = runTodo(t, "add", "--dry-run", "--independent-scope", "separate", "wording about #1943")
	if err != nil || !strings.Contains(stderr, "would admit the card on --independent-scope") {
		t.Fatalf("dry-run scope: err=%v stderr=%q", err, stderr)
	}
	if queueDigest(t, store) != before {
		t.Fatal("--dry-run changed the queue")
	}
}

func TestTodoAddRefGuard_ConcurrentAddsLeaveOneActiveCard(t *testing.T) {
	root, store := todoFixture(t)
	const workers = 8
	var wg sync.WaitGroup
	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = addAtRoot(root, fmt.Sprintf("worker %d wording about #77 and its fallout", i))
		}()
	}
	wg.Wait()
	admitted := 0
	for _, e := range errs {
		if e == nil {
			admitted++
		}
	}
	if admitted != 1 || activeCards(t, store, "#77") != 1 {
		t.Fatalf("admitted = %d, active cards on #77 = %d, want exactly 1 of %d concurrent adds", admitted, activeCards(t, store, "#77"), workers)
	}
}

// The MCP todo_add tool runs the same lookup, attach and scope paths.
func TestTodoAddMCP_SameLookupAsTheCLI(t *testing.T) {
	root, _ := sdMoaiFixture(t)
	t.Setenv(config.EnvFactoryRole, "operator")
	call := func(args map[string]any) (string, error) {
		args["project_root"] = root
		return sdCallTool(t, handleTodoAdd, args)
	}
	if _, err := call(map[string]any{"text": "stabilise the flaky restore test from PR #1943"}); err != nil {
		t.Fatal(err)
	}
	if _, err := call(map[string]any{"text": "wording about #1943"}); err == nil {
		t.Fatal("mcp todo_add admitted a shared #N")
	}
	if _, err := call(map[string]any{"text": "wording about #1943", "independent_scope": " "}); err == nil {
		t.Fatal("mcp todo_add admitted an empty independent_scope")
	}
	if out, err := call(map[string]any{"text": "note", "attach": "t1"}); err != nil || !strings.Contains(out, "attached t1") {
		t.Fatalf("mcp attach: out=%q err=%v", out, err)
	}
	if _, err := call(map[string]any{"text": "separate docs card about #1943", "independent_scope": "docs only"}); err != nil {
		t.Fatalf("mcp independent_scope: %v", err)
	}
}

// The standard codemaps-debt issuing path calls appendAnalyzedCard, which now
// runs the lookup; its card text names no `#N`, so the lookup finds nothing
// to refuse and the path is unchanged.
func TestCodemapsDebtCardTextCarriesNoIssueReference(t *testing.T) {
	text, issue, reason := codemapsDebtCardText(graph.CheckResult{Layers: []graph.LayerReport{codemapsLayer(40, 40, "1593d26233255c3623d24727fe747ae95e4447da")}})
	if !issue {
		t.Fatalf("fixture must issue; declined with %q", reason)
	}
	if len(todoTextRefs(text)) != 0 {
		t.Fatalf("codemaps debt card text %q carries a #N, so the lookup could refuse its issuance", text)
	}
}
