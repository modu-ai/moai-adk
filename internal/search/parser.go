//go:build !race

package search

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// Message represents a single conversation message parsed from a JSONL session file.
type Message struct {
	SessionID   string
	Role        string
	Text        string
	Timestamp   string
	GitBranch   string
	ProjectPath string
}

// noisePatterns lists XML tag prefixes used internally by Claude Code.
// Messages containing these patterns are excluded from the search index.
var noisePatterns = []string{
	"<local-command-caveat>",
	"<command-name>",
	"<system-reminder>",
	"<function_calls>",
}

// minTextLength is the minimum rune count for a message to be indexed.
// Messages shorter than this threshold carry little search value.
const minTextLength = 20

// jsonlRecord represents a single record in a JSONL session file.
type jsonlRecord struct {
	Type      string       `json:"type"`
	Timestamp string       `json:"timestamp"`
	SessionID string       `json:"sessionId"`
	Message   jsonlMessage `json:"message"`
}

// jsonlMessage is the message object within a JSONL record.
type jsonlMessage struct {
	Role    string         `json:"role"`
	Content []jsonlContent `json:"content"`
}

// jsonlContent is a content block within a message.
type jsonlContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ParseJSONL parses a JSONL session file and returns a list of searchable messages.
// gitBranch and projectPath are attached as metadata to each message.
//
// Filtering rules:
//   - Records whose type is not "user" or "assistant" are skipped.
//   - Messages containing XML noise tags are excluded.
//   - Messages whose trimmed text is shorter than minTextLength runes are excluded.
//   - Invalid JSON lines are silently skipped.
func ParseJSONL(path, gitBranch, projectPath string) ([]Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var messages []Message
	scanner := bufio.NewScanner(f)
	// Increase the line buffer to 4 MB to handle large tool-output messages.
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var rec jsonlRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			// Skip invalid JSON lines.
			continue
		}

		// Only process user and assistant message types.
		if rec.Type != "user" && rec.Type != "assistant" {
			continue
		}

		// Prefer message.role; fall back to record type.
		role := rec.Message.Role
		if role == "" {
			role = rec.Type
		}
		if role != "user" && role != "assistant" {
			continue
		}

		// Extract text-type content blocks only.
		var textParts []string
		for _, c := range rec.Message.Content {
			if c.Type == "text" && c.Text != "" {
				textParts = append(textParts, c.Text)
			}
		}
		if len(textParts) == 0 {
			continue
		}

		text := strings.Join(textParts, " ")

		// Exclude messages containing internal noise patterns.
		if containsNoise(text) {
			continue
		}

		// Exclude messages shorter than minTextLength runes.
		if len([]rune(strings.TrimSpace(text))) < minTextLength {
			continue
		}

		messages = append(messages, Message{
			SessionID:   rec.SessionID,
			Role:        role,
			Text:        text,
			Timestamp:   rec.Timestamp,
			GitBranch:   gitBranch,
			ProjectPath: projectPath,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return an empty slice (not nil) so callers do not need a nil check.
	if messages == nil {
		messages = []Message{}
	}

	return messages, nil
}

// containsNoise returns true if text contains any of the noise tag patterns.
func containsNoise(text string) bool {
	for _, pattern := range noisePatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}
