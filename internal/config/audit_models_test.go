package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestAuditConfigYAMLRoundTrip is the dedicated AuditConfig drift guard
// (SPEC-V3R6-AUDIT-MODEL-PIN-001 AC-AMP-001 / plan.md M1 N2). It was chosen
// over a symmetryCases extension because checkSymmetry navigates ONE level
// (rawRoot[yamlTopKey]) and cannot reach the nested workflow.audit block, and
// an AuditConfig-tagged case would require the template block to carry
// model+gates as well (4 struct tags vs 2 yaml keys).
//
// Guard contract: marshal a fully-populated AuditConfig → unmarshal →
// field-for-field equality; dropping either new yaml key (codex / glm) breaks
// the round-trip because the corresponding struct field falls back to zero —
// the key↔field binding is load-bearing, not incidental. Removing a struct
// field fails compilation of the explicit field references below.
func TestAuditConfigYAMLRoundTrip(t *testing.T) {
	populated := AuditConfig{
		Model: AuditModelCodex,
		Gates: AuditGates{Claude: AuditGateRequired, Codex: AuditGateRequired, GLM: AuditGateAdvisory},
		Codex: ModelEffort{Model: "gpt-5.6-sol", Effort: "high"},
		GLM:   ModelEffort{Model: "glm-5.3", Effort: "max"},
	}

	data, err := yaml.Marshal(populated)
	if err != nil {
		t.Fatalf("marshal populated AuditConfig: %v", err)
	}

	var back AuditConfig
	if err := yaml.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal marshaled AuditConfig: %v", err)
	}

	// Field-for-field equality — explicit references so a dropped struct field
	// is a compile error, not a silent pass.
	if back.Model != populated.Model {
		t.Errorf("Model: got %q, want %q", back.Model, populated.Model)
	}
	if back.Gates != populated.Gates {
		t.Errorf("Gates: got %+v, want %+v", back.Gates, populated.Gates)
	}
	if back.Codex != populated.Codex {
		t.Errorf("Codex pin: got %+v, want %+v", back.Codex, populated.Codex)
	}
	if back.GLM != populated.GLM {
		t.Errorf("GLM pin: got %+v, want %+v", back.GLM, populated.GLM)
	}

	// The marshaled yaml must carry both new sub-keys with their nested
	// model/effort pairs — a missing key would make the unmarshal arm below
	// vacuously pass.
	for _, key := range []string{"codex:", "glm:", "model:", "effort:"} {
		if !containsLine(data, key) {
			t.Errorf("marshaled yaml missing key %q:\n%s", key, data)
		}
	}

	// Key-drop arms: yaml WITHOUT the codex key unmarshals a zero Codex pin
	// (and likewise for glm) — proving each yaml key is what populates its
	// struct field.
	withoutCodex := "model: codex\nglm:\n    model: glm-5.3\n    effort: max\n"
	var droppedCodex AuditConfig
	if err := yaml.Unmarshal([]byte(withoutCodex), &droppedCodex); err != nil {
		t.Fatalf("unmarshal codex-dropped yaml: %v", err)
	}
	if droppedCodex.Codex != (ModelEffort{}) {
		t.Errorf("Codex pin populated without a codex: key: got %+v, want zero", droppedCodex.Codex)
	}
	if droppedCodex.GLM.Model != "glm-5.3" || droppedCodex.GLM.Effort != "max" {
		t.Errorf("GLM pin lost by codex-drop arm: got %+v", droppedCodex.GLM)
	}

	withoutGLM := "model: codex\ncodex:\n    model: gpt-5.6-sol\n    effort: high\n"
	var droppedGLM AuditConfig
	if err := yaml.Unmarshal([]byte(withoutGLM), &droppedGLM); err != nil {
		t.Fatalf("unmarshal glm-dropped yaml: %v", err)
	}
	if droppedGLM.GLM != (ModelEffort{}) {
		t.Errorf("GLM pin populated without a glm: key: got %+v, want zero", droppedGLM.GLM)
	}
	if droppedGLM.Codex.Model != "gpt-5.6-sol" || droppedGLM.Codex.Effort != "high" {
		t.Errorf("Codex pin lost by glm-drop arm: got %+v", droppedGLM.Codex)
	}
}

// TestAuditConfigYAMLWorkflowWrapperLoad verifies the pin block loads through
// the workflow.yaml `workflow:` wrapper exactly as loadWorkflowAuditSection
// (internal/cli) unmarshals it — the loader contract AC-AMP-001 cites.
func TestAuditConfigYAMLWorkflowWrapperLoad(t *testing.T) {
	const fixture = "workflow:\n" +
		"    audit:\n" +
		"        model: multi\n" +
		"        codex:\n" +
		"            model: gpt-5.6-sol\n" +
		"            effort: high\n" +
		"        glm:\n" +
		"            model: glm-5.3\n" +
		"            effort: max\n"
	var wrapper struct {
		Workflow struct {
			Audit AuditConfig `yaml:"audit"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal([]byte(fixture), &wrapper); err != nil {
		t.Fatalf("unmarshal workflow wrapper fixture: %v", err)
	}
	got := wrapper.Workflow.Audit
	if got.Model != AuditModelMulti {
		t.Errorf("Model: got %q, want %q", got.Model, AuditModelMulti)
	}
	if got.Codex.Model != "gpt-5.6-sol" || got.Codex.Effort != "high" {
		t.Errorf("Codex pin: got %+v, want {gpt-5.6-sol high}", got.Codex)
	}
	if got.GLM.Model != "glm-5.3" || got.GLM.Effort != "max" {
		t.Errorf("GLM pin: got %+v, want {glm-5.3 max}", got.GLM)
	}
}

// containsLine reports whether the marshaled bytes contain the given key with
// surrounding whitespace tolerance (yaml.v3 indents nested keys).
func containsLine(data []byte, key string) bool {
	s := string(data)
	for _, line := range splitLines(s) {
		if trimSpaces(line) == key || hasKey(line, key) {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

func trimSpaces(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}

func hasKey(line, key string) bool {
	t := trimSpaces(line)
	return len(t) >= len(key) && t[:len(key)] == key
}
