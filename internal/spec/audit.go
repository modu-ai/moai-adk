// Package spec — audit engine for SPEC-V3R6-LIFECYCLE-SYNC-GATE-001.
//
// audit.go implements `moai spec audit` core per design.md §B.2 + §A.3.
// Scans .moai/specs/SPEC-*/ directories, classifies each via era.go heuristics,
// and emits drift findings for V3R6 SPECs with cross-tab pattern violations.
//
// Output: AuditResult populated with grandfathered count + modern-era clean count
// + drift_findings slice. JSON schema documented in AC-LSG-007 (spec.md
// frontmatter).
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AuditOptions configures Audit() invocations.
type AuditOptions struct {
	// BaseDir is the root directory containing .moai/specs/ (e.g., project root).
	BaseDir string
	// FilterEra restricts findings to a single era (e.g., "V3R6"). Empty → all eras.
	FilterEra string
	// FilterSpec restricts findings to a single SPEC-ID (exact match on the
	// directory name under .moai/specs/, e.g., "SPEC-V3R6-ORCH-IGGDA-001").
	// Empty → no SPEC-ID filter (all SPECs). Additive to FilterEra: the two MAY
	// compose (filter to one SPEC within one era). When FilterSpec matches no
	// SPEC, the result carries empty drift_findings (graceful, not an error).
	// SPEC-V3R6-ORCH-IGGDA-001 M5.
	FilterSpec string
	// IncludeGrandfathered surfaces V2.x / V3R2-R4 / V3R5 SPECs in findings with
	// severity: INFO (no drift; observational only).
	IncludeGrandfathered bool
	// Strict escalates drift findings to non-zero exit code (consumed by CLI layer).
	Strict bool
}

// AuditResult is the structured audit output. Marshaled to JSON for AC-LSG-007.
type AuditResult struct {
	AuditedAt      time.Time      `json:"audited_at"`
	TotalSpecs     int            `json:"total_specs"`
	Grandfathered  int            `json:"grandfathered"`
	ModernEraClean int            `json:"modern_era_clean"`
	DriftFindings  []DriftFinding `json:"drift_findings"`
	// TokensSpent surfaces the per-SPEC tokens_spent integer parsed from the
	// audited SPEC's progress.md §I Token Accounting section (M4 audit surface,
	// REQ-TA-011/012). Nil when the SPEC has no §I section or no tokens_spent
	// line — never fabricated. Populated when exactly one SPEC is audited (the
	// --filter-spec single-SPEC case); multi-SPEC audits leave it nil because
	// aggregating across SPECs with different confidence qualifiers would be
	// misleading (precision-honesty, spec.md §G anti-pattern).
	TokensSpent *int `json:"tokens_spent,omitempty"`
}

// DriftFinding represents a single audit finding.
type DriftFinding struct {
	SpecID      string         `json:"spec_id"`
	Era         string         `json:"era"`
	FindingType string         `json:"finding_type"` // SyncStatusDrift | EraAutoDetected | AuditError | ...
	Severity    string         `json:"severity"`     // MUST-FIX | INFO
	Remediation string         `json:"remediation,omitempty"`
	Details     map[string]any `json:"details,omitempty"`
}

const (
	// FindingSyncStatusDrift is the single surviving V3R6 drift dimension under the
	// 3-phase lifecycle (SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-019): §E.2 run-evidence
	// + §E.4 sync marker + sync_commit_sha are all present (sync complete) but spec.md
	// status != completed. Re-anchored from the legacy Y_Y_Y_Y_StatusDrift predicate
	// (which keyed on §E.5 + mx_commit_sha) to the 3-marker predicate.
	FindingSyncStatusDrift = "SyncStatusDrift"
	// FindingEraAutoDetected is an INFO finding emitted when frontmatter `era:` field
	// is absent and ClassifyEra inferred the era via heuristics (AC-LSG-013).
	FindingEraAutoDetected = "EraAutoDetected"

	// --- Deprecated aliases (backward compatibility for git-history JSON consumers) ---
	// The three §E.5/mx_commit_sha-keyed findings below are RETIRED under the 3-phase
	// lifecycle (REQ-LR-019). FindingY_Y_Y_Y_StatusDrift is kept as an alias of
	// FindingSyncStatusDrift so older JSON readers still match; FindingY_N_N_Y and
	// FindingY_Y_N_Y retain their string values for historical JSON compatibility but
	// are NO LONGER EMITTED by checkV3R6Drift. See SPEC-V3R6-LIFECYCLE-REDESIGN-001
	// design.md §B.4 for the retirement rationale (the 4-section end-state would
	// otherwise trip Y_N_N_Y MUST-FIX on every non-completed V3R6 SPEC catalog-wide).
	FindingY_Y_Y_Y_StatusDrift = "SyncStatusDrift" // alias of FindingSyncStatusDrift (retired predicate name)
	FindingY_N_N_Y             = "Y_N_N_Y"         // retired — no longer emitted
	FindingY_Y_N_Y             = "Y_Y_N_Y"         // retired — no longer emitted
)

// specStatusPattern extracts `status:` field from spec.md frontmatter.
var specStatusPattern = regexp.MustCompile(`(?m)^status:\s*(.+?)\s*$`)

// sectionIHeading is the canonical progress.md §I Token Accounting heading.
// It MUST match the literal produced by the M3 §I writer
// (internal/tokenusage.SectionIHeading). A local constant is used instead of a
// cross-package import because (a) no import cycle exists between internal/spec
// and internal/tokenusage, but (b) keeping the parser self-contained avoids
// coupling the audit engine's read path to the writer's package; the two
// literals are paired by a round-trip test contract (the M4 test fixture
// mirrors BuildSectionI output verbatim). If the heading ever changes in M3,
// this constant MUST be updated in lockstep.
const sectionIHeading = "## §I Token Accounting"

// parseTokensSpentFromSectionI extracts the integer value of the
// `- tokens_spent:` field from the `## §I Token Accounting` section of a
// progress.md body. Returns nil when the section is absent, the field line is
// absent, or the value is not a base-10 integer — never fabricates a value
// (REQ-TA-012 precision-honesty: absent evidence is reported as nil, not zero).
//
// The section span is the §I heading line through the next top-level (`## `)
// heading or EOF, matching the M3 writer's applySectionI span contract so the
// parser and writer agree on section boundaries.
func parseTokensSpentFromSectionI(progressContent string) *int {
	if progressContent == "" {
		return nil
	}
	lines := strings.Split(progressContent, "\n")
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			if inSection {
				break // reached the next top-level heading — §I section ended
			}
			if trimmed == sectionIHeading {
				inSection = true
			}
			continue
		}
		if !inSection {
			continue
		}
		const prefix = "- tokens_spent:"
		if strings.HasPrefix(trimmed, prefix) {
			valStr := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			n, err := strconv.Atoi(valStr)
			if err != nil {
				return nil // malformed value — treat as absent, do not fabricate
			}
			return &n
		}
	}
	return nil
}

// Audit scans .moai/specs/SPEC-*/ under opts.BaseDir, classifies each SPEC by
// era, and emits drift findings for V3R6 SPECs with cross-tab pattern violations.
//
// V2.x / V3R2-R4 / V3R5 SPECs are grandfather-clause-protected (AC-LSG-017):
// they are counted in Grandfathered but DriftFindings excludes them by default.
// Use IncludeGrandfathered to surface them as INFO findings.
func Audit(opts AuditOptions) (*AuditResult, error) {
	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	specsDir := filepath.Join(baseDir, ".moai", "specs")

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No specs directory yet — return empty result, not an error.
			return &AuditResult{
				AuditedAt:     time.Now().UTC(),
				DriftFindings: []DriftFinding{},
			}, nil
		}
		return nil, fmt.Errorf("read specs dir %s: %w", specsDir, err)
	}

	result := &AuditResult{
		AuditedAt:     time.Now().UTC(),
		DriftFindings: []DriftFinding{},
	}

	// Sort entries for deterministic output (test determinism).
	var specDirs []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "SPEC-") {
			specDirs = append(specDirs, e.Name())
		}
	}
	sort.Strings(specDirs)

	// M4 audit surface (REQ-TA-011/012): track the parsed §I tokens_spent so
	// AuditResult can surface it for the single-SPEC / --filter-spec case.
	// Multi-SPEC audits leave TokensSpent nil (aggregating across SPECs with
	// different confidence qualifiers would mislead — precision-honesty).
	auditedCount := 0
	var singleSpecTokensSpent *int

	for _, specName := range specDirs {
		// Apply SPEC-ID filter (SPEC-V3R6-ORCH-IGGDA-001 M5) at the TOP of the
		// loop — this gates the AuditError branch too (avoids leaking AuditError
		// findings for specs outside the filter). Additive to FilterEra (applied
		// below after classification, since era is only known post-auditSpec).
		// Empty FilterSpec = no filter.
		if opts.FilterSpec != "" && specName != opts.FilterSpec {
			continue
		}

		specDir := filepath.Join(specsDir, specName)
		findings, classified, tokensSpent, err := auditSpec(specDir, specName, opts)
		if err != nil {
			// Surface per-spec errors as findings with FindingType: "AuditError"
			// rather than aborting the entire run.
			result.DriftFindings = append(result.DriftFindings, DriftFinding{
				SpecID:      specName,
				FindingType: "AuditError",
				Severity:    "INFO",
				Details:     map[string]any{"error": err.Error()},
			})
			continue
		}

		result.TotalSpecs++
		auditedCount++
		singleSpecTokensSpent = tokensSpent // last-write wins; only surfaced when auditedCount==1
		if classified.EraFinal() {
			result.Grandfathered++
		} else if classified == EraV3R6 {
			// V3R6 is "clean" when there are no MUST-FIX drift findings.
			// INFO findings (EraAutoDetected, Grandfathered) do not disqualify.
			hasMustFix := false
			for _, f := range findings {
				if f.Severity == "MUST-FIX" {
					hasMustFix = true
					break
				}
			}
			if !hasMustFix {
				result.ModernEraClean++
			}
		}

		// Apply era filter (SPEC-ID filter applied at top of loop — M5).
		if opts.FilterEra != "" && string(classified) != opts.FilterEra {
			continue
		}

		result.DriftFindings = append(result.DriftFindings, findings...)
	}

	// M4: surface tokens_spent only when exactly one SPEC was audited (the
	// --filter-spec case and the single-fixture case). For multi-SPEC audits
	// the field stays nil — a per-SPEC breakdown is a separate concern and
	// summing across SPECs would fabricate a misleading aggregate.
	if auditedCount == 1 {
		result.TokensSpent = singleSpecTokensSpent
	}

	// SPEC-OBSERVE-HYGIENE-001 M1 (REQ-OBH-001): consume the
	// status-transition-audit.log as a write-site cross-check. The log records
	// status values captured at Write/Edit time — a signal the per-SPEC git
	// history scan above cannot see. Absent/corrupt log degrades to zero
	// findings (graceful, never an error per EC-1).
	result.DriftFindings = append(result.DriftFindings, crossCheckTransitionLog(baseDir, opts)...)

	return result, nil
}

// auditSpec audits a single SPEC directory and returns (findings, classifiedEra,
// tokensSpent, error). tokensSpent is the parsed §I value (nil when the SPEC
// has no populated §I section — never fabricated; surfaced by Audit() on the
// aggregate AuditResult for the single-SPEC / --filter-spec case).
func auditSpec(specDir, specID string, opts AuditOptions) ([]DriftFinding, Era, *int, error) {
	signals, err := LoadEraSignalsFromDir(specDir)
	if err != nil {
		return nil, EraUnclassified, nil, fmt.Errorf("load era signals: %w", err)
	}
	tokensSpent := parseTokensSpentFromSectionI(signals.ProgressMDContent)

	era, heuristic := ClassifyEra(signals)
	var findings []DriftFinding

	// AC-LSG-013 — EraAutoDetected INFO finding when frontmatter era was absent
	// and classification was performed via heuristics.
	if signals.FrontmatterEra == "" && era != EraUnclassified {
		findings = append(findings, DriftFinding{
			SpecID:      specID,
			Era:         string(era),
			FindingType: FindingEraAutoDetected,
			Severity:    "INFO",
			Details: map[string]any{
				"heuristic_matched": heuristic,
			},
		})
	}

	// AC-LSG-017 — grandfather-clause-protected eras emit no MUST-FIX findings.
	if era.EraFinal() {
		if opts.IncludeGrandfathered {
			findings = append(findings, DriftFinding{
				SpecID:      specID,
				Era:         string(era),
				FindingType: "Grandfathered",
				Severity:    "INFO",
				Details: map[string]any{
					"reason": "pre-V3R6 era — grandfather clause active",
				},
			})
		}
		return findings, era, tokensSpent, nil
	}

	// EraUnclassified — emit INFO finding for visibility but no MUST-FIX action.
	if era == EraUnclassified {
		findings = append(findings, DriftFinding{
			SpecID:      specID,
			Era:         string(era),
			FindingType: "EraUnclassified",
			Severity:    "INFO",
			Details: map[string]any{
				"heuristic_matched": heuristic,
				"reason":            "no era heuristic matched; consider explicit era: field",
			},
		})
		return findings, era, tokensSpent, nil
	}

	// V3R6 — check cross-tab pattern for drift
	driftFinding := checkV3R6Drift(specDir, specID, signals)
	if driftFinding != nil {
		findings = append(findings, *driftFinding)
	}

	return findings, era, tokensSpent, nil
}

// checkV3R6Drift performs the V3R6 status-drift detection under the 3-phase
// lifecycle (SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-019).
//
// The single surviving drift dimension is SyncStatusDrift: §E.2 run-evidence +
// §E.4 sync marker + sync_commit_sha are all present (sync phase complete) but
// spec.md status != completed. The two former §E.5/mx_commit_sha-keyed findings
// (Y_N_N_Y, Y_Y_N_Y) are RETIRED — under the mandated 4-section end-state
// (§E.5 absent), Y_N_N_Y would otherwise fire MUST-FIX on every non-completed
// V3R6 SPEC catalog-wide (the D2 false-positive storm).
//
// Returns nil when no drift detected (clean V3R6 SPEC).
func checkV3R6Drift(specDir, specID string, signals EraSignals) *DriftFinding {
	// Parse spec.md status
	specMDPath := filepath.Join(specDir, "spec.md")
	specContent, err := os.ReadFile(specMDPath)
	if err != nil {
		return nil
	}
	statusMatch := specStatusPattern.FindStringSubmatch(string(specContent))
	if len(statusMatch) < 2 {
		return nil // no status field — skip
	}
	specStatus := strings.TrimSpace(statusMatch[1])

	hasRunEvidence := hasProgressMarker(signals.ProgressMDContent, "§E.2")
	hasSyncMarker := hasProgressMarker(signals.ProgressMDContent, "§E.4")
	syncSHA := extractProgressField(signals.ProgressMDContent, "sync_commit_sha")

	// If status is already completed, no drift.
	if specStatus == "completed" {
		return nil
	}
	// Terminal states (superseded / archived / rejected) — no drift.
	if specStatus == "superseded" || specStatus == "archived" || specStatus == "rejected" {
		return nil
	}

	// SyncStatusDrift: §E.2 run-evidence + §E.4 sync marker + sync_commit_sha present
	// (sync phase complete) but spec.md status != completed. This is the re-anchored
	// successor to the legacy Y_Y_Y_Y_StatusDrift predicate (REQ-LR-019).
	if hasRunEvidence && hasSyncMarker && syncSHA != "" {
		return &DriftFinding{
			SpecID:      specID,
			Era:         string(EraV3R6),
			FindingType: FindingSyncStatusDrift,
			Severity:    "MUST-FIX",
			Remediation: fmt.Sprintf("moai spec close %s --backfill-only", specID),
			Details: map[string]any{
				"spec_md_status":  specStatus,
				"sync_commit_sha": syncSHA,
				"reason":          "§E.2 + §E.4 + sync_commit_sha present (sync complete) but status != completed",
			},
		}
	}

	return nil
}
