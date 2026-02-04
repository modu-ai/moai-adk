// Package ops provides optimized Git operations with parallel execution,
// result caching, and connection pooling.
package ops

import (
	"time"
)

// GitOperationType represents the type of Git operation.
type GitOperationType string

const (
	OpBranch GitOperationType = "branch"
	OpCommit GitOperationType = "commit"
	OpStatus GitOperationType = "status"
	OpLog    GitOperationType = "log"
	OpDiff   GitOperationType = "diff"
	OpRemote GitOperationType = "remote"
	OpConfig GitOperationType = "config"
)

// GitCommand represents a Git command specification.
type GitCommand struct {
	OperationType   GitOperationType `json:"operationType"`
	Args            []string         `json:"args"`
	CacheTTLSeconds int              `json:"cacheTTLSeconds"`
	RetryCount      int              `json:"retryCount"`
	TimeoutSeconds  int              `json:"timeoutSeconds"`
	WorkDir         string           `json:"workDir,omitempty"`
}

// GitResult represents the result of a Git operation.
type GitResult struct {
	Success       bool             `json:"success"`
	Stdout        string           `json:"stdout"`
	Stderr        string           `json:"stderr"`
	ReturnCode    int              `json:"returnCode"`
	ExecutionTime time.Duration    `json:"executionTime"`
	Cached        bool             `json:"cached"`
	CacheHit      bool             `json:"cacheHit"`
	OperationType GitOperationType `json:"operationType"`
	Command       []string         `json:"command"`
	Error         error            `json:"-"`
}

// ProjectInfo represents comprehensive Git project information.
type ProjectInfo struct {
	Branch     string    `json:"branch"`
	LastCommit string    `json:"lastCommit"`
	CommitTime string    `json:"commitTime"`
	Changes    int       `json:"changes"`
	FetchTime  time.Time `json:"fetchTime"`
}

// Statistics represents performance and cache statistics.
type Statistics struct {
	Operations OperationStats `json:"operations"`
	Cache      CacheStats     `json:"cache"`
	Queue      QueueStats     `json:"queue"`
}

// OperationStats tracks operation-level metrics.
type OperationStats struct {
	Total            int     `json:"total"`
	CacheHits        int     `json:"cacheHits"`
	CacheMisses      int     `json:"cacheMisses"`
	CacheHitRate     float64 `json:"cacheHitRate"`
	Errors           int     `json:"errors"`
	AvgExecutionTime float64 `json:"avgExecutionTime"`
	TotalTime        int64   `json:"totalTime"`
}

// CacheStats tracks cache-level metrics.
type CacheStats struct {
	Size        int     `json:"size"`
	SizeLimit   int     `json:"sizeLimit"`
	Utilization float64 `json:"utilization"`
}

// QueueStats tracks queue-level metrics.
type QueueStats struct {
	Pending int `json:"pending"`
}

// GitOperationsManager manages optimized Git operations.
type GitOperationsManager interface {
	// ExecuteCommand executes a single Git command.
	ExecuteCommand(cmd GitCommand) GitResult

	// ExecuteParallel executes multiple Git commands in parallel.
	ExecuteParallel(cmds []GitCommand) []GitResult

	// GetProjectInfo returns comprehensive project information.
	GetProjectInfo() ProjectInfo

	// GetStatistics returns performance and cache statistics.
	GetStatistics() Statistics

	// ClearCache clears cache entries for a specific operation type.
	// Returns the number of entries cleared.
	ClearCache(opType GitOperationType) int

	// Shutdown gracefully shuts down the manager.
	Shutdown()
}

// ManagerConfig holds configuration for GitOperationsManager.
type ManagerConfig struct {
	// MaxWorkers is the maximum number of concurrent workers (default: 4).
	MaxWorkers int `json:"maxWorkers"`

	// CacheSizeLimit is the maximum number of cache entries (default: 100).
	CacheSizeLimit int `json:"cacheSizeLimit"`

	// DefaultTTLSeconds is the default cache TTL in seconds (default: 60).
	DefaultTTLSeconds int `json:"defaultTTLSeconds"`

	// DefaultTimeoutSeconds is the default command timeout (default: 10).
	DefaultTimeoutSeconds int `json:"defaultTimeoutSeconds"`

	// DefaultRetryCount is the default retry count (default: 2).
	DefaultRetryCount int `json:"defaultRetryCount"`

	// WorkDir is the working directory for Git commands.
	WorkDir string `json:"workDir"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() ManagerConfig {
	return ManagerConfig{
		MaxWorkers:            4,
		CacheSizeLimit:        100,
		DefaultTTLSeconds:     60,
		DefaultTimeoutSeconds: 10,
		DefaultRetryCount:     2,
		WorkDir:               "",
	}
}
