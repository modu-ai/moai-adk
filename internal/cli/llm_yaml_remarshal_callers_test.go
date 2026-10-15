package cli

// llm_yaml_remarshal_callers_test.go — the llm.yaml re-marshal contract for each
// real caller of persistTeamMode (SPEC-UPDATE-REPAIR-RC30-001 AC-RC30-23, U-24;
// REQ-RC30-007). persistTeamMode writes llm.yaml through saveLLMSection, which
// re-marshals the typed config.LLMConfig. That type has no field for
// deployment_mode or for an unmodelled key, and a typed marshal keeps no
// comments, so a caller that performs a real write drops all three.
//
// The two real callers are the cc reset path (resetTeamModeForCC, launcher.go,
// reached on every moai cc launch) and the glm path (persistTeamMode(root,
// "glm"), reached on every moai glm launch). Each subtest starts from the same
// llm.yaml and asserts that the header comment, the deployment_mode record, and
// the unmodelled key all survive the call. This test is the RED that precedes
// the saveLLMSection fix (GREEN 1(a)); it does not change saveLLMSection.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// llmRemarshalHeader is the header comment the user keeps above llm.yaml.
const llmRemarshalHeader = "# T1619-LLM-REMARSHAL-HEADER: user note above the llm section"

// writeRemarshalLLMYAML writes the caller fixture: the header comment, then an
// llm section holding the team_mode under test, the deployment_mode record that
// ApplyDeployMode nests under llm, and an unmodelled key that LLMConfig does not
// carry.
func writeRemarshalLLMYAML(t *testing.T, root, teamMode string) {
	t.Helper()
	src := llmRemarshalHeader + "\n" +
		"llm:\n" +
		fmt.Sprintf("  team_mode: %q\n", teamMode) +
		"  deployment_mode: plugin\n" +
		"  future_flag: keep-me\n"
	writeTestFile(t, root, ".moai/config/sections/llm.yaml", src)
}

// assertRemarshalKeepsKeys reads the llm.yaml a caller left behind and fails
// with the caller's name and the key or comment it lost.
func assertRemarshalKeepsKeys(t *testing.T, caller, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("caller %s: read llm.yaml: %v", caller, err)
	}
	raw := string(data)
	var doc struct {
		LLM map[string]any `yaml:"llm"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("caller %s: parse llm.yaml: %v\n%s", caller, err, raw)
	}
	if got := doc.LLM["deployment_mode"]; got != "plugin" {
		t.Errorf("caller %s dropped llm.deployment_mode (want %q, got %v); llm.yaml after:\n%s", caller, "plugin", got, raw)
	}
	if got := doc.LLM["future_flag"]; got != "keep-me" {
		t.Errorf("caller %s dropped llm.future_flag (want %q, got %v); llm.yaml after:\n%s", caller, "keep-me", got, raw)
	}
	if !strings.Contains(raw, llmRemarshalHeader) {
		t.Errorf("caller %s dropped the header comment %q; llm.yaml after:\n%s", caller, llmRemarshalHeader, raw)
	}
}

// TestLLMYamlRemarshalCallers is AC-RC30-23: each real caller of persistTeamMode
// keeps every key and the header comment of llm.yaml. A failing run names the
// caller and the key it lost.
func TestLLMYamlRemarshalCallers(t *testing.T) {
	callers := []struct {
		name     string
		teamMode string
		call     func(t *testing.T, root string)
	}{
		{
			// The moai cc launch. team_mode is set so that the reset actually runs.
			name:     "resetTeamModeForCC",
			teamMode: "glm",
			call: func(t *testing.T, root string) {
				t.Helper()
				if msg := resetTeamModeForCC(root); !strings.Contains(msg, "Team mode disabled") {
					t.Fatalf("resetTeamModeForCC did not reset team_mode (returned %q); the case would pass without a write", msg)
				}
			},
		},
		{
			// The moai glm launch. team_mode starts unset so that the write runs.
			// A file already at team_mode "glm" hits the no-rewrite gate and would
			// pass without writing.
			name:     "persistTeamMode_glm",
			teamMode: "",
			call: func(t *testing.T, root string) {
				t.Helper()
				if err := persistTeamMode(root, "glm"); err != nil {
					t.Fatalf("persistTeamMode(root, \"glm\"): %v", err)
				}
			},
		},
	}
	for _, c := range callers {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			writeRemarshalLLMYAML(t, root, c.teamMode)
			c.call(t, root)
			assertRemarshalKeepsKeys(t, c.name, root)
		})
	}
}
