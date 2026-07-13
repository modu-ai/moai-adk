// Package v4manifest — schema constants (design §C.2 + §E).
//
// The 5 execution primitives, 2 isolation modes, 5 effort levels, 4 model
// tiers, and 6 pattern-catalog entries are reproduced verbatim from design.md
// §C.2 (schema validation rules) and §E (6-pattern catalog). These sets are
// closed: validation rejects any value not in the set.
package v4manifest

// Execution primitives (design §C.2). The Runner dispatches verbatim per
// specialist.primitive — no heuristic re-derivation (REQ-HV4-005 / AC-HV4-005b).
const (
	PrimitiveSubAgent          = "sub-agent"
	PrimitiveDynamicWorkflow   = "dynamic-workflow"
	PrimitiveWorktree          = "worktree"
	PrimitiveGoal              = "/goal"
	PrimitiveAdversarialFanOut = "adversarial-fan-out"
)

// validPrimitives is the closed set of the 5 execution primitives. A
// specialist.primitive MUST be exactly one of these (design §C.2).
var validPrimitives = map[string]bool{
	PrimitiveSubAgent:          true,
	PrimitiveDynamicWorkflow:   true,
	PrimitiveWorktree:          true,
	PrimitiveGoal:              true,
	PrimitiveAdversarialFanOut: true,
}

// Isolation modes (design §C.2). Per-specialist, conditional (REQ-HV4-007).
const (
	IsolationNone    = "none"
	IsolationWorktree = "worktree"
)

// validIsolations is the closed set of the 2 isolation modes.
var validIsolations = map[string]bool{
	IsolationNone:     true,
	IsolationWorktree: true,
}

// Effort levels (design §C.1). Per SPEC-V3R6-WORKFLOW-EFFORT-MAP-001
// purpose-driven taxonomy. low/medium/high/xhigh/max.
const (
	EffortLow    = "low"
	EffortMedium = "medium"
	EffortHigh   = "high"
	EffortXhigh  = "xhigh"
	EffortMax    = "max"
)

// validEfforts is the closed set of the 5 effort levels.
var validEfforts = map[string]bool{
	EffortLow:    true,
	EffortMedium: true,
	EffortHigh:   true,
	EffortXhigh:  true,
	EffortMax:    true,
}

// Model tiers (design §C.1). inherit/haiku/sonnet/opus.
const (
	ModelInherit = "inherit"
	ModelHaiku   = "haiku"
	ModelSonnet  = "sonnet"
	ModelOpus    = "opus"
)

// validModels is the closed set of the 4 model tiers.
var validModels = map[string]bool{
	ModelInherit: true,
	ModelHaiku:   true,
	ModelSonnet:  true,
	ModelOpus:    true,
}

// Schedule mechanisms. "loop" is the native /loop scheduler (session-scoped);
// "cron" is the Cron tools registration (persistent across sessions).
const (
	MechanismLoop = "loop"
	MechanismCron = "cron"
)

// validMechanisms is the closed set of the 2 schedule mechanisms.
var validMechanisms = map[string]bool{
	MechanismLoop: true,
	MechanismCron: true,
}

// ScheduleModeDiscoveryOnly is the sole valid schedule mode literal.
// Scheduled harness runs are discovery-only: read-only analysis, findings
// persisted to a queue surface, no writes/commits/pushes, no run-phase entry.
const ScheduleModeDiscoveryOnly = "discovery-only"

// 6-pattern catalog (design §E). Patterns are selected/combined dynamically
// by the PLAN phase; the selection is recorded in manifest.patterns.
// AC-HV4-004b requires patterns[] entries to be from this catalog (no custom
// patterns).
const (
	PatternPipeline                 = "Pipeline"
	PatternFanOutFanIn              = "Fan-out/Fan-in"
	PatternExpertPool               = "Expert Pool"
	PatternProducerReviewer         = "Producer-Reviewer"
	PatternSupervisor               = "Supervisor"
	PatternHierarchicalDelegation   = "Hierarchical Delegation"
)

// validPatterns is the closed set of the 6-pattern catalog.
var validPatterns = map[string]bool{
	PatternPipeline:               true,
	PatternFanOutFanIn:            true,
	PatternExpertPool:             true,
	PatternProducerReviewer:       true,
	PatternSupervisor:             true,
	PatternHierarchicalDelegation: true,
}

// ─── Sub-agent 4-color tier model (SPEC-WEBCONF-SIMPLIFY-001 M1) ──────────────
//
// The tier is a display-time reasoning-role classification of a sub-agent. It
// is the source of truth for the moai-web agentfm color badge. It is
// INDEPENDENT of each agent's live effort frontmatter value (design.md §B):
//
//   - Tier  answers "what reasoning role does this agent play?"  (stable, chosen)
//   - Effort answers "what effort did the user last set?"         (mutable, spawn-time)
//
// The REJECTED alternative — deriving the tier from the agent's effort
// frontmatter — was rejected because it requires rewriting 20 agent files,
// creates a circular dependency for the 3 absent-effort agents, and makes the
// badge flap when the user changes effort (design.md §G). The name-keyed
// table below is a chosen mapping, NOT derived from any frontmatter.

// Tier is the 4-color reasoning-role classification of a sub-agent. It is a
// distinct named type so a raw effort string cannot be passed where a Tier is
// expected, preserving the tier-vs-effort distinction at the type level.
type Tier string

// The 4 tier values (closed set). The string identity is color-based (NOT
// effort-based) to avoid even surface-level conflation with the effort
// closed set {low, medium, high, xhigh, max}.
const (
	TierRed       Tier = "red"       // 🔴 — deep reasoning
	TierOrange    Tier = "orange"    // 🟠 — heavy reasoning
	TierBlue      Tier = "blue"      // 🔵 — moderate reasoning
	TierLightBlue Tier = "lightblue" // 🩵 — light / narrow-scope
)

// validTiers is the closed set of the 4 tiers.
var validTiers = map[Tier]bool{
	TierRed:       true,
	TierOrange:    true,
	TierBlue:      true,
	TierLightBlue: true,
}

// tierColors maps each tier to its render glyph (design.md §C/D). The UI
// layer (moai-web agentfm, M5) renders the glyph as the color badge. An
// unmapped tier yields "" so the caller can surface a neutral "custom" /
// "unmapped" badge (the max/inherit override sentinels are handled at the UI
// layer, not in this table — C-8).
var tierColors = map[Tier]string{
	TierRed:       "🔴",
	TierOrange:    "🟠",
	TierBlue:      "🔵",
	TierLightBlue: "🩵",
}

// tierSuggestions is the tier → suggested-(model, effort) table (design.md §D).
// The suggestion is applied ONLY on explicit user action (click a tier color);
// applying it writes effort/model via the existing agentfm.Patch mechanism —
// it does NOT write a new tier: frontmatter key, and it does NOT change the
// name→tier table (C-7).
var tierSuggestions = map[Tier]struct {
	Model  string
	Effort string
}{
	TierRed:       {Model: ModelOpus, Effort: EffortXhigh},
	TierOrange:    {Model: ModelOpus, Effort: EffortHigh},
	TierBlue:      {Model: ModelSonnet, Effort: EffortMedium},
	TierLightBlue: {Model: ModelHaiku, Effort: EffortLow},
}

// agentTiers is the name-keyed lookup table: agent file stem (the base name
// of the .md under .claude/agents/{moai,harness}/, matching agentfm.
// AgentInfo.Name's contract) → Tier. This is the SSOT for badge color
// (plan.md §F M1.2 + design.md §C). Distribution: 🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 = 20.
//
// @MX:ANCHOR: [AUTO] sub-agent tier SSOT — name-keyed lookup table (Option A)
// @MX:REASON: display-only tier invariant; 3+ consumers (AgentTier accessor, moai-web agentfm render, tests). Mutating this map changes badge colors project-wide.
// @MX:SPEC: SPEC-WEBCONF-SIMPLIFY-001 M1.2 / design.md §C
var agentTiers = map[string]Tier{
	// 🔴 — deep reasoning (×4): plan-phase authoring, independent audit,
	// high-reasoning consultation, skeptical 4-dimension scoring.
	"manager-spec":  TierRed,
	"plan-auditor":  TierRed,
	"super-advisor": TierRed,
	"sync-auditor":  TierRed,

	// 🟠 — heavy reasoning bounded by the SPEC (×4): run-phase DDD/TDD,
	// design pipeline, harness/specialist generation, cross-platform E2E.
	"manager-develop": TierOrange,
	"manager-design":  TierOrange,
	"builder-harness": TierOrange,
	"e2e-specialist":  TierOrange,

	// 🔵 — moderate / procedural reasoning (×5): sync-phase docs, git ops,
	// TRUST 5 checks, template editing, workflow pattern application.
	"manager-docs":            TierBlue,
	"manager-git":             TierBlue,
	"quality-specialist":      TierBlue,
	"cli-template-specialist": TierBlue,
	"workflow-specialist":     TierBlue,

	// 🩵 — light / narrow-scope (×7): the hns-* harness specialists and
	// hook-ci. The 3 hns-oss-docs-* agents here have NO effort frontmatter
	// at all — under Option A they still get 🩵 from this table directly,
	// eliminating the circular-dependency that defeated effort-derivation.
	"hns-github-specialist":                     TierLightBlue,
	"hns-oss-docs-content-author-specialist":    TierLightBlue,
	"hns-oss-docs-locale-translator-specialist": TierLightBlue,
	"hns-oss-docs-structure-curator-specialist": TierLightBlue,
	"hns-release-specialist":                    TierLightBlue,
	"hns-release-update-specialist":             TierLightBlue,
	"hook-ci-specialist":                        TierLightBlue,
}

// AgentTier returns the Tier for the named agent, keyed by the agent's file
// stem (the base name of its .md under .claude/agents/{moai,harness}/).
// The ok return is false when the name has no table entry (e.g. a future
// agent not yet added — design.md EC-6 renders no badge until the table is
// extended). The lookup is display-only and never touches the agent's
// frontmatter.
//
// @MX:ANCHOR: [AUTO] primary tier lookup accessor (name → Tier)
// @MX:REASON: fan_in >= 3 (settings re-export, moai-web agentfm, tests); the badge-color entry point for every agent row.
func AgentTier(name string) (t Tier, ok bool) {
	t, ok = agentTiers[name]
	return
}

// TierColor returns the render glyph (emoji) for the tier, or "" for an
// unmapped tier. The UI layer renders a neutral badge on empty.
func TierColor(t Tier) string {
	return tierColors[t]
}

// TierSuggestedModelEffort returns the suggested (model, effort) pair for the
// tier (design.md §D), or ("", "") for an unmapped tier. The suggestion is
// applied only on explicit user action and writes via agentfm.Patch — it does
// NOT add a tier: frontmatter key (C-7).
//
// @MX:ANCHOR: [AUTO] tier → suggested-(model, effort) suggestion contract
// @MX:REASON: fan_in >= 3 (settings re-export, moai-web tier-click auto-fill, tests); pins the 4-row suggestion table that drive the auto-fill UX.
func TierSuggestedModelEffort(t Tier) (model, effort string) {
	s, ok := tierSuggestions[t]
	if !ok {
		return "", ""
	}
	return s.Model, s.Effort
}

// AllAgentTiers returns a snapshot copy of the name→tier table. Use this for
// enumeration (e.g. test assertions on catalog coverage). The returned map is
// a copy so callers cannot mutate the package-level table.
func AllAgentTiers() map[string]Tier {
	out := make(map[string]Tier, len(agentTiers))
	for k, v := range agentTiers {
		out[k] = v
	}
	return out
}

// IsValidTier reports whether t is one of the 4 closed-set tier values.
// Useful for input validation at surfaces that accept a Tier literal.
func IsValidTier(t Tier) bool {
	return validTiers[t]
}
