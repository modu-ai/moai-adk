package escalation_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// disarmRecords returns the detection-disarmed records of the worktree.
func disarmRecords(t *testing.T, w *escalationtest.Worktree) []escalation.Record {
	t.Helper()
	rs, _ := recordsOfClass(t, w, escalation.ClassDetectionDisarmed)
	return rs
}

// kinds returns the log entry kinds in order.
func kinds(lg escalation.CardLog) []string {
	var out []string
	for _, e := range lg.Entries {
		out = append(out, e.Kind)
	}
	return out
}

// assertDisarmedBeforeNotArmed checks that the one disarmed entry precedes
// every not-armed line written after it.
func assertOneDisarm(t *testing.T, w *escalationtest.Worktree, reason string) {
	t.Helper()
	dr := disarmRecords(t, w)
	if len(dr) != 1 || dr[0].ContractRef != "disarm:"+reason {
		t.Fatalf("detection-disarmed records = %+v, want one disarm:%s", dr, reason)
	}
	lg := cardLog(t, w)
	d := entriesOf(lg, escalation.LineDisarmed)
	if len(d) != 1 || d[0].Reason != reason || d[0].Fingerprint != dr[0].Fingerprint {
		t.Fatalf("disarmed entries = %+v", d)
	}
	ks := kinds(lg)
	di := strings.Index(strings.Join(ks, ","), "disarmed")
	ni := strings.Index(strings.Join(ks, ","), "not-armed")
	if ni >= 0 && ni < di {
		t.Errorf("a not-armed line precedes the disarmed entry: %v", ks)
	}
}

// AC-AE-016 (REQ-AE-017): contract loss writes exactly one
// detection-disarmed record and one disarmed entry; afterwards an
// out-of-scope write trips nothing and an over-budget operation still trips.
func TestDisarmContractLossWritesOneRecord(t *testing.T) {
	contractRel := ".moai/specs/SPEC-A-001/contract.yaml"
	cases := []struct {
		name, reason string
		act          func(t *testing.T, w *escalationtest.Worktree)
	}{
		{"signature-removed", "signature-invalid", func(t *testing.T, w *escalationtest.Worktree) {
			text := string(w.Read(contractRel))
			w.Write(contractRel, text[:strings.Index(text, "\nsignature:")+1])
			preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		}},
		{"contract-deleted", "contract-absent", func(t *testing.T, w *escalationtest.Worktree) {
			if err := os.Remove(w.Path(contractRel)); err != nil {
				t.Fatal(err)
			}
			preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		}},
		{"byte-changed", "signature-invalid", func(t *testing.T, w *escalationtest.Worktree) {
			w.Replace(contractRel, "Implement the fixture feature", "Implement another feature")
			commitCheckpoint(contractSettings(t, w), w)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateStore(t)
			w := armedWith(t, "t9001", budgetEdit("3", "2"), "")
			tc.act(t, w)
			assertOneDisarm(t, w, tc.reason)

			preWrite(contractSettings(t, w), w, "internal/bar/out-of-scope.go")
			if rs, _ := recordsOfClass(t, w, escalation.ClassOwnershipMove); len(rs) != 0 {
				t.Errorf("out-of-scope write tripped after disarm: %+v", rs)
			}
			for i := 0; i < 4; i++ {
				postBash(t, w, "ls", false, "")
			}
			if rs, _ := recordsOfClass(t, w, escalation.ClassBudgetExceeded); len(rs) != 1 {
				t.Errorf("over-budget operation after disarm wrote %d budget records", len(rs))
			}
			assertOneDisarm(t, w, tc.reason)
		})
	}
}

// AC-AE-017 (REQ-AE-017): terminal status, a re-signed card change, and a
// second claimant each write one detection-disarmed record; a second reason
// after the first increments that record and writes nothing new.
func TestDisarmTransitions(t *testing.T) {
	t.Run("terminal-status", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", nil, "")
		w.Replace(".moai/specs/SPEC-A-001/spec.md", "status: in-progress", "status: completed")
		commitCheckpoint(contractSettings(t, w), w)
		assertOneDisarm(t, w, "terminal-status")

		if err := os.Remove(w.Path(".moai/specs/SPEC-A-001/contract.yaml")); err != nil {
			t.Fatal(err)
		}
		preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		dr := disarmRecords(t, w)
		if len(dr) != 1 || dr[0].Occurrences != 2 {
			t.Errorf("after a second reason: records = %+v, want one with occurrences 2", dr)
		}
		if n := len(entriesOf(cardLog(t, w), escalation.LineDisarmed)); n != 1 {
			t.Errorf("disarmed entries = %d, want 1", n)
		}
	})
	t.Run("card-changed", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", nil, "")
		w.Resign("SPEC-A-001", func(s string) string { return strings.Replace(s, "card: t9001", "card: t9002", 1) })
		preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		assertOneDisarm(t, w, "card-mismatch")
	})
	t.Run("second-claimant", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", nil, "")
		w.AddSpec("SPEC-B-001", escalationtest.SpecOptions{Unsigned: true})
		preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		assertOneDisarm(t, w, "card-mismatch")
	})
}

// AC-AE-018 (REQ-AE-017): state-file tampering and a mid-log edit or a tail
// cut are judged state-tamper from the card log — one record, one disarmed
// entry each; removing the armed value never reads as a plain disarm.
func TestStateTamperJudgedFromCardLog(t *testing.T) {
	cases := []struct {
		name string
		act  func(t *testing.T, files escalation.CardFiles)
	}{
		{"a-armed-value-removed", func(t *testing.T, files escalation.CardFiles) {
			var st map[string]any
			data, _ := os.ReadFile(files.State)
			if err := json.Unmarshal(data, &st); err != nil {
				t.Fatal(err)
			}
			delete(st, "armed")
			out, _ := json.Marshal(st)
			writeRaw(t, files.State, out)
		}},
		{"b-state-rewritten", func(t *testing.T, files escalation.CardFiles) {
			data, _ := os.ReadFile(files.State)
			writeRaw(t, files.State, append(data, ' '))
		}},
		{"c-middle-line-altered", func(t *testing.T, files escalation.CardFiles) {
			data, _ := os.ReadFile(files.Log)
			lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
			if len(lines) < 3 {
				t.Fatalf("log too short: %d lines", len(lines))
			}
			lines[1] = strings.Replace(lines[1], `"card":"t9001"`, `"card":"t9001","x":1`, 1)
			writeRaw(t, files.Log, []byte(strings.Join(lines, "\n")+"\n"))
		}},
		{"d-tail-truncated", func(t *testing.T, files escalation.CardFiles) {
			data, _ := os.ReadFile(files.Log)
			lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
			cut := -1
			for i, l := range lines {
				if strings.Contains(l, `"kind":"armed"`) {
					cut = i
				}
			}
			if cut < 0 {
				t.Fatal("no armed line")
			}
			writeRaw(t, files.Log, []byte(strings.Join(lines[:cut], "\n")+"\n"))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateStore(t)
			w := armedWith(t, "t9001", nil, "")
			postBash(t, w, "ls", false, "") // one more state entry: armed, state, state
			files, _ := escalation.CardFilesFor(w.Root, w.Card)
			tc.act(t, files)
			preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
			assertOneDisarm(t, w, "state-tamper")
		})
	}
}

func writeRaw(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// AC-AE-020 (REQ-AE-017): a card log that does not show the card armed reads
// as never armed unless unaccounted evidence of arming exists.
func TestMissingCardLogReading(t *testing.T) {
	armedState := func(t *testing.T) (*escalationtest.Worktree, escalation.CardFiles, []byte) {
		w := armedWith(t, "t9001", nil, "")
		files, _ := escalation.CardFilesFor(w.Root, w.Card)
		state, _ := os.ReadFile(files.State)
		return w, files, state
	}
	t.Run("a-nothing", func(t *testing.T) {
		isolateStore(t)
		w := escalationtest.NewWorktree(t, "t9001")
		w.WriteContractMode("")
		w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
		preWrite(contractSettings(t, w), w, "internal/fixture/a.go")
		if len(disarmRecords(t, w)) != 0 {
			t.Fatal("never-armed card wrote a disarm record")
		}
		if !cardLog(t, w).Armed() {
			t.Error("never-armed card did not arm through resolution")
		}
	})
	t.Run("b-state-armed", func(t *testing.T) {
		isolateStore(t)
		w, files, _ := armedState(t)
		if err := os.Remove(files.Log); err != nil {
			t.Fatal(err)
		}
		preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		assertOneDisarm(t, w, "state-tamper")
		if ks := kinds(cardLog(t, w)); ks[0] != escalation.LineDisarmed {
			t.Errorf("first entry of the new log = %v", ks)
		}
	})
	t.Run("c-disarm-record", func(t *testing.T) {
		isolateStore(t)
		w, files, _ := armedState(t)
		fp := escalation.Fingerprint(escalation.ClassDetectionDisarmed, strings.Repeat("e", 64))
		if _, err := escalation.WriteRecord(w.Root, escalation.Record{Card: "t9001", Kind: escalation.KindOperational,
			Class: escalation.ClassDetectionDisarmed, Fingerprint: fp, ContractRef: "disarm:signature-invalid",
			Observation: "earlier", Options: []string{"a", "b"}}, fixedNow); err != nil {
			t.Fatal(err)
		}
		_ = os.Remove(files.Log)
		_ = os.Remove(files.State)
		preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		var tamper int
		for _, r := range disarmRecords(t, w) {
			if r.ContractRef == "disarm:state-tamper" {
				tamper++
			}
		}
		if tamper != 1 {
			t.Fatalf("state-tamper records = %d, want 1", tamper)
		}
		if ks := kinds(cardLog(t, w)); ks[0] != escalation.LineDisarmed {
			t.Errorf("first entry of the new log = %v", ks)
		}
		// Accounted: a later hook does not judge the same evidence again.
		preWrite(contractSettings(t, w), w, "internal/fixture/c.go")
		if n := len(entriesOf(cardLog(t, w), escalation.LineDisarmed)); n != 1 {
			t.Errorf("evidence judged again: %d disarmed entries", n)
		}
	})
	t.Run("d-empty-log", func(t *testing.T) {
		isolateStore(t)
		w, files, _ := armedState(t)
		writeRaw(t, files.Log, nil)
		preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
		assertOneDisarm(t, w, "state-tamper")
	})
}

// AC-AE-019 part 1: another card's log never affects this card.
func TestOtherCardLogDoesNotAffectThisCard(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	isolateStore(t)
	p := escalationtest.NewGitWorktree(t, "proj")
	p.WriteContractMode("")
	p.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Card: "t9001"})
	p.AddSpec("SPEC-B-001", escalationtest.SpecOptions{Card: "t9002"})
	p.Commit("base")
	w1 := escalationtest.NewLinkedWorktree(t, p, "t9001")
	w2 := escalationtest.NewLinkedWorktree(t, p, "t9002")
	s := contractSettings(t, w1)
	preWrite(s, w1, "internal/fixture/a.go")
	contractB := w2.Path(".moai/specs/SPEC-B-001/contract.yaml")
	saved, _ := os.ReadFile(contractB)
	for i := 0; len(cardLog(t, w2).Entries) < 50; i++ {
		switch i {
		case 10:
			_ = os.Remove(contractB) // disarm t9002
		case 12:
			writeRaw(t, contractB, saved) // re-arm t9002
		}
		escalation.Observe(s, escalation.Event{Hook: escalation.HookPostToolUse, CWD: w2.Root, ToolName: "Bash", Command: "ls"})
		preWrite(s, w1, "internal/fixture/a.go")
		if i > 200 {
			t.Fatal("t9002 log did not reach 50 entries")
		}
	}
	lg1 := cardLog(t, w1)
	if !lg1.ChainIntact || !lg1.Armed() {
		t.Errorf("t9001 chain intact=%v armed=%v", lg1.ChainIntact, lg1.Armed())
	}
	for _, e := range lg1.Entries {
		if e.Card != "t9001" {
			t.Errorf("t9001 log carries an entry for %s", e.Card)
		}
	}
	if len(disarmRecords(t, w1)) != 0 {
		t.Errorf("t9001 wrote disarm records")
	}
	k2 := kinds(cardLog(t, w2))
	if !strings.Contains(strings.Join(k2, ","), "disarmed") || strings.Count(strings.Join(k2, ","), "armed") < 2 {
		t.Errorf("t9002 did not run a disarm-and-rearm cycle: %v", k2)
	}
	f1, _ := escalation.CardFilesFor(w1.Root, "t9001")
	f2, _ := escalation.CardFilesFor(w2.Root, "t9002")
	if !strings.HasSuffix(f1.Log, "/t9001.log.jsonl") && !strings.HasSuffix(f1.Log, `\t9001.log.jsonl`) ||
		!strings.HasSuffix(f2.Log, "t9002.log.jsonl") {
		t.Errorf("log names %s %s", f1.Log, f2.Log)
	}
	if f1.Log[:len(f1.Log)-len("t9001.log.jsonl")] != f2.Log[:len(f2.Log)-len("t9002.log.jsonl")] {
		t.Errorf("the two cards do not share one store: %s vs %s", f1.Log, f2.Log)
	}
}

// AC-AE-019 part 2: two concurrent PostToolUse hooks of one card append in
// sequence under the per-card lock.
func TestSameCardConcurrentHooksKeepChain(t *testing.T) {
	isolateStore(t)
	w := armedWith(t, "t9001", nil, "")
	files, _ := escalation.CardFilesFor(w.Root, w.Card)
	before, _, _ := escalation.ReadCardState(files.State)
	s := contractSettings(t, w)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			escalation.Observe(s, escalation.Event{Hook: escalation.HookPostToolUse, CWD: w.Root, ToolName: "Bash", Command: "ls"})
		}()
	}
	wg.Wait()
	lg := cardLog(t, w)
	if !lg.ChainIntact {
		t.Error("chain broken by concurrent hooks")
	}
	if len(disarmRecords(t, w)) != 0 {
		t.Error("concurrent hooks produced a disarm record")
	}
	after, _, _ := escalation.ReadCardState(files.State)
	if after.Counters.Operations != before.Counters.Operations+2 {
		t.Errorf("operations %d -> %d, want +2", before.Counters.Operations, after.Counters.Operations)
	}
	data, _ := os.ReadFile(files.State)
	st := entriesOf(lg, escalation.LineState)
	if st[len(st)-1].StateSHA256 != escalation.SHA256Hex(data) {
		t.Error("state file does not match the latest state entry")
	}
}
