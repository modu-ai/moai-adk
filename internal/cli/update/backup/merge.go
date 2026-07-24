package backup

import (
	"fmt"
	"maps"

	"gopkg.in/yaml.v3"
)

// MergeYAML3Way performs a 3-way merge of YAML documents.
// It uses baseData (old template defaults) to detect user changes:
//   - If user value == base value: user didn't change it → use new template value
//   - If user value != base value: user customized it → preserve user value
//
// System fields (like template_version) always use new values regardless.
//
// Moved from internal/cli/update.go during M3d-A decomposition (SPEC-CLI-TUX-V3-003).
// Behavior is byte-identical to the pre-decomposition implementation; only the
// package location and export status changed.
func MergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error) {
	var newMap, oldMap, baseMap map[string]any

	if err := yaml.Unmarshal(newData, &newMap); err != nil {
		return nil, fmt.Errorf("unmarshal new YAML: %w", err)
	}
	if err := yaml.Unmarshal(oldData, &oldMap); err != nil {
		return nil, fmt.Errorf("unmarshal old YAML: %w", err)
	}
	if err := yaml.Unmarshal(baseData, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base YAML: %w", err)
	}

	merged := DeepMerge3Way(newMap, oldMap, baseMap)
	return yaml.Marshal(merged)
}

// DeepMerge3Way recursively performs 3-way merge of maps.
// Decision logic for each key:
//   - old == base → user didn't change → use new value
//   - old != base → user changed → preserve old value
//   - key only in new → new field added by template → use new value
//   - key only in old → removed from template → drop it
func DeepMerge3Way(newMap, oldMap, baseMap map[string]any) map[string]any {
	result := make(map[string]any)

	// System fields that always use new values
	systemFields := map[string]bool{
		"template_version": true,
		"version":          true,
	}

	// Start with all new values as the base result
	for k, newV := range newMap {
		// System fields always use new value
		if systemFields[k] {
			result[k] = newV
			continue
		}

		oldV, oldExists := oldMap[k]
		baseV, baseExists := baseMap[k]

		if !oldExists {
			// Key only in new template → add it (new field)
			result[k] = newV
			continue
		}

		// Both new and old exist
		newMapVal, newIsMap := newV.(map[string]any)
		oldMapVal, oldIsMap := oldV.(map[string]any)

		if newIsMap && oldIsMap {
			// Both are maps → recurse
			baseMapVal, baseIsMap := baseV.(map[string]any)
			if !baseIsMap {
				baseMapVal = make(map[string]any)
			}
			result[k] = DeepMerge3Way(newMapVal, oldMapVal, baseMapVal)
		} else {
			// Scalar or list values
			if !baseExists {
				// No base value → user added this; preserve user value
				result[k] = oldV
			} else if ValuesEqual(oldV, baseV) {
				// User didn't change from template default → use new template value
				result[k] = newV
			} else {
				// User changed from template default → preserve user value
				result[k] = oldV
			}
		}
	}

	// Keys only in old (not in new template) are dropped:
	// they were removed from the template, so we don't carry them forward.

	return result
}

// ValuesEqual compares two interface{} values for equality.
// Handles string, int, float, bool, and nil comparisons.
func ValuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// MergeYAMLDeep performs a deep merge of two YAML documents (2-way fallback).
// The newData takes precedence for structure, but values from oldData are preserved
// when the key exists in both.
func MergeYAMLDeep(newData, oldData []byte) ([]byte, error) {
	var newMap, oldMap map[string]any

	if err := yaml.Unmarshal(newData, &newMap); err != nil {
		return nil, fmt.Errorf("unmarshal new YAML: %w", err)
	}
	if err := yaml.Unmarshal(oldData, &oldMap); err != nil {
		return nil, fmt.Errorf("unmarshal old YAML: %w", err)
	}

	// Deep merge old values into new structure
	merged := DeepMergeMaps(newMap, oldMap)

	return yaml.Marshal(merged)
}

// DeepMergeMaps recursively merges oldMap into newMap, preserving old values.
// System fields (version, template_version) always use new values.
func DeepMergeMaps(newMap, oldMap map[string]any) map[string]any {
	result := make(map[string]any)

	// System fields that should always use new values (not preserved from old
	// config). MUST stay in parity with DeepMerge3Way's systemFields: when the
	// restore degrades to this 2-way fallback (template defaults unavailable),
	// a backup carrying an old moai.version must not override the deployed one.
	systemFields := map[string]bool{
		"template_version": true,
		"version":          true,
	}

	// Copy all new values
	maps.Copy(result, newMap)

	// Merge old values, preserving when they exist
	for k, v := range oldMap {
		// Skip system fields - always use new value
		if systemFields[k] {
			continue
		}

		if newV, exists := newMap[k]; exists {
			// Both exist, check if they are maps
			newMapVal, newIsMap := newV.(map[string]any)
			oldMapVal, oldIsMap := v.(map[string]any)

			if newIsMap && oldIsMap {
				// Recursively merge nested maps
				result[k] = DeepMergeMaps(newMapVal, oldMapVal)
			} else {
				// Keep old value (preserve user setting)
				result[k] = v
			}
		} else {
			// Only exists in old, add it
			result[k] = v
		}
	}

	return result
}
