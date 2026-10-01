package config

import (
	"fmt"
	"sort"
	"strings"
)

// agent_tiers.go — the tier-axis assignment surface (SPEC-AGENT-TIER-001 M2).
//
// The tier tokens are CONFIGURATION-KEY names, not effort values (REQ-TIER-001,
// the Q2 decision): each names one {model, effort} pair resolved through
// AgentTierPair onto the DefaultClaudeTier* constants. The assignment is
// machine-driven configuration resolution (REQ-TIER-010): agent definition
// files carry no model/effort/tier frontmatter and no consumer passes values
// by hand at spawn time.
//
// Audit surfaces are excluded from the tier matrix (REQ-TIER-007): plan-auditor,
// sync-auditor, and the claude/codex/glm audit backends resolve to NO tier —
// they ride the workflow.audit pins exclusively.

// AgentTier tokens — the closed tier set (REQ-TIER-007).
const (
	// AgentTierMax is the accuracy-first tier {sonnet-5-5, max} (70.6% @ ~$11).
	AgentTierMax = "max"
	// AgentTierMedium is the cost-efficiency tier {sonnet-5-5, high} (45% @ ~$2.3).
	AgentTierMedium = "medium"
	// AgentTierLow is the light-work tier {sonnet-5-5, medium} (29% @ ~$0.8).
	AgentTierLow = "low"
)

// ValidAgentTiers returns the closed set for workflow.agent_tiers assignments,
// derived from the AgentTier* tokens.
func ValidAgentTiers() []string {
	return []string{AgentTierMax, AgentTierMedium, AgentTierLow}
}

// validAgentTiersSet is the membership set behind Validate.
var validAgentTiersSet = map[string]struct{}{
	AgentTierMax:    {},
	AgentTierMedium: {},
	AgentTierLow:    {},
}

// DefaultAgentTierClasses returns the default agent-class → tier assignment
// (REQ-TIER-008): super-advisor and manager-spec (design judgments) → max;
// manager-develop, manager-docs, e2e-tester, and the general implementation
// lane ("lane") → medium; Explore and light-fix/observation/summary tasks →
// low. A project overrides per class through workflow.agent_tiers.classes;
// unlisted classes keep these built-ins.
func DefaultAgentTierClasses() map[string]string {
	return map[string]string{
		"super-advisor":   AgentTierMax,
		"manager-spec":    AgentTierMax,
		"manager-develop": AgentTierMedium,
		"manager-docs":    AgentTierMedium,
		"e2e-tester":      AgentTierMedium,
		"explore":         AgentTierLow,
		"lane":            AgentTierMedium,
	}
}

// auditSurfaceClasses are the agent classes EXCLUDED from the tier matrix
// (REQ-TIER-007): the two independent auditors and the three audit backends.
// ResolveAgentClassTier returns "" for them whatever the table says — the
// audit surfaces resolve exclusively through the workflow.audit pins.
var auditSurfaceClasses = map[string]struct{}{
	"plan-auditor": {},
	"sync-auditor": {},
	"audit-claude": {},
	"audit-codex":  {},
	"audit-glm":    {},
}

// IsAuditSurfaceClass reports whether the class is excluded from the tier
// matrix (REQ-TIER-007) and resolves through the workflow.audit pins instead.
func IsAuditSurfaceClass(class string) bool {
	_, ok := auditSurfaceClasses[strings.TrimSpace(class)]
	return ok
}

// ResolveAgentClassTier returns the tier token for an agent class:
// the user's workflow.agent_tiers.classes entry wins per class, an unlisted
// class (or a stored empty value — the console's "unset" submit) falls back
// to DefaultAgentTierClasses, and an audit surface always resolves to ""
// (no tier — REQ-TIER-007) whatever the table carries.
func ResolveAgentClassTier(tiers AgentTiersConfig, class string) string {
	name := strings.TrimSpace(class)
	if IsAuditSurfaceClass(name) {
		return ""
	}
	if tiers.Classes != nil {
		if tier, ok := tiers.Classes[name]; ok && tier != "" {
			return tier
		}
	}
	return DefaultAgentTierClasses()[name]
}

// AgentTierPair resolves a tier token to its {model, effort} pair
// (REQ-TIER-002): max → DefaultClaudeTierMax, medium → DefaultClaudeTierMedium,
// low → DefaultClaudeTierLow. An unknown token resolves nothing (ok=false) —
// callers treat that as a configuration defect, never as a default.
func AgentTierPair(tier string) (ModelEffort, bool) {
	switch tier {
	case AgentTierMax:
		return DefaultClaudeTierMax, true
	case AgentTierMedium:
		return DefaultClaudeTierMedium, true
	case AgentTierLow:
		return DefaultClaudeTierLow, true
	default:
		return ModelEffort{}, false
	}
}

// Validate enforces the closed tier set on every configured class assignment
// (REQ-TIER-009): a token outside {max, medium, low} is an error naming the
// offending class, the token, the valid set, and the file the loader wraps
// around it. An empty/nil Classes table is valid (all classes then resolve to
// the built-in defaults), and a stored EMPTY token is the console's "unset"
// submit — it falls back to the default rather than failing the load.
func (c AgentTiersConfig) Validate() error {
	// Sort for a deterministic error when several tokens are invalid.
	classes := make([]string, 0, len(c.Classes))
	for class := range c.Classes {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	for _, class := range classes {
		tier := c.Classes[class]
		if tier == "" {
			continue // unset — ResolveAgentClassTier falls back to the default
		}
		if _, ok := validAgentTiersSet[tier]; !ok {
			return fmt.Errorf("workflow.agent_tiers.classes[%s] = %q invalid: want one of max|medium|low", class, tier)
		}
	}
	return nil
}

// AgentTierClassOrder returns the known agent classes in domain order (the
// order display surfaces list them): design judgments first, then the
// implementation classes, then the light-work class. Unknown classes a project
// configures are appended after these by the caller if it wants them shown.
func AgentTierClassOrder() []string {
	return []string{
		"super-advisor",
		"manager-spec",
		"manager-develop",
		"manager-docs",
		"e2e-tester",
		"explore",
		"lane",
	}
}

// AgentTierChart is one tier's chart grounding for display surfaces
// (REQ-TIER-011): the tier token plus its Terminal-Bench 4.0 accuracy and
// cost-per-attempt figures, rendered verbatim.
type AgentTierChart struct {
	// Tier is the tier token from the closed set (ValidAgentTiers).
	Tier string
	// Score is the chart accuracy figure, e.g. "70.6%".
	Score string
	// Cost is the chart cost-per-attempt figure, e.g. "~$11".
	Cost string
	// Pair is the tier's {model, effort} resolution, e.g. "sonnet-5-5 @ max".
	Pair string
}

// AgentTierChartTable returns the three tiers with their chart grounding, in
// tier order (max, medium, low). The figures are the SPEC-critical evidence
// base (Terminal-Bench 4.0) and live HERE only — display surfaces consume
// this table and never restate the numbers.
func AgentTierChartTable() []AgentTierChart {
	return []AgentTierChart{
		{Tier: AgentTierMax, Score: "70.6%", Cost: "~$11", Pair: DefaultClaudeTierMaxModel + " @ " + DefaultClaudeTierMaxEffort},
		{Tier: AgentTierMedium, Score: "45%", Cost: "~$2.3", Pair: DefaultClaudeTierMediumModel + " @ " + DefaultClaudeTierMediumEffort},
		{Tier: AgentTierLow, Score: "29%", Cost: "~$0.8", Pair: DefaultClaudeTierLowModel + " @ " + DefaultClaudeTierLowEffort},
	}
}
