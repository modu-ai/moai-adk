package cli

// SPEC-AGENT-MODEL-INHERIT-001 M4: subagents inherit the main session's model
// and effort, so the per-agent assignment surfaces of the CLI are retired.
//   - `moai model` is gone (REQ-AMI-008, AC-AMI-008).
//   - --profile / --model-policy / --high / --medium-alias / --low on init and
//     --profile on update are accepted, warn, and write nothing (REQ-AMI-016,
//     AC-AMI-016, design D10/D13).
//   - neither the init nor the reconfigure wizard asks the agent model-policy
//     question (REQ-AMI-015, design D13).
// Not parallel-safe: the flag tests mutate the shared initCmd / updateCmd.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

func TestModelCommandIsRemoved(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "model" {
			t.Fatalf("`moai model` is still registered (%s)", c.Short)
		}
	}
}

// assertDeprecationWarning checks the warning names the flag, the inheritance
// rule and the command that owns the main-session policy.
func assertDeprecationWarning(t *testing.T, out, flag string) {
	t.Helper()
	for _, want := range []string{flag, "deprecated", "subagents now inherit the main session's model and effort", "moai profile setup"} {
		if !strings.Contains(out, want) {
			t.Errorf("warning for %s lacks %q:\n%s", flag, want, out)
		}
	}
}

func resetDeprecatedInitFlags(t *testing.T) {
	t.Helper()
	resetInitFlagsForProfile(t)
	for _, f := range []string{"high", "medium-alias", "low"} {
		_ = initCmd.Flags().Set(f, "false")
	}
}

func TestInitDeprecatedModelFlags_WarnAndAcceptAnyValue(t *testing.T) {
	cases := []struct{ flag, value string }{
		{"profile", "high"},
		{"profile", "bogus"},
		{"model-policy", "low"},
		{"model-policy", "subscription"},
		{"high", "true"},
		{"medium-alias", "true"},
		{"low", "true"},
	}
	for _, tc := range cases {
		t.Run(tc.flag+"="+tc.value, func(t *testing.T) {
			resetDeprecatedInitFlags(t)
			t.Cleanup(func() { resetDeprecatedInitFlags(t) })
			var errBuf bytes.Buffer
			initCmd.SetErr(&errBuf)
			t.Cleanup(func() { initCmd.SetErr(nil) })
			if err := initCmd.Flags().Set(tc.flag, tc.value); err != nil {
				t.Fatal(err)
			}
			if err := validateInitFlags(initCmd, []string{}); err != nil {
				t.Fatalf("--%s=%s must be accepted, got: %v", tc.flag, tc.value, err)
			}
			assertDeprecationWarning(t, errBuf.String(), "--"+tc.flag)
		})
	}

	// No deprecated flag → no warning.
	resetDeprecatedInitFlags(t)
	var errBuf bytes.Buffer
	initCmd.SetErr(&errBuf)
	t.Cleanup(func() { initCmd.SetErr(nil) })
	if err := validateInitFlags(initCmd, []string{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(errBuf.String(), "deprecated") {
		t.Errorf("a run with no deprecated flag printed a warning:\n%s", errBuf.String())
	}
}

func TestUpdateDeprecatedProfileFlag_WarnsAndIsAccepted(t *testing.T) {
	var errBuf bytes.Buffer
	updateCmd.SetErr(&errBuf)
	t.Cleanup(func() { updateCmd.SetErr(nil); _ = updateCmd.Flags().Set("profile", "") })
	for _, v := range []string{"low", "bogus"} {
		errBuf.Reset()
		if err := updateCmd.Flags().Set("profile", v); err != nil {
			t.Fatal(err)
		}
		if err := validateUpdateFlags(updateCmd, nil); err != nil {
			t.Fatalf("update --profile %s must be accepted, got: %v", v, err)
		}
		assertDeprecationWarning(t, errBuf.String(), "--profile")
	}
}

// runInitAt runs the init command non-interactively into a fresh root with the
// given extra flags and returns the deployed llm.yaml.
func runInitAt(t *testing.T, extra map[string]string) string {
	t.Helper()
	root := t.TempDir()
	var buf bytes.Buffer
	initCmd.SetOut(&buf)
	initCmd.SetErr(&buf)
	resetDeprecatedInitFlags(t)
	t.Cleanup(func() {
		resetDeprecatedInitFlags(t)
		_ = initCmd.Flags().Set("root", "")
		initCmd.SetOut(nil)
		initCmd.SetErr(nil)
	})
	flags := map[string]string{"root": root, "non-interactive": "true", "name": "flag-test", "language": "Go", "mode": "tdd"}
	for k, v := range extra {
		flags[k] = v
	}
	for k, v := range flags {
		if err := initCmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set %s: %v", k, err)
		}
	}
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init RunE: %v\n%s", err, buf.String())
	}
	b, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestInitDeprecatedModelFlags_WriteNothing(t *testing.T) {
	baseline := runInitAt(t, nil)
	for _, extra := range []map[string]string{
		{"profile": "high"},
		{"model-policy": "low"},
		{"high": "true"},
		{"low": "true"},
	} {
		if got := runInitAt(t, extra); got != baseline {
			t.Errorf("init with %v wrote a different llm.yaml than a run without it", extra)
		}
	}
}

func TestWizardsDoNotAskTheAgentModelPolicy(t *testing.T) {
	root := t.TempDir()
	for name, qs := range map[string][]wizard.Question{
		"default":     wizard.DefaultQuestions(root),
		"reconfigure": wizard.ReconfigureQuestions(root),
		"init":        wizard.InitQuestions(root),
	} {
		if len(qs) == 0 {
			t.Fatalf("%s question set is empty — the check would be vacuous", name)
		}
		if wizard.QuestionByID(qs, "model_policy") != nil {
			t.Errorf("the %s wizard still asks model_policy", name)
		}
	}
}
