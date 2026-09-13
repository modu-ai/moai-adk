package codexbridge

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// A Claude compaction is classified as a normal summary turn: it completes
// through the standard turn path, its text is the public summary the caller
// contrasts against the authenticated PostCompact, and the bridge issues zero
// explicit compact RPCs of any kind.
func TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs(t *testing.T) {
	e, rpc, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("compact")
	first, err := e.Step(ctx, q)
	if err != nil || first.Tool == nil {
		t.Fatal(err)
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-1"
	q.PrefixDigest = "prefix-2"
	q.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputText", Text: "ok"}}, Success: true}}
	last, err := e.Step(ctx, q)
	if err != nil || !last.Done || last.Text == "" {
		t.Fatal("summary turn did not complete normally", last, err)
	}

	rpc.mu.Lock()
	methods := append([]string(nil), rpc.methods...)
	rpc.mu.Unlock()
	for _, m := range methods {
		if strings.Contains(strings.ToLower(m), "compact") {
			t.Fatal("bridge issued an explicit compact RPC", methods)
		}
	}

	// The returned summary passes the authenticated PostCompact contrastor and
	// rebases exactly once; a different summary for the same notice never does.
	const scope = "profile-generation-1"
	ledger := receipt.NewRebaseLedger(scope, 0)
	base := receipt.CompactBase{Scope: scope, Epoch: 1, Summary: receipt.Hash([]byte(last.Text))}
	if err = ledger.Rebase(base, []byte(last.Text)); err != nil {
		t.Fatal("authenticated summary rejected", err)
	}
	forged := receipt.CompactBase{Scope: scope, Epoch: 2, Summary: receipt.Hash([]byte(last.Text))}
	if err = ledger.Rebase(forged, []byte(last.Text + " synthesized tail")); !strings.Contains(err.Error(), "does not match") {
		t.Fatal("non-exact summary accepted", err)
	}
}

// A server-initiated compaction notification carries no RPC identity and is
// classified as noise: it never fails the bridge and never interrupts the
// running summary turn.
func TestCompactionNotificationDoesNotFailTheTurn(t *testing.T) {
	e, rpc, _ := fixture(t)
	rpc.suppressTool = true
	rpc.startEntered = make(chan struct{})
	rpc.startWait = make(chan struct{})
	close(rpc.startWait)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type outcome struct {
		seg Segment
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		seg, err := e.Step(ctx, request("notify"))
		done <- outcome{seg, err}
	}()
	<-rpc.startEntered
	notification, _ := json.Marshal(map[string]any{"threadId": "thread-1"})
	delta, _ := json.Marshal(map[string]any{"threadId": "thread-1", "turnId": "turn-thread-1", "itemId": "text", "delta": "compacted summary"})
	completed, _ := json.Marshal(map[string]any{"threadId": "thread-1", "turn": map[string]string{"id": "turn-thread-1", "status": "completed"}})
	rpc.mu.Lock()
	rpc.events <- codexapp.Message{Method: "thread/compacted", Params: notification}
	rpc.events <- codexapp.Message{Method: "item/agentMessage/delta", Params: delta}
	rpc.events <- codexapp.Message{Method: "turn/completed", Params: completed}
	rpc.mu.Unlock()
	got := <-done
	if got.err != nil || !got.seg.Done || got.seg.Text != "compacted summary" {
		t.Fatal("compaction notification disturbed the turn", got.seg, got.err)
	}
}
