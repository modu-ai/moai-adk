package harness

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func writeYAML(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	return path
}

func TestUpdateWorkflowYAML_InjectsHarness(t *testing.T) {
	dir := t.TempDir()
	src := `workflow:
  team:
    enabled: true
`
	path := writeYAML(t, dir, "workflow.yaml", src)
	agents := []AgentRef{{Name: "ios-architect", Path: ".claude/agents/my-harness/ios-architect.md", InvokeIn: []string{"plan", "run"}}}
	skills := []SkillRef{{Name: "ios-patterns", Path: ".claude/skills/my-harness-ios-patterns/SKILL.md", TriggersIn: []string{"plan", "run"}}}
	chains := []ChainRule{{Phase: "run", BeforeSpecialist: "ios-architect", AfterSpecialist: "swiftui-engineer"}}

	if err := UpdateWorkflowYAML(path, "ios-mobile", "SPEC-PROJ-INIT-001", agents, skills, chains); err != nil {
		t.Fatalf("UpdateWorkflowYAML: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wf, ok := doc["workflow"].(map[string]any)
	if !ok {
		t.Fatal("workflow section missing")
	}
	team, ok := wf["team"].(map[string]any)
	if !ok || team["enabled"] != true {
		t.Errorf("team section not preserved: %v", wf["team"])
	}
	harness, ok := wf["harness"].(map[string]any)
	if !ok {
		t.Fatal("harness section missing")
	}
	if harness["enabled"] != true {
		t.Errorf("harness.enabled = %v, want true", harness["enabled"])
	}
	if harness["domain"] != "ios-mobile" {
		t.Errorf("domain = %v", harness["domain"])
	}
	if harness["spec_id"] != "SPEC-PROJ-INIT-001" {
		t.Errorf("spec_id = %v", harness["spec_id"])
	}
}

func TestUpdateWorkflowYAML_Idempotent(t *testing.T) {
	dir := t.TempDir()
	src := "workflow:\n  team:\n    enabled: true\n"
	path := writeYAML(t, dir, "workflow.yaml", src)
	agents := []AgentRef{{Name: "a", Path: "p", InvokeIn: []string{"plan"}}}
	if err := UpdateWorkflowYAML(path, "d", "S-1", agents, nil, nil); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := UpdateWorkflowYAML(path, "d2", "S-1", agents, nil, nil); err != nil {
		t.Fatalf("second call: %v", err)
	}
	// After second call, only one harness section should exist (yaml mapping
	// keys are unique by definition so we just verify domain reflects update).
	data, _ := os.ReadFile(path)
	var doc map[string]any
	_ = yaml.Unmarshal(data, &doc)
	wf := doc["workflow"].(map[string]any)
	harness := wf["harness"].(map[string]any)
	if harness["domain"] != "d2" {
		t.Errorf("expected idempotent overwrite, got domain=%v", harness["domain"])
	}
}

func TestUpdateWorkflowYAML_NoWorkflowKey(t *testing.T) {
	dir := t.TempDir()
	src := "other:\n  thing: 1\n"
	path := writeYAML(t, dir, "workflow.yaml", src)
	if err := UpdateWorkflowYAML(path, "d", "S", nil, nil, nil); err != nil {
		t.Fatalf("expected to create workflow: %v", err)
	}
	data, _ := os.ReadFile(path)
	var doc map[string]any
	_ = yaml.Unmarshal(data, &doc)
	if _, ok := doc["other"]; !ok {
		t.Errorf("other key dropped")
	}
	if _, ok := doc["workflow"]; !ok {
		t.Errorf("workflow not created")
	}
}

func TestUpdateWorkflowYAML_EmptyPath(t *testing.T) {
	if err := UpdateWorkflowYAML("", "d", "S", nil, nil, nil); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestUpdateWorkflowYAML_NotFound(t *testing.T) {
	if err := UpdateWorkflowYAML(filepath.Join(t.TempDir(), "missing.yaml"), "d", "S", nil, nil, nil); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestUpdateWorkflowYAML_TopLevelNotMapping(t *testing.T) {
	dir := t.TempDir()
	path := writeYAML(t, dir, "workflow.yaml", "- list\n- entries\n")
	if err := UpdateWorkflowYAML(path, "d", "S", nil, nil, nil); err == nil {
		t.Fatal("expected error for non-mapping top level")
	}
}
