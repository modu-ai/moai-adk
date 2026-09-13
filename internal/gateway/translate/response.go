package translate

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
)

// Response converts one complete ordinary Responses JSON result. It never
// returns a partial successful response alongside a validation error.
func (c *ResponseContext) Response(body []byte) ([]byte, error) {
	if c == nil || len(body) > budget(c.limits.MaxOutputBytes) {
		return nil, errors.New("response size or context invalid")
	}
	r, e := objectJSON(body)
	if e != nil {
		return nil, e
	}
	var wire struct {
		Output []json.RawMessage `json:"output"`
	}
	if json.Unmarshal(body, &wire) != nil {
		return nil, errors.New("invalid response output")
	}
	c.rawOutput = map[int][]byte{}
	for i, raw := range wire.Output {
		c.rawOutput[i] = append([]byte(nil), raw...)
	}
	out, e := c.response(r)
	if e != nil {
		return nil, e
	}
	return json.Marshal(out)
}
func (c *ResponseContext) response(r map[string]any) (map[string]any, error) {
	id, e := stringField(r, "id")
	if e != nil {
		return nil, e
	}
	if model, e := stringField(r, "model"); e != nil || model != c.model {
		return nil, errors.New("response model mismatch")
	}
	items, ok := r["output"].([]any)
	if !ok || len(items) == 0 {
		return nil, errors.New("no public output")
	}
	envelope, e := outputEnvelopeWithRaw(items, c.rawOutput)
	if e != nil {
		return nil, e
	}
	if envelope != nil && c.limits.History == nil {
		return nil, errors.New("opaque response authorization unavailable")
	}
	blocks := []any{}
	if envelope != nil {
		blocks = append(blocks, map[string]any{"type": "redacted_thinking", "data": envelope.Data()})
	}
	ids := map[string]bool{}
	calls := map[string]bool{}
	hasTool := false
	for _, v := range items {
		item, e := asObject(v)
		if e != nil {
			return nil, e
		}
		id, e := stringField(item, "id")
		if e != nil || ids[id] {
			return nil, errors.New("invalid or duplicate output item ID")
		}
		ids[id] = true
		if item["type"] == "reasoning" {
			continue
		}
		bs, call, e := c.item(item)
		if e != nil {
			return nil, e
		}
		if call != "" {
			if calls[call] {
				return nil, errors.New("duplicate output call ID")
			}
			if envelope != nil {
				bound, err := opaque.BindToolID(call, envelope)
				if err != nil {
					return nil, err
				}
				bs[0].(map[string]any)["id"] = bound
			}
			calls[call] = true
			hasTool = true
		}
		blocks = append(blocks, bs...)
	}
	publicBlocks := blocks
	if envelope != nil {
		publicBlocks = blocks[1:]
	}
	if len(publicBlocks) == 0 {
		return nil, errors.New("no public output")
	}
	stop, e := stopReason(r, hasTool)
	if e != nil {
		return nil, e
	}
	if c.title {
		if e = validateTitle(publicBlocks, stop); e != nil {
			return nil, e
		}
	}
	usage, e := responseUsage(r)
	if e != nil {
		return nil, e
	}
	result := map[string]any{"id": id, "type": "message", "role": "assistant", "model": c.model, "content": blocks, "stop_reason": stop, "stop_sequence": nil, "usage": usage}
	if c.limits.History != nil {
		if stop == "max_tokens" {
			return nil, errors.New("incomplete response cannot publish history")
		}
		var history []any
		if json.Unmarshal(c.messages, &history) != nil {
			return nil, errors.New("request history missing")
		}
		history = append(history, map[string]any{"role": "assistant", "content": blocks})
		encoded, _ := json.Marshal(history)
		if err := c.limits.History.Publish(c.ctx, c.model, c.limits.CredentialScope, encoded); err != nil {
			return nil, err
		}
	}
	return result, nil
}
func (c *ResponseContext) item(item map[string]any) ([]any, string, error) {
	if item["status"] != "completed" {
		return nil, "", errors.New("unfinished output item")
	}
	typ, e := stringField(item, "type")
	if e != nil {
		return nil, "", e
	}
	switch typ {
	case "message":
		if e = keys(item, "id", "type", "role", "status", "content", "phase"); e != nil {
			return nil, "", e
		}
		if item["role"] != "assistant" {
			return nil, "", errors.New("invalid output role")
		}
		if phase, ok := item["phase"]; ok && phase != nil {
			value, valid := phase.(string)
			if !valid || value != "commentary" && value != "final_answer" {
				return nil, "", errors.New("invalid output phase")
			}
		}
		content, ok := item["content"].([]any)
		if !ok || len(content) == 0 {
			return nil, "", errors.New("empty output message")
		}
		blocks := []any{}
		for _, v := range content {
			p, e := asObject(v)
			if e != nil {
				return nil, "", e
			}
			s, e := outputText(p)
			if e != nil {
				return nil, "", e
			}
			blocks = append(blocks, map[string]any{"type": "text", "text": s})
		}
		return blocks, "", nil
	case "function_call":
		if e = keys(item, "id", "type", "status", "call_id", "name", "arguments"); e != nil {
			return nil, "", e
		}
		call, e := stringField(item, "call_id")
		if e != nil {
			return nil, "", e
		}
		name, e := stringField(item, "name")
		if e != nil {
			return nil, "", e
		}
		original, ok := c.names[name]
		if !ok {
			return nil, "", errors.New("unmapped tool name")
		}
		args, e := stringFieldAllowEmpty(item, "arguments")
		if e != nil {
			return nil, "", e
		}
		input, e := objectJSON([]byte(args))
		if e != nil {
			return nil, "", errors.New("invalid function arguments")
		}
		return []any{map[string]any{"type": "tool_use", "id": call, "name": original, "input": input}}, call, nil
	default:
		return nil, "", fmt.Errorf("unsupported output item %q; opaque reasoning is not enabled", typ)
	}
}
func outputText(p map[string]any) (string, error) {
	if e := keys(p, "type", "text", "annotations", "logprobs"); e != nil {
		return "", e
	}
	if p["type"] != "output_text" {
		return "", errors.New("unsupported public content")
	}
	for _, k := range []string{"annotations", "logprobs"} {
		if v, ok := p[k]; ok {
			a, yes := v.([]any)
			if !yes || len(a) != 0 {
				return "", fmt.Errorf("unsupported nonempty %s", k)
			}
		}
	}
	return stringFieldAllowEmpty(p, "text")
}
func stopReason(r map[string]any, hasTool bool) (string, error) {
	switch r["status"] {
	case "completed":
		if r["error"] != nil || r["incomplete_details"] != nil {
			return "", errors.New("conflicting completion metadata")
		}
		if hasTool {
			return "tool_use", nil
		}
		return "end_turn", nil
	case "incomplete":
		d, ok := r["incomplete_details"].(map[string]any)
		if ok && d["reason"] == "max_output_tokens" && r["error"] == nil {
			return "max_tokens", nil
		}
	}
	return "", errors.New("upstream did not complete successfully")
}
func responseUsage(r map[string]any) (map[string]any, error) {
	m, ok := r["usage"].(map[string]any)
	if !ok {
		return nil, errors.New("usage missing")
	}
	input, e := integer(m["input_tokens"])
	if e != nil || input < 0 {
		return nil, errors.New("invalid input usage")
	}
	output, e := integer(m["output_tokens"])
	if e != nil || output < 0 {
		return nil, errors.New("invalid output usage")
	}
	return map[string]any{"input_tokens": input, "output_tokens": output}, nil
}
