package bodp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeGit overrides gitCommand for the duration of a test.
//
// Tests using fakeGit / fakeGh MUST NOT call t.Parallel() because the
// override mutates package-level state.
func fakeGit(t *testing.T, fn func(args ...string) (string, error)) {
	t.Helper()
	orig := gitCommand
	gitCommand = fn
	t.Cleanup(func() { gitCommand = orig })
}

func fakeGh(t *testing.T, fn func(args ...string) (string, error)) {
	t.Helper()
	orig := ghCommand
	ghCommand = fn
	t.Cleanup(func() { ghCommand = orig })
}

// writeSpecWithDependsOn writes a minimal SPEC fixture with the supplied
// depends_on list at .moai/specs/<specID>/spec.md under repoRoot.
func writeSpecWithDependsOn(t *testing.T, repoRoot, specID string, dependsOn []string) {
	t.Helper()
	dir := filepath.Join(repoRoot, ".moai", "specs", specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	body := "---\nid: " + specID + "\ndepends_on: [" + strings.Join(dependsOn, ", ") + "]\n---\n# spec\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
}

// TestRelatedness_AllNegative_RecommendsMain replays the canonical AC-CIAUT-018
// fixture: chore branch + new unrelated SPEC + clean working tree + no open PR
// → ChoiceMain @ origin/main.
func TestRelatedness_AllNegative_RecommendsMain(t *testing.T) {
	fakeGit(t, func(args ...string) (string, error) {
		return "", nil
	})
	fakeGh(t, func(args ...string) (string, error) {
		return "[]\n", nil
	})

	got, err := Check(CheckInput{
		CurrentBranch: "chore/translation-batch-b",
		NewSpecID:     "SPEC-V3R3-CI-AUTONOMY-001",
		RepoRoot:      t.TempDir(),
		EntryPoint:    EntryPlanBranch,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if got.SignalA || got.SignalB || got.SignalC {
		t.Errorf("expected all signals false, got A=%t B=%t C=%t", got.SignalA, got.SignalB, got.SignalC)
	}
	if got.Recommended != ChoiceMain {
		t.Errorf("expected ChoiceMain, got %v", got.Recommended)
	}
	if got.BaseBranch != "origin/main" {
		t.Errorf("expected BaseBranch origin/main, got %q", got.BaseBranch)
	}
	if got.Rationale == "" || !strings.Contains(got.Rationale, "main") {
		t.Errorf("expected Rationale mentions main, got %q", got.Rationale)
	}
}

// TestRelatedness_SignalA_RecommendsStacked replays canonical AC-CIAUT-019:
// new SPEC declares depends_on: [SPEC-AUTH-001] and current branch name
// contains that SPEC ID → ChoiceStacked @ currentBranch.
func TestRelatedness_SignalA_RecommendsStacked(t *testing.T) {
	fakeGit(t, func(args ...string) (string, error) {
		return "", nil
	})
	fakeGh(t, func(args ...string) (string, error) {
		return "[]\n", nil
	})

	repoRoot := t.TempDir()
	writeSpecWithDependsOn(t, repoRoot, "SPEC-AUTH-002", []string{"SPEC-AUTH-001"})

	got, err := Check(CheckInput{
		CurrentBranch: "feat/SPEC-AUTH-001-base",
		NewSpecID:     "SPEC-AUTH-002",
		RepoRoot:      repoRoot,
		EntryPoint:    EntryPlanBranch,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !got.SignalA {
		t.Errorf("expected SignalA=true, got false")
	}
	if got.Recommended != ChoiceStacked {
		t.Errorf("expected ChoiceStacked, got %v", got.Recommended)
	}
	if got.BaseBranch != "feat/SPEC-AUTH-001-base" {
		t.Errorf("expected BaseBranch feat/SPEC-AUTH-001-base, got %q", got.BaseBranch)
	}
	if !strings.Contains(got.Rationale, "stacked") && !strings.Contains(got.Rationale, "의존성") {
		t.Errorf("expected Rationale mentions stacked or 의존성, got %q", got.Rationale)
	}
}

// TestRelatedness_SignalB_RecommendsContinue: untracked .moai/specs/<NewSpecID>/
// in working tree → ChoiceContinue (current branch is the right place).
func TestRelatedness_SignalB_RecommendsContinue(t *testing.T) {
	fakeGit(t, func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "status" {
			return "?? .moai/specs/SPEC-FOO-001/\n", nil
		}
		return "", nil
	})
	fakeGh(t, func(args ...string) (string, error) {
		return "[]\n", nil
	})

	got, err := Check(CheckInput{
		CurrentBranch: "feat/foo",
		NewSpecID:     "SPEC-FOO-001",
		RepoRoot:      t.TempDir(),
		EntryPoint:    EntryPlanWorktree,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !got.SignalB {
		t.Errorf("expected SignalB=true, got false")
	}
	if got.Recommended != ChoiceContinue {
		t.Errorf("expected ChoiceContinue, got %v", got.Recommended)
	}
	if got.BaseBranch != "" {
		t.Errorf("expected BaseBranch empty for continue, got %q", got.BaseBranch)
	}
}

// TestRelatedness_SignalC_RecommendsStackedWithGotchaWarning: open PR exists
// with current branch as head → ChoiceStacked + parent-merge gotcha rationale
// (REQ-CIAUT-047b verbatim).
func TestRelatedness_SignalC_RecommendsStackedWithGotchaWarning(t *testing.T) {
	fakeGit(t, func(args ...string) (string, error) {
		return "", nil
	})
	fakeGh(t, func(args ...string) (string, error) {
		return `[{"number":42}]`, nil
	})

	got, err := Check(CheckInput{
		CurrentBranch: "feat/X",
		NewSpecID:     "SPEC-Y-001",
		RepoRoot:      t.TempDir(),
		EntryPoint:    EntryPlanBranch,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !got.SignalC {
		t.Errorf("expected SignalC=true, got false")
	}
	if got.Recommended != ChoiceStacked {
		t.Errorf("expected ChoiceStacked, got %v", got.Recommended)
	}
	if got.BaseBranch != "feat/X" {
		t.Errorf("expected BaseBranch feat/X, got %q", got.BaseBranch)
	}
	if !strings.Contains(got.Rationale, "parent-merge gotcha") && !strings.Contains(got.Rationale, "§18.11") {
		t.Errorf("expected Rationale contains parent-merge gotcha or §18.11, got %q", got.Rationale)
	}
}

// TestRelatedness_SignalsAandC_RecommendsStacked: signal A + signal C both
// positive → ChoiceStacked @ currentBranch (decision matrix row 5).
func TestRelatedness_SignalsAandC_RecommendsStacked(t *testing.T) {
	fakeGit(t, func(args ...string) (string, error) {
		return "", nil
	})
	fakeGh(t, func(args ...string) (string, error) {
		return `[{"number":99}]`, nil
	})

	repoRoot := t.TempDir()
	writeSpecWithDependsOn(t, repoRoot, "SPEC-A-002", []string{"SPEC-A-001"})

	got, err := Check(CheckInput{
		CurrentBranch: "feat/SPEC-A-001-base",
		NewSpecID:     "SPEC-A-002",
		RepoRoot:      repoRoot,
		EntryPoint:    EntryPlanBranch,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !got.SignalA || !got.SignalC {
		t.Errorf("expected SignalA=true, SignalC=true, got A=%t C=%t", got.SignalA, got.SignalC)
	}
	if got.Recommended != ChoiceStacked {
		t.Errorf("expected ChoiceStacked, got %v", got.Recommended)
	}
}

// TestExtractFrontmatter covers the YAML frontmatter slicing edge cases.
func TestExtractFrontmatter(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{
			name: "valid frontmatter",
			raw:  "---\nid: SPEC-X\ndepends_on: [SPEC-Y]\n---\n# body\n",
			want: "id: SPEC-X\ndepends_on: [SPEC-Y]",
			ok:   true,
		},
		{
			name: "no leading delimiter",
			raw:  "# just a body\nid: SPEC-X\n",
			ok:   false,
		},
		{
			name: "no closing delimiter",
			raw:  "---\nid: SPEC-X\n",
			ok:   false,
		},
		{
			name: "CRLF line endings normalize",
			raw:  "---\r\nid: SPEC-X\r\n---\r\n# body\r\n",
			want: "id: SPEC-X",
			ok:   true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := extractFrontmatter([]byte(c.raw))
			if ok != c.ok {
				t.Fatalf("ok = %t, want %t", ok, c.ok)
			}
			if c.ok && string(got) != c.want {
				t.Errorf("frontmatter = %q, want %q", got, c.want)
			}
		})
	}
}

// TestParseDependsOn_FileMissing covers the missing-file error path.
func TestParseDependsOn_FileMissing(t *testing.T) {
	repoRoot := t.TempDir()
	if _, err := parseDependsOn(repoRoot, "SPEC-NONE-001"); err == nil {
		t.Fatalf("expected error for missing spec file, got nil")
	}
}

// TestParseDependsOn_NoFrontmatter covers the missing-frontmatter error path.
func TestParseDependsOn_NoFrontmatter(t *testing.T) {
	repoRoot := t.TempDir()
	dir := filepath.Join(repoRoot, ".moai", "specs", "SPEC-PLAIN-001")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("# title only\nbody\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := parseDependsOn(repoRoot, "SPEC-PLAIN-001"); err == nil {
		t.Fatalf("expected error for missing frontmatter, got nil")
	}
}

// TestRelatedness_GhCommandUnavailable_GracefulSkip: gh missing → SignalC=false
// without aborting Check; other signals still evaluated normally.
func TestRelatedness_GhCommandUnavailable_GracefulSkip(t *testing.T) {
	fakeGit(t, func(args ...string) (string, error) {
		return "", nil
	})
	fakeGh(t, func(args ...string) (string, error) {
		return "", errors.New("exec: \"gh\": executable file not found in $PATH")
	})

	got, err := Check(CheckInput{
		CurrentBranch: "feat/X",
		NewSpecID:     "SPEC-X-001",
		RepoRoot:      t.TempDir(),
		EntryPoint:    EntryPlanBranch,
	})
	if err != nil {
		t.Fatalf("Check returned error on gh failure (graceful skip expected): %v", err)
	}
	if got.SignalC {
		t.Errorf("expected SignalC=false on gh failure (graceful skip), got true")
	}
	if got.Recommended != ChoiceMain {
		t.Errorf("expected ChoiceMain when all signals negative, got %v", got.Recommended)
	}
}
