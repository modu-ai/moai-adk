package codexbridge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

type fakeRPC struct {
	mu             sync.Mutex
	events         chan codexapp.Message
	next           int
	responses      int
	interrupts     []string
	fail           error
	startEntered   chan struct{}
	startWait      chan struct{}
	transform      func(codexapp.Message) codexapp.Message
	suppressTool   bool
	respondFailure func() error
}

func (f *fakeRPC) Call(ctx context.Context, method string, p any, out any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m := p.(map[string]any)
	var result any
	switch method {
	case "thread/start":
		f.next++
		result = map[string]any{"thread": map[string]string{"id": fmt.Sprint("thread-", f.next)}}
	case "turn/start":
		if f.startEntered != nil {
			close(f.startEntered)
			<-f.startWait
		}
		thread := m["threadId"].(string)
		turn := "turn-" + thread
		result = map[string]any{"turn": map[string]string{"id": turn}}
		raw, _ := json.Marshal(map[string]any{"threadId": thread, "turnId": turn, "callId": "call-" + thread, "tool": "echo", "arguments": map[string]string{"text": "hello"}})
		event := codexapp.Message{ID: json.RawMessage(fmt.Sprintf("%q", thread)), Method: "item/tool/call", Params: raw}
		if f.transform != nil {
			event = f.transform(event)
		}
		if !f.suppressTool {
			f.events <- event
		}
	case "turn/interrupt":
		f.interrupts = append(f.interrupts, m["threadId"].(string))
		result = map[string]any{}
	default:
		return errors.New("unexpected method")
	}
	raw, _ := json.Marshal(result)
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}
func (f *fakeRPC) Respond(ctx context.Context, id json.RawMessage, p any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses++
	if f.respondFailure != nil {
		return f.respondFailure()
	}
	var thread string
	json.Unmarshal(id, &thread)
	raw, _ := json.Marshal(map[string]any{"threadId": thread, "turnId": "turn-" + thread, "itemId": "text", "delta": "answer"})
	f.events <- codexapp.Message{Method: "item/agentMessage/delta", Params: raw}
	raw, _ = json.Marshal(map[string]any{"threadId": thread, "turn": map[string]string{"id": "turn-" + thread, "status": "completed"}})
	f.events <- codexapp.Message{Method: "turn/completed", Params: raw}
	return nil
}
func (f *fakeRPC) Events() <-chan codexapp.Message { return f.events }
func (f *fakeRPC) Err() error                      { return f.fail }
func request(id string) Request {
	return Request{Owner: codextools.Binding{ConversationID: id, AccountScope: "profile-generation-1"}, Model: "gpt-5.6-sol", CWD: "/tmp", Tools: []codextools.Definition{{Name: "echo", InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)}}, Input: []any{map[string]string{"type": "text", "text": "hello"}}, PrefixDigest: "prefix-1"}
}
func fixture(t *testing.T) (*Engine, *fakeRPC, *FileStore) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	rpc := &fakeRPC{events: make(chan codexapp.Message, 32)}
	engine, err := New(context.Background(), rpc, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { engine.Close() })
	return engine, rpc, store
}
func TestToolSegmentRetainsRPCAndCompletesExactlyOnce(t *testing.T) {
	e, rpc, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("a")
	first, err := e.Step(ctx, q)
	if err != nil || first.Tool == nil || first.Done {
		t.Fatal(first, err)
	}
	if rpc.responses != 0 {
		t.Fatal("premature RPC response")
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-1"
	q.PrefixDigest = "prefix-2"
	q.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputText", Text: "done"}}, Success: true}}
	last, err := e.Step(ctx, q)
	if err != nil || !last.Done || last.Text != "answer" {
		t.Fatal(last, err)
	}
	if _, err = e.Step(ctx, q); err == nil {
		t.Fatal("result replay accepted")
	}
	if rpc.responses != 1 {
		t.Fatal(rpc.responses)
	}
}
func TestTwoConversationSwapAndCancellation(t *testing.T) {
	e, rpc, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a := request("a")
	b := request("b")
	sa, err := e.Step(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	sb, err := e.Step(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	a.Input = nil
	a.ExpectedPrefix = "prefix-1"
	a.Results = []ToolResult{{ID: sb.Tool.ID, Success: true, Content: []Content{{Type: "inputText", Text: "wrong"}}}}
	if _, err := e.Step(ctx, a); err == nil {
		t.Fatal("foreign result consumed")
	}
	if rpc.responses != 0 {
		t.Fatal("foreign RPC response")
	}
	if err := e.Cancel(ctx, a.Owner); err != nil {
		t.Fatal(err)
	}
	a.Results[0].ID = sa.Tool.ID
	if _, err := e.Step(ctx, a); err == nil {
		t.Fatal("canceled late result accepted")
	}
	b.Input = nil
	b.ExpectedPrefix = "prefix-1"
	b.Results = []ToolResult{{ID: sb.Tool.ID, Success: true, Content: []Content{{Type: "inputText", Text: "ok"}}}}
	if last, err := e.Step(ctx, b); err != nil || !last.Done {
		t.Fatal(last, err)
	}
	if len(rpc.interrupts) != 1 || rpc.interrupts[0] != "thread-1" {
		t.Fatal(rpc.interrupts)
	}
}
func TestRestartPendingNeverReplays(t *testing.T) {
	e, _, store := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("crash")
	if _, err := e.Step(ctx, q); err != nil {
		t.Fatal(err)
	}
	e.Close()
	freshRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
	fresh, err := New(ctx, freshRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	if _, err := fresh.Step(ctx, q); !errors.Is(err, ErrRecovery) {
		t.Fatal(err)
	}
	if freshRPC.next != 0 || freshRPC.responses != 0 {
		t.Fatal("old RPC replayed")
	}
}

func TestPendingBarrierIsDurableBeforeExposure(t *testing.T) {
	e, _, store := fixture(t)
	q := request("durable")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	segment, err := e.Step(ctx, q)
	if err != nil || segment.Tool == nil {
		t.Fatal(segment, err)
	}
	raw, err := os.ReadFile(filepath.Join(store.dir, storeKey(q.Owner)+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved record
	if err = json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Phase != "waiting" || saved.CallID != "call-thread-1" || saved.TurnID != "turn-thread-1" || saved.Owner.AccountScope != q.Owner.AccountScope {
		t.Fatal(saved)
	}
	if bytes.Contains(raw, []byte("hello")) || bytes.Contains(raw, []byte(segment.Tool.ID)) {
		t.Fatal("state retained tool payload or public result credential")
	}
}

func TestEOFAfterToolResultDoesNotCompleteSuccessfully(t *testing.T) {
	e, rpc, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("eof")
	first, err := e.Step(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	close(rpc.events)
	<-e.done
	q.Input = nil
	q.ExpectedPrefix = "prefix-1"
	q.Results = []ToolResult{{ID: first.Tool.ID, Success: true}}
	segment, err := e.Step(ctx, q)
	if !errors.Is(err, io.EOF) || segment.Done || rpc.responses != 0 {
		t.Fatal(segment, err, rpc.responses)
	}
}

func TestOutputLimitHasNoSuccessTerminal(t *testing.T) {
	e, rpc, _ := fixture(t)
	e.cfg.MaxOutputBytes = 3 // before any Step/event consumption
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("limit")
	first, err := e.Step(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-1"
	q.Results = []ToolResult{{ID: first.Tool.ID, Success: true}}
	segment, err := e.Step(ctx, q)
	if !errors.Is(err, ErrLimit) || segment.Done || rpc.responses != 1 {
		t.Fatal(segment, err, rpc.responses)
	}
}

func TestCancelBeforeTurnStartReplyInterruptsAllocatedTurn(t *testing.T) {
	e, rpc, _ := fixture(t)
	rpc.startEntered = make(chan struct{})
	rpc.startWait = make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("cancel-start")
	done := make(chan error, 1)
	go func() { _, err := e.Step(ctx, q); done <- err }()
	select {
	case <-rpc.startEntered:
	case err := <-done:
		t.Fatal("turn/start never reached", err)
	}
	if err := e.Cancel(ctx, q.Owner); err != nil {
		t.Fatal(err)
	}
	close(rpc.startWait)
	if err := <-done; err == nil {
		t.Fatal("canceled turn succeeded")
	}
	e.workers.Wait()
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	if len(rpc.interrupts) != 1 || rpc.interrupts[0] != "thread-1" {
		t.Fatal("allocated turn not interrupted", rpc.interrupts)
	}
}

func TestUntrustedTurnEventsFailClosed(t *testing.T) {
	for _, kind := range []string{"wrong-turn", "unknown-rpc", "missing-rpc-id", "unknown-tool", "failed-turn", "invalid-arguments"} {
		t.Run(kind, func(t *testing.T) {
			e, rpc, _ := fixture(t)
			rpc.transform = func(event codexapp.Message) codexapp.Message {
				var p map[string]any
				json.Unmarshal(event.Params, &p)
				switch kind {
				case "wrong-turn":
					p["turnId"] = "foreign-turn"
				case "unknown-rpc":
					event.Method = "item/commandExecution/requestApproval"
				case "missing-rpc-id":
					event.ID = nil
				case "unknown-tool":
					p["tool"] = "not-authorized"
				case "invalid-arguments":
					p["arguments"] = map[string]any{"text": 42}
				case "failed-turn":
					event.ID = nil
					event.Method = "turn/completed"
					p["turn"] = map[string]string{"id": "turn-thread-1", "status": "failed"}
				}
				event.Params, _ = json.Marshal(p)
				return event
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			segment, err := e.Step(ctx, request(kind))
			if err == nil || segment.Done || segment.Tool != nil || rpc.responses != 0 {
				t.Fatal(segment, err, rpc.responses)
			}
		})
	}
}
func TestContextCancellationInterruptsWaitingTurn(t *testing.T) {
	e, rpc, _ := fixture(t)
	rpc.suppressTool = true
	rpc.startEntered = make(chan struct{})
	rpc.startWait = make(chan struct{})
	close(rpc.startWait)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		seg, err := e.Step(ctx, request("timeout"))
		if seg.Done {
			result <- errors.New("canceled turn succeeded")
		} else {
			result <- err
		}
	}()
	select {
	case <-rpc.startEntered:
	case err := <-result:
		t.Fatal("turn never started", err)
	case <-time.After(5 * time.Second):
		t.Fatal("turn never started")
	}
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	e.workers.Wait()
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	if len(rpc.interrupts) != 1 {
		t.Fatal(rpc.interrupts)
	}
}

func TestResultBarrierBeforeUncertainRPCResponse(t *testing.T) {
	e, rpc, store := fixture(t)
	q := request("uncertain")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first, err := e.Step(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	rpc.respondFailure = func() error {
		raw, err := os.ReadFile(filepath.Join(store.dir, storeKey(q.Owner)+".json"))
		if err != nil {
			t.Fatal(err)
		}
		var saved record
		if err = json.Unmarshal(raw, &saved); err != nil {
			t.Fatal(err)
		}
		if saved.Phase != "responding" || saved.CallID == "" {
			t.Fatal(saved)
		}
		return io.ErrUnexpectedEOF
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-1"
	q.Results = []ToolResult{{ID: first.Tool.ID, Success: true}}
	if segment, err := e.Step(ctx, q); !errors.Is(err, io.ErrUnexpectedEOF) || segment.Done {
		t.Fatal(segment, err)
	}
	e.Close()
	freshRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
	fresh, err := New(ctx, freshRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	q.Results = nil
	q.ExpectedPrefix = ""
	q.Input = []any{map[string]string{"type": "text", "text": "retry"}}
	if _, err = fresh.Step(ctx, q); !errors.Is(err, ErrRecovery) {
		t.Fatal(err)
	}
	if freshRPC.responses != 0 || freshRPC.next != 0 {
		t.Fatal("uncertain response replayed")
	}
}
func TestUnsafeStateDirectoryRejected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode test; Windows ACL runtime evidence belongs to CI")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenStore(dir); err == nil {
		t.Fatal("public state directory accepted")
	}
}

func TestScopeDelimiterCannotAliasAnotherConversation(t *testing.T) {
	e, rpc, _ := fixture(t)
	q := request("b\x00c")
	q.Owner.AccountScope = "a"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := e.Step(ctx, q); !errors.Is(err, ErrScope) {
		t.Fatal("ambiguous scope admitted", err)
	}
	if rpc.next != 0 {
		t.Fatal("invalid scope reached App Server")
	}
}

func (f *fakeRPC) DiscardRequest(json.RawMessage) error { return nil }

func TestStopRetainsAlreadyQueuedLateStartIdentity(t *testing.T) {
	e, _, _ := fixture(t)
	q := request("queued-start")
	c, err := e.get(q)
	if err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	c.owner.ThreadID = "owned-thread"
	c.phase = "failed"
	c.queue <- codexapp.Message{Method: "turn/started", Params: json.RawMessage(`{"threadId":"owned-thread","turn":{"id":"late-turn"}}`)}
	e.stop(c)
	got := c.turn
	c.mu.Unlock()
	if got != "late-turn" {
		t.Fatal("stop discarded allocated turn identity", got)
	}
}
