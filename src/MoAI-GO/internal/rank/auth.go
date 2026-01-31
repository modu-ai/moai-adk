package rank

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

const (
	// AuthTimeout is the maximum time to wait for OAuth callback
	AuthTimeout = 5 * time.Minute
	// MinPort is the start of the port range for the callback server
	MinPort = 8080
	// MaxPort is the end of the port range for the callback server
	MaxPort = 8180
)

// OAuthResult contains the result of an OAuth flow
type OAuthResult struct {
	APIKey    string `json:"api_key"`
	Username  string `json:"username"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
	Error     string `json:"error"`
}

// generateState creates a cryptographically secure state token
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// findAvailablePort finds an available port in the range [MinPort, MaxPort]
func findAvailablePort() (int, error) {
	for port := MinPort; port <= MaxPort; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			if closeErr := ln.Close(); closeErr != nil {
				continue
			}
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port in range %d-%d", MinPort, MaxPort)
}

// browserOpener is the function used to open URLs in a browser, overridable in tests
var browserOpener = openBrowserDefault

// openBrowser opens the given URL in the default browser
func openBrowser(url string) error {
	return browserOpener(url)
}

func openBrowserDefault(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Login performs the OAuth flow: starts a local server, opens browser, waits for callback
func Login() (*Credentials, error) {
	state, err := generateState()
	if err != nil {
		return nil, err
	}

	port, err := findAvailablePort()
	if err != nil {
		return nil, err
	}

	resultCh := make(chan *OAuthResult, 1)
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify state parameter
		if query.Get("state") != state {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body><h2>Authentication failed: invalid state</h2></body></html>")
			resultCh <- &OAuthResult{Error: "invalid state parameter"}
			return
		}

		// Check for error
		if errMsg := query.Get("error"); errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body><h2>Authentication failed: %s</h2></body></html>", errMsg)
			resultCh <- &OAuthResult{Error: errMsg}
			return
		}

		// Extract credentials from callback
		result := &OAuthResult{
			APIKey:    query.Get("api_key"),
			Username:  query.Get("username"),
			UserID:    query.Get("user_id"),
			CreatedAt: query.Get("created_at"),
		}

		if result.APIKey == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body><h2>Authentication failed: no API key received</h2></body></html>")
			resultCh <- &OAuthResult{Error: "no API key in callback"}
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<html><body><h2>Authentication successful!</h2><p>You can close this window.</p></body></html>")
		resultCh <- result
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server in background
	go func() {
		if listenErr := server.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
			resultCh <- &OAuthResult{Error: fmt.Sprintf("server error: %v", listenErr)}
		}
	}()

	// Build auth URL
	authURL := fmt.Sprintf("%s/api/auth/cli?redirect_uri=%s&state=%s",
		"https://rank.mo.ai.kr", redirectURI, state)

	// Open browser
	if browserErr := openBrowser(authURL); browserErr != nil {
		fmt.Printf("Could not open browser. Please visit:\n%s\n", authURL)
	}

	// Wait for callback or timeout
	ctx, cancel := context.WithTimeout(context.Background(), AuthTimeout)
	defer cancel()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	select {
	case result := <-resultCh:
		if result.Error != "" {
			return nil, fmt.Errorf("authentication failed: %s", result.Error)
		}
		creds := &Credentials{
			APIKey:    result.APIKey,
			Username:  result.Username,
			UserID:    result.UserID,
			CreatedAt: result.CreatedAt,
		}
		if err := SaveCredentials(creds); err != nil {
			return nil, fmt.Errorf("failed to save credentials: %w", err)
		}
		return creds, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("authentication timed out after %v", AuthTimeout)
	}
}

// LoginWithAPIKey performs login using a provided API key (non-interactive)
func LoginWithAPIKey(apiKey string) (*Credentials, error) {
	client := NewClient(apiKey)
	rankResp, err := client.GetRank()
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	creds := &Credentials{
		APIKey:    apiKey,
		Username:  rankResp.Username,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := SaveCredentials(creds); err != nil {
		return nil, fmt.Errorf("failed to save credentials: %w", err)
	}
	return creds, nil
}

// AuthURL returns the full OAuth URL for manual login
func AuthURL() string {
	return fmt.Sprintf("%s/api/auth/cli", "https://rank.mo.ai.kr")
}

// StatusResponse represents a rank status check result
type StatusResponse struct {
	LoggedIn bool
	Username string
	Ranks    map[string]RankInfo
	Tokens   TokenStats
}

// GetStatus returns the current rank status
func GetStatus() (*StatusResponse, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return &StatusResponse{LoggedIn: false}, nil
	}

	client := NewClient(creds.APIKey)
	rankResp, rankErr := client.GetRank()
	if rankErr != nil {
		return &StatusResponse{
			LoggedIn: true,
			Username: creds.Username,
		}, rankErr
	}

	return &StatusResponse{
		LoggedIn: true,
		Username: rankResp.Username,
		Ranks:    rankResp.Ranks,
		Tokens:   rankResp.Tokens,
	}, nil
}
