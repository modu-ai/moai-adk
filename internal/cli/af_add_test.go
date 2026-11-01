package cli

// N59 (audit of 1c05083f71): no verb put a card into the ledger, so the card flow could be exercised only by tests that
// called the store. `moai af add <id>` is that verb: it hands the ledger's AddCard a READY card and adds no rule of its
// own (the ledger refuses a duplicate id, a class that is not A, B or C, a path that leaves the repository and a
// dependency that does not exist). It binds the repository like every other verb.

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afAddVerb runs one verb and fails the test when it is refused.
func afAddVerb(t *testing.T, args ...string) string {
	t.Helper()
	out, errOut, err := afRunBoth(args...)
	if err != nil {
		t.Fatalf("moai af %s: %v\nstdout %.300s\nstderr %.300s", strings.Join(args, " "), err, out, errOut)
	}
	return out
}

// A card that exists only because the verb added it goes claim, evidence, enqueue, land and DONE through the verbs, at
// the default ledger location. The test setup makes no store call for the card.
func TestAFAdd_ACardAddedByTheVerbRunsFromClaimToDone(t *testing.T) {
	afSkipWithoutSh(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	repo, c1 := afN38Other(t) // the project: the dedicated trunk and a commit on a branch off it
	t.Setenv("CLAUDE_PROJECT_DIR", repo)

	out := afAddVerb(t, "add", "t1", "--title", "the first card", "--class", "A", "--path", "x.txt", "--priority", "5")
	var added struct {
		CardID string `json:"card_id"`
		Stage  string `json:"stage"`
	}
	if err := json.Unmarshal([]byte(out), &added); err != nil || added.CardID != "t1" || added.Stage != string(core.StageReady) {
		t.Fatalf("add printed %q: %+v %v, want t1 in READY", out, added, err)
	}

	var claimed struct {
		CardID string `json:"card_id"`
		Epoch  int64  `json:"epoch"`
	}
	if err := json.Unmarshal([]byte(afAddVerb(t, "claim", "--session", "s-t1", "--worktree", repo)), &claimed); err != nil || claimed.CardID != "t1" {
		t.Fatalf("claim: %+v %v", claimed, err)
	}
	e := strconv.FormatInt(claimed.Epoch, 10)
	for _, kind := range []string{"green", "gate"} {
		afAddVerb(t, "evidence", "t1", "--epoch", e, "--kind", kind, "--subject-sha", c1, "--result", "pass")
	}
	afAddVerb(t, "enqueue", "t1", "--epoch", e, "--candidate", c1)
	if out := afAddVerb(t, "land", "t1", "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("land: %s", out)
	}
	afAddVerb(t, "move", "t1", "--epoch", e, "--stage", "MERGING", "--remote-integrated")
	afAddVerb(t, "move", "t1", "--epoch", e, "--stage", "DONE")

	db, err := afLedgerPath("")
	if err != nil {
		t.Fatal(err)
	}
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	cards, err := st.Cards(context.Background())
	if err != nil || len(cards) != 1 || cards[0].Position.Stage != core.StageDone {
		t.Fatalf("cards %+v %v, want t1 DONE", cards, err)
	}
	if cards[0].Title != "the first card" || cards[0].Spec.Class != "A" || cards[0].Priority != 5 || len(cards[0].Spec.Paths) != 1 || cards[0].Spec.Paths[0] != "x.txt" {
		t.Errorf("the card does not carry what add was given: %+v", cards[0])
	}
}

// The minimal call is an id: the title is the id, the class is C, the card is READY with nothing else. What the ledger
// refuses, the verb refuses with the ledger unchanged.
func TestAFAdd_TheMinimalCallAndTheRefusalsOfTheLedger(t *testing.T) {
	afSkipWithoutSh(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	repo, _ := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	afAddVerb(t, "add", "m1")
	db, err := afLedgerPath("")
	if err != nil {
		t.Fatal(err)
	}
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	cards, err := st.Cards(context.Background())
	if err != nil || len(cards) != 1 {
		t.Fatalf("cards %+v %v", cards, err)
	}
	if c := cards[0]; c.ID != "m1" || c.Title != "m1" || c.Spec.Class != "C" || c.Position.Stage != core.StageReady {
		t.Errorf("the minimal card: %+v", c)
	}
	for name, args := range map[string][]string{
		"a duplicate id":                    {"add", "m1"},
		"a class that is not A, B or C":     {"add", "m2", "--class", "Z"},
		"a path that leaves the repository": {"add", "m3", "--path", "../x"},
		"a dependency that does not exist":  {"add", "m4", "--depends", "nosuch"},
	} {
		before := afSnapshot(t, st)
		if out, _, err := afRunBoth(args...); err == nil {
			t.Errorf("%s: accepted: %.200s", name, out)
		}
		if after := afSnapshot(t, st); len(after.Cards) != len(before.Cards) {
			t.Errorf("%s: the refusal changed the cards: %d -> %d", name, len(before.Cards), len(after.Cards))
		}
	}
	// a card that depends on one that exists is accepted
	afAddVerb(t, "add", "m5", "--depends", "m1")
	// the verb binds the repository like every other verb: this ledger is bound to the project's, so the same ledger
	// file reached from another repository is refused (the default location would be another project's ledger)
	if got := afN49Bound(t, st); got != afN38GitDir(t, repo) {
		t.Fatalf("add did not bind the ledger to the project: %q", got)
	}
	other, _ := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", other)
	if _, _, err := afRunBoth("add", "m6", "--db", db); err == nil || !strings.Contains(err.Error(), "bound to") {
		t.Errorf("add from another repository: %v, want the ledger's repository refusal", err)
	}
}

func TestAFAdd_TheHelpSaysWhatTheVerbTakes(t *testing.T) {
	flat := afHelpOf(t, "add")
	for _, want := range []string{"--title", "--class", "--path", "--depends", "READY", "is bound to one repository"} {
		if !strings.Contains(flat, want) {
			t.Errorf("`moai af add --help` does not say %q", want)
		}
	}
}
