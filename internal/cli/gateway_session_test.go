package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/profile"
)

func TestGatewaySessionStartsPrivateChildAndKeepsSecretsOutOfOverlay(t *testing.T) {
	catalog, _ := gateway.NewSessionCatalog(nil)
	stopped := false
	binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog, SessionEnv: []string{"FIXTURE_SESSION=private"}, Start: func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
		if strings.Contains(string(opts.Config.Overlay), "private") || strings.Contains(strings.Join(opts.Args, " "), "private") {
			t.Fatal("private data materialized")
		}
		var overlay map[string]any
		if err := json.Unmarshal(opts.Config.Overlay, &overlay); err != nil {
			t.Fatal(err)
		}
		if overlay["permissions"] == nil || overlay["teammateMode"] != "in-process" {
			t.Fatal(overlay)
		}
		return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "/fixture/settings.json", Stop: func() { stopped = true }}, nil
	}})
	got, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", Inherited: []string{"KEEP=parent"}, Args: []string{"--settings", `{"env":{"Z_AI_API_KEY":"private","KEEP":"user"},"permissions":{"allow":["Read"]}}`}})
	if err != nil {
		t.Fatal(err)
	}
	stop()
	if !stopped || got.ChildSettings != "/fixture/settings.json" {
		t.Fatal("missing child handoff/cleanup")
	}
	for _, item := range got.ChildEnv {
		if strings.HasPrefix(item, "Z_AI_API_KEY=") {
			t.Fatal("user settings credential survived")
		}
	}
	args, err := replaceGatewaySettingsArgs([]string{"--settings", "old", "--", "--settings=prompt"}, "new")
	if err != nil || !reflect.DeepEqual(args, []string{"--settings", "new", "--", "--settings=prompt"}) {
		t.Fatalf("%v %v", args, err)
	}
}

func TestGatewayConversationNewUsesActualCWDAndPrivateSecureNamespace(t *testing.T) {
	root := t.TempDir()
	families, err := conversation.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(root, "worktree")
	d, err := prepareGatewayConversation(gatewayLaunchRequest{
		CWD: cwd, Project: filepath.Join(root, "project"),
		SecureStorage: "/private/secure", SecureStorageSet: true,
	}, families)
	if err != nil {
		t.Fatal(err)
	}
	if d.CWD != cwd || d.ConfigDir == "" || len(d.Args) != 2 || d.Args[0] != "--session-id" || d.Args[1] != d.UUID {
		t.Fatalf("descriptor=%+v", d)
	}
	if got := gatewayConversationPassthrough([]string{"--resume", "old", "--fork-session", "--session-id", "old", "--verbose"}); !reflect.DeepEqual(got, []string{"--verbose"}) {
		t.Fatalf("passthrough=%v", got)
	}
	if _, err := prepareGatewayConversation(gatewayLaunchRequest{CWD: "relative", Project: "project"}, families); err == nil {
		t.Fatal("relative CWD accepted")
	}
}

func TestGatewaySessionFailureAlwaysStopsStartedChild(t *testing.T) {
	catalog, _ := gateway.NewSessionCatalog(nil)
	for _, tt := range []struct {
		name, mode, address, overlay, model string
		startErr                            bool
	}{
		{name: "wrong mode", mode: "glm"}, {name: "startup failure", mode: "gpt", startErr: true}, {name: "invalid handoff", mode: "gpt", address: "0.0.0.0:3", overlay: "fixture"}, {name: "missing overlay", mode: "gpt", address: "127.0.0.1:3"}, {name: "unknown model", mode: "gpt", address: "127.0.0.1:3", overlay: "fixture", model: "unknown"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			starts, stops := 0, 0
			binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog, Start: func(context.Context, gateway.StartOptions) (gatewayStartedChild, error) {
				starts++
				if tt.startErr {
					return gatewayStartedChild{}, errors.New("start failed")
				}
				return gatewayStartedChild{Address: tt.address, OverlayPath: tt.overlay, Stop: func() { stops++ }}, nil
			}})
			_, _, err := binding.Prepare(gatewayLaunchRequest{Mode: tt.mode, ExplicitModel: tt.model})
			if err == nil {
				t.Fatal("accepted failure")
			}
			if starts > 0 && !tt.startErr && stops != 1 {
				t.Fatalf("started child leaked, stops %d", stops)
			}
		})
	}
	binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog})
	if _, _, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt"}); err == nil {
		t.Fatal("missing executable accepted")
	}
}

func TestGatewayCLIChildHelper(t *testing.T) {
	if os.Getenv("GATEWAY_CLI_HELPER") != "1" {
		return
	}
	cmd := newGatewayChildCommand(func(json.RawMessage) (http.Handler, error) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }), nil
	})
	cmd.SetArgs(nil)
	cmd.SetIn(os.Stdin)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(io.Discard)
	if err := cmd.Execute(); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func TestGatewaySessionRealSupervisorHandoffAndCleanup(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	catalog, _ := gateway.NewSessionCatalog(nil)
	binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog, Child: gateway.StartOptions{Executable: executable, Args: []string{"-test.run=^TestGatewayCLIChildHelper$"}, Env: append(os.Environ(), "GATEWAY_CLI_HELPER=1"), StartupTimeout: 3 * time.Second, Config: gateway.ChildConfig{Lifetime: 10 * time.Second, PollInterval: 20 * time.Millisecond}}})
	plan, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stop)
	var address string
	for _, item := range plan.ChildEnv {
		if strings.HasPrefix(item, "ANTHROPIC_BASE_URL=") {
			address = strings.TrimPrefix(item, "ANTHROPIC_BASE_URL=")
		}
	}
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(address)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatal(response.StatusCode)
	}
	stop()
	if _, err := os.Stat(plan.ChildSettings); !os.IsNotExist(err) {
		t.Fatalf("owned overlay survived: %v", err)
	}
	response, err = client.Get(address)
	if err == nil {
		response.Body.Close()
		t.Fatal("listener survived stop")
	}
}

func TestGatewaySessionSettingsCannotRedirectSelectedProfile(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	catalog, _ := gateway.NewSessionCatalog(nil)
	binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog, Start: func(context.Context, gateway.StartOptions) (gatewayStartedChild, error) {
		return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "fixture", Stop: func() {}}, nil
	}})
	plan, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", ProfileName: "work", Args: []string{"--settings", `{"env":{"CLAUDE_CONFIG_DIR":"wrong-profile"}}`}})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	want := "CLAUDE_CONFIG_DIR=" + profile.GetProfileDir("work")
	count := 0
	for _, item := range plan.ChildEnv {
		if strings.HasPrefix(item, "CLAUDE_CONFIG_DIR=") {
			count++
			if item != want {
				t.Fatalf("profile override %q", item)
			}
		}
	}
	if count != 1 {
		t.Fatalf("profile entries %d", count)
	}
}

func TestGatewaySessionDisablesInheritedFallbackWithoutMutatingCaller(t *testing.T) {
	catalog, _ := gateway.NewSessionCatalog(nil)
	for _, args := range [][]string{nil, {"--settings", `{"fallbackModel":["claude-sonnet-5"],"availableModels":["gpt-6-astra"]}`}} {
		caller := map[string]any{"fallbackModel": []string{"claude-opus-5"}}
		binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog, Overlay: caller, Start: func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
			var doc map[string]json.RawMessage
			if err := json.Unmarshal(opts.Config.Overlay, &doc); err != nil {
				t.Fatal(err)
			}
			if string(doc["fallbackModel"]) != "[]" {
				t.Errorf("fallbackModel = %s, want []", doc["fallbackModel"])
			}
			if string(doc["availableModels"]) != `["gpt-6-astra","gpt-5.6-sol","gpt-5.6-terra","gpt-5.6-luna"]` {
				t.Fatal("availableModels not bound to provider")
			}
			return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "/fixture/settings.json"}, nil
		}})
		if _, _, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", Args: args}); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(caller, map[string]any{"fallbackModel": []string{"claude-opus-5"}}) {
			t.Fatal("caller overlay mutated")
		}
	}
}

func TestGatewaySessionRejectsExplicitFallbackBeforeChildStart(t *testing.T) {
	for _, args := range [][]string{{"--fallback-model", "claude-sonnet-5"}, {"--fallback-model=claude-opus-5"}, {"--fallback-model"}, {"--settings", `{}`, "--fallback-model="}} {
		starts := 0
		binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Start: func(context.Context, gateway.StartOptions) (gatewayStartedChild, error) {
			starts++
			return gatewayStartedChild{}, errors.New("unexpected start")
		}})
		_, _, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", Args: args})
		if err == nil || !strings.Contains(err.Error(), "--fallback-model is unsupported in gateway mode") || starts != 0 {
			t.Errorf("args=%v err=%v starts=%d", args, err, starts)
		}
	}
}

func TestGatewaySessionClaudePickerHasFriendlyContextLabel(t *testing.T) {
	rows, _ := gatewayNativeModels("claude", config.GLMModels{})
	catalog, _ := gateway.NewCatalog(rows)
	binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "claude", Catalog: catalog, Start: func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
		var doc struct {
			Picker struct {
				Options []struct {
					Model string `json:"model"`
					Label string `json:"label"`
				} `json:"options"`
			} `json:"modelPicker"`
		}
		if err := json.Unmarshal(opts.Config.Overlay, &doc); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range doc.Picker.Options {
			if row.Model == "claude-opus-5[1m]" {
				found = true
				if row.Label != "Opus 5 (1M context)" {
					t.Fatal(row.Label)
				}
			}
		}
		if !found {
			t.Fatal("qualified Opus missing")
		}
		return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "fixture", Stop: func() {}}, nil
	}})
	_, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: "claude", ExplicitModel: "claude-opus-5[1m]"})
	if stop != nil {
		defer stop()
	}
	if err != nil {
		t.Fatal(err)
	}
}
