package cli

// spec_ceiling.go — SPEC-AUDIT-CEILING-002: `moai spec ceiling` evaluates a
// SPEC's plan-audit iteration ceiling and — with --record — writes the one
// ceiling policy outcome record. It is the recording path's only caller this
// SPEC builds (D7): the path is a user- and agent-invocable CLI surface.
//
// C1/C-HRA-008: this verb asks nothing of anyone — no prompt, no interactive
// input of any kind. Read-only by default; --record is the explicit write.
//
// @MX:NOTE: [AUTO] the ceiling recording path's only caller this SPEC builds (D7) — evaluate, print, and with --record write through runtime.RecordCeilingOutcome
// @MX:SPEC: SPEC-AUDIT-CEILING-002

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditverdict"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/runtime"
	"github.com/spf13/cobra"
)

// newSpecCeilingCmd builds the `moai spec ceiling` verb.
func newSpecCeilingCmd() *cobra.Command {
	var (
		evidence []string
		record   bool
	)
	cmd := &cobra.Command{
		Use:   "ceiling <SPEC-ID> [--evidence <dir>]... [--record]",
		Short: "Evaluate a SPEC's plan-audit iteration ceiling",
		Long: `Evaluate a SPEC's plan-audit iteration ceiling (SPEC-AUDIT-CEILING-002).

The round count derives from durable iteration evidence on disk — the
SPEC-scoped .moai/reports/<SPEC-ID>/ directory plus every directory listed
with --evidence, one recorded file one round. The ceiling resolves from the
configured plan_audit_tier_ceilings (an absent or unknown tier resolves to L).

Read-only by default: it prints the evaluation. With --record, when the
ceiling applies to a non-admitted latest verdict, the outcome is recorded
through the one recording path to .moai/state/audit-ceiling/<SPEC-ID>.json
(a clean admitted PASS at the ceiling records nothing).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSpecCeiling(cmd, args[0], evidence, record)
		},
	}
	cmd.Flags().StringArrayVar(&evidence, "evidence", nil, "explicitly listed evidence directory (repeatable); an absent listing is an error")
	cmd.Flags().BoolVar(&record, "record", false, "write the ceiling outcome record when the ceiling applies")
	return cmd
}

// runSpecCeiling evaluates the ceiling for specID and optionally records the
// outcome. Every branch prints; the verb never asks a question.
func runSpecCeiling(cmd *cobra.Command, specID string, listedDirs []string, record bool) error {
	root := resolveProjectDir()
	if root == "" {
		return fmt.Errorf("spec ceiling: could not resolve the project root (run from the project, or set CLAUDE_PROJECT_DIR)")
	}
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if info, err := os.Stat(specDir); err != nil || !info.IsDir() {
		return fmt.Errorf("spec ceiling: no SPEC directory at %s", specDir)
	}

	ceilings, onFinalHit, err := loadCeilingConfig(root)
	if err != nil {
		return fmt.Errorf("spec ceiling: %w", err)
	}

	in := runtime.CeilingInput{
		SpecID:          specID,
		SpecEvidenceDir: filepath.Join(root, ".moai", "reports", specID),
		ListedDirs:      listedDirs,
		Tier:            auditverdict.SpecTier(specDir),
		Threshold:       auditverdict.PlanThreshold(specDir),
		Ceilings:        ceilings,
		OnFinalHit:      onFinalHit,
	}
	outcome, hit, err := runtime.EvaluatePlanAuditCeiling(in)
	if err != nil {
		return fmt.Errorf("spec ceiling: %w", err)
	}
	ceiling, err := runtime.ResolvePlanAuditCeiling(in.Tier, ceilings)
	if err != nil {
		return fmt.Errorf("spec ceiling: %w", err)
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "SPEC %s — tier %s, plan threshold %.2f\n", specID, in.Tier, in.Threshold)
	_, _ = fmt.Fprintf(out, "ceiling: %d  policy: %q\n", ceiling, onFinalHit)
	if !hit {
		_, _ = fmt.Fprintln(out, "no ceiling outcome: the ceiling does not apply (below ceiling, or a clean admitted PASS at the ceiling)")
		return nil
	}
	_, _ = fmt.Fprintf(out, "ceiling HIT at %d rounds — latest verdict %q (admitted=%t) — disposition %q\n", outcome.Count, outcome.VerdictLabel, outcome.VerdictAdmitted, outcome.Disposition)
	if outcome.SplitProposalRef != "" {
		_, _ = fmt.Fprintf(out, "split-proposal reference: %s\n", outcome.SplitProposalRef)
	}
	if len(outcome.DebtIDs) > 0 {
		_, _ = fmt.Fprintf(out, "debts: %s\n", strings.Join(outcome.DebtIDs, ", "))
	}

	if !record {
		_, _ = fmt.Fprintln(out, "read-only evaluation (pass --record to write the outcome record)")
		return nil
	}
	if err := runtime.RecordCeilingOutcome(specID, outcome); err != nil {
		return fmt.Errorf("spec ceiling: %w", err)
	}
	_, _ = fmt.Fprintf(out, "recorded: %s\n", filepath.Join(runtime.AuditCeilingStateDir, specID+".json"))
	return nil
}

// loadCeilingConfig reads plan_audit_tier_ceilings and
// plan_audit_ceiling_policy.on_final_hit from the tree's harness.yaml; a
// missing harness.yaml resolves the shipped defaults (an absent file is a
// defaults project, not an error).
func loadCeilingConfig(root string) (map[string]int, string, error) {
	cfg, err := config.LoadHarnessConfig(filepath.Join(root, ".moai", "config", "sections", "harness.yaml"))
	if errors.Is(err, config.ErrConfigNotFound) {
		def := config.DefaultPlanAuditCeilingPolicy()
		return config.DefaultPlanAuditTierCeilings(), def.OnFinalHit, nil
	}
	if err != nil {
		return nil, "", err
	}
	return cfg.PlanAuditTierCeilings, cfg.PlanAuditCeilingPolicy.OnFinalHit, nil
}
