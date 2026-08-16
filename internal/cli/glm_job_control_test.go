package cli

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// glm job-control tests — glm_job_status / glm_job_result / glm_job_cancel.
// Mirrors the SPEC-CODEX-PHASE2-001 M4 job-control tests (codex_job_control_test.go),
// minus the turn-id/process branches the GLM HTTP backend has no counterpart
// for: cancel revokes the job's context (asserted via the live-map join and
// the record's cancelled status), it signals no pid.

// callGLMJobTool invokes one of the three glm job-control handlers.
func callGLMJobTool(t *testing.T, fn func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := fn(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	if err != nil {
		t.Fatalf("glm job-control handler returned a Go error (the tool must return structured results only): %v", err)
	}
	if res == nil {
		t.Fatal("glm job-control handler returned a nil result")
	}
	return res
}

// withShortGLMCancelGrace shortens the cancel grace window so a test that must
// let it expire does not spend the distributed default waiting.
func withShortGLMCancelGrace(t *testing.T, d time.Duration) {
	t.Helper()
	prev := glmJobCancelGrace
	glmJobCancelGrace = d
	t.Cleanup(func() { glmJobCancelGrace = prev })
}

// ─── glm_job_status ───

// TestGLMJobStatus_ReturnsRecordAndNotFound proves glm_job_status returns the
// job's record for a known id and a structured not-found result — IsError set,
// not a Go error — for an unknown one, plus the missing-argument case.
func TestGLMJobStatus_ReturnsRecordAndNotFound(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newGLMJobRegistry(root)

	rec, err := reg.create(glmJobSpec{Model: "glm-5.3", RequestSummary: "summarize the diff"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	res := callGLMJobTool(t, handleGLMJobStatus, map[string]any{"job_id": rec.ID})
	if res.IsError {
		t.Fatalf("known job id must not produce an error result: %+v", res)
	}
	got := structuredMap(t, res)
	for field, want := range map[string]string{
		"id":              rec.ID,
		"status":          glmJobStatusQueued,
		"model":           "glm-5.3",
		"request_summary": "summarize the diff",
	} {
		if v, _ := got[field].(string); v != want {
			t.Errorf("record field %s = %q, want %q", field, v, want)
		}
	}

	unknown := callGLMJobTool(t, handleGLMJobStatus, map[string]any{"job_id": "job-does-not-exist"})
	if !unknown.IsError {
		t.Error("an unknown job id must return a structured not-found result with IsError set")
	}
	if txt := resultText(unknown); !containsAny(txt, "job-does-not-exist") {
		t.Errorf("not-found result must name the id it could not find; got %q", txt)
	}

	missing := callGLMJobTool(t, handleGLMJobStatus, map[string]any{})
	if !missing.IsError {
		t.Error("a missing job_id argument must return a structured error result")
	}
}

// TestGLMJobStatus_MalformedRecordIsReported covers the §B edge case: a record
// file that is present but malformed is reported as unreadable rather than
// crashing the server.
func TestGLMJobStatus_MalformedRecordIsReported(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newGLMJobRegistry(root)
	rec, err := reg.create(glmJobSpec{Model: "glm-5.3", RequestSummary: "x"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	writeFileForTest(t, reg.pathFor(rec.ID), "{ this is not json")

	res := callGLMJobTool(t, handleGLMJobStatus, map[string]any{"job_id": rec.ID})
	if !res.IsError {
		t.Fatal("a malformed record must be reported as an error result, not decoded as an empty record")
	}
}

// ─── glm_job_result ───

// TestGLMJobResult_TerminalReturnsOutputRunningReturnsStatus proves
// glm_job_result returns the recorded output for a terminal job and the
// current status — without waiting for the call — for a running one.
func TestGLMJobResult_TerminalReturnsOutputRunningReturnsStatus(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newGLMJobRegistry(root)

	done, err := reg.create(glmJobSpec{Model: "glm-5.3", RequestSummary: "finished job"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(done.ID, func(r *GLMJobRecord) {
		r.Status = glmJobStatusCompleted
		r.Output = "the summary is ready"
	}); err != nil {
		t.Fatalf("update job: %v", err)
	}

	res := callGLMJobTool(t, handleGLMJobResult, map[string]any{"job_id": done.ID})
	if res.IsError {
		t.Fatalf("terminal job must not produce an error result: %+v", res)
	}
	got := structuredMap(t, res)
	if out, _ := got["output"].(string); out != "the summary is ready" {
		t.Errorf("output = %q, want the recorded task output", out)
	}
	if term, _ := got["terminal"].(bool); !term {
		t.Error("a completed job must be reported as terminal")
	}

	running, err := reg.create(glmJobSpec{Model: "glm-5.3", RequestSummary: "in flight"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(running.ID, func(r *GLMJobRecord) { r.Status = glmJobStatusRunning }); err != nil {
		t.Fatalf("update job: %v", err)
	}

	start := time.Now()
	res = callGLMJobTool(t, handleGLMJobResult, map[string]any{"job_id": running.ID})
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("a non-terminal job must be reported without waiting for the call; it took %s", elapsed)
	}
	got = structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusRunning {
		t.Errorf("status = %q, want %q", st, glmJobStatusRunning)
	}
	if term, _ := got["terminal"].(bool); term {
		t.Error("a running job must not be reported as terminal")
	}
	if out, _ := got["output"].(string); out != "" {
		t.Errorf("a running job must carry no output; got %q", out)
	}

	unknown := callGLMJobTool(t, handleGLMJobResult, map[string]any{"job_id": "job-nope"})
	if !unknown.IsError {
		t.Error("an unknown job id must return a structured not-found result with IsError set")
	}
}

// ─── glm_job_cancel ───

// TestGLMJobCancel_CancelsRunningJob proves cancel records the job cancelled,
// revokes the in-flight call's context, and reports the outcome — all within a
// bounded time. The context-aware blocking doer mirrors a real http.Client:
// revoking the context ends the in-flight request, so the goroutine observes
// the cancellation and ends inside the grace window.
func TestGLMJobCancel_CancelsRunningJob(t *testing.T) {
	root := t.TempDir()
	jobID, reg := startHangingGLMJob(t, root)
	withShortGLMCancelGrace(t, 5*time.Second) // ample: the revoked context ends the call promptly

	start := time.Now()
	res := callGLMJobTool(t, handleGLMJobCancel, map[string]any{"job_id": jobID})
	elapsed := time.Since(start)

	if res.IsError {
		t.Fatalf("cancelling a running job this server owns must not error: %+v", res)
	}
	if elapsed > 5*time.Second {
		t.Errorf("cancel must return within a bounded time; it took %s", elapsed)
	}

	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusCancelled {
		t.Errorf("result status = %q, want %q", st, glmJobStatusCancelled)
	}
	if sent, _ := got["cancel_sent"].(bool); !sent {
		t.Error("result must report that the cancellation was sent")
	}

	waitForGLMJobToStop(t, jobID)
	rec, err := reg.load(jobID)
	if err != nil {
		t.Fatalf("load record: %v", err)
	}
	if rec.Status != glmJobStatusCancelled {
		t.Errorf("recorded status = %q, want %q", rec.Status, glmJobStatusCancelled)
	}
}

// TestGLMJobCancel_NoRecordWriteAfterCancel pins the registry-side guard: once
// a job is recorded cancelled, its goroutine must make NO further write to the
// record. updated_at is the witness — it is refreshed on EVERY registry write,
// so an unchanged value after the goroutine stops is proof no write occurred.
func TestGLMJobCancel_NoRecordWriteAfterCancel(t *testing.T) {
	root := t.TempDir()
	jobID, reg := startHangingGLMJob(t, root)
	withShortGLMCancelGrace(t, 60*time.Millisecond)

	res := callGLMJobTool(t, handleGLMJobCancel, map[string]any{"job_id": jobID})
	if res.IsError {
		t.Fatalf("cancel must succeed: %+v", res)
	}
	atCancel, err := reg.load(jobID)
	if err != nil {
		t.Fatalf("load record right after cancel: %v", err)
	}

	// Let the goroutine run its terminal transition — the write this test
	// forbids.
	waitForGLMJobToStop(t, jobID)

	after, err := reg.load(jobID)
	if err != nil {
		t.Fatalf("load record after the goroutine stopped: %v", err)
	}
	if after.Status != glmJobStatusCancelled {
		t.Errorf("status = %q, want %q — a late return must not overwrite a cancel", after.Status, glmJobStatusCancelled)
	}
	if !after.UpdatedAt.Equal(atCancel.UpdatedAt) {
		t.Errorf("updated_at moved from %s to %s: a record write landed AFTER glm_job_cancel returned; "+
			"a job already recorded cancelled must produce no further write",
			atCancel.UpdatedAt.Format(time.RFC3339Nano), after.UpdatedAt.Format(time.RFC3339Nano))
	}
}

// TestGLMJobCancel_RefusesStaleRecord proves a record naming a running job the
// running server did not start — the shape a record left behind by a previous
// server lifetime takes — is refused rather than cancelled.
func TestGLMJobCancel_RefusesStaleRecord(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newGLMJobRegistry(root)

	// A stale record: running, with a model and a summary — but no live entry,
	// because this server lifetime never started it.
	stale, err := reg.create(glmJobSpec{Model: "glm-5.3", RequestSummary: "left behind by a previous server lifetime"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(stale.ID, func(r *GLMJobRecord) { r.Status = glmJobStatusRunning }); err != nil {
		t.Fatalf("update job: %v", err)
	}
	if _, live := glmLiveJobs.Load(stale.ID); live {
		t.Fatal("precondition: no live entry may exist for a stale record")
	}

	res := callGLMJobTool(t, handleGLMJobCancel, map[string]any{"job_id": stale.ID})
	if !res.IsError {
		t.Fatalf("a stale record must be refused with a structured result: %+v", res)
	}
	if txt := resultText(res); !containsAny(txt, stale.ID) {
		t.Errorf("the refusal must name the job it refused; got %q", txt)
	}
	if rec, err := reg.load(stale.ID); err != nil || rec.Status != glmJobStatusRunning {
		t.Errorf("a refused cancel must leave the record untouched; got %q (err %v)", rec.Status, err)
	}
}

// TestGLMJobCancel_TerminalJobSendsNothing covers the §B edge case: cancel on
// an already-terminal job returns the terminal status and sends nothing.
func TestGLMJobCancel_TerminalJobSendsNothing(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newGLMJobRegistry(root)

	rec, err := reg.create(glmJobSpec{Model: "glm-5.3", RequestSummary: "already done"})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := reg.update(rec.ID, func(r *GLMJobRecord) {
		r.Status = glmJobStatusCompleted
		r.Output = "done"
	}); err != nil {
		t.Fatalf("update job: %v", err)
	}

	res := callGLMJobTool(t, handleGLMJobCancel, map[string]any{"job_id": rec.ID})
	if res.IsError {
		t.Fatalf("cancelling a terminal job is a no-op, not an error: %+v", res)
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusCompleted {
		t.Errorf("status = %q, want the terminal status %q left untouched", st, glmJobStatusCompleted)
	}
	if sent, _ := got["cancel_sent"].(bool); sent {
		t.Error("a terminal job must not have a cancellation sent for it")
	}

	unknown := callGLMJobTool(t, handleGLMJobCancel, map[string]any{"job_id": "job-nope"})
	if !unknown.IsError {
		t.Error("an unknown job id must return a structured not-found result with IsError set")
	}
	missing := callGLMJobTool(t, handleGLMJobCancel, map[string]any{})
	if !missing.IsError {
		t.Error("a missing job_id argument must return a structured error result")
	}
}
