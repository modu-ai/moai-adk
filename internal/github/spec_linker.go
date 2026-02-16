package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RegistryFileName is the JSON file that maps GitHub issues to SPEC IDs.
const RegistryFileName = "github-spec-registry.json"

// RegistryVersion is the current schema version of the registry file.
const RegistryVersion = "1.0.0"

// Registry holds the bidirectional mappings between GitHub issues and SPECs.
type Registry struct {
	Version  string        `json:"version"`
	Mappings []SpecMapping `json:"mappings"`
}

// SpecMapping represents a single issue-to-SPEC link.
type SpecMapping struct {
	IssueNumber int       `json:"issue_number"`
	SpecID      string    `json:"spec_id"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}

// SpecLinker manages bidirectional links between GitHub issues and SPEC documents.
type SpecLinker interface {
	// LinkIssueToSpec creates a new mapping between an issue and a SPEC.
	LinkIssueToSpec(issueNum int, specID string) error

	// GetLinkedSpec returns the SPEC ID linked to the given issue number.
	GetLinkedSpec(issueNum int) (string, error)

	// GetLinkedIssue returns the issue number linked to the given SPEC ID.
	GetLinkedIssue(specID string) (int, error)

	// ListMappings returns all current mappings.
	ListMappings() []SpecMapping
}

// fileSpecLinker implements SpecLinker using a JSON file.
type fileSpecLinker struct {
	mu           sync.Mutex
	registryPath string
	registry     *Registry
}

// NewSpecLinker creates a SpecLinker that stores mappings in the given project root.
// Registry file is stored at {projectRoot}/.moai/github-spec-registry.json.
func NewSpecLinker(projectRoot string) (SpecLinker, error) {
	registryPath := filepath.Join(projectRoot, ".moai", RegistryFileName)
	linker := &fileSpecLinker{
		registryPath: registryPath,
	}
	if err := linker.load(); err != nil {
		return nil, fmt.Errorf("new spec linker: %w", err)
	}
	return linker, nil
}

// load reads the registry from disk or creates an empty one.
func (l *fileSpecLinker) load() error {
	data, err := os.ReadFile(l.registryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			l.registry = &Registry{
				Version:  RegistryVersion,
				Mappings: []SpecMapping{},
			}
			return nil
		}
		return fmt.Errorf("load registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		// Backup corrupt file and start fresh.
		_ = os.Rename(l.registryPath, l.registryPath+".corrupt")
		l.registry = &Registry{
			Version:  RegistryVersion,
			Mappings: []SpecMapping{},
		}
		return nil
	}

	if reg.Mappings == nil {
		reg.Mappings = []SpecMapping{}
	}
	l.registry = &reg
	return nil
}

// save writes the registry to disk atomically.
func (l *fileSpecLinker) save() error {
	data, err := json.MarshalIndent(l.registry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(l.registryPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	// Atomic write: temp file + rename.
	tmp, err := os.CreateTemp(dir, ".registry-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	return os.Rename(tmpName, l.registryPath)
}

// LinkIssueToSpec creates a new mapping between an issue and a SPEC.
func (l *fileSpecLinker) LinkIssueToSpec(issueNum int, specID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check for duplicate.
	for _, m := range l.registry.Mappings {
		if m.IssueNumber == issueNum {
			return fmt.Errorf("issue #%d: %w", issueNum, ErrMappingExists)
		}
	}

	l.registry.Mappings = append(l.registry.Mappings, SpecMapping{
		IssueNumber: issueNum,
		SpecID:      specID,
		CreatedAt:   time.Now().UTC(),
		Status:      "active",
	})

	return l.save()
}

// GetLinkedSpec returns the SPEC ID linked to the given issue number.
func (l *fileSpecLinker) GetLinkedSpec(issueNum int) (string, error) {
	for _, m := range l.registry.Mappings {
		if m.IssueNumber == issueNum {
			return m.SpecID, nil
		}
	}
	return "", fmt.Errorf("issue #%d: %w", issueNum, ErrMappingNotFound)
}

// GetLinkedIssue returns the issue number linked to the given SPEC ID.
func (l *fileSpecLinker) GetLinkedIssue(specID string) (int, error) {
	for _, m := range l.registry.Mappings {
		if m.SpecID == specID {
			return m.IssueNumber, nil
		}
	}
	return 0, fmt.Errorf("spec %s: %w", specID, ErrMappingNotFound)
}

// ListMappings returns all current mappings.
func (l *fileSpecLinker) ListMappings() []SpecMapping {
	result := make([]SpecMapping, len(l.registry.Mappings))
	copy(result, l.registry.Mappings)
	return result
}
