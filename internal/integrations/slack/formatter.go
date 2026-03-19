package slack

import "github.com/modu-ai/moai-adk/internal/integrations"

// SlackMessage represents a Slack Block Kit message.
type SlackMessage struct {
	Text   string  `json:"text"`
	Blocks []Block `json:"blocks,omitempty"`
}

// Block represents a Slack Block Kit block.
type Block struct {
	Type string `json:"type"`
	Text *Text  `json:"text,omitempty"`
}

// Text represents a Slack Block Kit text object.
type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// FormatMessage formats a NotifyEvent into a Slack Block Kit message.
func FormatMessage(event integrations.NotifyEvent) SlackMessage {
	emoji := eventEmoji(event.Type)
	header := emoji + " " + event.Title

	blocks := []Block{
		{Type: "header", Text: &Text{Type: "plain_text", Text: header}},
		{Type: "section", Text: &Text{Type: "mrkdwn", Text: event.Message}},
	}

	if len(event.Details) > 0 {
		detailText := ""
		for k, v := range event.Details {
			detailText += "*" + k + "*: " + v + "\n"
		}
		blocks = append(blocks, Block{
			Type: "section",
			Text: &Text{Type: "mrkdwn", Text: detailText},
		})
	}

	return SlackMessage{
		Text:   header,
		Blocks: blocks,
	}
}

func eventEmoji(eventType string) string {
	switch eventType {
	case "spec_complete":
		return "\u2705"
	case "quality_failure":
		return "\u274c"
	case "pr_created":
		return "\U0001f500"
	case "budget_alert":
		return "\u26a0\ufe0f"
	default:
		return "\u2139\ufe0f"
	}
}
