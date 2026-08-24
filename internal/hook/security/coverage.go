package security

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// LanguageCoverage is the result of resolving a project's ast-grep rules
// configuration and deriving which languages that configuration actually
// carries rules for.
//
// Three states are distinguishable, and the caller MUST treat them differently:
//
//   - ConfigPath == ""            no configuration resolved at all. A scan can
//     produce no finding, so the caller skips it.
//   - ConfigPath != "", !Known    a configuration resolved but the covered-
//     language set could not be derived (unreadable, unparseable, an unwalkable
//     ruleDir, or an empty derived set). The caller escalates and scans anyway:
//     an unreadable or empty result is never evidence that a language is
//     uncovered.
//   - ConfigPath != "", Known     the derived set is authoritative; a language
//     absent from it has no rule that could produce a finding.
type LanguageCoverage struct {
	// ConfigPath is the resolved sgconfig path, or "" when none resolved.
	ConfigPath string
	// Languages is the sorted set of ast-grep language names the resolved
	// configuration declares at least one rule for. Meaningful only when Known.
	Languages []string
	// Known reports whether Languages is an authoritative answer.
	Known bool
}

// Covers reports whether the derived set names the given language. It answers
// false for an unknown derivation, so callers MUST check Known before treating
// a false as "this language has no rules".
func (c LanguageCoverage) Covers(language string) bool {
	if !c.Known || language == "" {
		return false
	}
	for _, l := range c.Languages {
		if l == language {
			return true
		}
	}
	return false
}

// ResolveCoverage resolves the rules configuration for projectDir and derives
// the set of languages it declares at least one rule for, by walking the
// configuration's ruleDirs and reading each rule document's `language:` field.
//
// The result is cached for the lifetime of this RuleManager. The hook process
// is short-lived, so this is a single load per invocation rather than a
// long-lived cache with invalidation concerns.
func (rm *ruleManager) ResolveCoverage(projectDir string) LanguageCoverage {
	rm.coverageMu.Lock()
	defer rm.coverageMu.Unlock()

	if rm.coverageCache == nil {
		rm.coverageCache = make(map[string]LanguageCoverage)
	}
	if cached, ok := rm.coverageCache[projectDir]; ok {
		return cached
	}

	cov := deriveCoverage(rm.FindRulesConfig(projectDir))
	rm.coverageCache[projectDir] = cov
	return cov
}

// deriveCoverage does the derivation for an already-resolved config path. Every
// failure path returns Known=false (fail-open), never an empty known set.
func deriveCoverage(configPath string) LanguageCoverage {
	if configPath == "" {
		return LanguageCoverage{}
	}
	cov := LanguageCoverage{ConfigPath: configPath}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cov
	}
	var config sgConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return cov
	}

	found := map[string]struct{}{}
	for _, rule := range config.Rules {
		if lang := strings.TrimSpace(rule.Language); lang != "" {
			found[lang] = struct{}{}
		}
	}

	base := filepath.Dir(configPath)
	for _, dir := range config.RuleDirs {
		ruleDir := dir
		if !filepath.IsAbs(ruleDir) {
			ruleDir = filepath.Join(base, ruleDir)
		}
		if err := collectRuleLanguages(ruleDir, found); err != nil {
			// A ruleDir that cannot be read or walked leaves the derivation
			// incomplete, and an incomplete set would skip languages whose
			// rules simply could not be seen.
			return cov
		}
	}

	if len(found) == 0 {
		return cov
	}

	cov.Languages = make([]string, 0, len(found))
	for lang := range found {
		cov.Languages = append(cov.Languages, lang)
	}
	sort.Strings(cov.Languages)
	cov.Known = true
	return cov
}

// collectRuleLanguages walks one rule directory, adding every rule document's
// declared language to found. Any read or parse failure is returned as an
// error so the caller can mark the whole derivation unknown.
func collectRuleLanguages(ruleDir string, found map[string]struct{}) error {
	info, err := os.Stat(ruleDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("ruleDir is not a directory: " + ruleDir)
	}

	return filepath.WalkDir(ruleDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".yml", ".yaml":
		default:
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		// Rule files are multi-document YAML: the shipped ruleset packs one
		// document per language into a single file.
		dec := yaml.NewDecoder(file)
		for {
			var doc struct {
				Language string `yaml:"language"`
			}
			decErr := dec.Decode(&doc)
			if errors.Is(decErr, io.EOF) {
				return nil
			}
			if decErr != nil {
				return decErr
			}
			if lang := strings.TrimSpace(doc.Language); lang != "" {
				found[lang] = struct{}{}
			}
		}
	})
}

// coverageState carries the per-manager derivation cache. It is embedded in
// ruleManager rather than held globally so each manager (and therefore each
// test) starts from a clean cache.
type coverageState struct {
	coverageMu    sync.Mutex
	coverageCache map[string]LanguageCoverage
}
