package quality

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// Pass-path reasons for the ast-grep gate step. Both accompany a pass: a
// missing or failing optional scanner must never turn into a blocked commit.
//
// They are package-level named constants rather than inline literals so their
// distinctness is assertable directly. Only the unavailable branch is
// reachable end-to-end through the gate — RunAstGrepGateV2 constructs its
// scanner with a fixed binary name and offers no injection seam — so a test
// that could only reach one branch would never catch the two collapsing into
// one string.
const (
	// astGrepReasonScannerUnavailable reports that the sg CLI could not be
	// resolved, so no rule ran at all. Distinct from "ran and found nothing".
	astGrepReasonScannerUnavailable = "ast-grep scan skipped: the sg CLI was not found, so no rules ran " +
		"(install from " + astgrep.InstallURL + ")"

	// astGrepReasonScanDegraded reports that the scanner was present but the
	// scan itself failed, so its (empty) result is not a clean bill of health.
	astGrepReasonScanDegraded = "ast-grep scan degraded: the scanner returned an error, so its findings are incomplete"
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
		// Graceful degradation: an optional scanner never blocks a commit.
		// The pass now carries a reason, because passing with an empty string
		// is exactly what a clean scan returns — which left an absent scanner
		// indistinguishable from scanned-and-clean.
		if errors.Is(err, astgrep.ErrScannerUnavailable) {
			return true, astGrepReasonScannerUnavailable
		}
		return true, astGrepReasonScanDegraded
	}

	if len(findings) == 0 {
		return true, ""
	}

	// SPEC-GATE-ASTGREP-REPAIR-001 M2 (REQ-GAR-004/005): exclude findings whose
	// file paths fall under roots ast-grep should not surface. The primary
	// mechanism (sgconfig.yml `globs:`) is NOT supported by ast-grep 0.40.5 in
	// config-mode (empirically verified — the field silently breaks the scan).
	// .gitignore already excludes .claude/worktrees/ for the default scan, but
	// test files, testdata, and vendor are not gitignored, so this filter is the
	// authoritative exclusion boundary. Applied at the gate layer (NOT in
	// scanner.go, whose Scan body is preserved by REQ-GAR-010) so both the
	// PreToolUse path and the standalone `moai gate` CLI path share one filter.
	findings = filterExcludedPaths(findings)

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
	// No code-level default: gate.yaml ast_grep_gate.rules_dir is the SSOT
	// (the template ships the key with an explicit value). Empty means
	// "unconfigured" — the scanner reports 0 findings for a nonexistent
	// directory.
	RulesDir string
	// BlockOnError causes the gate to block a commit when error-severity matches are found.
	BlockOnError bool
	// WarnOnlyMode prevents blocking even when error-severity matches are found.
	// When true, matches are reported as warnings but never block.
	WarnOnlyMode bool
}

// DefaultAstGrepGateConfig returns default configuration. RulesDir is empty
// (t50): this default is the last resort when no config can be loaded at all,
// and guessing a rules path the project never declared would silently point
// the advisory gate at nothing (or at the wrong ruleset) — "unconfigured" is
// the honest value and behaves identically for a project with no rules.
func DefaultAstGrepGateConfig() *AstGrepGateConfig {
	return &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     "",
		BlockOnError: true,
		WarnOnlyMode: false,
	}
}

const astGrepScanTimeout = 30 * time.Second

// astGrepExcludedPathPatterns lists path substrings that mark a finding's file
// as out-of-scope for the ast-grep gate. SPEC-GATE-ASTGREP-REPAIR-001 M2
// (REQ-GAR-004/005): worktree duplicates, vendored code, test sources, and
// test fixtures do not represent project quality regressions and are filtered
// out at the gate layer before reporting/blocking.
//
// Substring match on the FORWARD-SLASH path keeps the filter cross-platform
// and resolution-free: ast-grep emits forward-slash paths in its JSON output
// regardless of OS, so no filepath.ToSlash conversion is needed at the call
// site. The `_test.go` entry is suffix-shaped (no leading slash) so it catches
// both top-level and nested test files.
var astGrepExcludedPathPatterns = []string{
	".claude/worktrees/",
	"/vendor/",
	"/testdata/",
	"_test.go",
}

// @MX:NOTE: filterExcludedPaths is the authoritative path-exclusion boundary
// for ast-grep scan findings, applied at the gate layer (NOT in scanner.go,
// whose Scan body is preserved by REQ-GAR-010). The filter exists because
// ast-grep 0.40.5 has no viable globs exclusion in sgconfig.yml config-mode,
// so the exclusion must live downstream of the scan. Shared by both the
// PreToolUse path and the standalone `moai gate` CLI — both call
// RunAstGrepGateV2, which calls this filter (SPEC-GATE-ASTGREP-REPAIR-001 M2 / D2).
//
// filterExcludedPaths returns findings whose file paths do NOT fall under any
// excluded root. It is case-sensitive and uses substring containment on the
// finding's File field as-returned by the scanner.
func filterExcludedPaths(findings []astgrep.Finding) []astgrep.Finding {
	if len(findings) == 0 {
		return findings
	}
	out := make([]astgrep.Finding, 0, len(findings))
	for _, f := range findings {
		if isExcludedPath(f.File) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// isExcludedPath reports whether the given path matches any excluded pattern.
func isExcludedPath(path string) bool {
	for _, p := range astGrepExcludedPathPatterns {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

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
