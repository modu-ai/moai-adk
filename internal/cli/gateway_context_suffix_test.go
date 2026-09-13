package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestGatewayClaudeContextSuffixFromFactoryDefault(t *testing.T) {
	rows, err := gatewayNativeModels("claude", config.GLMModels{})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := gateway.NewCatalog(rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"opus[1m]", "claude-opus-5[1m]", "sonnet[1m]", "claude-sonnet-5[1m]"} {
		for _, explicit := range []bool{false, true} {
			in := gatewayPrepareInput{Mode: "claude", ClaudeDefault: input, Catalog: catalog, Address: "127.0.0.1:4321", Inherited: []string{"CLAUDE_CODE_DISABLE_1M_CONTEXT=1"}}
			if explicit {
				in.ExplicitModel = input
				in.ClaudeDefault = ""
			}
			plan, err := prepareGatewayLaunch(in)
			if err != nil {
				t.Fatalf("model=%s explicit=%v: %v", input, explicit, err)
			}
			if plan.InitialModel != expandModelString(input) {
				t.Fatal(plan.InitialModel)
			}
			entry, err := plan.Catalog.Resolve(plan.InitialModel)
			if err != nil || entry.Provider != gateway.ProviderAnthropic || entry.UpstreamID != strings.TrimSuffix(plan.InitialModel, "[1m]") {
				t.Fatalf("route=%+v err=%v", entry, err)
			}
			for _, env := range plan.ChildEnv {
				if env == "CLAUDE_CODE_DISABLE_1M_CONTEXT=1" {
					t.Fatal("explicit Claude 1M request disabled")
				}
			}
		}
	}
}

func TestGatewayContextSuffixCannotCrossProvider(t *testing.T) {
	catalog, err := gateway.NewSessionCatalog([]string{"glm-test"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ mode, model string }{{"gpt", "claude-opus-5[1m]"}, {"glm", "claude-opus-5[1m]"}, {"gpt", "gpt-5.6-sol[1m]"}, {"claude", "claude-unknown[1m]"}, {"claude", "claude-opus-5[1m][1m]"}} {
		_, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: tt.mode, ExplicitModel: tt.model, Catalog: catalog, Address: "127.0.0.1:4321"})
		if err == nil {
			t.Fatalf("unsupported route admitted: %+v", tt)
		}
	}
}

func TestGatewayGPTDeclaresCatalogWindowForEverySelectableModel(t *testing.T) {
	rows := gatewayGPTModels()
	catalog, err := gateway.NewCatalog(rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		plan, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: "gpt", ExplicitModel: row.RouteID, Catalog: catalog, Address: "127.0.0.1:1234", Inherited: []string{"CLAUDE_CODE_MAX_CONTEXT_TOKENS=200000"}})
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, env := range plan.ChildEnv {
			if strings.HasPrefix(env, "CLAUDE_CODE_MAX_CONTEXT_TOKENS=") {
				count++
				if env != "CLAUDE_CODE_MAX_CONTEXT_TOKENS=872000" {
					t.Fatal(env)
				}
			}
		}
		if count != 1 {
			t.Fatalf("%s: declared windows=%d", row.RouteID, count)
		}
	}
}
