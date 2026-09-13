package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

// gatewayAssemblyFixture catalogs mirror the production binding builders so
// the assembly test reads the same provider-exclusive rows a real launch sees.
func gatewayAssemblyFixture(t *testing.T, mode string) (gateway.CatalogSnapshot, config.GLMModels) {
	t.Helper()
	switch mode {
	case "gpt":
		catalog, err := gateway.NewCatalog(gatewayGPTModels())
		if err != nil {
			t.Fatal(err)
		}
		return catalog, config.GLMModels{}
	case "claude":
		rows, err := gatewayNativeModels("claude", config.GLMModels{})
		if err != nil {
			t.Fatal(err)
		}
		catalog, err := gateway.NewCatalog(rows)
		if err != nil {
			t.Fatal(err)
		}
		return catalog, config.GLMModels{}
	case "glm":
		tiers := config.GLMModels{High: "glm-tier-high", Medium: "glm-tier-medium", Low: "glm-tier-low", Fable: "glm-tier-fable"}
		rows, err := gatewayNativeModels("glm", tiers)
		if err != nil {
			t.Fatal(err)
		}
		catalog, err := gateway.NewCatalog(rows)
		if err != nil {
			t.Fatal(err)
		}
		return catalog, tiers
	}
	t.Fatalf("unknown mode %q", mode)
	return gateway.CatalogSnapshot{}, config.GLMModels{}
}

// TestGatewayLaunchAssemblyCarriesAuthDisplay pins the AC-MG-026 (a) assembly
// combination: one launch assembly carries, for every launcher mode, the
// provider-specific auth-method display and the App Server output-policy
// marker alongside the provider-exclusive catalog. The displayed auth method
// must match the authentication the child actually receives (session env), so
// the display cannot drift from the real selection (AS-014/AS-021 mock level).
func TestGatewayLaunchAssemblyCarriesAuthDisplay(t *testing.T) {
	cases := []struct {
		mode, sessionEnv, wantAuthDisplay string
		glmKey                            string
	}{
		{mode: "gpt", sessionEnv: "ANTHROPIC_AUTH_TOKEN=fixture-token", wantAuthDisplay: "subscription"},
		{mode: "claude", sessionEnv: "ANTHROPIC_CUSTOM_HEADERS=X-MoAI-Session-Token: fixture-token", wantAuthDisplay: "subscription"},
		{mode: "glm", glmKey: "fixture-glm-key", wantAuthDisplay: "existing-credential"},
	}
	for _, tc := range cases {
		t.Run(tc.mode, func(t *testing.T) {
			catalog, tiers := gatewayAssemblyFixture(t, tc.mode)
			var overlay map[string]any
			binding := newGatewaySessionBinding(gatewaySessionOptions{
				Mode: tc.mode, Catalog: catalog, GLM: tiers, GLMKey: tc.glmKey,
				SessionEnv: []string{tc.sessionEnv},
				Start: func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
					if err := json.Unmarshal(opts.Config.Overlay, &overlay); err != nil {
						return gatewayStartedChild{}, err
					}
					return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "/fixture/settings.json", Stop: func() {}}, nil
				},
			})
			plan, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: tc.mode, Inherited: []string{"KEEP=parent"}})
			if err != nil {
				t.Fatal(err)
			}
			stop()
			if overlay["authMethod"] != tc.wantAuthDisplay {
				t.Fatalf("authMethod display = %v, want %q (overlay: %v)", overlay["authMethod"], tc.wantAuthDisplay, overlay)
			}
			if overlay["outputPolicy"] != "app-server" {
				t.Fatalf("outputPolicy display = %v, want %q", overlay["outputPolicy"], "app-server")
			}
			// Display/selection consistency: the same auth surface the display
			// names must be what the child env actually carries.
			if tc.mode != "glm" && !authDisplayMatchesSessionEnv(tc.wantAuthDisplay, plan.ChildEnv) {
				t.Fatalf("auth display %q has no matching session env in %v", tc.wantAuthDisplay, plan.ChildEnv)
			}
			if tc.mode == "glm" && !authDisplayMatchesSessionEnv(tc.wantAuthDisplay, plan.ChildEnv) {
				t.Fatalf("glm auth display %q has no matching session env in %v", tc.wantAuthDisplay, plan.ChildEnv)
			}
		})
	}
}

// authDisplayMatchesSessionEnv ties the displayed auth method to the concrete
// child-env authentication carrier for its method class.
func authDisplayMatchesSessionEnv(display string, env []string) bool {
	for _, item := range env {
		switch display {
		case "subscription":
			if strings.HasPrefix(item, "ANTHROPIC_AUTH_TOKEN=") || strings.HasPrefix(item, "ANTHROPIC_CUSTOM_HEADERS=X-MoAI-Session-Token:") {
				return true
			}
		case "existing-credential":
			if strings.HasPrefix(item, "Z_AI_API_KEY=") {
				return true
			}
		case "api-key":
			if strings.HasPrefix(item, "ANTHROPIC_API_KEY=") {
				return true
			}
		}
	}
	return false
}

// TestGatewayNativeModelCapabilitiesPerProvider pins the AS-020 capability
// re-verdict of the production native model declarations: GLM routes are
// text-only with the officially documented nominal 200K context, Claude
// routes declare image acceptance with the 1M nominal context. The declared
// numbers are nominal metadata, never an acceptance guarantee.
func TestGatewayNativeModelCapabilitiesPerProvider(t *testing.T) {
	glmRows, err := gatewayNativeModels("glm", config.GLMModels{High: "glm-tier-high", Medium: "glm-tier-medium", Low: "glm-tier-low", Fable: "glm-tier-fable"})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range glmRows {
		if row.Provider != gateway.ProviderZAI {
			t.Fatalf("glm row %s has provider %s", row.RouteID, row.Provider)
		}
		if row.Capabilities.Images {
			t.Fatalf("glm row %s declares image acceptance; GLM is text-only (AS-020)", row.RouteID)
		}
		if row.Capabilities.ContextTokens != 200000 {
			t.Fatalf("glm row %s nominal context = %d, want the documented 200000", row.RouteID, row.Capabilities.ContextTokens)
		}
	}
	claudeRows, err := gatewayNativeModels("claude", config.GLMModels{})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range claudeRows {
		if row.Provider != gateway.ProviderAnthropic {
			t.Fatalf("claude row %s has provider %s", row.RouteID, row.Provider)
		}
		if !row.Capabilities.Images {
			t.Fatalf("claude row %s lost its declared image capability", row.RouteID)
		}
		if row.Capabilities.ContextTokens != 1000000 {
			t.Fatalf("claude row %s nominal context = %d, want 1000000", row.RouteID, row.Capabilities.ContextTokens)
		}
	}
}

// TestGatewayLaunchTransportGateControl is the AC-MG-026 (a) control group:
// until the AS-014~AS-022 verification gate has passed, BOTH wait-error sites
// (internal/cli/gpt.go and internal/cli/launcher.go) retain the transport
// wait error and a binding-less GPT launch refuses to proceed. When the gate
// passes, the literals are removed and this test flips with them — the flip
// is the (a) GREEN act, never a silent drift.
func TestGatewayLaunchTransportGateControl(t *testing.T) {
	const waitLiteral = "awaiting transport verification"

	// 1. Both known non-test gate sites still carry the literal (source scan,
	// same judgment surface as the AC's grep command).
	for _, name := range []string{"gpt.go", "launcher.go"} {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Count(string(raw), waitLiteral); got != 1 {
			t.Fatalf("%s carries %d occurrences of the wait literal, want 1 (control group)", name, got)
		}
	}

	// 2. A binding-less GPT launch through the shared path ends with the wait
	// error and never reaches project resolution or launch assembly.
	err := unifiedLaunchWithGateway("default", "gpt", nil, nil)
	if err == nil || !strings.Contains(err.Error(), waitLiteral) {
		t.Fatalf("binding-less gpt launch error = %v, want the transport wait error", err)
	}

	// 3. The gpt command with no launch service refuses the same way.
	cmd := newGPTCommand(gptCommandServices{})
	if err := cmd.RunE(cmd, nil); err == nil || !strings.Contains(err.Error(), waitLiteral) {
		t.Fatalf("nil-launch gpt command error = %v, want the transport wait error", err)
	}

	// 4. A nil binding for a non-gpt mode is unaffected by the gate (claude
	// keeps its legacy non-gateway path).
	if err := unifiedLaunchDefaultWithFactory("default", "cg", nil, nil); !errors.Is(err, errCGRetired) {
		t.Fatalf("cg mode error = %v, want errCGRetired", err)
	}
}
