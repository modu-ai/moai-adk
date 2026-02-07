// Package lifecycle provides session lifecycle utilities for MoAI-ADK.
// It handles auto-cleanup on SessionEnd and work state persistence between sessions.
package lifecycle

import (
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// SessionCleanup handles SessionEnd cleanup.
type SessionCleanup interface {
	// CleanTempFiles removes temporary files from .moai/temp/.
	CleanTempFiles() (*CleanupResult, error)

	// ClearCaches clears session-specific caches.
	ClearCaches() error

	// GenerateCleanupReport generates a human-readable cleanup report.
	GenerateCleanupReport() string
}

// CleanupResult represents the result of cleanup operations.
type CleanupResult struct {
	FilesDeleted int           `json:"filesDeleted"`
	DirsDeleted  int           `json:"dirsDeleted"`
	BytesFreed   int64         `json:"bytesFreed"`
	Errors       []string      `json:"errors"`
	Duration     time.Duration `json:"duration"`
}

// CleanupConfig configures cleanup behavior.
type CleanupConfig struct {
	// TempDir is the directory for temporary files.
	TempDir string

	// CacheDir is the directory for cache files.
	CacheDir string

	// LogDir is the directory for log files.
	LogDir string

	// SessionLogPattern is the pattern for session log files to clean.
	SessionLogPattern string

	// PreserveMemory indicates whether to preserve .moai/memory/*.json files.
	PreserveMemory bool
}

// DefaultCleanupConfig returns the default cleanup configuration.
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		TempDir:           defs.MoAIDir + "/temp",
		CacheDir:          defs.MoAIDir + "/cache/temp",
		LogDir:            defs.MoAIDir + "/" + defs.LogsSubdir,
		SessionLogPattern: "session-*.log",
		PreserveMemory:    true,
	}
}

// WorkState persists work state between sessions.
type WorkState interface {
	// Save persists the work state to storage.
	Save(state *WorkStateData) error

	// Load retrieves the work state from storage.
	Load() (*WorkStateData, error)
}

// WorkStateData represents work session state.
// Simplified to focus on crash recovery rather than file position tracking.
type WorkStateData struct {
	ActiveFiles    []string  `json:"activeFiles,omitempty"`
	ContextSummary string    `json:"contextSummary,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// WorkStateConfig configures work state persistence.
type WorkStateConfig struct {
	// StoragePath is the path for work state storage.
	StoragePath string
}

// DefaultWorkStateConfig returns the default work state configuration.
func DefaultWorkStateConfig() WorkStateConfig {
	return WorkStateConfig{
		StoragePath: defs.MoAIDir + "/" + defs.MemorySubdir + "/last-session-state.json",
	}
}
