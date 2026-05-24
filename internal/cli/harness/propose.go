// Package harness hosts the `moai harness <verb>` CLI surface extensions
// authored by SPEC-V3R6-HARNESS-PROPOSAL-GEN-001. The package intentionally
// lives in its own subdirectory (separate from the historical
// internal/cli/harness*.go files in package `cli`) so the C-HRA-008-class
// boundary guard (TestPropose_NoAskUserQuestion) can scan every source file
// under one directory without false positives from unrelated files.
//
// HARD subagent boundary: NO source file in this package may invoke
// AskUserQuestion. Prose comments describing orchestrator behavior are
// permitted; the parenthesized invocation form is forbidden — see the
// boundary guard in propose_boundary_test.go. SPEC: REQ-PGN-012,
// REQ-PGN-013.
package harness

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// NewProposeCmd is the `moai harness propose` cobra factory.
//
// The function is exported so that the existing parent command factory in
// `internal/cli/harness_route.go` (package `cli`) can register the
// subcommand without creating a duplicate parent — see plan.md §3.2
// architectural decision (Option A).
//
// The orchestrator presents an AskUserQuestion gate (Approve / Modify /
// Reject) after consuming this command's JSON output when --auto is set and
// at least one actionable proposal exists. The gate itself lives entirely
// outside this binary; this command emits a structured payload and exits.
func NewProposeCmd() *cobra.Command {
	flags := proposalgen.OutputFlags{}
	cmd := &cobra.Command{
		Use:   "propose",
		Short: "Generate draft SPEC proposals from harness learning history",
		Long: `Consume .moai/harness/learning-history/tier-promotions.jsonl and emit draft
SPEC proposal candidates that map onto actionable learning patterns. The
generator is a graceful no-op when the input file is absent, empty, or
contains only system-event patterns (current data state as of 2026-05-24).

The CLI emits a single JSON object to stdout describing the proposal set,
including a machine-readable 'reason' diagnostic and an 'auto_delegate' flag
that signals the orchestrator's AskUserQuestion gate handoff.

This subcommand never invokes AskUserQuestion. User interaction (Approve /
Modify / Reject) is owned exclusively by the orchestrator per the V3R4
subagent boundary HARD contract.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPropose(cmd, flags)
		},
	}

	cmd.Flags().BoolVar(&flags.Auto, "auto", false,
		"Signal orchestrator to launch AskUserQuestion gate when proposals exist")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false,
		"Evaluate and report candidates without writing to .moai/proposals/")
	cmd.Flags().IntVar(&flags.Limit, "limit", proposalgen.DefaultLimit,
		"Maximum number of proposals to emit (sorted by Confidence descending)")
	cmd.Flags().StringVar(&flags.InputPath, "input", "",
		"Override the tier-promotions.jsonl path (default: "+proposalgen.DefaultInputPath+")")
	cmd.Flags().StringVar(&flags.OutputDir, "output-dir", "",
		"Override the proposals output directory (default: "+proposalgen.DefaultOutputDir+")")
	return cmd
}

// runPropose orchestrates the read → map → optionally scaffold pipeline and
// emits the GeneratorResult as JSON to stdout.
//
// Exit semantics per REQ-PGN-009:
//   - 0 on success (including no-op).
//   - 1 on unrecoverable filesystem / IO error (returned via cobra's RunE).
//   - 2 on malformed CLI flags (cobra handles automatically).
func runPropose(cmd *cobra.Command, flags proposalgen.OutputFlags) error {
	inputPath := flags.InputPath
	if inputPath == "" {
		inputPath = proposalgen.DefaultInputPath
	}
	outputDir := flags.OutputDir
	if outputDir == "" {
		outputDir = proposalgen.DefaultOutputDir
	}
	limit := flags.Limit
	if limit <= 0 {
		limit = proposalgen.DefaultLimit
	}

	promotions, malformed, err := proposalgen.ReadPromotions(inputPath)
	if err != nil {
		return fmt.Errorf("propose: read promotions: %w", err)
	}

	candidates := proposalgen.MapPromotions(promotions)
	evaluated := countEvaluated(promotions)

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := proposalgen.GeneratorResult{
		Proposals:         candidates,
		MalformedLines:    malformed,
		EvaluatedPatterns: evaluated,
	}

	switch {
	case len(promotions) == 0:
		result.Reason = "tier-promotions.jsonl absent or empty"
	case len(candidates) == 0:
		result.Reason = "no-actionable-patterns"
	default:
		result.Reason = "ok"
	}

	// auto_delegate handoff per REQ-PGN-010: true only when --auto AND at
	// least one proposal exists. The orchestrator interprets this flag to
	// decide whether to launch the AskUserQuestion gate.
	result.AutoDelegate = flags.Auto && len(candidates) > 0

	if !flags.DryRun && len(candidates) > 0 {
		if _, err := proposalgen.WriteProposals(outputDir, candidates); err != nil {
			return fmt.Errorf("propose: write proposals: %w", err)
		}
	}

	if result.Proposals == nil {
		result.Proposals = []proposalgen.ProposalCandidate{}
	}

	out, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("propose: marshal result: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return nil
}

// countEvaluated returns the number of unique pattern_key values observed in
// the input promotions. The CLI reports this as evaluated_patterns in the
// JSON output to surface the mapper's evaluation surface to the orchestrator.
func countEvaluated(promotions []harness.Promotion) int {
	seen := make(map[string]struct{}, len(promotions))
	for _, p := range promotions {
		seen[p.PatternKey] = struct{}{}
	}
	return len(seen)
}
