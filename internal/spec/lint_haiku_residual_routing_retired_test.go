package spec

import "testing"

// The two routing surfaces of HaikuResidualRule — workflow.yaml routing blocks
// and internal/config/model_routing.go — are retired with the per-agent model
// routing they policed (SPEC-AGENT-MODEL-INHERIT-001 M4, design D12). A haiku
// mention on either no longer yields a finding; the agent-definition and
// claude_models surfaces still do.
func TestHaikuResidualRule_RoutingSurfacesRetired(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeHaikuFixtureFile(t, root, ".moai/config/sections/workflow.yaml",
		"workflow:\n    workflow_agents:\n        read-only-extract: { model: haiku, effort: low }\n")
	writeHaikuFixtureFile(t, root, "internal/config/model_routing.go",
		"package config\nvar validRoutingModels = map[string]bool{\"haiku\": true}\n")
	if findings := (&HaikuResidualRule{baseDir: root}).CheckAll(nil); len(findings) != 0 {
		t.Errorf("retired routing surfaces still yield findings: %+v", findings)
	}

	// Positive control: an agent definition surface still fires.
	writeHaikuFixtureFile(t, root, ".claude/agents/moai/some-agent.md", "---\nname: some-agent\nmodel: haiku\n---\n")
	if findings := (&HaikuResidualRule{baseDir: root}).CheckAll(nil); len(findings) != 1 {
		t.Errorf("agent-definition surface: got %d findings, want 1: %+v", len(findings), findings)
	}
}
