// Package translate converts the explicitly supported public Messages subset to
// Responses. Opaque reasoning and unimplemented policies fail closed.
package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Limits supplies measured context usage; no token estimate or truncation is
// performed here. Positive ContextTokens requires a caller measurement.
type Limits struct {
	PolicyProfile                               PolicyProfile
	MaxBodyBytes, MaxEventBytes, MaxOutputBytes int
	ContextTokens                               int64
	InputTokens                                 *int64
	// NativeReceiptAuthorize is supplied only by the private gateway factory.
	// Native policy syntax never authorizes history by itself.
	NativeReceiptAuthorize func(context.Context, []byte, NativePolicy) error
	History                HistoryAuthority
	CredentialScope        string
}

// ResponseContext owns an immutable, request-local tool-name reverse map.
type ResponseContext struct {
	ctx       context.Context
	messages  []byte
	rawOutput map[int][]byte
	title     bool
	names     map[string]string
	model     string
	limits    Limits
}

var safeName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func budget(n int) int {
	if n > 0 {
		return n
	}
	return 8 << 20
}

// Request validates the complete input before returning bytes suitable for egress.
// @MX:ANCHOR: [AUTO] Public request translation boundary.
// @MX:REASON: Routing, streaming and non-streaming adapters share this boundary.
func Request(model string, body []byte, limits Limits) ([]byte, *ResponseContext, error) {
	return RequestContext(context.Background(), model, body, limits)
}

// RequestContext is Request with a caller-owned cancellation context for the
// private native receipt check.
func RequestContext(ctx context.Context, model string, body []byte, limits Limits) ([]byte, *ResponseContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if model == "" || len(body) > budget(limits.MaxBodyBytes) {
		return nil, nil, errors.New("invalid model or request size")
	}
	if limits.ContextTokens > 0 && (limits.InputTokens == nil || *limits.InputTokens < 0 || *limits.InputTokens > limits.ContextTokens) {
		return nil, nil, errors.New("context measurement missing or exceeds limit")
	}
	root, err := objectJSON(body)
	if err != nil {
		return nil, nil, err
	}
	allowed := []string{"model", "messages", "system", "tools", "metadata", "max_tokens", "stream", "temperature", "top_p", "stop_sequences", "tool_choice"}
	if limits.PolicyProfile != "" {
		allowed = append(allowed, "thinking", "output_config", "context_management")
	}
	if err = keys(root, allowed...); err != nil {
		return nil, nil, err
	}
	max, err := integer(root["max_tokens"])
	if err != nil || max <= 0 {
		return nil, nil, errors.New("max_tokens must be a positive integer")
	}
	out := map[string]any{"model": model, "max_output_tokens": max, "store": false}
	var policy NativePolicy
	if limits.PolicyProfile != "" {
		if limits.PolicyProfile != PolicyGPTNative {
			return nil, nil, errors.New("wrong Responses policy profile")
		}
		switch model {
		case "gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna":
		default:
			return nil, nil, errors.New("unverified policy model")
		}
		policy, err = nativePolicy(root, limits.PolicyProfile)
		if err != nil {
			return nil, nil, err
		}
		if policy.Effort != "" && !gptModelEffortAllowed(model, policy.Effort) {
			return nil, nil, errors.New("unsupported effort")
		}
		if policy.KeepAll || policy.UserID != "" {
			if limits.NativeReceiptAuthorize == nil || limits.NativeReceiptAuthorize(ctx, body, policy) != nil {
				return nil, nil, errors.New("native receipt authorization required")
			}
		}
		policy.applyGPT(out, root)
	}
	if v, ok := root["stream"]; ok {
		b, yes := v.(bool)
		if !yes {
			return nil, nil, errors.New("stream must be boolean")
		}
		out["stream"] = b
	}
	for _, k := range []string{"temperature", "top_p"} {
		if v, ok := root[k]; ok {
			if _, yes := v.(json.Number); !yes {
				return nil, nil, fmt.Errorf("invalid %s", k)
			}
			out[k] = v
		}
	}
	if _, ok := root["stop_sequences"]; ok {
		return nil, nil, errors.New("stop_sequences unsupported by Responses")
	}
	if v, ok := root["metadata"]; ok {
		m, yes := v.(map[string]any)
		if !yes {
			return nil, nil, errors.New("metadata must be object")
		}
		if _, nativeID := m["user_id"]; nativeID && policy.UserID == "" {
			return nil, nil, errors.New("native receipt authorization required")
		}
		for _, x := range m {
			if _, yes := x.(string); !yes {
				return nil, nil, errors.New("metadata values must be strings")
			}
		}
		if policy.UserID == "" {
			out["metadata"] = m
		}
	}
	if v, ok := root["system"]; ok {
		texts, e := textContent(v)
		if e != nil {
			return nil, nil, e
		}
		out["instructions"] = strings.Join(texts, "\n\n")
	}
	c := &ResponseContext{ctx: ctx, title: policy.Title, names: map[string]string{}, model: model, limits: limits}
	forward := map[string]string{}
	// Reserve all safe names before generating aliases, including historical tools.
	names, err := collectNames(root)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range names {
		if safeName.MatchString(n) {
			forward[n] = n
			c.names[n] = n
		}
	}
	for _, n := range names {
		if _, ok := forward[n]; ok {
			continue
		}
		for i := 0; ; i++ {
			alias := "moai_tool_" + strconv.Itoa(i)
			if _, used := c.names[alias]; !used {
				forward[n] = alias
				c.names[alias] = n
				break
			}
		}
	}
	toolDefinitions := map[string]any{}
	loadedTools := map[string]bool{}
	if v, ok := root["tools"]; ok {
		list, yes := v.([]any)
		if !yes {
			return nil, nil, errors.New("tools must be array")
		}
		tools := []any{}
		seen := map[string]bool{}
		for _, v := range list {
			m, e := asObject(v)
			if e != nil {
				return nil, nil, e
			}
			if e = keys(m, "name", "description", "input_schema", "cache_control", "defer_loading"); e != nil {
				return nil, nil, e
			}
			if e = cacheHint(m); e != nil {
				return nil, nil, e
			}
			deferLoading := false
			if deferred, ok := m["defer_loading"]; ok {
				var yes bool
				deferLoading, yes = deferred.(bool)
				if !yes {
					return nil, nil, errors.New("invalid defer loading setting")
				}
			}
			name, e := stringField(m, "name")
			if e != nil || seen[name] {
				return nil, nil, errors.New("invalid or duplicate tool name")
			}
			seen[name] = true
			schema, ok := m["input_schema"].(map[string]any)
			if !ok {
				return nil, nil, errors.New("tool schema must be object")
			}
			tool := map[string]any{"type": "function", "name": forward[name], "parameters": schema, "strict": false}
			if d, ok := m["description"]; ok {
				if _, yes := d.(string); !yes {
					return nil, nil, errors.New("tool description must be string")
				}
				tool["description"] = d
			}
			toolDefinitions[name] = tool
			// Deferred schemas become visible only after a validated tool reference.
			if !deferLoading {
				tools = append(tools, tool)
				loadedTools[name] = true
			}
		}
		out["tools"] = tools
	}
	if v, ok := root["tool_choice"]; ok {
		m, e := asObject(v)
		if e != nil {
			return nil, nil, e
		}
		if e = keys(m, "type", "name", "disable_parallel_tool_use"); e != nil {
			return nil, nil, e
		}
		typ, e := stringField(m, "type")
		if e != nil {
			return nil, nil, e
		}
		switch typ {
		case "auto", "none":
			out["tool_choice"] = typ
		case "any":
			out["tool_choice"] = "required"
		case "tool":
			name, e := stringField(m, "name")
			if e != nil || forward[name] == "" {
				return nil, nil, errors.New("unknown chosen tool")
			}
			out["tool_choice"] = map[string]any{"type": "function", "name": forward[name]}
		default:
			return nil, nil, errors.New("unsupported tool choice")
		}
		if v, ok := m["disable_parallel_tool_use"]; ok {
			b, yes := v.(bool)
			if !yes {
				return nil, nil, errors.New("invalid parallel tool setting")
			}
			out["parallel_tool_calls"] = !b
		}
	}
	messages, ok := root["messages"].([]any)
	if !ok || len(messages) == 0 {
		return nil, nil, errors.New("messages must be nonempty array")
	}
	c.messages, _ = json.Marshal(messages)
	if limits.History != nil {
		if err := limits.History.Check(ctx, model, limits.CredentialScope, c.messages); err != nil {
			return nil, nil, err
		}
	}
	input := []any{}
	rawReplay := []opaque.Item{}
	calls := map[string]bool{}
	answered := map[string]bool{}
	restoredIDs := map[string]string{}
	for _, v := range messages {
		m, e := asObject(v)
		if e != nil {
			return nil, nil, e
		}
		if e = keys(m, "role", "content"); e != nil {
			return nil, nil, e
		}
		role, e := stringField(m, "role")
		if e != nil || (role != "user" && role != "assistant" && role != "system") {
			return nil, nil, errors.New("unsupported message role")
		}
		blocks, e := contentBlocks(m["content"])
		if e != nil {
			return nil, nil, e
		}
		var envelope *opaque.Envelope
		if role == "assistant" {
			envelope, e = messageEnvelope(blocks)
			if e != nil {
				return nil, nil, e
			}
			if envelope != nil && limits.History == nil {
				return nil, nil, errors.New("opaque history authorization unavailable")
			}
		}
		startInput := len(input)
		pending := []any{}
		flush := func() {
			if len(pending) > 0 {
				input = append(input, map[string]any{"role": role, "content": pending})
				pending = []any{}
			}
		}
		for _, b := range blocks {
			typ, e := stringField(b, "type")
			if e != nil {
				return nil, nil, e
			}
			switch typ {
			case "redacted_thinking":
				if role != "assistant" || envelope == nil {
					return nil, nil, errors.New("untrusted reasoning block")
				}
				continue
			case "text":
				if e = keys(b, "type", "text", "cache_control"); e != nil {
					return nil, nil, e
				}
				if e = cacheHint(b); e != nil {
					return nil, nil, e
				}
				s, e := stringFieldAllowEmpty(b, "text")
				if e != nil {
					return nil, nil, e
				}
				kind := "input_text"
				if role == "assistant" {
					kind = "output_text"
				}
				pending = append(pending, map[string]any{"type": kind, "text": s})
			case "tool_use":
				flush()
				if role != "assistant" {
					return nil, nil, errors.New("tool_use requires assistant")
				}
				if e = keys(b, "type", "id", "name", "input", "cache_control"); e != nil {
					return nil, nil, e
				}
				if e = cacheHint(b); e != nil {
					return nil, nil, e
				}
				id, e := stringField(b, "id")
				if e != nil || calls[id] {
					return nil, nil, errors.New("invalid or duplicate call ID")
				}
				originalID := id
				if envelope != nil {
					id, e = opaque.RestoreToolID(id, envelope)
					if e != nil {
						return nil, nil, e
					}
				} else if strings.HasPrefix(id, opaque.ToolPrefix) {
					return nil, nil, errors.New("tool marker missing reasoning envelope")
				}
				restoredIDs[originalID] = id
				name, e := stringField(b, "name")
				if e != nil {
					return nil, nil, e
				}
				args, ok := b["input"].(map[string]any)
				if !ok {
					return nil, nil, errors.New("tool input must be object")
				}
				encoded, _ := json.Marshal(args)
				calls[id] = true
				input = append(input, map[string]any{"type": "function_call", "call_id": id, "name": forward[name], "arguments": string(encoded)})
			case "tool_result":
				flush()
				if role != "user" {
					return nil, nil, errors.New("tool_result requires user")
				}
				if e = keys(b, "type", "tool_use_id", "content", "is_error", "cache_control"); e != nil {
					return nil, nil, e
				}
				if e = cacheHint(b); e != nil {
					return nil, nil, e
				}
				id, e := stringField(b, "tool_use_id")
				if restored, ok := restoredIDs[id]; ok {
					id = restored
				}
				if e != nil || !calls[id] || answered[id] {
					return nil, nil, errors.New("unpaired or duplicate tool result")
				}
				isError := false
				if flag, ok := b["is_error"]; ok {
					var yes bool
					isError, yes = flag.(bool)
					if !yes {
						return nil, nil, errors.New("is_error must be boolean")
					}
				}
				texts, references, e := toolResultContent(b["content"], forward, toolDefinitions)
				for _, name := range references {
					if !loadedTools[name] {
						out["tools"] = append(out["tools"].([]any), toolDefinitions[name])
						loadedTools[name] = true
					}
				}
				if e != nil {
					return nil, nil, e
				}
				answered[id] = true
				result := []any{}
				// Responses carries tool success/failure in its output content.
				// Preserve the failure flag even when the tool returned no text.
				if isError {
					result = append(result, map[string]any{"type": "input_text", "text": "Tool execution failed (is_error=true)."})
				}
				for _, text := range texts {
					result = append(result, map[string]any{"type": "input_text", "text": text})
				}
				input = append(input, map[string]any{"type": "function_call_output", "call_id": id, "output": result})
			default:
				return nil, nil, fmt.Errorf("unsupported content type %q (opaque reasoning is not enabled)", typ)
			}
		}
		flush()
		if envelope != nil {
			public := append([]any(nil), input[startInput:]...)
			if layout := envelope.PublicItems(); len(layout) > 0 {
				public, e = restorePublicBoundaries(public, layout)
				if e != nil {
					return nil, nil, e
				}
			}
			input = input[:startInput]
			next := 0
			for _, item := range envelope.Items() {
				count := item.OutputIndex - next
				if count < 0 || count > len(public) {
					return nil, nil, errors.New("reasoning output position mismatch")
				}
				input = append(input, public[:count]...)
				public = public[count:]
				raw, e := objectJSON(item.Raw)
				if e != nil {
					return nil, nil, e
				}
				input = append(input, raw)
				rawReplay = append(rawReplay, item)
				next = item.OutputIndex + 1
			}
			input = append(input, public...)
		}
	}
	for id := range calls {
		if !answered[id] {
			return nil, nil, errors.New("unpaired tool call")
		}
	}
	out["input"] = input
	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, nil, err
	}
	for _, item := range rawReplay {
		value, err := objectJSON(item.Raw)
		if err != nil {
			return nil, nil, err
		}
		canonical, _ := json.Marshal(value)
		encoded = bytes.Replace(encoded, canonical, item.Raw, 1)
	}
	return encoded, c, nil
}
func collectNames(root map[string]any) ([]string, error) {
	var names []string
	if tools, ok := root["tools"].([]any); ok {
		for _, v := range tools {
			m, e := asObject(v)
			if e != nil {
				return nil, e
			}
			n, e := stringField(m, "name")
			if e != nil {
				return nil, e
			}
			names = append(names, n)
		}
	}
	if ms, ok := root["messages"].([]any); ok {
		for _, v := range ms {
			m, e := asObject(v)
			if e != nil {
				return nil, e
			}
			bs, e := contentBlocks(m["content"])
			if e != nil {
				return nil, e
			}
			for _, b := range bs {
				if b["type"] == "tool_use" {
					n, e := stringField(b, "name")
					if e != nil {
						return nil, e
					}
					names = append(names, n)
				}
			}
		}
	}
	return names, nil
}
func contentBlocks(v any) ([]map[string]any, error) {
	if s, ok := v.(string); ok {
		return []map[string]any{{"type": "text", "text": s}}, nil
	}
	a, ok := v.([]any)
	if !ok {
		return nil, errors.New("content must be string or array")
	}
	r := make([]map[string]any, 0, len(a))
	for _, v := range a {
		m, e := asObject(v)
		if e != nil {
			return nil, e
		}
		r = append(r, m)
	}
	return r, nil
}
func textContent(v any) ([]string, error) {
	bs, e := contentBlocks(v)
	if e != nil {
		return nil, e
	}
	texts := []string{}
	for _, b := range bs {
		if b["type"] != "text" {
			return nil, errors.New("only text is supported here")
		}
		if e = keys(b, "type", "text", "cache_control"); e != nil {
			return nil, e
		}
		if e = cacheHint(b); e != nil {
			return nil, e
		}
		s, e := stringFieldAllowEmpty(b, "text")
		if e != nil {
			return nil, e
		}
		texts = append(texts, s)
	}
	return texts, nil
}

// Cache hints affect provider caching, not public content. Only the documented
// ephemeral hint is accepted; its text/schema remains intact in Responses.
func cacheHint(m map[string]any) error {
	if v, ok := m["cache_control"]; ok {
		c, e := asObject(v)
		if e != nil {
			return e
		}
		if e = keys(c, "type", "ttl"); e != nil {
			return e
		}
		if c["type"] != "ephemeral" {
			return errors.New("unsupported cache hint")
		}
		if ttl, ok := c["ttl"]; ok && ttl != "5m" && ttl != "1h" {
			return errors.New("unsupported cache ttl")
		}
	}
	return nil
}
func keys(m map[string]any, allowed ...string) error {
	for k := range m {
		ok := false
		for _, a := range allowed {
			if a == k {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("unsupported field %q", k)
		}
	}
	return nil
}
func asObject(v any) (map[string]any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("expected object")
	}
	return m, nil
}
func stringFieldAllowEmpty(m map[string]any, k string) (string, error) {
	s, ok := m[k].(string)
	if !ok {
		return "", fmt.Errorf("%s must be string", k)
	}
	return s, nil
}
func stringField(m map[string]any, k string) (string, error) {
	s, e := stringFieldAllowEmpty(m, k)
	if e != nil || s == "" {
		return "", fmt.Errorf("%s must be nonempty string", k)
	}
	return s, nil
}
func integer(v any) (int64, error) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, errors.New("expected integer")
	}
	return n.Int64()
}
func objectJSON(b []byte) (map[string]any, error) {
	if !utf8.Valid(b) || !validUnicodeEscapes(b) {
		return nil, errors.New("invalid UTF-8 JSON")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	v, e := readJSON(d, 0)
	if e != nil {
		return nil, e
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, errors.New("trailing JSON data")
	}
	return asObject(v)
}
func readJSON(d *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, errors.New("JSON nesting exceeds limit")
	}
	t, e := d.Token()
	if e != nil {
		return nil, errors.New("malformed JSON")
	}
	switch t {
	case json.Delim('{'):
		m := map[string]any{}
		for d.More() {
			k, e := d.Token()
			if e != nil {
				return nil, errors.New("malformed JSON key")
			}
			s, ok := k.(string)
			if !ok {
				return nil, errors.New("non-string JSON key")
			}
			if _, ok = m[s]; ok {
				return nil, errors.New("duplicate JSON key")
			}
			v, e := readJSON(d, depth+1)
			if e != nil {
				return nil, e
			}
			m[s] = v
		}
		if _, e = d.Token(); e != nil {
			return nil, errors.New("unclosed object")
		}
		return m, nil
	case json.Delim('['):
		a := []any{}
		for d.More() {
			v, e := readJSON(d, depth+1)
			if e != nil {
				return nil, e
			}
			a = append(a, v)
		}
		if _, e = d.Token(); e != nil {
			return nil, errors.New("unclosed array")
		}
		return a, nil
	default:
		if _, ok := t.(json.Delim); ok {
			return nil, errors.New("unexpected delimiter")
		}
		return t, nil
	}
}

// validUnicodeEscapes supplements encoding/json, which replaces unpaired UTF-16
// surrogates. JSON syntax remains the standard library decoder's responsibility.
func validUnicodeEscapes(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != 92 {
			continue
		}
		i++
		if i >= len(b) {
			return false
		}
		if b[i] != 'u' {
			continue
		}
		if i+4 >= len(b) {
			return false
		}
		n, e := strconv.ParseUint(string(b[i+1:i+5]), 16, 16)
		if e != nil {
			return false
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return false
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+6 >= len(b) || b[i+1] != 92 || b[i+2] != 'u' {
				return false
			}
			low, e := strconv.ParseUint(string(b[i+3:i+7]), 16, 16)
			if e != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}

// ValidateJSONObject checks JSON shape, duplicate keys, depth and Unicode scalar
// escapes without applying Responses-specific message or policy constraints.
func ValidateJSONObject(body []byte) error {
	_, err := objectJSON(body)
	return err
}

// restorePublicBoundaries distinguishes multiple parts of one public message
// from multiple public messages. Receipt authorization has already checked the
// envelope and the public content before this projection is reconstructed.
func restorePublicBoundaries(public []any, layout []opaque.PublicItem) ([]any, error) {
	flat := []map[string]any{}
	for _, v := range public {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, errors.New("invalid public replay")
		}
		if m["role"] == "assistant" {
			parts, ok := m["content"].([]any)
			if !ok {
				return nil, errors.New("invalid public replay")
			}
			for _, part := range parts {
				flat = append(flat, map[string]any{"role": "assistant", "content": []any{part}})
			}
		} else {
			flat = append(flat, m)
		}
	}
	out := []any{}
	for _, item := range layout {
		if item.Blocks > len(flat) {
			return nil, errors.New("public replay boundary mismatch")
		}
		if item.Type == "message" {
			parts := []any{}
			for _, m := range flat[:item.Blocks] {
				if m["role"] != "assistant" {
					return nil, errors.New("public replay type mismatch")
				}
				parts = append(parts, m["content"].([]any)...)
			}
			m := map[string]any{"role": "assistant", "content": parts}
			if item.PhaseNull {
				m["phase"] = nil
			}
			if item.Phase != "" {
				m["phase"] = item.Phase
			}
			out = append(out, m)
		} else {
			if flat[0]["type"] != "function_call" {
				return nil, errors.New("public replay type mismatch")
			}
			out = append(out, flat[0])
		}
		flat = flat[item.Blocks:]
	}
	if len(flat) != 0 {
		return nil, errors.New("public replay boundary mismatch")
	}
	return out, nil
}
