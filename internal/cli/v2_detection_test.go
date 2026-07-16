// Package cli — v2_detection_test.go
//
// Table-driven tests for v2 fingerprint detection per AC-VVCR-001 of
// SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001. Tests cover all 3 signal sources
// (system.yaml version, .agency/ directory, DeprecatedPaths enumeration)
// across their state matrix: positive, negative, and Signal-1 sub-states
// (v2.* prefix / empty string / file missing → all positive per Option α).
//
// Per CLAUDE.local.md §6 [HARD]: all temp directories use t.TempDir() for
// auto-cleanup; no project root writes.

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTestFile is a test helper that writes a file inside the given root,
// creating any necessary parent directories.
func writeTestFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir parent of %s: %v", relPath, err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

// makeTestDir is a test helper that creates an empty directory under root.
func makeTestDir(t *testing.T, root, relPath string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(abs, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", relPath, err)
	}
}

// TestDetectV2Fingerprint_Signal1_VersionField covers Signal 1 sub-states.
//
// Per AC-VVCR-001 the Signal-1 positive cases are:
//   - moai.version starts with "v2."  (legacy version installed)
//   - moai.version is empty string    (drift / unmigrated)
//   - system.yaml file is missing     (uninitialized v3 metadata)
//
// And the Signal-1 negative case is:
//   - moai.version starts with "v3."  (current major release)
func TestDetectV2Fingerprint_Signal1_VersionField(t *testing.T) {
	tests := []struct {
		name       string
		systemYAML string // empty string ⇒ omit file
		omitFile   bool
		wantSignal bool
		wantDetail string // substring expected in SignalDetails["version_signal"]
	}{
		{
			name:       "v2.0.0 prefix triggers Signal 1",
			systemYAML: "moai:\n    version: v2.0.0\n",
			wantSignal: true,
			wantDetail: "v2",
		},
		{
			name:       "v2.16.1 prefix triggers Signal 1",
			systemYAML: "moai:\n    version: v2.16.1\n",
			wantSignal: true,
			wantDetail: "v2",
		},
		{
			name:       "empty version field triggers Signal 1",
			systemYAML: "moai:\n    version: \"\"\n",
			wantSignal: true,
			wantDetail: "empty",
		},
		{
			name:       "missing version field triggers Signal 1",
			systemYAML: "moai:\n    template_version: v3.0.0-rc2\n",
			wantSignal: true,
			wantDetail: "empty",
		},
		{
			name:       "missing system.yaml file triggers Signal 1",
			omitFile:   true,
			wantSignal: true,
			wantDetail: "missing",
		},
		{
			name:       "v3.0.0-rc2 negative case",
			systemYAML: "moai:\n    version: v3.0.0-rc2\n",
			wantSignal: false,
		},
		{
			name:       "v3.1.0 negative case",
			systemYAML: "moai:\n    version: v3.1.0\n",
			wantSignal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if !tt.omitFile {
				writeTestFile(t, root, ".moai/config/sections/system.yaml", tt.systemYAML)
			}

			fp, err := detectV2Fingerprint(root)
			if err != nil {
				t.Fatalf("detectV2Fingerprint: unexpected error: %v", err)
			}
			if fp.V2DetectedViaVersion != tt.wantSignal {
				t.Errorf("V2DetectedViaVersion = %v, want %v", fp.V2DetectedViaVersion, tt.wantSignal)
			}
			if tt.wantSignal && tt.wantDetail != "" {
				detail := fp.SignalDetails["version_signal"]
				if detail == "" {
					t.Errorf("SignalDetails[\"version_signal\"] is empty; want substring %q", tt.wantDetail)
				}
				// Substring match (case-insensitive for "empty" / "missing").
				if !containsLowerSubstring(detail, tt.wantDetail) {
					t.Errorf("SignalDetails[\"version_signal\"] = %q; want substring %q", detail, tt.wantDetail)
				}
			}
		})
	}
}

// TestDetectV2Fingerprint_Signal2_AgencyDir covers Signal 2.
func TestDetectV2Fingerprint_Signal2_AgencyDir(t *testing.T) {
	tests := []struct {
		name       string
		createDir  bool
		wantSignal bool
	}{
		{
			name:       ".agency/ present triggers Signal 2",
			createDir:  true,
			wantSignal: true,
		},
		{
			name:       ".agency/ absent does not trigger Signal 2",
			createDir:  false,
			wantSignal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			// Ensure system.yaml is v3-clean so Signal 1 stays negative.
			writeTestFile(t, root, ".moai/config/sections/system.yaml",
				"moai:\n    version: v3.0.0-rc2\n")
			if tt.createDir {
				makeTestDir(t, root, ".agency")
			}

			fp, err := detectV2Fingerprint(root)
			if err != nil {
				t.Fatalf("detectV2Fingerprint: unexpected error: %v", err)
			}
			if fp.V2DetectedViaAgencyDir != tt.wantSignal {
				t.Errorf("V2DetectedViaAgencyDir = %v, want %v", fp.V2DetectedViaAgencyDir, tt.wantSignal)
			}
		})
	}
}

// TestDetectV2Fingerprint_Signal3_DeprecatedPath covers Signal 3 using
// real entries from defs.DeprecatedPaths (Category B v.2.x-era subset).
func TestDetectV2Fingerprint_Signal3_DeprecatedPath(t *testing.T) {
	tests := []struct {
		name         string
		seedPaths    []string // paths under project root to create
		wantSignal   bool
		wantInDetail string // expected substring in deprecated_signal_first_hit
	}{
		{
			name:         "agency agent path triggers Signal 3",
			seedPaths:    []string{".claude/agents/moai/planner.md"},
			wantSignal:   true,
			wantInDetail: ".claude/agents/moai/planner.md",
		},
		{
			name:         "retired manager path triggers Signal 3",
			seedPaths:    []string{".claude/agents/moai/manager-strategy.md"},
			wantSignal:   true,
			wantInDetail: "manager-strategy",
		},
		{
			name:         "rc1-stage core/ directory triggers Signal 3",
			seedPaths:    []string{".claude/agents/core/manager-develop.md"},
			wantSignal:   true,
			wantInDetail: ".claude/agents/core",
		},
		{
			name:       "no deprecated paths present → Signal 3 negative",
			seedPaths:  nil,
			wantSignal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			// Keep Signal 1 negative.
			writeTestFile(t, root, ".moai/config/sections/system.yaml",
				"moai:\n    version: v3.0.0-rc2\n")
			for _, p := range tt.seedPaths {
				writeTestFile(t, root, p, "stub\n")
			}

			fp, err := detectV2Fingerprint(root)
			if err != nil {
				t.Fatalf("detectV2Fingerprint: unexpected error: %v", err)
			}
			if fp.V2DetectedViaDeprecatedPath != tt.wantSignal {
				t.Errorf("V2DetectedViaDeprecatedPath = %v, want %v",
					fp.V2DetectedViaDeprecatedPath, tt.wantSignal)
			}
			if tt.wantSignal && tt.wantInDetail != "" {
				detail := fp.SignalDetails["deprecated_signal_first_hit"]
				if !containsSubstring(detail, tt.wantInDetail) {
					t.Errorf("deprecated_signal_first_hit = %q; want substring %q",
						detail, tt.wantInDetail)
				}
			}
		})
	}
}

// TestDetectV2Fingerprint_IsV2_Aggregation covers the final IsV2 boolean
// resolution (any one of the 3 signals positive ⇒ IsV2 true).
func TestDetectV2Fingerprint_IsV2_Aggregation(t *testing.T) {
	tests := []struct {
		name        string
		systemYAML  string
		omitSysYAML bool
		makeAgency  bool
		seedPath    string
		wantIsV2    bool
	}{
		{
			name:       "all 3 signals negative → IsV2 false",
			systemYAML: "moai:\n    version: v3.0.0-rc2\n",
			wantIsV2:   false,
		},
		{
			name:       "Signal 1 alone (v2.* version) → IsV2 true",
			systemYAML: "moai:\n    version: v2.16.1\n",
			wantIsV2:   true,
		},
		{
			// REQ-CRR-001 (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002): a populated
			// v3.* version short-circuits Signal 2 (.agency/ residue).
			name:       "v3.* + .agency/ residue → IsV2 false (REQ-CRR-001 v3-version negative-override)",
			systemYAML: "moai:\n    version: v3.0.0-rc2\n",
			makeAgency: true,
			wantIsV2:   false,
		},
		{
			// REQ-CRR-001: a populated v3.* version short-circuits Signal 3
			// (DeprecatedPaths residue).
			name:       "v3.* + deprecated path residue → IsV2 false (REQ-CRR-001 v3-version negative-override)",
			systemYAML: "moai:\n    version: v3.0.0-rc2\n",
			seedPath:   ".claude/agents/moai/manager-strategy.md",
			wantIsV2:   false,
		},
		{
			name:        "Signal 1 (file missing) alone → IsV2 true",
			omitSysYAML: true,
			wantIsV2:    true,
		},
		{
			name:       "Signals 1+2+3 combined → IsV2 true",
			systemYAML: "moai:\n    version: v2.16.1\n",
			makeAgency: true,
			seedPath:   ".claude/agents/core/manager-develop.md",
			wantIsV2:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if !tt.omitSysYAML {
				writeTestFile(t, root, ".moai/config/sections/system.yaml", tt.systemYAML)
			}
			if tt.makeAgency {
				makeTestDir(t, root, ".agency")
			}
			if tt.seedPath != "" {
				writeTestFile(t, root, tt.seedPath, "stub\n")
			}

			fp, err := detectV2Fingerprint(root)
			if err != nil {
				t.Fatalf("detectV2Fingerprint: unexpected error: %v", err)
			}
			if fp.IsV2 != tt.wantIsV2 {
				t.Errorf("IsV2 = %v, want %v (fp = %+v)", fp.IsV2, tt.wantIsV2, fp)
			}
		})
	}
}

// TestDetectV2Fingerprint_V3Override_AC_CRR_002 is the AC-CRR-002 reproduction
// test (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002) at the fingerprint level.
//
// A genuine v3 project (populated v3.* system.yaml) carrying stale legacy
// residue (a lingering .agency/ dir AND a DeprecatedPath seed) MUST resolve
// IsV2 = false. This is the loop-termination contract: when IsV2 is false,
// runUpdate does NOT enter the clean-reinstall path, so a second `moai update`
// converges instead of looping (#1084).
//
// REQ-CRR-001: a confirmed v3.* version short-circuits Signal 2 / Signal 3.
//
// Pre-fix expectation: FAIL — the current aggregation is a pure disjunction,
// so Signal 2 (.agency/) OR Signal 3 (deprecated path) drives IsV2 to true
// despite the populated v3 version.
func TestDetectV2Fingerprint_V3Override_AC_CRR_002(t *testing.T) {
	root := t.TempDir()

	// Populated v3.* version — the negative-override trigger.
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0\n")

	// Stale legacy residue that would otherwise trip Signal 2 / Signal 3.
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "stub\n")

	fp, err := detectV2Fingerprint(root)
	if err != nil {
		t.Fatalf("detectV2Fingerprint: unexpected error: %v", err)
	}

	if fp.IsV2 {
		t.Errorf("AC-CRR-002: IsV2 = true; want false "+
			"(REQ-CRR-001 v3-version negative-override must short-circuit "+
			"Signal 2/3 legacy residue)\nfp = %+v", fp)
	}

	// Verify the override MECHANISM, not just the outcome: V3VersionConfirmed
	// must be true (the override fired), and the per-signal flags must still
	// reflect that legacy residue WAS detected (Signal 2 + Signal 3 positive)
	// — proving the override short-circuited rather than the residue being
	// absent.
	if !fp.V3VersionConfirmed {
		t.Errorf("AC-CRR-001(a)/AC-CRR-002: V3VersionConfirmed = false; want true "+
			"(populated v3.* version must set the override flag)\nfp = %+v", fp)
	}
	if !fp.V2DetectedViaAgencyDir {
		t.Errorf("AC-CRR-002: V2DetectedViaAgencyDir = false; want true "+
			"(.agency/ residue WAS present — override must short-circuit, not hide)")
	}
	if !fp.V2DetectedViaDeprecatedPath {
		t.Errorf("AC-CRR-002: V2DetectedViaDeprecatedPath = false; want true "+
			"(deprecated path residue WAS present — override must short-circuit, not hide)")
	}
	// AC-CRR-001(b): the version_signal detail MUST name the v3-version
	// negative-override so callers (telemetry / --dry-run) report WHY IsV2
	// was forced false.
	versionDetail := fp.SignalDetails["version_signal"]
	if !containsLowerSubstring(versionDetail, "v3-version negative-override") {
		t.Errorf("AC-CRR-001(b): SignalDetails[\"version_signal\"] = %q; "+
			"want substring \"v3-version negative-override\"", versionDetail)
	}
}

// TestDetectV2Fingerprint_EmptyProject covers the edge case of a completely
// empty project root: no system.yaml, no .agency/, no deprecated paths.
// Per Option α, missing system.yaml triggers Signal 1 alone → IsV2 true.
func TestDetectV2Fingerprint_EmptyProject(t *testing.T) {
	root := t.TempDir()

	fp, err := detectV2Fingerprint(root)
	if err != nil {
		t.Fatalf("detectV2Fingerprint: unexpected error: %v", err)
	}
	if !fp.IsV2 {
		t.Errorf("empty project: IsV2 = false; want true (Option α — missing system.yaml is positive Signal 1)")
	}
	if !fp.V2DetectedViaVersion {
		t.Errorf("empty project: V2DetectedViaVersion = false; want true")
	}
	if fp.V2DetectedViaAgencyDir {
		t.Errorf("empty project: V2DetectedViaAgencyDir = true; want false")
	}
	if fp.V2DetectedViaDeprecatedPath {
		t.Errorf("empty project: V2DetectedViaDeprecatedPath = true; want false")
	}
}

// TestDetectV2Fingerprint_NonexistentRoot verifies error handling for a
// project root that does not exist. The function MUST return an error
// rather than panicking.
func TestDetectV2Fingerprint_NonexistentRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nonexistent-subdir")

	_, err := detectV2Fingerprint(root)
	if err == nil {
		t.Errorf("expected error for nonexistent project root; got nil")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func containsSubstring(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) &&
		(haystack == needle || stringContains(haystack, needle)))
}

func containsLowerSubstring(haystack, needle string) bool {
	return containsSubstring(stringToLower(haystack), stringToLower(needle))
}

// stringContains is a tiny re-implementation to avoid pulling strings into
// the test file's import list before the implementation file exists. It is
// equivalent to strings.Contains and exists only because the SPEC TDD cycle
// authors the test before the source.
func stringContains(s, substr string) bool {
	n := len(substr)
	if n == 0 {
		return true
	}
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == substr {
			return true
		}
	}
	return false
}

func stringToLower(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}
