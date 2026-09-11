package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/specid"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecLintCmd 'moai spec lint' subcommand creates.
// SPEC-V3R2-SPC-003 implementation.
//
// @MX:NOTE: [AUTO] newSpecLintCmd is the spec lint CLI entry point following cobra pattern.
func newSpecLintCmd() *cobra.Command {
	var (
		jsonOutput     bool
		sarifOutput    bool
		strict         bool
		format         string
		baselinePath   string
		updateBaseline bool
		reason         string
	)

	cmd := &cobra.Command{
		Use:   "lint [SPEC-ID | path/to/spec.md | SPEC directory ...]",
		Short: "Lint SPEC documents for EARS compliance and structural validity",
		Long: `Accepted arguments (SPEC-SPEC-LINT-ID-ARG-001) — three shapes, mixable in
one invocation, and with no argument the whole corpus is scanned as before:
- a SPEC-ID, e.g. SPEC-SPC-001, resolved to .moai/specs/<SPEC-ID>/spec.md
  under the project root (the same rule 'moai spec view' uses)
- a file path, e.g. .moai/specs/SPEC-SPC-001/spec.md
- a SPEC directory, e.g. .moai/specs/SPEC-SPC-001, read as the spec.md inside it

An argument that looks like a SPEC-ID but resolves to nothing is an argument
error (exit code 3) naming the path that was tried — not a finding about a
document.

Validate SPEC documents against:
- EARS modality compliance (SHALL, WHEN, WHILE, WHERE, IF)
- REQ ID uniqueness
- AC→REQ coverage (100% required)
- Frontmatter schema validation
- Dependency DAG (no cycles, all deps exist)
- Out of Scope section presence
- Zone registry cross-references

Baseline ratchet (--baseline):
  Gate on the CHANGE rather than on the standing inventory. Errors always fail.
  A rule whose non-advisory warning count exceeds its recorded count fails and
  is named with its delta; anything at or below its recorded count passes, and
  the warning total stays in the output. A decrease passes and is reported —
  the baseline file is never rewritten implicitly. Rebaselining is the explicit
  --update-baseline --reason "<text>" path, which records the tree SHA, the
  date, and the reason.

Exit codes:
0 = success (no errors)
1 = errors found
2 = linter crash
3 = invalid arguments`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// validate arguments. Invalid argument combination exits 3 per
			// REQ-CONT-001-005 (SPEC-CLIFIX-CONTRACT-001 M2).
			if jsonOutput && sarifOutput {
				return &exitCodeError{code: 3, msg: "cannot use --json and --sarif together"}
			}
			if err := validateBaselineFlags(cmd.ErrOrStderr(), baselinePath, updateBaseline, reason, jsonOutput, sarifOutput, strict); err != nil {
				return err
			}

			// Determine BaseDir: prioritize .moai/specs/ in current working directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory verification error: %w", err)
			}

			baseDir := detectBaseDir(cwd)
			registryPath := detectRegistryPath(cwd)

			// SPEC-SPEC-LINT-ID-ARG-001 — resolve the argument shapes
			// (SPEC-ID / file path / SPEC directory) into lint targets. With
			// no arguments this is a no-op, which is what keeps the whole-
			// corpus scan on detectBaseDir(cwd) exactly as it was.
			targets, resolveErr := resolveLintTargets(cmd.ErrOrStderr(), args)
			if resolveErr != nil {
				return resolveErr
			}

			linterOpts := spec.LinterOptions{
				RegistryPath: registryPath,
				BaseDir:      baseDir,
				Strict:       strict,
			}

			linter := spec.NewLinter(linterOpts)
			report, lintErr := linter.Lint(targets)
			if lintErr != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "linter error: %v\n", lintErr)
				return &exitCodeError{code: 2, msg: fmt.Sprintf("spec lint: linter error: %v", lintErr)}
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

			if baselinePath != "" {
				return runBaselineGate(cmd, report, baselinePath, updateBaseline, reason, cwd)
			}

			if report.HasErrors() {
				return &exitCodeError{code: 1, msg: "spec lint: error-severity findings detected"}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&sarifOutput, "sarif", false, "output in SARIF 2.1.0 format")
	cmd.Flags().BoolVar(&strict, "strict", false, "treat warnings as errors")
	cmd.Flags().StringVar(&format, "format", "table", "output format (table)")
	cmd.Flags().StringVar(&baselinePath, "baseline", "", "gate against a checked-in per-rule baseline file instead of --strict")
	cmd.Flags().BoolVar(&updateBaseline, "update-baseline", false, "recompute and rewrite the --baseline file (requires --reason)")
	cmd.Flags().StringVar(&reason, "reason", "", "mandatory non-empty rationale recorded by --update-baseline")

	return cmd
}

// validateBaselineFlags rejects every incoherent baseline flag combination with
// exit 3. Each rejection is explicit on purpose: a silently-ignored flag, or a
// silent fallback to today's behaviour, is the failure mode the baseline gate
// exists to remove.
//
// Each rejection is WRITTEN to stderr through argumentError, not only carried by
// the returned error: an ExitCoder returned from RunE is rendered as nothing, so
// a rejection that lived only inside the error would exit 3 in silence.
func validateBaselineFlags(stderr io.Writer, baselinePath string, updateBaseline bool, reason string, jsonOutput, sarifOutput, strict bool) error {
	if reason != "" && !updateBaseline {
		return argumentError(stderr, "--reason is only meaningful with --update-baseline")
	}
	if updateBaseline && baselinePath == "" {
		return argumentError(stderr, "--update-baseline requires --baseline <path> naming the file to write")
	}
	if updateBaseline && strings.TrimSpace(reason) == "" {
		return argumentError(stderr, `--update-baseline requires a non-empty --reason "<text>"; an unexplained rebaseline is baseline manipulation`)
	}
	if baselinePath == "" {
		return nil
	}
	// The JSON and SARIF payloads carry the raw finding list and have no
	// baseline dimension, so accepting the flag there would silently ignore it.
	if jsonOutput {
		return argumentError(stderr, "cannot use --baseline with --json (the JSON payload carries no baseline verdict)")
	}
	if sarifOutput {
		return argumentError(stderr, "cannot use --baseline with --sarif (the SARIF payload carries no baseline verdict)")
	}
	// --strict and --baseline are two escalation policies over the same
	// dimension. Running both splits the gate's meaning across two flags.
	if strict {
		return argumentError(stderr, "cannot use --baseline and --strict together: both gate on non-advisory warnings, and --baseline supersedes --strict for that purpose")
	}
	if !updateBaseline {
		if _, statErr := os.Stat(baselinePath); statErr != nil {
			return argumentError(stderr, "--baseline file not found: %s (create it with --update-baseline --reason \"<text>\")", baselinePath)
		}
	}
	return nil
}

// runBaselineGate applies the baseline policy in its load-bearing order:
// errors first (REQ-SLGS-009), then the per-rule non-advisory ratchet.
//
// The ordering is observable rather than merely internal: both causes exit 1,
// so the output names WHICH cause fired. An implementation that consulted the
// baseline first would still exit 1 and would be indistinguishable by exit code
// alone.
//
// @MX:ANCHOR: [AUTO] runBaselineGate is the sole decision point of the spec-lint baseline ratchet.
// @MX:REASON: [AUTO] Error-before-baseline ordering here is what stops the gate becoming a false-green machine.
func runBaselineGate(cmd *cobra.Command, report *spec.Report, path string, update bool, reason, cwd string) error {
	out := cmd.OutOrStdout()

	if update {
		b := spec.NewBaselineFromReport(report, resolveTreeSHA(cwd), time.Now().Format("2006-01-02"), strings.TrimSpace(reason))
		if writeErr := spec.WriteBaseline(path, b); writeErr != nil {
			// Written to stderr for the same reason as argumentError: the
			// exit-2 error is rendered as nothing at the terminal.
			msg := fmt.Sprintf("spec lint: %v", writeErr)
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), msg)
			return &exitCodeError{code: 2, msg: msg}
		}
		recorded := 0
		for _, n := range b.Rules {
			recorded += n
		}
		fmt.Fprintf(out, "\nbaseline: UPDATED — %s\n", path)                                                   //nolint:errcheck // report stream
		fmt.Fprintf(out, "  tree_sha: %s\n", b.TreeSHA)                                                        //nolint:errcheck // report stream
		fmt.Fprintf(out, "  date:     %s\n", b.UpdatedAt)                                                      //nolint:errcheck // report stream
		fmt.Fprintf(out, "  reason:   %s\n", b.Reason)                                                         //nolint:errcheck // report stream
		fmt.Fprintf(out, "  recorded: %d non-advisory warning(s) across %d rule(s)\n", recorded, len(b.Rules)) //nolint:errcheck // report stream

		// The baseline records no errors, so a rebaseline must not read as a
		// clean run while one stands (REQ-SLGS-009).
		if report.HasErrors() {
			fmt.Fprintf(out, "baseline: ERROR-GATED — error-severity findings stand; the baseline never absorbs them\n") //nolint:errcheck // report stream
			return &exitCodeError{code: 1, msg: "spec lint: error-severity findings detected"}
		}
		return nil
	}

	baseline, loadErr := spec.LoadBaseline(path)
	if loadErr != nil {
		return argumentError(cmd.ErrOrStderr(), "spec lint: %v", loadErr)
	}

	verdict := spec.CompareBaseline(report, baseline)

	// Error-first. Reported as its own cause, ahead of any ratchet verdict.
	if verdict.ErrorGated() {
		fmt.Fprintf(out, "\nbaseline: ERROR-GATED — %d error-severity finding(s); the baseline never absorbs an error\n", verdict.ErrorCount) //nolint:errcheck // report stream
		printBaselineInventory(cmd, verdict, baseline)
		return &exitCodeError{code: 1, msg: "spec lint: error-severity findings detected"}
	}

	if len(verdict.Increases) > 0 {
		fmt.Fprintf(out, "\nbaseline: EXCEEDED — %s\n", path) //nolint:errcheck // report stream
		for _, d := range verdict.Increases {
			fmt.Fprintf(out, "  %s: recorded %d -> current %d (+%d)\n", d.Code, d.Recorded, d.Current, d.Delta()) //nolint:errcheck // report stream
		}
		printBaselineInventory(cmd, verdict, baseline)
		return &exitCodeError{code: 1, msg: "spec lint: baseline exceeded"}
	}

	fmt.Fprintf(out, "\nbaseline: OK — %s\n", path) //nolint:errcheck // report stream
	for _, d := range verdict.Decreases {
		fmt.Fprintf(out, "  decreased: %s recorded %d -> current %d (%d)\n", d.Code, d.Recorded, d.Current, d.Delta()) //nolint:errcheck // report stream
	}
	printBaselineInventory(cmd, verdict, baseline)
	return nil
}

// printBaselineInventory keeps the standing inventory observable even while the
// gate is green — a green gate that hides the inventory trades one blind spot
// for another (REQ-SLGS-004, REQ-SLGS-010).
func printBaselineInventory(cmd *cobra.Command, c spec.BaselineComparison, b *spec.Baseline) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "  inventory: %d warning(s) total (advisory included), %d non-advisory tracked across %d recorded rule(s)\n", //nolint:errcheck // report stream
		c.TotalWarnings, c.NonAdvisoryWarnings, len(b.Rules))
	if b.TreeSHA != "" || b.UpdatedAt != "" {
		fmt.Fprintf(out, "  recorded at: %s (%s)\n", b.TreeSHA, b.UpdatedAt) //nolint:errcheck // report stream
	}
}

// resolveTreeSHA reports the short HEAD SHA of the tree the measurement was
// taken against. A count with no tree attribution is not a measurement, so the
// field is always populated — "unknown" when git cannot answer (a corpus that
// is not a git repository, which is the normal case under test).
func resolveTreeSHA(dir string) string {
	sha, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	trimmed := strings.TrimSpace(string(sha))
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
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

// resolveLintTargets maps each CLI argument onto the path that should be
// linted. SPEC-SPEC-LINT-ID-ARG-001 (REQ-SLI-001, REQ-SLI-003, REQ-SLI-005).
//
// An empty argument list is returned untouched: the whole-corpus scan is a
// different code path with a different base directory (detectBaseDir(cwd)), and
// leaving it alone is what keeps a no-argument invocation unchanged.
//
// The project root is looked up LAZILY — only once an ID-shaped argument is
// actually seen. Doing it eagerly would make a path-only invocation fail
// outside a project, and the path form is this SPEC's control group.
// The diagnostic is WRITTEN to stderr as well as carried by the returned
// error, because the error's message never reaches the terminal on this path:
// an ExitCoder returned from RunE sets the process exit code and is rendered
// as nothing. That silence is pre-existing — the older exit-3 case in this same
// command ("cannot use --json and --sarif together") is silent too — but
// REQ-SLI-005 asks for a message naming the path that was tried, and an exit
// code alone does not name anything.
func resolveLintTargets(stderr io.Writer, args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

	var projectRoot string
	targets := make([]string, 0, len(args))
	for _, arg := range args {
		// Path signals win first (plan.md D1: signal → shape → path). The
		// pre-exclusion is what stops a path being misread as an ID, which is
		// the worse of the two misreadings: the user believes the file they
		// named was the file that was checked.
		if hasPathSignal(arg) {
			targets = append(targets, expandSpecDirArg(arg))
			continue
		}

		if !specid.HasCanonicalSpecIDShape(arg) {
			// Not a path signal and not ID-shaped: treat it as a path, which
			// is the pre-existing behaviour for anything unrecognised.
			targets = append(targets, expandSpecDirArg(arg))
			continue
		}

		// Sibling-command idiom (spec view/status/close): sanitize before the
		// ID reaches filepath.Join. Note this call is UNREACHABLE defence in
		// depth here — ValidateSpecID rejects absolute paths, ".." and path
		// separators, and every one of those carries a path signal that the
		// branch above has already consumed. It is kept because the idiom is
		// the boundary contract, not because a rejection is expected; no
		// acceptance criterion asserts that it fires.
		if err := specid.ValidateSpecID(arg); err != nil {
			return nil, argumentError(stderr, "spec lint: %v", err)
		}

		if projectRoot == "" {
			root, err := findProjectRootFn()
			if err != nil {
				return nil, argumentError(stderr, "spec lint: cannot resolve SPEC-ID %q: failed to find project root: %v", arg, err)
			}
			projectRoot = root
		}

		specPath := filepath.Join(projectRoot, ".moai", "specs", arg, "spec.md")
		if _, err := os.Stat(specPath); err != nil {
			// An argument-shape mistake is an argument error, not a finding
			// about a document (REQ-SLI-005). The path that was tried is named
			// so the reader can see what the ID was taken to mean. The first
			// failure ends the run: a partial run followed by exit 3 blurs
			// what was actually checked (plan.md D2).
			return nil, argumentError(stderr, "spec lint: no SPEC document found for %q (tried: %s)", arg, specPath)
		}
		targets = append(targets, specPath)
	}

	return targets, nil
}

// argumentError writes the diagnostic where the user can read it and returns
// the exit-3 error that carries the same text to the process exit code.
func argumentError(stderr io.Writer, format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintln(stderr, msg)
	return &exitCodeError{code: 3, msg: msg}
}

// hasPathSignal reports whether the argument carries a syntactic path signal.
// Structural only — no filesystem lookup — so the classification cannot change
// underneath a caller depending on what happens to exist (plan.md D1 rejects
// probe-ordered discrimination for exactly that reason).
func hasPathSignal(arg string) bool {
	return strings.ContainsAny(arg, `/\`) || strings.HasSuffix(arg, ".md")
}

// expandSpecDirArg turns a SPEC directory into the spec.md inside it, and
// leaves everything else alone — including paths that do not exist, which must
// keep producing the same ParseFailure they produced before (REQ-SLI-004).
func expandSpecDirArg(arg string) string {
	if info, err := os.Stat(arg); err == nil && info.IsDir() {
		return filepath.Join(arg, "spec.md")
	}
	return arg
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
