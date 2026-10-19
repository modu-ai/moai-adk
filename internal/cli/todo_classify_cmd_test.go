// todo_classify_cmd_test.go — SPEC-FACTORY-DEFAULT-MODE-001 M2 acceptance
// tests on the `moai todo classify` verb: the one-locked-write human judgment
// on a queued (AC-FDM-005) and a picked (AC-FDM-006) card, the refusal matrix
// with byte-identical writes (AC-FDM-007), and the serial-slot conflict
// refusal on an in-flight serial reclassification (AC-FDM-011).
package cli

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// fcSeedCard adds one card and hands back the fixture root and store, so a
// refusal-matrix seed starts from a known single-card queue.
func fcSeedCard(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "classify target"); err != nil {
		t.Fatalf("add: %v", err)
	}
	return root, store
}

func TestTodoClassifyQueuedCard(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "queued target"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// A prior judgment whose priority/blocked the verb must leave untouched.
	fcClassify(t, store, "t1", factory.ClassPriorityHigh, true, factory.ClassModeParallelizable)
	if _, _, err := runTodo(t, "add", "bystander card"); err != nil {
		t.Fatalf("add bystander: %v", err)
	}
	before, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}

	out, _, err := runTodo(t, "classify", "t1", "--mode", "serial", "--reason", "touches the release pipeline")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if n := strings.Count(out, "classified t1 mode=serial"); n != 1 {
		t.Errorf("stdout = %q, want exactly one confirmation line", out)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	beforeByID := make(map[string]factory.BacklogItem, len(before.Items))
	for _, it := range before.Items {
		beforeByID[it.ID] = it
	}
	var target, bystander *factory.BacklogItem
	for i := range rec.Items {
		switch rec.Items[i].ID {
		case "t1":
			target = &rec.Items[i]
		case "t2":
			bystander = &rec.Items[i]
		}
	}
	if target == nil || bystander == nil {
		t.Fatalf("queue lost items: %+v", rec.Items)
	}
	c := target.Classification
	if c == nil {
		t.Fatalf("classified card carries no classification")
	}
	if c.Mode != factory.ClassModeSerial || c.Decider != factory.DeciderIdentityHuman {
		t.Errorf("classification = %+v, want mode serial under decider human", c)
	}
	if c.Reason != "touches the release pipeline" {
		t.Errorf("reason = %q, want the given reason verbatim", c.Reason)
	}
	if c.ClassifiedAt == "" {
		t.Fatalf("classification carries no classified-at stamp")
	}
	if _, err := time.Parse(time.RFC3339, c.ClassifiedAt); err != nil {
		t.Errorf("classified-at %q is not an RFC 3339 stamp: %v", c.ClassifiedAt, err)
	}
	if c.Priority != factory.ClassPriorityHigh || !c.Blocked {
		t.Errorf("priority/blocked = %q/%v, want the prior judgment's high/true untouched", c.Priority, c.Blocked)
	}
	// Every other item is verbatim-unchanged: the add path legitimately gave
	// the bystander its own default judgment, and the verb must not have
	// touched it.
	if beforeB := beforeByID["t2"]; !reflect.DeepEqual(*bystander, beforeB) {
		t.Errorf("bystander card changed:\n got %+v\nwant %+v", *bystander, beforeB)
	}
}

// TestTodoClassifyPickedCard — AC-FDM-006: a picked card (leased, in the
// factory record — the shape of a fail-safe-admitted card already in flight)
// records the same one-locked-write judgment with --mode parallelizable. This
// is the fallback-path closure check: the leader's correction path for a
// fail-safe admission is the verb, not a database edit.
func TestTodoClassifyPickedCard(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked, OwnerLabel: "lane-2"})

	out, _, err := runTodo(t, "classify", "t1", "--mode", "parallelizable", "--reason", "fail-safe admission was wrong; the card touches only its own tree")
	if err != nil {
		t.Fatalf("classify picked: %v", err)
	}
	if n := strings.Count(out, "classified t1 mode=parallelizable"); n != 1 {
		t.Errorf("stdout = %q, want exactly one confirmation line", out)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	c := rec.Items[0].Classification
	if c == nil {
		t.Fatalf("classified card carries no classification")
	}
	if c.Mode != factory.ClassModeParallelizable || c.Decider != factory.DeciderIdentityHuman || c.Reason != "fail-safe admission was wrong; the card touches only its own tree" {
		t.Errorf("classification = %+v, want parallelizable under decider human with the given reason verbatim", c)
	}
	if rec.Items[0].State != factory.BacklogStatePicked {
		t.Errorf("classify moved the picked card to %s, want the state untouched", rec.Items[0].State)
	}
}

func TestTodoClassifyRefusals(t *testing.T) {
	archiveSeed := func(t *testing.T, store *factory.BacklogStore) {
		t.Helper()
		if err := store.Mutate(func(rec *factory.BacklogRecord) error {
			rec.Archived = append(rec.Archived, factory.BacklogArchiveEntry{Item: rec.Items[0]})
			rec.Items = rec.Items[1:]
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	stateSeed := func(state factory.BacklogState) func(t *testing.T, store *factory.BacklogStore) {
		return func(t *testing.T, store *factory.BacklogStore) {
			t.Helper()
			if err := store.Mutate(func(rec *factory.BacklogRecord) error {
				rec.Items[0].State = state
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}
	}
	noSeed := func(t *testing.T, store *factory.BacklogStore) {}

	cases := []struct {
		name     string
		seed     func(t *testing.T, store *factory.BacklogStore)
		args     []string
		wantExit int
	}{
		{"held card", stateSeed(factory.BacklogStateHold), []string{"classify", "t1", "--mode", "serial", "--reason", "x"}, 2},
		{"dropped card", stateSeed(factory.BacklogStateDropped), []string{"classify", "t1", "--mode", "serial", "--reason", "x"}, 2},
		{"terminal card (archived, id no longer an item)", archiveSeed, []string{"classify", "t1", "--mode", "serial", "--reason", "x"}, 2},
		{"id absent from the queue", noSeed, []string{"classify", "t404", "--mode", "serial", "--reason", "x"}, 2},
		{"empty reason", noSeed, []string{"classify", "t1", "--mode", "serial", "--reason", ""}, 2},
		{"mode outside the closed set", noSeed, []string{"classify", "t1", "--mode", "conflicting", "--reason", "x"}, 2},
		{"no decider flag exists", noSeed, []string{"classify", "t1", "--mode", "serial", "--decider", "human", "--reason", "x"}, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, store := fcSeedCard(t)
			tc.seed(t, store)
			before := readBacklogBytes(t, root)
			_, _, err := runTodo(t, tc.args...)
			if err == nil {
				t.Fatalf("classify %v refused nothing", tc.args)
			}
			if tc.wantExit >= 0 {
				var code *exitCodeError
				if !errors.As(err, &code) || code.ExitCode() != tc.wantExit {
					t.Fatalf("refusal error = %v, want a usage exit-%d refusal", err, tc.wantExit)
				}
				if !strings.Contains(err.Error(), "nothing written") {
					t.Errorf("refusal %q does not name that nothing was written", err)
				}
			}
			if after := readBacklogBytes(t, root); string(after) != string(before) {
				t.Fatalf("refused classify changed the queue file")
			}
		})
	}
}

func TestTodoClassifySerialSlotConflictRefusal(t *testing.T) {
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	conflictFixture := func(t *testing.T) (string, *factory.BacklogStore) {
		t.Helper()
		root, store := fcFixture(t)
		// t1: the picked target, in the factory record with a live lease.
		// t2: another DISTINCT in-flight serial card holding the slot.
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStateQueued)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
		fcPlace(t, root,
			homestate.Card{CardID: "t1", State: homestate.CardPicked, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live, Stage: homestate.CardRun},
			homestate.Card{CardID: "t2", State: homestate.CardRun, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: live, Stage: homestate.CardRun},
		)
		return root, store
	}

	t.Run("serial refused while another distinct in-flight card holds the slot", func(t *testing.T) {
		root, store := conflictFixture(t)
		before := readBacklogBytes(t, root)
		_, _, err := runTodo(t, "classify", "t1", "--mode", "serial", "--reason", "make it serial")
		var code *exitCodeError
		if !errors.As(err, &code) || code.ExitCode() != 2 {
			t.Fatalf("conflict error = %v, want a usage exit-2 refusal", err)
		}
		if !strings.Contains(err.Error(), "t2") {
			t.Errorf("refusal %q does not name the slot holder", err)
		}
		for _, want := range []string{"wait", "reclassify"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal %q is missing recovery advice naming %q", err, want)
			}
		}
		if after := readBacklogBytes(t, root); string(after) != string(before) {
			t.Fatalf("refused conflict classify changed the queue file")
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if c := rec.Items[0].Classification; c == nil || c.Mode != factory.ClassModeParallelizable {
			t.Errorf("target classification = %+v, want the prior parallelizable record untouched", rec.Items[0].Classification)
		}
	})

	t.Run("same target with parallelizable succeeds beside the holder", func(t *testing.T) {
		_, store := conflictFixture(t)
		out, _, err := runTodo(t, "classify", "t1", "--mode", "parallelizable", "--reason", "keep it parallel")
		if err != nil {
			t.Fatalf("classify parallelizable beside the holder: %v", err)
		}
		if !strings.Contains(out, "classified t1 mode=parallelizable") {
			t.Errorf("stdout = %q, want the confirmation line", out)
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if c := rec.Items[0].Classification; c == nil || c.Mode != factory.ClassModeParallelizable || c.Decider != factory.DeciderIdentityHuman {
			t.Errorf("classification = %+v, want the human judgment recorded", c)
		}
	})

	t.Run("serial succeeds when no other serial card is in flight", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStateQueued)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		fcPlace(t, root,
			homestate.Card{CardID: "t1", State: homestate.CardPicked, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live, Stage: homestate.CardRun},
			homestate.Card{CardID: "t2", State: homestate.CardRun, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: live, Stage: homestate.CardRun},
		)
		out, _, err := runTodo(t, "classify", "t1", "--mode", "serial", "--reason", "the slot is free")
		if err != nil {
			t.Fatalf("classify serial with the slot free: %v", err)
		}
		if !strings.Contains(out, "classified t1 mode=serial") {
			t.Errorf("stdout = %q, want the confirmation line", out)
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if c := rec.Items[0].Classification; c == nil || c.Mode != factory.ClassModeSerial || c.Decider != factory.DeciderIdentityHuman {
			t.Errorf("classification = %+v, want the serial human judgment recorded", c)
		}
	})
}
