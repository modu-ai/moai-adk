package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The doctrine-surface guard for SPEC-EVIDENCE-CITATION-CANON-001.
//
// A gitignored path cannot be reached at audit time, so a doctrine document may
// not name `.moai/state/verify` as a citation target. The directory keeps its
// other roles (machine-local scratch, the keyed snapshot store, the SSE watch
// source), and those roles are recorded here as allowlist entries rather than
// left to a reader's judgement.
//
// Two properties keep this guard from going quietly vacuous:
//
//   - It reads BOTH the repository copy and the `internal/template/templates`
//     mirror. A guard that reads only the mirror stays green while the rule is
//     deleted from the repository copy — a failure shape this repository
//     already contains (gitignore_agents_mirror_test.go reads only the mirror).
//   - It reports what it visited, not just what it found. A count alone cannot
//     say which subtree went missing, and the two smallest subtrees fit inside
//     the floor's slack — so subtree presence is asserted as set equality,
//     separately from the floor.

const forbiddenCitation = ".moai/state/verify"

// evidenceCitationSubtrees are the doctrine surfaces, relative to a `.claude`
// root. Both trees carry all four.
var evidenceCitationSubtrees = []string{"rules", "agents", "output-styles", "skills"}

// citationTree is one `.claude` root the guard walks.
type citationTree struct {
	Name      string
	ClaudeDir string
	// MinFiles is derived from the measured population, NOT from the number of
	// violations: the two are different populations, and after the repair the
	// violation count is zero and would anchor nothing. Measured on this tree:
	// repo root 363 files, template mirror 338. The floor of 300 stops a
	// collapse to any single subtree (the largest, `skills`, is 251 < 300).
	MinFiles int
}

func evidenceCitationTrees() []citationTree {
	return []citationTree{
		{Name: "repo-root", ClaudeDir: filepath.Join("..", "..", ".claude"), MinFiles: 300},
		{Name: "template-mirror", ClaudeDir: filepath.Join("templates", ".claude"), MinFiles: 300},
	}
}

// allowEntry exempts ONE occurrence shape in ONE file.
//
// The unit is deliberate: an entry carrying only a file path would exempt that
// whole file, and a later genuine citation added to it would pass unseen. The
// Literal is the exact substring the exempt line must contain, so an entry
// covers the occurrence that was reviewed and nothing else.
type allowEntry struct {
	// File is relative to a `.claude` root, so one entry covers both trees —
	// which is also what makes mirror drift visible rather than exempt.
	File string
	// Literal must appear on the line for the occurrence to be exempt.
	Literal string
	// Why records the reviewed reason. Documentation, not matched against.
	Why string
}

// Every entry below was reviewed against one occurrence the scanner reported
// while this list was empty — none was written from memory of what the corpus
// probably contains.
var evidenceCitationAllowlist = []allowEntry{
	{
		File:    "rules/moai/core/agent-common-protocol.md",
		Literal: "is **machine-local scratch**",
		Why:     "names the directory as scratch — the sentence that withdraws the citation role",
	},
	{
		File:    "rules/moai/core/agent-common-protocol-reference.md",
		Literal: "is **machine-local scratch**",
		Why:     "same naming, in the detail companion",
	},
	{
		File:    "rules/moai/core/agent-common-protocol-reference.md",
		Literal: ".moai/state/verify/snapshots",
		Why:     "machine carve-out 1: the HEAD-SHA-keyed snapshot store (internal/verify/store.go)",
	},
	{
		File:    "rules/moai/core/agent-common-protocol-reference.md",
		Literal: "whose fsnotify map watches",
		Why:     "machine carve-out 2: the SSE watch source (internal/web/events.go)",
	},
	{
		File:    "agents/moai/manager-lead.md",
		Literal: "to **machine-local scratch**",
		Why:     "Context-Folding Step 1 names the capture target as scratch before the export step",
	},
	{
		File:    "agents/moai/manager-lead.md",
		Literal: "mkdir -p .moai/state/verify/$MOAI_SESSION_ID/",
		Why:     "scratch capture recipe — a write instruction, not a citation",
	},
	{
		File:    "agents/moai/manager-lead.md",
		Literal: "| tee .moai/state/verify/$MOAI_SESSION_ID/",
		Why:     "same recipe, the tee line",
	},
	{
		File:    "skills/moai/workflows/gate.md",
		Literal: "record its result into the shared snapshot store",
		Why:     "carve-out: `moai verify record` writes the keyed store; no human reads this path",
	},
	{
		File:    "skills/moai/workflows/loop.md",
		Literal: "The mechanical read surface is the shared diagnostic snapshot",
		Why:     "carve-out: the loop's completion evaluator reads the same keyed store",
	},
}

// evidenceCitationAllowlistSize pins the list length. Adding a line to the
// allowlist is the cheapest way to make a violation disappear, so the length
// carries a price: it cannot grow without this constant moving too.
const evidenceCitationAllowlistSize = 9

// validateAllowlist rejects the shapes that would hollow the list out.
func validateAllowlist(entries []allowEntry) error {
	for i, e := range entries {
		if strings.TrimSpace(e.File) == "" {
			return fmt.Errorf("allowlist[%d]: empty file path", i)
		}
		if strings.TrimSpace(e.Literal) == "" {
			return fmt.Errorf("allowlist[%d]: entry for %q carries no literal — "+
				"a file-only entry exempts the whole file", i, e.File)
		}
	}
	return nil
}

type citationViolation struct {
	Tree string
	File string
	Line int
	Text string
}

func (v citationViolation) String() string {
	return fmt.Sprintf("%s:%s:%d: %s", v.Tree, v.File, v.Line, strings.TrimSpace(v.Text))
}

type scanReport struct {
	Trees      []string
	Subtrees   map[string][]string
	Scanned    map[string]int
	Violations map[string][]citationViolation
}

// scanContent is the whole decision procedure for one file. The walkers and the
// mutation fixtures both go through it, so a fixture demonstration is a
// demonstration of the code that runs against the corpus.
func scanContent(tree, file, content string, allow []allowEntry) []citationViolation {
	var out []citationViolation
	for i, line := range strings.Split(content, "\n") {
		if !strings.Contains(line, forbiddenCitation) {
			continue
		}
		exempt := false
		for _, e := range allow {
			if e.File == file && strings.Contains(line, e.Literal) {
				exempt = true
				break
			}
		}
		if !exempt {
			out = append(out, citationViolation{Tree: tree, File: file, Line: i + 1, Text: line})
		}
	}
	return out
}

func scanEvidenceCitations(trees []citationTree, subtrees []string, allow []allowEntry) (*scanReport, error) {
	rep := &scanReport{
		Subtrees:   map[string][]string{},
		Scanned:    map[string]int{},
		Violations: map[string][]citationViolation{},
	}
	for _, tree := range trees {
		rep.Trees = append(rep.Trees, tree.Name)
		for _, sub := range subtrees {
			root := filepath.Join(tree.ClaudeDir, sub)
			info, err := os.Stat(root)
			if err != nil || !info.IsDir() {
				return nil, fmt.Errorf("tree %s: subtree %q is not readable: %v", tree.Name, sub, err)
			}
			rep.Subtrees[tree.Name] = append(rep.Subtrees[tree.Name], sub)
			err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
					return nil
				}
				rel, err := filepath.Rel(tree.ClaudeDir, path)
				if err != nil {
					return err
				}
				rel = filepath.ToSlash(rel)
				raw, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				rep.Scanned[tree.Name]++
				rep.Violations[tree.Name] = append(rep.Violations[tree.Name],
					scanContent(tree.Name, rel, string(raw), allow)...)
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("tree %s: walk %q: %w", tree.Name, sub, err)
			}
		}
		sort.Strings(rep.Subtrees[tree.Name])
	}
	return rep, nil
}

// --- Direction 1: the pass direction (AC-ECC-010) ---

func TestEvidenceCitation_Corpus(t *testing.T) {
	if err := validateAllowlist(evidenceCitationAllowlist); err != nil {
		t.Fatalf("allowlist invalid: %v", err)
	}
	trees := evidenceCitationTrees()
	rep, err := scanEvidenceCitations(trees, evidenceCitationSubtrees, evidenceCitationAllowlist)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	for _, tree := range trees {
		t.Logf("tree %s: scanned %d .md files across %v, %d violation(s)",
			tree.Name, rep.Scanned[tree.Name], rep.Subtrees[tree.Name], len(rep.Violations[tree.Name]))
		if got := rep.Scanned[tree.Name]; got < tree.MinFiles {
			t.Errorf("tree %s: scanned %d files, want >= %d — the scan scope collapsed",
				tree.Name, got, tree.MinFiles)
		}
		if vs := rep.Violations[tree.Name]; len(vs) != 0 {
			t.Errorf("tree %s: %d un-allowlisted citation(s) of %q:", tree.Name, len(vs), forbiddenCitation)
			for _, v := range vs {
				t.Errorf("  %s", v)
			}
		}
	}
}

// --- Direction 4/5 + mirror parity (AC-ECC-015) ---

func TestEvidenceCitation_Visitation(t *testing.T) {
	trees := evidenceCitationTrees()
	rep, err := scanEvidenceCitations(trees, evidenceCitationSubtrees, evidenceCitationAllowlist)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	// 1. both tree roots visited.
	wantTrees := []string{"repo-root", "template-mirror"}
	if strings.Join(rep.Trees, ",") != strings.Join(wantTrees, ",") {
		t.Errorf("visited trees = %v, want %v — a one-tree guard stays green while "+
			"the other tree loses the rule", rep.Trees, wantTrees)
	}

	// 2. subtree set equality, per tree. Set equality rather than a subset:
	//    `agents` (21 files) and `output-styles` (3) both fit inside the floor's
	//    slack, and those two hold this SPEC's strongest cases.
	want := append([]string(nil), evidenceCitationSubtrees...)
	sort.Strings(want)
	for _, tree := range trees {
		got := rep.Subtrees[tree.Name]
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("tree %s: visited subtrees = %v, want exactly %v", tree.Name, got, want)
		}
	}

	// 3. per-tree floor, reported separately — never an aggregate sum.
	for _, tree := range trees {
		if got := rep.Scanned[tree.Name]; got < tree.MinFiles {
			t.Errorf("tree %s: scanned %d, want >= %d", tree.Name, got, tree.MinFiles)
		}
	}

	// 4. mirror parity on this item.
	if a, b := len(rep.Violations["repo-root"]), len(rep.Violations["template-mirror"]); a != b {
		t.Errorf("mirror parity: repo-root has %d violation(s), template-mirror has %d — "+
			"the two copies disagree about this rule", a, b)
	}
}

func TestEvidenceCitation_TreeVisitMutant(t *testing.T) {
	only := evidenceCitationTrees()[:1] // drop the mirror
	rep, err := scanEvidenceCitations(only, evidenceCitationSubtrees, evidenceCitationAllowlist)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	// The same comparison TestEvidenceCitation_Visitation makes must now fail.
	if strings.Join(rep.Trees, ",") == strings.Join([]string{"repo-root", "template-mirror"}, ",") {
		t.Fatalf("dropping a tree root did not change the visited-tree list: %v", rep.Trees)
	}
	// The floor still passes — which is the point: the floor cannot see this.
	if rep.Scanned["repo-root"] < 300 {
		t.Errorf("expected the surviving tree to still clear its floor, got %d", rep.Scanned["repo-root"])
	}
}

func TestEvidenceCitation_SubtreeVisitMutant(t *testing.T) {
	var reduced []string
	for _, s := range evidenceCitationSubtrees {
		if s != "agents" {
			reduced = append(reduced, s)
		}
	}
	rep, err := scanEvidenceCitations(evidenceCitationTrees(), reduced, evidenceCitationAllowlist)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	// The two assertions must move independently, or the subtree assertion is
	// only a restatement of the floor.
	for _, tree := range evidenceCitationTrees() {
		if got := rep.Scanned[tree.Name]; got < tree.MinFiles {
			t.Errorf("tree %s: dropping `agents` fell below the floor (%d < %d); the "+
				"independence demonstration needs the floor to still PASS here",
				tree.Name, got, tree.MinFiles)
		}
		want := append([]string(nil), evidenceCitationSubtrees...)
		sort.Strings(want)
		if strings.Join(rep.Subtrees[tree.Name], ",") == strings.Join(want, ",") {
			t.Errorf("tree %s: subtree set still equals the full set after dropping `agents` — "+
				"the set-equality assertion is not doing its job", tree.Name)
		}
	}
}

// --- Direction 2: synthetic mutant (AC-ECC-011) ---

func TestEvidenceCitation_SyntheticMutant(t *testing.T) {
	const fixture = "" +
		"evidence: .moai/reports/t999/verify/gotest.log\n" +
		"evidence: .moai/state/verify/abc123/gotest.log\n"

	got := scanContent("fixture", "rules/fixture.md", fixture, evidenceCitationAllowlist)
	if len(got) != 1 {
		t.Fatalf("want exactly 1 violation (the old form), got %d: %v", len(got), got)
	}
	if got[0].Line != 2 {
		t.Errorf("flagged line %d, want line 2 — the new tracked-path form must not be flagged", got[0].Line)
	}
	if !strings.Contains(got[0].Text, ".moai/state/verify/abc123") {
		t.Errorf("flagged the wrong line: %s", got[0])
	}
}

// --- Direction 3: real-sentence mutant (AC-ECC-012) ---

func TestEvidenceCitation_RealSentenceMutant(t *testing.T) {
	// Verbatim from agent-common-protocol.md:268 as it stood BEFORE this SPEC's
	// repair. A synthetic fixture only shows the scanner recognising a shape it
	// invented; this shows it catching a sentence the repository actually had.
	const preRepair = "- **Evidence persistence.** The cited path must still resolve at audit " +
		"time, so evidence is persisted under `.moai/state/verify/<session>/` rather than left " +
		"in `/tmp`, which the OS clears. A claim whose cited evidence path no longer resolves " +
		"is an unattributed claim (`verification-claim-integrity.md` §2)."

	got := scanContent("fixture", "rules/moai/core/agent-common-protocol.md", preRepair, evidenceCitationAllowlist)
	if len(got) != 1 {
		t.Fatalf("the pre-repair sentence was not flagged (%d violations) — the guard would "+
			"not have caught the defect it exists for", len(got))
	}
}

// --- Direction 6: allowlist unit + size (AC-ECC-013) ---

func TestEvidenceCitation_Allowlist(t *testing.T) {
	t.Run("Unit", func(t *testing.T) {
		if err := validateAllowlist(evidenceCitationAllowlist); err != nil {
			t.Fatalf("the live allowlist is invalid: %v", err)
		}
		// A file-only entry is the shape that would silently exempt a whole
		// file. Assert the validator refuses it, rather than grepping today's
		// list — a grep is true only of the entries that exist right now.
		bad := []allowEntry{{File: "rules/moai/core/agent-common-protocol.md", Literal: "  "}}
		if err := validateAllowlist(bad); err == nil {
			t.Error("validator accepted an entry with no literal — that entry would exempt the whole file")
		}
		if err := validateAllowlist([]allowEntry{{File: "", Literal: "x"}}); err == nil {
			t.Error("validator accepted an entry with no file")
		}
	})

	t.Run("Size", func(t *testing.T) {
		if got := len(evidenceCitationAllowlist); got != evidenceCitationAllowlistSize {
			t.Errorf("allowlist has %d entries, want %d — growing the allowlist is the "+
				"cheapest way to erase a violation, so its length is pinned",
				got, evidenceCitationAllowlistSize)
		}
	})
}
