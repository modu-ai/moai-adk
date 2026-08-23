package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// initializer_todo_test.go is AC-T-008 — the defence against wiring the wizard
// answer to a writer nothing calls.
//
// The card that motivated this SPEC named applyAutonomyTierFromWizard as the
// precedent to follow; that function has no production caller, so following it
// would produce a question that is asked, stored in WizardResult, and then
// dropped. Asserting that the question EXISTS would not have caught that. These
// tests assert the answer reaches the file, which does.

func todoBoolPtr(v bool) *bool { return &v }

// deployedWorkflowYAML is a stand-in for the template-deployed workflow.yaml:
// comments and unrelated keys, and NO todo block — which is exactly what
// distributed users have, since M6 deliberately does not ship one.
const deployedWorkflowYAML = `# Workflow configuration
workflow:
    # Execution mode for the run phase.
    execution_mode: team
    worktree:
        auto_create: false
`

// TestWritePhase1ConfigsPersistsTodoEnabled pins the false answer landing in
// workflow.yaml as a nested todo.enabled key, upserted into a document that
// never carried the block, with the surrounding comments and keys intact.
func TestWritePhase1ConfigsPersistsTodoEnabled(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)
	if err := os.WriteFile(workflowPath, []byte(deployedWorkflowYAML), 0o644); err != nil {
		t.Fatalf("seed workflow.yaml: %v", err)
	}

	opts := InitOptions{
		ProjectRoot: root,
		ProjectMode: "personal",
		TodoEnabled: todoBoolPtr(false),
	}
	if err := WritePhase1Configs(opts, &InitResult{}); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	got, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow.yaml: %v", err)
	}

	// Read it back through the real loader rather than grepping the text: the
	// point of the AC is that the runtime surfaces see the answer, and the
	// loader is what they see it through.
	cfg, err := config.NewLoader().Load(filepath.Join(root, defs.MoAIDir))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.TodoEnabled() {
		t.Fatalf("workflow.yaml did not persist todo.enabled: false; file:\n%s", got)
	}

	// Preservation: the patch must not flatten the document it edited.
	for _, want := range []string{"# Workflow configuration", "execution_mode: team", "auto_create: false"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("workflow.yaml lost %q after the todo patch; file:\n%s", want, got)
		}
	}
}

// TestWritePhase1ConfigsSkipsTodoWhenUnanswered is the non-interactive path.
// A nil answer means the question was never asked, and an unasked question
// must not write anything — writing `enabled: true` would be equivalent in
// meaning but would put a key in every user's config for no reason, and
// writing `false` would silently disable a default-on feature.
func TestWritePhase1ConfigsSkipsTodoWhenUnanswered(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)
	if err := os.WriteFile(workflowPath, []byte(deployedWorkflowYAML), 0o644); err != nil {
		t.Fatalf("seed workflow.yaml: %v", err)
	}

	opts := InitOptions{ProjectRoot: root, ProjectMode: "personal"}
	if err := WritePhase1Configs(opts, &InitResult{}); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	got, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow.yaml: %v", err)
	}
	if string(got) != deployedWorkflowYAML {
		t.Errorf("workflow.yaml changed with no todo answer:\n%s", got)
	}
}

// TestWritePhase1ConfigsTodoNoWorkflowFile covers the no-deployer fallback:
// with no workflow.yaml on disk, the answer still has to land somewhere the
// loader will find it.
func TestWritePhase1ConfigsTodoNoWorkflowFile(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{
		ProjectRoot: root,
		ProjectMode: "personal",
		TodoEnabled: todoBoolPtr(false),
	}
	if err := WritePhase1Configs(opts, &InitResult{}); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	if _, err := os.Stat(filepath.Join(sectionsDir, defs.WorkflowYAML)); err != nil {
		t.Fatalf("workflow.yaml not created on the fallback path: %v", err)
	}
	cfg, err := config.NewLoader().Load(filepath.Join(root, defs.MoAIDir))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.TodoEnabled() {
		t.Fatal("fallback-created workflow.yaml does not disable todo")
	}
}
