package astgrep

import (
	"bytes"
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

// loadFileSkipOnError는 단일 YAML 파일에서 규칙을 로딩합니다.
// 멀티 문서 YAML에서 특정 문서 파싱이 실패해도 이후 문서를 계속 로딩합니다 (F5 fix).
//
// yaml.v3 Decoder는 파싱 실패 후 상태를 복구하지 않으므로, 파일을 "---" 구분자로
// 분리하여 각 문서를 독립적으로 파싱하는 방식을 사용합니다.
func (l *RuleLoader) loadFileSkipOnError(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening rule file %s: %w", path, err)
	}

	// "---" 구분자로 문서 분리 (멀티 문서 YAML 지원)
	// 각 문서를 개별적으로 파싱하여 파싱 실패 격리
	docs := splitYAMLDocs(data)

	var rules []Rule
	var parseErrors int

	for i, doc := range docs {
		trimmed := bytes.TrimSpace(doc)
		if len(trimmed) == 0 {
			continue
		}

		var rule Rule
		if err := yaml.Unmarshal(trimmed, &rule); err != nil {
			parseErrors++
			slog.Warn("규칙 파일 문서 파싱 실패, 다음 문서로 진행",
				"path", path, "doc_index", i, "error", err)
			continue
		}
		if rule.ID == "" {
			continue
		}
		rules = append(rules, rule)
	}

	if parseErrors > 0 {
		slog.Warn("규칙 파일에서 파싱 오류 발생",
			"path", path, "valid_rules", len(rules), "parse_errors", parseErrors)
	}

	return rules, nil
}

// splitYAMLDocs는 YAML 멀티 문서 데이터를 "---" 구분자로 분리합니다.
// 각 문서를 개별 파싱을 위한 독립적인 바이트 슬라이스로 반환합니다.
func splitYAMLDocs(data []byte) [][]byte {
	var docs [][]byte
	// \n--- 또는 파일 시작 --- 로 분리
	// bytes.Split은 정확히 "---"만 있는 행을 기준으로 분리
	lines := bytes.Split(data, []byte("\n"))
	var current [][]byte
	for _, line := range lines {
		if bytes.Equal(bytes.TrimSpace(line), []byte("---")) {
			if len(current) > 0 {
				docs = append(docs, bytes.Join(current, []byte("\n")))
			}
			current = nil
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		docs = append(docs, bytes.Join(current, []byte("\n")))
	}
	return docs
}
