package template

import (
	"io/fs"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// embeddedWorkflowYAMLPath is the shipped workflow.yaml inside the embedded
// filesystem (EmbeddedTemplates strips the "templates/" prefix).
const embeddedWorkflowYAMLPath = ".moai/config/sections/workflow.yaml"

// TestAC_CONTRACT_020 decodes the EMBEDDED template workflow.yaml and asserts
// the shipped workflow.autonomy defaults (AC-CONTRACT-020 Go half; the
// neutrality scan and the local-values grep are shell checks).
func TestAC_CONTRACT_020(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	raw, err := fs.ReadFile(fsys, embeddedWorkflowYAMLPath)
	if err != nil {
		t.Fatalf("read %s from embedded FS: %v", embeddedWorkflowYAMLPath, err)
	}

	var generic struct {
		Workflow struct {
			Autonomy struct {
				Mode     string         `yaml:"mode"`
				Contract map[string]any `yaml:"contract"`
				Kickoff  map[string]any `yaml:"kickoff"`
			} `yaml:"autonomy"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal embedded %s: %v", embeddedWorkflowYAMLPath, err)
	}
	a := generic.Workflow.Autonomy
	if a.Mode != "guided" {
		t.Errorf("workflow.autonomy.mode = %q, want guided", a.Mode)
	}
	if got := a.Contract["second_review"]; got != "required" {
		t.Errorf("workflow.autonomy.contract.second_review = %v, want required", got)
	}
	if a.Kickoff == nil {
		t.Error("workflow.autonomy.kickoff section is absent from the template")
	}
	if _, ok := a.Kickoff["decider"]; ok {
		t.Errorf("workflow.autonomy.kickoff carries a decider key (%v); the template must omit it", a.Kickoff["decider"])
	}

	var typed struct {
		Workflow config.WorkflowConfig `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(raw, &typed); err != nil {
		t.Fatalf("unmarshal embedded %s into WorkflowConfig: %v", embeddedWorkflowYAMLPath, err)
	}
	s := config.ResolveAutonomy(typed.Workflow)
	if s.Decider != "human" {
		t.Errorf("effective decider for the template as shipped = %q, want human", s.Decider)
	}
	if len(s.Warnings) != 0 || s.DeciderError != nil {
		t.Errorf("template as shipped resolves with warnings %v / error %v", s.Warnings, s.DeciderError)
	}
}
