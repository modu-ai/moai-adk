package cli

// @MX:NOTE: [AUTO] ast-edit is the write-side counterpart of ast-grep; ast-grep only reads.
// @MX:WARN: [AUTO] A non-dry run rewrites source files in place.
// @MX:REASON: [AUTO] Registering it separately from ast-grep keeps a Bash(moai ast-grep:*)
// permission grant read-only — a --fix flag on ast-grep would have widened that grant to writes.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// astEditTimeout bounds a whole ast-edit invocation, including every per-rule pass.
const astEditTimeout = 5 * time.Minute

// astEditFlags holds flag values for the ast-edit subcommand.
type astEditFlags struct {
	dry      bool
	pattern  string
	rewrite  string
	lang     string
	rulesDir string
	rule     string
	format   string
}

// NewAstEditCmd creates and returns the `moai ast-edit` Cobra command.
// REQ-AGE-001, REQ-AGE-002, REQ-AGE-003, REQ-AGE-004
func NewAstEditCmd() *cobra.Command {
	flags := &astEditFlags{}

	cmd := &cobra.Command{
		Use:   "ast-edit [path]",
		Short: "Rewrite code using ast-grep",
		Long: `Applies ast-grep rewrites to the specified path using the ast-grep (sg) CLI.

This command MODIFIES FILES IN PLACE. Run it with --dry first to preview the
changes; --dry computes and prints the same result without touching any file.

Two modes:
  pattern  --pattern <p> --rewrite <r>   rewrite every match of <p> to <r>
  rule     (no --pattern)                apply the fix: field of loaded rules

Examples:
  moai ast-edit --dry --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/
  moai ast-edit --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/
  moai ast-edit --dry ./internal/
  moai ast-edit --rule go-error-not-wrapped ./internal/`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return runAstEdit(cmd, flags, path)
		},
	}

	cmd.Flags().BoolVar(&flags.dry, "dry", false, "Preview the changes without modifying any file")
	cmd.Flags().StringVar(&flags.pattern, "pattern", "", "ast-grep pattern to match (requires --rewrite)")
	cmd.Flags().StringVar(&flags.rewrite, "rewrite", "", "Replacement pattern (requires --pattern)")
	cmd.Flags().StringVar(&flags.lang, "lang", "", "Language of the target code (e.g. go, python, typescript)")
	cmd.Flags().StringVar(&flags.rulesDir, "rules-dir", ".moai/config/astgrep-rules", "ast-grep rules directory path")
	cmd.Flags().StringVar(&flags.rule, "rule", "", "Apply only the rule with this ID (rule mode)")
	cmd.Flags().StringVar(&flags.format, "format", "text", "Output format: text, json")

	return cmd
}

// runAstEdit dispatches to pattern mode or rule mode and reports the result.
func runAstEdit(cmd *cobra.Command, flags *astEditFlags, path string) error {
	if err := validateAstEditFlags(flags); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), astEditTimeout)
	defer cancel()

	analyzer := astgrep.NewAnalyzer(".")
	if !analyzer.IsSGAvailable(ctx) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "ast-grep (sg) is not installed; nothing to apply.")
		return nil
	}

	var (
		result *astgrep.ReplaceResult
		err    error
	)
	if flags.pattern != "" {
		result, err = applyPatternEdit(ctx, analyzer, flags, path)
	} else {
		result, err = applyRuleEdits(ctx, analyzer, flags, path, cmd)
	}
	if err != nil {
		return err
	}

	return renderAstEditResult(cmd, flags, result)
}

// validateAstEditFlags rejects flag combinations that cannot be executed.
func validateAstEditFlags(flags *astEditFlags) error {
	if flags.pattern != "" && flags.rewrite == "" {
		return fmt.Errorf("--pattern requires --rewrite (the replacement to apply)")
	}
	if flags.rewrite != "" && flags.pattern == "" {
		return fmt.Errorf("--rewrite requires --pattern (the code to match)")
	}
	if flags.pattern != "" && flags.rule != "" {
		return fmt.Errorf("--rule applies to rule mode and cannot be combined with --pattern")
	}
	if flags.format != "text" && flags.format != "json" {
		return fmt.Errorf("unsupported --format %q (want text or json)", flags.format)
	}
	return nil
}

// applyPatternEdit runs a single pattern rewrite. REQ-AGE-003.
func applyPatternEdit(ctx context.Context, analyzer *astgrep.SGAnalyzer, flags *astEditFlags, path string) (*astgrep.ReplaceResult, error) {
	result, err := analyzer.PatternReplace(ctx, flags.pattern, flags.rewrite, flags.lang, path, flags.dry)
	if err != nil {
		return nil, fmt.Errorf("pattern rewrite: %w", err)
	}
	return result, nil
}

// applyRuleEdits applies the fix: field of every loaded rule that declares one.
// Rules without a fix: are detection-only and are skipped with a counted notice.
// REQ-AGE-004.
func applyRuleEdits(ctx context.Context, analyzer *astgrep.SGAnalyzer, flags *astEditFlags, path string, cmd *cobra.Command) (*astgrep.ReplaceResult, error) {
	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(flags.rulesDir)
	if err != nil {
		return nil, fmt.Errorf("load rules from %s: %w", flags.rulesDir, err)
	}

	aggregate := &astgrep.ReplaceResult{DryRun: flags.dry}
	var skipped int

	for _, rule := range rules {
		if flags.rule != "" && rule.ID != flags.rule {
			continue
		}
		// A rule is rewritable only when the loader surfaced both a pattern to
		// match and a fix to apply. Rules written with ast-grep's nested `rule:`
		// block (kind/regex/any) land here with an empty Pattern — the loader
		// reads only the flat `pattern:` field — so they are skipped, not
		// silently passed to sg with an empty pattern.
		if rule.Fix == "" || rule.Pattern == "" {
			skipped++
			continue
		}

		lang := flags.lang
		if lang == "" {
			lang = rule.Language
		}

		result, err := analyzer.PatternReplace(ctx, rule.Pattern, rule.Fix, lang, path, flags.dry)
		if err != nil {
			return nil, fmt.Errorf("apply rule %s: %w", rule.ID, err)
		}

		aggregate.MatchesFound += result.MatchesFound
		aggregate.FilesModified += result.FilesModified
		aggregate.Changes = append(aggregate.Changes, result.Changes...)
	}

	if skipped > 0 && flags.format == "text" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "skipped %d detection-only rule(s) (no fix: field)\n", skipped)
	}

	return aggregate, nil
}

// renderAstEditResult prints the replacement result in the requested format.
func renderAstEditResult(cmd *cobra.Command, flags *astEditFlags, result *astgrep.ReplaceResult) error {
	out := cmd.OutOrStdout()

	if flags.format == "json" {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			return fmt.Errorf("encode result: %w", err)
		}
		return nil
	}

	if result.DryRun {
		_, _ = fmt.Fprintf(out, "dry run: %d match(es) in %d file(s) would be rewritten\n",
			result.MatchesFound, result.FilesModified)
	} else {
		_, _ = fmt.Fprintf(out, "rewrote %d match(es) across %d file(s)\n",
			result.MatchesFound, result.FilesModified)
	}

	for _, change := range result.Changes {
		_, _ = fmt.Fprintf(out, "  %s:%d:%d\n", change.FilePath, change.Line, change.Column)
	}

	return nil
}
