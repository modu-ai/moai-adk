// todo_classify_llm.go — SPEC-TCD-LLM-DECIDER-001 M2: the LLM-backed
// classification decider (REQ-TLD-001): one in-process HTTPS call to the
// project's configured GLM endpoint, authenticated through the existing GLM
// credential reader, validated through the single classification validator
// with the decider identity FORCED to llm before validation (REQ-TLD-004 —
// a response claiming human/jev is NOT a failure; the identity is
// overwritten and the judgment accepted).
//
// Every LLM-side failure — absent credentials, connection failure, non-2xx,
// timeout, non-JSON, out-of-set fields — is an error return. The fallback
// branch stays in the seam (todo_classify.go's notice + fail-safe default),
// never synthesized inside the decider (plan OD-C): the entry points print
// the ONE notice and promote the defaults.
//
// The transport mirrors the mcp_glm.go pattern (plan OD-B): the same
// Anthropic-compatible endpoint shape and credential reader, its own
// injectable endpoint/doer/key-loader seams, its own (shorter) timeout from
// internal/config/defaults.go.
//
// @MX:SPEC: SPEC-TCD-LLM-DECIDER-001
package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoLLMHTTPDoer abstracts the classification POST so tests inject a canned
// response without any network — the mcp_glm.go glmHTTPDoer pattern, with
// its own seam variable so the audit and classification paths swap
// independently.
type todoLLMHTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// The decider's transport seams. Production resolves the configured GLM
// endpoint, an *http.Client bounded by the config default, and the existing
// GLM credential reader; tests aim all three at fakes (plan §D.5: zero real
// network in the suite).
var (
	// todoLLMEndpoint resolves the classification request URL — a var so
	// tests aim the decider at an httptest server. Production returns the
	// configured GLM endpoint; the path constant is shared with the audit
	// path so the spelling stays single-sourced (plan §D.2).
	todoLLMEndpoint = func() string {
		return config.DefaultGLMBaseURL + glmMessagesPath
	}

	todoLLMHTTPClient todoLLMHTTPDoer = &http.Client{Timeout: config.DefaultTodoClassifyLLMTimeout}

	todoLLMKeyLoader = loadGLMKey
)

// todoLLMJudgment is the JSON object the system prompt constrains the model
// to emit. The claimed decider field is parsed only to be discarded — the
// recorded identity is forced to llm (REQ-TLD-004).
type todoLLMJudgment struct {
	Priority string `json:"priority"`
	Blocked  bool   `json:"blocked"`
	Mode     string `json:"mode"`
	Decider  string `json:"decider"`
	Reason   string `json:"reason"`
}

// llmCardDecider implements kanban.CardDecider over the configured GLM
// endpoint (REQ-TLD-001). It is stateless: the endpoint, credential, HTTP
// client, and timeout resolve through the package seams above.
type llmCardDecider struct{}

// newLLMCardDecider returns the LLM-backed decider the selector installs.
func newLLMCardDecider() llmCardDecider { return llmCardDecider{} }

// todoDeciderIsLLM reports whether dec is the LLM-backed decider whose
// judgment the entry points compute BEFORE the queue lock (REQ-TLD-005 /
// plan OD-D). Every other decider — deterministic, static, unavailable —
// resolves inside the locked write as before.
func todoDeciderIsLLM(dec kanban.CardDecider) bool {
	_, ok := dec.(llmCardDecider)
	return ok
}

// Classify asks the configured GLM endpoint to judge the card text and
// returns the validated classification with its decider identity forced to
// llm. Every failure is an error return — the caller's seam owns the
// fail-safe degradation (OD-C).
func (llmCardDecider) Classify(text string) (kanban.CardClassification, error) {
	key := todoLLMKeyLoader()
	if key == "" {
		return kanban.CardClassification{}, errors.New("llm decider: GLM credential not configured")
	}
	body, err := json.Marshal(glmMessagesRequest{
		Model:     glmTaskDefaultModel,
		MaxTokens: config.DefaultTodoClassifyLLMMaxTokens,
		System:    todoLLMClassifySystemPrompt(),
		Messages:  []glmMessage{{Role: "user", Content: text}},
	})
	if err != nil {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: build request: %w", err)
	}
	httpReq, err := http.NewRequest(http.MethodPost, todoLLMEndpoint(), bytes.NewReader(body))
	if err != nil {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("anthropic-version", glmAnthropicVersion)
	resp, err := todoLLMHTTPClient.Do(httpReq)
	if err != nil {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: endpoint returned HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: read response: %w", err)
	}
	return parseTodoLLMJudgment(raw)
}

// parseTodoLLMJudgment extracts and validates the judgment from a response
// envelope. The decider identity is forced to llm BEFORE validation, so a
// claimed human/jev identity never fails the judgment and never reaches the
// record (REQ-TLD-004).
func parseTodoLLMJudgment(raw []byte) (kanban.CardClassification, error) {
	var env glmMessagesResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: malformed response: %w", err)
	}
	text := ""
	for i := range env.Content {
		if (env.Content[i].Type == "" || env.Content[i].Type == "text") && env.Content[i].Text != "" {
			text = env.Content[i].Text
			break
		}
	}
	if text == "" {
		return kanban.CardClassification{}, errors.New("llm decider: response carried no text content")
	}
	var j todoLLMJudgment
	if err := json.Unmarshal([]byte(extractJSONObject(text)), &j); err != nil {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: judgment json: %w", err)
	}
	cls := kanban.CardClassification{
		Priority: j.Priority,
		Blocked:  j.Blocked,
		Mode:     j.Mode,
		Decider:  kanban.DeciderIdentityLLM,
		Reason:   todoLLMSingleLineReason(j.Reason),
	}
	// The single validator, single spelling (C3): the closed sets are
	// judged only here, through the existing kanban validator.
	if err := kanban.ValidateCardClassification(cls); err != nil {
		return kanban.CardClassification{}, fmt.Errorf("llm decider: %w", err)
	}
	return cls, nil
}

// todoLLMSingleLineReason bounds the recorded reason to a single line: the
// model's one-line rationale, with any embedded line break collapsed.
func todoLLMSingleLineReason(reason string) string {
	reason = strings.ReplaceAll(reason, "\r", " ")
	reason = strings.ReplaceAll(reason, "\n", " ")
	return strings.TrimSpace(reason)
}

// todoLLMClassifySystemPrompt constrains the model to emit ONLY the
// judgment JSON over the closed sets — defensive output validation: the
// response is schema-validated before use, never acted on raw.
func todoLLMClassifySystemPrompt() string {
	return "You classify a work card for a task queue. Respond with ONLY a single " +
		"JSON object (no prose, no code fences) with exactly these fields: " +
		`{"priority":"high|normal|low","blocked":true|false,` +
		`"mode":"serial|parallelizable","reason":"<one line>"}. ` +
		"priority high = urgent or multi-system, normal = default, low = cosmetic " +
		"or deferrable. blocked true = waiting on another card or an external " +
		"system. mode parallelizable = safe to run concurrently with other cards, " +
		"serial = must not overlap other cards. reason is one short line naming " +
		"the driver of the judgment."
}

// todoPreClassifyLLM resolves the LLM decider's judgment BEFORE the caller's
// locked write (REQ-TLD-005, plan OD-D): an in-lock LLM round-trip
// serializes concurrent adds behind the queue lock for a judgment-duration
// each and holds every locked reader behind a network stall. A non-LLM
// decider passes through unchanged. The LLM judgment — healthy or fail-safe
// — rides a StaticCardDecider into the lock, so todoClassifyInLock
// re-stamps classified-at at the write and the visibility invariant
// (REQ-TCD-001) is untouched: only the computation moved, never the
// attachment. On judgment failure the ONE seam notice prints here and the
// fail-safe default is carried — the fallback branch stays in the seam
// (OD-C), never synthesized inside the decider.
func todoPreClassifyLLM(dec kanban.CardDecider, text string, errOut io.Writer) kanban.CardDecider {
	llm, ok := dec.(llmCardDecider)
	if !ok {
		return dec
	}
	cls, err := llm.Classify(text)
	if err != nil {
		_, _ = fmt.Fprintln(errOut, todoClassificationFallbackNotice)
		cls = kanban.DefaultCardClassification()
	}
	return kanban.StaticCardDecider{Class: cls}
}
