package cli

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

	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"github.com/modu-ai/moai-adk/internal/glmcred"
)

func factoryDependencies(t *testing.T) gatewayFactoryDependencies {
	t.Helper()
	return gatewayFactoryDependencies{Models: []gateway.ModelEntry{{RouteID: "claude-opus-5", UpstreamID: "claude-opus-5", Provider: gateway.ProviderAnthropic, AuthMethod: gateway.AuthOAuthPassthrough, Capabilities: gateway.Capabilities{ContextTokens: 100, Streaming: true}}}, Transport: &http.Transport{}, Limits: translate.Limits{MaxBodyBytes: 4096, MaxEventBytes: 4096, MaxOutputBytes: 8192}, MeasureInput: func(gateway.ModelEntry, []byte) (int64, error) { return 2, nil }, AnthropicVersion: "2023-06-01", AllowedBetas: []string{"oauth-2025-04-20"}}
}

const factoryPayload = `{"version":1,"session_token":"private-session","model_ids":["claude-opus-5"]}`

type gatewayAuthBroker func(context.Context, string, bool) error

func (f gatewayAuthBroker) Run(ctx context.Context, home string, refresh bool) error {
	return f(ctx, home, refresh)
}

func gatewayAuthToken(expiry time.Time) string {
	claims, _ := json.Marshal(map[string]any{"exp": expiry.Unix()})
	return "e30." + base64.RawURLEncoding.EncodeToString(claims) + ".gateway-test"
}

func seedGatewayAuthStore(t *testing.T, dir string) {
	t.Helper()
	s, err := auth.OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_, err = s.Login(context.Background(), gatewayAuthBroker(func(_ context.Context, home string, _ bool) error {
		authBody, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": gatewayAuthToken(time.Now().Add(time.Hour)), "refresh_token": "refresh", "id_token": "identity", "account_id": "account-1"}})
		return os.WriteFile(filepath.Join(home, "auth.json"), authBody, 0600)
	}))
	if err != nil {
		t.Fatal(err)
	}
}

func TestGatewayFactoryRejectsPayloadBeforeResources(t *testing.T) {
	d := factoryDependencies(t)
	opens := 0
	d.OpenStore = func() (*auth.Store, error) { opens++; return nil, errors.New("must not open") }
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{`null`, `{}`, `{"version":1,"version":1}`, `{"version":1,"session_token":"secret","model_ids":["unknown"]}`, `{"version":1,"session_token":"secret","model_ids":["claude-opus-5","claude-opus-5"]}`, strings.TrimSuffix(factoryPayload, "}") + `,"url":"https://evil.invalid"}`, `{"version":1,"session_token":"x","model_ids":[]}`} {
		if h, err := factory(json.RawMessage(payload)); err == nil || h != nil {
			t.Fatal("invalid payload accepted")
		}
	}
	if opens != 0 {
		t.Fatal("invalid payload opened store")
	}
	d.MeasureInput = nil
	if _, err := newGatewayHandlerFactory(d); err == nil {
		t.Fatal("missing metadata admitted")
	}
}

func TestGatewayFactoryNativeRoutingAndCancellation(t *testing.T) {
	var sends atomic.Int32
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sends.Add(1)
		if r.Host != "api.anthropic.com" || r.URL.Path != "/v1/messages" || r.Header.Get("Authorization") != "Bearer native-token" || r.Header.Get("X-MoAI-Session-Token") != "" {
			t.Error("route or credential crossed")
		}
		w.WriteHeader(401)
	}))
	defer upstream.Close()
	d := factoryDependencies(t)
	tr := upstream.Client().Transport.(*http.Transport).Clone()
	cfg := tr.TLSClientConfig.Clone()
	tr.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "api.anthropic.com:443" {
			return nil, errors.New("wrong destination")
		}
		return (&tls.Dialer{Config: cfg}).DialContext(ctx, "tcp", upstream.Listener.Addr().String())
	}
	d.Transport = tr
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	h, err := factory(json.RawMessage(factoryPayload))
	if err != nil {
		t.Fatal(err)
	}
	defer h.(io.Closer).Close()
	send := func(ctx context.Context, token string) int {
		r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-opus-5","max_tokens":2,"messages":[{"role":"user","content":"hello"}]}`)).WithContext(ctx)
		r.Header.Set("X-MoAI-Session-Token", "private-session")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	if send(context.Background(), "Bearer native-token") != 401 || sends.Load() != 1 {
		t.Fatal("native401 lost")
	}
	send(context.Background(), "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	send(ctx, "Bearer other")
	if sends.Load() != 1 {
		t.Fatal("unauthenticated/cancelled sent")
	}
}

func TestGatewayFactoryStoreOwnershipAndUnavailableLogin(t *testing.T) {
	d := factoryDependencies(t)
	d.Models = []gateway.ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100}}}
	var brokerCalls atomic.Int32
	d.Broker = factoryBroker{calls: &brokerCalls}
	d.RefreshVerifier = func(context.Context, auth.CredentialRef) error {
		t.Error("missing login verified")
		return errors.New("must not verify missing login")
	}
	var store *auth.Store
	d.OpenStore = func() (*auth.Store, error) {
		var err error
		dir, pathErr := filepath.EvalSymlinks(t.TempDir())
		if pathErr != nil {
			return nil, pathErr
		}
		store, err = auth.OpenStore(filepath.Join(dir, "auth"))
		return store, err
	}
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	h, err := factory(json.RawMessage(strings.ReplaceAll(factoryPayload, "claude-opus-5", "gpt-5.6-sol")))
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"gpt-5.6-sol","max_tokens":2,"messages":[{"role":"user","content":"hi"}]}`))
	r.Header.Set("Authorization", "Bearer private-session")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 || brokerCalls.Load() != 0 {
		t.Fatal("missing login not closed")
	}
	if err := h.(io.Closer).Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Status(context.Background()); err == nil {
		t.Fatal("store remained open")
	}
	if err := h.(io.Closer).Close(); err != nil {
		t.Fatal("close not idempotent")
	}
}

func TestGatewayFactoryNativeReceiptAuthorizationReachesSubscription(t *testing.T) {
	sessionID := "11111111-1111-4111-8111-111111111111"
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	receiptDir := filepath.Join(home, "receipt")
	if err := os.MkdirAll(receiptDir, 0700); err != nil {
		t.Fatal(err)
	}
	receipts, err := receipt.OpenStore(context.Background(), receiptDir, sessionID, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := receipts.Close(); err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "chatgpt.com" || r.URL.Path != "/backend-api/codex/responses" {
			t.Error("subscription route crossed")
		}
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("invalid translated request")
		}
		if _, ok := body["max_output_tokens"]; ok {
			t.Error("subscription max_output_tokens escaped")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp","model":"gpt-5.6-sol","status":"in_progress"}}`,
			`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg","type":"message","role":"assistant","status":"in_progress","phase":"final_answer","content":[]}}`,
			`data: {"type":"response.content_part.added","output_index":0,"item_id":"msg","content_index":0,"part":{"type":"output_text","text":"","annotations":[]}}`,
			`data: {"type":"response.output_text.delta","output_index":0,"item_id":"msg","content_index":0,"delta":"ok"}`,
			`data: {"type":"response.output_text.done","output_index":0,"item_id":"msg","content_index":0,"text":"ok"}`,
			`data: {"type":"response.content_part.done","output_index":0,"item_id":"msg","content_index":0,"part":{"type":"output_text","text":"ok","annotations":[]}}`,
			`data: {"type":"response.output_item.done","output_index":0,"item":{"id":"msg","type":"message","role":"assistant","status":"completed","phase":"final_answer","content":[{"type":"output_text","text":"ok","annotations":[],"logprobs":[]}]}}`,
			`data: {"type":"response.completed","response":{"id":"resp","model":"gpt-5.6-sol","status":"completed","output":[{"id":"msg","type":"message","role":"assistant","status":"completed","phase":"final_answer","content":[{"type":"output_text","text":"ok","annotations":[],"logprobs":[]}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`,
		}, "\n\n")+"\n\n")
	}))
	defer upstream.Close()
	tr := upstream.Client().Transport.(*http.Transport).Clone()
	tlsConfig := tr.TLSClientConfig.Clone()
	tr.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "chatgpt.com:443" {
			return nil, errors.New("wrong destination")
		}
		return (&tls.Dialer{Config: tlsConfig}).DialContext(ctx, "tcp", upstream.Listener.Addr().String())
	}
	d := factoryDependencies(t)
	d.Models = []gateway.ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100, Streaming: true}}}
	d.Transport = tr
	d.Limits.PolicyProfile = translate.PolicyGPTNative
	d.SendOptions = auth.SendOptions{WriteTimeout: time.Second, PollInterval: 5 * time.Millisecond, MaxBodyBytes: 1 << 20}
	d.Broker = gatewayAuthBroker(func(_ context.Context, brokerHome string, _ bool) error {
		authBody, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": gatewayAuthToken(time.Now().Add(time.Hour)), "refresh_token": "refresh", "id_token": "identity", "account_id": "account-1"}})
		return os.WriteFile(filepath.Join(brokerHome, "auth.json"), authBody, 0600)
	})
	d.RefreshVerifier = func(context.Context, auth.CredentialRef) error { return nil }
	authDir := filepath.Join(home, "auth")
	seedGatewayAuthStore(t, authDir)
	d.OpenStore = func() (*auth.Store, error) { return auth.OpenStore(authDir) }
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: "private", ModelIDs: []string{"gpt-5.6-sol"}, Conversation: &gatewayPrivateConversation{FamilyID: sessionID, SessionID: sessionID, ReceiptDir: receiptDir}})
	if err != nil {
		t.Fatal(err)
	}
	h, err := factory(payload)
	if err != nil {
		t.Fatal(err)
	}
	defer h.(io.Closer).Close()
	body := `{"model":"gpt-5.6-sol","max_tokens":2,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"},"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]},"metadata":{"user_id":"{\"account_uuid\":\"\",\"device_id\":\"device\",\"session_id\":\"` + sessionID + `\"}"},"messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer private")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"text":"ok"`) {
		t.Fatalf("authorized native request status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestGatewayFactoryNativeReceiptAuthorizationRejectsForeignSession(t *testing.T) {
	sessionID := "22222222-2222-4222-8222-222222222222"
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	receiptDir := filepath.Join(home, "receipt")
	if err := os.MkdirAll(receiptDir, 0700); err != nil {
		t.Fatal(err)
	}
	receipts, err := receipt.OpenStore(context.Background(), receiptDir, sessionID, true)
	if err != nil {
		t.Fatal(err)
	}
	_ = receipts.Close()
	d := factoryDependencies(t)
	d.Models = []gateway.ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100}}}
	d.Limits.PolicyProfile = translate.PolicyGPTNative
	d.Broker = factoryBroker{}
	d.RefreshVerifier = func(context.Context, auth.CredentialRef) error { return nil }
	authDir := filepath.Join(home, "auth")
	seedGatewayAuthStore(t, authDir)
	d.OpenStore = func() (*auth.Store, error) { return auth.OpenStore(authDir) }
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: "private", ModelIDs: []string{"gpt-5.6-sol"}, Conversation: &gatewayPrivateConversation{FamilyID: sessionID, SessionID: sessionID, ReceiptDir: receiptDir}})
	h, err := factory(payload)
	if err != nil {
		t.Fatal(err)
	}
	defer h.(io.Closer).Close()
	body := `{"model":"gpt-5.6-sol","max_tokens":2,"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]},"metadata":{"user_id":"{\"account_uuid\":\"\",\"device_id\":\"device\",\"session_id\":\"33333333-3333-4333-8333-333333333333\"}"},"messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer private")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("foreign native session status=%d body=%s", res.Code, res.Body.String())
	}
}

type factoryClosingHandler struct{ closed bool }

func (h *factoryClosingHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (h *factoryClosingHandler) Close() error                                 { h.closed = true; return nil }
func TestGatewayFactoryChildStartupClosesOwnedHandler(t *testing.T) {
	h := &factoryClosingHandler{}
	cmd := newGatewayChildCommand(func(json.RawMessage) (http.Handler, error) { return h, nil })
	cmd.SetIn(strings.NewReader(`{"parent_pid":1,"parent_fingerprint":"invalid","lifetime":1000000000,"poll_interval":1000000,"payload":{}}` + "\n"))
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err == nil {
		t.Fatal("bad parent started")
	}
	if !h.closed {
		t.Fatal("startup leaked handler")
	}
}

type factoryBroker struct{ calls *atomic.Int32 }

func (factoryBroker) Run(context.Context, string, bool) error { return errors.New("unexpected broker") }

func TestGatewayFactoryFailedInitializationClosesStore(t *testing.T) {
	d := factoryDependencies(t)
	d.Models = append(d.Models, gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100}})
	d.Broker = factoryBroker{}
	d.RefreshVerifier = func(context.Context, auth.CredentialRef) error { return nil }
	d.AllowedBetas = []string{"invalid beta"}
	var store *auth.Store
	d.OpenStore = func() (*auth.Store, error) {
		dir, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			return nil, err
		}
		store, err = auth.OpenStore(filepath.Join(dir, "auth"))
		return store, err
	}
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"version":1,"session_token":"private","model_ids":["claude-opus-5","gpt-5.6-sol"]}`
	if h, err := factory(json.RawMessage(payload)); err == nil || h != nil {
		t.Fatal("bad adapter started")
	}
	if store == nil {
		t.Fatal("cleanup path not reached")
	}
	if _, err := store.Status(context.Background()); err == nil {
		t.Fatal("initialization leaked store")
	}
}

func TestGatewayFactoryCloseCancelsAndJoinsRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started, done := make(chan struct{}), make(chan struct{})
	h := &gatewayOwnedHandler{ctx: ctx, cancel: cancel, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done() })}
	go func() { defer close(done); h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request not started")
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("close did not join")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 503 {
		t.Fatal("closed handler reused")
	}
}

func TestGatewayFactoryRejectsUnverifiedModelsAndGPTVerifier(t *testing.T) {
	for _, change := range []func(*gatewayFactoryDependencies){
		func(d *gatewayFactoryDependencies) { d.Models[0].AuthMethod = gateway.AuthAPIKey },
		func(d *gatewayFactoryDependencies) { d.Models[0].UpstreamID = "unknown" },
		func(d *gatewayFactoryDependencies) { d.Models[0].Capabilities.ContextTokens = 0 },
		func(d *gatewayFactoryDependencies) {
			d.Models = []gateway.ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100}}}
		},
	} {
		d := factoryDependencies(t)
		change(&d)
		if _, err := newGatewayHandlerFactory(d); err == nil {
			t.Fatal("unverified dependency admitted")
		}
	}
}

func TestGatewayFactoryGLMExactRouteAndBetaRemoval(t *testing.T) {
	t.Setenv(glmcred.EnvTestGLMKey, "synthetic-stored-glm")
	var sends atomic.Int32
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sends.Add(1)
		if r.Host != "api.z.ai" || r.URL.Path != "/api/anthropic/v1/messages" || r.Header.Get("Authorization") != "Bearer synthetic-stored-glm" || r.Header.Get("Anthropic-Beta") != "" || r.Header.Get("X-MoAI-Session-Token") != "" {
			t.Error("GLM credential or route crossed")
		}
		w.WriteHeader(401)
	}))
	defer upstream.Close()
	d := factoryDependencies(t)
	d.Models = []gateway.ModelEntry{{RouteID: "glm-approved", UpstreamID: "glm-approved", Provider: gateway.ProviderZAI, AuthMethod: gateway.AuthExistingGLM, Capabilities: gateway.Capabilities{ContextTokens: 100}}}
	tr := upstream.Client().Transport.(*http.Transport).Clone()
	cfg := tr.TLSClientConfig.Clone()
	tr.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "api.z.ai:443" {
			return nil, errors.New("wrong destination")
		}
		return (&tls.Dialer{Config: cfg}).DialContext(ctx, "tcp", upstream.Listener.Addr().String())
	}
	d.Transport = tr
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	h, err := factory(json.RawMessage(strings.ReplaceAll(factoryPayload, "claude-opus-5", "glm-approved")))
	if err != nil {
		t.Fatal(err)
	}
	defer h.(io.Closer).Close()
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"glm-approved","max_tokens":2,"messages":[{"role":"user","content":"hi"}]}`))
	r.Header.Set("X-Api-Key", "private-session")
	r.Header.Set("Authorization", "Bearer must-not-forward")
	r.Header.Set("Anthropic-Beta", "oauth-2025-04-20")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 || sends.Load() != 1 {
		t.Fatal("GLM route not executed")
	}
}

func TestGatewayFactoryExactPayloadKeysBeforeStore(t *testing.T) {
	d := factoryDependencies(t)
	d.Models = []gateway.ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100}}}
	d.Broker = factoryBroker{}
	d.RefreshVerifier = func(context.Context, auth.CredentialRef) error { return nil }
	opens := 0
	d.OpenStore = func() (*auth.Store, error) { opens++; return nil, errors.New("synthetic unavailable store") }
	f, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	fields := []string{`"version":1`, `"session_token":"synthetic"`, `"model_ids":["gpt-5.6-sol"]`}
	for i, field := range fields {
		key, value, _ := strings.Cut(field, ":")
		for _, alias := range []string{strings.ToUpper(key), strings.Replace(key, string(key[1]), strings.ToUpper(string(key[1])), 1)} {
			for _, replacement := range []string{alias + ":" + value, field + "," + alias + ":" + value, alias + ":" + value + "," + field, field + "," + field} {
				parts := append([]string(nil), fields...)
				parts[i] = replacement
				h, e := f(json.RawMessage("{" + strings.Join(parts, ",") + "}"))
				if h != nil {
					h.(io.Closer).Close()
				}
				if e == nil || h != nil || opens != 0 {
					t.Errorf("field %d alias or duplicate reached resource", i)
					opens = 0
				}
			}
		}
	}
	_, _ = f(json.RawMessage("{" + strings.Join(fields, ",") + "}"))
	if opens != 1 {
		t.Fatal("canonical positive control did not reach resource")
	}
}

func TestGatewayFactorySubscriptionSwitchKeepsAuthorizedHistory(t *testing.T) {
	sessionID := "11111111-1111-4111-8111-111111111111"
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	receiptDir := filepath.Join(home, "receipt")
	if err := os.MkdirAll(receiptDir, 0700); err != nil {
		t.Fatal(err)
	}
	receipts, err := receipt.OpenStore(context.Background(), receiptDir, sessionID, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := receipts.Close(); err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "chatgpt.com" || r.URL.Path != "/backend-api/codex/responses" {
			t.Error("subscription route crossed")
		}
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("invalid translated request")
		}
		if _, ok := body["max_output_tokens"]; ok {
			t.Error("subscription max_output_tokens escaped")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.ReplaceAll(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp","model":"gpt-5.6-sol","status":"in_progress"}}`,
			`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg","type":"message","role":"assistant","status":"in_progress","phase":"final_answer","content":[]}}`,
			`data: {"type":"response.content_part.added","output_index":0,"item_id":"msg","content_index":0,"part":{"type":"output_text","text":"","annotations":[]}}`,
			`data: {"type":"response.output_text.delta","output_index":0,"item_id":"msg","content_index":0,"delta":"ok"}`,
			`data: {"type":"response.output_text.done","output_index":0,"item_id":"msg","content_index":0,"text":"ok"}`,
			`data: {"type":"response.content_part.done","output_index":0,"item_id":"msg","content_index":0,"part":{"type":"output_text","text":"ok","annotations":[]}}`,
			`data: {"type":"response.output_item.done","output_index":0,"item":{"id":"msg","type":"message","role":"assistant","status":"completed","phase":"final_answer","content":[{"type":"output_text","text":"ok","annotations":[],"logprobs":[]}]}}`,
			`data: {"type":"response.completed","response":{"id":"resp","model":"gpt-5.6-sol","status":"completed","output":[{"id":"msg","type":"message","role":"assistant","status":"completed","phase":"final_answer","content":[{"type":"output_text","text":"ok","annotations":[],"logprobs":[]}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`,
		}, "\n\n")+"\n\n", "gpt-5.6-sol", body["model"].(string)))
	}))
	defer upstream.Close()
	tr := upstream.Client().Transport.(*http.Transport).Clone()
	tlsConfig := tr.TLSClientConfig.Clone()
	tr.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "chatgpt.com:443" {
			return nil, errors.New("wrong destination")
		}
		return (&tls.Dialer{Config: tlsConfig}).DialContext(ctx, "tcp", upstream.Listener.Addr().String())
	}
	d := factoryDependencies(t)
	d.Models = []gateway.ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100, Streaming: true}}}
	d.Models = append(d.Models, gateway.ModelEntry{RouteID: "gpt-6-astra", UpstreamID: "gpt-6-astra", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: 100, Streaming: true}})
	d.Transport = tr
	d.Limits.PolicyProfile = translate.PolicyGPTNative
	d.SendOptions = auth.SendOptions{WriteTimeout: time.Second, PollInterval: 5 * time.Millisecond, MaxBodyBytes: 1 << 20}
	d.Broker = gatewayAuthBroker(func(_ context.Context, brokerHome string, _ bool) error {
		authBody, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": gatewayAuthToken(time.Now().Add(time.Hour)), "refresh_token": "refresh", "id_token": "identity", "account_id": "account-1"}})
		return os.WriteFile(filepath.Join(brokerHome, "auth.json"), authBody, 0600)
	})
	d.RefreshVerifier = func(context.Context, auth.CredentialRef) error { return nil }
	authDir := filepath.Join(home, "auth")
	seedGatewayAuthStore(t, authDir)
	d.OpenStore = func() (*auth.Store, error) { return auth.OpenStore(authDir) }
	factory, err := newGatewayHandlerFactory(d)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: "private", ModelIDs: []string{"gpt-5.6-sol", "gpt-6-astra"}, Conversation: &gatewayPrivateConversation{FamilyID: sessionID, SessionID: sessionID, ReceiptDir: receiptDir}})
	if err != nil {
		t.Fatal(err)
	}
	h, err := factory(payload)
	if err != nil {
		t.Fatal(err)
	}
	defer h.(io.Closer).Close()
	body := `{"model":"gpt-5.6-sol","max_tokens":2,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"},"context_management":{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]},"metadata":{"user_id":"{\"account_uuid\":\"\",\"device_id\":\"device\",\"session_id\":\"` + sessionID + `\"}"},"messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer private")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"text":"ok"`) {
		t.Fatalf("authorized native request status=%d body=%s", res.Code, res.Body.String())
	}
	var first map[string]any
	if err = json.Unmarshal(res.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	var next map[string]any
	if err = json.Unmarshal([]byte(body), &next); err != nil {
		t.Fatal(err)
	}
	next["model"] = "gpt-6-astra"
	next["messages"] = []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": first["content"]}, map[string]any{"role": "user", "content": "continue"}}
	raw, _ := json.Marshal(next)
	req = httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(raw)))
	req.Header.Set("Authorization", "Bearer private")
	res = httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"text":"ok"`) {
		t.Fatalf("model switch status=%d body=%s", res.Code, res.Body.String())
	}

}
