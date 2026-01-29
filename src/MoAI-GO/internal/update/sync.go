package update

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/anthropics/moai-adk-go/internal/output"
	templatepkg "github.com/anthropics/moai-adk-go/pkg/templates"
)

// Protected directories that should never be modified during update
var protectedDirs = map[string]bool{
	".moai/project/": true,
	".moai/specs/":   true,
}

// SyncManager handles template synchronization
type SyncManager struct {
	projectDir        string
	templateFS        fs.FS
	dryRun            bool
	preserveProtected bool
}

// NewSyncManager creates a new sync manager
func NewSyncManager(projectDir string) *SyncManager {
	return &SyncManager{
		projectDir:        projectDir,
		templateFS:        templatepkg.GetTemplatesFS(),
		dryRun:            false,
		preserveProtected: true,
	}
}

// SetDryRun sets whether to perform a dry run
func (sm *SyncManager) SetDryRun(dryRun bool) {
	sm.dryRun = dryRun
}

// Sync performs the template synchronization
func (sm *SyncManager) Sync() (*SyncResult, error) {
	result := &SyncResult{
		FilesUpdated:  make([]string, 0),
		FilesAdded:    make([]string, 0),
		FilesRemoved:  make([]string, 0),
		ProtectedDirs: make([]string, 0),
	}

	// Walk through embedded templates
	err := fs.WalkDir(sm.templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories in templates (they'll be created as needed)
		if d.IsDir() {
			return nil
		}

		// Check if this is a protected path
		if sm.isProtected(path) {
			result.ProtectedDirs = append(result.ProtectedDirs, path)
			return nil
		}

		// Construct target path
		targetPath := filepath.Join(sm.projectDir, path)

		// Check if file exists
		_, statErr := os.Stat(targetPath)
		exists := statErr == nil

		// Read template content
		content, err := fs.ReadFile(sm.templateFS, path)
		if err != nil {
			return fmt.Errorf("error reading template %s: %w", path, err)
		}

		// Create directory structure
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("error creating directory for %s: %w", path, err)
		}

		// Write file
		if !sm.dryRun {
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("error writing file %s: %w", path, err)
			}
		}

		// Track change
		if exists {
			result.FilesUpdated = append(result.FilesUpdated, path)
		} else {
			result.FilesAdded = append(result.FilesAdded, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// isProtected checks if a path is in a protected directory
func (sm *SyncManager) isProtected(path string) bool {
	if !sm.preserveProtected {
		return false
	}

	for protectedDir := range protectedDirs {
		if path == protectedDir || (len(path) >= len(protectedDir) && path[:len(protectedDir)] == protectedDir) {
			return true
		}
	}

	return false
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	FilesUpdated  []string
	FilesAdded    []string
	FilesRemoved  []string
	ProtectedDirs []string
}

// GetSummary returns a summary of the sync result
func (sr *SyncResult) GetSummary() string {
	summary := fmt.Sprintf("Added %d files, Updated %d files", len(sr.FilesAdded), len(sr.FilesUpdated))
	if len(sr.ProtectedDirs) > 0 {
		summary += fmt.Sprintf(", Preserved %d protected directories", len(sr.ProtectedDirs))
	}
	return summary
}

// PrintSummary prints the sync result summary
func (sr *SyncResult) PrintSummary() {
	fmt.Println(output.InfoStyle.Render("Update Summary:"))
	fmt.Printf("  • Added: %d files\n", len(sr.FilesAdded))
	fmt.Printf("  • Updated: %d files\n", len(sr.FilesUpdated))
	if len(sr.ProtectedDirs) > 0 {
		fmt.Printf("  • Preserved: %d protected directories\n", len(sr.ProtectedDirs))
	}
}
