// Package cli — GLM (z.ai) task delegation.
//
// glm_task.go owns the task-delegation surface for the GLM backend: it sends
// the caller's prompt to z.ai and returns the model's text (sync form), or
// hands the call to a background goroutine and returns a job id (background
// form). It mirrors the codex task surface (codex_task.go) — the structural
// SSOT for the delegation-family task layer — and sits ON TOP of the z.ai
// plumbing in mcp_glm.go (glmMessagesRequest type, endpoint + header
// constants, the glmHTTPDoer abstraction). Its HTTP CLIENT seam is
// deliberately its own (glmTaskHTTPClient, no client Timeout) so the task
// bound is not shadowed by the audit client's ceiling.
//
// Two properties are load-bearing (mirrored from the codex surface):
//
//   - a background job is a GOROUTINE inside this server process, not a
//     detached subprocess. Every in-flight job is lost when the server exits,
//     and a record found running with no glmLiveJobs entry is stale by
//     construction — the cancel path refuses it rather than acting on it.
//
//   - the wall-clock bound is DefaultGLMTaskTimeout, enforced on BOTH arms via
//     context.WithTimeout. The task path deliberately uses its own HTTP client
//     with NO client Timeout (glmTaskHTTPClient below): http.Client.Timeout
//     caps the ENTIRE request including the body read, so reusing the audit
//     path's 120s client would silently re-cap the configured task bound —
//     exactly the shadowing this seam exists to prevent.
//
//   - fail-open. GLM is OPTIONAL: a missing key, an unauthenticated endpoint,
//     a transport failure, and a malformed response are all structured failed
//     results the orchestrator translates — never a Go error and never a
//     panic. The two arms that DO set IsError are caller-actionable faults: a
//     missing prompt, and a state directory that cannot be written.
//
// Nothing here invokes AskUserQuestion (subagent boundary).
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// glmTaskToolName is the MCP tool name.
	glmTaskToolName = "glm_task"

	// glmTaskAgentKey is the profile-matrix agent key used to resolve the
	// default task model + effort via the SSOT (template.ResolveAgentModelEffort).
	// super-advisor is the GLM delegation family's wired consumer, so a generic
	// task resolves on the same matrix cell its caller runs under — the
	// auditor-shaped glmAuditAgentKey is NOT reused, because a task is not a
	// review and should not inherit an auditor's model choice.
	glmTaskAgentKey = "super-advisor"
)

// glmLiveJobs holds the cancel function of every RUNNING background job, keyed
// by job id. It is the seam the cancel path reads: cancelling a GLM job means
// revoking the context its in-flight HTTP call runs under, and only the
// goroutine this server lifetime started has one here.
//
// It is in-process state by construction, matching the in-process execution
// model: an entry exists only for a job this server lifetime started and is
// removed the moment the job reaches a terminal status. A record found running
// with no entry here is stale (a previous server lifetime), which is exactly
// the case the cancel path refuses rather than acts on.
var glmLiveJobs sync.Map // job id (string) → context.CancelFunc

// glmTaskHTTPClient is the task-path HTTP seam. It deliberately carries NO
// client Timeout (Timeout 0 = unbounded): DefaultGLMTaskTimeout is the sole
// wall-clock bound on a glm_task call and is enforced through the request
// context in BOTH arms (sync and background), so the configured bound is the
// one that actually fires. A client Timeout here would shadow it the way the
// shared audit client's 120s ceiling (glmAuditHTTPTimeout) otherwise would.
// The audit path keeps glmHTTPClient and its 120s ceiling untouched — the two
// surfaces bound different callers and stay deliberately separate.
var glmTaskHTTPClient glmHTTPDoer = &http.Client{}

// glmTaskTimeoutMessage names the bound that ended a call. It reads the
// duration at call time rather than baking it in, so a shortened bound
// reports the value that actually fired (mirrors codexTaskTimeoutMessage).
func glmTaskTimeoutMessage() string {
	return "glm_task timed out after " + config.DefaultGLMTaskTimeout.String() +
		" (the bound glm_task imposes on its own calls); the request was abandoned"
}

// GLMTaskResult is the structured result glm_task returns. It mirrors
// CodexTaskResult's shape minus the fields with no GLM counterpart
// (thread/turn ids, the write opt-in, thread resume): a task produces output,
// and its provenance is the model it ran on.
type GLMTaskResult struct {
	// Status is a job-status value (glmJobStatus*). A foreground task reports
	// its terminal status; a background task reports the status at hand-off.
	Status string `json:"status"`

	// Background reports which form ran. JobID is set only for the background
	// form; Output only for the foreground form.
	Background bool   `json:"background"`
	JobID      string `json:"job_id,omitempty"`
	Output     string `json:"output,omitempty"`

	// Model is the z.ai model id the task ran on (the caller's override or the
	// resolved default), echoed so the caller knows what produced the output.
	Model string `json:"model"`

	// Note carries a human-readable statement; Error names a failure the tool
	// absorbed rather than raised.
	Note  string `json:"note,omitempty"`
	Error string `json:"error,omitempty"`
}

// resolveGLMTaskModel resolves the default GLM task model via the model/effort
// SSOT, keyed on the task family's consumer (glmTaskAgentKey). Same rule as
// the audit resolver: a GLM session uses the matrix cell, anything else falls
// back to the canonical GLM model.
func resolveGLMTaskModel() string {
	return resolveGLMModelForAgent(glmTaskAgentKey)
}

// handleGLMTask is the handler for the `glm_task` MCP tool.
//
// Foreground (background=false): drives the call to completion and returns the
// model's text. Background (background=true): creates the job record, hands
// the call to a goroutine, and returns the job id immediately.
//
// Fail-open: a missing or unreachable GLM yields a structured result carrying
// the failure, never a Go error. The two arms that DO set IsError are
// caller-actionable faults — a missing prompt, and a state directory that
// cannot be written.
func handleGLMTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt := req.GetString("prompt", "")
	background := req.GetBool("background", false)
	token := extractProgressToken(req)
	notifyMCPProgress(ctx, token, 0, "glm 태스크 시작 — 프롬프트 접수")

	if prompt == "" {
		return toolErr(glmTaskToolName, errors.New("prompt is required")), nil
	}

	result := GLMTaskResult{Background: background}

	// Fail-open: a missing key short-circuits BEFORE any HTTP call (the same
	// ordering as glm_audit). The workflow decides what to do without GLM.
	key := glmKeyLoader()
	if key == "" {
		result.Status = glmJobStatusFailed
		result.Error = "GLM API key not configured (~/.moai/.env.glm)"
		return toolJSON(glmTaskToolName, result), nil
	}

	model := req.GetString("model", "")
	if model == "" {
		model = resolveGLMTaskModel() // SSOT, keyed on glmTaskAgentKey
	}
	result.Model = model

	system := req.GetString("system", "")
	maxTokens := req.GetInt("max_tokens", 0)
	if maxTokens <= 0 {
		maxTokens = config.DefaultGLMTaskMaxTokens
	}

	if !background {
		// The task bound applies to the sync form too (the codex mirror's
		// runCodexTaskTurn bounds both forms): DefaultGLMTaskTimeout is the
		// sole wall-clock cap, enforced here because the caller's context may
		// carry no deadline at all and the task-path HTTP client carries none.
		ctx, cancel := context.WithTimeout(ctx, config.DefaultGLMTaskTimeout)
		defer cancel()
		output, err := callGLMTask(ctx, key, model, prompt, system, maxTokens, token)
		if err != nil {
			result.Status = glmJobStatusFailed
			result.Error = err.Error()
			return toolJSON(glmTaskToolName, result), nil
		}
		result.Status = glmJobStatusCompleted
		result.Output = output
		return toolJSON(glmTaskToolName, result), nil
	}

	// Background: the record is created BEFORE the call is handed off, so an
	// unwritable state directory is reported to the caller as a structured
	// error rather than surfacing later as a job nobody can observe.
	registry := newGLMJobRegistry(projectDirResolver())
	rec, err := registry.create(glmJobSpec{
		Model:          model,
		RequestSummary: prompt,
	})
	if err != nil {
		return toolErr(glmTaskToolName, err), nil
	}

	// The live entry is the cancel handle: revoking this context aborts the
	// in-flight HTTP call the job goroutine waits on.
	jobCtx, cancel := context.WithCancel(ctx)
	glmLiveJobs.Store(rec.ID, cancel)

	if _, err := registry.update(rec.ID, func(r *GLMJobRecord) { r.Status = glmJobStatusRunning }); err != nil {
		glmLiveJobs.Delete(rec.ID)
		cancel()
		return toolErr(glmTaskToolName, err), nil
	}

	go runGLMBackgroundJob(jobCtx, cancel, registry, rec.ID, key, model, prompt, system, maxTokens)

	result.Status = glmJobStatusRunning
	result.JobID = rec.ID
	return toolJSON(glmTaskToolName, result), nil
}

// runGLMBackgroundJob drives one background job's call to completion and
// records the outcome. It runs as a goroutine inside this server process, so
// the job dies with the server; nothing here attempts to survive that.
//
// The live-map entry is removed and the cancel context released on EVERY exit
// path, so a terminal record is never left with a live entry the cancel path
// could still address.
func runGLMBackgroundJob(ctx context.Context, cancel context.CancelFunc, registry *glmJobRegistry, jobID, key, model, prompt, system string, maxTokens int) {
	defer func() {
		glmLiveJobs.Delete(jobID)
		cancel() // release the WithCancel context; idempotent once fired
	}()

	// The bound the tool imposes on its own calls — same constant, same
	// enforcement, as the sync arm above. It is REAL: glmTaskHTTPClient
	// carries no client Timeout, so nothing shadows it.
	ctx, cancelTimeout := context.WithTimeout(ctx, config.DefaultGLMTaskTimeout)
	defer cancelTimeout()

	output, err := callGLMTask(ctx, key, model, prompt, system, maxTokens, nil)

	// A job cancelled while the call was in flight keeps its cancelled status:
	// the call returning afterwards must not overwrite it with completed or
	// failed. The guard lives in the REGISTRY (updateUnlessCancelled), because
	// a mutator can only decline to change the record — it cannot decline the
	// write, and a write landing after glm_job_cancel returned is exactly what
	// has to stop.
	if err != nil {
		_, _ = registry.updateUnlessCancelled(jobID, func(r *GLMJobRecord) {
			r.Status = glmJobStatusFailed
			r.Error = err.Error()
		})
		return
	}
	_, _ = registry.updateUnlessCancelled(jobID, func(r *GLMJobRecord) {
		r.Status = glmJobStatusCompleted
		r.Output = output
	})
}

// callGLMTask posts the prompt to z.ai and returns the model's first text
// content block as-is — a task produces plain text, not a ReviewOutput, so no
// JSON-schema parsing or fence-stripping applies. Every failure is a named
// error the caller renders into a structured failed result (fail-open); a
// deadline expiry is reported by name via glmTaskTimeoutMessage.
//
// It is the task-layer sibling of callGLMAudit: the z.ai plumbing
// (glmMessagesRequest type, endpoint + header constants, the glmHTTPDoer
// abstraction) is REUSED, not forked — but the CLIENT seam is deliberately
// separate (glmTaskHTTPClient, no client Timeout) so DefaultGLMTaskTimeout
// is the bound that governs, not the audit client's 120s ceiling. The two
// paths differ in what they do with the response (audit parses a ReviewOutput
// and fails open to VerdictInconclusive; task returns raw text and reports
// errors upward), which is why the send cores are siblings rather than one
// shared function.
func callGLMTask(ctx context.Context, key, model, prompt, system string, maxTokens int, token mcp.ProgressToken) (string, error) {
	body, err := json.Marshal(glmMessagesRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []glmMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("cannot build z.ai request: %w", err)
	}

	url := config.DefaultGLMBaseURL + glmMessagesPath
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("cannot build z.ai request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("anthropic-version", glmAnthropicVersion)

	notifyMCPProgress(ctx, token, 0.1, "z.ai에 태스크 요청 전송 중...")
	resp, err := glmTaskHTTPClient.Do(httpReq)
	if err != nil {
		// The task bound firing is a named, structured failure — never a raw
		// "Client.Timeout exceeded" and never a Go error escaping the handler.
		if errors.Is(err, context.DeadlineExceeded) {
			return "", errors.New(glmTaskTimeoutMessage())
		}
		return "", fmt.Errorf("z.ai request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Unauthenticated (401/403) / rate-limited / server error ⇒ named failure.
		return "", fmt.Errorf("z.ai returned HTTP %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read z.ai response: %w", err)
	}
	notifyMCPProgress(ctx, token, 0.8, "z.ai 응답 수신 — 텍스트 추출 중...")

	var env glmMessagesResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", fmt.Errorf("malformed z.ai response: %w", err)
	}
	if len(env.Content) == 0 || env.Content[0].Text == "" {
		return "", errors.New("z.ai response carried no content")
	}
	return env.Content[0].Text, nil
}
