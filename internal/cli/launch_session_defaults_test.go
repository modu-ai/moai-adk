package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Factory session launch defaults (card t1808): the leader (-f) and lane (-l)
// of each launcher start from the llm.defaults.sessions pair. The tests pin
// the argv / settings payload / env each launcher builds per mode.

// factorySessionEnv stamps (or clears) the factory role and backend the
// launch reads, isolating the test from the shell it runs in.
func factorySessionEnv(t *testing.T, role, backend string) {
	t.Helper()
	for _, k := range []string{config.EnvFactoryRole, config.EnvMoaiFactoryWorker, config.EnvMoaiFactoryWorkers, config.EnvFactoryRunID, config.EnvFactoryBackend} {
		t.Setenv(k, "")
	}
	switch role {
	case config.SessionRoleLeader:
		t.Setenv(config.EnvFactoryRunID, "run-test")
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	case config.SessionRoleLane:
		t.Setenv(config.EnvFactoryRunID, "run-test")
		t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	}
	t.Setenv(config.EnvFactoryBackend, backend)
}

// writeSessionLLM writes an llm.yaml carrying the given sessions block under
// the cross-session config root (the launch's project root in these tests).
func writeSessionLLM(t *testing.T, root, sessions string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte("llm:\n  defaults:\n    sessions:\n"+sessions), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFactorySessionDefaultResolution(t *testing.T) {
	cases := []struct {
		role, backend string
		wantBackend   string
		want          config.ModelEffort
	}{
		{config.SessionRoleLeader, factory.BackendClaude, config.SessionBackendCC, config.ModelEffort{Model: "opus", Effort: "high"}},
		{config.SessionRoleLane, factory.BackendClaude, config.SessionBackendCC, config.ModelEffort{Model: "sonnet", Effort: "medium"}},
		{config.SessionRoleLeader, factory.BackendGLM, config.SessionBackendGLM, config.ModelEffort{Model: "opus", Effort: "high"}},
		{config.SessionRoleLane, factory.BackendGLM, config.SessionBackendGLM, config.ModelEffort{Model: "sonnet", Effort: "medium"}},
	}
	for _, tc := range cases {
		t.Run(tc.role+"/"+tc.backend, func(t *testing.T) {
			factorySessionEnv(t, tc.role, tc.backend)
			backend, got, ok := factorySessionDefault(os.Environ(), t.TempDir())
			if !ok || backend != tc.wantBackend || got != tc.want {
				t.Errorf("factorySessionDefault = %s %+v (ok %v), want %s %+v", backend, got, ok, tc.wantBackend, tc.want)
			}
		})
	}
	t.Run("no factory role", func(t *testing.T) {
		factorySessionEnv(t, "", factory.BackendClaude)
		if _, _, ok := factorySessionDefault(os.Environ(), t.TempDir()); ok {
			t.Error("a non-factory launch must not resolve a session default")
		}
	})
	t.Run("custom mode reads llm.yaml", func(t *testing.T) {
		factorySessionEnv(t, config.SessionRoleLane, factory.BackendClaude)
		root := t.TempDir()
		writeSessionLLM(t, root, "      mode: custom\n      cc:\n        lane: {model: opus, effort: low}\n")
		if _, got, _ := factorySessionDefault(os.Environ(), root); got != (config.ModelEffort{Model: "opus", Effort: "low"}) {
			t.Errorf("custom cc lane = %+v, want opus/low", got)
		}
	})
}

// TestFactorySettingsCarrySessionEffort: a cc factory launch's transient
// settings payload carries the session default effort, replacing the profile
// effort; CLAUDE_CODE_EFFORT_LEVEL is never involved.
func TestFactorySettingsCarrySessionEffort(t *testing.T) {
	for _, tc := range []struct{ role, want string }{
		{config.SessionRoleLeader, "high"},
		{config.SessionRoleLane, "medium"},
	} {
		t.Run(tc.role, func(t *testing.T) {
			withCrossSessionConfig(t, "")
			withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "low"}, nil)
			factorySessionEnv(t, tc.role, factory.BackendClaude)
			t.Setenv(config.EnvClaudeCodeEffortLevel, "")

			flag, cleanup := prepareFactorySettings("dev", nil)
			t.Cleanup(cleanup)
			effort, has := injectedSettingsEffort(t, flag)
			if !has || effort != tc.want {
				t.Errorf("injected effortLevel = %v (settings %v), want %s", effort, has, tc.want)
			}
			if n, _ := countEffortFlags(flag); n != 0 {
				t.Errorf("a non-max session effort must ride the settings payload, not --effort (flags %v)", flag)
			}
			if os.Getenv(config.EnvClaudeCodeEffortLevel) != "" {
				t.Error("CLAUDE_CODE_EFFORT_LEVEL must never be set by the launcher (it locks /effort)")
			}
		})
	}
	t.Run("glm factory keeps the profile effort in the Claude payload", func(t *testing.T) {
		withCrossSessionConfig(t, "")
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "low"}, nil)
		factorySessionEnv(t, config.SessionRoleLeader, factory.BackendGLM)
		flag, cleanup := prepareFactorySettings("dev", nil)
		t.Cleanup(cleanup)
		if effort, _ := injectedSettingsEffort(t, flag); effort != "low" {
			t.Errorf("glm payload effortLevel = %v, want the profile low (GLM effort rides the z.ai mapping)", effort)
		}
	})
}

// TestFactorySessionModelArgs: the launch model is the session default unless
// the operator passed --model; the glm model stays a slot alias.
func TestFactorySessionModelArgs(t *testing.T) {
	factorySessionEnv(t, config.SessionRoleLane, factory.BackendClaude)
	root := t.TempDir()
	if got := factorySessionModel(os.Environ(), root, false, "fromprofile"); got != "sonnet" {
		t.Errorf("cc lane model = %q, want sonnet (session default over the profile)", got)
	}
	if got := factorySessionModel(os.Environ(), root, true, "opus"); got != "opus" {
		t.Errorf("explicit --model must win, got %q", got)
	}
	factorySessionEnv(t, config.SessionRoleLeader, factory.BackendGLM)
	if got := factorySessionModel(os.Environ(), root, false, ""); got != "opus" {
		t.Errorf("glm leader model = %q, want the opus slot alias", got)
	}
	factorySessionEnv(t, "", factory.BackendClaude)
	if got := factorySessionModel(os.Environ(), root, false, "fromprofile"); got != "fromprofile" {
		t.Errorf("a non-factory launch keeps its model, got %q", got)
	}
}

// TestFactorySessionGLMWire pins what a glm factory launch puts on the wire:
// the session default is Claude vocabulary, translated through the
// llm.glm.effort tier table (medium -> high, high -> max) and wired verbatim,
// so `glm -l` reaches z.ai at high and `glm -f` at max. A flash slot stays max.
func TestFactorySessionGLMWire(t *testing.T) {
	models := config.GLMModels{High: config.DefaultGLM53, Medium: config.DefaultGLM53, Low: config.DefaultGLM53Flash, Fable: config.DefaultGLM53}
	tier := config.GLMTierEffort{High: "max", Medium: "high", Low: "low", Fable: "max"}
	for _, tc := range []struct{ role, model, want string }{
		{config.SessionRoleLane, "sonnet", "high"},
		{config.SessionRoleLeader, "opus", "max"},
	} {
		t.Run(tc.role, func(t *testing.T) {
			factorySessionEnv(t, tc.role, factory.BackendGLM)
			state, ok := factorySessionGLMState(os.Environ(), t.TempDir(), tier)
			if !ok || state != tc.want {
				t.Fatalf("factorySessionGLMState = %q (ok %v), want %q", state, ok, tc.want)
			}
			env := buildEnvForGLMLaunchState(models, tc.model, state, []string{config.EnvAnthropicReasoningEffort + "=stale"})
			if got := launchEnvValue(env, config.EnvAnthropicReasoningEffort); got != tc.want {
				t.Errorf("%s=%q, want %q", config.EnvAnthropicReasoningEffort, got, tc.want)
			}
			if n := slices.IndexFunc(env, func(e string) bool { return e == config.EnvAnthropicReasoningEffort+"=stale" }); n >= 0 {
				t.Error("the inherited reasoning effort must be replaced")
			}
		})
	}
	t.Run("flash slot pins max", func(t *testing.T) {
		env := buildEnvForGLMLaunchState(models, "haiku", "low", nil)
		if got := launchEnvValue(env, config.EnvAnthropicReasoningEffort); got != "max" {
			t.Errorf("flash slot effort = %q, want max", got)
		}
	})
	t.Run("cc launch has no glm state", func(t *testing.T) {
		factorySessionEnv(t, config.SessionRoleLane, factory.BackendClaude)
		if _, ok := factorySessionGLMState(os.Environ(), t.TempDir(), tier); ok {
			t.Error("a cc launch must not resolve a glm state")
		}
	})
}

// TestCodexSessionDefaultArgs: `moai codex` (leader) and `moai codex -l`
// (lane) carry -m and -c model_reasoning_effort; an operator -m / -c effort
// suppresses the injection.
func TestCodexSessionDefaultArgs(t *testing.T) {
	root := t.TempDir()
	if got := codexSessionDefaultArgs(root, config.SessionRoleLeader, nil); !slices.Equal(got,
		[]string{"-m", config.CodexModelSol, "-c", `model_reasoning_effort="high"`}) {
		t.Errorf("leader args = %v", got)
	}
	if got := codexSessionDefaultArgs(root, config.SessionRoleLane, nil); !slices.Equal(got,
		[]string{"-m", config.CodexModelSol, "-c", `model_reasoning_effort="medium"`}) {
		t.Errorf("lane args = %v", got)
	}
	if got := codexSessionDefaultArgs(root, config.SessionRoleLeader, []string{"-m", "other"}); !slices.Equal(got,
		[]string{"-c", `model_reasoning_effort="high"`}) {
		t.Errorf("operator -m keeps only the effort, got %v", got)
	}
	if got := codexSessionDefaultArgs(root, config.SessionRoleLeader, []string{"-c", "model_reasoning_effort=low"}); !slices.Equal(got,
		[]string{"-m", config.CodexModelSol}) {
		t.Errorf("operator -c effort keeps only the model, got %v", got)
	}
	writeSessionLLM(t, root, "      mode: custom\n      codex:\n        lane: {effort: xhigh}\n")
	if got := codexSessionDefaultArgs(root, config.SessionRoleLane, nil); !slices.Equal(got,
		[]string{"-m", config.CodexModelSol, "-c", `model_reasoning_effort="xhigh"`}) {
		t.Errorf("custom lane args = %v", got)
	}
}

// TestCodexPlainLaunchCarriesNoModelArgs: a plain `moai codex` is the
// operator's ordinary session — its argv carries no -m and no
// model_reasoning_effort (the project .codex/config.toml decides), even with
// the real session-default seam in place. The leader pair is reserved for a
// future `moai codex -f`.
func TestCodexPlainLaunchCarriesNoModelArgs(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	codexSessionDefaultArgsFn = codexSessionDefaultArgs // restore the real body
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t, "--", "--sandbox", "workspace-write"); err != nil {
		t.Fatal(err)
	}
	got := codexArgvTail(t, cap)
	for _, a := range got {
		if a == "-m" || strings.Contains(a, "model_reasoning_effort") {
			t.Errorf("plain moai codex argv carries %q: %q", a, got)
		}
	}
	if want := []string{"--sandbox", "workspace-write"}; !slices.Equal(got, want) {
		t.Errorf("argv tail = %q, want %q", got, want)
	}
}
