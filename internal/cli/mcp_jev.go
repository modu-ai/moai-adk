// mcp_jev.go — SPEC-JEV-GOAL-DIST-001 M8a. The moai-mcp tool wrapper over
// internal/jev: a thin CALLER, not a layer. It holds no transport code of its
// own — request construction, size bounds, secret screening, retry policy,
// and credential handling all belong to internal/jev, so there is exactly one
// call path and one set of bounds to test. The wrapper is inert behind the
// workflow.jev.enabled gate: while the gate is false — the shipped template
// default — it constructs no request, makes no network call, and reports
// gated-unavailable. The capability is display-only: an answer is a labelled
// model signal a person reads, never a completion verdict, a merge approval,
// a queue mutation, or any other decision that is hard to undo.
package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
)

// jevAskToolName is the registered MCP tool name. It MUST match the catalog
// entry in internal/mcp and the catalogue documentation in
// .claude/rules/moai/core/moai-mcp-tools.md (both copies).
const jevAskToolName = "jev_ask"

// newJevAskClient is the construction seam: it assembles the production
// client fully wired (credential loader included), so tests stubbing it
// replace the whole assembly rather than patching fields after the fact. The
// handler calls it only AFTER the gate has passed, so with the gate off the
// seam is never invoked and a test observes a zero-construction count instead
// of trusting an absence of network activity. Replaced in tests.
var newJevAskClient = func() *jev.Client {
	c := jev.New(true)
	c.LoadCredential = jevcred.Load
	return c
}

// registerJevAskTool registers the gated judgment tool. Split out of
// registerMoaiMCPTools to keep the registration table scannable; the closure
// argument is the same `add` helper the table uses, so the per-tool
// enablement map and the catalog-equality guard see this tool identically.
func registerJevAskTool(add func(name string, tool mcp.Tool, handler server.ToolHandlerFunc)) {
	add(jevAskToolName, mcp.NewTool(
		jevAskToolName,
		mcp.WithDescription("Ask the gated TypeSafe System One judgment capability one typed question set over one supplied state; returns typed answers with the model's probability. Wraps jev.Client.Ask. DISPLAY-ONLY: the answer is a labelled model signal a person reads — never a completion predicate, merge approval, queue mutation, or gate input. Gated behind workflow.jev.enabled (shipped default false): with the gate off this tool constructs no request, makes no network call, and reports gated-unavailable; while the chain's fitness measurement gate stands unrun the capability is not presented as available."),
		mcp.WithString("state", mcp.Required(), mcp.Description("The state to judge (plain text). One state per call; several questions over it stay mutually consistent.")),
		mcp.WithArray("questions", mcp.Required(), mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"question_id": map[string]any{"type": "string"},
				"question":    map[string]any{"type": "string"},
				"kind":        map[string]any{"type": "string", "enum": []string{"noul", "choice", "score"}},
				"choices":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
			"required": []string{"question_id", "question"},
		}), mcp.Description("Typed questions to ask over the state (kind: noul = yes/no, choice = one of choices, score = bounded numeric).")),
		mcp.WithBoolean("repeatable", mcp.Description("Declare this exact call provably side-effect-free to permit pre-delivery retry. Defaults false — an undeclared request is never retried.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleJevAsk)
}

// @MX:WARN: [AUTO] gate-read path — handleJevAsk reads workflow.jev.enabled on every call and must keep it first and short-circuiting
// @MX:REASON: [AUTO] a reordered or degraded gate read would make an opt-in capability reachable from any agent session while a project that never opted in could still be charged for a call; the gate-off branch constructs no client, so the zero-request property depends on this ordering.
func handleJevAsk(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// The gate is read FIRST and short-circuits: while the capability is
	// disabled no client is constructed, so no request can exist (the doctor
	// check shares this ordering under the same constraint family).
	enabled, readErr := jevEnabled(resolveProjectDir())
	if readErr != nil || !enabled {
		reason := "workflow.jev.enabled is false (shipped default); no request constructed, no network call"
		if readErr != nil {
			reason = "workflow.jev gate unreadable; treated as disabled; no request constructed, no network call"
		}
		return toolJSON(jevAskToolName, map[string]any{
			"status":    "gated_unavailable",
			"available": false,
			"reason":    reason,
			"opt_in":    "set workflow.jev.enabled: true in .moai/config/sections/workflow.yaml; a TypeSafe credential at ~/.moai/.env.typesafe is then also required",
		}), nil
	}

	state, err := req.RequireString("state")
	if err != nil {
		return toolErr(jevAskToolName, errors.New("state is required: the plain-text state to judge")), nil
	}
	questions, err := decodeJevAskQuestions(req)
	if err != nil {
		return toolErr(jevAskToolName, err), nil
	}

	client := newJevAskClient()
	res := client.Ask(ctx, jev.Request{
		State:      state,
		Questions:  questions,
		Repeatable: req.GetBool("repeatable", false),
	})

	out := map[string]any{
		"available":    res.OK(),
		"availability": string(res.Availability),
	}
	if res.Condition != "" {
		out["condition"] = res.Condition
	}
	if notice := res.NoticeLine(); notice != "" {
		out["notice"] = notice
	}
	if res.OK() {
		out["answers"] = res.Answers
		out["usage"] = res.Usage
	}
	return toolJSON(jevAskToolName, out), nil
}

// decodeJevAskQuestions maps the tool's questions argument onto the jev
// request's typed questions. Structural problems (not an array, an item that
// is not an object) are tool errors; semantic problems (an empty list, an
// unknown kind, oversize input, credential-shaped payloads) stay with
// internal/jev, which returns typed unavailability for them — one validation
// layer per concern.
func decodeJevAskQuestions(req mcp.CallToolRequest) ([]jev.Question, error) {
	raw, ok := req.GetArguments()["questions"].([]any)
	if !ok {
		return nil, errors.New("questions is required: a non-empty array of {question_id, question, kind?, choices?}")
	}
	questions := make([]jev.Question, 0, len(raw))
	for i, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("questions[%d] must be an object", i)
		}
		id, _ := m["question_id"].(string)
		text, _ := m["question"].(string)
		q := jev.Question{ID: id, Text: text}
		if kind, ok := m["kind"].(string); ok && kind != "" {
			q.Kind = jev.AnswerKind(kind)
		}
		if choices, ok := m["choices"].([]any); ok {
			for _, c := range choices {
				if s, ok := c.(string); ok {
					q.Choices = append(q.Choices, s)
				}
			}
		}
		questions = append(questions, q)
	}
	return questions, nil
}
