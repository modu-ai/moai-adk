// Package webhook provides an HTTP server for receiving external trigger events.
package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Server receives external webhook events.
type Server struct {
	port      int
	secretEnv string
	handler   EventHandler
	srv       *http.Server
}

// EventHandler processes incoming webhook events.
type EventHandler func(event map[string]any) error

// NewServer creates a new webhook server.
func NewServer(port int, secretEnv string, handler EventHandler) *Server {
	return &Server{
		port:      port,
		secretEnv: secretEnv,
		handler:   handler,
	}
}

// Start begins listening for webhook events.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)
	mux.HandleFunc("/health", s.handleHealth)

	s.srv = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	slog.Info("webhook server starting", "port", s.port)
	return s.srv.ListenAndServe()
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// Verify signature — reject if secret is not configured
	secret := os.Getenv(s.secretEnv)
	if secret == "" {
		http.Error(w, "webhook secret not configured", http.StatusForbidden)
		return
	}
	{
		sig := r.Header.Get("X-Signature-256")
		if !verifySignature(body, sig, secret) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	var event map[string]any
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if s.handler != nil {
		if err := s.handler(event); err != nil {
			slog.Error("webhook handler error", "error", err)
			http.Error(w, "handler error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
}

func verifySignature(payload []byte, signature, secret string) bool {
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
