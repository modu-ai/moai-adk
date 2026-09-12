package gateway

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
)

// ServerConfig has no implicit session-auth header or body-size default. The
// launcher chooses the transport header only after its authentication gate.
type ServerConfig struct {
	ManagedAuthority  *AppServerAuthority
	SessionHeader     string
	SessionToken      string
	MaxBodyBytes      int64
	Catalog           CatalogSnapshot
	ResolveCredential func(context.Context, ModelEntry) (CredentialRef, error)
	Adapters          map[ProviderID]Adapter
}

type Server struct {
	managedAuthority  *AppServerAuthority
	sessionHeader     string
	tokenHash         [32]byte
	maxBodyBytes      int64
	catalog           CatalogSnapshot
	resolveCredential func(context.Context, ModelEntry) (CredentialRef, error)
	adapters          map[ProviderID]Adapter
}

func NewServer(c ServerConfig) (*Server, error) {
	if c.SessionHeader == "" || c.SessionToken == "" || c.MaxBodyBytes <= 0 || c.MaxBodyBytes == math.MaxInt64 {
		return nil, errors.New("gateway session header, token and positive bounded body limit required")
	}
	for _, entry := range c.Catalog.Entries() {
		if entry.AuthMethod == AuthOAuthPassthrough && !strings.EqualFold(c.SessionHeader, "X-MoAI-Session-Token") {
			return nil, errors.New("OAuth gateway requires the measured X-MoAI-Session-Token header")
		}
	}
	for _, r := range c.SessionHeader {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", r)) {
			return nil, errors.New("invalid gateway session header")
		}
	}
	if strings.ContainsAny(c.SessionToken, "\r\n") {
		return nil, errors.New("invalid gateway session token")
	}
	s := &Server{managedAuthority: c.ManagedAuthority, sessionHeader: http.CanonicalHeaderKey(c.SessionHeader), tokenHash: sha256.Sum256([]byte(c.SessionToken)), maxBodyBytes: c.MaxBodyBytes, catalog: c.Catalog, resolveCredential: c.ResolveCredential, adapters: make(map[ProviderID]Adapter, len(c.Adapters))}
	for id, a := range c.Adapters {
		s.adapters[id] = a
	}
	return s, nil
}

// ServeHTTP authenticates before inspecting the path or reading the body. It
// never forwards unknown paths or arbitrary URLs. Only an OAuth catalog entry
// may resolve an ingress Bearer into a request-scoped provider credential.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := r.Header.Values(s.sessionHeader)
	provided := ""
	if len(values) == 1 {
		provided = values[0]
	}
	hash := sha256.Sum256([]byte(provided))
	if subtle.ConstantTimeCompare(hash[:], s.tokenHash[:]) != 1 {
		writeError(w, 401, "authentication_error")
		return
	}
	if code, kind := pathPolicy(r); code != 0 {
		writeError(w, code, kind)
		return
	}
	if r.URL.Path == "/v1/models" {
		s.models(w)
		return
	}
	body, err := readRequestBody(r, s.maxBodyBytes)
	if err != nil {
		code := 400
		if errors.Is(err, errBodyTooLarge) {
			code = 413
		}
		writeError(w, code, "invalid_request_error")
		return
	}
	validation, err := IsValidationRequest(body)
	if err != nil {
		writeError(w, 400, "invalid_request_error")
		return
	}
	var payload map[string]json.RawMessage
	_ = json.Unmarshal(body, &payload)
	var model string
	if err := json.Unmarshal(payload["model"], &model); err != nil || model == "" {
		writeError(w, 400, "invalid_request_error")
		return
	}
	entry, err := s.catalog.Resolve(model)
	if err != nil {
		writeError(w, 404, "not_found_error")
		return
	}
	if r.URL.Path == "/v1/messages/count_tokens" {
		writeCount(w, entry, payload)
		return
	}
	var ref CredentialRef
	var generation uint64
	var managed *ManagedGrant
	if entry.AuthMethod == AuthAppServer {
		managed, err = s.managedAuthority.Authorize(r.Context(), entry)
	} else {
		ref, generation, err = s.credential(r.Context(), entry, r.Header)
	}
	if err != nil {
		writeError(w, 401, "authentication_error")
		return
	}
	if validation {
		writeJSON(w, 200, map[string]any{"id": "msg_gateway_validation", "type": "message", "role": "assistant", "model": model, "content": []any{map[string]string{"type": "text", "text": ""}}, "stop_reason": "end_turn", "stop_sequence": nil, "usage": map[string]int{"input_tokens": 0, "output_tokens": 0}})
		return
	}
	adapter := s.adapters[entry.Provider]
	if adapter == nil {
		writeError(w, 503, "api_error")
		return
	}
	if err := r.Context().Err(); err != nil {
		writeError(w, 408, "timeout_error")
		return
	}
	if managed != nil {
		err = s.managedAuthority.Check(r.Context(), managed)
	} else {
		var current uint64
		current, err = ref.Generation()
		if err == nil && current != generation {
			err = ErrManagedAuthority
		}
	}
	if err != nil {
		writeError(w, 401, "authentication_error")
		return
	}
	headers := make(http.Header)
	for _, key := range []string{"Anthropic-Version", "Anthropic-Beta"} {
		if !strings.EqualFold(key, s.sessionHeader) {
			if values := r.Header.Values(key); len(values) != 0 {
				headers[key] = append([]string(nil), values...)
			}
		}
	}
	response, err := adapter.Send(r.Context(), RoutedRequest{Entry: entry, Body: body, Headers: headers, Credential: ref, Generation: generation, Managed: managed})
	if err != nil || response == nil || response.Body == nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		writeError(w, 502, "api_error")
		return
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 599 {
		writeError(w, 502, "api_error")
		return
	}
	for _, key := range []string{"Content-Type", "Retry-After", "Request-Id"} {
		if v := response.Header.Get(key); v != "" {
			w.Header().Set(key, v)
		}
	}
	w.WriteHeader(response.StatusCode)
	// A truncated/erroring body is copied as-is. Never synthesize a terminal event
	// or retry once the response status or bytes have been exposed to the caller.
	_, _ = io.Copy(flushingWriter{w}, response.Body)
}

type flushingWriter struct{ http.ResponseWriter }

func (w flushingWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
	return n, err
}

func (s *Server) models(w http.ResponseWriter) {
	rows := make([]map[string]string, 0)
	for _, entry := range s.catalog.Entries() {
		rows = append(rows, map[string]string{"id": entry.RouteID})
	}
	writeJSON(w, 200, map[string]any{"data": rows})
}
func writeCount(w http.ResponseWriter, entry ModelEntry, payload map[string]json.RawMessage) {
	count, err := estimateInputProjection(payload)
	if err != nil {
		writeError(w, 400, "invalid_request_error")
		return
	}
	writeJSON(w, 200, map[string]any{"input_tokens": count, "accuracy": "estimate", "algorithm": "json-runes-div4", "provider": entry.Provider})
}
func writeError(w http.ResponseWriter, status int, kind string) {
	writeJSON(w, status, map[string]any{"type": "error", "error": map[string]string{"type": kind, "message": http.StatusText(status)}})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
