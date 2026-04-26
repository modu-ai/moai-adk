package constitution

import (
	"fmt"
	"regexp"
)

// ruleIDPattern is the regular expression constant that validates IDs in the CONST-V3R2-NNN format.
// NNN is a numeric string of 3 or more digits.
const ruleIDPattern = `^CONST-V3R2-\d{3,}$`

// ruleIDRegexp is the compiled ruleIDPattern.
var ruleIDRegexp = regexp.MustCompile(ruleIDPattern)

// Rule represents a single entry in the zone registry.
// It has exactly 6 exported fields mapped to the registry YAML schema via yaml.v3 tags.
// Directly implements SPEC-V3R2-CON-001 REQ-CON-001-004.
//
// The orphan field (unexported) is set internally by the loader when the referenced file is absent.
// It has no yaml tag, so it does not appear in the registry file and is not visible as exported via Go reflect.
type Rule struct {
	// ID is the unique identifier in the CONST-V3R2-NNN format.
	ID string `yaml:"id"`
	// Zone is the zone of the clause (Frozen or Evolvable).
	Zone Zone `yaml:"zone"`
	// File is the path to the rule file containing the clause.
	File string `yaml:"file"`
	// Anchor is the in-file location anchor (#section-name or Lline-number).
	Anchor string `yaml:"anchor"`
	// Clause is the verbatim text of the HARD clause.
	Clause string `yaml:"clause"`
	// CanaryGate indicates whether shadow evaluation is required on amendment.
	// Frozen → true, Evolvable → false (default).
	CanaryGate bool `yaml:"canary_gate"`

	// orphan is set to true by the loader when the referenced file does not exist on disk.
	// Unexported, so it is not serialized to YAML and is not visible as exported via Go reflect.
	orphan bool
}

// Orphan reports whether the referenced file does not exist on disk.
// Set by the loader during LoadRegistry after verifying file existence.
func (r Rule) Orphan() bool {
	return r.orphan
}

// withOrphan is an internal helper that returns a new Rule with the orphan flag set.
func (r Rule) withOrphan(orphan bool) Rule {
	r.orphan = orphan
	return r
}

// Validate validates the required fields of the Rule.
// Returns an error when a required field is empty or the ID format is invalid.
func (r Rule) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("rule ID is empty")
	}
	if !ruleIDRegexp.MatchString(r.ID) {
		return fmt.Errorf("rule ID %q does not match pattern %q", r.ID, ruleIDPattern)
	}
	if r.File == "" {
		return fmt.Errorf("rule %q: File field is empty", r.ID)
	}
	if r.Clause == "" {
		return fmt.Errorf("rule %q: Clause field is empty", r.ID)
	}
	return nil
}
