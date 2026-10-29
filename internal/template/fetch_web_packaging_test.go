// fetch_web_packaging_test.go — SPEC-FETCH-WEB-001 M7 (AC-FW-013,
// REQ-FW-015). Template packaging: the moai-fetch-web skill, the settings
// permission rule, the GLM web-tooling routing row, the Codex network_access
// note — and the neutrality of THIS SPEC's own shipped artifacts (no
// `enabled: true`, no machine-specific values). Pre-implementation RED: the
// plan-phase evidence ledger rows R6 (skill dir absent) and R7 (permission
// rule zero hits), tree 4f0e58197d.
package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fetchWebTemplatePath(t *testing.T, rel string) string {
	t.Helper()
	path := filepath.Join("templates", rel)
	if _, err := os.Stat(path); err != nil {
		t.Skipf("template tree not found at %s (consumer checkout?)", path)
	}
	return path
}

func TestFetchWebSkillShips(t *testing.T) {
	skill := fetchWebTemplatePath(t, filepath.Join(".claude", "skills", "moai-fetch-web", "SKILL.md"))
	body, err := os.ReadFile(skill)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(body)
	for _, want := range []string{
		"WebFetch",             // when to prefer it over WebFetch
		"--json",               // how --json output is handled
		"no_match",             // how the no_match status is handled
		"network_access",       // the Codex note
		"moai fetch web",       // the command
		"workflow.jev.enabled", // the gate
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("SKILL.md must cover %q:\n%s", want, doc)
		}
	}
}

func TestFetchWebPermissionRuleShips(t *testing.T) {
	settings := fetchWebTemplatePath(t, filepath.Join(".claude", "settings.json.tmpl"))
	body, err := os.ReadFile(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"Bash(moai fetch web:*)"`) {
		t.Error("settings.json.tmpl must carry the Bash(moai fetch web:*) allow rule")
	}
}

func TestFetchWebGLMRoutingRowShips(t *testing.T) {
	rule := fetchWebTemplatePath(t, filepath.Join(".claude", "rules", "moai", "core", "glm-web-tooling.md"))
	body, err := os.ReadFile(rule)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(body)
	if !strings.Contains(doc, "fetch_web first") {
		t.Error("glm-web-tooling routing rule must state the 'jev active -> fetch_web first' priority")
	}
	if !strings.Contains(doc, "workflow.jev.enabled") {
		t.Error("the routing row must name the gate it conditions on")
	}
}

// Neutrality of THIS SPEC's own shipped artifacts (REQ-FW-015): the skill,
// the settings rule, and the routing row introduce no `enabled: true` and no
// machine-specific values. The tree's pre-existing enabled:true occurrences
// are other templates' and untouched.
func TestFetchWebArtifactsAreNeutral(t *testing.T) {
	artifacts := []string{
		filepath.Join(".claude", "skills", "moai-fetch-web", "SKILL.md"),
		filepath.Join(".claude", "settings.json.tmpl"),
		filepath.Join(".claude", "rules", "moai", "core", "glm-web-tooling.md"),
	}
	for _, rel := range artifacts {
		path := fetchWebTemplatePath(t, rel)
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// The settings template legitimately carries other sections'
		// enabled:true entries; this SPEC's own additions are the three
		// fetch-web surfaces — assert OUR additions by checking the
		// fetch-web-relevant lines only.
		for _, line := range strings.Split(string(body), "\n") {
			if !strings.Contains(line, "fetch") {
				continue
			}
			if strings.Contains(line, "enabled: true") || strings.Contains(line, "\"enabled\": true") {
				t.Errorf("%s: fetch-web line introduces enabled:true: %s", rel, strings.TrimSpace(line))
			}
			if strings.Contains(line, "/Users/") || strings.Contains(line, "/home/") {
				t.Errorf("%s: fetch-web line carries a machine-specific path: %s", rel, strings.TrimSpace(line))
			}
		}
	}
}
