package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The shipped workflow.yaml names the escalation detector's one configuration
// key explicitly under A1's autonomy.escalation block: new_api_detector:
// graph. The template as shipped resolves it without a warning, and the
// block's comment carries no card id, SPEC id, date, or commit SHA.
func TestTemplateAutonomyEscalationNewAPIDetector(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	raw, err := fs.ReadFile(fsys, embeddedWorkflowYAMLPath)
	if err != nil {
		t.Fatalf("read %s: %v", embeddedWorkflowYAMLPath, err)
	}

	var generic struct {
		Workflow struct {
			Autonomy struct {
				Escalation map[string]any `yaml:"escalation"`
			} `yaml:"autonomy"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := generic.Workflow.Autonomy.Escalation["new_api_detector"]; got != "graph" {
		t.Errorf("workflow.autonomy.escalation.new_api_detector = %v, want graph written explicitly", got)
	}

	var typed struct {
		Workflow config.WorkflowConfig `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(raw, &typed); err != nil {
		t.Fatalf("unmarshal into WorkflowConfig: %v", err)
	}
	s := config.ResolveAutonomy(typed.Workflow)
	if s.NewAPIDetector != "graph" || len(s.Warnings) != 0 {
		t.Errorf("template resolves new_api_detector=%q warnings=%v", s.NewAPIDetector, s.Warnings)
	}

	// Neutrality of the autonomy block (comments included).
	text := string(raw)
	start := strings.Index(text, "    autonomy:")
	end := strings.Index(text[start:], "    workflow_agents:")
	if start < 0 || end < 0 {
		t.Fatal("autonomy block not found")
	}
	block := text[start : start+end]
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`\bt\d{2,5}\b`),            // card id
		regexp.MustCompile(`SPEC-[A-Z0-9-]+`),         // SPEC id
		regexp.MustCompile(`\b20\d\d-\d\d-\d\d\b`),    // date
		regexp.MustCompile(`\b[0-9a-f]{9,40}\b`),      // commit SHA
		regexp.MustCompile(`(?i)\bmoai escalation\b`), // undecided CLI verb
	} {
		if m := re.FindString(block); m != "" {
			t.Errorf("autonomy block carries %q (pattern %s)", m, re)
		}
	}
}
