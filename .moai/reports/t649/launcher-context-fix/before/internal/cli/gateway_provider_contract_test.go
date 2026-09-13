package cli

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestGatewayProviderContractRejectsForeignModel(t *testing.T) {
	catalog, _ := gateway.NewSessionCatalog([]string{"glm-test"})
	for _, pair := range [][2]string{{"claude", "gpt-6-astra"}, {"gpt", "claude-opus-5"}, {"glm", "gpt-6-astra"}} {
		_, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: pair[0], ExplicitModel: pair[1], Catalog: catalog, Address: "127.0.0.1:1234", GLMKey: "fixture", GLM: config.GLMModels{High: "glm-test", Medium: "glm-test", Low: "glm-test", Fable: "glm-test"}})
		if err == nil {
			t.Errorf("%s accepted foreign model %s", pair[0], pair[1])
		}
	}
}

func TestGatewayProviderContractPickerAndSlots(t *testing.T) {
	catalog, _ := gateway.NewSessionCatalog(nil)
	var overlay map[string]any
	binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog, Start: func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
		if err := json.Unmarshal(opts.Config.Overlay, &overlay); err != nil {
			t.Fatal(err)
		}
		return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "/fixture/settings.json", Stop: func() {}}, nil
	}})
	plan, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", ExplicitModel: "gpt-6-astra"})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	picker, ok := overlay["modelPicker"].(map[string]any)
	if !ok || picker["replaceBuiltInOptions"] != true {
		t.Errorf("missing private modelPicker: %v", overlay)
	} else {
		options, ok := picker["options"].([]any)
		if !ok || len(options) != 4 {
			t.Errorf("GPT picker options=%v", picker["options"])
		}
		for _, v := range options {
			row := v.(map[string]any)
			if !strings.HasPrefix(row["model"].(string), "gpt-") {
				t.Errorf("foreign picker row %v", row)
			}
		}
	}
	if overlay["model"] != "gpt-6-astra" {
		t.Errorf("default=%v", overlay["model"])
	}
	values := map[string]string{}
	for _, s := range plan.ChildEnv {
		k, v, _ := strings.Cut(s, "=")
		values[k] = v
	}
	for _, tier := range []string{"OPUS", "SONNET", "HAIKU", "FABLE"} {
		if values["ANTHROPIC_DEFAULT_"+tier+"_MODEL"] != "gpt-6-astra" {
			t.Errorf("%s slot not bound", tier)
		}
	}
	for _, e := range plan.Catalog.Entries() {
		if e.Provider != gateway.ProviderOpenAI {
			t.Errorf("foreign catalog route %s", e.RouteID)
		}
	}
}

func TestNativeProviderProductionBindingKeepsCatalogAndFamilySeparate(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	models := config.GLMModels{High: "glm-a", Medium: "glm-b", Low: "glm-c", Fable: "glm-d"}
	for _, mode := range []string{"claude", "glm"} {
		binding, err := newNativeGatewayBindingWithStart(mode, models, "fixture", func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
			var payload gatewayPrivatePayload
			if err := json.Unmarshal(opts.Config.Payload, &payload); err != nil {
				t.Fatal(err)
			}
			for _, id := range payload.ModelIDs {
				if mode == "claude" && !strings.HasPrefix(id, "claude-") || mode == "glm" && !strings.HasPrefix(id, "glm-") {
					t.Errorf("foreign private payload %s", id)
				}
			}
			return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "/fixture/settings.json", Stop: func() {}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		project := t.TempDir()
		plan, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: mode, ClaudeDefault: "opus", CWD: project, Project: project, Inherited: []string{"ANTHROPIC_API_KEY=stale", "ANTHROPIC_CUSTOM_HEADERS=stale"}})
		if err != nil {
			t.Fatal(err)
		}
		stop()
		for _, s := range plan.ChildEnv {
			if strings.HasSuffix(s, "=stale") {
				t.Errorf("stale provider key survived %s", s)
			}
			if strings.HasPrefix(s, "CLAUDE_CONFIG_DIR=") && !strings.Contains(s, "gateway-conversations-"+mode) {
				t.Errorf("provider family shared: %s", s)
			}
		}
	}
}

func TestDefaultLaunchSelectsNativeGatewayBeforeAnyLegacyMutation(t *testing.T) {
	for _, mode := range []string{"claude", "glm"} {
		selected := ""
		sentinel := errors.New("binding selection observed")
		err := unifiedLaunchDefaultWithFactory("default", mode, nil, func(got string) (*gatewayLaunchBinding, error) { selected = got; return nil, sentinel })
		if !errors.Is(err, sentinel) || selected != mode {
			t.Fatalf("mode=%s selection=%s err=%v", mode, selected, err)
		}
	}
}

func TestGatewayExactResumeDoesNotAlsoPassContinue(t *testing.T) {
	fakeMoaiProject(t)
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude")))
	old := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = old })
	calls := 0
	execOrSpawnClaudeFunc = func(_ string, args, env []string) error {
		calls++
		if strings.Contains(strings.Join(args, " "), "--continue") {
			t.Fatal("exact resume also passed continue")
		}
		return nil
	}
	binding := &gatewayLaunchBinding{Mode: "gpt", Prepare: func(gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
		return gateway.LaunchPlan{InitialModel: "gpt-6-astra", Args: []string{"--resume", "11111111-1111-4111-8111-111111111111"}}, nil, nil
	}, Continue: func(string, []string, []string) error { return errors.New("ambiguous continue invoked") }}
	if err := launchClaudeWithGateway("default", []string{"--continue"}, binding); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("exact resume process not launched")
	}
}

func TestGatewayExactModelContextAndGPTToolExposure(t *testing.T) {
	catalog, _ := gateway.NewSessionCatalog([]string{"glm-test"})
	for _, mode := range []string{"claude", "gpt", "glm"} {
		plan, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: mode, Catalog: catalog, Address: "127.0.0.1:1234", Inherited: []string{"CLAUDE_CODE_DISABLE_1M_CONTEXT=0", "ENABLE_TOOL_SEARCH=true"}, GLMKey: "fixture", GLM: config.GLMModels{High: "glm-test", Medium: "glm-test", Low: "glm-test", Fable: "glm-test"}})
		if err != nil {
			t.Fatal(err)
		}
		values := map[string]string{}
		for _, s := range plan.ChildEnv {
			k, v, _ := strings.Cut(s, "=")
			values[k] = v
		}
		if values["CLAUDE_CODE_DISABLE_1M_CONTEXT"] != "1" {
			t.Errorf("%s permits implicit unregistered [1m] ID", mode)
		}
		if mode == "gpt" && values["ENABLE_TOOL_SEARCH"] != "false" {
			t.Error("GPT permits unsupported tool_reference deferral")
		}
	}
}
