package cli

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// SPEC-CODEX-PHASE2-001 M4 — job control (REQ-CX2-009, REQ-CX2-010,
// REQ-CX2-011, REQ-CX2-012). Verifies AC-CX2-012 through AC-CX2-015 plus the
// §B edge cases that bind the job-control tools (cancel on an already-terminal
// job, an unknown id, a malformed record).
//
// Two fixtures carry the weight here:
//
//   - hangingCodexSession replays a canned transcript and then BLOCKS instead of
//     signalling EOF, so a background job stays genuinely in flight while the
//     cancel path runs against it. The canned runner of M1-M3 ends its lines and
//     returns EOF, which completes the turn — a completed turn is exactly the
//     case cancellation does not exercise.
//
//   - the process-termination seam is recorded, never executed. fakeCodexConnPID
//     (424242) is an implausible pid by construction, and AC-CX2-015 asks for the
//     absence of a signal to be asserted "by a test that records signal attempts
//     rather than by observing a live process".

// ─── fixtures ───

// hangingCodexSession is a session runner whose connection replays `lines` and
// then blocks on recv until close (or unblock) is called. It records every sent
// request line under a mutex, because the cancel path sends turn/interrupt from
// the calling goroutine while the job goroutine is still reading.
type hangingCodexSession struct {
	lines []string

	mu   sync.Mutex
	sent []string

	block  chan struct{}
	unlock sync.Once
}

func newHangingCodexSession(lines []string) *hangingCodexSession {
	return &hangingCodexSession{lines: lines, block: make(chan struct{})}
}

func (s *hangingCodexSession) start(context.Context, string, []string) (codexConn, error) {
	return &hangingCodexConn{s: s}, nil
}

// unblock releases the blocked reader so the job goroutine can finish. Tests
// call it in cleanup: the fake termination seam records a kill rather than
// performing one, so nothing else would ever end the turn.
func (s *hangingCodexSession) unblock() { s.unlock.Do(func() { close(s.block) }) }

// sentLines returns a copy of the recorded request lines.
func (s *hangingCodexSession) sentLines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.sent...)
}

type hangingCodexConn struct {
	s   *hangingCodexSession
	idx int // read only by the single reader goroutine
}

func (c *hangingCodexConn) send(line string) error {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	c.s.sent = append(c.s.sent, line)
	return nil
}

func (c *hangingCodexConn) recv() (string, bool) {
	if c.idx < len(c.s.lines) {
		line := c.s.lines[c.idx]
		c.idx++
		return line, true
	}
	<-c.s.block
	return "", false
}

func (c *hangingCodexConn) close() error { c.s.unblock(); return nil }

// pid makes the hanging connection satisfy codexProcessConn, so the session
// reports a spawned-process pid without spawning a process.
func (c *hangingCodexConn) pid() int { return fakeCodexConnPID }

// withHangingCodexSession installs a session that replays lines and then blocks.
func withHangingCodexSession(t *testing.T, lines []string) *hangingCodexSession {
	t.Helper()
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	sess := newHangingCodexSession(lines)
	codexSession = sess
	t.Cleanup(func() {
		sess.unblock()
		codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess
	})
	return sess
}

// hangingTaskScript is the transcript of a task turn that starts and never
// completes: handshake acks, the turn/start ack, and (when turnID is non-empty)
// the turn/started notification the turnId is read from.
func hangingTaskScript(turnID string) []string {
	lines := []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn-ack","status":"inProgress"}}}`,
	}
	if turnID != "" {
		lines = append(lines,
			`{"method":"turn/started","params":{"threadId":"tid-fake","turn":{"id":`+jsonString(turnID)+`,"status":"inProgress"}}}`)
	}
	return lines
}

// withRecordedCodexTerminator swaps the process-termination seam for a recorder
// and returns a reader of the pids termination was attempted on. No signal is
// ever sent.
func withRecordedCodexTerminator(t *testing.T) func() []int {
	t.Helper()
	var (
		mu   sync.Mutex
		pids []int
	)
	prev := codexTerminateProcess
	codexTerminateProcess = func(pid int) error {
		mu.Lock()
		defer mu.Unlock()
		pids = append(pids, pid)
		return nil
	}
	t.Cleanup(func() { codexTerminateProcess = prev })
	return func() []int {
		mu.Lock()
		defer mu.Unlock()
		return append([]int(nil), pids...)
	}
}

// withShortCancelGrace shortens the cancel grace window so a test that must let
// it expire does not spend the distributed default waiting.
func withShortCancelGrace(t *testing.T, d time.Duration) {
	t.Helper()
	prev := codexJobCancelGrace
	codexJobCancelGrace = d
	t.Cleanup(func() { codexJobCancelGrace = prev })
}

// resultText concatenates the text content of a tool result. An IsError result
// carries its explanation there rather than in StructuredContent (see toolErr).
func resultText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

// writeFileForTest overwrites path with body.
func writeFileForTest(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// callCodexJobTool invokes one of the three job-control handlers.
func callCodexJobTool(t *testing.T, fn func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := fn(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	if err != nil {
		t.Fatalf("job-control handler returned a Go error (the tool must return structured results only): %v", err)
	}
	if res == nil {
		t.Fatal("job-control handler returned a nil result")
	}
	return res
}

// startHangingBackgroundJob starts a background codex_task whose turn never
// completes, and returns the job id plus the registry rooted at root.
func startHangingBackgroundJob(t *testing.T, root, turnID string) (string, *codexJobRegistry) {
	t.Helper()
	withCodexProjectDir(t, root)
	sess := withHangingCodexSession(t, hangingTaskScript(turnID))

	res := callCodexTask(t, map[string]any{"prompt": "refactor the parser", "background": true})
	if res.IsError {
		t.Fatalf("background codex_task returned IsError: %+v", res)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background codex_task returned no job id: %+v", res)
	}
	reg := newCodexJobRegistry(root)

	// This helper starts a goroutine (the background job), so it owns joining it.
	// Registered AFTER withHangingCodexSession so that LIFO cleanup order runs
	// this one FIRST: release the blocked turn, then wait for its goroutine to
	// stop before t.TempDir's RemoveAll runs. Without the join, RemoveAll races
	// the goroutine's own MkdirAll + record write and the cleanup fails with
	// "directory not empty" — a test that lets work outlive it (acceptance.md §B
	// "no goroutine leaks past the test").
	t.Cleanup(func() {
		sess.unblock()
		waitForJobToStop(t, jobID)
	})

	// The job must be observably running before a cancel is meaningful, and when
	// a turnId is expected it must have landed in the record first — the cancel
	// path reads it from there.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rec, err := reg.load(jobID)
		if err == nil && rec.Status == codexJobStatusRunning && (turnID == "" || rec.TurnID == turnID) {
			return jobID, reg
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached running (turn id %q) before the deadline", jobID, turnID)
	return "", nil
}

// waitForJobToStop blocks until the job has left the live-session map, which is
// the join point for the goroutine runCodexBackgroundJob runs in: that goroutine
// writes its terminal registry transition BEFORE the deferred Delete, so an
// absent entry means every write it will ever make has already landed.
func waitForJobToStop(t *testing.T, jobID string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, live := codexLiveJobSessions.Load(jobID); !live {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Errorf("job %s goroutine did not stop within the deadline; it would outlive the test", jobID)
}

// sentInterrupt returns the first recorded request line naming turn/interrupt.
func sentInterrupt(lines []string) (string, bool) {
	for _, l := range lines {
		if strings.Contains(l, codexMethodTurnInterrupt) {
			return l, true
		}
	}
	return "", false
}

// ─── AC-CX2-012 (REQ-CX2-009) — codex_job_status ───

// TestCodexJobStatus_ReturnsRecordAndNotFound proves codex_job_status returns
// the job's record for a known id and a structured not-found result — IsError
// set, not a Go error that aborts the call — for an unknown one.
func TestCodexJobStatus_ReturnsRecordAndNotFound(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)

	rec, err := reg.create(codexJobSpec{
		ThreadID:       "tid-fake",
		TurnID:         "trn-status",
		PID:            fakeCodexConnPID,
		Mode:           codexTaskMode,
		RequestSummary: "summarize the diff",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	res := callCodexJobTool(t, handleCodexJobStatus, map[string]any{"job_id": rec.ID})
	if res.IsError {
		t.Fatalf("known job id must not produce an error result: %+v", res)
	}
	got := structuredMap(t, res)
	for field, want := range map[string]string{
		"id":        rec.ID,
		"status":    codexJobStatusQueued,
		"thread_id": "tid-fake",
		"turn_id":   "trn-status",
		"mode":      codexTaskMode,
	} {
		if v, _ := got[field].(string); v != want {
			t.Errorf("record field %s = %q, want %q", field, v, want)
		}
	}
	if pid, _ := got["pid"].(float64); int(pid) != fakeCodexConnPID {
		t.Errorf("record pid = %v, want %d", got["pid"], fakeCodexConnPID)
	}

	unknown := callCodexJobTool(t, handleCodexJobStatus, map[string]any{"job_id": "job-does-not-exist"})
	if !unknown.IsError {
		t.Error("an unknown job id must return a structured not-found result with IsError set")
	}
	if txt := resultText(unknown); !strings.Contains(txt, "job-does-not-exist") {
		t.Errorf("not-found result must name the id it could not find; got %q", txt)
	}

	missing := callCodexJobTool(t, handleCodexJobStatus, map[string]any{})
	if !missing.IsError {
		t.Error("a missing job_id argument must return a structured error result")
	}
}

// TestCodexJobStatus_MalformedRecordIsReported covers acceptance.md §B: a record
// file that is present but malformed is reported as unreadable rather than
// crashing the server.
func TestCodexJobStatus_MalformedRecordIsReported(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	rec, err := reg.create(codexJobSpec{ThreadID: "tid-fake", Mode: codexTaskMode, RequestSummary: "x"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	writeFileForTest(t, reg.pathFor(rec.ID), "{ this is not json")

	res := callCodexJobTool(t, handleCodexJobStatus, map[string]any{"job_id": rec.ID})
	if !res.IsError {
		t.Fatal("a malformed record must be reported as an error result, not decoded as an empty record")
	}
}

// ─── AC-CX2-013 (REQ-CX2-010) — codex_job_result ───

// TestCodexJobResult_TerminalReturnsOutputRunningReturnsStatus proves
// codex_job_result returns the recorded output for a terminal job and the
// current status — without waiting for the turn — for a running one.
func TestCodexJobResult_TerminalReturnsOutputRunningReturnsStatus(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)

	done, err := reg.create(codexJobSpec{ThreadID: "tid-fake", Mode: codexTaskMode, RequestSummary: "finished job"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(done.ID, func(r *CodexJobRecord) {
		r.Status = codexJobStatusCompleted
		r.Output = "the refactor is complete"
	}); err != nil {
		t.Fatalf("update job: %v", err)
	}

	res := callCodexJobTool(t, handleCodexJobResult, map[string]any{"job_id": done.ID})
	if res.IsError {
		t.Fatalf("terminal job must not produce an error result: %+v", res)
	}
	got := structuredMap(t, res)
	if out, _ := got["output"].(string); out != "the refactor is complete" {
		t.Errorf("output = %q, want the recorded task output", out)
	}
	if term, _ := got["terminal"].(bool); !term {
		t.Error("a completed job must be reported as terminal")
	}

	running, err := reg.create(codexJobSpec{ThreadID: "tid-fake", Mode: codexTaskMode, RequestSummary: "in flight"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(running.ID, func(r *CodexJobRecord) { r.Status = codexJobStatusRunning }); err != nil {
		t.Fatalf("update job: %v", err)
	}

	start := time.Now()
	res = callCodexJobTool(t, handleCodexJobResult, map[string]any{"job_id": running.ID})
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("a non-terminal job must be reported without waiting for the turn; the call took %s", elapsed)
	}
	got = structuredMap(t, res)
	if st, _ := got["status"].(string); st != codexJobStatusRunning {
		t.Errorf("status = %q, want %q", st, codexJobStatusRunning)
	}
	if term, _ := got["terminal"].(bool); term {
		t.Error("a running job must not be reported as terminal")
	}
	if out, _ := got["output"].(string); out != "" {
		t.Errorf("a running job must carry no output; got %q", out)
	}

	unknown := callCodexJobTool(t, handleCodexJobResult, map[string]any{"job_id": "job-nope"})
	if !unknown.IsError {
		t.Error("an unknown job id must return a structured not-found result with IsError set")
	}
}

// ─── AC-CX2-014 (REQ-CX2-011) — codex_job_cancel ───

// TestCodexJobCancel_SendsInterruptAndTerminates proves the cancel path sends
// the M0-confirmed turn/interrupt carrying BOTH required params — the record's
// threadId and turnId — records the job cancelled, and terminates the process
// this server spawned when the turn does not end within the grace window, all
// within a bounded time.
func TestCodexJobCancel_SendsInterruptAndTerminates(t *testing.T) {
	root := t.TempDir()
	jobID, reg := startHangingBackgroundJob(t, root, "trn-cancel")
	sess, _ := codexSession.(*hangingCodexSession)
	if sess == nil {
		t.Fatal("expected the hanging session runner to be installed")
	}
	terminated := withRecordedCodexTerminator(t)
	withShortCancelGrace(t, 60*time.Millisecond)

	start := time.Now()
	res := callCodexJobTool(t, handleCodexJobCancel, map[string]any{"job_id": jobID})
	elapsed := time.Since(start)

	if res.IsError {
		t.Fatalf("cancelling a running job this server owns must not error: %+v", res)
	}
	if elapsed > 5*time.Second {
		t.Errorf("cancel must return within a bounded time; it took %s", elapsed)
	}

	line, ok := sentInterrupt(sess.sentLines())
	if !ok {
		t.Fatalf("no %s request was sent; sent lines: %v", codexMethodTurnInterrupt, sess.sentLines())
	}
	for _, want := range []string{`"threadId":"tid-fake"`, `"turnId":"trn-cancel"`} {
		if !strings.Contains(line, want) {
			t.Errorf("%s params must carry %s; got %s", codexMethodTurnInterrupt, want, line)
		}
	}

	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != codexJobStatusCancelled {
		t.Errorf("result status = %q, want %q", st, codexJobStatusCancelled)
	}
	if sent, _ := got["interrupt_sent"].(bool); !sent {
		t.Error("result must report that the interrupt was sent")
	}
	if term, _ := got["process_terminated"].(bool); !term {
		t.Error("result must report that the process was terminated after the grace window")
	}

	if pids := terminated(); len(pids) != 1 || pids[0] != fakeCodexConnPID {
		t.Errorf("termination targeted %v, want exactly [%d] — only the pid this server spawned for the job", pids, fakeCodexConnPID)
	}

	rec, err := reg.load(jobID)
	if err != nil {
		t.Fatalf("load record: %v", err)
	}
	if rec.Status != codexJobStatusCancelled {
		t.Errorf("recorded status = %q, want %q", rec.Status, codexJobStatusCancelled)
	}

	// The cancelled status must survive the turn returning afterwards: the fake
	// termination recorded a kill rather than performing one, so the goroutine is
	// still blocked until the transcript is released here.
	sess.unblock()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, live := codexLiveJobSessions.Load(jobID); !live {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	final, err := reg.load(jobID)
	if err != nil {
		t.Fatalf("load record after the turn returned: %v", err)
	}
	if final.Status != codexJobStatusCancelled {
		t.Errorf("status after the turn returned = %q, want %q — a late turn must not overwrite a cancel", final.Status, codexJobStatusCancelled)
	}
}

// TestCodexJobCancel_EmptyTurnIDDegradesToTermination covers the branch M2 and
// M3 both left open: when no turn/started was ever observed the record carries
// an empty turn_id, and turn/interrupt has no second argument to send. The
// cancel must degrade to process termination rather than sending a malformed
// interrupt or failing.
func TestCodexJobCancel_EmptyTurnIDDegradesToTermination(t *testing.T) {
	root := t.TempDir()
	jobID, reg := startHangingBackgroundJob(t, root, "")
	sess, _ := codexSession.(*hangingCodexSession)
	if sess == nil {
		t.Fatal("expected the hanging session runner to be installed")
	}
	terminated := withRecordedCodexTerminator(t)
	withShortCancelGrace(t, 60*time.Millisecond)

	if rec, err := reg.load(jobID); err != nil || rec.TurnID != "" {
		t.Fatalf("precondition: record must carry an empty turn_id; got %+v (err %v)", rec, err)
	}

	res := callCodexJobTool(t, handleCodexJobCancel, map[string]any{"job_id": jobID})
	if res.IsError {
		t.Fatalf("an unaddressable turn must still cancel: %+v", res)
	}
	if line, ok := sentInterrupt(sess.sentLines()); ok {
		t.Errorf("no %s may be sent without a turnId; sent %s", codexMethodTurnInterrupt, line)
	}

	got := structuredMap(t, res)
	if sent, _ := got["interrupt_sent"].(bool); sent {
		t.Error("result must not claim an interrupt was sent")
	}
	if note, _ := got["note"].(string); note == "" {
		t.Error("result must state why the turn could not be interrupted")
	}
	if term, _ := got["process_terminated"].(bool); !term {
		t.Error("the cancel must degrade to terminating the process this server spawned")
	}
	if pids := terminated(); len(pids) != 1 || pids[0] != fakeCodexConnPID {
		t.Errorf("termination targeted %v, want exactly [%d]", pids, fakeCodexConnPID)
	}
	if rec, err := reg.load(jobID); err != nil || rec.Status != codexJobStatusCancelled {
		t.Errorf("recorded status = %+v (err %v), want %q", rec.Status, err, codexJobStatusCancelled)
	}
}

// TestCodexJobCancel_NoRecordWriteAfterCancel pins the production half of the
// M4 re-verification defect: once a job is recorded cancelled, its goroutine
// must make NO further write to the record.
//
// The status guard in runCodexBackgroundJob prevented the terminal transition
// from overwriting the cancelled STATUS, but it lived inside the mutator — and a
// mutator cannot decline a write, because registry.update persists whatever the
// mutator leaves behind, refreshing updated_at even when nothing changed. So a
// write still landed after the cancel call had returned. That is observable to a
// caller as a record mutating after the tool said it was done, and under the
// in-process execution model the goroutine sits inside the long-lived
// mcp-server, so it is not a test artifact.
//
// updated_at is the witness: it is refreshed on EVERY registry write, so an
// unchanged value after the turn returns is proof no write occurred.
func TestCodexJobCancel_NoRecordWriteAfterCancel(t *testing.T) {
	root := t.TempDir()
	jobID, reg := startHangingBackgroundJob(t, root, "trn-nowrite")
	sess, _ := codexSession.(*hangingCodexSession)
	if sess == nil {
		t.Fatal("expected the hanging session runner to be installed")
	}
	withRecordedCodexTerminator(t)
	withShortCancelGrace(t, 60*time.Millisecond)

	res := callCodexJobTool(t, handleCodexJobCancel, map[string]any{"job_id": jobID})
	if res.IsError {
		t.Fatalf("cancel must succeed: %+v", res)
	}
	atCancel, err := reg.load(jobID)
	if err != nil {
		t.Fatalf("load record right after cancel: %v", err)
	}

	// Release the turn the cancel interrupted. Its goroutine now runs its
	// terminal transition — the write this test forbids.
	sess.unblock()
	waitForJobToStop(t, jobID)

	after, err := reg.load(jobID)
	if err != nil {
		t.Fatalf("load record after the turn returned: %v", err)
	}
	if after.Status != codexJobStatusCancelled {
		t.Errorf("status = %q, want %q — a late turn must not overwrite a cancel", after.Status, codexJobStatusCancelled)
	}
	if !after.UpdatedAt.Equal(atCancel.UpdatedAt) {
		t.Errorf("updated_at moved from %s to %s: a record write landed AFTER codex_job_cancel returned; "+
			"a job already recorded cancelled must produce no further write",
			atCancel.UpdatedAt.Format(time.RFC3339Nano), after.UpdatedAt.Format(time.RFC3339Nano))
	}
}

// ─── AC-CX2-015 (REQ-CX2-012) — ownership refusal ───

// TestCodexJobCancel_RefusesRecordThisServerDidNotSpawn proves a record naming a
// pid the running server did not spawn in the current process lifetime — the
// shape a record left behind by a previous server lifetime takes — is refused
// rather than signalled. The absence of a signal is asserted by recording
// termination attempts, never by observing a live process.
func TestCodexJobCancel_RefusesRecordThisServerDidNotSpawn(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	terminated := withRecordedCodexTerminator(t)

	// A stale record: running, with a thread, a turn, and a pid — but no live
	// session, because this server lifetime never started it.
	stale, err := reg.create(codexJobSpec{
		ThreadID:       "tid-stale",
		TurnID:         "trn-stale",
		PID:            999999,
		Mode:           codexTaskMode,
		RequestSummary: "left behind by a previous server lifetime",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(stale.ID, func(r *CodexJobRecord) { r.Status = codexJobStatusRunning }); err != nil {
		t.Fatalf("update job: %v", err)
	}
	if _, live := codexLiveJobSessions.Load(stale.ID); live {
		t.Fatal("precondition: no live session may exist for a stale record")
	}

	res := callCodexJobTool(t, handleCodexJobCancel, map[string]any{"job_id": stale.ID})
	if !res.IsError {
		t.Fatalf("a stale record must be refused with a structured result: %+v", res)
	}
	if pids := terminated(); len(pids) != 0 {
		t.Errorf("no process may be signalled for a record this server did not spawn; attempted %v", pids)
	}
	if txt := resultText(res); !strings.Contains(txt, stale.ID) {
		t.Errorf("the refusal must name the job it refused; got %q", txt)
	}
}

// TestCodexJobCancel_TerminalJobSendsNothing covers acceptance.md §B: cancel on
// an already-terminal job returns the terminal status and sends nothing.
func TestCodexJobCancel_TerminalJobSendsNothing(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	terminated := withRecordedCodexTerminator(t)

	rec, err := reg.create(codexJobSpec{ThreadID: "tid-fake", TurnID: "trn-done", Mode: codexTaskMode, RequestSummary: "already done"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(rec.ID, func(r *CodexJobRecord) {
		r.Status = codexJobStatusCompleted
		r.Output = "done"
	}); err != nil {
		t.Fatalf("update job: %v", err)
	}

	res := callCodexJobTool(t, handleCodexJobCancel, map[string]any{"job_id": rec.ID})
	if res.IsError {
		t.Fatalf("cancelling a terminal job is a no-op, not an error: %+v", res)
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != codexJobStatusCompleted {
		t.Errorf("status = %q, want the terminal status %q left untouched", st, codexJobStatusCompleted)
	}
	if sent, _ := got["interrupt_sent"].(bool); sent {
		t.Error("a terminal job must not have an interrupt sent for it")
	}
	if pids := terminated(); len(pids) != 0 {
		t.Errorf("a terminal job must not have a process signalled for it; attempted %v", pids)
	}

	unknown := callCodexJobTool(t, handleCodexJobCancel, map[string]any{"job_id": "job-nope"})
	if !unknown.IsError {
		t.Error("an unknown job id must return a structured not-found result with IsError set")
	}
	missing := callCodexJobTool(t, handleCodexJobCancel, map[string]any{})
	if !missing.IsError {
		t.Error("a missing job_id argument must return a structured error result")
	}
}
