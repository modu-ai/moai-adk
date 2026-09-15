package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestGPTContextMetadataFailsClosedBeforeBinding(t *testing.T) {
	for _, kind := range []string{"missing", "valid", "future", "malformed", "duplicate-key", "missing-model", "different", "zero", "fraction", "string", "max-only"} {
		t.Run(kind, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("MOAI_HOME", home)
			profile := filepath.Join(home, "gpt-appserver")
			if err := os.Mkdir(profile, 0700); err != nil {
				t.Fatal(err)
			}
			window := 272000
			if kind == "future" {
				window = 300000
			}
			rows := []map[string]any{}
			for _, row := range gatewayGPTModels() {
				rows = append(rows, map[string]any{"slug": row.RouteID, "context_window": window, "max_context_window": 872000})
			}
			switch kind {
			case "missing-model":
				rows = rows[:3]
			case "different":
				rows[3]["context_window"] = 250000
			case "zero":
				rows[0]["context_window"] = 0
			case "fraction":
				rows[0]["context_window"] = 272000.5
			case "string":
				rows[0]["context_window"] = "272000"
			case "max-only":
				delete(rows[0], "context_window")
			}
			raw, err := json.Marshal(map[string]any{"models": rows})
			if err != nil {
				t.Fatal(err)
			}
			if kind == "malformed" {
				raw = []byte(`not-json`)
			}
			if kind == "duplicate-key" {
				raw = []byte(strings.Replace(string(raw), `"context_window":272000`, `"context_window":872000,"context_window":272000`, 1))
			}
			if kind != "missing" {
				if err := os.WriteFile(filepath.Join(profile, "models_cache.json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err = newGPTGatewayBinding()
			good := kind == "missing" || kind == "valid" || kind == "future"
			if (err == nil) != good {
				t.Fatalf("binding error=%v, want success=%v", err, good)
			}
			models, err := installedGatewayGPTModels()
			if (err == nil) != good {
				t.Fatalf("handler error=%v, want success=%v", err, good)
			}
			if good {
				catalog, err := gateway.NewCatalog(models)
				if err != nil {
					t.Fatal(err)
				}
				plan, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: "gpt", ExplicitModel: "gpt-5.6-sol", Catalog: catalog, Address: "127.0.0.1:1234"})
				if err != nil {
					t.Fatal(err)
				}
				want, err := json.Marshal(window)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains("\n"+strings.Join(plan.ChildEnv, "\n")+"\n", "\nCLAUDE_CODE_MAX_CONTEXT_TOKENS="+string(want)+"\n") {
					t.Fatal("launch window missing", plan.ChildEnv)
				}
				for _, row := range models {
					if row.Capabilities.ContextTokens != window {
						t.Fatalf("window=%d want %d", row.Capabilities.ContextTokens, window)
					}
				}
			}
		})
	}
}

func TestGPTContextSnapshotSurvivesCacheChange(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	models := gatewayGPTModels()
	for i := range models {
		models[i].Capabilities.ContextTokens = 300000
	}
	payload, err := marshalGatewayPrivatePayload("snapshot-test", models)
	if err != nil {
		t.Fatal(err)
	}
	// No cache exists in the child, but the parent already verified its snapshot.
	_, rows, err := decodeGPTAppServerPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Capabilities.ContextTokens != 300000 {
			t.Fatalf("snapshot lost: got %d", row.Capabilities.ContextTokens)
		}
	}
}

func TestGPTContextSnapshotRejectsInvalidExplicitWindow(t *testing.T) {
	for _, window := range []string{"0", "-1", "null", "2000001", "1.5", `"272000"`} {
		raw := []byte(`{"version":1,"session_token":"snapshot","model_ids":["gpt-5.6-sol"],"context_tokens":` + window + `}`)
		if _, _, err := decodeGPTAppServerPayload(raw); err == nil {
			t.Errorf("invalid explicit window accepted: %s", window)
		}
	}
}
