package gateway

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// retryTLS mirrors nativeTLS with an injected dial-failure budget: a positive
// budget fails exactly that many dials (the stale-connection reproduction)
// before letting traffic through. The returned counter records every dial, so
// a test can prove each attempt opened a fresh connection.
func retryTLS(t *testing.T, h http.HandlerFunc, dialFailures *atomic.Int32) (*http.Transport, *atomic.Int32) {
	t.Helper()
	s := httptest.NewTLSServer(h)
	t.Cleanup(s.Close)
	tr := s.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	cfg := tr.TLSClientConfig.Clone()
	calls := &atomic.Int32{}
	tr.DialTLSContext = func(ctx context.Context, n, addr string) (net.Conn, error) {
		calls.Add(1)
		if addr != "api.anthropic.com:443" {
			t.Errorf("unexpected destination %s", addr)
		}
		if dialFailures.Add(-1) >= 0 {
			return nil, errors.New("dial tcp: connection reset by peer")
		}
		return (&tls.Dialer{Config: cfg}).DialContext(ctx, "tcp", s.Listener.Addr().String())
	}
	return tr, calls
}

// dialBudget returns a dial-failure budget that fails exactly n dials before
// letting traffic through.
func dialBudget(n int32) *atomic.Int32 {
	b := &atomic.Int32{}
	b.Store(n)
	return b
}

// gatewayErrorBody parses the canonical synthetic error body the adapters
// render for failed egress.
func gatewayErrorBody(t *testing.T, body string) (kind, message string) {
	t.Helper()
	var parsed struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(body), &parsed) != nil {
		t.Fatalf("gateway error body is not the canonical shape: %s", body)
	}
	return parsed.Error.Type, parsed.Error.Message
}

func nativeHandler(status int, payload string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// A write error is expected when a retry attempt aborts the connection
		// mid-response; only the final attempt's bytes reach the assertions.
		_, _ = io.WriteString(w, payload)
	}
}

// Regression (t697): a stub upstream 5xx once followed by 200 must succeed on
// the pre-stream retry, each attempt on a fresh connection.
func TestUpstreamRetryRecoversSingleFiveHundred(t *testing.T) {
	var seen atomic.Int32
	tr, dials := retryTLS(t, func(w http.ResponseWriter, r *http.Request) {
		if seen.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, nativeOutput); err != nil {
			t.Errorf("write response: %v", err)
		}
	}, dialBudget(0))
	a, e := NewAnthropicAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Send(context.Background(), nativeQ(t))
	body := oaiRead(t, r, e)
	if body != nativeOutput {
		t.Fatalf("retry did not recover from a single upstream 502: %s", body)
	}
	if dials.Load() != 2 {
		t.Fatalf("upstream connections = %d, want 2 (fresh connection per attempt)", dials.Load())
	}
}

// Regression (t697): a persistently failing upstream exhausts the bounded
// retries and the rendered 502 names the upstream origin and the attempt
// count, distinct from a gateway connection failure.
func TestUpstreamRetryExhaustionLabelsUpstreamOrigin(t *testing.T) {
	tr, dials := retryTLS(t, nativeHandler(http.StatusBadGateway, ""), dialBudget(0))
	a, e := NewAnthropicAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Send(context.Background(), nativeQ(t))
	if e != nil {
		t.Fatal(e)
	}
	if r.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", r.StatusCode)
	}
	kind, message := gatewayErrorBody(t, oaiRead(t, r, nil))
	if kind != "api_error" {
		t.Fatalf("error kind = %q, want api_error", kind)
	}
	if !strings.Contains(message, "upstream returned 502") || !strings.Contains(message, "after 3 attempts") {
		t.Fatalf("exhaustion message missing upstream origin label: %q", message)
	}
	if strings.Contains(message, "gateway upstream connection failed") {
		t.Fatalf("upstream-origin 502 mislabeled as gateway connection failure: %q", message)
	}
	if dials.Load() != 3 {
		t.Fatalf("upstream connections = %d, want 3 (bounded retry)", dials.Load())
	}
}

// Regression (t697): a dropped connection on the first dial recovers on the
// retry; a connection that never comes back exhausts with the underlying
// reason delivered to the client.
func TestUpstreamRetryDialFailureRecovers(t *testing.T) {
	tr, dials := retryTLS(t, nativeHandler(http.StatusOK, nativeOutput), dialBudget(1))
	a, e := NewAnthropicAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Send(context.Background(), nativeQ(t))
	body := oaiRead(t, r, e)
	if body != nativeOutput {
		t.Fatalf("retry did not recover from a dropped connection: %s", body)
	}
	if dials.Load() != 2 {
		t.Fatalf("upstream connections = %d, want 2", dials.Load())
	}
}

func TestUpstreamRetryDialExhaustionCarriesReason(t *testing.T) {
	tr, dials := retryTLS(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected send: every dial must fail")
	}, dialBudget(5))
	a, e := NewAnthropicAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Send(context.Background(), nativeQ(t))
	if e != nil {
		t.Fatal(e)
	}
	if r.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", r.StatusCode)
	}
	kind, message := gatewayErrorBody(t, oaiRead(t, r, nil))
	if kind != "api_error" {
		t.Fatalf("error kind = %q, want api_error", kind)
	}
	if !strings.Contains(message, "gateway upstream connection failed after 3 attempts") || !strings.Contains(message, "connection reset by peer") {
		t.Fatalf("exhaustion message missing gateway origin and underlying reason: %q", message)
	}
	if dials.Load() != 3 {
		t.Fatalf("upstream connections = %d, want 3 (bounded retry)", dials.Load())
	}
}

// The idempotency boundary is hard: once a response has started, the request
// is never re-driven, even when the body dies mid-flight.
func TestUpstreamNoRetryAfterResponseStarted(t *testing.T) {
	tr, dials := retryTLS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, nativeOutput[:20]); err != nil {
			t.Errorf("write response: %v", err)
		}
		w.(http.Flusher).Flush()
		panic(http.ErrAbortHandler)
	}, dialBudget(0))
	a, e := NewAnthropicAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Send(context.Background(), nativeQ(t))
	if e != nil {
		t.Fatal(e)
	}
	if r.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", r.StatusCode)
	}
	if dials.Load() != 1 {
		t.Fatalf("upstream connections = %d, want 1 (no retry after the response started)", dials.Load())
	}
}

// Deterministic client errors are terminal on the first attempt.
func TestUpstreamFourHundredIsNotRetried(t *testing.T) {
	tr, dials := retryTLS(t, nativeHandler(http.StatusBadRequest, ""), dialBudget(0))
	a, e := NewAnthropicAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Send(context.Background(), nativeQ(t))
	if e != nil {
		t.Fatal(e)
	}
	if r.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", r.StatusCode)
	}
	kind, _ := gatewayErrorBody(t, oaiRead(t, r, nil))
	if kind != "invalid_request_error" {
		t.Fatalf("error kind = %q, want invalid_request_error", kind)
	}
	if dials.Load() != 1 {
		t.Fatalf("upstream connections = %d, want 1 (4xx is never retried)", dials.Load())
	}
}

// The OpenAI adapter shares the same egress retry wiring.
func TestOpenAIUpstreamRetryWiring(t *testing.T) {
	t.Run("recovers", func(t *testing.T) {
		var seen atomic.Int32
		tr, dials := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
			if seen.Add(1) == 1 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := io.WriteString(w, oaiOutput); err != nil {
				t.Errorf("write response: %v", err)
			}
		})
		a, e := NewOpenAIAdapter(oaiConfig(tr))
		if e != nil {
			t.Fatal(e)
		}
		r, e := a.Send(context.Background(), oaiRequest(t))
		body := oaiRead(t, r, e)
		// The API-key path converts the upstream payload to the native shape.
		converted := `{"content":[{"text":"hello","type":"text"}],"id":"r","model":"gpt-5.6-sol","role":"assistant","stop_reason":"end_turn","stop_sequence":null,"type":"message","usage":{"input_tokens":1,"output_tokens":2}}`
		if body != converted {
			t.Fatalf("retry did not recover from a single upstream 502: %s", body)
		}
		if dials.Load() != 2 {
			t.Fatalf("upstream connections = %d, want 2", dials.Load())
		}
	})
	t.Run("labels exhaustion", func(t *testing.T) {
		tr, dials := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		})
		a, e := NewOpenAIAdapter(oaiConfig(tr))
		if e != nil {
			t.Fatal(e)
		}
		r, e := a.Send(context.Background(), oaiRequest(t))
		if e != nil {
			t.Fatal(e)
		}
		if r.StatusCode != http.StatusBadGateway {
			t.Fatalf("status = %d, want 502", r.StatusCode)
		}
		_, message := gatewayErrorBody(t, oaiRead(t, r, nil))
		if !strings.Contains(message, "upstream returned 502") || !strings.Contains(message, "after 3 attempts") {
			t.Fatalf("exhaustion message missing upstream origin label: %q", message)
		}
		if dials.Load() != 3 {
			t.Fatalf("upstream connections = %d, want 3", dials.Load())
		}
	})
}
