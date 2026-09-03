package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// writeContinuationFixture lays down a sections dir carrying body as
// workflow.yaml (or nothing when body is empty) and returns the sections dir.
func writeContinuationFixture(t *testing.T, body string) string {
	t.Helper()
	sectionsDir := filepath.Join(t.TempDir(), defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(sectionsDir, defs.WorkflowYAML), []byte(body), 0o644); err != nil {
			t.Fatalf("write workflow.yaml: %v", err)
		}
	}
	return sectionsDir
}

// TestWriteWorkflowProjectContinuationYAML covers the M4 persistence half: the
// answer lands at workflow.project.continuation through yamlpatch, an unasked
// question writes nothing, and the no-deployer path creates a minimal block.
func TestWriteWorkflowProjectContinuationYAML(t *testing.T) {
	t.Run("unasked question leaves the file byte-identical", func(t *testing.T) {
		body := "workflow:\n    project:\n        continuation: card\n"
		sectionsDir := writeContinuationFixture(t, body)
		path := filepath.Join(sectionsDir, defs.WorkflowYAML)

		result := &InitResult{}
		if err := writeWorkflowProjectContinuationYAML(sectionsDir, InitOptions{}, result); err != nil {
			t.Fatalf("write: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read back: %v", err)
		}
		if string(got) != body {
			t.Errorf("unasked question mutated the file:\ngot  %q\nwant %q", string(got), body)
		}
	})

	t.Run("answer patches the existing key", func(t *testing.T) {
		sectionsDir := writeContinuationFixture(t, "workflow:\n    project:\n        continuation: card\n    default_mode: \"\"\n")
		path := filepath.Join(sectionsDir, defs.WorkflowYAML)

		opts := InitOptions{ProjectContinuation: config.ProjectContinuationPipeline}
		if err := writeWorkflowProjectContinuationYAML(sectionsDir, opts, &InitResult{}); err != nil {
			t.Fatalf("write: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read back: %v", err)
		}
		if !strings.Contains(string(got), "continuation: pipeline") {
			t.Errorf("patched file does not carry continuation: pipeline:\n%s", string(got))
		}
		// The neighbouring key survives — yamlpatch is a patch, not a rewrite.
		if !strings.Contains(string(got), "default_mode") {
			t.Errorf("patch destroyed the neighbouring default_mode key:\n%s", string(got))
		}
	})

	t.Run("upserts into a workflow.yaml carrying no project block", func(t *testing.T) {
		sectionsDir := writeContinuationFixture(t, "workflow:\n    default_mode: \"\"\n")
		path := filepath.Join(sectionsDir, defs.WorkflowYAML)

		opts := InitOptions{ProjectContinuation: config.ProjectContinuationNone}
		if err := writeWorkflowProjectContinuationYAML(sectionsDir, opts, &InitResult{}); err != nil {
			t.Fatalf("write: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read back: %v", err)
		}
		if !strings.Contains(string(got), "continuation: none") {
			t.Errorf("upsert did not create the nested key:\n%s", string(got))
		}
	})

	t.Run("no-deployer fallback creates a minimal block", func(t *testing.T) {
		sectionsDir := writeContinuationFixture(t, "")
		path := filepath.Join(sectionsDir, defs.WorkflowYAML)

		result := &InitResult{}
		opts := InitOptions{ProjectContinuation: config.ProjectContinuationPipeline}
		if err := writeWorkflowProjectContinuationYAML(sectionsDir, opts, result); err != nil {
			t.Fatalf("write: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read back: %v", err)
		}
		want := "workflow:\n    project:\n        continuation: pipeline\n"
		if string(got) != want {
			t.Errorf("fallback content = %q, want %q", string(got), want)
		}
		if len(result.CreatedFiles) != 1 {
			t.Errorf("CreatedFiles = %v, want exactly one entry", result.CreatedFiles)
		}
	})
}
