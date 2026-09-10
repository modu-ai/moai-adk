package astgrep

// Severity-contract primitives for SPEC-ASTGREP-LANG16-001 M3.
//
// REQ-A16-012 makes a promotion to `error` conditional on two clauses, and
// REQ-A16-011 makes a security rule's pattern owe an anchor. Both read as prose
// obligations on the YAML, so both are unenforced until something loads the
// YAML and asserts them. That is what this file exists for: it parses the
// shipped multi-document ruleset and the repo-side rule-test cases so the
// severity contract is a test rather than a review convention.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Severity values used by the shipped ruleset. `error` returns a deny on the
// pre-write path; `warning` reports without refusing the write.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// AnchorPrefixes are the two evidence forms REQ-A16-011's third clause accepts:
// a named reference documenting the matched head symbol, or a recorded probe
// showing the pattern matching real code. Identical to the evidence rule the
// coverage matrix imposes on an EXEMPT cell — a presence claim owes no less
// than an absence claim.
var AnchorPrefixes = []string{"cite:", "probe:"}

// ShippedRule is one document of the shipped ruleset.
type ShippedRule struct {
	ID       string `yaml:"id"`
	Language string `yaml:"language"`
	Severity string `yaml:"severity"`
	Metadata struct {
		OWASP  string `yaml:"owasp"`
		CWE    string `yaml:"cwe"`
		Anchor string `yaml:"anchor"`
	} `yaml:"metadata"`

	// SourceFile is the ruleset file the document came from; the security
	// family is a directory, so membership is read from the path.
	SourceFile string `yaml:"-"`
}

// IsSecurity reports whether the rule belongs to a security family — clause 1
// of REQ-A16-012. Membership is necessary and NOT sufficient for `error`.
func (r ShippedRule) IsSecurity() bool {
	return filepath.Base(filepath.Dir(r.SourceFile)) == "security"
}

// HasAnchor reports whether metadata.anchor carries one of the two accepted
// evidence forms.
func (r ShippedRule) HasAnchor() bool {
	trimmed := strings.TrimSpace(r.Metadata.Anchor)
	for _, p := range AnchorPrefixes {
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

// LoadShippedRules parses every rule document under a ruleset directory tree.
// sgconfig.yml is a project configuration, not a rule, and is skipped.
func LoadShippedRules(rulesetDir string) ([]ShippedRule, error) {
	var rules []ShippedRule
	err := filepath.WalkDir(rulesetDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", path, walkErr)
		}
		if d.IsDir() || filepath.Base(path) == "sgconfig.yml" {
			return nil
		}
		if ext := filepath.Ext(path); ext != ".yml" && ext != ".yaml" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		dec := yaml.NewDecoder(strings.NewReader(string(data)))
		for {
			var r ShippedRule
			decErr := dec.Decode(&r)
			if decErr != nil {
				break // io.EOF, or a non-rule document; ID check below filters
			}
			if r.ID == "" {
				continue
			}
			r.SourceFile = path
			rules = append(rules, r)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// RuleTestCases is one rule-test document: the valid/invalid case pair whose
// structural proof REQ-A16-012 clause 2 rests on.
type RuleTestCases struct {
	ID      string   `yaml:"id"`
	Valid   []string `yaml:"valid"`
	Invalid []string `yaml:"invalid"`

	SourceFile string `yaml:"-"`
}

// LoadRuleTestCases parses every rule-test document under a test root, keyed
// by rule id. ast-grep keys cases by id alone, so several rule documents
// sharing an id share one case document.
func LoadRuleTestCases(testDir string) (map[string]RuleTestCases, error) {
	out := map[string]RuleTestCases{}
	err := filepath.WalkDir(testDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", path, walkErr)
		}
		if d.IsDir() {
			// __snapshots__ holds generated snapshots, not case documents.
			if d.Name() == "__snapshots__" {
				return fs.SkipDir
			}
			return nil
		}
		if ext := filepath.Ext(path); ext != ".yml" && ext != ".yaml" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		var tc RuleTestCases
		if err := yaml.Unmarshal(data, &tc); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		if tc.ID == "" {
			return nil
		}
		tc.SourceFile = path
		out[tc.ID] = tc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
