// Package cli — v2_detection_matrix_test.go
//
// Version-signal normalization guards for SPEC-UPDATE-REINSTALL-LOOP-002 M1
// (REQ-RIL2-001..009, NFR-RIL2-001; AC-RIL2-001/002/003).
//
// Two independent guards live here:
//
//   - TestProbeVersionSignal_NormalizedMatrix — a table-driven classification
//     matrix over a RESIDUE-CARRYING fixture (.agency/ present, so Signals 2
//     and 3 are positive). All three columns are asserted per row; the
//     `Signal 1` column is load-bearing because IsV2 is true for every
//     non-v3-confirmed row on this fixture under both the old and the new rule.
//
//   - TestProbeVersionSignal_NoDestructiveWidening — the NFR-RIL2-001
//     monotonicity assertion over a RESIDUE-FREE fixture (no .agency/, no
//     deprecated paths), where IsV2 is a pure function of Signal 1 and the
//     comparison can actually move. The pre-change rule is reproduced as an
//     explicit local reference implementation so the implication is checked
//     against a real second implementation rather than a restatement of the
//     new rule's own output.

package cli

import (
	"fmt"
	"strings"
	"testing"
)

// writeVersionedProject creates a t.TempDir() project whose system.yaml carries
// the given moai.version. The version is written as a QUOTED YAML scalar so a
// bare numeric such as `3` is decoded as a string rather than failing the
// unmarshal (a decode failure would be classified as "file unparseable" and
// would silently turn the fixture into a Signal-1-positive case).
func writeVersionedProject(t *testing.T, version string) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		fmt.Sprintf("moai:\n    version: %q\n", version))
	return root
}

// makeResidueCarryingProject builds a project carrying `.agency/` at the root.
// `.agency` is both the Signal 2 artifact and a defs.DeprecatedPaths entry, so
// one directory drives Signals 2 and 3 positive simultaneously — which is what
// makes the v3-confirmed negative-override observable in the IsV2 column.
func makeResidueCarryingProject(t *testing.T, version string) string {
	t.Helper()
	root := writeVersionedProject(t, version)
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy\n")
	return root
}

// makeResidueFreeProject builds a project containing ONLY
// `.moai/config/sections/system.yaml` — no `.agency/`, no deprecated paths.
// Signals 2 and 3 are negative, so IsV2 reduces to Signal 1 and a widening of
// the version rule becomes observable.
func makeResidueFreeProject(t *testing.T, version string) string {
	t.Helper()
	return writeVersionedProject(t, version)
}

// TestProbeVersionSignal_NormalizedMatrix is the AC-RIL2-001 / AC-RIL2-002
// classification matrix. Each row asserts all three columns; subtest names are
// the literal version strings so `--- PASS: .../<version>` lines are greppable.
func TestProbeVersionSignal_NormalizedMatrix(t *testing.T) {
	cases := []struct {
		name           string
		version        string
		wantSignal1    bool
		wantIsV2       bool
		wantV3Confirms bool
	}{
		// Nine-row §A Defect 1 matrix.
		{"v3.0.1", "v3.0.1", false, false, true},
		{"3.0.1", "3.0.1", false, false, true},
		{"V3.0.1", "V3.0.1", false, false, true},
		{"v4.0.0", "v4.0.0", false, false, true},
		{"4.0.0", "4.0.0", false, false, true},
		{"empty", "", true, true, false},
		{"v2.5.0", "v2.5.0", true, true, false},
		{"2.5.0", "2.5.0", false, true, false},
		{"V2.5.0", "V2.5.0", false, true, false},

		// REQ-RIL2-005 — prerelease and build metadata classify by major alone.
		{"3.0.1-rc13", "3.0.1-rc13", false, false, true},
		{"3.0.0+build.5", "3.0.0+build.5", false, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := makeResidueCarryingProject(t, tc.version)

			gotSignal1, gotV3, detail := probeVersionSignal(root)
			if gotSignal1 != tc.wantSignal1 {
				t.Errorf("Signal 1 = %v, want %v (version %q, detail %q)",
					gotSignal1, tc.wantSignal1, tc.version, detail)
			}
			if gotV3 != tc.wantV3Confirms {
				t.Errorf("V3VersionConfirmed = %v, want %v (version %q, detail %q)",
					gotV3, tc.wantV3Confirms, tc.version, detail)
			}

			fp, err := detectV2Fingerprint(root)
			if err != nil {
				t.Fatalf("detectV2Fingerprint: %v", err)
			}
			if fp.IsV2 != tc.wantIsV2 {
				t.Errorf("IsV2 = %v, want %v (version %q; version=%v agency=%v deprecated=%v v3=%v)",
					fp.IsV2, tc.wantIsV2, tc.version,
					fp.V2DetectedViaVersion, fp.V2DetectedViaAgencyDir,
					fp.V2DetectedViaDeprecatedPath, fp.V3VersionConfirmed)
			}
			if fp.V2DetectedViaVersion != tc.wantSignal1 {
				t.Errorf("fingerprint V2DetectedViaVersion = %v, want %v (version %q)",
					fp.V2DetectedViaVersion, tc.wantSignal1, tc.version)
			}
			if fp.V3VersionConfirmed != tc.wantV3Confirms {
				t.Errorf("fingerprint V3VersionConfirmed = %v, want %v (version %q)",
					fp.V3VersionConfirmed, tc.wantV3Confirms, tc.version)
			}
		})
	}
}

// referenceProbeVersionSignalPreChange is a verbatim reproduction of the
// four-branch `switch` that probeVersionSignal carried BEFORE this SPEC (the
// literal `strings.HasPrefix(v, "v3.")` rule). It is deliberately a second,
// standalone implementation — NOT a call into probeVersionSignal — so that the
// monotonicity implication below is checked against a real reference rather
// than being tautologically true (AC-RIL2-003).
func referenceProbeVersionSignalPreChange(version string) (positive bool, v3Confirmed bool) {
	v := strings.TrimSpace(version)
	switch {
	case v == "":
		return true, false
	case strings.HasPrefix(v, "v2."):
		return true, false
	case strings.HasPrefix(v, "v3."):
		return false, true
	default:
		return false, false
	}
}

// TestProbeVersionSignal_NoDestructiveWidening is the NFR-RIL2-001 monotonicity
// assertion (AC-RIL2-003): on a residue-free project, no input that the
// pre-change rule classified as non-v2 may classify as v2 after the change.
//
// The fixture MUST be residue-free — on the residue-carrying fixture IsV2 is
// true for every non-v3-confirmed input under both rules, so the implication
// would hold trivially.
func TestProbeVersionSignal_NoDestructiveWidening(t *testing.T) {
	inputs := []string{
		"v3.0.1", "3.0.1", "V3.0.1", "v4.0.0", "4.0.0",
		"", "v2.5.0",
		"2.5.0",  // widened by a major == 2 rule that ignores the v prefix
		"V2.5.0", // same
		"1.9.0",
		"abc", // widened by reading "unparseable" as "major digits unparseable"
		"3",
	}

	for _, version := range inputs {
		t.Run(fmt.Sprintf("input=%q", version), func(t *testing.T) {
			root := makeResidueFreeProject(t, version)

			refPositive, refV3 := referenceProbeVersionSignalPreChange(version)
			// Residue-free ⇒ Signals 2 and 3 are negative, so the aggregation
			// reduces to: IsV2 = !v3Confirmed && signal1.
			referenceIsV2 := !refV3 && refPositive

			fp, err := detectV2Fingerprint(root)
			if err != nil {
				t.Fatalf("detectV2Fingerprint: %v", err)
			}

			if !referenceIsV2 && fp.IsV2 {
				t.Errorf("NFR-RIL2-001 violated for %q: reference rule classified IsV2=false, "+
					"new rule classifies IsV2=true — the destructive path widened "+
					"(version=%v agency=%v deprecated=%v v3=%v)",
					version, fp.V2DetectedViaVersion, fp.V2DetectedViaAgencyDir,
					fp.V2DetectedViaDeprecatedPath, fp.V3VersionConfirmed)
			}
		})
	}
}
