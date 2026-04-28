package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// astGrepFlags holds flag values for the ast-grep subcommand.
type astGrepFlags struct {
	format   string
	lang     string
	severity string
	dry      bool
	rulesDir string
}

// NewAstGrepCmd creates and returns the `moai ast-grep` Cobra command.
// REQ-ASTG-UPG-020, REQ-ASTG-UPG-021
func NewAstGrepCmd() *cobra.Command {
	flags := &astGrepFlags{}

	cmd := &cobra.Command{
		Use:   "ast-grep [path]",
		Short: "Scan code using ast-grep",
		Long: `Applies code quality and security rules to the specified path using the ast-grep (sg) CLI.

Supported output formats:
  text   - Human-readable text (default)
  json   - Machine-readable JSON array
  sarif  - SARIF 2.1.0 format (for uploading to GitHub code scanning)

Examples:
  moai ast-grep ./
  moai ast-grep --format=sarif --lang=go ./internal/
  moai ast-grep --severity=error ./
  moai ast-grep --dry ./`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return runAstGrep(cmd, flags, path)
		},
	}

	// Register flags (REQ-ASTG-UPG-021)
	cmd.Flags().StringVar(&flags.format, "format", "text", "Output format: text, json, sarif")
	cmd.Flags().StringVar(&flags.lang, "lang", "", "Scan only the specified language (e.g. go, python, typescript)")
	cmd.Flags().StringVar(&flags.severity, "severity", "", "Minimum severity to display (error, warning, info)")
	cmd.Flags().BoolVar(&flags.dry, "dry", false, "Print only the list of rules that would be applied without running the actual scan")
	cmd.Flags().StringVar(&flags.rulesDir, "rules-dir", ".moai/config/astgrep-rules", "ast-grep rules directory path")

	return cmd
}

// runAstGrep runs the ast-grep scan and outputs the results.
func runAstGrep(cmd *cobra.Command, flags *astGrepFlags, path string) error {
	cfg := &astgrep.ScannerConfig{
		RulesDir:     flags.rulesDir,
		SGBinary:     "sg",
		WarnOnlyMode: false,
	}

	// --dry: print only the list of rules
	if flags.dry {
		return runDryMode(cmd, cfg, flags)
	}

	scanner := astgrep.NewScanner(cfg)
	ctx := cmd.Context()
	if ctx == nil {
		ctx = cmd.Root().Context()
	}

	findings, err := scanner.Scan(ctx, path)
	if err != nil {
		return fmt.Errorf("ast-grep scan: %w", err)
	}

	// Apply --lang filter.
	if flags.lang != "" {
		findings = filterByLang(findings, flags.lang)
	}

	// Apply --severity filter.
	if flags.severity != "" {
		findings = filterBySeverity(findings, flags.severity)
	}

	// Output results in the selected format.
	switch strings.ToLower(flags.format) {
	case "json":
		return outputJSON(cmd, findings)
	case "sarif":
		return outputSARIF(cmd, findings)
	default: // "text"
		outputText(cmd, findings)
	}

	// exit code 1 when error-severity findings are found (AC4)
	if astgrep.HasErrors(findings) {
		os.Exit(1)
	}

	return nil
}

// runDryMode prints only the list of rules when the --dry flag is set.
func runDryMode(cmd *cobra.Command, cfg *astgrep.ScannerConfig, flags *astGrepFlags) error {
	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(cfg.RulesDir)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "cannot read rules directory: %v\n", err)
		return nil
	}

	if len(rules) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no rules to apply")
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rules to apply (%d):\n", len(rules))
	for _, r := range rules {
		lang := r.Language
		if lang == "" {
			lang = "all"
		}
		if flags.lang != "" && !strings.EqualFold(lang, flags.lang) && lang != "all" {
			continue
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s - %s (%s)\n", r.Severity, r.ID, r.Message, lang)
	}

	return nil
}

// outputText prints findings in text format.
func outputText(cmd *cobra.Command, findings []astgrep.Finding) {
	out := cmd.OutOrStdout()
	if len(findings) == 0 {
		_, _ = fmt.Fprintln(out, "no findings")
		return
	}

	_, _ = fmt.Fprintf(out, "findings (%d):\n\n", len(findings))
	for _, f := range findings {
		_, _ = fmt.Fprintln(out, f.String())
		if f.Note != "" {
			_, _ = fmt.Fprintf(out, "  note: %s\n", f.Note)
		}
	}
}

// outputJSON prints findings as a JSON array.
func outputJSON(cmd *cobra.Command, findings []astgrep.Finding) error {
	if findings == nil {
		findings = []astgrep.Finding{}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		return fmt.Errorf("JSON encoding: %w", err)
	}

	return nil
}

// outputSARIF prints findings in SARIF 2.1.0 format.
func outputSARIF(cmd *cobra.Command, findings []astgrep.Finding) error {
	// Detect the sg version (falls back to "unknown" on failure).
	sgVersion := detectSGVersion()

	output, err := astgrep.ToSARIF(findings, sgVersion)
	if err != nil {
		return fmt.Errorf("SARIF generation: %w", err)
	}

	_, err = cmd.OutOrStdout().Write(output)
	return err
}

// sgVersionOnce and sgVersionResult are package-level variables for sync.Once caching in detectSGVersion.
// Using a pointer allows tests to swap in a new instance (sync.Once cannot be copied).
// REQ-UTIL-002-009: sg --version is executed at most once within the same process.
var (
	sgVersionOnce   = new(sync.Once)
	sgVersionResult string
)

// sgVersionExec is the function responsible for running sg --version.
// It is injectable and can be replaced in tests, enabling unit tests without the sg binary.
// REQ-UTIL-002-008: runs sg --version within a 5-second timeout and returns the trimmed stdout.
var sgVersionExec = func(sgBinary string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, sgBinary, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// detectSGVersion parses the output of sg --version and returns the version string.
// Results are cached via sync.Once, so sg --version runs at most once per process.
// Returns "unknown" on error (binary not found, timeout, abnormal exit, or empty output).
// REQ-UTIL-002-008, REQ-UTIL-002-009
func detectSGVersion() string {
	sgVersionOnce.Do(func() {
		v, err := sgVersionExec("sg")
		if err != nil {
			sgVersionResult = "unknown"
			return
		}
		v = strings.TrimSpace(v)
		if v == "" {
			sgVersionResult = "unknown"
			return
		}
		sgVersionResult = v
	})
	return sgVersionResult
}

// NOTE: sgVersionOnce is a *sync.Once (pointer type).
// Tests can achieve the equivalent of resetting sync.Once by replacing the pointer with a new instance.

// filterByLang returns only findings produced by rules targeting the specified language.
// Returns all findings when lang is empty.
// Findings with an empty Language are treated as language-neutral rules and always included.
// Comparison is case-insensitive.
func filterByLang(findings []astgrep.Finding, lang string) []astgrep.Finding {
	if lang == "" {
		return findings
	}
	target := strings.ToLower(lang)
	out := make([]astgrep.Finding, 0, len(findings))
	for _, f := range findings {
		fl := strings.ToLower(f.Language)
		// Include findings with no language info (language-neutral rules).
		if fl == "" || fl == target {
			out = append(out, f)
		}
	}
	return out
}

// filterBySeverity returns only findings at or above the specified severity.
func filterBySeverity(findings []astgrep.Finding, minSeverity string) []astgrep.Finding {
	var filtered []astgrep.Finding
	for _, f := range findings {
		switch strings.ToLower(minSeverity) {
		case "error":
			if f.IsError() {
				filtered = append(filtered, f)
			}
		case "warning":
			if f.IsError() || f.IsWarning() {
				filtered = append(filtered, f)
			}
		default: // info: include all
			filtered = append(filtered, f)
		}
	}
	return filtered
}
