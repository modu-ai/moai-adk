package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// glm_task + glm job registry tests. Mirrors the SPEC-CODEX-PHASE2-001 M3
// task-layer tests (codex_task_test.go): these exercise the TASK layer over
// the existing z.ai plumbing (mcp_glm.go seams — glmKeyLoader / glmHTTPClient),
// never a second HTTP client.
//
// Tests MUST NOT require a real GLM key or network. Every fail-open path
// returns a structured failed result, never a Go error, never IsError, never
// AskUserQuestion (subagent boundary).

// ─── fixtures ───

// glmTextResp builds an Anthropic-compatible /v1/messages response whose
// single text block carries the given plain payload. glm_task returns raw
// text (NOT a ReviewOutput JSON), so the helper takes a string, not a review.
func glmTextResp(text string) string {
	b, _ := json.Marshal(map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	})
	return string(b)
}

// blockingGLMDoer is an injectable HTTP doer whose Do blocks until released OR
// the request context is revoked — mirroring how a real http.Client aborts an
// in-flight request when its context is canceled. A stub that returns would
// complete the task, and a completed task is exactly the case cancellation
// does not exercise (same rationale as hangingCodexSession).
type blockingGLMDoer struct {
	release chan struct{}
	once    sync.Once
}

func newBlockingGLMDoer() *blockingGLMDoer {
	return &blockingGLMDoer{release: make(chan struct{})}
}

func (s *blockingGLMDoer) Do(req *http.Request) (*http.Response, error) {
	select {
	case <-s.release:
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(glmTextResp("late output"))),
			Header:     make(http.Header),
		}, nil
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}
}

// unblock releases the blocked Do so a job goroutine waiting on it can finish.
func (s *blockingGLMDoer) unblock() { s.once.Do(func() { close(s.release) }) }

// callGLMTaskTool invokes the handler with the given arguments. Named for the
// TOOL (the codex sibling is callCodexTask) so it cannot collide with the
// production HTTP core callGLMTask.
func callGLMTaskTool(t *testing.T, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := handleGLMTask(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	})
	if err != nil {
		t.Fatalf("handleGLMTask returned a Go error (the tool must return structured results only): %v", err)
	}
	if res == nil {
		t.Fatal("handleGLMTask returned a nil result")
	}
	return res
}

// waitForGLMJobToStop blocks until the job has left the live-job map, which is
// the join point for the goroutine runGLMBackgroundJob runs in: that goroutine
// writes its terminal registry transition BEFORE the deferred Delete, so an
// absent entry means every write it will ever make has already landed.
func waitForGLMJobToStop(t *testing.T, jobID string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, live := glmLiveJobs.Load(jobID); !live {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Errorf("job %s goroutine did not stop within the deadline; it would outlive the test", jobID)
}

// startHangingGLMJob starts a background glm_task whose HTTP call blocks until
// released (or its context is revoked), and returns the job id plus the
// registry rooted at root.
//
// This helper starts a goroutine (the background job), so it owns joining it.
// The join cleanup is registered AFTER withGLMSeams so LIFO cleanup order runs
// it FIRST: release the blocked call, then wait for its goroutine to stop
// before t.TempDir's RemoveAll runs (mirrors startHangingBackgroundJob).
func startHangingGLMJob(t *testing.T, root string) (string, *glmJobRegistry) {
	t.Helper()
	// The project-root seam (projectDirResolver) is shared with the codex
	// family; withCodexProjectDir swaps exactly that seam.
	withCodexProjectDir(t, root)
	doer := newBlockingGLMDoer()
	withGLMSeams(t, "test-glm-key", doer)

	res := callGLMTaskTool(t, map[string]any{"prompt": "summarize the queue", "background": true})
	if res.IsError {
		t.Fatalf("background glm_task returned IsError: %+v", res)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background glm_task returned no job id: %+v", res)
	}
	reg := newGLMJobRegistry(root)

	t.Cleanup(func() {
		doer.unblock()
		waitForGLMJobToStop(t, jobID)
	})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if rec, err := reg.load(jobID); err == nil && rec.Status == glmJobStatusRunning {
			return jobID, reg
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached running before the deadline", jobID)
	return "", nil
}

// ─── sync form ───

// TestGLMTask_SyncReturnsOutput proves the sync form drives the call to
// completion and returns its raw text output, hitting the z.ai
// Anthropic-compatible endpoint directly with the prompt and the default
// token bound.
func TestGLMTask_SyncReturnsOutput(t *testing.T) {
	stub := &stubGLMDoer{body: glmTextResp("the answer is 42")}
	withGLMSeams(t, "test-glm-key", stub)

	res := callGLMTaskTool(t, map[string]any{"prompt": "what is the answer"})
	if res.IsError {
		t.Fatalf("sync glm_task IsError: %+v", res)
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusCompleted {
		t.Errorf("status = %q, want %q", st, glmJobStatusCompleted)
	}
	if bg, _ := got["background"].(bool); bg {
		t.Error("background = true, want false for the sync form")
	}
	if out, _ := got["output"].(string); out != "the answer is 42" {
		t.Errorf("output = %q, want the stub text", out)
	}
	if m, _ := got["model"].(string); m == "" {
		t.Error("result carries no resolved model")
	}

	// The request must target the direct z.ai Anthropic endpoint — NOT the
	// z.ai MCP server, NOT any gateway.
	if !strings.HasPrefix(stub.gotURL, "https://api.z.ai/api/anthropic/v1/messages") {
		t.Errorf("URL = %q, want prefix https://api.z.ai/api/anthropic/v1/messages", stub.gotURL)
	}
	if strings.Contains(stub.gotURL, "/api/mcp/") {
		t.Errorf("URL %q hits the z.ai MCP server, not the direct Anthropic endpoint", stub.gotURL)
	}

	var sent glmMessagesRequest
	if err := json.Unmarshal([]byte(stub.gotBody), &sent); err != nil {
		t.Fatalf("request body is not a glmMessagesRequest: %v (body=%q)", err, stub.gotBody)
	}
	if len(sent.Messages) != 1 || sent.Messages[0].Role != "user" || sent.Messages[0].Content != "what is the answer" {
		t.Errorf("messages = %+v, want one user turn carrying the prompt", sent.Messages)
	}
	if sent.MaxTokens != config.DefaultGLMTaskMaxTokens {
		t.Errorf("max_tokens = %d, want the default %d", sent.MaxTokens, config.DefaultGLMTaskMaxTokens)
	}
}

// TestGLMTask_ForwardsSystemModelAndMaxTokens proves the optional arguments
// reach the request body and the echoed model, rather than being silently
// dropped or overridden by the resolver.
func TestGLMTask_ForwardsSystemModelAndMaxTokens(t *testing.T) {
	stub := &stubGLMDoer{body: glmTextResp("ok")}
	withGLMSeams(t, "k", stub)

	res := callGLMTaskTool(t, map[string]any{
		"prompt":     "do the thing",
		"system":     "be terse",
		"model":      "glm-test-model",
		"max_tokens": 512,
	})
	if res.IsError {
		t.Fatalf("glm_task IsError: %+v", res)
	}
	if m, _ := structuredMap(t, res)["model"].(string); m != "glm-test-model" {
		t.Errorf("echoed model = %q, want the explicit override", m)
	}

	var sent glmMessagesRequest
	if err := json.Unmarshal([]byte(stub.gotBody), &sent); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if sent.System != "be terse" {
		t.Errorf("system = %q, want the caller-supplied system prompt", sent.System)
	}
	if sent.Model != "glm-test-model" {
		t.Errorf("model = %q, want the explicit override", sent.Model)
	}
	if sent.MaxTokens != 512 {
		t.Errorf("max_tokens = %d, want 512", sent.MaxTokens)
	}
}

// TestGLMTask_MissingPromptIsStructuredError proves a missing prompt is the
// one caller-actionable fault that sets IsError (mirroring codex_task).
func TestGLMTask_MissingPromptIsStructuredError(t *testing.T) {
	withGLMSeams(t, "k", &stubGLMDoer{body: glmTextResp("x")})

	res := callGLMTaskTool(t, map[string]any{})
	if !res.IsError {
		t.Error("a missing prompt must return a structured IsError result")
	}
}

// ─── fail-open ───

// TestGLMTask_FailOpenOnMissingKey proves a missing key is a structured failed
// result (never IsError, never a Go error) and that no HTTP call is made — the
// key check short-circuits before z.ai.
func TestGLMTask_FailOpenOnMissingKey(t *testing.T) {
	stub := &stubGLMDoer{body: glmTextResp("x")}
	withGLMSeams(t, "", stub)

	res := callGLMTaskTool(t, map[string]any{"prompt": "p"})
	if res.IsError {
		t.Fatalf("missing key must fail-open to a structured result, not IsError: %+v", res)
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusFailed {
		t.Errorf("status = %q, want %q (fail-open)", st, glmJobStatusFailed)
	}
	if errText, _ := got["error"].(string); !strings.Contains(errText, "~/.moai/.env.glm") {
		t.Errorf("error = %q, want it to name the key file ~/.moai/.env.glm", errText)
	}
	if stub.gotURL != "" {
		t.Errorf("missing key still hit z.ai at %q", stub.gotURL)
	}
}

// TestGLMTask_FailOpenOnHTTPError proves a transport failure is a structured
// failed result, not a hard error.
func TestGLMTask_FailOpenOnHTTPError(t *testing.T) {
	withGLMSeams(t, "k", &stubGLMDoer{err: errBoom})

	res := callGLMTaskTool(t, map[string]any{"prompt": "p"})
	if res.IsError {
		t.Fatalf("transport error must fail-open to a structured result, not IsError")
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusFailed {
		t.Errorf("status = %q, want %q (fail-open)", st, glmJobStatusFailed)
	}
	if errText, _ := got["error"].(string); !strings.Contains(errText, "boom") {
		t.Errorf("error = %q, want it to carry the transport failure", errText)
	}
}

// TestGLMTask_FailOpenOnUnauthenticatedStatus proves a 401 from z.ai is a
// structured failed result (an unauthenticated GLM is an unavailable GLM).
func TestGLMTask_FailOpenOnUnauthenticatedStatus(t *testing.T) {
	withGLMSeams(t, "k", &stubGLMDoer{status: http.StatusUnauthorized, body: "{}"})

	res := callGLMTaskTool(t, map[string]any{"prompt": "p"})
	if res.IsError {
		t.Fatalf("HTTP 401 must fail-open to a structured result, not IsError")
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusFailed {
		t.Errorf("status = %q, want %q on 401", st, glmJobStatusFailed)
	}
	if errText, _ := got["error"].(string); !strings.Contains(errText, "401") {
		t.Errorf("error = %q, want it to name the HTTP status", errText)
	}
}

// TestGLMTask_FailOpenOnMalformedResponse proves a body that is not a
// /v1/messages envelope, and one that carries no content block, are both
// structured failed results rather than empty-output "successes".
func TestGLMTask_FailOpenOnMalformedResponse(t *testing.T) {
	for name, body := range map[string]string{
		"not json":         "<<<not-json>>>",
		"no content":       "{}",
		"empty text block": glmTextResp(""),
	} {
		t.Run(name, func(t *testing.T) {
			withGLMSeams(t, "k", &stubGLMDoer{body: body})

			res := callGLMTaskTool(t, map[string]any{"prompt": "p"})
			if res.IsError {
				t.Fatalf("malformed response must fail-open to a structured result, not IsError")
			}
			got := structuredMap(t, res)
			if st, _ := got["status"].(string); st != glmJobStatusFailed {
				t.Errorf("status = %q, want %q (malformed ⇒ fail-open)", st, glmJobStatusFailed)
			}
			if errText, _ := got["error"].(string); errText == "" {
				t.Error("the failure must be named, not silent")
			}
		})
	}
}

// ─── background form ───

// TestGLMTask_BackgroundLifecycle proves the background form hands back a job
// id immediately with status running, the goroutine records the completed
// output, and the live-map entry is gone once the job ends.
func TestGLMTask_BackgroundLifecycle(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withGLMSeams(t, "k", &stubGLMDoer{body: glmTextResp("background done")})

	res := callGLMTaskTool(t, map[string]any{"prompt": "long task", "background": true})
	if res.IsError {
		t.Fatalf("background glm_task IsError: %+v", res)
	}
	got := structuredMap(t, res)
	jobID, _ := got["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background glm_task returned no job id: %+v", got)
	}
	if st, _ := got["status"].(string); st != glmJobStatusRunning {
		t.Errorf("status = %q, want %q at hand-off", st, glmJobStatusRunning)
	}
	if bg, _ := got["background"].(bool); !bg {
		t.Error("background = false, want true")
	}
	if out, _ := got["output"].(string); out != "" {
		t.Errorf("the background hand-off must carry no output; got %q", out)
	}
	t.Cleanup(func() { waitForGLMJobToStop(t, jobID) })

	reg := newGLMJobRegistry(root)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rec, err := reg.load(jobID)
		if err == nil && rec.Status == glmJobStatusCompleted && rec.Output == "background done" {
			if _, live := glmLiveJobs.Load(jobID); live {
				t.Error("live-map entry must be gone after the job ended")
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached completed with its output", jobID)
}

// TestGLMTask_BackgroundRecordRedactsSecrets proves the job record carries a
// redacted, bounded request summary: a credential-shaped prompt value must not
// survive into the record.
func TestGLMTask_BackgroundRecordRedactsSecrets(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withGLMSeams(t, "k", &stubGLMDoer{body: glmTextResp("ok")})

	res := callGLMTaskTool(t, map[string]any{
		"prompt":     "use api_key=supersecret123 then summarize the queue",
		"background": true,
	})
	if res.IsError {
		t.Fatalf("background glm_task IsError: %+v", res)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	t.Cleanup(func() { waitForGLMJobToStop(t, jobID) })

	reg := newGLMJobRegistry(root)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rec, err := reg.load(jobID)
		if err == nil && rec.Status == glmJobStatusCompleted {
			if !strings.Contains(rec.RequestSummary, "[redacted]") {
				t.Errorf("request summary = %q, want a redacted credential", rec.RequestSummary)
			}
			if strings.Contains(rec.RequestSummary, "supersecret123") {
				t.Errorf("request summary leaked the credential: %q", rec.RequestSummary)
			}
			if rec.Model == "" {
				t.Error("record carries no model")
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached completed", jobID)
}

// ─── model resolution ───

// TestResolveGLMTaskModel_SSOT proves model resolution goes through the SSOT
// (template.ResolveAgentModelEffort) and falls back to the documented GLM
// default when no llm.yaml is resolvable — never a Claude id.
func TestResolveGLMTaskModel_SSOT(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	old := projectDirResolver
	projectDirResolver = func() string { return "" } // no sections dir available
	t.Cleanup(func() { projectDirResolver = old })

	m := resolveGLMTaskModel()
	if m == "" {
		t.Fatal("resolveGLMTaskModel returned empty for a missing llm.yaml (want the GLM default)")
	}
	if m != config.DefaultGLMHigh {
		t.Errorf("resolveGLMTaskModel = %q, want the documented fallback %q", m, config.DefaultGLMHigh)
	}
	if strings.HasPrefix(m, "claude") {
		t.Errorf("resolveGLMTaskModel = %q; a Claude id cannot be a GLM default", m)
	}
}
