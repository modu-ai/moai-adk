// Package linear provides Linear issue tracking integration for MoAI-ADK.
package linear

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/modu-ai/moai-adk/internal/integrations"
)

// Client interacts with the Linear GraphQL API.
type Client struct {
	apiKeyEnv  string
	teamID     string
	httpClient *http.Client
}

// NewClient creates a new Linear API client.
func NewClient(apiKeyEnv, teamID string) *Client {
	return &Client{
		apiKeyEnv:  apiKeyEnv,
		teamID:     teamID,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Send sends a notification event to Linear.
func (c *Client) Send(event integrations.NotifyEvent) error {
	apiKey := os.Getenv(c.apiKeyEnv)
	if apiKey == "" {
		return fmt.Errorf("linear API key not set in env var %s", c.apiKeyEnv)
	}

	switch event.Type {
	case "spec_complete":
		return c.createIssue(apiKey, event)
	case "quality_failure", "pr_created":
		return c.addComment(apiKey, event)
	default:
		return nil
	}
}

// Test verifies Linear connectivity.
func (c *Client) Test() error {
	apiKey := os.Getenv(c.apiKeyEnv)
	if apiKey == "" {
		return fmt.Errorf("linear API key not set in env var %s", c.apiKeyEnv)
	}

	query := `{"query": "{ viewer { id name } }"}`
	_, err := c.graphQL(apiKey, query)
	return err
}

// Name returns the notifier name.
func (c *Client) Name() string { return "Linear" }

// IsEnabled returns whether the client is configured.
func (c *Client) IsEnabled() bool {
	return os.Getenv(c.apiKeyEnv) != "" && c.teamID != ""
}

// graphQLRequest represents a GraphQL request with variables to prevent injection.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

func (c *Client) createIssue(apiKey string, event integrations.NotifyEvent) error {
	reqBody := graphQLRequest{
		Query: `mutation CreateIssue($teamId: String!, $title: String!, $description: String!) {
			issueCreate(input: { teamId: $teamId, title: $title, description: $description }) {
				success
				issue { id identifier url }
			}
		}`,
		Variables: map[string]any{
			"teamId":      c.teamID,
			"title":       event.Title,
			"description": event.Message,
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal linear request: %w", err)
	}

	_, err = c.graphQL(apiKey, string(payload))
	return err
}

func (c *Client) addComment(apiKey string, event integrations.NotifyEvent) error {
	if event.SpecID == "" {
		return nil
	}
	// Comment addition requires issue ID lookup — simplified for now
	return nil
}

func (c *Client) graphQL(apiKey, query string) ([]byte, error) {
	req, err := http.NewRequest("POST", "https://api.linear.app/graphql", bytes.NewBufferString(query))
	if err != nil {
		return nil, fmt.Errorf("create linear request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linear API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linear API returned status %d", resp.StatusCode)
	}

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode linear response: %w", err)
	}
	return result, nil
}
