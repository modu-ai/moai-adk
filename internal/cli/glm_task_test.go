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

// withGLMTaskSeams swaps the key loader + the TASK-path HTTP client and
// restores them on cleanup. Mirrors withGLMSeams, but for the dedicated task
// seam: callGLMTask reads glmTaskHTTPClient (unbounded client — the bound is
// ctx-only), NOT the audit path's glmHTTPClient, so a task test that swaps
// only the audit seam would let the REAL client hit the network.
func withGLMTaskSeams(t *testing.T, key string, doer glmHTTPDoer) {
	t.Helper()
	oldKey, oldHTTP := glmKeyLoader, glmTaskHTTPClient
	glmKeyLoader = func() string { return key }
	if doer != nil {
		glmTaskHTTPClient = doer
	}
	t.Cleanup(func() {
		glmKeyLoader = oldKey
		glmTaskHTTPClient = oldHTTP
	})
}

// withShortGLMTaskTimeout shortens the task wall-clock bound so a deadline
// test does not wait the distributed 600s. This is the seam that proves the
// CONSTANT governs: with the bound well under the old audit-client ceiling
// (120s), a call that ends at the shortened bound demonstrates
// DefaultGLMTaskTimeout is the effective cap, not a hidden client Timeout.
func withShortGLMTaskTimeout(t *testing.T, d time.Duration) {
	t.Helper()
	prev := config.DefaultGLMTaskTimeout
	config.DefaultGLMTaskTimeout = d
	t.Cleanup(func() { config.DefaultGLMTaskTimeout = prev })
}

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
	withGLMTaskSeams(t, "test-glm-key", doer)

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
	withGLMTaskSeams(t, "test-glm-key", stub)

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
//
// It asserts the NON-factory path, so it must not inherit the ambient factory
// signal: a developer running the suite from inside a factory session carries
// MOAI_FACTORY_WORKERS in the environment, and the override-ignoring branch in
// glm_task.go keys on its mere presence. Without this clear the test reports a
// dropped override that the factory guard deliberately dropped — a false defect
// whose factory-mode counterpart is pinned separately by
// TestGLMTaskFactoryModeIgnoresModelOverride.
func TestGLMTask_ForwardsSystemModelAndMaxTokens(t *testing.T) {
	clearFactoryTestEnv(t)
	stub := &stubGLMDoer{body: glmTextResp("ok")}
	withGLMTaskSeams(t, "k", stub)

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
	withGLMTaskSeams(t, "k", &stubGLMDoer{body: glmTextResp("x")})

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
	withGLMTaskSeams(t, "", stub)

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
	withGLMTaskSeams(t, "k", &stubGLMDoer{err: errBoom})

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
	withGLMTaskSeams(t, "k", &stubGLMDoer{status: http.StatusUnauthorized, body: "{}"})

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
			withGLMTaskSeams(t, "k", &stubGLMDoer{body: body})

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
	withGLMTaskSeams(t, "k", &stubGLMDoer{body: glmTextResp("background done")})

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
	withGLMTaskSeams(t, "k", &stubGLMDoer{body: glmTextResp("ok")})

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

// ─── the task wall-clock bound (F1 regression) ───

// TestGLMTask_SyncDeadlineGovernsCall proves DefaultGLMTaskTimeout is the
// EFFECTIVE bound on the SYNC arm: with the constant shortened to 100ms (well
// under the audit client's old 120s ceiling) and a doer that blocks until its
// request context is revoked, the call must return the deadline failure AT the
// shortened bound — naming the bound that fired and coming back in seconds,
// not minutes. Against the defect this guards, the sync arm carried no
// WithTimeout at all and the shared 120s client was the only cap.
func TestGLMTask_SyncDeadlineGovernsCall(t *testing.T) {
	withShortGLMTaskTimeout(t, 100*time.Millisecond)
	withGLMTaskSeams(t, "k", newBlockingGLMDoer())

	start := time.Now()
	res := callGLMTaskTool(t, map[string]any{"prompt": "slow task"})
	elapsed := time.Since(start)

	if res.IsError {
		t.Fatalf("a deadline expiry must be a structured failed result, not IsError: %+v", res)
	}
	if elapsed >= 2*time.Second {
		t.Errorf("the call took %s; the shortened task bound (100ms) must govern, not a client ceiling", elapsed)
	}
	got := structuredMap(t, res)
	if st, _ := got["status"].(string); st != glmJobStatusFailed {
		t.Errorf("status = %q, want %q (deadline ⇒ fail-open)", st, glmJobStatusFailed)
	}
	if errText, _ := got["error"].(string); !strings.Contains(errText, "timed out after 100ms") {
		t.Errorf("error = %q, want it to name the bound that fired (timed out after 100ms)", errText)
	}
}

// TestGLMTask_BackgroundDeadlineRecordsFailed is the background-arm
// equivalent: the shortened bound ends the job's in-flight call and the
// goroutine records the failure with the bound named.
func TestGLMTask_BackgroundDeadlineRecordsFailed(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withShortGLMTaskTimeout(t, 100*time.Millisecond)
	withGLMTaskSeams(t, "k", newBlockingGLMDoer())

	res := callGLMTaskTool(t, map[string]any{"prompt": "slow background task", "background": true})
	if res.IsError {
		t.Fatalf("background glm_task IsError: %+v", res)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatal("background glm_task returned no job id")
	}
	t.Cleanup(func() { waitForGLMJobToStop(t, jobID) })

	reg := newGLMJobRegistry(root)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rec, err := reg.load(jobID)
		if err == nil && rec.Status == glmJobStatusFailed {
			if !strings.Contains(rec.Error, "timed out after 100ms") {
				t.Errorf("recorded error = %q, want it to name the bound that fired", rec.Error)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached failed within the shortened bound", jobID)
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

// TestGLMTaskFactoryModeIgnoresModelOverride is the Go half of the factory
// no-model-override discipline (t85 lead loop): when the MCP server process
// runs inside a factory session (MOAI_FACTORY_WORKERS inherited from the
// session env), a caller-supplied model is IGNORED in favor of the SSOT
// default and the result says so — a per-call override would split the
// session's caches and can bypass the ANTHROPIC_DEFAULT_*_MODEL slot-to-GLM
// tier mapping the launcher established. Outside factory mode the override
// is honored exactly as before.
func TestGLMTaskFactoryModeIgnoresModelOverride(t *testing.T) {
	stub := &stubGLMDoer{body: glmTextResp("ok")}
	withGLMTaskSeams(t, "test-glm-key", stub)

	t.Run("factory mode ignores the override", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "8")

		got := structuredMap(t, callGLMTaskTool(t, map[string]any{
			"prompt": "do the thing",
			"model":  "caller-picked-model",
		}))
		if m, _ := got["model"].(string); m != resolveGLMTaskModel() {
			t.Errorf("model = %q, want the resolved SSOT default %q", m, resolveGLMTaskModel())
		}
		note, _ := got["note"].(string)
		if !strings.Contains(note, "factory mode") || !strings.Contains(note, "caller-picked-model") {
			t.Errorf("note = %q, want it to name the ignored override and factory mode", note)
		}
	})

	t.Run("outside factory mode the override is honored", func(t *testing.T) {
		clearFactoryTestEnv(t)

		got := structuredMap(t, callGLMTaskTool(t, map[string]any{
			"prompt": "do the thing",
			"model":  "caller-picked-model",
		}))
		if m, _ := got["model"].(string); m != "caller-picked-model" {
			t.Errorf("model = %q, want the caller's override honored outside factory mode", m)
		}
		if note, _ := got["note"].(string); note != "" {
			t.Errorf("note = %q, want none outside factory mode", note)
		}
	})
}
