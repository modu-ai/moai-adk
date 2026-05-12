package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// @MX:NOTE: [AUTO] MergedSettings implements 8-tier priority merge system with provenance tracking
//
// MergedSettings holds the merged configuration from all 8 tiers.
// Values are keyed by "section.field" format.
type MergedSettings struct {
	values map[string]Value[any]
}

// NewMergedSettings creates an empty MergedSettings instance.
func NewMergedSettings() *MergedSettings {
	return &MergedSettings{
		values: make(map[string]Value[any]),
	}
}

// Get retrieves a value by key.
// Returns the value and true if the key exists, zero value and false otherwise.
func (m *MergedSettings) Get(key string) (Value[any], bool) {
	v, ok := m.values[key]
	return v, ok
}

// Set stores a value with its key.
func (m *MergedSettings) Set(key string, v Value[any]) {
	m.values[key] = v
}

// Keys returns all keys in the merged settings.
func (m *MergedSettings) Keys() []string {
	keys := make([]string, 0, len(m.values))
	for k := range m.values {
		keys = append(keys, k)
	}
	return keys
}

// All returns the complete map of keys to values.
func (m *MergedSettings) All() map[string]Value[any] {
	return m.values
}

// MergeAll walks all tiers in priority order and produces merged settings.
// For each (section, field) key, the first non-zero value wins.
// Provenance is captured from the winning tier, including SchemaVersion when
// loadYAMLFile embedded the schemaVersionKey sentinel.
//
// When policy.strict_mode is true in the SrcPolicy tier, any lower tier providing
// a non-zero value for a policy-designated key causes PolicyOverrideRejected to be returned.
//
// This maps to REQ-V3R2-RT-005-005, REQ-V3R2-RT-005-010, REQ-V3R2-RT-005-022, AC-01, AC-07.
//
// @MX:NOTE: [AUTO] SPEC-V3R2-RT-005 — deterministic 8-tier merge. Identical tier inputs MUST produce byte-identical merged output (cache-prefix discipline; problem-catalog P-C05). Map iteration order is non-deterministic; sort keys before serialization (see merge.go:dumpJSON).
//
// Parameters:
//   - tiers: map of source to raw configuration values (map[string]any where key is "section.field")
//   - origins: map of source to the file path where those values came from
//   - loadedAt: timestamp to use for all provenance records
//
// Returns a MergedSettings with all values merged, or an error if merging fails.
func MergeAll(
	tiers map[Source]map[string]any,
	origins map[Source]string,
	loadedAt time.Time,
) (*MergedSettings, error) {
	result := NewMergedSettings()

	// Extract per-tier schema versions from the sentinel key.
	// These are not config keys and must never appear in the merged output.
	tierSchemaVersions := make(map[Source]int)
	for source, tierData := range tiers {
		if sv, ok := tierData[schemaVersionSentinel]; ok {
			if svInt, ok := sv.(int); ok {
				tierSchemaVersions[source] = svInt
			}
		}
	}

	// @MX:WARN: [AUTO] SPEC-V3R2-RT-005 REQ-022 — policy override enforcement.
	// @MX:REASON: Disabling this check (e.g., commenting out the strict_mode enforcement) defeats the constitutional governance contract. Enterprise rollouts depend on PolicyOverrideRejected to enforce policy precedence.

	// Determine strict_mode from the SrcPolicy tier.
	// policy.strict_mode: true means lower tiers cannot override policy-designated keys.
	// REQ-V3R2-RT-005-022, AC-07.
	strictMode := false
	if policyTier, ok := tiers[SrcPolicy]; ok {
		if sm, ok := policyTier["policy.strict_mode"]; ok {
			if smBool, ok := sm.(bool); ok {
				strictMode = smBool
			}
		}
	}

	// Collect policy-designated keys (all non-sentinel keys present in SrcPolicy tier).
	// Only relevant when strictMode is true.
	policyKeys := make(map[string]bool)
	if strictMode {
		if policyTier, ok := tiers[SrcPolicy]; ok {
			for key := range policyTier {
				if key != schemaVersionSentinel {
					policyKeys[key] = true
				}
			}
		}
	}

	// Collect all keys across all tiers, excluding the sentinel.
	allKeys := make(map[string]bool)
	for _, tierData := range tiers {
		for key := range tierData {
			if key == schemaVersionSentinel {
				continue
			}
			allKeys[key] = true
		}
	}

	// For each key, walk tiers in priority order and take the first non-zero value.
	for key := range allKeys {
		var winningValue any
		var winningSource Source
		var winningOrigin string
		var winningSchemaVersion int
		var overriddenBy []string

		// Walk tiers in priority order (SrcPolicy = 0 is highest)
		for _, source := range AllSources() {
			tierData, tierExists := tiers[source]
			if !tierExists {
				continue
			}

			value, valueExists := tierData[key]
			if !valueExists {
				continue
			}

			// Check if value is zero (use reflection to handle all types)
			if isZero(value) {
				continue
			}

			// If we already have a winning value, this tier is being overridden
			if winningValue != nil {
				// strict_mode enforcement: lower tier must not override policy-designated keys.
				// REQ-V3R2-RT-005-022: WHILE policy.strict_mode: true, override is rejected.
				if strictMode && policyKeys[key] && winningSource == SrcPolicy {
					return nil, &PolicyOverrideRejected{
						Key:             key,
						PolicySource:    origins[SrcPolicy],
						AttemptedSource: origins[source],
					}
				}
				origin := origins[source]
				if origin != "" {
					overriddenBy = append(overriddenBy, origin)
				}
				continue
			}

			// This is the winning value
			winningValue = value
			winningSource = source
			winningOrigin = origins[source]
			winningSchemaVersion = tierSchemaVersions[source]
		}

		// If we found a winning value, store it with provenance
		if winningValue != nil {
			provenance := Provenance{
				Source:        winningSource,
				Origin:        winningOrigin,
				Loaded:        loadedAt,
				SchemaVersion: winningSchemaVersion,
				OverriddenBy:  overriddenBy,
			}

			result.Set(key, Value[any]{
				V: winningValue,
				P: provenance,
			})
		}
	}

	return result, nil
}

// isZero uses reflection to check if a value is the zero value for its type.
// This handles all types including nil interfaces, empty strings, zero numbers, etc.
func isZero(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.String:
		return rv.String() == ""
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

// Diff compares two MergedSettings and returns the differences.
// Keys that are identical in both are omitted from the result.
//
// This maps to REQ-V3R2-RT-005-007 and REQ-V3R2-RT-005-051.
func Diff(a, b *MergedSettings) map[string]struct{ A, B Value[any] } {
	result := make(map[string]struct{ A, B Value[any] })

	// Collect all keys from both settings
	allKeys := make(map[string]bool)
	for _, k := range a.Keys() {
		allKeys[k] = true
	}
	for _, k := range b.Keys() {
		allKeys[k] = true
	}

	// Check each key for differences
	for key := range allKeys {
		valA, hasA := a.Get(key)
		valB, hasB := b.Get(key)

		// Key exists in both - check if values are equal
		if hasA && hasB {
			if !valuesEqual(valA.V, valB.V) {
				result[key] = struct{ A, B Value[any] }{A: valA, B: valB}
			}
			continue
		}

		// Key exists only in A
		if hasA && !hasB {
			result[key] = struct{ A, B Value[any] }{A: valA, B: Value[any]{}}
			continue
		}

		// Key exists only in B
		if !hasA && hasB {
			result[key] = struct{ A, B Value[any] }{A: Value[any]{}, B: valB}
		}
	}

	return result
}

// valuesEqual checks if two values are equal using reflection.
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	ra := reflect.ValueOf(a)
	rb := reflect.ValueOf(b)

	if ra.Type() != rb.Type() {
		return false
	}

	return reflect.DeepEqual(a, b)
}

// Dump writes the merged settings to a writer in a human-readable format.
// This maps to REQ-V3R2-RT-005-006.
func (m *MergedSettings) Dump(format string) (string, error) {
	switch format {
	case "json", "":
		return m.dumpJSON()
	case "yaml":
		return m.dumpYAML()
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// dumpJSON serializes merged settings to byte-stable JSON.
// Keys are sorted alphabetically via sort.Strings to guarantee deterministic output.
// Uses encoding/json.MarshalIndent with 2-space indent for readability.
//
// REQ-V3R2-RT-005-006, AC-02: byte-stable JSON with provenance structure.
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M3 GREEN — sort.Strings + MarshalIndent for determinism.
func (m *MergedSettings) dumpJSON() (string, error) {
	// Collect and sort keys for deterministic output (AC-02 byte-stability).
	keys := make([]string, 0, len(m.values))
	for k := range m.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build output in sorted key order using a slice of key-value pairs
	// to preserve insertion order in the JSON output.
	type jsonEntry struct {
		Value      any    `json:"value"`
		Source     string `json:"source"`
		Origin     string `json:"origin"`
		Loaded     string `json:"loaded"`
		Overridden bool   `json:"overridden"`
		Default    bool   `json:"default,omitempty"`
	}

	output := make(map[string]jsonEntry, len(keys))
	for _, key := range keys {
		value := m.values[key]
		output[key] = jsonEntry{
			Value:      value.V,
			Source:     value.P.Source.String(),
			Origin:     value.P.Origin,
			Loaded:     value.P.Loaded.Format("2006-01-02T15:04:05Z07:00"),
			Overridden: len(value.P.OverriddenBy) > 0,
			Default:    value.IsDefault(),
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("dumpJSON marshal: %w", err)
	}
	return string(data), nil
}

// dumpYAML serializes merged settings to YAML with source comments.
// Keys are sorted alphabetically for deterministic output.
// Each key is followed by a `# source: <tier>` comment as required by REQ-V3R2-RT-005-030.
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M3 GREEN — alphabetical sort + source comments.
func (m *MergedSettings) dumpYAML() (string, error) {
	// Collect and sort keys for deterministic output (AC-09).
	keys := make([]string, 0, len(m.values))
	for k := range m.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, key := range keys {
		value := m.values[key]
		fmt.Fprintf(&sb, "%s: %v # source: %s\n", key, value.V, value.P.Source.String())
	}

	return sb.String(), nil
}
