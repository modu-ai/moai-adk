// Package ciwatch provides helpers for CI watch loop (Wave 2, T2).
// It classifies GitHub Actions check runs as required vs auxiliary
// by reading the SSoT at .github/required-checks.yml.
package ciwatch

import (
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Classifier resolves whether a given CI check name is required for a branch.
// It caches the loaded SSoT for the process lifetime (single-read pattern).
type Classifier struct {
	rc *config.RequiredChecks
}

// NewClassifier loads the required-checks SSoT from projectRoot and returns
// a Classifier. Returns an error if the SSoT file is missing or malformed.
func NewClassifier(projectRoot string) (*Classifier, error) {
	rc, err := config.LoadRequiredChecks(projectRoot)
	if err != nil {
		return nil, err
	}
	return &Classifier{rc: rc}, nil
}

// IsRequired reports whether checkName is a required CI check for the given
// branch name. Branch patterns in the SSoT are matched using filepath.Match
// (e.g. "release/*" matches "release/v1.0.0"). Returns false if checkName is
// in the auxiliary list, or if no branch pattern matches.
func (c *Classifier) IsRequired(checkName, branch string) bool {
	// Auxiliary checks are never required, regardless of branch.
	if c.rc.IsAuxiliary(checkName) {
		return false
	}

	for pattern, bc := range c.rc.Branches {
		if !matchBranch(pattern, branch) {
			continue
		}
		for _, ctx := range bc.Contexts {
			if ctx == checkName {
				return true
			}
		}
	}
	return false
}

// matchBranch reports whether branchName satisfies the given pattern.
// Exact match wins; otherwise filepath.Match is used for glob patterns
// like "release/*".
func matchBranch(pattern, branchName string) bool {
	if pattern == branchName {
		return true
	}
	// filepath.Match is POSIX path-style; branch names use "/" so it works.
	matched, err := filepath.Match(pattern, branchName)
	if err != nil {
		// Invalid glob pattern — treat as no-match rather than panic.
		return false
	}
	return matched
}

// RequiredNames returns the list of required check names for the given branch.
// Auxiliary checks are excluded. Returns nil if no pattern matches.
func (c *Classifier) RequiredNames(branch string) []string {
	seen := map[string]bool{}
	var out []string

	for pattern, bc := range c.rc.Branches {
		if !matchBranch(pattern, branch) {
			continue
		}
		for _, ctx := range bc.Contexts {
			if !seen[ctx] && !c.rc.IsAuxiliary(ctx) {
				seen[ctx] = true
				out = append(out, ctx)
			}
		}
	}
	return out
}

// AuxiliaryNames returns the auxiliary (advisory-only) check names from the SSoT.
func (c *Classifier) AuxiliaryNames() []string {
	result := make([]string, len(c.rc.Auxiliary))
	copy(result, c.rc.Auxiliary)
	return result
}

// IsAuxiliary reports whether checkName is an auxiliary (advisory-only) check.
func (c *Classifier) IsAuxiliary(checkName string) bool {
	return c.rc.IsAuxiliary(checkName)
}

// IsKnown reports whether a check name appears anywhere in the SSoT
// (either as required for some branch or as auxiliary).
func (c *Classifier) IsKnown(checkName string) bool {
	if c.rc.IsAuxiliary(checkName) {
		return true
	}
	for _, bc := range c.rc.Branches {
		for _, ctx := range bc.Contexts {
			if strings.EqualFold(ctx, checkName) {
				return true
			}
		}
	}
	return false
}
