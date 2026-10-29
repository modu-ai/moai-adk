package cli

// SPEC-TUI-CHECKLIST-001 M0 — AC-TCL-013 baseline characterization.
//
// The checklist surface (internal/tui/checklist.go + the CLI session) is a NEW
// presentation layered over init. Every environment where it must stay off
// (non-TTY, CI, TERM=dumb, --verbose, a replaced writer) keeps today's output,
// byte for byte (REQ-TCL-012). This file pins that output BEFORE any code
// change lands: six fixed runs — init and update, each clean / warn / fail —
// captured through buffer writers (the non-TTY path) and compared with the
// committed goldens under testdata/checklist/fallback/.
//
// The goldens were produced by the unmodified baseline code. The three update
// runs are a regression boundary for the shared files this card edits (spec.md
// C-7), not an update-side checklist. Every injection uses a seam that exists in
// the baseline code; this commit adds no production seam.
//
// Regenerate with UPDATE_GOLDEN=1 only for a deliberate baseline re-capture
// (plan.md section B); a normal run never writes.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/merge"
)

// checklistFallbackGoldenDir resolves the golden directory from this source
// file, so it stays valid after a run moves the process into its fixed project
// directory.
func checklistFallbackGoldenDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "checklist", "fallback")
}

// Normalization applied to every captured stream, in this order and nothing
// else (acceptance.md AC-TCL-013). The ROOT substitution is a fixed-string
// replacement of the test's temp directory; the rest are regular expressions.
var checklistFallbackRules = []struct {
	re   *regexp.Regexp
	repl string
}{
	{regexp.MustCompile(`\b\d{8}_\d{6}\b`), "<TS>"},
	{regexp.MustCompile(`\b\d{8}T\d{6}Z\b`), "<TS>"},
	{regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}[:-]\d{2}[:-]\d{2}Z\b`), "<ISO_TS>"},
	{regexp.MustCompile(`\bv\d+\.\d+\.\d+(-[0-9A-Za-z.]+)?\b`), "<VER>"},
	{regexp.MustCompile(`\bgo1\.\d+(\.\d+)?\b`), "<GO>"},
	{regexp.MustCompile(`\b\d+(\.\d+)?(ms|s)\b`), "<DUR>"},
}

// checklistFallbackNormalize applies the closed normalization list. It never
// adds or removes a line: every rule substitutes inside a line.
func checklistFallbackNormalize(raw, root string) string {
	s := strings.ReplaceAll(raw, root, "<ROOT>")
	for _, r := range checklistFallbackRules {
		s = r.re.ReplaceAllString(s, r.repl)
	}
	return s
}

// checklistFallbackEnv isolates one fixed run: a private home (the user-asset
// ensure writes under it), a private project directory entered through
// t.Chdir (which keeps PWD on the TempDir spelling), and a scrubbed
// environment so the captured bytes do not depend on the developer's shell.
// It returns the TempDir root (replaced by <ROOT>) and the project directory.
func checklistFallbackEnv(t *testing.T) (root, projectDir string) {
	t.Helper()
	root = t.TempDir()
	home := filepath.Join(root, "home")
	projectDir = filepath.Join(root, "proj")
	for _, d := range []string{home, projectDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	requireNotRealHome(t, home)

	origHome := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = origHome })

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CI", "")
	t.Setenv("MOAI_REDUCED_MOTION", "")
	t.Setenv("MOAI_DISTRIBUTE_ALL", "")
	t.Chdir(projectDir)
	return root, projectDir
}

// checklistFallbackInit runs `moai init` non-interactively in the current
// directory through buffer writers and returns the raw stdout, stderr and the
// run error.
func checklistFallbackInit(t *testing.T) (stdout, stderr string, err error) {
	t.Helper()
	cmd := initModeTestCmd()
	for name, val := range map[string]string{
		"non-interactive": "true",
		"language":        "go",
		"mode":            "tdd",
		"no-hooks":        "false",
	} {
		if setErr := cmd.Flags().Set(name, val); setErr != nil {
			t.Fatalf("set --%s: %v", name, setErr)
		}
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	runErr := runInit(cmd, nil)
	return out.String(), errBuf.String(), runErr
}

// checklistFallbackUpdate runs `moai update` in the current directory through
// buffer writers.
func checklistFallbackUpdate(t *testing.T, force, yes bool) (stdout, stderr string, err error) {
	t.Helper()
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = &Dependencies{UpdateChecker: &mockUpdateChecker{}}

	var out, errBuf bytes.Buffer
	updateCmd.SetOut(&out)
	updateCmd.SetErr(&errBuf)
	updateCmd.SetContext(context.Background())
	set := func(name string, v bool) {
		if setErr := updateCmd.Flags().Set(name, fmt.Sprint(v)); setErr != nil {
			t.Fatalf("set --%s: %v", name, setErr)
		}
	}
	set("check", false)
	set("force", force)
	set("yes", yes)
	t.Cleanup(func() {
		set("force", false)
		set("yes", false)
	})
	runErr := updateCmd.RunE(updateCmd, nil)
	return out.String(), errBuf.String(), runErr
}

// checklistFallbackRepoInit makes the fixed project a repository so the hook
// installers take their normal path.
func checklistFallbackRepoInit(t *testing.T, dir string) {
	t.Helper()
	c := exec.Command("git", "init", "-q")
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("repository init: %v\n%s", err, out)
	}
}

// checklistFallbackCompare compares (or, under UPDATE_GOLDEN=1, rewrites) one
// golden and returns the normalized text. A missing golden is a hard failure
// so a deleted baseline cannot pass.
func checklistFallbackCompare(t *testing.T, name, root, raw string) string {
	t.Helper()
	got := checklistFallbackNormalize(raw, root)
	if strings.Count(got, "\n") != strings.Count(raw, "\n") {
		t.Fatalf("%s: normalization changed the line count (%d -> %d)", name,
			strings.Count(raw, "\n"), strings.Count(got, "\n"))
	}
	path := filepath.Join(checklistFallbackGoldenDir(), name+".golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		return got
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("baseline golden %s is missing: %v", path, err)
	}
	if got != string(want) {
		t.Fatalf("%s differs from the baseline golden %s\n%s", name, path, checklistFallbackFirstDiff(string(want), got))
	}
	return got
}

// checklistFallbackFirstDiff names the first differing line as an
// expected/actual pair, so a run-varying value outside the normalization list
// is reported by value rather than hidden.
func checklistFallbackFirstDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return fmt.Sprintf("first difference at line %d\n  expected: %q\n  actual:   %q\n(a run-varying value outside the normalization list is a blocker, not something to normalize silently)", i+1, wl, gl)
		}
	}
	return "outputs differ only in length"
}

// checklistFallbackRequireOnce asserts that frag appears exactly once in s.
func checklistFallbackRequireOnce(t *testing.T, label, s, frag string) {
	t.Helper()
	if n := strings.Count(s, frag); n != 1 {
		t.Errorf("%s: %q appears %d times, want exactly 1", label, frag, n)
	}
}

// checklistFallbackWithError appends the run's returned error to the captured
// stderr the way the command's caller would surface it, so a "fail" golden
// carries the injected failure sentence (the command returns the error rather
// than printing it).
func checklistFallbackWithError(stderr string, err error) string {
	if err == nil {
		return stderr
	}
	return stderr + "[returned error] " + err.Error() + "\n"
}

// checklistFallbackProject builds the fixed update project: the same binary
// initializes an empty directory, then the directory becomes a repository.
func checklistFallbackProject(t *testing.T) (root, projectDir string) {
	t.Helper()
	root, projectDir = checklistFallbackEnv(t)
	if _, stderr, err := checklistFallbackInit(t); err != nil {
		t.Fatalf("fixture init: %v\nstderr:\n%s", err, stderr)
	}
	checklistFallbackRepoInit(t, projectDir)
	return root, projectDir
}

// TestChecklistFallbackMatchesBaseline is AC-TCL-013: the buffer-writer
// (non-TTY) path of six fixed runs equals the committed baseline goldens.
func TestChecklistFallbackMatchesBaseline(t *testing.T) {
	// The runs swap package seams and the process-wide cobra commands, so no
	// subtest is parallel.

	t.Run("init_clean", func(t *testing.T) {
		root, _ := checklistFallbackEnv(t)
		stdout, stderr, err := checklistFallbackInit(t)
		if err != nil {
			t.Fatalf("init: %v\nstderr:\n%s", err, stderr)
		}
		so := checklistFallbackCompare(t, "init-clean-stdout", root, stdout)
		se := checklistFallbackCompare(t, "init-clean-stderr", root, stderr)
		// The five direct-print fragments the checklist must leave in place
		// (spec.md section 3.3 clean-run definition), each exactly once.
		for _, frag := range []string{
			"Deploying core templates only (slim mode).",
			"Common skills and agents install into your user folders.",
			"user-asset ensure: ",
			"shared across concurrent sessions",
		} {
			checklistFallbackRequireOnce(t, "init-clean-stdout", so, frag)
		}
		checklistFallbackRequireOnce(t, "init-clean-stderr", se, "Initializing MoAI project...")
	})

	t.Run("init_warn", func(t *testing.T) {
		root, _ := checklistFallbackEnv(t)
		orig := applyAutonomyTierBundleFn
		applyAutonomyTierBundleFn = func(_, _, _, _ string) error {
			return errors.New("injected tier bundle failure")
		}
		t.Cleanup(func() { applyAutonomyTierBundleFn = orig })

		stdout, stderr, err := checklistFallbackInit(t)
		if err != nil {
			t.Fatalf("init: %v\nstderr:\n%s", err, stderr)
		}
		checklistFallbackCompare(t, "init-warn-stdout", root, stdout)
		se := checklistFallbackCompare(t, "init-warn-stderr", root, stderr)
		if !strings.Contains(se, "Failed to apply autonomy tier bundle") {
			t.Errorf("init-warn-stderr lacks the injected warning sentence")
		}
	})

	t.Run("init_fail", func(t *testing.T) {
		root, _ := checklistFallbackEnv(t)
		injectDeployFailExecutor(t)

		stdout, stderr, err := checklistFallbackInit(t)
		if err == nil {
			t.Fatal("init succeeded; the injected deployment failure did not surface")
		}
		checklistFallbackCompare(t, "init-fail-stdout", root, stdout)
		se := checklistFallbackCompare(t, "init-fail-stderr", root, checklistFallbackWithError(stderr, err))
		if !strings.Contains(se, "template deployment") {
			t.Errorf("init-fail-stderr lacks the injected failure sentence")
		}
	})

	t.Run("update_clean", func(t *testing.T) {
		root, _ := checklistFallbackProject(t)
		stdout, stderr, err := checklistFallbackUpdate(t, true, true)
		if err != nil {
			t.Fatalf("update: %v\nstderr:\n%s", err, stderr)
		}
		so := checklistFallbackCompare(t, "update-clean-stdout", root, stdout)
		checklistFallbackCompare(t, "update-clean-stderr", root, stderr)
		// Positive control: the run reached the shared hook-install lines.
		checklistFallbackRequireOnce(t, "update-clean-stdout", so, "Pre-push hook installed")
	})

	t.Run("update_warn", func(t *testing.T) {
		root, projectDir := checklistFallbackProject(t)
		hook := filepath.Join(projectDir, ".git", "hooks", "pre-push")
		f, err := os.OpenFile(hook, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			t.Fatalf("open hook: %v", err)
		}
		if _, err := f.WriteString("\n# local edit by the fixture\n"); err != nil {
			t.Fatalf("append hook: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("close hook: %v", err)
		}

		stdout, stderr, runErr := checklistFallbackUpdate(t, true, true)
		if runErr != nil {
			t.Fatalf("update: %v\nstderr:\n%s", runErr, stderr)
		}
		checklistFallbackCompare(t, "update-warn-stdout", root, stdout)
		se := checklistFallbackCompare(t, "update-warn-stderr", root, stderr)
		if !strings.Contains(se, "user-modified pre-push hook was replaced") {
			t.Errorf("update-warn-stderr lacks the injected warning sentence")
		}
	})

	t.Run("update_fail", func(t *testing.T) {
		root, _ := checklistFallbackProject(t)
		orig := confirmViaPreviewFn
		confirmViaPreviewFn = func(_ merge.MergeAnalysis, _ string) (bool, error) {
			return false, errors.New("injected confirmation failure")
		}
		t.Cleanup(func() { confirmViaPreviewFn = orig })

		stdout, stderr, err := checklistFallbackUpdate(t, true, false)
		if err == nil {
			t.Fatal("update succeeded; the injected confirmation failure did not surface")
		}
		checklistFallbackCompare(t, "update-fail-stdout", root, stdout)
		se := checklistFallbackCompare(t, "update-fail-stderr", root, checklistFallbackWithError(stderr, err))
		if !strings.Contains(se, "confirm merge for") {
			t.Errorf("update-fail-stderr lacks the injected failure sentence")
		}
	})
}
