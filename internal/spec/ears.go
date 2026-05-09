package spec

import (
	"fmt"
	"regexp"
	"strings"
)

// Acceptance represents an Acceptance Criteria in EARS format.
// Supports hierarchical structure up to 3 levels deep.
type Acceptance struct {
	ID             string        // AC-<DOMAIN>-<NNN>-<NN> format (e.g., AC-SPC-001-05)
	Given          string        // Precondition
	When           string        // Trigger event
	Then           string        // Expected result
	RequirementIDs []string      // Referenced REQ IDs (e.g., ["SPC-001", "SPC-002"])
	Children       []Acceptance  // Child Acceptance Criteria (hierarchical)
}

// MaxDepth is the maximum allowed depth of Acceptance tree
const MaxDepth = 3

// topLevelIDPattern validates top-level AC ID format
// AC-SPC-001-05
var topLevelIDPattern = regexp.MustCompile(`^AC-[A-Z0-9]+-[0-9]+-[0-9]+$`)

// ValidateID validates if Acceptance ID is in valid format
func (a *Acceptance) ValidateID() error {
	if !topLevelIDPattern.MatchString(a.ID) {
		return fmt.Errorf("invalid acceptance criteria ID format: %s (expected AC-<DOMAIN>-<NNN>-<NN>)", a.ID)
	}
	return nil
}

func (a *Acceptance) GenerateChildID(depth int, index int) (string, error) {
	if depth < 1 || depth > 2 {
		return "", fmt.Errorf("invalid depth for child generation: %d (must be 1 or 2)", depth)
	}

	if depth == 1 {
		if index < 0 || index > 25 {
			return "", fmt.Errorf("invalid index for level 1 child: %d (must be 0-25)", index)
		}
		suffix := string(rune('a' + index))
		return fmt.Sprintf("%s.%s", a.ID, suffix), nil
	}

	// a.i, .a.ii, ..., .a.xxvi
	romanNumerals := []string{
		"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x",
		"xi", "xii", "xiii", "xiv", "xv", "xvi", "xvii", "xviii", "xix", "xx",
		"xxi", "xxii", "xxiii", "xxiv", "xxv", "xxvi",
	}
	if index < 0 || index >= len(romanNumerals) {
		return "", fmt.Errorf("invalid index for level 2 child: %d (must be 0-25)", index)
	}

	return fmt.Sprintf("%s.%s", a.ID, romanNumerals[index]), nil
}

// Depth returns the depth of current node
func (a *Acceptance) Depth() int {
	parts := strings.Split(a.ID, ".")
	// AC-XXX-NNN-NN (parts length 1)
	// AC-XXX-NNN-NN.a (parts length 2)
	// AC-XXX-NNN-NN.a.i (parts length 3)
	return len(parts) - 1
}

// InheritGiven inherits parent's Given when child's Given is empty
func (a *Acceptance) InheritGiven(parent *Acceptance) {
	if a.Given == "" && parent != nil && parent.Given != "" {
		a.Given = parent.Given
	}
}

// IsLeaf verifies if this node is a leaf (has no children)
func (a *Acceptance) IsLeaf() bool {
	return len(a.Children) == 0
}

// CountLeaves returns the total number of leaf nodes in the tree
func (a *Acceptance) CountLeaves() int {
	if a.IsLeaf() {
		return 1
	}

	count := 0
	for i := range a.Children {
		count += a.Children[i].CountLeaves()
	}
	return count
}

// ValidateDepth validates that tree depth does not exceed MaxDepth
func (a *Acceptance) ValidateDepth() error {
	if a.Depth() >= MaxDepth {
		return &MaxDepthExceeded{
			ID:    a.ID,
			Depth: a.Depth(),
			Max:   MaxDepth - 1,
		}
	}

	for i := range a.Children {
		if err := a.Children[i].ValidateDepth(); err != nil {
			return err
		}
	}

	return nil
}

// ExtractRequirementMappings extracts (maps REQ-...) pattern from text
func ExtractRequirementMappings(text string) []string {
	pattern := regexp.MustCompile(`\(?\s*(?:maps|MAPS)\s+REQ-([A-Z0-9-]+)\s*\)?`)
	matches := pattern.FindAllStringSubmatch(text, -1)

	var reqIDs []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			reqID := match[1]
			if !seen[reqID] {
				seen[reqID] = true
				reqIDs = append(reqIDs, reqID)
			}
		}
	}

	return reqIDs
}

// ValidateRequirementMappings validates that all leaf nodes have REQ mapping
func (a *Acceptance) ValidateRequirementMappings() []error {
	var errors []error

	a.validateRequirementMappingsRecursive(&errors)
	return errors
}

func (a *Acceptance) validateRequirementMappingsRecursive(errors *[]error) {
	if a.IsLeaf() {
		if len(a.RequirementIDs) == 0 {
			*errors = append(*errors, &MissingRequirementMapping{ACID: a.ID})
		}
	} else {
		for i := range a.Children {
			a.Children[i].validateRequirementMappingsRecursive(errors)
		}
	}
}
