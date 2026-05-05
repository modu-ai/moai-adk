package cli

import (
"fmt"
"os"
"path/filepath"
"strings"
"text/tabwriter"

"github.com/spf13/cobra"

"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecLintCmd 'moai spec lint' subcommand creates.
// SPEC-V3R2-SPC-003 implementation.
//
// @MX:NOTE: [AUTO] newSpecLintCmd is the spec lint CLI entry point following cobra pattern.
func newSpecLintCmd() *cobra.Command {
var (
jsonOutput bool
sarifOutput bool
strict bool
format string
)

cmd := &cobra.Command{
Use: "lint [spec.md...]",
Short: "Lint SPEC documents for EARS compliance and structural validity",
Long: `Validate SPEC documents against:
- EARS modality compliance (SHALL, WHEN, WHILE, WHERE, IF)
- REQ ID uniqueness
- AC→REQ coverage (100% required)
- Frontmatter schema validation
- Dependency DAG (no cycles, all deps exist)
- Out of Scope section presence
- Zone registry cross-references

Exit codes:
0 = success (no errors)
1 = errors found
2 = linter crash
3 = invalid arguments`,
Args: cobra.ArbitraryArgs,
RunE: func(cmd *cobra.Command, args []string) error {
// validate arguments
if jsonOutput && sarifOutput {
return fmt.Errorf("cannot use --json and --sarif together")
}

// Determine BaseDir: prioritize .moai/specs/ in current working directory
cwd, err := os.Getwd()
if err != nil {
return fmt.Errorf("working directory verification error: %w", err)
}

baseDir := detectBaseDir(cwd)
registryPath := detectRegistryPath(cwd)

linterOpts := spec.LinterOptions{
RegistryPath: registryPath,
BaseDir: baseDir,
Strict: strict,
}

linter := spec.NewLinter(linterOpts)
report, lintErr := linter.Lint(args)
if lintErr != nil {
_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "linter error: %v\n", lintErr)
os.Exit(2)
}

// select output format
switch {
case jsonOutput:
data, marshalErr := report.ToJSON()
if marshalErr != nil {
return fmt.Errorf("JSON serialization error: %w", marshalErr)
}
_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))

case sarifOutput:
data, marshalErr := report.ToSARIF()
if marshalErr != nil {
return fmt.Errorf("SARIF serialization error: %w", marshalErr)
}
_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))

default:
// human-readable table (default or --format table)
_ = format // format=table is same as default value
printTable(cmd, report)
}

if report.HasErrors() {
os.Exit(1)
}

return nil
},
}

cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
cmd.Flags().BoolVar(&sarifOutput, "sarif", false, "output in SARIF 2.1.0 format")
cmd.Flags().BoolVar(&strict, "strict", false, "treat warnings as errors")
cmd.Flags().StringVar(&format, "format", "table", "output format (table)")

return cmd
}

// printTable outputs findings in human-readable table format.
func printTable(cmd *cobra.Command, report *spec.Report) {
out := cmd.OutOrStdout()
if len(report.Findings) == 0 {
_, _ = fmt.Fprintln(out, "✓ No findings — all SPEC documents are valid")
return
}

w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
_, _ = fmt.Fprintln(w, "SEVERITY\tCODE\tFILE\tLINE\tMESSAGE")
_, _ = fmt.Fprintln(w, "--------\t----\t----\t----\t-------")

for _, f := range report.Findings {
_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
strings.ToUpper(string(f.Severity)),
f.Code,
f.File,
f.Line,
f.Message,
)
}
_ = w.Flush()

// output summary
var errCount, warnCount int
for _, f := range report.Findings {
switch f.Severity {
case spec.SeverityError:
errCount++
case spec.SeverityWarning:
warnCount++
}
}
_, _ = fmt.Fprintf(out, "\n%d error(s), %d warning(s)\n", errCount, warnCount)
}

// detectBaseDir determine project base directory.
// if .moai/specs/ directory exists, use that as base.
func detectBaseDir(cwd string) string {
specsDir := filepath.Join(cwd, ".moai", "specs")
if _, err := os.Stat(specsDir); err == nil {
return specsDir
}
return cwd
}

// detectRegistryPath detect zone registry file path.
func detectRegistryPath(cwd string) string {
candidate := filepath.Join(cwd, ".claude", "rules", "moai", "core", "zone-registry.md")
if _, err := os.Stat(candidate); err == nil {
return candidate
}
return ""
}
