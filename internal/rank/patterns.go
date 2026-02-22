package rank

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"gopkg.in/yaml.v3"
)

// PatternStore manages exclusion and inclusion patterns for metrics sync.
type PatternStore struct {
	mu       sync.RWMutex
	path     string
	patterns *PatternConfig
}

// PatternConfig holds the pattern configuration.
type PatternConfig struct {
	ExcludePatterns []string `yaml:"exclude_patterns"`
	IncludePatterns []string `yaml:"include_patterns"`
}

// NewPatternStore creates a new PatternStore with the given config directory.
func NewPatternStore(configDir string) (*PatternStore, error) {
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		configDir = filepath.Join(homeDir, ".moai", "config")
	}

	path := filepath.Join(configDir, "rank.yaml")

	store := &PatternStore{
		path: path,
		patterns: &PatternConfig{
			ExcludePatterns: []string{},
			IncludePatterns: []string{},
		},
	}

	if err := store.load(); err != nil {
		// If file doesn't exist, that's okay - start with empty patterns
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load patterns: %w", err)
		}
	}

	return store, nil
}

// load reads the pattern configuration from disk.
func (s *PatternStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var config PatternConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("unmarshal patterns: %w", err)
	}

	s.patterns = &config
	return nil
}

// save writes the pattern configuration to disk.
func (s *PatternStore) save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(s.patterns)
	if err != nil {
		return fmt.Errorf("marshal patterns: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write patterns: %w", err)
	}

	return nil
}

// AddExclude adds a pattern to the exclusion list.
func (s *PatternStore) AddExclude(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if slices.Contains(s.patterns.ExcludePatterns, pattern) {
		return fmt.Errorf("pattern already exists in exclude list")
	}

	s.patterns.ExcludePatterns = append(s.patterns.ExcludePatterns, pattern)
	return s.save()
}

// RemoveExclude removes a pattern from the exclusion list.
func (s *PatternStore) RemoveExclude(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.patterns.ExcludePatterns {
		if p == pattern {
			s.patterns.ExcludePatterns = append(
				s.patterns.ExcludePatterns[:i],
				s.patterns.ExcludePatterns[i+1:]...,
			)
			return s.save()
		}
	}

	return fmt.Errorf("pattern not found in exclude list")
}

// AddInclude adds a pattern to the inclusion list.
func (s *PatternStore) AddInclude(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if slices.Contains(s.patterns.IncludePatterns, pattern) {
		return fmt.Errorf("pattern already exists in include list")
	}

	s.patterns.IncludePatterns = append(s.patterns.IncludePatterns, pattern)
	return s.save()
}

// RemoveInclude removes a pattern from the inclusion list.
func (s *PatternStore) RemoveInclude(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.patterns.IncludePatterns {
		if p == pattern {
			s.patterns.IncludePatterns = append(
				s.patterns.IncludePatterns[:i],
				s.patterns.IncludePatterns[i+1:]...,
			)
			return s.save()
		}
	}

	return fmt.Errorf("pattern not found in include list")
}

// ListPatterns returns all patterns.
func (s *PatternStore) ListPatterns() (exclude, include []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exclude = make([]string, len(s.patterns.ExcludePatterns))
	copy(exclude, s.patterns.ExcludePatterns)

	include = make([]string, len(s.patterns.IncludePatterns))
	copy(include, s.patterns.IncludePatterns)

	return exclude, include
}

// ShouldExclude checks if a given path matches any exclusion pattern.
func (s *PatternStore) ShouldExclude(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Simple prefix matching for now
	// Can be extended to support glob patterns
	return slices.Contains(s.patterns.ExcludePatterns, path)
}

// ShouldInclude checks if a given path matches any inclusion pattern.
func (s *PatternStore) ShouldInclude(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return slices.Contains(s.patterns.IncludePatterns, path)
}

// GetConfig returns a copy of the current pattern configuration.
func (s *PatternStore) GetConfig() PatternConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config := PatternConfig{
		ExcludePatterns: make([]string, len(s.patterns.ExcludePatterns)),
		IncludePatterns: make([]string, len(s.patterns.IncludePatterns)),
	}
	copy(config.ExcludePatterns, s.patterns.ExcludePatterns)
	copy(config.IncludePatterns, s.patterns.IncludePatterns)

	return config
}
