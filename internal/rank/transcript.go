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

// claudeDesktopConfigDir returns the Claude Desktop (Electron app) configuration directory
// based on the platform.
func claudeDesktopConfigDir() (string, error) {
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

// claudeCodeDir returns the Claude Code CLI configuration directory (~/.claude/).
func claudeCodeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".claude"), nil
}

// globJSONL collects .jsonl files matching the given pattern, ignoring glob errors.
func globJSONL(pattern string) []string {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	return matches
}

// FindTranscripts finds all Claude Code transcript JSONL files.
// It searches multiple locations in priority order:
//  1. ~/.claude/projects/*/*.jsonl  (Claude Code CLI - new format, UUID-based filenames)
//  2. ~/.claude/transcripts/*.jsonl (Claude Code CLI - old/legacy format)
//  3. Claude Desktop paths           (fallback for Desktop users)
func FindTranscripts() ([]string, error) {
	seen := make(map[string]struct{})
	var results []string

	addUnique := func(paths []string) {
		for _, p := range paths {
			if _, exists := seen[p]; !exists {
				seen[p] = struct{}{}
				results = append(results, p)
			}
		}
	}

	// Priority 1: Claude Code CLI new format (~/.claude/projects/*/*.jsonl)
	if codeDir, err := claudeCodeDir(); err == nil {
		addUnique(globJSONL(filepath.Join(codeDir, "projects", "*", "*.jsonl")))
	}

	// Priority 2: Claude Code CLI old/legacy format (~/.claude/transcripts/*.jsonl)
	if codeDir, err := claudeCodeDir(); err == nil {
		addUnique(globJSONL(filepath.Join(codeDir, "transcripts", "*.jsonl")))
	}

	// Priority 3: Claude Desktop paths (fallback)
	if desktopDir, err := claudeDesktopConfigDir(); err == nil {
		addUnique(globJSONL(filepath.Join(desktopDir, "*", "transcripts", "*.jsonl")))
	}

	return results, nil
}

// isValidSessionID validates a session ID to prevent path traversal attacks.
// Valid session IDs contain only alphanumeric characters, hyphens, and underscores.
func isValidSessionID(sessionID string) bool {
	if sessionID == "" || len(sessionID) > 128 {
		return false
	}
	for _, c := range sessionID {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return false
		}
	}
	return true
}

// FindTranscriptForSession finds the transcript file for a specific session ID.
// It searches Claude Code CLI paths first, then falls back to Desktop paths.
// Returns the path if found, empty string otherwise.
func FindTranscriptForSession(sessionID string) string {
	if !isValidSessionID(sessionID) {
		return ""
	}

	// Priority 1: Claude Code CLI new format (~/.claude/projects/*/<sessionID>*.jsonl)
	if codeDir, err := claudeCodeDir(); err == nil {
		pattern := filepath.Join(codeDir, "projects", "*", sessionID+"*.jsonl")
		if matches := globJSONL(pattern); len(matches) > 0 {
			return matches[0]
		}
	}

	// Priority 2: Claude Code CLI old/legacy format (~/.claude/transcripts/<sessionID>*.jsonl)
	if codeDir, err := claudeCodeDir(); err == nil {
		pattern := filepath.Join(codeDir, "transcripts", sessionID+"*.jsonl")
		if matches := globJSONL(pattern); len(matches) > 0 {
			return matches[0]
		}
	}

	// Priority 3: Claude Desktop paths (fallback)
	if desktopDir, err := claudeDesktopConfigDir(); err == nil {
		pattern := filepath.Join(desktopDir, "*", "transcripts", sessionID+"*.jsonl")
		if matches := globJSONL(pattern); len(matches) > 0 {
			return matches[0]
		}
	}

	return ""
}
