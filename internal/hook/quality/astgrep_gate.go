package quality

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// AstGrepGateConfig holds configuration for ast-grep quality gate scanning.
type AstGrepGateConfig struct {
	// Enabled controls whether ast-grep scanning is performed.
	Enabled bool
	// RulesDir is the directory containing domain-specific ast-grep rule files.
	// Default: ".moai/config/astgrep-rules"
	RulesDir string
	// BlockOnError causes the gate to block a commit when error-severity matches are found.
	BlockOnError bool
	// WarnOnlyMode prevents blocking even when error-severity matches are found.
	// When true, matches are reported as warnings but never block.
	WarnOnlyMode bool
}

// DefaultAstGrepGateConfig returns default configuration.
func DefaultAstGrepGateConfig() *AstGrepGateConfig {
	return &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: true,
		WarnOnlyMode: false,
	}
}

// astGrepScanMatch represents a single match entry from sg scan --json output.
type astGrepScanMatch struct {
	File     string `json:"file"`
	Lines    string `json:"lines,omitempty"`
	Text     string `json:"text,omitempty"`
	RuleID   string `json:"ruleId,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
	Range    struct {
		Start struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"start"`
	} `json:"range"`
}

const astGrepScanTimeout = 30 * time.Second

// RunAstGrepGate scans the project using domain-specific ast-grep rules.
// Returns (passed, output) where passed=false blocks the commit.
// When sg CLI is not available, returns (true, "") gracefully.
// When the rules directory does not exist, returns (true, "") gracefully.
func RunAstGrepGate(ctx context.Context, projectDir string, cfg *AstGrepGateConfig) (bool, string) {
	if cfg == nil || !cfg.Enabled {
		return true, ""
	}

	// Check if sg CLI is available — pass silently if not found
	if _, err := exec.LookPath("sg"); err != nil {
		return true, ""
	}

	rulesDir := filepath.Join(projectDir, cfg.RulesDir)
	if _, err := os.Stat(rulesDir); err != nil {
		// Pass silently when the rules directory does not exist
		return true, ""
	}

	// Load rule files
	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDirectory(rulesDir)
	if err != nil || len(rules) == 0 {
		return true, ""
	}

	// Apply 30-second timeout for the full scan
	scanCtx, cancel := context.WithTimeout(ctx, astGrepScanTimeout)
	defer cancel()

	// Use config-based scan if sgconfig.yml exists, otherwise scan per-rule pattern
	var allMatches []astGrepScanMatch
	sgconfigPath := filepath.Join(rulesDir, "sgconfig.yml")
	if _, err := os.Stat(sgconfigPath); err == nil {
		matches, scanErr := runSGConfig(scanCtx, sgconfigPath, projectDir)
		if scanErr == nil {
			allMatches = matches
		}
	} else {
		for _, rule := range rules {
			if rule.Pattern == "" || rule.Language == "" {
				continue
			}
			matches, scanErr := runSGRule(scanCtx, rule, projectDir)
			if scanErr != nil {
				continue
			}
			allMatches = append(allMatches, matches...)
		}
	}

	if len(allMatches) == 0 {
		return true, ""
	}

	// Format results and classify by severity
	var sb strings.Builder
	sb.WriteString("ast-grep domain rule scan results:\n\n")
	hasError := false
	for _, m := range allMatches {
		sev := strings.ToLower(m.Severity)
		if sev == "" {
			sev = "warning"
		}
		line := m.Range.Start.Line + 1 // 0-indexed → 1-indexed (convert to human-readable line number)
		ruleID := m.RuleID
		if ruleID == "" {
			ruleID = "unknown"
		}
		msg := m.Message
		if msg == "" {
			msg = m.Lines
		}
		fmt.Fprintf(&sb, "%s:%d: [%s] %s (%s)\n", m.File, line, ruleID, msg, sev)
		if sev == "error" {
			hasError = true
		}
	}

	output := strings.TrimSpace(sb.String())

	// Block when error-severity matches exist, WarnOnlyMode is off, and BlockOnError is enabled
	if hasError && !cfg.WarnOnlyMode && cfg.BlockOnError {
		return false, fmt.Sprintf("quality gate failed: ast-grep domain rules\n\n%s", output)
	}

	return true, output
}

// runSGConfig runs sg scan using an sgconfig.yml file.
func runSGConfig(ctx context.Context, configPath, projectDir string) ([]astGrepScanMatch, error) {
	cmd := exec.CommandContext(ctx, "sg", "scan", "--config", configPath, "--json", projectDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ignore the error: sg may return a non-zero exit code when matches are found
	_ = cmd.Run()

	return parseSGScanOutput(stdout.Bytes())
}

// runSGRule runs sg for a single rule using its pattern and language.
func runSGRule(ctx context.Context, rule astgrep.Rule, projectDir string) ([]astGrepScanMatch, error) {
	cmd := exec.CommandContext(ctx, "sg", "run",
		"--pattern", rule.Pattern,
		"--lang", rule.Language,
		"--json",
		projectDir,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ignore the error: sg may return a non-zero exit code when matches are found
	_ = cmd.Run()

	matches, err := parseSGScanOutput(stdout.Bytes())
	if err != nil {
		return nil, err
	}

	// Inject rule metadata into matches
	for i := range matches {
		if matches[i].RuleID == "" {
			matches[i].RuleID = rule.ID
		}
		if matches[i].Severity == "" {
			matches[i].Severity = rule.Severity
		}
		if matches[i].Message == "" {
			matches[i].Message = rule.Message
		}
	}

	return matches, nil
}

// parseSGScanOutput parses the JSON array output from sg scan/run --json.
func parseSGScanOutput(output []byte) ([]astGrepScanMatch, error) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return nil, nil
	}

	var matches []astGrepScanMatch
	if err := json.Unmarshal(trimmed, &matches); err != nil {
		return nil, fmt.Errorf("parse sg output: %w", err)
	}
	return matches, nil
}
