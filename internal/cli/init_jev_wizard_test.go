package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// init_jev_wizard_test.go — the wizard entrance of the shared Jev persistence
// path (SPEC-JEV-OPTIN-MEASURE-001 REQ-JEVO-002, AC-JEVO-002).

const jevInitWorkflowFixture = `workflow:
    # sibling comment
    default_mode: ""
    jev:
        enabled: false
`

func jevInitFixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(jevInitWorkflowFixture), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return root
}

func jevInitReadWorkflow(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatalf("read workflow.yaml: %v", err)
	}
	return string(b)
}

// TestApplyJevFromWizard_OptInPersists: the accepted opt-in reaches the file.
func TestApplyJevFromWizard_OptInPersists(t *testing.T) {
	root := jevInitFixtureRoot(t)
	if err := applyJevFromWizard(true, &wizard.WizardResult{JevEnabled: true}, root); err != nil {
		t.Fatalf("applyJevFromWizard: %v", err)
	}
	if got := jevInitReadWorkflow(t, root); !strings.Contains(got, "enabled: true") {
		t.Errorf("opt-in not persisted:\n%s", got)
	}
}

// TestApplyJevFromWizard_WizardDidNotRun_WritesNothing is the clobber guard:
// a non-interactive init must not write the wizard's zero value over a value
// the user set earlier.
func TestApplyJevFromWizard_WizardDidNotRun_WritesNothing(t *testing.T) {
	root := jevInitFixtureRoot(t)
	// Seed an enabled state, as an earlier interactive init would have.
	if err := applyJevFromWizard(true, &wizard.WizardResult{JevEnabled: true}, root); err != nil {
		t.Fatalf("seed: %v", err)
	}
	before := jevInitReadWorkflow(t, root)
	// Positive control: the seed actually took, so the comparison below is
	// measuring preservation rather than an unchanged no-op fixture.
	if !strings.Contains(before, "enabled: true") {
		t.Fatalf("positive control failed: the seed did not enable\n%s", before)
	}

	if err := applyJevFromWizard(false, &wizard.WizardResult{}, root); err != nil {
		t.Fatalf("applyJevFromWizard(wizardRan=false): %v", err)
	}
	if after := jevInitReadWorkflow(t, root); after != before {
		t.Errorf("a non-interactive run rewrote the file\nbefore:\n%s\nafter:\n%s", before, after)
	}

	// Nil result is the same case and must be equally inert.
	if err := applyJevFromWizard(true, nil, root); err != nil {
		t.Fatalf("applyJevFromWizard(nil result): %v", err)
	}
	if after := jevInitReadWorkflow(t, root); after != before {
		t.Errorf("a nil wizard result rewrote the file\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestApplyJevFromWizard_DeclineDisables: a declined answer from a wizard that
// DID run turns a previously-enabled capability off. The opt-in is a live
// answer in both directions.
func TestApplyJevFromWizard_DeclineDisables(t *testing.T) {
	root := jevInitFixtureRoot(t)
	if err := applyJevFromWizard(true, &wizard.WizardResult{JevEnabled: true}, root); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := applyJevFromWizard(true, &wizard.WizardResult{JevEnabled: false}, root); err != nil {
		t.Fatalf("applyJevFromWizard(decline): %v", err)
	}
	got := jevInitReadWorkflow(t, root)
	if strings.Contains(got, "enabled: true") {
		t.Errorf("decline did not disable:\n%s", got)
	}
	if !strings.Contains(got, "sibling comment") {
		t.Error("the sibling comment was destroyed by the write")
	}
}
