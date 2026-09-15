package codexbridge

import (
	"context"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

type sandboxContractRPC struct {
	*instructionRPC
	requests map[string]map[string]any
}

func (r *sandboxContractRPC) Call(ctx context.Context, method string, params any, out any) error {
	if method == "thread/start" || method == "thread/resume" || method == "turn/start" {
		r.requests[method] = params.(map[string]any)
	}
	return r.instructionRPC.Call(ctx, method, params, out)
}

func TestWorkspaceSandboxOnStartResumeAndTurn(t *testing.T) {
	_, _, store := fixture(t)
	q := request("workspace-sandbox")
	for _, method := range []string{"thread/start", "thread/resume"} {
		rpc := &sandboxContractRPC{instructionRPC: &instructionRPC{fakeRPC: &fakeRPC{events: make(chan codexapp.Message, 32), suppressTool: true}}, requests: map[string]map[string]any{}}
		engine, err := New(context.Background(), rpc, Config{Store: store})
		if err != nil {
			t.Fatal(err)
		}
		_, err = engine.Step(context.Background(), q)
		engine.Close()
		if err != nil {
			t.Fatal(err)
		}
		if got := rpc.requests[method]["sandbox"]; got != "workspace-write" {
			t.Errorf("%s sandbox=%v want workspace-write", method, got)
		}
		if got := rpc.requests[method]["approvalPolicy"]; got != "never" {
			t.Errorf("%s approvalPolicy=%v want never", method, got)
		}
		want := map[string]any{"type": "workspaceWrite", "networkAccess": false}
		if got := rpc.requests["turn/start"]["sandboxPolicy"]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s subsequent turn sandboxPolicy=%v want %v", method, got, want)
		}
		q.Resume = true
		q.ExpectedPrefix = q.PrefixDigest
		q.PrefixDigest += "-resumed"
	}
}
