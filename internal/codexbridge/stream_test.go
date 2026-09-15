package codexbridge

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

func TestStreamTextArrivesBeforeDurableCompletion(t *testing.T) {
	e, rpc, store := fixture(t)
	rpc.suppressTool = true
	q := request("stream")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	text := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		seg, err := e.StepStream(ctx, q, func(delta string) error { text <- delta; return nil })
		if err == nil && (!seg.Done || seg.Text != "early") {
			err = errors.New("incomplete segment")
		}
		done <- err
	}()
	waitActive(t, e, q)
	rpc.events <- codexapp.Message{Method: "item/agentMessage/delta", Params: json.RawMessage(`{"threadId":"thread-1","turnId":"turn-thread-1","delta":"early"}`)}
	select {
	case got := <-text:
		if got != "early" {
			t.Fatal(got)
		}
	case <-ctx.Done():
		t.Fatal("text buffered until completion")
	}
	b, _, err := store.Barrier(q.Owner)
	if err != nil || b.Phase != "active" {
		t.Fatal(b, err)
	}
	select {
	case err := <-done:
		t.Fatal("completed before terminal event", err)
	default:
	}
	rpc.events <- codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-thread-1","status":"completed"}}`)}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	b, _, err = store.Barrier(q.Owner)
	if err != nil || b.Phase != "idle" {
		t.Fatal(b, err)
	}
}

func waitActive(t *testing.T, e *Engine, q Request) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		e.mu.Lock()
		c := e.conversations[storeKey(q.Owner)]
		e.mu.Unlock()
		if c != nil {
			c.mu.Lock()
			active := c.turn != "" && c.phase == "active"
			c.mu.Unlock()
			if active {
				return
			}
		}
		select {
		case <-deadline:
			t.Fatal("turn did not become active")
		case <-time.After(time.Millisecond):
		}
	}
}

func TestQueuedStepHonorsCancellationWithoutPoisoningOwner(t *testing.T) {
	e, rpc, _ := fixture(t)
	rpc.suppressTool = true
	q := request("queued")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	first := make(chan error, 1)
	go func() { _, err := e.Step(ctx, q); first <- err }()
	waitActive(t, e, q)
	queued, stop := context.WithCancel(ctx)
	stop()
	second := make(chan error, 1)
	go func() { _, err := e.Step(queued, q); second <- err }()
	select {
	case err := <-second:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("canceled request waits behind running turn")
	}
	rpc.events <- codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-thread-1","status":"completed"}}`)}
	if err := <-first; err != nil {
		t.Fatal(err)
	}
}

func TestConfirmedTextOnlyCancellationAllowsNewTurnWithoutReplay(t *testing.T) {
	e, rpc, store := fixture(t)
	rpc.suppressTool = true
	q := request("recover-cancel")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := e.Step(ctx, q); done <- err }()
	waitActive(t, e, q)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	b, _, err := store.Barrier(q.Owner)
	if err != nil || b.Phase == "idle" {
		t.Fatal("cancellation accepted without terminal acknowledgement", b, err)
	}
	rpc.events <- codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-thread-1","status":"interrupted"}}`)}
	deadline := time.After(2 * time.Second)
	for {
		b, _, err = store.Barrier(q.Owner)
		if err != nil {
			t.Fatal(err)
		}
		if b.Phase == "idle" {
			break
		}
		select {
		case <-deadline:
			t.Fatal("confirmed cancellation did not recover", b)
		case <-time.After(time.Millisecond):
		}
	}
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "next-input"
	q.Input = []any{map[string]string{"type": "text", "text": "a distinct new request"}}
	nextCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	go func() { _, err := e.Step(nextCtx, q); done <- err }()
	waitActive(t, e, q)
	rpc.events <- codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-thread-1-2","status":"completed"}}`)}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	if len(rpc.models) != 2 || rpc.responses != 0 {
		t.Fatalf("unexpected replay: turns=%d tool responses=%d", len(rpc.models), rpc.responses)
	}
}

func TestOutputSchemaReachesTurnStart(t *testing.T) {
	e, rpc, _ := fixture(t)
	q := request("schema")
	q.OutputSchema = json.RawMessage(`{"type":"object","properties":{"ok":{"type":"boolean"}},"required":["ok"],"additionalProperties":false}`)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := e.Step(ctx, q); err != nil {
		t.Fatal(err)
	}
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	if len(rpc.schemas) != 1 || string(rpc.schemas[0]) != string(q.OutputSchema) {
		t.Fatal(rpc.schemas)
	}
}

func TestCompletedEphemeralOwnerReleasesCapacity(t *testing.T) {
	e, rpc, _ := fixture(t)
	e.cfg.MaxConversations = 1
	rpc.transform = func(m codexapp.Message) codexapp.Message {
		return codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-thread-1","status":"completed"}}`)}
	}
	q := request("utility")
	q.Ephemeral = true
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if seg, err := e.Step(ctx, q); err != nil || !seg.Done {
		t.Fatal(seg, err)
	}
	e.mu.Lock()
	count := len(e.conversations)
	threads := len(e.threads)
	e.mu.Unlock()
	if count != 0 || threads != 0 {
		t.Fatalf("completed utility retains capacity: conversations=%d threads=%d", count, threads)
	}
}

func TestCancellationRetainsAlreadyQueuedTerminalAcknowledgement(t *testing.T) {
	e, rpc, store := fixture(t)
	rpc.suppressTool = true
	q := request("queued-terminal")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := e.StepStream(ctx, q, func(string) error {
			e.mu.Lock()
			c := e.conversations[storeKey(q.Owner)]
			e.mu.Unlock()
			if !c.queue.push(codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-thread-1","status":"completed"}}`)}) {
				t.Error("queue admission failed")
			}
			cancel()
			return context.Canceled
		})
		done <- err
	}()
	waitActive(t, e, q)
	rpc.events <- codexapp.Message{Method: "item/agentMessage/delta", Params: json.RawMessage(`{"threadId":"thread-1","turnId":"turn-thread-1","delta":"last"}`)}
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	b, _, err := store.Barrier(q.Owner)
	if err != nil || b.Phase != "idle" {
		t.Fatal("queued completion lost during cancellation", b, err)
	}
}

func TestToolResultImageReachesAppServer(t *testing.T) {
	e, rpc, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	q := request("image-tool")
	first, err := e.Step(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	q.Input = nil
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "image-result"
	q.Results = []ToolResult{{ID: first.Tool.ID, Success: true, Content: []Content{{Type: "inputImage", ImageURL: "data:image/png;base64,aGVsbG8="}}}}
	if result, err := e.Step(ctx, q); err != nil || !result.Done {
		t.Fatal(result, err)
	}
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	if rpc.responses != 1 || len(rpc.resultContents) != 1 || rpc.resultContents[0].ImageURL != q.Results[0].Content[0].ImageURL {
		t.Fatal(rpc.responses, rpc.resultContents)
	}
}

func TestToolBurstAcceptsInformationalEventsAndPreservesText(t *testing.T) {
	for _, kind := range []string{"thread/tokenUsage/updated", "item/started", "item/completed", "item/agentMessage/delta"} {
		t.Run(kind, func(t *testing.T) {
			e, rpc, _ := fixture(t)
			rpc.transform = func(m codexapp.Message) codexapp.Message {
				rpc.events <- m
				return codexapp.Message{Method: kind, Params: json.RawMessage(`{"threadId":"thread-1","turnId":"turn-thread-1","delta":"interleaved"}`)}
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			q := request("burst")
			first, err := e.Step(ctx, q)
			if err != nil || first.Tool == nil {
				t.Fatal(first, err)
			}
			q.Input = nil
			q.ExpectedPrefix = q.PrefixDigest
			q.PrefixDigest = "results"
			q.Results = []ToolResult{{ID: first.Tool.ID, Success: true}}
			last, err := e.Step(ctx, q)
			if err != nil || !last.Done {
				t.Fatal(last, err)
			}
			want := "answer"
			if kind == "item/agentMessage/delta" {
				want = "interleavedanswer"
			}
			if last.Text != want {
				t.Fatal(last.Text, want)
			}
		})
	}
}

func TestRejectedToolResultDoesNotConsumeOrPoisonPendingCall(t *testing.T) {
	for _, kind := range []string{"input", "count", "unknown", "content", "image", "size", "cwd", "prefix"} {
		t.Run(kind, func(t *testing.T) {
			e, rpc, store := fixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			q := request("validation")
			first, err := e.Step(ctx, q)
			if err != nil {
				t.Fatal(err)
			}
			q.Input = nil
			q.ExpectedPrefix = q.PrefixDigest
			q.PrefixDigest = "result"
			q.Results = []ToolResult{{ID: first.Tool.ID, Success: true, Content: []Content{{Type: "inputText", Text: "ok"}}}}
			bad := q
			switch kind {
			case "input":
				bad.Input = []any{"unexpected"}
			case "count":
				bad.Results = nil
			case "unknown":
				bad.Results = []ToolResult{{ID: "foreign"}}
			case "content":
				bad.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "unknown"}}}}
			case "image":
				bad.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputImage"}}}}
			case "size":
				e.cfg.MaxOutputBytes = 1
			case "cwd":
				bad.CWD = "/foreign"
			case "prefix":
				bad.ExpectedPrefix = "foreign"
			}
			if _, err := e.Step(ctx, bad); err == nil {
				t.Fatal("invalid result accepted")
			}
			barrier, _, err := store.Barrier(q.Owner)
			if err != nil || barrier.Phase != "waiting" {
				t.Fatal(barrier, err)
			}
			rpc.mu.Lock()
			responses := rpc.responses
			rpc.mu.Unlock()
			if responses != 0 {
				t.Fatal("invalid result transmitted")
			}
			e.cfg.MaxOutputBytes = 8 << 20
			if last, err := e.Step(ctx, q); err != nil || !last.Done {
				t.Fatal("valid continuation poisoned", last, err)
			}
		})
	}
}
