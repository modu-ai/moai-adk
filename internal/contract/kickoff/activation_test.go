package kickoff_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract/kickoff"
	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// repoRoot is the repository holding this package (the directory with go.mod).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func git(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(gitenv.Env(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// firstCommit is the oldest commit whose diff adds or removes token in path
// ("" when none).
func firstCommit(t *testing.T, repo, token, path string) string {
	t.Helper()
	out, err := git(t, repo, "log", "--reverse", "--format=%H", "-S"+token, "--", path)
	if err != nil || out == "" {
		return ""
	}
	return strings.Split(out, "\n")[0]
}

// firstAdded is the oldest commit that added a file under dir ("" when none).
func firstAdded(t *testing.T, repo, dir string) string {
	t.Helper()
	out, err := git(t, repo, "log", "--reverse", "--diff-filter=A", "--format=%H", "--", dir)
	if err != nil || out == "" {
		return ""
	}
	return strings.Split(out, "\n")[0]
}

const enabledToken = "autonomousKickoffEnabled = true"

// orderFindings checks that the commit turning autonomous Kickoff on is a
// strict descendant of the commits that first added revoke and the store.
func orderFindings(t *testing.T, repo string) []string {
	t.Helper()
	on := firstCommit(t, repo, enabledToken, "internal/contract/kickoff/kickoff.go")
	if on == "" {
		return []string{"the activation commit is not in the history"}
	}
	var f []string
	for _, dir := range []string{"internal/contract/revoke", "internal/contract/receipt"} {
		c := firstAdded(t, repo, dir)
		switch c {
		case "":
			f = append(f, dir+" was never added")
		case on:
			f = append(f, dir+" landed in the activation commit itself")
		default:
			if _, err := git(t, repo, "merge-base", "--is-ancestor", c, on); err != nil {
				f = append(f, dir+" is not an ancestor of the activation commit")
			}
		}
	}
	return f
}

// TestAutonomousKickoffActivationOrder (AC-GR-017 order half).
func TestAutonomousKickoffActivationOrder(t *testing.T) {
	t.Run("falsifier/same-commit", func(t *testing.T) {
		repo := t.TempDir()
		for _, a := range [][]string{{"init", "-q"}, {"config", "user.name", "f"}, {"config", "user.email", "f@example.com"}} {
			if _, err := git(t, repo, a...); err != nil {
				t.Fatal(err)
			}
		}
		for rel, body := range map[string]string{
			"internal/contract/revoke/r.go":        "package revoke\n",
			"internal/contract/receipt/s.go":       "package receipt\n",
			"internal/contract/kickoff/kickoff.go": "package kickoff\n\nconst " + enabledToken + "\n",
		} {
			p := filepath.Join(repo, filepath.FromSlash(rel))
			_ = os.MkdirAll(filepath.Dir(p), 0o755)
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := git(t, repo, "add", "-A"); err != nil {
			t.Fatal(err)
		}
		if _, err := git(t, repo, "commit", "-q", "-m", "all at once"); err != nil {
			t.Fatal(err)
		}
		if f := orderFindings(t, repo); len(f) == 0 {
			t.Fatal("order check accepted activation in the same commit as its conditions")
		} else {
			t.Logf("observed: %v", f)
		}
	})
	t.Run("tree", func(t *testing.T) {
		if !kickoff.AutonomousKickoffEnabled() {
			t.Logf("autonomousKickoffEnabled = false; checking that llm and llm+jev signatures are refused")
			for _, c := range []struct{ decider, signer string }{{"llm", "llm"}, {"llm+jev", "llm+jev"}} {
				f := newFx(t)
				var mut = jevApprove
				if c.decider == "llm" {
					mut = nil
				}
				f.signReceipt(c.decider, c.signer, mut, true)
				in := f.checkIn(c.decider, kickoff.AutonomousKickoffEnabled())
				res, err := kickoff.Check(in)
				if err != nil {
					t.Fatal(err)
				}
				if res.Pass || res.Reason != kickoff.ReasonInactive {
					t.Errorf("%s signature under the compiled state: pass=%v reason=%q, want inactive", c.decider, res.Pass, res.Reason)
				}
			}
			return
		}
		for _, f := range orderFindings(t, repoRoot(t)) {
			t.Error(f)
		}
	})
}

// linkageMarker is one location the Jev doctrine amendment touches, and the
// token that shows it amended.
type linkageMarker struct{ path, token string }

var linkageMarkers = []linkageMarker{
	{".moai/specs/SPEC-JEV-CORE-001/spec.md", "[AMENDED 2026-09-26"},
	{".claude/rules/moai/core/moai-mcp-tools-catalogue.md", "contract-mode Kickoff"},
	{"internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md", "contract-mode Kickoff"},
	{".moai/config/sections/workflow.yaml", "contract-mode Kickoff"},
	{"internal/template/templates/.moai/config/sections/workflow.yaml", "contract-mode Kickoff"},
	{"CLAUDE.local.md", "`moai contract decide` 가 Jev 를 두 번째 신호로"},
	{"internal/contract/kickoff/kickoff.go", "JevDoctrineAmended = true"},
}

// linkageFindings requires the amendment markers and the doctrine constant to
// be all present or all absent, and — when present — to first appear in one
// commit.
func linkageFindings(t *testing.T, repo string) []string {
	t.Helper()
	var present, absent []string
	for _, m := range linkageMarkers {
		data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(m.path)))
		if err == nil && strings.Contains(string(data), m.token) {
			present = append(present, m.path)
		} else {
			absent = append(absent, m.path)
		}
	}
	if len(present) == 0 {
		return nil
	}
	var f []string
	if len(absent) > 0 {
		f = append(f, "partial amendment: present in "+strings.Join(present, ", ")+"; absent in "+strings.Join(absent, ", "))
	}
	first := ""
	for _, m := range linkageMarkers {
		c := firstCommit(t, repo, m.token, m.path)
		if c == "" {
			continue
		}
		if first == "" {
			first = c
		} else if c != first {
			f = append(f, "marker in "+m.path+" first appears in "+c+", not "+first)
		}
	}
	return f
}

// TestJevAmendmentLinkage (AC-GR-017 linkage half).
func TestJevAmendmentLinkage(t *testing.T) {
	fixture := func(t *testing.T, skip func(linkageMarker) bool) string {
		repo := t.TempDir()
		for _, a := range [][]string{{"init", "-q"}, {"config", "user.name", "f"}, {"config", "user.email", "f@example.com"}} {
			if _, err := git(t, repo, a...); err != nil {
				t.Fatal(err)
			}
		}
		for _, m := range linkageMarkers {
			body := "base\n"
			if !skip(m) {
				body += m.token + "\n"
			}
			p := filepath.Join(repo, filepath.FromSlash(m.path))
			_ = os.MkdirAll(filepath.Dir(p), 0o755)
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := git(t, repo, "add", "-A"); err != nil {
			t.Fatal(err)
		}
		if _, err := git(t, repo, "commit", "-q", "-m", "fixture"); err != nil {
			t.Fatal(err)
		}
		return repo
	}
	constant := func(m linkageMarker) bool { return m.path == "internal/contract/kickoff/kickoff.go" }
	falsifiers := map[string]func(linkageMarker) bool{
		"markers-only":   constant,
		"constant-only":  func(m linkageMarker) bool { return !constant(m) },
		"all-but-sec-29": func(m linkageMarker) bool { return m.path == "CLAUDE.local.md" },
	}
	for name, skip := range falsifiers {
		t.Run("falsifier/"+name, func(t *testing.T) {
			if f := linkageFindings(t, fixture(t, skip)); len(f) == 0 {
				t.Fatalf("linkage check accepted %s", name)
			} else {
				t.Logf("observed: %v", f)
			}
		})
	}
	t.Run("all-in-one-commit", func(t *testing.T) {
		if f := linkageFindings(t, fixture(t, func(linkageMarker) bool { return false })); len(f) != 0 {
			t.Fatalf("complete linkage rejected: %v", f)
		}
	})
	t.Run("tree", func(t *testing.T) {
		t.Logf("JevDoctrineAmended = %v", kickoff.JevDoctrineAmended)
		for _, f := range linkageFindings(t, repoRoot(t)) {
			t.Error(f)
		}
	})
}
