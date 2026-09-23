package codexadapter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// loosenings returns one line per row of table that loosens a deny or a
// needs_input decision: an outcome the host resolves as allow, or — on Codex —
// the empty no-opinion object (REQ-HPR-007).
func loosenings(table []Translation) []string {
	var out []string
	for _, row := range table {
		if row.Decision != DecisionDeny && row.Decision != DecisionNeedsInput {
			continue
		}
		label := string(row.Harness) + "/" + string(row.Event) + "/" + string(row.Decision) + " → " + string(row.Outcome)
		if HostResolvesAsAllow(row.Harness, row.Event, row.Outcome) {
			out = append(out, label+" is in the host-resolves-as-allow set")
		}
		if row.Harness == HarnessCodex && row.Outcome == OutcomeNoOpinion {
			out = append(out, label+" is the empty no-opinion object")
		}
	}
	return out
}

// TestDecisionTranslationNeverLoosens is AC-HPR-006 (SPEC-DUAL-HARNESS-HOOK-PARITY-001,
// design.md §D1): the translation table is data, so never-loosens is a table
// property, checked row by row and on the rendered bytes.
func TestDecisionTranslationNeverLoosens(t *testing.T) {
	table := TranslationTable()

	t.Run("table declares every event x harness x decision exactly once", func(t *testing.T) {
		seen := map[string]int{}
		for _, row := range table {
			seen[string(row.Harness)+"/"+string(row.Event)+"/"+string(row.Decision)]++
		}
		for _, h := range []Harness{HarnessClaude, HarnessCodex} {
			for _, ev := range DecisionBearingEvents() {
				for _, d := range Decisions() {
					key := string(h) + "/" + string(ev) + "/" + string(d)
					if seen[key] != 1 {
						t.Errorf("%s declared %d times, want 1", key, seen[key])
					}
				}
			}
		}
		want := 2 * len(DecisionBearingEvents()) * len(Decisions())
		if len(table) != want {
			t.Errorf("table has %d rows, want %d", len(table), want)
		}
	})

	t.Run("no deny or needs_input row loosens", func(t *testing.T) {
		for _, l := range loosenings(table) {
			t.Error(l)
		}
	})

	t.Run("rendered deny and needs_input bytes are never empty or allow", func(t *testing.T) {
		for _, row := range table {
			if row.Decision != DecisionDeny && row.Decision != DecisionNeedsInput {
				continue
			}
			label := string(row.Harness) + "/" + string(row.Event) + "/" + string(row.Decision)
			raw, err := Render(row.Event, row.Outcome, "approval required: run `moai` to grant it")
			if err != nil {
				t.Errorf("%s: render: %v", label, err)
				continue
			}
			if row.Harness == HarnessCodex && strings.TrimSpace(string(raw)) == "{}" {
				t.Errorf("%s rendered the empty object", label)
			}
			var v map[string]any
			if err := json.Unmarshal(raw, &v); err != nil {
				t.Errorf("%s: rendered invalid JSON %s", label, raw)
				continue
			}
			if strings.Contains(string(raw), `"allow"`) {
				t.Errorf("%s rendered an allow: %s", label, raw)
			}
		}
	})

	t.Run("a Codex needs_input conversion is recorded, never silent", func(t *testing.T) {
		for _, ev := range DecisionBearingEvents() {
			row, ok := Lookup(HarnessCodex, ev, DecisionNeedsInput)
			if !ok {
				t.Fatalf("no Codex needs_input row for %s", ev)
			}
			if row.Outcome != OutcomeDeny || !row.DiscardRecorded {
				t.Errorf("Codex %s needs_input → %s (discard recorded %v), want deny with a discard record", ev, row.Outcome, row.DiscardRecorded)
			}
		}
	})

	t.Run("a deny render without a reason is refused", func(t *testing.T) {
		for _, ev := range DecisionBearingEvents() {
			if _, err := Render(ev, OutcomeDeny, "  "); err == nil {
				t.Errorf("%s: blank-reason deny rendered; Codex rejects it and it would become a no-op", ev)
			}
		}
	})

	t.Run("the checker names a restored ask drop", func(t *testing.T) {
		// Mutation (acceptance.md §B rule 6, applied in-test): restore the
		// card-t590 ask → {} drop on the Codex PreToolUse needs_input row.
		mutated := append([]Translation(nil), table...)
		hit := false
		for i := range mutated {
			r := mutated[i]
			if r.Harness == HarnessCodex && r.Event == hook.EventPreToolUse && r.Decision == DecisionNeedsInput {
				mutated[i].Outcome = OutcomeNoOpinion
				hit = true
			}
		}
		if !hit {
			t.Fatal("mutation target row absent; the mutation would be vacuous")
		}
		got := loosenings(mutated)
		if len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), "codex/PreToolUse/needs_input") {
			t.Fatalf("restored drop not detected: %v", got)
		}
	})
}
