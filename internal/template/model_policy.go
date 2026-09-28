package template

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
// Medium since SPEC-CLI-WIZARD-RESTRUCTURE-001 (REQ-WIZ-009): it matches the
// most common plan and is the tier the init wizard pre-selects. ModelPolicyHigh
// is unchanged and remains a fully selectable tier — only the DEFAULT moved.
const DefaultModelPolicy = ModelPolicyMedium

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

// ModelIDOpus55 is the canonical model ID for Claude Opus 5.5 — the current
// target of the "opus" alias, which requires Claude Code v2.1.280 or later.
// Opus 5.5 has a 1M-token context window and 128K max output, is priced at
// $4/$20 per MTok, keeps adaptive thinking always on, and defaults to the
// `medium` effort level (defaults differ per model: `high` on most other
// effort-capable models, `xhigh` on Opus 4.7).
// Used by launcher.go to route the model and by profile translations.
const ModelIDOpus55 = "claude-opus-5-5"

// ModelIDOpus48 is the superseded canonical model ID for Claude Opus 4.8, now
// replaced by ModelIDOpus55. Retained as a named constant because historical
// prefs files still carry it; it resolves back to the "opus" alias via
// ModelDeprecatedCanonicalIDs (deprecated-id normalization).
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
	"opus":     ModelIDOpus55,
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
	ModelIDOpus48:       "opus",
	"claude-opus-5":     "opus", // superseded by ModelIDOpus55
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
	// 1M unification — opus/sonnet/fable are exposed ONLY as their [1m] variants
	// (1M context always on); haiku has no 1M variant and stays as the base alias.
	// The base aliases (opus/sonnet/fable) and the opusplan routing alias remain
	// valid input values (ModelAliasTable is unchanged), but are no longer
	// offered as picker options.
	return []string{
		"fable[1m]",
		"opus[1m]",
		"sonnet[1m]",
		"haiku",
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

// ModelInherit is the exported "inherit" model sentinel — the vocabulary for
// "no per-agent model routing", kept for the served-model expectation logic.
// whose profile model is inherit (user-added agents with no group membership)
// are never injected, and under a GLM backend every mapped agent's model is
// session-inherited (the t66 fold: the launcher maps llm.glm.models onto the
// session-global ANTHROPIC_DEFAULT_*_MODEL env, so per-agent model cells carry
// no differentiation there). Inherit is never written as a model: value. The
// built-in Explore now has an explicit explore group cell (sonnet/low) and is
// no longer an inherit agent.
const ModelInherit = "inherit"

// MapModelPolicyToEffort translates a template.ModelPolicy value
// ({high, medium, low}) to the runtime-LAUNCH effort level vocabulary
// (EffortLevelHigh/Medium/Low): high→high, medium→medium, low→low. This is the
// runtime-LAUNCH effort projection and remains DISTINCT from
// MapModelPolicyToTier — it targets the EFFORT axis, not the TIER axis, even
// though both are now identity mappings on the shared vocabulary. An empty or
// unrecognized policy returns "" (no override), so an absent model_policy
// preserves today's launch behavior byte-identically.
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
