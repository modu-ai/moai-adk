package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLiveBudgetStopsBeforeExceeding pins the ABORTED boundary shared by the
// LIVE tests (REQ-DHR-023, AC-DHR-012): the call that would exceed the limit
// is refused before it starts, and a run that ends at exactly the limit is
// never reported as aborted.
func TestLiveBudgetStopsBeforeExceeding(t *testing.T) {
	now := time.Unix(1000, 0)
	b := newLiveBudget(3, 100*time.Second)
	b.now = func() time.Time { return now }
	b.start = now
	for i := 1; i <= 3; i++ {
		if ok, reason := b.take(); !ok {
			t.Fatalf("call %d refused within budget: %s", i, reason)
		}
	}
	if b.used != 3 || b.aborted() {
		t.Fatalf("exactly-at-budget run: used=%d aborted=%v, want used=3 and not aborted", b.used, b.aborted())
	}
	ok, reason := b.take()
	if ok || b.used != 3 || !strings.Contains(reason, "invocation") {
		t.Fatalf("4th call: ok=%v used=%d reason=%q, want refusal before the call", ok, b.used, reason)
	}
	if !b.aborted() {
		t.Fatal("a refused call must mark the run aborted")
	}
}

func TestLiveBudgetStopsAtTimeWindow(t *testing.T) {
	now := time.Unix(1000, 0)
	b := newLiveBudget(8, 900*time.Second)
	b.now = func() time.Time { return now }
	b.start = now
	if ok, _ := b.take(); !ok {
		t.Fatal("first call refused")
	}
	now = now.Add(900 * time.Second)
	ok, reason := b.take()
	if ok || !strings.Contains(reason, "time") || b.used != 1 {
		t.Fatalf("call at the window edge: ok=%v used=%d reason=%q", ok, b.used, reason)
	}
	if got := b.remaining(); got != 0 {
		t.Fatalf("remaining after window = %v, want 0", got)
	}
}

// TestLiveEvidenceTagLineMatchesFile pins the evidence channel of
// acceptance.md §A: the tag line carries the sha256 of the file bytes.
func TestLiveEvidenceTagLineMatchesFile(t *testing.T) {
	dir := t.TempDir()
	line, err := writeLiveEvidenceFile(dir, "ev.json", "AC018", "claude-claude", map[string]any{"case": "claude-claude"})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "ev.json"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	want := "AC018_EVIDENCE_SHA256 claude-claude " + hex.EncodeToString(sum[:])
	if line != want {
		t.Fatalf("tag line %q, want %q", line, want)
	}
	line, err = writeLiveEvidenceFile(dir, "ev2.json", "AC012", "", map[string]any{})
	if err != nil || !strings.HasPrefix(line, "AC012_EVIDENCE_SHA256 ") || len(line) != len("AC012_EVIDENCE_SHA256 ")+64 {
		t.Fatalf("caseless tag line %q err=%v", line, err)
	}
}

// TestLiveEvidenceDirSkipsWithNotRun pins the explicit NOT_RUN skip when the
// evidence channel is absent.
func TestLiveEvidenceDirSkipsWithNotRun(t *testing.T) {
	if got := liveEvidenceSkipReason(""); !strings.Contains(got, "NOT_RUN") || !strings.Contains(got, envT1100EvidenceDir) {
		t.Fatalf("empty dir skip reason %q must name NOT_RUN and %s", got, envT1100EvidenceDir)
	}
	if got := liveEvidenceSkipReason("../x"); got != "" {
		t.Fatalf("set dir must not skip, got %q", got)
	}
}

const syntheticParentRollout = `{"type":"session_meta","payload":{"id":"p1","source":"exec","cli_version":"0.156.1"}}
{"type":"response_item","payload":{"type":"function_call","name":"spawn_agent","call_id":"c0","arguments":"{\"agent_type\":\"plan-auditor\"}"}}
`

const syntheticSubagentRollout = `{"type":"session_meta","payload":{"id":"s1","agent_role":"plan-auditor","source":{"subagent":{"thread_spawn":{"parent_thread_id":"p1","depth":1,"agent_role":"plan-auditor"}}}}}
{"type":"turn_context","payload":{"sandbox_policy":{"type":"read-only"},"approval_policy":"never"}}
{"type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c1","input":"tools.exec_command({cmd:\"printf audit > audit-probe-plan-auditor.txt\"})"}}
{"type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"c1","output":[{"type":"input_text","text":"zsh: operation not permitted: audit-probe-plan-auditor.txt"}]}}
{"type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"c2","arguments":"{\"cmd\":\"ls\"}"}}
{"type":"response_item","payload":{"type":"function_call_output","call_id":"c2","output":"README.md"}}
{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"draft"}]}}
{"type":"event_msg","payload":{"type":"task_complete","last_agent_message":"VERDICT plan-auditor nonce=abc write=denied"}}
`

const syntheticGuardianRollout = `{"type":"session_meta","payload":{"id":"g1","source":{"subagent":{"other":"guardian"}}}}
{"type":"event_msg","payload":{"type":"task_complete","last_agent_message":"approved"}}
`

// TestParseCodexRolloutSubagent pins the rollout reading AC-DHR-012/023 rely
// on: the role comes from the thread_spawn session_meta, the sandbox from the
// last turn_context, the returned text from task_complete, and the attempt
// output from the tool output paired by call_id with an input naming the probe.
func TestParseCodexRolloutSubagent(t *testing.T) {
	sub := parseCodexRollout([]byte(syntheticSubagentRollout))
	if !sub.Subagent || sub.Role != "plan-auditor" || sub.Sandbox != "read-only" {
		t.Fatalf("subagent meta: %+v", sub)
	}
	if sub.Final != "VERDICT plan-auditor nonce=abc write=denied" {
		t.Fatalf("final %q", sub.Final)
	}
	got := sub.outputsMentioning("audit-probe-plan-auditor")
	if !strings.Contains(got, "operation not permitted") || strings.Contains(got, "README.md") {
		t.Fatalf("attempt output %q", got)
	}
	if !liveDenialPattern.MatchString(got) {
		t.Fatalf("denial pattern missed %q", got)
	}
	parent := parseCodexRollout([]byte(syntheticParentRollout))
	if parent.Subagent || parent.Role != "" {
		t.Fatalf("parent parsed as subagent: %+v", parent)
	}
	guard := parseCodexRollout([]byte(syntheticGuardianRollout))
	if !guard.Subagent || guard.Role != "" {
		t.Fatalf("guardian must be a roleless subagent: %+v", guard)
	}
	noFinal := parseCodexRollout([]byte(strings.Replace(syntheticSubagentRollout, `"last_agent_message":"VERDICT plan-auditor nonce=abc write=denied"`, `"last_agent_message":null`, 1)))
	if noFinal.Final != "draft" {
		t.Fatalf("fallback final %q, want last assistant message", noFinal.Final)
	}
}

// TestLiveProcsReapKillsTheGroup pins the cleanup contract: a tracked
// process still running at reap time is killed and reported as cleaned.
func TestLiveProcsReapKillsTheGroup(t *testing.T) {
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not available on this platform")
	}
	p := newLiveProcs(t)
	cmd := exec.Command(sleep, "30")
	if err := p.start(cmd); err != nil {
		t.Fatal(err)
	}
	pid := cmd.Process.Pid
	cleaned := p.reap()
	if len(cleaned) != 1 || cleaned[0] != pid {
		t.Fatalf("cleaned %v, want [%d]", cleaned, pid)
	}
	if !liveProcGroupGone(pid) {
		t.Fatalf("process group %d still alive after reap", pid)
	}
	if again := p.reap(); len(again) != 1 {
		t.Fatalf("reap must be idempotent, got %v", again)
	}
}
