//go:build repro_t2034

package cli

// repro_t2034_3_test.go — card t2034 claim 3 (reproduce only, never default-run).
//
// Claim (card text, "go.mod만 바꾼 카드의 검사 누락"): a card whose only change
// is go.mod is not covered by the fast required check. Scenario: a temporary
// Go module with a package; branch `main` is the release tip; the card branch
// changes only go.mod (variants: only go.sum, only go.work). Expected-correct:
// the lint and test steps of factoryFastCheckSteps cover the whole module
// ("./..."), not nothing. A FAIL means REPRODUCED.

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func repro3Git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@e", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func repro3Write(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// repro3Repo builds main with a module root (go.mod, go.sum) and pkg/a, then a
// card branch with `change` applied.
func repro3Repo(t *testing.T, change func(dir string)) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	repro3Git(t, dir, "init", "-q", "-b", "main")
	repro3Write(t, dir, "go.mod", "module example.com/m\n\ngo 1.23\n")
	repro3Write(t, dir, "go.sum", "")
	repro3Write(t, dir, "pkg/a/a.go", "package a\n")
	repro3Git(t, dir, "add", "go.mod", "go.sum", "pkg/a/a.go")
	repro3Git(t, dir, "commit", "-qm", "base")
	repro3Git(t, dir, "checkout", "-qb", "card")
	change(dir)
	repro3Git(t, dir, "add", "-A")
	repro3Git(t, dir, "commit", "-qm", "card")
	return dir
}

func repro3Covers(steps [][]string, want string) bool {
	if len(steps) < 3 {
		return false
	}
	return slices.Contains(steps[1], want) && slices.Contains(steps[2], want)
}

func TestRepro3GoModOnlyCard(t *testing.T) {
	cases := []struct {
		name    string
		change  func(dir string)
		want    string // package argument both lint and test steps must carry
		control bool
	}{
		{"claim: only go.mod changed", func(d string) {
			repro3Write(t, d, "go.mod", "module example.com/m\n\ngo 1.23\n\nrequire example.com/x v1.0.0\n")
		}, "./...", false},
		{"variant: only go.sum changed", func(d string) { repro3Write(t, d, "go.sum", "example.com/x v1.0.0 h1:abc=\n") }, "./...", false},
		{"variant: only go.work added", func(d string) { repro3Write(t, d, "go.work", "go 1.23\n\nuse .\n") }, "./...", false},
		{"control: only pkg/a/a.go changed (narrow scope expected)", func(d string) { repro3Write(t, d, "pkg/a/a.go", "package a\n\nvar X = 1\n") }, "./pkg/a", true},
	}
	swept := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			swept++
			dir := repro3Repo(t, tc.change)
			steps, err := factoryFastCheckSteps(dir, "main")
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("steps: %s", joinSteps(steps))
			if !repro3Covers(steps, tc.want) {
				t.Errorf("lint/test steps do not carry %q: %s", tc.want, joinSteps(steps))
			}
			if tc.control && slices.Contains(steps[1], "./...") {
				t.Errorf("control widened to ./... although only a package file changed: %s", joinSteps(steps))
			}
		})
	}
	if swept != len(cases) {
		t.Fatalf("swept %d of %d", swept, len(cases))
	}
}

// TestRepro3RedObservedOnKnownBadInput shows the assertion can go red: a card
// with NO change against the release tip yields a build-only plan, which the
// repro3Covers predicate must reject.
func TestRepro3RedObservedOnKnownBadInput(t *testing.T) {
	dir := repro3Repo(t, func(d string) { repro3Write(t, d, "notes.txt", "x\n") })
	steps, err := factoryFastCheckSteps(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if repro3Covers(steps, "./...") {
		t.Fatalf("predicate accepted a build-only plan: %s", joinSteps(steps))
	}
	t.Logf("known-bad plan (non-Go change only): %s", joinSteps(steps))
}

func joinSteps(steps [][]string) string {
	var parts []string
	for _, s := range steps {
		parts = append(parts, strings.Join(s, " "))
	}
	return strings.Join(parts, " | ")
}
