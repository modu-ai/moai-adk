package rank

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// @MX:NOTE: [AUTO] OAuth 인증 흐름은 로컬 HTTP 콜백 서버를 사용하여 자격 증명을 수집합니다. DefaultOAuthTimeout은 300초입니다.
// OAuth authentication constants.
const (
	DefaultOAuthTimeout = 300 * time.Second
	OAuthPortMin        = 8080
	OAuthPortMax        = 8180
	StateTokenBytes     = 32
)

// BrowserOpener abstracts browser opening for testability.
type BrowserOpener interface {
	Open(url string) error
}

// OAuthHandler defines the interface for OAuth authentication flows.
type OAuthHandler interface {
	// StartOAuthFlow initiates the GitHub OAuth flow and waits for credentials.
	StartOAuthFlow(ctx context.Context, timeout time.Duration) (*Credentials, error)
}

// OAuthConfig holds configuration for the OAuth flow.
type OAuthConfig struct {
	BaseURL string
	Browser BrowserOpener
}

// DefaultOAuthHandler implements OAuthHandler using a local HTTP callback server.
type DefaultOAuthHandler struct {
	config OAuthConfig
}

// Compile-time interface check.
var _ OAuthHandler = (*DefaultOAuthHandler)(nil)

// NewOAuthHandler creates a new DefaultOAuthHandler with the given configuration.
func NewOAuthHandler(config OAuthConfig) *DefaultOAuthHandler {
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	return &DefaultOAuthHandler{config: config}
}

// GenerateStateToken generates a cryptographically random state token for CSRF protection.
// Returns a hex-encoded string of StateTokenBytes random bytes.
func GenerateStateToken() (string, error) {
	b := make([]byte, StateTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate state token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// FindAvailablePort finds an available TCP port in the range [OAuthPortMin, OAuthPortMax].
// Returns the port number and the listener. The caller must close the listener.
func FindAvailablePort() (int, net.Listener, error) {
	for port := OAuthPortMin; port <= OAuthPortMax; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			continue // port in use, try next
		}
		return port, ln, nil
	}
	return 0, nil, fmt.Errorf("no available port in range %d-%d", OAuthPortMin, OAuthPortMax)
}

// CallbackResult holds the result received from the OAuth callback.
type CallbackResult struct {
	Credentials *Credentials
	Error       error
}

// StartOAuthFlow initiates the GitHub OAuth authentication flow.
// It starts a local HTTP server to receive the callback, opens the browser,
// and waits for the user to complete authentication.
func (h *DefaultOAuthHandler) StartOAuthFlow(ctx context.Context, timeout time.Duration) (*Credentials, error) {
	if timeout == 0 {
		timeout = DefaultOAuthTimeout
	}

	// Generate CSRF state token.
	state, err := GenerateStateToken()
	if err != nil {
		return nil, err
	}

	// Find available port for callback server.
	port, ln, err := FindAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("find callback port: %w", err)
	}

	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	resultCh := make(chan CallbackResult, 1)

	// Create callback handler.
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		h.handleCallback(w, r, state, resultCh)
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// @MX:WARN: [AUTO] 고루틴이 독립적으로 실행되어 server.Shutdown()으로만 종료됩니다. HTTP 서버 라이프사이클에 주의가 필요합니다.
	// @MX:REASON: 고루틴이 context를 공유하지 않아 부모 컨텍스트 취소 시 즉시 종료되지 않습니다
	// Start server in background.
	go func() {
		if serveErr := server.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			resultCh <- CallbackResult{Error: fmt.Errorf("callback server: %w", serveErr)}
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	// Build auth URL and open browser.
	// Use /api/auth/cli endpoint (matching Python moai-adk implementation)
	authURL := fmt.Sprintf("%s/api/auth/cli?redirect_uri=%s&state=%s", h.config.BaseURL, callbackURL, state)
	if h.config.Browser != nil {
		if openErr := h.config.Browser.Open(authURL); openErr != nil {
			return nil, fmt.Errorf("open browser: %w", openErr)
		}
	}

	// Wait for callback or timeout.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case result := <-resultCh:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.Credentials, nil
	case <-ctx.Done():
		return nil, &AuthenticationError{Message: "OAuth flow timed out"}
	}
}

// handleCallback processes the OAuth callback request.
// Validates the state token and extracts credentials from the response.
func (h *DefaultOAuthHandler) handleCallback(w http.ResponseWriter, r *http.Request, expectedState string, resultCh chan<- CallbackResult) {
	query := r.URL.Query()

	// Validate state token for CSRF protection (R-SEC-003).
	receivedState := query.Get("state")
	if receivedState != expectedState {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "Invalid state token")
		resultCh <- CallbackResult{Error: &AuthenticationError{Message: "state token mismatch"}}
		return
	}

	// Check for error response.
	if errMsg := query.Get("error"); errMsg != "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Authentication error: %s", errMsg)
		resultCh <- CallbackResult{Error: &AuthenticationError{Message: errMsg}}
		return
	}

	// Try new flow: credentials directly in query params.
	apiKey := query.Get("api_key")
	if apiKey != "" {
		creds := &Credentials{
			APIKey:    apiKey,
			Username:  query.Get("username"),
			UserID:    query.Get("user_id"),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Authentication successful! You can close this window.")
		resultCh <- CallbackResult{Credentials: creds}
		return
	}

	// Try new flow: credentials in JSON body (POST callback).
	if r.Method == http.MethodPost && r.Body != nil {
		var creds Credentials
		if decodeErr := json.NewDecoder(r.Body).Decode(&creds); decodeErr == nil && creds.APIKey != "" {
			if creds.CreatedAt == "" {
				creds.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "Authentication successful!")
			resultCh <- CallbackResult{Credentials: &creds}
			return
		}
	}

	// Legacy flow: exchange authorization code for API key.
	code := query.Get("code")
	if code != "" {
		creds, exchangeErr := h.exchangeCode(code)
		if exchangeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, "Failed to exchange authorization code")
			resultCh <- CallbackResult{Error: exchangeErr}
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Authentication successful! You can close this window.")
		resultCh <- CallbackResult{Credentials: creds}
		return
	}

	// No recognizable credential format.
	w.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprint(w, "Missing credentials in callback")
	resultCh <- CallbackResult{Error: &AuthenticationError{Message: "no credentials in callback"}}
}

// exchangeCode exchanges an authorization code for credentials via the Rank API.
func (h *DefaultOAuthHandler) exchangeCode(code string) (*Credentials, error) {
	payload := struct {
		Code string `json:"code"`
	}{Code: code}

	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal exchange request: %w", marshalErr)
	}

	url := h.config.BaseURL + "/api/auth/cli/token"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create exchange request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Body = http.NoBody

	// Re-create request with body.
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, &AuthenticationError{
			Message: fmt.Sprintf("code exchange failed with status %d", resp.StatusCode),
		}
	}

	var creds Credentials
	if decodeErr := json.NewDecoder(resp.Body).Decode(&creds); decodeErr != nil {
		return nil, fmt.Errorf("decode exchange response: %w", decodeErr)
	}

	if creds.CreatedAt == "" {
		creds.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	return &creds, nil
}
