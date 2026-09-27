package escalation_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// The verify environment reads the constitution registry from the worktree,
// as `moai contract verify` does: rule IDs and the distinct Frozen-zone files.
func TestLoadVerifyEnvReadsRegistry(t *testing.T) {
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteMode("contract", false)
	w.Write("CLAUDE.md", "# fixture\n")
	w.Write(".claude/rules/moai/core/zone-registry.md", "# registry\n\n```yaml\n"+
		"- id: CONST-V3R2-001\n  zone: Frozen\n  file: CLAUDE.md\n  anchor: \"#a\"\n  clause: first\n  canary_gate: true\n"+
		"- id: CONST-V3R2-002\n  zone: Frozen\n  file: CLAUDE.md\n  anchor: \"#b\"\n  clause: second\n  canary_gate: true\n"+
		"- id: CONST-V3R2-003\n  zone: Evolvable\n  file: CLAUDE.md\n  anchor: \"#c\"\n  clause: third\n"+
		"```\n")
	env := escalation.LoadVerifyEnv(w.Root, contractSettings(t, w))
	if !slices.Equal(env.RegistryRuleIDs, []string{"CONST-V3R2-001", "CONST-V3R2-002", "CONST-V3R2-003"}) {
		t.Errorf("rule IDs = %v", env.RegistryRuleIDs)
	}
	if !slices.Equal(env.RegistryFrozenFiles, []string{"CLAUDE.md"}) {
		t.Errorf("frozen files = %v, want the one distinct Frozen file", env.RegistryFrozenFiles)
	}
	if env.Policy.Mode != "contract" || !env.Policy.PushDevelop || env.Policy.BudgetDefault.Operations != 40 {
		t.Errorf("policy = %+v", env.Policy)
	}

	// No registry: empty lists, not nil and not an error.
	bare := escalationtest.NewWorktree(t, "t9002")
	env = escalation.LoadVerifyEnv(bare.Root, contractSettings(t, bare))
	if env.RegistryRuleIDs == nil || len(env.RegistryRuleIDs) != 0 || len(env.RegistryFrozenFiles) != 0 {
		t.Errorf("absent registry env = %+v", env)
	}
}

// A linked worktree carries `.git` as a file; the root is found the same way.
func TestFindWorktreeRootAcceptsGitFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "t9001")
	if err := os.MkdirAll(filepath.Join(root, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: /elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := escalation.FindWorktreeRoot(filepath.Join(root, "a", "b"))
	if !ok || got != root {
		t.Errorf("FindWorktreeRoot = %q, %v; want %q", got, ok, root)
	}
}

// A contract whose card field cannot be read is not a claimant; it yields a
// warning naming it, and resolution continues with the rest.
func TestResolverWarnsOnUnreadableCardField(t *testing.T) {
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteMode("contract", false)
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	w.Write(".moai/specs/SPEC-B-001/contract.yaml", "card: [not, a, string]\n")

	res := resolveFrom(t, w)
	if !res.Armed || res.SpecID != "SPEC-A-001" {
		t.Fatalf("armed=%v spec=%q, want armed against SPEC-A-001", res.Armed, res.SpecID)
	}
	if len(res.Lines) != 1 || res.Lines[0].Kind != escalation.LineWarning || res.Lines[0].Cause != escalation.CauseCardUnreadable ||
		!slices.Equal(res.Lines[0].Specs, []string{"SPEC-B-001"}) {
		t.Errorf("lines = %+v, want one card-unreadable warning naming SPEC-B-001", res.Lines)
	}
}

// An unreadable contract is a fault returned as an error (REQ-AE-004 is the
// caller's to apply), never a silent not-armed.
func TestResolverUnreadableContractIsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits do not deny reads on Windows")
	}
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteMode("contract", false)
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	p := w.Path(".moai/specs/SPEC-A-001/contract.yaml")
	if err := os.Chmod(p, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0o644) })
	if f, err := os.Open(p); err == nil {
		_ = f.Close()
		t.Skip("running with privileges that ignore file modes")
	}
	_, err := escalation.Resolve(w.Root, escalation.VerifyEnv{})
	if err == nil || !errors.Is(err, os.ErrPermission) {
		t.Errorf("Resolve error = %v, want a permission fault", err)
	}
}
