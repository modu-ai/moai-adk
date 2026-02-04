package astgrep

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleLoader loads and manages ast-grep rules from YAML files.
type RuleLoader struct {
	rules []Rule
}

// NewRuleLoader creates a new RuleLoader instance.
func NewRuleLoader() *RuleLoader {
	return &RuleLoader{}
}

// LoadFromFile loads ast-grep rules from a single YAML file.
// Supports multi-document YAML (--- separator).
// Returns an error if the file does not exist or contains invalid YAML.
func (l *RuleLoader) LoadFromFile(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("rule file not found: %s", path)
		}
		return nil, fmt.Errorf("open rule file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var rules []Rule
	decoder := yaml.NewDecoder(f)

	for {
		var rule Rule
		if err := decoder.Decode(&rule); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("parse rule file %s: %w", path, err)
		}
		// Skip empty documents (can occur in multi-doc YAML)
		if rule.ID == "" {
			continue
		}
		rules = append(rules, rule)
	}

	l.rules = append(l.rules, rules...)
	return rules, nil
}

// LoadFromDirectory loads ast-grep rules from all .yml and .yaml files
// in the specified directory. Non-YAML files are ignored.
func (l *RuleLoader) LoadFromDirectory(dir string) ([]Rule, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read rule directory %s: %w", dir, err)
	}

	var allRules []Rule
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		rules, err := l.LoadFromFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("load rules from %s: %w", entry.Name(), err)
		}
		allRules = append(allRules, rules...)
	}

	return allRules, nil
}

// GetRulesForLanguage returns all loaded rules that match the specified language.
func (l *RuleLoader) GetRulesForLanguage(language string) []Rule {
	var filtered []Rule
	lang := strings.ToLower(language)
	for _, r := range l.rules {
		if strings.ToLower(r.Language) == lang {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// Rules returns all loaded rules.
func (l *RuleLoader) Rules() []Rule {
	result := make([]Rule, len(l.rules))
	copy(result, l.rules)
	return result
}
