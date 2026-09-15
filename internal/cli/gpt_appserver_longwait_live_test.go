package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

// longwaitEventRPC observes the single event stream consumed by the bridge.
// It records method counts only, never message bodies or credential material.
type longwaitEventRPC struct {
	*codexapp.Client
	events chan codexapp.Message
	mu     sync.Mutex
	counts map[string]int
}

func (r *longwaitEventRPC) Events() <-chan codexapp.Message { return r.events }
func (r *longwaitEventRPC) histogram() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	raw, _ := json.Marshal(r.counts)
	return string(raw)
}

func TestSharedGPTLiveLongToolWait(t *testing.T) {
	if os.Getenv("MOAI_GPT_LONG_WAIT_LIVE") != "1" {
		t.Skip("long subscription probe requires MOAI_GPT_LONG_WAIT_LIVE=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	client, err := codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }() // Best-effort teardown; primary assertions own failures.
	if _, err := client.Initialize(ctx, "moai-longwait-live", "1"); err != nil {
		t.Fatal(err)
	}
	rpc := &longwaitEventRPC{Client: client, events: make(chan codexapp.Message), counts: make(map[string]int)}
	forwardCtx, stopForward := context.WithCancel(ctx)
	forwardDone := make(chan struct{})
	// @MX:WARN: [AUTO] Forwarding lifetime is bounded by test cancellation.
	// @MX:REASON: One consumer preserves the original App Server event ordering.
	go func() {
		defer close(forwardDone)
		defer close(rpc.events)
		for {
			select {
			case <-forwardCtx.Done():
				return
			case msg, ok := <-client.Events():
				if !ok {
					return
				}
				rpc.mu.Lock()
				rpc.counts[msg.Method]++
				rpc.mu.Unlock()
				select {
				case rpc.events <- msg:
				case <-forwardCtx.Done():
					return
				}
			}
		}
	}()
	defer func() { stopForward(); <-forwardDone }()
	store := managedGPTTestStore(t)
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	q := codexbridge.Request{Owner: codextools.Binding{ConversationID: fmt.Sprintf("longwait-%d", time.Now().UnixNano()), AccountScope: "longwait-live"}, Model: "gpt-5.6-sol", Effort: "low", CWD: t.TempDir(), PrefixDigest: "longwait-1",
		Tools: []codextools.Definition{{Name: "delayed_echo", Description: "Echo the supplied synthetic text after a long operator-controlled delay. Do not retry a pending call.", InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)}},
		Input: []any{map[string]string{"type": "text", "text": "Call delayed_echo exactly three times, with LONG_WAIT_A, LONG_WAIT_B, and LONG_WAIT_C. Launch all three calls concurrently using Promise.all in code mode if supported; otherwise perform all three using the available tool interface. Wait patiently for all three responses; calls may take over two minutes, so never retry a pending call. Then reply exactly LONG_WAIT_COMPLETE LONG_WAIT_A LONG_WAIT_B LONG_WAIT_C. Never use other tools."}},
	}
	started := time.Now()
	seg, err := engine.Step(ctx, q)
	if err != nil {
		t.Fatalf("tool start: %v methods=%s", err, rpc.histogram())
	}
	calls := seg.Tools
	if len(calls) == 0 && seg.Tool != nil {
		calls = []codexbridge.Tool{*seg.Tool}
	}
	if len(calls) == 0 {
		t.Fatalf("no pending tools; done=%v methods=%s", seg.Done, rpc.histogram())
	}
	t.Logf("pending_tools=%d initial_elapsed=%s histogram=%s", len(calls), time.Since(started).Round(time.Millisecond), rpc.histogram())
	waitStarted := time.Now()
	timer := time.NewTimer(130 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
wait:
	for {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
			barrier, _, barrierErr := store.Barrier(q.Owner)
			t.Logf("pending_elapsed=%s phase=%s barrier_error=%v client_error=%v histogram=%s", time.Since(waitStarted).Round(time.Millisecond), barrier.Phase, barrierErr, client.Err(), rpc.histogram())
		case <-timer.C:
			break wait
		}
	}
	t.Logf("delayed_result_elapsed=%s histogram=%s", time.Since(waitStarted).Round(time.Millisecond), rpc.histogram())
	seen := make(map[string]bool)
	maxPending := len(calls)
	for round := 0; round < 4; round++ {
		q.Input = nil
		q.ExpectedPrefix = q.PrefixDigest
		q.PrefixDigest = fmt.Sprintf("longwait-%d", round+2)
		q.Results = nil
		for _, call := range calls {
			var args struct {
				Text string `json:"text"`
			}
			if call.Name != "delayed_echo" || json.Unmarshal(call.Arguments, &args) != nil || (args.Text != "LONG_WAIT_A" && args.Text != "LONG_WAIT_B" && args.Text != "LONG_WAIT_C") {
				t.Fatal("unexpected synthetic tool arguments")
			}
			if seen[args.Text] {
				t.Fatalf("duplicate synthetic tool call: %s", args.Text)
			}
			seen[args.Text] = true
			q.Results = append(q.Results, codexbridge.ToolResult{ID: call.ID, Success: true, Content: []codexbridge.Content{{Type: "inputText", Text: args.Text}}})
		}
		seg, err = engine.Step(ctx, q)
		if err != nil {
			t.Fatalf("delayed continuation: %v total=%s histogram=%s", err, time.Since(started).Round(time.Millisecond), rpc.histogram())
		}
		if seg.Done {
			for _, marker := range []string{"LONG_WAIT_A", "LONG_WAIT_B", "LONG_WAIT_C"} {
				if !seen[marker] || !strings.Contains(seg.Text, marker) {
					t.Fatalf("requested call/result missing: marker=%s call_seen=%v result=%q", marker, seen[marker], seg.Text)
				}
			}
			if !strings.Contains(seg.Text, "LONG_WAIT_COMPLETE") {
				t.Fatalf("completion marker missing: %q", seg.Text)
			}
			t.Logf("done=true calls_returned=%d max_segment_pending=%d total=%s result=%q histogram=%s", len(seen), maxPending, time.Since(started).Round(time.Millisecond), seg.Text, rpc.histogram())
			return
		}
		calls = seg.Tools
		if len(calls) == 0 && seg.Tool != nil {
			calls = []codexbridge.Tool{*seg.Tool}
		}
		if len(calls) == 0 {
			t.Fatal("neither completion nor next synthetic call")
		}
		if len(calls) > maxPending {
			maxPending = len(calls)
		}
	}
	t.Fatal("tool continuation exceeded bounded rounds")
}
