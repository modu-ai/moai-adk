package lifecycle

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// sessionCleanupImpl implements SessionCleanup.
type sessionCleanupImpl struct {
	config CleanupConfig
	mu     sync.Mutex
	result *CleanupResult
}

// Compile-time interface compliance check.
var _ SessionCleanup = (*sessionCleanupImpl)(nil)

// NewSessionCleanup creates a new SessionCleanup instance.
func NewSessionCleanup(config CleanupConfig) *sessionCleanupImpl {
	return &sessionCleanupImpl{
		config: config,
		result: &CleanupResult{
			Errors: make([]string, 0),
		},
	}
}

// CleanTempFiles removes temporary files from .moai/temp/.
// REQ-HOOK-360: Clean .moai/temp/ directory.
// REQ-HOOK-362: Continue cleanup even if errors occur, log errors.
func (c *sessionCleanupImpl) CleanTempFiles() (*CleanupResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	start := time.Now()
	c.result = &CleanupResult{
		Errors: make([]string, 0),
	}

	// Clean temp directory
	c.cleanDirectory(c.config.TempDir)

	// Clean session log files if pattern is specified
	if c.config.SessionLogPattern != "" && c.config.LogDir != "" {
		c.cleanSessionLogs()
	}

	c.result.Duration = time.Since(start)
	return c.result, nil
}

// cleanDirectory recursively removes all files and subdirectories.
func (c *sessionCleanupImpl) cleanDirectory(dirPath string) {
	if dirPath == "" {
		return
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return // Directory doesn't exist, nothing to clean
		}
		c.addError(fmt.Sprintf("failed to stat %s: %v", dirPath, err))
		return
	}

	if !info.IsDir() {
		return
	}

	// Collect all files and directories
	var files []string
	var dirs []string

	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			c.addError(fmt.Sprintf("failed to walk %s: %v", path, err))
			return nil // Continue walking despite errors
		}

		// Skip the root directory itself
		if path == dirPath {
			return nil
		}

		if d.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
			// Get file size
			if info, err := d.Info(); err == nil {
				c.result.BytesFreed += info.Size()
			}
		}
		return nil
	})

	if err != nil {
		c.addError(fmt.Sprintf("failed to walk directory %s: %v", dirPath, err))
	}

	// Delete files first
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			c.addError(fmt.Sprintf("failed to remove file %s: %v", file, err))
		} else {
			c.result.FilesDeleted++
		}
	}

	// Delete directories in reverse order (deepest first)
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		if err := os.Remove(dir); err != nil {
			// Directory might not be empty due to earlier errors
			c.addError(fmt.Sprintf("failed to remove dir %s: %v", dir, err))
		} else {
			c.result.DirsDeleted++
		}
	}
}

// cleanSessionLogs removes session-specific log files.
func (c *sessionCleanupImpl) cleanSessionLogs() {
	if c.config.LogDir == "" {
		return
	}

	// Check if log directory exists
	if _, err := os.Stat(c.config.LogDir); os.IsNotExist(err) {
		return
	}

	pattern := filepath.Join(c.config.LogDir, c.config.SessionLogPattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		c.addError(fmt.Sprintf("failed to glob session logs: %v", err))
		return
	}

	for _, file := range matches {
		info, err := os.Stat(file)
		if err != nil {
			c.addError(fmt.Sprintf("failed to stat %s: %v", file, err))
			continue
		}

		if err := os.Remove(file); err != nil {
			c.addError(fmt.Sprintf("failed to remove log %s: %v", file, err))
		} else {
			c.result.FilesDeleted++
			c.result.BytesFreed += info.Size()
		}
	}
}

// ClearCaches clears session-specific caches.
// REQ-HOOK-360: Clear .moai/cache/temp/ directory.
func (c *sessionCleanupImpl) ClearCaches() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.config.CacheDir == "" {
		return nil
	}

	// Check if cache directory exists
	if _, err := os.Stat(c.config.CacheDir); os.IsNotExist(err) {
		return nil
	}

	// Remove all files in cache directory
	entries, err := os.ReadDir(c.config.CacheDir)
	if err != nil {
		slog.Warn("failed to read cache directory",
			"dir", c.config.CacheDir,
			"error", err.Error(),
		)
		return nil // Don't fail on cache clearing errors
	}

	for _, entry := range entries {
		path := filepath.Join(c.config.CacheDir, entry.Name())
		if entry.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				slog.Warn("failed to remove cache subdirectory",
					"path", path,
					"error", err.Error(),
				)
			}
		} else {
			if err := os.Remove(path); err != nil {
				slog.Warn("failed to remove cache file",
					"path", path,
					"error", err.Error(),
				)
			}
		}
	}

	return nil
}

// GenerateCleanupReport generates a human-readable cleanup report.
// REQ-HOOK-361: Report cleanup results.
func (c *sessionCleanupImpl) GenerateCleanupReport() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.result == nil {
		return "Cleanup: No cleanup performed"
	}

	var sb strings.Builder

	sb.WriteString("Cleanup Summary\n")
	sb.WriteString(strings.Repeat("-", 30) + "\n")

	fmt.Fprintf(&sb, "Files deleted: %d\n", c.result.FilesDeleted)
	fmt.Fprintf(&sb, "Directories deleted: %d\n", c.result.DirsDeleted)
	fmt.Fprintf(&sb, "Space freed: %s\n", formatBytes(c.result.BytesFreed))
	fmt.Fprintf(&sb, "Duration: %v\n", c.result.Duration.Round(time.Millisecond))

	if len(c.result.Errors) > 0 {
		fmt.Fprintf(&sb, "Errors: %d\n", len(c.result.Errors))
		for i, errMsg := range c.result.Errors {
			if i >= 5 { // Limit error display
				fmt.Fprintf(&sb, "  ... and %d more errors\n", len(c.result.Errors)-5)
				break
			}
			fmt.Fprintf(&sb, "  - %s\n", errMsg)
		}
	}

	return sb.String()
}

// addError adds an error message to the result and logs it.
func (c *sessionCleanupImpl) addError(msg string) {
	c.result.Errors = append(c.result.Errors, msg)
	slog.Warn("cleanup error", "message", msg)
}

// formatBytes formats a byte count as a human-readable string.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
