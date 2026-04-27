// Package harness — chaining-rules.yaml read/write.
package harness

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ChainEntry is one row in `.moai/harness/chaining-rules.yaml` describing how
// a phase-bound agent invocation is augmented with my-harness/* agents.
type ChainEntry struct {
	Phase         string            `yaml:"phase"`
	When          map[string]string `yaml:"when"`
	InsertBefore  []string          `yaml:"insert_before"`
	InsertAfter   []string          `yaml:"insert_after"`
}

// ChainingRules is the top-level document for chaining-rules.yaml.
type ChainingRules struct {
	Version int          `yaml:"version"`
	Chains  []ChainEntry `yaml:"chains"`
}

// WriteChainingRules marshals rules to YAML and writes to yamlPath. The caller
// is responsible for path validation; layer code does NOT call EnsureAllowed.
// Tests freely use t.TempDir() paths.
func WriteChainingRules(yamlPath string, rules ChainingRules) error {
	if yamlPath == "" {
		return errors.New("WriteChainingRules: empty path")
	}
	// Normalize empty arrays to non-nil so they marshal as `[]` not omitted.
	for i := range rules.Chains {
		if rules.Chains[i].InsertBefore == nil {
			rules.Chains[i].InsertBefore = []string{}
		}
		if rules.Chains[i].InsertAfter == nil {
			rules.Chains[i].InsertAfter = []string{}
		}
		if rules.Chains[i].When == nil {
			rules.Chains[i].When = map[string]string{}
		}
	}
	data, err := yaml.Marshal(rules)
	if err != nil {
		return fmt.Errorf("WriteChainingRules: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0o644); err != nil {
		return fmt.Errorf("WriteChainingRules: write %s: %w", yamlPath, err)
	}
	return nil
}

// ReadChainingRules reads and unmarshals the chaining-rules.yaml document
// at yamlPath. Returns a populated ChainingRules struct on success.
func ReadChainingRules(yamlPath string) (ChainingRules, error) {
	var rules ChainingRules
	if yamlPath == "" {
		return rules, errors.New("ReadChainingRules: empty path")
	}
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return rules, fmt.Errorf("ReadChainingRules: read %s: %w", yamlPath, err)
	}
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return rules, fmt.Errorf("ReadChainingRules: parse %s: %w", yamlPath, err)
	}
	return rules, nil
}
