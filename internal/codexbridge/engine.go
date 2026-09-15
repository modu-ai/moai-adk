// Package codexbridge keeps an official App Server turn alive across Claude
// HTTP tool segments. Callers supply authenticated scope and tool snapshots.
package codexbridge

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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
	Store *FileStore
	// QueueSize is a deprecated compatibility field, no longer an event-count
	// limit. Mailboxes are bounded by MaxOutputBytes (minimum 64 KiB).
	QueueSize                        int
	MaxConversations, MaxOutputBytes int
	// ForkPrefix contrasts a fork child's claimed inherited prefix with the
	// durable receipt ledger before the child conversation is created
	// (AC-MG-026 (c)). The gateway wires the receipt-backed authority here;
	// nil keeps the caller-asserted acceptance so bridge-only tests stay
	// self-contained.
	ForkPrefix func(claimed string) error
}
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}
type ToolResult struct {
	ID      string
	Content []Content
	Success bool
}
type Request struct {
	Owner                            codextools.Binding
	Model, Effort, CWD, Instructions string
	ExpectedPrefix, PrefixDigest     string
	Tools                            []codextools.Definition
	References                       []codextools.Reference
	Input                            []any
	Results                          []ToolResult
	Resume                           bool
	OutputSchema                     json.RawMessage
	// Ephemeral utility owners release in-memory capacity after durable completion.
	Ephemeral bool
	// Set only for a classified Claude hook-agent owner. StructuredOutput is a
	// terminal decision; Claude does not send its tool_result back to the bridge.
	HookAgentTerminalTool bool
	// Fork starts a brand-new App Server thread that inherits exactly the
	// completed prefix named by ExpectedPrefix — the explicit session fork
	// boundary. Never combined with Resume.
	Fork bool
}
type Tool struct {
	ID, Name  string
	Arguments json.RawMessage
}
type Segment struct {
	// nil means upstream usage was not observed, never a measured zero.
	Usage *Usage
	// ContextUsage is the latest generation's input occupancy, not segment billing.
	ContextUsage *Usage
	Text         string
	Tool         *Tool
	Tools        []Tool
	Done         bool
}
type pending struct {
	rpcID            json.RawMessage
	callID, publicID string
}
type conversation struct {
	lateCleanupScheduled            bool
	recoverableCancel               bool
	interruptedTurn                 string
	attach                          bool
	op                              chan struct{}
	mu                              sync.Mutex
	owner                           codextools.Binding
	registry                        *codextools.Registry
	model, cwd, prefix, phase, turn string
	instructionDigest               string
	pending                         []*pending
	deferred                        *codexapp.Message
	usageTotal                      *Usage
	usageTurn                       string
	queue                           *eventQueue
	failure                         error
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

func (c *conversation) stopError() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failure != nil {
		return c.failure
	}
	return ErrRecovery
}

// interruptOverflow leaves the shared reader available for RPC replies and
// other owners. No tool is replayed; cleanup has a bounded process-owned life.
func (e *Engine) interruptOverflow(c *conversation) {
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

// @MX:WARN: [AUTO] One reader owns event demultiplexing for all conversations.
// @MX:REASON: Bounded mailboxes isolate a slow owner without blocking shared RPC replies.
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
				TurnID   string `json:"turnId"`
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
				e.recoverCanceled(c, event)
				e.lateStarted(c, event)
				if len(event.ID) > 0 {
					_ = e.rpc.DiscardRequest(event.ID)
				}
				continue
			default:
			}
			if informationalEvent(event) && (envelope.TurnID == "" || envelope.TurnID == c.turn) {
				c.mu.Unlock()
				continue
			}
			if !c.queue.push(event) {
				c.failure = eventQueueLimitError{}
				c.phase = "failed"
				_ = e.save(c, "failed")
				e.stop(c)
				if len(event.ID) > 0 {
					_ = e.rpc.DiscardRequest(event.ID)
				}
				c.mu.Unlock()
				e.interruptOverflow(c)
				continue
			}
			c.mu.Unlock()
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
		if q.Fork {
			return nil, ErrScope
		}
		return e.resume(q, key)
	}
	inherited := ""
	if q.Fork {
		// A fork child inherits exactly the boundary prefix and nothing else;
		// the boundary is contrasted with the durable receipt ledger here
		// when the gateway wired a ForkPrefix authority (AC-MG-026 (c)).
		if len(q.Results) != 0 || q.ExpectedPrefix == "" || len(q.ExpectedPrefix) > 256 {
			return nil, ErrScope
		}
		if e.cfg.ForkPrefix != nil {
			if err := e.cfg.ForkPrefix(q.ExpectedPrefix); err != nil {
				return nil, fmt.Errorf("%w: fork inherited prefix rejected: %v", ErrScope, err)
			}
		}
		inherited = q.ExpectedPrefix
	} else if len(q.Results) != 0 || q.ExpectedPrefix != "" {
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
	c := &conversation{op: make(chan struct{}, 1), owner: q.Owner, model: q.Model, cwd: q.CWD, prefix: inherited, phase: "new", usageTotal: &Usage{}, queue: newEventQueue(e.cfg.MaxOutputBytes), stopped: make(chan struct{})}
	e.conversations[key] = c
	return c, nil
}
func (e *Engine) save(c *conversation, phase string) error {
	// The reader can fail an owner while its HTTP consumer is at a boundary.
	// A selected timer/event may not overwrite that durable terminal failure.
	if c.failure != nil && phase != "failed" {
		return c.failure
	}
	call := ""
	if len(c.pending) != 0 {
		call = c.pending[0].callID
	}
	return e.cfg.Store.save(record{Owner: c.owner, Phase: phase, TurnID: c.turn, CallID: call, Prefix: c.prefix, Model: c.model, CWD: c.cwd, InstructionDigest: c.instructionDigest})
}

func instructionDigest(instructions string) string {
	sum := sha256.Sum256([]byte(instructions))
	return hex.EncodeToString(sum[:])
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
	c := &conversation{op: make(chan struct{}, 1), owner: owner, registry: registry, model: q.Model, cwd: q.CWD, prefix: rec.Prefix, instructionDigest: rec.InstructionDigest, phase: "idle", attach: true, queue: newEventQueue(e.cfg.MaxOutputBytes), stopped: make(chan struct{})}
	e.conversations[key] = c
	e.threads[thread] = c
	return c, nil
}
func (e *Engine) stop(c *conversation) *codexapp.Message {
	var terminal *codexapp.Message
	if c.deferred != nil && c.deferred.Method == "turn/completed" {
		terminal = c.deferred
	}
	c.deferred = nil
	c.stopOnce.Do(func() { close(c.stopped) })
	for _, pending := range c.pending {
		_ = e.rpc.DiscardRequest(pending.rpcID)
	}
	c.pending = nil
	for {
		event, ok := c.queue.pop()
		if ok {
			if event.Method == "turn/completed" {
				copy := event
				terminal = &copy
			}
			if c.turn == "" {
				c.turn = startedTurnID(event, c.owner.ThreadID)
			}
			if len(event.ID) > 0 {
				c.recoverableCancel = false
				_ = e.rpc.DiscardRequest(event.ID)
			}
		} else {
			return terminal
		}
	}
}

// Step returns at a tool boundary without interrupting the running App Server
// turn. Prefixes and references must be derived from authenticated HTTP history,
// never from untrusted model text. Input contains only the new turn's delta.
func (e *Engine) Step(ctx context.Context, q Request) (seg Segment, err error) {
	return e.StepStream(ctx, q, nil)
}

// StepStream publishes validated text deltas while the turn runs. A callback
// failure cancels the operation; callers may emit a successful terminal only
// after this method returns a durable tool barrier or completed segment.
func (e *Engine) StepStream(ctx context.Context, q Request, onText func(string) error) (seg Segment, err error) {
	if q.HookAgentTerminalTool && (!q.Ephemeral || !strings.Contains(q.Owner.ConversationID, ":hook-agent:")) {
		return seg, ErrScope
	}
	if err := ctx.Err(); err != nil {
		return seg, err
	}
	c, err := e.get(q)
	if err != nil {
		return seg, err
	}
	select {
	case c.op <- struct{}{}:
		defer func() { <-c.op }()
	case <-ctx.Done():
		return seg, ctx.Err()
	case <-e.ctx.Done():
		return seg, e.failure()
	}
	if err := ctx.Err(); err != nil {
		return seg, err
	}
	c.mu.Lock()
	if c.failure != nil {
		err := c.failure
		c.mu.Unlock()
		return seg, err
	}
	if c.phase == "failed" || c.phase == "canceled" || c.phase == "retired" {
		c.mu.Unlock()
		return seg, ErrRecovery
	}
	phase := c.phase
	// AS4: only an idle barrier may switch the model — the new value repins the
	// conversation below and rides the next turn/start. A waiting or active
	// turn keeps its pinned model so a switch can never land on the wrong turn.
	if phase != "idle" && c.model != q.Model {
		c.mu.Unlock()
		return seg, fmt.Errorf("%w: active model mismatch", ErrScope)
	}
	if c.cwd != q.CWD {
		c.mu.Unlock()
		return seg, fmt.Errorf("%w: working directory mismatch", ErrScope)
	}
	if c.prefix != q.ExpectedPrefix {
		c.mu.Unlock()
		return seg, fmt.Errorf("%w: history prefix mismatch", ErrScope)
	}
	if phase == "waiting" {
		if len(q.Input) != 0 {
			c.mu.Unlock()
			return seg, fmt.Errorf("%w: new input arrived while tools are pending", ErrScope)
		}
		if len(q.Results) != len(c.pending) || len(c.pending) == 0 {
			c.mu.Unlock()
			return seg, fmt.Errorf("%w: tool result count mismatch", ErrScope)
		}
		pendingIDs := make(map[string]bool, len(c.pending))
		for _, pending := range c.pending {
			pendingIDs[pending.publicID] = true
		}
		size := 0
		seen := make(map[string]bool, len(q.Results))
		for _, result := range q.Results {
			if !pendingIDs[result.ID] || seen[result.ID] {
				c.mu.Unlock()
				return seg, fmt.Errorf("%w: unknown or duplicate tool result", ErrScope)
			}
			seen[result.ID] = true
			for _, content := range result.Content {
				if content.Type != "inputText" && content.Type != "inputImage" {
					c.mu.Unlock()
					return seg, fmt.Errorf("%w: unsupported tool result content", ErrScope)
				}
				if content.Type == "inputImage" && content.ImageURL == "" {
					c.mu.Unlock()
					return seg, fmt.Errorf("%w: empty tool result image", ErrScope)
				}
				size += len(content.Text) + len(content.ImageURL)
			}
		}
		if size > e.cfg.MaxOutputBytes {
			c.mu.Unlock()
			return seg, ErrLimit
		}
	} else if len(q.Results) != 0 {
		c.mu.Unlock()
		return seg, fmt.Errorf("%w: unexpected tool result", ErrScope)
	} else if len(q.Input) == 0 {
		c.mu.Unlock()
		return seg, fmt.Errorf("%w: empty turn input", ErrScope)
	}
	c.mu.Unlock()
	// A new owner's local preflight cannot have allocated a remote thread.
	// Once thread/start is entered, even a returned error may hide allocation;
	// persist a barrier and never retry that RPC automatically.
	remoteMayHaveStarted := phase != "new"
	defer func() {
		if err != nil && remoteMayHaveStarted {
			c.mu.Lock()
			// Recovery requires both an unambiguous allocated turn and a
			// terminal acknowledgement. Unanswered tools never qualify.
			c.recoverableCancel = errors.Is(err, context.Canceled) && c.phase == "active" && c.turn != "" && len(c.pending) == 0
			c.phase = "failed"
			_ = e.save(c, "failed")
			terminal := e.stop(c)
			c.mu.Unlock()
			if terminal != nil {
				e.recoverCanceled(c, *terminal)
			}
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
		oldPrefix, oldInstructions := c.prefix, c.instructionDigest
		c.instructionDigest = instructionDigest(q.Instructions)
		c.prefix = q.PrefixDigest
		err = e.save(c, "starting")
		if err != nil {
			c.prefix, c.instructionDigest = oldPrefix, oldInstructions
		} else {
			c.registry = registry
		}
		c.mu.Unlock()
		if err != nil {
			return seg, err
		}
		var started struct {
			Thread struct {
				ID string `json:"id"`
			} `json:"thread"`
		}
		start := map[string]any{"model": q.Model, "cwd": q.CWD, "dynamicTools": registry.NativeTools(), "approvalPolicy": "never", "sandbox": "workspace-write", "environments": []any{}}
		// Apply per thread too: a shared owner may predate the process-level
		// defaults, and model metadata can override feature flags alone.
		start["config"] = map[string]any{"agents.enabled": false, "features.multi_agent": false, "features.multi_agent_v2": false}
		remoteMayHaveStarted = true
		err = e.rpc.Call(ctx, "thread/start", start, &started)
		if err != nil {
			var rejected *codexapp.RPCError
			if errors.As(err, &rejected) {
				// An explicit negative acknowledgement precedes any turn/start.
				// Allow a caller's new request, not an automatic RPC retry. Keep
				// the durable starting barrier conservative across process loss.
				remoteMayHaveStarted = false
				c.mu.Lock()
				c.prefix, c.instructionDigest = oldPrefix, oldInstructions
				c.registry = nil
				rollbackErr := e.save(c, "starting")
				c.mu.Unlock()
				if rollbackErr != nil {
					remoteMayHaveStarted = true
					return seg, rollbackErr
				}
			}
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
		pending := append([]*pending(nil), c.pending...)
		results := make(map[string]ToolResult, len(q.Results))
		for _, result := range q.Results {
			results[result.ID] = result
		}
		if len(q.References) > 0 {
			err = c.registry.Discover(c.owner, q.References, q.Tools)
		}
		for _, p := range pending {
			if err == nil {
				err = c.registry.Complete(c.owner, c.turn, p.callID, q.Tools)
			}
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
		for _, p := range pending {
			result := results[p.publicID]
			err = e.rpc.Respond(ctx, p.rpcID, map[string]any{"contentItems": result.Content, "success": result.Success})
			if err != nil {
				return seg, err
			}
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
		if c.failure != nil {
			err := c.failure
			c.mu.Unlock()
			return seg, err
		}
		if c.phase == "canceled" {
			c.mu.Unlock()
			return seg, ErrRecovery
		}
		c.prefix = q.PrefixDigest
		c.model = q.Model
		c.instructionDigest = instructionDigest(q.Instructions)
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
			// This protects unloaded resumes. App Server may ignore overrides for
			// an already loaded thread; it is not a migration of an old live thread.
			params := map[string]any{"threadId": thread, "sandbox": "workspace-write", "approvalPolicy": "never", "config": map[string]any{"agents.enabled": false, "features.multi_agent": false, "features.multi_agent_v2": false}}
			if err = e.rpc.Call(ctx, "thread/resume", params, &resumed); err == nil && resumed.Thread.ID != thread {
				err = ErrProtocol
			}
			if err != nil {
				return seg, err
			}
			c.mu.Lock()
			c.attach = false
			c.instructionDigest = instructionDigest(q.Instructions)
			err = e.save(c, "active")
			c.mu.Unlock()
			if err != nil {
				return seg, err
			}
		}
		turn := map[string]any{"threadId": thread, "model": q.Model, "input": q.Input, "environments": []any{}}
		// Reapply on turns because loaded-thread resume can retain its old
		// read-only policy. External Claude tools still enforce their permissions.
		turn["sandboxPolicy"] = map[string]any{"type": "workspaceWrite", "networkAccess": false}
		// Loaded-thread resume ignores developerInstructions, even while idle.
		// A turn override updates the native developer context without a second
		// RPC or replaying history. An explicit empty string clears instructions
		// rather than selecting Codex's built-in collaboration preset.
		settings := map[string]any{"model": q.Model, "developer_instructions": q.Instructions}
		if q.Effort != "" {
			settings["reasoning_effort"] = q.Effort
		}
		turn["collaborationMode"] = map[string]any{"mode": "default", "settings": settings}
		if q.Effort != "" {
			turn["effort"] = q.Effort
		}
		if len(q.OutputSchema) != 0 {
			turn["outputSchema"] = q.OutputSchema
		}
		err = e.startTurn(ctx, c, turn)
		if err != nil {
			return seg, err
		}
	}
	argumentFailures := 0
	terminalHook := false
	var terminalInterruptErr error
	rejectArguments := func(id json.RawMessage) error {
		argumentFailures++
		if argumentFailures > 3 {
			return fmt.Errorf("%w: tool argument correction limit exceeded", ErrProtocol)
		}
		return e.rpc.Respond(ctx, id, map[string]any{"success": false, "contentItems": []Content{{Type: "inputText", Text: "Tool arguments failed JSON schema validation. No tool was executed. Read the tool's declared schema, provide all required fields with the correct types, and retry with corrected arguments."}}})
	}
	for {
		event, nextErr := e.nextEvent(ctx, c)
		if nextErr != nil {
			if terminalInterruptErr != nil {
				return seg, errors.Join(ErrProtocol, terminalInterruptErr, nextErr)
			}
			return seg, nextErr
		}
		c.mu.Lock()
		if c.failure != nil {
			err := c.failure
			c.mu.Unlock()
			return seg, err
		}
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
		turnBound := event.Method != "thread/tokenUsage/updated"
		if decodeErr != nil || p.ThreadID != c.owner.ThreadID || (turnBound && p.TurnID != "" && p.TurnID != c.turn) {
			c.mu.Unlock()
			return seg, fmt.Errorf("%w: invalid %s event binding", ErrProtocol, event.Method)
		}
		switch event.Method {
		case "item/tool/call":
			if terminalHook {
				c.mu.Unlock()
				return seg, fmt.Errorf("%w: tool after terminal hook decision", ErrProtocol)
			}
			if len(event.ID) == 0 || p.TurnID != c.turn {
				c.mu.Unlock()
				return seg, ErrProtocol
			}
			invocation, beginErr := c.registry.Begin(c.owner, codextools.Call{TurnID: c.turn, CallID: p.CallID, Tool: p.Tool, Arguments: p.Arguments}, q.Tools)
			if beginErr != nil {
				c.mu.Unlock()
				if errors.Is(beginErr, codextools.ErrArguments) {
					if err := rejectArguments(event.ID); err != nil {
						return seg, err
					}
					continue
				}
				return seg, beginErr
			}
			nonce := make([]byte, 24)
			if _, err = rand.Read(nonce); err != nil {
				c.mu.Unlock()
				return seg, err
			}
			publicID := "toolu_moai_" + hex.EncodeToString(nonce)
			if q.HookAgentTerminalTool && invocation.Name == "StructuredOutput" {
				if len(c.pending) != 0 {
					c.mu.Unlock()
					return seg, ErrProtocol
				}
				if err = c.registry.Complete(c.owner, c.turn, p.CallID, q.Tools); err != nil {
					c.mu.Unlock()
					return seg, err
				}
				c.phase = "responding"
				err = e.save(c, "responding")
				c.mu.Unlock()
				if err != nil {
					return seg, err
				}
				tool := Tool{ID: publicID, Name: invocation.Name, Arguments: invocation.Arguments}
				seg.Tool = &tool
				seg.Tools = []Tool{tool}
				terminalHook = true
				terminalCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				ctx = terminalCtx
				if err = e.rpc.Respond(ctx, event.ID, map[string]any{"success": true, "contentItems": []Content{{Type: "inputText", Text: "Hook decision received. No further action is required."}}}); err != nil {
					return seg, err
				}
				terminalInterruptErr = e.interrupt(ctx, c)
				continue
			}
			c.pending = append(c.pending, &pending{rpcID: append(json.RawMessage(nil), event.ID...), callID: p.CallID, publicID: publicID})
			tool := Tool{ID: publicID, Name: invocation.Name, Arguments: invocation.Arguments}
			seg.Tools = append(seg.Tools, tool)
			if seg.Tool == nil {
				first := tool
				seg.Tool = &first
			}
			c.mu.Unlock()

			// App Server emits a turn's dynamic-tool requests as one ordered
			// burst. A short quiet boundary collects that burst so Claude gets
			// one parallel tool_use batch and can return results in any order.
			quiet := time.NewTimer(5 * time.Millisecond)
			for {
				select {
				case <-c.queue.ready:
					extra, ok := c.queue.pop()
					if !ok {
						continue
					}
					if !quiet.Stop() {
						<-quiet.C
					}
					quiet.Reset(5 * time.Millisecond)
					if extra.Method != "item/tool/call" {
						var binding struct {
							ThreadID string `json:"threadId"`
							TurnID   string `json:"turnId"`
						}
						if len(extra.ID) > 0 || json.Unmarshal(extra.Params, &binding) != nil || binding.ThreadID != c.owner.ThreadID || (extra.Method != "thread/tokenUsage/updated" && binding.TurnID != "" && binding.TurnID != c.turn) {
							quiet.Stop()
							return seg, ErrProtocol
						}
						switch extra.Method {
						case "thread/tokenUsage/updated":
							c.mu.Lock()
							c.observeUsage(extra.Params, &seg)
							c.mu.Unlock()
							continue
						case "item/started", "item/completed":
							continue
						default:
							// Preserve content and terminal order for the next
							// segment rather than dropping a non-tool event.
							quiet.Stop()
							c.mu.Lock()
							c.deferred = &extra
							c.phase = "waiting"
							err = e.save(c, "waiting")
							c.mu.Unlock()
							return seg, err
						}
					}
					if len(extra.ID) == 0 {
						quiet.Stop()
						return seg, ErrProtocol
					}
					var next struct {
						ThreadID  string          `json:"threadId"`
						TurnID    string          `json:"turnId"`
						CallID    string          `json:"callId"`
						Tool      string          `json:"tool"`
						Arguments json.RawMessage `json:"arguments"`
					}
					if json.Unmarshal(extra.Params, &next) != nil || next.ThreadID != c.owner.ThreadID || next.TurnID != c.turn {
						quiet.Stop()
						return seg, ErrProtocol
					}
					if q.HookAgentTerminalTool && next.Tool == "StructuredOutput" {
						quiet.Stop()
						return seg, fmt.Errorf("%w: terminal hook decision mixed with pending tools", ErrProtocol)
					}
					c.mu.Lock()
					invocation, beginErr = c.registry.Begin(c.owner, codextools.Call{TurnID: c.turn, CallID: next.CallID, Tool: next.Tool, Arguments: next.Arguments}, q.Tools)
					if beginErr != nil {
						c.mu.Unlock()
						if errors.Is(beginErr, codextools.ErrArguments) {
							if err := rejectArguments(extra.ID); err != nil {
								quiet.Stop()
								return seg, err
							}
							continue
						}
						quiet.Stop()
						return seg, beginErr
					}
					if _, err = rand.Read(nonce); err != nil {
						c.mu.Unlock()
						quiet.Stop()
						return seg, err
					}
					publicID = "toolu_moai_" + hex.EncodeToString(nonce)
					c.pending = append(c.pending, &pending{rpcID: append(json.RawMessage(nil), extra.ID...), callID: next.CallID, publicID: publicID})
					seg.Tools = append(seg.Tools, Tool{ID: publicID, Name: invocation.Name, Arguments: invocation.Arguments})
					c.mu.Unlock()
				case <-quiet.C:
					c.mu.Lock()
					c.phase = "waiting"
					err = e.save(c, "waiting")
					c.mu.Unlock()
					return seg, err
				case <-ctx.Done():
					quiet.Stop()
					return seg, ctx.Err()
				case <-e.ctx.Done():
					quiet.Stop()
					return seg, e.failure()
				case <-c.stopped:
					quiet.Stop()
					return seg, c.stopError()
				}
			}
		case "thread/tokenUsage/updated":
			c.observeUsage(event.Params, &seg)
		case "item/agentMessage/delta":
			if terminalHook {
				c.mu.Unlock()
				continue
			}
			if len(seg.Text)+len(p.Delta) > e.cfg.MaxOutputBytes {
				c.mu.Unlock()
				return seg, ErrLimit
			}
			seg.Text += p.Delta
			c.mu.Unlock()
			if onText != nil {
				if err := onText(p.Delta); err != nil {
					return seg, err
				}
			}
			continue
		case "turn/completed":
			validTerminal := p.Turn.Status == "completed" || (terminalHook && p.Turn.Status == "interrupted")
			if p.Turn.ID != c.turn || !validTerminal {
				c.mu.Unlock()
				return seg, fmt.Errorf("%w: invalid turn/completed state", ErrProtocol)
			}
			c.phase = "idle"
			err = e.save(c, "idle")
			seg.Done = err == nil
			c.mu.Unlock()
			if seg.Done && q.Ephemeral {
				if err = e.cfg.Store.removeCompleted(c.owner); err != nil {
					seg.Done = false
					return seg, err
				}
				e.mu.Lock()
				c.mu.Lock()
				c.phase = "retired"
				delete(e.threads, c.owner.ThreadID)
				delete(e.conversations, storeKey(q.Owner))
				c.mu.Unlock()
				e.mu.Unlock()
			}
			return seg, err
		case "error":
			c.mu.Unlock()
			return seg, fmt.Errorf("%w: App Server error event", ErrProtocol)
		default:
			if len(event.ID) > 0 {
				c.mu.Unlock()
				return seg, fmt.Errorf("%w: unsupported App Server request %s", ErrProtocol, event.Method)
			}
		}
		c.mu.Unlock()
	}
}

func (e *Engine) nextEvent(ctx context.Context, c *conversation) (codexapp.Message, error) {
	c.mu.Lock()
	if c.failure != nil {
		err := c.failure
		c.mu.Unlock()
		return codexapp.Message{}, err
	}
	if c.deferred != nil {
		event := *c.deferred
		c.deferred = nil
		c.mu.Unlock()
		return event, nil
	}
	c.mu.Unlock()
	for {
		select {
		case <-ctx.Done():
			return codexapp.Message{}, ctx.Err()
		case <-e.ctx.Done():
			return codexapp.Message{}, e.failure()
		case <-c.stopped:
			return codexapp.Message{}, c.stopError()
		case <-c.queue.ready:
			if event, ok := c.queue.pop(); ok {
				return event, nil
			}
		}
	}
}

// recoverCanceled accepts only the terminal event for the exact canceled turn.
// It replaces the stopped in-memory owner without replaying input or tool
// responses. A crash before this durable idle barrier remains fail-closed.
func (e *Engine) recoverCanceled(c *conversation, event codexapp.Message) {
	if event.Method != "turn/completed" || len(event.ID) != 0 {
		return
	}
	var p struct {
		ThreadID string                      `json:"threadId"`
		Turn     struct{ ID, Status string } `json:"turn"`
	}
	if json.Unmarshal(event.Params, &p) != nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.recoverableCancel || p.ThreadID != c.owner.ThreadID || p.Turn.ID != c.turn || (p.Turn.Status != "interrupted" && p.Turn.Status != "completed") || e.ctx.Err() != nil {
		return
	}
	if err := e.save(c, "idle"); err != nil {
		return
	}
	next := &conversation{op: make(chan struct{}, 1), owner: c.owner, registry: c.registry, model: c.model, cwd: c.cwd, prefix: c.prefix, instructionDigest: c.instructionDigest, usageTotal: c.usageTotal, usageTurn: c.usageTurn, phase: "idle", queue: newEventQueue(e.cfg.MaxOutputBytes), stopped: make(chan struct{})}
	c.recoverableCancel = false
	e.conversations[storeKey(c.owner)] = next
	e.threads[c.owner.ThreadID] = next
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
