// Package cli — SPEC-CODEX-PHASE2-001 M4 codex job control.
//
// codex_job_control.go owns the three tools that observe and stop a background
// codex task: codex_job_status (REQ-CX2-009), codex_job_result (REQ-CX2-010),
// and codex_job_cancel (REQ-CX2-011 / REQ-CX2-012). It reads the durable record
// the registry writes (codex_jobs.go) and the live session the task handler
// holds (codex_task.go); it opens no session and spawns no process of its own.
//
// Three properties are load-bearing:
//
//   - Ownership is decided by codexLiveJobSessions membership, NOT by the pid in
//     the record. A background job is a goroutine inside THIS process, so an
//     entry exists exactly for the jobs this server lifetime started and is
//     removed the moment one ends. A record naming a running job with no entry
//     is stale by construction — a leftover from a previous server lifetime —
//     and is REFUSED rather than signalled (REQ-CX2-012). The record's pid is
//     never the signal target, so a tampered or stale record cannot direct a
//     signal anywhere; only the live session's own connection can.
//
//   - Cancellation targets ONE pid: the process the live session spawned for
//     that job. There is no pattern match, no name match, and no sweep — a
//     `pkill codex` would kill a developer's interactive session (plan.md §F
//     AP-4). No pid reattachment is attempted, because under the in-process
//     execution model there is nothing to reattach to.
//
//   - The call is bounded. turn/interrupt is sent and NOT awaited; the outcome
//     is observed as the job ending, waited for through a grace window, after
//     which the process is terminated. The tool therefore returns within the
//     grace window plus one poll, never waiting on the turn itself.
//
// Nothing here invokes AskUserQuestion: an unknown job, a stale record, and an
// unreadable record are all structured results the orchestrator translates
// (REQ-CX2-014 / C2).
//
// @MX:SPEC: SPEC-CODEX-PHASE2-001
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// MCP tool names. Registration with JSON Schema and read-only hints is M5;
	// the handlers and their behavior are M4.
	codexJobStatusToolName = "codex_job_status"
	codexJobResultToolName = "codex_job_result"
	codexJobCancelToolName = "codex_job_cancel"

	// codexJobIDArg is the argument every job-control tool takes.
	codexJobIDArg = "job_id"

	// Result notes. Each states an outcome the caller would otherwise have to
	// infer from a missing field.
	codexJobCancelTerminalNote = "the job had already reached a terminal status; nothing was sent and no process was signalled."
	codexJobCancelNoTurnIDNote = "no turn id is recorded for this job (its turn/started notification was never observed), " +
		"so turn/interrupt — which requires both threadId and turnId — could not be sent. " +
		"The cancel fell back to terminating the codex process this server spawned for the job."
	codexJobCancelExitedNote    = "the job ended within the grace window; no process termination was needed."
	codexJobCancelNoProcessNote = "the job's session has no backing process to terminate."
	codexJobResultPendingNote   = "the job has not reached a terminal status; no output is recorded yet."
)

// errCodexJobIDRequired is returned when the job_id argument is absent.
var errCodexJobIDRequired = errors.New("job_id is required")

// codexJobCancelGrace is how long cancel waits for an interrupted turn to end
// before terminating the process. The value is the SSOT in
// internal/config/defaults.go (REQ-CX2-015); it is a var here only so a test can
// shorten the window it must let expire.
var codexJobCancelGrace = config.DefaultCodexJobCancelGrace

// codexTerminateProcess is the process-termination seam. Tests swap it to record
// the pids termination was attempted on, which is how AC-CX2-015 asserts the
// ABSENCE of a signal — observing a live process could not distinguish "we did
// not signal it" from "we signalled it and it survived".
var codexTerminateProcess = terminateCodexProcess

// terminateCodexProcess terminates exactly one process by pid.
//
// It is a single Kill rather than a SIGTERM-then-SIGKILL escalation, and that is
// deliberate on two grounds. The graceful phase already happened — turn/interrupt
// was sent and the grace window elapsed — so a second graceful stage would only
// extend an already-bounded call. And os.Process.Kill is uniform across darwin,
// linux, and windows, whereas syscall.SIGTERM does not exist on windows and would
// have forced a build-tag split for a stage the flow does not need (plan.md §B5).
func terminateCodexProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("refusing to terminate pid %d: not a process this server spawned", pid)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("codex job cancel: find process %d: %w", pid, err)
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("codex job cancel: terminate process %d: %w", pid, err)
	}
	return nil
}

// ─── result shapes ───

// CodexJobResultOutput is what codex_job_result returns. Terminal is reported
// explicitly so a caller never has to re-derive the lifecycle from an empty
// Output — a completed job with no output and a running job with none yet are
// different answers to the same question.
type CodexJobResultOutput struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Terminal bool   `json:"terminal"`
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Note     string `json:"note,omitempty"`
}

// CodexJobCancelOutput is what codex_job_cancel returns. InterruptSent and
// ProcessTerminated are reported separately because they are separate facts: a
// job with no recorded turn id is terminated without an interrupt, and a job
// that yields to the interrupt is never terminated at all.
type CodexJobCancelOutput struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`

	InterruptSent     bool `json:"interrupt_sent"`
	ProcessTerminated bool `json:"process_terminated"`

	// PID is the process termination targeted, or 0 when none was.
	PID int `json:"pid,omitempty"`

	Note  string `json:"note,omitempty"`
	Error string `json:"error,omitempty"`
}

// ─── codex_job_status (REQ-CX2-009) ───

// handleCodexJobStatus returns the job's record. An unknown id is a structured
// not-found result with IsError set — not a Go error, which would abort the tool
// call and leave the caller unable to distinguish "no such job" from "the server
// broke" (AC-CX2-012). A record file that is present but malformed is reported
// the same way rather than decoded into an empty record.
func handleCodexJobStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rec, errResult := loadCodexJobFor(codexJobStatusToolName, req)
	if errResult != nil {
		return errResult, nil
	}
	return toolJSON(codexJobStatusToolName, rec), nil
}

// ─── codex_job_result (REQ-CX2-010) ───

// handleCodexJobResult returns the recorded output of a terminal job, or the
// current status of a non-terminal one. It never waits for the turn: a running
// job is reported as running and the caller decides whether to ask again.
func handleCodexJobResult(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rec, errResult := loadCodexJobFor(codexJobResultToolName, req)
	if errResult != nil {
		return errResult, nil
	}

	out := CodexJobResultOutput{
		JobID:    rec.ID,
		Status:   rec.Status,
		Terminal: codexJobTerminal(rec.Status),
	}
	if out.Terminal {
		out.Output = rec.Output
		out.Error = rec.Error
	} else {
		out.Note = codexJobResultPendingNote
	}
	return toolJSON(codexJobResultToolName, out), nil
}

// ─── codex_job_cancel (REQ-CX2-011, REQ-CX2-012) ───

// handleCodexJobCancel stops a running job: send turn/interrupt on that job's
// own session carrying both required params from the record, then terminate the
// process this server spawned for it if the turn does not end within the grace
// window.
//
// The ownership check comes before anything is sent (REQ-CX2-012): the live
// session map is the record of what this server lifetime actually started, so a
// job absent from it is refused rather than signalled.
func handleCodexJobCancel(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rec, errResult := loadCodexJobFor(codexJobCancelToolName, req)
	if errResult != nil {
		return errResult, nil
	}

	out := CodexJobCancelOutput{JobID: rec.ID, Status: rec.Status}

	// Already finished: report the terminal status and send nothing
	// (acceptance.md §B).
	if codexJobTerminal(rec.Status) {
		out.Note = codexJobCancelTerminalNote
		return toolJSON(codexJobCancelToolName, out), nil
	}

	// Ownership (REQ-CX2-012). An entry exists only for a job this server
	// lifetime started and is deleted the moment one ends, so its absence on a
	// non-terminal record means the record outlived the process that owned it.
	stored, live := codexLiveJobSessions.Load(rec.ID)
	session, _ := stored.(*codexSessionHandle)
	if !live || session == nil {
		return toolErr(codexJobCancelToolName, fmt.Errorf(
			"job %s is recorded %q but this server holds no live session for it — "+
				"background jobs run in-process and do not survive a restart, so the record is stale; "+
				"refusing to signal a process this server did not spawn",
			rec.ID, rec.Status)), nil
	}

	// Record the cancel BEFORE interrupting. A turn that completes while the
	// interrupt is in flight would otherwise write completed or failed over the
	// cancel; runCodexBackgroundJob keeps a cancelled status untouched, and this
	// ordering is what puts the status there in time for that guard to see it.
	if updated, err := currentCodexJobRegistry().update(rec.ID, func(r *CodexJobRecord) {
		r.Status = codexJobStatusCancelled
	}); err != nil {
		out.Error = err.Error()
	} else {
		out.Status = updated.Status
	}

	// Interrupt. A record with no turnId cannot supply the method's second
	// required argument, so the cancel degrades to process termination rather
	// than sending a malformed request (M2/M3 carried this branch forward).
	if rec.TurnID == "" {
		out.Note = codexJobCancelNoTurnIDNote
	} else if err := session.sendTurnInterrupt(rec.ThreadID, rec.TurnID); err != nil {
		out.Note = appendCodexNote(out.Note, "turn/interrupt could not be sent: "+err.Error()+
			" The cancel fell back to terminating the codex process this server spawned for the job.")
	} else {
		out.InterruptSent = true
	}

	// Grace window: give the interrupted turn a bounded chance to end on its own.
	if codexAwaitJobExit(rec.ID, codexJobCancelGrace) {
		out.Note = appendCodexNote(out.Note, codexJobCancelExitedNote)
		return toolJSON(codexJobCancelToolName, out), nil
	}

	// Termination targets the pid of the connection THIS server's live session
	// holds for this job — never the pid in the record, which is untrusted input
	// from a file, and never a name or pattern (plan.md §F AP-4).
	pid := session.pid()
	if pid <= 0 {
		out.Note = appendCodexNote(out.Note, codexJobCancelNoProcessNote)
		return toolJSON(codexJobCancelToolName, out), nil
	}
	out.PID = pid
	if err := codexTerminateProcess(pid); err != nil {
		out.Error = appendCodexNote(out.Error, err.Error())
	} else {
		out.ProcessTerminated = true
	}
	return toolJSON(codexJobCancelToolName, out), nil
}

// ─── shared helpers ───

// currentCodexJobRegistry returns the registry rooted at the current project.
func currentCodexJobRegistry() *codexJobRegistry {
	return newCodexJobRegistry(projectDirResolver())
}

// loadCodexJobFor reads the job named by the request's job_id argument. It
// returns either the record or the structured error result the tool should
// return verbatim — never both, and never a Go error.
func loadCodexJobFor(tool string, req mcp.CallToolRequest) (CodexJobRecord, *mcp.CallToolResult) {
	jobID := req.GetString(codexJobIDArg, "")
	if jobID == "" {
		return CodexJobRecord{}, toolErr(tool, errCodexJobIDRequired)
	}
	rec, err := currentCodexJobRegistry().load(jobID)
	if err != nil {
		if codexJobNotFound(err) {
			return CodexJobRecord{}, toolErr(tool, fmt.Errorf("no job record for id %s", jobID))
		}
		return CodexJobRecord{}, toolErr(tool, err)
	}
	return rec, nil
}

// codexAwaitJobExit waits up to grace for the job to leave the live-session map,
// which happens on every exit path of runCodexBackgroundJob. It reports whether
// the job ended. Polling the map rather than the record file is deliberate: the
// map entry is removed by the goroutine itself, so its absence means the turn is
// genuinely done rather than merely marked done.
func codexAwaitJobExit(jobID string, grace time.Duration) bool {
	deadline := time.Now().Add(grace)
	for {
		if _, live := codexLiveJobSessions.Load(jobID); !live {
			return true
		}
		if !time.Now().Before(deadline) {
			return false
		}
		time.Sleep(config.DefaultCodexJobCancelPoll)
	}
}
