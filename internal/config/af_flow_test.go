package config

// af_flow_test.go — card t2045 (R456): the flow is named AF-Flow and the
// config value is `af-flow`. The legacy value `agentic-trunk` is still READ
// and interpreted as `af-flow`; every other spelling stays invalid.

import "testing"

const afFlowPrefixLine = "        release_branch_prefix: release/\n"

// afFlowRouteLine is the release route's explicit opt-in (card t2098): the
// route arms only when release_route is true, never on a bare prefix.
const afFlowRouteLine = "        release_route: true\n"

func TestAFFlowAllowedValue(t *testing.T) {
	got := AllowedWorkflows()
	if len(got) != 5 {
		t.Fatalf("AllowedWorkflows() has %d entries, want 5: %v", len(got), got)
	}
	if got[4] != "af-flow" {
		t.Errorf("AllowedWorkflows()[4] = %q, want af-flow", got[4])
	}
	for _, v := range got {
		if v == "agentic-trunk" {
			t.Errorf("AllowedWorkflows() lists the legacy value %q; it is accepted on read, not listed", v)
		}
	}
}

func TestAFFlowAndLegacyClassifyAndTarget(t *testing.T) {
	for _, v := range []string{"af-flow", "agentic-trunk"} {
		t.Run(v, func(t *testing.T) {
			if d := ClassifyWorkflowDisposition(v); d != DispositionValidNonGitFlow {
				t.Errorf("ClassifyWorkflowDisposition(%q) = %q, want %q", v, d, DispositionValidNonGitFlow)
			}
			if tgt := WorkflowIntegrationTarget(v, ModeProfile{DevelopBranch: "develop"}); tgt != "main" {
				t.Errorf("WorkflowIntegrationTarget(%q) = %q, want main", v, tgt)
			}
		})
	}
}

func TestAFFlowLoaderNormalizesLegacy(t *testing.T) {
	for _, v := range []string{"af-flow", "agentic-trunk"} {
		t.Run(v, func(t *testing.T) {
			root := t.TempDir()
			dispositionWrite(t, root, gitStrategyBody("manual", v, afFlowPrefixLine))
			cfg := LoadGitFlowIntegrationConfig(root)
			if cfg.Workflow != "af-flow" {
				t.Errorf("Workflow = %q, want af-flow", cfg.Workflow)
			}
			if cfg.Disposition != DispositionValidNonGitFlow {
				t.Errorf("Disposition = %q, want %q", cfg.Disposition, DispositionValidNonGitFlow)
			}
			if cfg.IntegrationTarget != "main" {
				t.Errorf("IntegrationTarget = %q, want main", cfg.IntegrationTarget)
			}
			if !cfg.AgenticTrunkWorkflow || !cfg.UsesLocalIntegration() {
				t.Errorf("AgenticTrunkWorkflow=%v UsesLocalIntegration=%v, want both true", cfg.AgenticTrunkWorkflow, cfg.UsesLocalIntegration())
			}
			// Card t2098: the release route is opt-in, default OFF by operator
			// decision (the af-flow landing surface is local main) — a
			// non-empty release_branch_prefix alone arms nothing.
			if got := LoadAgenticTrunkReleaseBranch(root); got != "" {
				t.Errorf("LoadAgenticTrunkReleaseBranch = %q, want \"\" (release_route absent)", got)
			}
			// release_route: true is the explicit opt-in (card t2098); the
			// fixed release branch resolves only through it.
			dispositionWrite(t, root, gitStrategyBody("manual", v, afFlowPrefixLine+afFlowRouteLine))
			if got := LoadAgenticTrunkReleaseBranch(root); got != "release/main-batch" {
				t.Errorf("LoadAgenticTrunkReleaseBranch = %q, want release/main-batch (release_route: true)", got)
			}
		})
	}
}

func TestAFFlowWrongValuesInvalid(t *testing.T) {
	for _, v := range []string{"trunk-based", "af_flow", "AF-Flow", "agentic_trunk", "Agentic-Trunk", "af-flow ", ""} {
		if d := ClassifyWorkflowDisposition(v); d != DispositionInvalid {
			t.Errorf("ClassifyWorkflowDisposition(%q) = %q, want invalid", v, d)
		}
	}
}
