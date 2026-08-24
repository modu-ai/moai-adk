package hook

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestPreTool_GitCommitDoesNotRunTheQualityGate pins the commit path out of the
// PreToolUse hook's synchronous budget.
//
// PreToolUse used to run the full vet+lint+test quality gate inline whenever the
// Bash command was a git commit, and deny the tool call if it failed. Against a
// repository of any size that does not fit the budget it is given: the gate's own
// ceilings are 30s/60s/120s per step while the settings.json entry for this hook
// allows 10s, so Claude Code killed the hook before the gate could finish.
// Measured on this repository: 30,033 / 30,016 / 30,020 ms, all three ending in a
// deny whose reason named a budget that had not tripped.
//
// A gate that cannot finish inside its budget has produced no evidence, so it must
// not deny on that basis. The gate itself is not lost: the git pre-commit hook
// installed by `moai init` / `moai update` shells out to `moai gate`
// (internal/cli/hook_install_precommit.go), which is the same gate with no 10s
// ceiling, a clean abort instead of a tool-permission deny, and a documented
// SKIP_MOAI_PRECOMMIT bypass.
//
// The fixture below fails `go vet`, so the gate would deny if it still ran here.
// Deliberately tiny: the point is that the gate does not run, and a fixture that
// took 30s to prove it would be measuring the wrong thing.
func TestPreTool_GitCommitDoesNotRunTheQualityGate(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH — the gate's vet step cannot run either way")
	}

	// semi-auto (the default a user without MOAI_AUTONOMY_TIER gets) is the tier
	// that used to keep the gate ON. Setting it explicitly stops an inherited
	// value in the developer's environment from masking the regression.
	t.Setenv(config.EnvAutonomyTier, string(config.AutonomyTierSemiAuto))

	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module sample\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	vetBad := "package sample\n\nimport \"fmt\"\n\n// VetBad triggers go vet's printf verb/arg check.\nfunc VetBad() { fmt.Printf(\"%d\", \"string-arg\") }\n"
	if err := os.WriteFile(filepath.Join(repo, "bad.go"), []byte(vetBad), 0o644); err != nil {
		t.Fatalf("write bad.go: %v", err)
	}

	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     DefaultSecurityPolicy(),
		projectDir: repo,
	}

	toolInput, err := json.Marshal(map[string]string{"command": `git commit -m "add a thing"`})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}
	input := &HookInput{
		SessionID:     "sess-commit-gate",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     json.RawMessage(toolInput),
		CWD:           repo,
	}

	got, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("expected a non-nil hook output")
	}
	if got.HookSpecificOutput.PermissionDecision == DecisionDeny {
		t.Fatalf("PreToolUse denied a plain `git commit` against a vet-failing fixture — "+
			"the quality gate is still running inside the hook's synchronous budget. "+
			"reason: %s", got.HookSpecificOutput.PermissionDecisionReason)
	}
}
