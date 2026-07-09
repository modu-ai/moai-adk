// Package harness — v4 harness reference-integrity smoke gate (SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-005).
//
// Doctor scans every v4 harness under .claude/commands/harness/ and verifies a
// 4-axis reference-integrity contract:
//
//  1. entry command file exists (given by ListHarnesses enumeration)
//  2. manifest.json exists + schema-valid (v4manifest.Validate reuse)
//  3. Runner (harness-<name>-run.js) exists + its MANIFEST_PATH constant resolves
//     to a real file
//  4. each specialist agent referenced by the Runner resolves to an existing
//     .claude/agents/harness/<name>.md file
//
// Severity policy (design D2): findings are ERROR or INFO. A command-only thin
// harness (no manifest.json / Runner — the github / release maintainer harnesses)
// is reported as an INFO note and NEVER counted as an ERROR; the Runner/agent
// axes apply only when the manifest declares those artifacts. The exit code is
// non-zero only when >= 1 ERROR-severity finding exists.
//
// Runner parsing is a regex heuristic, NOT a JS AST parse (Enforce Simplicity /
// AP-2): the gate's purpose is to detect the B5 defect class (wrong MANIFEST_PATH
// constant + unresolved specialist agent reference), not to fully parse the JS.
//
// HARD subagent boundary (C-HRA-008): this file MUST NOT prompt the user via the
// deferred user-question tool — the CLI surfaces structured output; the
// orchestrator owns user interaction.
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
	"github.com/spf13/cobra"
)

// DoctorSeverity is a finding severity: ERROR (counts toward non-zero exit) or
// INFO (advisory note; never blocks).
type DoctorSeverity string

const (
	// SeverityError is a reference-integrity defect that fails the smoke gate.
	SeverityError DoctorSeverity = "ERROR"
	// SeverityInfo is an advisory note (e.g. a command-only thin harness).
	SeverityInfo DoctorSeverity = "INFO"
)

// DoctorFinding is a single cross-reference finding for one harness.
type DoctorFinding struct {
	Harness  string         `json:"harness"`
	Axis     string         `json:"axis"` // "manifest" | "runner" | "agent"
	Severity DoctorSeverity `json:"severity"`
	Message  string         `json:"message"`
}

// DoctorReport is the aggregate smoke-gate result.
type DoctorReport struct {
	Harnesses  int             `json:"harnesses"`
	Findings   []DoctorFinding `json:"findings"`
	ErrorCount int             `json:"error_count"`
	InfoCount  int             `json:"info_count"`
}

// runnerManifestPathRE extracts the MANIFEST_PATH string constant from a Runner
// JS: `const MANIFEST_PATH = "...";` (single, double, or backtick quotes).
var runnerManifestPathRE = regexp.MustCompile("MANIFEST_PATH\\s*=\\s*[\"'`]([^\"'`]+)[\"'`]")

// runnerSpecialistRE extracts specialist agent name references
// (harness-<x>-specialist) mentioned anywhere in a Runner JS.
var runnerSpecialistRE = regexp.MustCompile(`harness-[a-z0-9-]+-specialist`)

// Doctor runs the v4 harness reference-integrity smoke gate over projectRoot.
//
// @MX:ANCHOR: [AUTO] Doctor is the smoke-gate entry consumed by the CLI command
// and (per REQ-HEP-006) the Builder ACTIVATE contract.
// @MX:REASON: [AUTO] fan_in >= 2 growing: NewHarnessDoctorCmd, doctor_test.go,
// Builder ACTIVATE gate — the canonical v4 reference-integrity checker.
func Doctor(projectRoot string) (DoctorReport, error) {
	report := DoctorReport{}
	entries, err := ListHarnesses(projectRoot)
	if err != nil {
		return report, err
	}
	report.Harnesses = len(entries)
	for _, e := range entries {
		report.Findings = append(report.Findings, checkHarness(projectRoot, e)...)
	}
	// SPEC-HARNESS-RATCHET-REWIRE-001 REQ-HRR-008: pipeline-dormancy check
	// (independent of the v4 harness reference-integrity scan). Detects when the
	// classify half has produced promotions but the propose back half has never
	// run (proposals dir absent) — the starved-pipeline condition.
	report.Findings = append(report.Findings, checkPipelineDormancy(projectRoot)...)
	for _, f := range report.Findings {
		switch f.Severity {
		case SeverityError:
			report.ErrorCount++
		case SeverityInfo:
			report.InfoCount++
		}
	}
	return report, nil
}

// checkHarness applies the 4-axis cross-reference check to a single harness.
func checkHarness(projectRoot string, e HarnessEntry) []DoctorFinding {
	var findings []DoctorFinding

	// Axis 2 (manifest): a command-only thin harness (manifest absent) is an INFO
	// note — the Runner/agent axes apply only when the manifest is declared.
	if e.ManifestMissing {
		return []DoctorFinding{{
			Harness:  e.Name,
			Axis:     "manifest",
			Severity: SeverityInfo,
			Message:  "command-only thin harness (no manifest.json / Runner) — Runner/agent axes not applicable",
		}}
	}

	data, err := os.ReadFile(e.ManifestPath)
	if err != nil {
		return append(findings, DoctorFinding{
			Harness: e.Name, Axis: "manifest", Severity: SeverityError,
			Message: fmt.Sprintf("manifest read failed: %v", err),
		})
	}
	var m v4manifest.Manifest
	if uErr := json.Unmarshal(data, &m); uErr != nil {
		return append(findings, DoctorFinding{
			Harness: e.Name, Axis: "manifest", Severity: SeverityError,
			Message: fmt.Sprintf("manifest JSON invalid: %v", uErr),
		})
	}
	if vErr := v4manifest.Validate(m); vErr != nil {
		findings = append(findings, DoctorFinding{
			Harness: e.Name, Axis: "manifest", Severity: SeverityError,
			Message: fmt.Sprintf("manifest schema invalid: %v", vErr),
		})
	}

	// Axis 3 (runner): the declared Runner exists and its MANIFEST_PATH resolves.
	if m.RunnerWorkflow == "" {
		return append(findings, DoctorFinding{
			Harness: e.Name, Axis: "runner", Severity: SeverityError,
			Message: "manifest.runner_workflow is empty — no Runner declared",
		})
	}
	runnerPath := filepath.Join(projectRoot, v4WorkflowsDir, m.RunnerWorkflow)
	runnerData, rErr := os.ReadFile(runnerPath)
	if rErr != nil {
		return append(findings, DoctorFinding{
			Harness: e.Name, Axis: "runner", Severity: SeverityError,
			Message: fmt.Sprintf("Runner %s not found: %v", m.RunnerWorkflow, rErr),
		})
	}
	runnerSrc := string(runnerData)

	if mp := runnerManifestPathRE.FindStringSubmatch(runnerSrc); mp != nil {
		declared := mp[1]
		if _, sErr := os.Stat(filepath.Join(projectRoot, declared)); sErr != nil {
			findings = append(findings, DoctorFinding{
				Harness: e.Name, Axis: "runner", Severity: SeverityError,
				Message: fmt.Sprintf("Runner MANIFEST_PATH %q does not resolve to an existing file", declared),
			})
		}
	}

	// Axis 4 (agent): each specialist agent referenced by the Runner resolves.
	seen := map[string]bool{}
	for _, ref := range runnerSpecialistRE.FindAllString(runnerSrc, -1) {
		if seen[ref] {
			continue
		}
		seen[ref] = true
		if _, sErr := os.Stat(filepath.Join(projectRoot, v4AgentsDir, ref+".md")); sErr != nil {
			findings = append(findings, DoctorFinding{
				Harness: e.Name, Axis: "agent", Severity: SeverityError,
				Message: fmt.Sprintf("Runner references specialist agent %q but %s does not exist", ref, filepath.Join(v4AgentsDir, ref+".md")),
			})
		}
	}

	return findings
}

// checkPipelineDormancy detects the starved-learning-loop condition
// (SPEC-HARNESS-RATCHET-REWIRE-001 REQ-HRR-008): tier-promotions.jsonl contains
// ≥1 promotion but .moai/harness/proposals/ has never been created. This means
// the classify half turned but the propose→apply back half never ran — the
// pipeline is dormant. The finding is advisory (SeverityInfo); it never fails
// the smoke gate.
func checkPipelineDormancy(projectRoot string) []DoctorFinding {
	promoPath := filepath.Join(projectRoot, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	proposalsDir := filepath.Join(projectRoot, ".moai", "harness", "proposals")

	promoCount, err := countPromotionLines(promoPath)
	if err != nil || promoCount < 1 {
		return nil // no promotions → nothing starved (or unreadable — silent)
	}
	if _, statErr := os.Stat(proposalsDir); statErr == nil {
		return nil // proposals dir exists → back half has run → not dormant
	}
	return []DoctorFinding{{
		Harness:  "(learning-loop)",
		Axis:     "pipeline-dormancy",
		Severity: SeverityInfo,
		Message:  fmt.Sprintf("pipeline dormancy: %d promotion(s) exist in tier-promotions.jsonl but .moai/harness/proposals/ is absent — the propose→apply back half has never run", promoCount),
	}}
}

// countPromotionLines counts non-blank lines in the tier-promotions.jsonl file.
// Returns (0, nil) when the file does not exist (normal first-run state).
func countPromotionLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			count++
		}
	}
	return count, nil
}

// NewHarnessDoctorCmd is the `moai harness doctor` factory (REQ-HEP-005). It runs
// the reference-integrity smoke gate and exits non-zero when >= 1 ERROR finding
// exists. A project with zero harnesses exits 0.
func NewHarnessDoctorCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Reference-integrity smoke gate for v4 harnesses",
		Long: `Run a reference-integrity smoke gate over every v4 harness.

For each harness under .claude/commands/harness/, the checker verifies 4 axes:
  1. entry command file exists
  2. manifest.json exists + schema-valid
  3. Runner (harness-<name>-run.js) exists + its MANIFEST_PATH constant resolves
  4. each specialist agent referenced by the Runner resolves to an existing
     .claude/agents/harness/<name>.md file

Command-only thin harnesses (no manifest / Runner) are reported as INFO notes,
not ERRORs. Exit code is non-zero only when >= 1 ERROR-severity finding exists.
A project with zero harnesses exits 0.

Use --json for machine-readable output.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRootV4(cmd)
			if err != nil {
				return err
			}
			report, err := Doctor(root)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if jsonOutput {
				out, mErr := json.MarshalIndent(report, "", "  ")
				if mErr != nil {
					return fmt.Errorf("harness doctor: json marshal: %w", mErr)
				}
				_, _ = fmt.Fprintln(w, string(out))
			} else if report.Harnesses == 0 {
				_, _ = fmt.Fprintln(w, "No v4 harnesses found under .claude/commands/harness/.")
			} else {
				_, _ = fmt.Fprintf(w, "Scanned %d harness(es): %d ERROR, %d INFO.\n", report.Harnesses, report.ErrorCount, report.InfoCount)
				for _, f := range report.Findings {
					_, _ = fmt.Fprintf(w, "  [%s] %s (%s): %s\n", f.Severity, f.Harness, f.Axis, f.Message)
				}
			}

			if report.ErrorCount > 0 {
				return fmt.Errorf("harness doctor: %d ERROR-severity finding(s)", report.ErrorCount)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	// Local --project-root so the command is usable standalone (mirrors the
	// parent harness command's persistent flag; resolveProjectRootV4 reads it).
	cmd.Flags().String("project-root", "", "project root path (default: current directory)")
	return cmd
}
