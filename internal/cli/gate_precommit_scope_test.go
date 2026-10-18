package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// gate_precommit_scope_test.go — the runner side of gate.pre_commit.test_scope.mode
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-011 / AC-PAS-011).
//
// Under the pre-commit marker with the heavy gate opted in, an unknown scope
// mode is rejected: the run fails with a message naming the key and the
// accepted values, rather than silently choosing a scope.

// TestRunGatePreCommitInvalidScopeModeRejected rejects an unknown mode before any
// step runs.
func TestRunGatePreCommitInvalidScopeModeRejected(t *testing.T) {
	root := writeGatePreCommitFixture(t,
		"gate:\n  pre_commit:\n    enabled: true\n    test_scope:\n      mode: bogus\n")
	t.Setenv(config.EnvPreCommitMarker, "1")

	errOut, err := runGateInFixture(t, root)
	if err == nil {
		t.Fatalf("an unknown test_scope.mode was accepted; stderr: %s", errOut)
	}
	msg := err.Error() + "\n" + errOut
	for _, want := range []string{"gate.pre_commit.test_scope.mode", "affected", "full", "bogus"} {
		if !strings.Contains(msg, want) {
			t.Errorf("rejection does not contain %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "quality gate failed") {
		t.Errorf("the gate ran its steps before rejecting the mode:\n%s", msg)
	}
}

// A valid mode is not rejected: full runs the module as before.
func TestRunGatePreCommitFullScopeModeAccepted(t *testing.T) {
	root := writeGatePreCommitFixture(t,
		"gate:\n  pre_commit:\n    enabled: true\n    test_scope:\n      mode: full\n")
	t.Setenv(config.EnvPreCommitMarker, "1")

	_, err := runGateInFixture(t, root)
	if err == nil || !strings.Contains(err.Error(), "quality gate failed") {
		t.Fatalf("mode full did not run the module's tests (the fixture's go test fails): err = %v", err)
	}
}

// Outside the pre-commit marker the key is never read: a standalone run with an
// unknown mode behaves exactly as before.
func TestRunGateStandaloneIgnoresScopeMode(t *testing.T) {
	root := writeGatePreCommitFixture(t,
		"gate:\n  pre_commit:\n    enabled: true\n    test_scope:\n      mode: bogus\n")
	t.Setenv(config.EnvPreCommitMarker, "")

	_, err := runGateInFixture(t, root)
	if err == nil || !strings.Contains(err.Error(), "quality gate failed") {
		t.Fatalf("standalone run: err = %v, want the existing quality gate failed contract (the key is not read)", err)
	}
}
