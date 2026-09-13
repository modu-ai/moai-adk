package gateway

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func TestNativeThinkingPolicyAndPreservation(t *testing.T) {
	raw := []byte(`{"model":"claude-opus-5","max_tokens":32000,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"},"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]},"messages":[{"role":"assistant","content":[{"type":"thinking","thinking":"","signature":"synthetic-signature"},{"type":"redacted_thinking","data":"synthetic"}]},{"role":"user","content":"next"}]}`)
	if _, _, e := nativeRequestProfile(raw, true); e != nil {
		t.Fatal(e)
	}
	if _, _, e := nativeRequest(raw); e == nil {
		t.Fatal("default profile enabled")
	}
	output := `{"type":"message","id":"m","model":"claude-opus-5","role":"assistant","content":[{"type":"thinking","thinking":"","signature":"sig"},{"type":"redacted_thinking","data":"opaque"},{"type":"text","text":"answer"}],"stop_reason":"end_turn"}`
	if e := nativeResponseProfile([]byte(output), "claude-opus-5", true); e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{strings.Replace(output, `"signature":"sig"`, `"signature":null`, 1), strings.Replace(output, `"data":"opaque"`, `"data":null`, 1), strings.Replace(output, `"thinking":""`, `"thinking":"","unknown":1`, 1)} {
		if e := nativeResponseProfile([]byte(bad), "claude-opus-5", true); e == nil {
			t.Fatal("accepted", bad)
		}
	}
}
func TestNativeThinkingStream(t *testing.T) {
	s := nativeSSE()
	s = strings.Replace(s, `"type":"text","text":""`, `"type":"thinking","thinking":"","signature":""`, 1)
	s = strings.Replace(s, `"type":"text_delta","text":"hello"`, `"type":"thinking_delta","thinking":""`, 1)
	marker := `event: content_block_stop`
	pos := strings.Index(s, marker)
	signature := "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"signature_delta\",\"signature\":\"signed\"}}\n\n"
	s = s[:pos] + signature + s[pos:]
	body := newNativeBody(context.Background(), io.NopCloser(strings.NewReader(s)), "canonical", translate.Limits{MaxOutputBytes: 1 << 20, MaxEventBytes: 1 << 16})
	body.nativePolicy = true
	got, e := io.ReadAll(body)
	_ = body.Close() // nativeBody.Close always returns nil
	if e != nil || string(got) != s {
		t.Fatalf("%v %s", e, got)
	}
	for _, bad := range []string{strings.Replace(s, `"signature":"signed"`, `"signature":null`, 1), strings.Replace(s, `"type":"thinking_delta"`, `"type":"text_delta"`, 1), strings.Replace(s, `"type":"signature_delta"`, `"type":"unknown"`, 1), s[:strings.Index(s, "event: message_stop")]} {
		b := newNativeBody(context.Background(), io.NopCloser(strings.NewReader(bad)), "canonical", translate.Limits{MaxOutputBytes: 1 << 20, MaxEventBytes: 1 << 16})
		b.nativePolicy = true
		if _, e := io.ReadAll(b); e == nil {
			t.Fatal("accepted malformed stream")
		}
		_ = b.Close() // nativeBody.Close always returns nil
	}
}
func TestInputEstimateIncludesOutputSchemaOnly(t *testing.T) {
	base := `{"messages":[],"output_config":{"effort":"high","format":{"type":"json_schema","schema":{"type":"object","properties":{"title":{"type":"string"}},"required":["title"],"additionalProperties":false}}}}`
	got, e := EstimateInputTokens(ModelEntry{}, []byte(base))
	if e != nil || got <= 4 {
		t.Fatalf("%d %v", got, e)
	}
	other, e := EstimateInputTokens(ModelEntry{}, []byte(strings.Replace(base, `"high"`, `"long ignored policy value"`, 1)))
	if e != nil || got != other {
		t.Fatalf("policy overhead incorrectly counted %d %d", got, other)
	}
}

func TestNativePolicyWireAndProviderIsolation(t *testing.T) {
	raw := `{ "model":"claude-opus-5", "max_tokens":32000,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"},"messages":[{"role":"user","content":"synthetic"}] }`
	tr, calls := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) {
		got, _ := io.ReadAll(r.Body)
		if string(got) != raw {
			t.Error("same-model native wire rewritten")
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, strings.ReplaceAll(nativeOutput, "canonical", "claude-opus-5")); err != nil {
			t.Error(err)
		}
	})
	cfg := nativeConfig(tr)
	cfg.Limits.PolicyProfile = translate.PolicyAnthropicNative
	a, e := NewAnthropicAdapter(cfg)
	if e != nil {
		t.Fatal(e)
	}
	q := nativeQ(t)
	q.Entry.UpstreamID = "claude-opus-5"
	q.Body = []byte(raw)
	resp, e := a.Send(context.Background(), q)
	if e != nil || resp.StatusCode != 200 {
		t.Fatal(e, resp)
	}
	if cerr := resp.Body.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	q.Entry.UpstreamID = "unverified"
	resp, e = a.Send(context.Background(), q)
	if e != nil || resp.StatusCode != 400 || calls.Load() != 1 {
		t.Fatal("unknown profile model sent")
	}
	if cerr := resp.Body.Close(); cerr != nil {
		t.Fatal(cerr)
	}
}

func TestInputEstimateRejectsMalformedOutputProjection(t *testing.T) {
	for _, raw := range []string{`{"output_config":1}`, `{"output_config":{"format":"invalid"}}`} {
		if _, e := EstimateInputTokens(ModelEntry{}, []byte(raw)); e == nil {
			t.Fatal("invalid projection accepted")
		}
	}
}
func TestNativePolicyRequestRejectsMalformedBeforeWire(t *testing.T) {
	tr, calls := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid policy reached endpoint") })
	cfg := nativeConfig(tr)
	cfg.Limits.PolicyProfile = translate.PolicyAnthropicNative
	a, e := NewAnthropicAdapter(cfg)
	if e != nil {
		t.Fatal(e)
	}
	for _, extra := range []string{`,"thinking":null`, `,"output_config":{"effort":"high","unknown":true}`, `,"context_management":{}`, `,"output_config":{"format":{"type":"json_schema","schema":{}}}`} {
		q := nativeQ(t)
		q.Entry.UpstreamID = "claude-opus-5"
		q.Body = []byte(`{"max_tokens":1,"messages":[{"role":"user","content":"x"}]` + extra + `}`)
		r, e := a.Send(context.Background(), q)
		if e != nil || r.StatusCode != 400 {
			t.Fatal(e, r)
		}
		if cerr := r.Body.Close(); cerr != nil {
			t.Fatal(cerr)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid native requests sent")
	}
}
