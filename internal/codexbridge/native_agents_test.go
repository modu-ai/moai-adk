package codexbridge

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

type nativeAgentsRPC struct {
	*instructionRPC
	configs map[string]map[string]any
}

func (r *nativeAgentsRPC) Call(ctx context.Context, method string, params any, out any) error {
	if method == "thread/start" || method == "thread/resume" {
		r.configs[method], _ = params.(map[string]any)["config"].(map[string]any)
	}
	return r.instructionRPC.Call(ctx, method, params, out)
}

func TestThreadConfigDisablesNativeAgentsOnStartAndResume(t *testing.T) {
	_, _, store := fixture(t)
	q := request("native-agents-disabled")
	for _, method := range []string{"thread/start", "thread/resume"} {
		rpc := &nativeAgentsRPC{
			instructionRPC: &instructionRPC{fakeRPC: &fakeRPC{events: make(chan codexapp.Message, 32), suppressTool: true}},
			configs:        make(map[string]map[string]any),
		}
		engine, err := New(context.Background(), rpc, Config{Store: store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := engine.Step(context.Background(), q); err != nil {
			engine.Close()
			t.Fatal(err)
		}
		engine.Close()
		for _, key := range []string{"agents.enabled", "features.multi_agent", "features.multi_agent_v2"} {
			if value, ok := rpc.configs[method][key]; !ok || value != false {
				t.Errorf("%s missing false config override %q: %v", method, key, rpc.configs[method])
			}
		}
		q.Resume = true
		q.ExpectedPrefix = q.PrefixDigest
		q.PrefixDigest += "-resumed"
	}
}
