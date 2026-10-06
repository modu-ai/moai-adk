package template

// The cutover scripts take the retiring integration branch and the target base
// branch as arguments and name neither (SPEC-GITHUB-FLOW-DEFAULT-001, card t1453,
// leader decision D17 option 2).
//
// The sweep guard (github_flow_sweep_guard_test.go) judges `scripts/*.sh` live text
// that names the retiring branch. The three cutover scripts were created by this
// card with that name hard-coded, so they would be flagged by the very guard the
// card builds. The leader rejected a named exclusion and a cap raise; the scripts
// now carry no such literal at all.
//
// What this asserts, per script, on the script text as it stands in the tree:
//
//  1. the sweep engine (sweepScan, the same code the tree run uses, with an EMPTY
//     allow list) reports no violating line: no strong and no weak finding;
//  2. a plain whole-word, case-insensitive match for the retiring branch name finds
//     nothing, so a comment-only mention counts as well (the engine exempts a few
//     lines, for example a verb use or a retirement marker, which a plain grep does
//     not);
//  3. the sweep is not empty: the script was read, it has lines, and the engine
//     still flags the literal in a known-bad sample (a positive control, so a
//     classifier that stopped matching cannot turn this green).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// cutoverScriptsForLiteralGuard are the three scripts this card created for the cutover.
var cutoverScriptsForLiteralGuard = []string{
	"scripts/cutover-precheck.sh",
	"scripts/cutover-rehearsal.sh",
	"scripts/cutover-protection-compare.sh",
}

// cutoverBranchWordRE is the plain whole-word match: `develop` as a word, so it also
// catches `develop-tip`, a hyphenated mention the engine classifies as an identifier.
var cutoverBranchWordRE = regexp.MustCompile(`(?i)\bdevelop\b`)

func TestCutoverScriptsNameNoBranchLiteral(t *testing.T) {
	root := filepath.Join("..", "..")

	// Positive control: the engine and the plain match both flag a known-bad line.
	probe := "LOCAL_REF=\"develop\"\n"
	flagged := 0
	for _, f := range sweepScan("scripts/probe.sh", probe, nil) {
		if f.violation() {
			flagged++
		}
	}
	if flagged == 0 || !cutoverBranchWordRE.MatchString(probe) {
		t.Fatalf("positive control failed: engine violations=%d, plain match=%v; the guard cannot see a literal", flagged, cutoverBranchWordRE.MatchString(probe))
	}

	for _, rel := range cutoverScriptsForLiteralGuard {
		t.Run(filepath.Base(rel), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				t.Fatalf("read %s: %v", rel, err)
			}
			text := string(raw)
			if strings.Count(text, "\n") < 50 {
				t.Fatalf("%s has %d lines; an empty sweep asserts nothing", rel, strings.Count(text, "\n"))
			}

			for _, f := range sweepScan(rel, text, nil) {
				if f.violation() {
					t.Errorf("sweep engine: %s", f)
				}
			}
			for i, line := range strings.Split(text, "\n") {
				if cutoverBranchWordRE.MatchString(line) {
					t.Errorf("plain whole-word match: %s:%d: %s", rel, i+1, strings.TrimSpace(line))
				}
			}
		})
	}
}
