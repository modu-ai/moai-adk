package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/report/planhtml"
)

// newPlanCmd constructs the `moai plan` Cobra parent command with its first
// subcommand `render-html` (SPEC-GOAL-HTML-WIRING-001 REQ-WIRE-004/005/006/007).
//
// IMPORTANT: this CLI surface MUST NOT invoke AskUserQuestion (subagent boundary,
// C-HRA-008). It returns exit codes + structured stdout/stderr only; the
// orchestrator owns all user interaction. The `moai plan` CLI verb is DISTINCT
// from the `/moai plan` slash-command skill — the former is a Go binary
// subcommand, the latter routes through the Claude Code skill system (spec.md
// §G out-of-scope: convergence is a separate follow-up).
func newPlanCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan-phase CLI utilities (render-html, …)",
		Long: `Plan-phase CLI utilities.

The 'render-html' subcommand renders the plan-phase HTML report for a SPEC,
consuming the plan-auditor review markdown (fail-open) and writing a
self-contained .html to .moai/reports/plan-html/<SPEC-ID>-plan.html.

NOTE: this CLI verb is distinct from the '/moai plan' slash-command skill. The
skill routes through the Claude Code skill system to author SPEC artifacts via
manager-spec; this CLI verb is a Go binary subcommand that emits the HTML
report from already-authored artifacts.`,
		SilenceUsage: true,
	}
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")

	renderHTMLCmd := &cobra.Command{
		Use:          "render-html <SPEC-ID>",
		Short:        "Render the plan-phase HTML report for a SPEC",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runPlanRenderHTML(c, args[0], jsonOutput)
		},
	}
	cmd.AddCommand(renderHTMLCmd)
	return cmd
}

// runPlanRenderHTML resolves the SPEC dir, finds the most-recent plan-audit
// review file, calls planhtml.RenderPlanHTML, and writes the bytes to
// .moai/reports/plan-html/<SPEC-ID>-plan.html.
//
//   - Missing SPEC dir (REQ-WIRE-007): non-zero exit + stderr naming the
//     SPEC-ID + NO html file written.
//   - Missing/unparseable review file (REQ-WIRE-006 / REQ-GHF-007 fail-open):
//     the report is still written with the "audit verdict unavailable"
//     placeholder AND exit 0.
func runPlanRenderHTML(cmd *cobra.Command, specID string, jsonOutput bool) error {
	root := resolveProjectDir()
	if root == "" {
		return fmt.Errorf("plan render-html: cannot resolve project root (set CLAUDE_PROJECT_DIR or run from the project directory)")
	}
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if _, err := os.Stat(specDir); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"plan render-html: SPEC directory not found for %s (expected %s)\n",
			specID, specDir)
		return fmt.Errorf("plan render-html: missing SPEC directory for %s", specID)
	}
	reviewFile := mostRecentReview(root, specID) // "" when absent (fail-open)

	raw, err := planhtml.RenderPlanHTML(specDir, reviewFile)
	if err != nil {
		return fmt.Errorf("plan render-html: %w", err)
	}
	outDir := filepath.Join(root, ".moai", "reports", "plan-html")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("plan render-html mkdir %s: %w", outDir, err)
	}
	outPath := filepath.Join(outDir, specID+"-plan.html")
	if err := os.WriteFile(outPath, raw, 0o644); err != nil {
		return fmt.Errorf("plan render-html write %s: %w", outPath, err)
	}

	if jsonOutput {
		return emitOK(cmd, true, map[string]any{
			"action":  "render-html",
			"spec_id": specID,
			"path":    outPath,
			"bytes":   len(raw),
		})
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"rendered plan report for %s to %s\n", specID, outPath)
	return nil
}

// mostRecentReview returns the path of the highest-numbered
// .moai/reports/plan-audit/<SPEC-ID>-review-<N>.md file, or "" if none exists
// (the renderer's fail-open path). <N> is parsed numerically so review-10 sorts
// above review-9.
func mostRecentReview(root, specID string) string {
	dir := filepath.Join(root, ".moai", "reports", "plan-audit")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	prefix := specID + "-review-"
	var best int = -1
	var bestName string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".md") {
			continue
		}
		numStr := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".md")
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if n > best {
			best = n
			bestName = name
		}
	}
	if best < 0 {
		return ""
	}
	return filepath.Join(dir, bestName)
}
