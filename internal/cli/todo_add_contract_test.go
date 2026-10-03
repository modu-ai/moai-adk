package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestTodoAddReturnedIDAddressesALiveRow — GitHub #1732 (card t1313): every
// id the `todo add` response prints must address a LIVE queued row in the
// same store. The near-duplicate adjudication (mechanical analyser or the
// Consumer C Jev probe) records a finding beside the admitted card; it must
// never leave a printed id without the row it names, because the response is
// the only receipt the issuing session has — a ticket written against a
// printed id dies when the row does not exist.
//
// The probe is stubbed to reproduce the reported observation exactly (jev
// p=0.94 naming an older card), and the seeded texts walk the mechanical
// analyser into its own near-duplicate branch, so both adjudication sources
// are live while the contract is judged.
func TestTodoAddReturnedIDAddressesALiveRow(t *testing.T) {
	_, store := todoFixture(t)
	installJevProbe(t, alwaysNearDuplicateOf("t1", 0.94))

	texts := []string{
		"alpha one",                  // seed the similar row the probe will name
		"alpha one revised",          // mechanical near-duplicate branch
		"please alpha one now",       // near-duplicate + jev judgment at 0.94
		"alpha one with extra words", // third near-duplicate
	}
	var printed []string
	for _, text := range texts {
		out, _, err := runTodo(t, "add", text)
		if err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
		id, _, ok := strings.Cut(strings.TrimSpace(out), " ")
		if !ok || !strings.HasPrefix(id, "t") {
			t.Fatalf("add %q: response %q carries no issued id", text, out)
		}
		printed = append(printed, id)
	}

	// The contract, judged on the store the responses were issued from.
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load record: %v", err)
	}
	live := map[string]bool{}
	for _, it := range rec.Items {
		if it.State == factory.BacklogStateQueued {
			live[it.ID] = true
		}
	}
	for _, id := range printed {
		if !live[id] {
			t.Errorf("add printed %s but no live queued row carries it — the response and the queue state diverge (GitHub #1732)", id)
		}
	}
}
