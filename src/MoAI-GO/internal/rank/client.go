package rank

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RankInfo contains ranking data for a specific time period
type RankInfo struct {
	Position          int     `json:"position"`
	TotalParticipants int     `json:"total_participants"`
	CompositeScore    float64 `json:"composite_score"`
}

// UserRankResponse is the API response for user rank
type UserRankResponse struct {
	Username string              `json:"username"`
	Ranks    map[string]RankInfo `json:"ranks"` // daily, weekly, monthly, all_time
	Tokens   TokenStats          `json:"tokens"`
}

// TokenStats contains token usage statistics
type TokenStats struct {
	Total  int64 `json:"total"`
	Input  int64 `json:"input"`
	Output int64 `json:"output"`
}

// defaultBaseURL is the default API base URL, overridable in tests
var defaultBaseURL = APIBaseURL

// Client handles HTTP communication with the MoAI Rank API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new rank API client
func NewClient(apiKey string) *Client {
	return NewClientWithURL(apiKey, defaultBaseURL)
}

// NewClientWithURL creates a rank API client with a custom base URL
func NewClientWithURL(apiKey, baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetRank fetches the user's rank information
func (c *Client) GetRank() (*UserRankResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/users/rank", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rank server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("rank API error (status %d): failed to read response body", resp.StatusCode)
		}
		return nil, fmt.Errorf("rank API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result UserRankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse rank response: %w", err)
	}
	return &result, nil
}
