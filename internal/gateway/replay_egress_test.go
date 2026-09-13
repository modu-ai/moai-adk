package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"io"
	"net/http"
	"strings"
	"testing"
)

type auditHistory struct{}

func (auditHistory) Check(context.Context, string, string, []byte) error   { return nil }
func (auditHistory) Publish(context.Context, string, string, []byte) error { return nil }
func TestRegressionNativeEgressPreservesOpaqueItemBytes(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) { regressionNativeEgress(t, stream) })
	}
}
func regressionNativeEgress(t *testing.T, stream bool) {
	raw := `{ "type":"reasoning", "id":"rs_audit", "summary":[], "encrypted_content":"synthetic_cipher" }`
	env, e := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(raw)}})
	if e != nil {
		t.Fatal(e)
	}
	input := map[string]any{"model": "gpt-5.6-sol", "max_tokens": 10, "stream": stream, "messages": []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "redacted_thinking", "data": env.Data()}, map[string]any{"type": "text", "text": "answer"}}}, map[string]any{"role": "user", "content": "continue"}}}
	body, _ := json.Marshal(input)
	// The handler goroutine must hand the observed bytes through a channel: a
	// shared variable would be a data race against the test goroutine's read.
	sent := make(chan string, 1)
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sent <- string(b)
		w.Header().Set("Content-Type", "text/event-stream")
		if _, err := io.WriteString(w, oaiSSE()); err != nil {
			t.Error(err)
		}
	})
	store, ref, gen := oaiStore(t)
	cfg := oaiConfig(tr)
	cfg.Subscription = store
	cfg.Limits.History = auditHistory{}
	adapter, e := NewOpenAIAdapter(cfg)
	if e != nil {
		t.Fatal(e)
	}
	resp, e := adapter.Send(context.Background(), RoutedRequest{Entry: oaiEntry(AuthPKCE), Body: body, Credential: ref, Generation: gen})
	respBody := oaiRead(t, resp, e)
	if calls.Load() != 1 || resp.StatusCode != 200 {
		// The collected body is the only diagnostic the 502 short-circuit in
		// the adapter carries; surface it verbatim for CI failure logs.
		t.Fatalf("probe failed status=%d calls=%d body=%s", resp.StatusCode, calls.Load(), respBody)
	}
	select {
	case observed := <-sent:
		t.Logf("observed subscription request: %s", observed)
		if !strings.Contains(observed, raw) {
			t.Fatal("opaque item original bytes lost at subscription egress")
		}
	default:
		t.Fatal("subscription request was not observed")
	}
}
