package gateway

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
)

type AppServerAdapterConfig struct {
	Engine    *codexbridge.Engine
	Authority *AppServerAuthority
	// Prepare is a trusted owner boundary, called only after gateway session
	// authentication. AS4/AS5 must establish conversation identity, verify history
	// prefixes and join tool references to the current authenticated tool snapshot.
	// Never implement this by trusting a model-supplied conversation/thread ID.
	Prepare func(context.Context, RoutedRequest) (codexbridge.Request, error)
}
type AppServerAdapter struct{ cfg AppServerAdapterConfig }

func NewAppServerAdapter(cfg AppServerAdapterConfig) (*AppServerAdapter, error) {
	if cfg.Engine == nil || cfg.Authority == nil || cfg.Prepare == nil {
		return nil, ErrManagedAuthority
	}
	return &AppServerAdapter{cfg: cfg}, nil
}
func (a *AppServerAdapter) Send(ctx context.Context, r RoutedRequest) (response *http.Response, err error) {
	// A fresh value per Send prevents reuse or concurrent requests from sharing diagnostics.
	r.SummaryDiagnostic = &AppServerSummaryDiagnostic{}
	defer func() {
		if errors.Is(err, codexbridge.ErrScope) && r.SummaryDiagnostic.LastRole != "" {
			err = &appServerDiagnosticError{err: err, diagnostic: *r.SummaryDiagnostic}
		}
	}()
	if r.Entry.Provider != ProviderOpenAI || r.Entry.AuthMethod != AuthAppServer || r.Credential != nil {
		return nil, ErrManagedAuthority
	}
	if err := a.cfg.Authority.Check(ctx, r.Managed); err != nil {
		return nil, err
	}
	// The operator approved App Server output policy for both subscription and
	// API-key modes. Claude max_tokens is not an equivalent generation cap here;
	// the bridge still enforces its independent response-byte and cancel limits.
	q, err := a.cfg.Prepare(ctx, r)
	if err != nil {
		return nil, err
	}
	if q.Owner.AccountScope != r.Managed.Scope() || !appServerPreparedModelAllowed(r.Entry.UpstreamID, q) {
		return nil, ErrManagedAuthority
	}
	var options struct {
		Stream bool `json:"stream"`
	}
	if err = json.Unmarshal(r.Body, &options); err != nil {
		return nil, err
	}
	if err = a.cfg.Authority.Check(ctx, r.Managed); err != nil {
		return nil, err
	}
	if options.Stream {
		return appServerStream(ctx, r.Entry.RouteID, func(ctx context.Context, emit func(string) error) (codexbridge.Segment, error) {
			return a.cfg.Engine.StepStream(ctx, q, emit)
		})
	}
	segment, err := a.cfg.Engine.Step(ctx, q)
	if err != nil {
		return nil, err
	}
	return appServerResponse(r.Entry.RouteID, options.Stream, segment)
}

// Only trusted Prepare can select the model already owning a pending tool turn.
// Engine independently enforces that owner's model, prefix, and exact result IDs.
func appServerPreparedModelAllowed(selected string, q codexbridge.Request) bool {
	if q.Model == selected {
		return true
	}
	if len(q.Input) != 0 || len(q.Results) == 0 || q.ExpectedPrefix == "" || q.Resume || q.Ephemeral || q.Fork {
		return false
	}
	switch q.Model {
	case "gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-6-astra":
		return true
	default:
		return false
	}
}

type appServerStreamBody struct {
	*io.PipeReader
	cancel context.CancelFunc
}

func (b *appServerStreamBody) Close() error { b.cancel(); return b.PipeReader.Close() }

// Text is delivered immediately; tools and success terminals are emitted only
// after StepStream has persisted its barrier. Closing the HTTP body interrupts
// generation and unblocks both the producer and a pending pipe write.
// @MX:WARN: [AUTO] A stream producer lives until completion or body cancellation.
// @MX:REASON: Pipe closure and request cancellation bound blocked writes.
func appServerStream(ctx context.Context, model string, run func(context.Context, func(string) error) (codexbridge.Segment, error)) (*http.Response, error) {
	ctx, cancel := context.WithCancel(ctx)
	reader, writer := io.Pipe()
	ready := make(chan error, 1)
	var once sync.Once
	signal := func(err error) { once.Do(func() { ready <- err }) }
	stopClose := context.AfterFunc(ctx, func() { _ = reader.CloseWithError(ctx.Err()) })
	go func() {
		defer cancel()
		defer stopClose()
		// PipeWriter.Close always returns nil; stream errors are sent above it.
		defer func() { _ = writer.Close() }()
		emit := func(kind string, value any) error {
			raw, err := json.Marshal(value)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", kind, raw)
			return err
		}
		started, textOpen := false, false
		start := func() error {
			if started {
				return nil
			}
			nonce := make([]byte, 16)
			if _, err := rand.Read(nonce); err != nil {
				return err
			}
			started = true
			signal(nil)
			return emit("message_start", map[string]any{"type": "message_start", "message": map[string]any{"id": "msg_moai_" + hex.EncodeToString(nonce), "type": "message", "role": "assistant", "model": model, "content": []any{}, "stop_reason": nil, "stop_sequence": nil, "usage": map[string]int{"input_tokens": 0, "output_tokens": 0}}})
		}
		segment, err := run(ctx, func(delta string) error {
			if delta == "" {
				return nil
			}
			if err := start(); err != nil {
				return err
			}
			if !textOpen {
				textOpen = true
				if err := emit("content_block_start", map[string]any{"type": "content_block_start", "index": 0, "content_block": map[string]string{"type": "text", "text": ""}}); err != nil {
					return err
				}
			}
			return emit("content_block_delta", map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]string{"type": "text_delta", "text": delta}})
		})
		if err == nil && !segment.Done && segment.Tool == nil && len(segment.Tools) == 0 {
			err = codexbridge.ErrProtocol
		}
		if err != nil {
			if !started {
				signal(err)
				return
			}
			_, kind, cause := appServerError(err)
			_ = emit("error", map[string]any{"type": "error", "error": map[string]string{"type": kind, "message": cause}})
			return
		}
		if err = start(); err != nil {
			signal(err)
			return
		}
		index := 0
		if textOpen {
			if emit("content_block_stop", map[string]any{"type": "content_block_stop", "index": index}) != nil {
				return
			}
			index++
		}
		tools := segment.Tools
		if len(tools) == 0 && segment.Tool != nil {
			tools = []codexbridge.Tool{*segment.Tool}
		}
		for _, tool := range tools {
			if emit("content_block_start", map[string]any{"type": "content_block_start", "index": index, "content_block": map[string]any{"type": "tool_use", "id": tool.ID, "name": tool.Name, "input": map[string]any{}}}) != nil {
				return
			}
			if emit("content_block_delta", map[string]any{"type": "content_block_delta", "index": index, "delta": map[string]string{"type": "input_json_delta", "partial_json": string(tool.Arguments)}}) != nil {
				return
			}
			if emit("content_block_stop", map[string]any{"type": "content_block_stop", "index": index}) != nil {
				return
			}
			index++
		}
		reason := "end_turn"
		if len(tools) > 0 {
			reason = "tool_use"
		}
		if emit("message_delta", map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": reason, "stop_sequence": nil}, "usage": appServerSegmentUsage(segment)}) != nil {
			return
		}
		_ = emit("message_stop", map[string]string{"type": "message_stop"})
	}()
	if err := <-ready; err != nil {
		cancel()
		_ = reader.Close()
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: &appServerStreamBody{PipeReader: reader, cancel: cancel}}, nil
}

// appServerResponse handles non-streaming responses and completed-segment
// fixtures. Live streaming uses appServerStream to avoid buffering text.
func appServerResponse(model string, stream bool, segment codexbridge.Segment) (*http.Response, error) {
	if !segment.Done && segment.Tool == nil && len(segment.Tools) == 0 {
		return nil, codexbridge.ErrProtocol
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	id := "msg_moai_" + hex.EncodeToString(nonce)
	blocks := []any{}
	if segment.Text != "" {
		blocks = append(blocks, map[string]any{"type": "text", "text": segment.Text})
	}
	reason := "end_turn"
	tools := segment.Tools
	if len(tools) == 0 && segment.Tool != nil {
		tools = []codexbridge.Tool{*segment.Tool}
	}
	if len(tools) != 0 {
		reason = "tool_use"
		for _, tool := range tools {
			blocks = append(blocks, map[string]any{"type": "tool_use", "id": tool.ID, "name": tool.Name, "input": tool.Arguments})
		}
	}
	message := map[string]any{"id": id, "type": "message", "role": "assistant", "model": model, "content": blocks, "stop_reason": reason, "stop_sequence": nil, "usage": appServerSegmentUsage(segment)}
	var body bytes.Buffer
	contentType := "application/json"
	if !stream {
		if err := json.NewEncoder(&body).Encode(message); err != nil {
			return nil, err
		}
	} else {
		contentType = "text/event-stream"
		emit := func(event string, value any) error {
			raw, err := json.Marshal(value)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(&body, "event: %s\ndata: %s\n\n", event, raw)
			return err
		}
		start := map[string]any{"id": id, "type": "message", "role": "assistant", "model": model, "content": []any{}, "stop_reason": nil, "stop_sequence": nil, "usage": map[string]int{"input_tokens": 0, "output_tokens": 0}}
		if err := emit("message_start", map[string]any{"type": "message_start", "message": start}); err != nil {
			return nil, err
		}
		for i, block := range blocks {
			b := block.(map[string]any)
			var delta any
			if b["type"] == "text" {
				delta = map[string]any{"type": "text_delta", "text": b["text"]}
				b = map[string]any{"type": "text", "text": ""}
			} else {
				raw := b["input"].(json.RawMessage)
				delta = map[string]any{"type": "input_json_delta", "partial_json": string(raw)}
				b = map[string]any{"type": "tool_use", "id": b["id"], "name": b["name"], "input": map[string]any{}}
			}
			if err := emit("content_block_start", map[string]any{"type": "content_block_start", "index": i, "content_block": b}); err != nil {
				return nil, err
			}
			if err := emit("content_block_delta", map[string]any{"type": "content_block_delta", "index": i, "delta": delta}); err != nil {
				return nil, err
			}
			if err := emit("content_block_stop", map[string]any{"type": "content_block_stop", "index": i}); err != nil {
				return nil, err
			}
		}
		if err := emit("message_delta", map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": reason, "stop_sequence": nil}, "usage": appServerSegmentUsage(segment)}); err != nil {
			return nil, err
		}
		if err := emit("message_stop", map[string]string{"type": "message_stop"}); err != nil {
			return nil, err
		}
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(bytes.NewReader(body.Bytes()))}, nil
}

func appServerUsage(usage *codexbridge.Usage) map[string]int64 {
	if usage == nil {
		// Required protocol placeholders; nil remains explicitly unknown inside
		// the bridge. No tokenizer estimate or measured-zero claim is made.
		return map[string]int64{"input_tokens": 0, "output_tokens": 0}
	}
	return map[string]int64{"input_tokens": usage.InputTokens - usage.CachedInputTokens - usage.CacheWriteInputTokens, "cache_read_input_tokens": usage.CachedInputTokens, "cache_creation_input_tokens": usage.CacheWriteInputTokens, "output_tokens": usage.OutputTokens}
}

func appServerSegmentUsage(segment codexbridge.Segment) map[string]int64 {
	usage := appServerUsage(segment.Usage)
	if segment.ContextUsage != nil {
		context := appServerUsage(segment.ContextUsage)
		for _, key := range []string{"input_tokens", "cache_read_input_tokens", "cache_creation_input_tokens"} {
			usage[key] = context[key]
		}
	}
	return usage
}
