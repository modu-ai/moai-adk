package convention

import "fmt"

// Manager coordinates convention loading, detection, and validation.
type Manager struct {
	convention *Convention
	repoPath   string
}

// NewManager creates a Manager for the given repository path.
func NewManager(repoPath string) *Manager {
	return &Manager{repoPath: repoPath}
}

// LoadConvention loads a convention by name (built-in) or from config.
// If name is "auto", it auto-detects from the repository history and
// falls back to conventional-commits on failure.
func (m *Manager) LoadConvention(name string) error {
	if name == "auto" {
		result, err := Detect(m.repoPath, 100)
		if err != nil {
			// Fallback to conventional-commits.
			conv, parseErr := ParseBuiltin("conventional-commits")
			if parseErr != nil {
				return fmt.Errorf("load convention: auto-detect failed and fallback failed: %w", parseErr)
			}
			m.convention = conv
			return nil
		}
		m.convention = result.Convention
		return nil
	}

	// Try built-in first.
	conv, err := ParseBuiltin(name)
	if err == nil {
		m.convention = conv
		return nil
	}

	return fmt.Errorf("load convention %q: %w", name, err)
}

// LoadFromConfig loads a convention from a ConventionConfig struct.
func (m *Manager) LoadFromConfig(cfg ConventionConfig) error {
	conv, err := Parse(cfg)
	if err != nil {
		return fmt.Errorf("load from config: %w", err)
	}
	m.convention = conv
	return nil
}

// ValidateMessage validates a single commit message against the loaded convention.
func (m *Manager) ValidateMessage(message string) ValidationResult {
	return Validate(message, m.convention)
}

// ValidateMessages validates multiple commit messages.
func (m *Manager) ValidateMessages(messages []string) []ValidationResult {
	results := make([]ValidationResult, len(messages))
	for i, msg := range messages {
		results[i] = Validate(msg, m.convention)
	}
	return results
}

// Convention returns the currently loaded convention, or nil if none loaded.
func (m *Manager) Convention() *Convention {
	return m.convention
}
