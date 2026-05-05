package constitution

import (
	"fmt"
	"regexp"
)

// ruleIDPattern is a regex constant for validating IDs in CONST-V3R2-NNN format.
// NNN is 3 or more digits.
const ruleIDPattern = `^CONST-V3R2-\d{3,}$`

// ruleIDRegexp is the compiled ruleIDPattern.
var ruleIDRegexp = regexp.MustCompile(ruleIDPattern)

// Rule represents a single entry in the zone registry.
// Has exactly 6 exported fields, mapped to registry YAML schema via yaml.v3 tags.
// Directly implements SPEC-V3R2-CON-001 REQ-CON-001-004.
//
// Orphan field (unexported) is set internally by the loader when the referenced file is missing.
// Has no yaml tag, so it does not appear in registry files.
type Rule struct {
	// ID is the unique identifier in CONST-V3R2-NNN format.
	ID string `yaml:"id"`
	// Zone is the zone of the clause (Frozen or Evolvable).
	Zone Zone `yaml:"zone"`
	// File is the path to the rule file where the clause is located.
	File string `yaml:"file"`
	// Anchor is the location anchor within the file (#sectionname or Llinenumber).
	Anchor string `yaml:"anchor"`
	// Clause is the verbatim text of the HARD clause.
	Clause string `yaml:"clause"`
	// CanaryGate indicates whether shadow evaluation is required during amendment.
	// Frozen → true, Evolvable → false (default).
	CanaryGate bool `yaml:"canary_gate"`

	// orphan is set to true by the loader when the referenced file does not exist on disk.
	// Unexported, so it is not serialized to YAML and not visible as exported via Go reflect.
	orphan bool
}

// Orphan returns whether the referenced file does not exist on disk.
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
// Returns an error if any field is empty or if the ID format is invalid.
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
