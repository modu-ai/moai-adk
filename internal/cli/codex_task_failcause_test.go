package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// t514 / GH #1687 — codex_task reported an IMMEDIATE failure with the generic
// "turn timed out after 10m0s" wording.
//
// Two distinct defects sit behind that one symptom, and each gets its own
// reproduction here:
//
//   - Misreport: runCodexTaskTurn selected on the DERIVED context's Done
//     channel and named its own bound unconditionally, so a turn that ended
//     because the CALLER's context ended was reported as having exhausted a
//     10-minute bound it never approached.
//
//   - Root cause of the sub-second failure: a background job was handed the
//     request-scoped context, which the MCP host ends the moment the handler
//     returns — which for background=true is immediately. Every background job
//     therefore died within milliseconds of being created.
//
// Regression runs in BOTH directions (the card's [HARD]): a real bound expiry
// must still say timeout, and a caller-ended turn must say what actually ended
// it. A one-directional guard lets the other side come back silently.
//
// The stalling transport (newStallingCodexConn / withCodexConn) and the bound
// override (withCodexTaskTimeout) are the AC-CX2-018 fixtures, reused verbatim.

// openStalledTaskSession opens a session whose turn never completes.
func openStalledTaskSession(t *testing.T, turnID string) (*codexSessionHandle, map[string]any) {
	t.Helper()
	withCodexConn(t, newStallingCodexConn(codexHandshakeLines(turnID)))
	params := map[string]any{
		"prompt":        "stall",
		"cwd":           t.TempDir(),
		"sandboxPolicy": codexSandboxPolicy(false),
	}
	session, err := openCodexSessionOn(context.Background(), "/fake/codex", params, "")
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	t.Cleanup(func() { _ = session.close() })
	return session, params
}

// ─── direction 1: the caller's context ended — NOT this tool's bound ───

func TestCodexTaskTurn_CallerCancelIsNotReportedAsTimeout(t *testing.T) {
	withCodexTaskTimeout(t, 10*time.Minute)
	session, params := openStalledTaskSession(t, "trn-cancel")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	out, err := runCodexTaskTurn(ctx, session, params)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("an abandoned turn must return an error")
	}
	// The measurement that makes the misreport visible: the turn ended in
	// milliseconds, while the reported cause named a 10-minute bound.
	if elapsed > 30*time.Second {
		t.Fatalf("turn took %v; the cancellation path is not being exercised", elapsed)
	}
	t.Logf("turn ended after %v (bound was %v)", elapsed, config.DefaultCodexTaskTimeout)

	if strings.Contains(out.Summary, "timed out after") {
		t.Errorf("a caller-cancelled turn that ended in %v is reported as a bound expiry: %q", elapsed, out.Summary)
	}
	if !strings.Contains(out.Summary, "cancelled") {
		t.Errorf("summary must name the actual cause (caller cancellation); got %q", out.Summary)
	}
	if out.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q", out.Verdict, VerdictInconclusive)
	}
}

// ─── direction 2: this tool's own bound really did expire ───

func TestCodexTaskTurn_RealBoundExpiryStillReportsTimeout(t *testing.T) {
	withCodexTaskTimeout(t, 120*time.Millisecond)
	session, params := openStalledTaskSession(t, "trn-expire")

	start := time.Now()
	out, err := runCodexTaskTurn(context.Background(), session, params)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("an expired turn must return an error")
	}
	if elapsed < 100*time.Millisecond {
		t.Fatalf("turn ended after %v, before the %v bound could fire", elapsed, config.DefaultCodexTaskTimeout)
	}
	if !strings.Contains(out.Summary, "timed out after") {
		t.Errorf("a real bound expiry must still be reported as a timeout; got %q", out.Summary)
	}
	if !strings.Contains(out.Summary, config.DefaultCodexTaskTimeout.String()) {
		t.Errorf("the timeout message must name the bound that actually fired (%v); got %q",
			config.DefaultCodexTaskTimeout, out.Summary)
	}
}

// ─── root cause: a background job must outlive the request context ───

func TestCodexTask_BackgroundJobSurvivesRequestContextEnd(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexSession(t, codexTaskScript("trn-bg-detached", "background work done"))

	// The MCP host ends the request context when the handler returns — which,
	// for background=true, is immediately.
	ctx, cancel := context.WithCancel(context.Background())
	res, err := handleCodexTask(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"prompt": "audit the module", "background": true,
		}},
	})
	cancel()
	if err != nil {
		t.Fatalf("handleCodexTask returned a Go error: %v", err)
	}

	got := structuredMap(t, res)
	jobID, _ := got["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background result must carry a job id: %+v", got)
	}

	rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
	if rec.Status != codexJobStatusCompleted {
		t.Fatalf("background job ended %q with error %q; the request context ending must not kill it",
			rec.Status, rec.Error)
	}
	if rec.Output != "background work done" {
		t.Errorf("recorded output = %q, want the task output", rec.Output)
	}
}
