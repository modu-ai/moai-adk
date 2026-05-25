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
	"strings"
	"time"
)

// AuditOptions configures Audit() invocations.
type AuditOptions struct {
	// BaseDir is the root directory containing .moai/specs/ (e.g., project root).
	BaseDir string
	// FilterEra restricts findings to a single era (e.g., "V3R6"). Empty → all eras.
	FilterEra string
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
}

// DriftFinding represents a single audit finding.
type DriftFinding struct {
	SpecID      string         `json:"spec_id"`
	Era         string         `json:"era"`
	FindingType string         `json:"finding_type"` // Y_N_N_Y | Y_Y_N_Y | Y_Y_Y_Y_StatusDrift | EraAutoDetected
	Severity    string         `json:"severity"`     // MUST-FIX | INFO
	Remediation string         `json:"remediation,omitempty"`
	Details     map[string]any `json:"details,omitempty"`
}

const (
	// FindingY_N_N_Y indicates sync section present but mx absent + status drift.
	FindingY_N_N_Y = "Y_N_N_Y"
	// FindingY_Y_N_Y indicates §E.2 + §E.5 present but mx_commit_sha missing + status drift.
	FindingY_Y_N_Y = "Y_Y_N_Y"
	// FindingY_Y_Y_Y_StatusDrift indicates all 4 phase markers present + valid SHAs
	// but spec.md status != completed (the modern-era violation pattern this SPEC fixes).
	FindingY_Y_Y_Y_StatusDrift = "Y_Y_Y_Y_StatusDrift"
	// FindingEraAutoDetected is an INFO finding emitted when frontmatter `era:` field
	// is absent and ClassifyEra inferred the era via heuristics (AC-LSG-013).
	FindingEraAutoDetected = "EraAutoDetected"
)

// specStatusPattern extracts `status:` field from spec.md frontmatter.
var specStatusPattern = regexp.MustCompile(`(?m)^status:\s*(.+?)\s*$`)

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

	for _, specName := range specDirs {
		specDir := filepath.Join(specsDir, specName)
		findings, classified, err := auditSpec(specDir, specName, opts)
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

		// Apply era filter
		if opts.FilterEra != "" && string(classified) != opts.FilterEra {
			continue
		}

		result.DriftFindings = append(result.DriftFindings, findings...)
	}

	return result, nil
}

// auditSpec audits a single SPEC directory and returns (findings, classifiedEra, error).
func auditSpec(specDir, specID string, opts AuditOptions) ([]DriftFinding, Era, error) {
	signals, err := LoadEraSignalsFromDir(specDir)
	if err != nil {
		return nil, EraUnclassified, fmt.Errorf("load era signals: %w", err)
	}

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
		return findings, era, nil
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
		return findings, era, nil
	}

	// V3R6 — check cross-tab pattern for drift
	driftFinding := checkV3R6Drift(specDir, specID, signals)
	if driftFinding != nil {
		findings = append(findings, *driftFinding)
	}

	return findings, era, nil
}

// checkV3R6Drift performs the V3R6 cross-tab status drift detection per AC-LSG-009.
// Detects three drift patterns:
//   - Y_N_N_Y: spec.md status != completed AND sync present but mx absent
//   - Y_Y_N_Y: §E.2 + §E.5 present but mx_commit_sha missing + status drift
//   - Y_Y_Y_Y_StatusDrift: all phase markers + SHAs present but status != completed
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

	hasSyncSection := hasProgressMarker(signals.ProgressMDContent, "§E.2")
	hasMxSection := hasProgressMarker(signals.ProgressMDContent, "§E.5")
	syncSHA := extractProgressField(signals.ProgressMDContent, "sync_commit_sha")
	mxSHA := extractProgressField(signals.ProgressMDContent, "mx_commit_sha")

	// If status is already completed, no drift.
	if specStatus == "completed" {
		return nil
	}
	// Terminal states (superseded / archived / rejected) — no drift.
	if specStatus == "superseded" || specStatus == "archived" || specStatus == "rejected" {
		return nil
	}

	// AC-LSG-009 — Y_Y_Y_Y_StatusDrift: all 4 phase markers + valid SHAs but status != completed.
	if hasSyncSection && hasMxSection && syncSHA != "" && mxSHA != "" {
		return &DriftFinding{
			SpecID:      specID,
			Era:         string(EraV3R6),
			FindingType: FindingY_Y_Y_Y_StatusDrift,
			Severity:    "MUST-FIX",
			Remediation: fmt.Sprintf("moai spec close %s --backfill-only", specID),
			Details: map[string]any{
				"spec_md_status":   specStatus,
				"sync_commit_sha":  syncSHA,
				"mx_commit_sha":    mxSHA,
				"reason":           "all phase markers present + valid SHAs but status != completed",
			},
		}
	}

	// Y_Y_N_Y: §E.2 + §E.5 present but mx_commit_sha missing + status drift.
	if hasSyncSection && hasMxSection && mxSHA == "" {
		return &DriftFinding{
			SpecID:      specID,
			Era:         string(EraV3R6),
			FindingType: FindingY_Y_N_Y,
			Severity:    "MUST-FIX",
			Remediation: fmt.Sprintf("moai spec close %s --backfill-only", specID),
			Details: map[string]any{
				"spec_md_status":  specStatus,
				"sync_commit_sha": syncSHA,
				"reason":          "§E.5 section present but mx_commit_sha missing",
			},
		}
	}

	// Y_N_N_Y: sync section present but mx absent + status drift.
	if hasSyncSection && !hasMxSection {
		return &DriftFinding{
			SpecID:      specID,
			Era:         string(EraV3R6),
			FindingType: FindingY_N_N_Y,
			Severity:    "MUST-FIX",
			Remediation: fmt.Sprintf("moai spec close %s --backfill-only", specID),
			Details: map[string]any{
				"spec_md_status": specStatus,
				"reason":         "§E.2 sync section present but §E.5 Mx section absent",
			},
		}
	}

	return nil
}
