package statusline

// agentic_trunk_test.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M3 (card
// t1831). AC-AGT-012: the workflow label renders beside the branch only for
// a valid non-github-flow value, and every github-flow / absent / unreadable
// / invalid state stays byte-identical to the pre-change render with no
// extra git invocation (REQ-PLO-009's cases, pinned here for composition).

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// statuslineGitStrategyFixture writes a git-strategy.yaml naming the given
// workflow (an empty body writes no file) and returns the project root.
func statuslineGitStrategyFixture(t *testing.T, workflow string) string {
	t.Helper()
	root := t.TempDir()
	if workflow == "" {
		return root
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n    mode: manual\n    manual:\n        workflow: " + workflow + "\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// branchFixture is the repoBranchFixture shape the golden strings below pin.
func branchFixture() *StatusData {
	return &StatusData{
		Git: GitStatusData{
			Branch:    "main",
			Available: true,
			Modified:  7,
			Untracked: 21,
		},
	}
}

// goldenBranch is the byte-exact pre-change render of branchFixture: the
// golden every absent/invalid state must reproduce (REQ-AGT-013, REQ-PLO-009
// composition).
const goldenBranch = "🅱️ main +28"

// TestAgenticTrunkStatuslineLabel is AC-AGT-012's label half: a valid
// non-github-flow workflow value — agentic-trunk and git-flow both — renders
// the workflow value in parentheses after the branch segment's dirty count.
func TestAgenticTrunkStatuslineLabel(t *testing.T) {
	t.Run("agentic-trunk label after the dirty count", func(t *testing.T) {
		d := branchFixture()
		d.WorkflowLabel = resolveWorkflowLabel(statuslineGitStrategyFixture(t, "agentic-trunk"))
		if d.WorkflowLabel != "agentic-trunk" {
			t.Fatalf("provider = %q, want agentic-trunk", d.WorkflowLabel)
		}
		want := goldenBranch + " (agentic-trunk)"
		if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
			t.Errorf("render = %q, want %q", got, want)
		}
	})

	t.Run("git-flow label after the dirty count", func(t *testing.T) {
		d := branchFixture()
		d.WorkflowLabel = resolveWorkflowLabel(statuslineGitStrategyFixture(t, "git-flow"))
		if d.WorkflowLabel != "git-flow" {
			t.Fatalf("provider = %q, want git-flow", d.WorkflowLabel)
		}
		want := goldenBranch + " (git-flow)"
		if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
			t.Errorf("render = %q, want %q", got, want)
		}
	})

	t.Run("label precedes nothing — it is the segment tail", func(t *testing.T) {
		d := branchFixture()
		d.WorkflowLabel = "agentic-trunk"
		got := newTestRenderer().renderRepoBranchSegment(d)
		if !strings.HasSuffix(got, " (agentic-trunk)") {
			t.Errorf("render %q does not end with the label tail", got)
		}
		if strings.Contains(got, "(agentic-trunk) (") {
			t.Errorf("render %q carries the label twice", got)
		}
	})
}

// TestAgenticTrunkStatuslineAbsent is AC-AGT-012's byte-identity half: the
// github-flow, absent-file, unreadable-file and invalid-value states render
// byte-equal to the pre-change golden — the label provider contributes
// nothing (and with it REQ-PLO-009's cases stay byte-identical).
func TestAgenticTrunkStatuslineAbsent(t *testing.T) {
	unreadable := t.TempDir()
	dir := filepath.Join(unreadable, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte("git_strategy: [unterminated\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := map[string]string{
		"github-flow":   statuslineGitStrategyFixture(t, "github-flow"),
		"absent file":   statuslineGitStrategyFixture(t, ""),
		"unreadable":    unreadable,
		"invalid value": statuslineGitStrategyFixture(t, "svn-flow"),
		"empty value":   statuslineGitStrategyFixture(t, "''"),
	}
	for name, root := range cases {
		t.Run(name, func(t *testing.T) {
			if got := resolveWorkflowLabel(root); got != "" {
				t.Errorf("provider = %q, want empty", got)
			}
			d := branchFixture()
			d.WorkflowLabel = resolveWorkflowLabel(root)
			if got := newTestRenderer().renderRepoBranchSegment(d); got != goldenBranch {
				t.Errorf("render = %q, want the byte-identical golden %q", got, goldenBranch)
			}
		})
	}

	t.Run("no label on the data renders the golden", func(t *testing.T) {
		if got := newTestRenderer().renderRepoBranchSegment(branchFixture()); got != goldenBranch {
			t.Errorf("render = %q, want %q", got, goldenBranch)
		}
	})
}

// TestAgenticTrunkStatuslineNoExtraGit is AC-AGT-012's no-invocation half:
// inside a real git repository, the provider and the labeled render spawn no
// git process — the shim logs every git invocation and stays empty.
func TestAgenticTrunkStatuslineNoExtraGit(t *testing.T) {
	root := statuslineGitStrategyFixture(t, "agentic-trunk")
	gitOut := func(args ...string) string {
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	gitOut("-C", root, "init", "-q")

	// A git shim that logs every invocation, prepended to PATH.
	shimBin := t.TempDir()
	log := filepath.Join(shimBin, "git-calls.log")
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("locate real git: %v", err)
	}
	shim := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + log + "\nexec " + realGit + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(shimBin, "git"), []byte(shim), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	// The provider reads the yaml through the config reader only.
	for i := 0; i < 3; i++ {
		if got := resolveWorkflowLabel(root); got != "agentic-trunk" {
			t.Fatalf("provider = %q under the shim, want agentic-trunk", got)
		}
	}
	// The labeled render runs the provider again.
	d := branchFixture()
	d.WorkflowLabel = resolveWorkflowLabel(root)
	if got := newTestRenderer().renderRepoBranchSegment(d); !strings.HasSuffix(got, " (agentic-trunk)") {
		t.Fatalf("render = %q, want the labeled tail", got)
	}

	invoked, rerr := os.ReadFile(log)
	if rerr != nil {
		if !os.IsNotExist(rerr) {
			t.Fatalf("read git-call log: %v", rerr)
		}
		invoked = nil
	}
	if len(strings.TrimSpace(string(invoked))) != 0 {
		t.Errorf("the provider/render path invoked git %d time(s): %q — the label reads the yaml, never git",
			len(strings.Split(strings.TrimSpace(string(invoked)), "\n")), invoked)
	}
}
