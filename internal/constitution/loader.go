package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// designMirrorStart is the design subsystem mirror entry ID range start (051).
const designMirrorStart = 51

// designMirrorEnd is the design subsystem mirror entry ID range end (099, inclusive).
const designMirrorEnd = 99

// designOverflowEnd is the design mirror overflow range end (149, inclusive).
const designOverflowEnd = 149

// Registry represents a loaded zone registry.
// Entries is the full list of entries, and lookup is an internal map for O(1) ID→index lookup.
type Registry struct {
	// Entries is the list of loaded Rules.
	Entries []Rule
	// Warnings is the list of warning messages generated during loading (orphan, overflow, etc.).
	Warnings []string
	// lookup is an ID→index internal map (supports O(1) Get).
	lookup map[string]int
}

// Get retrieves a Rule by ID in O(1) time.
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

// FilterByZone returns a list of Rules for the specified zone.
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

// rawEntry is an internal struct for YAML unmarshaling.
// Separately defined to parse the Zone field as a string first.
type rawEntry struct {
	ID         string `yaml:"id"`
	Zone       string `yaml:"zone"`
	File       string `yaml:"file"`
	Anchor     string `yaml:"anchor"`
	Clause     string `yaml:"clause"`
	CanaryGate bool   `yaml:"canary_gate"`
}

// LoadRegistry loads a zone registry markdown file from the specified path.
//
// Parsing strategy (plan.md §7 OQ1 Decision):
//  1. Extract the first ```yaml ... ``` code fence from the file
//  2. Unmarshal to []rawEntry using gopkg.in/yaml.v3
//  3. Validate Rules: duplicate IDs are fatal errors, orphan files generate warnings + Orphan=true
//
// projectDir is the base directory for filepath.Clean + scope restriction.
// Limits the registry file path to within projectDir to prevent path traversal.
func LoadRegistry(path, projectDir string) (*Registry, error) {
	// Prevent path traversal: clean path and verify it's within projectDir
	cleanPath := filepath.Clean(path)
	cleanProjectDir := filepath.Clean(projectDir)
	if filepath.IsAbs(cleanPath) {
		// If absolute path, convert to relative path from projectDir for scope verification
		rel, err := filepath.Rel(cleanProjectDir, cleanPath)
		if err == nil && strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("registry path %q escapes project dir %q", path, projectDir)
		}
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("error reading registry file %q: %w", path, err)
	}

	yamlBlock, err := extractYAMLFence(string(data))
	if err != nil {
		return nil, fmt.Errorf("error extracting YAML code fence: %w", err)
	}

	var raw []rawEntry
	if err := yaml.Unmarshal([]byte(yamlBlock), &raw); err != nil {
		return nil, fmt.Errorf("YAML parsing error: %w", err)
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
			return nil, fmt.Errorf("entry %d (id=%q) zone parsing error: %w", i, re.ID, parseErr)
		}

		rule := Rule{
			ID:         re.ID,
			Zone:       z,
			File:       re.File,
			Anchor:     re.Anchor,
			Clause:     re.Clause,
			CanaryGate: re.CanaryGate,
		}

		// ID validation
		if validErr := rule.Validate(); validErr != nil {
			return nil, fmt.Errorf("entry %d validation error: %w", i, validErr)
		}

		// Duplicate ID detection (fatal)
		if _, exists := reg.lookup[rule.ID]; exists {
			return nil, fmt.Errorf("duplicate ID detected: %q", rule.ID)
		}

		// Orphan file check: verify file path actually exists
		// Interpret relative path from projectDir
		rulePath := rule.File
		if !filepath.IsAbs(rulePath) {
			rulePath = filepath.Join(cleanProjectDir, rulePath)
		}
		if _, statErr := os.Stat(filepath.Clean(rulePath)); statErr != nil {
			rule = rule.withOrphan(true)
			reg.Warnings = append(reg.Warnings,
				fmt.Sprintf("orphan: file %q not found (rule %s)", rule.File, rule.ID))
		}

		// Design mirror overflow detection (exceeds 051-099 range)
		idNum := extractIDNumber(rule.ID)
		if idNum >= designMirrorStart && idNum <= designMirrorEnd {
			mirrorCount++
		}
		if idNum > designMirrorEnd && idNum <= designOverflowEnd {
			// Using overflow range
			reg.Warnings = append(reg.Warnings,
				fmt.Sprintf("design mirror overflow: ID %s uses extended range (100-149)", rule.ID))
		}

		reg.lookup[rule.ID] = len(reg.Entries)
		reg.Entries = append(reg.Entries, rule)
	}

	// Detect design mirror 49 slot exceeded
	if mirrorCount > designMirrorEnd-designMirrorStart+1 {
		reg.Warnings = append(reg.Warnings,
			fmt.Sprintf("design mirror overflow: mirror entry count %d exceeds allowed 49 slots", mirrorCount))
	}

	// Warnings are stored in reg.Warnings and returned.
	// Caller verifies orphan, overflow warnings via reg.Warnings.
	// REQ-CON-001-040: orphan does not return error, only sets Orphan=true flag.
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

	// Position after openFence
	afterOpen := start + len(openFence)
	// Find closing fence (after openFence)
	end := strings.Index(content[afterOpen:], closeFence)
	if end == -1 {
		return "", fmt.Errorf("closing ``` fence not found")
	}

	return content[afterOpen : afterOpen+end], nil
}

// extractIDNumber extracts the numeric part from CONST-V3R2-NNN format.
// Returns -1 on parsing failure.
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
