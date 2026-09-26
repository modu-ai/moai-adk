// Package cli — SPEC-CODEX-PHASE2-001 M3 codex_task.
//
// codex_task.go owns the task-delegation surface: it drives a codex turn with
// the caller's prompt (REQ-CX2-006), gates the writing sandbox behind an
// explicit project opt-in (REQ-CX2-007), and can continue a recorded thread
// instead of opening a new one (REQ-CX2-008, scoped by
// SPEC-CODEX-RESUME-SCOPE-001: an explicit thread_id or work_key selects the
// thread, and a selector-less resume_last is refused when more than one thread
// is recorded rather than guessing by recency). It sits ON TOP of the
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
	"strconv"
	"strings"
	"sync"
	"unicode"

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

// Resume selection (SPEC-CODEX-RESUME-SCOPE-001). resume_basis names WHICH rule
// chose the thread that was sent; error_code names why a call was refused.
const (
	codexResumeBasisThreadID   = "thread_id"
	codexResumeBasisWorkKey    = "work_key"
	codexResumeBasisSoleThread = "sole_thread"

	codexTaskErrThreadNotRecorded = "thread_not_recorded"
	codexTaskErrResumeAmbiguous   = "resume_ambiguous"
	codexTaskErrInvalidWorkKey    = "invalid_work_key"

	// codexTaskWorkKeyMaxBytes bounds a work_key (REQ-CRS-005): it is stored in
	// every record and returned in every ambiguity refusal.
	codexTaskWorkKeyMaxBytes = 128

	// codexTaskMaxCandidates bounds the candidate list of a resume_ambiguous
	// refusal (REQ-CRS-006); candidate_total still reports the full count.
	codexTaskMaxCandidates = 10
)

// codexTaskNoWorkKeyThreadNote states that a work_key matched no recorded
// thread, so a new one was opened (REQ-CRS-004).
func codexTaskNoWorkKeyThreadNote(workKey string) string {
	return "resume_last was requested with work_key " + strconv.Quote(workKey) +
		" but no prior thread is recorded for that work_key; a new thread was opened."
}

// validCodexWorkKey reports whether a supplied work_key is usable
// (REQ-CRS-005): non-empty after trimming, at most codexTaskWorkKeyMaxBytes, and
// free of control characters.
func validCodexWorkKey(key string) bool {
	if strings.TrimSpace(key) == "" || len(key) > codexTaskWorkKeyMaxBytes {
		return false
	}
	for _, r := range key {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// codexTaskTimeoutMessage names the bound that ended a turn (REQ-CX2-017). It
// reads the duration at call time rather than baking it in, so a shortened
// bound reports the value that actually fired.
func codexTaskTimeoutMessage() string {
	return "codex_task turn timed out after " + config.DefaultCodexTaskTimeout.String() +
		" (the bound codex_task imposes on its own turns); the turn was abandoned and the session torn down"
}

var errCodexTaskSessionDeadline = errors.New("codex_task session deadline")

func codexTaskHandshakeTimeoutMessage() string {
	return "codex_task handshake timed out after " + config.DefaultCodexTaskTimeout.String() +
		" (the bound codex_task imposes on background sessions); the session was torn down"
}

// codexTaskCallerEndedMessage names the CALLER's context as what ended the turn
// (t514 / GH #1687). It exists because the timeout wording above was reported
// for turns that ended in milliseconds: both causes reach the same select arm,
// and naming only the bound sent every diagnosis toward slow models and slow
// networks when the turn had never been given a chance to run.
//
// The distinction is drawn from the PARENT context rather than the derived one:
// the derived context reports DeadlineExceeded for the caller's deadline and
// for ours alike, so only the parent's own error separates them.
func codexTaskCallerEndedMessage(cause error) string {
	reason := "cancelled by the caller"
	if errors.Is(cause, context.DeadlineExceeded) {
		reason = "cancelled when the caller's own deadline expired"
	}
	return "codex_task turn was " + reason + ", well before the " +
		config.DefaultCodexTaskTimeout.String() +
		" bound codex_task imposes on its own turns; the turn was abandoned and the session torn down"
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
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, config.DefaultCodexTaskTimeout)
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
		// Our bound fired only if the parent is still alive; otherwise the
		// caller ended the turn and the bound is not what happened to it.
		msg := codexTaskTimeoutMessage()
		nextStep := "re-run with a narrower prompt, or raise the codex_task bound"
		if parentErr := parent.Err(); parentErr != nil && !errors.Is(context.Cause(parent), errCodexTaskSessionDeadline) {
			msg = codexTaskCallerEndedMessage(parentErr)
			nextStep = "re-run with a context that outlives the turn; the codex_task bound was never reached"
		}
		return ReviewOutput{
			Verdict:   VerdictInconclusive,
			Summary:   msg,
			Findings:  []Finding{},
			NextSteps: []string{nextStep},
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

	// ResumeThreadID is the thread id SENT in thread/resume and ResumeBasis the
	// rule that chose it (thread_id / work_key / sole_thread). Both are set once
	// the request is written — so a rejected resume still names what was asked —
	// and absent when no thread/resume was sent (SPEC-CODEX-RESUME-SCOPE-001
	// REQ-CRS-001). ThreadID above stays the id codex RETURNED; the two differ
	// exactly when ResumedThread is false despite a resume.
	ResumeThreadID string `json:"resume_thread_id,omitempty"`
	ResumeBasis    string `json:"resume_basis,omitempty"`

	// UnusedSelectors lists supplied selector inputs that did not decide the
	// thread because thread_id took precedence (REQ-CRS-008).
	UnusedSelectors []string `json:"unused_selectors,omitempty"`

	// Note carries the human-readable statements REQ-CX2-007 / REQ-CX2-008
	// require; Error names a failure the tool absorbed rather than raised.
	Note  string `json:"note,omitempty"`
	Error string `json:"error,omitempty"`

	// ErrorCode is the machine-readable code of a refusal (REQ-CRS-010):
	// thread_not_recorded, invalid_work_key, or resume_ambiguous. A refusal is a
	// structured result with status failed, never an MCP tool error.
	ErrorCode string `json:"error_code,omitempty"`

	// CandidateTotal and Candidates describe a resume_ambiguous refusal: the
	// number of distinct recorded threads and the newest of them, each carrying
	// the selector values (thread_id, work_key) the next call can pass
	// (REQ-CRS-006).
	CandidateTotal int                    `json:"candidate_total,omitempty"`
	Candidates     []codexThreadCandidate `json:"candidates,omitempty"`
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

// The request context ends when a background tool call returns. Keep a
// separate, bounded process lifetime and cancel it when the MCP server exits.
var codexLiveJobCancels sync.Map // job id (string) → context.CancelFunc
var codexBackgroundJobs sync.WaitGroup

func stopCodexBackgroundJobs() {
	codexLiveJobCancels.Range(func(_, value any) bool {
		value.(context.CancelFunc)()
		return true
	})
	codexBackgroundJobs.Wait()
}

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
	threadID := req.GetString("thread_id", "")
	workKey := req.GetString("work_key", "")
	_, workKeySupplied := req.GetArguments()["work_key"]
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

	// Every refusal below is decided BEFORE the session opens, so a refused call
	// starts no codex process (REQ-CRS-003/005/006).
	if workKeySupplied && !validCodexWorkKey(workKey) {
		return codexTaskRefusal(result, codexTaskErrInvalidWorkKey,
			"work_key must be non-empty after trimming, at most "+strconv.Itoa(codexTaskWorkKeyMaxBytes)+
				" bytes, and free of control characters"), nil
	}
	workKey = strings.TrimSpace(workKey)

	resumeThreadID, resumeBasis := "", ""
	switch {
	case threadID != "":
		// An explicit thread wins over every other selector (REQ-CRS-008), but
		// only a thread THIS project recorded may be resumed (REQ-CRS-003).
		if !registry.hasThread(threadID) {
			return codexTaskRefusal(result, codexTaskErrThreadNotRecorded,
				"thread_id "+strconv.Quote(threadID)+" is not recorded in this project's codex job registry"), nil
		}
		resumeThreadID, resumeBasis = threadID, codexResumeBasisThreadID
		if resumeLast {
			result.UnusedSelectors = append(result.UnusedSelectors, "resume_last")
		}
		if workKeySupplied {
			result.UnusedSelectors = append(result.UnusedSelectors, "work_key")
		}
	case resumeLast && workKeySupplied:
		// work_key scopes resume_last to this work item's records (REQ-CRS-004).
		if id, ok := registry.latestThreadForWorkKey(workKey); ok {
			resumeThreadID, resumeBasis = id, codexResumeBasisWorkKey
		} else {
			result.Note = appendCodexNote(result.Note, codexTaskNoWorkKeyThreadNote(workKey))
		}
	case resumeLast:
		// No selector: resume only when the choice is not a guess. One distinct
		// thread is resumed as before (REQ-CRS-007); several are refused with the
		// candidates the caller can select from (REQ-CRS-006) — recency across
		// work items is never the deciding rule (REQ-CRS-002).
		threads := registry.recordedThreads()
		switch len(threads) {
		case 0:
			result.Note = appendCodexNote(result.Note, codexTaskNoPriorThreadNote)
		case 1:
			resumeThreadID, resumeBasis = threads[0].ThreadID, codexResumeBasisSoleThread
		default:
			result.CandidateTotal = len(threads)
			result.Candidates = threads[:min(len(threads), codexTaskMaxCandidates)]
			return codexTaskRefusal(result, codexTaskErrResumeAmbiguous,
				"resume_last matched "+strconv.Itoa(len(threads))+" recorded threads; "+
					"pass thread_id or work_key from candidates to choose one"), nil
		}
	}

	turnParams := map[string]any{
		"prompt": prompt,
		"cwd":    projectDir,
		// EVERY turn carries the policy explicitly — see codexSandboxPolicy.
		"sandboxPolicy": codexSandboxPolicy(writeGranted),
	}

	// The SESSION's context decides the codex child's lifetime: the production
	// runner spawns it with exec.CommandContext and stops reading stdout when
	// that context ends (card t1186). A foreground session stays on the request
	// context. A background session must outlive the request — the host ends it
	// the moment this handler returns — so it gets a detached context the job
	// goroutine cancels on exit. The task deadline also covers the handshake,
	// while the request still bounds it through the AfterFunc tie.
	sessionCtx, cancelSession := ctx, context.CancelFunc(func() {})
	untieRequest := func() bool { return true }
	if background {
		sessionCtx, cancelSession = context.WithTimeoutCause(
			context.WithoutCancel(ctx), config.DefaultCodexTaskTimeout, errCodexTaskSessionDeadline)
		untieRequest = context.AfterFunc(ctx, cancelSession)
	}

	notifyMCPProgress(ctx, token, 0.2, "codex 세션 오픈 중...")
	session, err := openCodexSessionOn(sessionCtx, binaryPath, turnParams, resumeThreadID)
	// Report the resume once its request was written, including when codex then
	// rejected it; a failure before the write reports none (REQ-CRS-001).
	if resumeThreadID != "" && (err == nil || codexThreadRequestWasSent(err)) {
		result.ResumeThreadID, result.ResumeBasis = resumeThreadID, resumeBasis
	}
	if background && !untieRequest() && err == nil {
		// The request ended before hand-off; the session is already being torn
		// down, so report that instead of starting a job on a dead session.
		_ = session.close()
		err = errors.New("codex_task request was cancelled before the background job was handed off")
	}
	if err != nil {
		cancelSession()
		var sErr *codexSessionError
		result.Status = codexJobStatusFailed
		if background && ctx.Err() == nil && errors.Is(context.Cause(sessionCtx), errCodexTaskSessionDeadline) {
			result.Error = codexTaskHandshakeTimeoutMessage()
		} else if errors.As(err, &sErr) {
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
		WorkKey:        workKey, // recorded whichever selector decided (REQ-CRS-005)
	})
	if err != nil {
		_ = session.close()
		cancelSession()
		return toolErr(codexTaskToolName, err), nil
	}

	// The turnId must land in the record while the turn is still RUNNING, which
	// is the only window in which it is useful for cancellation.
	session.setTurnStartedObserver(registry.turnIDRecorder(rec.ID))
	codexLiveJobSessions.Store(rec.ID, session)

	if _, err := registry.update(rec.ID, func(r *CodexJobRecord) { r.Status = codexJobStatusRunning }); err != nil {
		codexLiveJobSessions.Delete(rec.ID)
		_ = session.close()
		cancelSession()
		return toolErr(codexTaskToolName, err), nil
	}

	// The job is DETACHED from the request context (t514 / GH #1687). The MCP
	// host ends that context when the handler returns, and for background=true
	// the handler returns immediately — so a job handed the request context was
	// cancelled within milliseconds of being created, every time, and reported
	// the failure as a 10-minute bound expiry. Values (the progress token among
	// them) are carried through; only the cancellation is dropped. The turn
	// stays bounded by the session deadline, including its handshake; the turn
	// also retains its own bound through runCodexTaskTurn.
	codexLiveJobCancels.Store(rec.ID, cancelSession)
	codexBackgroundJobs.Add(1)
	go runCodexBackgroundJob(sessionCtx, cancelSession, registry, rec.ID, session, turnParams)

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
func runCodexBackgroundJob(ctx context.Context, cancelSession context.CancelFunc, registry *codexJobRegistry, jobID string, session *codexSessionHandle, params map[string]any) {
	defer func() {
		codexLiveJobSessions.Delete(jobID)
		_ = session.close()
		codexLiveJobCancels.Delete(jobID)
		cancelSession()
		codexBackgroundJobs.Done()
	}()

	out, runErr := runCodexTaskTurn(ctx, session, params)
	out, runErr = codexBackgroundDeadlineResult(ctx, out, runErr)

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

func codexBackgroundDeadlineResult(ctx context.Context, out ReviewOutput, runErr error) (ReviewOutput, error) {
	if runErr != nil && errors.Is(context.Cause(ctx), errCodexTaskSessionDeadline) {
		// The child can close stdout at the deadline before the turn's
		// select observes ctx.Done. Report the deadline, not that EOF race.
		out.Summary = codexTaskTimeoutMessage()
		runErr = errors.New(out.Summary)
	}
	return out, runErr
}

// codexTaskRefusal shapes a refusal as REQ-CRS-010 requires: a structured
// result with status failed and the refusal's code, NOT an MCP tool error. The
// caller can act on it — pick a candidate, fix the input — so it is an answer,
// not a failed call.
func codexTaskRefusal(result CodexTaskResult, code, message string) *mcp.CallToolResult {
	result.Status = codexJobStatusFailed
	result.ErrorCode = code
	result.Error = message
	return toolJSON(codexTaskToolName, result)
}

// appendCodexNote joins two result notes, keeping both statements rather than
// letting the second silently replace the first.
func appendCodexNote(existing, add string) string {
	if existing == "" {
		return add
	}
	return existing + " " + add
}
