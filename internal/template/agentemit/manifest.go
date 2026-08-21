// manifest.go — the Codex mapping manifest types and embedded loader.
package agentemit

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	_ "embed"
)

//go:embed agents-codex.yaml
var manifestYAML []byte

// LayoutConfig selects where emitted TOMLs live under .codex/agents/.
type LayoutConfig struct {
	Mode         string `yaml:"mode"`         // subdirectory | flat_prefix
	Subdirectory string `yaml:"subdirectory"` // subdir name (subdirectory mode)
	FlatPrefix   string `yaml:"flat_prefix"`  // filename prefix (flat_prefix mode)
}

// FieldConfig is one optional-field emission switch.
type FieldConfig struct {
	Emit  bool              `yaml:"emit"`
	Value string            `yaml:"value"` // emitted value (sandbox_mode)
	Map   map[string]string `yaml:"map"`   // source->target map (model_reasoning_effort)
	// AcceptedValues is the probe-measured value set for scalar fields
	// (sandbox_mode). When non-empty and Emit is true, Value must belong to
	// it — the emitter's own enforcement of the measured enumeration.
	AcceptedValues []string `yaml:"accepted_values"`
}

// ClassDisposition is one semantic class row of the mapping table.
type ClassDisposition struct {
	Class       string   `yaml:"class"`
	Disposition string   `yaml:"disposition"`
	Field       string   `yaml:"field"`
	Value       []string `yaml:"value"`
	Rationale   string   `yaml:"rationale"`
}

// DocumentedDrop records a Claude capability with no Codex equivalent.
type DocumentedDrop struct {
	ID        string `yaml:"id"`
	Rationale string `yaml:"rationale"`
}

// CorrespondenceNote records a no-artifact correspondence.
type CorrespondenceNote struct {
	ID   string `yaml:"id"`
	Note string `yaml:"note"`
}

// Manifest is the machine-readable Codex mapping manifest.
type Manifest struct {
	CodexMeasuredVersion string                  `yaml:"codex_measured_version"`
	Layout               LayoutConfig            `yaml:"layout"`
	Fields               map[string]*FieldConfig `yaml:"fields"`
	ToolClasses          map[string]string       `yaml:"tool_classes"`
	Classes              []ClassDisposition      `yaml:"classes"`
	DocumentedDrops      []DocumentedDrop        `yaml:"documented_drops"`
	CorrespondenceNotes  []CorrespondenceNote    `yaml:"correspondence_notes"`
}

// LoadManifest parses and self-validates the embedded default manifest.
// Self-validation is fail-closed (AC-013's mechanical face): every class row
// must carry a known disposition and a non-empty rationale, and every
// tool-class value must reference an existing disposition row — an
// unmapped class or a silent discard is a manifest bug, not a warning.
func LoadManifest() (Manifest, error) {
	return ParseManifest(manifestYAML)
}

// ParseManifest decodes and self-validates manifest bytes (the embedded
// default, or a caller-supplied variant — M4's wiring generator reuses this
// as its output-checking layer). Exported for the M4 seam.
func ParseManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("agentemit: manifest parse: %w", err)
	}
	if m.CodexMeasuredVersion == "" {
		return Manifest{}, fmt.Errorf("agentemit: manifest missing codex_measured_version")
	}
	switch m.Layout.Mode {
	case "subdirectory", "flat_prefix":
	default:
		return Manifest{}, fmt.Errorf("agentemit: manifest layout mode %q invalid (want subdirectory|flat_prefix)", m.Layout.Mode)
	}
	if len(m.ToolClasses) == 0 {
		return Manifest{}, fmt.Errorf("agentemit: manifest has no tool_classes")
	}
	validDisposition := map[string]bool{
		"no-field": true, "consequence": true, "emit-field": true,
		"documented-drop": true, "deferred-m1": true, "omit": true,
		"correspondence-note": true,
	}
	rows := map[string]bool{}
	for _, c := range m.Classes {
		if !validDisposition[c.Disposition] {
			return Manifest{}, fmt.Errorf("agentemit: manifest class %s has invalid disposition %q", c.Class, c.Disposition)
		}
		if strings.TrimSpace(c.Rationale) == "" {
			return Manifest{}, fmt.Errorf("agentemit: manifest class %s disposition %q carries no rationale (no silent discard)", c.Class, c.Disposition)
		}
		rows[c.Class] = true
	}
	for token, class := range m.ToolClasses {
		if !rows[class] {
			return Manifest{}, fmt.Errorf("agentemit: manifest tool %s maps to class %s with no disposition row", token, class)
		}
	}
	if fc, ok := m.Fields["model_reasoning_effort"]; ok && fc.Emit && len(fc.Map) == 0 {
		return Manifest{}, fmt.Errorf("agentemit: manifest emits model_reasoning_effort with an empty map")
	}
	if fc, ok := m.Fields["sandbox_mode"]; ok && fc.Emit && len(fc.AcceptedValues) > 0 {
		accepted := make(map[string]bool, len(fc.AcceptedValues))
		for _, v := range fc.AcceptedValues {
			accepted[v] = true
		}
		if !accepted[fc.Value] {
			return Manifest{}, fmt.Errorf("agentemit: manifest sandbox_mode value %q is outside the measured value set %v — never emit an unconfirmed value", fc.Value, fc.AcceptedValues)
		}
	}
	return m, nil
}
