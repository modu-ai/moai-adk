package config

// model_routing.go — typed loader + accessor for the model_routing block
// (SPEC-TOKEN-ROUTING-001, Token-Economy Epic B).
//
// This file owns the per-spawn COST axis: Tier x Phase -> {model, effort}.
// It is ORTHOGONAL to Phase 0.95's Mode 1-6 shape axis — this file decides
// model/effort, Phase 0.95 decides spawn shape. The two compose; they never
// compete (REQ-TR-011).

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Closed-set enums for model_routing validation (REQ-TR-004, REQ-TR-007).
//
// The effort set matches the canonical workflow_agents effort enum per
// SPEC-V3R6-WORKFLOW-EFFORT-MAP-001 REQ-WEM-006, so the two maps' effort
// sets cannot drift apart (REQ-TR-007). The model set extends the live
// workflow_agents values with "glm" for cost-routing purposes (REQ-TR-012).
var (
	validRoutingModels = map[string]bool{
		"inherit": true,
		"haiku":   true,
		"sonnet":  true,
		"opus":    true,
		"glm":     true,
	}
	validRoutingEfforts = map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"xhigh":  true,
		"max":    true,
	}
	validRoutingTiers = map[string]bool{
		"S": true,
		"M": true,
		"L": true,
	}
	validRoutingPhases = map[string]bool{
		"plan": true,
		"run":  true,
		"sync": true,
		"mx":   true,
	}
	// validRoutingPerfTiers is the closed set of performance tiers for the
	// No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001 M3, design.md §D.2).
	validRoutingPerfTiers = map[string]bool{
		"max":    true,
		"medium": true,
		"low":    true,
	}
)

// defaultRoutingEntry is the documented fallback returned when a (Tier, Phase)
// pair is absent from the declared model_routing map (REQ-TR-002). The caller
// observes the fallback via FallbackApplied=true on the returned entry.
var defaultRoutingEntry = ModelRoutingEntry{
	Model:           "inherit",
	Effort:          "medium",
	FallbackApplied: true,
}

// routingKey returns the composite map key "<TIER>-<phase>".
func routingKey(tier, phase string) string {
	return tier + "-" + phase
}

// RouteModelFor returns the routing entry for the given (Tier, Phase) pair,
// reading from the Workflow.ModelRouting map. Behaviors:
//   - Happy path: the pair is present and valid -> returned verbatim with
//     FallbackApplied=false (REQ-TR-003).
//   - Absent pair: the map is nil or lacks the key -> the documented default
//     entry is returned with FallbackApplied=true (REQ-TR-002). This never
//     fails; absence is a fallback, not an error.
//   - Invalid tier/phase INPUT (tier not in {S,M,L} or phase not in
//     {plan,run,sync,mx}): returns an error. The caller's input is wrong,
//     which is distinct from a map that merely omits a valid pair (REQ-TR-004
//     input-validation edge case, acceptance.md §D.2).
//
// Closed-set VALIDATION of the loaded model/effort values happens at
// ValidateModelRouting time (called by the loader after unmarshal), NOT here.
// RouteModelFor trusts the map was validated on load; callers that construct a
// Config without going through the loader SHOULD call ValidateModelRouting
// first.
//
// @MX:ANCHOR: [AUTO] RouteModelFor is the spawn-time cost-routing accessor
// @MX:REASON: fan_in >= 3 expected (orchestrator spawn call-site, tests, Epic C/D consumers)
//
// RouteModelFor returns the routing entry for the given (specTier, phase,
// perfTier) triple, reading from Workflow.ModelRoutingProfiles (the No-Haiku
// 3-tier map, SPEC-AGENT-ARCH-V2-001 M3, design.md §D.5). Behaviors mirror the
// flat-map accessor: happy path returns the entry with FallbackApplied=false;
// an absent perfTier or (specTier, phase) pair returns the documented default
// entry with FallbackApplied=true; invalid specTier/phase/perfTier INPUT
// returns an error. Closed-set VALIDATION of loaded model/effort values
// happens at ValidateModelRoutingProfiles time (loader), NOT here.
func (c *Config) RouteModelFor(specTier, phase, perfTier string) (ModelRoutingEntry, error) {
	if !validRoutingTiers[specTier] {
		return ModelRoutingEntry{}, &ValidationError{
			Field:   "tier",
			Message: fmt.Sprintf("tier %q not in closed set {S, M, L}", specTier),
			Value:   specTier,
		}
	}
	if !validRoutingPhases[phase] {
		return ModelRoutingEntry{}, &ValidationError{
			Field:   "phase",
			Message: fmt.Sprintf("phase %q not in closed set {plan, run, sync, mx}", phase),
			Value:   phase,
		}
	}
	if !validRoutingPerfTiers[perfTier] {
		return ModelRoutingEntry{}, &ValidationError{
			Field:   "perfTier",
			Message: fmt.Sprintf("perfTier %q not in closed set {max, medium, low}", perfTier),
			Value:   perfTier,
		}
	}

	if c == nil || c.Workflow.ModelRoutingProfiles == nil {
		return defaultRoutingEntry, nil
	}

	tierPhaseMap, ok := c.Workflow.ModelRoutingProfiles[perfTier]
	if !ok {
		return defaultRoutingEntry, nil
	}

	entry, ok := tierPhaseMap[routingKey(specTier, phase)]
	if !ok {
		return defaultRoutingEntry, nil
	}

	// A present entry overrides the default; clear any stale fallback flag the
	// YAML may have carried so the indicator reflects the runtime decision.
	entry.FallbackApplied = false
	return entry, nil
}

// ValidateModelRoutingProfiles checks every entry in the
// Workflow.ModelRoutingProfiles map against the closed sets (perfTier, key
// structure, model, effort). Returns a ValidationError naming the first
// offending location. A nil/empty map is valid (every lookup will fall back).
func (c *Config) ValidateModelRoutingProfiles() error {
	if c == nil || len(c.Workflow.ModelRoutingProfiles) == 0 {
		return nil
	}

	perfTiers := make([]string, 0, len(c.Workflow.ModelRoutingProfiles))
	for pt := range c.Workflow.ModelRoutingProfiles {
		perfTiers = append(perfTiers, pt)
	}
	sort.Strings(perfTiers)

	for _, pt := range perfTiers {
		if !validRoutingPerfTiers[pt] {
			return &ValidationError{
				Field:   fmt.Sprintf("model_routing_profiles.%s", pt),
				Message: fmt.Sprintf("perfTier %q not in closed set {max, medium, low}", pt),
				Value:   pt,
			}
		}
		tierPhaseMap := c.Workflow.ModelRoutingProfiles[pt]

		keys := make([]string, 0, len(tierPhaseMap))
		for k := range tierPhaseMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			entry := tierPhaseMap[k]

			parts := strings.SplitN(k, "-", 2)
			if len(parts) != 2 || !validRoutingTiers[parts[0]] || !validRoutingPhases[parts[1]] {
				return &ValidationError{
					Field:   fmt.Sprintf("model_routing_profiles.%s.%s", pt, k),
					Message: "key is not <TIER>-<phase> with Tier in {S,M,L} and Phase in {plan,run,sync,mx}",
					Value:   k,
				}
			}

			if !validRoutingModels[entry.Model] {
				return &ValidationError{
					Field:   fmt.Sprintf("model_routing_profiles.%s.%s.model", pt, k),
					Message: fmt.Sprintf("model %q not in closed set {inherit, haiku, sonnet, opus, glm}", entry.Model),
					Value:   entry.Model,
				}
			}
			if !validRoutingEfforts[entry.Effort] {
				return &ValidationError{
					Field:   fmt.Sprintf("model_routing_profiles.%s.%s.effort", pt, k),
					Message: fmt.Sprintf("effort %q not in closed set {low, medium, high, xhigh, max}", entry.Effort),
					Value:   entry.Effort,
				}
			}
		}
	}
	return nil
}

// ValidateModelRouting checks every entry in the Workflow.ModelRouting map
// against the closed sets (REQ-TR-004, REQ-TR-007). Returns a ValidationError
// naming the first offending key + field. A nil/empty map is valid (every
// lookup will fall back).
func (c *Config) ValidateModelRouting() error {
	if c == nil || len(c.Workflow.ModelRouting) == 0 {
		return nil
	}

	// Deterministic iteration order for stable error messages.
	keys := make([]string, 0, len(c.Workflow.ModelRouting))
	for k := range c.Workflow.ModelRouting {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		entry := c.Workflow.ModelRouting[k]

		// Validate the key structure "<TIER>-<phase>".
		parts := strings.SplitN(k, "-", 2)
		if len(parts) != 2 || !validRoutingTiers[parts[0]] || !validRoutingPhases[parts[1]] {
			return &ValidationError{
				Field:   fmt.Sprintf("model_routing.%s", k),
				Message: "key is not <TIER>-<phase> with Tier in {S,M,L} and Phase in {plan,run,sync,mx}",
				Value:   k,
			}
		}

		if !validRoutingModels[entry.Model] {
			return &ValidationError{
				Field:   fmt.Sprintf("model_routing.%s.model", k),
				Message: fmt.Sprintf("model %q not in closed set {inherit, haiku, sonnet, opus, glm}", entry.Model),
				Value:   entry.Model,
			}
		}
		if !validRoutingEfforts[entry.Effort] {
			return &ValidationError{
				Field:   fmt.Sprintf("model_routing.%s.effort", k),
				Message: fmt.Sprintf("effort %q not in closed set {low, medium, high, xhigh, max}", entry.Effort),
				Value:   entry.Effort,
			}
		}
	}
	return nil
}

// ValidateModelRoutingFromYAML parses a raw workflow.yaml byte slice and
// validates the model_routing block against the closed sets. Used by tests
// that need to validate fixture YAML without constructing a full Config.
// Returns nil if the block is absent or empty (valid — all lookups fall back).
func ValidateModelRoutingFromYAML(data []byte) error {
	var wrapper struct {
		Workflow struct {
			ModelRouting map[string]ModelRoutingEntry `yaml:"model_routing"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("parse workflow.yaml: %w", err)
	}

	cfg := &Config{}
	cfg.Workflow.ModelRouting = wrapper.Workflow.ModelRouting
	return cfg.ValidateModelRouting()
}

// LoadModelRouting reads workflow.yaml at the given path and returns just the
// model_routing map (unvalidated — callers may call ValidateModelRouting on a
// constructed Config, or ValidateModelRoutingFromYAML on the raw bytes).
// Returns nil (no error) when the file or block is absent — absence is a
// fallback, not a failure.
func LoadModelRouting(path string) (map[string]ModelRoutingEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var wrapper struct {
		Workflow struct {
			ModelRouting map[string]ModelRoutingEntry `yaml:"model_routing"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if wrapper.Workflow.ModelRouting == nil {
		slog.Debug("model_routing block absent; all lookups will fall back", "path", path)
	}
	return wrapper.Workflow.ModelRouting, nil
}
