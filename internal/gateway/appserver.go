package gateway

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
func (a *AppServerAdapter) Send(ctx context.Context, r RoutedRequest) (*http.Response, error) {
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
	if q.Owner.AccountScope != r.Managed.Scope() || q.Model != r.Entry.UpstreamID {
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
	segment, err := a.cfg.Engine.Step(ctx, q)
	if err != nil {
		return nil, err
	}
	return appServerResponse(r.Entry.RouteID, options.Stream, segment)
}

// Segment buffering is deliberate: no successful SSE terminal is emitted until
// the tool barrier is durable or the actual turn/completed event is accepted.
// This does not claim incremental token delivery latency or measured token usage.
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
	message := map[string]any{"id": id, "type": "message", "role": "assistant", "model": model, "content": blocks, "stop_reason": reason, "stop_sequence": nil, "usage": map[string]int{"input_tokens": 0, "output_tokens": 0}}
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
		if err := emit("message_delta", map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": reason, "stop_sequence": nil}, "usage": map[string]int{"output_tokens": 0}}); err != nil {
			return nil, err
		}
		if err := emit("message_stop", map[string]string{"type": "message_stop"}); err != nil {
			return nil, err
		}
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(bytes.NewReader(body.Bytes()))}, nil
}
