package translate

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"
)

type blockEvent struct {
	kind   string
	fields map[string]any
}

type streamBlock struct {
	pending           []blockEvent
	index             int
	text              strings.Builder
	deltaDone, closed bool
}
type streamItem struct {
	assigned             int
	base                 int
	id, kind, call, name string
	blocks               map[int]*streamBlock
	closed               bool
	final                map[string]any
}
type streamState struct {
	subscription  bool
	c             *ResponseContext
	w             io.Writer
	result        *map[string]any
	responseID    string
	publishOutput bool
	items         map[int]*streamItem
	ids, calls    map[string]bool
	nextBlock     int
	nextItem      int
}

// Stream owns and closes upstream on every exit, including cancellation. It
// forwards deltas once and emits success only after all final items agree with
// their deltas. The caller owns HTTP status handling before invoking Stream.
// @MX:WARN: [AUTO] Cancellation closes a potentially blocked upstream reader.
// @MX:REASON: context.AfterFunc is stopped on return; no unbounded reader goroutine.
func (c *ResponseContext) Stream(ctx context.Context, upstream io.ReadCloser, w io.Writer) error {
	return c.stream(ctx, upstream, w, false)
}

// SubscriptionStream accepts a missing or empty terminal output only after all
// indexed output_item.done records have passed the ordinary stream checks.
// It is reserved for the fixed subscription endpoint; API-key streams use Stream.
func (c *ResponseContext) SubscriptionStream(ctx context.Context, upstream io.ReadCloser, w io.Writer) error {
	return c.stream(ctx, upstream, w, true)
}

func (c *ResponseContext) stream(ctx context.Context, upstream io.ReadCloser, w io.Writer, subscription bool) error {
	return c.streamResult(ctx, upstream, w, subscription, nil)
}

// SubscriptionResponse collects a fixed subscription SSE response and returns
// the validated ordinary Messages JSON equivalent. It is used only when the
// client requested a non-stream response but the subscription endpoint accepts
// streaming requests exclusively.
func (c *ResponseContext) SubscriptionResponse(ctx context.Context, upstream io.ReadCloser) ([]byte, error) {
	var result map[string]any
	if err := c.streamResult(ctx, upstream, io.Discard, true, &result); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("subscription response missing")
	}
	return json.Marshal(result)
}

func (c *ResponseContext) streamResult(ctx context.Context, upstream io.ReadCloser, w io.Writer, subscription bool, result *map[string]any) error {
	if upstream == nil {
		return errors.New("upstream missing")
	}
	defer func() { _ = upstream.Close() }() // the context hook below already closes; a second close error carries no signal
	stop := context.AfterFunc(ctx, func() { _ = upstream.Close() })
	defer stop()
	if c == nil || w == nil {
		return errors.New("stream context or writer missing")
	}
	c.rawOutput = map[int][]byte{}
	s := &streamState{subscription: subscription, c: c, w: w, result: result, items: map[int]*streamItem{}, ids: map[string]bool{}, calls: map[string]bool{}}
	max := budget(c.limits.MaxEventBytes)
	scanner := bufio.NewScanner(upstream)
	remaining := budget(c.limits.MaxOutputBytes)
	// Count ScanLines' raw advance before it strips CRLF. Counting tokens loses
	// carriage returns; counting reader read-ahead would include unparsed events.
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		advance, token, err := bufio.ScanLines(data, atEOF)
		if advance > remaining || (advance == 0 && len(data) > remaining) {
			return 0, nil, errors.New("stream exceeds output byte limit")
		}
		remaining -= advance
		return advance, token, err
	})
	scanner.Buffer(make([]byte, min(max, 4096)), max)
	var data strings.Builder
	eventName := ""
	for scanner.Scan() {
		if e := ctx.Err(); e != nil {
			return e
		}
		line := scanner.Text()
		if subscription && !utf8.ValidString(line) {
			return errors.New("invalid UTF-8 SSE")
		}
		if line == "" {
			if data.Len() == 0 {
				eventName = ""
				continue
			}
			raw := strings.TrimSuffix(data.String(), "\n")
			event, e := objectJSON([]byte(raw))
			if e != nil {
				return e
			}
			typ, e := stringField(event, "type")
			if e != nil {
				return e
			}
			if eventName != "" && eventName != typ {
				return errors.New("SSE event type mismatch")
			}
			if typ == "response.output_item.done" {
				var wire struct {
					OutputIndex int             `json:"output_index"`
					Item        json.RawMessage `json:"item"`
				}
				if json.Unmarshal([]byte(raw), &wire) != nil {
					return errors.New("invalid raw output item")
				}
				c.rawOutput[wire.OutputIndex] = append([]byte(nil), wire.Item...)
			}
			done, e := s.event(typ, event)
			if e != nil {
				return e
			}
			if done {
				return ctx.Err()
			}
			data.Reset()
			eventName = ""
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, found := strings.Cut(line, ":")
		if !found {
			value = ""
		}
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "data":
			if data.Len()+len(value)+1 > max {
				return errors.New("SSE event exceeds limit")
			}
			data.WriteString(value)
			data.WriteByte('\n')
		case "event":
			if eventName != "" {
				return errors.New("duplicate SSE event field")
			}
			eventName = value
		case "id", "retry":
			return errors.New("SSE resume fields unsupported")
		default:
			return errors.New("unsupported SSE field")
		}
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := scanner.Err(); e != nil {
		return fmt.Errorf("upstream SSE read: %w", e)
	}
	return errors.New("upstream EOF before terminal event")
}

func (s *streamState) emit(typ string, m map[string]any) error {
	if s.c.limits.History != nil && !s.publishOutput {
		return nil
	}
	m["type"] = typ
	b, e := json.Marshal(m)
	if e != nil {
		return e
	}
	if _, e = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", typ, b); e != nil {
		return e
	}
	if f, ok := s.w.(interface{ Flush() }); ok {
		f.Flush()
	}
	return nil
}
func index(m map[string]any, key string) (int, error) {
	n, e := integer(m[key])
	if e != nil || n < 0 || n > 100000 {
		return 0, errors.New("invalid stream index")
	}
	return int(n), nil
}
func (s *streamState) find(m map[string]any) (*streamItem, error) {
	idx, e := index(m, "output_index")
	if e != nil {
		return nil, e
	}
	item, ok := s.items[idx]
	if !ok || item.closed {
		return nil, errors.New("missing or closed stream item")
	}
	id, e := stringField(m, "item_id")
	if e != nil || id != item.id {
		return nil, errors.New("stream item ID mismatch")
	}
	return item, nil
}
func (s *streamState) block(m map[string]any) (*streamItem, *streamBlock, error) {
	item, e := s.find(m)
	if e != nil {
		return nil, nil, e
	}
	if item.kind != "message" {
		return nil, nil, errors.New("text event on function call")
	}
	idx, e := index(m, "content_index")
	if e != nil {
		return nil, nil, e
	}
	b, ok := item.blocks[idx]
	if !ok || b.closed {
		return nil, nil, errors.New("missing or closed text block")
	}
	return item, b, nil
}
func (s *streamState) event(typ string, m map[string]any) (bool, error) {
	if typ == "response.created" {
		if s.responseID != "" {
			return false, errors.New("duplicate response start")
		}
		r, e := asObject(m["response"])
		if e != nil {
			return false, e
		}
		id, e := stringField(r, "id")
		if e != nil || r["model"] != s.c.model || r["status"] != "in_progress" {
			return false, errors.New("invalid response start")
		}
		if out, ok := r["output"]; ok {
			a, yes := out.([]any)
			if !yes || len(a) != 0 {
				return false, errors.New("response start already contains output")
			}
		}
		s.responseID = id
		return false, s.emit("message_start", map[string]any{"message": map[string]any{"id": id, "type": "message", "role": "assistant", "model": s.c.model, "content": []any{}, "stop_reason": nil, "stop_sequence": nil, "usage": map[string]any{"input_tokens": 0, "output_tokens": 0}}})
	}
	if s.responseID == "" {
		return false, errors.New("event before response start")
	}
	switch typ {
	case "response.in_progress":
		r, e := asObject(m["response"])
		if e != nil || r["id"] != s.responseID || r["status"] != "in_progress" {
			return false, errors.New("invalid progress event")
		}
		return false, nil
	case "response.output_item.added":
		return false, s.addItem(m)
	case "response.content_part.added":
		item, e := s.find(m)
		if e != nil {
			return false, e
		}
		idx, e := index(m, "content_index")
		if e != nil {
			return false, e
		}
		if item.kind != "message" || idx != len(item.blocks) {
			return false, errors.New("duplicate or out of order text part")
		}
		p, e := asObject(m["part"])
		if e != nil {
			return false, e
		}
		text, e := outputText(p)
		if e != nil || text != "" {
			return false, errors.New("text part must start empty")
		}
		b := &streamBlock{index: -1}
		item.blocks[idx] = b
		if e = s.emitBlock(b, "content_block_start", map[string]any{"content_block": map[string]any{"type": "text", "text": ""}}); e != nil {
			return false, e
		}
		return false, s.layout()
	case "response.output_text.delta", "response.output_text.done", "response.content_part.done":
		_, b, e := s.block(m)
		if e != nil {
			return false, e
		}
		switch typ {
		case "response.output_text.delta":
			if b.deltaDone {
				return false, errors.New("delta after text done")
			}
			text, e := stringFieldAllowEmpty(m, "delta")
			if e != nil {
				return false, e
			}
			b.text.WriteString(text)
			return false, s.emitBlock(b, "content_block_delta", map[string]any{"index": b.index, "delta": map[string]any{"type": "text_delta", "text": text}})
		case "response.output_text.done":
			text, e := stringFieldAllowEmpty(m, "text")
			if e != nil || b.deltaDone || text != b.text.String() {
				return false, errors.New("text done mismatch")
			}
			b.deltaDone = true
			return false, nil
		default:
			p, e := asObject(m["part"])
			if e != nil {
				return false, e
			}
			text, e := outputText(p)
			if e != nil || !b.deltaDone || text != b.text.String() {
				return false, errors.New("text part done mismatch")
			}
			b.closed = true
			return false, s.emitBlock(b, "content_block_stop", map[string]any{"index": b.index})
		}
	case "response.function_call_arguments.delta", "response.function_call_arguments.done":
		item, e := s.find(m)
		if e != nil {
			return false, e
		}
		if item.kind != "function_call" {
			return false, errors.New("arguments event on text")
		}
		b := item.blocks[0]
		if b.deltaDone {
			return false, errors.New("arguments already done")
		}
		if typ == "response.function_call_arguments.delta" {
			text, e := stringFieldAllowEmpty(m, "delta")
			if e != nil {
				return false, e
			}
			b.text.WriteString(text)
			return false, s.emitBlock(b, "content_block_delta", map[string]any{"index": b.index, "delta": map[string]any{"type": "input_json_delta", "partial_json": text}})
		}
		args, e := stringFieldAllowEmpty(m, "arguments")
		if e != nil || args != b.text.String() {
			return false, errors.New("arguments done mismatch")
		}
		if _, e = objectJSON([]byte(args)); e != nil {
			return false, errors.New("invalid function arguments")
		}
		b.deltaDone = true
		return false, nil
	case "response.reasoning_summary_part.added", "response.reasoning_summary_text.delta", "response.reasoning_summary_text.done", "response.reasoning_summary_part.done":
		item, e := s.find(m)
		if e != nil {
			return false, e
		}
		if item.kind != "reasoning" {
			return false, errors.New("summary on non-reasoning item")
		}
		return false, nil
	case "response.output_item.done":
		return false, s.doneItem(m)
	case "response.completed", "response.incomplete":
		r, e := asObject(m["response"])
		if e != nil {
			return false, e
		}
		if r["id"] != s.responseID || typ != "response."+fmt.Sprint(r["status"]) {
			return false, errors.New("response terminal mismatch")
		}
		value, present := r["output"]
		out, ok := value.([]any)
		if s.subscription && (!present || ok && len(out) == 0) {
			out = make([]any, len(s.items))
			for i := range out {
				item, exists := s.items[i]
				if !exists || !item.closed {
					return false, errors.New("terminal before item completion")
				}
				out[i] = item.final
			}
			r["output"] = out
			ok = true
		}
		if !ok || len(out) != len(s.items) {
			return false, errors.New("terminal output count mismatch")
		}
		for i, v := range out {
			item, ok := s.items[i]
			if !ok || !item.closed || !reflect.DeepEqual(item.final, v) {
				return false, errors.New("terminal output does not match closed items")
			}
		}
		converted, e := s.c.response(r)
		if e != nil {
			return false, e
		}
		if s.c.limits.History != nil {
			s.publishOutput = true
			if e = s.emit("message_start", map[string]any{"message": map[string]any{"id": s.responseID, "type": "message", "role": "assistant", "model": s.c.model, "content": []any{}, "stop_reason": nil, "stop_sequence": nil, "usage": map[string]any{"input_tokens": 0, "output_tokens": 0}}}); e != nil {
				return false, e
			}
			for i, value := range converted["content"].([]any) {
				block := value.(map[string]any)
				start := map[string]any{}
				for k, v := range block {
					start[k] = v
				}
				var delta map[string]any
				switch block["type"] {
				case "text":
					start["text"] = ""
					delta = map[string]any{"type": "text_delta", "text": block["text"]}
				case "tool_use":
					start["input"] = map[string]any{}
					raw, _ := json.Marshal(block["input"])
					delta = map[string]any{"type": "input_json_delta", "partial_json": string(raw)}
				}
				if e = s.emit("content_block_start", map[string]any{"index": i, "content_block": start}); e != nil {
					return false, e
				}
				if delta != nil {
					if e = s.emit("content_block_delta", map[string]any{"index": i, "delta": delta}); e != nil {
						return false, e
					}
				}
				if e = s.emit("content_block_stop", map[string]any{"index": i}); e != nil {
					return false, e
				}
			}
		}
		if s.result != nil {
			*s.result = converted
		}
		if e = s.emit("message_delta", map[string]any{"delta": map[string]any{"stop_reason": converted["stop_reason"], "stop_sequence": nil}, "usage": converted["usage"]}); e != nil {
			return false, e
		}
		return true, s.emit("message_stop", map[string]any{})
	default:
		return false, fmt.Errorf("unsupported or failed upstream event %q", typ)
	}
}
func (s *streamState) addItem(m map[string]any) error {
	idx, e := index(m, "output_index")
	if e != nil {
		return e
	}
	if idx != len(s.items) {
		return errors.New("duplicate or out of order output index")
	}
	raw, e := asObject(m["item"])
	if e != nil {
		return e
	}
	id, e := stringField(raw, "id")
	if e != nil || s.ids[id] {
		return errors.New("invalid or duplicate item ID")
	}
	kind, e := stringField(raw, "type")
	if e != nil {
		return e
	}
	if kind != "reasoning" && raw["status"] != "in_progress" {
		return errors.New("item must start in progress")
	}
	item := &streamItem{base: -1, id: id, kind: kind, blocks: map[int]*streamBlock{}}
	switch kind {
	case "reasoning":
		if s.c.limits.History == nil {
			return errors.New("opaque response authorization unavailable")
		}
		if e = keys(raw, "type", "id", "summary", "encrypted_content", "content", "status"); e != nil {
			return e
		}
		if status, ok := raw["status"]; ok && status != "in_progress" {
			return errors.New("invalid initial reasoning status")
		}
	case "message":
		if e = keys(raw, "id", "type", "role", "status", "content", "phase"); e != nil {
			return e
		}
		if phase, ok := raw["phase"]; ok && phase != nil {
			value, valid := phase.(string)
			if !valid || value != "commentary" && value != "final_answer" {
				return errors.New("invalid message phase")
			}
		}
		content, ok := raw["content"].([]any)
		if raw["role"] != "assistant" || !ok || len(content) != 0 {
			return errors.New("invalid initial message")
		}
	case "function_call":
		if e = keys(raw, "id", "type", "status", "call_id", "name", "arguments"); e != nil {
			return e
		}
		item.call, e = stringField(raw, "call_id")
		if e != nil || s.calls[item.call] {
			return errors.New("invalid or duplicate call ID")
		}
		item.name, e = stringField(raw, "name")
		if e != nil {
			return e
		}
		original, ok := s.c.names[item.name]
		if !ok || raw["arguments"] != "" {
			return errors.New("invalid initial tool call")
		}
		b := &streamBlock{index: -1}
		item.blocks[0] = b
		s.calls[item.call] = true
		if e = s.emitBlock(b, "content_block_start", map[string]any{"index": b.index, "content_block": map[string]any{"type": "tool_use", "id": item.call, "name": original, "input": map[string]any{}}}); e != nil {
			return e
		}
	default:
		return errors.New("unsupported output item; opaque reasoning is not enabled")
	}
	s.ids[id] = true
	s.items[idx] = item
	return s.layout()
}
func (s *streamState) doneItem(m map[string]any) error {
	idx, e := index(m, "output_index")
	if e != nil {
		return e
	}
	item, ok := s.items[idx]
	if !ok || item.closed {
		return errors.New("missing or already done item")
	}
	raw, e := asObject(m["item"])
	if e != nil {
		return e
	}
	if raw["id"] != item.id || raw["type"] != item.kind {
		return errors.New("item done identity mismatch")
	}
	if item.kind == "reasoning" {
		if _, e = outputEnvelope([]any{raw}); e != nil {
			return e
		}
		item.closed = true
		item.final = raw
		return s.layout()
	}
	if _, _, e = s.c.item(raw); e != nil {
		return e
	}
	if item.kind == "function_call" {
		b := item.blocks[0]
		if raw["call_id"] != item.call || raw["name"] != item.name || raw["arguments"] != b.text.String() || !b.deltaDone {
			return errors.New("tool item done mismatch")
		}
		b.closed = true
		if e = s.emitBlock(b, "content_block_stop", map[string]any{"index": b.index}); e != nil {
			return e
		}
	} else {
		parts := raw["content"].([]any)
		if len(parts) != len(item.blocks) {
			return errors.New("text item part count mismatch")
		}
		for i, p := range parts {
			b, ok := item.blocks[i]
			if !ok || !b.closed || p.(map[string]any)["text"] != b.text.String() {
				return errors.New("text item done mismatch")
			}
		}
	}
	item.closed = true
	item.final = raw
	return s.layout()
}

// emitBlock holds only output whose canonical offset is not yet knowable. The
// upstream byte budget also bounds this request-local queue; nothing is retried.
func (s *streamState) emitBlock(b *streamBlock, kind string, fields map[string]any) error {
	if b.index < 0 {
		b.pending = append(b.pending, blockEvent{kind, fields})
		return nil
	}
	fields["index"] = b.index
	return s.emit(kind, fields)
}

// layout advances across items with known widths. A function call has one block
// from its start; a message's final width is known only at output_item.done.
// The current message's parts can stream immediately at base+content_index.
func (s *streamState) layout() error {
	for s.nextItem < len(s.items) {
		item := s.items[s.nextItem]
		if item.base < 0 {
			item.base = s.nextBlock
		}
		for i := item.assigned; i < len(item.blocks); i++ {
			b := item.blocks[i]
			b.index = item.base + i
			for _, event := range b.pending {
				if err := s.emitBlock(b, event.kind, event.fields); err != nil {
					return err
				}
			}
			b.pending = nil
			item.assigned++
		}
		if item.kind != "function_call" && !item.closed {
			return nil
		}
		s.nextBlock = item.base + len(item.blocks)
		s.nextItem++
	}
	return nil
}
