package worktree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RegistryData maps project names to their worktree entries
// Format: { "project-name": { "SPEC-ID": WorktreeInfo } }
type RegistryData map[string]map[string]*WorktreeInfo

// Registry manages worktree metadata persistence
type Registry struct {
	mu       sync.Mutex
	path     string
	data     RegistryData
	repoName string
}

// NewRegistry creates a registry for the given worktree root and repo name
func NewRegistry(worktreeRoot, repoName string) (*Registry, error) {
	regPath := filepath.Join(worktreeRoot, ".moai-worktree-registry.json")
	r := &Registry{
		path:     regPath,
		data:     make(RegistryData),
		repoName: repoName,
	}
	if err := r.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}
	return r, nil
}

// Register adds or updates a worktree entry
func (r *Registry) Register(info *WorktreeInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data[r.repoName] == nil {
		r.data[r.repoName] = make(map[string]*WorktreeInfo)
	}
	r.data[r.repoName][info.SpecID] = info
	return r.save()
}

// Unregister removes a worktree entry
func (r *Registry) Unregister(specID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data[r.repoName] != nil {
		delete(r.data[r.repoName], specID)
		if len(r.data[r.repoName]) == 0 {
			delete(r.data, r.repoName)
		}
	}
	return r.save()
}

// Get returns worktree info for a given SPEC ID
func (r *Registry) Get(specID string) *WorktreeInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data[r.repoName] != nil {
		return r.data[r.repoName][specID]
	}
	// Search all projects
	for _, entries := range r.data {
		if info, ok := entries[specID]; ok {
			return info
		}
	}
	return nil
}

// List returns all worktrees for the current project
func (r *Registry) List() []*WorktreeInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*WorktreeInfo
	if entries, ok := r.data[r.repoName]; ok {
		for _, info := range entries {
			result = append(result, info)
		}
	}
	return result
}

// ListAll returns all worktrees across all projects
func (r *Registry) ListAll() []*WorktreeInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*WorktreeInfo
	for _, entries := range r.data {
		for _, info := range entries {
			result = append(result, info)
		}
	}
	return result
}

// load reads the registry file from disk
func (r *Registry) load() error {
	content, err := os.ReadFile(r.path)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		return nil
	}
	return json.Unmarshal(content, &r.data)
}

// save writes the registry to disk
func (r *Registry) save() error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, append(content, '\n'), 0644)
}
