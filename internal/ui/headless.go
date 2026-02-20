package ui

import (
	"maps"
	"os"

	"github.com/mattn/go-isatty"
)

// HeadlessManager manages headless (non-interactive) mode detection
// and default values for UI components running without a TTY.
type HeadlessManager struct {
	forced   *bool
	defaults map[string]string
}

// NewHeadlessManager creates a HeadlessManager that detects
// headless mode from the TTY state of os.Stdin.
func NewHeadlessManager() *HeadlessManager {
	return &HeadlessManager{}
}

// IsHeadless returns true when the UI should operate in headless mode.
// ForceHeadless overrides TTY detection. Otherwise, it checks whether
// os.Stdin is connected to a terminal.
func (h *HeadlessManager) IsHeadless() bool {
	if h.forced != nil {
		return *h.forced
	}
	return !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// ForceHeadless overrides TTY detection. Pass true to force headless mode,
// or false to force interactive mode regardless of TTY state.
func (h *HeadlessManager) ForceHeadless(force bool) {
	h.forced = &force
}

// ClearForce removes any forced override, reverting to automatic TTY detection.
func (h *HeadlessManager) ClearForce() {
	h.forced = nil
}

// SetDefaults stores default values used in headless mode.
// Keys should match the field names expected by each component
// (e.g., "project_name", "language", "framework").
func (h *HeadlessManager) SetDefaults(defaults map[string]string) {
	if len(defaults) == 0 {
		h.defaults = nil
		return
	}
	h.defaults = make(map[string]string, len(defaults))
	maps.Copy(h.defaults, defaults)
}

// GetDefault retrieves a default value by key. The second return value
// indicates whether the key was found.
func (h *HeadlessManager) GetDefault(key string) (string, bool) {
	if h.defaults == nil {
		return "", false
	}
	v, ok := h.defaults[key]
	return v, ok
}

// HasDefaults returns true when at least one default value has been set.
func (h *HeadlessManager) HasDefaults() bool {
	return len(h.defaults) > 0
}
