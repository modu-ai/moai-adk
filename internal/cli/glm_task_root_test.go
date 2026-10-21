package cli

// t1632 finding 2 — REQ-MN-004: the glm_* job family resolves the CALLER's
// root. The producer (glm_task background) must write the job record under
// the root the request names — not the server's spawn tree — and the
// consumers (glm_job_status/result/cancel) must read the same root, or a job
// written under the named root becomes unobservable to the caller that
// created it.

import (
	"testing"
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
