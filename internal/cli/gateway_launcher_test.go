package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/profile"
)

func TestGatewayLaunchUsesPreparedEnvironmentForNormalAndContinue(t *testing.T) {
	for _, cont := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "continue"}[cont], func(t *testing.T) {
			fakeMoaiProject(t)
			t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
			old := execOrSpawnClaudeFunc
			t.Cleanup(func() { execOrSpawnClaudeFunc = old })
			want := []string{"PREPARED=value"}
			calls := 0
			stopped := false
			execOrSpawnClaudeFunc = func(_ string, args, env []string) error {
				calls++
				if !reflect.DeepEqual(env, want) {
					t.Fatalf("normal env %v", env)
				}
				if !strings.Contains(strings.Join(args, " "), "--model gpt-5.6-sol") {
					t.Fatal(args)
				}
				return nil
			}
			binding := &gatewayLaunchBinding{Mode: "gpt", Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
				if in.ExplicitModel != "gpt-5.6-sol" {
					t.Fatalf("explicit model %q", in.ExplicitModel)
				}
				return gateway.LaunchPlan{InitialModel: "gpt-5.6-sol", ChildEnv: want}, func() { stopped = true }, nil
			}, Continue: func(_ string, _ []string, env []string) error {
				calls++
				if !reflect.DeepEqual(env, want) {
					t.Fatalf("continue env %v", env)
				}
				return nil
			}}
			args := []string{"--model=gpt-5.6-sol"}
			if cont {
				args = append(args, "--continue")
			}
			if err := launchClaudeWithGateway("", args, binding); err != nil {
				t.Fatal(err)
			}
			if calls != 1 || !stopped {
				t.Fatalf("calls %d cleanup %v", calls, stopped)
			}
		})
	}
}

func TestUnifiedGatewayLaunchSkipsLegacyModeMutation(t *testing.T) {
	root := fakeMoaiProject(t)
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
	t.Setenv("MOAI_NO_PROFILE_FALLBACK", "1")
	t.Setenv("ANTHROPIC_BASE_URL", "inherited")
	old := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = old })
	execOrSpawnClaudeFunc = func(string, []string, []string) error { return nil }
	binding := &gatewayLaunchBinding{Mode: "gpt", Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
		if in.Mode != "gpt" {
			t.Fatal(in.Mode)
		}
		return gateway.LaunchPlan{InitialModel: "gpt-5.6-sol", ChildEnv: []string{"READY=yes"}}, nil, nil
	}}
	if err := unifiedLaunchWithGateway("default", "gpt", nil, binding); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ANTHROPIC_BASE_URL") != "inherited" {
		t.Fatal("gateway altered parent environment")
	}
	data, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err == nil && strings.Contains(string(data), "mode:") {
		t.Fatalf("legacy mode persisted: %s", data)
	}
}

func TestGatewayContinueFallbackRetainsOneSupervisor(t *testing.T) {
	fakeMoaiProject(t)
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
	old := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = old })
	prepared, attempts, stops := 0, 0, 0
	want := []string{"GATEWAY=single"}
	execOrSpawnClaudeFunc = func(_ string, args, env []string) error {
		attempts++
		if !reflect.DeepEqual(env, want) || strings.Contains(strings.Join(args, " "), "--continue") {
			t.Fatalf("fallback %v %v", args, env)
		}
		return nil
	}
	binding := &gatewayLaunchBinding{Mode: "gpt", Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
		prepared++
		if in.ExplicitModel != "" {
			t.Fatal("prompt model parsed as launcher flag")
		}
		return gateway.LaunchPlan{InitialModel: "gpt-5.6-sol", ChildEnv: want}, func() { stops++ }, nil
	}, Continue: func(_ string, _ []string, env []string) error {
		attempts++
		if !reflect.DeepEqual(env, want) {
			t.Fatal(env)
		}
		return exec.Command("/bin/sh", "-c", "exit 1").Run()
	}}
	if err := launchClaudeWithGateway("", []string{"--continue", "--", "--model=prompt"}, binding); err != nil {
		t.Fatal(err)
	}
	if prepared != 1 || attempts != 2 || stops != 1 {
		t.Fatalf("prepare %d attempts %d stops %d", prepared, attempts, stops)
	}
}

func TestLegacyContinueDoesNotRecordNewWorktreeSpawn(t *testing.T) {
	root := fakeMoaiProject(t)
	t.Setenv("MOAI_HOME", "")
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
	store, err := resolveChainStore()
	if err != nil {
		t.Fatal(err)
	}
	before := len(store.BuildNodes())
	if err := launchClaudeDefault("", []string{"--continue", "--worktree", "fixture"}); err != nil {
		t.Fatal(err)
	}
	if got := len(store.BuildNodes()); got != before {
		t.Fatalf("continued session recorded a new spawn: before %d after %d", before, got)
	}
}

// gatewayEffortFixture prepares a gateway launch whose profile resolves effort
// "high" while the project carries a persisted GLM mode — the stored mode must
// not route a gateway launch onto the GLM effort translation.
func gatewayEffortFixture(t *testing.T) {
	t.Helper()
	root := fakeMoaiProject(t)
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("llm:\n  team_mode: glm\n  glm:\n    models:\n      high: glm-5.3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if active, _, _ := resolveGLMBackendForLaunch(root); !active {
		t.Fatal("fixture did not activate persisted GLM mode")
	}
	oldBase := profile.BaseDirOverride
	profile.BaseDirOverride = t.TempDir()
	t.Cleanup(func() { profile.BaseDirOverride = oldBase })
	if err := profile.WritePreferences("", profile.ProfilePreferences{Model: "opus", EffortLevel: "high"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_NO_PROFILE_FALLBACK", "1")
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
	t.Setenv(config.EnvAnthropicReasoningEffort, "unchanged-fixture")
	oldExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = oldExec })
	execOrSpawnClaudeFunc = func(string, []string, []string) error { return nil }
}

// gatewaySettingsSource returns the --settings value a gateway launch request
// carries, or "" when none is present.
func gatewaySettingsSource(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			return ""
		}
		if args[i] == settingsFlagLong && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(args[i], settingsFlagLong+"=") {
			return strings.TrimPrefix(args[i], settingsFlagLong+"=")
		}
	}
	return ""
}

// TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode pins where a gateway
// launch carries the profile effort (card t668): in the injected --settings
// payload as effortLevel, NEVER in CLAUDE_CODE_EFFORT_LEVEL. That variable is
// an override — while it is set Claude Code refuses every in-session /effort or
// /model effort change — and no gateway component reads it (the adapter takes
// effort off each request body). The payload is also run through
// prepareGatewayOverlay, the merge the real session binding materializes as the
// child settings file, so a merge that dropped the key would turn this red.
func TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode(t *testing.T) {
	for _, mode := range []string{"claude", "gpt", "glm"} {
		t.Run(mode, func(t *testing.T) {
			gatewayEffortFixture(t)
			t.Setenv(config.EnvClaudeCodeEffortLevel, "fixture")
			if err := os.Unsetenv(config.EnvClaudeCodeEffortLevel); err != nil {
				t.Fatal(err)
			}
			prepared := false
			binding := &gatewayLaunchBinding{Mode: mode, Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
				prepared = true
				for _, item := range in.Inherited {
					if strings.HasPrefix(item, config.EnvClaudeCodeEffortLevel+"=") {
						t.Errorf("gateway launch env gained %s; the override freezes in-session effort changes", item)
					}
				}
				if got := envValue(in.Inherited, config.EnvAnthropicReasoningEffort); got != "unchanged-fixture" {
					t.Errorf("legacy GLM effort translation survived: %q", got)
				}
				source := gatewaySettingsSource(in.Args)
				if source == "" {
					t.Fatalf("gateway launch carries no --settings payload: %v", in.Args)
				}
				if got := readSettingsPayload(t, source)[effortSettingsKey]; got != "high" {
					t.Errorf("injected %s = %v, want high", effortSettingsKey, got)
				}
				data, _, _, err := prepareGatewayOverlay(in.Args, in.Inherited, map[string]any{"model": "fixture"})
				if err != nil {
					t.Fatal(err)
				}
				var merged map[string]any
				if err := json.Unmarshal(data, &merged); err != nil {
					t.Fatal(err)
				}
				if merged[effortSettingsKey] != "high" {
					t.Errorf("gateway child settings %s = %v, want high", effortSettingsKey, merged[effortSettingsKey])
				}
				return gateway.LaunchPlan{InitialModel: "gpt-5.6-sol", ChildEnv: in.Inherited}, nil, nil
			}}
			if err := unifiedLaunchWithGateway("", mode, nil, binding); err != nil {
				t.Fatal(err)
			}
			if !prepared {
				t.Fatal("gateway binding was never prepared")
			}
		})
	}
}

// TestGatewayLaunchEnvPreservesInheritedEffort is the other half of the t595
// invariant on the gateway branch: an inherited CLAUDE_CODE_EFFORT_LEVEL is the
// user's documented per-session override, so the launch passes it through
// untouched — neither replaced by the profile effort nor stripped.
func TestGatewayLaunchEnvPreservesInheritedEffort(t *testing.T) {
	for _, mode := range []string{"claude", "gpt", "glm"} {
		t.Run(mode, func(t *testing.T) {
			gatewayEffortFixture(t)
			t.Setenv(config.EnvClaudeCodeEffortLevel, "xhigh")
			binding := &gatewayLaunchBinding{Mode: mode, Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
				count := 0
				for _, item := range in.Inherited {
					if strings.HasPrefix(item, config.EnvClaudeCodeEffortLevel+"=") {
						count++
					}
				}
				if count != 1 {
					t.Errorf("%s appears %d times in the gateway launch env, want exactly 1", config.EnvClaudeCodeEffortLevel, count)
				}
				if got := envValue(in.Inherited, config.EnvClaudeCodeEffortLevel); got != "xhigh" {
					t.Errorf("%s = %q, want the inherited xhigh", config.EnvClaudeCodeEffortLevel, got)
				}
				return gateway.LaunchPlan{InitialModel: "gpt-5.6-sol", ChildEnv: in.Inherited}, nil, nil
			}}
			if err := unifiedLaunchWithGateway("", mode, nil, binding); err != nil {
				t.Fatal(err)
			}
		})
	}
}
