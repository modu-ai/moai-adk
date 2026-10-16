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
	// No code-level default: where a project keeps its rules is project
	// configuration (gate.yaml ast_grep_gate.rules_dir), resolved by callers
	// (CLI commands, gate loaders). Empty means "no rules configured" — the
	// scanner reports 0 findings for a nonexistent directory
	// (REQ-ASTG-UPG-012).
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
// RulesDir is deliberately empty: a hardcoded path cannot serve both the
// template deployment path (.moai/config/astgrep-rules, where distributed
// users' rules live) and per-project overrides (this repository tracks its
// ruleset at .moai/astgrep-rules), so rules-dir resolution belongs to the
// config layer (gate.yaml ast_grep_gate.rules_dir), not the scanner package.
func DefaultScannerConfig() *ScannerConfig {
	return &ScannerConfig{
		RulesDir: "",
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

// ErrScannerUnavailable identifies "the sg scanner binary could not be
// resolved" as a condition distinct from both a clean scan and a scan failure.
// Scan returns it wrapped, so the message can name the binary it sought and
// carry install guidance while errors.Is still matches through the wrap.
//
// It exists because reporting an absent scanner as an empty successful scan is
// a false all-clear: a CI pipeline gating on `moai ast-grep` passed green over
// code that was never scanned.
var ErrScannerUnavailable = errors.New("astgrep: sg scanner binary unavailable")

// ErrRulesetLoad identifies "sg ran but refused to load the ruleset". It is the
// condition behind a non-zero sg exit with no JSON on stdout and a diagnostic on
// stderr. It is distinct from a clean scan (no findings) and from
// ErrScannerUnavailable (no sg binary at all).
var ErrRulesetLoad = errors.New("astgrep: ruleset failed to load")

// RulesetLoadError reports that sg refused to load a ruleset, so no rule ran.
// Detail is the first non-blank line of sg's stderr, which names the failing
// rule file (for example "Error: Cannot parse rule <path>"). errors.Is matches
// ErrRulesetLoad through Unwrap, and errors.As recovers Detail.
type RulesetLoadError struct {
	Detail string
}

// Error renders the load failure with sg's first stderr line.
func (e *RulesetLoadError) Error() string {
	return "ast-grep ruleset failed to load: " + e.Detail
}

// Unwrap lets errors.Is match the ErrRulesetLoad sentinel.
func (e *RulesetLoadError) Unwrap() error { return ErrRulesetLoad }

// firstNonBlankLine returns the first line of s that has non-space content,
// trimmed of surrounding whitespace, or "" when s has none.
func firstNonBlankLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// ScanExecError reports that sg exited non-zero with no findings on stdout for a
// reason other than a ruleset load failure: a usage error, a crash, or a killed
// child. It fails loud and is never a clean scan. Detail is sg's first non-blank
// stderr line and may be empty; Err is the exit error from os/exec.
type ScanExecError struct {
	Detail string
	Err    error
}

// Error renders the failure with sg's first stderr line when there is one.
func (e *ScanExecError) Error() string {
	if e.Detail != "" {
		return "ast-grep scan failed: " + e.Detail
	}
	if e.Err != nil {
		return "ast-grep scan failed: " + e.Err.Error()
	}
	return "ast-grep scan failed"
}

// Unwrap exposes the exit error from os/exec.
func (e *ScanExecError) Unwrap() error { return e.Err }

// rulesetLoadDiagnostics are the first stderr lines sg prints when it cannot load
// a ruleset (measured on sg 0.40.5). Any other diagnostic is an execution error,
// so an unrecognized wording still fails loud.
var rulesetLoadDiagnostics = []string{
	"Error: Cannot parse rule",
	"Error: Cannot read rule directory",
	"Error: Cannot parse configuration",
	"Error: Cannot read configuration",
}

// isRulesetLoadDiagnostic reports whether detail is one of sg's ruleset-load
// diagnostics.
func isRulesetLoadDiagnostic(detail string) bool {
	for _, prefix := range rulesetLoadDiagnostics {
		if strings.HasPrefix(detail, prefix) {
			return true
		}
	}
	return false
}

// InstallURL is the canonical ast-grep install page. The scanner's unavailable
// error, the `moai ast-grep` CLI guidance, the commit quality gate's skip
// reason, and the `moai doctor` check all cite it from here so the four cannot
// drift apart.
const InstallURL = "https://ast-grep.github.io/guide/quick-start.html"

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
// @MX:ANCHOR: [AUTO] ValidateBinary is a security-critical validation used across 3+ packages
// @MX:REASON: [AUTO] fan_in >= 3 — Scanner, CLI, quality gate, PostToolUse hook all call this
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
// @MX:REASON: [AUTO] fan_in >= 3: quality gate hook, PostToolUse hook, and CLI subcommand all call this method
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
// Returns an error wrapping ErrScannerUnavailable when the sg CLI cannot be
// resolved — that condition is NOT reported as an empty successful scan,
// because callers cannot tell the two apart.
// Returns a *RulesetLoadError (matching ErrRulesetLoad) when sg refuses to load the
// ruleset: a load failure is never reported as an empty, clean scan.
// Returns a *ScanExecError when sg exits non-zero with no findings for any other
// reason, and an error when the scan target does not exist. Neither is a clean scan.
// Returns ([]Finding{}, nil) when the rules directory is empty or does not exist (REQ-ASTG-UPG-012).
// Returns an error when SGBinary is an untrusted path (F2 security check).
func (s *Scanner) Scan(ctx context.Context, path string) ([]Finding, error) {
	// Binary path security validation (F2): untrusted paths return an error immediately.
	// This runs before the availability probe, so a rejected path yields a
	// distinct, non-sentinel error.
	if err := ValidateBinary(s.cfg.SGBinary); err != nil {
		return []Finding{}, fmt.Errorf("sg binary validation failed (SGBinary=%q): %w", s.cfg.SGBinary, err)
	}

	// The sg CLI is not resolvable, so no rule ran. Report it as the sentinel
	// rather than as a clean result: callers gate on this.
	if !s.isSGAvailable() {
		binary := s.cfg.SGBinary
		if binary == "" {
			binary = "sg"
		}
		slog.Warn("ast-grep (sg) CLI not found; skipping scan",
			"binary", binary,
			"hint", "install from "+InstallURL)
		return []Finding{}, fmt.Errorf("%q not found in PATH; install from %s: %w", binary, InstallURL, ErrScannerUnavailable)
	}

	// The target must exist before sg is asked to scan it. Behind `--`, a missing
	// path is an empty scan that exits 0 (measured on sg 0.40.5), so this check is
	// what keeps a missing target loud.
	if _, err := os.Stat(path); err != nil {
		return []Finding{}, fmt.Errorf("ast-grep scan target: %w", err)
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

	// `--` keeps a target whose name starts with "-" from being read as a flag.
	cmd := exec.CommandContext(scanCtx, binary, "scan", "--config", configPath, "--json", "--", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	// F4: debug-log any stderr content (language is unknown in config-based scans).
	if stderr.Len() > 0 {
		slog.Debug("sg scan stderr (config-based)", "config", configPath, "stderr", stderr.String())
	}

	// sg exits non-zero in three situations that must not be conflated. With
	// error-severity matches it exits non-zero WITH the JSON on stdout: a normal
	// scan, parsed below. With a ruleset it cannot load it exits non-zero with
	// NOTHING on stdout and a load diagnostic on stderr: RulesetLoadError. Any
	// other non-zero exit with nothing on stdout (a usage error, a crash, a killed
	// child) is a ScanExecError. Neither is ever a clean scan (GH #1824).
	if runErr != nil && stdout.Len() == 0 {
		detail := firstNonBlankLine(stderr.String())
		if isRulesetLoadDiagnostic(detail) {
			return nil, &RulesetLoadError{Detail: detail}
		}
		return nil, &ScanExecError{Detail: detail, Err: runErr}
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
