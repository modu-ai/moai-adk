package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ModelPolicy represents the token consumption tier for agent models.
type ModelPolicy string

const (
	// ModelPolicyHigh uses explicit opus for most agents (Max $200 plan, highest quality).
	ModelPolicyHigh ModelPolicy = "high"
	// ModelPolicyMedium uses opus for critical agents, sonnet for standard, haiku for mechanical (Max $100 plan).
	ModelPolicyMedium ModelPolicy = "medium"
	// ModelPolicyLow uses no opus (Plus $20 plan). Sonnet for core agents, Haiku for the rest.
	ModelPolicyLow ModelPolicy = "low"
)

// DefaultModelPolicy is the default model policy for new projects.
const DefaultModelPolicy = ModelPolicyHigh

// ValidModelPolicies returns all valid model policy values.
func ValidModelPolicies() []string {
	return []string{string(ModelPolicyHigh), string(ModelPolicyMedium), string(ModelPolicyLow)}
}

// IsValidModelPolicy checks if the given string is a valid model policy.
func IsValidModelPolicy(s string) bool {
	switch ModelPolicy(s) {
	case ModelPolicyHigh, ModelPolicyMedium, ModelPolicyLow:
		return true
	}
	return false
}

// ModelIDOpus48 is the canonical model ID for Claude Opus 4.8.
// Used by launcher.go to route the new model and by profile translations.
const ModelIDOpus48 = "claude-opus-4-8"

// ModelAliasTable is the single source of truth mapping short model aliases
// (the user-facing wizard picker values) to their canonical Claude Code model
// ids. Add a new row whenever a new alias is introduced; every call site that
// needs the alias→id resolution MUST read from this table rather than
// hard-coding a literal, so the mapping stays in one place.
//
// The forward direction (alias → canonical id) is used by expandModelString in
// launcher.go. The reverse direction (canonical id → alias) is performed by
// ModelAliasFromCanonicalID, which consults ModelAliasTable for the current id
// and ModelDeprecatedCanonicalIDs for superseded ids that still appear in
// historical prefs files.
//
// opusplan is a Claude Code native routing alias (Opus for planning, Sonnet for
// coding) with no standalone full-id form; it maps to itself so the table is
// total over the wizard picker surface.
//
// @MX:ANCHOR: [AUTO] ModelAliasTable — single SSOT for alias↔canonical-id mapping
// @MX:REASON: [AUTO] fan_in >= 3 (launcher.go expandModelString + profile_setup.go normalizeModel + settings/schema.go modelOptions); hardcoding-prevention per CLAUDE.local.md §14
var ModelAliasTable = map[string]string{
	"opus":     ModelIDOpus48,
	"sonnet":   "claude-sonnet-5",
	"fable":    "claude-fable-5",
	"haiku":    "claude-haiku-4-5",
	"opusplan": "opusplan", // CC-native routing alias, no full-id expansion
}

// ModelDeprecatedCanonicalIDs maps superseded canonical model ids to their
// short alias, so wizard migration can normalize historical prefs values that
// predate the current canonical id. A new row is added whenever a model is
// bumped to a newer version (e.g. claude-opus-4-7 → claude-opus-4-8); the old
// id stays here so existing prefs files keep resolving to the right alias.
//
// This is the reverse-companion of ModelAliasTable. The current canonical id
// lives ONLY in ModelAliasTable; deprecated predecessors live ONLY here, so
// there is exactly one home for each id and no duplication.
var ModelDeprecatedCanonicalIDs = map[string]string{
	"claude-opus-4-6":   "opus",
	"claude-opus-4-7":   "opus",
	"claude-sonnet-4-6": "sonnet",
}

// ModelAliasCanonicalID returns the canonical Claude Code model id for the
// given short alias. It is the programmatic accessor for ModelAliasTable so
// callers do not reach into the map literal directly. When the alias is absent
// from the table the input is returned unchanged (callers may then treat it as
// an already-canonical id or an unknown value).
func ModelAliasCanonicalID(alias string) string {
	if id, ok := ModelAliasTable[alias]; ok {
		return id
	}
	return alias
}

// ModelAliasFromCanonicalID returns the short alias for a canonical model id,
// performing the reverse lookup of ModelAliasTable. It also consults
// ModelDeprecatedCanonicalIDs so historical prefs values carrying superseded
// ids still resolve to the correct alias. Used by wizard migration paths that
// must normalize deprecated full-id prefs values back to the user-facing alias
// surface. When the id is absent from both tables the input is returned
// unchanged (caller decides how to handle unknown ids).
func ModelAliasFromCanonicalID(canonicalID string) string {
	for alias, id := range ModelAliasTable {
		if id == canonicalID {
			return alias
		}
	}
	if alias, ok := ModelDeprecatedCanonicalIDs[canonicalID]; ok {
		return alias
	}
	return canonicalID
}

// ModelAliasPickerValues returns the ordered list of short aliases presented to
// users in the profile wizard model picker. The list is the user-facing surface
// that the wizard, the settings schema, and the advanced wizard gate all
// consume, so they stay in sync by reading from here rather than re-declaring
// the literal array. The [1m] variants are listed alongside the base alias
// because the [1m] suffix is a Claude Code native context-window modifier, not
// a separate model.
func ModelAliasPickerValues() []string {
	return []string{
		"opus", "opus[1m]",
		"sonnet", "sonnet[1m]",
		"fable", "fable[1m]",
		"haiku",
		"opusplan",
	}
}

// Effort level constants for the 5-tier effort system.
// These are separate from ModelPolicy (3-tier). ModelPolicy selects the model;
// effort levels control reasoning depth within a model session.
// Supported by Claude Code v2.1.68+ for Opus 4.6 and Opus 4.7.
const (
	// EffortLevelLow is the fastest, least thorough effort level.
	EffortLevelLow = "low"
	// EffortLevelMedium is the balanced default effort level.
	EffortLevelMedium = "medium"
	// EffortLevelHigh activates deep reasoning for complex tasks.
	EffortLevelHigh = "high"
	// EffortLevelXHigh is extended high reasoning for Opus 4.7+.
	// Not supported on Opus 4.6.
	EffortLevelXHigh = "xhigh"
	// EffortLevelMax is the maximum effort level.
	// On Opus 4.6, max is the highest supported level.
	// On Opus 4.7+, xhigh and max are both available.
	EffortLevelMax = "max"
)

// effortLineRegex matches an existing effort: field in YAML frontmatter.
var effortLineRegex = regexp.MustCompile(`(?m)^effort:\s*\S+`)

// frontmatterOpenPrefix is the opening --- delimiter of a YAML frontmatter block.
// insertEffortInFrontmatter locates the second (closing) --- via string splitting.
var frontmatterOpenPrefix = "---\n"

// insertEffortInFrontmatter inserts "effort: <level>" into the YAML frontmatter block
// of content. Returns unchanged content when:
//   - the file does not start with "---\n" (no frontmatter)
//   - no closing "---" is found
//   - an effort: line already exists (caller is responsible for this guard)
func insertEffortInFrontmatter(content []byte, effortLevel string) []byte {
	s := string(content)
	if !strings.HasPrefix(s, frontmatterOpenPrefix) {
		return content // No YAML frontmatter
	}
	// Locate the closing --- after the opening one.
	// Content after the opening "---\n" (offset 4) is searched for "\n---".
	rest := s[len(frontmatterOpenPrefix):]
	closingIdx := strings.Index(rest, "\n---")
	if closingIdx == -1 {
		return content // Malformed frontmatter — leave untouched
	}
	// closingIdx points at the "\n" before "---". Insert before the newline.
	insertPos := len(frontmatterOpenPrefix) + closingIdx + 1 // position of the closing "---" line
	return []byte(s[:insertPos] + "effort: " + effortLevel + "\n" + s[insertPos:])
}

// modelLineRegex matches the "model:" line in YAML frontmatter.
var modelLineRegex = regexp.MustCompile(`(?m)^model:\s*\S+`)

// Performance tier tokens — the canonical {max, medium, low} vocabulary of the
// --model-policy CLI flag (SPEC-AGENT-ARCH-V2-001 M3c) and the tier axis of the
// plan_type × tier profile structure. Named constants per CLAUDE.local.md §14
// (no magic literals for the closed-set tokens).
const (
	// PerformanceTierMax is the highest-quality tier column.
	PerformanceTierMax = "max"
	// PerformanceTierMedium is the balanced default tier column (default-when-absent per plan.md D6).
	PerformanceTierMedium = "medium"
	// PerformanceTierLow is the economical tier column.
	PerformanceTierLow = "low"
)

// performanceTierRegex matches the performance_tier: line in llm.yaml.
var performanceTierRegex = regexp.MustCompile(`(?m)^(\s*)performance_tier:\s*["']?[\w-]*["']?`)

// ValidPerformanceTiers returns the closed set of valid No-Haiku performance
// tiers (SPEC-AGENT-ARCH-V2-001 M3c, design.md §D.2).
func ValidPerformanceTiers() []string {
	return []string{PerformanceTierMax, PerformanceTierMedium, PerformanceTierLow}
}

// IsValidPerformanceTier checks if the given string is a valid performance tier.
func IsValidPerformanceTier(s string) bool {
	for _, t := range ValidPerformanceTiers() {
		if s == t {
			return true
		}
	}
	return false
}

// ApplyPerformanceTier patches the performance_tier field in llm.yaml under
// the given project root (SPEC-AGENT-ARCH-V2-001 M3c, REQ-AA2-010). It reads
// .moai/config/sections/llm.yaml, replaces the performance_tier: line with
// the new tier value, and writes the file back. Returns nil if the file is
// absent (graceful no-op). The tier MUST be validated by the caller.
//
// @MX:ANCHOR: [AUTO] ApplyPerformanceTier — performance_tier persistence entry point
// @MX:REASON: fan_in >= 3 expected (init.go, update.go, tests)
func ApplyPerformanceTier(projectRoot, tier string) error {
	llmPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read llm.yaml: %w", err)
	}

	replacement := "${1}performance_tier: " + tier
	newContent := performanceTierRegex.ReplaceAll(content, []byte(replacement))

	if string(newContent) == string(content) {
		return nil
	}

	if err := os.WriteFile(llmPath, newContent, 0o644); err != nil {
		return fmt.Errorf("write llm.yaml: %w", err)
	}
	return nil
}

// modelInherit is the "inherit" model sentinel. Agents whose profile model is
// inherit (only Explore) are skipped by the apply pass — inherit is never
// written as a model: value, and Explore has no agent file on disk.
const modelInherit = "inherit"

// TierProfileEntry is one atomic {model, effort} assignment for a single agent
// at one (plan_type, tier) coordinate. It replaces the split agentModelMap
// ([3]string) + agentEffortMap representation with one unit that carries both
// dimensions together (REQ-MTP-006).
type TierProfileEntry struct {
	// Model is the Claude Code short model alias (opus/sonnet/fable/inherit).
	Model string
	// Effort is the reasoning effort level (low/medium/high/xhigh/max).
	Effort string
}

// tierProfileRow holds the three tier columns for one agent, indexed by
// tierColumnIndex: [max, medium, low].
type tierProfileRow = [3]TierProfileEntry

// tierColumnIndex maps a performance tier token to its column index in a
// tierProfileRow. Unknown tiers fall back to the medium column (plan.md D6).
var tierColumnIndex = map[string]int{
	PerformanceTierMax:    0,
	PerformanceTierMedium: 1,
	PerformanceTierLow:    2,
}

// tierProfiles is the single source of truth for the plan_type × tier × agent →
// {model, effort} assignment, replacing the dual agentModelMap + agentEffortMap
// (REQ-MTP-006). Outer key: plan type (config.PlanTypeAPI | config.PlanTypeSubscription).
// Middle key: agent name. Value: the three tier columns (max, medium, low).
//
// The api rows equal spec.md §B.6 (Plan A rev2) and the subscription rows equal
// spec.md §B.7 (Plan B), copied cell-for-cell — these values are settled design
// input (the original 60 cells verified 60/60 against the model-tier-redesign
// report; the e2e-specialist row added by SPEC-E2E-REVIVAL-001 mirrors the
// builder-harness execution-class profile, 66 cells total) and MUST NOT be
// re-derived or "improved" without a spec change. Auditors, manager-design, and
// super-advisor carry explicit rows (moved off implicit inherit); Explore keeps
// an inherit row for display/derivation surfaces only (the apply pass skips it).
//
// @MX:NOTE: [AUTO] SSOT for plan_type × tier agent {model, effort}; api = spec §B.6, subscription = spec §B.7 (verbatim, do not re-derive)
var tierProfiles = map[string]map[string]tierProfileRow{
	config.PlanTypeAPI: {
		// Agent            max                       medium                    low
		"manager-spec":    {{"fable", "high"}, {"fable", "high"}, {"opus", "high"}},
		"plan-auditor":    {{"fable", "high"}, {"fable", "high"}, {"opus", "high"}},
		"sync-auditor":    {{"fable", "high"}, {"opus", "high"}, {"opus", "medium"}},
		"manager-design":  {{"fable", "high"}, {"fable", "high"}, {"opus", "high"}},
		"super-advisor":   {{"fable", "xhigh"}, {"fable", "high"}, {"opus", "high"}},
		"manager-develop": {{"fable", "high"}, {"opus", "high"}, {"opus", "medium"}},
		"builder-harness": {{"opus", "high"}, {"opus", "medium"}, {"opus", "medium"}},
		"e2e-specialist":  {{"opus", "high"}, {"opus", "medium"}, {"opus", "medium"}},
		"manager-docs":    {{"sonnet", "medium"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"manager-git":     {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"Explore":         {{modelInherit, "medium"}, {modelInherit, "low"}, {modelInherit, "low"}},
	},
	config.PlanTypeSubscription: {
		// Agent            max                       medium                    low
		"manager-spec":    {{"opus", "high"}, {"opus", "high"}, {"opus", "medium"}},
		"plan-auditor":    {{"opus", "high"}, {"opus", "medium"}, {"sonnet", "high"}},
		"sync-auditor":    {{"opus", "high"}, {"opus", "medium"}, {"sonnet", "high"}},
		"manager-design":  {{"opus", "high"}, {"opus", "medium"}, {"sonnet", "high"}},
		"super-advisor":   {{"opus", "xhigh"}, {"opus", "high"}, {"opus", "medium"}},
		"manager-develop": {{"sonnet", "high"}, {"sonnet", "high"}, {"sonnet", "high"}},
		"builder-harness": {{"sonnet", "high"}, {"sonnet", "medium"}, {"sonnet", "medium"}},
		"e2e-specialist":  {{"sonnet", "high"}, {"sonnet", "medium"}, {"sonnet", "medium"}},
		"manager-docs":    {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"manager-git":     {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"Explore":         {{modelInherit, "medium"}, {modelInherit, "low"}, {modelInherit, "low"}},
	},
}

// tierProfileAgentOrder is the canonical display/derivation order of the 11
// retained agents for the model-policy preview surfaces (REQ-MTP-009 — Explore is
// included for display even though the apply pass skips its inherit row). The order
// matches the spec.md §B.6/§B.7 matrix row order so the web preview reads top-down
// like the design tables.
var tierProfileAgentOrder = []string{
	"manager-spec",
	"plan-auditor",
	"sync-auditor",
	"manager-design",
	"super-advisor",
	"manager-develop",
	"builder-harness",
	"e2e-specialist",
	"manager-docs",
	"manager-git",
	"Explore",
}

// TierProfileAgents returns a copy of the canonical display order of the agents
// carried by the plan_type tier profiles. The web /model-policy preview iterates
// this to derive its per-tier {model, effort} cells from the single Go structure
// (REQ-MTP-023 — the web layer must NOT re-declare the matrix as a second literal).
// A defensive copy is returned so callers cannot mutate the package-level order.
func TierProfileAgents() []string {
	out := make([]string, len(tierProfileAgentOrder))
	copy(out, tierProfileAgentOrder)
	return out
}

// GetTierProfileEntry returns the {model, effort} assignment for an agent under
// the given plan type and tier (REQ-MTP-006..009). The bool is false when the
// agent has no row in the plan's profile (unknown/user-added agent) — the apply
// pass uses the false return to skip the file byte-identically (REQ-MTP-013).
// An unrecognized plan type falls back to the subscription profile (backward-
// compatibility default per REQ-MTP-002); an unrecognized tier falls back to the
// medium column (plan.md D6 default-when-absent).
func GetTierProfileEntry(planType, agentName, tier string) (TierProfileEntry, bool) {
	plan, ok := tierProfiles[planType]
	if !ok {
		plan = tierProfiles[config.DefaultPlanType] // subscription fallback (BC)
	}
	row, ok := plan[agentName]
	if !ok {
		return TierProfileEntry{}, false
	}
	idx, ok := tierColumnIndex[tier]
	if !ok {
		idx = tierColumnIndex[PerformanceTierMedium] // D6 default-when-absent
	}
	return row[idx], true
}

// MapModelPolicyToTier translates a legacy template.ModelPolicy value
// ({high, medium, low}) to the canonical performance tier ({max, medium, low}):
// high→max, medium→medium, low→low (REQ-MTP-012). It maps the TIER dimension
// ONLY and returns NO plan value — plan selection is owned solely by the
// effective plan_type resolution (config.LLMConfig.EffectivePlanType), NEVER by
// this function. An empty or unrecognized policy falls back to the medium tier
// (plan.md D6 default-when-absent).
func MapModelPolicyToTier(policy ModelPolicy) string {
	switch policy {
	case ModelPolicyHigh:
		return PerformanceTierMax
	case ModelPolicyMedium:
		return PerformanceTierMedium
	case ModelPolicyLow:
		return PerformanceTierLow
	default:
		return PerformanceTierMedium
	}
}

// MapModelPolicyToEffort translates a legacy template.ModelPolicy value
// ({high, medium, low}) to the runtime-LAUNCH effort level vocabulary
// (EffortLevelHigh/Medium/Low): high→high, medium→medium, low→low. This is the
// runtime-LAUNCH effort projection of the legacy vocabulary and is DISTINCT from
// MapModelPolicyToTier — which projects high→max on the TIER axis {max, medium,
// low} for the init/update apply pass. Reuse MapModelPolicyToTier for the tier
// dimension and this function for the effort dimension; NEVER substitute one for
// the other, since the vocabularies differ (high→max in tier vs high→high in
// effort). An empty or unrecognized policy returns "" (no override), so an absent
// model_policy preserves today's launch behavior byte-identically.
func MapModelPolicyToEffort(policy ModelPolicy) string {
	switch policy {
	case ModelPolicyHigh:
		return EffortLevelHigh
	case ModelPolicyMedium:
		return EffortLevelMedium
	case ModelPolicyLow:
		return EffortLevelLow
	default:
		return ""
	}
}

// NormalizeToTier resolves any performance-policy string to the canonical tier
// vocabulary {max, medium, low}. It accepts the canonical performance-tier
// tokens verbatim and bridges the legacy ModelPolicy vocabulary via
// MapModelPolicyToTier (high→max). Empty or unrecognized input falls back to the
// medium tier (plan.md D6). This is the call-site resolver used by the init and
// update apply paths, whose ModelPolicy vocabularies differ ({max,medium,low}
// vs {high,medium,low}).
func NormalizeToTier(s string) string {
	if IsValidPerformanceTier(s) {
		return s
	}
	return MapModelPolicyToTier(ModelPolicy(s))
}

// planTypeLineRegex matches an uncommented plan_type: field in llm.yaml,
// capturing the value. A commented line (`# plan_type: ...`) does not match
// because "#" is not leading whitespace.
var planTypeLineRegex = regexp.MustCompile(`(?m)^\s*plan_type:\s*["']?([\w-]+)["']?`)

// ResolveProjectPlanType reads the effective plan type from the project's
// llm.yaml (REQ-MTP-002). An absent file, an absent/commented plan_type key, or
// an empty value resolves to the subscription default via
// config.LLMConfig.EffectivePlanType; any explicit value passes through verbatim
// (GetTierProfileEntry falls back to subscription for an out-of-set value).
func ResolveProjectPlanType(projectRoot string) string {
	llmPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		return config.DefaultPlanType
	}
	var pt string
	if m := planTypeLineRegex.FindSubmatch(content); len(m) >= 2 {
		pt = string(m[1])
	}
	return config.LLMConfig{PlanType: pt}.EffectivePlanType()
}

// planTypePersistRegex matches the plan_type: line in llm.yaml for a
// value-replacing write, capturing the leading indentation (group 1). Unlike
// planTypeLineRegex (which captures the value and requires a non-empty value),
// this uses `[\w-]*` so an empty value (`plan_type: ""`) is also matched and
// rewritten — mirroring performanceTierRegex.
var planTypePersistRegex = regexp.MustCompile(`(?m)^(\s*)plan_type:\s*["']?[\w-]*["']?`)

// ApplyPlanType patches the plan_type field in llm.yaml under the given project
// root (REQ-MTP-016/018), mirroring ApplyPerformanceTier. It reads
// .moai/config/sections/llm.yaml, replaces the plan_type: line with the new
// value (preserving indentation), and writes the file back. Returns nil when the
// file is absent (graceful no-op) or when the plan_type line already carries the
// target value. The planType MUST be validated by the caller (config.IsValidPlanType).
//
// @MX:ANCHOR: [AUTO] ApplyPlanType — plan_type persistence entry point
// @MX:REASON: [AUTO] fan_in >= 2 (init.go, update.go); shipped-plan_type SSOT persistence, mirrors ApplyPerformanceTier
func ApplyPlanType(projectRoot, planType string) error {
	llmPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read llm.yaml: %w", err)
	}

	replacement := "${1}plan_type: " + planType
	newContent := planTypePersistRegex.ReplaceAll(content, []byte(replacement))

	if string(newContent) == string(content) {
		return nil
	}

	if err := os.WriteFile(llmPath, newContent, 0o644); err != nil {
		return fmt.Errorf("write llm.yaml: %w", err)
	}
	return nil
}

// performanceTierValueRegex captures the persisted performance_tier value
// (group 1) from llm.yaml, requiring a non-empty value (an empty
// `performance_tier: ""` falls through to the medium default).
var performanceTierValueRegex = regexp.MustCompile(`(?m)^\s*performance_tier:\s*["']?([\w-]+)["']?`)

// ResolveProjectPerformanceTier reads the persisted performance_tier from the
// project's llm.yaml (REQ-MTP-018 — the update path's tier source). An absent
// file, an absent/commented performance_tier key, or an empty value resolves to
// the medium default (plan.md D6 default-when-absent). Any explicit value passes
// through verbatim; NormalizeToTier at the apply site bridges legacy vocabularies.
func ResolveProjectPerformanceTier(projectRoot string) string {
	llmPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		return PerformanceTierMedium
	}
	if m := performanceTierValueRegex.FindSubmatch(content); len(m) >= 2 && len(m[1]) > 0 {
		return string(m[1])
	}
	return PerformanceTierMedium
}

// ApplyTierProfile patches each shipped agent file's model: AND effort:
// frontmatter in a single traversal, using the {model, effort} pair from the
// plan_type × tier profile (spec.md §B.6/§B.7). It replaces the former two-pass
// ApplyModelPolicy + ApplyEffortPolicy sequence at both production call sites
// (REQ-MTP-010).
//
// Precedence (REQ-MTP-011, plan.md D1 replace-both): BOTH the existing model:
// line AND the existing effort: line are REPLACED with the profile values. The
// tier profile is the single source of truth for shipped-agent model/effort; the
// historical preserve-existing-effort behavior is intentionally dropped because
// all shipped agent files already carry effort:, which would otherwise render
// the entire effort matrix inert. A user who wants to pin a value re-pins after
// init/update (identical to today's model: replace semantics).
//
// An agent whose name has no row in the profile (unknown/user-added) is left
// byte-identical (REQ-MTP-013). Agents whose profile model is "inherit" (Explore)
// are skipped — Explore has no agent file, and inherit is never written as a
// model: value. Manifest hashes are updated for every written file.
//
// @MX:ANCHOR: [AUTO] ApplyTierProfile — unified model+effort apply pass; called from initializer and update paths
// @MX:REASON: [AUTO] fan_in >= 2 (initializer.go + update.go); public API boundary replacing ApplyModelPolicy + ApplyEffortPolicy; replace-both precedence is load-bearing (preserve-effort would render the effort matrix inert)
func ApplyTierProfile(projectRoot, planType, tier string, mgr manifest.Manager) error {
	// Agents are consolidated under .claude/agents/moai/.
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "moai")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // agents directory absent — nothing to patch
		}
		return fmt.Errorf("read agents directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		agentName := strings.TrimSuffix(entry.Name(), ".md")
		profile, ok := GetTierProfileEntry(planType, agentName, tier)
		if !ok {
			continue // unknown/user-added agent — preserve byte-identical (REQ-MTP-013)
		}
		if profile.Model == modelInherit {
			continue // inherit agents (Explore) are never patched
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read agent file %q: %w", entry.Name(), err)
		}

		// Replace-both precedence (D1): rewrite model: AND effort: to the profile
		// values in one traversal.
		newContent := modelLineRegex.ReplaceAll(content, []byte("model: "+profile.Model))
		if effortLineRegex.Match(newContent) {
			newContent = effortLineRegex.ReplaceAll(newContent, []byte("effort: "+profile.Effort))
		} else {
			newContent = insertEffortInFrontmatter(newContent, profile.Effort)
		}

		if string(newContent) == string(content) {
			continue // already at the target model + effort — no write
		}

		if err := os.WriteFile(filePath, newContent, 0o644); err != nil {
			return fmt.Errorf("write agent file %q: %w", entry.Name(), err)
		}

		// Update manifest hash for the patched file.
		relPath := filepath.Join(".claude", "agents", "moai", entry.Name())
		hash := manifest.HashBytes(newContent)
		if err := mgr.Track(relPath, manifest.TemplateManaged, hash); err != nil {
			return fmt.Errorf("track patched agent %q: %w", entry.Name(), err)
		}
	}

	return nil
}
