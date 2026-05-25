// Package cli — v2_detection.go
//
// Implements the v2 fingerprint heuristic for SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
// (REQ-VVCR-001..002, AC-VVCR-001). The detector inspects three independent
// signals and resolves the IsV2 boolean as the disjunction: ANY positive
// signal ⇒ IsV2 true. This drives the clean-reinstall code path in
// `runUpdate` (see update.go integration in M5).
//
// Signal sources:
//
//   - Signal 1 (V2DetectedViaVersion): `.moai/config/sections/system.yaml`
//     `moai.version` field. Positive when the version string starts with
//     "v2.", OR is empty, OR the system.yaml file is missing entirely.
//     Empty / missing branches reflect Option α (broader detection): v3
//     projects always carry system.yaml with populated moai.version, so
//     drift / absence is a positive v2 signal.
//
//   - Signal 2 (V2DetectedViaAgencyDir): existence of `.agency/` legacy
//     directory at project root. v.2.x exclusive artifact.
//
//   - Signal 3 (V2DetectedViaDeprecatedPath): existence of ANY path
//     enumerated in defs.DeprecatedPaths (43 entries: Category A 9 +
//     Category B 31 + Category C 3 per spec.md §A.4).
//
// The SignalDetails map carries per-signal diagnostic strings used by
// telemetry and `--dry-run` output (REQ-VVCR-028 / REQ-VVCR-029).

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// V2Fingerprint is the result of detectV2Fingerprint.
//
// @MX:ANCHOR: Output contract consumed by runCleanReinstall (M4) and
// runUpdate (M5) integration.
// @MX:REASON: Field set mirrors AC-VVCR-001 verification expectations;
// modifications MUST update both this struct and the acceptance.md AC entry.
type V2Fingerprint struct {
	// IsV2 is true when ANY of the 3 signals is positive (disjunction).
	IsV2 bool

	// Per-signal positive flags (used by telemetry).
	V2DetectedViaVersion        bool
	V2DetectedViaAgencyDir      bool
	V2DetectedViaDeprecatedPath bool

	// SignalDetails carries per-signal diagnostic strings. Keys:
	//   "version_signal"             — what triggered Signal 1
	//   "agency_signal"              — what triggered Signal 2
	//   "deprecated_signal_first_hit" — first DeprecatedPaths entry that hit
	SignalDetails map[string]string
}

// detectV2Fingerprint inspects projectRoot and returns a V2Fingerprint
// describing which v2 signals are positive.
//
// The function is read-only: it MUST NOT modify any file or directory.
// It returns an error only when projectRoot itself is unreadable (e.g.,
// the path does not exist); individual signal probes that encounter
// fs.ErrNotExist treat the missing artifact as the appropriate signal
// outcome (e.g., missing system.yaml → positive Signal 1 per Option α;
// missing .agency/ → negative Signal 2).
func detectV2Fingerprint(projectRoot string) (V2Fingerprint, error) {
	if projectRoot == "" {
		return V2Fingerprint{}, errors.New("detectV2Fingerprint: empty projectRoot")
	}

	// Verify the root itself exists. We do not require it to be a directory
	// strictly — Stat is sufficient to surface a clear error before
	// individual signal probes run.
	if _, err := os.Stat(projectRoot); err != nil {
		return V2Fingerprint{}, fmt.Errorf("detectV2Fingerprint: stat projectRoot: %w", err)
	}

	fp := V2Fingerprint{
		SignalDetails: make(map[string]string, 3),
	}

	// ---------------------------------------------------------------
	// Signal 1: system.yaml moai.version reading
	// ---------------------------------------------------------------
	versionPositive, versionDetail := probeVersionSignal(projectRoot)
	fp.V2DetectedViaVersion = versionPositive
	if versionDetail != "" {
		fp.SignalDetails["version_signal"] = versionDetail
	}

	// ---------------------------------------------------------------
	// Signal 2: .agency/ legacy directory presence
	// ---------------------------------------------------------------
	agencyPositive, agencyDetail := probeAgencyDirSignal(projectRoot)
	fp.V2DetectedViaAgencyDir = agencyPositive
	if agencyDetail != "" {
		fp.SignalDetails["agency_signal"] = agencyDetail
	}

	// ---------------------------------------------------------------
	// Signal 3: DeprecatedPaths enumeration
	// ---------------------------------------------------------------
	deprecatedPositive, deprecatedDetail := probeDeprecatedPathSignal(projectRoot)
	fp.V2DetectedViaDeprecatedPath = deprecatedPositive
	if deprecatedDetail != "" {
		fp.SignalDetails["deprecated_signal_first_hit"] = deprecatedDetail
	}

	// Aggregation: any one positive ⇒ IsV2 true.
	fp.IsV2 = fp.V2DetectedViaVersion ||
		fp.V2DetectedViaAgencyDir ||
		fp.V2DetectedViaDeprecatedPath

	return fp, nil
}

// systemYAMLMoaiBlock is the minimal subset of system.yaml needed for
// Signal 1. We unmarshal only the `moai.version` field to keep the
// detector decoupled from the full system.yaml schema.
type systemYAMLMoaiBlock struct {
	Moai struct {
		Version string `yaml:"version"`
	} `yaml:"moai"`
}

// probeVersionSignal returns (positive, detail).
//
// Per AC-VVCR-001 Option α:
//   - file missing       → positive, detail "system.yaml missing"
//   - empty version      → positive, detail "moai.version empty"
//   - v2.* prefix         → positive, detail "moai.version starts with v2."
//   - other (e.g. v3.*)  → negative, detail empty
//
// Parse errors are treated as positive Signal 1 with a descriptive detail —
// a malformed system.yaml in a project running `moai update` is more likely
// to be a partial v2 migration than a deliberately corrupted v3 file.
func probeVersionSignal(projectRoot string) (bool, string) {
	sysYAMLPath := filepath.Join(projectRoot,
		".moai", "config", "sections", "system.yaml")

	data, err := os.ReadFile(sysYAMLPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, "system.yaml missing"
		}
		// Other read errors (permission denied, etc.) — treat as positive
		// to err on the side of clean-reinstall over silent skip.
		return true, fmt.Sprintf("system.yaml read error: %v", err)
	}

	var block systemYAMLMoaiBlock
	if err := yaml.Unmarshal(data, &block); err != nil {
		return true, fmt.Sprintf("system.yaml parse error: %v", err)
	}

	v := strings.TrimSpace(block.Moai.Version)
	switch {
	case v == "":
		return true, "moai.version empty"
	case strings.HasPrefix(v, "v2."):
		return true, fmt.Sprintf("moai.version starts with v2. (%s)", v)
	default:
		// v3.* or any other non-v2 value → negative.
		return false, ""
	}
}

// probeAgencyDirSignal returns (positive, detail).
func probeAgencyDirSignal(projectRoot string) (bool, string) {
	agencyDir := filepath.Join(projectRoot, ".agency")
	info, err := os.Stat(agencyDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, ""
		}
		// Stat error other than not-exist — treat as negative; an
		// inaccessible .agency/ is unusual but not a v2 indicator.
		return false, ""
	}
	if !info.IsDir() {
		// A file named .agency at the root is not the v2 directory.
		return false, ""
	}
	return true, ".agency/ present at project root"
}

// probeDeprecatedPathSignal returns (positive, first-hit-path).
//
// Iterates defs.DeprecatedPaths and returns positive as soon as the first
// hit is observed. The iteration order matches the declaration order in
// defs/dirs.go so the diagnostic string is stable across invocations
// (useful for telemetry deduplication).
func probeDeprecatedPathSignal(projectRoot string) (bool, string) {
	for _, entry := range defs.DeprecatedPaths {
		abs := filepath.Join(projectRoot, filepath.FromSlash(entry.Path))
		if _, err := os.Stat(abs); err == nil {
			return true, entry.Path
		}
	}
	return false, ""
}
