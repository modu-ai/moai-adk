package rank

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- ComputeSignature Tests ---

func TestComputeSignature_Deterministic(t *testing.T) {
	sig1 := ComputeSignature("key", "12345", "body")
	sig2 := ComputeSignature("key", "12345", "body")
	if sig1 != sig2 {
		t.Error("same inputs should produce same signature")
	}
}

func TestComputeSignature_DifferentKeys(t *testing.T) {
	sig1 := ComputeSignature("key1", "12345", "body")
	sig2 := ComputeSignature("key2", "12345", "body")
	if sig1 == sig2 {
		t.Error("different keys should produce different signatures")
	}
}

func TestComputeSignature_DifferentTimestamps(t *testing.T) {
	sig1 := ComputeSignature("key", "11111", "body")
	sig2 := ComputeSignature("key", "22222", "body")
	if sig1 == sig2 {
		t.Error("different timestamps should produce different signatures")
	}
}

func TestComputeSignature_DifferentBodies(t *testing.T) {
	sig1 := ComputeSignature("key", "12345", "body1")
	sig2 := ComputeSignature("key", "12345", "body2")
	if sig1 == sig2 {
		t.Error("different bodies should produce different signatures")
	}
}

func TestComputeSignature_EmptyBody(t *testing.T) {
	sig := ComputeSignature("key", "12345", "")
	if sig == "" {
		t.Error("signature should not be empty")
	}
	if len(sig) != 64 { // SHA-256 hex = 64 chars
		t.Errorf("expected 64 hex chars, got %d", len(sig))
	}
}

func TestComputeSignature_HexEncoded(t *testing.T) {
	sig := ComputeSignature("key", "ts", "body")
	for _, c := range sig {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("signature contains non-hex character: %c", c)
		}
	}
}

// --- NewClient Tests ---

func TestNewClient_Defaults(t *testing.T) {
	t.Setenv("MOAI_RANK_API_URL", "")

	c := NewClient("test-key")
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey test-key, got %q", c.apiKey)
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("expected baseURL %q, got %q", DefaultBaseURL, c.baseURL)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewClient_EnvOverride(t *testing.T) {
	t.Setenv("MOAI_RANK_API_URL", "https://env.example.com")

	c := NewClient("key")
	if c.baseURL != "https://env.example.com" {
		t.Errorf("expected env URL, got %q", c.baseURL)
	}
}

func TestNewClient_WithBaseURL(t *testing.T) {
	t.Setenv("MOAI_RANK_API_URL", "")

	c := NewClient("key", WithBaseURL("https://custom.example.com"))
	if c.baseURL != "https://custom.example.com" {
		t.Errorf("expected custom URL, got %q", c.baseURL)
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 99 * time.Second}
	c := NewClient("key", WithHTTPClient(custom))
	if c.httpClient != custom {
		t.Error("expected custom HTTP client")
	}
}

func TestNewClient_EmptyKey(t *testing.T) {
	c := NewClient("")
	if c.apiKey != "" {
		t.Error("expected empty apiKey")
	}
}

// --- CheckStatus Tests ---

func TestCheckStatus_Success(t *testing.T) {
	status := ApiStatus{
		Status:    "ok",
		Version:   "1.0.0",
		Timestamp: "2026-01-15T10:00:00Z",
	}
	body, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	c := NewClient("", WithBaseURL(srv.URL))
	result, err := c.CheckStatus(context.Background())
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if result.Status != "ok" {
		t.Errorf("expected status ok, got %q", result.Status)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", result.Version)
	}
}

func TestCheckStatus_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer srv.Close()

	c := NewClient("", WithBaseURL(srv.URL))
	_, err := c.CheckStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	apiErr, ok := err.(*ApiError)
	if !ok {
		t.Fatalf("expected ApiError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestCheckStatus_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := NewClient("", WithBaseURL(srv.URL))
	_, err := c.CheckStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if _, ok := err.(*ClientError); !ok {
		t.Errorf("expected ClientError, got %T", err)
	}
}

// --- GetUserRank Tests ---

func TestGetUserRank_Success(t *testing.T) {
	rank := UserRank{
		Username: "testuser",
		Stats: &UserRankStats{
			TotalTokens:   50000,
			TotalSessions: 10,
			InputTokens:   30000,
			OutputTokens:  20000,
		},
	}
	body, err := json.Marshal(rank)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/rank" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify auth headers are present.
		if r.Header.Get("X-API-Key") == "" {
			t.Error("missing X-API-Key header")
		}
		if r.Header.Get("X-Timestamp") == "" {
			t.Error("missing X-Timestamp header")
		}
		if r.Header.Get("X-Signature") == "" {
			t.Error("missing X-Signature header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	c := NewClient("test-api-key", WithBaseURL(srv.URL))
	result, err := c.GetUserRank(context.Background())
	if err != nil {
		t.Fatalf("GetUserRank failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("expected username testuser, got %q", result.Username)
	}
	if result.Stats == nil {
		t.Fatal("expected Stats to be non-nil")
	}
	if result.Stats.TotalTokens != 50000 {
		t.Errorf("expected 50000 tokens, got %d", result.Stats.TotalTokens)
	}
}

func TestGetUserRank_NoAPIKey(t *testing.T) {
	c := NewClient("", WithBaseURL("http://localhost"))
	_, err := c.GetUserRank(context.Background())
	if err == nil {
		t.Fatal("expected error with empty API key")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestGetUserRank_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient("bad-key", WithBaseURL(srv.URL))
	_, err := c.GetUserRank(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestGetUserRank_Forbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient("key", WithBaseURL(srv.URL))
	_, err := c.GetUserRank(context.Background())
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

// --- GetLeaderboard Tests ---

func TestGetLeaderboard_Success(t *testing.T) {
	entries := []LeaderboardEntry{
		{Rank: 1, Username: "top", TotalTokens: 100000, CompositeScore: 95.5, SessionCount: 50},
		{Rank: 2, Username: "second", TotalTokens: 80000, CompositeScore: 88.2, SessionCount: 40},
	}
	body, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/leaderboard" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("period") != "weekly" {
			t.Errorf("expected period weekly, got %q", q.Get("period"))
		}
		if q.Get("limit") != "10" {
			t.Errorf("expected limit 10, got %q", q.Get("limit"))
		}
		if q.Get("offset") != "0" {
			t.Errorf("expected offset 0, got %q", q.Get("offset"))
		}
		// Should not have auth headers (public API).
		if r.Header.Get("X-API-Key") != "" {
			t.Error("leaderboard should not include auth headers")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	c := NewClient("", WithBaseURL(srv.URL))
	result, err := c.GetLeaderboard(context.Background(), "weekly", 10, 0)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Username != "top" {
		t.Errorf("expected first user top, got %q", result[0].Username)
	}
}

func TestGetLeaderboard_LimitClamping(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		expected string
	}{
		{"below_min", 0, "1"},
		{"negative", -5, "1"},
		{"above_max", 200, "100"},
		{"normal", 50, "50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got := r.URL.Query().Get("limit")
				if got != tt.expected {
					t.Errorf("expected limit %s, got %s", tt.expected, got)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("[]"))
			}))
			defer srv.Close()

			c := NewClient("", WithBaseURL(srv.URL))
			_, err := c.GetLeaderboard(context.Background(), "daily", tt.limit, 0)
			if err != nil {
				t.Fatalf("GetLeaderboard failed: %v", err)
			}
		})
	}
}

// --- SubmitSession Tests ---

func TestSubmitSession_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sessions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Verify auth headers.
		if r.Header.Get("X-API-Key") == "" {
			t.Error("missing X-API-Key header")
		}
		if r.Header.Get("X-Signature") == "" {
			t.Error("missing X-Signature header")
		}
		if r.Header.Get("User-Agent") != UserAgent {
			t.Errorf("expected User-Agent %q, got %q", UserAgent, r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient("test-key", WithBaseURL(srv.URL))
	session := &SessionSubmission{
		SessionHash:  "abc123",
		EndedAt:      "2026-01-15T10:00:00Z",
		InputTokens:  1000,
		OutputTokens: 500,
	}
	err := c.SubmitSession(context.Background(), session)
	if err != nil {
		t.Fatalf("SubmitSession failed: %v", err)
	}
}

func TestSubmitSession_TokenClamping(t *testing.T) {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		if decErr := decoder.Decode(&receivedBody); decErr != nil {
			t.Errorf("decode body: %v", decErr)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient("test-key", WithBaseURL(srv.URL))
	session := &SessionSubmission{
		SessionHash:         "hash",
		EndedAt:             "2026-01-15T10:00:00Z",
		InputTokens:         999_999_999_999,
		OutputTokens:        999_999_999_999,
		CacheCreationTokens: 999_999_999_999,
		CacheReadTokens:     999_999_999_999,
	}
	err := c.SubmitSession(context.Background(), session)
	if err != nil {
		t.Fatal(err)
	}

	// Verify tokens were clamped.
	maxInputFloat := float64(MaxInputTokens)
	maxOutputFloat := float64(MaxOutputTokens)
	if v, ok := receivedBody["inputTokens"].(float64); ok && v > maxInputFloat {
		t.Errorf("inputTokens should be clamped to %v, got %v", maxInputFloat, v)
	}
	if v, ok := receivedBody["outputTokens"].(float64); ok && v > maxOutputFloat {
		t.Errorf("outputTokens should be clamped to %v, got %v", maxOutputFloat, v)
	}
}

func TestSubmitSession_NoAPIKey(t *testing.T) {
	c := NewClient("", WithBaseURL("http://localhost"))
	session := &SessionSubmission{SessionHash: "hash", EndedAt: "now"}
	err := c.SubmitSession(context.Background(), session)
	if err == nil {
		t.Fatal("expected error with empty API key")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

// --- SubmitSessionsBatch Tests ---

func TestSubmitSessionsBatch_Success(t *testing.T) {
	batchResult := BatchResult{Success: true, Processed: 2, Succeeded: 2, Failed: 0}
	body, err := json.Marshal(batchResult)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sessions/batch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	c := NewClient("key", WithBaseURL(srv.URL))
	sessions := []*SessionSubmission{
		{SessionHash: "h1", EndedAt: "now", InputTokens: 100},
		{SessionHash: "h2", EndedAt: "now", InputTokens: 200},
	}
	result, err := c.SubmitSessionsBatch(context.Background(), sessions)
	if err != nil {
		t.Fatalf("SubmitSessionsBatch failed: %v", err)
	}
	if !result.Success {
		t.Error("expected success true")
	}
	if result.Processed != 2 {
		t.Errorf("expected 2 processed, got %d", result.Processed)
	}
}

func TestSubmitSessionsBatch_ExceedsMax(t *testing.T) {
	c := NewClient("key")
	sessions := make([]*SessionSubmission, MaxBatchSize+1)
	for i := range sessions {
		sessions[i] = &SessionSubmission{SessionHash: "h"}
	}

	_, err := c.SubmitSessionsBatch(context.Background(), sessions)
	if err == nil {
		t.Fatal("expected error for exceeding batch size")
	}
	if _, ok := err.(*ClientError); !ok {
		t.Errorf("expected ClientError, got %T", err)
	}
}

func TestSubmitSessionsBatch_ExactMax(t *testing.T) {
	batchResult := BatchResult{Success: true, Processed: MaxBatchSize, Succeeded: MaxBatchSize}
	body, err := json.Marshal(batchResult)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	c := NewClient("key", WithBaseURL(srv.URL))
	sessions := make([]*SessionSubmission, MaxBatchSize)
	for i := range sessions {
		sessions[i] = &SessionSubmission{SessionHash: "h", EndedAt: "now"}
	}

	result, err := c.SubmitSessionsBatch(context.Background(), sessions)
	if err != nil {
		t.Fatalf("batch of exactly %d should succeed: %v", MaxBatchSize, err)
	}
	if result.Processed != MaxBatchSize {
		t.Errorf("expected %d processed, got %d", MaxBatchSize, result.Processed)
	}
}

// --- ComputeSessionHash Tests ---

func TestComputeSessionHash_Deterministic(t *testing.T) {
	h1 := ComputeSessionHash("2026-01-15T10:00:00Z", 1000, 500, 200, 100, "claude-sonnet-4-20250514")
	h2 := ComputeSessionHash("2026-01-15T10:00:00Z", 1000, 500, 200, 100, "claude-sonnet-4-20250514")

	// Same inputs should produce the same hash (deterministic).
	if h1 != h2 {
		t.Errorf("same inputs should produce same hash, got %q and %q", h1, h2)
	}
}

func TestComputeSessionHash_DifferentInputs(t *testing.T) {
	h1 := ComputeSessionHash("2026-01-15T10:00:00Z", 1000, 500, 200, 100, "claude-sonnet-4-20250514")
	h2 := ComputeSessionHash("2026-01-15T11:00:00Z", 1000, 500, 200, 100, "claude-sonnet-4-20250514")

	if h1 == h2 {
		t.Error("different endedAt should produce different hashes")
	}

	h3 := ComputeSessionHash("2026-01-15T10:00:00Z", 1000, 500, 200, 100, "claude-opus-4-20250514")
	if h1 == h3 {
		t.Error("different modelName should produce different hashes")
	}

	h4 := ComputeSessionHash("2026-01-15T10:00:00Z", 1000, 500, 300, 100, "claude-sonnet-4-20250514")
	if h1 == h4 {
		t.Error("different cacheCreationTokens should produce different hashes")
	}
}

func TestComputeSessionHash_Format(t *testing.T) {
	hash := ComputeSessionHash("2026-01-15T10:00:00Z", 100, 50, 10, 5, "claude-sonnet-4-20250514")

	if len(hash) != 64 { // SHA-256 hex = 64 chars
		t.Errorf("expected 64 hex chars, got %d", len(hash))
	}

	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("hash contains non-hex character: %c", c)
		}
	}
}

// --- Token Clamping Tests ---

func TestClampTokensTo(t *testing.T) {
	max := int64(MaxInputTokens)
	tests := []struct {
		name     string
		input    int64
		max      int64
		expected int64
	}{
		{"below_max", 1000, max, 1000},
		{"at_max", max, max, max},
		{"above_max", max + 1, max, max},
		{"zero", 0, max, 0},
		{"large", 999_999_999_999, max, max},
		{"output_max", int64(MaxOutputTokens) + 1, int64(MaxOutputTokens), int64(MaxOutputTokens)},
		{"cache_max", int64(MaxCacheTokens) + 1, int64(MaxCacheTokens), int64(MaxCacheTokens)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampTokensTo(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("clampTokensTo(%d, %d) = %d, want %d", tt.input, tt.max, got, tt.expected)
			}
		})
	}
}

func TestClampSessionTokens(t *testing.T) {
	s := &SessionSubmission{
		InputTokens:         999_999_999_999,
		OutputTokens:        999_999_999_999,
		CacheCreationTokens: 999_999_999_999,
		CacheReadTokens:     999_999_999_999,
	}
	clampSessionTokens(s)

	if s.InputTokens != int64(MaxInputTokens) {
		t.Errorf("InputTokens not clamped: got %d, want %d", s.InputTokens, MaxInputTokens)
	}
	if s.OutputTokens != int64(MaxOutputTokens) {
		t.Errorf("OutputTokens not clamped: got %d, want %d", s.OutputTokens, MaxOutputTokens)
	}
	if s.CacheCreationTokens != int64(MaxCacheTokens) {
		t.Errorf("CacheCreationTokens not clamped: got %d, want %d", s.CacheCreationTokens, MaxCacheTokens)
	}
	if s.CacheReadTokens != int64(MaxCacheTokens) {
		t.Errorf("CacheReadTokens not clamped: got %d, want %d", s.CacheReadTokens, MaxCacheTokens)
	}
}

// --- Error Type Tests ---

func TestClientError_Error(t *testing.T) {
	err := &ClientError{Message: "test error"}
	expected := "rank client error: test error"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestAuthenticationError_Error(t *testing.T) {
	err := &AuthenticationError{Message: "no key"}
	expected := "rank authentication error: no key"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestApiError_Error(t *testing.T) {
	err := &ApiError{Message: "not found", StatusCode: 404}
	expected := "rank API error (status 404): not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestApiError_WithDetails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid input","field":"email"}`))
	}))
	defer srv.Close()

	c := NewClient("", WithBaseURL(srv.URL))
	_, err := c.CheckStatus(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*ApiError)
	if !ok {
		t.Fatalf("expected ApiError, got %T", err)
	}
	if apiErr.Message != "invalid input" {
		t.Errorf("expected message from JSON body, got %q", apiErr.Message)
	}
	if apiErr.Details["field"] != "email" {
		t.Errorf("expected field detail, got %v", apiErr.Details)
	}
}

// --- Request Header Tests ---

func TestDoRequest_UserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != UserAgent {
			t.Errorf("expected User-Agent %q, got %q", UserAgent, r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := NewClient("", WithBaseURL(srv.URL))
	_, _ = c.CheckStatus(context.Background())
}

func TestDoRequest_ContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := NewClient("", WithBaseURL(srv.URL))
	_, _ = c.CheckStatus(context.Background())
}

// --- Context Cancellation Tests ---

func TestCheckStatus_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	c := NewClient("", WithBaseURL(srv.URL))
	_, err := c.CheckStatus(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// --- Data Model JSON Tests ---

func TestApiStatus_JSON(t *testing.T) {
	raw := `{"status":"ok","version":"2.0","timestamp":"2026-01-15T10:00:00Z"}`
	var s ApiStatus
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Status != "ok" {
		t.Errorf("expected ok, got %q", s.Status)
	}
	if s.Version != "2.0" {
		t.Errorf("expected 2.0, got %q", s.Version)
	}
}

func TestUserRank_JSON(t *testing.T) {
	raw := `{
		"username": "test",
		"rankings": {
			"daily": {"position": 1, "compositeScore": 99.5, "totalParticipants": 100}
		},
		"stats": {
			"totalTokens": 50000,
			"totalSessions": 10,
			"inputTokens": 30000,
			"outputTokens": 20000
		},
		"lastUpdated": "2026-01-15"
	}`
	var ur UserRank
	if err := json.Unmarshal([]byte(raw), &ur); err != nil {
		t.Fatal(err)
	}
	if ur.Username != "test" {
		t.Errorf("expected test, got %q", ur.Username)
	}
	if ur.Rankings == nil {
		t.Fatal("expected rankings to be non-nil")
	}
	if ur.Rankings.Daily == nil {
		t.Fatal("expected daily rank info")
	}
	if ur.Rankings.Daily.Position != 1 {
		t.Errorf("expected position 1, got %d", ur.Rankings.Daily.Position)
	}
}

func TestLeaderboardEntry_JSON(t *testing.T) {
	raw := `{"rank":1,"username":"leader","totalTokens":100000,"compositeScore":99.9,"sessionCount":100,"isPrivate":false}`
	var e LeaderboardEntry
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		t.Fatal(err)
	}
	if e.Rank != 1 {
		t.Errorf("expected rank 1, got %d", e.Rank)
	}
	if e.IsPrivate {
		t.Error("expected isPrivate false")
	}
}

func TestSessionSubmission_JSON(t *testing.T) {
	s := SessionSubmission{
		SessionHash:  "hash123",
		EndedAt:      "2026-01-15T10:00:00Z",
		InputTokens:  1000,
		OutputTokens: 500,
		ModelName:    "claude-opus-4-5-20251101",
		ToolUsage:    map[string]int{"read": 5, "write": 3},
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["sessionHash"] != "hash123" {
		t.Errorf("expected hash123, got %v", parsed["sessionHash"])
	}
	if parsed["modelName"] != "claude-opus-4-5-20251101" {
		t.Errorf("expected model name, got %v", parsed["modelName"])
	}
}

func TestBatchResult_JSON(t *testing.T) {
	raw := `{"success":true,"processed":10,"succeeded":9,"failed":1}`
	var br BatchResult
	if err := json.Unmarshal([]byte(raw), &br); err != nil {
		t.Fatal(err)
	}
	if !br.Success {
		t.Error("expected success true")
	}
	if br.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", br.Failed)
	}
}
