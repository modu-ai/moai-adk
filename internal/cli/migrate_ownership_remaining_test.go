package cli

// migrate_ownership_remaining_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M3
// (card t1865), AC-OWR-14: the remaining-file list, and the 313 unregistered
// copies the field run left in `.agents/skills`.
//
// The fixture plants 313 unregistered files whose bytes are read from REAL
// embedded template files (the first 313 of `.claude/skills` in path order,
// placed at the same relative path under `.agents/skills`), so a byte-equality
// promotion would actually fire. Beside them sits a registered managed copy with
// every removal precondition, which the same run removes first: the keep is
// shown to be a decision, not an unreachable removal arm.
//
// Everything is judged twice: from the migration's structured rows (one row per
// file, class, reason, missing evidence items, deterministic path order) and from
// the rendered output of the real update flow (the end-of-run terminal block),
// where the default view is one line per reason with a count and `--verbose`
// adds one row per file. The renderer is exercised only through that flow, so the
// criterion stays testable before and after the renderer exists.
//
// The assertion functions are pure; the mutant sub-tests feed them wrong
// implementations (empty reasons, every path printed by default, a byte-equality
// promoting evidence source) and require a non-empty problem list.

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// plantedCopyCount is the number of unregistered copies the field run left.
const plantedCopyCount = 313

// plantedCopy is one planted unregistered file and its bytes.
type plantedCopy struct {
	rel  string
	data []byte
}

// plantUnregisteredMirrorCopies plants the 313 copies in root. Each copy's bytes
// are those of the embedded template file it was read from; the premise checks
// run here so a fixture that no longer matches the embedded tree fails loudly.
func plantUnregisteredMirrorCopies(t *testing.T, root string) []plantedCopy {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	var sources []string
	walkErr := fs.WalkDir(embedded, ".claude/skills", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			sources = append(sources, p)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("list embedded skill files: %v", walkErr)
	}
	sort.Strings(sources)
	if len(sources) < plantedCopyCount {
		t.Fatalf("the embedded tree holds %d skill files, want at least %d", len(sources), plantedCopyCount)
	}
	planted := make([]plantedCopy, 0, plantedCopyCount)
	for _, src := range sources[:plantedCopyCount] {
		data, readErr := fs.ReadFile(embedded, src)
		if readErr != nil {
			t.Fatalf("embedded lookup of %s: %v", src, readErr)
		}
		rel := ".agents/skills/" + strings.TrimPrefix(src, ".claude/skills/")
		ownAddUnregistered(t, root, rel, data)
		onDisk, diskErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if diskErr != nil || !bytes.Equal(onDisk, data) {
			t.Fatalf("premise: %s does not carry the embedded bytes of %s", rel, src)
		}
		planted = append(planted, plantedCopy{rel: rel, data: data})
	}
	if len(planted) != plantedCopyCount {
		t.Fatalf("planted %d copies, want %d", len(planted), plantedCopyCount)
	}
	return planted
}

func manifestKeySet(t *testing.T, root string) map[string]bool {
	t.Helper()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	keys := map[string]bool{}
	for k := range mgr.Manifest().Files {
		keys[k] = true
	}
	return keys
}

// problemTally folds many identical-shaped problems into one line carrying the
// count and the first example, so a failing criterion over 313 files reports a
// bounded list instead of 313 lines.
type problemTally struct {
	order []string
	count map[string]int
	first map[string]string
}

func newProblemTally() *problemTally {
	return &problemTally{count: map[string]int{}, first: map[string]string{}}
}

func (pt *problemTally) add(category, example string) {
	if pt.count[category] == 0 {
		pt.order = append(pt.order, category)
		pt.first[category] = example
	}
	pt.count[category]++
}

func (pt *problemTally) problems() []string {
	var out []string
	for _, c := range pt.order {
		out = append(out, fmt.Sprintf("%s: %d occurrence(s), first: %s", c, pt.count[c], pt.first[c]))
	}
	return out
}

// remainingListProblems is the pure structural assertion of AC-OWR-14: every
// planted copy is intact, unremoved, unadopted, and listed exactly once as an
// unregistered file kept for want of a provenance record with all three evidence
// items missing; the whole list is in path order and every row has a known
// reason; the manifest gained no key.
func remainingListProblems(root string, run ownershipRun, planted []plantedCopy, keysBefore map[string]bool) []string {
	pt := newProblemTally()
	rows := map[string][]remainingRow{}
	for _, r := range run.rc.remaining {
		rows[r.Rel] = append(rows[r.Rel], r)
	}
	for _, p := range planted {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p.rel)))
		switch {
		case err != nil:
			pt.add("a planted copy is gone or unreadable", p.rel)
		case !bytes.Equal(got, p.data):
			pt.add("a planted copy's bytes changed", p.rel)
		}
		if slices.Contains(run.removed, p.rel) {
			pt.add("a planted copy is reported as removed", p.rel)
		}
		if matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", "removed-assets", filepath.FromSlash(p.rel))); len(matches) != 0 {
			pt.add("a planted copy has a removed-assets copy", p.rel)
		}
		rs := rows[p.rel]
		if len(rs) != 1 {
			pt.add(fmt.Sprintf("a planted copy has %d remaining rows, want exactly 1", len(rs)), p.rel)
			continue
		}
		r := rs[0]
		if r.Class != ownershipUnregistered {
			pt.add(fmt.Sprintf("a planted copy's class is not %q", ownershipUnregistered), fmt.Sprintf("%s is %q", p.rel, r.Class))
		}
		if r.Reason != keepNoProvenanceRecord {
			pt.add(fmt.Sprintf("a planted copy's reason is not %q", keepNoProvenanceRecord), fmt.Sprintf("%s is %q", p.rel, r.Reason))
		}
		if !slices.Equal(r.Missing, allEvidenceItems) {
			pt.add("a planted copy's missing-evidence list is not all three items", fmt.Sprintf("%s has %v", p.rel, r.Missing))
		}
	}
	for i, r := range run.rc.remaining {
		if !isKnownKeepReason(r.Reason) {
			pt.add("a row's reason is empty or outside the closed set", fmt.Sprintf("row %d (%s) is %q", i, r.Rel, r.Reason))
		}
	}
	if !sort.SliceIsSorted(run.rc.remaining, func(i, j int) bool { return run.rc.remaining[i].Rel < run.rc.remaining[j].Rel }) {
		pt.add("the remaining rows are not in path order", "the list as returned")
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err == nil {
		for k := range mgr.Manifest().Files {
			if !keysBefore[k] {
				pt.add("the manifest gained a key", k)
			}
		}
	}
	return pt.problems()
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// terminalBlockLines renders the update run's end-of-run terminal block into
// lines without colour codes.
func terminalBlockLines() []string {
	var b strings.Builder
	renderUpdateTerminalBlock(&b, resolveTheme())
	return strings.Split(strings.TrimRight(ansiRe.ReplaceAllString(b.String(), ""), "\n"), "\n")
}

// defaultRenderProblems is the pure assertion on the DEFAULT view: no planted
// path appears in the terminal block or in the flow's standard output, the block
// carries a line for the reason with a count at least the planted number, and
// the block is not file-sized.
func defaultRenderProblems(block []string, stdout string, planted []plantedCopy) []string {
	var problems []string
	inBlock, inOut := 0, 0
	for _, p := range planted {
		if slices.ContainsFunc(block, func(l string) bool { return strings.Contains(l, p.rel) }) {
			inBlock++
		}
		if strings.Contains(stdout, p.rel) {
			inOut++
		}
	}
	if inBlock != 0 {
		problems = append(problems, fmt.Sprintf("%d planted paths are printed in the default terminal block", inBlock))
	}
	if inOut != 0 {
		problems = append(problems, fmt.Sprintf("%d planted paths are printed in the default standard output", inOut))
	}
	grouped := false
	for _, l := range block {
		if !strings.Contains(l, string(keepNoProvenanceRecord)) {
			continue
		}
		for _, f := range regexp.MustCompile(`\d+`).FindAllString(l, -1) {
			if n, err := strconv.Atoi(f); err == nil && n >= len(planted) {
				grouped = true
			}
		}
	}
	if !grouped {
		problems = append(problems, fmt.Sprintf("no default line names %q with a count of at least %d", keepNoProvenanceRecord, len(planted)))
	}
	if len(block) >= 40 {
		problems = append(problems, fmt.Sprintf("the default terminal block has %d lines; it is file-sized", len(block)))
	}
	return problems
}

// verboseRenderProblems is the pure assertion on the --verbose view: every
// planted path has exactly one row carrying the reason reasonOf names for it, and
// the rows are in path order.
func verboseRenderProblems(block []string, planted []plantedCopy, reasonOf func(rel string) keepReason) []string {
	pt := newProblemTally()
	var order []string
	for _, p := range planted {
		var hits []string
		for _, l := range block {
			if strings.Contains(l, p.rel) {
				hits = append(hits, l)
			}
		}
		if len(hits) != 1 {
			pt.add(fmt.Sprintf("a planted path has %d verbose rows, want exactly 1", len(hits)), p.rel)
			continue
		}
		if want := reasonOf(p.rel); !strings.Contains(hits[0], string(want)) {
			pt.add(fmt.Sprintf("a verbose row does not carry its reason %q", want), hits[0])
		}
		order = append(order, p.rel)
	}
	position := func(rel string) int {
		for i, l := range block {
			if strings.Contains(l, rel) {
				return i
			}
		}
		return -1
	}
	sorted := append([]string(nil), order...)
	sort.Strings(sorted)
	last := -1
	for _, rel := range sorted {
		pos := position(rel)
		if pos < last {
			pt.add("the verbose rows are not in path order", rel)
			break
		}
		last = pos
	}
	return pt.problems()
}

// claudeOriginPromotingSource is the wrong implementation AC-OWR-14 injects: it
// claims all three evidence items for every `.agents/skills` file whose bytes
// equal the embedded `.claude/skills` file at the same relative path.
type claudeOriginPromotingSource struct{ root string }

func (s claudeOriginPromotingSource) generations() []priorGeneration {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return nil
	}
	files := map[string]priorRecord{}
	_ = filepath.WalkDir(filepath.Join(s.root, ".agents", "skills"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(s.root, p)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		src := ".claude/skills/" + strings.TrimPrefix(rel, ".agents/skills/")
		disk, rErr := os.ReadFile(p)
		tpl, tErr := fs.ReadFile(embedded, src)
		if rErr == nil && tErr == nil && bytes.Equal(disk, tpl) {
			h := manifest.HashBytes(disk)
			files[rel] = priorRecord{provenance: manifest.TemplateManaged, templateHash: h, deployedHash: h, deployedVersion: "v-mutant"}
		}
		return nil
	})
	return []priorGeneration{{name: "mutant-claude-origin", files: files}}
}

func TestRemainingListKeeps313UnregisteredCopies(t *testing.T) {
	enableRemovalForTest(t)
	ownProfileSandbox(t)

	t.Run("Premise", func(t *testing.T) {
		root := buildMigrationFixture(t)
		planted := plantUnregisteredMirrorCopies(t, root)
		if len(planted) != plantedCopyCount {
			t.Fatalf("planted %d, want %d", len(planted), plantedCopyCount)
		}
	})

	t.Run("StructuredRowsKeepEveryCopy", func(t *testing.T) {
		home := installMigrationUserCounterparts(t)
		requireUserCounterpart(t, home, migIdenticalSkill)
		root := buildMigrationFixture(t)
		planted := plantUnregisteredMirrorCopies(t, root)
		before := manifestKeySet(t, root)
		run := runOwnershipMigration(t, root, home)
		// The reach control comes first: the registered sibling with every
		// removal precondition is removed in this very run.
		if !slices.Contains(run.removed, migIdenticalSkill) {
			t.Fatalf("the reach control %s was not removed (removed=%v): the keep below would prove nothing", migIdenticalSkill, run.removed)
		}
		for _, p := range remainingListProblems(root, run, planted, before) {
			t.Error(p)
		}
	})

	t.Run("MutantEmptyReasonRowsIsCaught", func(t *testing.T) {
		home := installMigrationUserCounterparts(t)
		root := buildMigrationFixture(t)
		planted := plantUnregisteredMirrorCopies(t, root)
		before := manifestKeySet(t, root)
		run := runOwnershipMigration(t, root, home)
		blanked := make([]remainingRow, len(run.rc.remaining))
		copy(blanked, run.rc.remaining)
		for i := range blanked {
			blanked[i].Reason = ""
		}
		run.rc = &migrationRunContext{remaining: blanked}
		if got := remainingListProblems(root, run, planted, before); len(got) == 0 {
			t.Errorf("the structural assertion stayed green in front of rows with an empty reason")
		}
	})

	t.Run("MutantPrintingEveryPathByDefaultIsCaught", func(t *testing.T) {
		root := buildMigrationFixture(t)
		planted := plantUnregisteredMirrorCopies(t, root)
		var block, out []string
		for _, p := range planted {
			block = append(block, "kept: "+p.rel+" [no-provenance-record]")
			out = append(out, "migration: unregistered mirror copy: "+p.rel)
		}
		if got := defaultRenderProblems(block, strings.Join(out, "\n"), planted); len(got) == 0 {
			t.Errorf("the default-render assertion stayed green in front of a renderer that prints every path")
		}
	})

	t.Run("MutantBytesEqualPromotionIsCaught", func(t *testing.T) {
		home := installMigrationUserCounterparts(t)
		root := buildMigrationFixture(t)
		planted := plantUnregisteredMirrorCopies(t, root)
		before := manifestKeySet(t, root)
		swapProvenanceSource(t, func(projectRoot string) provenanceSource {
			return claudeOriginPromotingSource{root: projectRoot}
		})
		run := runOwnershipMigration(t, root, home)
		if got := remainingListProblems(root, run, planted, before); len(got) == 0 {
			t.Errorf("the structural assertion stayed green in front of a byte-equality promoting evidence source")
		}
	})

	t.Run("RenderedByTheUpdateFlow", func(t *testing.T) {
		home := installMigrationUserCounterparts(t)
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		requireUserCounterpart(t, home, migIdenticalSkill)
		root := buildMigrationFixture(t)
		planted := plantUnregisteredMirrorCopies(t, root)
		prevVerbose := updateVerboseMode
		t.Cleanup(func() { updateVerboseMode = prevVerbose; updateLedger.reset() })

		// Run 1: a plain run (the project version differs from the binary, so it
		// reaches reconciliation), default view.
		updateLedger.reset()
		updateVerboseMode = false
		run1 := runUpdateFlowPlain(t, root)
		if run1.Skipped {
			t.Fatalf("run 1 was skipped before reconciliation; it judged nothing")
		}
		assertFileAbsent(t, root, migIdenticalSkill) // the reach control was removed
		for _, p := range defaultRenderProblems(terminalBlockLines(), run1.Out, planted) {
			t.Errorf("default view: %s", p)
		}

		// Run 2: --force, --verbose. A plain same-version re-run would be skipped,
		// so the run is forced and its reach is asserted (rule 7).
		updateLedger.reset()
		updateVerboseMode = true
		run2 := runUpdateFlowForced(t, root)
		if run2.NewPriorManifest == "" {
			t.Fatalf("run 2 left no new manifest.prior.json; the Backup stage was not reached")
		}
		if got, err := os.ReadFile(run2.NewPriorManifest); err != nil || !bytes.Equal(got, run2.StartManifest) {
			t.Errorf("run 2: manifest.prior.json is not the manifest as the run loaded it (%v)", err)
		}
		block2 := terminalBlockLines()
		// Run 1's start manifest is kept as a prior generation for run 2. It
		// registered two files under `.claude/skills` whose bytes equal two of
		// the planted copies, and design.md D-2 reads a record with the same hash
		// under ANOTHER key as "the original deployed path is missing" — not as
		// "no record at all". Those two rows therefore carry that narrower
		// reason in run 2; every other copy keeps no-provenance-record.
		sameHashElsewhere := map[string]bool{
			".agents/skills/moai-foundation-core/SKILL.md": true, // sibling of migIdenticalSkill
			".agents/skills/moai-workflow-tdd/SKILL.md":    true, // sibling of migModifiedSkill
		}
		reasonOf := func(rel string) keepReason {
			if sameHashElsewhere[rel] {
				return evidenceIncompleteReason(evidenceOriginalPath)
			}
			return keepNoProvenanceRecord
		}
		for _, p := range verboseRenderProblems(block2, planted, reasonOf) {
			t.Errorf("verbose view: %s", p)
		}

		for _, p := range planted {
			got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p.rel)))
			if err != nil || !bytes.Equal(got, p.data) {
				t.Errorf("%s did not survive the two update runs byte for byte (%v)", p.rel, err)
			}
		}
	})

	t.Run("MixedReasonsEveryRowHasOne", func(t *testing.T) {
		home := installMigrationUserCounterparts(t)
		root := buildMigrationFixture(t)
		for _, rel := range []string{profAgentA, profAgentB} {
			ownAddRegisteredManaged(t, root, rel, ownEmbeddedBytes(t, rel))
		}
		ownAddRegisteredUserCreated(t, root, ".claude/skills/mine/SKILL.md", []byte("# mine\n"))
		for i := range 5 {
			ownAddUnregistered(t, root, fmt.Sprintf(".agents/skills/moai-custom-%d/SKILL.md", i), []byte("# unknown origin\n"))
		}
		mapProjectToWork(t, root)
		run := runOwnershipMigration(t, root, home)
		reasons := map[keepReason]int{}
		for _, r := range run.rc.remaining {
			if !isKnownKeepReason(r.Reason) {
				t.Errorf("%s: reason %q is empty or outside the closed set", r.Rel, r.Reason)
			}
			reasons[r.Reason]++
		}
		for _, want := range []keepReason{keepProfileReach, keepUserModified, keepUserCreated, keepNoProvenanceRecord} {
			if reasons[want] == 0 {
				t.Errorf("no row carries the reason %q in the mixed fixture (reasons: %v)", want, reasons)
			}
		}
		if len(reasons) < 4 {
			t.Errorf("the mixed fixture produced %d distinct reasons, want at least 4", len(reasons))
		}
	})
}
