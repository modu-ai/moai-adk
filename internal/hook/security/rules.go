package security

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ruleManager implements RuleManager interface.
type ruleManager struct {
	defaultRules []string
}

// NewRuleManager creates a new RuleManager instance.
func NewRuleManager() RuleManager {
	return &ruleManager{
		defaultRules: getDefaultOWASPRules(),
	}
}

// FindRulesConfig finds the rules configuration file in project directory.
// Implements REQ-HOOK-110.
// Search order:
// 1. sgconfig.yml in project root
// 2. sgconfig.yaml in project root
// 3. .ast-grep/sgconfig.yml
// 4. .ast-grep/sgconfig.yaml
// 5. .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml
func (rm *ruleManager) FindRulesConfig(projectDir string) string {
	searchPaths := []string{
		filepath.Join(projectDir, "sgconfig.yml"),
		filepath.Join(projectDir, "sgconfig.yaml"),
		filepath.Join(projectDir, ".ast-grep", "sgconfig.yml"),
		filepath.Join(projectDir, ".ast-grep", "sgconfig.yaml"),
		filepath.Join(projectDir, ".claude", "skills", "moai-tool-ast-grep", "rules", "sgconfig.yml"),
		filepath.Join(projectDir, ".claude", "skills", "moai-tool-ast-grep", "rules", "sgconfig.yaml"),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// LoadRules loads rules from a configuration file.
// Implements REQ-HOOK-110, REQ-HOOK-112.
func (rm *ruleManager) LoadRules(configPath string) ([]string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config sgConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	rules := []string{}

	// Extract rule IDs from config
	for ruleID := range config.Rules {
		rules = append(rules, ruleID)
	}

	// Add rule directories if specified
	for _, dir := range config.RuleDirs {
		rules = append(rules, fmt.Sprintf("dir:%s", dir))
	}

	return rules, nil
}

// GetDefaultRules returns the default OWASP security rules.
// Implements REQ-HOOK-111.
func (rm *ruleManager) GetDefaultRules() []string {
	return rm.defaultRules
}

// GetEffectiveRules returns rules to use for a project.
// If project has a config, use that; otherwise use defaults.
// Implements REQ-HOOK-111, REQ-HOOK-112.
func (rm *ruleManager) GetEffectiveRules(projectDir string) []string {
	configPath := rm.FindRulesConfig(projectDir)
	if configPath == "" {
		return rm.GetDefaultRules()
	}

	rules, err := rm.LoadRules(configPath)
	if err != nil {
		// REQ-HOOK-112: Fall back to default rules on invalid config
		return rm.GetDefaultRules()
	}

	if len(rules) == 0 {
		return rm.GetDefaultRules()
	}

	return rules
}

// sgConfig represents ast-grep configuration file structure.
type sgConfig struct {
	RuleDirs []string          `yaml:"ruleDirs"`
	Rules    map[string]sgRule `yaml:"rules"`
	Utils    map[string]any    `yaml:"utils"`
}

// sgRule represents a single ast-grep rule.
type sgRule struct {
	Severity string `yaml:"severity"`
	Message  string `yaml:"message"`
	Pattern  string `yaml:"pattern"`
	Language string `yaml:"language"`
	Fix      string `yaml:"fix"`
}

// getDefaultOWASPRules returns built-in OWASP security rules.
// Per SPEC 4.3.
func getDefaultOWASPRules() []string {
	return []string{
		// SQL Injection patterns
		`rule:sql-injection
  severity: error
  message: Potential SQL injection vulnerability
  patterns:
    - execute($$$QUERY)
    - cursor.execute($QUERY)
    - db.query($QUERY)
    - connection.execute($QUERY)`,

		// XSS patterns
		`rule:xss-vulnerability
  severity: error
  message: Potential XSS vulnerability
  patterns:
    - innerHTML = $INPUT
    - document.write($INPUT)
    - element.innerHTML = $INPUT`,

		// Hardcoded secrets
		`rule:hardcoded-secret
  severity: error
  message: Hardcoded secret or password detected
  patterns:
    - password = $PASSWORD
    - api_key = $KEY
    - secret = $SECRET
    - token = $TOKEN
    - apiKey = $KEY
    - secretKey = $KEY`,

		// Insecure random
		`rule:insecure-random
  severity: warning
  message: Insecure random number generator - use crypto module instead
  patterns:
    - Math.random()
    - random.random()
    - rand()`,

		// Eval usage
		`rule:dangerous-eval
  severity: error
  message: Dangerous use of eval() - avoid executing dynamic code
  patterns:
    - eval($$$CODE)
    - exec($$$CODE)
    - Function($$$CODE)`,

		// Command injection
		`rule:command-injection
  severity: error
  message: Potential command injection vulnerability
  patterns:
    - os.system($CMD)
    - subprocess.call($CMD, shell=True)
    - exec($CMD)
    - child_process.exec($CMD)`,

		// Path traversal
		`rule:path-traversal
  severity: error
  message: Potential path traversal vulnerability
  patterns:
    - open($PATH)
    - fs.readFile($PATH)
    - require($PATH)`,

		// Insecure deserialization
		`rule:insecure-deserialization
  severity: error
  message: Insecure deserialization - avoid pickle/marshal with untrusted data
  patterns:
    - pickle.loads($DATA)
    - yaml.load($DATA)
    - marshal.Unmarshal($DATA)`,

		// Debug mode in production
		`rule:debug-enabled
  severity: warning
  message: Debug mode should be disabled in production
  patterns:
    - DEBUG = True
    - debug: true
    - app.run(debug=True)`,

		// CORS wildcard
		`rule:cors-wildcard
  severity: warning
  message: CORS wildcard allows any origin - restrict in production
  patterns:
    - Access-Control-Allow-Origin: *
    - cors({ origin: '*' })`,
	}
}
