package cli

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestClaudeLauncherConsumesOnlyLauncherSeparator(t *testing.T) {
	for _, tc := range []struct {
		name string
		tail []string
	}{
		{"print flags", []string{"--print", "--output-format", "json", "--tools", "", "--max-turns", "1", "Reply exactly MOAI_GPT_LIVE_OK."}},
		{"literal delimiter and prompt", []string{"--print", "--", "--model literal prompt $HOME; echo nope"}},
		{"launcher-looking flags stay verbatim", []string{"--model", "gpt-5.6-terra", "--permission-mode", "plan", "--no-chrome"}},
		{"empty tail", []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fakeMoaiProject(t)
			t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
			old := execOrSpawnClaudeFunc
			t.Cleanup(func() { execOrSpawnClaudeFunc = old })
			called := false
			execOrSpawnClaudeFunc = func(_ string, args, _ []string) error {
				called = true
				want := append([]string{"claude", "--no-chrome", "--model", "gpt-5.6-sol"}, tc.tail...)
				if !reflect.DeepEqual(args, want) {
					t.Fatalf("Claude argv = %q, want %q", args, want)
				}
				return nil
			}
			binding := &gatewayLaunchBinding{Mode: "gpt", Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
				if in.ExplicitModel != "gpt-5.6-sol" || !reflect.DeepEqual(append([]string{}, in.Args...), tc.tail) {
					t.Fatalf("prepared model = %q, tail = %q; want sol and %q", in.ExplicitModel, in.Args, tc.tail)
				}
				return gateway.LaunchPlan{InitialModel: "gpt-5.6-sol"}, nil, nil
			}}
			args := append([]string{"--model", "gpt-5.6-sol", "--"}, tc.tail...)
			if err := launchClaudeWithGateway("", args, binding); err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("Claude was not launched")
			}
		})
	}
}
