// factory_card_eligible_pull_test.go — SPEC-CARD-ELIGIBLE-PULL-001 M2 gate
// scenario tests (card t1839): the OFF dormancy of the generic selection
// arms (AC-CEP-001), the ON five-condition eligibility gate with one
// disqualification cell per token (AC-CEP-002), the rule-paired eligible
// lease / ineligible refusal fixture (AC-CEP-003), the traceless skip and
// refusal (AC-CEP-005), and the machine-readable refusal grammar on both
// surfaces (AC-CEP-006).
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (§B of acceptance.md); ./internal/cli runs only
// through the anchored -run selectors naming one of these tests.
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"gopkg.in/yaml.v3"
)

// cepSelfDispatchYAML sets workflow.factory.self_dispatch in the fixture's
// workflow.yaml, MERGING into any existing file (sync-audit round-3 F10): a
// plain overwrite silently disabled sibling keys other fixtures had already
// written — the quota fixtures' `quota_gate.enabled: true` (qasEnableGate) —
// killing those cells' quota subject while they stayed green. The merge is a
// generic map round-trip: every key the fixture (or a real deployment) wrote
// survives; only the factory block is set.
func cepSelfDispatchYAML(t *testing.T, root string, on bool) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	path := filepath.Join(dir, "workflow.yaml")
	var doc map[string]interface{}
	if raw, err := os.ReadFile(path); err == nil && len(raw) > 0 {
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("merge self_dispatch into %s: parse existing yaml: %v", path, err)
		}
	}
	if doc == nil {
		doc = map[string]interface{}{}
	}
	workflow, _ := doc["workflow"].(map[string]interface{})
	if workflow == nil {
		workflow = map[string]interface{}{}
		doc["workflow"] = workflow
	}
	factory, _ := workflow["factory"].(map[string]interface{})
	if factory == nil {
		factory = map[string]interface{}{}
		workflow["factory"] = factory
	}
	factory["self_dispatch"] = on
	out, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("merge self_dispatch into %s: encode: %v", path, err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// cepRefusedLines picks the `refused <token>: <detail>` lines out of a
// no-card response.
func cepRefusedLines(out string) []string {
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "refused ") {
			lines = append(lines, l)
		}
	}
	return lines
}

// cepAssertRefusal asserts out carries exactly one refused line with the
// token naming cardID.
func cepAssertRefusal(t *testing.T, out, token, cardID string) {
	t.Helper()
	var hits []string
	for _, l := range cepRefusedLines(out) {
		if strings.HasPrefix(l, "refused "+token+": ") {
			hits = append(hits, l)
		}
	}
	if len(hits) != 1 {
		t.Fatalf("refused %s lines for %s: %d (%v), want exactly 1", token, cardID, len(hits), cepRefusedLines(out))
	}
	if !strings.Contains(hits[0], cardID) {
		t.Errorf("refused %s line %q does not name %s", token, hits[0], cardID)
	}
}

// cepAssertNoRefusal asserts the quiet-skip surface: no refused line at all.
func cepAssertNoRefusal(t *testing.T, where, out string) {
	t.Helper()
	if lines := cepRefusedLines(out); len(lines) > 0 {
		t.Errorf("%s: refusal lines %v, want none", where, lines)
	}
}

// cepAssertRowFrozen asserts the candidate's row is byte-for-byte the state,
// owner and version it had before the pass.
func cepAssertRowFrozen(t *testing.T, root, cardID string, before homestate.Card) {
	t.Helper()
	after := fcCard(t, root, cardID)
	if after.State != before.State || after.OwnerLabel != before.OwnerLabel || after.Version != before.Version {
		t.Errorf("%s row changed across the skip: %s v%d owner=%q → %s v%d owner=%q",
			cardID, before.State, before.Version, before.OwnerLabel, after.State, after.Version, after.OwnerLabel)
	}
}

// cepMarkText rewrites one queue item's text (the hold-marker fixture).
func cepMarkText(t *testing.T, store *factory.BacklogStore, cardID, text string) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == cardID {
				r.Items[i].Text = text
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("mark %s text: %v", cardID, err)
	}
}

// cepNoCard runs bare `next` as lane-1 and asserts the no-card answer: exit
// 3 and the no-card line on stdout. It returns stdout for refusal asserts.
func cepNoCard(t *testing.T, root string) string {
	t.Helper()
	out, _, err := runFactory(t, "next", "--run", fcRun)
	sdExit3(t, "gate refusal", err)
	if !strings.Contains(out, "no card is available") {
		t.Fatalf("stdout = %q, want the no-card line", out)
	}
	return out
}

// AC-CEP-001 — flag OFF (default): the generic arms are dormant for unowned
// cards; the assigned path, own bundles, and queued promotion keep serving.
func TestFactoryEligiblePullOffDormancy(t *testing.T) {
	t.Run("unowned picked row is not leased; the queued sibling is promoted instead", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardPicked})
		before := fcCard(t, root, "t2")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out, _, err := runFactory(t, "next", "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
			t.Fatalf("t1 = %s holder=%q, want leased/lane-1 through arm (c) — the queued card is served with the generic arms dormant", c.State, c.LeaseHolder)
		}
		cepAssertRowFrozen(t, root, "t2", before)
		if got, want := fcQueueItemState(t, store, "t2"), factory.BacklogStatePicked; got != want {
			t.Errorf("t2 queue state = %s, want %s (no pick change, no trace)", got, want)
		}
		cepAssertNoRefusal(t, "OFF dormancy", out)
	})

	t.Run("queue-picked card with no row is neither recorded nor leased", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStateQueued)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out, _, err := runFactory(t, "next", "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		if fcHasCard(t, root, "t1") {
			t.Fatalf("t1 gained a record row although arm (b2) is dormant — the skip must be traceless")
		}
		if c := fcCard(t, root, "t2"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
			t.Fatalf("t2 = %s holder=%q, want leased/lane-1 through arm (c)", c.State, c.LeaseHolder)
		}
		cepAssertNoRefusal(t, "OFF dormancy b2", out)
	})

	t.Run("lone unowned picked card answers no-card with no reason line", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRowFrozen(t, root, "t1", before)
		cepAssertNoRefusal(t, "OFF quiet skip", out)
		if got, want := fcQueueItemState(t, store, "t1"), factory.BacklogStatePicked; got != want {
			t.Errorf("t1 queue state = %s, want %s", got, want)
		}
	})

	t.Run("unreadable config fails safe to dormancy", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		cepSetSpecID(t, store, "t1", "SPEC-GATE-001")
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		before := fcCard(t, root, "t1")
		// A workflow.yaml the loader must reject: the kill-switch default is
		// the fail-safe, so the arms stay dormant rather than arming on a
		// half-read config.
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir config sections: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte("workflow: ["), 0o644); err != nil {
			t.Fatalf("write broken workflow.yaml: %v", err)
		}
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRowFrozen(t, root, "t1", before)
		cepAssertNoRefusal(t, "unreadable config", out)
	})
}

// cepSetSpecID attaches a spec id to one queue item — the eligible-candidate
// shape under the amended (0.3.0) fail-closed condition 2: a card with no
// spec id anywhere is never leased through the generic arms.
func cepSetSpecID(t *testing.T, store *factory.BacklogStore, cardID, spec string) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == cardID {
				s := spec
				r.Items[i].SpecID = &s
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("set %s spec id: %v", cardID, err)
	}
}

// fcQueueItemState reads one item's queue state.
func fcQueueItemState(t *testing.T, store *factory.BacklogStore, cardID string) factory.BacklogState {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == cardID {
			return it.State
		}
	}
	t.Fatalf("queue item %s not found", cardID)
	return ""
}

// AC-CEP-002 — flag ON: one disqualification cell per closed-set token. Each
// fixture leaves the candidate unowned-picked with exactly one defect.
func TestFactoryEligiblePullOnGate(t *testing.T) {
	t.Run("i queue item held: refused not-admitted", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStateHold)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRefusal(t, out, "not-admitted", "t1")
		cepAssertRowFrozen(t, root, "t1", before)
	})

	t.Run("ii class-C high without spec: refused spec-missing", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityHigh, false, factory.ClassModeParallelizable)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRefusal(t, out, "spec-missing", "t1")
		cepAssertRowFrozen(t, root, "t1", before)
	})

	// Amendment 0.3.0 (F1 ruling): condition 2 is FAIL-CLOSED and
	// classification-agnostic — a spec-less candidate never leases through
	// the generic arms whatever its priority. This cell is the repair
	// verification: under the removed priority==high surrogate, a
	// normal-priority no-spec candidate leased.
	t.Run("ii-b normal priority without spec: refused spec-missing (fail-closed, amendment 0.3.0)", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRefusal(t, out, "spec-missing", "t1")
		cepAssertRowFrozen(t, root, "t1", before)
	})

	t.Run("iii picked plus hold marker: refused hold-marker", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		cepSetSpecID(t, store, "t1", "SPEC-GATE-001")
		cepMarkText(t, store, "t1", autoRankHoldMarker+" parked until the audit lands")
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRefusal(t, out, "hold-marker", "t1")
		cepAssertRowFrozen(t, root, "t1", before)
	})

	t.Run("iv unmerged after-predecessor: refused predecessor-unmerged", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		cepSetSpecID(t, store, "t1", "SPEC-GATE-001")
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked, HintAfter: "t0"})
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRefusal(t, out, "predecessor-unmerged", "t1")
		cepAssertRowFrozen(t, root, "t1", before)
	})

	t.Run("v serial slot held: refused serial-slot", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		cepSetSpecID(t, store, "t1", "SPEC-GATE-005")
		sdRegisterLane(t, root, "lane-1")
		live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
		fcPlace(t, root,
			homestate.Card{CardID: "t1", State: homestate.CardPicked},
			homestate.Card{CardID: "t2", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live},
		)
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out := cepNoCard(t, root)
		cepAssertRefusal(t, out, "serial-slot", "t1")
		cepAssertRowFrozen(t, root, "t1", before)
	})
}

// AC-CEP-003 — flag ON, rule-paired: the eligible sibling leases through the
// generic arm (no over-blocking) while the ineligible sibling, once it is the
// remaining candidate, is refused with its token (no over-passing). Both
// assertions run on one fixture.
func TestFactoryEligiblePullOnEligible(t *testing.T) {
	root, store := fcFixture(t)
	cepSelfDispatchYAML(t, root, true)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityHigh, false, factory.ClassModeParallelizable)
	// Amendment 0.3.0: the eligible sibling carries a spec id — condition 2
	// is fail-closed, so eligibility now requires the id whatever the class.
	cepSetSpecID(t, store, "t1", "SPEC-ELIGIBLE-001")
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardPicked},
		homestate.Card{CardID: "t2", State: homestate.CardPicked},
	)

	// No over-blocking: the eligible candidate leases through arm (b).
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)
	out, _, err := runFactory(t, "next", "--run", fcRun)
	if err != nil {
		t.Fatalf("next leases the eligible candidate: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t1 = %s holder=%q, want leased/lane-1 through arm (b) — the gate must not block eligible cards", c.State, c.LeaseHolder)
	}
	cepAssertNoRefusal(t, "leased pass", out)

	// No over-passing: the spec-missing sibling is refused with its token.
	before := fcCard(t, root, "t2")
	sdLaneEnv(t, "lane-2", "")
	out2 := cepNoCard(t, root)
	cepAssertRefusal(t, out2, "spec-missing", "t2")
	cepAssertRowFrozen(t, root, "t2", before)
}

// AC-CEP-005 — skip and refusal are traceless: no new row, no claim, and the
// candidate's row and queue bytes unchanged, on both the OFF skip and the ON
// refusal paths; the exit stays the no-card 3.
func TestFactoryEligiblePullTraceless(t *testing.T) {
	for _, tc := range []struct {
		name string
		on   bool
	}{
		{"off quiet skip", false},
		{"on gate refusal", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, store := fcFixture(t)
			if tc.on {
				cepSelfDispatchYAML(t, root, true)
			}
			fcQueue(t, store, factory.BacklogStatePicked)
			fcClassify(t, store, "t1", factory.ClassPriorityHigh, false, factory.ClassModeParallelizable)
			sdRegisterLane(t, root, "lane-1")
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
			before, qBefore := fcCard(t, root, "t1"), sdQueueBytes(t, store)
			sdLaneEnv(t, "lane-1", "")
			t.Chdir(root)
			out := cepNoCard(t, root)
			cepAssertRowFrozen(t, root, "t1", before)
			if qAfter := sdQueueBytes(t, store); qAfter != qBefore {
				t.Errorf("queue bytes changed across the %s:\nbefore %s\nafter  %s", tc.name, qBefore, qAfter)
			}
			if tc.on {
				// The refusal LINE is the AC-CEP-002/006 surface (M4) and is
				// asserted by the gate and grammar cells; this cell pins the
				// tracelessness alone, which is AC-CEP-005's own claim.
				_ = out
			} else {
				cepAssertNoRefusal(t, tc.name, out)
			}
		})
	}
}

// AC-CEP-006 — the refusal surface is machine-readable on both surfaces:
// the first line stays the existing no-card line, every following line
// matches the closed-token grammar exactly once per disqualified candidate,
// the exit stays 3, and the cobra verb and the MCP tool render the same
// text from the one shared source.
func TestFactoryEligiblePullRefusalGrammar(t *testing.T) {
	root, store := fcFixture(t)
	cepSelfDispatchYAML(t, root, true)
	fcQueue(t, store, factory.BacklogStateHold, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityHigh, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t3", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	// t3 carries a spec id so its FIRST disqualifier is the marker (condition 3),
	// not the fail-closed spec check (condition 2) — each grammar line demos a
	// distinct token.
	cepSetSpecID(t, store, "t3", "SPEC-GRAMMAR-003")
	cepMarkText(t, store, "t3", autoRankHoldMarker+" parked")
	sdRegisterLane(t, root, "lane-1")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardPicked},
		homestate.Card{CardID: "t2", State: homestate.CardPicked},
		homestate.Card{CardID: "t3", State: homestate.CardPicked},
	)
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	grammar := regexp.MustCompile(`^refused (not-admitted|spec-missing|hold-marker|predecessor-unmerged|serial-slot): .+$`)
	// cepNoCardBlock picks the response block off a surface's text: the
	// no-card first line plus the contiguous `refused` lines that follow.
	// The cobra verb's error also prints a usage block after the response —
	// that noise is not part of the answer and stops the block.
	block := func(text string) []string {
		lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
		var b []string
		for i, l := range lines {
			if i == 0 || strings.HasPrefix(l, "refused ") {
				b = append(b, l)
				continue
			}
			break
		}
		return b
	}
	parse := func(t *testing.T, label, text string) {
		t.Helper()
		lines := block(text)
		if len(lines) == 0 || lines[0] != "no card is available" {
			t.Fatalf("%s first line = %q, want the existing no-card line", label, firstOr(lines))
		}
		rest := lines[1:]
		if len(rest) != 3 {
			t.Fatalf("%s carried %d refusal lines (%v), want exactly one per disqualified candidate", label, len(rest), rest)
		}
		seen := map[string]bool{}
		for _, l := range rest {
			if !grammar.MatchString(l) {
				t.Errorf("%s line %q does not match the closed-token grammar", label, l)
				continue
			}
			seen[l] = true
		}
		for _, id := range []string{"t1", "t2", "t3"} {
			if !anyContains(seen, id) {
				t.Errorf("%s: no refusal line names %s", label, id)
			}
		}
	}

	cliOut := cepNoCard(t, root)
	parse(t, "cli", cliOut)

	// The MCP tool renders the same text from the same shared source; the
	// refusal pass left the candidates untouched, so the same fixture serves
	// the second surface.
	mcpOut, mcpErr := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun})
	if mcpErr != nil {
		t.Fatalf("mcp factory_next on a refused pass: %v", mcpErr)
	}
	parse(t, "mcp", mcpOut)
	if strings.Join(block(mcpOut), "\n") != strings.Join(block(cliOut), "\n") {
		t.Errorf("surface texts diverge:\ncli %q\nmcp %q", strings.Join(block(cliOut), "\n"), strings.Join(block(mcpOut), "\n"))
	}
}

func firstOr(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

func anyContains(set map[string]bool, sub string) bool {
	for l := range set {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}
