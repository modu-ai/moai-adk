package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// manifestFileName is an alias for defs.ManifestJSON kept for test compatibility.
const manifestFileName = defs.ManifestJSON

// Manager defines the interface for manifest CRUD operations and change detection.
type Manager interface {
	// Load reads the manifest from {projectRoot}/.moai/manifest.json.
	// If the file does not exist, returns an empty manifest with nil error.
	// If the file is corrupt, backs up to .corrupt and returns a fresh manifest
	// along with an ErrManifestCorrupt error.
	Load(projectRoot string) (*Manifest, error)

	// Manifest returns the currently loaded in-memory manifest without
	// reading from disk. Returns nil if Load has not been called yet.
	Manifest() *Manifest

	// Save persists the in-memory manifest to disk atomically.
	Save() error

	// Track registers or updates a file entry in the manifest.
	// It computes the current file hash and creates/updates the FileEntry.
	Track(path string, provenance Provenance, templateHash string) error

	// GetEntry retrieves a file entry by its relative path.
	GetEntry(path string) (*FileEntry, bool)

	// DetectChanges computes current hashes of all tracked files and
	// returns a list of files whose hashes differ from the recorded value.
	DetectChanges() ([]ChangedFile, error)

	// Remove deletes a file entry from the manifest.
	Remove(path string) error
}

// manifestManager is the concrete implementation of Manager.
type manifestManager struct {
	projectRoot  string
	manifest     *Manifest
	manifestPath string
}

// NewManager creates a new Manager instance.
// Load must be called before using other methods.
func NewManager() Manager {
	return &manifestManager{}
}

// Manifest returns the currently loaded in-memory manifest without disk I/O.
func (m *manifestManager) Manifest() *Manifest {
	return m.manifest
}

// Load reads and parses the manifest from disk.
func (m *manifestManager) Load(projectRoot string) (*Manifest, error) {
	m.projectRoot = filepath.Clean(projectRoot)
	m.manifestPath = filepath.Join(m.projectRoot, defs.MoAIDir, defs.ManifestJSON)

	data, err := os.ReadFile(m.manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.manifest = NewManifest()
			return m.manifest, nil
		}
		return nil, fmt.Errorf("manifest load: %w", err)
	}

	var mf Manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		// Backup corrupt file
		corruptPath := m.manifestPath + ".corrupt"
		_ = os.Rename(m.manifestPath, corruptPath)

		m.manifest = NewManifest()
		return m.manifest, fmt.Errorf("%w: %v", ErrManifestCorrupt, err)
	}

	// Ensure Files map is initialized
	if mf.Files == nil {
		mf.Files = make(map[string]FileEntry)
	}

	m.manifest = &mf
	return m.manifest, nil
}

// Save persists the manifest to disk using atomic write (temp file + rename).
func (m *manifestManager) Save() error {
	if m.manifest == nil {
		return fmt.Errorf("manifest save: %w", ErrManifestNotFound)
	}

	data, err := json.MarshalIndent(m.manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("manifest save marshal: %w", err)
	}

	// Append trailing newline for POSIX compliance
	data = append(data, '\n')

	// Ensure parent directory exists
	dir := filepath.Dir(m.manifestPath)
	if err := os.MkdirAll(dir, defs.DirPerm); err != nil {
		return fmt.Errorf("manifest save mkdir: %w", err)
	}

	// Atomic write: temp file + rename
	if err := atomicWriteFile(m.manifestPath, data); err != nil {
		return fmt.Errorf("manifest save: %w", err)
	}

	return nil
}

// Track registers or updates a file entry with the current file hash.
func (m *manifestManager) Track(path string, provenance Provenance, templateHash string) error {
	if m.manifest == nil {
		return fmt.Errorf("manifest track: %w", ErrManifestNotFound)
	}

	absPath := filepath.Join(m.projectRoot, filepath.Clean(path))
	currentHash, err := HashFile(absPath)
	if err != nil {
		return fmt.Errorf("manifest track hash: %w", err)
	}

	m.manifest.Files[path] = FileEntry{
		Provenance:   provenance,
		TemplateHash: templateHash,
		DeployedHash: currentHash,
		CurrentHash:  currentHash,
	}

	return nil
}

// GetEntry retrieves a file entry by relative path.
func (m *manifestManager) GetEntry(path string) (*FileEntry, bool) {
	if m.manifest == nil {
		return nil, false
	}

	entry, ok := m.manifest.Files[path]
	if !ok {
		return nil, false
	}

	return &entry, true
}

// DetectChanges compares recorded hashes with current file hashes.
func (m *manifestManager) DetectChanges() ([]ChangedFile, error) {
	if m.manifest == nil {
		return nil, fmt.Errorf("manifest detect changes: %w", ErrManifestNotFound)
	}

	var changes []ChangedFile

	for path, entry := range m.manifest.Files {
		absPath := filepath.Join(m.projectRoot, filepath.Clean(path))
		currentHash, err := HashFile(absPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// File was deleted
				changes = append(changes, ChangedFile{
					Path:       path,
					OldHash:    entry.CurrentHash,
					NewHash:    "",
					Provenance: entry.Provenance,
				})
				continue
			}
			return nil, fmt.Errorf("manifest detect changes: %w", err)
		}

		if currentHash != entry.CurrentHash {
			changes = append(changes, ChangedFile{
				Path:       path,
				OldHash:    entry.CurrentHash,
				NewHash:    currentHash,
				Provenance: entry.Provenance,
			})
		}
	}

	return changes, nil
}

// Remove deletes a file entry from the manifest.
func (m *manifestManager) Remove(path string) error {
	if m.manifest == nil {
		return fmt.Errorf("manifest remove: %w", ErrManifestNotFound)
	}

	delete(m.manifest.Files, path)
	return nil
}

// atomicWriteFile writes data to path atomically using temp file + rename.
func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".manifest-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // clean up on error path

	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck // best-effort close on error path
		return fmt.Errorf("write temp: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	return os.Rename(tmpName, path)
}
