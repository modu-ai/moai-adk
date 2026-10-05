package cli

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// t532 (t514 derivative) — glm_task's background jobs died within
// milliseconds of being created, every time. The probe that opened the card
// recorded status="failed" after 5.248208ms.
//
// The root cause is the one t514 fixed on the codex side: the background job
// was derived from the REQUEST-scoped context, which the MCP host ends the
// moment the handler returns — and for background=true the handler returns
// immediately.
//
// Two things had to be true for the existing suite to miss it, and both are
// repaired here rather than worked around:
//
//   - Every background test drives the handler through callGLMTaskTool, which
//     passes a context.Background() that is never cancelled. The host cancels.
//     startBackgroundGLMJobWithEndedRequest below does what the host does.
//
//   - stubGLMDoer never consults req.Context(), so it returns a canned success
//     even for a request whose context is already dead. That makes the stub
//     STRONGER than the thing it models: a real http.Client returns
//     context.Canceled there. ctxAwareGLMDoer is the faithful model, and
//     without it the reproduction below passes against the defect.
//
// Scope note: the codex one-line repair is NOT copied here. GLM stores a
// cancel function in glmLiveJobs and drives glm_job_cancel through it, so a
// bare context.WithoutCancel would silently sever cancellation. The repair
// must be WithCancel layered ON TOP of WithoutCancel, and the second test is
// the guard for exactly that — a wrong repair passes the first test and fails
// the second.

// ctxAwareGLMDoer returns a canned response, but only for a request whose
// context is still alive — which is what an http.Client does. The existing
// stubGLMDoer answers a dead request as readily as a live one.
type ctxAwareGLMDoer struct{ body string }

func (s *ctxAwareGLMDoer) Do(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

// startBackgroundGLMJobWithEndedRequest drives the background handler with a
// request context that ends the instant the handler returns, mirroring the MCP
// host, and hands back the job id.
func startBackgroundGLMJobWithEndedRequest(t *testing.T, prompt string) string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	res, err := handleGLMTask(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"prompt": prompt, "background": true,
		}},
	})
	cancel() // the host ends the request scope here
	if err != nil {
		t.Fatalf("handleGLMTask returned a Go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("background glm_task IsError: %+v", res)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background glm_task returned no job id: %+v", res)
	}
	return jobID
}

// ─── direction 1: the job must outlive the request context ───

func TestGLMTask_BackgroundJobSurvivesRequestContextEnd(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withGLMTaskSeams(t, "k", &ctxAwareGLMDoer{body: glmTextResp("background work done")})

	start := time.Now()
	jobID := startBackgroundGLMJobWithEndedRequest(t, "audit the queue")
	t.Cleanup(func() { waitForGLMJobToStop(t, jobID) })

	reg := newGLMJobRegistry(root)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rec, err := reg.load(jobID)
		if err == nil && rec.Status == glmJobStatusCompleted {
			if rec.Output != "background work done" {
				t.Errorf("recorded output = %q, want the task output", rec.Output)
			}
			return
		}
		// The defect's signature: a terminal FAILED record, reached in
		// milliseconds, naming the request context's end.
		if err == nil && rec.Status == glmJobStatusFailed {
			t.Fatalf("job ended %q after %v with error %q — the request context ending killed it",
				rec.Status, time.Since(start), rec.Error)
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached completed", jobID)
}

// ─── direction 2: cancellation must still reach the detached job ───

// TestGLMTask_DetachedBackgroundJobIsStillCancellable is the guard the repair
// shape demands. Detaching the job from the request context must not detach it
// from glm_job_cancel: the cancel function glmLiveJobs stores has to keep
// aborting the in-flight call. A repair that hands runGLMBackgroundJob a bare
// context.WithoutCancel passes the test above and fails this one.
func TestGLMTask_DetachedBackgroundJobIsStillCancellable(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	doer := newBlockingGLMDoer()
	withGLMTaskSeams(t, "k", doer)

	jobID := startBackgroundGLMJobWithEndedRequest(t, "a task that blocks")
	t.Cleanup(func() {
		doer.unblock()
		waitForGLMJobToStop(t, jobID)
	})

	reg := newGLMJobRegistry(root)

	// `running` is written by the HANDLER, before the goroutine starts, so
	// reading it proves nothing about survival — it is already on disk when
	// the request context ends. What proves survival is that the job is STILL
	// live after a settle window: against the defect the goroutine is gone,
	// with a terminal record, within milliseconds.
	settle := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(settle) {
		rec, err := reg.load(jobID)
		if err == nil && rec.Status != glmJobStatusRunning {
			t.Fatalf("job went terminal (%q, error %q) while its call was still blocked — the request context ending killed it",
				rec.Status, rec.Error)
		}
		if _, live := glmLiveJobs.Load(jobID); !live {
			t.Fatal("the job goroutine exited while its call was still blocked — the request context ending killed it")
		}
		time.Sleep(10 * time.Millisecond)
	}

	res := callGLMJobTool(t, handleGLMJobCancel, map[string]any{"job_id": jobID})
	if res.IsError {
		t.Fatalf("glm_job_cancel on a live detached job IsError: %+v", res)
	}
	if st, _ := structuredMap(t, res)["status"].(string); st != glmJobStatusCancelled {
		t.Errorf("cancel result status = %q, want %q", st, glmJobStatusCancelled)
	}

	// The load-bearing half: the cancel must actually REACH the goroutine and
	// abort the blocked call. doer.unblock() has not run yet, so the job
	// leaving the live map can only be the revoked context.
	stopped := time.Now().Add(3 * time.Second)
	for time.Now().Before(stopped) {
		if _, live := glmLiveJobs.Load(jobID); !live {
			final, err := reg.load(jobID)
			if err != nil {
				t.Fatalf("load cancelled record: %v", err)
			}
			if final.Status != glmJobStatusCancelled {
				t.Errorf("recorded status = %q, want %q", final.Status, glmJobStatusCancelled)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the blocked call was never aborted by glm_job_cancel — the job's context no longer reaches it")
}
