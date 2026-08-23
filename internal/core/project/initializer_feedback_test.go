package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// initializer_feedback_test.go is the defence against wiring the wizard answer
// to a writer nothing calls. Asserting that the question exists would not catch
// a question that is asked, stored in WizardResult, and then dropped; these
// tests assert the answer reaches the file, which does.

func feedbackBoolPtr(v bool) *bool { return &v }

// deployedFeedbackYAML is a stand-in for the template-deployed feedback.yaml:
// comments and the repository key, and no auto_submit key.
const deployedFeedbackYAML = `feedback:
    # Target repository for the feedback workflow (GitHub owner/repo slug).
    repository: modu-ai/moai-adk
`

// TestWritePhase1ConfigsPersistsFeedbackAutoSubmit pins the true answer landing
// in feedback.yaml as a nested auto_submit key, upserted into a document that
// never carried it, with the surrounding comment and key intact.
func TestWritePhase1ConfigsPersistsFeedbackAutoSubmit(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	feedbackPath := filepath.Join(sectionsDir, defs.FeedbackYAML)
	if err := os.WriteFile(feedbackPath, []byte(deployedFeedbackYAML), 0o644); err != nil {
		t.Fatalf("seed feedback.yaml: %v", err)
	}

	opts := InitOptions{
		ProjectRoot:        root,
		ProjectMode:        "personal",
		FeedbackAutoSubmit: feedbackBoolPtr(true),
	}
	if err := WritePhase1Configs(opts, &InitResult{}); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	got, err := os.ReadFile(feedbackPath)
	if err != nil {
		t.Fatalf("read feedback.yaml: %v", err)
	}

	// Read it back through the real loader rather than grepping the text: the
	// point is that the runtime surfaces see the answer, and the loader is what
	// they see it through.
	cfg, err := config.NewLoader().Load(filepath.Join(root, defs.MoAIDir))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeedbackAutoSubmit() {
		t.Fatalf("feedback.yaml did not persist auto_submit: true; file:\n%s", got)
	}

	// Preservation: the patch must not flatten the document it edited.
	for _, want := range []string{"# Target repository", "repository: modu-ai/moai-adk"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("feedback.yaml lost %q after the auto_submit patch; file:\n%s", want, got)
		}
	}
}

// TestWritePhase1ConfigsSkipsFeedbackWhenUnanswered is the non-interactive
// path. A nil answer means the question was never asked, and an unasked
// question must not write anything.
func TestWritePhase1ConfigsSkipsFeedbackWhenUnanswered(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	feedbackPath := filepath.Join(sectionsDir, defs.FeedbackYAML)
	if err := os.WriteFile(feedbackPath, []byte(deployedFeedbackYAML), 0o644); err != nil {
		t.Fatalf("seed feedback.yaml: %v", err)
	}

	opts := InitOptions{ProjectRoot: root, ProjectMode: "personal"}
	if err := WritePhase1Configs(opts, &InitResult{}); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	got, err := os.ReadFile(feedbackPath)
	if err != nil {
		t.Fatalf("read feedback.yaml: %v", err)
	}
	if string(got) != deployedFeedbackYAML {
		t.Errorf("feedback.yaml changed with no auto_submit answer:\n%s", got)
	}
}

// TestWritePhase1ConfigsFeedbackNoFile covers the no-deployer fallback: with no
// feedback.yaml on disk, the answer still has to land somewhere the loader
// will find it.
func TestWritePhase1ConfigsFeedbackNoFile(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{
		ProjectRoot:        root,
		ProjectMode:        "personal",
		FeedbackAutoSubmit: feedbackBoolPtr(true),
	}
	if err := WritePhase1Configs(opts, &InitResult{}); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	if _, err := os.Stat(filepath.Join(sectionsDir, defs.FeedbackYAML)); err != nil {
		t.Fatalf("feedback.yaml not created on the fallback path: %v", err)
	}
	cfg, err := config.NewLoader().Load(filepath.Join(root, defs.MoAIDir))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.FeedbackAutoSubmit() {
		t.Fatal("fallback-created feedback.yaml does not enable auto_submit")
	}
}
