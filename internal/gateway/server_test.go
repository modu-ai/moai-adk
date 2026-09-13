package gateway

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
)

type testCredential struct {
	provider   ProviderID
	generation atomic.Uint64
	change     bool
	absent     bool
}

func (c *testCredential) Provider() ProviderID { return c.provider }
func (c *testCredential) Generation() (uint64, error) {
	if c.absent {
		return 0, auth.ErrCredentialAbsent
	}
	if c.change {
		return c.generation.Add(1), nil
	}
	return c.generation.Load(), nil
}
func (c *testCredential) Apply(*http.Request) error { return nil }
func (c *testCredential) Redacted() string          { return "test:credential" }

type testAdapter struct {
	calls  atomic.Int64
	body   string
	status int
	err    error
	check  func(RoutedRequest)
}

func (a *testAdapter) Send(_ context.Context, r RoutedRequest) (*http.Response, error) {
	a.calls.Add(1)
	if a.check != nil {
		a.check(r)
	}
	if a.err != nil {
		return nil, a.err
	}
	status := a.status
	if status == 0 {
		status = 200
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(a.body))}, nil
}

// Stored-credential fixtures intentionally retain explicit API-key rows; native
// request-scoped subscription behavior is exercised by the OAuth tests.
func testStoredCredentialCatalog(glm []string) (CatalogSnapshot, error) {
	c, err := NewSessionCatalog(glm)
	if err != nil {
		return c, err
	}
	entries := c.Entries()
	for i := range entries {
		if entries[i].Provider == ProviderAnthropic {
			entries[i].AuthMethod = AuthAPIKey
		}
	}
	return NewCatalog(entries)
}
func fixtureServer(t *testing.T, limit int64, ref CredentialRef, a *testAdapter) *Server {
	t.Helper()
	catalog, err := testStoredCredentialCatalog([]string{"glm-test"})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "local-secret", MaxBodyBytes: limit, Catalog: catalog, ResolveCredential: func(context.Context, ModelEntry) (CredentialRef, error) { return ref, nil }, Adapters: map[ProviderID]Adapter{ProviderOpenAI: a, ProviderZAI: a, ProviderAnthropic: a}})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func request(s *Server, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("X-Test-Session", "local-secret")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

const validationBody = `{"model":"gpt-6-astra","max_tokens":1,"messages":[{"role":"user","content":"hello"}]}`
const turnBody = `{"model":"gpt-6-astra","max_tokens":2,"messages":[{"role":"user","content":"hello"}]}`

type forbiddenBody struct{ reads atomic.Int64 }

func (b *forbiddenBody) Read([]byte) (int, error) {
	b.reads.Add(1)
	return 0, errors.New("body must not be read")
}
func (*forbiddenBody) Close() error { return nil }
func TestSessionAuthPrecedesBodyAndPath(t *testing.T) {
	a := &testAdapter{}
	s := fixtureServer(t, 4096, &testCredential{provider: ProviderOpenAI}, a)
	for _, token := range []string{"", "wrong", "local-secret,other"} {
		body := &forbiddenBody{}
		r := httptest.NewRequest("POST", "/v1/unknown-path", nil)
		r.Body = body
		r.Header.Set("X-Test-Session", token)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != 401 || body.reads.Load() != 0 || a.calls.Load() != 0 {
			t.Fatalf("auth got status=%d reads=%d calls=%d", w.Code, body.reads.Load(), a.calls.Load())
		}
	}
}
func TestIngressValidationStaysLocal(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		ref        CredentialRef
		code       int
		errorType  string
	}{{"unknown", strings.ReplaceAll(validationBody, "gpt-6-astra", "gpt-6-astro"), nil, 404, "not_found_error"}, {"absent", validationBody, nil, 401, "authentication_error"}, {"credential absent", validationBody, &testCredential{provider: ProviderOpenAI, absent: true}, 401, "authentication_error"}, {"present", validationBody, &testCredential{provider: ProviderOpenAI}, 200, ""}} {
		t.Run(tc.name, func(t *testing.T) {
			a := &testAdapter{}
			s := fixtureServer(t, 4096, tc.ref, a)
			w := request(s, "POST", "/v1/messages?beta=true", tc.body)
			if w.Code != tc.code || a.calls.Load() != 0 {
				t.Fatalf("status %d body %s calls %d", w.Code, w.Body, a.calls.Load())
			}
			if tc.errorType != "" && !strings.Contains(w.Body.String(), tc.errorType) {
				t.Fatal(w.Body.String())
			}
			if tc.code == 200 && (!strings.Contains(w.Body.String(), `"type":"message"`) || !strings.Contains(w.Body.String(), `"stop_reason":"end_turn"`)) {
				t.Fatal(w.Body.String())
			}
		})
	}
}
func TestIngressRejectsInvalidAndUnsupportedWithoutEgress(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body string
		code                     int
	}{{"unknown path", "POST", "/v1/complete", turnBody, 404}, {"unknown path2", "POST", "/v1/unknown-path", turnBody, 404}, {"unknown method", "PUT", "/v1/messages", turnBody, 405}, {"query", "POST", "/v1/messages?target=https://evil.invalid", turnBody, 400}, {"duplicate query", "POST", "/v1/messages?beta=true&beta=true", turnBody, 400}, {"encoded path", "POST", "/v1/%6dessages", turnBody, 404}, {"duplicate model", "POST", "/v1/messages", `{"model":"gpt-6-astra","model":"glm-test"}`, 400}, {"malformed", "POST", "/v1/messages", `{`, 400}, {"missing model", "POST", "/v1/messages", `{}`, 400}, {"unsupported model", "POST", "/v1/messages", strings.ReplaceAll(turnBody, "gpt-6-astra", "gpt-5.3"), 404}, {"too large", "POST", "/v1/messages", strings.Repeat("x", 513), 413}} {
		t.Run(tc.name, func(t *testing.T) {
			a := &testAdapter{}
			s := fixtureServer(t, 512, &testCredential{provider: ProviderOpenAI}, a)
			w := request(s, tc.method, tc.path, tc.body)
			if w.Code != tc.code || a.calls.Load() != 0 {
				t.Fatalf("status %d want %d calls %d body %s", w.Code, tc.code, a.calls.Load(), w.Body)
			}
		})
	}
}
func TestIngressGzipBounds(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		malformed  bool
		code       int
	}{{"valid", validationBody, false, 200}, {"expanded too large", strings.Repeat("x", 2049), false, 413}, {"malformed", "not gzip", true, 400}} {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			if tc.malformed {
				b.WriteString(tc.body)
			} else {
				z := gzip.NewWriter(&b)
				_, _ = z.Write([]byte(tc.body))
				_ = z.Close()
			}
			a := &testAdapter{}
			s := fixtureServer(t, 2048, &testCredential{provider: ProviderOpenAI}, a)
			r := httptest.NewRequest("POST", "/v1/messages", &b)
			r.Header.Set("X-Test-Session", "local-secret")
			r.Header.Set("Content-Encoding", "gzip")
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			if w.Code != tc.code || a.calls.Load() != 0 {
				t.Fatalf("status %d calls %d body %s", w.Code, a.calls.Load(), w.Body)
			}
		})
	}
}
func TestIngressGenerationAndProviderGuard(t *testing.T) {
	for _, ref := range []CredentialRef{nil, &testCredential{provider: ProviderZAI}, &testCredential{provider: ProviderOpenAI, change: true}} {
		a := &testAdapter{}
		s := fixtureServer(t, 4096, ref, a)
		w := request(s, "POST", "/v1/messages", turnBody)
		if w.Code != 401 || a.calls.Load() != 0 {
			t.Fatalf("status %d calls %d", w.Code, a.calls.Load())
		}
	}
}
func TestIngressAdapterNoFallback(t *testing.T) {
	for _, status := range []int{200, 401, 429, 503} {
		a := &testAdapter{body: `{"result":"selected"}`, status: status}
		s := fixtureServer(t, 4096, &testCredential{provider: ProviderOpenAI}, a)
		w := request(s, "POST", "/v1/messages", turnBody)
		if w.Code != status || a.calls.Load() != 1 || w.Body.String() != a.body {
			t.Fatalf("status %d calls %d body %s", w.Code, a.calls.Load(), w.Body)
		}
	}
	a := &testAdapter{err: errors.New("private credential should not escape")}
	s := fixtureServer(t, 4096, &testCredential{provider: ProviderOpenAI}, a)
	w := request(s, "POST", "/v1/messages", turnBody)
	if w.Code != 502 || strings.Contains(w.Body.String(), "private credential") || a.calls.Load() != 1 {
		t.Fatal(w.Body.String())
	}
}

func TestModelsAndCountAreLocalAndExplicit(t *testing.T) {
	a := &testAdapter{}
	s := fixtureServer(t, 4096, nil, a)
	for _, path := range []string{"/v1/models", "/v1/models?limit=1000"} {
		w := request(s, "GET", path, "")
		if w.Code != 200 || a.calls.Load() != 0 {
			t.Fatalf("models %d %s", w.Code, w.Body)
		}
		for _, id := range []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna", "claude-opus-5", "claude-sonnet-5", "glm-test"} {
			if !strings.Contains(w.Body.String(), `"id":"`+id+`"`) {
				t.Fatalf("model %s missing: %s", id, w.Body)
			}
		}
	}
	for _, id := range []string{"gpt-6-astra", "claude-opus-5", "glm-test"} {
		body := strings.ReplaceAll(turnBody, "gpt-6-astra", id)
		w := request(s, "POST", "/v1/messages/count_tokens", body)
		again := request(s, "POST", "/v1/messages/count_tokens", body)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"accuracy":"estimate"`) || !strings.Contains(w.Body.String(), `"algorithm":"json-runes-div4"`) || w.Body.String() != again.Body.String() || a.calls.Load() != 0 {
			t.Fatalf("count %s: %d %s", id, w.Code, w.Body)
		}
	}
}
func TestServerRequiresExplicitPolicy(t *testing.T) {
	for _, cfg := range []ServerConfig{{}, {SessionHeader: "X-Key", SessionToken: "secret"}, {SessionHeader: "X-Key", SessionToken: "secret", MaxBodyBytes: -1}, {SessionHeader: "Bad Header", SessionToken: "secret", MaxBodyBytes: 1}} {
		if _, err := NewServer(cfg); err == nil {
			t.Fatal("missing policy accepted")
		}
	}
}

func TestConcurrentIngressRoutesAreRequestLocal(t *testing.T) {
	catalog, err := testStoredCredentialCatalog([]string{"glm-test"})
	if err != nil {
		t.Fatal(err)
	}
	adapters := map[ProviderID]Adapter{}
	for _, provider := range []ProviderID{ProviderOpenAI, ProviderAnthropic, ProviderZAI} {
		adapters[provider] = &testAdapter{body: string(provider), check: func(r RoutedRequest) {
			if r.Entry.Provider != provider || r.Credential.Provider() != provider {
				t.Errorf("provider crossed: %s -> %s", provider, r.Entry.Provider)
			}
		}}
	}
	s, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "local-secret", MaxBodyBytes: 4096, Catalog: catalog, ResolveCredential: func(_ context.Context, e ModelEntry) (CredentialRef, error) {
		return &testCredential{provider: e.Provider}, nil
	}, Adapters: adapters})
	if err != nil {
		t.Fatal(err)
	}
	// Mutating the caller's registration map must not change the session.
	delete(adapters, ProviderOpenAI)
	for _, id := range []string{"gpt-6-astra", "claude-opus-5", "glm-test"} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			for range 20 {
				w := request(s, "POST", "/v1/messages", strings.ReplaceAll(turnBody, "gpt-6-astra", id))
				entry, _ := catalog.Resolve(id)
				if w.Code != 200 || w.Body.String() != string(entry.Provider) {
					t.Fatalf("%s: %d %s", id, w.Code, w.Body)
				}
			}
		})
	}
}

func TestIngressCancelledAndUnavailableAdapter(t *testing.T) {
	a := &testAdapter{}
	s := fixtureServer(t, 4096, &testCredential{provider: ProviderOpenAI}, a)
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(turnBody))
	ctx, cancel := context.WithCancel(r.Context())
	cancel()
	r = r.WithContext(ctx)
	r.Header.Set("X-Test-Session", "local-secret")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 408 || a.calls.Load() != 0 {
		t.Fatalf("cancelled: %d calls %d", w.Code, a.calls.Load())
	}
	s.adapters = map[ProviderID]Adapter{}
	w = request(s, "POST", "/v1/messages", turnBody)
	if w.Code != 503 || a.calls.Load() != 0 {
		t.Fatalf("unavailable adapter: %d", w.Code)
	}
}

func TestIngressHeaderAndReadFailuresStayLocal(t *testing.T) {
	for _, scenario := range []string{"multiple session", "read error", "unsupported encoding", "invalid gzip checksum"} {
		t.Run(scenario, func(t *testing.T) {
			a := &testAdapter{}
			s := fixtureServer(t, 4096, &testCredential{provider: ProviderOpenAI}, a)
			r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(validationBody))
			r.Header.Set("X-Test-Session", "local-secret")
			code := 400
			switch scenario {
			case "multiple session":
				r.Header.Add("X-Test-Session", "local-secret")
				code = 401
			case "read error":
				r.Body = &forbiddenBody{}
			case "unsupported encoding":
				r.Header.Set("Content-Encoding", "br")
			case "invalid gzip checksum":
				var b bytes.Buffer
				z := gzip.NewWriter(&b)
				_, _ = z.Write([]byte(validationBody))
				_ = z.Close()
				data := b.Bytes()
				data[len(data)-1] ^= 1
				r.Body = io.NopCloser(bytes.NewReader(data))
				r.Header.Set("Content-Encoding", "gzip")
			}
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			if w.Code != code || a.calls.Load() != 0 {
				t.Fatalf("status %d calls %d", w.Code, a.calls.Load())
			}
		})
	}
}

func TestCredentialResolverFailuresNeverSend(t *testing.T) {
	for _, resolve := range []func(context.Context, ModelEntry) (CredentialRef, error){nil, func(context.Context, ModelEntry) (CredentialRef, error) { return nil, errors.New("private failure") }} {
		a := &testAdapter{}
		s := fixtureServer(t, 4096, nil, a)
		s.resolveCredential = resolve
		w := request(s, "POST", "/v1/messages", turnBody)
		if w.Code != 401 || a.calls.Load() != 0 || strings.Contains(w.Body.String(), "private failure") {
			t.Fatalf("resolver failure: %d %s", w.Code, w.Body)
		}
	}
}

func TestModelsContainsExactlySessionCatalog(t *testing.T) {
	a := &testAdapter{}
	s := fixtureServer(t, 4096, nil, a)
	w := request(s, "GET", "/v1/models?limit=1000", "")
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{}
	for _, entry := range s.catalog.Entries() {
		expected[entry.RouteID] = true
	}
	if len(result.Data) != len(expected) {
		t.Fatalf("model set count %d want %d", len(result.Data), len(expected))
	}
	for _, entry := range result.Data {
		if !expected[entry.ID] {
			t.Fatalf("extra or duplicate model %q", entry.ID)
		}
		delete(expected, entry.ID)
	}
	if len(expected) != 0 {
		t.Fatalf("missing models: %v", expected)
	}
}

func TestAdapterReceivesOnlyClonedProtocolHeaders(t *testing.T) {
	a := &testAdapter{body: "ok"}
	s := fixtureServer(t, 4096, &testCredential{provider: ProviderOpenAI}, a)
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(turnBody))
	for key, value := range map[string]string{"X-Test-Session": "local-secret", "Authorization": "Bearer private", "X-Api-Key": "private-key", "Anthropic-Version": "2023-06-01", "Anthropic-Beta": "test-beta", "X-Other-Secret": "hidden"} {
		r.Header.Set(key, value)
	}
	a.check = func(in RoutedRequest) {
		if len(in.Headers) != 2 || in.Headers.Get("Anthropic-Version") != "2023-06-01" || in.Headers.Get("Anthropic-Beta") != "test-beta" {
			t.Fatalf("protocol header selection: %v", in.Headers)
		}
		for _, key := range []string{"Authorization", "X-Api-Key", "X-Test-Session", "X-Other-Secret"} {
			if in.Headers.Get(key) != "" {
				t.Fatalf("secret header escaped: %s", key)
			}
		}
		in.Headers["Anthropic-Beta"][0] = "mutated"
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 || r.Header.Get("Anthropic-Beta") != "test-beta" {
		t.Fatalf("header alias or response: %d %s", w.Code, r.Header.Get("Anthropic-Beta"))
	}
}

func TestSessionHeaderIsNeverForwardedEvenWhenProtocolNamed(t *testing.T) {
	a := &testAdapter{body: "ok", check: func(r RoutedRequest) {
		if r.Headers.Get("Anthropic-Beta") != "" {
			t.Fatal("session token forwarded in protocol header")
		}
	}}
	s := fixtureServer(t, 4096, &testCredential{provider: ProviderOpenAI}, a)
	s.sessionHeader = "Anthropic-Beta"
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(turnBody))
	r.Header.Set("Anthropic-Beta", "local-secret")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}

func TestServerCancellationReachesCredentialResolver(t *testing.T) {
	catalog, err := testStoredCredentialCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	adapter := &testAdapter{}
	entered := make(chan struct{})
	resolved := make(chan error, 1)
	s, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "local-secret", MaxBodyBytes: 4096, Catalog: catalog, Adapters: map[ProviderID]Adapter{ProviderOpenAI: adapter}, ResolveCredential: func(ctx context.Context, _ ModelEntry) (CredentialRef, error) {
		close(entered)
		<-ctx.Done()
		resolved <- ctx.Err()
		return nil, ctx.Err()
	}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(turnBody)).WithContext(ctx)
	r.Header.Set("X-Test-Session", "local-secret")
	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { defer close(done); s.ServeHTTP(w, r) }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("resolver not entered")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after cancellation")
	}
	if err := <-resolved; !errors.Is(err, context.Canceled) {
		t.Fatalf("resolver error %v", err)
	}
	if adapter.calls.Load() != 0 || w.Code != http.StatusUnauthorized {
		t.Fatalf("calls=%d code=%d", adapter.calls.Load(), w.Code)
	}
}
