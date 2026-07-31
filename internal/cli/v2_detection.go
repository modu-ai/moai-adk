// Package cli — v2_detection.go
//
// Implements the v2 fingerprint heuristic for SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
// (REQ-VVCR-001..002, AC-VVCR-001). The detector inspects three independent
// signals and resolves the IsV2 boolean as the disjunction: ANY positive
// signal ⇒ IsV2 true. This drives the clean-reinstall code path in
// `runUpdate` (see update.go integration in M5).
//
// REQ-CRR-001 (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002) adds a v3-version
// negative-override: when probeVersionSignal confirms a populated v3.*
// version, the final IsV2 aggregation short-circuits to false regardless of
// Signal 2 (.agency/) or Signal 3 (DeprecatedPaths) being positive. This
// prevents an infinite `moai update` loop on a genuine v3 project carrying
// stale legacy residue (#1084). Option α broader detection (empty/missing
// system.yaml → positive Signal 1) is preserved — the override fires ONLY on
// a confirmed populated v3.* string.
//
// Signal sources:
//
//   - Signal 1 (V2DetectedViaVersion): `.moai/config/sections/system.yaml`
//     `moai.version` field. Positive when the normalized major version is 2
//     and the string carried a leading "v", OR the version is empty, OR the
//     system.yaml file is missing/unparseable entirely.
//     Empty / missing branches reflect Option α (broader detection): v3
//     projects always carry system.yaml with populated moai.version, so
//     drift / absence is a positive v2 signal.
//
//   - Signal 2 (V2DetectedViaAgencyDir): existence of `.agency/` legacy
//     directory at project root. v.2.x exclusive artifact.
//
//   - Signal 3 (V2DetectedViaDeprecatedPath): existence of ANY path
//     enumerated in defs.DeprecatedPaths. See internal/defs/dirs.go for the
//     authoritative entry count and per-category (A/B/C/D) breakdown.
//
// The SignalDetails map carries per-signal diagnostic strings used by
// telemetry and `--dry-run` output (REQ-VVCR-028 / REQ-VVCR-029).

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	// IsV2 is true when ANY of the 3 signals is positive (disjunction),
	// UNLESS V3VersionConfirmed is true (REQ-CRR-001 v3-version
	// negative-override short-circuits the disjunction to false).
	IsV2 bool

	// Per-signal positive flags (used by telemetry).
	V2DetectedViaVersion        bool
	V2DetectedViaAgencyDir      bool
	V2DetectedViaDeprecatedPath bool

	// V3VersionConfirmed is true when system.yaml carries a populated
	// moai.version starting with "v3.". When true, IsV2 is forced to false
	// regardless of Signal 2/3 state (REQ-CRR-001 v3-version
	// negative-override — prevents the infinite `moai update` loop #1084).
	V3VersionConfirmed bool

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
	versionPositive, v3Confirmed, versionDetail := probeVersionSignal(projectRoot)
	fp.V2DetectedViaVersion = versionPositive
	fp.V3VersionConfirmed = v3Confirmed
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

	// Aggregation: any one positive ⇒ IsV2 true, UNLESS a populated v3.*
	// version is confirmed (REQ-CRR-001 v3-version negative-override). A
	// genuine v3 project carrying stale legacy residue (.agency/ dir or
	// deprecated paths) MUST NOT be classified as v2 — doing so causes an
	// infinite `moai update` loop (#1084). The override short-circuits
	// Signal 2/3 while leaving Option α (empty/missing system.yaml →
	// positive Signal 1) intact.
	fp.IsV2 = !fp.V3VersionConfirmed && (fp.V2DetectedViaVersion ||
		fp.V2DetectedViaAgencyDir ||
		fp.V2DetectedViaDeprecatedPath)

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

// probeVersionSignal returns (positive, v3Confirmed, detail).
//
// The version string is normalized before comparison (REQ-RIL2-001): trim
// surrounding whitespace, strip at most one leading "v"/"V", parse the leading
// digit run as the major version. Classification then follows:
//   - file missing        → positive, v3Confirmed=false, detail "system.yaml missing"
//   - empty version       → positive, v3Confirmed=false, detail "moai.version empty"
//   - major >= 3          → negative, v3Confirmed=true,  detail names the override (REQ-RIL2-002)
//   - major == 2, "v" pfx → positive, v3Confirmed=false, detail names the normalized major (REQ-RIL2-003)
//   - any other parseable → negative, v3Confirmed=false, detail empty (REQ-RIL2-006)
//   - non-numeric leading → negative, v3Confirmed=false, detail empty (REQ-RIL2-006)
//
// The empty/missing branches (Option α broader detection) are preserved
// unchanged — the v3 negative-override fires ONLY on a confirmed populated
// version whose normalized major is 3 or greater, NOT on drift/absence.
//
// FILE parse errors are treated as positive Signal 1 with a descriptive detail
// (REQ-RIL2-004) — a malformed system.yaml in a project running `moai update`
// is more likely to be a partial v2 migration than a deliberately corrupted v3
// file. A well-formed file carrying an unrecognized version string is a
// DIFFERENT case, governed by REQ-RIL2-006 (negative), because classifying it
// positive would widen the destructive path (NFR-RIL2-001).
func probeVersionSignal(projectRoot string) (bool, bool, string) {
	sysYAMLPath := filepath.Join(projectRoot,
		".moai", "config", "sections", "system.yaml")

	data, err := os.ReadFile(sysYAMLPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, false, "system.yaml missing"
		}
		// Other read errors (permission denied, etc.) — treat as positive
		// to err on the side of clean-reinstall over silent skip.
		return true, false, fmt.Sprintf("system.yaml read error: %v", err)
	}

	var block systemYAMLMoaiBlock
	if err := yaml.Unmarshal(data, &block); err != nil {
		return true, false, fmt.Sprintf("system.yaml parse error: %v", err)
	}

	v := strings.TrimSpace(block.Moai.Version)
	if v == "" {
		return true, false, "moai.version empty"
	}

	major, prefix, parsed := normalizeVersionMajor(v)
	switch {
	case !parsed:
		// REQ-RIL2-006: a well-formed system.yaml carrying a version whose
		// leading component is not numeric ("abc") is Signal 1 NEGATIVE with
		// no override, so Signals 2/3 decide. "Unparseable" in REQ-RIL2-004
		// means the FILE failed to parse (handled above), not that the major
		// digits failed to parse — reading it the other way would flip a
		// residue-free project to IsV2=true and widen the destructive path
		// (NFR-RIL2-001).
		return false, false, ""
	case major >= 3:
		// REQ-RIL2-002 v3-version negative-override (widened from the former
		// literal "v3." prefix test by REQ-RIL2-001 normalization): any
		// normalized major >= 3 confirms a genuine v3 project. The detail
		// names the override so callers (telemetry / --dry-run / the reason
		// string surfaced by runUpdate) can report WHY IsV2 was forced false
		// despite legacy residue.
		return false, true, fmt.Sprintf(
			"v3-version negative-override (REQ-RIL2-002): moai.version normalized major %d >= 3 (%s) — Signal 2/3 short-circuited",
			major, v)
	case major == 2 && prefix == 'v':
		// REQ-RIL2-003: a prefixed major-2 string is Signal 1 positive. The
		// prefix scope is deliberate and matches the pre-change behaviour
		// exactly — a bare "2.x" (and an uppercase-prefixed "V2.x", which the
		// pre-change literal "v2." test also did not match) falls through to
		// REQ-RIL2-006 below and stays Signal 1 NEGATIVE, because making
		// either positive would flip a residue-free project from IsV2=false to
		// IsV2=true (NFR-RIL2-001, which admits no exception).
		return true, false, fmt.Sprintf("moai.version normalized major 2 with %q prefix (%s)", string(prefix), v)
	default:
		// REQ-RIL2-006: any other parseable major (1, bare/uppercase-prefixed
		// 2, ...) → negative, no override. Reproduces the pre-change default
		// branch.
		return false, false, ""
	}
}

// normalizeVersionMajor parses a moai.version string per REQ-RIL2-001:
// trim surrounding whitespace, strip at most one leading "v" or "V", then read
// the leading run of decimal digits as the major version.
//
// Prerelease and build suffixes ("3.0.1-rc13", "3.0.0+build.5") are classified
// by the major component alone, since everything after the digit run is ignored
// (REQ-RIL2-005).
//
// Returns the parsed major, the stripped prefix byte (0 when the string carried
// no "v"/"V"), and parsed=false when the leading component is not numeric or
// does not fit an int.
func normalizeVersionMajor(version string) (major int, prefix byte, parsed bool) {
	s := strings.TrimSpace(version)
	if s == "" {
		return 0, 0, false
	}
	if s[0] == 'v' || s[0] == 'V' {
		prefix = s[0]
		s = s[1:]
	}

	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, prefix, false
	}

	major, err := strconv.Atoi(s[:end])
	if err != nil {
		// Overflow on an absurd digit run. Treat as unparseable → Signal 1
		// negative, which is the non-widening direction (NFR-RIL2-001).
		return 0, prefix, false
	}
	return major, prefix, true
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

// isMoAIProject reports whether projectRoot carries a positive moai-project
// marker. REQ-CRR-005 / AC-CRR-004: the clean-reinstall path requires a genuine
// project context; an arbitrary non-project directory MUST NOT trigger
// clean-reinstall (#1086).
//
// Positive marker (AC-CRR-004(a)): presence of `.moai/config/sections/system.yaml`
// as a regular file. A bare `.moai/` directory is intentionally insufficient —
// the system.yaml file is the canonical v3 project marker whose absence in a
// non-project cwd is the #1086 regression root cause. When this returns false,
// the clean-reinstall gate in runUpdate (update.go) refuses entry regardless of
// any legacy residue (.agency/, deprecated paths) that detectV2Fingerprint may
// have flagged.
//
// Edge-5 (symlink): os.Stat (not Lstat) follows symlinks, so a system.yaml
// symlink-to-regular-file also satisfies the marker. A symlink loop yields a
// Stat error → treated as marker absent. A directory named system.yaml is
// rejected by the !IsDir() check.
//
// Edge-6 (Windows): filepath.Join handles path separators, so the predicate
// resolves identically on macOS/Linux/Windows.
func isMoAIProject(projectRoot string) bool {
	marker := filepath.Join(projectRoot,
		defs.MoAIDir, "config", "sections", "system.yaml")
	info, err := os.Stat(marker)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
