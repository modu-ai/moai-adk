package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// participationHomeFor points MOAI_HOME at a temporary directory and returns
// the path of the user-scoped consent file. The real home is never touched.
func participationHomeFor(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")
	return filepath.Join(home, "config", "participation.yaml")
}

func readParticipationFileFor(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(raw)
}

// TestApplyParticipationFromWizard covers the four persistence-gate arms:
// (a) wizard not run, (b) wizard run with CI empty, (c) wizard run with CI
// set, (d) nil result — and the decline-over-true direction.
func TestApplyParticipationFromWizard(t *testing.T) {
	t.Run("wizard_not_run_writes_nothing", func(t *testing.T) {
		path := participationHomeFor(t)
		res := &wizard.WizardResult{ParticipationEnabled: true}
		if err := applyParticipationFromWizard(false, res, t.TempDir()); err != nil {
			t.Fatalf("applyParticipationFromWizard: %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("consent file exists after a wizard that did not run: %v", err)
		}
	})

	t.Run("wizard_run_ci_empty_persists_both_values", func(t *testing.T) {
		path := participationHomeFor(t)
		res := &wizard.WizardResult{ParticipationEnabled: true}
		if err := applyParticipationFromWizard(true, res, t.TempDir()); err != nil {
			t.Fatalf("applyParticipationFromWizard: %v", err)
		}
		body := readParticipationFileFor(t, path)
		for _, want := range []string{"enabled: true", "asked: true"} {
			if !strings.Contains(body, want) {
				t.Errorf("consent file missing %q:\n%s", want, body)
			}
		}

		// A decline over an existing true writes false through the writer.
		res.ParticipationEnabled = false
		if err := applyParticipationFromWizard(true, res, t.TempDir()); err != nil {
			t.Fatalf("applyParticipationFromWizard (decline): %v", err)
		}
		body = readParticipationFileFor(t, path)
		if !strings.Contains(body, "enabled: false") {
			t.Errorf("decline did not persist enabled false:\n%s", body)
		}
		if !strings.Contains(body, "asked: true") {
			t.Errorf("decline lost asked true:\n%s", body)
		}
	})

	t.Run("wizard_run_ci_set_writes_nothing", func(t *testing.T) {
		path := participationHomeFor(t)
		t.Setenv("CI", "1")
		res := &wizard.WizardResult{ParticipationEnabled: true}
		if err := applyParticipationFromWizard(true, res, t.TempDir()); err != nil {
			t.Fatalf("applyParticipationFromWizard: %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("consent file written under CI: %v", err)
		}
	})

	t.Run("nil_result_writes_nothing", func(t *testing.T) {
		path := participationHomeFor(t)
		if err := applyParticipationFromWizard(true, nil, t.TempDir()); err != nil {
			t.Fatalf("applyParticipationFromWizard: %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("consent file written for a nil result: %v", err)
		}
	})
}
