package cli

// doctor_worktree_base_test.go — SPEC-WORKTREE-BASEREF-001 M4 (card t313).
// AC-WBR-009: the diagnostic distinguishes FOUR states, and the two non-OK
// messages are NOT equal — their repairs differ (run the alignment vs correct
// the setting), so a single shared "base branch problem" message would tell the
// user nothing actionable.

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// errNoOriginHeadSentinel stands in for the error git returns when
// refs/remotes/origin/HEAD does not exist (no origin remote configured).
var errNoOriginHeadSentinel = errors.New("fatal: ref refs/remotes/origin/HEAD is not a symbolic ref")

// swapDoctorWorktreeBaseSeams installs fakes for the git-facing seams the
// diagnostic reads, and restores them on cleanup.
func swapDoctorWorktreeBaseSeams(t *testing.T, head string, headErr error, resolvable bool) {
	t.Helper()
	origHead := doctorWorktreeBaseOriginHead
	origResolvable := doctorWorktreeBaseResolvable
	doctorWorktreeBaseOriginHead = func() (string, error) { return head, headErr }
	doctorWorktreeBaseResolvable = func(string) bool { return resolvable }
	t.Cleanup(func() {
		doctorWorktreeBaseOriginHead = origHead
		doctorWorktreeBaseResolvable = origResolvable
	})
}

func TestDoctorWorktreeBaseBranchFourStates(t *testing.T) {
	// (c) and (d) are captured here so the distinctness assertion below can
	// compare them.
	var mismatchMsg, unresolvableMsg string

	t.Run("empty setting is OK", func(t *testing.T) {
		root := t.TempDir()
		writeWorktreeBaseConfig(t, root, "")
		swapDoctorWorktreeBaseSeams(t, "main", nil, true)

		got := checkWorktreeBaseBranch(root, false)
		if got.Name != "Worktree Base Branch" {
			t.Errorf("Name = %q, want %q (moai doctor --check filters by exact name equality)", got.Name, "Worktree Base Branch")
		}
		if got.Status != uikit.CheckOK {
			t.Errorf("empty setting: status = %q, want %q", got.Status, uikit.CheckOK)
		}
	})

	t.Run("matching setting is OK", func(t *testing.T) {
		root := t.TempDir()
		writeWorktreeBaseConfig(t, root, "develop")
		swapDoctorWorktreeBaseSeams(t, "develop", nil, true)

		got := checkWorktreeBaseBranch(root, false)
		if got.Status != uikit.CheckOK {
			t.Errorf("match: status = %q, want %q", got.Status, uikit.CheckOK)
		}
	})

	t.Run("resolvable mismatch is non-OK and names the repair command", func(t *testing.T) {
		root := t.TempDir()
		writeWorktreeBaseConfig(t, root, "develop")
		swapDoctorWorktreeBaseSeams(t, "main", nil, true)

		got := checkWorktreeBaseBranch(root, false)
		if got.Status == uikit.CheckOK {
			t.Errorf("mismatch: status = %q, want a non-OK status", got.Status)
		}
		if !strings.Contains(got.Message, "git remote set-head origin develop") {
			t.Errorf("mismatch message %q does not name the repair command", got.Message)
		}
		mismatchMsg = got.Message
	})

	t.Run("unresolvable value is non-OK and names the offending value", func(t *testing.T) {
		root := t.TempDir()
		writeWorktreeBaseConfig(t, root, "no-such-branch")
		swapDoctorWorktreeBaseSeams(t, "main", nil, false)

		got := checkWorktreeBaseBranch(root, false)
		if got.Status == uikit.CheckOK {
			t.Errorf("unresolvable: status = %q, want a non-OK status", got.Status)
		}
		if !strings.Contains(got.Message, "no-such-branch") {
			t.Errorf("unresolvable message %q does not name the offending value", got.Message)
		}
		if !strings.Contains(got.Message, "setting") {
			t.Errorf("unresolvable message %q does not direct the user to correct the setting", got.Message)
		}
		unresolvableMsg = got.Message
	})

	if mismatchMsg == "" || unresolvableMsg == "" {
		t.Fatal("both non-OK sub-tests must run before the distinctness assertion")
	}
	if mismatchMsg == unresolvableMsg {
		t.Errorf("the two non-OK messages are identical (%q); collapsing them FAILS the criterion because the repairs differ", mismatchMsg)
	}
}

// TestDoctorWorktreeBaseBranchNoOriginRemoteIsInformational pins acceptance.md
// §D.2: with no origin remote configured, consumer 1 is a silent no-op and the
// doctor item reports informationally rather than as a failure.
func TestDoctorWorktreeBaseBranchNoOriginRemoteIsInformational(t *testing.T) {
	root := t.TempDir()
	writeWorktreeBaseConfig(t, root, "develop")
	swapDoctorWorktreeBaseSeams(t, "", errNoOriginHeadSentinel, true)

	got := checkWorktreeBaseBranch(root, false)
	if got.Status != uikit.CheckOK {
		t.Errorf("no origin/HEAD: status = %q, want %q (informational, not a failure)", got.Status, uikit.CheckOK)
	}
}

// TestDoctorWorktreeBaseBranchRegisteredByName pins that the item is reachable
// through `moai doctor --check 'Worktree Base Branch'`, which filters by exact
// name equality — the name is part of the contract, not an implementation
// detail.
func TestDoctorWorktreeBaseBranchRegisteredByName(t *testing.T) {
	// runDiagnosticChecks filters by exact name equality, exactly as
	// `moai doctor --check` does — so a non-empty result proves registration.
	got := runDiagnosticChecks(false, "Worktree Base Branch")
	if len(got) != 1 {
		t.Fatalf("moai doctor --check 'Worktree Base Branch' matched %d checks, want exactly 1 "+
			"(the item is unregistered, or its Name string drifted from the contract)", len(got))
	}
	if got[0].Name != "Worktree Base Branch" {
		t.Errorf("matched check Name = %q, want %q", got[0].Name, "Worktree Base Branch")
	}
}
