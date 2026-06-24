package toolpolicy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a tool-policy YAML file, then validates every entry
// against the 6-field schema (REQ-TPS-002). A malformed YAML or an entry with
// an invalid risk_tier/decision/missing-required-field returns a clear error
// rather than silently producing wrong policy (acceptance.md EC-5).
func Load(path string) (*PolicyDocument, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tool-policy read %q: %w", path, err)
	}
	return parseDocument(raw, path)
}

// parseDocument parses YAML bytes into a PolicyDocument and validates it.
// Split from Load so tests can feed fixture bytes without touching disk.
func parseDocument(raw []byte, source string) (*PolicyDocument, error) {
	var doc PolicyDocument
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("tool-policy parse %q: %w", source, err)
	}
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("tool-policy validate %q: %w", source, err)
	}
	return &doc, nil
}

// Validate checks that every entry carries the 6 required fields with valid
// enum values. Returns the first violation encountered (acceptance.md EC-5:
// the loader MUST return a clear error on malformed YAML, not silently
// produce wrong policy).
func (d *PolicyDocument) Validate() error {
	if d == nil {
		return fmt.Errorf("policy document is nil")
	}
	for i, e := range d.Entries {
		if e.Tool == "" {
			return fmt.Errorf("entry[%d]: tool is required (6-field schema)", i)
		}
		if e.RiskTier == "" {
			return fmt.Errorf("entry[%d] tool=%q: risk_tier is required", i, e.Tool)
		}
		if !e.RiskTier.IsValid() {
			return fmt.Errorf("entry[%d] tool=%q: risk_tier %q not in {read,write,irreversible}", i, e.Tool, e.RiskTier)
		}
		if e.Decision == "" {
			return fmt.Errorf("entry[%d] tool=%q: decision is required", i, e.Tool)
		}
		if !e.Decision.IsValid() {
			return fmt.Errorf("entry[%d] tool=%q: decision %q not in {allow,deny,ask}", i, e.Tool, e.Decision)
		}
		if e.OwnerAgent == "" {
			return fmt.Errorf("entry[%d] tool=%q: owner_agent is required", i, e.Tool)
		}
		if e.Audit == "" {
			return fmt.Errorf("entry[%d] tool=%q: audit is required", i, e.Tool)
		}
		// ArgsPattern is intentionally NOT required to be non-empty: a bare
		// tool-level entry (e.g., "Read" with no args restriction) has an
		// empty args_pattern, which SettingsSpecifier renders as just "Read".
	}
	return nil
}

// LoadFromProjectDir resolves .moai/config/sections/tool-policy.yaml relative
// to the given project directory and loads it. Convenience for CLI callers
// that already have the project root.
func LoadFromProjectDir(projectDir string) (*PolicyDocument, error) {
	path := filepath.Join(projectDir, ".moai", "config", "sections", "tool-policy.yaml")
	return Load(path)
}

// SortedBySpecifier returns the entries sorted by their settings specifier
// ("Tool" or "Tool(args)") in ascending lexicographic order. Used by codegen
// to guarantee deterministic output (AC-TPS-013 idempotency — the generated
// permissions block must be byte-identical across runs with no ordering
// nondeterminism).
func (d *PolicyDocument) SortedBySpecifier() []PolicyEntry {
	out := make([]PolicyEntry, len(d.Entries))
	copy(out, d.Entries)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].SettingsSpecifier() < out[j].SettingsSpecifier()
	})
	return out
}

// FilterByRiskTier returns entries matching the given risk tier.
func (d *PolicyDocument) FilterByRiskTier(t RiskTier) []PolicyEntry {
	var out []PolicyEntry
	for _, e := range d.Entries {
		if e.RiskTier == t {
			out = append(out, e)
		}
	}
	return out
}

// FilterByDecision returns entries matching the given decision.
func (d *PolicyDocument) FilterByDecision(dec Decision) []PolicyEntry {
	var out []PolicyEntry
	for _, e := range d.Entries {
		if e.Decision == dec {
			out = append(out, e)
		}
	}
	return out
}

// FilterByTool returns entries whose Tool field matches the given tool name.
func (d *PolicyDocument) FilterByTool(tool string) []PolicyEntry {
	if tool == "" {
		return d.Entries
	}
	var out []PolicyEntry
	for _, e := range d.Entries {
		if e.Tool == tool {
			out = append(out, e)
		}
	}
	return out
}
