package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

// gate_precommit_test.go — the pre-commit-context runner contract
// (SPEC-PRECOMMIT-GATE-SCOPE-001 M1 / AC-006 / AC-009, operator decision 2).
//
// The git pre-commit hook invokes `moai gate` with MOAI_PRECOMMIT=1. Under
// that marker the runner honors gate.pre_commit.enabled: false (the default)
// skips the project-wide heavy steps and passes. Without the marker — a
// standalone `moai gate` run — the key is never read and the existing
// gate.enabled contract is unchanged. One branch point, marker-gated; the
// shared config.Enabled switch is never flipped (Known Issue 2).
//
// The fixture is a real Go module whose test step FAILS. That pre-existing
// project-wide failure is the probe: a runner that (wrongly) executes the
// heavy steps under the marker makes these tests red.

// writeGatePreCommitFixture builds a project with a pre-existing failing
// project-wide check (go test fails) and an optional gate.yaml body. It
// returns the fixture root.
func writeGatePreCommitFixture(t *testing.T, gateYAML string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"go.mod":       "module fixture\n\ngo 1.21\n",
		"main.go":      "package main\n\nfunc main() {}\n",
		"fail_test.go": "package main\n\nimport \"testing\"\n\nfunc TestPreExistingFailure(t *testing.T) {\n\tt.Fatal(\"pre-existing project-wide failure\")\n}\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if gateYAML != "" {
		p := filepath.Join(root, ".moai", "config", "sections", "gate.yaml")
		if err := os.WriteFile(p, []byte(gateYAML), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// runGateInFixture points the gate at the fixture and runs runGate, returning
// the error plus everything the run wrote to stderr.
func runGateInFixture(t *testing.T, root string) (string, error) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	var errOut bytes.Buffer
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&errOut)
	err := runGate(cmd, nil)
	return errOut.String(), err
}

// TestRunGatePreCommitMarkerSkipsHeavyGate is the AC-009 contrast's first
// half: under MOAI_PRECOMMIT=1 with the default (opt-out) posture, the runner
// passes with exit 0 and never executes the project-wide heavy steps — the
// fixture's failing go test never runs.
func TestRunGatePreCommitMarkerSkipsHeavyGate(t *testing.T) {
	root := writeGatePreCommitFixture(t, "")
	t.Setenv(config.EnvPreCommitMarker, "1")

	errOut, err := runGateInFixture(t, root)
	if err != nil {
		t.Fatalf("runGate under marker + default opt-out returned error: %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(errOut, "gate.pre_commit.enabled") {
		t.Errorf("skip notice does not name the remedy key gate.pre_commit.enabled; stderr: %s", errOut)
	}
	if strings.Contains(errOut, "quality gate failed") {
		t.Errorf("heavy steps ran under the marker (gate failed output present); stderr: %s", errOut)
	}
}

// TestRunGateStandaloneUnchanged is the characterization baseline (written
// and observed BEFORE the runner branch landed): without the marker, the same
// failing fixture fails the gate under the existing contract. This is the
// second half of the AC-009 contrast and the immutability proof for the
// standalone `moai gate` behavior.
func TestRunGateStandaloneUnchanged(t *testing.T) {
	root := writeGatePreCommitFixture(t, "")
	t.Setenv(config.EnvPreCommitMarker, "")

	_, err := runGateInFixture(t, root)
	if err == nil {
		t.Fatal("standalone moai gate passed on a project with a failing test — the existing contract regressed")
	}
	if !strings.Contains(err.Error(), "quality gate failed") {
		t.Errorf("error = %v, want the existing quality gate failed contract", err)
	}
}

// TestRunGatePreCommitOptInRunsHeavyGate is AC-006's runner half: with the
// marker present AND gate.pre_commit.enabled: true, the heavy steps run —
// the fixture's pre-existing failure blocks (gate fails).
func TestRunGatePreCommitOptInRunsHeavyGate(t *testing.T) {
	root := writeGatePreCommitFixture(t, "gate:\n  pre_commit:\n    enabled: true\n")
	t.Setenv(config.EnvPreCommitMarker, "1")

	_, err := runGateInFixture(t, root)
	if err == nil {
		t.Fatal("opted-in pre-commit gate passed despite the failing project-wide test — the heavy steps did not run")
	}
}

// TestRunGatePreCommitOptInPassingProject passes pins the benign direction of
// the opt-in: marker present, opted in, healthy project — the gate passes.
func TestRunGatePreCommitOptInPassingProjectPasses(t *testing.T) {
	root := writeGatePreCommitFixture(t, "gate:\n  pre_commit:\n    enabled: true\n")
	t.Setenv(config.EnvPreCommitMarker, "1")
	// Repair the fixture: overwrite the failing test so the project is healthy.
	if err := os.WriteFile(filepath.Join(root, "fail_test.go"),
		[]byte("package main\n\nimport \"testing\"\n\nfunc TestPreExistingFailure(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	errOut, err := runGateInFixture(t, root)
	if err != nil {
		t.Fatalf("opted-in gate on a healthy project returned error: %v\nstderr: %s", err, errOut)
	}
}

// TestRunGateGateDisabledStillShortCircuitsUnderMarker pins precedence: the
// existing gate.enabled=false short-circuit still passes under the marker —
// the branch point never flips config.Enabled.
func TestRunGateGateDisabledStillShortCircuitsUnderMarker(t *testing.T) {
	root := writeGatePreCommitFixture(t, "gate:\n  enabled: false\n")
	t.Setenv(config.EnvPreCommitMarker, "1")

	_, err := runGateInFixture(t, root)
	if err != nil {
		t.Fatalf("gate.enabled=false under the marker returned error: %v", err)
	}
}
