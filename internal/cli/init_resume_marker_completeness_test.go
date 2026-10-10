package cli

// init_resume_marker_completeness_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// codex sync-audit finding 1: the init marker must hold against the deployment-
// completeness arm of initResumeCheckpoint.
//
// An init that dies after ApplyDeployMode wrote deployment_mode into llm.yaml, but
// before the MCP and Codex wiring finished, leaves the init marker standing. The
// completeness arm read deployment_mode as "this project is complete" and refused the
// rerun with "already initialized". The marker is the project-scoped interruption
// evidence, so it must route the rerun into the resume.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInitResumeCheckpointInitMarkerOverridesCompleteness reproduces sync-audit finding 1.
// The homeDir carries no pending journal and no user manifest, so the init marker is the
// only interruption evidence in play. The two controls pin the behaviour that must not move:
// a healthy project keeps the original redirect, and a project the executor never deployed
// keeps it too (the llm.yaml presence check stays ahead of the marker override).
func TestInitResumeCheckpointInitMarkerOverridesCompleteness(t *testing.T) {
	writeLLM := func(t *testing.T, projectRoot, body string) {
		t.Helper()
		sections := filepath.Join(projectRoot, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("interrupted_init_resumes_despite_deployment_mode", func(t *testing.T) {
		homeDir := t.TempDir()
		projectRoot := t.TempDir()
		if err := persistInitPending(projectRoot, nil); err != nil {
			t.Fatal(err)
		}
		// ApplyDeployMode ran (deployment_mode present); the marker still stands.
		writeLLM(t, projectRoot, "harness: claude\ndeployment_mode: auto\n")
		if !initResumeCheckpoint(homeDir, projectRoot) {
			t.Fatal("an init marker that stands must route the rerun into the resume, " +
				"even when llm.yaml carries deployment_mode: initResumeCheckpoint reported not resumable")
		}
	})

	t.Run("control_healthy_project_without_marker_keeps_redirect", func(t *testing.T) {
		homeDir := t.TempDir()
		projectRoot := t.TempDir()
		writeLLM(t, projectRoot, "harness: claude\ndeployment_mode: auto\n")
		if initResumeCheckpoint(homeDir, projectRoot) {
			t.Fatal("a complete project with no interruption evidence was routed into resume")
		}
	})

	t.Run("control_marker_without_llm_yaml_keeps_redirect", func(t *testing.T) {
		homeDir := t.TempDir()
		projectRoot := t.TempDir()
		if err := persistInitPending(projectRoot, nil); err != nil {
			t.Fatal(err)
		}
		if initResumeCheckpoint(homeDir, projectRoot) {
			t.Fatal("a project the executor never deployed was routed into resume by the init marker " +
				"alone: the llm.yaml presence check must stay ahead of the marker override")
		}
	})
}
