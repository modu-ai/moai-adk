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

	// GeneratedManaged indicates a file a MoAI generator (not the template
	// deployer) writes into, where user-owned and MoAI-owned parts can share
	// one file. Ownership is decided per part (FileEntry.Parts), never for
	// the whole file.
	GeneratedManaged Provenance = "generated_managed"
)

// IsValid checks if the Provenance value is one of the defined constants.
func (p Provenance) IsValid() bool {
	switch p {
	case TemplateManaged, UserModified, UserCreated, Deprecated, GeneratedManaged:
		return true
	}
	return false
}

// PartKind names the unit of ownership inside a generated file.
type PartKind string

const (
	// PartWholeFile is the whole file, owned only when MoAI created it.
	PartWholeFile PartKind = "whole-file"
	// PartHookHandler is one hook handler, keyed by event and command.
	PartHookHandler PartKind = "hook-handler"
	// PartJSONKey is one top-level JSON key.
	PartJSONKey PartKind = "json-key"
	// PartTOMLTable is one TOML table MoAI appended as a whole.
	PartTOMLTable PartKind = "toml-table"
	// PartTOMLKey is one TOML assignment MoAI inserted into a table it does
	// not own.
	PartTOMLKey PartKind = "toml-key"
)

// PartOrigin records who put a part in a generated file.
type PartOrigin string

const (
	// OriginCreated means MoAI wrote the part and may remove it.
	OriginCreated PartOrigin = "created"
	// OriginPreexisting means the part was present before any MoAI wiring
	// evidence existed; it is user-owned.
	OriginPreexisting PartOrigin = "preexisting"
	// OriginUnknown means the part was present, earlier wiring evidence
	// existed, and no part record did; ownership cannot be established.
	OriginUnknown PartOrigin = "unknown"
)

// Part is one ownership record inside a generated file. A preexisting or
// unknown part is never promoted to created by a later run.
type Part struct {
	Kind   PartKind   `json:"kind"`
	Key    string     `json:"key"`
	Origin PartOrigin `json:"origin"`
	// Hash is the sha256 of the bytes MoAI wrote for the part; empty for a
	// preexisting or unknown part.
	Hash string `json:"hash,omitempty"`
	// Region is the exact byte region MoAI inserted (TOML parts only),
	// including the separator it inserted with it.
	Region string `json:"region,omitempty"`
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
	// Parts is the per-part ownership of a GeneratedManaged file.
	Parts []Part `json:"parts,omitempty"`
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

// @MX:ANCHOR: [AUTO] Manifest data structure creation - called from 3 or more paths including Load, corrupt recovery, and Track
// @MX:REASON: [AUTO] fan_in=3, responsible for Files map initialization; single exit point preventing nil map panics
// NewManifest creates a new empty Manifest with initialized Files map.
func NewManifest() *Manifest {
	return &Manifest{
		Files: make(map[string]FileEntry),
	}
}
