// Package codexbridge keeps an official App Server turn alive across Claude
// HTTP tool segments. Callers supply authenticated scope and tool snapshots.
package codexbridge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

var (
	ErrRecovery = errors.New("app server conversation requires explicit recovery")
	ErrScope    = errors.New("invalid App Server conversation or tool result")
	ErrProtocol = errors.New("invalid App Server turn event")
	ErrLimit    = errors.New("app server bridge limit exceeded")
)

type RPC interface {
	Call(context.Context, string, any, any) error
	Respond(context.Context, json.RawMessage, any) error
	DiscardRequest(json.RawMessage) error
	Events() <-chan codexapp.Message
	Err() error
}
type Config struct {
	Store                                       *FileStore
	QueueSize, MaxConversations, MaxOutputBytes int
}
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type ToolResult struct {
	ID      string
	Content []Content
	Success bool
}
type Request struct {
	Owner                                    codextools.Binding
	Model, CWD, ExpectedPrefix, PrefixDigest string
	Tools                                    []codextools.Definition
	References                               []codextools.Reference
	Input                                    []any
	Results                                  []ToolResult
	Resume                                   bool
}
type Tool struct {
	ID, Name  string
	Arguments json.RawMessage
}
type Segment struct {
	Text string
	Tool *Tool
	Done bool
}
type pending struct {
	rpcID            json.RawMessage
	callID, publicID string
}
type conversation struct {
	lateCleanupScheduled            bool
	interruptedTurn                 string
	attach                          bool
	op                              sync.Mutex
	mu                              sync.Mutex
	owner                           codextools.Binding
	registry                        *codextools.Registry
	model, cwd, prefix, phase, turn string
	pending                         *pending
	queue                           chan codexapp.Message
	stopped                         chan struct{}
	stopOnce                        sync.Once
}
type Engine struct {
	rpc           RPC
	cfg           Config
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	conversations map[string]*conversation
	threads       map[string]*conversation
	err           error
	done          chan struct{}
	workers       sync.WaitGroup
}

func New(ctx context.Context, rpc RPC, cfg Config) (*Engine, error) {
	if ctx == nil || ctx.Err() != nil || rpc == nil || cfg.Store == nil {
		return nil, ErrScope
	}
	if cfg.QueueSize == 0 {
		cfg.QueueSize = 64
	}
	if cfg.MaxConversations == 0 {
		cfg.MaxConversations = 128
	}
	if cfg.MaxOutputBytes == 0 {
		cfg.MaxOutputBytes = 8 << 20
	}
	if cfg.QueueSize < 1 || cfg.QueueSize > 4096 || cfg.MaxConversations < 1 || cfg.MaxConversations > 1024 || cfg.MaxOutputBytes < 1 || cfg.MaxOutputBytes > 64<<20 {
		return nil, ErrLimit
	}
	life, cancel := context.WithCancel(ctx)
	e := &Engine{rpc: rpc, cfg: cfg, ctx: life, cancel: cancel, conversations: map[string]*conversation{}, threads: map[string]*conversation{}, done: make(chan struct{})}
	go e.read()
	return e, nil
}
func (e *Engine) Close() { e.mu.Lock(); e.cancel(); e.mu.Unlock(); <-e.done; e.workers.Wait() }
func (e *Engine) failure() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.err != nil {
		return e.err
	}
	return ErrRecovery
}
func (e *Engine) fail(err error) {
	e.mu.Lock()
	if e.err == nil {
		e.err = err
	}
	e.mu.Unlock()
	e.cancel()
}

// @MX:WARN: [AUTO] One reader owns event demultiplexing for all conversations.
// @MX:REASON: A full queue fails the shared transport closed instead of blocking RPC replies.
func (e *Engine) read() {
	defer close(e.done)
	for {
		select {
		case <-e.ctx.Done():
			return
		case event, ok := <-e.rpc.Events():
			if !ok {
				err := e.rpc.Err()
				if err == nil {
					err = io.EOF
				}
				e.fail(err)
				return
			}
			var envelope struct {
				ThreadID string `json:"threadId"`
			}
			if json.Unmarshal(event.Params, &envelope) != nil {
				e.fail(ErrProtocol)
				return
			}
			if envelope.ThreadID == "" {
				if len(event.ID) > 0 {
					e.fail(ErrProtocol)
					return
				}
				continue
			}
			e.mu.Lock()
			c := e.threads[envelope.ThreadID]
			e.mu.Unlock()
			if c == nil {
				if len(event.ID) > 0 {
					e.fail(ErrProtocol)
					return
				}
				continue
			}
			c.mu.Lock()
			select {
			case <-c.stopped:
				c.mu.Unlock()
				e.lateStarted(c, event)
				if len(event.ID) > 0 {
					_ = e.rpc.DiscardRequest(event.ID)
				}
				continue
			default:
			}
			select {
			case c.queue <- event:
				c.mu.Unlock()
			default:
				c.mu.Unlock()
				e.fail(ErrLimit)
				return
			}
		}
	}
}
func validOwner(o codextools.Binding) bool {
	return !strings.ContainsRune(o.AccountScope, 0) && !strings.ContainsRune(o.ConversationID, 0) && o.AccountScope != "" && o.ConversationID != "" && len(o.AccountScope) <= 256 && len(o.ConversationID) <= 256 && o.ThreadID == ""
}
func (e *Engine) get(q Request) (*conversation, error) {
	if !validOwner(q.Owner) || q.Model == "" || len(q.Model) > 256 || q.PrefixDigest == "" || len(q.PrefixDigest) > 256 {
		return nil, ErrScope
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ctx.Err() != nil {
		if e.err != nil {
			return nil, e.err
		}
		return nil, ErrRecovery
	}
	key := storeKey(q.Owner)
	if c := e.conversations[key]; c != nil {
		return c, nil
	}
	if q.Resume {
		return e.resume(q, key)
	}
	if len(q.Results) != 0 || q.ExpectedPrefix != "" {
		return nil, ErrScope
	}
	if len(e.conversations) >= e.cfg.MaxConversations {
		return nil, ErrLimit
	}
	exists, err := e.cfg.Store.exists(q.Owner)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrRecovery
	}
	c := &conversation{owner: q.Owner, model: q.Model, cwd: q.CWD, phase: "new", queue: make(chan codexapp.Message, e.cfg.QueueSize), stopped: make(chan struct{})}
	e.conversations[key] = c
	return c, nil
}
func (e *Engine) save(c *conversation, phase string) error {
	call := ""
	if c.pending != nil {
		call = c.pending.callID
	}
	return e.cfg.Store.save(record{Owner: c.owner, Phase: phase, TurnID: c.turn, CallID: call, Prefix: c.prefix, Model: c.model, CWD: c.cwd})
}

// resume rebuilds a conversation from a durable idle barrier so a new Engine
// can continue the owned App Server thread. The thread identity comes only
// from the record the starting process wrote — never from the caller — and
// the pinned model, working directory and completed public prefix must match
// the resume request exactly. Legacy AS3 barriers are rejected, never
// migrated: they deliberately omit the pinned binding fields a resume needs.
// e.mu must be held.
func (e *Engine) resume(q Request, key string) (*conversation, error) {
	if len(q.Results) != 0 || q.ExpectedPrefix == "" {
		return nil, ErrScope
	}
	rec, found, err := e.cfg.Store.load(q.Owner)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrScope
	}
	if rec.Schema != storeSchema {
		return nil, fmt.Errorf("%w: barrier predates the resumable schema", ErrRecovery)
	}
	if rec.Phase != "idle" {
		return nil, fmt.Errorf("%w: incomplete %q barrier cannot be resumed", ErrRecovery, rec.Phase)
	}
	thread := rec.Owner.ThreadID
	if thread == "" || len(thread) > 256 {
		return nil, ErrProtocol
	}
	if rec.Owner.ConversationID != q.Owner.ConversationID || rec.Owner.AccountScope != q.Owner.AccountScope {
		return nil, ErrScope
	}
	if rec.Model != q.Model || rec.CWD != q.CWD || rec.Prefix != q.ExpectedPrefix {
		return nil, ErrScope
	}
	if len(e.conversations) >= e.cfg.MaxConversations {
		return nil, ErrLimit
	}
	if e.threads[thread] != nil {
		return nil, ErrProtocol
	}
	registry, err := codextools.New(q.Owner, q.Tools)
	if err != nil {
		return nil, err
	}
	owner, err := registry.BindThread(thread)
	if err != nil {
		return nil, err
	}
	c := &conversation{owner: owner, registry: registry, model: q.Model, cwd: q.CWD, prefix: rec.Prefix, phase: "idle", attach: true, queue: make(chan codexapp.Message, e.cfg.QueueSize), stopped: make(chan struct{})}
	e.conversations[key] = c
	e.threads[thread] = c
	return c, nil
}
func (e *Engine) stop(c *conversation) {
	c.stopOnce.Do(func() { close(c.stopped) })
	if c.pending != nil {
		_ = e.rpc.DiscardRequest(c.pending.rpcID)
		c.pending = nil
	}
	for {
		select {
		case event := <-c.queue:
			if c.turn == "" {
				c.turn = startedTurnID(event, c.owner.ThreadID)
			}
			if len(event.ID) > 0 {
				_ = e.rpc.DiscardRequest(event.ID)
			}
		default:
			return
		}
	}
}

// Step returns at a tool boundary without interrupting the running App Server
// turn. Prefixes and references must be derived from authenticated HTTP history,
// never from untrusted model text. Input contains only the new turn's delta.
func (e *Engine) Step(ctx context.Context, q Request) (seg Segment, err error) {
	c, err := e.get(q)
	if err != nil {
		return seg, err
	}
	c.op.Lock()
	defer c.op.Unlock()
	c.mu.Lock()
	if c.phase == "failed" || c.phase == "canceled" {
		c.mu.Unlock()
		return seg, ErrRecovery
	}
	phase := c.phase
	// AS4: only an idle barrier may switch the model — the new value repins the
	// conversation below and rides the next turn/start. A waiting or active
	// turn keeps its pinned model so a switch can never land on the wrong turn.
	if (phase != "idle" && c.model != q.Model) || c.cwd != q.CWD || c.prefix != q.ExpectedPrefix {
		c.mu.Unlock()
		return seg, ErrScope
	}
	if phase == "waiting" {
		if len(q.Input) != 0 || len(q.Results) != 1 || c.pending == nil || q.Results[0].ID != c.pending.publicID {
			c.mu.Unlock()
			return seg, ErrScope
		}
		size := 0
		for _, content := range q.Results[0].Content {
			if content.Type != "inputText" {
				c.mu.Unlock()
				return seg, ErrScope
			}
			size += len(content.Text)
		}
		if size > e.cfg.MaxOutputBytes {
			c.mu.Unlock()
			return seg, ErrLimit
		}
	} else if len(q.Results) != 0 || len(q.Input) == 0 {
		c.mu.Unlock()
		return seg, ErrScope
	}
	c.mu.Unlock()
	// Any failure after an operation starts is uncertain. Persist a barrier and
	// never retry an App Server RPC or a tool response automatically.
	defer func() {
		if err != nil {
			c.mu.Lock()
			c.phase = "failed"
			_ = e.save(c, "failed")
			e.stop(c)
			c.mu.Unlock()
			interruptCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = e.interrupt(interruptCtx, c)
		}
	}()
	if phase == "new" {
		registry, makeErr := codextools.New(q.Owner, q.Tools)
		if makeErr != nil {
			return seg, makeErr
		}
		c.mu.Lock()
		c.registry = registry
		c.prefix = q.PrefixDigest
		err = e.save(c, "starting")
		c.mu.Unlock()
		if err != nil {
			return seg, err
		}
		var started struct {
			Thread struct {
				ID string `json:"id"`
			} `json:"thread"`
		}
		err = e.rpc.Call(ctx, "thread/start", map[string]any{"model": q.Model, "cwd": q.CWD, "dynamicTools": registry.NativeTools(), "approvalPolicy": "never", "sandbox": "read-only", "environments": []any{}}, &started)
		if err != nil {
			return seg, err
		}
		if started.Thread.ID == "" || len(started.Thread.ID) > 256 {
			return seg, ErrProtocol
		}
		owner, bindErr := registry.BindThread(started.Thread.ID)
		if bindErr != nil {
			return seg, bindErr
		}
		c.mu.Lock()
		c.owner = owner
		c.mu.Unlock()
		e.mu.Lock()
		if e.threads[owner.ThreadID] != nil {
			e.mu.Unlock()
			return seg, ErrProtocol
		}
		e.threads[owner.ThreadID] = c
		e.mu.Unlock()
	}
	if phase == "waiting" {
		c.mu.Lock()
		if c.phase != "waiting" {
			c.mu.Unlock()
			return seg, ErrRecovery
		}
		p := c.pending
		if len(q.References) > 0 {
			err = c.registry.Discover(c.owner, q.References, q.Tools)
		}
		if err == nil {
			err = c.registry.Complete(c.owner, c.turn, p.callID, q.Tools)
		}
		if err == nil {
			c.prefix = q.PrefixDigest
			err = e.save(c, "responding")
		}
		if err != nil {
			c.mu.Unlock()
			return seg, err
		}
		c.phase = "responding"
		c.mu.Unlock()
		err = e.rpc.Respond(ctx, p.rpcID, map[string]any{"contentItems": q.Results[0].Content, "success": q.Results[0].Success})
		if err != nil {
			return seg, err
		}
		c.mu.Lock()
		c.pending = nil
		c.phase = "active"
		err = e.save(c, "active")
		c.mu.Unlock()
		if err != nil {
			return seg, err
		}
	} else {
		c.mu.Lock()
		if c.phase == "canceled" {
			c.mu.Unlock()
			return seg, ErrRecovery
		}
		c.prefix = q.PrefixDigest
		c.model = q.Model
		c.phase = "active"
		c.turn = ""
		err = e.save(c, "active")
		thread := c.owner.ThreadID
		attach := c.attach
		c.mu.Unlock()
		if err != nil {
			return seg, err
		}
		if attach {
			// Reattach to the owned thread before the first turn of a resumed
			// conversation. Resume history and raw reasoning are never injected.
			var resumed struct {
				Thread struct {
					ID string `json:"id"`
				} `json:"thread"`
			}
			if err = e.rpc.Call(ctx, "thread/resume", map[string]any{"threadId": thread}, &resumed); err == nil && resumed.Thread.ID != thread {
				err = ErrProtocol
			}
			if err != nil {
				return seg, err
			}
			c.mu.Lock()
			c.attach = false
			c.mu.Unlock()
		}
		err = e.startTurn(ctx, c, map[string]any{"threadId": thread, "model": q.Model, "input": q.Input, "environments": []any{}})
		if err != nil {
			return seg, err
		}
	}
	for {
		select {
		case <-ctx.Done():
			return seg, ctx.Err()
		case <-e.ctx.Done():
			return seg, e.failure()
		case <-c.stopped:
			return seg, ErrRecovery
		case event := <-c.queue:
			c.mu.Lock()
			if c.phase == "canceled" {
				c.mu.Unlock()
				return seg, ErrRecovery
			}
			var p struct {
				ThreadID  string                      `json:"threadId"`
				TurnID    string                      `json:"turnId"`
				CallID    string                      `json:"callId"`
				Tool      string                      `json:"tool"`
				Arguments json.RawMessage             `json:"arguments"`
				Delta     string                      `json:"delta"`
				Turn      struct{ ID, Status string } `json:"turn"`
			}
			decodeErr := json.Unmarshal(event.Params, &p)
			if decodeErr != nil || p.ThreadID != c.owner.ThreadID || (p.TurnID != "" && p.TurnID != c.turn) {
				c.mu.Unlock()
				return seg, ErrProtocol
			}
			switch event.Method {
			case "item/tool/call":
				if len(event.ID) == 0 || p.TurnID != c.turn {
					c.mu.Unlock()
					return seg, ErrProtocol
				}
				invocation, beginErr := c.registry.Begin(c.owner, codextools.Call{TurnID: c.turn, CallID: p.CallID, Tool: p.Tool, Arguments: p.Arguments}, q.Tools)
				if beginErr != nil {
					c.mu.Unlock()
					return seg, beginErr
				}
				nonce := make([]byte, 24)
				if _, err = rand.Read(nonce); err != nil {
					c.mu.Unlock()
					return seg, err
				}
				publicID := "toolu_moai_" + hex.EncodeToString(nonce)
				c.pending = &pending{rpcID: append(json.RawMessage(nil), event.ID...), callID: p.CallID, publicID: publicID}
				c.phase = "waiting"
				err = e.save(c, "waiting")
				if err == nil {
					seg.Tool = &Tool{ID: publicID, Name: invocation.Name, Arguments: invocation.Arguments}
				}
				c.mu.Unlock()
				return seg, err
			case "item/agentMessage/delta":
				if len(seg.Text)+len(p.Delta) > e.cfg.MaxOutputBytes {
					c.mu.Unlock()
					return seg, ErrLimit
				}
				seg.Text += p.Delta
			case "turn/completed":
				if p.Turn.ID != c.turn || p.Turn.Status != "completed" {
					c.mu.Unlock()
					return seg, ErrProtocol
				}
				c.phase = "idle"
				err = e.save(c, "idle")
				seg.Done = err == nil
				c.mu.Unlock()
				return seg, err
			case "error":
				c.mu.Unlock()
				return seg, ErrProtocol
			default:
				if len(event.ID) > 0 {
					c.mu.Unlock()
					return seg, ErrProtocol
				}
			}
			c.mu.Unlock()
		}
	}
}
func (e *Engine) interrupt(ctx context.Context, c *conversation) error {
	c.mu.Lock()
	thread, turn := c.owner.ThreadID, c.turn
	if turn != "" && c.interruptedTurn == turn {
		c.mu.Unlock()
		return nil
	}
	if turn != "" {
		c.interruptedTurn = turn
	}
	c.mu.Unlock()
	if thread == "" || turn == "" {
		return nil
	}
	return e.rpc.Call(ctx, "turn/interrupt", map[string]any{"threadId": thread, "turnId": turn}, nil)
}
func (e *Engine) Cancel(ctx context.Context, owner codextools.Binding) error {
	if !validOwner(owner) {
		return ErrScope
	}
	e.mu.Lock()
	c := e.conversations[storeKey(owner)]
	e.mu.Unlock()
	if c == nil {
		return ErrScope
	}
	c.mu.Lock()
	c.phase = "canceled"
	err := e.save(c, "canceled")
	e.stop(c)
	c.mu.Unlock()
	if err != nil {
		return err
	}
	return e.interrupt(ctx, c)
}

// startTurn retains RPC correlation after the HTTP caller departs. Each worker
// has a process-owned deadline and can interrupt only its allocated turn.
func (e *Engine) startTurn(ctx context.Context, c *conversation, params map[string]any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	result := make(chan error, 1)
	e.mu.Lock()
	if e.ctx.Err() != nil {
		e.mu.Unlock()
		return ErrRecovery
	}
	e.workers.Add(1)
	e.mu.Unlock()
	go func() {
		defer e.workers.Done()
		startCtx, cancel := context.WithTimeout(e.ctx, 5*time.Second)
		defer cancel()
		var started struct {
			Turn struct {
				ID string `json:"id"`
			} `json:"turn"`
		}
		err := e.rpc.Call(startCtx, "turn/start", params, &started)
		if err == nil && (started.Turn.ID == "" || len(started.Turn.ID) > 256) {
			err = ErrProtocol
		}
		if err == nil {
			c.mu.Lock()
			c.turn = started.Turn.ID
			abandoned := ctx.Err() != nil || c.phase == "failed" || c.phase == "canceled"
			if !abandoned {
				err = e.save(c, "active")
			}
			c.mu.Unlock()
			if abandoned {
				cleanup, stop := context.WithTimeout(context.Background(), time.Second)
				_ = e.interrupt(cleanup, c)
				stop()
				if ctx.Err() != nil {
					err = ctx.Err()
				} else {
					err = ErrRecovery
				}
			}
		}
		result <- err
	}()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-c.stopped:
		return ErrRecovery
	case <-e.ctx.Done():
		return e.failure()
	}
}

// lateStarted closes the timeout gap: a canceled thread may announce allocation
// after its start RPC expired. One cleanup worker per conversation is permitted.
func (e *Engine) lateStarted(c *conversation, event codexapp.Message) {
	c.mu.Lock()
	id := startedTurnID(event, c.owner.ThreadID)
	if id == "" || (c.phase != "failed" && c.phase != "canceled") || c.lateCleanupScheduled || (c.turn != "" && c.turn != id) || c.interruptedTurn == id {
		c.mu.Unlock()
		return
	}
	c.turn = id
	c.lateCleanupScheduled = true
	c.mu.Unlock()
	e.mu.Lock()
	if e.ctx.Err() != nil {
		e.mu.Unlock()
		return
	}
	e.workers.Add(1)
	e.mu.Unlock()
	go func() {
		defer e.workers.Done()
		ctx, cancel := context.WithTimeout(e.ctx, time.Second)
		defer cancel()
		_ = e.interrupt(ctx, c)
	}()
}

func startedTurnID(event codexapp.Message, thread string) string {
	if event.Method != "turn/started" || len(event.ID) != 0 {
		return ""
	}
	var params struct {
		ThreadID string `json:"threadId"`
		Turn     struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if json.Unmarshal(event.Params, &params) != nil || params.ThreadID != thread || params.Turn.ID == "" || len(params.Turn.ID) > 256 {
		return ""
	}
	return params.Turn.ID
}
