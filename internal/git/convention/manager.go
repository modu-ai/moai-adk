package convention

import "fmt"

// defaultAutoDetectSampleSize is the fallback sample size used when an
// AutoDetectOptions carries a non-positive SampleSize.
const defaultAutoDetectSampleSize = 100

// defaultAutoDetectFallback is the built-in convention used when no fallback
// is configured in AutoDetectOptions.
const defaultAutoDetectFallback = "conventional-commits"

// AutoDetectOptions carries the auto-detection knobs forwarded into
// LoadConvention. It mirrors the live levers of models.AutoDetectionConfig
// (enabled / sample_size / confidence_threshold / fallback) without importing
// pkg/models — the caller (internal/cli/hook_pre_push.go) translates the config
// struct into this package-local options struct, keeping the convention engine
// decoupled from the config schema.
type AutoDetectOptions struct {
	Enabled             bool
	SampleSize          int
	ConfidenceThreshold float64
	Fallback            string
}

// Manager coordinates convention loading, detection, and validation.
type Manager struct {
	convention *Convention
	repoPath   string
}

// NewManager creates a Manager for the given repository path.
// @MX:ANCHOR: [AUTO] NewManager is a core factory called from 4+ packages
// @MX:REASON: Multi-package dependency — signature changes have wide blast radius
func NewManager(repoPath string) *Manager {
	return &Manager{repoPath: repoPath}
}

// LoadConvention loads a convention by name (built-in) or, when name is "auto",
// auto-detects it from the repository history honoring the supplied
// AutoDetectOptions (Fix A): the enabled flag gates whether detection runs, the
// sample size is forwarded to Detect (instead of a hardcoded 100), the confidence
// threshold gates acceptance of the detected convention, and the configured
// fallback is used when detection is disabled, errors, or scores below the
// threshold. For a named built-in convention the options are ignored.
func (m *Manager) LoadConvention(name string, opts AutoDetectOptions) error {
	if name == "auto" {
		conv, err := m.resolveAuto(opts)
		if err != nil {
			return err
		}
		m.convention = conv
		return nil
	}

	// Named built-in convention — auto-detection options do not apply.
	conv, err := ParseBuiltin(name)
	if err != nil {
		return fmt.Errorf("load convention %q: %w", name, err)
	}
	m.convention = conv
	return nil
}

// resolveAuto applies the Fix A auto-detection contract and returns the resolved
// Convention. It never mutates the Manager — the caller assigns the result.
func (m *Manager) resolveAuto(opts AutoDetectOptions) (*Convention, error) {
	// When auto-detection is disabled, use the configured fallback directly.
	if !opts.Enabled {
		return m.loadFallback(opts.Fallback)
	}

	sampleSize := opts.SampleSize
	if sampleSize <= 0 {
		sampleSize = defaultAutoDetectSampleSize
	}

	result, err := Detect(m.repoPath, sampleSize)
	if err != nil || result.Confidence < opts.ConfidenceThreshold {
		// Detection failed or scored below the confidence threshold → fallback.
		return m.loadFallback(opts.Fallback)
	}
	return result.Convention, nil
}

// loadFallback parses the configured fallback convention, defaulting to
// conventional-commits when no fallback is specified.
func (m *Manager) loadFallback(fallback string) (*Convention, error) {
	if fallback == "" {
		fallback = defaultAutoDetectFallback
	}
	conv, err := ParseBuiltin(fallback)
	if err != nil {
		return nil, fmt.Errorf("load convention: fallback %q failed: %w", fallback, err)
	}
	return conv, nil
}

// SetMaxLength overrides the maximum commit header length on the loaded
// convention (Fix B). It is applied after LoadConvention to forward the
// configured validation.max_length, since a built-in convention parsed by
// ParseBuiltin carries its own default MaxLength. A non-positive value or a nil
// loaded convention is a no-op (preserving the built-in default).
func (m *Manager) SetMaxLength(maxLength int) {
	if m.convention == nil || maxLength <= 0 {
		return
	}
	m.convention.MaxLength = maxLength
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
