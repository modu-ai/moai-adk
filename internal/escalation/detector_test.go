package escalation_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// isolateStore points MOAI_HOME at a per-test directory, so the contract
// store (and every card state file and log) lives under the OS temporary
// directory and never touches the operator's store.
func isolateStore(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv(config.EnvHome, home)
	return home
}

// cardLog reads the card audit log of the worktree's card.
func cardLog(t *testing.T, w *escalationtest.Worktree) escalation.CardLog {
	t.Helper()
	files, err := escalation.CardFilesFor(w.Root, w.Card)
	if err != nil {
		t.Fatalf("CardFilesFor: %v", err)
	}
	lg, err := escalation.ReadCardLog(files.Log)
	if err != nil {
		t.Fatalf("ReadCardLog: %v", err)
	}
	return lg
}

// entriesOf returns the log entries of one kind.
func entriesOf(lg escalation.CardLog, kind string) []escalation.LogEntry {
	var out []escalation.LogEntry
	for _, e := range lg.Entries {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// records lists the card's escalation record files by base name.
func records(t *testing.T, w *escalationtest.Worktree) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(escalation.RecordDir(w.Root, w.Card), "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range paths {
		names = append(names, filepath.Base(p))
	}
	return names
}

// recordsOfClass returns the parsed records of one class and their raw bytes.
func recordsOfClass(t *testing.T, w *escalationtest.Worktree, class string) ([]escalation.Record, []string) {
	t.Helper()
	var rs []escalation.Record
	var raw []string
	for _, name := range records(t, w) {
		if !strings.HasPrefix(name, class+"-") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(escalation.RecordDir(w.Root, w.Card), name))
		if err != nil {
			t.Fatal(err)
		}
		r, err := escalation.ParseRecord(data)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		rs = append(rs, r)
		raw = append(raw, string(data))
	}
	return rs, raw
}

// preWrite runs a PreToolUse Write observation for a worktree-relative path.
func preWrite(s config.AutonomySettings, w *escalationtest.Worktree, rel string) {
	escalation.Observe(s, escalation.Event{Hook: escalation.HookPreToolUse, CWD: w.Root,
		ToolName: "Write", FilePath: w.Path(rel)})
}

// commitCheckpoint runs the PostToolUse observation of a successful commit.
func commitCheckpoint(s config.AutonomySettings, w *escalationtest.Worktree) {
	escalation.Observe(s, escalation.Event{Hook: escalation.HookPostToolUse, CWD: w.Root,
		ToolName: "Bash", Command: "git commit -m wip"})
}

// armedWorktree returns a contract-mode worktree whose card is armed against
// SPEC-A-001 by a first PreToolUse observation inside ownership.write.
func armedWorktree(t *testing.T, card string) (*escalationtest.Worktree, config.AutonomySettings) {
	t.Helper()
	w := escalationtest.NewWorktree(t, card)
	w.WriteContractMode("")
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	s := contractSettings(t, w)
	preWrite(s, w, "internal/fixture/a.go")
	if !cardLog(t, w).Armed() {
		t.Fatalf("card %s did not arm on first observation; log %+v", card, cardLog(t, w).Entries)
	}
	if got := records(t, w); len(got) != 0 {
		t.Fatalf("arming wrote records: %v", got)
	}
	return w, s
}

// AC-AE-005 (REQ-AE-004): an unreadable resolved contract on a never-armed
// card is a fault — the call proceeds, the card log gains a not-checked line
// naming the fault, and no escalation record is written.
func TestFaultIsNotChecked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits do not deny reads on Windows")
	}
	isolateStore(t)
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteContractMode("")
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	p := w.Path(".moai/specs/SPEC-A-001/contract.yaml")
	if err := os.Chmod(p, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0o644) })
	if f, err := os.Open(p); err == nil {
		_ = f.Close()
		t.Skip("running with privileges that ignore file modes")
	}

	preWrite(contractSettings(t, w), w, "internal/bar/x.go") // returns: the call proceeds

	lg := cardLog(t, w)
	nc := entriesOf(lg, escalation.LineNotChecked)
	if len(nc) != 1 {
		t.Fatalf("not-checked lines = %d, want 1: %+v", len(nc), lg.Entries)
	}
	if !strings.Contains(nc[0].Detail, "permission denied") || nc[0].Class == "" {
		t.Errorf("not-checked line does not name the class and the fault: %+v", nc[0])
	}
	if lg.Armed() {
		t.Error("card armed despite the fault")
	}
	if got := records(t, w); len(got) != 0 {
		t.Errorf("fault wrote records: %v", got)
	}
}

// acceptanceReasons returns the acceptance-class reasons verify reports now.
func acceptanceReasons(t *testing.T, w *escalationtest.Worktree) []string {
	t.Helper()
	in, err := contract.LoadDir(w.Path(".moai/specs/SPEC-A-001"))
	if err != nil {
		t.Fatal(err)
	}
	in.Policy = escalationtest.Policy()
	in.SpecStatus = "in-progress"
	var out []string
	for _, r := range contract.Verify(in).Reasons {
		if slices.Contains(escalation.AcceptanceReasons, r) {
			out = append(out, r)
		}
	}
	return out
}

// AC-AE-006 (REQ-AE-005): an acceptance.md change after signing, seen at the
// commit checkpoint with contract.yaml bytes unchanged, writes one
// acceptance-change record citing recorded and measured values and every
// acceptance reason, plus exactly one detection-disarmed record with
// contract_ref disarm:signature-invalid. An unchanged file writes neither; a
// plan-phase edit to an unsigned contract's SPEC writes no record at all.
func TestAcceptanceChangeTrips(t *testing.T) {
	accRel := ".moai/specs/SPEC-A-001/acceptance.md"
	cases := []struct {
		name   string
		mutate func(w *escalationtest.Worktree)
	}{
		{"one-byte", func(w *escalationtest.Worktree) { w.Replace(accRel, "then it passes.", "then it passed.") }},
		{"ac-count", func(w *escalationtest.Worktree) { w.Write(accRel, signtest.AcceptanceThree) }},
		{"ambiguous", func(w *escalationtest.Worktree) { w.Write(accRel, signtest.AcceptanceAmbiguous) }},
		{"deleted", func(w *escalationtest.Worktree) {
			if err := os.Remove(w.Path(accRel)); err != nil {
				t.Fatal(err)
			}
		}},
	}
	sawSignatureAcceptance := false
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateStore(t)
			w, s := armedWorktree(t, "t9001")
			contractBefore := w.Read(".moai/specs/SPEC-A-001/contract.yaml")
			c, err := contract.Decode(contractBefore)
			if err != nil {
				t.Fatal(err)
			}
			recordedSHA := *c.Acceptance.SHA256
			recordedCount := *c.Acceptance.ACCount

			tc.mutate(w)
			want := acceptanceReasons(t, w)
			if len(want) == 0 {
				t.Fatal("fixture produced no acceptance reason")
			}
			if slices.Contains(want, "signature_acceptance_mismatch") {
				sawSignatureAcceptance = true
			}
			commitCheckpoint(s, w)

			if got := string(w.Read(".moai/specs/SPEC-A-001/contract.yaml")); got != string(contractBefore) {
				t.Fatal("contract.yaml bytes changed")
			}
			if n := len(records(t, w)); n != 2 {
				t.Fatalf("records = %v, want exactly one acceptance-change and one detection-disarmed", records(t, w))
			}
			acc, accRaw := recordsOfClass(t, w, escalation.ClassAcceptanceChange)
			if len(acc) != 1 {
				t.Fatalf("acceptance-change records = %d", len(acc))
			}
			if acc[0].Kind != escalation.KindContract || acc[0].EscalateOn != "acceptance-change" ||
				!strings.HasPrefix(acc[0].ContractRef, "contract.yaml:") {
				t.Errorf("acceptance-change frontmatter = %+v", acc[0])
			}
			for _, needle := range append([]string{recordedSHA, strconv.Itoa(recordedCount)}, want...) {
				if !strings.Contains(accRaw[0], needle) {
					t.Errorf("acceptance-change record does not cite %q:\n%s", needle, accRaw[0])
				}
			}
			dis, _ := recordsOfClass(t, w, escalation.ClassDetectionDisarmed)
			if len(dis) != 1 || dis[0].ContractRef != "disarm:signature-invalid" || dis[0].Kind != escalation.KindOperational {
				t.Fatalf("detection-disarmed records = %+v, want one with disarm:signature-invalid", dis)
			}
			lg := cardLog(t, w)
			if d := entriesOf(lg, escalation.LineDisarmed); len(d) != 1 || d[0].Reason != "signature-invalid" {
				t.Errorf("disarmed entries = %+v, want exactly one signature-invalid", d)
			}
			if lg.Armed() {
				t.Error("card still armed after the acceptance change")
			}
		})
	}
	if !sawSignatureAcceptance {
		t.Error("no fixture made verify report signature_acceptance_mismatch")
	}

	t.Run("unchanged", func(t *testing.T) {
		isolateStore(t)
		w, s := armedWorktree(t, "t9001")
		commitCheckpoint(s, w)
		if got := records(t, w); len(got) != 0 {
			t.Errorf("unchanged acceptance wrote records: %v", got)
		}
		if !cardLog(t, w).Armed() {
			t.Error("unchanged acceptance disarmed the card")
		}
	})

	t.Run("plan-phase-unsigned", func(t *testing.T) {
		isolateStore(t)
		w := escalationtest.NewWorktree(t, "t9001")
		w.WriteContractMode("")
		w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Unsigned: true})
		s := contractSettings(t, w)
		preWrite(s, w, "internal/fixture/a.go")
		w.Write(accRel, signtest.AcceptanceThree)
		commitCheckpoint(s, w)
		if got := records(t, w); len(got) != 0 {
			t.Errorf("plan-phase edit wrote records: %v", got)
		}
		if cardLog(t, w).Armed() {
			t.Error("unsigned contract armed")
		}
	})
}

// REQ-AE-023 arming record: the card state file carries the arming snapshot,
// and the log's latest state entry carries the state file's digest.
func TestArmingWritesStateAndChainedLog(t *testing.T) {
	isolateStore(t)
	w, _ := armedWorktree(t, "t9001")
	files, err := escalation.CardFilesFor(w.Root, w.Card)
	if err != nil {
		t.Fatal(err)
	}
	st, ok, err := escalation.ReadCardState(files.State)
	if err != nil || !ok || st.Armed == nil {
		t.Fatalf("state = %+v ok=%v err=%v, want an arming snapshot", st, ok, err)
	}
	a := st.Armed
	if a.SpecID != "SPEC-A-001" || a.Card != "t9001" || len(a.ContractSHA256) != 64 || len(a.ContractDigest) != 64 ||
		a.Budget.Operations != 40 || !slices.Contains(a.EffectiveNever, ".moai/specs/SPEC-A-001/contract.yaml") ||
		!slices.Contains(a.Write, "internal/fixture/**") || !slices.Contains(a.Scratch, ".moai/state/verify/**") ||
		!slices.Contains(a.FrozenFiles, "**/CLAUDE.md") {
		t.Errorf("arming snapshot = %+v", a)
	}
	lg := cardLog(t, w)
	if !lg.ChainIntact {
		t.Error("card log chain broken after arming")
	}
	armed := entriesOf(lg, escalation.LineArmed)
	if len(armed) != 1 || armed[0].Spec != "SPEC-A-001" || armed[0].ContractSHA256 != a.ContractSHA256 {
		t.Errorf("armed entries = %+v", armed)
	}
	states := entriesOf(lg, escalation.LineState)
	data, err := os.ReadFile(files.State)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) == 0 || states[len(states)-1].StateSHA256 != escalation.SHA256Hex(data) {
		t.Errorf("latest state entry does not carry the state file digest")
	}
	if filepath.Base(files.Log) != "t9001.log.jsonl" || filepath.Base(files.State) != "t9001.json" {
		t.Errorf("card file names = %s, %s", files.Log, files.State)
	}

	// A second observation of an armed card appends nothing new for arming.
	preWrite(contractSettings(t, w), w, "internal/fixture/b.go")
	if n := len(entriesOf(cardLog(t, w), escalation.LineArmed)); n != 1 {
		t.Errorf("armed entries after a second hook = %d, want 1", n)
	}
}
