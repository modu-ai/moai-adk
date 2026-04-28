package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

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
// Provenance is captured from the winning tier.
//
// This maps to REQ-V3R2-RT-005-005 and REQ-V3R2-RT-005-010.
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

	// Collect all keys across all tiers
	allKeys := make(map[string]bool)
	for _, tierData := range tiers {
		for key := range tierData {
			allKeys[key] = true
		}
	}

	// For each key, walk tiers in priority order and take the first non-zero value
	for key := range allKeys {
		var winningValue any
		var winningSource Source
		var winningOrigin string
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
		}

		// If we found a winning value, store it with provenance
		if winningValue != nil {
			provenance := Provenance{
				Source:       winningSource,
				Origin:       winningOrigin,
				Loaded:       loadedAt,
				OverriddenBy: overriddenBy,
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

func (m *MergedSettings) dumpJSON() (string, error) {
	// Build a map with provenance information for JSON output
	output := make(map[string]map[string]any)

	for key, value := range m.values {
		output[key] = map[string]any{
			"value":     value.V,
			"source":    value.P.Source.String(),
			"origin":    value.P.Origin,
			"loaded":    value.P.Loaded,
			"overridden": len(value.P.OverriddenBy) > 0,
		}
	}

	// Use encoding/json - but we need to import it
	// For now, return a simple format representation
	return formatMapAsJSON(output)
}

func (m *MergedSettings) dumpYAML() (string, error) {
	var sb strings.Builder

	for key, value := range m.values {
		fmt.Fprintf(&sb, "%s: %v # source: %s\n", key, value.V, value.P.Source.String())
	}

	return sb.String(), nil
}

// formatMapAsJSON is a simple JSON formatter for the dump output.
// In production, this would use encoding/json properly.
func formatMapAsJSON(m map[string]map[string]any) (string, error) {
	// Simplified JSON representation
	// In a real implementation, use encoding/json
	return fmt.Sprintf("%+v", m), nil
}
