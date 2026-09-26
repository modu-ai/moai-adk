package toolpolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// PowerShell deny parity between tool-policy.yaml and the distributed
// settings template. The drift check compares the YAML with the local
// .claude/settings.json only; the generator's --local-only flow updates both
// of those and leaves the template untouched, so a PowerShell deny added to
// or dropped from the YAML could ship without reaching the template. This
// guard closes that gap for the PowerShell namespace.

// psYAMLDenySet returns the settings specifiers of every non-env-gated
// PowerShell deny entry in the policy (env-gated entries are skipped, as the
// generator skips them).
func psYAMLDenySet(doc *PolicyDocument) map[string]bool {
	set := map[string]bool{}
	for _, e := range doc.Entries {
		if e.Tool == "PowerShell" && e.Decision == DecisionDeny && e.EnvGate == nil {
			set[e.SettingsSpecifier()] = true
		}
	}
	return set
}

// psSettingsDenySet returns every "PowerShell(" row of the deny list in a
// settings body. Go-template directive lines inside the permissions region
// (the .tmpl carries {{- if ...}} lines in its allow list) are dropped before
// the region is decoded, so the same reader serves the template and the
// rendered local settings.
func psSettingsDenySet(t *testing.T, body []byte) map[string]bool {
	t.Helper()
	region, err := locatePermissionsRegion(body)
	if err != nil {
		t.Fatalf("locate permissions region: %v", err)
	}
	var kept []string
	for _, line := range strings.Split(string(body[region.start:region.end]), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "{{") {
			continue
		}
		kept = append(kept, line)
	}
	var wrapper struct {
		Permissions struct {
			Deny []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal([]byte("{"+strings.Join(kept, "\n")+"}"), &wrapper); err != nil {
		t.Fatalf("decode permissions region: %v", err)
	}
	set := map[string]bool{}
	for _, r := range wrapper.Permissions.Deny {
		if strings.HasPrefix(r, "PowerShell(") {
			set[r] = true
		}
	}
	return set
}

// psSetDiff lists the specifiers present in exactly one of the two sets,
// labelled by side, in sorted order.
func psSetDiff(yamlSet, settingsSet map[string]bool, settingsLabel string) []string {
	var out []string
	for s := range yamlSet {
		if !settingsSet[s] {
			out = append(out, "only-in-yaml: "+s)
		}
	}
	for s := range settingsSet {
		if !yamlSet[s] {
			out = append(out, "only-in-"+settingsLabel+": "+s)
		}
	}
	sort.Strings(out)
	return out
}

// psParityInputs loads the committed policy and the template body, and
// returns the local settings path.
func psParityInputs(t *testing.T) (*PolicyDocument, []byte, string) {
	t.Helper()
	yamlPath, settingsPath := driftCommittedPaths(t)
	doc, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("load policy: %v", err)
	}
	root := filepath.Dir(filepath.Dir(settingsPath))
	tmpl, err := os.ReadFile(TemplateSettingsPath(root))
	if err != nil {
		t.Fatalf("read settings template: %v", err)
	}
	return doc, tmpl, settingsPath
}

// TestToolPolicyPowerShellTemplateParity_Committed asserts the YAML's
// PowerShell deny set equals the template's PowerShell deny rows, and equals
// the local settings' rows as well.
func TestToolPolicyPowerShellTemplateParity_Committed(t *testing.T) {
	doc, tmpl, settingsPath := psParityInputs(t)
	yamlSet := psYAMLDenySet(doc)
	if len(yamlSet) == 0 {
		t.Fatal("tool-policy.yaml declares no PowerShell deny entries; the parity guard would compare nothing")
	}
	if diff := psSetDiff(yamlSet, psSettingsDenySet(t, tmpl), "template"); len(diff) > 0 {
		t.Errorf("PowerShell deny rules differ between tool-policy.yaml and settings.json.tmpl (%d):\n%s\n"+
			"add or remove the same rows in internal/template/templates/.claude/settings.json.tmpl",
			len(diff), strings.Join(diff, "\n"))
	}
	local, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read local settings: %v", err)
	}
	if diff := psSetDiff(yamlSet, psSettingsDenySet(t, local), "local"); len(diff) > 0 {
		t.Errorf("PowerShell deny rules differ between tool-policy.yaml and .claude/settings.json (%d):\n%s",
			len(diff), strings.Join(diff, "\n"))
	}
}

// TestToolPolicyPowerShellTemplateParity_DetectsMutation proves the guard goes
// red on the two drifts the local-only drift check lets through: a PowerShell
// deny added to the YAML only, and one removed from the YAML only.
func TestToolPolicyPowerShellTemplateParity_DetectsMutation(t *testing.T) {
	doc, tmpl, _ := psParityInputs(t)
	tmplSet := psSettingsDenySet(t, tmpl)

	t.Run("yaml_adds_rule", func(t *testing.T) {
		yamlSet := psYAMLDenySet(doc)
		yamlSet["PowerShell(wipefs:*)"] = true
		got := psSetDiff(yamlSet, tmplSet, "template")
		want := []string{"only-in-yaml: PowerShell(wipefs:*)"}
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("diff = %q, want %q", got, want)
		}
	})
	t.Run("yaml_drops_rule", func(t *testing.T) {
		const dropped = "PowerShell(git push -f:*)"
		yamlSet := psYAMLDenySet(doc)
		if !yamlSet[dropped] {
			t.Fatalf("committed policy lacks %s; the mutation has nothing to drop", dropped)
		}
		delete(yamlSet, dropped)
		got := psSetDiff(yamlSet, tmplSet, "template")
		want := []string{"only-in-template: " + dropped}
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("diff = %q, want %q", got, want)
		}
	})
}
