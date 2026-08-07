package astx

import (
	"os"
	"sort"
)

// SpecIDsFromCapabilityMap reads 001's capability-map.md and returns the
// sorted-unique set of SPEC IDs from its `spec-id` column, WITHOUT walking
// implementation paths or running tree-sitter extraction. It reuses the
// existing unexported `parseCapabilityMap` parser (header-driven join, so the
// column may appear in any position).
//
// This is the lightweight alternative to `EnrichRows` for callers that need
// only the SPEC ID set (e.g. navigator-sync's `loadSpecsFromCapabilityMap`).
// `EnrichRows` runs the full enrichment pipeline — per-row
// `filepath.WalkDir` of the implementation-path and tree-sitter `Extract()`
// on every file — which is wasted work when only the spec-id column is
// consumed.
//
// Fail-open: an unreadable or absent capability-map returns an empty slice
// and a nil error, mirroring `EnrichRows` (REQ-NT-002).
func SpecIDsFromCapabilityMap(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		// Fail-open: absent/unreadable capability-map → empty result (REQ-NT-002).
		return nil, nil
	}
	rows := parseCapabilityMap(content)
	seen := map[string]bool{}
	var out []string
	for _, r := range rows {
		id := r["spec-id"]
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}
