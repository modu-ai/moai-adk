package cli

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// card t1186 — codex_task background=true failed ~7ms after creation with
// "codex turn/start rejected: codex stdout closed before response to id=3".
//
// t514 detached the job GOROUTINE from the request context, but the SESSION was
// opened before that, on the request context itself. The production runner
// spawns codex with exec.CommandContext(ctx, …) and its readLoop exits on
// ctx.Done(), so the MCP host ending the request (the moment the background
// handler returns) killed the child and closed stdout. The job goroutine then
// sent turn/start (id=3: initialize=1, thread=2) into a dead session.
//
// TestCodexTask_BackgroundJobSurvivesRequestContextEnd did not catch it because
// fakeCodexSession ignores the context it is started with — a fake STRONGER than
// the process it models. ctxBoundCodexSession is the faithful model: its conn
// dies when the context used to OPEN it ends, exactly as exec.CommandContext +
// readLoop do.

// ctxBoundCodexSession starts ctxBoundCodexConns. lines[:gateAt] are delivered
// freely; lines[gateAt:] wait for release(), so a test can end the request
// context between the handshake and the turn deterministically.
type ctxBoundCodexSession struct {
	lines  []string
	gateAt int

	mu      sync.Mutex
	conn    *ctxBoundCodexConn
	openCtx context.Context

	releaseCh   chan struct{}
	releaseOnce sync.Once

	onStart func() // optional hook run inside start, after the ctx tie
}

func (s *ctxBoundCodexSession) start(ctx context.Context, _ string, _ []string) (codexConn, error) {
	c := &ctxBoundCodexConn{s: s, killed: make(chan struct{})}
	// exec.CommandContext kills the child when ctx ends; readLoop stops too.
	context.AfterFunc(ctx, c.kill)
	s.mu.Lock()
	s.conn, s.openCtx = c, ctx
	s.mu.Unlock()
	if s.onStart != nil {
		s.onStart()
	}
	return c, nil
}

func (s *ctxBoundCodexSession) release() { s.releaseOnce.Do(func() { close(s.releaseCh) }) }

func (s *ctxBoundCodexSession) current() (*ctxBoundCodexConn, context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn, s.openCtx
}

type ctxBoundCodexConn struct {
	s        *ctxBoundCodexSession
	idx      int // single reader goroutine
	killed   chan struct{}
	killOnce sync.Once
}

func (c *ctxBoundCodexConn) kill() { c.killOnce.Do(func() { close(c.killed) }) }

func (c *ctxBoundCodexConn) isKilled() bool {
	select {
	case <-c.killed:
		return true
	default:
		return false
	}
}

// send always lands: the live failure wrote turn/start into the pipe before the
// kill was observed, then read EOF waiting for its response.
func (c *ctxBoundCodexConn) send(string) error { return nil }

func (c *ctxBoundCodexConn) recv() (string, bool) {
	if c.isKilled() {
		return "", false // stdout closed
	}
	if c.idx >= len(c.s.lines) {
		<-c.killed // transcript exhausted: a live codex just stays silent
		return "", false
	}
	if c.idx >= c.s.gateAt {
		select {
		case <-c.s.releaseCh:
		case <-c.killed:
			return "", false
		}
		if c.isKilled() {
			return "", false
		}
	}
	l := c.s.lines[c.idx]
	c.idx++
	return l, true
}

func (c *ctxBoundCodexConn) close() error { c.kill(); return nil }
func (c *ctxBoundCodexConn) pid() int     { return fakeCodexConnPID }

// waitKilled reports whether the conn died within d.
func (c *ctxBoundCodexConn) waitKilled(d time.Duration) bool {
	select {
	case <-c.killed:
		return true
	case <-time.After(d):
		return false
	}
}

func withCtxBoundCodexSession(t *testing.T, lines []string, gateAt int) *ctxBoundCodexSession {
	t.Helper()
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	s := &ctxBoundCodexSession{lines: lines, gateAt: gateAt, releaseCh: make(chan struct{})}
	codexSession = s
	t.Cleanup(func() {
		s.release()
		if c, _ := s.current(); c != nil {
			c.kill()
		}
		codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess
	})
	return s
}

// startCtxBoundBackgroundJob drives handleCodexTask with background=true and a
// request context the caller ends; it returns the job id and the cancel func.
func startCtxBoundBackgroundJob(t *testing.T) (string, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	res, err := handleCodexTask(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"prompt": "audit the module", "background": true}},
	})
	if err != nil {
		t.Fatalf("handleCodexTask returned a Go error: %v", err)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background result must carry a job id: %+v", structuredMap(t, res))
	}
	t.Cleanup(func() { waitForJobToStop(t, jobID) })
	return jobID, cancel
}

// ─── RED reproduction ───

func TestCodexTask_BackgroundSessionOutlivesRequestContext(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	// Handshake (ids 1, 2) flows freely; everything from the turn/start ack on
	// waits until the request context has already ended.
	sess := withCtxBoundCodexSession(t, codexTaskScript("trn-bg-t1186", "background work done"), 2)

	jobID, cancel := startCtxBoundBackgroundJob(t)

	// What the MCP host does once the handler has returned.
	cancel()
	conn, _ := sess.current()
	conn.waitKilled(200 * time.Millisecond) // a request-bound session dies here
	sess.release()

	rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
	if rec.Status != codexJobStatusCompleted {
		t.Fatalf("background job ended %q with error %q; ending the request must not kill the session",
			rec.Status, rec.Error)
	}
	if rec.Output != "background work done" {
		t.Errorf("recorded output = %q, want the task output", rec.Output)
	}
}

// ─── guards the detached session must keep ───

// awaitRunningJob waits until the job is recorded running with turnID.
func awaitRunningJob(t *testing.T, reg *codexJobRegistry, jobID, turnID string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if rec, err := reg.load(jobID); err == nil && rec.Status == codexJobStatusRunning && rec.TurnID == turnID {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached running with turn id %q", jobID, turnID)
}

// assertSessionEnded checks the child is gone AND the session context was
// cancelled, so nothing the session spawned outlives the job.
func assertSessionEnded(t *testing.T, sess *ctxBoundCodexSession) {
	t.Helper()
	conn, openCtx := sess.current()
	if !conn.waitKilled(3 * time.Second) {
		t.Error("the codex session (child) was never torn down")
	}
	deadline := time.Now().Add(3 * time.Second)
	for openCtx.Err() == nil && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if openCtx.Err() == nil {
		t.Error("the session context was never cancelled; the child would stay bound to nothing")
	}
}

// (a) codex_job_cancel still terminates a running background job's child.
func TestCodexTask_BackgroundDetachedSessionIsCancellable(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	lines := hangingTaskScript("trn-bg-cancel")
	sess := withCtxBoundCodexSession(t, lines, len(lines))
	withShortCancelGrace(t, 60*time.Millisecond)
	prevTerm := codexTerminateProcess
	var (
		termMu     sync.Mutex
		terminated []int
	)
	codexTerminateProcess = func(pid int) error { // kills the fake child, never a real pid
		termMu.Lock()
		terminated = append(terminated, pid)
		termMu.Unlock()
		if c, _ := sess.current(); c != nil {
			c.kill()
		}
		return nil
	}
	t.Cleanup(func() { codexTerminateProcess = prevTerm })

	jobID, cancel := startCtxBoundBackgroundJob(t)
	cancel() // the host ends the request; the job must still be running
	reg := newCodexJobRegistry(root)
	awaitRunningJob(t, reg, jobID, "trn-bg-cancel")
	if c, _ := sess.current(); c.isKilled() {
		t.Fatal("the session died with the request context; nothing is left to cancel")
	}

	res := callCodexJobTool(t, handleCodexJobCancel, map[string]any{"job_id": jobID})
	if res.IsError {
		t.Fatalf("cancel returned IsError: %s", resultText(res))
	}
	if term, _ := structuredMap(t, res)["process_terminated"].(bool); !term {
		t.Errorf("cancel must terminate the child after the grace window: %+v", structuredMap(t, res))
	}
	termMu.Lock()
	gotPIDs := append([]int(nil), terminated...)
	termMu.Unlock()
	if len(gotPIDs) != 1 || gotPIDs[0] != fakeCodexConnPID {
		t.Errorf("termination targeted %v, want [%d]", gotPIDs, fakeCodexConnPID)
	}
	assertSessionEnded(t, sess)
	waitForJobToStop(t, jobID)
	if rec, _ := reg.load(jobID); rec.Status != codexJobStatusCancelled {
		t.Errorf("recorded status = %q, want %q", rec.Status, codexJobStatusCancelled)
	}
}

// (b) the codex_task turn bound still ends a hung background job.
func TestCodexTask_BackgroundDetachedSessionHonorsTurnBound(t *testing.T) {
	withCodexTaskTimeout(t, 120*time.Millisecond)
	root := t.TempDir()
	withCodexProjectDir(t, root)
	lines := hangingTaskScript("trn-bg-bound")
	sess := withCtxBoundCodexSession(t, lines, len(lines))

	jobID, cancel := startCtxBoundBackgroundJob(t)
	cancel()

	rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
	if rec.Status != codexJobStatusFailed || !strings.Contains(rec.Error, "timed out after") {
		t.Fatalf("a hung background job must end at the bound; got %q / %q", rec.Status, rec.Error)
	}
	assertSessionEnded(t, sess)
}

// (c) foreground stays tied to the request: cancelling it ends the child.
func TestCodexTask_ForegroundSessionStillBoundToRequest(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	sess := withCtxBoundCodexSession(t, codexTaskScript("trn-fg-cancel", "never"), 2)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan *mcp.CallToolResult, 1)
	go func() {
		res, _ := handleCodexTask(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]any{"prompt": "x"}},
		})
		done <- res
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()

	var res *mcp.CallToolResult
	select {
	case res = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("foreground codex_task did not return after its request was cancelled")
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != codexJobStatusFailed {
		t.Errorf("status = %q, want %q", st, codexJobStatusFailed)
	}
	if _, openCtx := sess.current(); openCtx.Err() == nil {
		t.Error("the foreground session must be opened on the request context")
	}
	assertSessionEnded(t, sess)
}

// (d) until hand-off the request still bounds the background handshake.
func TestCodexTask_BackgroundHandshakeStillBoundToRequest(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	sess := withCtxBoundCodexSession(t, codexTaskScript("trn-hs", "never"), 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sess.onStart = cancel // the request ends mid-handshake

	res, err := handleCodexTask(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"prompt": "x", "background": true}},
	})
	if err != nil {
		t.Fatalf("Go error: %v", err)
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != codexJobStatusFailed {
		t.Errorf("status = %q, want %q (a cancelled handshake must not start a job)", st, codexJobStatusFailed)
	}
	if id, _ := got["job_id"].(string); id != "" {
		t.Errorf("no job may be created when the request ends before hand-off; got %q", id)
	}
	assertSessionEnded(t, sess)
}
