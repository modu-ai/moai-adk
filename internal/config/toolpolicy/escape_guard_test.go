package toolpolicy

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// escapeGuardMutate copies the committed policy into a temp dir, rewrites the
// drive-root rule with an escaped colon, and regenerates the settings copy
// from the mutated policy so the two stay set-equal. It returns the mutated
// paths and the escaped specifier it introduced.
func escapeGuardMutate(t *testing.T) (yamlPath, settingsPath, escaped string) {
	t.Helper()
	committedYAML, committedSettings := driftCommittedPaths(t)
	yamlPath, settingsPath = driftCopyCommitted(t, committedYAML, committedSettings)

	body, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read yaml copy: %v", err)
	}
	const clean = `args_pattern: "rm -rf C:/:*"`
	if !bytes.Contains(body, []byte(clean)) {
		t.Fatalf("committed policy lacks %s; the mutation has nothing to rewrite", clean)
	}
	mutated := bytes.Replace(body, []byte(clean), []byte(`args_pattern: "rm -rf C\\:/:*"`), 1)
	driftWriteFixture(t, filepath.Dir(yamlPath), filepath.Base(yamlPath), string(mutated))

	doc, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("load mutated yaml: %v", err)
	}
	block, _, err := BuildPermissions(doc, "")
	if err != nil {
		t.Fatalf("build permissions from mutated yaml: %v", err)
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	perms := map[string]any{"allow": block.Allow, "deny": block.Deny}
	if len(block.Ask) > 0 {
		perms["ask"] = block.Ask
	}
	if err := enc.Encode(map[string]any{"permissions": perms}); err != nil {
		t.Fatalf("encode regenerated settings: %v", err)
	}
	driftWriteFixture(t, filepath.Dir(settingsPath), filepath.Base(settingsPath), out.String())
	return yamlPath, settingsPath, `Bash(rm -rf C\:/:*)`
}

// escapedColonSpecifiers returns every Bash or PowerShell specifier in the policy at path
// whose pattern contains `\:`. Claude Code matches that backslash literally,
// so such a rule never denies a real drive path such as "C:/". Env-gated
// entries are included: they are still rules that a build can emit.
func escapedColonSpecifiers(t *testing.T, path string) []string {
	t.Helper()
	doc, err := Load(path)
	if err != nil {
		t.Fatalf("load policy %q: %v", path, err)
	}
	var found []string
	for _, e := range doc.Entries {
		if (e.Tool == "Bash" || e.Tool == "PowerShell") && strings.Contains(e.ArgsPattern, `\:`) {
			found = append(found, e.SettingsSpecifier())
		}
	}
	return found
}

// TestToolPolicyEscapeGuard_CommittedPolicy is the source-layer twin of the
// template guard in internal/template: the policy that generates the local
// settings must not carry an escaped colon either.
func TestToolPolicyEscapeGuard_CommittedPolicy(t *testing.T) {
	yamlPath, _ := driftCommittedPaths(t)
	if found := escapedColonSpecifiers(t, yamlPath); len(found) > 0 {
		t.Errorf("tool-policy.yaml Bash/PowerShell rules escape ':' and cannot match a real drive path: %q", found)
	}
}

// TestToolPolicyEscapeGuard_DetectsMutation proves the guard goes red on the
// exact input the drift check lets through.
func TestToolPolicyEscapeGuard_DetectsMutation(t *testing.T) {
	yamlPath, _, escaped := escapeGuardMutate(t)
	found := escapedColonSpecifiers(t, yamlPath)
	if len(found) != 1 || found[0] != escaped {
		t.Errorf("guard on the mutated policy = %q, want exactly [%q]", found, escaped)
	}
}

// TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt reproduces the gap: an
// escaped colon added to the policy and regenerated into settings leaves the
// two set-equal, so the drift check reports nothing.
func TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt(t *testing.T) {
	yamlPath, settingsPath, escaped := escapeGuardMutate(t)

	diff, err := driftSetDiff(yamlPath, settingsPath)
	if err != nil {
		t.Fatalf("driftSetDiff: %v", err)
	}
	if len(diff) != 0 {
		t.Fatalf("drift check reported %q; the reproduction expects set-equal inputs", diff)
	}
	settings, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read regenerated settings: %v", err)
	}
	if !strings.Contains(string(settings), strings.ReplaceAll(escaped, `\`, `\\`)) {
		t.Fatalf("regenerated settings do not carry %q", escaped)
	}
}

// TestToolPolicyEscapeGuard_DetectsPowerShellMutation proves the source-layer
// guard also covers the PowerShell namespace: an escaped colon in a
// PowerShell rule is as dead as one in a Bash rule.
func TestToolPolicyEscapeGuard_DetectsPowerShellMutation(t *testing.T) {
	committedYAML, committedSettings := driftCommittedPaths(t)
	yamlPath, _ := driftCopyCommitted(t, committedYAML, committedSettings)
	body, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read yaml copy: %v", err)
	}
	entry := "  - tool: \"PowerShell\"\n" +
		"    args_pattern: \"rm -rf C\\\\:/:*\"\n" +
		"    risk_tier: irreversible\n" +
		"    decision: deny\n" +
		"    owner_agent: orchestrator\n" +
		"    audit: \"escape guard mutation\"\n"
	driftWriteFixture(t, filepath.Dir(yamlPath), filepath.Base(yamlPath), string(body)+entry)

	want := `PowerShell(rm -rf C\:/:*)`
	found := escapedColonSpecifiers(t, yamlPath)
	if len(found) != 1 || found[0] != want {
		t.Errorf("guard on the mutated policy = %q, want exactly [%q]", found, want)
	}
}
