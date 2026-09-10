package spec

// progress_section_letter_guard_test.go — guards the two halves of card t408 so
// that repairing one without the other cannot pass.
//
// THE DEFECT. Two workflow skills told a SPEC author to write phase signals into
// `progress.md §F.1` (plan) and `progress.md §F.3` (sync). The canonical section
// map allocates those signals to `§E.1` and `§E.4`, and allocates progress.md
// `§F` to something else entirely — the orchestrator's Phase 4 Mode Selection
// log. A third document asserted a fixed three-file artifact set, which the Tier
// table contradicts at Tier S (two files, criteria inline in `spec.md` §3).
//
// WHY BOTH HALVES NEED A GUARD. Repairing only the wording leaves the documents
// already written in the old shape; repairing only those leaves the instruction
// that produces more of them. The card named that pair as its representative
// mutants, so each is asserted separately below.
//
// WHAT THE THIRD TEST ACTUALLY PINS, and how it differs from the card's premise.
// Lettering a PLAN signal as `§F.1` instead of `§E.1` has no effect on era
// classification: `ClassifyEra` matches `§E.2`-`§E.5` and the two SHA fields, and
// `§E.1` is not among them — so `§E.1` and `§F.1` are equally invisible to it.
// What DOES misclassify a SPEC is a progress.md whose RUN and SYNC signals are
// lettered outside `§E.*`, because then no marker matches and the file falls to
// H-2 ("progress.md without §E.* markers") — a modern SPEC reading as V3R2-R4.
// Measured on this tree: 592 progress.md files, 125 era-invisible, and only 3 of
// those 125 are invisible for this reason rather than from genuinely predating
// the section map. The pinned set below is therefore the population that letters
// ANY phase signal under `§F.N`, which is the shape the instruction produced.
//
// The set is pinned by EQUALITY, not by a ceiling. A ceiling cannot tell a new
// instance from a repaired one, and both are things a reader of a red build needs
// to know.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// phaseSignalCitationDocs are the workflow skills that name the progress.md
// section a phase owner writes into — root copies and their embedded mirrors.
// The mirror is what ships to user projects, so a root-only repair would leave
// every installed project reading the old instruction.
var phaseSignalCitationDocs = []string{
	".claude/skills/moai/workflows/plan.md",
	"internal/template/templates/.claude/skills/moai/workflows/plan.md",
	".claude/skills/moai/workflows/sync.md",
	"internal/template/templates/.claude/skills/moai/workflows/sync.md",
}

// specArtifactSetDocs are the copies of the skill that describes the SPEC
// artifact set.
var specArtifactSetDocs = []string{
	".claude/skills/moai-workflow-spec/SKILL.md",
	"internal/template/templates/.claude/skills/moai-workflow-spec/SKILL.md",
}

var (
	// progressFSectionCitation matches a citation of a NUMBERED progress.md §F
	// sub-section. Bare `progress.md §F` is deliberately NOT matched: that is the
	// legitimate Phase 4 Mode Selection log the section map allocates.
	progressFSectionCitation = regexp.MustCompile(`progress\.md\s*§F\.[0-9]`)

	// fixedArtifactCountClaim matches a HARD assertion of a fixed file count for
	// a SPEC directory, whatever the number — pinning "3" would pass the moment
	// someone updated it to "4" while the set stayed tier-dependent.
	fixedArtifactCountClaim = regexp.MustCompile(`MUST contain all [0-9]+ files`)

	// progressPhaseSignalUnderF matches a progress.md HEADING that letters a
	// phase signal under §F.N.
	progressPhaseSignalUnderF = regexp.MustCompile(`(?m)^#{2,6}\s*§F\.[0-9][^\n]*?(Plan-phase|Run-phase|Sync-phase|Mx-phase)`)
)

// progressFPhaseLetteringGrandfathered is the measured population of progress.md
// files that letter a phase signal under §F.N, as of this guard's introduction.
//
// These are historical records of SPECs that are already closed; the card's
// repair is to the instruction that produced them, not to the records. Three of
// them (marked) are additionally invisible to ClassifyEra because they carry no
// §E.* marker at all — a separate defect with a separate fix, recorded rather
// than silently rewritten here.
var progressFPhaseLetteringGrandfathered = []string{
	"SPEC-DEVPROT-REQUIRED-001",
	"SPEC-EVIDENCE-CLAIM-INVARIANT-001", // also era-invisible
	"SPEC-FOURDIM-PHANTOM-001",
	"SPEC-HARNESS-OUTCOME-CAPTURE-001",
	"SPEC-PREPUSH-WIRING-001",
	"SPEC-SEC-HARDEN-003",
	"SPEC-STOP-EVIDENCE-WRITER-001",
	"SPEC-V3R6-AGENT-TEAM-REBUILD-001",    // also era-invisible
	"SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001", // also era-invisible
	"SPEC-WEB-CONSOLE-007",
}

// TestPhaseSignalCitationsNameTheParsedSection is the wording half.
//
// Sentinel on failure: PROGRESS_SECTION_CITATION_DRIFT
func TestPhaseSignalCitationsNameTheParsedSection(t *testing.T) {
	root := repoRoot(t)

	for _, rel := range phaseSignalCitationDocs {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("PROGRESS_SECTION_CITATION_DRIFT: cannot read %s: %v "+
				"(if the file moved, update phaseSignalCitationDocs deliberately rather than "+
				"letting the scan go quiet)", rel, err)
		}
		text := string(data)

		if m := progressFSectionCitation.FindString(text); m != "" {
			t.Errorf("PROGRESS_SECTION_CITATION_DRIFT: %s cites %q. progress.md §F is the orchestrator's "+
				"Phase 4 Mode Selection log and has no numbered phase-signal sub-sections; the plan signal "+
				"is §E.1 and the sync signal is §E.4 (spec-frontmatter-schema.md § progress.md Section Map).", rel, m)
		}

		// Anti-vacuity: the file must still SAY which section the phase owner
		// writes into. Deleting the citation would satisfy the check above while
		// leaving the author with no instruction at all.
		if !strings.Contains(text, "progress.md §E.") {
			t.Errorf("PROGRESS_SECTION_CITATION_DRIFT: %s no longer cites any `progress.md §E.` section. "+
				"The repair for a wrong citation is a correct citation, not a deleted one.", rel)
		}
	}
}

// TestSpecArtifactSetIsTierScoped is the artifact-set half.
//
// Sentinel on failure: SPEC_ARTIFACT_SET_DRIFT
func TestSpecArtifactSetIsTierScoped(t *testing.T) {
	root := repoRoot(t)

	for _, rel := range specArtifactSetDocs {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("SPEC_ARTIFACT_SET_DRIFT: cannot read %s: %v", rel, err)
		}
		text := string(data)

		if m := fixedArtifactCountClaim.FindString(text); m != "" {
			t.Errorf("SPEC_ARTIFACT_SET_DRIFT: %s asserts %q. The artifact set is tier-dependent — Tier S is "+
				"spec.md + plan.md with the criteria inline in spec.md §3 — so any fixed count contradicts "+
				"spec-workflow.md § SPEC Complexity Tier for at least one tier.", rel, m)
		}

		// Anti-vacuity: deleting the clause must not pass. The document has to
		// point at the table that actually decides the set.
		if !strings.Contains(text, "SPEC Complexity Tier") {
			t.Errorf("SPEC_ARTIFACT_SET_DRIFT: %s does not name the tier table that determines the artifact set. "+
				"Removing the wrong count is only half the repair; the reader still needs to be sent somewhere.", rel)
		}
	}
}

// TestProgressPhaseSignalsUseParsedSectionLetters is the corpus half: it pins
// the population that already carries the shape, so a NEW one reddens.
//
// Sentinel on failure: PROGRESS_SECTION_LETTERING_DRIFT
func TestProgressPhaseSignalsUseParsedSectionLetters(t *testing.T) {
	root := repoRoot(t)
	specsDir := filepath.Join(root, ".moai", "specs")

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		t.Skipf("no .moai/specs directory at %s (%v) — this guard measures this repository's own corpus", specsDir, err)
	}

	var found []string
	scanned := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(specsDir, e.Name(), "progress.md"))
		if err != nil {
			continue // a SPEC without progress.md is out of this guard's subject
		}
		scanned++
		if progressPhaseSignalUnderF.Match(data) {
			found = append(found, e.Name())
		}
	}

	// Anti-vacuity floor, measured on this tree at introduction: 592 progress.md
	// files. A scan that suddenly sees a handful is reading the wrong directory,
	// and would report an empty finding set for that reason rather than because
	// the corpus is clean.
	const minScanned = 400
	if scanned < minScanned {
		t.Fatalf("PROGRESS_SECTION_LETTERING_DRIFT: scanned only %d progress.md files under %s, expected at least %d. "+
			"The corpus cannot have shrunk by that much; this is a broken scan, and a broken scan reports zero findings.",
			scanned, specsDir, minScanned)
	}

	sort.Strings(found)
	want := append([]string(nil), progressFPhaseLetteringGrandfathered...)
	sort.Strings(want)

	if strings.Join(found, "\n") != strings.Join(want, "\n") {
		t.Errorf("PROGRESS_SECTION_LETTERING_DRIFT: the set of progress.md files lettering a phase signal under §F.N moved.\n"+
			"  scanned: %d files\n  want (%d): %v\n  got  (%d): %v\n"+
			"A NEW entry means the retired instruction is being followed again — the phase signals belong under §E.N, "+
			"which is the namespace internal/spec/era.go parses. A MISSING entry means one was repaired, which is good: "+
			"remove it from progressFPhaseLetteringGrandfathered in the same commit.",
			scanned, len(want), want, len(found), found)
	}
}
