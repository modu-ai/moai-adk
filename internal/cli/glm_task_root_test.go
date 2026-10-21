package cli

// t1632 finding 2 — REQ-MN-004: the glm_* job family resolves the CALLER's
// root. The producer (glm_task background) must write the job record under
// the root the request names — not the server's spawn tree — and the
// consumers (glm_job_status/result/cancel) must read the same root, or a job
// written under the named root becomes unobservable to the caller that
// created it.

import (
	"testing"
	"time"
)

// The fixture needs a repository: the caller's root must be a registered
// worktree of the server's repository to pass the REQ-MN-005 confinement.
// newUntrackedFixture builds exactly that layout (P = primary, W = linked).

// TestGLMTaskBackgroundJobRecordLandsUnderNamedRoot pins the producer half:
// background glm_task with project_root naming worktree W must write its job
// record under W, not under the server's tree P.
func TestGLMTaskBackgroundJobRecordLandsUnderNamedRoot(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	doer := newBlockingGLMDoer()
	withGLMTaskSeams(t, "k", doer)
	withCodexProjectDir(t, fx.P) // stamped last: the server is rooted at P

	jobID := startBackgroundGLMJobWithEndedRequest(t, fx.W, "root confinement probe")
	t.Cleanup(func() {
		doer.unblock()
		waitForGLMJobToStop(t, jobID)
	})

	if _, err := newGLMJobRegistry(fx.W).load(jobID); err != nil {
		t.Fatalf("background glm_task with project_root=%s wrote no job record under the named root: %v", fx.W, err)
	}
	if _, err := newGLMJobRegistry(fx.P).load(jobID); err == nil {
		t.Fatal("the background job record landed under the SERVER tree — the caller's named root was ignored (REQ-MN-004)")
	}
}

// TestGLMJobStatusReadsTheNamedRoot is the consumer half of the same rule:
// glm_job_status must read the record at the root the producer wrote it to.
// A revert of either half alone fails here — create/read parity is the
// contract (design.md §D3: producers and consumers resolve the same rule).
func TestGLMJobStatusReadsTheNamedRoot(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	doer := newBlockingGLMDoer()
	withGLMTaskSeams(t, "k", doer)
	withCodexProjectDir(t, fx.P)

	jobID := startBackgroundGLMJobWithEndedRequest(t, fx.W, "root parity probe")
	t.Cleanup(func() {
		doer.unblock()
		waitForGLMJobToStop(t, jobID)
	})

	res := callGLMJobTool(t, handleGLMJobStatus, map[string]any{"job_id": jobID, "project_root": fx.W})
	if res.IsError {
		t.Fatalf("glm_job_status could not read the job record at the root glm_task wrote it to: %s", resultText(res))
	}
}

// TestGLMJobCancelUpdatesTheNamedRoot pins the cancel's WRITE half of the same
// rule (r6b): the cancelled-status update must land in the registry the
// request names — the same one the record was read from. A cancel naming
// worktree B whose update routes to the server's tree P leaves the record at B
// uncancelled while _root claims B (the defect shape r6 fixed for the codex
// family's update path).
func TestGLMJobCancelUpdatesTheNamedRoot(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	doer := newBlockingGLMDoer()
	withGLMTaskSeams(t, "k", doer)
	withCodexProjectDir(t, fx.P) // stamped last: the server is rooted at P
	withShortGLMCancelGrace(t, 5*time.Second)

	// The record lives under W — the root the request named (the producer half
	// is pinned above).
	jobID := startBackgroundGLMJobWithEndedRequest(t, fx.W, "cancel parity probe")
	t.Cleanup(func() {
		doer.unblock()
		waitForGLMJobToStop(t, jobID)
	})

	res := callGLMJobTool(t, handleGLMJobCancel, map[string]any{"job_id": jobID, "project_root": fx.W})
	if res.IsError {
		t.Fatalf("cancelling a running job this server owns must not error: %s", resultText(res))
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusCancelled {
		t.Errorf("result status = %q, want %q — the cancelled-status update must land in the registry the request names", st, glmJobStatusCancelled)
	}
	rec, err := newGLMJobRegistry(fx.W).load(jobID)
	if err != nil {
		t.Fatalf("load the record under the named root: %v", err)
	}
	if rec.Status != glmJobStatusCancelled {
		t.Errorf("record at the named root = %q, want %q — the cancel's status update wrote somewhere other than the request root (REQ-MN-004)", rec.Status, glmJobStatusCancelled)
	}
}
