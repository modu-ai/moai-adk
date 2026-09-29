// todo_classify_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M2 acceptance
// tests on the `moai todo add` path: classification recorded through the
// decider seam inside the locked write (AC-TCD-001), the fail-safe fallback
// with its one-line stderr notice (AC-TCD-003), the --classification-file
// validated input and its refusals (AC-TCD-004), the sorted queue order and
// the sorted position output (AC-TCD-005/-006).
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// installTodoCardDecider replaces the add path's classification seam for the
// test's duration.
func installTodoCardDecider(t *testing.T, dec kanban.CardDecider) {
	t.Helper()
	old := todoCardDecider
	todoCardDecider = dec
	t.Cleanup(func() { todoCardDecider = old })
}

// failingCardDecider is the AC-TCD-003 mutation: a decider that is
// unavailable.
type failingCardDecider struct{}

func (failingCardDecider) Classify(string) (kanban.CardClassification, error) {
	return kanban.CardClassification{}, errors.New("decider unavailable (injected)")
}

// judgementCardDecider is the healthy positive control: it returns a real
// judgment.
func judgementCardDecider(prio, mode string, blocked bool) kanban.CardDecider {
	return kanban.StaticCardDecider{Class: kanban.CardClassification{
		Priority: prio, Blocked: blocked, Mode: mode, Decider: kanban.DeciderIdentityLLM,
		Reason: "injected healthy judgment",
	}}
}

func queueOrder(t *testing.T, store *kanban.BacklogStore) []string {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var ids []string
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStateQueued {
			ids = append(ids, it.ID)
		}
	}
	return ids
}

// TestTodoAddRecordsClassificationInLockedWrite — AC-TCD-001: `todo add`
// classifies the card through the decider seam inside the same locked write
// that appends it, and the --json queue read carries every field.
func TestTodoAddRecordsClassificationInLockedWrite(t *testing.T) {
	_, store := todoFixture(t)
	installTodoCardDecider(t, judgementCardDecider(kanban.ClassPriorityHigh, kanban.ClassModeParallelizable, false))

	out, _, err := runTodo(t, "add", "classified card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])

	out, _, err = runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !strings.Contains(out, `"classification"`) {
		t.Errorf("--json queue read carries no classification key; add output was %q", out)
	}
	for _, it := range rec.Items {
		if it.ID != id {
			continue
		}
		c := it.Classification
		if c == nil {
			t.Fatalf("card %s carries no classification", id)
		}
		if c.Priority != kanban.ClassPriorityHigh || c.Blocked || c.Mode != kanban.ClassModeParallelizable || c.Decider != kanban.DeciderIdentityLLM {
			t.Errorf("card %s classification = %+v, want the decider's judgment", id, c)
		}
		if c.ClassifiedAt == "" {
			t.Errorf("card %s classification carries no classified_at stamp", id)
		}
	}
}

// TestTodoAddDeciderFailureFallsBackWithNotice — AC-TCD-003: a decider that
// is unavailable promotes the fail-safe default (normal / false / serial /
// default) with EXACTLY ONE stderr notice line, and admission never blocks.
// The positive control (healthy decider records its real values) lives in
// TestTodoAddRecordsClassificationInLockedWrite; the default-vs-failure
// separation is asserted here by the notice firing only on the failure arm.
func TestTodoAddDeciderFailureFallsBackWithNotice(t *testing.T) {
	_, store := todoFixture(t)
	installTodoCardDecider(t, failingCardDecider{})

	out, errOut, err := runTodo(t, "add", "unclassified fallback card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	if !strings.Contains(errOut, "decider unavailable") {
		t.Errorf("stderr notice %q does not name the fallback cause", errOut)
	}
	// Count the FALLBACK lines, not the whole stderr stream: the fixture
	// environment legitimately adds its own disclosure lines (the temp-root
	// queue notice), which are not part of the AC.
	if n := strings.Count(errOut, "classification decider unavailable"); n != 1 {
		t.Errorf("stderr carries %d fallback notices, want exactly 1 (got %q)", n, errOut)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID != id {
			continue
		}
		c := it.Classification
		if c == nil {
			t.Fatalf("fallback card %s carries no classification", id)
		}
		want := kanban.DefaultCardClassification()
		if c.Priority != want.Priority || c.Blocked != want.Blocked || c.Mode != want.Mode || c.Decider != want.Decider {
			t.Errorf("fallback classification = %+v, want the fail-safe defaults %+v", c, want)
		}
	}
}

// TestTodoAddDefaultDeciderPrintsNoNotice — the deterministic default is a
// judgment, not a failure: the plain add path records decider=default and
// prints no fallback notice.
func TestTodoAddDefaultDeciderPrintsNoNotice(t *testing.T) {
	_, store := todoFixture(t)
	installTodoCardDecider(t, kanban.DefaultCardDecider{})

	out, errOut, err := runTodo(t, "add", "plain card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if strings.Contains(errOut, "classification decider unavailable") {
		t.Errorf("plain add printed a fallback notice on stderr %q, want none (the deterministic default is a judgment, not a failure)", errOut)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID == id && (it.Classification == nil || it.Classification.Decider != kanban.DeciderIdentityDefault) {
			t.Errorf("plain add classification = %+v, want decider=default", it.Classification)
		}
	}
}

// writeClassificationFile writes one classification JSON fixture.
func writeClassificationFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "classification.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestTodoAddClassificationFileValidatesAndRecords — AC-TCD-004: a valid
// supplied classification is recorded verbatim with its decider identity;
// out-of-set values and the jev identity are usage refusals (exit 2) with
// nothing written.
func TestTodoAddClassificationFileValidatesAndRecords(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		_, store := todoFixture(t)
		path := writeClassificationFile(t, `{"priority":"low","blocked":true,"mode":"parallelizable","decider":"human","reason":"operator said so"}`)
		if _, _, err := runTodo(t, "add", "supplied card", "--classification-file", path); err != nil {
			t.Fatalf("add: %v", err)
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		c := rec.Items[0].Classification
		if c == nil || c.Priority != kanban.ClassPriorityLow || !c.Blocked || c.Mode != kanban.ClassModeParallelizable || c.Decider != kanban.DeciderIdentityHuman {
			t.Errorf("supplied classification = %+v, want the supplied judgment recorded", c)
		}
	})
	for _, tc := range []struct {
		name string
		body string
	}{
		{"priority out of set", `{"priority":"urgent","mode":"serial","decider":"llm"}`},
		{"mode out of set", `{"priority":"normal","mode":"serially","decider":"llm"}`},
		{"decider jev", `{"priority":"normal","mode":"serial","decider":"jev"}`},
		{"malformed json", `{"priority":"normal"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, store := todoFixture(t)
			path := writeClassificationFile(t, tc.body)
			_, _, err := runTodo(t, "add", "refused card", "--classification-file", path)
			var code *exitCodeError
			if !errors.As(err, &code) || code.ExitCode() != 2 {
				t.Fatalf("refusal error = %v, want a usage exit-2 refusal", err)
			}
			rec, err := store.LoadPure()
			if err != nil {
				t.Fatal(err)
			}
			if len(rec.Items) != 0 {
				t.Errorf("refused add wrote %d cards, want nothing written", len(rec.Items))
			}
		})
	}
}

// TestTodoAddClassificationFileUnavailableFallsBack — REQ-TCD-003's
// transport-unavailable arm: a --classification-file that cannot be read
// falls back to the defaults with the one-line notice instead of blocking
// admission.
func TestTodoAddClassificationFileUnavailableFallsBack(t *testing.T) {
	_, store := todoFixture(t)
	missing := filepath.Join(t.TempDir(), "absent.json")
	out, errOut, err := runTodo(t, "add", "fallback card", "--classification-file", missing)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(errOut, "decider unavailable") {
		t.Errorf("stderr %q does not carry the fallback notice", errOut)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Items) != 1 || rec.Items[0].ID != id || rec.Items[0].Classification == nil ||
		rec.Items[0].Classification.Decider != kanban.DeciderIdentityDefault {
		t.Errorf("fallback add recorded %+v, want one card with the default classification", rec.Items)
	}
}

// TestTodoAddSortsQueueAndPrintsSortedPosition — AC-TCD-005 and AC-TCD-006:
// adds A(normal), B(high), C(low), D(normal, blocked) land the queue in
// B, A, C, D order, and the NEXT add's printed position is its sorted
// 1-based position — a high card added last prints position 3, not 5.
func TestTodoAddSortsQueueAndPrintsSortedPosition(t *testing.T) {
	_, store := todoFixture(t)
	installTodoCardDecider(t, failingCardDecider{}) // replaced per add below

	add := func(text string, dec kanban.CardDecider) string {
		t.Helper()
		installTodoCardDecider(t, dec)
		out, _, err := runTodo(t, "add", text)
		if err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
		return strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	}
	high := func(text string) string {
		return add(text, kanban.StaticCardDecider{Class: kanban.CardClassification{Priority: kanban.ClassPriorityHigh, Mode: kanban.ClassModeSerial, Decider: kanban.DeciderIdentityLLM}})
	}
	normal := func(text string) string {
		return add(text, kanban.StaticCardDecider{Class: kanban.CardClassification{Priority: kanban.ClassPriorityNormal, Mode: kanban.ClassModeSerial, Decider: kanban.DeciderIdentityLLM}})
	}
	low := func(text string) string {
		return add(text, kanban.StaticCardDecider{Class: kanban.CardClassification{Priority: kanban.ClassPriorityLow, Mode: kanban.ClassModeSerial, Decider: kanban.DeciderIdentityLLM}})
	}
	blockedNormal := func(text string) string {
		return add(text, kanban.StaticCardDecider{Class: kanban.CardClassification{Priority: kanban.ClassPriorityNormal, Blocked: true, Mode: kanban.ClassModeSerial, Decider: kanban.DeciderIdentityLLM}})
	}

	idA := normal("A normal")
	idB := high("B high")
	idC := low("C low")
	idD := blockedNormal("D blocked")

	got := queueOrder(t, store)
	want := []string{idB, idA, idC, idD}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("queue order = %v, want %v", got, want)
	}

	// AC-TCD-006: the next add prints its sorted position. E is high, B is
	// high — same rank, so E lands right after B in insertion order: position
	// 2 in the sorted order B, E, A, C, D (never the append index 5).
	installTodoCardDecider(t, kanban.StaticCardDecider{Class: kanban.CardClassification{Priority: kanban.ClassPriorityHigh, Mode: kanban.ClassModeSerial, Decider: kanban.DeciderIdentityLLM}})
	out, _, err := runTodo(t, "add", "E high")
	if err != nil {
		t.Fatalf("add E: %v", err)
	}
	idE := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	printed := strings.Fields(strings.TrimSpace(out))[1]
	if printed != "2" {
		t.Errorf("add printed position %s for a high card, want 2 (sorted order B, E, A, C, D)", printed)
	}
	if got := queueOrder(t, store); fmt.Sprint(got) != fmt.Sprint([]string{idB, idE, idA, idC, idD}) {
		t.Errorf("queue order after E = %v, want B E A C D", got)
	}

	// AC-TCD-006: the list render prints the same sorted order.
	listOut, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	prev := -1
	for _, id := range []string{idB, idE, idA, idC, idD} {
		idx := strings.Index(listOut, id+"\t")
		if idx < 0 {
			t.Fatalf("list output missing card %s:\n%s", id, listOut)
		}
		if idx < prev {
			t.Errorf("list render is not in sorted order at %s:\n%s", id, listOut)
		}
		prev = idx
	}
}

// TestTodoStoreAddKeepsQueueSorted — REQ-TCD-005 covers every add that may
// change the order: the store's own Add (the GTD publication path) re-sorts
// inside the same locked write, so an unclassified append to a sorted queue
// stays sorted (an unclassified card ranks normal, which sorts ahead of an
// existing low card).
func TestTodoStoreAddKeepsQueueSorted(t *testing.T) {
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(t.TempDir()))
	lowID := ""
	if _, _, err := store.Add("seed low"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			c := kanban.CardClassification{Priority: kanban.ClassPriorityLow, Mode: kanban.ClassModeSerial, Decider: kanban.DeciderIdentityLLM}
			rec.Items[i].Classification = &c
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	lowID = rec.Items[0].ID

	id, _, err := store.Add("unclassified normal append")
	if err != nil {
		t.Fatal(err)
	}
	got := queueOrder(t, store)
	if fmt.Sprint(got) != fmt.Sprint([]string{id.ID, lowID}) {
		t.Errorf("store.Add left the queue unsorted: %v, want [%s %s]", got, id.ID, lowID)
	}
}
