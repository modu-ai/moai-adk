package safety

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FrozenGuard is a guard that prevents modification of constitutional files.
type FrozenGuard struct {
	frozenPaths []string
}

// NewFrozenGuard creates a guard with the default frozen paths.
func NewFrozenGuard() *FrozenGuard {
	return &FrozenGuard{
		frozenPaths: defaultFrozenPaths(),
	}
}

// NewFrozenGuardWithPaths creates a guard with custom frozen paths.
func NewFrozenGuardWithPaths(paths []string) *FrozenGuard {
	return &FrozenGuard{
		frozenPaths: paths,
	}
}

// IsFrozen returns true if the given path is in the frozen list.
// Uses substring matching to detect frozen paths in absolute paths.
// Applies filepath.Clean for path traversal defense and
// normalizes to forward slashes for Windows compatibility.
func (g *FrozenGuard) IsFrozen(path string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	for _, fp := range g.frozenPaths {
		if strings.Contains(cleaned, fp) {
			return true
		}
	}
	return false
}

// ValidateWrite returns an error if a write is attempted on a frozen path.
func (g *FrozenGuard) ValidateWrite(path string) error {
	if g.IsFrozen(path) {
		return fmt.Errorf("research/safety: cannot write to frozen path: %s", path)
	}
	return nil
}

// defaultFrozenPaths returns the default list of frozen file paths.
func defaultFrozenPaths() []string {
	return []string{
		".claude/rules/moai/core/moai-constitution.md",
		".claude/rules/agency/constitution.md",
		".claude/agents/moai/researcher.md",
		".moai/research/config.yaml",
	}
}
