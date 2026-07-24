package quality

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// RunAstGrepGateV2 runs the ast-grep quality gate using the unified Scanner.
// REQ-ASTG-UPG-030: quality gate hook calls the unified Scanner
// REQ-UTIL-002-012: returns (false, formatted-list) when suppression policy violations are found.
// Sole gate entry point: the legacy RunAstGrepGate (V1) path was removed once
// the config loader made this path reachable.
func RunAstGrepGateV2(ctx context.Context, projectDir string, cfg *AstGrepGateConfig) (bool, string) {
	if cfg == nil || !cfg.Enabled {
		return true, ""
	}

	// ── 1. Suppression policy check (sg-independent, pure-Go) ─────────────────
	// REQ-UTIL-002-010/011/012: verify @MX:REASON pairing for ast-grep-ignore comments
	sourceFiles := walkSourceFiles(projectDir)
	var allViolations []SuppressionViolation
	for _, fp := range sourceFiles {
		allViolations = append(allViolations, checkSuppressionPairing(fp)...)
	}

	if len(allViolations) > 0 {
		var sb strings.Builder
		sb.WriteString("ast-grep suppression policy violations:\n\n")
		for _, v := range allViolations {
			fmt.Fprintf(&sb, "[%s] %s\n", v.Type, v.Message)
		}
		return false, strings.TrimSpace(sb.String())
	}

	// ── 2. ast-grep scan (depends on sg CLI) ─────────────────────────────────
	rulesDir := filepath.Join(projectDir, cfg.RulesDir)
	scannerCfg := &astgrep.ScannerConfig{
		RulesDir:     rulesDir,
		SGBinary:     "sg",
		WarnOnlyMode: cfg.WarnOnlyMode,
		Timeout:      astGrepScanTimeout,
	}

	scanner := astgrep.NewScanner(scannerCfg)
	findings, err := scanner.Scan(ctx, projectDir)
	if err != nil {
		// Pass on scan error (graceful degradation)
		return true, ""
	}

	if len(findings) == 0 {
		return true, ""
	}

	// Format results
	var sb strings.Builder
	sb.WriteString("ast-grep domain rule scan results:\n\n")
	for _, f := range findings {
		sb.WriteString(f.String())
		sb.WriteString("\n")
	}
	output := strings.TrimSpace(sb.String())

	// Block when error-severity findings are found (unless WarnOnlyMode is enabled)
	if astgrep.HasErrors(findings) && !cfg.WarnOnlyMode && cfg.BlockOnError {
		return false, fmt.Sprintf("quality gate failed: ast-grep domain rules\n\n%s", output)
	}

	return true, output
}

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

const astGrepScanTimeout = 30 * time.Second

// SuppressionViolation represents a case where an ast-grep-ignore comment is not paired with @MX:REASON.
// REQ-UTIL-002-010/011
type SuppressionViolation struct {
	// Type is always "SUPPRESSION_WITHOUT_REASON".
	Type string `json:"type"`
	// File is the path of the file where the violation occurred.
	File string `json:"file"`
	// Line is the 1-indexed line number.
	Line int `json:"line"`
	// Message is a human-readable error message.
	Message string `json:"message"`
}

// commentPrefix returns the line comment prefix for the given file extension.
// Returns an empty string for unsupported extensions.
func commentPrefix(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".rb", ".py", ".pyi", ".ex", ".exs":
		return "#"
	case ".hs":
		return "--"
	case ".go", ".java", ".ts", ".tsx", ".js", ".jsx", ".rs", ".swift",
		".kt", ".kts", ".cs", ".scala", ".cpp", ".cc", ".cxx", ".hpp",
		".c", ".h", ".dart", ".zig":
		return "//"
	default:
		return ""
	}
}

// checkSuppressionPairing inspects a file line by line to find cases where
// an ast-grep-ignore comment is not paired with an adjacent @MX:REASON comment.
//
// Allowed distance: @MX:REASON may appear after 0–1 blank lines following ast-grep-ignore (REQ-UTIL-002-010).
// Not allowed: any non-empty comment or code line in between (REQ-UTIL-002-011).
//
// REQ-UTIL-002-010/011
func checkSuppressionPairing(filePath string) []SuppressionViolation {
	prefix := commentPrefix(filePath)
	if prefix == "" {
		return nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	ignoreMarker := prefix + " ast-grep-ignore"
	reasonPrefix := prefix + " @MX:REASON"

	// Read file line by line
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	var violations []SuppressionViolation
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != ignoreMarker {
			continue
		}

		// ast-grep-ignore found at line i (1-indexed: i+1)
		found := false
		blankCount := 0

		for j := i + 1; j < len(lines); j++ {
			nextTrimmed := strings.TrimSpace(lines[j])

			if nextTrimmed == "" {
				blankCount++
				if blankCount > 1 {
					// More than 1 blank line: too far
					break
				}
				continue
			}

			// Reached the first non-empty line
			// OK if @MX:REASON is present and followed by non-empty text
			if strings.HasPrefix(nextTrimmed, reasonPrefix) {
				rest := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, reasonPrefix))
				if rest != "" {
					found = true
				}
			}
			break
		}

		if !found {
			violations = append(violations, SuppressionViolation{
				Type: "SUPPRESSION_WITHOUT_REASON",
				File: filePath,
				Line: i + 1,
				Message: fmt.Sprintf(
					"ast-grep suppression at %s:%d requires adjacent '%s @MX:REASON <rationale>' on next line",
					filePath, i+1, prefix,
				),
			})
		}
	}

	return violations
}

// walkSourceFiles recursively walks source files under projectDir.
// Returns only files for which commentPrefix is non-empty.
// Suppression policy checks apply to production code, so *_test.go and common exclusion paths are skipped.
func walkSourceFiles(projectDir string) []string {
	var files []string
	_ = filepath.WalkDir(projectDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		// Exclude vendor, .git, node_modules, etc.
		parts := strings.Split(filepath.ToSlash(path), "/")
		for _, part := range parts {
			if part == "vendor" || part == ".git" || part == "node_modules" || part == "__pycache__" {
				return nil
			}
		}
		// Exclude *_test.go files: test files often use ast-grep-ignore as test fixture data
		// and should be excluded from suppression policy checks.
		if strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		if commentPrefix(path) != "" {
			files = append(files, path)
		}
		return nil
	})
	return files
}
