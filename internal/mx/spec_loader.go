package mx

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// specFrontmatter holds the fields parsed from a spec.md YAML frontmatter.
type specFrontmatter struct {
	ID        string      `yaml:"id"`
	Module    interface{} `yaml:"module"`
	DependsOn []string    `yaml:"depends_on"`
}

// LoadSpecModules walks projectRoot/.moai/specs/*/spec.md and returns a
// SPEC ID → []modulePath map (REQ-SPC-004-005).
//
// @MX:ANCHOR: [AUTO] LoadSpecModules — SPEC module path loader; CLI, Resolver, and SpecAssociator all configure path-based association via this function
// @MX:REASON: fan_in >= 3 — invoked from CLI mx_query.go, the Resolver init path, and the future codemaps generator
//
// Supported `module` field formats:
//   - String: "internal/mx/, cmd/moai/" → split on commas + TrimSpace
//   - YAML sequence: [internal/foo/, internal/bar/] → as-is
//   - Empty string: "" → empty slice
//
// Returns an empty map without error when .moai/specs/ does not exist.
func LoadSpecModules(projectRoot string) (map[string][]string, error) {
	specsDir := filepath.Join(projectRoot, ".moai", "specs")

	if _, err := os.Stat(specsDir); errors.Is(err, os.ErrNotExist) {
		return map[string][]string{}, nil
	}

	pattern := filepath.Join(specsDir, "*", "spec.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Sort for deterministic ordering
	sort.Strings(matches)

	result := make(map[string][]string)

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		fm, err := parseFrontmatter(data)
		if err != nil || fm.ID == "" {
			continue
		}

		result[fm.ID] = parseModuleField(fm.Module)
	}

	return result, nil
}

// LoadSpecDependencies walks projectRoot/.moai/specs/*/spec.md and returns a
// SPEC ID → depends_on SPEC-ID list map (the spec.md frontmatter edge list).
// It mirrors LoadSpecModules: same glob, same parseFrontmatter, same
// skip-on-error semantics — consumed by the graph writer (internal/graph) to
// persist SPEC→SPEC edges that already exist in frontmatter but are never
// aggregated.
//
// @MX:NOTE: [AUTO] LoadSpecDependencies — frontmatter depends_on edge-list loader for the graph writer
func LoadSpecDependencies(projectRoot string) (map[string][]string, error) {
	specsDir := filepath.Join(projectRoot, ".moai", "specs")

	if _, err := os.Stat(specsDir); errors.Is(err, os.ErrNotExist) {
		return map[string][]string{}, nil
	}

	pattern := filepath.Join(specsDir, "*", "spec.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Sort for deterministic ordering
	sort.Strings(matches)

	result := make(map[string][]string)

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		fm, err := parseFrontmatter(data)
		if err != nil || fm.ID == "" {
			continue
		}

		deps := make([]string, 0, len(fm.DependsOn))
		for _, d := range fm.DependsOn {
			if d = strings.TrimSpace(d); d != "" {
				deps = append(deps, d)
			}
		}
		result[fm.ID] = deps
	}

	return result, nil
}

// parseFrontmatter parses the YAML frontmatter (between the `---` delimiters)
// of a spec.md file.
func parseFrontmatter(data []byte) (specFrontmatter, error) {
	content := string(data)

	// Extract the frontmatter starting with `---`
	if !strings.HasPrefix(content, "---") {
		return specFrontmatter{}, nil
	}

	// Find the second `---` after the first one
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		// No closing `---`: treat the remainder of the file as frontmatter
		end = len(rest)
	}

	yamlContent := rest[:end]

	var fm specFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return specFrontmatter{}, err
	}

	return fm, nil
}

// parseModuleField converts the `module` field value into a []string.
// Supported types:
//   - string: comma-separated → split + TrimSpace, drop empty entries
//   - []interface{}: cast each element to string
func parseModuleField(v interface{}) []string {
	if v == nil {
		return []string{}
	}

	switch val := v.(type) {
	case string:
		if val == "" {
			return []string{}
		}
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result

	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
		return result

	default:
		return []string{}
	}
}
