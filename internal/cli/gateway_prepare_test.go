package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestGatewayPreparationBindsModeToInitialProvider(t *testing.T) {
	models := config.GLMModels{High: "glm-high", Medium: "glm-medium", Low: "glm-low", Fable: "glm-fable"}
	catalog, err := gateway.NewSessionCatalog([]string{models.High, models.Medium, models.Low, models.Fable})
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ mode, explicit, want, provider string }{
		{"claude", "", "claude-opus-5", "claude"}, {"gpt", "", "gpt-5.6-sol", "gpt"}, {"glm", "", "glm-high", "glm"},
	} {
		t.Run(tt.mode+tt.explicit, func(t *testing.T) {
			env := []string{"KEEP=value", "ANTHROPIC_AUTH_TOKEN=old", "ANTHROPIC_BASE_URL=https://old.invalid", "Z_AI_API_KEY=old", "ANTHROPIC_DEFAULT_OPUS_MODEL=old", "MOAI_BACKUP_AUTH_TOKEN=old"}
			got, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: tt.mode, ExplicitModel: tt.explicit, ClaudeDefault: "opus", GLM: models, GLMKey: "resolved", Catalog: catalog, Inherited: env, Address: "127.0.0.1:4321", SessionEnv: []string{"FIXTURE_SESSION_HEADER=private"}})
			if err != nil {
				t.Fatal(err)
			}
			if got.InitialModel != tt.want {
				t.Fatalf("model %q", got.InitialModel)
			}
			values := map[string]string{}
			for _, item := range got.ChildEnv {
				k, v, _ := strings.Cut(item, "=")
				values[k] = v
			}
			if values[config.EnvMoaiLaunchProvider] != tt.provider || values["KEEP"] != "value" || values["ANTHROPIC_BASE_URL"] != "http://127.0.0.1:4321" || values["ANTHROPIC_AUTH_TOKEN"] != "" || values["MOAI_BACKUP_AUTH_TOKEN"] != "" {
				t.Fatalf("unexpected child environment %v", values)
			}
			if tt.mode == "glm" {
				if values["Z_AI_API_KEY"] != "resolved" || values["ANTHROPIC_DEFAULT_OPUS_MODEL"] != "glm-high" || values["ANTHROPIC_DEFAULT_FABLE_MODEL"] != "glm-fable" {
					t.Fatal("GLM launcher values missing")
				}
			} else if values["Z_AI_API_KEY"] != "" || values["ANTHROPIC_DEFAULT_OPUS_MODEL"] != tt.want {
				t.Fatal("inherited GLM values survived")
			}
			if env[1] != "ANTHROPIC_AUTH_TOKEN=old" {
				t.Fatal("mutated parent")
			}
		})
	}
}

func TestGatewayPreparationRemovesExactFourteenInheritedKeys(t *testing.T) {
	catalog, _ := gateway.NewSessionCatalog(nil)
	keys := []string{"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL", "ANTHROPIC_DEFAULT_OPUS_MODEL", "ANTHROPIC_DEFAULT_SONNET_MODEL", "ANTHROPIC_DEFAULT_HAIKU_MODEL", "ANTHROPIC_DEFAULT_FABLE_MODEL", "MOAI_BACKUP_AUTH_TOKEN", "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS", "API_TIMEOUT_MS", "CLAUDE_CODE_AUTO_COMPACT_WINDOW", "CLAUDE_CODE_MAX_CONTEXT_TOKENS", "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC", "CLAUDE_CODE_TEAMMATE_DISPLAY", "MOAI_STATUSLINE_CONTEXT_SIZE"}
	var inherited []string
	for _, key := range keys {
		inherited = append(inherited, key+"=polluted")
	}
	got, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: "gpt", Catalog: catalog, Inherited: inherited, Address: "127.0.0.1:1234"})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range got.ChildEnv {
		if strings.HasSuffix(item, "=polluted") {
			t.Fatalf("inherited key survived: %s", item)
		}
	}
}

func TestGatewayGLMExplicitTierAliasesKeepConfiguredRouting(t *testing.T) {
	models := config.GLMModels{High: "glm-high", Medium: "glm-medium", Low: "glm-low", Fable: "glm-fable"}
	catalog, _ := gateway.NewSessionCatalog([]string{models.High, models.Medium, models.Low, models.Fable})
	for alias, want := range map[string]string{"opus": models.High, "sonnet": models.Medium, "haiku": models.Low, "fable": models.Fable} {
		got, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: "glm", ExplicitModel: alias, GLM: models, GLMKey: "fixture", Catalog: catalog, Address: "127.0.0.1:3"})
		if err != nil {
			t.Fatal(err)
		}
		if got.InitialModel != want {
			t.Fatalf("%s resolved %s, want %s", alias, got.InitialModel, want)
		}
	}
}
