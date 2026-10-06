// todo_merge_test.go — acceptance tests for the operator merge verb
// (SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-019, AC-TCI-018): it records and
// drops, refuses picked/closed/merged/cycle with the queue byte-identical,
// is refused in a lane, and is called by nobody in the analysis layer.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// mergedIntoFinding returns the merged-into finding between the two ids, or
// nil.
func mergedIntoFinding(findings []factory.BacklogFinding, from, into string) *factory.BacklogFinding {
	for i := range findings {
		f := &findings[i]
		if f.Relation == string(factory.CardRelationMergedInto) && f.SubjectID == from && f.RelatedID == into {
			return f
		}
	}
	return nil
}

// loadItem reads one card straight from the store.
func loadItem(t *testing.T, store *factory.BacklogStore, id string) factory.BacklogItem {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == id {
			return it
		}
	}
	t.Fatalf("no card %s in the fixture", id)
	return factory.BacklogItem{}
}

// pickItem moves a card to picked without walking the CLI.
func pickItem(t *testing.T, store *factory.BacklogStore, id string) {
	t.Helper()
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				rec.Items[i].State = factory.BacklogStatePicked
				return nil
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("pick %s: %v", id, err)
	}
}

// TestTodoMergeRecordsAndDrops — AC-TCI-018 (a): the fold, the drop, and the
// relation land in one write.
func TestTodoMergeRecordsAndDrops(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha task body", "Beta detail body", "Gamma body")

	out, _, err := runTodo(t, "merge", "t1", "t2")
	if err != nil {
		t.Fatalf("merge t1 t2: %v", err)
	}
	if !strings.Contains(out, "merged t2 into t1") {
		t.Fatalf("merge output = %q, want the merged line", out)
	}

	into := loadItem(t, store, "t1")
	want := "Alpha task body\n\n[merged from t2] Beta detail body"
	if into.Text != want {
		t.Fatalf("t1 text = %q, want %q", into.Text, want)
	}
	if into.State != factory.BacklogStateQueued {
		t.Fatalf("t1 state = %s, want queued (the merge moves nothing else)", into.State)
	}

	from := loadItem(t, store, "t2")
	if from.State != factory.BacklogStateDropped {
		t.Fatalf("t2 state = %s, want dropped", from.State)
	}
	if from.DroppedAt == nil || *from.DroppedAt == "" {
		t.Fatalf("t2 carries no dropped stamp: %+v", from.DroppedAt)
	}
	if !strings.HasPrefix(from.Text, todoDropMarkerOpen+"merged into t1"+todoDropMarkerClose) {
		t.Fatalf("t2 text = %q, want the dropped marker with reason merged into t1", from.Text)
	}
	if from.Issuance == nil || from.Issuance.DropReason != "merged into t1" {
		t.Fatalf("t2 drop-reason attribute = %+v, want merged into t1", from.Issuance)
	}

	findings := loadFindings(t, store)
	if f := mergedIntoFinding(findings, "t2", "t1"); f == nil {
		t.Fatalf("no merged-into t2→t1 finding recorded: %+v", findings)
	} else if f.Source != factory.BacklogSourceAgent {
		t.Fatalf("merged-into finding source = %q, want agent", f.Source)
	}
}

// TestTodoMergeRefusesPickedAndCycles — AC-TCI-018 (b): picked either side,
// an already-merged from, and a closed or already-merged into are refused,
// each leaving the queue byte-identical.
func TestTodoMergeRefusesPickedAndCycles(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha body", "Beta body", "Gamma body", "Delta body")
	pickItem(t, store, "t3")

	// picked <from>
	snapshot := snapshotItems(t, store)
	if _, _, err := runTodo(t, "merge", "t1", "t3"); err == nil {
		t.Fatal("merge into from a picked card was accepted")
	}
	assertItemsUnchanged(t, store, snapshot, "after picked-from refusal")

	// picked <into>
	if _, _, err := runTodo(t, "merge", "t3", "t1"); err == nil {
		t.Fatal("merge into a picked card was accepted")
	}
	assertItemsUnchanged(t, store, snapshot, "after picked-into refusal")

	// a real merge, then the already-merged refusals both ways
	if _, _, err := runTodo(t, "merge", "t1", "t2"); err != nil {
		t.Fatalf("merge t1 t2: %v", err)
	}
	snapshot = snapshotItems(t, store)
	// from-already-merged: merge t4 t2 (t2 carries the outgoing merged-into).
	if _, _, err := runTodo(t, "merge", "t4", "t2"); err == nil {
		t.Fatal("merge from an already-merged card was accepted")
	}
	// into-already-merged: merge t2 t4.
	if _, _, err := runTodo(t, "merge", "t2", "t4"); err == nil {
		t.Fatal("merge into an already-merged card was accepted")
	}
	assertItemsUnchanged(t, store, snapshot, "after already-merged refusals")
	// The refusal is the merged-into kind, not luck: t2 carries the outgoing
	// merged-into finding from the first merge.
	if f := mergedIntoFinding(loadFindings(t, store), "t2", "t1"); f == nil {
		t.Fatal("the first merge left no merged-into finding")
	}

	// closed <into>: a dropped card
	if _, _, err := runTodo(t, "drop", "t4", "discard me"); err != nil {
		t.Fatalf("drop t4: %v", err)
	}
	snapshot = snapshotItems(t, store)
	if _, _, err := runTodo(t, "merge", "t4", "t3"); err == nil {
		t.Fatal("merge into a dropped card was accepted")
	}
	assertItemsUnchanged(t, store, snapshot, "after closed-into refusal")

	// nonexistent id
	if _, _, err := runTodo(t, "merge", "t999", "t1"); err == nil {
		t.Fatal("merge from a nonexistent card was accepted")
	}
	assertItemsUnchanged(t, store, snapshot, "after nonexistent-id refusal")
}

// TestTodoMergeNeverInvokedByAnalysis — AC-TCI-018 (d): the merge entry
// function is called from no production file outside todo_merge.go — the
// analyser, analyze, relate, and add keep their never-folds shape by code
// shape, not by comment.
func TestTodoMergeNeverInvokedByAnalysis(t *testing.T) {
	// Tests run with the package directory as cwd: internal/cli is here, the
	// factory package one directory up. todo_merge.go itself is the declaring
	// file and is skipped — every OTHER production file must carry zero
	// references to the entry symbol.
	roots := []string{".", filepath.Join("..", "factory")}
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			if filepath.Base(path) == "todo_merge.go" {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if count := strings.Count(string(raw), "runTodoMerge"); count > 0 {
				t.Errorf("%s references runTodoMerge %d times — only todo_merge.go may call it", path, count)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
}

// TestTodoMergeRefusedInLane — AC-TCI-018 (c): a lane session's merge is
// refused by the standing queue guard, with the queue file byte-identical.
func TestTodoMergeRefusedInLane(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha body", "Beta body")
	snapshot := snapshotItems(t, store)

	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	_, errOut, err := runTodo(t, "merge", "t1", "t2")
	if err == nil {
		t.Fatal("a lane session's merge was accepted")
	}
	if !strings.Contains(errOut, "lane boundary") || !strings.Contains(errOut, "cannot mutate the queue") {
		t.Fatalf("lane refusal text = %q, want the lane-boundary queue-mutation wording", errOut)
	}
	assertItemsUnchanged(t, store, snapshot, "after lane refusal")
}
