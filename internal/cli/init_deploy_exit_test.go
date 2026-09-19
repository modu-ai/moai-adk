package cli

// SPEC-INIT-DEPLOY-EXIT-001 — CLI-layer contract for a template-deployment
// failure: RunE returns an error (so cobra exits non-zero), the success card is
// never printed, the warning summary still reaches stderr, and stdout carries
// no warning or error text.

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/project"
)

// deployFailExecutor reproduces what the real executor returns when template
// deployment fails: the InitResult (carrying warnings recorded before the
// failure) alongside the wrapped render error.
type deployFailExecutor struct {
	warnings []string
}

func (e *deployFailExecutor) SetReporter(_ project.ProgressReporter) {}

func (e *deployFailExecutor) Execute(_ context.Context, _ project.InitOptions) (*project.InitResult, error) {
	return &project.InitResult{Warnings: e.warnings},
		errors.New(`initialization: template deployment: deploy templates: template render ` +
			`".claude/hooks/moai/handle-agent-hook.sh.tmpl": unexpanded dynamic token detected: found "$REPO"`)
}

// injectDeployFailExecutor swaps the executor seam for the duration of the test.
func injectDeployFailExecutor(t *testing.T, warnings ...string) {
	t.Helper()
	orig := newInitPhaseExecutorFn
	newInitPhaseExecutorFn = func(
		_ project.Detector,
		_ project.MethodologyDetector,
		_ project.ProjectValidator,
		_ project.Initializer,
	) initPhaseExecutor {
		return &deployFailExecutor{warnings: warnings}
	}
	t.Cleanup(func() { newInitPhaseExecutorFn = orig })
}

// runInitWithFlags drives the init command non-interactively against root and
// returns the captured stdout / stderr plus its error.
func runInitWithFlags(t *testing.T, root string, extra map[string]string) (string, string, error) {
	t.Helper()

	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	initCmd.SetOut(out)
	initCmd.SetErr(errOut)

	flags := map[string]string{
		"root":            root,
		"name":            "deploy-exit-test",
		"language":        "go",
		"mode":            "tdd",
		"non-interactive": "true",
	}
	for k, v := range extra {
		flags[k] = v
	}
	for flag, value := range flags {
		if err := initCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set %s flag: %v", flag, err)
		}
	}
	t.Cleanup(func() {
		for _, flag := range []string{"root", "name", "language", "mode", "llm"} {
			if err := initCmd.Flags().Set(flag, ""); err != nil {
				t.Logf("reset %s: %v", flag, err)
			}
		}
		for _, flag := range []string{"non-interactive", "force"} {
			if err := initCmd.Flags().Set(flag, "false"); err != nil {
				t.Logf("reset %s: %v", flag, err)
			}
		}
	})

	runErr := initCmd.RunE(initCmd, []string{})
	return out.String(), errOut.String(), runErr
}

// AC-IDE-005 — the run fails, the success card is absent, and the error text
// says the tree is incomplete and names the failing template.
func TestInit_DeployFailureSuppressesSuccessCard(t *testing.T) {
	injectDeployFailExecutor(t)

	stdout, stderr, err := runInitWithFlags(t, t.TempDir(), nil)
	if err == nil {
		t.Fatal("init RunE returned nil on a template deployment failure; the exit code would be 0")
	}
	msg := err.Error()
	if !strings.Contains(msg, "handle-agent-hook.sh.tmpl") {
		t.Errorf("error does not name the failing template: %v", msg)
	}
	if !strings.Contains(msg, "INCOMPLETE") {
		t.Errorf("error does not state that the project tree is incomplete: %v", msg)
	}
	if !strings.Contains(msg, "moai init") {
		t.Errorf("error carries no next-action guidance: %v", msg)
	}
	for name, stream := range map[string]string{"stdout": stdout, "stderr": stderr} {
		if strings.Contains(stream, "MoAI project initialized") {
			t.Errorf("success card rendered on %s despite the failure: %q", name, stream)
		}
	}
}

// AC-IDE-006 — the warning summary still reaches stderr on the failure path,
// and stdout carries no warning or error text.
func TestInit_DeployFailurePreservesWarningSummary(t *testing.T) {
	const notice = "skill mirror fell back to copy for 9 skill(s)"
	injectDeployFailExecutor(t, notice)

	stdout, stderr, err := runInitWithFlags(t, t.TempDir(), nil)
	if err == nil {
		t.Fatal("init RunE returned nil on a template deployment failure")
	}
	if !strings.Contains(stderr, "warning(s) during init:") {
		t.Errorf("warning summary absent from stderr: %q", stderr)
	}
	if !strings.Contains(stderr, notice) {
		t.Errorf("executor result warning did not reach the summary: %q", stderr)
	}
	for _, leak := range []string{notice, "warning(s) during init:", "INCOMPLETE", "handle-agent-hook.sh.tmpl"} {
		if strings.Contains(stdout, leak) {
			t.Errorf("stdout carries warning/error text %q: %q", leak, stdout)
		}
	}
}

// AC-IDE-007 — the --force reinit path and the codex-only (--llm gpt) path
// behave identically to the default path.
func TestInit_DeployFailureConsistentAcrossPaths(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flags map[string]string
	}{
		{name: "force", flags: map[string]string{"force": "true"}},
		{name: "codex-only", flags: map[string]string{"llm": "gpt"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			injectDeployFailExecutor(t)

			stdout, stderr, err := runInitWithFlags(t, t.TempDir(), tc.flags)
			if err == nil {
				t.Fatalf("%s path returned nil on a template deployment failure", tc.name)
			}
			if !strings.Contains(err.Error(), "INCOMPLETE") {
				t.Errorf("%s path error does not state the tree is incomplete: %v", tc.name, err)
			}
			for name, stream := range map[string]string{"stdout": stdout, "stderr": stderr} {
				if strings.Contains(stream, "MoAI project initialized") {
					t.Errorf("%s path rendered the success card on %s", tc.name, name)
				}
			}
		})
	}
}
