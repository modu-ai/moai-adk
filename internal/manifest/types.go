// Package manifest provides file provenance tracking and change detection
// for the MoAI-ADK template deployment system.
//
// It implements ADR-007 (File Manifest Provenance) by tracking deployed files
// with triple-hash entries (template_hash, deployed_hash, current_hash) and
// four provenance classifications.
package manifest

import "errors"

// Provenance classifies the origin and ownership of a tracked file.
type Provenance string

const (
	// TemplateManaged indicates the file was deployed from a template
	// and has not been modified by the user. Safe to overwrite.
	TemplateManaged Provenance = "template_managed"

	// UserModified indicates the file was deployed from a template
	// but has been edited by the user. Requires 3-way merge.
	UserModified Provenance = "user_modified"

	// UserCreated indicates the file was created by the user
	// and is unrelated to any template. Never modify.
	UserCreated Provenance = "user_created"

	// Deprecated indicates the file has been removed in the new
	// template version. Notify user but preserve the file.
	Deprecated Provenance = "deprecated"
)

// IsValid checks if the Provenance value is one of the defined constants.
func (p Provenance) IsValid() bool {
	switch p {
	case TemplateManaged, UserModified, UserCreated, Deprecated:
		return true
	}
	return false
}

// Manifest represents the file tracking manifest stored at .moai/manifest.json.
type Manifest struct {
	Version    string               `json:"version"`
	DeployedAt string               `json:"deployed_at"`
	Files      map[string]FileEntry `json:"files"`
}

// FileEntry represents a single tracked file in the manifest.
type FileEntry struct {
	Provenance   Provenance `json:"provenance"`
	TemplateHash string     `json:"template_hash"`
	DeployedHash string     `json:"deployed_hash"`
	CurrentHash  string     `json:"current_hash"`
}

// ChangedFile represents a file whose content has changed since last tracking.
type ChangedFile struct {
	Path       string     `json:"path"`
	OldHash    string     `json:"old_hash"`
	NewHash    string     `json:"new_hash"`
	Provenance Provenance `json:"provenance"`
}

// Sentinel errors for the manifest package.
var (
	// ErrManifestNotFound indicates the manifest file does not exist.
	ErrManifestNotFound = errors.New("manifest: file not found")

	// ErrManifestCorrupt indicates the manifest JSON could not be parsed.
	ErrManifestCorrupt = errors.New("manifest: JSON parse error")

	// ErrEntryNotFound indicates the requested file entry does not exist.
	ErrEntryNotFound = errors.New("manifest: entry not found")

	// ErrHashMismatch indicates a hash verification failure.
	ErrHashMismatch = errors.New("manifest: hash verification failed")
)

// NewManifest creates a new empty Manifest with initialized Files map.
func NewManifest() *Manifest {
	return &Manifest{
		Files: make(map[string]FileEntry),
	}
}
