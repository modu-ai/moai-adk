package mx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// SchemaVersion is the current sidecar schema version.
	SchemaVersion = 2

	// SidecarFileName is the name of the sidecar index file.
	SidecarFileName = "mx-index.json"

	// ArchiveFileName is the name of the archive file for stale tags.
	ArchiveFileName = "mx-archive.json"
)

// Sidecar represents the machine-readable JSON sidecar index.
type Sidecar struct {
	// SchemaVersion is the version of the sidecar schema.
	SchemaVersion int `json:"schema_version"`

	// Tags is the list of all @MX tags in the project.
	Tags []Tag `json:"tags"`

	// ScannedAt is the timestamp when this index was last updated.
	ScannedAt time.Time `json:"scanned_at"`
}

// Archive contains tags that have been stale for more than 7 days.
type Archive struct {
	// SchemaVersion matches the sidecar schema version.
	SchemaVersion int `json:"schema_version"`

	// ArchivedTags is the list of removed tags.
	ArchivedTags []Tag `json:"archived_tags"`

	// ArchivedAt is when the archive was last updated.
	ArchivedAt time.Time `json:"archived_at"`
}

// Manager manages the sidecar index file.
type Manager struct {
	mu            sync.RWMutex
	stateDir      string
	sidecarPath   string
	archivePath   string
	currentTags   map[string]Tag // Key() -> Tag
}

// NewManager creates a new sidecar manager.
func NewManager(stateDir string) *Manager {
	return &Manager{
		stateDir:    stateDir,
		sidecarPath: filepath.Join(stateDir, SidecarFileName),
		archivePath: filepath.Join(stateDir, ArchiveFileName),
		currentTags: make(map[string]Tag),
	}
}

// loadWithoutLock reads the existing sidecar index from disk without acquiring the lock.
// Caller must hold m.mu.
func (m *Manager) loadWithoutLock() (*Sidecar, error) {
	data, err := os.ReadFile(m.sidecarPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Sidecar{SchemaVersion: SchemaVersion}, nil
		}
		return nil, fmt.Errorf("read sidecar: %w", err)
	}

	var sidecar Sidecar
	if err := json.Unmarshal(data, &sidecar); err != nil {
		// Corrupt sidecar - treat as empty and suggest repair
		return &Sidecar{SchemaVersion: SchemaVersion}, nil
	}

	// Rebuild in-memory index
	m.currentTags = make(map[string]Tag)
	for _, tag := range sidecar.Tags {
		m.currentTags[tag.Key()] = tag
	}

	return &sidecar, nil
}

// writeWithoutLock atomically writes the sidecar index to disk without acquiring the lock.
// Caller must hold m.mu.
func (m *Manager) writeWithoutLock(sidecar *Sidecar) error {
	// Ensure state directory exists
	if err := os.MkdirAll(m.stateDir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(sidecar, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sidecar: %w", err)
	}

	// Write to temp file in same directory
	tempPath := m.sidecarPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("write temp sidecar: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, m.sidecarPath); err != nil {
		_ = os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("rename sidecar: %w", err)
	}

	// Update in-memory index
	m.currentTags = make(map[string]Tag)
	for _, tag := range sidecar.Tags {
		m.currentTags[tag.Key()] = tag
	}

	return nil
}

// Load reads the existing sidecar index from disk.
// Returns empty sidecar if file doesn't exist or is corrupt.
func (m *Manager) Load() (*Sidecar, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.loadWithoutLock()
}

// Write atomically writes the sidecar index to disk.
// Uses temp-file + rename pattern to prevent partial reads.
// Uses temp-file + rename pattern to prevent partial reads.
func (m *Manager) Write(sidecar *Sidecar) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.writeWithoutLock(sidecar)
}

// UpdateFile updates the sidecar with new/changed/removed tags for a single file.
// This is used by PostToolUse hook for incremental updates.
func (m *Manager) UpdateFile(filePath string, newTags []Tag) (*Sidecar, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load current sidecar (without lock since we already hold it)
	sidecar, err := m.loadWithoutLock()
	if err != nil {
		return nil, err
	}

	// Remove old tags for this file
	var keptTags []Tag
	for _, tag := range sidecar.Tags {
		if tag.File != filePath {
			keptTags = append(keptTags, tag)
		}
	}

	// Add new tags
	now := time.Now()
	for i := range newTags {
		newTags[i].LastSeenAt = now
		keptTags = append(keptTags, newTags[i])
	}

	sidecar.Tags = keptTags
	sidecar.ScannedAt = now

	return sidecar, m.writeWithoutLock(sidecar)
}

// ArchiveStale removes tags older than 7 days and moves them to archive.
func (m *Manager) ArchiveStale() (*Archive, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sidecar, err := m.loadWithoutLock()
	if err != nil {
		return nil, err
	}

	var activeTags, staleTags []Tag
	now := time.Now()

	for _, tag := range sidecar.Tags {
		if tag.IsStale() {
			staleTags = append(staleTags, tag)
		} else {
			activeTags = append(activeTags, tag)
		}
	}

	if len(staleTags) == 0 {
		return nil, nil // Nothing to archive
	}

	// Update sidecar with only active tags
	sidecar.Tags = activeTags
	sidecar.ScannedAt = now
	if err := m.writeWithoutLock(sidecar); err != nil {
		return nil, err
	}

	// Load existing archive
	var archive Archive
	archiveData, err := os.ReadFile(m.archivePath)
	if err == nil {
		_ = json.Unmarshal(archiveData, &archive)
	}

	// Append stale tags to archive
	archive.ArchivedTags = append(archive.ArchivedTags, staleTags...)
	archive.SchemaVersion = SchemaVersion
	archive.ArchivedAt = now

	// Write archive atomically
	data, err := json.MarshalIndent(&archive, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal archive: %w", err)
	}

	tempPath := m.archivePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write temp archive: %w", err)
	}

	if err := os.Rename(tempPath, m.archivePath); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("rename archive: %w", err)
	}

	return &archive, nil
}

// GetTag returns a tag by its key, if present in the current index.
func (m *Manager) GetTag(key string) (Tag, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tag, ok := m.currentTags[key]
	return tag, ok
}

// GetAllTags returns all tags in the current index.
func (m *Manager) GetAllTags() []Tag {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tags := make([]Tag, 0, len(m.currentTags))
	for _, tag := range m.currentTags {
		tags = append(tags, tag)
	}
	return tags
}
