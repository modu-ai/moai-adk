// mcp_refusal.go — the structured policy-refusal contract
// (SPEC-MCP-NEUTRALITY-001 REQ-MN-001/-002/-003; design §D1/§D4). A policy
// refusal is a SUCCESSFUL MCP result whose payload describes the refusal:
// {refused: true, code, reason, hint}. The MCP error channel is never used
// for the five policy classes (serial-slot, owned, raced, parent-checkout,
// stale-run), so a caller can tell "the queue refused you" from "the server
// broke", and the failure observer's refused:true classification excludes
// refusals from the lessons inbox.
//
// The identity class (unknown attribution) refuses with refused:true and NO
// code: the closed set (plan §D3) names no member for it, and broadening the
// set is a SPEC amendment, not an ad-hoc string.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The two refusal codes the factory card layer did not already name. The
// other three members of the closed set — serial-slot, owned, raced (plan
// §D3) — are the nominate-refusal tokens factory_card.go already emits
// (factoryRefuseSerialSlot, factoryRefuseOwned, factoryRefuseRaced); those
// constants are reused verbatim so the CLI and MCP surfaces cannot drift.
const (
	// mcpRefusalParentCheckout refuses a lane verb that must run from the
	// parent checkout it names (REQ-MN-004 documents the requirement on all
	// three verbs; this is its enforcement arm).
	mcpRefusalParentCheckout = "parent-checkout"

	// mcpRefusalStaleRun refuses a run id that is not the currently active
	// run (REQ-MN-003): no lease, stage, or evidence state is read from
	// expired rows.
	mcpRefusalStaleRun = "stale-run"
)

// factoryUnknownAttributionSentinel names the unmarked-session refusal of the
// operator-only verbs (REQ-MN-003): no positive operator mark, no lane
// identity — nothing attributes the caller.
const factoryUnknownAttributionSentinel = "unknown attribution"

// mcpRefusal is the wire payload of a structured policy refusal (design §D1).
// Code stays empty for the identity class.
type mcpRefusal struct {
	Refused bool   `json:"refused"`
	Code    string `json:"code,omitempty"`
	Reason  string `json:"reason"`
	Hint    string `json:"hint,omitempty"`
}

// toolRefusal shapes a policy refusal as a successful result (is_error false)
// whose structured content AND text content block carry the same refusal JSON
// (design §D1 — one serializer, two blocks). The observer's refused:true
// classification keys on this payload.
//
// @MX:ANCHOR: [AUTO] the one refusal serializer every factory/todo/msg MCP handler shares
// @MX:REASON: 18 call sites plus the failure observer's refused:true classification key on this single function — a divergent serializer would split the wire shape
// @MX:SPEC: SPEC-MCP-NEUTRALITY-001
func toolRefusal(payload mcpRefusal) *mcp.CallToolResult {
	data, err := json.Marshal(payload)
	if err != nil {
		// Unreachable for this struct; the wire never panics on it.
		data = []byte(`{"refused":true,"reason":"refusal serialization failed"}`)
	}
	return &mcp.CallToolResult{
		Content:           []mcp.Content{mcp.TextContent{Type: "text", Text: string(data)}},
		StructuredContent: payload,
	}
}

// mcpIdentityRefusal is the unknown-attribution refusal (REQ-MN-003) with no
// closed-set code. The reason carries the same refusal wording the CLI
// prints, so the two surfaces keep one text.
func mcpIdentityRefusal(reason string) mcpRefusal {
	return mcpRefusal{Refused: true, Reason: reason}
}

// mcpUnknownAttributionReason renders the unmarked-session refusal reason for
// one tool surface — the ONE wording the CLI verb shares
// (factoryUnknownAttributionError), so the two surfaces cannot drift.
func mcpUnknownAttributionReason(tool string) string {
	return factoryUnknownAttributionError(tool).Error()
}

// mcpNominateRefusalPayload maps a nominated-lease refusal onto the closed
// code set: the three tokens that are members map to their codes, every other
// token reports false and keeps today's error-channel shape (the caller falls
// back to toolErr).
func mcpNominateRefusalPayload(r *factoryNominateRefusal) (mcpRefusal, bool) {
	switch r.Token {
	case factoryRefuseSerialSlot, factoryRefuseOwned, factoryRefuseRaced:
		return mcpRefusal{Refused: true, Code: r.Token, Reason: r.Detail}, true
	}
	return mcpRefusal{}, false
}

// mcpStaleRunCheck validates an EXPLICIT run id (REQ-MN-003): a run that is
// not the currently active one yields the stale-run refusal result, and the
// caller must stop — no lease, stage, or evidence state is read from expired
// rows. A nil result means the call may proceed: no explicit id was given (a
// path the caller resolves separately) or the id is the active run. A
// malformed id is an argument error and keeps the error channel.
func mcpStaleRunCheck(ctx context.Context, root, run, tool string) *mcp.CallToolResult {
	if strings.TrimSpace(run) == "" {
		return nil
	}
	if err := factorymsg.ValidateActiveRun(ctx, root, run); err != nil {
		if errors.Is(err, factorymsg.ErrRunNotActive) {
			return toolRefusal(mcpRefusal{
				Refused: true,
				Code:    mcpRefusalStaleRun,
				Reason:  fmt.Sprintf("run %q is not the currently active factory run", run),
				Hint:    "pass the active run id, or omit run to act in the single active run",
			})
		}
		return toolErr(tool, fmt.Errorf("%s: run %q: %w", tool, run, err))
	}
	return nil
}

// mcpFactoryActor resolves the caller's actor label (REQ-MN-003): the
// explicit actor argument wins; the spawn environment's factory-worker key
// (internal/config/envkeys.go EnvMoaiFactoryWorker) is the default only.
// Empty when neither carries a label — the caller refuses that as unknown
// attribution.
func mcpFactoryActor(req mcp.CallToolRequest) string {
	if actor := strings.TrimSpace(req.GetString("actor", "")); actor != "" {
		return actor
	}
	return strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker))
}

// mcpOperatorAuthority reports a POSITIVE operator mark (REQ-MN-003): the
// role variable names the operator role. The absence of lane variables grants
// nothing — an unmarked session is unknown attribution, and the operator-only
// verbs refuse it (the t1624 defect: an unmarked Codex session read as the
// operator because identity was decided by env absence).
func mcpOperatorAuthority() bool {
	return os.Getenv(config.EnvFactoryRole) == config.FactoryRoleOperator
}

// mcpLeaseHolderRefusal maps the F1 holder refusal (homestate.ErrLeaseHolder)
// onto the owned policy code with the CLI-identical text: the card is held by
// another endpoint. ok is false for every other error, which keeps the error
// channel.
func mcpLeaseHolderRefusal(err error) (*mcp.CallToolResult, bool) {
	if !errors.Is(err, homestate.ErrLeaseHolder) {
		return nil, false
	}
	return toolRefusal(mcpRefusal{Refused: true, Code: factoryRefuseOwned, Reason: err.Error()}), true
}

// mcpParentCheckoutRefusal wraps the parent-checkout assertion failure as its
// policy code; the reason keeps the assertion's own text naming the parent.
func mcpParentCheckoutRefusal(err error) *mcp.CallToolResult {
	return toolRefusal(mcpRefusal{
		Refused: true,
		Code:    mcpRefusalParentCheckout,
		Reason:  err.Error(),
		Hint:    "run the factory verbs from the parent checkout the project_root argument names",
	})
}

// mcpTodoOperatorGate is the REQ-MN-003 operator gate the todo MCP handlers
// share: only a positive operator mark proceeds. A lane-marked session keeps
// the CLI-identical boundary refusal; an unmarked session is unknown
// attribution. Both refuse as structured results (REQ-MN-001) and the queue
// bytes never move. A nil result admits the call.
func mcpTodoOperatorGate(tool, surface string) *mcp.CallToolResult {
	if mcpOperatorAuthority() {
		return nil
	}
	if factoryLaneRefusal() {
		return toolRefusal(mcpIdentityRefusal(todoLaneMutationRefusalText(surface)))
	}
	return toolRefusal(mcpIdentityRefusal(mcpUnknownAttributionReason(tool)))
}
