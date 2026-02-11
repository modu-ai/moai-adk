package hook

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/modu-ai/moai-adk/internal/rank"
)

// rankSessionHandler processes SessionEnd events and submits metrics to MoAI Rank API.
// It checks exclusion patterns and submits session data non-blocking.
// Errors are logged but don't break the hook chain (per REQ-HOOK-034).
type rankSessionHandler struct {
	patternStore *rank.PatternStore
	credStore    rank.CredentialStore
}

// NewRankSessionHandler creates a new rank session handler.
// The patternStore is used to check if a project should be excluded from metrics.
// The credStore is used to load API credentials for submission.
func NewRankSessionHandler(patternStore *rank.PatternStore, credStore rank.CredentialStore) Handler {
	return &rankSessionHandler{
		patternStore: patternStore,
		credStore:    credStore,
	}
}

// EventType returns EventSessionEnd.
func (h *rankSessionHandler) EventType() EventType {
	return EventSessionEnd
}

// Handle processes a SessionEnd event and submits metrics to MoAI Rank.
// SessionEnd hooks return empty JSON {} per Claude Code protocol.
// Errors are non-blocking: log warnings and return empty output.
func (h *rankSessionHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Check if user has credentials (not registered = skip silently)
	apiKey, err := h.credStore.GetAPIKey()
	if err != nil || apiKey == "" {
		slog.Debug("rank: no credentials configured, skipping session submission")
		return &HookOutput{}, nil
	}

	// Get project path from input (prefer ProjectDir, fallback to CWD)
	projectPath := input.ProjectDir
	if projectPath == "" {
		projectPath = input.CWD
	}

	// Check if project is excluded from metrics
	if h.patternStore != nil && h.patternStore.ShouldExclude(projectPath) {
		slog.Debug("rank: project excluded from metrics", "project", projectPath)
		return &HookOutput{}, nil
	}

	// Extract session metrics from HookInput
	// Note: HookInput doesn't contain all token fields directly, so we create
	// a minimal session submission. Future enhancements can parse additional data.
	submission, err := h.buildSessionSubmission(input)
	if err != nil {
		slog.Warn("rank: failed to build session submission", "error", err)
		return &HookOutput{}, nil
	}

	// Create rank client and submit session
	client := rank.NewClient(apiKey)
	if err := client.SubmitSession(ctx, submission); err != nil {
		slog.Warn("rank: failed to submit session", "error", err)
		return &HookOutput{}, nil
	}

	slog.Info("rank: session submitted successfully",
		"session_id", input.SessionID,
		"project", anonymizePath(projectPath),
	)

	// Mark session in sync state to prevent re-submission during sync
	transcriptPath := rank.FindTranscriptForSession(input.SessionID)
	if transcriptPath != "" {
		if syncState, syncErr := rank.NewSyncState(""); syncErr == nil {
			_ = syncState.MarkSynced(transcriptPath)
			_ = syncState.Save()
		}
	}

	// SessionEnd hooks return empty JSON {} per Claude Code protocol
	return &HookOutput{}, nil
}

// buildSessionSubmission creates a SessionSubmission from HookInput.
// It attempts to parse the transcript file for actual token usage.
func (h *rankSessionHandler) buildSessionSubmission(input *HookInput) (*rank.SessionSubmission, error) {
	now := time.Now()
	endedAt := now.Format(time.RFC3339)

	// Try to parse transcript for actual token usage
	var inputTokens, outputTokens, cacheCreation, cacheRead int64
	var startedAt string
	var durationSeconds int64
	var turnCount int
	var modelName string

	// Find transcript file for this session
	transcriptPath := rank.FindTranscriptForSession(input.SessionID)
	if transcriptPath != "" {
		if usage, err := rank.ParseTranscript(transcriptPath); err == nil {
			inputTokens = usage.InputTokens
			outputTokens = usage.OutputTokens
			cacheCreation = usage.CacheCreationTokens
			cacheRead = usage.CacheReadTokens
			startedAt = usage.StartedAt
			durationSeconds = usage.DurationSeconds
			turnCount = usage.TurnCount
			if usage.ModelName != "" {
				modelName = usage.ModelName
			}
		}
		// If parsing fails, fall back to zero values
	}

	// Use model from input if not found in transcript
	if modelName == "" && input.Model != "" {
		modelName = input.Model
	}

	// Generate session hash for deduplication
	sessionHash := rank.ComputeSessionHash(endedAt, inputTokens, outputTokens, cacheCreation, cacheRead, modelName)

	// Anonymize project path
	projectPath := input.ProjectDir
	if projectPath == "" {
		projectPath = input.CWD
	}
	anonymousProjectID := anonymizePath(projectPath)

	// Get device info for multi-device tracking
	deviceInfo := rank.GetDeviceInfo()

	submission := &rank.SessionSubmission{
		SessionHash:         sessionHash,
		EndedAt:             endedAt,
		InputTokens:         inputTokens,
		OutputTokens:        outputTokens,
		CacheCreationTokens: cacheCreation,
		CacheReadTokens:     cacheRead,
		AnonymousProjectID:  anonymousProjectID,
		StartedAt:           startedAt,
		DurationSeconds:     int(durationSeconds),
		TurnCount:           turnCount,
		ModelName:           modelName,
		DeviceID:            deviceInfo.DeviceID,
	}

	return submission, nil
}

// anonymizePath creates a one-way hash of the project path for privacy.
// This ensures the actual project path is never transmitted to the rank service.
func anonymizePath(path string) string {
	if path == "" {
		return ""
	}

	// Normalize the path
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Create SHA-256 hash for one-way anonymization
	hash := sha256.Sum256([]byte(absPath))
	fullHash := hex.EncodeToString(hash[:])
	if len(fullHash) > 16 {
		return fullHash[:16]
	}
	return fullHash
}

// --- Lazy Initialization Helpers ---

// InitRankSessionHandler creates and initializes a rank session handler
// with the default credential store and pattern store locations.
// This is a convenience function for deps initialization.
func InitRankSessionHandler() (Handler, error) {
	// Create credential store with default location
	credStore := rank.NewFileCredentialStore("")

	// Create pattern store with default location
	patternStore, err := rank.NewPatternStore("")
	if err != nil {
		return nil, fmt.Errorf("create pattern store: %w", err)
	}

	return NewRankSessionHandler(patternStore, credStore), nil
}

// InitRankSessionHandlerWithStores creates a rank session handler with
// explicit stores (useful for testing and custom configurations).
func InitRankSessionHandlerWithStores(patternStore *rank.PatternStore, credStore rank.CredentialStore) Handler {
	return NewRankSessionHandler(patternStore, credStore)
}

// EnsureRankSessionHandler initializes the rank session handler if credentials exist.
// Returns nil if no credentials are configured (not an error).
// This is the recommended way to initialize the handler for production use.
func EnsureRankSessionHandler() (Handler, error) {
	credStore := rank.NewFileCredentialStore("")
	if !credStore.HasCredentials() {
		// No credentials configured, return nil (not an error)
		return nil, nil
	}

	return InitRankSessionHandler()
}

// --- Environment Variable Configuration ---

// EnvRankEnabled checks if rank metrics are enabled via environment variable.
// Set MOAI_RANK_ENABLED=false to disable rank session submission.
func EnvRankEnabled() bool {
	val := os.Getenv("MOAI_RANK_ENABLED")
	if val == "" {
		return true // Default to enabled
	}
	enabled, err := strconv.ParseBool(val)
	if err != nil {
		return true // Default to enabled on parse error
	}
	return enabled
}

// EnvRankTimeout returns the timeout for rank submission from environment variable.
// Defaults to 10 seconds. Set MOAI_RANK_TIMEOUT to customize.
func EnvRankTimeout() time.Duration {
	val := os.Getenv("MOAI_RANK_TIMEOUT")
	if val == "" {
		return 10 * time.Second
	}
	seconds, err := strconv.Atoi(val)
	if err != nil {
		return 10 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
