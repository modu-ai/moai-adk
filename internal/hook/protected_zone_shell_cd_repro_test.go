package hook

// protected_zone_shell_cd_repro_test.go — RED reproduction for
// SPEC-ZONE-SHELL-CD-TRACKING-001 (card t1584), M1 step 1. Three cd shapes whose
// destination the walker does not track today: each one resets the
// possible-directory set to the project root, so the deletion that follows
// resolves against the wrong position and the guard answers allow. Every subtest
// asserts the NEW deny semantics; on the pre-fix guard each prints
// decision="allow" with an empty reason (the reset), which is the RED state the
// SPEC's AC-ZSCD-001..003 cite.
//
// The home is this file rather than a cd_track_ cell in zoneParsingMatrixCells:
// the in-project absolute shape needs the fixture root, which the runner's static
// cell command cannot carry — plan M1 step 1 names a reproduction subtest as the
// fallback for exactly that case. The shapes move into the matrix's cd_track_
// group at M2.

import (
	"fmt"
	"path/filepath"
	"testing"
)

// zoneCdReproShape is one reproduction shape: the zone the fixture declares, the
// cd command (built from the fixture root), and the defect the RED names.
type zoneCdReproShape struct {
	name    string
	paths   string
	command func(root string) string
	defect  string
}

// TestProtectedZoneShellCdTrackingRepro judges the three RED shapes through the
// real PreToolUse handler, one subtest per shape, each printing its verdict line
// so the per-shape result is countable in -v output.
func TestProtectedZoneShellCdTrackingRepro(t *testing.T) {
	shapes := []zoneCdReproShape{
		{
			name:    "shape1_dashdash_relative",
			paths:   "zone_dir/",
			command: func(string) string { return "cd -- zone_dir && rm a.log" },
			defect:  "the bare -- separator is counted as a second cd word, so zoneNextCwd resets the directory set",
		},
		{
			name:    "shape2_absolute_in_project",
			paths:   "zone_dir/",
			command: func(root string) string { return "cd " + filepath.Join(root, "zone_dir") + " && rm a.log" },
			defect:  "an in-project absolute destination takes zoneNextCwd's absolute reset, so the directory set degenerates to the root",
		},
		{
			name:    "shape3_dashdash_hyphen",
			paths:   "-zone/",
			command: func(string) string { return "cd -- -zone && rm a.log" },
			defect:  "a hyphen-leading operand after -- is read as an option and zoneNextCwd resets the directory set",
		},
	}
	for _, sh := range shapes {
		t.Run(sh.name, func(t *testing.T) {
			root := newZoneRoot(t, zoneShippedDoc(fmt.Sprintf("  probe_zone:\n    paths: [%q]\n", sh.paths)), "")
			h := zoneTestHandler(t, root)
			command := sh.command(root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
			t.Logf("cmd=%q decision=%q reason=%q", command, d, r)
			if d != DecisionDeny {
				t.Errorf("%q: decision=%q reason=%q, want deny — the cd destination was not tracked (%s)", command, d, r, sh.defect)
				return
			}
			wantZoneDeny(t, command, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}
