package config

// loader_workflow_disposition_test.go — SPEC-GITSTRAT-WORKFLOW-READER-001 M2
// (card t656).
//
// The multi-flow validation and interpretation-table tests. This file is the
// M2 owner BY CONTRACT (plan.md §F M2): the M1 characterization suite stays
// the sole owner of loader_integration_branch_test.go, which is what keeps
// AC-GWS-010's pinned-SHA diff executable.

import (
	"os"
	"path/filepath"
	"testing"
)

// dispositionWrite writes a git-strategy.yaml fixture under projectRoot. An
// empty body means "write no file at all" (the missing-file case).
func dispositionWrite(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		return
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// gitStrategyBody renders a manual-mode profile fixture with the given
// workflow value and extra profile lines appended under the profile key.
func gitStrategyBody(profileKey, workflow, extraLines string) string {
	body := "git_strategy:\n    mode: " + profileKey + "\n    " + profileKey + ":\n        workflow: " + workflow + "\n"
	if extraLines != "" {
		body += extraLines
	}
	return body
}

// TestWorkflowDisposition pins the three-way disposition (REQ-GWS-001/002):
// exactly the 4 allowed values classify, everything else — typos,
// trunk-based, empty — is invalid, and the raw offending value is carried
// through the reader seam.
func TestWorkflowDisposition(t *testing.T) {
	t.Run("pure classification of the allowed set", func(t *testing.T) {
		tests := []struct {
			value string
			want  WorkflowDisposition
		}{
			{value: "github-flow", want: DispositionValidNonGitFlow},
			{value: "git-flow", want: DispositionGitFlow},
			{value: "gitlab-flow", want: DispositionValidNonGitFlow},
			{value: "release-flow", want: DispositionValidNonGitFlow},
			// Deliberately excluded (operator decision): unsupported, must be
			// rejected/diagnosed, not tolerated.
			{value: "trunk-based", want: DispositionInvalid},
			{value: "git-flwo", want: DispositionInvalid}, // typo
			{value: "", want: DispositionInvalid},         // empty value present
			{value: "Git-Flow", want: DispositionInvalid}, // exact match only — the pre-existing discriminator was case-sensitive
		}
		for _, tt := range tests {
			if got := ClassifyWorkflowDisposition(tt.value); got != tt.want {
				t.Errorf("ClassifyWorkflowDisposition(%q) = %q, want %q", tt.value, got, tt.want)
			}
		}
	})

	t.Run("allowed set is exactly four flows", func(t *testing.T) {
		got := AllowedWorkflows()
		if len(got) != 4 {
			t.Fatalf("AllowedWorkflows() has %d entries, want 4: %v", len(got), got)
		}
		set := map[string]bool{}
		for _, v := range got {
			set[v] = true
		}
		for _, want := range []string{"github-flow", "git-flow", "gitlab-flow", "release-flow"} {
			if !set[want] {
				t.Errorf("AllowedWorkflows() is missing %q: %v", want, got)
			}
		}
	})

	// AC-GWS-002: the reader seam carries the offending raw value and reports
	// invalid — a typo is diagnosable, not silently "not git-flow".
	t.Run("typo fixture carries the raw value", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", "git-flwo", ""))
		got := LoadGitFlowIntegrationConfig(root)
		if got.Disposition != DispositionInvalid {
			t.Errorf("Disposition = %q, want %q", got.Disposition, DispositionInvalid)
		}
		if got.Workflow != "git-flwo" {
			t.Errorf("Workflow raw = %q, want git-flwo", got.Workflow)
		}
		if got.IsGitFlow() {
			t.Error("IsGitFlow() = true on an invalid workflow value")
		}
		if got.IntegrationTarget != "" {
			t.Errorf("IntegrationTarget = %q, want empty on invalid", got.IntegrationTarget)
		}
	})

	// AC-GWS-003: trunk-based exclusion is enforced by omission — invalid.
	t.Run("trunk-based fixture is invalid", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", "trunk-based", ""))
		got := LoadGitFlowIntegrationConfig(root)
		if got.Disposition != DispositionInvalid {
			t.Errorf("Disposition = %q, want %q (trunk-based is deliberately excluded)", got.Disposition, DispositionInvalid)
		}
		if got.Workflow != "trunk-based" {
			t.Errorf("Workflow raw = %q, want trunk-based", got.Workflow)
		}
	})

	t.Run("each allowed value classifies through the reader seam", func(t *testing.T) {
		tests := []struct {
			workflow string
			want     WorkflowDisposition
		}{
			{workflow: "github-flow", want: DispositionValidNonGitFlow},
			{workflow: "git-flow", want: DispositionGitFlow},
			{workflow: "gitlab-flow", want: DispositionValidNonGitFlow},
			{workflow: "release-flow", want: DispositionValidNonGitFlow},
		}
		for _, tt := range tests {
			root := t.TempDir()
			dispositionWrite(t, root, gitStrategyBody("manual", tt.workflow, ""))
			got := LoadGitFlowIntegrationConfig(root)
			if got.Disposition != tt.want {
				t.Errorf("workflow %q: Disposition = %q, want %q", tt.workflow, got.Disposition, tt.want)
			}
			if got.Workflow != tt.workflow {
				t.Errorf("workflow %q: raw Workflow = %q", tt.workflow, got.Workflow)
			}
		}
	})

	// The zero value is NOT invalid: an unreadable file or a missing active
	// profile is "unknown" (acceptance.md §D.3), which the doctor check must
	// distinguish from "value invalid".
	t.Run("no active profile stays unknown, not invalid", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, "git_strategy:\n    manual:\n        workflow: git-flow\n")
		got := LoadGitFlowIntegrationConfig(root)
		if got.Disposition != "" {
			t.Errorf("Disposition = %q, want the zero value (unknown) when no active profile exists", got.Disposition)
		}
		if got.Workflow != "" {
			t.Errorf("Workflow = %q, want empty when no active profile exists", got.Workflow)
		}
	})
}

// TestWorkflowTargetResolution pins the D2 flow-scoped integration-target
// table (REQ-GWS-004/006): each flow reads exactly one target source.
func TestWorkflowTargetResolution(t *testing.T) {
	t.Run("github-flow", func(t *testing.T) {
		// AC-GWS-005: the fixed default main, without reading any key — a
		// stale develop_branch must not leak into the resolution.
		profile := ModeProfile{DevelopBranch: "stale-develop", Environment: "staging"}
		if got := WorkflowIntegrationTarget("github-flow", profile); got != "main" {
			t.Errorf("github-flow target = %q, want main", got)
		}
	})

	t.Run("git-flow", func(t *testing.T) {
		// AC-GWS-006: develop_branch, trimmed.
		profile := ModeProfile{DevelopBranch: "  develop  "}
		if got := WorkflowIntegrationTarget("git-flow", profile); got != "develop" {
			t.Errorf("git-flow target = %q, want develop (trimmed)", got)
		}
	})

	t.Run("gitlab-flow", func(t *testing.T) {
		profile := ModeProfile{Environment: "staging"}
		if got := WorkflowIntegrationTarget("gitlab-flow", profile); got != "staging" {
			t.Errorf("gitlab-flow target = %q, want staging", got)
		}
	})

	// AC-GWS-007: empty environment key yields the caller-fallback neutral.
	t.Run("gitlab-flow-empty", func(t *testing.T) {
		profile := ModeProfile{Environment: ""}
		if got := WorkflowIntegrationTarget("gitlab-flow", profile); got != "" {
			t.Errorf("gitlab-flow empty-environment target = %q, want empty", got)
		}
	})

	// AC-GWS-008: release_branch_prefix flows through, trimmed.
	t.Run("release-flow", func(t *testing.T) {
		profile := ModeProfile{ReleaseBranchPrefix: " release/"}
		if got := WorkflowIntegrationTarget("release-flow", profile); got != "release/" {
			t.Errorf("release-flow target = %q, want release/ (trimmed)", got)
		}
	})

	t.Run("invalid value resolves no target", func(t *testing.T) {
		profile := ModeProfile{DevelopBranch: "develop", Environment: "staging", ReleaseBranchPrefix: "release/"}
		if got := WorkflowIntegrationTarget("git-flwo", profile); got != "" {
			t.Errorf("invalid workflow target = %q, want empty", got)
		}
	})
}
