package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/modu-ai/moai-adk/internal/config"
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

// ModelIDOpus5 is the canonical model ID for Claude Opus 5 — the current target
// of the "opus" alias and the default Opus model as of Claude Code v2.1.219
// (native 1M context). Opus 5 is priced identically to its predecessor Opus 4.8
// ($5/$25 per MTok) while Opus 4.8 has moved to the vendor's legacy model list,
// so the alias is advanced with no cost delta.
// Used by launcher.go to route the model and by profile translations.
const ModelIDOpus5 = "claude-opus-5"

// ModelIDOpus48 is the superseded canonical model ID for Claude Opus 4.8, now
// replaced by ModelIDOpus5. Retained as a named constant because historical
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
	"opus":     ModelIDOpus5,
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
	// opus[1m] is retained for back-compat even though Opus 5 is natively 1M:
	// the [1m] suffix remains a valid Claude Code context-window modifier and
	// historical prefs files may carry it (kept rather than dropped).
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

// Performance tier tokens — the canonical {high, medium, low} vocabulary of the
// --model-policy CLI flag and the legacy performance_tier axis (the read-time
// alias source for llm.profile). Named constants per CLAUDE.local.md §14.
//
// Since the top column was renamed max -> high, these tokens are now identical
// to both the ModelPolicy vocabulary above and config.ValidProfiles(); the three
// axes no longer disagree, so MapModelPolicyToTier is an identity.
const (
	// PerformanceTierHigh is the highest-quality tier column.
	PerformanceTierHigh = "high"
	// PerformanceTierMedium is the balanced default tier column.
	PerformanceTierMedium = "medium"
	// PerformanceTierLow is the economical tier column.
	PerformanceTierLow = "low"
	// LegacyPerformanceTierMax is the superseded name of the top tier, accepted
	// as a read-time alias and never written back.
	LegacyPerformanceTierMax = "max"
)

// performanceTierRegex matches the performance_tier: line in llm.yaml.
var performanceTierRegex = regexp.MustCompile(`(?m)^(\s*)performance_tier:\s*["']?[\w-]*["']?`)

// ValidPerformanceTiers returns the closed set of valid No-Haiku performance
// tiers. The legacy max alias is readable but never offered.
func ValidPerformanceTiers() []string {
	return []string{PerformanceTierHigh, PerformanceTierMedium, PerformanceTierLow}
}

// IsValidPerformanceTier checks if the given string is a valid performance tier,
// accepting the superseded max alias for the top tier.
func IsValidPerformanceTier(s string) bool {
	return slices.Contains(ValidPerformanceTiers(), config.NormalizeProfile(s))
}

// ApplyPerformanceTier patches the performance_tier field in llm.yaml under
// the given project root. It reads .moai/config/sections/llm.yaml, replaces the
// performance_tier: line with the new tier value, and writes the file back.
// Returns nil if the file is absent (graceful no-op). The tier MUST be
// validated by the caller.
//
// @MX:ANCHOR: [AUTO] ApplyPerformanceTier — performance_tier persistence entry point
// @MX:REASON: fan_in >= 2 (init.go, web save)
func ApplyPerformanceTier(projectRoot, tier string) error {
	// The superseded top-tier name is readable but never written back, so the two
	// persisted axes (profile + performance_tier) cannot disagree.
	tier = config.NormalizeProfile(tier)
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
// inherit (user-added agents with no group membership) are never injected —
// inherit is never written as a model: value. The built-in Explore now has an
// explicit explore group cell (sonnet/low) and is no longer an inherit agent.
const modelInherit = "inherit"

// MapModelPolicyToTier translates a template.ModelPolicy value
// ({high, medium, low}) to the canonical performance tier ({high, medium, low}).
// Since the top tier was renamed max -> high the mapping is now an IDENTITY on
// every member; the function is retained as the named projection site so call
// sites keep a single place to reason about the two axes. An empty or
// unrecognized policy falls back to the medium tier (default-when-absent).
func MapModelPolicyToTier(policy ModelPolicy) string {
	switch policy {
	case ModelPolicyHigh:
		return PerformanceTierHigh
	case ModelPolicyMedium:
		return PerformanceTierMedium
	case ModelPolicyLow:
		return PerformanceTierLow
	default:
		return PerformanceTierMedium
	}
}

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

// NormalizeToTier resolves any performance-policy string to the canonical tier
// vocabulary {high, medium, low}. Canonical tokens pass through; the superseded
// top-tier name "max" is folded to "high" via config.NormalizeProfile; anything
// else is bridged through MapModelPolicyToTier, which falls back to the medium
// tier. This is the call-site resolver used by the init/update apply paths and
// llm.profile persistence.
func NormalizeToTier(s string) string {
	if IsValidPerformanceTier(s) {
		return config.NormalizeProfile(s)
	}
	return MapModelPolicyToTier(ModelPolicy(s))
}

// performanceTierValueRegex captures the persisted performance_tier value
// (group 1) from llm.yaml, requiring a non-empty value (an empty
// `performance_tier: ""` falls through to the medium default).
var performanceTierValueRegex = regexp.MustCompile(`(?m)^\s*performance_tier:\s*["']?([\w-]+)["']?`)

// ResolveProjectPerformanceTier reads the persisted performance_tier from the
// project's llm.yaml (the legacy alias axis). An absent file, an absent/commented
// performance_tier key, or an empty value resolves to the medium default. Any
// explicit value passes through verbatim; NormalizeToTier at the apply site
// bridges legacy vocabularies.
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
