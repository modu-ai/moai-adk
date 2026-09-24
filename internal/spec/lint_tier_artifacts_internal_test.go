package spec

// lint_tier_artifacts_internal_test.go — drift and deployment guards for the
// Tier artifact-set parser (card t1121). The rule reads its required sets from
// spec-workflow.md at lint time, so a table edit that breaks the parser would
// silently change what the lint enforces; these tests redden CI instead.

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

var wantTierArtifactSets = map[string][]string{
	"S": {"spec.md", "plan.md"},
	"M": {"spec.md", "plan.md", "acceptance.md"},
	"L": {"spec.md", "plan.md", "acceptance.md", "design.md", "research.md"},
}

// TestTierArtifactSets_ShippedTemplateDrift parses the REAL shipped template
// rule on disk and pins the three sets.
func TestTierArtifactSets_ShippedTemplateDrift(t *testing.T) {
	path := filepath.Join("..", "template", "templates", filepath.FromSlash(tierArtifactRuleRelPath))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read shipped rule %s: %v", path, err)
	}
	got, ok := parseTierArtifactSets(string(data))
	if !ok {
		t.Fatalf("shipped rule %s: Tier table did not parse", path)
	}
	if !reflect.DeepEqual(got, wantTierArtifactSets) {
		t.Fatalf("shipped Tier sets drifted:\n got %v\nwant %v", got, wantTierArtifactSets)
	}
}

// TestTierArtifactSets_DeployedInEmbeddedTemplates proves user projects
// receive the SSOT the rule reads: the rule file is present in the embedded
// template filesystem `moai init` / `moai update` deploy, at the same relative
// path the rule resolves against the project root, and that embedded copy
// parses to the same three sets.
func TestTierArtifactSets_DeployedInEmbeddedTemplates(t *testing.T) {
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	data, err := fs.ReadFile(fsys, tierArtifactRuleRelPath)
	if err != nil {
		t.Fatalf("embedded templates lack %s: %v", tierArtifactRuleRelPath, err)
	}
	got, ok := parseTierArtifactSets(string(data))
	if !ok {
		t.Fatalf("embedded %s: Tier table did not parse", tierArtifactRuleRelPath)
	}
	if !reflect.DeepEqual(got, wantTierArtifactSets) {
		t.Fatalf("embedded Tier sets drifted:\n got %v\nwant %v", got, wantTierArtifactSets)
	}
}

// TestParseTierArtifactSets_Strict covers the parser's refusal shapes: a
// table missing one tier row, and a row whose artifact cell names no file,
// both report not-ok rather than a partial map.
func TestParseTierArtifactSets_Strict(t *testing.T) {
	header := "| Tier | Scope | Files | Artifact set | Threshold |\n|---|---|---|---|---|\n"
	cases := map[string]string{
		"no table":    "# nothing here\n",
		"missing L":   header + "| S (Simple) | a | b | spec.md + plan.md | 0.75 |\n| M (Medium) | a | b | spec.md + plan.md + acceptance.md | 0.80 |\n",
		"empty cell":  header + "| S (Simple) | a | b | two files | 0.75 |\n| M (Medium) | a | b | spec.md | 0.80 |\n| L (Large) | a | b | spec.md | 0.85 |\n",
		"no artifact": "| Tier | Scope |\n|---|---|\n| S (Simple) | spec.md |\n| M (Medium) | spec.md |\n| L (Large) | spec.md |\n",
	}
	for name, body := range cases {
		if got, ok := parseTierArtifactSets(body); ok {
			t.Errorf("%s: parsed ok = true, want false (got %v)", name, got)
		}
	}
}
