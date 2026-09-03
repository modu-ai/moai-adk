package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// precommit_gate_scope_e2e_test.go — end-to-end pre-commit hook verification
// (SPEC-PRECOMMIT-GATE-SCOPE-001 M4: AC-001 / AC-002 / AC-006-block / AC-007).
//
// The REAL hook content (preCommitHookContent) runs inside a real git
// repository in t.TempDir(), with a `moai` test double on PATH. The double
// mirrors the runner contract the M1 Go-level tests pin independently
// (gate_precommit_test.go drives the real runner): under MOAI_PRECOMMIT=1
// with the opt-in absent it exits 0 (the runner's skip), otherwise it exits 1
// (a project-wide check that fails — the pre-existing failure of AC-002).
// Neither side of the seam trusts the other; the hook side is what these
// tests prove — that the hook exports the marker, that the default posture
// lets unrelated commits through, and that the failure guidance carries the
// five remedy strings.

// gateE2EFakeMoai is the test double described above. NOT production code.
const gateE2EFakeMoai = `#!/bin/sh
# Test double for 'moai gate' (SPEC E2E fixture — mirrors the runner contract).
if [ "${MOAI_PRECOMMIT:-0}" = "1" ]; then
    enabled="$(awk '/^[[:space:]]*pre_commit:/{inblk=1; next} /^[^[:space:]#]/{inblk=0} inblk && $1 == "enabled:"{print $2; exit}' .moai/config/sections/gate.yaml 2>/dev/null)"
    [ "$enabled" = "true" ] || exit 0
fi
exit 1
`

// writeGateE2ERepo builds a git repo with the hook installed, an optional
// gate.yaml body, and (optionally) the fake moai on the hook's PATH. It
// returns the repo path and a commit func.
func writeGateE2ERepo(t *testing.T, gateYAML string, withFakeMoai bool) (string, func(msg string) (string, error)) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "hooks", "pre-commit"), []byte(preCommitHookContent), 0o755); err != nil {
		t.Fatal(err)
	}
	if gateYAML != "" {
		if err := os.WriteFile(filepath.Join(root, ".moai", "config", "sections", "gate.yaml"), []byte(gateYAML), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	hookPath := "/usr/bin:/bin:/usr/sbin:/sbin"
	if withFakeMoai {
		binDir := filepath.Join(root, ".e2e-bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(binDir, "moai"), []byte(gateE2EFakeMoai), 0o755); err != nil {
			t.Fatal(err)
		}
		hookPath = binDir + ":" + hookPath
	}

	commit := func(msg string) (string, error) {
		if err := os.WriteFile(filepath.Join(root, "unrelated.txt"), []byte(msg+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{
			{"add", "unrelated.txt"},
			{"-c", "user.email=t461@example.test", "-c", "user.name=t461", "commit", "-m", msg},
		} {
			cmd := exec.Command("git", args...)
			cmd.Dir = root
			cmd.Env = append(os.Environ(), "PATH="+hookPath, "SKIP_MOAI_PRECOMMIT=")
			out, err := cmd.CombinedOutput()
			if err != nil {
				return string(out), err
			}
		}
		return "", nil
	}
	return root, commit
}

// TestPrecommitE2EDefaultAllowsUnrelatedCommit is AC-002 end-to-end: a
// pre-existing project-wide failure (the double's exit 1 path) and an
// unrelated staged change, with no opt-in — the commit succeeds because the
// hook exports the marker and the runner skips the heavy steps by default.
func TestPrecommitE2EDefaultAllowsUnrelatedCommit(t *testing.T) {
	_, commit := writeGateE2ERepo(t, "gate:\n  enabled: true\n  skip_tests: false\n", true)

	if out, err := commit("unrelated change under default posture"); err != nil {
		t.Fatalf("commit blocked under the default posture (AC-002 violated):\n%s", out)
	}
}

// TestPrecommitE2EOptInBlocksAndGuides is AC-006's blocking direction plus
// AC-001's mechanical grep: with gate.pre_commit.enabled: true the failing
// project-wide check blocks the commit, and the hook's output names the
// config path, all four gate keys, and the retained SKIP_MOAI_PRECOMMIT=1.
func TestPrecommitE2EOptInBlocksAndGuides(t *testing.T) {
	_, commit := writeGateE2ERepo(t, "gate:\n  enabled: true\n  pre_commit:\n    enabled: true\n", true)

	out, err := commit("opted-in change with a failing project-wide check")
	if err == nil {
		t.Fatal("commit succeeded with the heavy gate opted in and failing (AC-006 blocking direction violated)")
	}
	for _, want := range []string{
		".moai/config/sections/gate.yaml",
		"gate.pre_commit.enabled",
		"gate.enabled",
		"gate.skip_tests",
		"gate.disabled_steps",
		"SKIP_MOAI_PRECOMMIT=1",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("failure output missing remedy string %q:\n%s", want, out)
		}
	}
}

// TestPrecommitE2EExplicitOptOutAllowsCommit pins the explicit-false direction:
// pre_commit.enabled: false behaves exactly like the absent default.
func TestPrecommitE2EExplicitOptOutAllowsCommit(t *testing.T) {
	_, commit := writeGateE2ERepo(t, "gate:\n  enabled: true\n  pre_commit:\n    enabled: false\n", true)

	if out, err := commit("explicit opt-out change"); err != nil {
		t.Fatalf("commit blocked despite explicit pre_commit.enabled: false:\n%s", out)
	}
}

// TestPrecommitE2ENoMoaiOnPathPasses is AC-007: without moai on PATH the hook
// skips the heavy gate and exits 0 (non-moai downstream projects unchanged).
func TestPrecommitE2ENoMoaiOnPathPasses(t *testing.T) {
	_, commit := writeGateE2ERepo(t, "", false)

	if out, err := commit("change in a project without moai on PATH"); err != nil {
		t.Fatalf("commit blocked with moai absent from PATH (AC-007 violated):\n%s", out)
	}
}
