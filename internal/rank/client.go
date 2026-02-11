// Package rank provides a client for the MoAI Rank API service.
// It supports session submission, leaderboard queries, user ranking,
// and HMAC-SHA256 authenticated requests.
package rank

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Constants for the Rank API client.
const (
	MaxInputTokens  = 50_000_000
	MaxOutputTokens = 10_000_000
	MaxCacheTokens  = 100_000_000
	MaxBatchSize    = 100
	DefaultTimeout  = 30 * time.Second
	DefaultBaseURL  = "https://rank.mo.ai.kr"
	APIVersion      = "v1"
	UserAgent       = "moai-adk/1.0"
)

// --- Error Types ---

// ClientError represents a general client-side error.
type ClientError struct {
	Message string
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("rank client error: %s", e.Message)
}

// AuthenticationError represents an authentication failure.
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("rank authentication error: %s", e.Message)
}

// ApiError represents an API response error.
type ApiError struct {
	Message    string
	StatusCode int
	Details    map[string]any
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("rank API error (status %d): %s", e.StatusCode, e.Message)
}

// --- Data Models ---

// ApiStatus represents the Rank API health status response.
type ApiStatus struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp string            `json:"timestamp"`
	Endpoints map[string]string `json:"endpoints,omitempty"`
}

// RankInfo holds ranking position details for a time period.
type RankInfo struct {
	Position          int     `json:"position"`
	CompositeScore    float64 `json:"compositeScore"`
	TotalParticipants int     `json:"totalParticipants"`
}

// UserRankStats holds aggregate statistics for a user.
type UserRankStats struct {
	TotalTokens   int64 `json:"totalTokens"`
	TotalSessions int   `json:"totalSessions"`
	InputTokens   int64 `json:"inputTokens"`
	OutputTokens  int64 `json:"outputTokens"`
}

// UserRankRankings holds ranking positions for all time periods.
type UserRankRankings struct {
	Daily   *RankInfo `json:"daily,omitempty"`
	Weekly  *RankInfo `json:"weekly,omitempty"`
	Monthly *RankInfo `json:"monthly,omitempty"`
	AllTime *RankInfo `json:"allTime,omitempty"`
}

// UserRank represents the full ranking information for a user.
type UserRank struct {
	Username    string            `json:"username"`
	Rankings    *UserRankRankings `json:"rankings,omitempty"`
	Stats       *UserRankStats    `json:"stats,omitempty"`
	LastUpdated string            `json:"lastUpdated"`
}

// LeaderboardEntry represents a single entry on the leaderboard.
type LeaderboardEntry struct {
	Rank            int     `json:"rank"`
	UserID          string  `json:"userId"`
	Username        string  `json:"username"`
	AvatarURL       string  `json:"avatarUrl"`
	TotalTokens     int64   `json:"totalTokens"`
	CompositeScore  float64 `json:"compositeScore"`
	SessionCount    int     `json:"sessionCount"`
	EfficiencyScore float64 `json:"efficiencyScore"`
	IsPrivate       bool    `json:"isPrivate"`
}

// SessionSubmission holds session data for submission to the Rank API.
type SessionSubmission struct {
	SessionHash         string         `json:"sessionHash"`
	EndedAt             string         `json:"endedAt"`
	InputTokens         int64          `json:"inputTokens"`
	OutputTokens        int64          `json:"outputTokens"`
	CacheCreationTokens int64          `json:"cacheCreationTokens"`
	CacheReadTokens     int64          `json:"cacheReadTokens"`
	ModelName           string         `json:"modelName,omitempty"`
	AnonymousProjectID  string         `json:"anonymousProjectId,omitempty"`
	StartedAt           string         `json:"startedAt,omitempty"`
	DurationSeconds     int            `json:"durationSeconds,omitempty"`
	TurnCount           int            `json:"turnCount,omitempty"`
	ToolUsage           map[string]int `json:"toolUsage,omitempty"`
	CodeMetrics         map[string]int `json:"codeMetrics,omitempty"`
	DeviceID            string         `json:"deviceId,omitempty"`
}

// BatchResult represents the result of a batch session submission.
type BatchResult struct {
	Success   bool `json:"success"`
	Processed int  `json:"processed"`
	Succeeded int  `json:"succeeded"`
	Failed    int  `json:"failed"`
}

// --- Client Interface ---

// Client defines the interface for interacting with the Rank API.
type Client interface {
	CheckStatus(ctx context.Context) (*ApiStatus, error)
	GetUserRank(ctx context.Context) (*UserRank, error)
	GetLeaderboard(ctx context.Context, period string, limit, offset int) ([]LeaderboardEntry, error)
	SubmitSession(ctx context.Context, session *SessionSubmission) error
	SubmitSessionsBatch(ctx context.Context, sessions []*SessionSubmission) (*BatchResult, error)
}

// --- Client Options ---

// ClientOption configures the RankClient.
type ClientOption func(*RankClient)

// WithBaseURL sets a custom base URL for the Rank API.
func WithBaseURL(url string) ClientOption {
	return func(c *RankClient) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *RankClient) {
		c.httpClient = httpClient
	}
}

// --- Client Implementation ---

// RankClient implements Client for the MoAI Rank API.
type RankClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Compile-time interface check.
var _ Client = (*RankClient)(nil)

// NewClient creates a new RankClient.
// If apiKey is empty, authenticated endpoints will fail with AuthenticationError.
func NewClient(apiKey string, opts ...ClientOption) *RankClient {
	c := &RankClient{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	// Check environment variable override for base URL.
	if envURL := os.Getenv("MOAI_RANK_API_URL"); envURL != "" {
		c.baseURL = envURL
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// --- HMAC Authentication ---

// ComputeSignature calculates the HMAC-SHA256 signature for a request.
// Signature = HMAC-SHA256(apiKey, timestamp + ":" + body)
func ComputeSignature(apiKey, timestamp, body string) string {
	message := timestamp + ":" + body
	mac := hmac.New(sha256.New, []byte(apiKey))
	// hmac.Write never returns an error per the hash.Hash contract.
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// getAuthHeaders generates the authentication headers for an API request.
// Returns AuthenticationError if no API key is configured.
func (c *RankClient) getAuthHeaders(body string) (map[string]string, error) {
	if c.apiKey == "" {
		return nil, &AuthenticationError{Message: "API key not configured"}
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := ComputeSignature(c.apiKey, timestamp, body)

	return map[string]string{
		"X-API-Key":   c.apiKey,
		"X-Timestamp": timestamp,
		"X-Signature": signature,
	}, nil
}

// --- HTTP Helpers ---

// doRequest performs an HTTP request and returns the response body.
func (c *RankClient) doRequest(ctx context.Context, method, path string, body []byte, authenticated bool) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("create request: %v", err)}
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")

	if authenticated {
		bodyStr := ""
		if body != nil {
			bodyStr = string(body)
		}
		headers, authErr := c.getAuthHeaders(bodyStr)
		if authErr != nil {
			return nil, authErr
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("read response: %v", err)}
	}

	// Try to parse response envelope
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data,omitempty"`
		Error   *struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details,omitempty"`
		} `json:"error,omitempty"`
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, &AuthenticationError{Message: "authentication failed"}
	}

	// Try to parse as envelope
	if err := json.Unmarshal(respBody, &envelope); err == nil && (envelope.Data != nil || envelope.Error != nil) {
		// Server uses envelope format
		if resp.StatusCode >= 400 || !envelope.Success {
			apiErr := &ApiError{
				Message:    fmt.Sprintf("API returned status %d", resp.StatusCode),
				StatusCode: resp.StatusCode,
			}
			if envelope.Error != nil {
				apiErr.Message = envelope.Error.Message
				apiErr.Details = envelope.Error.Details
			}
			return nil, apiErr
		}
		if envelope.Data != nil {
			return []byte(envelope.Data), nil
		}
	}

	// Fallback: non-envelope response
	if resp.StatusCode >= 400 {
		apiErr := &ApiError{
			Message:    fmt.Sprintf("API returned status %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
		}
		var details map[string]any
		if jsonErr := json.Unmarshal(respBody, &details); jsonErr == nil {
			apiErr.Details = details
			if msg, ok := details["message"].(string); ok {
				apiErr.Message = msg
			}
		}
		return nil, apiErr
	}

	return respBody, nil
}

// --- API Methods ---

// CheckStatus verifies the Rank API service availability.
func (c *RankClient) CheckStatus(ctx context.Context) (*ApiStatus, error) {
	body, err := c.doRequest(ctx, http.MethodGet, "/api/"+APIVersion+"/status", nil, false)
	if err != nil {
		return nil, err
	}

	var status ApiStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("parse status response: %v", err)}
	}

	return &status, nil
}

// GetUserRank returns the current user's ranking information.
// Requires authentication.
func (c *RankClient) GetUserRank(ctx context.Context) (*UserRank, error) {
	body, err := c.doRequest(ctx, http.MethodGet, "/api/"+APIVersion+"/rank", nil, true)
	if err != nil {
		return nil, err
	}

	var userRank UserRank
	if err := json.Unmarshal(body, &userRank); err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("parse rank response: %v", err)}
	}

	return &userRank, nil
}

// GetLeaderboard returns the leaderboard for the specified period.
// Public API - no authentication required. Limit is clamped to [1, 100].
func (c *RankClient) GetLeaderboard(ctx context.Context, period string, limit, offset int) ([]LeaderboardEntry, error) {
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	path := fmt.Sprintf("/api/leaderboard?period=%s&limit=%d&offset=%d", period, limit, offset)
	body, err := c.doRequest(ctx, http.MethodGet, path, nil, false)
	if err != nil {
		return nil, err
	}

	var entries []LeaderboardEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("parse leaderboard response: %v", err)}
	}

	return entries, nil
}

// SubmitSession submits a single session metric.
// Requires HMAC authentication. Token fields are clamped to per-field maximums.
func (c *RankClient) SubmitSession(ctx context.Context, session *SessionSubmission) error {
	clampSessionTokens(session)

	body, err := json.Marshal(session)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("marshal session: %v", err)}
	}

	_, reqErr := c.doRequest(ctx, http.MethodPost, "/api/"+APIVersion+"/sessions", body, true)
	return reqErr
}

// SubmitSessionsBatch submits up to 100 sessions at once.
// Returns an error if more than 100 sessions are provided.
func (c *RankClient) SubmitSessionsBatch(ctx context.Context, sessions []*SessionSubmission) (*BatchResult, error) {
	if len(sessions) > MaxBatchSize {
		return nil, &ClientError{
			Message: fmt.Sprintf("batch size %d exceeds maximum of %d", len(sessions), MaxBatchSize),
		}
	}

	for _, s := range sessions {
		clampSessionTokens(s)
	}

	payload := struct {
		Sessions []*SessionSubmission `json:"sessions"`
	}{Sessions: sessions}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("marshal batch: %v", err)}
	}

	respBody, reqErr := c.doRequest(ctx, http.MethodPost, "/api/"+APIVersion+"/sessions/batch", body, true)
	if reqErr != nil {
		return nil, reqErr
	}

	var result BatchResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("parse batch response: %v", err)}
	}

	return &result, nil
}

// --- Utility Functions ---

// ComputeSessionHash generates a deterministic SHA-256 hash for a session.
// The hash is computed from immutable session properties to ensure the same
// session always produces the same hash, enabling deduplication.
func ComputeSessionHash(endedAt string, inputTokens, outputTokens, cacheCreationTokens, cacheReadTokens int64, modelName string) string {
	data := fmt.Sprintf("%s:%d:%d:%d:%d:%s", endedAt, inputTokens, outputTokens, cacheCreationTokens, cacheReadTokens, modelName)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// clampTokensTo clamps a token value to the specified maximum.
func clampTokensTo(value, max int64) int64 {
	if value > max {
		return max
	}
	return value
}

// clampSessionTokens applies per-field token clamping to a session.
func clampSessionTokens(s *SessionSubmission) {
	s.InputTokens = clampTokensTo(s.InputTokens, MaxInputTokens)
	s.OutputTokens = clampTokensTo(s.OutputTokens, MaxOutputTokens)
	s.CacheCreationTokens = clampTokensTo(s.CacheCreationTokens, MaxCacheTokens)
	s.CacheReadTokens = clampTokensTo(s.CacheReadTokens, MaxCacheTokens)
}
