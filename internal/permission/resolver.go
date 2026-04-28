package permission

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// ResolveContext provides contextual information for permission resolution.
type ResolveContext struct {
	// Mode is the agent's permission mode.
	Mode PermissionMode

	// IsFork indicates if this agent is a fork (spawned by a parent session).
	// When true and Mode is bubble, prompts are routed to the parent session.
	IsFork bool

	// ParentAvailable indicates if the parent session is reachable.
	// When false in bubble mode, permission requests are denied.
	ParentAvailable bool

	// ForkDepth is the nesting depth of the fork (0 = not a fork, 1 = direct child, etc.).
	// When depth > 3, all modes except plan are degraded to bubble with a warning.
	ForkDepth int

	// StrictMode indicates if bypassPermissions mode should be rejected.
	// When true, spawning an agent with ModeBypassPermissions fails.
	StrictMode bool

	// IsInteractive indicates if the session is interactive (can prompt the user).
	// When false, "ask" decisions are degraded to "deny".
	IsInteractive bool

	// HookResponse is the response from the PreToolUse hook, if any.
	// Hook decisions override all tier-based rules for a single tool call.
	HookResponse *hook.HookResponse

	// RulesByTier is the set of permission rules indexed by source tier.
	// The resolver walks these tiers in priority order (SrcPolicy to SrcBuiltin).
	RulesByTier map[config.Source][]PermissionRule
}

// ResolveResult is the outcome of permission resolution.
type ResolveResult struct {
	// Decision is the final permission decision.
	Decision Decision

	// ResolvedBy is the source tier that supplied the decision.
	ResolvedBy config.Source

	// Origin is the file path that contributed the winning rule.
	// For SrcBuiltin, this may be "pre-allowlist" or "bypassPermissions mode".
	Origin string

	// UpdatedInput contains the modified tool input from hooks.
	// If the hook mutated the input, the resolver re-runs pattern matching.
	UpdatedInput json.RawMessage

	// SystemMessage is a warning message to display to the user.
	// Used for fork depth warnings, mode degradation notices, etc.
	SystemMessage string

	// Trace contains the resolution trace for --trace output.
	// This includes every tier inspected and why it matched or missed.
	Trace ResolutionTrace
}

// ResolutionTrace provides detailed information about the resolution process.
type ResolutionTrace struct {
	// Tool being invoked.
	Tool string

	// Input arguments for the tool.
	Input string

	// Tries is the list of tiers inspected in order.
	Tries []TierTry
}

// TierTry records the outcome of checking a single tier.
type TierTry struct {
	Tier    config.Source
	Matched bool
	Rule    *PermissionRule
	Reason  string
}

// PermissionResolver resolves tool permissions through an 8-tier stack.
//
// The resolver walks tiers in priority order:
// 1. Hook decision (overrides all tiers if present)
// 2. SrcPolicy (system-wide policy)
// 3. SrcUser (user-specific config)
// 4. SrcProject (project-specific config)
// 5. SrcLocal (local overrides)
// 6. SrcPlugin (plugin-contributed, reserved for v3.1+)
// 7. SrcSkill (skill frontmatter config)
// 8. SrcSession (session-scoped rules)
// 9. SrcBuiltin (pre-allowlist and defaults)
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-004
type PermissionResolver struct {
	mu sync.RWMutex

	// Pre-allowlist rules (SrcBuiltin tier)
	preAllowlist []PermissionRule
}

// NewPermissionResolver creates a new permission resolver with the pre-allowlist loaded.
func NewPermissionResolver() *PermissionResolver {
	return &PermissionResolver{
		preAllowlist: PreAllowlist(),
	}
}

// Resolve determines the permission decision for a tool invocation.
// It walks the 8-tier stack in priority order and returns the first non-empty decision.
//
// The resolution process:
// 1. Check bypassPermissions mode (short-circuit if allowed)
// 2. Check plan mode (deny writes)
// 3. Apply hook decision if present
// 4. Walk tiers from highest to lowest priority
// 5. Return the first matching rule, or the tier's default action
// 6. Handle non-interactive mode (convert "ask" to "deny")
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-004, REQ-V3R2-RT-002-005
func (r *PermissionResolver) Resolve(tool string, input json.RawMessage, ctx ResolveContext) (*ResolveResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inputStr := string(input)
	trace := ResolutionTrace{
		Tool:  tool,
		Input: inputStr,
	}

	// Handle UpdatedInput from hook - re-run resolution with mutated input
	// Only re-run if hook has UpdatedInput but NO PermissionDecision
	if ctx.HookResponse != nil && len(ctx.HookResponse.UpdatedInput) > 0 && ctx.HookResponse.PermissionDecision == "" {
		newCtx := ctx
		newCtx.HookResponse = nil // Clear hook to avoid infinite loop
		newInput := ctx.HookResponse.UpdatedInput

		result, err := r.Resolve(tool, newInput, newCtx)
		if err != nil {
			return nil, err
		}

		result.UpdatedInput = newInput
		result.Trace = trace // Preserve original trace context
		return result, nil
	}

	// Step 1: Handle bypassPermissions mode (short-circuit)
	if ctx.Mode == ModeBypassPermissions && !ctx.StrictMode {
		if ctx.IsFork {
			// Fork agents with bypassPermissions are degraded to bubble
			result := r.handleBypassInFork(tool, inputStr, ctx, &trace)
			return result, nil
		}
		return &ResolveResult{
			Decision:   DecisionAllow,
			ResolvedBy: config.SrcBuiltin,
			Origin:     "bypassPermissions mode",
			Trace:      trace,
		}, nil
	}

	// Step 2: Handle plan mode (deny all writes)
	if ctx.Mode == ModePlan && IsWriteOperation(tool, inputStr) {
		trace.Tries = append(trace.Tries, TierTry{
			Tier:    config.SrcBuiltin,
			Matched: true,
			Reason:  "plan mode denies writes",
		})
		return &ResolveResult{
			Decision:      DecisionDeny,
			ResolvedBy:    config.SrcBuiltin,
			Origin:        "plan mode denies writes",
			SystemMessage: "plan mode denies writes",
			Trace:         trace,
		}, nil
	}

	// Step 3: Apply hook decision (overrides all tiers)
	if ctx.HookResponse != nil && ctx.HookResponse.PermissionDecision != "" {
		hookDecision := Decision(ctx.HookResponse.PermissionDecision)
		trace.Tries = append(trace.Tries, TierTry{
			Tier:    config.Source(999), // Hook tier is above SrcPolicy
			Matched: true,
			Reason:  fmt.Sprintf("hook decision: %s", hookDecision),
		})
		return &ResolveResult{
			Decision:   hookDecision,
			ResolvedBy: config.SrcBuiltin, // Hooks are not a tier, use builtin for provenance
			Origin:     "PreToolUse hook",
			Trace:      trace,
		}, nil
	}

	// Step 3.5: Handle fork depth exceeding limit (degrade to bubble)
	// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-023
	if ctx.ForkDepth > 3 && ctx.Mode != ModePlan && ctx.Mode != ModeBubble {
		return &ResolveResult{
			Decision:      DecisionAsk,
			ResolvedBy:    config.SrcBuiltin,
			Origin:        "fork depth limit",
			SystemMessage: fmt.Sprintf("Fork depth %d exceeds limit - mode degraded to bubble", ctx.ForkDepth),
			Trace:         trace,
		}, nil
	}

	// Step 4: Walk tiers in priority order
	tiers := []config.Source{
		config.SrcPolicy,
		config.SrcUser,
		config.SrcProject,
		config.SrcLocal,
		config.SrcPlugin,
		config.SrcSkill,
		config.SrcSession,
		config.SrcBuiltin,
	}

	for _, tier := range tiers {
		result := r.checkTier(tool, inputStr, tier, ctx, &trace)
		if result != nil {
			return result, nil
		}
		// Continue to next tier if no match
	}

	// Step 5: No rule matched - return default ask (or deny in non-interactive)
	defaultDecision := DecisionAsk
	if !ctx.IsInteractive {
		defaultDecision = DecisionDeny
		r.logUnreachablePrompt(tool, inputStr)
	}

	return &ResolveResult{
		Decision:   defaultDecision,
		ResolvedBy: config.SrcBuiltin,
		Origin:     "no matching rule (default)",
		Trace:      trace,
	}, nil
}

// checkTier checks a single tier for a matching rule.
// Returns a ResolveResult if a decision is reached, otherwise continues to next tier.
func (r *PermissionResolver) checkTier(tool, input string, tier config.Source, ctx ResolveContext, trace *ResolutionTrace) *ResolveResult {
	rules := ctx.RulesByTier[tier]

	// Check pre-allowlist for SrcBuiltin tier
	if tier == config.SrcBuiltin {
		rules = append(rules, r.preAllowlist...)
	}

	// Check each rule in the tier
	for i := range rules {
		rule := &rules[i]
		if rule.Matches(tool, input) {
			trace.Tries = append(trace.Tries, TierTry{
				Tier:    tier,
				Matched: true,
				Rule:    rule,
				Reason:  fmt.Sprintf("rule matched: %s", rule.Pattern),
			})

			// Handle bubble mode for forks
			if ctx.Mode == ModeBubble && ctx.IsFork && rule.Action == DecisionAsk {
				return r.handleBubbleAsk(tool, input, tier, rule, ctx, trace)
			}

			return &ResolveResult{
				Decision:   rule.Action,
				ResolvedBy: tier,
				Origin:     rule.Origin,
				Trace:      *trace,
			}
		}
	}

	// No rule matched in this tier
	trace.Tries = append(trace.Tries, TierTry{
		Tier:    tier,
		Matched: false,
		Reason:  "no matching rule",
	})

	return nil
}

// handleBubbleAsk handles "ask" decisions in bubble mode for fork agents.
// Routes the prompt to the parent session instead of the fork's own channel.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-012
func (r *PermissionResolver) handleBubbleAsk(_ string, _ string, tier config.Source, rule *PermissionRule, ctx ResolveContext, trace *ResolutionTrace) *ResolveResult {
	if !ctx.ParentAvailable {
		return &ResolveResult{
			Decision:      DecisionDeny,
			ResolvedBy:    tier,
			Origin:        rule.Origin,
			SystemMessage: "Bubble target parent unavailable — decision deferred",
			Trace:         *trace,
		}
	}

	// Return "ask" - the orchestrator will route to parent AskUserQuestion
	return &ResolveResult{
		Decision: DecisionAsk,
		ResolvedBy: tier,
		Origin:     rule.Origin,
		SystemMessage: "Bubble mode: routing to parent session",
		Trace:      *trace,
	}
}

// handleBypassInFork handles bypassPermissions mode for fork agents.
// Fork agents with bypassPermissions are degraded to bubble mode.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-023
func (r *PermissionResolver) handleBypassInFork(_ string, _ string, ctx ResolveContext, trace *ResolutionTrace) *ResolveResult {
	var systemMsg string
	if ctx.ForkDepth > 3 {
		systemMsg = fmt.Sprintf("Fork depth %d exceeds limit - bypassPermissions degraded to bubble", ctx.ForkDepth)
	} else {
		systemMsg = "Fork agent with bypassPermissions - degraded to bubble mode"
	}

	return &ResolveResult{
		Decision:      DecisionAsk,
		ResolvedBy:    config.SrcBuiltin,
		Origin:        "bypassPermissions mode (fork)",
		SystemMessage: systemMsg,
		Trace:         *trace,
	}
}

// logUnreachablePrompt logs a permission prompt that couldn't be delivered
// because the session is non-interactive.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-041
func (r *PermissionResolver) logUnreachablePrompt(tool, input string) {
	logPath := filepath.Join(".moai", "logs", "permission.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)

	entry := fmt.Sprintf("[%s] Unreachable prompt: tool=%s input=%s\n",
		formatNow(), tool, truncate(input, 200))

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Silently fail if we can't write the log
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(entry)
}

// formatNow returns the current timestamp in ISO 8601 format.
func formatNow() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// truncate truncates a string to a maximum length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ValidateMode checks if a permission mode is allowed in the current context.
// Returns an error if the mode is rejected (e.g., bypassPermissions in strict mode).
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-022
func (r *PermissionResolver) ValidateMode(mode PermissionMode, isFork bool, strictMode bool, forkDepth int) error {
	if mode == ModeBypassPermissions && strictMode {
		return fmt.Errorf("permission mode rejected: bypassPermissions not allowed in strict mode")
	}

	// Note: isFork + ModeBypassPermissions is allowed — degradation handled in Resolve

	if forkDepth > 3 && mode != ModePlan && mode != ModeBubble {
		// Mode will be degraded to bubble in Resolve
		// Return a warning (not an error)
		return fmt.Errorf("fork depth %d exceeds limit - mode degraded to bubble", forkDepth)
	}

	return nil
}

// ExportTrace exports the resolution trace as JSON for --trace output.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-015
func (r *ResolveResult) ExportTrace() (string, error) {
	data, err := json.MarshalIndent(r.Trace, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal trace: %w", err)
	}
	return string(data), nil
}

// String returns a human-readable summary of the resolution result.
func (r *ResolveResult) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Decision: %s\n", r.Decision)
	fmt.Fprintf(&sb, "Resolved by: %s\n", r.ResolvedBy)
	fmt.Fprintf(&sb, "Origin: %s\n", r.Origin)
	if r.SystemMessage != "" {
		fmt.Fprintf(&sb, "System message: %s\n", r.SystemMessage)
	}
	if len(r.UpdatedInput) > 0 {
		fmt.Fprintf(&sb, "Updated input: %s\n", string(r.UpdatedInput))
	}
	return sb.String()
}
