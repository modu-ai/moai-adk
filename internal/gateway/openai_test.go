package gateway

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

const oaiInput = `{"model":"gpt-5.6-sol","max_tokens":10,"messages":[{"role":"user","content":"hello"}]}`
const oaiOutput = `{"id":"r","model":"gpt-5.6-sol","status":"completed","output":[{"id":"m","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"hello"}]}],"usage":{"input_tokens":1,"output_tokens":2}}`

func oaiEntry(method AuthMethod) ModelEntry {
	return ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: ProviderOpenAI, AuthMethod: method, Capabilities: Capabilities{ContextTokens: 100, Tools: true, Streaming: true}}
}
func oaiTLS(t *testing.T, h http.HandlerFunc) (*http.Transport, *atomic.Int32) {
	t.Helper()
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	tr := srv.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	calls := &atomic.Int32{}
	cfg := tr.TLSClientConfig.Clone()
	tr.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		calls.Add(1)
		if address != "api.openai.com:443" && address != "chatgpt.com:443" {
			return nil, errors.New("foreign endpoint")
		}
		return (&tls.Dialer{Config: cfg}).DialContext(ctx, "tcp", srv.Listener.Addr().String())
	}
	return tr, calls
}
func oaiConfig(tr *http.Transport) OpenAIConfig {
	return OpenAIConfig{Transport: tr, Limits: translate.Limits{MaxBodyBytes: 1 << 20, MaxOutputBytes: 1 << 20, MaxEventBytes: 1 << 16}, MeasureInput: func(ModelEntry, []byte) (int64, error) { return 1, nil }, SendOptions: auth.SendOptions{WriteTimeout: time.Second, PollInterval: 5 * time.Millisecond, MaxBodyBytes: 1 << 20}}
}
func oaiRequest(t *testing.T) RoutedRequest {
	t.Helper()
	ref, e := auth.NewAPIKey("synthetic-api-secret")
	if e != nil {
		t.Fatal(e)
	}
	gen, e := ref.Generation()
	if e != nil {
		t.Fatal(e)
	}
	return RoutedRequest{Entry: oaiEntry(AuthAPIKey), Body: []byte(oaiInput), Credential: ref, Generation: gen}
}
func oaiRead(t *testing.T, r *http.Response, e error) string {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
	if r == nil {
		t.Fatal("nil response")
	}
	defer func() { _ = r.Body.Close() }() // body fully read below; no write-back to lose
	b, e := io.ReadAll(r.Body)
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func TestOpenAIAPIKeyEndpointAndPublicResponse(t *testing.T) {
	seen := make(chan *http.Request, 1)
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var v map[string]any
		_ = json.Unmarshal(b, &v)
		if v["model"] != "gpt-5.6-sol" || v["max_output_tokens"] != float64(10) || v["store"] != false || v["truncation"] != "disabled" {
			t.Error("translation", string(b))
		}
		seen <- r
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, oaiOutput); err != nil {
			t.Error(err)
		}
	})
	a, e := NewOpenAIAdapter(oaiConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	q := oaiRequest(t)
	q.Headers = http.Header{"Authorization": {"foreign"}, "X-Api-Key": {"foreign"}}
	r, e := a.Send(context.Background(), q)
	body := oaiRead(t, r, e)
	if !strings.Contains(body, `"stop_reason":"end_turn"`) {
		t.Fatal(body)
	}
	got := <-seen
	if got.Host != "api.openai.com" || got.URL.Path != "/v1/responses" || got.Header.Get("Authorization") != "Bearer synthetic-api-secret" || got.Header.Get("X-Api-Key") != "" || calls.Load() != 1 {
		t.Fatal("endpoint/header boundary")
	}
}
func TestOpenAILocalRejectionsNeverSend(t *testing.T) {
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected send") })
	for _, kind := range []string{"policy", "measurement absent", "measurement error", "context absent", "context overflow", "wrong provider", "changed generation", "stream capability", "tool capability", "unknown auth", "subscription absent"} {
		cfg := oaiConfig(tr)
		q := oaiRequest(t)
		switch kind {
		case "policy":
			q.Body = []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"thinking":{},"max_tokens":10`, 1))
		case "measurement absent":
			cfg.MeasureInput = nil
		case "measurement error":
			cfg.MeasureInput = func(ModelEntry, []byte) (int64, error) { return 0, errors.New("private") }
		case "context absent":
			q.Entry.Capabilities.ContextTokens = 0
		case "context overflow":
			cfg.MeasureInput = func(ModelEntry, []byte) (int64, error) { return 101, nil }
		case "wrong provider":
			q.Entry.Provider = ProviderZAI
		case "changed generation":
			q.Generation++
		case "stream capability":
			q.Entry.Capabilities.Streaming = false
			q.Body = []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"stream":true,"max_tokens":10`, 1))
		case "tool capability":
			q.Entry.Capabilities.Tools = false
			q.Body = []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"tools":[{"name":"t","input_schema":{}}],"max_tokens":10`, 1))
		case "unknown auth":
			q.Entry.AuthMethod = "unknown"
		case "subscription absent":
			q.Entry.AuthMethod = AuthPKCE
		}
		a, e := NewOpenAIAdapter(cfg)
		if e != nil {
			t.Fatal(e)
		}
		r, e := a.Send(context.Background(), q)
		body := oaiRead(t, r, e)
		if r.StatusCode < 400 || r.StatusCode >= 500 || strings.Contains(body, "private") {
			t.Fatal(kind, r.StatusCode, body)
		}
	}
	if calls.Load() != 0 {
		t.Fatal(calls.Load())
	}
}
func TestOpenAIStatusAndNoRedirect(t *testing.T) {
	for _, status := range []int{401, 429, 302, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "12")
				w.Header().Set("Location", "https://foreign.invalid/steal")
				w.WriteHeader(status)
				if _, err := io.WriteString(w, `{"secret":"synthetic-api-secret"}`); err != nil {
					t.Error(err)
				}
			})
			a, _ := NewOpenAIAdapter(oaiConfig(tr))
			r, e := a.Send(context.Background(), oaiRequest(t))
			body := oaiRead(t, r, e)
			want := status
			wantCalls := int32(1)
			if status == 302 {
				want = 502
			}
			if status >= 500 && status <= 599 {
				// 5xx is retryable at the pre-stream boundary (t697): the
				// always-failing stub exhausts the bounded attempts.
				wantCalls = 3
			}
			if r.StatusCode != want || strings.Contains(body, "secret") || r.Header.Get("Location") != "" || calls.Load() != wantCalls {
				t.Fatal(r.StatusCode, body, calls.Load())
			}
			if (status == 429 || status == 503) && r.Header.Get("Retry-After") != "12" {
				t.Fatal("retry-after lost")
			}
		})
	}
}

type oaiBroker func(context.Context, string, bool) error

func (f oaiBroker) Run(c context.Context, h string, r bool) error { return f(c, h, r) }
func oaiStore(t *testing.T) (*auth.Store, CredentialRef, uint64) {
	t.Helper()
	dir, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	s, e := auth.OpenStore(filepath.Join(dir, "auth"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = s.Close() })
	gen, e := s.Login(context.Background(), oaiBroker(func(ctx context.Context, home string, refresh bool) error {
		claims, _ := json.Marshal(map[string]any{"exp": time.Now().Add(time.Hour).Unix()})
		v := map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": "e30." + base64.RawURLEncoding.EncodeToString(claims) + ".synthetic", "refresh_token": "synthetic-refresh", "id_token": "synthetic-id", "account_id": "synthetic-account"}}
		b, _ := json.Marshal(v)
		return os.WriteFile(filepath.Join(home, "auth.json"), b, 0600)
	}))
	if e != nil {
		t.Fatal(e)
	}
	ref, e := s.Resolve()
	if e != nil {
		t.Fatal(e)
	}
	return s, ref, gen
}
func TestOpenAISubscriptionUsesStoreAuthorizedBoundary(t *testing.T) {
	s, ref, gen := oaiStore(t)
	seen := make(chan string, 1)
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("request JSON")
		}
		if _, present := body["truncation"]; present {
			t.Error("subscription gained unverified truncation field")
		}
		seen <- r.Host + r.URL.Path + " " + r.Header.Get("ChatGPT-Account-Id")
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, oaiOutput); err != nil {
			t.Error(err)
		}
	})
	cfg := oaiConfig(tr)
	cfg.Subscription = s
	a, _ := NewOpenAIAdapter(cfg)
	q := RoutedRequest{Entry: oaiEntry(AuthPKCE), Body: []byte(oaiInput), Credential: ref, Generation: gen}
	r, e := a.Send(context.Background(), q)
	_ = oaiRead(t, r, e)
	if got := <-seen; got != "chatgpt.com/backend-api/codex/responses synthetic-account" {
		t.Fatal(got)
	}
	if _, e = s.Logout(context.Background()); e != nil {
		t.Fatal(e)
	}
	r, e = a.Send(context.Background(), q)
	_ = oaiRead(t, r, e)
	if r.StatusCode != 401 || calls.Load() != 1 {
		t.Fatal("logout crossed boundary")
	}
}
func TestOpenAIStreamEOFAndCloseJoin(t *testing.T) {
	for _, disconnect := range []bool{false, true} {
		tr, _ := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
			w.Header().Set("Content-Type", "text/event-stream")
			if _, err := io.WriteString(w, "data: {\"type\":\"response.created\",\"response\":{\"id\":\"r\",\"model\":\"gpt-5.6-sol\",\"status\":\"in_progress\"}}\n\n"); err != nil {
				t.Error(err)
			}
			w.(http.Flusher).Flush()
			if disconnect {
				select {
				case <-r.Context().Done():
				case <-time.After(2 * time.Second):
					t.Error("upstream cancellation missing")
				}
			}
		})
		a, _ := NewOpenAIAdapter(oaiConfig(tr))
		q := oaiRequest(t)
		q.Body = []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"stream":true,"max_tokens":10`, 1))
		r, e := a.Send(context.Background(), q)
		if e != nil {
			t.Fatal(e)
		}
		if disconnect {
			done := make(chan struct{})
			go func() { _ = r.Body.Close(); close(done) }()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("close did not join converter")
			}
		} else {
			b, e := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if e == nil || strings.Contains(string(b), "message_stop") {
				t.Fatal("EOF became success", string(b), e)
			}
		}
	}
}

func TestOpenAISubscriptionRejectsDifferentStore(t *testing.T) {
	s, _, _ := oaiStore(t)
	_, ref, gen := oaiStore(t)
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, oaiOutput); err != nil {
			t.Error(err)
		}
	})
	cfg := oaiConfig(tr)
	cfg.Subscription = s
	a, _ := NewOpenAIAdapter(cfg)
	r, e := a.Send(context.Background(), RoutedRequest{Entry: oaiEntry(AuthPKCE), Body: []byte(oaiInput), Credential: ref, Generation: gen})
	_ = oaiRead(t, r, e)
	if r.StatusCode != 401 || calls.Load() != 0 {
		t.Fatalf("foreign store accepted: status=%d calls=%d", r.StatusCode, calls.Load())
	}
}

type oaiChangingRef struct {
	CredentialRef
	generation uint64
	mutate     bool
}

func (r *oaiChangingRef) Generation() (uint64, error) { return r.generation, nil }
func (r *oaiChangingRef) Apply(q *http.Request) error {
	if r.mutate {
		q.URL.Host = "foreign.invalid"
	} else {
		r.generation++
	}
	return nil
}
func TestOpenAIRechecksGenerationAndDestinationAfterApply(t *testing.T) {
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected send") })
	a, _ := NewOpenAIAdapter(oaiConfig(tr))
	for _, mutate := range []bool{false, true} {
		q := oaiRequest(t)
		q.Credential = &oaiChangingRef{CredentialRef: q.Credential, generation: q.Generation, mutate: mutate}
		r, e := a.Send(context.Background(), q)
		_ = oaiRead(t, r, e)
		if r.StatusCode != 401 {
			t.Fatal(r.StatusCode)
		}
	}
	if calls.Load() != 0 {
		t.Fatal(calls.Load())
	}
}

func TestOpenAIErrorStatusSurvivesServer(t *testing.T) {
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) { t.Error("local policy reached upstream") })
	a, _ := NewOpenAIAdapter(oaiConfig(tr))
	q := oaiRequest(t)
	catalog, e := NewCatalog([]ModelEntry{q.Entry})
	if e != nil {
		t.Fatal(e)
	}
	srv, e := NewServer(ServerConfig{SessionHeader: "X-Session", SessionToken: "synthetic-session", MaxBodyBytes: 1 << 20, Catalog: catalog, Adapters: map[ProviderID]Adapter{ProviderOpenAI: a}, ResolveCredential: func(context.Context, ModelEntry) (CredentialRef, error) { return q.Credential, nil }})
	if e != nil {
		t.Fatal(e)
	}
	body := strings.Replace(oaiInput, `"max_tokens":10`, `"thinking":{},"max_tokens":10`, 1)
	req := httptest.NewRequest("POST", "http://local/v1/messages", strings.NewReader(body))
	req.Header.Set("X-Session", "synthetic-session")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "invalid_request_error") || calls.Load() != 0 {
		t.Fatal(rec.Code, rec.Body.String())
	}
}
func oaiSSE() string {
	events := []string{
		`{"type":"response.created","response":{"id":"r","model":"gpt-5.6-sol","status":"in_progress"}}`,
		`{"type":"response.output_item.added","output_index":0,"item":{"id":"m","type":"message","role":"assistant","status":"in_progress","content":[]}}`,
		`{"type":"response.content_part.added","output_index":0,"item_id":"m","content_index":0,"part":{"type":"output_text","text":""}}`,
		`{"type":"response.output_text.delta","output_index":0,"item_id":"m","content_index":0,"delta":"hello"}`,
		`{"type":"response.output_text.done","output_index":0,"item_id":"m","content_index":0,"text":"hello"}`,
		`{"type":"response.content_part.done","output_index":0,"item_id":"m","content_index":0,"part":{"type":"output_text","text":"hello"}}`,
		`{"type":"response.output_item.done","output_index":0,"item":{"id":"m","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"hello"}]}}`,
		`{"type":"response.completed","response":` + oaiOutput + `}`}
	var b strings.Builder
	for _, e := range events {
		b.WriteString("data: " + e + "\n\n")
	}
	return b.String()
}
func TestOpenAIStreamSuccessAndContextCancellation(t *testing.T) {
	for _, cancelEarly := range []bool{false, true} {
		tr, _ := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close() // server-side discard; assertions read the client side
			w.Header().Set("Content-Type", "text/event-stream")
			if _, err := io.WriteString(w, oaiSSE()); err != nil {
				t.Error(err)
			}
		})
		a, _ := NewOpenAIAdapter(oaiConfig(tr))
		q := oaiRequest(t)
		q.Body = []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"stream":true,"max_tokens":10`, 1))
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if cancelEarly {
			cancel()
		}
		r, e := a.Send(ctx, q)
		if cancelEarly {
			if !errors.Is(e, context.Canceled) || r != nil {
				t.Fatal("cancellation ignored")
			}
			continue
		}
		body := oaiRead(t, r, e)
		cancel()
		if strings.Count(body, "event: message_stop") != 1 || !strings.Contains(body, `"stop_reason":"end_turn"`) {
			t.Fatal(body)
		}
	}
}

func TestOpenAIMalformedUpstreamAndConfiguration(t *testing.T) {
	for _, cfg := range []OpenAIConfig{{}, {Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}, {Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}} {
		if _, e := NewOpenAIAdapter(cfg); e == nil {
			t.Fatal("unsafe transport accepted")
		}
	}
	for _, kind := range []string{"mime", "json", "size", "truncated", "stream mime", "opaque", "invalid retry"} {
		tr, _ := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch kind {
			case "mime":
				w.Header().Set("Content-Type", "invalid")
				if _, err := io.WriteString(w, oaiOutput); err != nil {
					t.Error(err)
				}
			case "json":
				if _, err := io.WriteString(w, "{"); err != nil {
					t.Error(err)
				}
			case "size":
				if _, err := io.WriteString(w, strings.Repeat("x", 101)); err != nil {
					t.Error(err)
				}
			case "truncated":
				w.Header().Set("Content-Length", "1000")
				if _, err := io.WriteString(w, "{"); err != nil {
					t.Error(err)
				}
			case "stream mime":
				if _, err := io.WriteString(w, oaiOutput); err != nil {
					t.Error(err)
				}
			case "opaque":
				if _, err := io.WriteString(w, strings.Replace(oaiOutput, `"type":"message"`, `"type":"reasoning"`, 1)); err != nil {
					t.Error(err)
				}
			case "invalid retry":
				w.Header().Set("Retry-After", "synthetic-secret")
				w.WriteHeader(429)
			}
		})
		cfg := oaiConfig(tr)
		if kind == "size" {
			cfg.Limits.MaxOutputBytes = 100
		}
		a, _ := NewOpenAIAdapter(cfg)
		q := oaiRequest(t)
		if kind == "stream mime" {
			q.Body = []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"stream":true,"max_tokens":10`, 1))
		}
		r, e := a.Send(context.Background(), q)
		body := oaiRead(t, r, e)
		want := 502
		if kind == "invalid retry" {
			want = 429
		}
		if r.StatusCode != want || r.Header.Get("Retry-After") != "" || strings.Contains(body, "synthetic") {
			t.Fatal(kind, r.StatusCode, body)
		}
	}
	tr := &http.Transport{DialTLSContext: func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("synthetic-private-transport")
	}}
	a, _ := NewOpenAIAdapter(oaiConfig(tr))
	r, e := a.Send(context.Background(), oaiRequest(t))
	body := oaiRead(t, r, e)
	// t697: on bounded-retry exhaustion the underlying transport reason is
	// delivered to the session-authenticated local client instead of a bare
	// "Bad Gateway".
	if r.StatusCode != 502 || !strings.Contains(body, "gateway upstream connection failed after 3 attempts") || !strings.Contains(body, "synthetic-private-transport") {
		t.Fatal(r.StatusCode, body)
	}
	for _, value := range []string{"", strings.Repeat("x", 65), "1\r\nsecret", "-1", "secret"} {
		if validRetryAfter(value) != "" {
			t.Fatal("bad retry")
		}
	}
	if validRetryAfter("Wed, 21 Oct 2015 07:28:00 GMT") != "Wed, 21 Oct 2015 07:28:00 GMT" {
		t.Fatal("date retry lost")
	}
}
