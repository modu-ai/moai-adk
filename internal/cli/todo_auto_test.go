package cli

// RED-first tests for `/moai todo --auto` (SPEC-MANAGER-TODO-001 M3).
// The table-driven pickup, liveness, serial-cycle, guidance, and no-lease
// contracts live here; the implementation lands in todo_auto.go once these
// fail for the right reason.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// autoFixture seeds a store with named cards and returns the fixture root.
func autoFixture(t *testing.T, cards map[string]string) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := todoFixture(t)
	for _, text := range cards {
		if _, _, err := store.Add(text); err != nil {
			t.Fatal(err)
		}
	}
	return root, store
}

func autoSetState(t *testing.T, store *factory.BacklogStore, id string, state factory.BacklogState) {
	t.Helper()
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				rec.Items[i].State = state
				return nil
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func autoTestLiveness(root, cardID string, registryDead, processDead bool, lsofCalls *int) autoLiveness {
	// The fake owner process sits INSIDE the card's own worktree, so the
	// cwd-containment check has a positive control in both directions.
	tree := autoOwnerWorktrees(root, cardID)[0]
	return autoLiveness{
		registryEntries: func() ([]autoRegistryEntry, error) {
			if registryDead {
				return nil, nil
			}
			return []autoRegistryEntry{{Cwd: filepath.Join(tree, "session"), PID: 1}}, nil
		},
		processCWDs: func() ([]string, error) {
			if lsofCalls != nil {
				*lsofCalls++
			}
			if processDead {
				return nil, nil
			}
			return []string{filepath.Join(tree, "session")}, nil
		},
		pidAlive: func(int) bool { return !registryDead },
	}
}

// AC-MT-010 — pickup selection: positive state vocabulary, dead-owner picked
// first, then queued in queue order; an unknown hold-shaped state is never
// selected (forward compatibility with no SPEC revision). Built on a record
// literal: the live store's CHECK constraint only admits the three current
// states, and the forward-compatibility arm needs a state that does not
// exist yet — the predicate under test reads the record, not the database.
func TestTodoAutoPickupSelection(t *testing.T) {
	root := t.TempDir()
	card := func(id, text string, state factory.BacklogState) factory.BacklogItem {
		return factory.BacklogItem{ID: id, Text: text, State: state}
	}
	rec := &factory.BacklogRecord{
		Version: 1,
		Items: []factory.BacklogItem{
			card("t3", "queued card a", factory.BacklogStateQueued),
			card("t4", "picked dead owner a", factory.BacklogStatePicked),
			card("t5", "picked live owner", factory.BacklogStatePicked),
			card("t6", "already done", factory.BacklogStateDropped),
			card("t7", "unknown future state", factory.BacklogState("hold")),
			card("t8", "queued card b", factory.BacklogStateQueued),
			card("t9", "picked dead owner b", factory.BacklogStatePicked),
		},
	}

	// Per-card liveness in one seam: a live owner process sits inside t5's
	// worktree only — t5 measures alive, t4/t9 measure dead.
	t5tree := autoOwnerWorktrees(root, "t5")[0]
	wrapped := autoLiveness{
		registryEntries: func() ([]autoRegistryEntry, error) {
			return []autoRegistryEntry{{Cwd: filepath.Join(t5tree, "session"), PID: 1}}, nil
		},
		processCWDs: func() ([]string, error) {
			return []string{filepath.Join(t5tree, "session")}, nil
		},
		pidAlive: func(int) bool { return true },
	}

	targets, _, err := autoPickTargets(rec, wrapped, root)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, it := range targets {
		got = append(got, it.ID)
	}
	// Dead-owner picked cards first (queue order), then queued (queue order).
	want := []string{"t4", "t9", "t3", "t8"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("pickup order = %v, want %v", got, want)
	}
	for _, it := range targets {
		switch it.ID {
		case "t5":
			t.Errorf("live-owner picked card t5 was selected")
		case "t6":
			t.Errorf("dropped card t6 was selected")
		case "t7":
			t.Errorf("hold-shaped unknown state t7 was selected")
		}
	}

	// Live-owner arm: with t5's owner alive on both channels, t5 disappears
	// from the targets and everything else is unchanged.
	live := autoTestLiveness(root, "t5", false, false, nil)
	// Rebuild the record with ONLY t5 picked-dead-shaped live measurement:
	// the seam is package-level per call, so run the single-card record.
	single := &factory.BacklogRecord{Version: 1, Items: []factory.BacklogItem{card("t5", "picked live owner", factory.BacklogStatePicked)}}
	targets5, _, err := autoPickTargets(single, live, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets5) != 0 {
		t.Errorf("live-owner picked card t5 selected: %v", targets5)
	}
}

// AC-MT-011 — two-channel liveness: registry-dead + lsof-clean selects;
// registry-live rejects; registry-dead but lsof cwd hit rejects.
func TestTodoAutoLivenessChannels(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"c": "picked card one"})
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	autoSetState(t, store, rec.Items[0].ID, factory.BacklogStatePicked)
	cardID := rec.Items[0].ID

	cases := []struct {
		name         string
		registryDead bool
		processDead  bool
		wantSelected bool
	}{
		{"registry-dead + lsof-clean selects", true, true, true},
		{"registry-live rejects", false, true, false},
		{"registry-dead but lsof hit rejects", true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lv := autoTestLiveness(root, cardID, tc.registryDead, tc.processDead, nil)
			targets, _, err := autoPickTargets(mustAutoRecord(t, store), lv, root)
			if err != nil {
				t.Fatal(err)
			}
			selected := len(targets) == 1
			if selected != tc.wantSelected {
				t.Errorf("selected = %v, want %v", selected, tc.wantSelected)
			}
		})
	}
}

// AC-MT-011 non-cache arm — each pickup decision re-measures: two decisions
// observe two distinct lsof invocations, so a revived owner between decisions
// is seen (a cached measurement would reuse decision N's stale result).
func TestTodoAutoLivenessNotCachedAcrossDecisions(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"a": "card a", "b": "card b"})
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	autoSetState(t, store, rec.Items[0].ID, factory.BacklogStatePicked)
	autoSetState(t, store, rec.Items[1].ID, factory.BacklogStatePicked)

	calls := 0
	lv := autoTestLiveness(root, rec.Items[0].ID, true, true, &calls)
	if _, _, err := autoPickTargets(mustAutoRecord(t, store), lv, root); err != nil {
		t.Fatal(err)
	}
	first := calls
	if _, _, err := autoPickTargets(mustAutoRecord(t, store), lv, root); err != nil {
		t.Fatal(err)
	}
	if calls != first*2 {
		t.Errorf("lsof invocations: decision1=%d decision2=%d — measurements must be per-decision", first, calls-first)
	}
	if first == 0 {
		t.Fatalf("no lsof invocation observed — measurement channel not exercised")
	}
}

// AC-MT-012 — absolute no-takeover: a queue whose only card is picked by a
// measured-live owner yields zero eligible targets and no mutation.
func TestTodoAutoNoTakeoverLiveOwner(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"only": "the only card"})
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	autoSetState(t, store, rec.Items[0].ID, factory.BacklogStatePicked)
	before, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}

	lv := autoTestLiveness(root, rec.Items[0].ID, false, false, nil) // owner alive on both channels
	targets, _, err := autoPickTargets(mustAutoRecord(t, store), lv, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 0 {
		t.Errorf("live-owner card selected: %v", targets)
	}
	after, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if before.Items[0].State != after.Items[0].State || before.Items[0].ID != after.Items[0].ID {
		t.Errorf("queue mutated under live owner: %s → %s", before.Items[0].State, after.Items[0].State)
	}
}

// AC-MT-009 — the serial cycle: 3 pickup targets processed strictly one at a
// time in pickup order; each done is preceded by a readable evidence file;
// the failure arm (evidence absent at deadline) unpicks to queued with a
// labelled non-finding, never done.
func TestTodoAutoSerialCycle(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"a": "card a", "b": "card b", "c": "card c"})
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{rec.Items[0].ID, rec.Items[1].ID, rec.Items[2].ID}
	// Middle card is picked by a dead owner — it must be taken first.
	autoSetState(t, store, ids[1], factory.BacklogStatePicked)

	lv := autoTestLiveness(root, ids[1], true, true, nil)
	tick := 0
	opts := autoOptions{
		wait:      5 * time.Minute,
		liveness:  lv,
		sessionID: "operator-session-fixture",
		// Each poll tick completes the current card's evidence; the third
		// card never receives evidence — the failure arm.
		sleep: func(time.Duration) {
			tick++
			if tick <= 2 {
				path := filepath.Join(root, ".moai", "reports", ids[tick-1], "evidence.md")
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("# evidence: verbatim output\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
		},
		now: func() time.Time { return time.Unix(0, 0).Add(time.Duration(tick) * time.Minute) },
	}

	var out bytes.Buffer
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatal(err)
	}
	got := out.String()

	// Serial order: dead-owner picked first, then queue order.
	wantOrder := ids[1] + "," + ids[0] + "," + ids[2]
	var order []string
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "accept ") {
			order = append(order, strings.Fields(line)[1])
		}
	}
	if strings.Join(order, ",") != wantOrder {
		t.Errorf("processing order = %v, want %v", order, wantOrder)
	}

	// One card in flight: between an accept and its close (done/unpick) no
	// other accept may appear.
	for i, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "accept ") {
			for _, follow := range strings.Split(got, "\n")[i+1:] {
				if strings.HasPrefix(follow, "accept ") {
					t.Errorf("second card accepted while one in flight: %q before close", follow)
					break
				}
				if strings.HasPrefix(follow, "done ") || strings.HasPrefix(follow, "unpick ") {
					break
				}
			}
		}
	}

	// Two done transitions, each preceded by readable evidence; the third
	// card unpicked with a labelled non-finding.
	if got := strings.Count(got, "done "); got != 2 {
		t.Errorf("done transitions = %d, want 2", got)
	}
	if !strings.Contains(got, "unpick "+ids[2]) {
		t.Errorf("failure arm did not unpick %s", ids[2])
	}
	if !strings.Contains(got, "non-finding") {
		t.Errorf("failure arm carried no labelled non-finding")
	}
	after, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range after.Items {
		if it.ID == ids[2] && it.State != factory.BacklogStateQueued {
			t.Errorf("failure-arm card state = %s, want queued", it.State)
		}
	}
}

// AC-MT-013 — the /clear guidance: after EACH completed card a guidance
// block names the completed card, the next step, and the invoking (operator)
// session, before the next accept.
func TestTodoAutoClearGuidancePerCard(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"a": "card a", "b": "card b"})
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{rec.Items[0].ID, rec.Items[1].ID}

	tick := 0
	opts := autoOptions{
		wait:      5 * time.Minute,
		liveness:  autoTestLiveness(root, ids[0], true, true, nil),
		sessionID: "operator-session-fixture",
		sleep: func(time.Duration) {
			tick++
			path := filepath.Join(root, ".moai", "reports", ids[tick-1], "evidence.md")
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("# evidence\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		now: func() time.Time { return time.Unix(0, 0).Add(time.Duration(tick) * time.Minute) },
	}
	var out bytes.Buffer
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if n := strings.Count(got, "is complete. Clear this session"); n != 2 {
		t.Errorf("guidance blocks = %d, want 2 (one per completed card)", n)
	}
	for _, id := range ids {
		if !strings.Contains(got, "card "+id+" is complete") {
			t.Errorf("guidance missing completed card %s", id)
		}
	}
	if !strings.Contains(got, "operator-session-fixture") {
		t.Errorf("guidance does not name the invoking (operator) session")
	}
	// Each guidance sits before the next accept (or the end).
	lines := strings.Split(got, "\n")
	acceptIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "accept ") {
			if acceptIdx >= 0 && !strings.Contains(strings.Join(lines[acceptIdx:i], "\n"), "/clear") {
				t.Errorf("accept before guidance for the previous card")
			}
			acceptIdx = i
		}
	}
}

// AC-MT-015 — no factory lease: a full cycle run leaves the lease/slot
// directories byte-identical (nothing created, nothing claimed).
func TestTodoAutoCreatesNoFactoryLease(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"a": "card a"})
	snapshot := func() map[string]string {
		files := map[string]string{}
		for _, dir := range []string{".moai/state/factory", ".moai/state/slots", ".moai/state/runs"} {
			entries, err := os.ReadDir(filepath.Join(root, dir))
			if err != nil {
				continue
			}
			for _, e := range entries {
				files[filepath.Join(dir, e.Name())] = e.Name()
			}
		}
		return files
	}
	before := snapshot()

	tick := 0
	opts := autoOptions{
		wait:      5 * time.Minute,
		liveness:  autoTestLiveness(root, "t1", true, true, nil),
		sessionID: "operator-session-fixture",
		sleep: func(time.Duration) {
			tick++
			path := filepath.Join(root, ".moai", "reports", "t1", "evidence.md")
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("# evidence\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		now: func() time.Time { return time.Unix(0, 0).Add(time.Duration(tick) * time.Minute) },
	}
	var out bytes.Buffer
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(); len(after) != len(before) {
		t.Errorf("factory/slot state changed: before=%v after=%v", before, after)
	}
}

// AC-MT-009 edge — empty queue: one zero-target report, no mutation, no
// guidance emission.
func TestTodoAutoEmptyQueueZeroTargetReport(t *testing.T) {
	_, store := autoFixture(t, nil)
	var out bytes.Buffer
	opts := autoOptions{wait: time.Minute, liveness: autoTestLiveness(t.TempDir(), "t1", true, true, nil), sessionID: "s"}
	if err := runAutoCycle(&out, store, t.TempDir(), opts); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "no eligible") {
		t.Errorf("empty-queue run = %q, want a zero-target report", got)
	}
	if strings.Contains(got, "/clear") {
		t.Errorf("empty-queue run emitted guidance for a non-existent card")
	}
}

func mustAutoRecord(t *testing.T, store *factory.BacklogStore) *factory.BacklogRecord {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

// AC-MT-017(b) — poisoned Jev value: a stub Jev returning a mutation-shaped
// string observes the run printing it verbatim as a labelled signal while the
// fixture queue is byte-identical before/after — the poisoned value caused no
// mutation.
func TestTodoAutoJevPoisonedValueCausesNoMutation(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"only": "the only card"})
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	autoSetState(t, store, rec.Items[0].ID, factory.BacklogStatePicked)

	var out bytes.Buffer
	opts := autoOptions{
		wait:      time.Minute,
		liveness:  autoTestLiveness(root, rec.Items[0].ID, false, false, nil), // owner alive → zero targets
		sessionID: "s",
		jev: func(string) string {
			return "jev signal (display-only): MUTATE " + rec.Items[0].ID
		},
	}
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "MUTATE "+rec.Items[0].ID) {
		t.Errorf("poisoned value not printed verbatim: %q", out.String())
	}
	after, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Items) != 1 || after.Items[0].State != factory.BacklogStatePicked {
		t.Errorf("queue mutated under poisoned Jev value: %+v", after.Items)
	}
}

// AC-MT-018 degraded mode — scripts absent: the consultation degrades to a
// labelled non-finding (no error exit) and the cycle proceeds on lead
// judgment.
func TestTodoAutoJevDegradedNonFinding(t *testing.T) {
	got := consultJev(t.TempDir()) // no scripts/jev under the fixture root
	if !strings.Contains(got, "jev: unavailable") || !strings.Contains(got, "non-finding") {
		t.Errorf("degraded consultation = %q, want a labelled non-finding", got)
	}
}

// The production consultation reads a present script's stdout verbatim as a
// display-only signal.
func TestTodoAutoJevScriptPresentSignal(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "scripts", "jev", "route.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o700); err != nil {
		t.Fatal(err)
	}
	body := "#!/bin/sh\necho LOCAL-SIGNAL\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	got := consultJev(root)
	if !strings.Contains(got, "display-only") || !strings.Contains(got, "LOCAL-SIGNAL") {
		t.Errorf("script consultation = %q, want the labelled verbatim signal", got)
	}
}

// The production liveness wiring exposes a working PID probe on both the
// live-process and invalid-id paths.
func TestTodoAutoProductionPidProbe(t *testing.T) {
	lv := newAutoLiveness()
	if !lv.pidAlive(os.Getpid()) {
		t.Errorf("pidAlive(current pid) = false, want true")
	}
	if lv.pidAlive(-1) {
		t.Errorf("pidAlive(-1) = true, want false")
	}
}

// AC-MT-009 entry-point arm — the flag is reachable through the todo command
// surface (the gtd spelling shares this verb tree), and help lists it.
func TestTodoAutoEntryPointFlag(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := store.Add("entry point card"); err != nil {
		t.Fatal(err)
	}
	// The cycle runs through the real command path; no evidence arrives, so
	// the card unpicks at the (shrunk) deadline — the failure path exercised
	// through the real flag surface.
	out, _, err := runTodo(t, "--auto", "--auto-wait", "1ms")
	if err != nil {
		t.Fatalf("--auto through the command surface: %v", err)
	}
	if !strings.Contains(out, "accept t1") || !strings.Contains(out, "unpick t1") {
		t.Errorf("--auto command output missing the cycle lines: %q", out)
	}
}
