package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RequiredChecksFile is the canonical path of the SSoT YAML within a project root.
const RequiredChecksFile = ".github/required-checks.yml"

// BranchChecks holds the required CI check context names for a branch pattern.
type BranchChecks struct {
	// Contexts is the list of required GitHub Actions check context names.
	Contexts []string `yaml:"contexts"`
}

// RequiredChecks is the top-level structure of .github/required-checks.yml.
// It maps branch name patterns (e.g. "main", "release/*") to their required checks
// and lists auxiliary checks that are advisory-only (not blocking PR merge).
type RequiredChecks struct {
	// Version is the schema version (currently 1).
	Version int `yaml:"version"`
	// Branches maps branch name/glob to BranchChecks (used for branch protection JSON).
	Branches map[string]BranchChecks `yaml:"branches"`
	// Auxiliary lists check context names that run on PRs but do NOT block merge.
	// Consumed by Wave 2 CI watch loop to discriminate required vs advisory failures.
	Auxiliary []string `yaml:"auxiliary,omitempty"`
}

// LoadRequiredChecks reads and parses .github/required-checks.yml from the given
// project root. Returns an error with a helpful message if the file is missing.
func LoadRequiredChecks(projectRoot string) (*RequiredChecks, error) {
	path := filepath.Join(projectRoot, RequiredChecksFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"required checks SSoT not found at %s: %w; create %s with the required CI context names",
			RequiredChecksFile, err, RequiredChecksFile,
		)
	}

	var rc RequiredChecks
	if err := yaml.Unmarshal(data, &rc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", RequiredChecksFile, err)
	}

	return &rc, nil
}

// IsAuxiliary reports whether the given check context name is in the Auxiliary list.
// Used by Wave 2 watch loop to filter advisory-only checks out of the
// fail-fast policy.
func (rc *RequiredChecks) IsAuxiliary(context string) bool {
	for _, a := range rc.Auxiliary {
		if a == context {
			return true
		}
	}
	return false
}
