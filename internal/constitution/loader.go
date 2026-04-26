package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// designMirrorStart is the start of the design subsystem mirror entry ID range (051).
const designMirrorStart = 51

// designMirrorEnd is the end of the design subsystem mirror entry ID range (099, inclusive).
const designMirrorEnd = 99

// designOverflowEnd is the end of the design mirror overflow range (149, inclusive).
const designOverflowEnd = 149

// Registry represents the loaded zone registry.
// Entries is the full list of entries; lookup is an internal map for O(1) ID→index lookup.
type Registry struct {
	// Entries is the list of loaded Rules.
	Entries []Rule
	// Warnings is the list of warning messages generated during loading (orphan, overflow, etc.).
	Warnings []string
	// lookup is the internal ID→index map (supports O(1) Get).
	lookup map[string]int
}

// Get looks up a Rule by ID in O(1).
func (r *Registry) Get(id string) (Rule, bool) {
	if r.lookup == nil {
		return Rule{}, false
	}
	idx, ok := r.lookup[id]
	if !ok {
		return Rule{}, false
	}
	return r.Entries[idx], true
}

// FilterByZone returns the list of Rules in the specified zone.
// Supports REQ-CON-001-012.
func (r *Registry) FilterByZone(z Zone) []Rule {
	var result []Rule
	for _, entry := range r.Entries {
		if entry.Zone == z {
			result = append(result, entry)
		}
	}
	return result
}

// rawEntry is the internal struct for YAML unmarshalling.
// Defined separately to parse the Zone field as a string first.
type rawEntry struct {
	ID         string `yaml:"id"`
	Zone       string `yaml:"zone"`
	File       string `yaml:"file"`
	Anchor     string `yaml:"anchor"`
	Clause     string `yaml:"clause"`
	CanaryGate bool   `yaml:"canary_gate"`
}

// LoadRegistry loads the zone registry markdown file at the specified path.
//
// Parsing strategy (plan.md §7 OQ1 Decision):
//  1. Extract the first ```yaml ... ``` code fence from the file
//  2. Unmarshal into []rawEntry using gopkg.in/yaml.v3
//  3. Rule validation: duplicate IDs are a fatal error; orphan files generate a warning + Orphan=true
//
// projectDir is the base directory for filepath.Clean and scope restriction.
// The registry file path is restricted to within projectDir to prevent path traversal.
func LoadRegistry(path, projectDir string) (*Registry, error) {
	// Path traversal prevention: clean the path and verify it is within projectDir scope.
	cleanPath := filepath.Clean(path)
	cleanProjectDir := filepath.Clean(projectDir)
	if filepath.IsAbs(cleanPath) {
		// For absolute paths, convert to a relative path from projectDir to check scope.
		rel, err := filepath.Rel(cleanProjectDir, cleanPath)
		if err == nil && strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("registry path %q escapes project dir %q", path, projectDir)
		}
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("registry file read error %q: %w", path, err)
	}

	yamlBlock, err := extractYAMLFence(string(data))
	if err != nil {
		return nil, fmt.Errorf("YAML code fence extraction error: %w", err)
	}

	var raw []rawEntry
	if err := yaml.Unmarshal([]byte(yamlBlock), &raw); err != nil {
		return nil, fmt.Errorf("YAML parse error: %w", err)
	}

	reg := &Registry{
		Entries:  make([]Rule, 0, len(raw)),
		Warnings: nil,
		lookup:   make(map[string]int, len(raw)),
	}

	var mirrorCount int

	for i, re := range raw {
		z, parseErr := ParseZone(re.Zone)
		if parseErr != nil {
			return nil, fmt.Errorf("entry %d (id=%q) zone parse error: %w", i, re.ID, parseErr)
		}

		rule := Rule{
			ID:         re.ID,
			Zone:       z,
			File:       re.File,
			Anchor:     re.Anchor,
			Clause:     re.Clause,
			CanaryGate: re.CanaryGate,
		}

		// Validate the ID.
		if validErr := rule.Validate(); validErr != nil {
			return nil, fmt.Errorf("entry %d validation error: %w", i, validErr)
		}

		// Detect duplicate IDs (fatal).
		if _, exists := reg.lookup[rule.ID]; exists {
			return nil, fmt.Errorf("duplicate ID detected: %q", rule.ID)
		}

		// Check for orphan files: verify that the file path actually exists.
		// Relative paths are resolved relative to projectDir.
		rulePath := rule.File
		if !filepath.IsAbs(rulePath) {
			rulePath = filepath.Join(cleanProjectDir, rulePath)
		}
		if _, statErr := os.Stat(filepath.Clean(rulePath)); statErr != nil {
			rule = rule.withOrphan(true)
			reg.Warnings = append(reg.Warnings,
				fmt.Sprintf("orphan: file %q not found (rule %s)", rule.File, rule.ID))
		}

		// Detect design mirror overflow (exceeds the 051-099 range).
		idNum := extractIDNumber(rule.ID)
		if idNum >= designMirrorStart && idNum <= designMirrorEnd {
			mirrorCount++
		}
		if idNum > designMirrorEnd && idNum <= designOverflowEnd {
			// Using overflow range.
			reg.Warnings = append(reg.Warnings,
				fmt.Sprintf("design mirror overflow: ID %s is using the extended range (100-149)", rule.ID))
		}

		reg.lookup[rule.ID] = len(reg.Entries)
		reg.Entries = append(reg.Entries, rule)
	}

	// Detect when design mirror count exceeds the 49-slot limit.
	if mirrorCount > designMirrorEnd-designMirrorStart+1 {
		reg.Warnings = append(reg.Warnings,
			fmt.Sprintf("design mirror overflow: mirror entry count %d exceeds the allowed 49 slots", mirrorCount))
	}

	// Warnings are stored in reg.Warnings and returned to the caller.
	// Orphan and overflow warnings are surfaced via reg.Warnings.
	// REQ-CON-001-040: orphans do not cause an error return; only the Orphan=true flag is set.
	return reg, nil
}

// extractYAMLFence extracts the first ```yaml ... ``` block from a markdown document.
func extractYAMLFence(content string) (string, error) {
	const openFence = "```yaml"
	const closeFence = "```"

	start := strings.Index(content, openFence)
	if start == -1 {
		return "", fmt.Errorf("```yaml code fence not found")
	}

	// Position immediately after openFence.
	afterOpen := start + len(openFence)
	// Find the closing fence (starting from after openFence).
	end := strings.Index(content[afterOpen:], closeFence)
	if end == -1 {
		return "", fmt.Errorf("closing ``` fence not found")
	}

	return content[afterOpen : afterOpen+end], nil
}

// extractIDNumber extracts the numeric part from the CONST-V3R2-NNN format.
// Returns -1 on parse failure.
func extractIDNumber(id string) int {
	const prefix = "CONST-V3R2-"
	if !strings.HasPrefix(id, prefix) {
		return -1
	}
	numStr := id[len(prefix):]
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return -1
	}
	return n
}
