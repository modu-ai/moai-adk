package cli

// t1632 l57-r6 finding 2 — REQ-MN-004: the codex_* job family resolves the
// CALLER's root, the same rule the glm_* family already carries (see
// glm_task_root_test.go). The producer (codex_task background) must write the
// job record under the root the request names — not the server's spawn tree —
// and the consumers (codex_job_status/result/cancel) must read the same root,
// or a cancel naming worktree B can cancel a job from tree A while the
// result's _root block claims B.

import (
	"testing"
)

// The fixture needs a repository: the caller's root must be a registered
// worktree of the server's repository to pass the REQ-MN-005 confinement.
// newUntrackedFixture builds exactly that layout (P = primary, W = linked).

// TestCodexTaskBackgroundJobRecordLandsUnderNamedRoot pins the producer half:
// background codex_task with project_root naming worktree W must write its job
// record under W, not under the server's tree P.
func TestCodexTaskBackgroundJobRecordLandsUnderNamedRoot(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	withCodexSession(t, codexTaskScript("trn-root", "background work done"))
	withCodexProjectDir(t, fx.P) // stamped last: the server is rooted at P
	withCodexServingDir(t, fx.P) // the anchor the root verification compares against

	res := callCodexTask(t, map[string]any{"prompt": "root confinement probe", "background": true, "project_root": fx.W})
	if res.IsError {
		t.Fatalf("background codex_task refused the registered worktree root: %s", resultText(res))
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background result must carry a job id: %+v", structuredMap(t, res))
	}
	t.Cleanup(func() { waitForJobToStop(t, jobID) })

	if _, err := newCodexJobRegistry(fx.W).load(jobID); err != nil {
		t.Fatalf("background codex_task with project_root=%s wrote no job record under the named root: %v", fx.W, err)
	}
	if _, err := newCodexJobRegistry(fx.P).load(jobID); err == nil {
		t.Fatal("the background job record landed under the SERVER tree — the caller's named root was ignored (REQ-MN-004)")
	}
}

// TestCodexJobStatusReadsTheNamedRoot is the consumer half of the same rule:
// codex_job_status must read the record under the root the REQUEST names, not
// the server's tree. The record is seeded directly under W so the consumer
// rule is pinned standalone (the producer half has its own test above — run
// together, a pre-fix producer AND a pre-fix consumer are consistently wrong
// at P, and the pair would stay green).
func TestCodexJobStatusReadsTheNamedRoot(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	withCodexProjectDir(t, fx.P) // the server is rooted at P

	rec, err := newCodexJobRegistry(fx.W).create(codexJobSpec{
		ThreadID:       "tid-root",
		Mode:           codexTaskMode,
		RequestSummary: "root parity probe",
	})
	if err != nil {
		t.Fatalf("seed the record under W: %v", err)
	}

	st := callCodexJobTool(t, handleCodexJobStatus, map[string]any{"job_id": rec.ID, "project_root": fx.W})
	if st.IsError {
		t.Fatalf("codex_job_status could not read the job record at the root the request names: %s", resultText(st))
	}
}
