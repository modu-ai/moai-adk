package cli

import (
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

func TestGatewayEffortUsesClaudeSettingsDespiteStoredGLMMode(t *testing.T) {
	for _, mode := range []string{"claude", "gpt", "glm"} {
		t.Run(mode, func(t *testing.T) {
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
			t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
			t.Setenv(config.EnvAnthropicReasoningEffort, "unchanged-fixture")
			oldExec := execOrSpawnClaudeFunc
			t.Cleanup(func() { execOrSpawnClaudeFunc = oldExec })
			execOrSpawnClaudeFunc = func(string, []string, []string) error { return nil }
			binding := &gatewayLaunchBinding{Mode: mode, Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
				if got := envValue(in.Inherited, config.EnvClaudeCodeEffortLevel); got != "high" {
					t.Fatalf("gateway Claude effort %q, want high", got)
				}
				if got := envValue(in.Inherited, config.EnvAnthropicReasoningEffort); got != "unchanged-fixture" {
					t.Fatalf("legacy GLM effort translation survived: %q", got)
				}
				return gateway.LaunchPlan{InitialModel: "gpt-5.6-sol", ChildEnv: in.Inherited}, nil, nil
			}}
			if err := launchClaudeWithGateway("", nil, binding); err != nil {
				t.Fatal(err)
			}
		})
	}
}
