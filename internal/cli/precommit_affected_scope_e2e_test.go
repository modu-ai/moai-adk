//go:build !windows

package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// precommit_affected_scope_e2e_test.go — AC-PAS-016: the affected test scope
// end to end through a REAL git pre-commit hook and the `moai` binary built
// from this tree (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-001 / -010).
//
// The throwaway module has three packages: a; b importing a; c standalone, each
// with a test. The hook exports MOAI_PRECOMMIT=1 and runs `moai gate`; only the
// Go test step is on. The hook is the repository's own .git/hooks/pre-commit —
// no git configuration is written and no shared configuration is touched.

const scopeE2EGateYAML = `gate:
  enabled: true
  disabled_steps:
    "go vet": false
    "golangci-lint": false
  typecheck:
    enabled: false
  pre_commit:
    enabled: true
  ast_grep_gate:
    enabled: false
  graph_freshness:
    enabled: false
`

// scopeE2EGit runs git in dir and returns combined output and the exit error.
func scopeE2EGit(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func scopeE2EMustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := scopeE2EGit(t, dir, args...)
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(out)
}

func scopeE2EWrite(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// newScopeE2ERepo builds the three-package module as a git repository with one
// base commit made before any hook exists.
func newScopeE2ERepo(t *testing.T, bTestBody string) string {
	t.Helper()
	root := t.TempDir()
	scopeE2EMustGit(t, root, "init", "-q")
	scopeE2EWrite(t, root, map[string]string{
		"go.mod":                          "module example.com/m\n\ngo 1.23\n",
		"a/a.go":                          "package a\n\nfunc A() int { return 1 }\n",
		"a/a_test.go":                     "package a\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n",
		"b/b.go":                          "package b\n\nimport \"example.com/m/a\"\n\nfunc B() int { return a.A() }\n",
		"b/b_test.go":                     bTestBody,
		"c/c.go":                          "package c\n\nfunc C() int { return 3 }\n",
		"c/c_test.go":                     "package c\n\nimport (\n\t\"os\"\n\t\"testing\"\n)\n\nfunc TestC(t *testing.T) {\n\tif p := os.Getenv(\"MOAI_E2E_MARKER\"); p != \"\" {\n\t\t_ = os.WriteFile(p, []byte(\"ran\"), 0o644)\n\t}\n}\n",
		".moai/config/sections/gate.yaml": scopeE2EGateYAML,
	})
	scopeE2EMustGit(t, root, "add", "-A")
	scopeE2EMustGit(t, root, "-c", "user.email=e2e@example.test", "-c", "user.name=e2e", "commit", "-q", "-m", "base")
	return root
}

// installScopeE2EHook installs the hook: MOAI_PRECOMMIT=1 moai gate, with
// extraPath ahead of PATH and the marker file location exported.
func installScopeE2EHook(t *testing.T, root, moai, extraPath, marker string) {
	t.Helper()
	pathPrefix := ""
	if extraPath != "" {
		pathPrefix = extraPath + string(os.PathListSeparator)
	}
	script := "#!/bin/sh\n" +
		"MOAI_E2E_MARKER='" + marker + "' MOAI_PRECOMMIT=1 CLAUDE_PROJECT_DIR='" + root + "' " +
		"PATH='" + pathPrefix + "'\"$PATH\" '" + moai + "' gate\n"
	hook := filepath.Join(root, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

// recordingGoWrapper writes a `go` on PATH that delegates every verb except
// test to the real go (so the gate's go list query works) and records only
// test invocations. It returns the wrapper directory and the log path.
func recordingGoWrapper(t *testing.T) (string, string) {
	t.Helper()
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("go not found: %v", err)
	}
	dir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "go-test.log")
	script := "#!/bin/sh\ncase \"$1\" in\n  test) echo \"$@\" >> '" + logPath + "'; exit 0 ;;\n  *) exec '" + realGo + "' \"$@\" ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(dir, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir, logPath
}

// TestPreCommitAffectedScopeE2E is AC-PAS-016 (release-blocking).
func TestPreCommitAffectedScopeE2E(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	t.Setenv("GOPROXY", "off")
	moai := buildMoaiBinary(t)

	t.Run("a change to a runs exactly a and b", func(t *testing.T) {
		root := newScopeE2ERepo(t, "package b\n\nimport \"testing\"\n\nfunc TestB(t *testing.T) {}\n")
		wrapDir, logPath := recordingGoWrapper(t)
		installScopeE2EHook(t, root, moai, wrapDir, filepath.Join(t.TempDir(), "marker"))

		scopeE2EWrite(t, root, map[string]string{"a/a.go": "package a\n\nfunc A() int { return 2 }\n"})
		scopeE2EMustGit(t, root, "add", "a/a.go")
		head0 := scopeE2EMustGit(t, root, "rev-parse", "HEAD")
		out, err := scopeE2EGit(t, root, "-c", "user.email=e2e@example.test", "-c", "user.name=e2e", "commit", "-q", "-m", "change a")
		if err != nil {
			t.Fatalf("the commit was refused: %v\n%s", err, out)
		}
		if head1 := scopeE2EMustGit(t, root, "rev-parse", "HEAD"); head1 == head0 {
			t.Fatal("HEAD did not move; the commit did not happen")
		}
		data, _ := os.ReadFile(logPath)
		got := strings.Join(slices.DeleteFunc(strings.Split(strings.TrimSpace(string(data)), "\n"), func(s string) bool { return s == "" }), ";")
		if want := "test ./a ./b"; got != want {
			t.Errorf("recorded go test calls = %q, want exactly %q\ncommit output:\n%s", got, want, out)
		}
	})

	t.Run("a failing dependent blocks the commit and the unaffected package never runs", func(t *testing.T) {
		root := newScopeE2ERepo(t, "package b\n\nimport \"testing\"\n\nfunc TestB(t *testing.T) { t.Fatal(\"b is broken\") }\n")
		marker := filepath.Join(t.TempDir(), "c-ran")
		installScopeE2EHook(t, root, moai, "", marker)

		scopeE2EWrite(t, root, map[string]string{"a/a.go": "package a\n\nfunc A() int { return 2 }\n"})
		scopeE2EMustGit(t, root, "add", "a/a.go")
		head0 := scopeE2EMustGit(t, root, "rev-parse", "HEAD")
		out, err := scopeE2EGit(t, root, "-c", "user.email=e2e@example.test", "-c", "user.name=e2e", "commit", "-q", "-m", "change a")
		if err == nil {
			t.Fatalf("the commit succeeded although b's test fails:\n%s", out)
		}
		if head1 := scopeE2EMustGit(t, root, "rev-parse", "HEAD"); head1 != head0 {
			t.Error("HEAD moved on a refused commit")
		}
		if !strings.Contains(out, "example.com/m/b") {
			t.Errorf("the output does not name b as failing:\n%s", out)
		}
		if _, statErr := os.Stat(marker); statErr == nil {
			t.Error("c's test ran: the marker file it writes exists; the unaffected package was tested")
		}
	})
}
