package merge

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// ErrZoneNotFound is returned by ReplaceEvolvableZone when the specified zone
// ID does not exist in the content.
var ErrZoneNotFound = errors.New("merge: evolvable zone not found")

// EvolvableZone represents a single evolvable zone within a file.
// Evolvable zones are delimited by HTML comment markers that survive moai update.
type EvolvableZone struct {
	// ID is the unique identifier from the start marker attribute.
	ID string

	// StartLine is the 0-based index of the start marker line.
	StartLine int

	// EndLine is the 0-based index of the end marker line (exclusive: line after zone content).
	EndLine int

	// Content is the text content between the start and end markers (not including markers).
	Content string
}

// reEvolvableStart matches <!-- moai:evolvable-start id="..." -->
// Captures: (1) id attribute value
var reEvolvableStart = regexp.MustCompile(`<!--\s*moai:evolvable-start\s+id="([^"]+)"\s*-->`)

// reEvolvableEnd matches <!-- moai:evolvable-end -->
var reEvolvableEnd = regexp.MustCompile(`<!--\s*moai:evolvable-end\s*-->`)

// ParseEvolvableZones parses evolvable zone markers from file content.
//
// Rules:
//   - id attribute is mandatory and must be unique within the file.
//   - Nested markers are rejected (error).
//   - Missing end marker produces a warning and discards the unclosed zone.
//
// Returns the ordered list of zones found in the file.
func ParseEvolvableZones(content string) ([]EvolvableZone, error) {
	lines := strings.Split(content, "\n")

	var zones []EvolvableZone
	seenIDs := make(map[string]bool)

	type openZone struct {
		id        string
		startLine int // line index of start marker
	}
	var current *openZone

	for i, line := range lines {
		if reEvolvableStart.MatchString(line) {
			if current != nil {
				// Nested start marker detected — reject.
				return nil, fmt.Errorf(
					"evolvable zone: nested start marker at line %d (inside zone %q opened at line %d)",
					i+1, current.id, current.startLine+1,
				)
			}

			m := reEvolvableStart.FindStringSubmatch(line)
			id := m[1]

			if seenIDs[id] {
				return nil, fmt.Errorf(
					"evolvable zone: duplicate id %q at line %d", id, i+1,
				)
			}
			seenIDs[id] = true

			current = &openZone{id: id, startLine: i}
			continue
		}

		if reEvolvableEnd.MatchString(line) {
			if current == nil {
				// End marker without matching start — warn and ignore.
				slog.Warn("evolvable zone: end marker without matching start",
					"line", i+1,
				)
				continue
			}

			// Collect content lines between markers (exclusive of marker lines).
			contentLines := lines[current.startLine+1 : i]
			zones = append(zones, EvolvableZone{
				ID:        current.id,
				StartLine: current.startLine,
				EndLine:   i, // exclusive: the end marker line index
				Content:   strings.Join(contentLines, "\n"),
			})
			current = nil
		}
	}

	// Unclosed zone at end of file — warn and discard.
	if current != nil {
		slog.Warn("evolvable zone: unclosed start marker (missing end marker), zone discarded",
			"id", current.id,
			"start_line", current.startLine+1,
		)
	}

	return zones, nil
}

// MergeEvolvableZones performs a "user wins inside zones" 3-way merge for files
// containing evolvable zone markers.
//
// Strategy:
//   - Content outside evolvable zones: normal line-based 3-way merge.
//   - Content inside evolvable zones (matched by ID): current (user) content wins.
//   - Zones added by updated template (new ID): added as-is.
//   - Zones removed from updated template: user content preserved with a warning.
//   - If marker parsing fails on any version, falls back to standard LineMerge.
func MergeEvolvableZones(base, current, updated string) (string, error) {
	// Parse zones from all three versions.  Errors are non-fatal — fall back.
	baseZones, errBase := ParseEvolvableZones(base)
	currentZones, errCurrent := ParseEvolvableZones(current)
	updatedZones, errUpdated := ParseEvolvableZones(updated)

	if errBase != nil || errCurrent != nil || errUpdated != nil {
		slog.Warn("evolvable zone merge: parse error, falling back to line merge",
			"base_err", errBase,
			"current_err", errCurrent,
			"updated_err", errUpdated,
		)
		result, err := mergeLineBased([]byte(base), []byte(current), []byte(updated))
		if err != nil {
			return current, fmt.Errorf("evolvable zone merge fallback: %w", err)
		}
		return string(result.Content), nil
	}

	// Build lookup maps: id → zone content.
	baseMap := zoneContentMap(baseZones)
	currentMap := zoneContentMap(currentZones)
	updatedMap := zoneContentMap(updatedZones)

	// Reconstruct updated file with zone substitutions applied.
	updatedLines := strings.Split(updated, "\n")
	var output []string

	i := 0
	for i < len(updatedLines) {
		line := updatedLines[i]

		if reEvolvableStart.MatchString(line) {
			m := reEvolvableStart.FindStringSubmatch(line)
			id := m[1]

			// Emit the start marker unchanged.
			output = append(output, line)
			i++

			// Find the matching end marker in updated content.
			endIdx := -1
			for j := i; j < len(updatedLines); j++ {
				if reEvolvableEnd.MatchString(updatedLines[j]) {
					endIdx = j
					break
				}
			}

			// Determine which content to emit inside the zone.
			if _, existsInCurrent := currentMap[id]; existsInCurrent {
				// Zone exists in user file — user content wins.
				output = append(output, currentMap[id])
			} else if _, existsInBase := baseMap[id]; existsInBase {
				// Zone was in base but user deleted it — respect user deletion.
				// Do NOT emit zone content (skip to end marker).
			} else {
				// Brand-new zone from template — use template content.
				if _, existsInUpdated := updatedMap[id]; existsInUpdated {
					output = append(output, updatedMap[id])
				}
			}

			// Skip past the template's zone body lines.
			if endIdx >= 0 {
				i = endIdx // will emit end marker below
			} else {
				// No end marker in updated — emit remaining as-is.
				for i < len(updatedLines) {
					output = append(output, updatedLines[i])
					i++
				}
				return strings.Join(output, "\n"), nil
			}
			continue
		}

		output = append(output, line)
		i++
	}

	// Preserve any user zones that were removed from the updated template.
	for _, zone := range currentZones {
		if _, stillInUpdated := updatedMap[zone.ID]; !stillInUpdated {
			slog.Warn("evolvable zone: template removed zone, preserving user content",
				"id", zone.ID,
			)
			output = append(output,
				fmt.Sprintf(`<!-- moai:evolvable-start id="%s" -->`, zone.ID),
				zone.Content,
				"<!-- moai:evolvable-end -->",
			)
		}
	}

	return strings.Join(output, "\n"), nil
}

// mergeEvolvableZone is the merge.Engine-compatible wrapper around MergeEvolvableZones.
// It returns a MergeResult using the EvolvableZoneMerge strategy.
func mergeEvolvableZone(base, current, updated []byte) (*MergeResult, error) {
	merged, err := MergeEvolvableZones(string(base), string(current), string(updated))
	if err != nil {
		return nil, err
	}
	return &MergeResult{
		Content:     []byte(merged),
		HasConflict: false,
		Conflicts:   nil,
		Strategy:    EvolvableZoneMerge,
	}, nil
}

// zoneContentMap creates a map from zone ID to its inner content.
func zoneContentMap(zones []EvolvableZone) map[string]string {
	m := make(map[string]string, len(zones))
	for _, z := range zones {
		m[z.ID] = z.Content
	}
	return m
}

// HasEvolvableZones reports whether the content contains at least one evolvable zone marker.
// This is a cheap check used by the strategy selector.
func HasEvolvableZones(content string) bool {
	return reEvolvableStart.MatchString(content)
}

// ReplaceEvolvableZone returns content with the named zone's body replaced by
// newZoneContent. Returns ErrZoneNotFound if zoneID does not exist.
//
// This function performs a direct in-place replacement rather than a 3-way merge.
// Designed for use in the evolution Apply path; unlike MergeEvolvableZones,
// it replaces only a specific zone within a single file rather than using base/current/updated versions.
//
// newZoneContent is the pure body content without marker lines.
// It is inserted between the zone start and end markers with no blank lines in the result.
func ReplaceEvolvableZone(content, zoneID, newZoneContent string) (string, error) {
	zones, err := ParseEvolvableZones(content)
	if err != nil {
		return "", fmt.Errorf("merge: parse zones for replacement: %w", err)
	}

	// Find the target zone
	var target *EvolvableZone
	for i := range zones {
		if zones[i].ID == zoneID {
			target = &zones[i]
			break
		}
	}
	if target == nil {
		return "", ErrZoneNotFound
	}

	lines := strings.Split(content, "\n")

	// target.StartLine: start marker line index (inclusive)
	// target.EndLine: end marker line index (end marker line; not included in content)
	before := lines[:target.StartLine+1]    // up to and including the start marker
	after := lines[target.EndLine:]         // from the end marker (inclusive)

	// Build new content: guarantee exactly one trailing newline
	trimmedNew := strings.TrimRight(newZoneContent, "\n")

	var sb strings.Builder
	sb.WriteString(strings.Join(before, "\n"))
	if trimmedNew != "" {
		sb.WriteString("\n")
		sb.WriteString(trimmedNew)
		sb.WriteString("\n")
	} else {
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Join(after, "\n"))

	return sb.String(), nil
}
