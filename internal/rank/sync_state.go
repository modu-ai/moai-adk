// Package rank provides sync state tracking for MoAI Rank session submission.
package rank

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// SyncedFile records metadata about a synced transcript file.
type SyncedFile struct {
	FileModTime time.Time `json:"fileModTime"`
	FileSize    int64     `json:"fileSize"`
	SyncedAt    time.Time `json:"syncedAt"`
}

// SyncStateData represents the persistent sync state structure.
type SyncStateData struct {
	Version      int                    `json:"version"`
	LastSyncTime time.Time              `json:"lastSyncTime"`
	SyncedFiles  map[string]*SyncedFile `json:"syncedFiles"`
}

// SyncState manages the sync state for transcript submissions.
type SyncState struct {
	path string
	data *SyncStateData
}

// NewSyncState creates a new SyncState with the default storage path.
// If basePath is empty, uses ~/.moai/rank/sync-state.json.
func NewSyncState(basePath string) (*SyncState, error) {
	if basePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		basePath = filepath.Join(homeDir, defs.MoAIDir, defs.RankSubdir, "sync-state.json")
	}

	s := &SyncState{
		path: basePath,
		data: &SyncStateData{
			Version:     1,
			SyncedFiles: make(map[string]*SyncedFile),
		},
	}

	// Try to load existing state
	if err := s.Load(); err != nil && !os.IsNotExist(err) {
		// If file exists but is corrupted, start fresh
		s.data = &SyncStateData{
			Version:     1,
			SyncedFiles: make(map[string]*SyncedFile),
		}
	}

	return s, nil
}

// Load reads the sync state from disk.
func (s *SyncState) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var state SyncStateData
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("parse sync state: %w", err)
	}

	if state.SyncedFiles == nil {
		state.SyncedFiles = make(map[string]*SyncedFile)
	}

	s.data = &state
	return nil
}

// Save writes the sync state to disk atomically.
func (s *SyncState) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create sync state directory: %w", err)
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sync state: %w", err)
	}

	// Atomic write via temp file
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write sync state: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		// Fallback: direct write if rename fails (cross-device)
		_ = os.Remove(tmpPath)
		if err := os.WriteFile(s.path, data, 0o600); err != nil {
			return fmt.Errorf("write sync state: %w", err)
		}
	}

	return nil
}

// IsSynced checks if a transcript file has been synced and hasn't changed since.
func (s *SyncState) IsSynced(transcriptPath string) bool {
	entry, ok := s.data.SyncedFiles[transcriptPath]
	if !ok {
		return false
	}

	// Check if file has been modified since last sync
	info, err := os.Stat(transcriptPath)
	if err != nil {
		return false
	}

	return info.ModTime().Equal(entry.FileModTime) && info.Size() == entry.FileSize
}

// MarkSynced records a transcript file as successfully synced.
func (s *SyncState) MarkSynced(transcriptPath string) error {
	info, err := os.Stat(transcriptPath)
	if err != nil {
		return fmt.Errorf("stat transcript: %w", err)
	}

	s.data.SyncedFiles[transcriptPath] = &SyncedFile{
		FileModTime: info.ModTime(),
		FileSize:    info.Size(),
		SyncedAt:    time.Now(),
	}

	s.data.LastSyncTime = time.Now()
	return nil
}

// Reset clears all sync state, forcing a full resync.
func (s *SyncState) Reset() {
	s.data = &SyncStateData{
		Version:     1,
		SyncedFiles: make(map[string]*SyncedFile),
	}
}

// SyncedCount returns the number of synced transcripts.
func (s *SyncState) SyncedCount() int {
	return len(s.data.SyncedFiles)
}

// CleanStale removes entries for transcript files that no longer exist.
func (s *SyncState) CleanStale() int {
	removed := 0
	for path := range s.data.SyncedFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			delete(s.data.SyncedFiles, path)
			removed++
		}
	}
	return removed
}
