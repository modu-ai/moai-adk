// reach_test.go — the reachability checker self-test (AC-MFB-013) and the
// link-classification unit tests of SPEC-MEMORY-FOLD-BUDGET-001 §1.5.
//
// The checker is falsified against the SPEC's frozen fixture store-A: the
// correct fold must pass all four invariants, five link-losing mutants must
// fail (a)-(c), and three exact-plan mutants must pass (a)-(c) and fail (d)
// alone. Every store here is in memory or the fixture read read-only; no test
// touches a store outside the fixture and t.TempDir().
package taxonomy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureStoreRel is the SPEC's frozen fixture, read read-only (REQ-MFB-012).
// Relative to this package directory.
const fixtureStoreRel = "../../../../.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A"

// fixtureArchiveIndex is the fixture's only archive-pattern file.
const fixtureArchiveIndex = "project_card_archive_2026_10.md"

// The fold-plan lines are FIXED in this test from the acceptance §1
// classification table, never read back from the code under test
// (AC-MFB-003). They are the fixture's MEMORY.md lines byte for byte.
var (
	line9001 = "- [t9001 alpha card — done, merged](project_card_t9001_alpha.md) — merged; detail lives in the topic file"
	line9002 = "- [t9002 beta card — done](.moai/reports/t9002/verdict.md) — the verdict path is repo-relative"
	line9006 = "- [t9006 zeta card — done](.moai/reports/t9006/verdict.md) — a second repo-relative verdict path with the same base name"

	// Link-free lines the exact-plan mutant m6 deletes on top of the fold.
	lineBlockquote = "> Synthetic fixture for SPEC-MEMORY-FOLD-BUDGET-001. Not a real memory store."
	lineOpenWork   = "## Open work"
)

func loadFixtureStore(t *testing.T) StoreSnapshot {
	t.Helper()
	snap, err := SnapshotStore(fixtureStoreRel)
	if err != nil {
		t.Fatalf("SnapshotStore(%s): %v", fixtureStoreRel, err)
	}
	if len(snap) == 0 {
		t.Fatalf("fixture store %s loaded empty: the self-test measured nothing", fixtureStoreRel)
	}
	return snap
}

func copySnapshot(s StoreSnapshot) StoreSnapshot {
	out := make(StoreSnapshot, len(s))
	for k, v := range s {
		out[k] = append([]byte(nil), v...)
	}
	return out
}

// removeLine drops one occurrence of line (with its newline) from data.
func removeLine(data []byte, line string) []byte {
	out := strings.Replace(string(data), line+"\n", "", 1)
	if out == string(data) {
		return data // unchanged: caller asserts the removal happened
	}
	return []byte(out)
}

// appendLines appends lines the way invariant (d2) expects the fold to: after
// one terminator when the content lacked one, each line with the file's
// existing (\n) terminator.
func appendLines(data []byte, lines ...string) []byte {
	out := string(data)
	if len(out) > 0 && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	for _, l := range lines {
		out += l + "\n"
	}
	return []byte(out)
}

// violationSet maps the checker's result for set-style assertions.
func violationSet(violations []string) map[string]bool {
	out := make(map[string]bool, len(violations))
	for _, v := range violations {
		out[v] = true
	}
	return out
}

func assertViolations(t *testing.T, violations []string, want ...string) {
	t.Helper()
	got := violationSet(violations)
	for _, w := range want {
		if !got[w] {
			t.Fatalf("violations %v do not name invariant (%s)", violations, w)
		}
	}
}

func assertNoViolations(t *testing.T, violations []string, notWant ...string) {
	t.Helper()
	got := violationSet(violations)
	for _, n := range notWant {
		if got[n] {
			t.Fatalf("violations %v unexpectedly name invariant (%s)", violations, n)
		}
	}
}

// correctFold returns the correct fold of t9001 over the fixture.
func correctFold(before StoreSnapshot) (StoreSnapshot, FoldPlan) {
	after := copySnapshot(before)
	after["MEMORY.md"] = removeLine(before["MEMORY.md"], line9001)
	after[fixtureArchiveIndex] = appendLines(before[fixtureArchiveIndex], line9001)
	plan := FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}}
	return after, plan
}

// TestReach_CorrectFoldPasses is AC-MFB-013's positive side: the correct fold
// passes all four invariants, and the fixture's measured model numbers
// (acceptance §1) hold.
func TestReach_CorrectFoldPasses(t *testing.T) {
	before := loadFixtureStore(t)

	if got := len(before["MEMORY.md"]); got != 1155 {
		t.Fatalf("fixture MEMORY.md = %d bytes, want 1155 (acceptance §1); the fixture drifted", got)
	}
	is := before.IndexSet()
	if len(is) != 2 || !is["MEMORY.md"] || !is[fixtureArchiveIndex] {
		t.Fatalf("I(S) = %v, want {MEMORY.md, %s} (acceptance §1)", is, fixtureArchiveIndex)
	}
	if r := before.ReachableSet(); len(r) != 11 {
		t.Fatalf("|R| = %d, want 11 (acceptance §1)", len(r))
	}
	if ts := before.TargetSet(); len(ts) != 14 {
		t.Fatalf("|T| = %d, want 14 (acceptance §1)", len(ts))
	}

	after, plan := correctFold(before)
	if violations := CheckFoldInvariants(before, after, plan); len(violations) != 0 {
		t.Fatalf("correct fold violates %v; all four invariants must hold", violations)
	}
	// The fold does shrink MEMORY.md — the decrease is information, and the
	// pass is never satisfied by it (§1.5: size is the motive, never a gate).
	if len(after["MEMORY.md"]) >= len(before["MEMORY.md"]) {
		t.Fatalf("fixture drift: the correct fold no longer shrinks MEMORY.md (%d → %d)", len(before["MEMORY.md"]), len(after["MEMORY.md"]))
	}
}

// TestReach_DeleteMutantFails: the line is removed from MEMORY.md and filed
// nowhere — (a), (b) and (c) must fail.
func TestReach_DeleteMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	after := copySnapshot(before)
	after["MEMORY.md"] = removeLine(before["MEMORY.md"], line9001)

	violations := CheckFoldInvariants(before, after, FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	assertViolations(t, violations, "a", "b", "c")
}

// TestReach_DropLinkMutantFails: the moved line loses its link target.
func TestReach_DropLinkMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	after := copySnapshot(before)
	after["MEMORY.md"] = removeLine(before["MEMORY.md"], line9001)
	after[fixtureArchiveIndex] = appendLines(before[fixtureArchiveIndex],
		"- t9001 alpha card — done, merged — detail lives in the topic file")

	violations := CheckFoldInvariants(before, after, FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	assertViolations(t, violations, "a", "b", "c")
}

// TestReach_UnlinkedFileMutantFails: the line is filed into a store file no
// index carries.
func TestReach_UnlinkedFileMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	after := copySnapshot(before)
	after["MEMORY.md"] = removeLine(before["MEMORY.md"], line9001)
	after["feedback_orphan.md"] = appendLines(before["feedback_orphan.md"], line9001)

	violations := CheckFoldInvariants(before, after, FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	assertViolations(t, violations, "a", "b", "c")
}

// TestReach_TruncateMutantFails: the correct fold plus a truncation of the
// MEMORY.md tail — reach and targets are lost.
func TestReach_TruncateMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	after := copySnapshot(before)
	mem := removeLine(before["MEMORY.md"], line9001)
	if len(mem) <= 441 {
		t.Fatalf("fixture drift: folded MEMORY.md is %d bytes, too small to truncate to 441", len(mem))
	}
	after["MEMORY.md"] = mem[:441]
	after[fixtureArchiveIndex] = appendLines(before[fixtureArchiveIndex], line9001)

	violations := CheckFoldInvariants(before, after, FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	assertViolations(t, violations, "a", "b")
}

// TestReach_NonIndexTopicMutantFails: the STRONG line is filed into a linked
// topic file that carries fewer than the doctor's threshold of resolved links,
// so it is not an index — the mutant fails (a), (b) and (c).
func TestReach_NonIndexTopicMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	after := copySnapshot(before)
	after["MEMORY.md"] = removeLine(before["MEMORY.md"], line9001)
	after["feedback_verify.md"] = appendLines(before["feedback_verify.md"], line9001)

	violations := CheckFoldInvariants(before, after, FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	assertViolations(t, violations, "a", "b", "c")
}

// TestReach_LinkFreeLineMutantFails (exact-plan mutant m6): the correct fold
// plus deletion of a link-free blockquote line and the "## Open work" heading.
// Invariants (a)-(c) hold; (d1) alone fails.
func TestReach_LinkFreeLineMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	after := copySnapshot(before)
	mem := removeLine(before["MEMORY.md"], line9001)
	mem = removeLine(mem, lineBlockquote)
	mem = removeLine(mem, lineOpenWork)
	if string(mem) == string(before["MEMORY.md"]) {
		t.Fatal("fixture drift: the link-free lines were not found, mutant not applied")
	}
	after["MEMORY.md"] = mem
	after[fixtureArchiveIndex] = appendLines(before[fixtureArchiveIndex], line9001)

	violations := CheckFoldInvariants(before, after, FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	assertViolations(t, violations, "d1")
	assertNoViolations(t, violations, "a", "b", "c")
	if len(violations) != 1 {
		t.Fatalf("violations = %v, want exactly [d1]", violations)
	}
}

// TestReach_UnlinkedHubMutantFails (exact-plan mutant m7): the store carries
// a planted hub — three resolved links, linked from no index, so it qualifies
// as an index on its own — and the fold files the three STRONG lines into the
// hub instead of the archive index. (a)-(c) hold; (d2) fails.
func TestReach_UnlinkedHubMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	before["feedback_hub.md"] = []byte("---\nname: hub\ndescription: planted hub\ntype: reference\n---\n\n" +
		"- [alpha](feedback_alpha.md) — hub link\n" +
		"- [beta](feedback_beta.md) — hub link\n" +
		"- [hangul](feedback_hangul.md) — hub link\n")
	if !before.IndexSet()["feedback_hub.md"] {
		t.Fatalf("fixture error: the planted hub does not qualify as an index (needs %d resolved links)", SecondaryIndexLinkThreshold())
	}

	removed := []string{line9001, line9002, line9006}
	after := copySnapshot(before)
	mem := before["MEMORY.md"]
	for _, l := range removed {
		mem = removeLine(mem, l)
	}
	after["MEMORY.md"] = mem
	after["feedback_hub.md"] = appendLines(before["feedback_hub.md"], removed...)

	violations := CheckFoldInvariants(before, after, FoldPlan{Removed: removed, Appended: removed})
	assertViolations(t, violations, "d2")
	assertNoViolations(t, violations, "a", "b", "c")
}

// TestReach_StrayEditMutantFails (exact-plan mutant m8): the correct fold plus
// one stray byte in a non-index topic file. Only (d3) fails.
func TestReach_StrayEditMutantFails(t *testing.T) {
	before := loadFixtureStore(t)
	after, plan := correctFold(before)
	after["feedback_alpha.md"] = append(before["feedback_alpha.md"], 'x')

	violations := CheckFoldInvariants(before, after, plan)
	assertViolations(t, violations, "d3")
	assertNoViolations(t, violations, "a", "b", "c")
	if len(violations) != 1 {
		t.Fatalf("violations = %v, want exactly [d3]", violations)
	}
}

// writeBoundaryStore builds a store whose hub file carries exactly links
// resolved links, read back through SnapshotStore.
func writeBoundaryStore(t *testing.T, links int) StoreSnapshot {
	t.Helper()
	dir := t.TempDir()

	var idx strings.Builder
	idx.WriteString("# Memory Index\n\n- [hub](hub.md) — the hub under test\n")
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(idx.String()), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	var hub strings.Builder
	for i := 0; i < links; i++ {
		name := fmt.Sprintf("t%d.md", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("topic\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		fmt.Fprintf(&hub, "- [t%d](t%d.md) — resolved link\n", i, i)
	}
	if err := os.WriteFile(filepath.Join(dir, "hub.md"), []byte(hub.String()), 0o644); err != nil {
		t.Fatalf("write hub.md: %v", err)
	}

	snap, err := SnapshotStore(dir)
	if err != nil {
		t.Fatalf("SnapshotStore: %v", err)
	}
	return snap
}

// TestReach_ThresholdBoundary flips index membership at threshold−1 versus
// threshold resolved links, read through the accessor (never a literal).
func TestReach_ThresholdBoundary(t *testing.T) {
	th := SecondaryIndexLinkThreshold()
	if th < 2 {
		t.Fatalf("threshold %d unexpectedly small; the boundary test needs th >= 2", th)
	}

	below := writeBoundaryStore(t, th-1)
	if below.IndexSet()["hub.md"] {
		t.Errorf("hub with threshold−1 (%d) resolved links must not be an index member", th-1)
	}
	at := writeBoundaryStore(t, th)
	if !at.IndexSet()["hub.md"] {
		t.Errorf("hub with threshold (%d) resolved links must be an index member", th)
	}
}

// TestReach_SizeAloneNeverPasses: every lossy transformation shrinks
// MEMORY.md by at least as much as the correct fold, and every one of them
// fails the checker — a size decrease is never a pass (§1.5).
func TestReach_SizeAloneNeverPasses(t *testing.T) {
	before := loadFixtureStore(t)
	correctAfter, _ := correctFold(before)
	correctSize := len(correctAfter["MEMORY.md"])

	build := map[string]func(StoreSnapshot) StoreSnapshot{
		"delete": func(s StoreSnapshot) StoreSnapshot {
			a := copySnapshot(s)
			a["MEMORY.md"] = removeLine(s["MEMORY.md"], line9001)
			return a
		},
		"drop-link": func(s StoreSnapshot) StoreSnapshot {
			a := copySnapshot(s)
			a["MEMORY.md"] = removeLine(s["MEMORY.md"], line9001)
			a[fixtureArchiveIndex] = appendLines(s[fixtureArchiveIndex], "- t9001 alpha card — done, merged — detail lives in the topic file")
			return a
		},
		"unlinked-file": func(s StoreSnapshot) StoreSnapshot {
			a := copySnapshot(s)
			a["MEMORY.md"] = removeLine(s["MEMORY.md"], line9001)
			a["feedback_orphan.md"] = appendLines(s["feedback_orphan.md"], line9001)
			return a
		},
		"truncate": func(s StoreSnapshot) StoreSnapshot {
			a := copySnapshot(s)
			a["MEMORY.md"] = removeLine(s["MEMORY.md"], line9001)[:441]
			a[fixtureArchiveIndex] = appendLines(s[fixtureArchiveIndex], line9001)
			return a
		},
		"non-index-topic": func(s StoreSnapshot) StoreSnapshot {
			a := copySnapshot(s)
			a["MEMORY.md"] = removeLine(s["MEMORY.md"], line9001)
			a["feedback_verify.md"] = appendLines(s["feedback_verify.md"], line9001)
			return a
		},
		"link-free": func(s StoreSnapshot) StoreSnapshot {
			a := copySnapshot(s)
			mem := removeLine(removeLine(s["MEMORY.md"], line9001), lineBlockquote)
			a["MEMORY.md"] = removeLine(mem, lineOpenWork)
			a[fixtureArchiveIndex] = appendLines(s[fixtureArchiveIndex], line9001)
			return a
		},
		"unlinked-hub": func(s StoreSnapshot) StoreSnapshot {
			s = copySnapshot(s) // never mutate the shared before-snapshot
			s["feedback_hub.md"] = []byte("---\nname: hub\ndescription: planted hub\ntype: reference\n---\n\n" +
				"- [alpha](feedback_alpha.md) — hub link\n" +
				"- [beta](feedback_beta.md) — hub link\n" +
				"- [hangul](feedback_hangul.md) — hub link\n")
			removed := []string{line9001, line9002, line9006}
			a := copySnapshot(s)
			mem := s["MEMORY.md"]
			for _, l := range removed {
				mem = removeLine(mem, l)
			}
			a["MEMORY.md"] = mem
			a["feedback_hub.md"] = appendLines(s["feedback_hub.md"], removed...)
			return a
		},
		"stray-edit": func(s StoreSnapshot) StoreSnapshot {
			a, _ := correctFold(s)
			a["feedback_alpha.md"] = append(s["feedback_alpha.md"], 'x')
			return a
		},
	}

	plan1 := FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}}
	plan3 := FoldPlan{Removed: []string{line9001, line9002, line9006}, Appended: []string{line9001, line9002, line9006}}
	plans := map[string]FoldPlan{
		"delete": plan1, "drop-link": plan1, "unlinked-file": plan1,
		"truncate": plan1, "non-index-topic": plan1, "link-free": plan1,
		"unlinked-hub": plan3, "stray-edit": plan1,
	}

	for name, buildFn := range build {
		after := buildFn(before)
		if got := len(after["MEMORY.md"]); got > correctSize {
			t.Errorf("mutant %s leaves MEMORY.md at %d bytes, larger than the correct fold's %d", name, got, correctSize)
		}
		if v := CheckFoldInvariants(before, after, plans[name]); len(v) == 0 {
			t.Errorf("mutant %s passes the checker: a transformation that only reduces bytes or lines never passes", name)
		}
	}
}

// TestClassifyLinkTarget covers §1.5's three-way classification.
func TestClassifyLinkTarget(t *testing.T) {
	cases := []struct {
		target string
		want   LinkClass
	}{
		{"project_card_t9001_alpha.md", LinkStoreLocal},
		{"./project_card_t9001_alpha.md", LinkStoreLocal},
		{"/Users/goos/abs/verdict.md", LinkAbsolute},
		{"/verdict.md", LinkAbsolute},
		{".moai/reports/t9002/verdict.md", LinkRepoRelative},
		{"../up/verdict.md", LinkRepoRelative},
		{`a\b.md`, LinkRepoRelative},
	}
	for _, tc := range cases {
		if got := ClassifyLinkTarget(tc.target); got != tc.want {
			t.Errorf("ClassifyLinkTarget(%q) = %q, want %q", tc.target, got, tc.want)
		}
	}
}

// TestExtractLinkTargets: full target text, any line shape, .md only.
func TestExtractLinkTargets(t *testing.T) {
	content := "# T\n" +
		"- [a](one.md) — x and mid-line [b](sub/two.md) on one line\n" +
		"- ![img](pic.png) and [code](`x.md`) shapes\n" +
		"- [c](./three.md)\n"
	got := ExtractLinkTargets(content)
	want := []string{"one.md", "sub/two.md", "./three.md"}
	if len(got) != len(want) {
		t.Fatalf("ExtractLinkTargets = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ExtractLinkTargets = %v, want %v", got, want)
		}
	}
}
