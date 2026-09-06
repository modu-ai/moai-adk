package cli

// integration_settings_drift_exitcode_test.go — AC-PSD-006.
//
// The verdict of this gate is a match count. Card t474 watched a gate invert
// while its grep still exited 0, and this card produced four more instances of
// the same shape while being built: a `$?` read after a pipe; `git ls-remote`
// answering rc=0 with empty stdout AND empty stderr; `gh api .../status`
// reporting success while its check-runs failed; and a `grep -c` hit that
// turned out to be the false branch. So no decision here may be taken from an
// exit code, and this test is what keeps that mechanical rather than intended.
//
// The sweep is an absence assertion, so it stands on a positive control: the
// file list must be non-empty and must contain the implementation files. A
// selector that matched nothing would otherwise report the same clean 0.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// settingsDriftSourceFiles are the predicate's and the gate's own files, both
// implementation and test, given relative to internal/cli.
func settingsDriftSourceFiles() []string {
	return []string{
		"../kanban/settings_drift.go",
		"../kanban/settings_drift_test.go",
		// integration.go carries the acquire wiring that propagates the
		// refusal to the caller — part of the gate, so part of the sweep
		// (sync-audit F5, a coverage gap rather than a live defect).
		"integration.go",
		"integration_settings_drift.go",
		"integration_settings_drift_test.go",
		"integration_settings_drift_report_test.go",
		"integration_settings_drift_exitcode_test.go",
	}
}

func TestSettingsDriftVerdictNeverReadsAnExitCode(t *testing.T) {
	t.Parallel()

	// Positive control: every named file exists and is non-empty. Without it a
	// typo in a path would silently shrink the swept population to nothing and
	// the absence assertion below would pass by vacuity.
	swept := map[string]string{}
	for _, rel := range settingsDriftSourceFiles() {
		data, err := os.ReadFile(rel) // #nosec G304 -- fixed in-repo relative paths
		if err != nil {
			t.Fatalf("control: cannot read %s: %v", rel, err)
		}
		if len(data) == 0 {
			t.Fatalf("control: %s is empty", rel)
		}
		swept[rel] = string(data)
	}
	if len(swept) != len(settingsDriftSourceFiles()) {
		t.Fatalf("control: swept %d files, expected %d", len(swept), len(settingsDriftSourceFiles()))
	}
	// The control also has to prove the swept set is the RIGHT set, not merely
	// a non-empty one: an implementation file that does not define the
	// predicate would satisfy "non-empty" while checking the wrong code.
	if !strings.Contains(swept["../kanban/settings_drift.go"], "func DetectSettingsDrift(") {
		t.Fatalf("control: ../kanban/settings_drift.go does not define DetectSettingsDrift")
	}
	if !strings.Contains(swept["integration_settings_drift.go"], "func newIntegrationPreflightCmd(") {
		t.Fatalf("control: integration_settings_drift.go does not define the preflight command")
	}

	// The absence assertion, on top of that control.
	forbidden := []string{"ExitCode", "ExitError", "$?"}
	for rel, body := range swept {
		for lineNo, line := range strings.Split(body, "\n") {
			// This file names the forbidden tokens in order to search for
			// them; excluding it by name would be a hole, so it is excluded by
			// the narrower fact that it is the sweep itself.
			if filepath.Base(rel) == "integration_settings_drift_exitcode_test.go" {
				continue
			}
			for _, tok := range forbidden {
				if strings.Contains(line, tok) {
					t.Errorf("%s:%d reads an exit code (%q): %s", rel, lineNo+1, tok, strings.TrimSpace(line))
				}
			}
		}
	}
}
