// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M2 — `moai spec audit` CLI subcommand.
//
// This file wires the cobra command surface for `moai spec audit`. All
// classification / drift detection logic lives in internal/spec/audit.go (M1
// deliverable); this layer is responsible for:
//   - cobra flag parsing (--json / --filter-era / --include-grandfathered /
//                         --strict / --base-dir)
//   - delegating to spec.Audit() with the parsed AuditOptions
//   - rendering JSON output per AC-LSG-007 schema:
//       {audited_at, total_specs, grandfathered, modern_era_clean, drift_findings}
//   - rendering a human-readable summary as the default output format
//   - mapping the strict-mode escalation to a non-zero exit code when drift
//     findings exist (orchestrator translates the exit code per the agent
//     common protocol).
//
// AC bindings (M2 scope):
//   - AC-LSG-002  — era classification 5 buckets surfaced through CLI
//   - AC-LSG-007  — JSON output schema
//   - AC-LSG-016  — NFR performance (the M1 benchmark verifies < 5s; M2 wires
//                   the CLI surface that the benchmark exercises end-to-end)
//
// SUBAGENT BOUNDARY (C-HRA-008): This file MUST NOT invoke AskUserQuestion
// or mcp__askuser__*. CLI runs in subagent context; orchestrator owns user
// interaction. See .claude/rules/moai/core/agent-common-protocol.md
// § User Interaction Boundary.

// @MX:NOTE: [AUTO] `moai spec audit` CLI subcommand for era classification + drift detection per SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M2
// @MX:ANCHOR: [AUTO] newSpecAuditCmd is the cobra entry point for `moai spec audit`; delegates to internal/spec.Audit()
// @MX:REASON: [AUTO] fan_in=2 (spec.go registers it via newSpecCmd; tests construct it directly); single source of truth for CLI surface

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecAuditCmd constructs the cobra command implementing `moai spec audit`.
//
// Flags:
//   --json                       Emit AuditResult as JSON on stdout per AC-LSG-007.
//   --filter-era <era>           Restrict drift_findings to one era
//                                (V2.x / V3R2-R4 / V3R5 / V3R6 / unclassified).
//   --include-grandfathered      Surface pre-V3R6 SPECs as INFO findings.
//   --strict                     Exit code 1 when any MUST-FIX drift finding exists.
//   --base-dir <path>            Project root (defaults to current working directory).
func newSpecAuditCmd() *cobra.Command {
	var (
		jsonOutput           bool
		filterEra            string
		filterSpec           string
		includeGrandfathered bool
		strict               bool
		baseDir              string
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit SPEC era classification and detect modern-era status drift",
		Long: `Scan .moai/specs/SPEC-*/ directories, classify each SPEC by era heuristic,
and emit drift findings for V3R6 SPECs with cross-tab pattern violations.

Era classification (5 buckets):
  - V2.x         pre-V3 SPECs (grandfather clause active per AC-LSG-017)
  - V3R2-R4      mid-cycle V3 SPECs (grandfather clause active)
  - V3R5         late-cycle V3 SPECs (grandfather clause active)
  - V3R6         modern-era SPECs (subject to drift detection)
  - unclassified no heuristic matched (INFO finding only)

Drift patterns (V3R6 only):
  - Y_N_N_Y               sync section present but Mx section absent + status drift
  - Y_Y_N_Y               §E.2 + §E.5 present but mx_commit_sha missing
  - Y_Y_Y_Y_StatusDrift   all 4 phase markers + valid SHAs but status != completed

Each MUST-FIX finding includes a remediation command (typically
`+"`moai spec close <SPEC-ID> --backfill-only`"+`) that resolves the drift.

JSON output schema (--json):
  {
    "audited_at": "RFC3339 timestamp",
    "total_specs": <integer>,
    "grandfathered": <integer>,
    "modern_era_clean": <integer>,
    "drift_findings": [
      {"spec_id": "...", "era": "...", "finding_type": "...", "severity": "...", "remediation": "..."}
    ]
  }

Exit codes:
  0 = success (audit completed; findings emitted)
  1 = strict mode + MUST-FIX drift findings detected
  2 = audit engine error (invalid spec directory, IO failure)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := spec.AuditOptions{
				BaseDir:              baseDir,
				FilterEra:            filterEra,
				FilterSpec:           filterSpec,
				IncludeGrandfathered: includeGrandfathered,
				Strict:               strict,
			}

			result, err := spec.Audit(opts)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error running audit: %v\n", err)
				return err
			}

			return renderAuditResult(cmd, result, jsonOutput, strict)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false,
		"Emit AuditResult as JSON on stdout (per AC-LSG-007 schema)")
	cmd.Flags().StringVar(&filterEra, "filter-era", "",
		"Restrict drift_findings to one era (V2.x / V3R2-R4 / V3R5 / V3R6 / unclassified)")
	cmd.Flags().StringVar(&filterSpec, "filter-spec", "",
		"Restrict drift_findings to one SPEC-ID (exact match, e.g. SPEC-V3R6-ORCH-IGGDA-001); empty = all SPECs")
	cmd.Flags().BoolVar(&includeGrandfathered, "include-grandfathered", false,
		"Surface pre-V3R6 SPECs as INFO findings (otherwise excluded)")
	cmd.Flags().BoolVar(&strict, "strict", false,
		"Exit code 1 when any MUST-FIX drift finding exists")
	cmd.Flags().StringVar(&baseDir, "base-dir", "",
		"Project root directory (default: current working directory)")

	return cmd
}

// renderAuditResult writes the AuditResult to the CLI's stdout in the requested
// format and applies the strict-mode escalation.
//
// The JSON path emits the AuditResult struct directly; the human-readable path
// emits a summary header + per-finding lines, designed for terminal scanning.
func renderAuditResult(cmd *cobra.Command, result *spec.AuditResult, jsonOutput, strict bool) error {
	out := cmd.OutOrStdout()

	if jsonOutput {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal audit result: %w", err)
		}
		_, _ = fmt.Fprintln(out, string(data))
	} else {
		renderAuditHuman(out, result)
	}

	// Strict mode: exit 1 if any MUST-FIX drift finding is present.
	if strict {
		for _, f := range result.DriftFindings {
			if f.Severity == "MUST-FIX" {
				return fmt.Errorf("strict mode: %d MUST-FIX drift finding(s) detected",
					countMustFix(result.DriftFindings))
			}
		}
	}

	return nil
}

// renderAuditHuman emits a human-readable summary of the audit result.
//
// Format:
//   Audit summary
//   =============
//   Total SPECs:        <N>
//   Grandfathered:      <N> (pre-V3R6 — protected)
//   Modern-era clean:   <N>
//   Drift findings:     <N>
//
//   Findings:
//     [MUST-FIX] SPEC-XXX (V3R6) — Y_Y_Y_Y_StatusDrift
//                Remediation: moai spec close SPEC-XXX --backfill-only
//     [INFO]     SPEC-YYY (V2.x) — Grandfathered
//
// The format prioritizes scannability over machine readability; downstream
// consumers that need structured data should use --json.
func renderAuditHuman(out interface{ Write(p []byte) (int, error) }, result *spec.AuditResult) {
	_, _ = fmt.Fprintln(out, "Audit summary")
	_, _ = fmt.Fprintln(out, "=============")
	_, _ = fmt.Fprintf(out, "Total SPECs:        %d\n", result.TotalSpecs)
	_, _ = fmt.Fprintf(out, "Grandfathered:      %d (pre-V3R6 — protected)\n", result.Grandfathered)
	_, _ = fmt.Fprintf(out, "Modern-era clean:   %d\n", result.ModernEraClean)
	_, _ = fmt.Fprintf(out, "Drift findings:     %d\n", len(result.DriftFindings))

	if len(result.DriftFindings) == 0 {
		_, _ = fmt.Fprintln(out, "\nNo drift findings — all modern-era SPECs clean.")
		return
	}

	_, _ = fmt.Fprintln(out, "\nFindings:")
	for _, f := range result.DriftFindings {
		eraSuffix := ""
		if f.Era != "" {
			eraSuffix = fmt.Sprintf(" (%s)", f.Era)
		}
		_, _ = fmt.Fprintf(out, "  [%s] %s%s — %s\n",
			f.Severity, f.SpecID, eraSuffix, f.FindingType)
		if f.Remediation != "" {
			_, _ = fmt.Fprintf(out, "             Remediation: %s\n", f.Remediation)
		}
		if reason, ok := f.Details["reason"].(string); ok && reason != "" {
			// Trim multi-line reasons to first line for compact human output.
			line := strings.SplitN(reason, "\n", 2)[0]
			_, _ = fmt.Fprintf(out, "             Reason: %s\n", line)
		}
	}
}

// countMustFix returns the number of MUST-FIX findings in the slice. Used only
// for the strict-mode error message.
func countMustFix(findings []spec.DriftFinding) int {
	n := 0
	for _, f := range findings {
		if f.Severity == "MUST-FIX" {
			n++
		}
	}
	return n
}
