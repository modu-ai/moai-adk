package agentlint

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// The workflow.model_routing_profiles block is retired with the per-agent
// model assignment it described (SPEC-AGENT-MODEL-INHERIT-001 D12): a leftover
// block — even one the old closed sets would reject — is no longer linted.
func TestWorkflowLint_ModelRoutingBlockIsNoLongerChecked(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, "workflow:\n  model_routing_profiles:\n    max:\n      S-run:\n        model: haiku\n        effort: high\n")
	cmd := &cobra.Command{Use: "lint", RunE: runWorkflowLint}
	cmd.Flags().String("path", workflowFile, "")
	cmd.Flags().String("format", "text", "")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := runWorkflowLint(cmd, nil); err != nil {
		t.Fatalf("a leftover model_routing_profiles block must not be a lint violation, got: %v\n%s", err, out.String())
	}
}

// LR-03 (missing effort) and LR-12 (effort drift from the per-agent matrix)
// are retired: MoAI agents declare no effort and inherit the main session's
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-005, AC-AMI-005, design D3). LR-13 stays
// as a value check on the now-optional key.
func TestEffortRules_LR03AndLR12Retired(t *testing.T) {
	cases := map[string]string{
		"no effort":           "---\nname: manager-spec\ndescription: d\ntools: Read, Write\n---\nbody\n",
		"former matrix drift": "---\nname: manager-spec\ndescription: d\ntools: Read, Write\neffort: low\n---\nbody\n",
		"invalid effort":      "---\nname: manager-spec\ndescription: d\ntools: Read, Write\neffort: ultra\n---\nbody\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "manager-spec.md")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			violations, err := lintAgentFile(path, false)
			if err != nil {
				t.Fatal(err)
			}
			var lr13 bool
			for _, v := range violations {
				if v.Rule == "LR-03" || v.Rule == "LR-12" {
					t.Errorf("retired rule %s fired: %s", v.Rule, v.Message)
				}
				lr13 = lr13 || v.Rule == "LR-13"
			}
			if want := name == "invalid effort"; lr13 != want {
				t.Errorf("LR-13 fired = %v, want %v (it still validates a declared effort)", lr13, want)
			}
		})
	}
}
