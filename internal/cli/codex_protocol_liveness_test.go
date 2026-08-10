package cli

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-CODEX-PHASE2-001 M6 — protocol liveness. Verifies AC-CX2-017 (a
// client-bound request is ANSWERED, and a file-change approval is answered with
// a denial) and AC-CX2-018 (a codex_task turn is bounded by a deadline the tool
// imposes, not by the caller's context).
//
// Both criteria carry a deliberate anti-vacuity device, and both are
// implemented here as written:
//
//   - AC-CX2-017's canned server WITHHOLDS turn/completed until it observes a
//     transmitted line whose id equals the request's id. A driver that drops
//     the request therefore never gets the turn to complete, and the test fails
//     by exhausting its own bound instead of passing silently.
//   - AC-CX2-018's bound is TWO-SIDED: the call must return after at least the
//     configured duration (so an unrelated early return fails) and within a
//     generous ceiling (so a hang fails).
//
// The assertions are stated against the TRANSMITTED LINE and its decoded
// decision value — never against the presence of a case in a dispatch switch,
// which a code-shape check could satisfy without a response ever being sent.

// ─── fixtures: canned conns that model the two live behaviors ───

// codexHandshakeLines is the canned ack prefix every session below replays:
// initialize (id 1), thread/start (id 2), turn/start (id 3), then the
// turn/started notification that carries the turn id.
func codexHandshakeLines(turnID string) []string {
	q := jsonString
	return []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":` + q(turnID) + `,"status":"inProgress"}}}`,
		`{"method":"turn/started","params":{"threadId":"tid-fake","turn":{"id":` + q(turnID) + `,"status":"inProgress"}}}`,
	}
}

// gatedCodexConn replays `pre`, then BLOCKS until a line carrying `wantID` is
// transmitted back, and only then releases `post`. It is the canned server
// AC-CX2-017 describes: the release is what makes the criterion falsifiable.
type gatedCodexConn struct {
	mu   sync.Mutex
	sent []string
	idx  int

	pre    []string
	post   []string
	wantID string

	released  chan struct{}
	releaseOn sync.Once
	closed    chan struct{}
	closeOn   sync.Once
}

func newGatedCodexConn(pre, post []string, wantID string) *gatedCodexConn {
	return &gatedCodexConn{
		pre: pre, post: post, wantID: wantID,
		released: make(chan struct{}),
		closed:   make(chan struct{}),
	}
}

func (c *gatedCodexConn) send(line string) error {
	c.mu.Lock()
	c.sent = append(c.sent, line)
	c.mu.Unlock()
	if codexLineIsResponseTo(line, c.wantID) {
		c.releaseOn.Do(func() { close(c.released) })
	}
	return nil
}

func (c *gatedCodexConn) recv() (string, bool) {
	c.mu.Lock()
	if c.idx < len(c.pre) {
		line := c.pre[c.idx]
		c.idx++
		c.mu.Unlock()
		return line, true
	}
	c.mu.Unlock()

	select {
	case <-c.released:
	case <-c.closed:
		return "", false
	}

	c.mu.Lock()
	if i := c.idx - len(c.pre); i < len(c.post) {
		line := c.post[i]
		c.idx++
		c.mu.Unlock()
		return line, true
	}
	c.mu.Unlock()

	<-c.closed
	return "", false
}

func (c *gatedCodexConn) close() error {
	c.closeOn.Do(func() { close(c.closed) })
	return nil
}

func (c *gatedCodexConn) sentLines() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.sent...)
}

// stallingCodexConn replays `pre` and then emits NOTHING further — no
// turn/completed and no EOF, which is exactly the shape a real codex holds when
// it is parked waiting on an approval. recv unblocks only on close, so the only
// thing that can end a turn on this conn is a bound the tool imposes itself.
type stallingCodexConn struct {
	mu   sync.Mutex
	sent []string
	idx  int

	pre     []string
	closed  chan struct{}
	closeOn sync.Once
}

func newStallingCodexConn(pre []string) *stallingCodexConn {
	return &stallingCodexConn{pre: pre, closed: make(chan struct{})}
}

func (c *stallingCodexConn) send(line string) error {
	c.mu.Lock()
	c.sent = append(c.sent, line)
	c.mu.Unlock()
	return nil
}

func (c *stallingCodexConn) recv() (string, bool) {
	c.mu.Lock()
	if c.idx < len(c.pre) {
		line := c.pre[c.idx]
		c.idx++
		c.mu.Unlock()
		return line, true
	}
	c.mu.Unlock()

	<-c.closed
	return "", false
}

func (c *stallingCodexConn) close() error {
	c.closeOn.Do(func() { close(c.closed) })
	return nil
}

// fixedConnSessionRunner hands out one prepared conn. Unlike fakeCodexSession it
// does not own the transcript, because these conns are stateful.
type fixedConnSessionRunner struct{ conn codexConn }

func (r fixedConnSessionRunner) start(context.Context, string, []string) (codexConn, error) {
	return r.conn, nil
}

// withCodexConn points the session + binary-resolution seams at conn.
func withCodexConn(t *testing.T, conn codexConn) {
	t.Helper()
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexSession = fixedConnSessionRunner{conn: conn}
	t.Cleanup(func() { codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess })
}

// withCodexTaskTimeout shortens the REQ-CX2-017 bound for the duration of a
// test. It is the reason the requirement asks for a var rather than a const: a
// compile-time constant would make AC-CX2-018 a ten-minute test.
func withCodexTaskTimeout(t *testing.T, d time.Duration) {
	t.Helper()
	prev := config.DefaultCodexTaskTimeout
	config.DefaultCodexTaskTimeout = d
	t.Cleanup(func() { config.DefaultCodexTaskTimeout = prev })
}

// codexLineIsResponseTo reports whether line is a JSON-RPC RESPONSE (an id and
// no method) whose id is wantID, compared as raw JSON text so `0` and `"0"` are
// not conflated.
func codexLineIsResponseTo(line, wantID string) bool {
	var m struct {
		ID     json.RawMessage `json:"id"`
		Method string          `json:"method"`
	}
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return false
	}
	if m.Method != "" || len(m.ID) == 0 {
		return false
	}
	return strings.TrimSpace(string(m.ID)) == wantID
}

// findCodexResponseTo returns the single transmitted response line carrying
// wantID, failing the test when none was transmitted.
func findCodexResponseTo(t *testing.T, sent []string, wantID string) map[string]any {
	t.Helper()
	for _, line := range sent {
		if !codexLineIsResponseTo(line, wantID) {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("transmitted response for id %s is not valid JSON: %v (%s)", wantID, err, line)
		}
		return m
	}
	t.Fatalf("no JSON-RPC response carrying id %s was transmitted; sent lines:\n%s",
		wantID, strings.Join(sent, "\n"))
	return nil
}

// ─── AC-CX2-017 — client-bound requests are answered, approvals are denied ───

// TestCodexTask_DeniesFileChangeApprovalRequest is AC-CX2-017's main arm. The
// canned server withholds turn/completed until it sees an id-matched response,
// so this test can only pass if the driver actually transmitted one.
func TestCodexTask_DeniesFileChangeApprovalRequest(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	// No workflow.yaml at all — the distributed default, so the turn did not opt
	// into writes and the answer must be a denial.
	withCodexTaskTimeout(t, 5*time.Second)

	pre := append(codexHandshakeLines("trn-approval"),
		`{"method":"item/fileChange/requestApproval","id":0,"params":{"threadId":"tid-fake","turnId":"trn-approval","itemId":"exec-1","startedAtMs":1786384004385,"reason":null,"grantRoot":null}}`)
	post := []string{
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn-approval","item":{"type":"agentMessage","id":"m1","text":"I could not write the file."}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn-approval","status":"completed"}}}`,
	}
	conn := newGatedCodexConn(pre, post, "0")
	withCodexConn(t, conn)

	res := callCodexTask(t, map[string]any{"prompt": "edit a file"})
	out := structuredMap(t, res)

	if got := out["status"]; got != codexJobStatusCompleted {
		t.Fatalf("the turn did not return cleanly: status=%v result=%v", got, out)
	}

	resp := findCodexResponseTo(t, conn.sentLines(), "0")
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("approval response carries no result object: %v", resp)
	}
	decision, _ := result["decision"].(string)
	if decision != codexApprovalDecisionDecline {
		t.Fatalf("the approval was not denied: decision=%q (want %q); full response %v",
			decision, codexApprovalDecisionDecline, resp)
	}
	if resp["jsonrpc"] != "2.0" {
		t.Errorf("response is not a JSON-RPC 2.0 envelope: %v", resp)
	}
}

// TestCodexTask_AnswersUnrecognizedClientRequest is AC-CX2-017's unknown-method
// arm. The request deliberately carries NO threadId, so a driver that filters
// by thread before answering would drop it and stall.
func TestCodexTask_AnswersUnrecognizedClientRequest(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexTaskTimeout(t, 5*time.Second)

	pre := append(codexHandshakeLines("trn-unknown"),
		`{"method":"item/somethingNobodyImplemented","id":42,"params":{"whatever":true}}`)
	post := []string{
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn-unknown","item":{"type":"agentMessage","id":"m1","text":"done"}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn-unknown","status":"completed"}}}`,
	}
	conn := newGatedCodexConn(pre, post, "42")
	withCodexConn(t, conn)

	res := callCodexTask(t, map[string]any{"prompt": "do something"})
	out := structuredMap(t, res)

	if got := out["status"]; got != codexJobStatusCompleted {
		t.Fatalf("the turn did not return cleanly: status=%v result=%v", got, out)
	}

	resp := findCodexResponseTo(t, conn.sentLines(), "42")
	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("an unrecognized request must be answered with a JSON-RPC error arm; got %v", resp)
	}
	if _, hasCode := errObj["code"]; !hasCode {
		t.Errorf("JSON-RPC error arm carries no code: %v", errObj)
	}
	if msg, _ := errObj["message"].(string); msg == "" {
		t.Errorf("JSON-RPC error arm carries no message: %v", errObj)
	}
}

// TestCodexReviewGate_DeniesFileChangeApprovalRequest covers the shared-driver
// blast radius: awaitCodexTurnReview is the review gate's loop too, that path
// never opts into writes, and its answer must likewise be a denial — with its
// (ReviewOutput, error) contract untouched (C7).
func TestCodexReviewGate_DeniesFileChangeApprovalRequest(t *testing.T) {
	pre := append(codexHandshakeLines("trn-review"),
		`{"method":"item/fileChange/requestApproval","id":3001,"params":{"threadId":"tid-fake","turnId":"trn-review","itemId":"exec-9","startedAtMs":1}}`)
	post := []string{
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn-review","item":{"type":"exitedReviewMode","id":"e1","review":` + jsonString("no findings") + `}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn-review","status":"completed"}}}`,
	}
	conn := newGatedCodexConn(pre, post, "3001")
	withCodexConn(t, conn)

	// REQ-CX2-017 binds codex_task only — the review path keeps its caller's
	// bound (C7) — so this test supplies its own. Without it a driver that stops
	// answering would make this test HANG rather than fail, which is the one
	// outcome a regression test must not have. Closing the conn is what unblocks
	// the reader: the driver blocks in recv, not on the context.
	type reviewOutcome struct {
		out ReviewOutput
		err error
	}
	done := make(chan reviewOutcome, 1)
	go func() {
		out, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart,
			map[string]any{"target": "uncommittedChanges"})
		done <- reviewOutcome{out: out, err: err}
	}()

	var got reviewOutcome
	select {
	case got = <-done:
	case <-time.After(10 * time.Second):
		_ = conn.close() // release the reader so the goroutine exits
		<-done
		t.Fatal("the review turn never completed: the approval request was not answered, so the canned server withheld turn/completed")
	}
	if got.err != nil {
		t.Fatalf("review turn did not complete: %v (summary %q)", got.err, got.out.Summary)
	}
	if got.out.Verdict == "" {
		t.Fatalf("review returned an empty verdict: %+v", got.out)
	}

	resp := findCodexResponseTo(t, conn.sentLines(), "3001")
	result, _ := resp["result"].(map[string]any)
	if decision, _ := result["decision"].(string); decision != codexApprovalDecisionDecline {
		t.Fatalf("the review path did not deny the approval: %v", resp)
	}
}

// ─── AC-CX2-018 — the task turn is bounded by the tool's own deadline ───

// TestCodexTask_ForegroundTurnBoundedByOwnDeadline is AC-CX2-018's foreground
// arm. context.Background() carries no deadline, so nothing but a bound the
// tool imposes itself can end this turn — and the two-sided time assertion is
// what stops an unrelated early return from passing.
func TestCodexTask_ForegroundTurnBoundedByOwnDeadline(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)

	const bound = 300 * time.Millisecond
	withCodexTaskTimeout(t, bound)
	withCodexConn(t, newStallingCodexConn(codexHandshakeLines("trn-stall")))

	start := time.Now()
	res := callCodexTask(t, map[string]any{"prompt": "hang forever"})
	elapsed := time.Since(start)

	if elapsed < bound {
		t.Fatalf("returned in %v, before the %v bound could have fired — the pass would not be attributable to the bound", elapsed, bound)
	}
	if ceiling := 30 * time.Second; elapsed > ceiling {
		t.Fatalf("returned in %v, past the %v ceiling — the turn was not bounded", elapsed, ceiling)
	}

	out := structuredMap(t, res)
	if got := out["status"]; got != codexJobStatusFailed {
		t.Fatalf("a timed-out turn must be reported as failed; got status=%v (%v)", got, out)
	}
	errText, _ := out["error"].(string)
	if !strings.Contains(strings.ToLower(errText), "timed out") {
		t.Fatalf("the result does not name the timeout: error=%q", errText)
	}
}

// TestCodexTask_BackgroundTurnTimeoutReachesTerminalStatus is AC-CX2-018's
// background arm: the job record must reach a terminal status on expiry rather
// than sitting at running forever.
func TestCodexTask_BackgroundTurnTimeoutReachesTerminalStatus(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)

	const bound = 300 * time.Millisecond
	withCodexTaskTimeout(t, bound)
	withCodexConn(t, newStallingCodexConn(codexHandshakeLines("trn-bg-stall")))

	start := time.Now()
	res := callCodexTask(t, map[string]any{"prompt": "hang forever", "background": true})
	out := structuredMap(t, res)
	jobID, _ := out["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background task returned no job id: %v", out)
	}

	registry := newCodexJobRegistry(root)
	var rec CodexJobRecord
	found := false
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		r, err := registry.read(jobID)
		if err == nil && r.Status != codexJobStatusRunning && r.Status != codexJobStatusQueued {
			rec, found = r, true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	elapsed := time.Since(start)

	if !found {
		t.Fatalf("job %s never reached a terminal status within the ceiling", jobID)
	}
	if elapsed < bound {
		t.Fatalf("job went terminal in %v, before the %v bound could have fired", elapsed, bound)
	}
	if rec.Status != codexJobStatusFailed {
		t.Fatalf("a timed-out background job must be failed; got %q", rec.Status)
	}
	if !strings.Contains(strings.ToLower(rec.Error), "timed out") {
		t.Fatalf("the job record does not name the timeout: error=%q", rec.Error)
	}
}

// TestCodexTaskTimeout_IsDistinctFromReviewGateBudget is AC-CX2-018's
// distinctness arm, held as a test rather than only as a one-off grep so a
// later edit that collapses the two budgets fails here (plan.md §F AP-8).
func TestCodexTaskTimeout_IsDistinctFromReviewGateBudget(t *testing.T) {
	src, err := os.ReadFile("codex_task.go")
	if err != nil {
		t.Fatalf("read codex_task.go: %v", err)
	}
	if strings.Contains(string(src), "DefaultCodexReviewGateTimeout") {
		t.Fatal("codex_task.go references DefaultCodexReviewGateTimeout: the task bound must be its own named value, not the review gate's budget")
	}
	if config.DefaultCodexTaskTimeout == config.DefaultCodexReviewGateTimeout {
		t.Errorf("the task bound and the review-gate budget hold the same value (%v); they are meant to be tuned independently",
			config.DefaultCodexTaskTimeout)
	}
}
