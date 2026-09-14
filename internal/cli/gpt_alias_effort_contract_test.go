package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestGPTAliasEffortMatrix(t *testing.T) {
	wantModels := map[string]string{
		"fable":  "gpt-6-astra",
		"opus":   "gpt-5.6-sol",
		"sonnet": "gpt-5.6-terra",
		"haiku":  "gpt-5.6-luna",
	}
	efforts := []string{"max", "high", "medium", "low", "ultra"}
	catalog, err := gateway.NewCatalog(gatewayGPTModels())
	if err != nil {
		t.Fatal(err)
	}

	for alias, wantModel := range wantModels {
		for _, effort := range efforts {
			alias, wantModel, effort := alias, wantModel, effort
			t.Run(alias+"/"+effort, func(t *testing.T) {
				plan, prepErr := prepareGatewayLaunch(gatewayPrepareInput{
					Mode:      "gpt",
					Catalog:   catalog,
					Address:   "127.0.0.1:43117",
					Inherited: []string{config.EnvClaudeCodeEffortLevel + "=" + effort},
				})
				if prepErr != nil {
					t.Fatalf("alias %s effort %s prepare: %v", alias, effort, prepErr)
				}
				values := gatewayContractEnv(plan.ChildEnv)
				if got := values["ANTHROPIC_DEFAULT_"+strings.ToUpper(alias)+"_MODEL"]; got != wantModel {
					t.Errorf("alias %s: got %s, want %s (effort %s)", alias, got, wantModel, effort)
				}
				if got := values[config.EnvClaudeCodeEffortLevel]; got != effort {
					t.Errorf("alias %s effort: got %s, want %s", alias, got, effort)
				}
			})
		}
	}

	t.Run("effort-omitted", func(t *testing.T) {
		plan, prepErr := prepareGatewayLaunch(gatewayPrepareInput{Mode: "gpt", Catalog: catalog, Address: "127.0.0.1:43117"})
		if prepErr != nil {
			t.Fatal(prepErr)
		}
		if got := gatewayContractEnv(plan.ChildEnv)[config.EnvClaudeCodeEffortLevel]; got != "" {
			t.Errorf("omitted effort: got %q, want empty", got)
		}
		if got := plan.InitialModel; got != "gpt-5.6-sol" {
			t.Errorf("default model: got %s, want gpt-5.6-sol", got)
		}
	})
}

type effortContractRPC struct {
	events chan codexapp.Message
	turn   map[string]any
}

func (r *effortContractRPC) Call(_ context.Context, method string, params any, out any) error {
	m, _ := params.(map[string]any)
	var result any
	switch method {
	case "thread/start":
		result = map[string]any{"thread": map[string]string{"id": "thread-effort"}}
	case "turn/start":
		r.turn = map[string]any{}
		for key, value := range m {
			r.turn[key] = value
		}
		result = map[string]any{"turn": map[string]string{"id": "turn-effort"}}
		raw, _ := json.Marshal(map[string]any{"threadId": "thread-effort", "turnId": "turn-effort", "callId": "call-effort", "tool": "echo", "arguments": map[string]string{"text": "hello"}})
		r.events <- codexapp.Message{ID: json.RawMessage(`"rpc-effort"`), Method: "item/tool/call", Params: raw}
	case "turn/interrupt":
		result = map[string]any{}
	default:
		return fmt.Errorf("unexpected effort RPC %s", method)
	}
	raw, _ := json.Marshal(result)
	return json.Unmarshal(raw, out)
}
func (r *effortContractRPC) Respond(context.Context, json.RawMessage, any) error { return nil }
func (r *effortContractRPC) DiscardRequest(json.RawMessage) error                { return nil }
func (r *effortContractRPC) Events() <-chan codexapp.Message                     { return r.events }
func (r *effortContractRPC) Err() error                                          { return nil }

func TestGPTAliasEffortRPCMatrix(t *testing.T) {
	aliases := []struct{ alias, model string }{{"fable", "gpt-6-astra"}, {"opus", "gpt-5.6-sol"}, {"sonnet", "gpt-5.6-terra"}, {"haiku", "gpt-5.6-luna"}}
	efforts := []string{"max", "high", "medium", "low", "ultra"}
	catalog, err := gateway.NewCatalog(gatewayGPTModels())
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range aliases {
		for _, effort := range efforts {
			row, effort := row, effort
			t.Run(row.alias+"/"+effort, func(t *testing.T) {
				plan, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: "gpt", Catalog: catalog, Address: "127.0.0.1:43117", Inherited: []string{config.EnvClaudeCodeEffortLevel + "=" + effort}})
				if err != nil {
					t.Fatal(err)
				}
				observedModel := gatewayContractEnv(plan.ChildEnv)["ANTHROPIC_DEFAULT_"+strings.ToUpper(row.alias)+"_MODEL"]
				turn := runEffortContractTurn(t, observedModel, effort)
				if got, _ := turn["model"].(string); got != row.model {
					t.Errorf("turn/start.model=%q want=%q alias=%s effort=%s", got, row.model, row.alias, effort)
				}
				if got, _ := turn["effort"].(string); got != effort {
					t.Errorf("turn/start.effort=%q want=%q alias=%s model=%s", got, effort, row.alias, row.model)
				}
			})
		}
	}
	t.Run("effort-omitted", func(t *testing.T) {
		turn := runEffortContractTurn(t, "gpt-5.6-sol", "")
		if got, _ := turn["model"].(string); got != "gpt-5.6-sol" {
			t.Errorf("turn/start.model=%q want=gpt-5.6-sol", got)
		}
		if got, exists := turn["effort"]; exists && got != "" {
			t.Errorf("turn/start.effort=%v want omitted", got)
		}
	})
}

func runEffortContractTurn(t *testing.T, model, effort string) map[string]any {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	store, err := codexbridge.OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	rpc := &effortContractRPC{events: make(chan codexapp.Message, 4)}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	q := codexbridge.Request{Owner: codextools.Binding{ConversationID: "effort-" + model + "-" + effort, AccountScope: "fixture"}, Model: model, CWD: dir, Tools: []codextools.Definition{{Name: "echo", InputSchema: json.RawMessage(`{"type":"object"}`)}}, Input: []any{map[string]string{"type": "text", "text": "hello"}}, PrefixDigest: "prefix-effort"}
	field := reflect.ValueOf(&q).Elem().FieldByName("Effort")
	if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
		field.SetString(effort)
	}
	segment, err := engine.Step(ctx, q)
	if err != nil || segment.Tool == nil {
		t.Fatalf("effort Engine.Step segment=%+v err=%v", segment, err)
	}
	return rpc.turn
}

func gatewayContractEnv(items []string) map[string]string {
	out := make(map[string]string, len(items))
	for _, item := range items {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			out[key] = value
		}
	}
	return out
}
