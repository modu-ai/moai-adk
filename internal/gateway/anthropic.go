package gateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// MessagesConfig enables native ordinary public Messages. Authentication mode
// is explicit in the constructor; opaque reasoning is not inferred. MeasureInput must be local.
type MessagesConfig struct {
	Transport        *http.Transport
	AnthropicVersion string
	AllowedBetas     []string
	Limits           translate.Limits
	MeasureInput     func(ModelEntry, []byte) (int64, error)
}
type MessagesAdapter struct {
	config   MessagesConfig
	provider ProviderID
	method   AuthMethod
	endpoint string
	betas    map[string]bool
}

var betaToken = regexp.MustCompile(`^[A-Za-z0-9-]{1,100}$`)

func NewAnthropicAdapter(c MessagesConfig) (*MessagesAdapter, error) {
	return newMessagesAdapter(c, ProviderAnthropic, AuthAPIKey, auth.AnthropicEndpoint)
}

// NewAnthropicOAuthAdapter uses the request-scoped native OAuth reference after M0.
func NewAnthropicOAuthAdapter(c MessagesConfig) (*MessagesAdapter, error) {
	return newMessagesAdapter(c, ProviderAnthropic, AuthOAuthPassthrough, auth.AnthropicEndpoint)
}
func newMessagesAdapter(c MessagesConfig, p ProviderID, m AuthMethod, endpoint string) (*MessagesAdapter, error) {
	if c.Transport == nil || c.Transport.Proxy != nil || c.Transport.TLSClientConfig != nil && c.Transport.TLSClientConfig.InsecureSkipVerify || c.AnthropicVersion == "" || strings.ContainsAny(c.AnthropicVersion, "\r\n") || c.Limits.MaxBodyBytes <= 0 || c.Limits.MaxEventBytes <= 0 || c.Limits.MaxOutputBytes <= 0 {
		return nil, errors.New("native Messages configuration invalid")
	}
	c.Transport = c.Transport.Clone()
	c.Transport.DisableKeepAlives = true
	c.Transport.ForceAttemptHTTP2 = false
	if c.Transport.ResponseHeaderTimeout <= 0 {
		c.Transport.ResponseHeaderTimeout = 30 * time.Second
	}
	a := &MessagesAdapter{config: c, provider: p, method: m, endpoint: endpoint, betas: map[string]bool{}}
	for _, b := range c.AllowedBetas {
		if !betaToken.MatchString(b) {
			return nil, errors.New("invalid allowed beta")
		}
		a.betas[b] = true
	}
	a.config.AllowedBetas = nil
	return a, nil
}

// Send preserves ordinary native Messages instead of imposing Responses schema.
// @MX:WARN: [AUTO] Provider, policy and credential checks precede fixed-endpoint egress.
// @MX:REASON: The final generation recheck is not an atomic GLM writer barrier.
func (a *MessagesAdapter) Send(ctx context.Context, q RoutedRequest) (*http.Response, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if q.Entry.Provider != a.provider || q.Entry.AuthMethod != a.method || q.Entry.UpstreamID == "" {
		return openAIError(400), nil
	}
	if len(q.Body) > a.config.Limits.MaxBodyBytes || q.Entry.Capabilities.ContextTokens <= 0 || a.config.MeasureInput == nil {
		return openAIError(400), nil
	}
	n, e := a.config.MeasureInput(q.Entry, q.Body)
	if e != nil || n < 0 || n > int64(q.Entry.Capabilities.ContextTokens) {
		return openAIError(400), nil
	}
	nativePolicy := a.config.Limits.PolicyProfile == translate.PolicyAnthropicNative
	if a.config.Limits.PolicyProfile != "" && (!nativePolicy || a.provider != ProviderAnthropic || (q.Entry.UpstreamID != "claude-opus-5" && q.Entry.UpstreamID != "claude-sonnet-5")) {
		return openAIError(400), nil
	}
	stream, hasTools, e := nativeRequestProfile(q.Body, nativePolicy)
	if e != nil || stream && !q.Entry.Capabilities.Streaming || hasTools && !q.Entry.Capabilities.Tools {
		return openAIError(400), nil
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(q.Body, &fields)
	body := q.Body
	var originalModel string
	_ = json.Unmarshal(fields["model"], &originalModel)
	if originalModel != q.Entry.UpstreamID {
		fields["model"], _ = json.Marshal(q.Entry.UpstreamID)
		body, _ = json.Marshal(fields)
	}
	if len(body) > a.config.Limits.MaxBodyBytes {
		return openAIError(400), nil
	}
	if q.Credential == nil || q.Credential.Provider() != a.provider {
		return openAIError(401), nil
	}
	generation, e := q.Credential.Generation()
	if e != nil || generation != q.Generation {
		return openAIError(401), nil
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Anthropic-Version", a.config.AnthropicVersion)
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	if a.provider == ProviderAnthropic {
		var allowed []string
		seen := map[string]bool{}
		for _, v := range q.Headers.Values("Anthropic-Beta") {
			for _, b := range strings.Split(v, ",") {
				b = strings.TrimSpace(b)
				if a.betas[b] && !seen[b] {
					allowed = append(allowed, b)
					seen[b] = true
				}
			}
		}
		if len(allowed) > 0 {
			req.Header.Set("Anthropic-Beta", strings.Join(allowed, ","))
		}
	}
	if e = q.Credential.Apply(req); e != nil {
		return openAIError(401), nil
	}
	if req.URL == nil || req.URL.String() != a.endpoint || req.Method != http.MethodPost || req.Host != "" && req.Host != req.URL.Host {
		return openAIError(401), nil
	}
	generation, e = q.Credential.Generation()
	if e != nil || generation != q.Generation {
		return openAIError(401), nil
	}
	up, e := a.config.Transport.RoundTrip(req)
	if e != nil {
		if up != nil && up.Body != nil {
			_ = up.Body.Close() // error-path discard; the mapped client error is already fixed
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return openAIError(502), nil
	}
	if up == nil || up.Body == nil {
		return openAIError(502), nil
	}
	if up.StatusCode != 200 {
		_ = up.Body.Close() // error-path discard; the upstream status is forwarded below
		status := up.StatusCode
		if status < 400 || status > 599 {
			status = 502
		}
		out := openAIError(status)
		if status == 429 || status == 503 {
			if v := validRetryAfter(up.Header.Get("Retry-After")); v != "" {
				out.Header.Set("Retry-After", v)
			}
		}
		return out, nil
	}
	media, _, e := mime.ParseMediaType(up.Header.Get("Content-Type"))
	if e != nil {
		_ = up.Body.Close() // error-path discard; the mapped client error is already fixed
		return openAIError(502), nil
	}
	if stream {
		if media != "text/event-stream" {
			_ = up.Body.Close() // error-path discard; the mapped client error is already fixed
			return openAIError(502), nil
		}
		b := newNativeBody(ctx, up.Body, q.Entry.UpstreamID, a.config.Limits)
		b.nativePolicy = nativePolicy
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: b}, nil
	}
	defer func() { _ = up.Body.Close() }() // body fully read before this point; no write-back to lose
	if media != "application/json" {
		return openAIError(502), nil
	}
	raw, e := io.ReadAll(io.LimitReader(up.Body, int64(a.config.Limits.MaxOutputBytes)+1))
	if e != nil || len(raw) > a.config.Limits.MaxOutputBytes {
		return openAIError(502), nil
	}
	if e = nativeResponseProfile(raw, q.Entry.UpstreamID, nativePolicy); e != nil {
		return openAIError(502), nil
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(raw)), ContentLength: int64(len(raw))}, nil
}
func nativeObject(raw []byte) (map[string]any, error) {
	if e := translate.ValidateJSONObject(raw); e != nil {
		return nil, e
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v map[string]any
	e := d.Decode(&v)
	return v, e
}

// @MX:WARN: [AUTO] Validates nested native public history without rewriting it.
// @MX:REASON: The shared JSON gate bounds depth before this tool-pair validation.
func nativeRequest(raw []byte) (bool, bool, error) { return nativeRequestProfile(raw, false) }
func nativeRequestProfile(raw []byte, profile bool) (bool, bool, error) {
	v, e := nativeObject(raw)
	if e != nil {
		return false, false, e
	}
	if profile {
		if _, e := translate.ValidateNativePolicy(raw, translate.PolicyAnthropicNative); e != nil {
			return false, false, e
		}
	}
	for k := range v {
		switch k {
		case "model", "messages", "system", "max_tokens", "stream", "tools", "tool_choice", "metadata", "temperature", "top_p", "top_k", "stop_sequences":
		case "output_config", "context_management":
			if !profile {
				return false, false, errors.New("native policy unsupported")
			}
		case "thinking":
			if profile {
				continue
			}
			m, ok := v[k].(map[string]any)
			if !ok || len(m) != 1 || m["type"] != "disabled" {
				return false, false, errors.New("opaque policy unsupported")
			}
		default:
			return false, false, errors.New("native policy unsupported")
		}
	}
	num, ok := v["max_tokens"].(json.Number)
	if !ok {
		return false, false, errors.New("max_tokens invalid")
	}
	n, e := num.Int64()
	if e != nil || n <= 0 {
		return false, false, errors.New("max_tokens invalid")
	}
	stream := false
	if s, ok := v["stream"]; ok {
		var valid bool
		stream, valid = s.(bool)
		if !valid {
			return false, false, errors.New("stream invalid")
		}
	}
	if sys, ok := v["system"]; ok {
		if _, e = nativeBlocksProfile(sys, "system", nil, nil, profile); e != nil {
			return false, false, e
		}
	}
	ms, ok := v["messages"].([]any)
	if !ok || len(ms) == 0 {
		return false, false, errors.New("messages invalid")
	}
	calls := map[string]bool{}
	results := map[string]bool{}
	hasTools := false
	for _, value := range ms {
		m, ok := value.(map[string]any)
		if !ok {
			return false, false, errors.New("message invalid")
		}
		role, _ := m["role"].(string)
		if role != "user" && role != "assistant" && role != "system" {
			return false, false, errors.New("role invalid")
		}
		tools, e := nativeBlocksProfile(m["content"], role, calls, results, profile)
		if e != nil {
			return false, false, e
		}
		hasTools = hasTools || tools
	}
	for id := range calls {
		if !results[id] {
			return false, false, errors.New("unfinished tool pair")
		}
	}
	if list, ok := v["tools"]; ok {
		tools, valid := list.([]any)
		if !valid {
			return false, false, errors.New("tools invalid")
		}
		seen := map[string]bool{}
		for _, value := range tools {
			m, ok := value.(map[string]any)
			if !ok {
				return false, false, errors.New("tool invalid")
			}
			name, _ := m["name"].(string)
			_, schema := m["input_schema"].(map[string]any)
			if name == "" || seen[name] || !schema || m["type"] != nil {
				return false, false, errors.New("tool invalid")
			}
			seen[name] = true
		}
		hasTools = hasTools || len(tools) > 0
	}
	return stream, hasTools, nil
}
func nativeBlocks(v any, role string, calls, results map[string]bool) (bool, error) {
	return nativeBlocksProfile(v, role, calls, results, false)
}
func nativeBlocksProfile(v any, role string, calls, results map[string]bool, profile bool) (bool, error) {
	if _, ok := v.(string); ok {
		return false, nil
	}
	bs, ok := v.([]any)
	if !ok {
		return false, errors.New("content invalid")
	}
	has := false
	for _, value := range bs {
		b, ok := value.(map[string]any)
		if !ok {
			return false, errors.New("block invalid")
		}
		switch b["type"] {
		case "thinking":
			if !profile || role != "assistant" || len(b) != 3 {
				return false, errors.New("unsupported thinking block")
			}
			if _, ok := b["thinking"].(string); !ok {
				return false, errors.New("invalid thinking text")
			}
			if _, ok := b["signature"].(string); !ok {
				return false, errors.New("invalid signature")
			}
		case "redacted_thinking":
			if !profile || role != "assistant" || len(b) != 2 {
				return false, errors.New("unsupported redacted thinking")
			}
			if data, ok := b["data"].(string); !ok || data == "" {
				return false, errors.New("invalid redacted data")
			}

		case "text":
			if _, ok := b["text"].(string); !ok {
				return false, errors.New("text invalid")
			}
		case "tool_use":
			if role != "assistant" || calls == nil {
				return false, errors.New("tool role invalid")
			}
			id, _ := b["id"].(string)
			name, _ := b["name"].(string)
			_, input := b["input"].(map[string]any)
			if id == "" || name == "" || !input || calls[id] {
				return false, errors.New("tool call invalid")
			}
			calls[id] = true
			has = true
		case "tool_result":
			if role != "user" || results == nil {
				return false, errors.New("result role invalid")
			}
			id, _ := b["tool_use_id"].(string)
			if id == "" || !calls[id] || results[id] {
				return false, errors.New("tool result unpaired")
			}
			if flag, ok := b["is_error"]; ok {
				if _, valid := flag.(bool); !valid {
					return false, errors.New("tool result error flag invalid")
				}
			}
			if _, e := nativeBlocks(b["content"], "system", nil, nil); e != nil {
				return false, e
			}
			results[id] = true
			has = true
		default:
			return false, errors.New("unsupported public content")
		}
	}
	return has, nil
}
func nativeStop(v any) bool {
	return v == "end_turn" || v == "tool_use" || v == "max_tokens" || v == "stop_sequence"
}
func nativeResponse(raw []byte, model string) error { return nativeResponseProfile(raw, model, false) }
func nativeResponseProfile(raw []byte, model string, profile bool) error {
	v, e := nativeObject(raw)
	if e != nil {
		return e
	}
	if v["type"] != "message" || v["role"] != "assistant" || v["model"] != model || !nativeStop(v["stop_reason"]) {
		return errors.New("invalid native response")
	}
	id, _ := v["id"].(string)
	if id == "" {
		return errors.New("response id invalid")
	}
	bs, ok := v["content"].([]any)
	if !ok || len(bs) == 0 {
		return errors.New("no public response")
	}
	_, e = nativeBlocksProfile(bs, "assistant", map[string]bool{}, nil, profile)
	return e
}

type nativeStreamBlock struct {
	kind      string
	closed    bool
	arguments strings.Builder
}
type nativeBody struct {
	nativePolicy             bool
	ctx                      context.Context
	upstream                 io.ReadCloser
	scanner                  *bufio.Scanner
	limits                   translate.Limits
	model                    string
	pending                  *bytes.Reader
	once                     sync.Once
	stop                     func() bool
	total                    int
	started, delta, terminal bool
	blocks                   map[int]*nativeStreamBlock
	ids                      map[string]bool
}

// newNativeBody creates no converter goroutine. Reads validate one bounded event
// before exposing it, while context cancellation closes a blocked HTTP body.
func newNativeBody(ctx context.Context, up io.ReadCloser, model string, limits translate.Limits) *nativeBody {
	b := &nativeBody{ctx: ctx, upstream: up, limits: limits, model: model, blocks: map[int]*nativeStreamBlock{}, ids: map[string]bool{}}
	b.scanner = bufio.NewScanner(up)
	// Count consumed wire bytes before ScanLines removes CRLF. Preserve bounded
	// incremental reads rather than buffering the entire provider response.
	b.scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		advance, token, err := bufio.ScanLines(data, atEOF)
		remaining := b.limits.MaxOutputBytes - b.total
		if advance > remaining || (advance == 0 && len(data) > remaining) {
			return 0, nil, errors.New("stream size exceeded")
		}
		b.total += advance
		return advance, token, err
	})
	b.scanner.Buffer(make([]byte, min(4096, limits.MaxEventBytes)), limits.MaxEventBytes)
	b.stop = context.AfterFunc(ctx, func() { _ = up.Close() })
	return b
}
func (b *nativeBody) Close() error {
	b.once.Do(func() { b.stop(); _ = b.upstream.Close() })
	return nil
}
func (b *nativeBody) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if e := b.ctx.Err(); e != nil {
		_ = b.Close()
		return 0, e
	}
	if b.pending != nil && b.pending.Len() > 0 {
		return b.pending.Read(p)
	}
	if b.terminal {
		_ = b.Close()
		return 0, io.EOF
	}
	raw, e := b.next()
	if e != nil {
		_ = b.Close()
		return 0, errors.New("native response stream failed")
	}
	b.pending = bytes.NewReader(raw)
	return b.pending.Read(p)
}
func (b *nativeBody) next() ([]byte, error) {
	var frame, data strings.Builder
	eventName := ""
	eventStart := b.total
	for b.scanner.Scan() {
		line := b.scanner.Text()
		// total counts consumed wire bytes before ScanLines removes CRLF.
		if b.total-eventStart > b.limits.MaxEventBytes {
			return nil, errors.New("stream size exceeded")
		}
		frame.WriteString(line)
		frame.WriteByte('\n')
		if line == "" {
			if data.Len() == 0 {
				frame.Reset()
				eventStart = b.total
				eventName = ""
				continue
			}
			v, e := nativeObject([]byte(strings.TrimSuffix(data.String(), "\n")))
			if e != nil {
				return nil, e
			}
			if eventName != "" && eventName != v["type"] {
				return nil, errors.New("event mismatch")
			}
			if e = b.event(v); e != nil {
				return nil, e
			}
			return []byte(frame.String()), nil
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		key, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch key {
		case "data":
			data.WriteString(value)
			data.WriteByte('\n')
		case "event":
			if eventName != "" {
				return nil, errors.New("duplicate event")
			}
			eventName = value
		default:
			return nil, errors.New("unknown SSE field")
		}
	}
	if e := b.scanner.Err(); e != nil {
		return nil, e
	}
	return nil, io.ErrUnexpectedEOF
}

// @MX:WARN: [AUTO] Native SSE state transitions must precede event exposure.
// @MX:REASON: Invalid indexes or unfinished blocks cannot produce success terminals.
func (b *nativeBody) event(v map[string]any) error {
	if v["type"] == "ping" {
		return nil
	}
	if v["type"] == "message_start" {
		m, ok := v["message"].(map[string]any)
		if !ok || b.started || m["model"] != b.model || m["role"] != "assistant" || m["type"] != "message" {
			return errors.New("bad message start")
		}
		content, ok := m["content"].([]any)
		id, _ := m["id"].(string)
		if !ok || len(content) != 0 || id == "" {
			return errors.New("bad start content")
		}
		b.started = true
		return nil
	}
	if !b.started {
		return errors.New("missing message start")
	}
	if v["type"] == "message_stop" {
		if !b.delta {
			return errors.New("missing final delta")
		}
		b.terminal = true
		return nil
	}
	if b.delta {
		return errors.New("event after final delta")
	}
	if v["type"] == "message_delta" {
		m, ok := v["delta"].(map[string]any)
		if !ok || !nativeStop(m["stop_reason"]) || len(b.blocks) == 0 {
			return errors.New("invalid final delta")
		}
		for _, block := range b.blocks {
			if !block.closed {
				return errors.New("unfinished output block")
			}
		}
		b.delta = true
		return nil
	}
	num, ok := v["index"].(json.Number)
	if !ok {
		return errors.New("index invalid")
	}
	i, e := num.Int64()
	if e != nil || i < 0 || i > 100000 {
		return errors.New("index invalid")
	}
	idx := int(i)
	switch v["type"] {
	case "content_block_start":
		if idx != len(b.blocks) {
			return errors.New("duplicate block")
		}
		block, ok := v["content_block"].(map[string]any)
		if !ok {
			return errors.New("block invalid")
		}
		if _, e := nativeBlocksProfile([]any{block}, "assistant", b.ids, nil, b.nativePolicy); e != nil {
			return e
		}
		kind, _ := block["type"].(string)
		b.blocks[idx] = &nativeStreamBlock{kind: kind}
		return nil
	case "content_block_delta":
		block, ok := b.blocks[idx]
		if !ok || block.closed {
			return errors.New("block not open")
		}
		delta, ok := v["delta"].(map[string]any)
		if !ok {
			return errors.New("delta invalid")
		}
		if b.nativePolicy && block.kind == "thinking" && len(delta) == 2 {
			if delta["type"] == "thinking_delta" {
				if _, ok := delta["thinking"].(string); ok {
					return nil
				}
			}
			if delta["type"] == "signature_delta" {
				if _, ok := delta["signature"].(string); ok {
					return nil
				}
			}
		}
		if block.kind == "text" && delta["type"] == "text_delta" {
			if _, ok := delta["text"].(string); ok {
				return nil
			}
		}
		if block.kind == "tool_use" && delta["type"] == "input_json_delta" {
			if s, ok := delta["partial_json"].(string); ok {
				block.arguments.WriteString(s)
				return nil
			}
		}
		return errors.New("unsupported delta")
	case "content_block_stop":
		block, ok := b.blocks[idx]
		if !ok || block.closed {
			return errors.New("block not open")
		}
		if block.kind == "tool_use" && block.arguments.Len() > 0 {
			if _, e := nativeObject([]byte(block.arguments.String())); e != nil {
				return e
			}
		}
		block.closed = true
		return nil
	default:
		return errors.New("unknown or failed native event")
	}
}
