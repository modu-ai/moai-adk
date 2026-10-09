package cli

// integration_remeasure_cmd_test.go — t1576 review rounds 2-3: the remeasure
// verb's verdict-to-exit mapping, its execution tree, and the integration
// surface regressions the review rounds landed. The record and the
// diagnostics land either way; an INVALID verdict returns the validation
// error so the exit code carries it (the INVALID print alone exited 0, and
// a calling script read the re-measure as passed).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestRemeasureVerdictErrorCarriesInvalid(t *testing.T) {
	valid := &factory.RemeasureRecord{
		Tree: "abc123def4567890abc123def4567890abc12345", Base: "b0" + strings.Repeat("0", 38),
		Command: "true", ExitCode: 0, BuildIdentity: "moai test",
	}
	verdict, err := remeasureVerdictError(valid)
	if err != nil || verdict != "valid" {
		t.Fatalf("a valid record reads valid: %q err=%v", verdict, err)
	}

	failed := *valid
	failed.ExitCode = 3
	verdict, err = remeasureVerdictError(&failed)
	if err == nil {
		t.Fatalf("an exit-3 record's INVALID verdict must return the error")
	}
	if !strings.Contains(err.Error(), "abc123def456") {
		t.Fatalf("the error must name the tree: %v", err)
	}
	if !strings.HasPrefix(verdict, "INVALID:") {
		t.Fatalf("the verdict line still carries INVALID for the record: %q", verdict)
	}
}

func TestRemeasureRunsInTheCallerWorktree(t *testing.T) {
	// t1576 review round 3: the verb ran the re-measure in resolveProjectDir
	// — a card worktree session's CLAUDE_PROJECT_DIR names the PRIMARY
	// checkout, so the record keyed the primary's tree while the card's
	// tree carried no record. The execution tree is the caller's working
	// directory; the project dir stays the record store root only.
	t.Setenv("CLAUDE_PROJECT_DIR", filepath.Join(t.TempDir(), "primary"))
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got, err := remeasureWorktreeDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != wd {
		t.Fatalf("the re-measure must run in the caller's worktree %s, got %s", wd, got)
	}
}

func TestConfiguredIntegrationBranchReadsFlowScopedTarget(t *testing.T) {
	// t1576 review round 3: under `workflow: github-flow` DevelopBranch is
	// empty by design and the base resolution died with `merge-base ..HEAD
	// … exit status 128`. The flow-scoped IntegrationTarget (main) is what
	// the re-measure names as its absorbed base.
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n" +
		"    mode: manual\n" +
		"    manual:\n" +
		"        workflow: github-flow\n" +
		"        develop_branch: fixture-integration\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	if got := configuredIntegrationBranch(); got != "main" {
		t.Fatalf("github-flow must resolve the integration target to main, got %q", got)
	}
}

func TestShellJoinArgsPreservesBoundaries(t *testing.T) {
	// t1576 review round 4: a bare strings.Join erased the argument
	// boundaries — `-run 'TestFirst|TestSecond'` reached sh -c unquoted and
	// sh read the regex's `|` as a pipe (exit 127 on a command that passes
	// when run by hand). The joined command string must keep the argument
	// one quoted word, and must still classify as the go test run it is.
	joined := shellJoinArgs([]string{"go", "test", "-json", "-run", "TestFirst|TestSecond", "."})
	if !strings.Contains(joined, "'TestFirst|TestSecond'") {
		t.Fatalf("the regex argument must stay one quoted argument: %q", joined)
	}
	// The legacy single-argument shell-line contract: a space-only argument
	// stays bare, so `exit 7` runs as two shell words.
	if got := shellJoinArgs([]string{"exit 7"}); got != "exit 7" {
		t.Fatalf("a space-only argument keeps the shell-line contract: %q", got)
	}
	// A single argument IS the shell line — the env-scrub compound arrives
	// this way and must not be re-quoted into a command name.
	if got := shellJoinArgs([]string{"unset REVIEW_TEST_VAR && true"}); got != "unset REVIEW_TEST_VAR && true" {
		t.Fatalf("a single compound shell line passes through verbatim: %q", got)
	}
	// In multi-argument form a space stays INSIDE its argument's boundary.
	if got := shellJoinArgs([]string{"printf", "[%s]", "a b"}); !strings.Contains(got, "'a b'") {
		t.Fatalf("a multi-argument space keeps its boundary: %q", got)
	}
	// A leading NAME=value argument is an assignment, not a word: the form
	// survives and only the value is quoted.
	if got := shellJoinArgs([]string{"GOPRIVATE=example.com/*", "go", "test", "-json", "./pkg/..."}); got != "GOPRIVATE='example.com/*' go test -json ./pkg/..." {
		t.Fatalf("a leading assignment keeps its form: %q", got)
	}
	count, structured, err := factory.ClassifyStructuredOutput(joined, strings.NewReader(`{"Action":"pass","Package":"p"}`+"\n"))
	if err != nil || !structured {
		t.Fatalf("the joined command must still classify as a go test run: count=%d structured=%v err=%v", count, structured, err)
	}
}

func TestIntegrationMergeAcceptsSessionFlag(t *testing.T) {
	// t1576 review round 3: the holder-refusal message advises `pass
	// --session <id>` but the verb never registered the flag — the advice
	// itself failed with `unknown flag: --session`.
	cmd := newIntegrationMergeCmd()
	cmd.SetArgs([]string{"merge", "--card", "t1", "--session", "sess-x"})
	err := cmd.Execute()
	if err != nil && strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("the --session flag must be registered: %v", err)
	}
}
