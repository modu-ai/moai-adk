// Package rank provides transcript parsing for MoAI Rank session submission.
package rank

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TranscriptUsage represents token usage extracted from a Claude Code transcript.
type TranscriptUsage struct {
	InputTokens         int64  `json:"input_tokens"`
	OutputTokens        int64  `json:"output_tokens"`
	CacheCreationTokens int64  `json:"cache_creation_tokens"`
	CacheReadTokens     int64  `json:"cache_read_tokens"`
	ModelName           string `json:"model_name"`
	StartedAt           string `json:"started_at,omitempty"`
	EndedAt             string `json:"ended_at,omitempty"`
	DurationSeconds     int64  `json:"duration_seconds,omitempty"`
	TurnCount           int    `json:"turn_count,omitempty"`
}

// transcriptMessage represents a single line in the JSONL transcript file.
type transcriptMessage struct {
	Timestamp string        `json:"timestamp"`
	Type      string        `json:"type"`
	Message   transcriptMsg `json:"message"`
	Model     string        `json:"model"`
}

// transcriptMsg represents the message content with usage data.
type transcriptMsg struct {
	Usage *transcriptUsage `json:"usage"`
	Model string           `json:"model"`
}

// transcriptUsage represents token usage information.
type transcriptUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

// ParseTranscript parses a Claude Code transcript JSONL file and extracts token usage.
// The transcript file contains one JSON object per line, with token usage in message.usage fields.
func ParseTranscript(transcriptPath string) (*TranscriptUsage, error) {
	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer func() {
		// Close errors are ignored for read-only files
		_ = file.Close()
	}()

	usage := &TranscriptUsage{}
	var firstTimestamp, lastTimestamp string
	turnCount := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg transcriptMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Skip invalid lines
			continue
		}

		// Track timestamps for duration calculation
		if msg.Timestamp != "" {
			if firstTimestamp == "" {
				firstTimestamp = msg.Timestamp
			}
			lastTimestamp = msg.Timestamp
		}

		// Count user turns
		if msg.Type == "user" {
			turnCount++
		}

		// Extract model name
		model := msg.Model
		if model != "" && usage.ModelName == "" {
			usage.ModelName = model
		}
		if msg.Message.Model != "" && usage.ModelName == "" {
			usage.ModelName = msg.Message.Model
		}

		// Extract token usage
		if msg.Message.Usage != nil {
			usage.InputTokens += msg.Message.Usage.InputTokens
			usage.OutputTokens += msg.Message.Usage.OutputTokens
			usage.CacheCreationTokens += msg.Message.Usage.CacheCreationInputTokens
			usage.CacheReadTokens += msg.Message.Usage.CacheReadInputTokens
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan transcript: %w", err)
	}

	// Set timing metadata
	usage.StartedAt = firstTimestamp
	usage.EndedAt = lastTimestamp
	usage.TurnCount = turnCount

	// Calculate duration if timestamps are available
	if firstTimestamp != "" && lastTimestamp != "" {
		start, err := time.Parse(time.RFC3339Nano, firstTimestamp)
		if err == nil {
			end, err := time.Parse(time.RFC3339Nano, lastTimestamp)
			if err == nil {
				usage.DurationSeconds = int64(end.Sub(start).Seconds())
			}
		}
	}

	return usage, nil
}

// claudeConfigDir returns the Claude Code configuration directory based on the platform.
func claudeConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	switch goos := runtime.GOOS; goos {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Claude"), nil
	case "linux":
		return filepath.Join(homeDir, ".config", "Claude"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "Claude"), nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", goos)
	}
}

// FindTranscripts finds all Claude Code transcript JSONL files in the user's home directory.
// Claude Code stores transcripts in platform-specific locations:
// - macOS: ~/Library/Application Support/Claude/*/transcripts/*.jsonl
// - Linux: ~/.config/Claude/*/transcripts/*.jsonl
// - Windows: %APPDATA%\Claude\*\transcripts\*.jsonl
func FindTranscripts() ([]string, error) {
	configDir, err := claudeConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get Claude config directory: %w", err)
	}

	// Claude Code transcript directory pattern
	pattern := filepath.Join(configDir, "*", "transcripts", "*.jsonl")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob transcripts: %w", err)
	}

	return matches, nil
}

// FindTranscriptForSession finds the transcript file for a specific session ID.
// Returns the path if found, empty string otherwise.
func FindTranscriptForSession(sessionID string) string {
	configDir, err := claudeConfigDir()
	if err != nil {
		return ""
	}

	// Search in all Claude directories
	pattern := filepath.Join(configDir, "*", "transcripts", sessionID+"*.jsonl")

	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}

	// Return the first match (most recent)
	return matches[0]
}
