package cli

// SPEC-INIT-WIZARD-REPAIR-001 chain ② runInit + update-wizard wiring tests
// (M2): the four workflow toggle flags persist to workflow.yaml through the
// initializer's tracker-gated writer; --worktree-auto-create beats the wizard
// advisory answer (REQ-005); the update-wizard workflow step is a no-op
// without a TTY and persists only deltas with one (REQ-007).

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// runInitForWorkflow runs a full non-interactive (or wizard-injected) init in
// a temp project with HOME isolated, returning the deployed workflow.yaml
// path. Mirrors runInitForAutonomyAtHome.
func runInitForWorkflow(t *testing.T, wizResult *wizard.WizardResult, flags map[string]string) (projectDir, workflowPath string) {
	t.Helper()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	projectDir = runInitForAutonomyAtHome(t, homeDir, wizResult, flags)
	return projectDir, filepath.Join(projectDir, ".moai", "config", "sections", defs.WorkflowYAML)
}

// embeddedTemplateWorkflow returns the shipped workflow.yaml bytes — the
// byte-identity baseline for a no-flags init.
func embeddedTemplateWorkflow(t *testing.T) []byte {
	t.Helper()
	embedFS, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	data, err := fs.ReadFile(embedFS, filepath.ToSlash(filepath.Join(defs.MoAIDir, "config", "sections", defs.WorkflowYAML)))
	if err != nil {
		t.Fatalf("read embedded workflow.yaml: %v", err)
	}
	return data
}

// TestRunInit_WorkflowToggleFlagsPersist asserts AC-005: explicitly passing
// each of the four toggle flags persists the matching workflow.yaml key.
func TestRunInit_WorkflowToggleFlagsPersist(t *testing.T) {
	_, workflowPath := runInitForWorkflow(t, nil, map[string]string{
		"branch-guard":          "true",
		"worktree-auto-create":  "true",
		"worktree-auto-merge":   "true",
		"worktree-auto-cleanup": "true",
	})

	got, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read deployed workflow.yaml: %v", err)
	}
	for _, want := range []string{
		"branch_guard:",
		"enabled: true",
		"auto_create: true",
		"auto_merge: true",
		"auto_cleanup: true",
	} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("deployed workflow.yaml missing %q; got:\n%s", want, got)
		}
	}
}

// TestRunInit_WorkflowToggleFlagsAbsentByteIdentical asserts AC-006: with
// none of the four flags passed (interactive or not), the deployed
// workflow.yaml is byte-identical to the template — no key synthesized, no
// comment disturbed.
func TestRunInit_WorkflowToggleFlagsAbsentByteIdentical(t *testing.T) {
	_, workflowPath := runInitForWorkflow(t, &wizard.WizardResult{}, nil)

	got, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read deployed workflow.yaml: %v", err)
	}
	want := embeddedTemplateWorkflow(t)
	if !bytes.Equal(got, want) {
		t.Errorf("no-flags init must deploy workflow.yaml byte-identical to the template;\n--template--\n%s\n--deployed--\n%s", want, got)
	}
}

// TestRunInit_WorktreeAutoCreateFlagBeatsWizard asserts AC-007: an explicit
// --worktree-auto-create=true wins over a wizard advisory answer of false;
// with the flag absent, the wizard answer applies (false — byte-identical to
// the template default, REQ-006).
func TestRunInit_WorktreeAutoCreateFlagBeatsWizard(t *testing.T) {
	wiz := &wizard.WizardResult{WorktreeAutoCreate: false}

	// Flag present: persisted true despite the wizard's false.
	_, workflowPath := runInitForWorkflow(t, wiz, map[string]string{
		"worktree-auto-create": "true",
	})
	got, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read deployed workflow.yaml: %v", err)
	}
	if !bytes.Contains(got, []byte("auto_create: true")) {
		t.Errorf("flag must beat the wizard answer: auto_create not true; got:\n%s", got)
	}

	// Flag absent: the wizard answer applies; false equals the template
	// default, so the file stays byte-identical (no tracker flipped).
	_, workflowPath2 := runInitForWorkflow(t, wiz, nil)
	got2, err := os.ReadFile(workflowPath2)
	if err != nil {
		t.Fatalf("read deployed workflow.yaml (flag absent): %v", err)
	}
	if want := embeddedTemplateWorkflow(t); !bytes.Equal(got2, want) {
		t.Errorf("flag-absent run with wizard answer false must stay byte-identical to the template;\n--template--\n%s\n--deployed--\n%s", want, got2)
	}
}

// TestRunWorkflowConfigStep_InteractiveDeltaPersist asserts REQ-007's
// interactive half: with a TTY-shaped stdin and answered deltas, only the
// answered keys change and every other line survives. Uses the package's
// isInteractiveStdin + stdinPromptSource seams.
func TestRunWorkflowConfigStep_InteractiveDeltaPersist(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	before := `workflow:
    execution_mode: auto
    worktree:
        auto_create: false
        auto_merge: false
        auto_cleanup: false
`
	workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)
	if err := os.WriteFile(workflowPath, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })
	origSrc := stdinPromptSource
	// One shared reader: stdinPromptSource is invoked once PER PROMPT, so a
	// closure that builds a fresh reader each call would re-read the first
	// answer every time. branch-guard -> y (delta); the other three -> n
	// (their current values, so no edit).
	answers := strings.NewReader("y\nn\nn\nn")
	stdinPromptSource = func() io.Reader { return answers }
	t.Cleanup(func() { stdinPromptSource = origSrc })

	var out bytes.Buffer
	if err := runWorkflowConfigStep(&out, root); err != nil {
		t.Fatalf("runWorkflowConfigStep: %v", err)
	}

	got, _ := os.ReadFile(workflowPath)
	if !bytes.Contains(got, []byte("enabled: true")) {
		t.Errorf("answered branch-guard delta must persist; got:\n%s", got)
	}
	for _, want := range []string{"auto_create: false", "auto_merge: false", "auto_cleanup: false"} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("unanswered toggle must keep its value (%q); got:\n%s", want, got)
		}
	}
}

// TestApplyWizardReconfigureSteps_RunsWorkflowStep asserts REQ-007's wiring:
// the update-wizard apply path runs the workflow step after applying the
// wizard config. On a non-TTY the step is a no-op (AC-008's second clause) —
// the workflow.yaml is untouched.
func TestApplyWizardReconfigureSteps_RunsWorkflowStep(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	before := "workflow:\n    execution_mode: auto\n"
	workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)
	if err := os.WriteFile(workflowPath, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := applyWizardReconfigureSteps(&out, root, &wizard.WizardResult{}); err != nil {
		t.Fatalf("applyWizardReconfigureSteps: %v", err)
	}
	got, _ := os.ReadFile(workflowPath)
	if !bytes.Equal(got, []byte(before)) {
		t.Errorf("non-TTY run must leave workflow.yaml untouched;\nwant:\n%s\ngot:\n%s", before, got)
	}
}
