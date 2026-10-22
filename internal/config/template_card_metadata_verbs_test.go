package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestTemplateCardMetadataVerbsShippedOff — SPEC-CARD-METADATA-VERBS-001 M5
// (AC-CMV-002): the shipped template workflow.yaml carries the
// card_metadata_verbs key as a COMMENTED-OUT block (the documented opt-in
// surface plan.md §F M5 prescribes), and parsed as shipped it is never true
// by default.
func TestTemplateCardMetadataVerbsShippedOff(t *testing.T) {
	repoRoot := findRepoRootFromCaller(t)
	templatePath := filepath.Join(repoRoot,
		"internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", templatePath, err)
	}

	// The M5 deliverable: the commented default-off key block is present.
	// (An absent key would satisfy "never true by default" vacuously — the
	// block's presence is what the milestone ships.)
	present := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "card_metadata_verbs:") {
			present = true
			break
		}
	}
	if !present {
		t.Fatal("the shipped template workflow.yaml carries no commented workflow.card_metadata_verbs block (plan.md §F M5)")
	}

	// AC-CMV-002: parsed as shipped, the key is absent or explicitly false —
	// never true by default.
	var doc struct {
		Workflow struct {
			CardMetadataVerbs struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"card_metadata_verbs"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("yaml.Unmarshal(%q): %v", templatePath, err)
	}
	if doc.Workflow.CardMetadataVerbs.Enabled {
		t.Fatal("the shipped template enables workflow.card_metadata_verbs by default; it must ship OFF (REQ-CMV-002)")
	}
}
