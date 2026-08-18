// Package cli — GLM (z.ai) job control.
//
// glm_job_control.go owns the three tools that observe and stop a background
// GLM task: glm_job_status, glm_job_result, and glm_job_cancel. It mirrors the
// codex job-control surface (codex_job_control.go) — the structural SSOT for
// the delegation-family job layer — reading the durable record the registry
// writes (glm_jobs.go) and the live cancel handle the task handler holds
// (glm_task.go); it makes no HTTP call of its own.
//
// Three properties are load-bearing (mirrored, minus the process stage):
//
//   - Ownership is decided by glmLiveJobs membership, NOT by anything in the
//     record. A background job is a goroutine inside THIS process, so an entry
//     exists exactly for the jobs this server lifetime started and is removed
//     the moment one ends. A record naming a running job with no entry is
//     stale by construction — a leftover from a previous server lifetime — and
//     is REFUSED rather than acted on.
//
//   - Cancellation revokes the job's context; there is no process to signal.
//     The z.ai call runs on a context derived from the stored cancel function,
//     so cancelling it aborts the in-flight HTTP request the same way a real
//     http.Client aborts on context cancellation.
//
//   - The call is bounded. The cancel function is invoked and NOT awaited
//     directly; the outcome is observed as the job ending, waited for through
//     a grace window. The tool therefore returns within the grace window plus
//     one poll, never waiting on the call itself.
//
// Nothing here invokes AskUserQuestion: an unknown job, a stale record, and an
// unreadable record are all structured results the orchestrator translates.
package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// MCP tool names.
	glmJobStatusToolName = "glm_job_status"
	glmJobResultToolName = "glm_job_result"
	glmJobCancelToolName = "glm_job_cancel"

	// glmJobIDArg is the argument every job-control tool takes.
	glmJobIDArg = "job_id"

	// Result notes. Each states an outcome the caller would otherwise have to
	// infer from a missing field.
	glmJobCancelTerminalNote = "the job had already reached a terminal status; nothing was sent."
	glmJobCancelExitedNote   = "the job ended within the grace window; its in-flight request was revoked."
	glmJobCancelGraceNote    = "the job did not end within the grace window; its context was revoked and it will " +
		"record its terminal status when the request returns."
	glmJobResultPendingNote = "the job has not reached a terminal status; no output is recorded yet."
)

// errGLMJobIDRequired is returned when the job_id argument is absent.
var errGLMJobIDRequired = errors.New("job_id is required")

// glmJobCancelGrace is how long cancel waits for the revoked job to end. The
// value's SSOT is internal/config/defaults.go (DefaultGLMJobCancelGrace,
// derived from the codex bound); it is a var here only so a test can shorten
// the window it must let expire.
var glmJobCancelGrace = config.DefaultGLMJobCancelGrace

// ─── result shapes ───

// GLMJobResultOutput is what glm_job_result returns. Terminal is reported
// explicitly so a caller never has to re-derive the lifecycle from an empty
// Output — a completed job with no output and a running job with none yet are
// different answers to the same question.
type GLMJobResultOutput struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Terminal bool   `json:"terminal"`
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Note     string `json:"note,omitempty"`
}

// GLMJobCancelOutput is what glm_job_cancel returns. It mirrors
// CodexJobCancelOutput minus the interrupt/process fields: cancelling a GLM
// job is ONE act (revoking the stored context), reported by cancel_sent.
type GLMJobCancelOutput struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`

	CancelSent bool `json:"cancel_sent"`

	Note  string `json:"note,omitempty"`
	Error string `json:"error,omitempty"`
}

// ─── glm_job_status ───

// handleGLMJobStatus returns the job's record. An unknown id is a structured
// not-found result with IsError set — not a Go error, which would abort the
// tool call and leave the caller unable to distinguish "no such job" from
// "the server broke". A record file that is present but malformed is reported
// the same way rather than decoded into an empty record.
func handleGLMJobStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rec, errResult := loadGLMJobFor(glmJobStatusToolName, req)
	if errResult != nil {
		return errResult, nil
	}
	return toolJSON(glmJobStatusToolName, rec), nil
}

// ─── glm_job_result ───

// handleGLMJobResult returns the recorded output of a terminal job, or the
// current status of a non-terminal one. It never waits for the call: a running
// job is reported as running and the caller decides whether to ask again.
func handleGLMJobResult(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rec, errResult := loadGLMJobFor(glmJobResultToolName, req)
	if errResult != nil {
		return errResult, nil
	}

	out := GLMJobResultOutput{
		JobID:    rec.ID,
		Status:   rec.Status,
		Terminal: glmJobTerminal(rec.Status),
	}
	if out.Terminal {
		out.Output = rec.Output
		out.Error = rec.Error
	} else {
		out.Note = glmJobResultPendingNote
	}
	return toolJSON(glmJobResultToolName, out), nil
}

// ─── glm_job_cancel ───

// handleGLMJobCancel stops a running job by revoking the context its in-flight
// call runs under, then waiting a bounded grace window for the job to end.
//
// The ownership check comes before anything is sent: the live-job map is the
// record of what this server lifetime actually started, so a job absent from
// it is refused rather than cancelled.
func handleGLMJobCancel(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rec, errResult := loadGLMJobFor(glmJobCancelToolName, req)
	if errResult != nil {
		return errResult, nil
	}

	out := GLMJobCancelOutput{JobID: rec.ID, Status: rec.Status}

	// Already finished: report the terminal status and send nothing.
	if glmJobTerminal(rec.Status) {
		out.Note = glmJobCancelTerminalNote
		return toolJSON(glmJobCancelToolName, out), nil
	}

	// Ownership. An entry exists only for a job this server lifetime started
	// and is deleted the moment one ends, so its absence on a non-terminal
	// record means the record outlived the process that owned it.
	stored, live := glmLiveJobs.Load(rec.ID)
	cancel, _ := stored.(context.CancelFunc)
	if !live || cancel == nil {
		return toolErr(glmJobCancelToolName, fmt.Errorf(
			"job %s is recorded %q but this server holds no live job for it — "+
				"background jobs run in-process and do not survive a restart, so the record is stale; "+
				"refusing to cancel a job this server did not start",
			rec.ID, rec.Status)), nil
	}

	// Record the cancel BEFORE revoking the context. A call that completes
	// while the cancellation is in flight would otherwise write completed or
	// failed over the cancel; runGLMBackgroundJob keeps a cancelled status
	// untouched, and this ordering is what puts the status there in time for
	// that guard to see it.
	if updated, err := currentGLMJobRegistry().update(rec.ID, func(r *GLMJobRecord) {
		r.Status = glmJobStatusCancelled
	}); err != nil {
		out.Error = err.Error()
	} else {
		out.Status = updated.Status
	}

	// Cancel: revoke the context the job's in-flight request runs under. There
	// is no process-termination stage — the GLM backend has no pid, and the
	// revoked context is what actually ends the call.
	cancel()
	out.CancelSent = true

	// Grace window: give the revoked call a bounded chance to end on its own.
	if glmAwaitJobExit(rec.ID, glmJobCancelGrace) {
		out.Note = glmJobCancelExitedNote
		return toolJSON(glmJobCancelToolName, out), nil
	}
	out.Note = glmJobCancelGraceNote
	return toolJSON(glmJobCancelToolName, out), nil
}

// ─── shared helpers ───

// currentGLMJobRegistry returns the registry rooted at the current project.
func currentGLMJobRegistry() *glmJobRegistry {
	return newGLMJobRegistry(projectDirResolver())
}

// loadGLMJobFor reads the job named by the request's job_id argument. It
// returns either the record or the structured error result the tool should
// return verbatim — never both, and never a Go error.
func loadGLMJobFor(tool string, req mcp.CallToolRequest) (GLMJobRecord, *mcp.CallToolResult) {
	jobID := req.GetString(glmJobIDArg, "")
	if jobID == "" {
		return GLMJobRecord{}, toolErr(tool, errGLMJobIDRequired)
	}
	rec, err := currentGLMJobRegistry().load(jobID)
	if err != nil {
		if glmJobNotFound(err) {
			return GLMJobRecord{}, toolErr(tool, fmt.Errorf("no job record for id %s", jobID))
		}
		return GLMJobRecord{}, toolErr(tool, err)
	}
	return rec, nil
}

// glmAwaitJobExit waits up to grace for the job to leave the live-job map,
// which happens on every exit path of runGLMBackgroundJob. It reports whether
// the job ended. Polling the map rather than the record file is deliberate:
// the map entry is removed by the goroutine itself, so its absence means the
// call is genuinely done rather than merely marked done.
func glmAwaitJobExit(jobID string, grace time.Duration) bool {
	deadline := time.Now().Add(grace)
	for {
		if _, live := glmLiveJobs.Load(jobID); !live {
			return true
		}
		if !time.Now().Before(deadline) {
			return false
		}
		time.Sleep(config.DefaultCodexJobCancelPoll) // shared poll interval with the codex cancel path
	}
}
