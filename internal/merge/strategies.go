package merge

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// textExtensions are file extensions that support line-based merge.
var textExtensions = map[string]bool{
	".md":   true,
	".txt":  true,
	".toml": true,
	".cfg":  true,
	".ini":  true,
	".sh":   true,
	".py":   true,
	".go":   true,
	".js":   true,
	".ts":   true,
	".css":  true,
	".html": true,
	".xml":  true,
	".rs":   true,
	".rb":   true,
	".java": true,
}

// binaryExtensions are file extensions treated as binary (overwrite only).
var binaryExtensions = map[string]bool{
	".bin":   true,
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".ico":   true,
	".zip":   true,
	".tar":   true,
	".gz":    true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".eot":   true,
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".pdf":   true,
	".mp3":   true,
	".mp4":   true,
}

// strategySelector implements StrategySelector.
type strategySelector struct{}

// NewStrategySelector creates a new StrategySelector instance.
func NewStrategySelector() StrategySelector {
	return &strategySelector{}
}

// SelectStrategy returns the appropriate merge strategy for a file path.
func (s *strategySelector) SelectStrategy(path string) MergeStrategy {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	// Exact filename matches first.
	switch base {
	case "CLAUDE.md":
		return SectionMerge
	case ".gitignore":
		return EntryMerge
	}

	// Extension-based matching.
	switch ext {
	case ".yaml", ".yml":
		return YAMLDeep
	case ".json":
		return JSONMerge
	}

	// Known text file extensions -> LineMerge.
	if textExtensions[ext] {
		return LineMerge
	}

	// Known binary extensions -> Overwrite.
	if binaryExtensions[ext] {
		return Overwrite
	}

	// Unknown extension: default to Overwrite for safety.
	return Overwrite
}

// mergeLineBased performs a line-by-line 3-way merge.
// It detects changes between base-current and base-updated, then combines them.
func mergeLineBased(base, current, updated []byte) (*MergeResult, error) {
	baseLines := splitLines(string(base))
	currentLines := splitLines(string(current))
	updatedLines := splitLines(string(updated))

	// Compute diffs: base->current (user changes) and base->updated (template changes).
	currentEdits := computeLineChanges(baseLines, currentLines)
	updatedEdits := computeLineChanges(baseLines, updatedLines)

	// Build merged result by walking through base lines.
	var merged []string
	var conflicts []Conflict

	n := len(baseLines)
	// Track which base lines have changes on each side.
	// For simplicity, use a per-line approach.

	ci := 0 // index into currentLines
	ui := 0 // index into updatedLines
	bi := 0 // index into baseLines

	for bi < n || ci < len(currentLines) || ui < len(updatedLines) {
		if bi >= n {
			// Past base: append remaining from whichever side has additions.
			for ci < len(currentLines) {
				if ui < len(updatedLines) {
					if currentLines[ci] == updatedLines[ui] {
						merged = append(merged, currentLines[ci])
						ci++
						ui++
					} else {
						// Both added different lines at the end.
						merged = append(merged, currentLines[ci])
						ci++
					}
				} else {
					merged = append(merged, currentLines[ci])
					ci++
				}
			}
			for ui < len(updatedLines) {
				merged = append(merged, updatedLines[ui])
				ui++
			}
			break
		}

		baseLine := baseLines[bi]
		currentChanged := currentEdits[bi]
		updatedChanged := updatedEdits[bi]

		switch {
		case !currentChanged && !updatedChanged:
			// Neither side changed this line.
			merged = append(merged, baseLine)
			bi++
			ci++
			ui++

		case currentChanged && !updatedChanged:
			// Only current (user) changed this line.
			if ci < len(currentLines) {
				merged = append(merged, currentLines[ci])
				ci++
			}
			// Skip the base line.
			bi++
			ui++

		case !currentChanged && updatedChanged:
			// Only updated (template) changed this line.
			if ui < len(updatedLines) {
				merged = append(merged, updatedLines[ui])
				ui++
			}
			bi++
			ci++

		default:
			// Both sides changed the same line.
			var curText, updText string
			if ci < len(currentLines) {
				curText = currentLines[ci]
			}
			if ui < len(updatedLines) {
				updText = updatedLines[ui]
			}

			if curText == updText {
				// Both changed to the same thing - no conflict.
				merged = append(merged, curText)
			} else {
				// Conflict.
				conflicts = append(conflicts, Conflict{
					StartLine: bi + 1,
					EndLine:   bi + 1,
					Base:      baseLine,
					Current:   curText,
					Updated:   updText,
				})
				// Include current version in output with conflict markers.
				merged = append(merged, curText)
			}
			bi++
			ci++
			ui++
		}
	}

	result := &MergeResult{
		Content:     []byte(strings.Join(merged, "\n")),
		HasConflict: len(conflicts) > 0,
		Conflicts:   conflicts,
		Strategy:    LineMerge,
	}

	return result, nil
}

// computeLineChanges compares base and modified, returning a map of base line
// indices that were changed (deleted or replaced).
func computeLineChanges(base, modified []string) map[int]bool {
	changes := make(map[int]bool)
	edits := DiffLines(base, modified)

	for _, e := range edits {
		if e.Op == OpDelete {
			changes[e.OldLine] = true
		}
	}

	return changes
}

// mergeEntryBased performs entry-based merge suitable for .gitignore-style files.
// Each line is treated as an independent entry. User additions are preserved,
// user deletions are respected, and new template entries are added.
func mergeEntryBased(base, current, updated []byte) (*MergeResult, error) {
	baseEntries := splitLines(string(base))
	currentEntries := splitLines(string(current))
	updatedEntries := splitLines(string(updated))

	baseSet := toSet(baseEntries)
	currentSet := toSet(currentEntries)

	// Entries user deleted (in base but not in current).
	userDeleted := make(map[string]bool)
	for _, e := range baseEntries {
		if !currentSet[e] {
			userDeleted[e] = true
		}
	}

	// Start with current entries (preserves user additions and order).
	seen := make(map[string]bool)
	var result []string
	for _, e := range currentEntries {
		if e == "" {
			continue
		}
		if !seen[e] {
			result = append(result, e)
			seen[e] = true
		}
	}

	// Add new template entries that are not already present
	// and were not deliberately deleted by the user.
	for _, e := range updatedEntries {
		if e == "" {
			continue
		}
		if !seen[e] && !userDeleted[e] {
			// Only add if it's genuinely new (not in base) or still in base.
			if !baseSet[e] {
				result = append(result, e)
				seen[e] = true
			}
		}
	}

	return &MergeResult{
		Content:     []byte(strings.Join(result, "\n")),
		HasConflict: false,
		Conflicts:   nil,
		Strategy:    EntryMerge,
	}, nil
}

// mergeOverwrite replaces the content entirely with the updated version.
func mergeOverwrite(current, updated []byte) (*MergeResult, error) {
	return &MergeResult{
		Content:     updated,
		HasConflict: false,
		Conflicts:   nil,
		Strategy:    Overwrite,
	}, nil
}

// mergeJSON performs a JSON object-level 3-way merge.
func mergeJSON(base, current, updated []byte) (*MergeResult, error) {
	var baseMap, currentMap, updatedMap map[string]any

	if err := json.Unmarshal(base, &baseMap); err != nil {
		return nil, fmt.Errorf("merge json: parse base: %w", err)
	}
	if err := json.Unmarshal(current, &currentMap); err != nil {
		return nil, fmt.Errorf("merge json: parse current: %w", err)
	}
	if err := json.Unmarshal(updated, &updatedMap); err != nil {
		return nil, fmt.Errorf("merge json: parse updated: %w", err)
	}

	merged, conflicts := deepMergeMap(baseMap, currentMap, updatedMap, "")

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("merge json: marshal result: %w", err)
	}

	return &MergeResult{
		Content:     data,
		HasConflict: len(conflicts) > 0,
		Conflicts:   conflicts,
		Strategy:    JSONMerge,
	}, nil
}

// mergeYAML performs a YAML structure-preserving deep merge.
func mergeYAML(base, current, updated []byte) (*MergeResult, error) {
	var baseMap, currentMap, updatedMap map[string]any

	if err := yaml.Unmarshal(base, &baseMap); err != nil {
		return nil, fmt.Errorf("merge yaml: parse base: %w", err)
	}
	if err := yaml.Unmarshal(current, &currentMap); err != nil {
		return nil, fmt.Errorf("merge yaml: parse current: %w", err)
	}
	if err := yaml.Unmarshal(updated, &updatedMap); err != nil {
		return nil, fmt.Errorf("merge yaml: parse updated: %w", err)
	}

	merged, conflicts := deepMergeMap(baseMap, currentMap, updatedMap, "")

	data, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("merge yaml: marshal result: %w", err)
	}

	return &MergeResult{
		Content:     data,
		HasConflict: len(conflicts) > 0,
		Conflicts:   conflicts,
		Strategy:    YAMLDeep,
	}, nil
}

// deepMergeMap performs a recursive 3-way merge on map structures.
// It returns the merged map and any conflicts detected.
func deepMergeMap(base, current, updated map[string]any, prefix string) (map[string]any, []Conflict) {
	result := make(map[string]any)
	var conflicts []Conflict

	// Collect all keys from all three maps.
	allKeys := make(map[string]bool)
	for k := range base {
		allKeys[k] = true
	}
	for k := range current {
		allKeys[k] = true
	}
	for k := range updated {
		allKeys[k] = true
	}

	for key := range allKeys {
		keyPath := key
		if prefix != "" {
			keyPath = prefix + "." + key
		}

		baseVal, inBase := base[key]
		curVal, inCurrent := current[key]
		updVal, inUpdated := updated[key]

		switch {
		case !inBase && inCurrent && !inUpdated:
			// User added key - preserve.
			result[key] = curVal

		case !inBase && !inCurrent && inUpdated:
			// Template added key - add.
			result[key] = updVal

		case !inBase && inCurrent && inUpdated:
			// Both added - check if same.
			if valuesEqual(curVal, updVal) {
				result[key] = curVal
			} else {
				conflicts = append(conflicts, Conflict{
					StartLine: 0,
					EndLine:   0,
					Base:      "",
					Current:   fmt.Sprintf("%v", curVal),
					Updated:   fmt.Sprintf("%v", updVal),
				})
				result[key] = curVal
			}

		case inBase && !inCurrent && inUpdated:
			// User deleted key - respect user deletion.
			// Don't include.

		case inBase && inCurrent && !inUpdated:
			// Template removed key - keep user's version.
			result[key] = curVal

		case inBase && inCurrent && inUpdated:
			baseChanged := !valuesEqual(baseVal, curVal)
			updChanged := !valuesEqual(baseVal, updVal)

			switch {
			case !baseChanged && !updChanged:
				// No changes.
				result[key] = baseVal

			case baseChanged && !updChanged:
				// Only user changed.
				result[key] = curVal

			case !baseChanged && updChanged:
				// Only template changed.
				result[key] = updVal

			default:
				// Both changed.
				if valuesEqual(curVal, updVal) {
					// Same change.
					result[key] = curVal
				} else {
					// Check if both are maps for recursive merge.
					curMap, curIsMap := toMapInterface(curVal)
					updMap, updIsMap := toMapInterface(updVal)
					baseMap, baseIsMap := toMapInterface(baseVal)

					if curIsMap && updIsMap && baseIsMap {
						subResult, subConflicts := deepMergeMap(baseMap, curMap, updMap, keyPath)
						result[key] = subResult
						conflicts = append(conflicts, subConflicts...)
					} else {
						// Conflict.
						conflicts = append(conflicts, Conflict{
							StartLine: 0,
							EndLine:   0,
							Base:      fmt.Sprintf("%v", baseVal),
							Current:   fmt.Sprintf("%v", curVal),
							Updated:   fmt.Sprintf("%v", updVal),
						})
						result[key] = curVal
					}
				}
			}
		}
	}

	return result, conflicts
}

// mergeSectionBased performs section-based merge for CLAUDE.md files.
// Sections are delimited by Markdown headings (## or ###).
func mergeSectionBased(base, current, updated []byte) (*MergeResult, error) {
	baseSections := parseSections(string(base))
	currentSections := parseSections(string(current))
	updatedSections := parseSections(string(updated))

	var conflicts []Conflict

	// Build maps for lookup.
	baseMap := sectionMap(baseSections)
	currentMap := sectionMap(currentSections)

	// Start with the updated template order.
	var resultParts []string
	usedSections := make(map[string]bool)

	for _, sec := range updatedSections {
		usedSections[sec.heading] = true

		baseSec, inBase := baseMap[sec.heading]
		curSec, inCurrent := currentMap[sec.heading]

		if !inBase {
			// New section from template.
			resultParts = append(resultParts, sec.heading+"\n"+sec.content)
		} else if !inCurrent {
			// Section removed by user - keep template version.
			resultParts = append(resultParts, sec.heading+"\n"+sec.content)
		} else {
			// Section exists in all three - check for changes.
			baseContent := baseSec.content
			curContent := curSec.content
			updContent := sec.content

			baseChanged := curContent != baseContent
			updChanged := updContent != baseContent

			switch {
			case !baseChanged && !updChanged:
				resultParts = append(resultParts, sec.heading+"\n"+baseContent)
			case baseChanged && !updChanged:
				resultParts = append(resultParts, sec.heading+"\n"+curContent)
			case !baseChanged && updChanged:
				resultParts = append(resultParts, sec.heading+"\n"+updContent)
			default:
				if curContent == updContent {
					resultParts = append(resultParts, sec.heading+"\n"+curContent)
				} else {
					conflicts = append(conflicts, Conflict{
						StartLine: 0,
						EndLine:   0,
						Base:      baseContent,
						Current:   curContent,
						Updated:   updContent,
					})
					resultParts = append(resultParts, sec.heading+"\n"+curContent)
				}
			}
		}
	}

	// Append user-added sections not in updated template.
	for _, sec := range currentSections {
		if !usedSections[sec.heading] {
			resultParts = append(resultParts, sec.heading+"\n"+sec.content)
		}
	}

	return &MergeResult{
		Content:     []byte(strings.Join(resultParts, "\n")),
		HasConflict: len(conflicts) > 0,
		Conflicts:   conflicts,
		Strategy:    SectionMerge,
	}, nil
}

// section represents a Markdown section with heading and content.
type section struct {
	heading string
	content string
}

// parseSections splits Markdown content into sections based on ## headings.
func parseSections(content string) []section {
	lines := strings.Split(content, "\n")
	var sections []section
	var currentHeading string
	var currentContent []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
			// Save previous section.
			if currentHeading != "" {
				sections = append(sections, section{
					heading: currentHeading,
					content: strings.Join(currentContent, "\n"),
				})
			}
			currentHeading = line
			currentContent = nil
		} else {
			if currentHeading != "" {
				currentContent = append(currentContent, line)
			}
			// Content before any heading is ignored for section merge.
		}
	}

	// Save last section.
	if currentHeading != "" {
		sections = append(sections, section{
			heading: currentHeading,
			content: strings.Join(currentContent, "\n"),
		})
	}

	return sections
}

// sectionMap creates a map from heading to section for quick lookup.
func sectionMap(sections []section) map[string]section {
	m := make(map[string]section)
	for _, s := range sections {
		m[s.heading] = s
	}
	return m
}

// toSet converts a slice of strings to a set (map[string]bool).
func toSet(items []string) map[string]bool {
	s := make(map[string]bool)
	for _, item := range items {
		if item != "" {
			s[item] = true
		}
	}
	return s
}

// valuesEqual compares two any values for equality.
func valuesEqual(a, b any) bool {
	aj, errA := json.Marshal(a)
	bj, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(aj) == string(bj)
}

// toMapInterface attempts to convert an any to map[string]any.
func toMapInterface(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		// YAML unmarshals maps with any keys.
		result := make(map[string]any)
		for k, val := range m {
			result[fmt.Sprintf("%v", k)] = val
		}
		return result, true
	}
	return nil, false
}
