package template

// profile_matrix.go — the per-agent {model, effort} profile matrix (Go-code
// SSOT), the agent→group membership layer, and the runtime resolver that maps
// the active profile + per-agent overrides → each agent's cell. Re-ported
// under SPEC-WEB-AGENTFM-RESTORE-001 M1 (operator card t1411) from the copy
// SPEC-AGENT-MODEL-INHERIT-001 M5 deleted (3fa8bd2ab), per the re-port rule
// (plan §D.4 — not a revert):
//
//   - The cells are RE-DERIVED from the CURRENT model matrix: every cell is
//     one of the config agent-tier pairs (config.DefaultClaudeTier{Max,
//     Medium,Low} — SPEC-AGENT-TIER-001), so the model ids and efforts live
//     in exactly one place and a tier bump flows through without a cell edit.
//     The blueprint's 2026-09-27 opus/sonnet cells are stale and their
//     divergence from these cells is not a defect (plan §B-1(b)). The old
//     cell LEVELS translate onto the tier axis: the old opus/high judgment
//     rows take the max pair, the old opus/medium rows the medium pair, and
//     the old low rows the low pair — preserving the old shape's per-row
//     monotonicity (high >= medium >= low) under the current values.
//   - The roster is the canonical retained catalog (template.RetainedAgents:
//     manager-todo in, mission-governor out at the 2026-10-02 measurement).
//     The display roster filters Explore (no definition file) from the
//     console rows; Explore keeps a mapped cell so pre-restore overrides for
//     it stay resolvable.
//   - The `llm.profiles` config-mirror lookup and the harness-class machinery
//     are NOT re-ported: the config mirror is not re-shipped (plan §D.1 — the
//     Go matrix stays the SSOT) and llm.harness_agents stays retired.
//   - The old profile_matrix.go's llm.yaml regex patchers (ApplyProfile /
//     ApplyHarness) are not re-ported — their only consumers died with them;
//     the restored console persists through the settings seam (plan §D.4-3).

import (
	"maps"
	"slices"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Performance-tier selector wire vocabulary — the closed set the restored
// console's profile selector offers (REQ-AFR-003). The top column keeps the
// historical "max" wire value and a max submission persists verbatim to
// llm.profile (AC-AFR-002); reads fold it back via config.NormalizeProfile.
// DISTINCT from config.ValidProfiles (the canonical {high, medium, low} —
// max readable, never offered) and from config.AgentTier* (the workflow
// class-tier axis, a different machine).
const (
	// PerformanceTierMax is the selector's top-column wire value.
	PerformanceTierMax = "max"
	// PerformanceTierMedium is the balanced selector value.
	PerformanceTierMedium = "medium"
	// PerformanceTierLow is the economical selector value.
	PerformanceTierLow = "low"
)

// ValidPerformanceTiers returns the selector wire vocabulary in display order.
func ValidPerformanceTiers() []string {
	return []string{PerformanceTierMax, PerformanceTierMedium, PerformanceTierLow}
}

// IsValidPerformanceTier reports whether s is a selector wire value. Strict
// membership — no alias folding: the selector persists what it offers, and
// anything else joins the atomic-reject set.
func IsValidPerformanceTier(s string) bool {
	return slices.Contains(ValidPerformanceTiers(), s)
}

// Profile agent-group keys. The groups partition the roster by model+effort
// class; the layer carries no routing information for the per-agent matrix
// (lookup is by agent NAME) — it survives for display grouping and as the
// web save path's matrix-membership validation gate.
const (
	// GroupSpecAuditors covers manager-spec, plan-auditor, sync-auditor.
	GroupSpecAuditors = "spec_auditors"
	// GroupDevelop covers manager-develop.
	GroupDevelop = "develop"
	// GroupAdvisor covers the non-writing super-advisor (mission-governor
	// retired from the catalog; the group keeps its name for old-config reads).
	GroupAdvisor = "advisor"
	// GroupDesignHarnessE2E covers manager-design, builder-harness, e2e-tester.
	GroupDesignHarnessE2E = "design_harness_e2e"
	// GroupDocs covers manager-docs.
	GroupDocs = "docs"
	// GroupGit covers manager-git.
	GroupGit = "git"
	// GroupLead covers manager-lead, the Tier L / factory coordinator.
	GroupLead = "lead"
	// GroupTodo covers manager-todo, the todo-queue management agent.
	GroupTodo = "todo"
	// GroupExplore covers the Anthropic built-in Explore read-only search
	// agent (not on the console roster; kept mapped for old-config reads).
	GroupExplore = "explore"
)

// agentGroupMembership is the agent-name → group layer. Agents with no entry
// (any user-added agent) resolve to the inherit sentinel and are never
// model-injected by the console surface.
var agentGroupMembership = map[string]string{
	"manager-spec":    GroupSpecAuditors,
	"plan-auditor":    GroupSpecAuditors,
	"sync-auditor":    GroupSpecAuditors,
	"manager-develop": GroupDevelop,
	"super-advisor":   GroupAdvisor,
	"manager-design":  GroupDesignHarnessE2E,
	"manager-lead":    GroupLead,
	"builder-harness": GroupDesignHarnessE2E,
	"e2e-tester":      GroupDesignHarnessE2E,
	"manager-docs":    GroupDocs,
	"manager-git":     GroupGit,
	"manager-todo":    GroupTodo,
	"Explore":         GroupExplore,
}

// AgentGroup returns the group an agent belongs to, and false when the agent
// has no membership (user-added agents).
func AgentGroup(agent string) (string, bool) {
	g, ok := agentGroupMembership[agent]
	return g, ok
}

// profileMatrixAgentOrder is the console's display/derivation roster for the
// client-side matrix island: the CANONICAL retained roster
// (template.RetainedAgents — the single roster literal in the tree) filtered
// to the agents with definition files. The built-in Explore has no file under
// .claude/agents/moai/ and is off the console surface, so it is filtered here
// by name — the one restatement this derivation needs, and the roster
// membership itself is rosterguard-asserted against the canonical literal.
var profileMatrixAgentOrder = func() []string {
	var out []string
	for _, name := range RetainedAgents() {
		if name == "Explore" {
			continue
		}
		out = append(out, name)
	}
	return out
}()

// ProfileMatrixAgents returns a defensive copy of the console roster order
// (the web/CLI matrix surfaces iterate this rather than restating a second
// literal).
func ProfileMatrixAgents() []string {
	out := make([]string, len(profileMatrixAgentOrder))
	copy(out, profileMatrixAgentOrder)
	return out
}

// defaultProfileMatrix is the per-agent model+effort Go-code SSOT: 13 mapped
// agents × 3 profile columns. Outer key: profile column {high, medium, low}.
// Inner key: agent NAME. Value: a {model, effort} pair in the OVERRIDE closed
// vocabulary — short model aliases and the 5-level efforts — because the
// console's select options (agentFMModelValues/agentFMEffortValues), the
// pin/clear comparison against the profile default (REQ-AFR-004: a submission
// equal to the default is CLEARED, so the two sides must be comparable), and
// config.validateAgentOverrides all speak that vocabulary.
//
// Cell re-derivation (plan §B-1(b)/§F M1 — the blueprint's 2026-09-27 inline
// literals are stale; divergence from them is not a defect):
//
//   - The MODEL axis derives from the CURRENT config defaults
//     config.NewDefaultLLMConfig().ClaudeModels = {high: opus, medium: sonnet,
//     low: haiku}: judgment/authoring rows take ClaudeModels.High, mechanical
//     rows ClaudeModels.Medium. The No-Haiku policy is carried over from the
//     pre-deletion matrix (a profile column never defaults an agent to haiku —
//     haiku stays available as an EXPLICIT override, REQ-AFR-006), so
//     ClaudeModels.Low never enters a cell.
//   - The EFFORT axis is the matrix's own judgment-weighted policy, expressed
//     with the EffortLevel* constants: the auditing/advising/coordinating rows
//     hold `high` in the upper columns; the authoring rows (manager-spec,
//     manager-develop) and the bounded-worker rows (e2e-tester, manager-todo)
//     hold `medium`; the mechanical rows (docs, git, Explore) hold `low`
//     profile-invariant. The economical column steps every non-advisor row
//     down one level. (The SPEC-AGENT-TIER-001 pairs
//     config.DefaultClaudeTier* were evaluated as the cell source and
//     rejected: their model values are full generation ids ("sonnet-5-5"),
//     outside the override closed set — a cell in that vocabulary would make
//     the REQ-AFR-004 clear-equals-default comparison unreachable and leave
//     the row select with no matching option.)
//   - The roster is the CURRENT catalog: mission-governor is gone,
//     manager-todo is in (following e2e-tester's bounded-worker shape).
//     Explore keeps its mapped cell so pre-restore overrides for it stay
//     resolvable, but it is not on the console roster.
//
// Invariants asserted by tests: models subset of {ClaudeModels.High,
// ClaudeModels.Medium} (zero haiku / fable / inherit); efforts subset of
// {low, medium, high}; rows monotone (depth never increases as the column
// descends); `inherit` survives only as the unmapped-agent fallback.
//
// @MX:ANCHOR: [AUTO] defaultProfileMatrix — per-agent model+effort SSOT (models from config claude_models defaults, efforts the judgment policy)
// @MX:REASON: [AUTO] fan_in >= 2 (ResolveAgentModelEffort resolver + web console matrix island); cells carry no inline model ids — config defaults and EffortLevel constants only
var defaultProfileMatrix = map[string]map[string]config.ModelEffort{
	config.ProfileHigh: {
		"manager-spec":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"plan-auditor":    {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"sync-auditor":    {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"manager-develop": {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"super-advisor":   {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"manager-design":  {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"manager-lead":    {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"builder-harness": {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"e2e-tester":      {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"manager-docs":    {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"manager-git":     {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"manager-todo":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"Explore":         {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
	},
	config.ProfileMedium: {
		"manager-spec":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"plan-auditor":    {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"sync-auditor":    {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"manager-develop": {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"super-advisor":   {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"manager-design":  {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"manager-lead":    {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"builder-harness": {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"e2e-tester":      {Model: defaultMatrixModelHigh, Effort: EffortLevelLow},
		"manager-docs":    {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"manager-git":     {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"manager-todo":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"Explore":         {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
	},
	config.ProfileLow: {
		"manager-spec":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"plan-auditor":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"sync-auditor":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"manager-develop": {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"super-advisor":   {Model: defaultMatrixModelHigh, Effort: EffortLevelHigh},
		"manager-design":  {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"manager-lead":    {Model: defaultMatrixModelHigh, Effort: EffortLevelMedium},
		"builder-harness": {Model: defaultMatrixModelHigh, Effort: EffortLevelLow},
		"e2e-tester":      {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"manager-docs":    {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"manager-git":     {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"manager-todo":    {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
		"Explore":         {Model: defaultMatrixModelMedium, Effort: EffortLevelLow},
	},
}

// defaultMatrixModelHigh/Medium are the two model aliases the matrix cells
// draw from — the CURRENT config claude_models defaults (High and Medium
// columns), resolved once at package init so the cells never restate the ids.
// ClaudeModels.Low (haiku) is deliberately unread: the No-Haiku policy keeps
// it out of every default cell.
var (
	defaultMatrixModelHigh   = config.NewDefaultLLMConfig().ClaudeModels.High
	defaultMatrixModelMedium = config.NewDefaultLLMConfig().ClaudeModels.Medium
)

// DefaultProfileMatrix returns a deep copy of the per-agent Go-code SSOT. A
// copy is returned so callers cannot mutate the package-level matrix.
func DefaultProfileMatrix() map[string]map[string]config.ModelEffort {
	out := make(map[string]map[string]config.ModelEffort, len(defaultProfileMatrix))
	for profile, agents := range defaultProfileMatrix {
		inner := make(map[string]config.ModelEffort, len(agents))
		maps.Copy(inner, agents)
		out[profile] = inner
	}
	return out
}

// ResolveAgentModelEffort resolves an agent's effective {model, effort} under
// the active profile with the D2 precedence:
//  1. llm.agent_overrides[agent] if present → wins;
//  2. else the Go-default per-agent cell under cfg.EffectiveProfile();
//  3. an unrecognized effective profile falls back to the medium column
//     (parity with the medium-default resolution of EffectiveProfile);
//  4. agent absent from the matrix → {inherit, ""}, unmapped.
//
// The returned bool `mapped` is false for the inherit case (a user-added
// agent that is not in the catalog), letting the caller skip model injection.
// Lookup is by agent NAME, not by group: the group layer carries no routing
// information here (AgentGroup survives for display and as the web save
// path's membership gate).
//
// This function is the restored console surface's SINGLE derivation site —
// internal/web must call it, never re-derive (REQ-AFR-010 lineage; the
// narrowed mcp_audit_surface guard enforces the definition ban).
//
// @MX:ANCHOR: [AUTO] ResolveAgentModelEffort — profile → per-agent {model, effort} resolver
// @MX:REASON: [AUTO] fan_in >= 2 (web console render + save paths + client matrix island); precedence order (override → Go default → medium fallback → inherit) is load-bearing
func ResolveAgentModelEffort(cfg config.LLMConfig, agent string) (me config.ModelEffort, mapped bool) {
	// (1) per-agent override wins.
	if ov, ok := cfg.AgentOverrides[agent]; ok {
		return ov, true
	}

	// (2) Go-default per-agent cell under the effective profile.
	profile := cfg.EffectiveProfile()
	if agents, ok := defaultProfileMatrix[profile]; ok {
		if cell, ok := agents[agent]; ok {
			return cell, true
		}
	}

	// (3) Unrecognized profile falls back to the medium column.
	if cell, ok := defaultProfileMatrix[config.ProfileMedium][agent]; ok {
		return cell, true
	}

	// (4) not in the catalog — inherit sentinel, never injected.
	return config.ModelEffort{Model: ModelInherit, Effort: ""}, false
}
