package template

// profile_matrix.go — SPEC-MODEL-PROFILE-MATRIX-001 M1/M2: the Matrix A
// per-agent-group model+effort profile matrix (Go-code SSOT), the agent-GROUP →
// agent-name membership SSOT, and the runtime resolver that maps the active
// profile + per-agent overrides → each agent's {model, effort}. This replaces
// the retired 66-cell tierProfiles (plan_type × tier) with a single 3-column
// profile axis (max/medium/low) consumed via runtime-arg spawn injection rather
// than agent-frontmatter mutation.

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// leadingWS returns the count of leading space/tab characters in a line.
func leadingWS(line string) int {
	n := 0
	for _, r := range line {
		if r == ' ' || r == '\t' {
			n++
			continue
		}
		break
	}
	return n
}

// stripRetiredLLMKeys removes the retired `plan_type:` line and the entire
// `claude_models:` block (header + its more-indented child lines) from llm.yaml
// content (REQ-MPM-005 write-time removal of retired fields). A line-based
// processor is used rather than a regex so the multi-line block is handled
// robustly by indentation depth. Returns the cleaned content.
func stripRetiredLLMKeys(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	out := make([]string, 0, len(lines))
	skipBlockIndent := -1 // -1 = not inside a stripped block
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if skipBlockIndent >= 0 {
			// Inside a stripped block: skip blank lines and lines indented deeper
			// than the block header; a line at the header's indent or shallower ends
			// the block.
			if trimmed == "" || leadingWS(line) > skipBlockIndent {
				continue
			}
			skipBlockIndent = -1 // block ended; fall through to normal handling
		}
		if strings.HasPrefix(trimmed, "plan_type:") {
			continue // drop the retired plan_type line
		}
		if strings.HasPrefix(trimmed, "claude_models:") {
			skipBlockIndent = leadingWS(line) // begin skipping the block
			continue
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n"))
}

// profileLineRegex matches the profile: line in llm.yaml for a value-replacing
// write, capturing the leading indentation (group 1). Uses `[\w-]*` so an empty
// value (`profile: ""`) is also matched and rewritten.
var profileLineRegex = regexp.MustCompile(`(?m)^(\s*)profile:\s*["']?[\w-]*["']?`)

// ApplyProfile patches the profile field in llm.yaml under the given project
// root (REQ-MPM-016), mirroring ApplyPerformanceTier. It reads
// .moai/config/sections/llm.yaml, replaces the profile: line with the new value
// (preserving indentation), and writes the file back. Returns nil when the file
// is absent (graceful no-op) or when the profile line already carries the target
// value. The profile MUST be validated by the caller (config.IsValidProfile).
//
// @MX:ANCHOR: [AUTO] ApplyProfile — llm.profile persistence entry point (init/update/web)
// @MX:REASON: [AUTO] fan_in >= 2 (initializer + update); the shipped-profile SSOT persistence, mirrors ApplyPerformanceTier
func ApplyProfile(projectRoot, profile string) error {
	// The superseded top-column name is readable but never written back.
	profile = config.NormalizeProfile(profile)
	llmPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read llm.yaml: %w", err)
	}

	original := content
	// Write-time removal of retired fields (REQ-MPM-005): strip plan_type + the
	// claude_models block before persisting the profile.
	content = stripRetiredLLMKeys(content)

	var newContent []byte
	if profileLineRegex.Match(content) {
		newContent = profileLineRegex.ReplaceAll(content, []byte("${1}profile: "+profile))
	} else {
		// Legacy config with no profile: key — insert one under the llm: root
		// (migration: REQ-MPM-005 write-time schema upgrade).
		newContent = llmRootRegex.ReplaceAll(content, []byte("${0}\n    profile: "+profile))
	}
	if string(newContent) == string(original) {
		return nil
	}

	if err := os.WriteFile(llmPath, newContent, 0o644); err != nil {
		return fmt.Errorf("write llm.yaml: %w", err)
	}
	return nil
}

// llmRootRegex matches the top-level `llm:` key line for profile insertion into
// a legacy config that has no profile: key.
var llmRootRegex = regexp.MustCompile(`(?m)^llm:[ \t]*$`)

// Profile agent-group keys (REQ-MPM-011). The seven groups partition the
// retained agents by model+effort class. git, docs, and explore rows are
// profile-invariant.
const (
	// GroupSpecAuditors covers manager-spec, plan-auditor, sync-auditor.
	GroupSpecAuditors = "spec_auditors"
	// GroupDevelop covers manager-develop.
	GroupDevelop = "develop"
	// GroupAdvisor covers super-advisor.
	GroupAdvisor = "advisor"
	// GroupDesignHarnessE2E covers manager-design, builder-harness, e2e-tester.
	GroupDesignHarnessE2E = "design_harness_e2e"
	// GroupDocs covers manager-docs.
	GroupDocs = "docs"
	// GroupGit covers manager-git.
	GroupGit = "git"
	// GroupExplore covers the Anthropic built-in Explore read-only search agent.
	// Assigned sonnet/low profile-invariantly (product decision); only
	// user-added agents fall through to the inherit sentinel now.
	GroupExplore = "explore"
)

// agentGroupMembership is the agent-name → group SSOT (REQ-MPM-011). Agents with
// no entry (any user-added agent) resolve to the inherit sentinel and are never
// model-injected (REQ-MPM-013 — scope narrowed by product decision: the
// built-in Explore now has an explicit group, so only user-added agents
// inherit).
var agentGroupMembership = map[string]string{
	"manager-spec":    GroupSpecAuditors,
	"plan-auditor":    GroupSpecAuditors,
	"sync-auditor":    GroupSpecAuditors,
	"manager-develop": GroupDevelop,
	"super-advisor":   GroupAdvisor,
	"manager-design":  GroupDesignHarnessE2E,
	"builder-harness": GroupDesignHarnessE2E,
	"e2e-tester":      GroupDesignHarnessE2E,
	"manager-docs":    GroupDocs,
	"manager-git":     GroupGit,
	"Explore":         GroupExplore,
}

// profileMatrixAgentOrder is the canonical display/derivation order of the 11
// retained agents for the model-profile preview surfaces (REQ-MPM-020). Explore
// is included in the display and now resolves to its own explore group cell
// (sonnet/low, profile-invariant), no longer the inherit sentinel.
var profileMatrixAgentOrder = []string{
	"manager-spec",
	"plan-auditor",
	"sync-auditor",
	"manager-develop",
	"super-advisor",
	"manager-design",
	"builder-harness",
	"e2e-tester",
	"manager-docs",
	"manager-git",
	"Explore",
}

// ProfileMatrixAgents returns a defensive copy of the canonical display order of
// the retained agents (REQ-MPM-020 — the web/CLI preview iterates this to derive
// per-agent cells from the single Go structure rather than a second literal).
func ProfileMatrixAgents() []string {
	out := make([]string, len(profileMatrixAgentOrder))
	copy(out, profileMatrixAgentOrder)
	return out
}

// defaultProfileMatrix is the per-AGENT model+effort Go-code SSOT: 11 mapped
// agents x 3 profiles = 33 cells. Outer key: profile {high, medium, low}. Inner
// key: retained agent NAME (not a group — the group layer is display-only now,
// because per-agent cells split two of the former groups). Value: {model,
// effort}. This is the authoritative fallback for any cell absent from config
// llm.profiles (REQ-MPM-009).
//
// Cell derivation (each row is monotone: high >= medium >= low). The cells are
// anchored on a published long-horizon coding-agent benchmark that measures
// score, cost per task, output tokens, and agent steps at every effort level:
//
//   - Opus 5 dominates Sonnet 5 at EVERY effort on that benchmark: Opus 5 at
//     `low` scores higher AND costs less per task than Sonnet 5 at any level,
//     because Sonnet 5 spends a multiple of the agent steps and output tokens
//     to finish the same long-horizon task. Unit token price is therefore not
//     the cost driver — completion efficiency is. Opus is consequently the
//     model for every multi-turn agentic row.
//   - `xhigh` is retired from the matrix: on Opus 5 it scores the same as
//     `high` while costing materially more, so it is strictly dominated. `max`
//     is the only level above `high`, so a row that wants more than `high`
//     takes `max`.
//   - The phase-weighted override: the benchmark-derived cost anchor put
//     manager-develop at `medium` and derived the rest from it. That anchor is
//     superseded by an operator policy weighting the phase that decides what
//     gets built over the phase that types it out — plan (manager-spec,
//     plan-auditor) takes `max`, while run (manager-develop), review
//     (sync-auditor) and sync (manager-docs) take `high`. `medium` remains the
//     knee of the Opus 5 cost/score curve; the policy knowingly spends past it,
//     so this matrix is NOT the cost-minimal derivation and must not be
//     re-derived back onto the anchor.
//   - Under a GLM backend only part of this is observable: z.ai's 3-state
//     reasoning control collapses {medium, high} onto one state and
//     {xhigh, max} onto another, so the run row is unchanged (its coding-max
//     override already forced the ceiling) and so is review. Plan and sync do
//     move. See glm_effort_overlay.go.
//   - Sonnet 5 is retained ONLY for single-shot, input-dominated, non-agentic
//     rows (Explore search, manager-git mechanics) where the multi-step
//     completion failure does not apply and the lower input price does.
//
// Invariants asserted by tests: zero haiku; zero fable; models subset of
// {opus, sonnet}; efforts subset of {low, medium, high, max} (no `xhigh` cell);
// `inherit` never appears inside the matrix — it survives only as the
// unmapped-agent fallback.
//
// @MX:ANCHOR: [AUTO] defaultProfileMatrix — per-agent model+effort SSOT (33 cells)
// @MX:REASON: [AUTO] fan_in >= 3 (ResolveAgentModelEffort resolver + moai model profile CLI + web preview + harness class derivation); cells are settled design input, re-derivation forbidden
var defaultProfileMatrix = map[string]map[string]config.ModelEffort{
	PerformanceTierHigh: {
		"manager-spec":    {Model: "opus", Effort: EffortLevelMax},
		"plan-auditor":    {Model: "opus", Effort: EffortLevelMax},
		"sync-auditor":    {Model: "opus", Effort: EffortLevelHigh},
		"manager-develop": {Model: "opus", Effort: EffortLevelMax},
		"super-advisor":   {Model: "opus", Effort: EffortLevelMax},
		"manager-design":  {Model: "opus", Effort: EffortLevelHigh},
		"builder-harness": {Model: "opus", Effort: EffortLevelHigh},
		"e2e-tester":      {Model: "opus", Effort: EffortLevelMedium},
		"manager-docs":    {Model: "opus", Effort: EffortLevelHigh},
		"manager-git":     {Model: "sonnet", Effort: EffortLevelLow},
		"Explore":         {Model: "sonnet", Effort: EffortLevelLow},
	},
	PerformanceTierMedium: {
		"manager-spec":    {Model: "opus", Effort: EffortLevelMax},
		"plan-auditor":    {Model: "opus", Effort: EffortLevelMax},
		"sync-auditor":    {Model: "opus", Effort: EffortLevelHigh},
		"manager-develop": {Model: "opus", Effort: EffortLevelHigh},
		"super-advisor":   {Model: "opus", Effort: EffortLevelHigh},
		"manager-design":  {Model: "opus", Effort: EffortLevelMedium},
		"builder-harness": {Model: "opus", Effort: EffortLevelMedium},
		"e2e-tester":      {Model: "opus", Effort: EffortLevelLow},
		"manager-docs":    {Model: "opus", Effort: EffortLevelHigh},
		"manager-git":     {Model: "sonnet", Effort: EffortLevelLow},
		"Explore":         {Model: "sonnet", Effort: EffortLevelLow},
	},
	PerformanceTierLow: {
		"manager-spec":    {Model: "opus", Effort: EffortLevelLow},
		"plan-auditor":    {Model: "opus", Effort: EffortLevelLow},
		"sync-auditor":    {Model: "opus", Effort: EffortLevelLow},
		"manager-develop": {Model: "opus", Effort: EffortLevelLow},
		"super-advisor":   {Model: "opus", Effort: EffortLevelMedium},
		"manager-design":  {Model: "opus", Effort: EffortLevelLow},
		"builder-harness": {Model: "opus", Effort: EffortLevelLow},
		"e2e-tester":      {Model: "sonnet", Effort: EffortLevelLow},
		"manager-docs":    {Model: "sonnet", Effort: EffortLevelLow},
		"manager-git":     {Model: "sonnet", Effort: EffortLevelLow},
		"Explore":         {Model: "sonnet", Effort: EffortLevelLow},
	},
}

// Harness purpose classes — the taxonomy `/moai:harness` classifies a generated
// specialist into. The names are reused verbatim from the `workflow_agents`
// purpose taxonomy in workflow.yaml so the two surfaces share one vocabulary.
const (
	HarnessClassReadOnlyExtract     = "read-only-extract"
	HarnessClassMechanicalTransform = "mechanical-transform"
	HarnessClassSynthesize          = "synthesize"
	HarnessClassResearch            = "research"
	HarnessClassVerifyJudge         = "verify-judge"
	HarnessClassImplement           = "implement"
	HarnessClassDesignArchitecture  = "design-architecture"
)

// HarnessAgentModel is the model every generated harness specialist is pinned to.
// Harness agents are model-uniform on purpose: they are persistent, user-owned
// specialists whose differentiation is reasoning DEPTH, not model tier, so the
// effort axis alone separates them. Pinning is safe now that every current
// non-haiku model carries a 1M context window — the former inherit-by-default
// rule existed to preserve a 1M entitlement that pinning would have lost.
const HarnessAgentModel = "opus"

// harnessClassRow maps each harness purpose class onto the retained-agent row
// whose EFFORT the class inherits from defaultProfileMatrix. Only the effort is
// borrowed; the model is always HarnessAgentModel. Keeping the derivation as a
// pointer into the matrix means the harness surface cannot drift from it.
var harnessClassRow = map[string]string{
	HarnessClassReadOnlyExtract:     "Explore",
	HarnessClassMechanicalTransform: "manager-git",
	HarnessClassSynthesize:          "manager-docs",
	HarnessClassResearch:            "plan-auditor",
	HarnessClassVerifyJudge:         "sync-auditor",
	HarnessClassImplement:           "manager-develop",
	HarnessClassDesignArchitecture:  "manager-design",
}

// HarnessClasses returns the purpose-class names in a stable display order.
func HarnessClasses() []string {
	return []string{
		HarnessClassReadOnlyExtract,
		HarnessClassMechanicalTransform,
		HarnessClassSynthesize,
		HarnessClassResearch,
		HarnessClassVerifyJudge,
		HarnessClassImplement,
		HarnessClassDesignArchitecture,
	}
}

// ResolveHarnessAgentModelEffort returns the {model, effort} a generated harness
// specialist of the given purpose class receives under the active profile.
// Precedence: config llm.harness_agents[profile][class].effort when present,
// else the effort of the class's matrix row. The model is ALWAYS
// HarnessAgentModel regardless of source. An unknown class falls back to the
// `implement` class. The bool reports whether the class was recognized.
//
// @MX:ANCHOR: [AUTO] ResolveHarnessAgentModelEffort — /moai:harness generation model+effort entry point
// @MX:REASON: [AUTO] fan_in >= 2 (builder-harness generation guidance + moai model profile --harness display); the single derivation site keeping harness frontmatter aligned with the profile matrix
func ResolveHarnessAgentModelEffort(cfg config.LLMConfig, class string) (config.ModelEffort, bool) {
	known := true
	if _, ok := harnessClassRow[class]; !ok {
		class = HarnessClassImplement
		known = false
	}

	profile := cfg.EffectiveProfile()

	if classes, ok := cfg.HarnessAgents[profile]; ok {
		if cell, ok := classes[class]; ok && strings.TrimSpace(cell.Effort) != "" {
			return config.ModelEffort{Model: HarnessAgentModel, Effort: cell.Effort}, known
		}
	}

	row := harnessClassRow[class]
	groups, ok := defaultProfileMatrix[profile]
	if !ok {
		groups = defaultProfileMatrix[PerformanceTierMedium]
	}
	return config.ModelEffort{Model: HarnessAgentModel, Effort: groups[row].Effort}, known
}

// DefaultProfileMatrix returns a deep copy of the per-agent Go-code SSOT
// (REQ-MPM-009/010). Used to mirror the matrix into the template llm.yaml and as
// the authoritative resolver fallback. A copy is returned so callers cannot
// mutate the package-level matrix.
func DefaultProfileMatrix() map[string]map[string]config.ModelEffort {
	out := make(map[string]map[string]config.ModelEffort, len(defaultProfileMatrix))
	for profile, agents := range defaultProfileMatrix {
		inner := make(map[string]config.ModelEffort, len(agents))
		maps.Copy(inner, agents)
		out[profile] = inner
	}
	return out
}

// AgentGroup returns the profile group an agent belongs to, and false when the
// agent has no membership (user-added agents) (REQ-MPM-011/013).
func AgentGroup(agent string) (string, bool) {
	g, ok := agentGroupMembership[agent]
	return g, ok
}

// ResolveAgentModelEffort resolves an agent's effective {model, effort} under
// the active profile with the D2 precedence (REQ-MPM-012):
//  1. llm.agent_overrides[agent] if present → wins;
//  2. else the active profile's per-agent cell from config llm.profiles;
//  3. else the Go-default per-agent cell (defaultProfileMatrix);
//  4. agent absent from the matrix → {inherit, ""} (REQ-MPM-013).
//
// The returned bool `mapped` is false for the inherit case (a user-added agent
// that is not in the retained catalog), letting the caller skip model injection.
// The active profile is read from cfg via EffectiveProfile (profile →
// performance_tier alias → medium).
//
// Lookup is by agent NAME, not by group: per-agent cells split two of the former
// groups, so the group layer no longer carries routing information and survives
// only as a display classification (see AgentGroup).
//
// @MX:ANCHOR: [AUTO] ResolveAgentModelEffort — profile → per-agent {model, effort} resolver
// @MX:REASON: [AUTO] fan_in >= 3 (moai model profile CLI + web preview + orchestrator spawn guidance); the runtime-arg injection SSOT replacing frontmatter mutation; precedence order (override → config profile → Go default → inherit) is load-bearing
func ResolveAgentModelEffort(cfg config.LLMConfig, agent string) (me config.ModelEffort, mapped bool) {
	// (1) per-agent override wins.
	if ov, ok := cfg.AgentOverrides[agent]; ok {
		return ov, true
	}

	profile := cfg.EffectiveProfile()

	// (2) config-mirror per-agent cell, when present.
	if agents, ok := cfg.Profiles[profile]; ok {
		if cell, ok := agents[agent]; ok {
			return cell, true
		}
	}

	// (3) Go-default per-agent cell (authoritative fallback).
	if agents, ok := defaultProfileMatrix[profile]; ok {
		if cell, ok := agents[agent]; ok {
			return cell, true
		}
	}

	// (3b) Unknown profile falls back to the medium column (parity with the
	// medium-default resolution of EffectiveProfile / plan.md D6).
	if cell, ok := defaultProfileMatrix[PerformanceTierMedium][agent]; ok {
		return cell, true
	}

	// (4) not in the retained catalog — inherit sentinel, never injected.
	return config.ModelEffort{Model: modelInherit, Effort: ""}, false
}
