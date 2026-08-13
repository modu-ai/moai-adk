// Package cli — SPEC-CODEX-PHASE2-001 M3 codex_task.
//
// codex_task.go owns the task-delegation surface: it drives a codex turn with
// the caller's prompt (REQ-CX2-006), gates the writing sandbox behind an
// explicit project opt-in (REQ-CX2-007), and can continue the last recorded
// thread instead of opening a new one (REQ-CX2-008). It sits ON TOP of the
// session client (mcp_codex.go) and the job registry (codex_jobs.go); it writes
// no transport and no second client (plan.md §F AP-1).
//
// Two properties are load-bearing:
//
//   - sandboxPolicy is transmitted on EVERY turn, readOnly included. The
//     protocol makes the field sticky on the thread ("this turn and subsequent
//     turns"), so an omitted field on a non-writing turn would inherit a
//     write-enabled policy from an earlier turn that opted in — the gate is read
//     at request time, but its effect outlives the request. See
//     codexSandboxPolicy in mcp_codex.go for the full reasoning.
//
//   - a background job is a GOROUTINE inside this server process, not a
//     detached subprocess (plan.md §D M0 decision). Every recorded pid is
//     therefore one this process spawned in this lifetime, and every in-flight
//     job is lost when the server exits. No reattachment is attempted and none
//     is recorded.
//
// Nothing here invokes AskUserQuestion: a missing prompt, a refused write, and
// an unwritable state directory are all structured results the orchestrator
// translates (REQ-CX2-014 / C2).
//
// @MX:SPEC: SPEC-CODEX-PHASE2-001
package cli

import (
	"context"
	"errors"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// codexTaskToolName is the MCP tool name. Registration is M5; the handler
	// and its behavior are M3.
	codexTaskToolName = "codex_task"

	// codexTaskMode is the mode a task job records. It is deliberately distinct
	// from the codex_audit review modes (native / adversarial): a task is not a
	// review, and a later reader of a job record should not have to guess.
	codexTaskMode = "task"

	// Result notes. REQ-CX2-007 requires a refused write to be STATED in the
	// result, and REQ-CX2-008 requires an unresumed thread to be stated too — a
	// silent downgrade in either case would leave the caller believing it got
	// something it did not.
	codexTaskWriteRefusedNote = "write was requested but not honored: this project has not opted in " +
		"(set workflow.codex.task.allow_write: true in .moai/config/sections/workflow.yaml). " +
		"The turn ran read-only."
	codexTaskNoPriorThreadNote = "resume_last was requested but no prior thread is recorded for this project; " +
		"a new thread was opened."
)

// codexTaskTimeoutMessage names the bound that ended a turn (REQ-CX2-017). It
// reads the duration at call time rather than baking it in, so a shortened
// bound reports the value that actually fired.
func codexTaskTimeoutMessage() string {
	return "codex_task turn timed out after " + config.DefaultCodexTaskTimeout.String() +
		" (the bound codex_task imposes on its own turns); the turn was abandoned and the session torn down"
}

// runCodexTaskTurn drives ONE task turn under a deadline the tool imposes
// itself (REQ-CX2-017).
//
// The bound exists because nothing else provides one on this path: the caller's
// context is whatever the MCP host supplied and may carry no deadline at all,
// and a turn can stop advancing for reasons that never close the connection — a
// live session parked on an unanswered approval request and did not return
// within 120 s (progress.md §E.2). A deadline on the context alone would not be
// enough either: the driver blocks in conn.recv(), which returns when the
// connection ends, so the turn is raced against the timer here instead.
//
// On expiry the reader goroutine is left blocked in recv and is unblocked by
// the caller's session tear-down, which both call sites already perform on
// every exit path. Nothing is closed here, because closing a live session from
// two goroutines would put two exec.Cmd.Wait calls in flight on the same
// process.
func runCodexTaskTurn(ctx context.Context, session *codexSessionHandle, params map[string]any) (ReviewOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DefaultCodexTaskTimeout)
	defer cancel()

	type turnOutcome struct {
		out ReviewOutput
		err error
	}
	done := make(chan turnOutcome, 1) // buffered: an abandoned turn must not block on send
	go func() {
		out, err := session.runTurn(ctx, codexMethodTurnStart, params)
		done <- turnOutcome{out: out, err: err}
	}()

	select {
	case outcome := <-done:
		return outcome.out, outcome.err
	case <-ctx.Done():
		msg := codexTaskTimeoutMessage()
		return ReviewOutput{
			Verdict:   VerdictInconclusive,
			Summary:   msg,
			Findings:  []Finding{},
			NextSteps: []string{"re-run with a narrower prompt, or raise the codex_task bound"},
		}, errors.New(msg)
	}
}

// CodexTaskResult is the structured result codex_task returns. It is a distinct
// shape from ReviewOutput: a task produces output, not a verdict, and folding it
// into the review schema would have forced a meaningless pass/fail on every
// task (and dragged synthesizeReviewOutput into scope — plan.md §F AP-6).
type CodexTaskResult struct {
	// Status is a job-status value (codexJobStatus*). A foreground task reports
	// its terminal status; a background task reports the status at hand-off.
	Status string `json:"status"`

	// Background reports which form ran. JobID is set only for the background
	// form; Output only for the foreground form.
	Background bool   `json:"background"`
	JobID      string `json:"job_id,omitempty"`
	Output     string `json:"output,omitempty"`

	// ThreadID and TurnID address the turn. TurnID is empty when the turn was
	// never observed starting.
	ThreadID string `json:"thread_id,omitempty"`
	TurnID   string `json:"turn_id,omitempty"`

	// WriteRequested and WriteGranted are reported separately on purpose: a
	// caller that asked for writes and did not get them must be able to tell
	// that apart from never having asked.
	WriteRequested bool `json:"write_requested"`
	WriteGranted   bool `json:"write_granted"`

	// ResumedThread reports whether a previously-recorded thread was continued.
	ResumedThread bool `json:"resumed_thread"`

	// Note carries the human-readable statements REQ-CX2-007 / REQ-CX2-008
	// require; Error names a failure the tool absorbed rather than raised.
	Note  string `json:"note,omitempty"`
	Error string `json:"error,omitempty"`
}

// codexLiveJobSessions holds the session handle of every RUNNING background job,
// keyed by job id, so the job's turn can still be addressed while it is in
// flight. It is the seam the M4 cancel path reads: turn/interrupt must be sent
// on the session the goroutine still holds, and the pid it may terminate is the
// one that session spawned.
//
// It is in-process state by construction, matching the in-process execution
// model: an entry exists only for a job this server lifetime started and is
// removed the moment the job reaches a terminal status. A record found running
// with no entry here is stale (a previous server lifetime), which is exactly the
// case REQ-CX2-012 requires the cancel path to refuse rather than signal.
var codexLiveJobSessions sync.Map // job id (string) → *codexSessionHandle

// handleCodexTask is the handler for the `codex_task` MCP tool (REQ-CX2-006).
//
// Foreground (background=false): drives the turn to completion and returns its
// output. Background (background=true): creates the job record, hands the turn
// to a goroutine, and returns the job id immediately.
//
// Fail-open (C1): a missing or unreachable codex yields a structured result
// carrying the failure, never a Go error and never a panic. The two arms that DO
// set IsError are caller-actionable faults rather than codex being unavailable —
// a missing prompt, and a state directory that cannot be written.
func handleCodexTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt := req.GetString("prompt", "")
	background := req.GetBool("background", false)
	writeRequested := req.GetBool("write", false)
	resumeLast := req.GetBool("resume_last", false)
	token := extractProgressToken(req)
	notifyMCPProgress(ctx, token, 0, "codex task 시작 — 프롬프트 접수")

	if prompt == "" {
		return toolErr(codexTaskToolName, errors.New("prompt is required")), nil
	}

	projectDir := projectDirResolver()
	writeGranted := writeRequested && readCodexTaskAllowWrite(projectDir)

	result := CodexTaskResult{
		Background:     background,
		WriteRequested: writeRequested,
		WriteGranted:   writeGranted,
	}
	if writeRequested && !writeGranted {
		result.Note = codexTaskWriteRefusedNote
	}

	notifyMCPProgress(ctx, token, 0.1, "codex 바이너리 확인 — 세션 준비 중...")
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		result.Status = codexJobStatusFailed
		result.Error = "codex binary not found in PATH"
		return toolJSON(codexTaskToolName, result), nil
	}

	registry := newCodexJobRegistry(projectDir)

	// resume_last continues the last recorded thread instead of opening a new
	// one. When nothing is recorded, a new thread is opened AND said so.
	resumeThreadID := ""
	if resumeLast {
		if id, ok := registry.latestThreadID(); ok {
			resumeThreadID = id
		} else {
			result.Note = appendCodexNote(result.Note, codexTaskNoPriorThreadNote)
		}
	}

	turnParams := map[string]any{
		"prompt": prompt,
		"cwd":    projectDir,
		// EVERY turn carries the policy explicitly — see codexSandboxPolicy.
		"sandboxPolicy": codexSandboxPolicy(writeGranted),
	}

	notifyMCPProgress(ctx, token, 0.2, "codex 세션 오픈 중...")
	session, err := openCodexSessionOn(ctx, binaryPath, turnParams, resumeThreadID)
	if err != nil {
		var sErr *codexSessionError
		result.Status = codexJobStatusFailed
		if errors.As(err, &sErr) {
			result.Error = sErr.summary
		} else {
			result.Error = err.Error()
		}
		return toolJSON(codexTaskToolName, result), nil
	}

	result.ThreadID = session.threadID
	result.ResumedThread = resumeThreadID != "" && session.threadID == resumeThreadID

	if !background {
		defer func() { _ = session.close() }()
		out, runErr := runCodexTaskTurn(ctx, session, turnParams)
		result.TurnID = session.currentTurnID()
		if runErr != nil {
			result.Status = codexJobStatusFailed
			result.Error = out.Summary
			return toolJSON(codexTaskToolName, result), nil
		}
		result.Status = codexJobStatusCompleted
		result.Output = out.Summary
		return toolJSON(codexTaskToolName, result), nil
	}

	// Background: the record is created BEFORE the turn is handed off, so an
	// unwritable state directory is reported to the caller as a structured error
	// (REQ-CX2-004 / AC-CX2-007) rather than surfacing later as a job nobody can
	// observe. The session is torn down on that path — a job that cannot be
	// recorded must not leave a codex process running.
	rec, err := registry.create(codexJobSpec{
		ThreadID:       session.threadID,
		PID:            session.pid(),
		Mode:           codexTaskMode,
		RequestSummary: prompt,
	})
	if err != nil {
		_ = session.close()
		return toolErr(codexTaskToolName, err), nil
	}

	// The turnId must land in the record while the turn is still RUNNING, which
	// is the only window in which it is useful for cancellation.
	session.setTurnStartedObserver(registry.turnIDRecorder(rec.ID))
	codexLiveJobSessions.Store(rec.ID, session)

	if _, err := registry.update(rec.ID, func(r *CodexJobRecord) { r.Status = codexJobStatusRunning }); err != nil {
		codexLiveJobSessions.Delete(rec.ID)
		_ = session.close()
		return toolErr(codexTaskToolName, err), nil
	}

	go runCodexBackgroundJob(ctx, registry, rec.ID, session, turnParams)

	result.Status = codexJobStatusRunning
	result.JobID = rec.ID
	return toolJSON(codexTaskToolName, result), nil
}

// runCodexBackgroundJob drives one background job's turn to completion and
// records the outcome. It runs as a goroutine inside this server process
// (plan.md §D M0 decision), so the job dies with the server; nothing here
// attempts to survive that.
//
// The live-session entry is removed and the session closed on EVERY exit path,
// so a terminal record is never left with a live entry the cancel path could
// still address.
func runCodexBackgroundJob(ctx context.Context, registry *codexJobRegistry, jobID string, session *codexSessionHandle, params map[string]any) {
	defer func() {
		codexLiveJobSessions.Delete(jobID)
		_ = session.close()
	}()

	out, runErr := runCodexTaskTurn(ctx, session, params)

	// A job cancelled while the turn was in flight keeps its cancelled status:
	// the turn returning afterwards must not overwrite it with completed or
	// failed (M4 sets that status). The guard lives in the REGISTRY rather than
	// in the mutator below, because a mutator can only decline to change the
	// record — it cannot decline the write, and a write landing after
	// codex_job_cancel returned is exactly what has to stop.
	_, _ = registry.updateUnlessCancelled(jobID, func(r *CodexJobRecord) {
		if runErr != nil {
			r.Status = codexJobStatusFailed
			r.Error = out.Summary
			return
		}
		r.Status = codexJobStatusCompleted
		r.Output = out.Summary
	})
}

// appendCodexNote joins two result notes, keeping both statements rather than
// letting the second silently replace the first.
func appendCodexNote(existing, add string) string {
	if existing == "" {
		return add
	}
	return existing + " " + add
}
