package cli

// SPEC-INIT-HARNESS-001 M1 — persistence of the resolved harness selection
// (AC-IH-001, REQ-IH-002): every init run — interactive or not, all three
// closed-set values INCLUDING the claude default — writes llm.harness to
// .moai/config/sections/llm.yaml. Explicit record over implicit absence: an
// absent key must not force doctor/update to infer "claude".

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"gopkg.in/yaml.v3"
)

// readLLMHarness loads the deployed llm.yaml and returns its llm.harness value
// ("" when the key is absent). It unmarshals into a map so the assertion is on
// the shipped file's key, not on any Go-side default fallback.
func readLLMHarness(t *testing.T, projectDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("read deployed llm.yaml: %v", err)
	}
	var doc map[string]map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse deployed llm.yaml: %v\nraw:\n%s", err, data)
	}
	llm, ok := doc["llm"]
	if !ok {
		return ""
	}
	h, _ := llm["harness"].(string)
	return h
}

// TestInitPersistsHarnessKey verifies AC-IH-001: the resolved harness value is
// written to llm.harness on every init run, for all three closed-set values
// AND for the flag-absent default (explicit record of claude included).
func TestInitPersistsHarnessKey(t *testing.T) {
	cases := []struct {
		name string
		flag string
		want string
	}{
		{"flag absent records claude", "", "claude"},
		{"claude records claude", "claude", "claude"},
		{"gpt records gpt", "gpt", "gpt"},
		{"both records both", "both", "both"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var flags map[string]string
			if tc.flag != "" {
				flags = map[string]string{"llm": tc.flag}
			}
			projectDir, _ := runInitForAutonomy(t, nil, flags)

			if got := readLLMHarness(t, projectDir); got != tc.want {
				t.Errorf("llm.harness after init = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestInitPersistsHarnessKeyFromWizard verifies the wizard path records the
// same key (REQ-IH-001: one resolution, two input sources — the persistence
// consumer must read the RESOLVED value either way). Flag absent, wizard
// answers codex.
func TestInitPersistsHarnessKeyFromWizard(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "gpt"}
	projectDir, _ := runInitForAutonomy(t, wiz, nil)

	if got := readLLMHarness(t, projectDir); got != "gpt" {
		t.Errorf("llm.harness after wizard gpt init = %q, want %q", got, "gpt")
	}
}
