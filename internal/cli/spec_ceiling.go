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
	"gopkg.in/yaml.v3"
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
	if err := runtime.RecordCeilingOutcome(root, specID, outcome); err != nil {
		return fmt.Errorf("spec ceiling: %w", err)
	}
	_, _ = fmt.Fprintf(out, "recorded: %s\n", filepath.Join(root, runtime.AuditCeilingStateDir, specID+".json"))
	return nil
}

// loadCeilingConfig reads plan_audit_tier_ceilings and
// plan_audit_ceiling_policy.on_final_hit from the tree's harness.yaml; a
// missing harness.yaml resolves the shipped defaults (an absent file is a
// defaults project, not an error).
//
// A harness.yaml the strict loader rejects wholesale is distinguished by WHAT
// cannot be read (CR2-P2-1), three ways:
//
//   - Policy-only unreadable (REQ-ACR-003's unreadable-policy arm): the
//     policy block's own values cannot be read — on_final_hit as a sequence,
//     auto_delta_rounds as a map — but the ceilings field can. Not a fatal
//     configuration error: the loose partial ceilings merge over the shipped
//     defaults (R3-P2-2, the same seeded-merge rule internal/config/loader.go
//     applies) and the empty policy string drops the evaluation to
//     disposition hold, record included.
//   - Ceilings malformed: the plan_audit_tier_ceilings field itself cannot
//     be decoded. Recovery covers ONLY the policy field, so the strict error
//     propagates even when the policy is the unreadable part, and the error
//     names the ceiling parse so the record abstention is auditable
//     (R3-P2-1).
//   - Every other load error stays fatal.
func loadCeilingConfig(root string) (map[string]int, string, error) {
	path := filepath.Join(root, ".moai", "config", "sections", "harness.yaml")
	cfg, err := config.LoadHarnessConfig(path)
	if err == nil {
		return cfg.PlanAuditTierCeilings, cfg.PlanAuditCeilingPolicy.OnFinalHit, nil
	}
	if errors.Is(err, config.ErrConfigNotFound) {
		def := config.DefaultPlanAuditCeilingPolicy()
		return config.DefaultPlanAuditTierCeilings(), def.OnFinalHit, nil
	}
	ceilings, ceilingsMalformed, policyUnreadable := looseReadCeilingPolicy(path)
	if ceilingsMalformed {
		// R3-P2-1: the ceilings field itself cannot be decoded — recovery
		// covers only the policy field, so the strict error propagates even
		// when the policy is the unreadable part. The message names the
		// ceiling field so the record abstention stays auditable.
		return nil, "", fmt.Errorf("harness.yaml plan_audit_tier_ceilings cannot be decoded (ceiling parse failed): %w", err)
	}
	if !policyUnreadable {
		return nil, "", err // the failure lives elsewhere in the configuration — stay strict
	}
	if len(ceilings) == 0 {
		ceilings = config.DefaultPlanAuditTierCeilings() // the key is absent entirely — the full defaults
	} else {
		merged := config.DefaultPlanAuditTierCeilings() // R3-P2-2: partial override over the shipped
		for tier, n := range ceilings {                 // defaults, the loader's seeded-merge rule
			merged[tier] = n
		}
		ceilings = merged
	}
	return ceilings, "", nil // unreadable policy value → the evaluation's hold arm
}

// looseReadCeilingPolicy re-reads the harness.yaml the strict loader rejected
// and reports how the two ceiling-relevant fields read:
//
//   - ceilingsMalformed: the plan_audit_tier_ceilings field is present but
//     cannot be decoded into map[string]int (the raw any survives, the typed
//     map does not) — R3-P2-1's propagate signal.
//   - policyUnreadable: the POLICY block's own values are what cannot be
//     read. The policy fields decode into any so every value shape survives
//     the tolerant pass; a present-but-non-string on_final_hit, or a
//     present-but-non-int auto_delta_rounds, is the unreadable policy.
//
// A readable policy block and a decodable ceilings field mean the strict
// failure lives elsewhere and stays fatal. The ceilings map decodes in a
// second pass into the typed map the strict loader uses, so a malformed
// ceilings field reports independently of the policy verdict.
func looseReadCeilingPolicy(path string) (ceilings map[string]int, ceilingsMalformed bool, policyUnreadable bool) {
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, false, false
	}
	var raw struct {
		Harness struct {
			PlanAuditTierCeilings  any `yaml:"plan_audit_tier_ceilings"`
			PlanAuditCeilingPolicy struct {
				AutoDeltaRounds any `yaml:"auto_delta_rounds"`
				OnFinalHit      any `yaml:"on_final_hit"`
			} `yaml:"plan_audit_ceiling_policy"`
		} `yaml:"harness"`
	}
	_ = yaml.Unmarshal(data, &raw) // tolerant by construction: any-typed fields cannot mismatch
	var typed struct {
		Harness struct {
			PlanAuditTierCeilings map[string]int `yaml:"plan_audit_tier_ceilings"`
		} `yaml:"harness"`
	}
	_ = yaml.Unmarshal(data, &typed) // the same typed decode the strict loader runs on this field
	if raw.Harness.PlanAuditTierCeilings != nil && typed.Harness.PlanAuditTierCeilings == nil {
		return nil, true, false // the ceilings field itself cannot be decoded — R3-P2-1
	}
	policy := raw.Harness.PlanAuditCeilingPolicy
	if policy.OnFinalHit == nil && policy.AutoDeltaRounds == nil {
		return nil, false, false // the policy block is absent — the strict failure lives elsewhere
	}
	unreadable := false
	if _, isString := policy.OnFinalHit.(string); policy.OnFinalHit != nil && !isString {
		unreadable = true
	}
	if _, isInt := policy.AutoDeltaRounds.(int); policy.AutoDeltaRounds != nil && !isInt {
		unreadable = true
	}
	if !unreadable {
		return nil, false, false // the policy values read fine — the strict failure lives elsewhere
	}
	return typed.Harness.PlanAuditTierCeilings, false, true
}
