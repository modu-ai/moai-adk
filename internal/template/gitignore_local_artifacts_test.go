package template

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTemplateGitignoreIgnoresLocalArtifacts guards the deployed .gitignore
// against re-admitting per-machine runtime artifacts into version control.
//
// templates/.gitignore is not a maintainer convenience: `moai init` writes it
// to the user's project root, so whatever it fails to cover is tracked in every
// project built with MoAI. Two directories are pure local output —
// .moai/logs/ (hook ledgers, audit trails) and .moai/reports/ (per-card
// evidence, generated analyses) — and both must stay out of the index while
// their .gitkeep scaffolds stay in it, so the directories still ship.
//
// The patterns are MEASURED, never inferred: a pattern's reach depends on where
// its slashes sit, and reading one wrong is exactly the defect this guards. The
// test writes the embedded .gitignore into a throwaway repository and asks git
// itself. A prior version of this file's sibling comment asserted that the
// root-anchored "logs/*" rule covered .moai/logs/; it never did, and no test
// contradicted it for as long as the claim stood.
func TestTemplateGitignoreIgnoresLocalArtifacts(t *testing.T) {
	t.Parallel()

	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	gitignore, err := fs.ReadFile(fsys, ".gitignore")
	if err != nil {
		t.Fatalf("read embedded .gitignore: %v", err)
	}

	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(gitBin, args...)
		cmd.Dir = repo
		if out, runErr := cmd.CombinedOutput(); runErr != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), runErr, out)
		}
	}
	run("init", "--quiet", ".")
	if writeErr := os.WriteFile(filepath.Join(repo, ".gitignore"), gitignore, 0o644); writeErr != nil {
		t.Fatalf("write .gitignore: %v", writeErr)
	}

	cases := []struct {
		path    string
		ignored bool
		why     string
	}{
		// Local artifacts — must never reach a user's index.
		{".moai/logs/hook-skip.log", true, "hook log"},
		{".moai/logs/agent-model-audit.jsonl", true, "ledger written beside the logs; *.log does not cover it"},
		{".moai/logs/agents/session-notes.txt", true, "nested log-directory content"},
		{".moai/reports/t123/verdict.md", true, "per-card evidence directory"},
		{".moai/reports/t123/gotest.log", true, "per-card evidence, non-markdown"},
		{".moai/reports/audit-summary.md", true, "loose generated report"},
		{".moai/reports/graph-report.md", true, "regenerating derived artifact"},
		{".moai/reports/plan-audit/SPEC-X.md", true, "plan-audit verdict"},

		// Scaffolds — must stay trackable so the directories ship.
		{".moai/logs/.gitkeep", false, "logs directory scaffold"},
		{".moai/reports/plan-audit/.gitkeep", false, "plan-audit directory scaffold"},

		// Control: a path the rules must not reach. Without it, a rule that
		// ignored everything would pass every assertion above.
		{".moai/project/product.md", false, "user project documentation"},
	}

	for _, tc := range cases {
		cmd := exec.Command(gitBin, "check-ignore", "-q", "--no-index", "--", tc.path)
		cmd.Dir = repo
		runErr := cmd.Run()

		// Exit 0 means the path is ignored, exit 1 means it is not; anything
		// else is git failing rather than answering.
		ignored := runErr == nil
		if runErr != nil {
			var exitErr *exec.ExitError
			if !errors.As(runErr, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("git check-ignore %s: %v", tc.path, runErr)
			}
		}

		if ignored != tc.ignored {
			t.Errorf("%s (%s): ignored=%v, want %v", tc.path, tc.why, ignored, tc.ignored)
		}
	}
}
