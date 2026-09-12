package gateway

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"
)

const nativeInput = `{"model":"route-alias","max_tokens":10,"stop_sequences":["STOP"],"messages":[{"role":"user","content":"hello"},{"role":"system","content":"in place"},{"role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"Read","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","is_error":true,"content":"file missing"}]}]}`
const nativeOutput = `{"id":"msg","type":"message","role":"assistant","model":"canonical","content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":2,"output_tokens":1}}`

func nativeConfig(tr *http.Transport) MessagesConfig {
	return MessagesConfig{Transport: tr, AnthropicVersion: "2023-06-01", AllowedBetas: []string{"allowed-beta"}, Limits: translate.Limits{MaxBodyBytes: 1 << 20, MaxOutputBytes: 1 << 20, MaxEventBytes: 1 << 16}, MeasureInput: func(ModelEntry, []byte) (int64, error) { return 2, nil }}
}
func nativeTLS(t *testing.T, h http.HandlerFunc) (*http.Transport, *atomic.Int32) {
	t.Helper()
	s := httptest.NewTLSServer(h)
	t.Cleanup(s.Close)
	tr := s.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	cfg := tr.TLSClientConfig.Clone()
	calls := &atomic.Int32{}
	tr.DialTLSContext = func(ctx context.Context, n, addr string) (net.Conn, error) {
		calls.Add(1)
		if addr != "api.anthropic.com:443" && addr != "api.z.ai:443" {
			t.Errorf("unexpected destination %s", addr)
		}
		return (&tls.Dialer{Config: cfg}).DialContext(ctx, "tcp", s.Listener.Addr().String())
	}
	return tr, calls
}
func nativeQ(t *testing.T) RoutedRequest {
	ref, e := auth.NewAnthropicAPIKey("synthetic-anthropic")
	if e != nil {
		t.Fatal(e)
	}
	return RoutedRequest{Entry: ModelEntry{RouteID: "route-alias", UpstreamID: "canonical", Provider: ProviderAnthropic, AuthMethod: AuthAPIKey, Capabilities: Capabilities{ContextTokens: 100, Tools: true, Streaming: true}}, Body: []byte(nativeInput), Credential: ref, Generation: 1, Headers: http.Header{"Anthropic-Beta": {"allowed-beta, forbidden-beta"}, "Authorization": {"inherited"}, "X-Api-Key": {"inherited"}}}
}
func TestAnthropicNativeMessagesPreserveAndFilter(t *testing.T) {
	seen := make(chan *http.Request, 1)
	tr, calls := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var d map[string]json.RawMessage
		_ = json.Unmarshal(b, &d)
		var model string
		_ = json.Unmarshal(d["model"], &model)
		if model != "canonical" || !strings.Contains(string(d["messages"]), `"is_error":true`) || !strings.Contains(string(d["messages"]), `"role":"system"`) || string(d["stop_sequences"]) != `["STOP"]` {
			t.Error(string(b))
		}
		seen <- r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, nativeOutput)
	})
	a, e := NewAnthropicAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Send(context.Background(), nativeQ(t))
	body := oaiRead(t, r, e)
	if body != nativeOutput {
		t.Fatal(body)
	}
	q := <-seen
	if q.Host != "api.anthropic.com" || q.URL.Path != "/v1/messages" || q.Header.Get("X-Api-Key") != "synthetic-anthropic" || q.Header.Get("Authorization") != "" || q.Header.Get("Anthropic-Beta") != "allowed-beta" || q.Header.Get("Anthropic-Version") != "2023-06-01" || calls.Load() != 1 {
		t.Fatal("header/endpoint boundary")
	}
}
func TestAnthropicNativeInvalidHistoryNoEgress(t *testing.T) {
	tr, calls := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected send") })
	for _, kind := range []string{"opaque", "pair", "context", "provider", "generation", "policy", "image", "json", "tools capability", "stream capability", "OAuth"} {
		q := nativeQ(t)
		cfg := nativeConfig(tr)
		switch kind {
		case "opaque":
			q.Body = []byte(`{"max_tokens":1,"messages":[{"role":"assistant","content":[{"type":"thinking","thinking":"private","signature":"opaque"}]}]}`)
		case "pair":
			q.Body = []byte(`{"max_tokens":1,"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"unpaired","name":"t","input":{}}]}]}`)
		case "context":
			cfg.MeasureInput = nil
		case "provider":
			q.Entry.Provider = ProviderOpenAI
		case "generation":
			q.Generation = 2
		case "policy":
			q.Body = []byte(strings.Replace(nativeInput, `"max_tokens":10`, `"thinking":{"type":"adaptive"},"max_tokens":10`, 1))
		case "image":
			q.Body = []byte(`{"max_tokens":1,"messages":[{"role":"user","content":[{"type":"image","source":{}}]}]}`)
		case "json":
			q.Body = []byte(`{"max_tokens":1,"max_tokens":2}`)
		case "tools capability":
			q.Entry.Capabilities.Tools = false
		case "stream capability":
			q.Entry.Capabilities.Streaming = false
			q.Body = []byte(strings.Replace(nativeInput, `"max_tokens":10`, `"stream":true,"max_tokens":10`, 1))
		case "OAuth":
			q.Entry.AuthMethod = AuthMethod("oauth-passthrough")
		}
		a, e := NewAnthropicAdapter(cfg)
		if e != nil {
			t.Fatal(e)
		}
		r, e := a.Send(context.Background(), q)
		_ = oaiRead(t, r, e)
		if r.StatusCode < 400 || r.StatusCode >= 500 {
			t.Fatal(kind, r.StatusCode)
		}
	}
	if calls.Load() != 0 {
		t.Fatal(calls.Load())
	}
}
func nativeSSE() string {
	return "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg\",\"type\":\"message\",\"model\":\"canonical\",\"role\":\"assistant\",\"content\":[]}}\n\nevent: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\nevent: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"hello\"}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":1}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
}
func TestAnthropicNativeStreamSuccessEOFAndClose(t *testing.T) {
	for _, kind := range []string{"success", "EOF", "disconnect", "error", "opaque"} {
		tr, _ := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			r.Body.Close()
			w.Header().Set("Content-Type", "text/event-stream")
			s := nativeSSE()
			switch kind {
			case "EOF":
				s = strings.Split(s, "event: message_stop")[0]
			case "error":
				s = "data: {\"type\":\"error\",\"error\":{\"message\":\"synthetic-secret\"}}\n\n"
			case "opaque":
				s = strings.Replace(s, `"type":"text","text":""`, `"type":"thinking","thinking":"private"`, 1)
			}
			io.WriteString(w, s)
			w.(http.Flusher).Flush()
			if kind == "disconnect" {
				select {
				case <-r.Context().Done():
				case <-time.After(2 * time.Second):
					t.Error("cancel missing")
				}
			}
		})
		a, _ := NewAnthropicAdapter(nativeConfig(tr))
		q := nativeQ(t)
		q.Body = []byte(strings.Replace(nativeInput, `"max_tokens":10`, `"stream":true,"max_tokens":10`, 1))
		r, e := a.Send(context.Background(), q)
		if e != nil {
			t.Fatal(e)
		}
		if kind == "disconnect" {
			r.Body.Close()
			continue
		}
		b, e := io.ReadAll(r.Body)
		r.Body.Close()
		if kind == "success" {
			if e != nil || string(b) != nativeSSE() {
				t.Fatal(string(b), e)
			}
		} else if e == nil || strings.Contains(string(b), "message_stop") || strings.Contains(string(b), "synthetic-secret") {
			t.Fatal(kind, string(b), e)
		}
	}
}

func TestAnthropicNativeStatusAndRedirectBoundaries(t *testing.T) {
	for _, status := range []int{401, 429, 302, 503} {
		tr, calls := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "9")
			w.Header().Set("Location", "https://foreign.invalid")
			w.WriteHeader(status)
			io.WriteString(w, "synthetic-private-body")
		})
		a, _ := NewAnthropicAdapter(nativeConfig(tr))
		r, e := a.Send(context.Background(), nativeQ(t))
		body := oaiRead(t, r, e)
		want := status
		if want == 302 {
			want = 502
		}
		if r.StatusCode != want || strings.Contains(body, "private") || r.Header.Get("Location") != "" || calls.Load() != 1 {
			t.Fatal(status, r.StatusCode, body)
		}
		if status == 429 && r.Header.Get("Retry-After") != "9" {
			t.Fatal("retry missing")
		}
	}
}
func TestAnthropicNativeParserAdversaries(t *testing.T) {
	for _, raw := range []string{
		`{`, `{"max_tokens":1.1,"messages":[]}`, `{"max_tokens":false,"messages":[]}`, `{"max_tokens":0,"messages":[]}`, `{"max_tokens":1,"messages":[]}`, `{"max_tokens":1,"stream":null,"messages":[{"role":"user","content":"x"}]}`, `{"max_tokens":1,"unknown_policy":{},"messages":[{"role":"user","content":"x"}]}`, `{"max_tokens":1,"messages":[null]}`, `{"max_tokens":1,"messages":[{"role":"alien","content":"x"}]}`, `{"max_tokens":1,"messages":[{"role":"user","content":null}]}`, `{"max_tokens":1,"messages":[{"role":"user","content":[null]}]}`, `{"max_tokens":1,"messages":[{"role":"user","content":[{"type":"text","text":4}]}]}`, `{"max_tokens":1,"messages":[{"role":"user","content":[{"type":"tool_use","id":"a","name":"t","input":{}}]}]}`, `{"max_tokens":1,"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"a","name":"t","input":[]}]}]}`, `{"max_tokens":1,"messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"a","content":"x"}]}]}`, `{"max_tokens":1,"system":[{"type":"tool_use"}],"messages":[{"role":"user","content":"x"}]}`, `{"max_tokens":1,"tools":null,"messages":[{"role":"user","content":"x"}]}`, `{"max_tokens":1,"tools":[null],"messages":[{"role":"user","content":"x"}]}`, `{"max_tokens":1,"tools":[{"name":"t","type":"computer","input_schema":{}}],"messages":[{"role":"user","content":"x"}]}`,
	} {
		if _, _, e := nativeRequest([]byte(raw)); e == nil {
			t.Fatal("accepted", raw)
		}
	}
	var base map[string]any
	_ = json.Unmarshal([]byte(nativeInput), &base)
	for _, kind := range []string{"duplicate call", "duplicate result", "bad error flag", "bad result content", "assistant result"} {
		var d map[string]any
		_ = json.Unmarshal([]byte(nativeInput), &d)
		ms := d["messages"].([]any)
		a := ms[2].(map[string]any)
		u := ms[3].(map[string]any)
		block := u["content"].([]any)[0].(map[string]any)
		switch kind {
		case "duplicate call":
			a["content"] = append(a["content"].([]any), a["content"].([]any)[0])
		case "duplicate result":
			u["content"] = append(u["content"].([]any), block)
		case "bad error flag":
			block["is_error"] = "true"
		case "bad result content":
			block["content"] = []any{map[string]any{"type": "document"}}
		case "assistant result":
			u["role"] = "assistant"
		}
		raw, _ := json.Marshal(d)
		if _, _, e := nativeRequest(raw); e == nil {
			t.Fatal(kind)
		}
	}
	base["thinking"] = map[string]any{"type": "disabled"}
	base["system"] = "system prompt"
	base["tools"] = []any{map[string]any{"name": "Read", "input_schema": map[string]any{"type": "object"}}}
	raw, _ := json.Marshal(base)
	if _, tools, e := nativeRequest(raw); e != nil || !tools {
		t.Fatal(e)
	}
	for _, raw := range []string{`{`, strings.Replace(nativeOutput, `"stop_reason":"end_turn"`, `"stop_reason":"unknown"`, 1), strings.Replace(nativeOutput, `"id":"msg"`, `"id":""`, 1), strings.Replace(nativeOutput, `[{"type":"text","text":"hello"}]`, `[]`, 1), strings.Replace(nativeOutput, `"type":"text"`, `"type":"redacted_thinking"`, 1)} {
		if e := nativeResponse([]byte(raw), "canonical"); e == nil {
			t.Fatal("bad response", raw)
		}
	}
}
func TestAnthropicNativeStreamStateAdversaries(t *testing.T) {
	valid := nativeSSE()
	frames := strings.Split(strings.TrimSuffix(valid, "\n\n"), "\n\n")
	cases := []string{frames[1] + "\n\n", frames[0] + "\n\n" + frames[0] + "\n\n", frames[0] + "\n\n" + frames[5] + "\n\n", strings.Join(frames[:3], "\n\n") + "\n\n" + frames[4] + "\n\n", strings.Join(frames[:4], "\n\n") + "\n\n" + frames[3] + "\n\n", strings.Replace(valid, `"index":0`, `"index":1`, 1), strings.Replace(valid, `"index":0`, `"index":0.5`, 1), strings.Replace(valid, `"type":"text_delta"`, `"type":"thinking_delta"`, 1), strings.Replace(valid, "event: message_start", "event: wrong", 1), "event: x\nevent: y\n\n", "id: x\n\n", "data: {\"type\":\"ping\",\"type\":\"ping\"}\n\n"}
	for i, s := range cases {
		b := newNativeBody(context.Background(), io.NopCloser(strings.NewReader(s)), "canonical", nativeConfig(&http.Transport{}).Limits)
		out, e := io.ReadAll(b)
		b.Close()
		if e == nil || strings.Contains(string(out), "message_stop") {
			t.Fatal(i, string(out), e)
		}
	}
	b := newNativeBody(context.Background(), io.NopCloser(strings.NewReader(": keepalive\n\ndata: {\"type\":\"ping\"}\n\n"+valid)), "canonical", nativeConfig(&http.Transport{}).Limits)
	if _, e := io.ReadAll(b); e != nil {
		t.Fatal(e)
	}
	b.Close()
	cfg := nativeConfig(&http.Transport{})
	cfg.Limits.MaxEventBytes = 16
	b = newNativeBody(context.Background(), io.NopCloser(strings.NewReader(valid)), "canonical", cfg.Limits)
	if _, e := io.ReadAll(b); e == nil {
		t.Fatal("event bound")
	}
	b.Close()
	cfg = nativeConfig(&http.Transport{})
	cfg.Limits.MaxOutputBytes = 16
	b = newNativeBody(context.Background(), io.NopCloser(strings.NewReader(valid)), "canonical", cfg.Limits)
	if _, e := io.ReadAll(b); e == nil {
		t.Fatal("total bound")
	}
	b.Close()
}

func TestAnthropicNativeInterleavedToolStreams(t *testing.T) {
	encode := func(v map[string]any) string { b, _ := json.Marshal(v); return "data: " + string(b) + "\n\n" }
	for _, broken := range []bool{false, true} {
		var s strings.Builder
		s.WriteString(strings.Split(nativeSSE(), "event: content_block_start")[0])
		for i, id := range []string{"call-a", "call-b"} {
			s.WriteString(encode(map[string]any{"type": "content_block_start", "index": i, "content_block": map[string]any{"type": "tool_use", "id": id, "name": "Read", "input": map[string]any{}}}))
		}
		for _, chunk := range []struct {
			idx  int
			text string
		}{{1, `{"b":`}, {0, `{"a":1}`}, {1, `2}`}} {
			text := chunk.text
			if broken && chunk.idx == 0 {
				text = "{"
			}
			s.WriteString(encode(map[string]any{"type": "content_block_delta", "index": chunk.idx, "delta": map[string]any{"type": "input_json_delta", "partial_json": text}}))
		}
		for _, i := range []int{1, 0} {
			s.WriteString(encode(map[string]any{"type": "content_block_stop", "index": i}))
		}
		s.WriteString(encode(map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": "tool_use"}}))
		s.WriteString(encode(map[string]any{"type": "message_stop"}))
		b := newNativeBody(context.Background(), io.NopCloser(strings.NewReader(s.String())), "canonical", nativeConfig(&http.Transport{}).Limits)
		out, e := io.ReadAll(b)
		b.Close()
		if broken {
			if e == nil || strings.Contains(string(out), "message_stop") {
				t.Fatal("broken arguments accepted")
			}
		} else if e != nil || string(out) != s.String() {
			t.Fatal("interleaved stream changed", e)
		}
	}
}
func TestAnthropicNativeCancellationClosesBlockedRead(t *testing.T) {
	r, w := io.Pipe()
	t.Cleanup(func() { r.Close(); w.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := newNativeBody(ctx, r, "canonical", nativeConfig(&http.Transport{}).Limits)
	defer b.Close()
	done := make(chan error, 1)
	go func() { _, e := b.Read(make([]byte, 100)); done <- e }()
	cancel()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("cancel hidden")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked read not closed")
	}
}

func TestAnthropicNativeRawWireBoundaries(t *testing.T) {
	for _, fragmented := range []bool{false, true} {
		for name, raw := range map[string]string{"LF": nativeSSE(), "CRLF": strings.ReplaceAll(nativeSSE(), "\n", "\r\n"), "mixed": strings.Replace(nativeSSE(), "\n", "\r\n", 3)} {
			for _, offset := range []int{-1, 0, 1} {
				limits := nativeConfig(&http.Transport{}).Limits
				limits.MaxOutputBytes = len(raw) + offset
				var reader io.Reader = strings.NewReader(raw)
				if fragmented {
					reader = iotest.OneByteReader(reader)
				}
				b := newNativeBody(context.Background(), io.NopCloser(reader), "canonical", limits)
				out, e := io.ReadAll(b)
				b.Close()
				success := strings.Contains(string(out), "event: message_stop")
				if offset < 0 && (e == nil || success) {
					t.Errorf("%s fragmented=%t exceeded raw limit", name, fragmented)
				}
				if offset >= 0 && (e != nil || !success) {
					t.Errorf("%s fragmented=%t offset=%d failed: %v", name, fragmented, offset, e)
				}
			}
		}
	}
	for _, raw := range []string{strings.TrimSuffix(nativeSSE(), "\n\n"), ": unterminated\r", strings.Split(nativeSSE(), "event: message_stop")[0]} {
		b := newNativeBody(context.Background(), io.NopCloser(iotest.OneByteReader(strings.NewReader(raw))), "canonical", nativeConfig(&http.Transport{}).Limits)
		out, e := io.ReadAll(b)
		b.Close()
		if e == nil || strings.Contains(string(out), "event: message_stop") {
			t.Fatal("EOF before complete terminal accepted")
		}
	}
}
