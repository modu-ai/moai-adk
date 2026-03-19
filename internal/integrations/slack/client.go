// Package slack provides Slack incoming webhook integration for MoAI-ADK notifications.
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/modu-ai/moai-adk/internal/integrations"
)

// Client sends notifications to Slack via incoming webhooks.
type Client struct {
	webhookURL string
	httpClient *http.Client
	events     []string
}

// NewClient creates a new Slack webhook client.
func NewClient(webhookURL string, events []string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		events:     events,
	}
}

// Send sends a notification event to Slack.
func (c *Client) Send(event integrations.NotifyEvent) error {
	if !c.shouldNotify(event.Type) {
		return nil
	}

	payload := FormatMessage(event)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}

	resp, err := c.httpClient.Post(c.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// Test sends a test message to verify connectivity.
func (c *Client) Test() error {
	return c.Send(integrations.NotifyEvent{
		Type:    "test",
		Title:   "MoAI-ADK Test",
		Message: "Slack integration is working correctly.",
	})
}

// Name returns the notifier name.
func (c *Client) Name() string { return "Slack" }

// IsEnabled returns whether the client is configured.
func (c *Client) IsEnabled() bool { return c.webhookURL != "" }

func (c *Client) shouldNotify(eventType string) bool {
	if eventType == "test" {
		return true
	}
	for _, e := range c.events {
		if e == eventType {
			return true
		}
	}
	return false
}
