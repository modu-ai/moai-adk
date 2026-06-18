// Package spec — era classification engine per SPEC-V3R6-LIFECYCLE-SYNC-GATE-001.
//
// era.go implements the era classification heuristic table documented in
// .moai/specs/SPEC-V3R6-LIFECYCLE-SYNC-GATE-001/design.md §C.2 (H-1..H-6).
//
// Era taxonomy (5 buckets per AC-LSG-002):
//   - V2.x       Pre-2026-02 — no progress.md
//   - V3R2-R4    2026-02 ~ 2026-03 — progress.md introduced; no sync_commit_sha
//   - V3R5       2026-03 ~ 2026-04 — sync section emerges; sync_commit_sha not enforced
//   - V3R6       2026-04 ~ present — 3-phase modern standard (plan/run/sync)
//   - unclassified — auto-detection ambiguous (H-6 fallback)
//
// Grandfather clause (design §C.3): V2.x / V3R2-R4 / V3R5 SPECs are protected
// from drift findings. Only V3R6 SPECs are subject to lifecycle invariants.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Era classifies a SPEC document by lifecycle standard.
type Era string

const (
	// EraV2x — pre-V3 SPECs without progress.md (H-1)
	EraV2x Era = "V2.x"
	// EraV3R2R4 — early-V3 SPECs with progress.md but no §E.* markers (H-2)
	EraV3R2R4 Era = "V3R2-R4"
	// EraV3R5 — V3R5 SPECs that have the §E.2 run-evidence start marker but are
	// missing sync_commit_sha (H-3). Classification is string-presence-based: the
	// var name hasSyncSection is a misnomer — §E.2 marks the §E-section progress
	// structure START (run-evidence), not the sync phase (sync lives at §E.4).
	EraV3R5 Era = "V3R5"
	// EraV3R6 — V3R6 SPECs with §E.2 + §E.4 + sync_commit_sha (new H-4, REQ-LR-005),
	// or via the legacy §E.2 + §E.5 + both commit_sha predicate during the migration
	// window (H-4-legacy, REQ-LR-006).
	EraV3R6 Era = "V3R6"
	// EraUnclassified — ambiguous, no heuristic matched (H-6)
	EraUnclassified Era = "unclassified"
)

// EraFinal returns true when the era is protected by the grandfather clause
// (V2.x / V3R2-R4 / V3R5). V3R6 SPECs are NOT grandfather-protected — they are
// subject to drift findings per AC-LSG-009.
func (e Era) EraFinal() bool {
	switch e {
	case EraV2x, EraV3R2R4, EraV3R5:
		return true
	default:
		return false
	}
}

// IsModern returns true when the era is V3R6 (subject to modern-era lifecycle
// invariants and drift detection).
func (e Era) IsModern() bool {
	return e == EraV3R6
}

// EraSignals captures the observable artifacts the heuristic table inspects.
// All fields are read-only inputs to ClassifyEra; the struct does not perform
// I/O itself — callers populate from disk.
type EraSignals struct {
	// FrontmatterEra is the optional `era:` field from spec.md frontmatter.
	// If non-empty, it overrides auto-detection (design §C.2 Override).
	FrontmatterEra string

	// ProgressMDExists indicates whether `progress.md` is present in the SPEC dir.
	ProgressMDExists bool

	// ProgressMDContent is the raw progress.md content (empty if absent).
	// Used to detect §E.* section markers and sync/mx commit_sha fields.
	ProgressMDContent string

	// FrontmatterPhase is the optional `phase:` field from spec.md frontmatter
	// (used by H-5 tie-breaker; e.g., "v3.0.0", "v3R6").
	FrontmatterPhase string

	// FrontmatterCreated is the optional `created:` field (YYYY-MM-DD)
	// used by H-5 date-based tie-breaker (>= 2026-04-01 → V3R6).
	FrontmatterCreated string
}

// ClassifyEra applies the H-1..H-6 heuristic table from design §C.2 to derive
// the era of a SPEC document. It is the SSOT for era classification logic.
//
// Heuristic order (first match wins, except FrontmatterEra override):
//
//	H-override: FrontmatterEra non-empty + valid → returned verbatim
//	H-1:        ProgressMDExists == false → V2.x
//	H-2:        progress.md present but no §E.{2,3,4,5} markers → V3R2-R4
//	H-3:        §E.2 run-evidence start marker present but sync_commit_sha empty/missing → V3R5
//	H-4:        §E.2 + §E.4 present AND sync_commit_sha non-empty → V3R6 (new H-4, REQ-LR-005)
//	H-4-legacy: §E.2 + §E.5 present AND sync_commit_sha + mx_commit_sha non-empty → V3R6
//	            (REQ-LR-006 dual-predicate migration window — legacy 5-section layout)
//	H-5:        H-4 ambiguous + (phase ~ v3.0|v3R6 OR created >= 2026-04-01) → V3R6
//	H-6:        no heuristic matched → unclassified
//
// The new H-4 predicate (§E.2 + §E.4 + sync_commit_sha) reflects the 3-phase
// lifecycle restoration (SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-005): §E.4 is
// the sync-phase marker, and sync_commit_sha is the sole required commit SHA.
// The legacy H-4-legacy fallback (§E.2 + §E.5 + both SHAs) preserves V3R6
// classification for SPECs still carrying the pre-redesign 5-section layout
// during the migration window (REQ-LR-006); it is defense-in-depth plus
// classification-rationale precision (the re-derived H-6 at-risk set is empty
// for the current catalog — every V3R6 SPEC is caught by H-5 even without the
// window — but an explicit predicate is a stronger signal than the H-5 date
// heuristic, and the catalog is moving).
//
// Detection is string-presence-based (hasProgressMarker): hasSyncSection tests
// the §E.2 run-evidence start marker (the var name is a historical misnomer
// retained for call-site stability — §E.2 marks run-evidence, not sync);
// hasSyncMarker tests §E.4 (the actual sync phase); hasMxSection tests §E.5
// (the legacy Mx-completion marker).
func ClassifyEra(signals EraSignals) (Era, string) {
	// H-override: explicit frontmatter `era:` field wins
	if signals.FrontmatterEra != "" {
		if era, ok := normalizeEra(signals.FrontmatterEra); ok {
			return era, "H-override (frontmatter era field)"
		}
		// Invalid value — fall through to auto-detect (do not silently accept)
	}

	// H-1: progress.md absent → V2.x
	if !signals.ProgressMDExists {
		return EraV2x, "H-1 (progress.md absent)"
	}

	// Parse progress.md signals.
	// hasSyncSection is a historical misnomer: it tests §E.2 (run-evidence start),
	// not the sync phase (which lives at §E.4 — tested by hasSyncMarker below).
	hasSyncSection := hasProgressMarker(signals.ProgressMDContent, "§E.2")
	hasSyncMarker := hasProgressMarker(signals.ProgressMDContent, "§E.4")
	hasMxSection := hasProgressMarker(signals.ProgressMDContent, "§E.5")
	syncSHA := extractProgressField(signals.ProgressMDContent, "sync_commit_sha")
	mxSHA := extractProgressField(signals.ProgressMDContent, "mx_commit_sha")

	// H-2: progress.md present but no §E.{2,3,4,5} markers → V3R2-R4
	if !hasAnyProgressMarker(signals.ProgressMDContent) {
		return EraV3R2R4, "H-2 (progress.md without §E.* markers)"
	}

	// H-3: §E.2 run-evidence start marker present but sync_commit_sha empty/missing → V3R5
	// (hasSyncSection tests literal §E.2 string presence — the run-evidence start
	// marker — not the sync phase, which lives at §E.4.)
	if hasSyncSection && syncSHA == "" {
		return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"
	}

	// H-4 (new H-4, REQ-LR-005): §E.2 run-evidence + §E.4 sync marker + sync_commit_sha → V3R6.
	// This is the canonical 3-phase predicate (plan/run/sync); §E.5 + mx_commit_sha
	// are no longer required.
	if hasSyncSection && hasSyncMarker && syncSHA != "" {
		return EraV3R6, "H-4 (§E.2 + §E.4 + sync_commit_sha)"
	}

	// H-4-legacy (REQ-LR-006 dual-predicate migration window): SPECs authored before
	// the redesign still carry §E.5 + mx_commit_sha. Treat them as V3R6 during the
	// migration window so they classify via an explicit predicate (precise rationale)
	// rather than falling through to the weaker H-5 date heuristic. The window is
	// defense-in-depth; the re-derived H-6 at-risk set is empty for the current catalog.
	if hasSyncSection && hasMxSection && syncSHA != "" && mxSHA != "" {
		return EraV3R6, "H-4-legacy (§E.2 + §E.5 + both commit_sha — migration window)"
	}

	// H-5: tie-breaker via phase or created date
	if matchesModernPhase(signals.FrontmatterPhase) ||
		isAfterModernThreshold(signals.FrontmatterCreated) {
		return EraV3R6, "H-5 (modern phase or created date)"
	}

	// H-6: no match
	return EraUnclassified, "H-6 (no heuristic matched)"
}

// normalizeEra normalizes a user-supplied era string to a canonical Era value.
// Returns (era, true) on match, (EraUnclassified, false) otherwise.
func normalizeEra(raw string) (Era, bool) {
	switch strings.TrimSpace(raw) {
	case "V2.x", "v2.x", "V2", "v2":
		return EraV2x, true
	case "V3R2-R4", "v3r2-r4", "V3R2", "V3R3", "V3R4":
		return EraV3R2R4, true
	case "V3R5", "v3r5":
		return EraV3R5, true
	case "V3R6", "v3r6":
		return EraV3R6, true
	case "unclassified":
		return EraUnclassified, true
	}
	return EraUnclassified, false
}

// hasProgressMarker reports whether the given §E.N marker appears in content.
// Match is heading-style: "## §E.2" or "### §E.2" etc.
func hasProgressMarker(content, marker string) bool {
	return strings.Contains(content, marker)
}

// hasAnyProgressMarker reports whether any §E.{2,3,4,5} section header appears.
func hasAnyProgressMarker(content string) bool {
	return hasProgressMarker(content, "§E.2") ||
		hasProgressMarker(content, "§E.3") ||
		hasProgressMarker(content, "§E.4") ||
		hasProgressMarker(content, "§E.5")
}

// extractProgressField extracts the value of a `field: value` pair from
// progress.md body. Returns the trimmed value or empty string.
// Recognizes both YAML-style (sync_commit_sha: abc123) and markdown-style
// (- `sync_commit_sha`: abc123) patterns.
func extractProgressField(content, field string) string {
	// Pattern 1: YAML-style at line start
	pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(field) + `\s*:\s*(.+?)\s*$`)
	if m := pattern.FindStringSubmatch(content); len(m) > 1 {
		return cleanFieldValue(m[1])
	}
	// Pattern 2: markdown list with backticks
	pattern2 := regexp.MustCompile("(?m)^\\s*[-*]\\s*`?" + regexp.QuoteMeta(field) + "`?\\s*:\\s*(.+?)\\s*$")
	if m := pattern2.FindStringSubmatch(content); len(m) > 1 {
		return cleanFieldValue(m[1])
	}
	return ""
}

// cleanFieldValue strips empty placeholders (null, none, "", ``) and returns
// only non-trivial values (typically a git SHA or quoted string).
func cleanFieldValue(raw string) string {
	v := strings.TrimSpace(raw)
	v = strings.Trim(v, "`")
	v = strings.Trim(v, `"`)
	v = strings.Trim(v, `'`)
	switch strings.ToLower(v) {
	case "", "null", "none", "tbd", "<pending>", "pending":
		return ""
	}
	return v
}

// matchesModernPhase reports whether phase string indicates V3R6 era.
func matchesModernPhase(phase string) bool {
	p := strings.ToLower(strings.TrimSpace(phase))
	if p == "" {
		return false
	}
	// Match v3.0.x, v3R6, v3r6 etc.
	if strings.Contains(p, "v3r6") {
		return true
	}
	// v3.0.* phase tag
	if strings.HasPrefix(p, "v3.0") || strings.HasPrefix(p, `"v3.0`) {
		return true
	}
	return false
}

// modernEraThreshold is the cutoff date for H-5 created-date heuristic.
const modernEraThreshold = "2026-04-01"

// isAfterModernThreshold reports whether created date is on/after modernEraThreshold.
// Compares ISO-8601 YYYY-MM-DD lexicographically (valid since both sides use 4-2-2).
func isAfterModernThreshold(created string) bool {
	c := strings.TrimSpace(created)
	if c == "" {
		return false
	}
	// Strict ISO-8601 lexicographic comparison
	return c >= modernEraThreshold
}

// LoadEraSignalsFromDir reads spec.md frontmatter + progress.md content from
// a SPEC directory and returns populated EraSignals. Used by audit.go.
//
// If the SPEC directory is missing required files, returns EraSignals with
// the best-effort populated subset (e.g., ProgressMDExists=false).
func LoadEraSignalsFromDir(specDir string) (EraSignals, error) {
	signals := EraSignals{}

	specMDPath := filepath.Join(specDir, "spec.md")
	specContent, err := os.ReadFile(specMDPath)
	if err != nil {
		return signals, fmt.Errorf("read spec.md: %w", err)
	}

	fm, _, fmErr := extractFrontmatter(string(specContent))
	if fmErr == nil {
		signals.FrontmatterPhase = fm.Phase
		signals.FrontmatterCreated = fm.Created
		// Era field is optional — added by SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
		signals.FrontmatterEra = fm.Era
	}

	progressMDPath := filepath.Join(specDir, "progress.md")
	if progressContent, perr := os.ReadFile(progressMDPath); perr == nil {
		signals.ProgressMDExists = true
		signals.ProgressMDContent = string(progressContent)
	}

	return signals, nil
}
