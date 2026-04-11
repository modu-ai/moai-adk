package astgrep

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
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

// LoadFromDir recursively loads ast-grep rules from all .yml/.yaml files in dir.
// Individual file parse errors are logged as warnings and skipped (partial success).
// Returns empty slice if dir does not exist. (REQ-ASTG-UPG-011)
func (l *RuleLoader) LoadFromDir(dir string) ([]Rule, error) {
	var allRules []Rule
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		rules, err := l.loadFileSkipOnError(path)
		if err != nil {
			slog.Warn("skipping rule file with parse error", "path", path, "error", err)
			return nil
		}
		allRules = append(allRules, rules...)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return []Rule{}, nil
		}
		return nil, fmt.Errorf("walking rules dir %s: %w", dir, err)
	}

	return allRules, nil
}

// loadFileSkipOnError loads rules from a single YAML file.
// Returns partial results on decode errors to allow partial success per file.
func (l *RuleLoader) loadFileSkipOnError(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening rule file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var rules []Rule
	decoder := yaml.NewDecoder(f)

	for {
		var rule Rule
		if err := decoder.Decode(&rule); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			slog.Warn("partial decode in rule file", "path", path, "error", err)
			break
		}
		if rule.ID == "" {
			continue
		}
		rules = append(rules, rule)
	}

	return rules, nil
}
