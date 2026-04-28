package astgrep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ScannerConfig holds configuration for the unified Scanner.
// REQ-ASTG-UPG-010, REQ-ASTG-UPG-011, REQ-ASTG-UPG-012
type ScannerConfig struct {
	// RulesDir is the ast-grep rules directory path (recursive search).
	// Default: ".moai/config/astgrep-rules"
	RulesDir string
	// SGBinary is the sg CLI binary name or path.
	// Default: "sg"
	SGBinary string
	// WarnOnlyMode prevents blocking even when error-severity findings are detected.
	WarnOnlyMode bool
	// Timeout is the total scan timeout.
	// Default: 30 seconds
	Timeout time.Duration
}

// DefaultScannerConfig returns a ScannerConfig with built-in default values.
func DefaultScannerConfig() *ScannerConfig {
	return &ScannerConfig{
		RulesDir: ".moai/config/astgrep-rules",
		SGBinary: "sg",
		Timeout:  30 * time.Second,
	}
}

// Finding represents a single result from an ast-grep scan.
// It is the canonical shared type used by the quality gate, post-tool hook, and CLI subcommands.
//
// @MX:ANCHOR: [AUTO] Finding is the canonical shared data type for scanner, CLI, and hook
// @MX:REASON: fan_in >= 3: Scanner.Scan, SARIF converter, and hook integration all use this type
type Finding struct {
	// RuleID is the rule identifier.
	RuleID string `json:"ruleId"`
	// Severity is the finding severity: "error", "warning", "info"
	Severity string `json:"severity"`
	// Message is the rule message.
	Message string `json:"message"`
	// File is the path of the file where the finding was detected.
	File string `json:"file"`
	// Line is the 1-indexed line number.
	Line int `json:"line"`
	// Column is the 0-indexed column number.
	Column int `json:"column,omitempty"`
	// EndLine is the 1-indexed end line number.
	EndLine int `json:"endLine,omitempty"`
	// EndColumn is the end column number.
	EndColumn int `json:"endColumn,omitempty"`
	// Note is the additional description from the rule.
	Note string `json:"note,omitempty"`
	// Metadata is additional metadata such as CWE/OWASP.
	Metadata map[string]string `json:"metadata,omitempty"`
	// Language is the target language of the rule that produced this finding.
	// Injected from rule.Language via the scanWithRules path.
	// Empty string in the scanWithConfig path (per-finding language is unknown).
	// The --lang filter always includes findings with an empty Language (language-neutral rules).
	Language string `json:"language,omitempty"`
}

// IsError reports whether the severity is "error".
func (f Finding) IsError() bool {
	return strings.ToLower(f.Severity) == "error"
}

// IsWarning reports whether the severity is "warning".
func (f Finding) IsWarning() bool {
	return strings.ToLower(f.Severity) == "warning"
}

// IsInfo reports whether the severity is "info" or empty string.
func (f Finding) IsInfo() bool {
	s := strings.ToLower(f.Severity)
	return s == "info" || s == ""
}

// String returns a human-readable representation of the Finding.
func (f Finding) String() string {
	sev := f.Severity
	if sev == "" {
		sev = "info"
	}
	return fmt.Sprintf("%s:%d: [%s] %s (%s)", f.File, f.Line, f.RuleID, f.Message, sev)
}

// HasErrors reports whether any finding in the slice has error severity.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.IsError() {
			return true
		}
	}
	return false
}

// ErrUntrustedBinary is returned when an untrusted binary path is specified.
var ErrUntrustedBinary = errors.New("astgrep: untrusted binary path")

// trustedBinaryPrefixes returns the list of allowed absolute path prefixes.
func trustedBinaryPrefixes() []string {
	home, _ := os.UserHomeDir()
	sep := string(os.PathSeparator)
	prefixes := []string{
		"/usr/bin/",
		"/usr/local/bin/",
		"/opt/homebrew/bin/",
	}
	if home != "" {
		prefixes = append(prefixes,
			filepath.Join(home, "go", "bin")+sep,
			filepath.Join(home, ".local", "bin")+sep,
			filepath.Join(home, ".cargo", "bin")+sep,
		)
	}
	return prefixes
}

// ValidateBinary checks the safety of the sg binary path.
// An empty string is allowed because it falls back to the default "sg".
// Bare names are restricted to "sg" or "ast-grep".
// Absolute paths must be in the trusted prefix list.
// Shell metacharacters or path traversal (..) return ErrUntrustedBinary.
func ValidateBinary(binary string) error {
	if binary == "" {
		// Empty value falls back to the default "sg".
		return nil
	}
	// Shell injection defense: block metacharacters.
	if strings.ContainsAny(binary, ";|&`$()<>\n\r") {
		return ErrUntrustedBinary
	}
	// Block path traversal.
	if strings.Contains(binary, "..") {
		return ErrUntrustedBinary
	}
	// Bare name: only allow fixed values from the allowlist.
	// filepath.IsAbs does not treat Unix-style "/usr/bin/sg" as absolute on Windows,
	// so path-ness is determined by checking for slashes/backslashes to ensure
	// cross-platform consistency.
	looksLikePath := strings.ContainsAny(binary, "/\\") || filepath.IsAbs(binary)
	if !looksLikePath {
		if binary == "sg" || binary == "ast-grep" {
			return nil
		}
		return ErrUntrustedBinary
	}
	// Absolute path: check trusted prefix list (after normalizing traversal with Clean).
	// Normalize both sides with ToSlash for consistent Unix/Windows handling.
	cleaned := filepath.ToSlash(filepath.Clean(binary))
	for _, p := range trustedBinaryPrefixes() {
		prefixSlash := strings.TrimRight(filepath.ToSlash(p), "/")
		if strings.HasPrefix(cleaned, prefixSlash+"/") || cleaned == prefixSlash {
			return nil
		}
	}
	return ErrUntrustedBinary
}

// Scanner is a unified code scanner based on ast-grep.
// It replaces the separate implementations in the quality gate and post-tool hook.
//
// @MX:ANCHOR: [AUTO] Scanner.Scan is the single entry point for all ast-grep scans
// @MX:REASON: fan_in >= 3: quality gate hook, PostToolUse hook, and CLI subcommand all call this method
type Scanner struct {
	cfg *ScannerConfig
}

// NewScanner creates a new Scanner with the given configuration.
// If cfg is nil, DefaultScannerConfig() is used.
func NewScanner(cfg *ScannerConfig) *Scanner {
	if cfg == nil {
		cfg = DefaultScannerConfig()
	}
	return &Scanner{cfg: cfg}
}

// sgScanMatch is the internal parsing struct for sg scan --json output.
type sgScanMatch struct {
	File     string `json:"file"`
	Lines    string `json:"lines,omitempty"`
	Text     string `json:"text,omitempty"`
	RuleID   string `json:"ruleId,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
	Note     string `json:"note,omitempty"`
	Range    struct {
		Start struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"start"`
		End struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"end"`
	} `json:"range"`
}

// isSGAvailable reports whether the sg CLI binary is available in PATH.
func (s *Scanner) isSGAvailable() bool {
	binary := s.cfg.SGBinary
	if binary == "" {
		binary = "sg"
	}
	_, err := exec.LookPath(binary)
	return err == nil
}

// Scan runs all rules against the given path.
// Returns ([]Finding{}, nil) when the sg CLI is not available (REQ-ASTG-UPG-012).
// Returns ([]Finding{}, nil) when the rules directory is empty or does not exist (REQ-ASTG-UPG-012).
// Returns an error when SGBinary is an untrusted path (F2 security check).
func (s *Scanner) Scan(ctx context.Context, path string) ([]Finding, error) {
	// Binary path security validation (F2): untrusted paths return an error immediately.
	if err := ValidateBinary(s.cfg.SGBinary); err != nil {
		return []Finding{}, fmt.Errorf("sg binary validation failed (SGBinary=%q): %w", s.cfg.SGBinary, err)
	}

	// Warn and skip when the sg CLI is not available (REQ-ASTG-UPG-012).
	if !s.isSGAvailable() {
		slog.Warn("ast-grep (sg) CLI not found; skipping scan",
			"binary", s.cfg.SGBinary,
			"hint", "install from https://ast-grep.github.io/guide/quick-start.html")
		return []Finding{}, nil
	}

	// Return empty results when the rules directory does not exist
	if _, err := os.Stat(s.cfg.RulesDir); err != nil {
		if os.IsNotExist(err) {
			return []Finding{}, nil
		}
		return []Finding{}, nil
	}

	// Use config-based scan when sgconfig.yml is present
	sgconfigPath := filepath.Join(s.cfg.RulesDir, "sgconfig.yml")
	if _, err := os.Stat(sgconfigPath); err == nil {
		return s.scanWithConfig(ctx, sgconfigPath, path)
	}

	// When sgconfig.yml is absent, load rules recursively and scan
	loader := NewRuleLoader()
	rules, err := loader.LoadFromDir(s.cfg.RulesDir)
	if err != nil || len(rules) == 0 {
		return []Finding{}, nil
	}

	return s.scanWithRules(ctx, rules, path)
}

// scanWithConfig scans using an sgconfig.yml file.
func (s *Scanner) scanWithConfig(ctx context.Context, configPath, path string) ([]Finding, error) {
	timeout := s.cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	binary := s.cfg.SGBinary
	if binary == "" {
		binary = "sg"
	}

	cmd := exec.CommandContext(scanCtx, binary, "scan", "--config", configPath, "--json", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ignore the error: sg may return a non-zero exit code when matches are found.
	_ = cmd.Run()

	// F4: debug-log any stderr content (language is unknown in config-based scans, so skip error propagation).
	if stderr.Len() > 0 {
		slog.Debug("sg scan stderr (config-based)", "config", configPath, "stderr", stderr.String())
	}

	if stdout.Len() == 0 {
		return []Finding{}, nil
	}

	return parseSGFindings(stdout.Bytes())
}

// runSingleRule executes sg run for a single rule and returns the findings.
// The context is protected with defer cancel() to prevent leaks (F3).
// stderr is debug-logged; an error is returned when stdout is empty and stderr is non-empty (F4).
//
// @MX:WARN: [AUTO] context.WithTimeout guarded by defer cancel() — previous implementation leaked cancel()
// @MX:REASON: F3 bug fix: without defer, cancel() was leaked on panic/early-return paths inside the loop
func (s *Scanner) runSingleRule(ctx context.Context, binary string, rule Rule, path string, timeout time.Duration) ([]Finding, error) {
	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(scanCtx, binary, "run",
		"--pattern", rule.Pattern,
		"--lang", rule.Language,
		"--json",
		path,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// F4: log stderr content.
		if stderr.Len() > 0 {
			slog.Debug("sg run stderr", "rule", rule.ID, "stderr", stderr.String())
		}
		// Treat as execution failure when stdout is empty but stderr has content.
		if stdout.Len() == 0 && stderr.Len() > 0 {
			return nil, fmt.Errorf("sg run failed (rule %s): %s", rule.ID, stderr.String())
		}
	}

	if stdout.Len() == 0 {
		return nil, nil
	}
	return parseSGFindings(stdout.Bytes())
}

// scanWithRules scans the path individually for each rule in the list.
func (s *Scanner) scanWithRules(ctx context.Context, rules []Rule, path string) ([]Finding, error) {
	timeout := s.cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	binary := s.cfg.SGBinary
	if binary == "" {
		binary = "sg"
	}

	allFindings := make([]Finding, 0)

	for _, rule := range rules {
		if rule.Pattern == "" || rule.Language == "" {
			continue
		}

		// F3: context leak prevented by defer cancel() inside runSingleRule.
		findings, err := s.runSingleRule(ctx, binary, rule, path, timeout)
		if err != nil {
			slog.Debug("rule execution failed, skipping", "rule", rule.ID, "error", err)
			continue
		}

		// Inject rule metadata (F1: includes Language field).
		for i := range findings {
			if findings[i].RuleID == "" {
				findings[i].RuleID = rule.ID
			}
			if findings[i].Severity == "" {
				findings[i].Severity = rule.Severity
			}
			if findings[i].Message == "" {
				findings[i].Message = rule.Message
			}
			// Language is always injected from the rule (target language of the rule, not the matched file).
			findings[i].Language = rule.Language
			// Propagate Note: copy from Rule.Note when the Finding has none (REQ-UTIL-002-003).
			if findings[i].Note == "" {
				findings[i].Note = rule.Note
			}
			// Propagate Metadata: copy from Rule.Metadata when the Finding has none (REQ-UTIL-002-004).
			if findings[i].Metadata == nil {
				findings[i].Metadata = rule.Metadata
			}
		}

		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

// parseSGFindings converts sg JSON output to a Finding slice.
func parseSGFindings(output []byte) ([]Finding, error) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return []Finding{}, nil
	}

	var matches []sgScanMatch
	if err := json.Unmarshal(trimmed, &matches); err != nil {
		return nil, fmt.Errorf("parse sg output: %w", err)
	}

	findings := make([]Finding, 0, len(matches))
	for _, m := range matches {
		f := Finding{
			RuleID:    m.RuleID,
			Severity:  m.Severity,
			Message:   m.Message,
			File:      m.File,
			Line:      m.Range.Start.Line + 1, // sg uses 0-indexed lines; Finding uses 1-indexed
			Column:    m.Range.Start.Column,
			EndLine:   m.Range.End.Line + 1,
			EndColumn: m.Range.End.Column,
			Note:      m.Note,
		}
		findings = append(findings, f)
	}

	return findings, nil
}
