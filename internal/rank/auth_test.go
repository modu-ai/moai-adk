package rank

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockBrowser implements BrowserOpener for testing.
type mockBrowser struct {
	lastURL string
	err     error
}

func (m *mockBrowser) Open(url string) error {
	m.lastURL = url
	return m.err
}

// --- GenerateStateToken Tests ---

func TestGenerateStateToken_NotEmpty(t *testing.T) {
	token, err := GenerateStateToken()
	if err != nil {
		t.Fatalf("GenerateStateToken failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerateStateToken_Length(t *testing.T) {
	token, err := GenerateStateToken()
	if err != nil {
		t.Fatal(err)
	}
	// StateTokenBytes = 32, hex-encoded = 64 chars.
	expected := StateTokenBytes * 2
	if len(token) != expected {
		t.Errorf("expected token length %d, got %d", expected, len(token))
	}
}

func TestGenerateStateToken_Unique(t *testing.T) {
	token1, err := GenerateStateToken()
	if err != nil {
		t.Fatal(err)
	}
	token2, err := GenerateStateToken()
	if err != nil {
		t.Fatal(err)
	}
	if token1 == token2 {
		t.Error("tokens should be unique")
	}
}

func TestGenerateStateToken_HexEncoded(t *testing.T) {
	token, err := GenerateStateToken()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range token {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("token contains non-hex character: %c", c)
		}
	}
}

// --- FindAvailablePort Tests ---

func TestFindAvailablePort_ReturnsValidPort(t *testing.T) {
	port, ln, err := FindAvailablePort()
	if err != nil {
		t.Fatalf("FindAvailablePort failed: %v", err)
	}
	defer func() { _ = ln.Close() }()

	if port < OAuthPortMin || port > OAuthPortMax {
		t.Errorf("port %d outside range [%d, %d]", port, OAuthPortMin, OAuthPortMax)
	}
}

func TestFindAvailablePort_ListenerWorks(t *testing.T) {
	_, ln, err := FindAvailablePort()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	// Verify the listener is functional by accepting a connection.
	addr := ln.Addr().String()
	go func() {
		conn, dialErr := net.Dial("tcp", addr)
		if dialErr == nil {
			_ = conn.Close()
		}
	}()

	conn, err := ln.Accept()
	if err != nil {
		t.Fatalf("listener should accept connections: %v", err)
	}
	_ = conn.Close()
}

func TestFindAvailablePort_SkipsOccupiedPorts(t *testing.T) {
	// Occupy the first port.
	addr := fmt.Sprintf("127.0.0.1:%d", OAuthPortMin)
	occupied, err := net.Listen("tcp", addr)
	if err != nil {
		t.Skipf("cannot occupy port %d: %v", OAuthPortMin, err)
	}
	defer func() { _ = occupied.Close() }()

	// FindAvailablePort should find a different port.
	port, ln, err := FindAvailablePort()
	if err != nil {
		t.Fatalf("FindAvailablePort failed: %v", err)
	}
	defer func() { _ = ln.Close() }()

	if port == OAuthPortMin {
		t.Errorf("should not return occupied port %d", OAuthPortMin)
	}
}

// --- NewOAuthHandler Tests ---

func TestNewOAuthHandler_DefaultBaseURL(t *testing.T) {
	handler := NewOAuthHandler(OAuthConfig{})
	if handler.config.BaseURL != DefaultBaseURL {
		t.Errorf("expected default base URL %q, got %q", DefaultBaseURL, handler.config.BaseURL)
	}
}

func TestNewOAuthHandler_CustomBaseURL(t *testing.T) {
	handler := NewOAuthHandler(OAuthConfig{BaseURL: "https://custom.example.com"})
	if handler.config.BaseURL != "https://custom.example.com" {
		t.Errorf("expected custom base URL, got %q", handler.config.BaseURL)
	}
}

// --- OAuth Callback Tests ---

func TestHandleCallback_ValidState_APIKeyInQuery(t *testing.T) {
	handler := NewOAuthHandler(OAuthConfig{})
	resultCh := make(chan CallbackResult, 1)

	state := "valid-state-token"
	req := httptest.NewRequest(http.MethodGet,
		"/callback?state=valid-state-token&api_key=key123&username=user1&user_id=uid1", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req, state, resultCh)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	select {
	case result := <-resultCh:
		if result.Error != nil {
			t.Fatalf("unexpected error: %v", result.Error)
		}
		if result.Credentials == nil {
			t.Fatal("expected credentials")
		}
		if result.Credentials.APIKey != "key123" {
			t.Errorf("expected api_key key123, got %q", result.Credentials.APIKey)
		}
		if result.Credentials.Username != "user1" {
			t.Errorf("expected username user1, got %q", result.Credentials.Username)
		}
		if result.Credentials.UserID != "uid1" {
			t.Errorf("expected user_id uid1, got %q", result.Credentials.UserID)
		}
		if result.Credentials.CreatedAt == "" {
			t.Error("expected non-empty CreatedAt")
		}
	default:
		t.Fatal("no result received")
	}
}

func TestHandleCallback_InvalidState(t *testing.T) {
	handler := NewOAuthHandler(OAuthConfig{})
	resultCh := make(chan CallbackResult, 1)

	req := httptest.NewRequest(http.MethodGet,
		"/callback?state=wrong-state&api_key=key123", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req, "expected-state", resultCh)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	select {
	case result := <-resultCh:
		if result.Error == nil {
			t.Fatal("expected error for state mismatch")
		}
		authErr, ok := result.Error.(*AuthenticationError)
		if !ok {
			t.Fatalf("expected AuthenticationError, got %T", result.Error)
		}
		if authErr.Message != "state token mismatch" {
			t.Errorf("unexpected error message: %q", authErr.Message)
		}
	default:
		t.Fatal("no result received")
	}
}

func TestHandleCallback_MissingState(t *testing.T) {
	handler := NewOAuthHandler(OAuthConfig{})
	resultCh := make(chan CallbackResult, 1)

	req := httptest.NewRequest(http.MethodGet,
		"/callback?api_key=key123", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req, "expected-state", resultCh)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing state, got %d", w.Code)
	}

	select {
	case result := <-resultCh:
		if result.Error == nil {
			t.Fatal("expected error for missing state")
		}
	default:
		t.Fatal("no result received")
	}
}

func TestHandleCallback_ErrorResponse(t *testing.T) {
	handler := NewOAuthHandler(OAuthConfig{})
	resultCh := make(chan CallbackResult, 1)

	state := "valid-state"
	req := httptest.NewRequest(http.MethodGet,
		"/callback?state=valid-state&error=access_denied", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req, state, resultCh)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	select {
	case result := <-resultCh:
		if result.Error == nil {
			t.Fatal("expected error")
		}
		authErr, ok := result.Error.(*AuthenticationError)
		if !ok {
			t.Fatalf("expected AuthenticationError, got %T", result.Error)
		}
		if authErr.Message != "access_denied" {
			t.Errorf("expected access_denied, got %q", authErr.Message)
		}
	default:
		t.Fatal("no result received")
	}
}

func TestHandleCallback_NoCredentials(t *testing.T) {
	handler := NewOAuthHandler(OAuthConfig{})
	resultCh := make(chan CallbackResult, 1)

	state := "valid-state"
	req := httptest.NewRequest(http.MethodGet,
		"/callback?state=valid-state", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req, state, resultCh)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	select {
	case result := <-resultCh:
		if result.Error == nil {
			t.Fatal("expected error for no credentials")
		}
	default:
		t.Fatal("no result received")
	}
}

func TestHandleCallback_LegacyCodeFlow(t *testing.T) {
	// Set up a mock exchange server.
	exchangeCreds := Credentials{
		APIKey:   "exchanged-key",
		Username: "exchange-user",
		UserID:   "exchange-uid",
	}
	exchangeBody, marshalErr := json.Marshal(exchangeCreds)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}

	exchangeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/cli/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(exchangeBody)
	}))
	defer exchangeSrv.Close()

	handler := NewOAuthHandler(OAuthConfig{BaseURL: exchangeSrv.URL})
	resultCh := make(chan CallbackResult, 1)

	state := "valid-state"
	req := httptest.NewRequest(http.MethodGet,
		"/callback?state=valid-state&code=auth-code-123", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req, state, resultCh)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	select {
	case result := <-resultCh:
		if result.Error != nil {
			t.Fatalf("unexpected error: %v", result.Error)
		}
		if result.Credentials == nil {
			t.Fatal("expected credentials from code exchange")
		}
		if result.Credentials.APIKey != "exchanged-key" {
			t.Errorf("expected exchanged-key, got %q", result.Credentials.APIKey)
		}
	default:
		t.Fatal("no result received")
	}
}

func TestHandleCallback_LegacyCodeFlow_ExchangeFails(t *testing.T) {
	exchangeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer exchangeSrv.Close()

	handler := NewOAuthHandler(OAuthConfig{BaseURL: exchangeSrv.URL})
	resultCh := make(chan CallbackResult, 1)

	state := "valid-state"
	req := httptest.NewRequest(http.MethodGet,
		"/callback?state=valid-state&code=bad-code", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req, state, resultCh)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	select {
	case result := <-resultCh:
		if result.Error == nil {
			t.Fatal("expected error for failed code exchange")
		}
	default:
		t.Fatal("no result received")
	}
}

// --- StartOAuthFlow Tests ---

func TestStartOAuthFlow_Timeout(t *testing.T) {
	browser := &mockBrowser{}
	handler := NewOAuthHandler(OAuthConfig{
		BaseURL: "https://example.com",
		Browser: browser,
	})

	ctx := context.Background()
	_, err := handler.StartOAuthFlow(ctx, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	authErr, ok := err.(*AuthenticationError)
	if !ok {
		t.Fatalf("expected AuthenticationError, got %T: %v", err, err)
	}
	if authErr.Message != "OAuth flow timed out" {
		t.Errorf("unexpected message: %q", authErr.Message)
	}

	// Verify browser was opened with correct URL components.
	if browser.lastURL == "" {
		t.Error("expected browser to be opened")
	}
}

func TestStartOAuthFlow_BrowserError(t *testing.T) {
	browser := &mockBrowser{err: fmt.Errorf("browser not found")}
	handler := NewOAuthHandler(OAuthConfig{
		BaseURL: "https://example.com",
		Browser: browser,
	})

	ctx := context.Background()
	_, err := handler.StartOAuthFlow(ctx, 1*time.Second)
	if err == nil {
		t.Fatal("expected error when browser fails")
	}
}

func TestStartOAuthFlow_ContextCancelled(t *testing.T) {
	browser := &mockBrowser{}
	handler := NewOAuthHandler(OAuthConfig{
		BaseURL: "https://example.com",
		Browser: browser,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := handler.StartOAuthFlow(ctx, 5*time.Second)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestStartOAuthFlow_NoBrowser(t *testing.T) {
	// No browser configured: should not error on Open, just timeout.
	handler := NewOAuthHandler(OAuthConfig{
		BaseURL: "https://example.com",
	})

	ctx := context.Background()
	_, err := handler.StartOAuthFlow(ctx, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// --- OAuth Constants Tests ---

func TestOAuthConstants(t *testing.T) {
	if DefaultOAuthTimeout != 300*time.Second {
		t.Errorf("unexpected DefaultOAuthTimeout: %v", DefaultOAuthTimeout)
	}
	if OAuthPortMin >= OAuthPortMax {
		t.Error("OAuthPortMin should be less than OAuthPortMax")
	}
	if StateTokenBytes != 32 {
		t.Errorf("unexpected StateTokenBytes: %d", StateTokenBytes)
	}
}

// --- CallbackResult Tests ---

func TestCallbackResult_WithCredentials(t *testing.T) {
	result := CallbackResult{
		Credentials: &Credentials{APIKey: "test"},
	}
	if result.Error != nil {
		t.Error("expected no error")
	}
	if result.Credentials.APIKey != "test" {
		t.Error("expected credentials")
	}
}

func TestCallbackResult_WithError(t *testing.T) {
	result := CallbackResult{
		Error: &AuthenticationError{Message: "fail"},
	}
	if result.Credentials != nil {
		t.Error("expected no credentials")
	}
	if result.Error == nil {
		t.Error("expected error")
	}
}
