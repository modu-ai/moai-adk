package escalation_test

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// m3Edit rewrites the fixture draft to the M3 ownership shape:
// write [internal/foo/**, .moai/specs/<ID>/**], never [internal/foo/secret/**],
// scratch [tmp-build/**].
func m3Edit(draft string) string {
	draft = strings.Replace(draft, `    - "internal/fixture/**"`, `    - "internal/foo/**"`, 1)
	draft = strings.Replace(draft, `    - "internal/x/**"`, `    - "internal/foo/secret/**"`, 1)
	return strings.Replace(draft, `    - ".moai/state/verify/**"`, `    - "tmp-build/**"`, 1)
}

// armedM3 returns an armed contract-mode worktree with the M3 ownership shape
// and any further draft edit applied.
func armedM3(t *testing.T, card string, extra func(string) string) (*escalationtest.Worktree, config.AutonomySettings) {
	t.Helper()
	w := escalationtest.NewWorktree(t, card)
	w.WriteContractMode("")
	edit := m3Edit
	if extra != nil {
		edit = func(d string) string { return extra(m3Edit(d)) }
	}
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Edit: edit})
	s := contractSettings(t, w)
	preWrite(s, w, "internal/foo/a.go")
	if !cardLog(t, w).Armed() {
		t.Fatalf("card did not arm: %+v", cardLog(t, w).Entries)
	}
	if got := records(t, w); len(got) != 0 {
		t.Fatalf("arming wrote records: %v", got)
	}
	return w, s
}

// refLine returns the contract line a contract.yaml:<n> reference names.
func refLine(t *testing.T, w *escalationtest.Worktree, ref string) string {
	t.Helper()
	n, err := strconv.Atoi(strings.TrimPrefix(ref, "contract.yaml:"))
	if err != nil || !strings.HasPrefix(ref, "contract.yaml:") {
		t.Fatalf("contract_ref %q is not contract.yaml:<line>", ref)
	}
	return lineAt(t, w.Read(".moai/specs/SPEC-A-001/contract.yaml"), n)
}

// notObservedEntries returns the log's not-observed entries.
func notObservedEntries(t *testing.T, w *escalationtest.Worktree) []escalation.LogEntry {
	t.Helper()
	return entriesOf(cardLog(t, w), escalation.LineNotObserved)
}

// bash observes one Bash call at PostToolUse.
func bash(s config.AutonomySettings, w *escalationtest.Worktree, cmd string, failed bool) {
	escalation.Observe(s, escalation.Event{Hook: escalation.HookPostToolUse, CWD: w.Root,
		ToolName: "Bash", Command: cmd, Failed: failed})
}

// AC-AE-007 (REQ-AE-006, REQ-AE-022): an invariant command that exits
// non-zero trips invariant-violation (command); the same command exiting 0 or
// another command exiting 1 writes none; the checkpoint lists the
// constitution: entry and the unexecuted command entry under not_observed.
func TestInvariantCommandFailureTrips(t *testing.T) {
	isolateStore(t)
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteContractMode("")
	w.WriteRegistry(escalationtest.RegistryEntry{ID: "CONST-V3R2-001", Zone: "Evolvable", File: "CLAUDE.md"})
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Edit: func(d string) string {
		return strings.Replace(d, `  - "go test ./..."`,
			"  - \"go test ./internal/spec/...\"\n  - \"constitution:CONST-V3R2-001\"\n  - \"make lint\"", 1)
	}})
	s := contractSettings(t, w)
	preWrite(s, w, "internal/fixture/a.go")
	if !cardLog(t, w).Armed() {
		t.Fatalf("card did not arm: %+v", cardLog(t, w).Entries)
	}

	bash(s, w, "go test ./internal/spec/...", false) // exit 0: no trip
	bash(s, w, "go test ./internal/other/...", true) // other command: no trip
	if got := records(t, w); len(got) != 0 {
		t.Fatalf("non-tripping calls wrote records: %v", got)
	}
	bash(s, w, "go test ./internal/spec/...", true)
	commitCheckpoint(s, w)

	rs, raw := recordsOfClass(t, w, escalation.ClassInvariantViolation)
	if len(rs) != 1 || rs[0].EscalateOn != "invariant-violation" || !strings.Contains(raw[0], "command") {
		t.Fatalf("invariant-violation records = %+v", rs)
	}
	if l := refLine(t, w, rs[0].ContractRef); !strings.Contains(l, "go test ./internal/spec/...") {
		t.Errorf("contract_ref points at %q", l)
	}
	no := notObservedEntries(t, w)
	if len(no) == 0 {
		t.Fatal("checkpoint wrote no not-observed entry")
	}
	last := no[len(no)-1].NotObserved
	if !slices.Contains(last, "constitution:CONST-V3R2-001") || !slices.Contains(last, "make lint") ||
		slices.Contains(last, "go test ./internal/spec/...") {
		t.Errorf("checkpoint not_observed = %v, want the constitution entry and the unexecuted command only", last)
	}
}

// AC-AE-009 (REQ-AE-008, REQ-AE-012): writes outside ownership.write, inside
// ownership.never, and to the signed contract.yaml and acceptance.md each trip
// ownership-move; a never match names the never line, the post-signing pair
// names effective_never; an owned write trips nothing.
func TestOwnershipMoveTrips(t *testing.T) {
	isolateStore(t)
	w, s := armedM3(t, "t9001", nil)
	preWrite(s, w, "internal/foo/z.go")
	if got := records(t, w); len(got) != 0 {
		t.Fatalf("owned write tripped: %v", got)
	}
	targets := []string{
		"internal/bar/x.go",
		"internal/foo/secret/y.go",
		".moai/specs/SPEC-A-001/contract.yaml",
		".moai/specs/SPEC-A-001/acceptance.md",
	}
	for _, rel := range targets {
		preWrite(s, w, rel)
	}
	rs, raw := recordsOfClass(t, w, escalation.ClassOwnershipMove)
	if len(rs) != 4 {
		t.Fatalf("ownership-move records = %d (%v), want 4", len(rs), records(t, w))
	}
	byTarget := map[string]int{}
	for i := range rs {
		for _, rel := range targets {
			if strings.Contains(raw[i], " "+rel+" (") {
				byTarget[rel] = i
			}
		}
	}
	if len(byTarget) != 4 {
		t.Fatalf("records do not name each target: %v", byTarget)
	}
	if l := refLine(t, w, rs[byTarget["internal/foo/secret/y.go"]].ContractRef); !strings.Contains(l, "internal/foo/secret/**") {
		t.Errorf("never record points at %q", l)
	}
	if l := refLine(t, w, rs[byTarget["internal/bar/x.go"]].ContractRef); !strings.Contains(l, "write:") {
		t.Errorf("write-miss record points at %q", l)
	}
	for _, rel := range targets[2:] {
		i := byTarget[rel]
		if !strings.Contains(raw[i], "effective_never") || !strings.Contains(raw[i], "after signing") {
			t.Errorf("%s record does not name post-signing immutability:\n%s", rel, raw[i])
		}
	}
}

// AC-AE-010 (REQ-AE-013, REQ-AE-022): exempt writes write nothing; the two
// contract-store writes trip ownership-move before the OS-temp exemption; an
// outside-root write no root covers trips when every root is determined, and
// is listed not-observed (no record) when the scratchpad root is not, while a
// store write still trips and is not listed.
func TestOwnershipExemptionsAndOutsideRoot(t *testing.T) {
	isolateStore(t)
	cfgDir := t.TempDir()
	t.Setenv(config.EnvClaudeConfigDir, cfgDir)
	w, s := armedM3(t, "t9001", nil)
	scratchpad := t.TempDir()
	files, err := escalation.CardFilesFor(w.Root, w.Card)
	if err != nil {
		t.Fatal(err)
	}
	memory := filepath.Join(cfgDir, "projects", escalation.MemorySlug(w.Root), "memory", "note.md")
	outside := filepath.Join(filepath.VolumeName(w.Root)+string(filepath.Separator), "escalation-outside-root-t9001", "out.txt")

	write := func(path, scratch string) {
		if !filepath.IsAbs(path) {
			path = w.Path(path)
		}
		escalation.Observe(s, escalation.Event{Hook: escalation.HookPreToolUse, CWD: w.Root,
			ToolName: "Write", FilePath: path, ScratchpadDir: scratch})
	}
	for _, p := range []string{
		".moai/reports/t9001/verdict.md",
		".moai/state/x.json",
		filepath.Join(t.TempDir(), "tmp-file.txt"),
		filepath.Join(scratchpad, "notes.txt"),
		memory,
		"tmp-build/out.txt",
	} {
		write(p, scratchpad)
	}
	if got := records(t, w); len(got) != 0 {
		t.Fatalf("exempt writes tripped: %v", got)
	}

	write(files.State, scratchpad)
	write(files.Log, scratchpad)
	write(outside, scratchpad)
	rs, raw := recordsOfClass(t, w, escalation.ClassOwnershipMove)
	if len(rs) != 3 {
		t.Fatalf("ownership-move records = %d (%v), want 3", len(rs), records(t, w))
	}
	var store int
	for _, r := range raw {
		if strings.Contains(r, "contract store") {
			store++
		}
	}
	if store != 2 {
		t.Errorf("records naming the contract store = %d, want 2", store)
	}

	// Scratchpad root undeterminable: the outside write is not-observed, a
	// store write still trips (and is not listed).
	w2, s2 := armedM3(t, "t9002", nil)
	s = s2
	w = w2
	files2, _ := escalation.CardFilesFor(w2.Root, w2.Card)
	outside2 := filepath.Join(filepath.Dir(outside), "out2.txt")
	write(outside2, "")
	if got := records(t, w2); len(got) != 0 {
		t.Fatalf("outside write with an undetermined root tripped: %v", got)
	}
	no := notObservedEntries(t, w2)
	if len(no) != 1 || !strings.Contains(strings.Join(no[0].NotObserved, " "), "out2.txt") {
		t.Fatalf("not-observed entries = %+v, want one naming the outside write", no)
	}
	write(files2.State, "")
	if rs, _ := recordsOfClass(t, w2, escalation.ClassOwnershipMove); len(rs) != 1 {
		t.Errorf("store write with an undetermined root: %d ownership-move records, want 1", len(rs))
	}
	if n := len(notObservedEntries(t, w2)); n != 1 {
		t.Errorf("store write was listed not-observed (%d entries)", n)
	}
}

// AC-AE-024 (REQ-AE-022): when the arming snapshot's ownership globs cannot be
// read, a write outside ownership trips nothing and is listed not-observed,
// and the next checkpoint lists ownership as not observed.
func TestUnreadableContractFieldIsNotObserved(t *testing.T) {
	isolateStore(t)
	w, s := armedM3(t, "t9001", nil)
	files, _ := escalation.CardFilesFor(w.Root, w.Card)
	data, err := os.ReadFile(files.State)
	if err != nil {
		t.Fatal(err)
	}
	// Replace the write glob list with an undecodable value and record the
	// rewrite in the log exactly as the detector does, so the state stays
	// consistent with its latest state entry (this is not a tamper case).
	st := string(data)
	start := strings.Index(st, `"write": [`)
	end := strings.Index(st[start:], "]")
	if start < 0 || end < 0 {
		t.Fatalf("state file has no write list:\n%s", st)
	}
	st = st[:start] + `"write": null` + st[start+end+1:]
	if err := os.WriteFile(files.State, []byte(st), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := escalation.AppendLog(files.Log, escalation.LogEntry{Kind: escalation.LineState, Card: w.Card,
		StateSHA256: escalation.SHA256Hex([]byte(st))}, fixedNow); err != nil {
		t.Fatal(err)
	}

	preWrite(s, w, "internal/bar/x.go")
	if got := records(t, w); len(got) != 0 {
		t.Fatalf("unreadable ownership still tripped: %v", got)
	}
	no := notObservedEntries(t, w)
	if len(no) != 1 || !slices.Contains(no[0].NotObserved, "ownership") {
		t.Fatalf("not-observed entries after the write = %+v", no)
	}
	commitCheckpoint(s, w)
	no = notObservedEntries(t, w)
	if last := no[len(no)-1]; !slices.Contains(last.NotObserved, "ownership") {
		t.Errorf("checkpoint not_observed = %v, want ownership listed", last.NotObserved)
	}
}
