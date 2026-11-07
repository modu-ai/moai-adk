package runtime

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gatesGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func gatesWriteWorkflow(t *testing.T, root, codexGate string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n  audit:\n    gates:\n      codex: " + codexGate + "\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A linked worktree whose own workflow.yaml still says `required` reads the
// gate set of the primary checkout, which says `advisory`.
func TestResolveRequiredBackends_LinkedWorktreeReadsPrimary(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(cfg, []byte("[user]\n\tname = T\n\temail = t@example.invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	primary := filepath.Join(base, "primary")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatal(err)
	}
	gatesGit(t, primary, "init", "-q", "-b", "main")
	gatesWriteWorkflow(t, primary, "required")
	gatesGit(t, primary, "add", "-A")
	gatesGit(t, primary, "commit", "-q", "-m", "base")
	linked := filepath.Join(base, "linked")
	gatesGit(t, primary, "worktree", "add", "-q", "-b", "card", linked)

	// The primary moves to advisory after the worktree was cut.
	gatesWriteWorkflow(t, primary, "advisory")

	gs, err := ResolveRequiredBackends(linked)
	if err != nil {
		t.Fatalf("ResolveRequiredBackends(linked): %v", err)
	}
	if len(gs.Required) != 0 {
		t.Fatalf("linked worktree read its own stale gates, want the primary's advisory: %v", gs.Required)
	}

	// Control: the stale worktree copy alone does say required.
	gatesWriteWorkflow(t, linked, "required")
	if gs, err = ResolveRequiredBackends(linked); err != nil || len(gs.Required) != 0 {
		t.Fatalf("worktree copy overrode the primary: %v %v", gs, err)
	}
	gatesWriteWorkflow(t, primary, "required")
	gs, err = ResolveRequiredBackends(linked)
	if err != nil || strings.Join(gs.Required, ",") != "codex" {
		t.Fatalf("primary required not honored: %v %v", gs, err)
	}
}

// A root that is not inside any repository falls back to its own section.
func TestResolveRequiredBackends_NoPrimaryFallsBackToRoot(t *testing.T) {
	root := t.TempDir()
	gatesWriteWorkflow(t, root, "required")
	gs, err := ResolveRequiredBackends(root)
	if err != nil || strings.Join(gs.Required, ",") != "codex" {
		t.Fatalf("fallback to the given root failed: %v %v", gs, err)
	}
}
