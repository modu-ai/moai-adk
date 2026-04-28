package permission

import (
	"encoding/json"
	"fmt"
)

// BubbleDispatcher handles the routing of permission prompts to parent sessions
// when fork agents operate in bubble mode.
//
// In bubble mode, when a fork agent encounters a tool that requires user
// confirmation (DecisionAsk), the prompt is routed to the parent session's
// AskUserQuestion channel instead of the fork's own mailbox. This ensures
// that permission decisions are made by the user who initiated the parent
// session, maintaining the security boundary.
//
// Reference: SPEC-V3R2-RT-002 P6 (Permission Bubble), REQ-V3R2-RT-002-012

// BubbleRequest represents a permission prompt that should be routed to the parent session.
type BubbleRequest struct {
	// Tool is the tool being invoked.
	Tool string

	// Input is the tool's input arguments.
	Input json.RawMessage

	// ForkDepth is the nesting depth of the fork making the request.
	ForkDepth int

	// Origin is the file path that contributed the rule triggering the bubble.
	Origin string

	// ParentSessionID is the ID of the parent session to route the prompt to.
	ParentSessionID string
}

// BubbleResponse represents the user's response to a bubbled permission request.
type BubbleResponse struct {
	// Decision is the user's permission decision.
	Decision Decision

	// Reason is an optional explanation for the decision (for audit logging).
	Reason string
}

// BubbleDispatcher manages the routing of bubble requests to parent sessions.
type BubbleDispatcher struct {
	// In a real implementation, this would maintain a registry of active sessions
	// and their communication channels. For now, we provide the structural interface.
}

// NewBubbleDispatcher creates a new bubble dispatcher.
func NewBubbleDispatcher() *BubbleDispatcher {
	return &BubbleDispatcher{}
}

// ShouldBubble returns true if a permission decision should be routed to the parent session.
//
// Bubble routing is triggered when:
// 1. The agent is a fork (IsFork = true)
// 2. The agent is in bubble mode (Mode = ModeBubble)
// 3. The decision is "ask"
// 4. The parent session is available
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-012
func (b *BubbleDispatcher) ShouldBubble(mode PermissionMode, isFork bool, decision Decision, parentAvailable bool) bool {
	return mode == ModeBubble &&
		isFork &&
		decision == DecisionAsk &&
		parentAvailable
}

// CreateBubbleRequest creates a bubble request from a resolution result.
func (b *BubbleDispatcher) CreateBubbleRequest(tool string, input json.RawMessage, result *ResolveResult, parentSessionID string) *BubbleRequest {
	return &BubbleRequest{
		Tool:            tool,
		Input:           input,
		ForkDepth:       0, // Will be populated from context
		Origin:          result.Origin,
		ParentSessionID: parentSessionID,
	}
}

// DispatchToParent sends a bubble request to the parent session.
// In a real implementation, this would use IPC to communicate with the parent process.
// For now, this is a placeholder that demonstrates the contract.
//
// The orchestrator is responsible for:
// 1. Receiving the bubble request
// 2. Invoking AskUserQuestion in the parent session context
// 3. Returning the user's response via HandleBubbleResponse
//
// Reference: agent-common-protocol.md §User Interaction Boundary
func (b *BubbleDispatcher) DispatchToParent(req *BubbleRequest) (*BubbleResponse, error) {
	// Placeholder implementation
	// In production, this would:
	// 1. Serialize the request to the parent session via IPC
	// 2. Wait for the parent session's AskUserQuestion response
	// 3. Deserialize and return the response

	return nil, fmt.Errorf("bubble dispatch not implemented - requires orchestrator integration")
}

// HandleBubbleResponse processes a bubble response from the parent session.
// This converts the parent's decision back into a ResolveResult.
func (b *BubbleDispatcher) HandleBubbleResponse(req *BubbleRequest, resp *BubbleResponse, originalTrace ResolutionTrace) *ResolveResult {
	return &ResolveResult{
		Decision:   resp.Decision,
		ResolvedBy: originalTrace.Tries[len(originalTrace.Tries)-1].Tier, // Use last-checked tier
		Origin:     req.Origin,
		SystemMessage: fmt.Sprintf("Bubbled to parent session - user decided: %s", resp.Decision),
		Trace:      originalTrace,
	}
}

// FormatBubblePrompt formats a bubble request into an AskUserQuestion prompt.
// This is used by the orchestrator when displaying the prompt to the user.
//
// The prompt includes:
// - The tool being invoked
// - The input arguments
// - The fork agent's context (depth, origin)
// - Clear indication that this is a forwarded request from a fork
func (b *BubbleDispatcher) FormatBubblePrompt(req *BubbleRequest) string {
	return fmt.Sprintf(
		"Fork agent is requesting permission to invoke:\n\nTool: %s\nInput: %s\n\nOrigin: %s\nFork depth: %d\n\nAllow this operation?",
		req.Tool,
		truncate(string(req.Input), 500),
		req.Origin,
		req.ForkDepth,
	)
}

// ValidateForkDepth checks if the fork depth exceeds the limit.
// Returns a warning system message if depth > 3.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-023
func (b *BubbleDispatcher) ValidateForkDepth(depth int, mode PermissionMode) (string, error) {
	if depth > 3 && mode != ModePlan && mode != ModeBubble {
		warning := fmt.Sprintf("Fork depth %d exceeds limit - permission mode degraded to bubble", depth)
		return warning, nil
	}
	return "", nil
}

// IsParentAvailable checks if the parent session is reachable.
// In a real implementation, this would check the session registry.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-050
func (b *BubbleDispatcher) IsParentAvailable(parentSessionID string) bool {
	// Placeholder implementation
	// In production, this would check the session registry for liveness
	return parentSessionID != ""
}
