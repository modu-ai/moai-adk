package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/testguard"
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

// TestCodexSessionDefaultArgs: the session-default argv builder yields -m and
// -c model_reasoning_effort for both roles (the lane pair rides `moai codex
// -l`; the leader pair is reserved for a future `moai codex -f` — a plain
// `moai codex` carries neither); an operator -m / -c effort suppresses the
// injection.
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

// glmSessionLLMBody is an llm.yaml body selecting the GLM backend with the
// shipped tier tables, followed by a defaults.sessions block.
const glmSessionLLMBody = "llm:\n  mode: glm\n  glm:\n    models: {high: glm-5.3, medium: glm-5.3, low: glm-5.3-flash, fable: glm-5.3}\n" +
	"    effort: {high: max, medium: high, low: low, fable: max}\n  defaults:\n    sessions:\n"

// sessionLaunchCapture drives the real runLaunchClaude for a factory session
// stamped by role/backend, with claude stubbed, and returns the argv and env
// the launch handed over plus the launcher notices. llmBody is the full
// llm.yaml content ("" = none). The profile model is haiku so a launch that
// lost the session-default wiring is visible as --model haiku.
//
// NOTE: callers must not use t.Parallel() — the helper chdirs and overrides
// package-level seams and environment variables.
func sessionLaunchCapture(t *testing.T, role, backend, llmBody string) (args, env []string, stderr string) {
	t.Helper()
	root := t.TempDir()
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	for _, d := range []string{filepath.Join(".moai", "config", "sections"), ".claude"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if llmBody != "" {
		if err := os.WriteFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"), []byte(llmBody), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	profileBase := t.TempDir()
	origBase := profile.BaseDirOverride
	t.Cleanup(func() { profile.BaseDirOverride = origBase })
	profile.BaseDirOverride = profileBase
	if err := os.WriteFile(filepath.Join(profileBase, "preferences.yaml"), []byte("user_name: tester\nmodel: haiku\neffort_level: low\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeConfigDir, profileBase)
	t.Setenv(config.EnvAnthropicModel, "")
	t.Setenv(config.EnvMoaiFactoryManaged, "")
	t.Setenv(config.EnvAnthropicReasoningEffort, "")
	factorySessionEnv(t, role, backend)
	t.Chdir(root)
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude-stub")))

	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(_ string, a, e []string) error {
		args, env = a, e
		return nil
	}
	var buf strings.Builder
	origErr := launcherStderr
	t.Cleanup(func() { launcherStderr = origErr })
	launcherStderr = &buf

	if err := runLaunchClaude("", nil); err != nil {
		t.Fatalf("runLaunchClaude: %v (a session-default problem must never block the launch)", err)
	}
	return args, env, buf.String()
}

// argAfter returns the value following flag in args ("" when absent).
func argAfter(args []string, flag string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

// TestFactoryLaunchWiresSessionDefaults drives the real Claude launch path
// (runLaunchClaude) for cc and glm factory sessions and pins the two wirings
// the helper-level tests cannot see: the session model replacing the profile
// model (launcher.go factorySessionModel call) and, under GLM, the session
// effort reaching ANTHROPIC_REASONING_EFFORT (launcher.go
// factorySessionGLMState branch). The glm lane cell is custom low so its
// z.ai state (tier.low = low) differs from what the non-factory slot path
// wires for the sonnet slot.
func TestFactoryLaunchWiresSessionDefaults(t *testing.T) {
	t.Run("cc lane", func(t *testing.T) {
		args, _, _ := sessionLaunchCapture(t, config.SessionRoleLane, factory.BackendClaude, "")
		if got := argAfter(args, "--model"); got != "sonnet" {
			t.Errorf("cc lane --model = %q, want sonnet (args %q)", got, args)
		}
	})
	t.Run("cc leader", func(t *testing.T) {
		args, _, _ := sessionLaunchCapture(t, config.SessionRoleLeader, factory.BackendClaude, "")
		if got := argAfter(args, "--model"); got != "opus" {
			t.Errorf("cc leader --model = %q, want opus (args %q)", got, args)
		}
	})
	t.Run("glm lane", func(t *testing.T) {
		body := glmSessionLLMBody + "      mode: custom\n      glm:\n        lane: {model: sonnet, effort: low}\n"
		args, env, _ := sessionLaunchCapture(t, config.SessionRoleLane, factory.BackendGLM, body)
		if got := argAfter(args, "--model"); got != "sonnet" {
			t.Errorf("glm lane --model = %q, want the sonnet slot alias (args %q)", got, args)
		}
		if got := launchEnvValue(env, config.EnvAnthropicReasoningEffort); got != "low" {
			t.Errorf("glm lane %s = %q, want low (session effort low through llm.glm.effort.low)", config.EnvAnthropicReasoningEffort, got)
		}
	})
	t.Run("glm leader", func(t *testing.T) {
		args, env, _ := sessionLaunchCapture(t, config.SessionRoleLeader, factory.BackendGLM, glmSessionLLMBody+"      mode: default\n")
		if got := argAfter(args, "--model"); got != "opus" {
			t.Errorf("glm leader --model = %q, want the opus slot alias (args %q)", got, args)
		}
		if got := launchEnvValue(env, config.EnvAnthropicReasoningEffort); got != "max" {
			t.Errorf("glm leader %s = %q, want max", config.EnvAnthropicReasoningEffort, got)
		}
	})
}

// TestFactoryLaunchRejectsInvalidSessionDefaults pins the fail-open rule for
// a custom llm.yaml cell that config.Validate refuses: the launch proceeds,
// the launcher warns naming the key and the bad value, the built-in default
// for that cell is used, and the bad value never reaches argv or env.
func TestFactoryLaunchRejectsInvalidSessionDefaults(t *testing.T) {
	t.Run("cc lane model", func(t *testing.T) {
		writeBody := "llm:\n  defaults:\n    sessions:\n      mode: custom\n      cc:\n        lane: {model: \"opus; x\"}\n"
		args, env, stderr := sessionLaunchCapture(t, config.SessionRoleLane, factory.BackendClaude, writeBody)
		if got := argAfter(args, "--model"); got != "sonnet" {
			t.Errorf("--model = %q, want the built-in sonnet", got)
		}
		assertSessionWarning(t, stderr, "llm.defaults.sessions.cc.lane.model", "opus; x")
		assertNeverCarried(t, append(args, env...), "opus; x")
	})
	t.Run("cc lane effort", func(t *testing.T) {
		root := withCrossSessionConfig(t, "")
		writeSessionLLM(t, root, "      mode: custom\n      cc:\n        lane: {effort: ultra}\n")
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "low"}, nil)
		factorySessionEnv(t, config.SessionRoleLane, factory.BackendClaude)
		var buf strings.Builder
		origErr := launcherStderr
		t.Cleanup(func() { launcherStderr = origErr })
		launcherStderr = &buf

		flag, cleanup := prepareFactorySettings("dev", nil)
		t.Cleanup(cleanup)
		if effort, _ := injectedSettingsEffort(t, flag); effort != "medium" {
			t.Errorf("injected effortLevel = %v, want the built-in medium", effort)
		}
		if n, v := countEffortFlags(flag); n != 0 {
			t.Errorf("flag = %v carries --effort %s; the bad value must not reach argv", flag, v)
		}
		assertSessionWarning(t, buf.String(), "llm.defaults.sessions.cc.lane.effort", "ultra")
	})
	t.Run("glm lane effort", func(t *testing.T) {
		body := glmSessionLLMBody + "      mode: custom\n      glm:\n        lane: {effort: ultra}\n"
		args, env, stderr := sessionLaunchCapture(t, config.SessionRoleLane, factory.BackendGLM, body)
		if got := launchEnvValue(env, config.EnvAnthropicReasoningEffort); got != "high" {
			t.Errorf("%s = %q, want high (built-in medium through llm.glm.effort.medium)", config.EnvAnthropicReasoningEffort, got)
		}
		assertSessionWarning(t, stderr, "llm.defaults.sessions.glm.lane.effort", "ultra")
		assertNeverCarried(t, append(args, env...), "ultra")
	})
	t.Run("codex lane effort and model", func(t *testing.T) {
		root := t.TempDir()
		writeSessionLLM(t, root, "      mode: custom\n      codex:\n        lane: {model: \"gpt\\nx\", effort: 'high\" model=\"evil'}\n")
		var buf strings.Builder
		origErr := launcherStderr
		t.Cleanup(func() { launcherStderr = origErr })
		launcherStderr = &buf
		got := codexSessionDefaultArgs(root, config.SessionRoleLane, nil)
		if want := []string{"-m", config.CodexModelSol, "-c", `model_reasoning_effort="medium"`}; !slices.Equal(got, want) {
			t.Errorf("codex lane args = %q, want the built-in %q", got, want)
		}
		assertSessionWarning(t, buf.String(), "llm.defaults.sessions.codex.lane.effort", `high" model="evil`)
		assertSessionWarning(t, buf.String(), "llm.defaults.sessions.codex.lane.model", "gpt\nx")
		assertNeverCarried(t, got, "evil")
	})
}

func assertSessionWarning(t *testing.T, stderr, key, value string) {
	t.Helper()
	for _, line := range strings.Split(stderr, "\n") {
		if strings.Contains(line, "Warning:") && strings.Contains(line, key) && strings.Contains(line, strconv.Quote(value)) {
			return
		}
	}
	t.Errorf("stderr carries no warning naming %s and %q:\n%s", key, value, stderr)
}

func assertNeverCarried(t *testing.T, carried []string, bad string) {
	t.Helper()
	for _, a := range carried {
		if strings.Contains(a, bad) {
			t.Errorf("the rejected value %q reached the launch: %q", bad, a)
		}
	}
}

// TestCodexLaneModelSurroundingWhitespace pins the normalization rule: a
// codex lane model carrying leading/trailing whitespace or control characters
// is trimmed, the trimmed value reaches argv, and the launcher writes exactly
// one warning line naming the key and the raw value. The launch proceeds.
func TestCodexLaneModelSurroundingWhitespace(t *testing.T) {
	root := t.TempDir()
	writeSessionLLM(t, root, "      mode: custom\n      codex:\n        lane: {model: \"\\ngpt-custom\\t\"}\n")
	var buf strings.Builder
	origErr := launcherStderr
	t.Cleanup(func() { launcherStderr = origErr })
	launcherStderr = &buf
	got := codexSessionDefaultArgs(root, config.SessionRoleLane, nil)
	if want := []string{"-m", "gpt-custom", "-c", `model_reasoning_effort="medium"`}; !slices.Equal(got, want) {
		t.Errorf("codex lane args = %q, want %q", got, want)
	}
	var warnings []string
	for _, line := range strings.Split(buf.String(), "\n") {
		if strings.Contains(line, "Warning:") {
			warnings = append(warnings, line)
		}
	}
	if len(warnings) != 1 {
		t.Fatalf("warning lines = %d, want exactly 1:\n%s", len(warnings), buf.String())
	}
	assertSessionWarning(t, buf.String(), "llm.defaults.sessions.codex.lane.model", "\ngpt-custom\t")
}

// TestCodexLaneLaunchCarriesSessionDefaults drives the real `moai codex -l`
// launch and pins the lane argv wiring (codex_launcher.go): the lane child
// carries -m <sol> and -c model_reasoning_effort="medium".
func TestCodexLaneLaunchCarriesSessionDefaults(t *testing.T) {
	testguard.RedirectHome(t)
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)
	prevAdd := codexMoaiHomeAddDirArgs
	codexMoaiHomeAddDirArgs = func(string) []string { return nil }
	var argv []string
	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		if argv == nil {
			argv = c.Args
		}
		return nil
	}
	t.Cleanup(func() {
		codexLookPath, codexDirectLaunchFn, codexMoaiHomeAddDirArgs = prevLook, prevDirect, prevAdd
	})
	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex -l: %v", err)
	}
	if got := argAfter(argv, "-m"); got != config.CodexModelSol {
		t.Errorf("lane argv -m = %q, want %s (argv %q)", got, config.CodexModelSol, argv)
	}
	if !slices.Contains(argv, `model_reasoning_effort="medium"`) {
		t.Errorf("lane argv lacks -c model_reasoning_effort=\"medium\" (argv %q)", argv)
	}
}
