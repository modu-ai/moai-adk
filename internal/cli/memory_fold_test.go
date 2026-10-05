// memory_fold_test.go — the fold verb (SPEC-MEMORY-FOLD-BUDGET-001 plan M3,
// acceptance AC-MFB-001 … AC-MFB-007).
//
// Every test drives `moai memory fold` through its cobra command against a
// read-only copy of the SPEC fixture or a generated store under t.TempDir()
// — never the operator's real store (REQ-MFB-012). The expected `removed`
// and `appended` lists are FIXED below from the acceptance §1 classification
// table, never read back from the command under test (AC-MFB-003).
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// The fixture's MEMORY.md lines, fixed from acceptance.md §1 byte for byte
// and never read back from the command under test.
var (
	line9001      = "- [t9001 alpha card — done, merged](project_card_t9001_alpha.md) — merged; detail lives in the topic file"
	line9002      = "- [t9002 beta card — done](.moai/reports/t9002/verdict.md) — the verdict path is repo-relative"
	line9006      = "- [t9006 zeta card — done](.moai/reports/t9006/verdict.md) — a second repo-relative verdict path with the same base name"
	line9004      = "- [t9004 delta note](feedback_delta_note.md) — title leads with a card id but the target names no card"
	line9005      = "- [Epsilon wrap-up](project_card_t9005_epsilon.md) — target names a card but the title does not"
	lineVerify    = "- [Verify before claiming](feedback_verify.md) — a lesson first recorded on t9001"
	lineArchive10 = "- [Card archive 2026-10](project_card_archive_2026_10.md) — closed cards for the month"
	lineBlock     = "> Synthetic fixture for SPEC-MEMORY-FOLD-BUDGET-001. Not a real memory store."

	fixtureArchive    = "project_card_archive_2026_10.md"
	fixtureOlderIndex = "project_card_archive_2026_09.md"
)

// foldResult carries one cobra invocation's three channels.
type foldResult struct {
	stdout string
	stderr string
	err    error
}

func runMemoryFold(t *testing.T, args ...string) foldResult {
	t.Helper()
	var out, errBuf bytes.Buffer
	cmd := newMemoryFoldCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return foldResult{stdout: out.String(), stderr: errBuf.String(), err: err}
}

// runFoldOK runs a fold expected to exit 0 and returns its stdout.
func runFoldOK(t *testing.T, args ...string) string {
	t.Helper()
	r := runMemoryFold(t, args...)
	if r.err != nil {
		t.Fatalf("memory fold %s exited with error: %v\nstderr: %s", strings.Join(args, " "), r.err, r.stderr)
	}
	return r.stdout
}

// runFoldRefused runs a fold expected to exit non-zero.
func runFoldRefused(t *testing.T, args ...string) foldResult {
	t.Helper()
	r := runMemoryFold(t, args...)
	if r.err == nil {
		t.Fatalf("memory fold %s exited 0, want non-zero\nstdout: %s", strings.Join(args, " "), r.stdout)
	}
	return r
}

// foldPlanJSON decodes the --json plan.
type foldPlanJSON struct {
	Store struct {
		Dir    string `json:"dir"`
		Origin string `json:"origin"`
	} `json:"store"`
	Card     string   `json:"card"`
	Archive  string   `json:"archive_index"`
	Removed  []string `json:"removed"`
	Appended []string `json:"appended"`
	Kept     []struct {
		Line   string `json:"line"`
		Class  string `json:"class"`
		Reason string `json:"reason"`
	} `json:"kept"`
	Unlinked []string `json:"unlinked_archive_files"`
	Applied  bool     `json:"applied"`
}

func decodeFoldPlan(t *testing.T, out string) foldPlanJSON {
	t.Helper()
	var plan foldPlanJSON
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &plan); err != nil {
		t.Fatalf("decode fold plan JSON: %v\noutput: %s", err, out)
	}
	return plan
}

// foldKeptEntry finds one kept entry by its line text.
func foldKeptEntry(t *testing.T, plan foldPlanJSON, line string) (class, reason string) {
	t.Helper()
	for _, k := range plan.Kept {
		if k.Line == line {
			return k.Class, k.Reason
		}
	}
	t.Fatalf("kept list %+v does not carry %q", plan.Kept, line)
	return "", ""
}

// seedFoldStore lays out a store under t.TempDir(): MEMORY.md with memory
// plus one file per entry of files.
func seedFoldStore(t *testing.T, memory string, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(memory), 0o644); err != nil {
		t.Fatalf("write MEMORY.md: %v", err)
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// minimalMemory is the index of the smallest foldable store: the archive
// link plus one foldable line.
func minimalMemory(foldLine string) string {
	return "# Memory Index\n\n" +
		"- [Archive](project_card_archive_2026_10.md) — closed cards\n" +
		foldLine + "\n"
}

// minimalArchive is an archive index carrying exactly the doctor's threshold
// of resolved links (feedback_a/b/c).
func minimalArchive() string {
	return "---\nname: a\ndescription: d\ntype: reference\n---\n" +
		"- [a](feedback_a.md) — one\n" +
		"- [b](feedback_b.md) — two\n" +
		"- [c](feedback_c.md) — three\n"
}

// crlfArchive is minimalArchive with Windows line endings.
func crlfArchive() string {
	return "---\r\nname: a\r\ndescription: d\r\ntype: reference\r\n---\r\n" +
		"- [a](feedback_a.md) — one\r\n" +
		"- [b](feedback_b.md) — two\r\n" +
		"- [c](feedback_c.md) — three\r\n"
}

// minimalArchiveNoFinalNL is an archive index lacking a final line
// terminator (§6 edge case: one is supplied before the appended lines).
func minimalArchiveNoFinalNL() string {
	return strings.TrimSuffix(minimalArchive(), "\n")
}

func minimalFiles() map[string]string {
	return map[string]string{
		"feedback_a.md":               "---\n---\n",
		"feedback_b.md":               "---\n---\n",
		"feedback_c.md":               "---\n---\n",
		"project_card_t9001_alpha.md": "---\n---\nbody\n",
	}
}

// snapshotDir reads dir through the M1 SnapshotStore.
func foldSnapshot(t *testing.T, dir string) taxonomy.StoreSnapshot {
	t.Helper()
	snap, err := taxonomy.SnapshotStore(dir)
	if err != nil {
		t.Fatalf("SnapshotStore(%s): %v", dir, err)
	}
	return snap
}

// foldRead reads one store file, failing the test on error.
func foldRead(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s/%s: %v", dir, name, err)
	}
	return string(data)
}

// requireSameStore asserts two hash maps describe the identical store.
func requireSameStore(t *testing.T, before, after map[string]string) {
	t.Helper()
	if len(before) != len(after) {
		t.Fatalf("store file list changed: %d files before, %d after", len(before), len(after))
	}
	for name, sum := range before {
		if after[name] != sum {
			t.Errorf("%s changed across the run", name)
		}
	}
}

// requireNoTempFiles asserts the fold left no temporary file behind
// (AC-MFB-001, AC-MFB-007).
func requireNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read store %s: %v", dir, err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".moai-fold") {
			t.Errorf("temporary file %s remained after the fold", e.Name())
		}
	}
}

// foldSetsHold asserts invariants (a) and (b) hold between two snapshots —
// the property the interrupted-apply cell (AC-MFB-007 ii) requires.
func foldSetsHold(t *testing.T, before, after taxonomy.StoreSnapshot) {
	t.Helper()
	reachBefore, reachAfter := before.ReachableSet(), after.ReachableSet()
	for f := range reachBefore {
		if !reachAfter[f] {
			t.Errorf("invariant (a): %s became unreachable", f)
		}
	}
	targetsBefore, targetsAfter := before.TargetSet(), after.TargetSet()
	for tgt := range targetsBefore {
		if !targetsAfter[tgt] {
			t.Errorf("invariant (b): target %s disappeared from the index set", tgt)
		}
	}
}

// removeFoldTestLines drops one occurrence of the line (with its terminator)
// from content — the test-side (d1) rebuild, independent of the command.
func removeFoldTestLines(content, line string) string {
	out := strings.Replace(content, line+"\n", "", 1)
	if out != content {
		return out
	}
	return strings.Replace(content, line+"\r\n", "", 1)
}

// appendFoldTestLines appends lines the way (d2) expects: one terminator
// when the content lacked one, then each line with the file's existing
// line ending.
func appendFoldTestLines(content string, lines ...string) string {
	ending := "\n"
	if strings.Contains(content, "\r\n") {
		ending = "\r\n"
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += ending
	}
	for _, l := range lines {
		content += l + ending
	}
	return content
}

// assertExactFold is the exactness assertion (AC-MFB-003 (d), AC-MFB-004):
// the checker's (a)-(d) verdict plus the test-side rebuilds of (d1) and
// (d2) from the FIXED plan, and (d3) file-for-file.
func assertExactFold(t *testing.T, before, after taxonomy.StoreSnapshot, plan taxonomy.FoldPlan) {
	t.Helper()
	if v := taxonomy.CheckFoldInvariants(before, after, plan); len(v) != 0 {
		t.Fatalf("fold violates invariants %v of the reachability model", v)
	}
	foldSetsHold(t, before, after)

	// (d1) MEMORY.md equals its prior bytes minus exactly the planned lines.
	wantMem := string(before["MEMORY.md"])
	for _, l := range plan.Removed {
		next := removeFoldTestLines(wantMem, l)
		if next == wantMem {
			t.Fatalf("test-side (d1) rebuild: planned line %q not found in the before snapshot", l)
		}
		wantMem = next
	}
	if got := string(after["MEMORY.md"]); got != wantMem {
		t.Fatalf("(d1): MEMORY.md does not equal prior bytes minus exactly the planned lines\nwant %q\ngot  %q", wantMem, got)
	}

	// (d2) the archive index equals its prior bytes plus exactly the
	// appended lines verbatim at the end.
	dest := before.ArchiveIndexName()
	if dest == "" {
		t.Fatal("no archive index in the before snapshot; (d2) unmeasurable")
	}
	wantArch := appendFoldTestLines(string(before[dest]), plan.Appended...)
	if got := string(after[dest]); got != wantArch {
		t.Fatalf("(d2): archive %s does not equal prior bytes plus the appended lines verbatim\nwant %q\ngot  %q", dest, wantArch, got)
	}

	// (d3) every other file is byte-identical and no file appeared or went.
	if len(before) != len(after) {
		t.Fatalf("(d3): store file list changed: %d before, %d after", len(before), len(after))
	}
	for name, data := range before {
		afterData, ok := after[name]
		if !ok {
			t.Fatalf("(d3): file %s disappeared", name)
		}
		if name == "MEMORY.md" || name == dest {
			continue
		}
		if string(data) != string(afterData) {
			t.Errorf("(d3): %s changed across the fold", name)
		}
	}
}

// orphanPaths lists the store's MEMORY_ORPHAN_NOT_INDEXED paths through the
// in-process doctor linkage audit.
func orphanPaths(t *testing.T, dir string) []string {
	t.Helper()
	findings, err := taxonomy.AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage(%s): %v", dir, err)
	}
	var out []string
	for _, f := range findings {
		if f.Code == "MEMORY_ORPHAN_NOT_INDEXED" {
			out = append(out, f.Path)
		}
	}
	sort.Strings(out)
	return out
}

// requireNoNewOrphans is the doctor-agreement cell of AC-MFB-003 and
// AC-MFB-006: the fold reports no MEMORY_ORPHAN_NOT_INDEXED finding the
// store did not carry before it. A fold that makes an under-threshold
// archive cross the threshold legitimately shrinks the set (its previously
// orphaned targets become reachable); a growth is the failure.
func requireNoNewOrphans(t *testing.T, dir string, base []string) {
	t.Helper()
	baseSet := map[string]bool{}
	for _, p := range base {
		baseSet[p] = true
	}
	for _, p := range orphanPaths(t, dir) {
		if !baseSet[p] {
			t.Fatalf("the fold created a new orphan: %s (base %v)", p, base)
		}
	}
}

// requireOrphansUnchanged is the exact-set form: the orphan set is the same
// before and after. Used where no threshold crossing can occur.
func requireOrphansUnchanged(t *testing.T, dir string, base []string) {
	t.Helper()
	now := orphanPaths(t, dir)
	if strings.Join(now, "|") != strings.Join(base, "|") {
		t.Fatalf("orphan set changed across the fold: base %v → now %v", base, now)
	}
}

// fixtureMultiLineVariant is the D39 variant: the fixture's t9001 gains a
// second STRONG line two lines below the first, one non-STRONG line between
// them.
func fixtureMultiLineVariant(t *testing.T) (dir, second, interim string) {
	t.Helper()
	dir = specFixtureCopy(t)
	second = "- [t9001 second entry](project_card_t9001_second.md) — second STRONG line"
	interim = "t9001 interim note with no link shape"
	mem := foldRead(t, dir, "MEMORY.md")
	edited := strings.Replace(mem, line9001+"\n", line9001+"\n"+interim+"\n"+second+"\n", 1)
	if edited == mem {
		t.Fatal("fixture edit failed: the first STRONG line was not found")
	}
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(edited), 0o644); err != nil {
		t.Fatalf("write variant MEMORY.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "project_card_t9001_second.md"), []byte("---\n---\nsecond\n"), 0o644); err != nil {
		t.Fatalf("write second topic: %v", err)
	}
	return dir, second, interim
}

// fixtureTwoArchives is the fixture plus a second (older) archive index,
// present and linked — both qualify, so A(S) is the greater name.
func fixtureTwoArchives(t *testing.T) string {
	t.Helper()
	dir := specFixtureCopy(t)
	older := "---\nname: old\ndescription: d\ntype: reference\n---\n" +
		"- [t9000 zero card](project_card_t9000_zero.md) — closed earlier\n" +
		"- [t8999 prior card](project_card_t8999_prior.md) — closed earlier\n" +
		"- [t8998 older card](project_card_t8998_older.md) — closed earlier\n"
	if err := os.WriteFile(filepath.Join(dir, fixtureOlderIndex), []byte(older), 0o644); err != nil {
		t.Fatalf("write older archive: %v", err)
	}
	mem := foldRead(t, dir, "MEMORY.md")
	edited := strings.Replace(mem, lineArchive10+"\n",
		lineArchive10+"\n- [Card archive 2026-09](project_card_archive_2026_09.md) — older month\n", 1)
	if edited == mem {
		t.Fatal("fixture edit failed: the 2026-10 archive link was not found")
	}
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(edited), 0o644); err != nil {
		t.Fatalf("write MEMORY.md: %v", err)
	}
	return dir
}

// fixtureRolloverVariant is the month-rollover store: the older archive is
// the linked one; the newer is present but unlinked (§1.5: never A(S)).
func fixtureRolloverVariant(t *testing.T) string {
	t.Helper()
	dir := fixtureTwoArchives(t)
	mem := foldRead(t, dir, "MEMORY.md")
	edited := strings.Replace(mem, lineArchive10+"\n", "", 1)
	if edited == mem {
		t.Fatal("fixture edit failed: the 2026-10 archive link was not found")
	}
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(edited), 0o644); err != nil {
		t.Fatalf("write MEMORY.md: %v", err)
	}
	return dir
}

// fixtureDifferingText pre-files an archive line with the t9001 line's
// link-target set on a different wording (REQ-MFB-005).
func fixtureDifferingText(t *testing.T) string {
	t.Helper()
	dir := specFixtureCopy(t)
	diff := strings.TrimSuffix(foldRead(t, dir, fixtureArchive), "\n") + "\n" +
		"- [t9001 alpha card — done, merged](project_card_t9001_alpha.md) — archived earlier with different words\n"
	if err := os.WriteFile(filepath.Join(dir, fixtureArchive), []byte(diff), 0o644); err != nil {
		t.Fatalf("write differing-text archive: %v", err)
	}
	return dir
}

// fixtureNLinkArchive trims the fixture's archive index to links resolved
// links — the under-threshold refusal pair of AC-MFB-006.
func fixtureNLinkArchive(t *testing.T, links int) string {
	t.Helper()
	dir := specFixtureCopy(t)
	all := []string{
		"- [t9000 zero card](project_card_t9000_zero.md) — closed earlier in the month",
		"- [t8999 prior card](project_card_t8999_prior.md) — closed earlier in the month",
		"- [t8998 older card](project_card_t8998_older.md) — closed earlier in the month",
	}
	if links < 1 || links > len(all) {
		t.Fatalf("fixtureNLinkArchive: links=%d out of range", links)
	}
	body := "---\nname: a\ndescription: d\ntype: reference\n---\n" + strings.Join(all[:links], "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, fixtureArchive), []byte(body), 0o644); err != nil {
		t.Fatalf("write trimmed archive: %v", err)
	}
	return dir
}

// TestMemoryFold_DryRunWritesNothing is AC-MFB-001: the fold previews by
// default, writes nothing, names its store, and refuses an invalid id
// before the store is read.
func TestMemoryFold_DryRunWritesNothing(t *testing.T) {
	dir := specFixtureCopy(t)
	before := storeHashes(t, dir)

	out := runFoldOK(t, "--card", "t9001", "--dir", dir)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 || lines[0] != "store: "+dir+" (--dir)" {
		t.Fatalf("first stdout line = %q, want \"store: %s (--dir)\"", lines[0], dir)
	}
	if !strings.Contains(out, line9001) {
		t.Errorf("preview does not list the STRONG line it would file:\n%s", out)
	}
	if !strings.Contains(out, fixtureArchive) {
		t.Errorf("preview does not name the archive index it would file into:\n%s", out)
	}
	if !strings.Contains(out, "nothing is written without --yes") {
		t.Errorf("preview does not state that it writes nothing:\n%s", out)
	}

	// --json is one valid JSON object carrying the same plan.
	jout := runFoldOK(t, "--card", "t9001", "--dir", dir, "--json")
	var probe map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(jout)), &probe); err != nil {
		t.Fatalf("--json is not one JSON object: %v\n%s", err, jout)
	}
	plan := decodeFoldPlan(t, jout)
	if plan.Store.Dir != dir || plan.Store.Origin != "--dir" {
		t.Errorf("plan store = %+v, want dir %s origin --dir", plan.Store, dir)
	}
	if plan.Card != "t9001" || plan.Archive != fixtureArchive {
		t.Errorf("plan card=%q archive=%q, want t9001/%s", plan.Card, plan.Archive, fixtureArchive)
	}
	if len(plan.Removed) != 1 || plan.Removed[0] != line9001 {
		t.Errorf("plan removed = %q, want exactly [%q]", plan.Removed, line9001)
	}
	if len(plan.Appended) != 1 || plan.Appended[0] != line9001 {
		t.Errorf("plan appended = %q, want exactly [%q]", plan.Appended, line9001)
	}

	// Bare digits are accepted as t<digits>.
	p2 := decodeFoldPlan(t, runFoldOK(t, "--card", "9001", "--dir", dir, "--json"))
	if p2.Card != "t9001" || len(p2.Removed) != 1 || p2.Removed[0] != line9001 {
		t.Errorf("--card 9001 plan = card %q removed %q, want t9001/[%q]", p2.Card, p2.Removed, line9001)
	}

	// Every file of the copy is byte-identical and no temp file remains.
	requireSameStore(t, before, storeHashes(t, dir))
	requireNoTempFiles(t, dir)

	// Invalid ids are refused with a non-zero exit before the store is
	// read: stdout empty.
	for _, bad := range []string{"t9001;x", "t"} {
		r := runFoldRefused(t, "--card", bad, "--dir", dir)
		if r.stdout != "" {
			t.Errorf("--card %q refused but stdout = %q, want empty", bad, r.stdout)
		}
	}
	requireSameStore(t, before, storeHashes(t, dir))
}

// TestMemoryFold_Classification is AC-MFB-002: only STRONG lines are
// planned; AMBIGUOUS and MENTION lines are kept with a reason; a whole-token
// card id never prefix-matches.
func TestMemoryFold_Classification(t *testing.T) {
	dir := specFixtureCopy(t)
	before := storeHashes(t, dir)

	cases := []struct {
		card    string
		removed []string
		kept    map[string]string // line → expected class
		noLine  bool
	}{
		{card: "t9001", removed: []string{line9001}, kept: map[string]string{lineVerify: "MENTION"}},
		{card: "t9002", removed: []string{line9002}},
		{card: "t9006", removed: []string{line9006}},
		{card: "t9004", kept: map[string]string{line9004: "AMBIGUOUS"}},
		{card: "t9005", kept: map[string]string{line9005: "AMBIGUOUS"}},
		{card: "t90", noLine: true},
		{card: "t9999", noLine: true},
	}
	for _, tc := range cases {
		plan := decodeFoldPlan(t, runFoldOK(t, "--card", tc.card, "--dir", dir, "--json"))
		if strings.Join(plan.Removed, "\x00") != strings.Join(tc.removed, "\x00") {
			t.Errorf("card %s: removed = %q, want %q", tc.card, plan.Removed, tc.removed)
		}
		for line, wantClass := range tc.kept {
			gotClass, _ := foldKeptEntry(t, plan, line)
			if gotClass != wantClass {
				t.Errorf("card %s: line %q kept as %q, want %q", tc.card, line, gotClass, wantClass)
			}
		}
		if tc.noLine {
			out := runFoldOK(t, "--card", tc.card, "--dir", dir)
			if !strings.Contains(out, "no line") {
				t.Errorf("card %s: output does not report no line:\n%s", tc.card, out)
			}
		}
	}

	// The AMBIGUOUS reasons name which condition held (acceptance §1:
	// t9004 title only, t9005 target only).
	plan9004 := decodeFoldPlan(t, runFoldOK(t, "--card", "t9004", "--dir", dir, "--json"))
	if _, reason := foldKeptEntry(t, plan9004, line9004); !strings.Contains(reason, "title") {
		t.Errorf("t9004 reason %q does not name the title condition", reason)
	}
	plan9005 := decodeFoldPlan(t, runFoldOK(t, "--card", "t9005", "--dir", dir, "--json"))
	if _, reason := foldKeptEntry(t, plan9005, line9005); !strings.Contains(reason, "target") {
		t.Errorf("t9005 reason %q does not name the target condition", reason)
	}

	requireSameStore(t, before, storeHashes(t, dir))

	// A generated variant whose link targets name two different card ids is
	// AMBIGUOUS (targets name more than one card id).
	twoIDs := "- [t9100 cards](project_card_t9100_alpha.md) and [wrap](project_card_t9101_note.md) — two card ids on one line"
	files := minimalFiles()
	files["project_card_t9100_alpha.md"] = "---\n---\n"
	files["project_card_t9101_note.md"] = "---\n---\n"
	dir2 := seedFoldStore(t, minimalMemory(twoIDs), files)
	plan2 := decodeFoldPlan(t, runFoldOK(t, "--card", "t9100", "--dir", dir2, "--json"))
	if len(plan2.Removed) != 0 {
		t.Errorf("two-card-id line was planned for removal: %q", plan2.Removed)
	}
	if class, reason := foldKeptEntry(t, plan2, twoIDs); class != "AMBIGUOUS" || !strings.Contains(reason, "more than one card id") {
		t.Errorf("two-card-id line = %s/%q, want AMBIGUOUS naming the conflict", class, reason)
	}
}

// TestMemoryFold_ReachabilityPreserved is AC-MFB-003: every applied fold
// preserves reachability, equals the plan exactly, and never creates a new
// doctor orphan — the primary criterion, judged against FIXED expected
// lists, never by size.
func TestMemoryFold_ReachabilityPreserved(t *testing.T) {
	cards := []struct {
		card     string
		removed  []string
		appended []string
	}{
		{card: "t9001", removed: []string{line9001}, appended: []string{line9001}},
		{card: "t9002", removed: []string{line9002}, appended: []string{line9002}},
		{card: "t9006", removed: []string{line9006}, appended: []string{line9006}},
	}
	for _, c := range cards {
		dir := specFixtureCopy(t)
		before := foldSnapshot(t, dir)
		baseOrphans := orphanPaths(t, dir)
		runFoldOK(t, "--card", c.card, "--yes", "--dir", dir)
		after := foldSnapshot(t, dir)
		assertExactFold(t, before, after, taxonomy.FoldPlan{Removed: c.removed, Appended: c.appended})
		requireOrphansUnchanged(t, dir, baseOrphans)
	}

	// A generated variant whose STRONG line is a grouped line carrying two
	// link targets keeps both targets.
	grouped := "- [t9100 cards](project_card_t9100_alpha.md) and [t9100 wrap](project_card_t9100_note.md) — grouped STRONG line"
	files := minimalFiles()
	files["project_card_archive_2026_10.md"] = minimalArchive()
	files["project_card_t9100_alpha.md"] = "---\n---\n"
	files["project_card_t9100_note.md"] = "---\n---\n"
	dirG := seedFoldStore(t, minimalMemory(grouped), files)
	beforeG := foldSnapshot(t, dirG)
	runFoldOK(t, "--card", "t9100", "--yes", "--dir", dirG)
	afterG := foldSnapshot(t, dirG)
	assertExactFold(t, beforeG, afterG, taxonomy.FoldPlan{Removed: []string{grouped}, Appended: []string{grouped}})
	archG := string(afterG[fixtureArchive])
	for _, target := range []string{"project_card_t9100_alpha.md", "project_card_t9100_note.md"} {
		if !strings.Contains(archG, "("+target+")") {
			t.Errorf("grouped variant: target %s did not reach the archive", target)
		}
	}

	// The three cards folded sequentially on ONE copy: after each fold the
	// cumulative plan's invariants hold, and the doctor's orphan set never
	// grows.
	dir := specFixtureCopy(t)
	before := foldSnapshot(t, dir)
	baseOrphans := orphanPaths(t, dir)
	plan := taxonomy.FoldPlan{}
	for _, c := range cards {
		runFoldOK(t, "--card", c.card, "--yes", "--dir", dir)
		plan.Removed = append(plan.Removed, c.removed...)
		plan.Appended = append(plan.Appended, c.appended...)
		after := foldSnapshot(t, dir)
		if v := taxonomy.CheckFoldInvariants(before, after, plan); len(v) != 0 {
			t.Fatalf("sequential fold of %s violates invariants %v", c.card, v)
		}
	}
	requireOrphansUnchanged(t, dir, baseOrphans)
}

// TestMemoryFold_VerbatimFiling is AC-MFB-004: the archived line is verbatim
// (byte for byte, trailing text included), lines follow the file's existing
// line ending, a missing final terminator is supplied, and in the D39
// multi-line variant the archive receives the STRONG lines in their
// MEMORY.md order with the non-STRONG line between them left in place.
func TestMemoryFold_VerbatimFiling(t *testing.T) {
	// The fixture: the archived line is byte-equal to the removed one, at
	// the end of the archive index.
	dir := specFixtureCopy(t)
	before := foldSnapshot(t, dir)
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	after := foldSnapshot(t, dir)
	assertExactFold(t, before, after, taxonomy.FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	if !strings.Contains(string(after[fixtureArchive]), "— merged; detail lives in the topic file") {
		t.Errorf("the t9001 line's trailing text did not travel with it")
	}

	// Windows line endings: the appended line follows the archive's
	// existing \r\n ending.
	dirC := seedFoldStore(t,
		"# Memory Index\r\n\r\n- [Archive](project_card_archive_2026_10.md) — closed cards\r\n"+line9001+"\r\n",
		map[string]string{
			"project_card_archive_2026_10.md": crlfArchive(),
			"feedback_a.md":                   "---\n---\n",
			"feedback_b.md":                   "---\n---\n",
			"feedback_c.md":                   "---\n---\n",
			"project_card_t9001_alpha.md":     "---\n---\n",
		})
	beforeC := foldSnapshot(t, dirC)
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dirC)
	afterC := foldSnapshot(t, dirC)
	assertExactFold(t, beforeC, afterC, taxonomy.FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	if strings.Contains(string(afterC[fixtureArchive]), line9001+"\n") {
		t.Errorf("a \\r\\n archive received a bare-\\n line ending")
	}

	// An archive index lacking a final line terminator gets one before the
	// appended lines (§6).
	dirN := seedFoldStore(t, minimalMemory(line9001), map[string]string{
		"project_card_archive_2026_10.md": minimalArchiveNoFinalNL(),
		"feedback_a.md":                   "---\n---\n",
		"feedback_b.md":                   "---\n---\n",
		"feedback_c.md":                   "---\n---\n",
		"project_card_t9001_alpha.md":     "---\n---\n",
	})
	beforeN := foldSnapshot(t, dirN)
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dirN)
	afterN := foldSnapshot(t, dirN)
	assertExactFold(t, beforeN, afterN, taxonomy.FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})

	// The D39 multi-line variant: two STRONG lines with one non-STRONG line
	// between them. The archive receives the two lines in their MEMORY.md
	// order; the line between stays in MEMORY.md in its original place.
	dirM, second, interim := fixtureMultiLineVariant(t)
	beforeM := foldSnapshot(t, dirM)
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dirM)
	afterM := foldSnapshot(t, dirM)
	planM := taxonomy.FoldPlan{Removed: []string{line9001, second}, Appended: []string{line9001, second}}
	assertExactFold(t, beforeM, afterM, planM)
	wantArch := string(beforeM[fixtureArchive]) + line9001 + "\n" + second + "\n"
	if got := string(afterM[fixtureArchive]); got != wantArch {
		t.Fatalf("(D39) archive order is not the MEMORY.md order\nwant %q\ngot  %q", wantArch, got)
	}
	if got := strings.Count(string(afterM["MEMORY.md"]), interim); got != 1 {
		t.Errorf("(D39) the non-STRONG line between the two appears %d times in MEMORY.md, want exactly 1", got)
	}
	if strings.Contains(string(afterM["MEMORY.md"]), second) {
		t.Errorf("(D39) the second STRONG line stayed in MEMORY.md")
	}

	// The temp-file-then-rename replacement preserves the original file's
	// permission bits: a 0600 index is not widened to 0644.
	dirP := specFixtureCopy(t)
	for _, name := range []string{"MEMORY.md", fixtureArchive} {
		if err := os.Chmod(filepath.Join(dirP, name), 0o600); err != nil {
			t.Fatalf("chmod %s: %v", name, err)
		}
	}
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dirP)
	for _, name := range []string{"MEMORY.md", fixtureArchive} {
		info, err := os.Stat(filepath.Join(dirP, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("%s permission = %o after the fold, want the original 0600", name, got)
		}
	}
}

// TestMemoryFold_ArchiveSelection is AC-MFB-004's selection half and the
// month rollover: among the archive-pattern files MEMORY.md links and that
// exist, the fold files into the greatest name; a matching file MEMORY.md
// does not link is never the destination and is listed as information.
func TestMemoryFold_ArchiveSelection(t *testing.T) {
	// Two linked archives: the greater name receives the lines.
	dir := fixtureTwoArchives(t)
	before := foldSnapshot(t, dir)
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	after := foldSnapshot(t, dir)
	want10 := string(before[fixtureArchive]) + line9001 + "\n"
	if got := string(after[fixtureArchive]); got != want10 {
		t.Fatalf("the greater archive name did not receive the line\nwant %q\ngot  %q", want10, got)
	}
	if string(after[fixtureOlderIndex]) != string(before[fixtureOlderIndex]) {
		t.Errorf("the older archive %s changed", fixtureOlderIndex)
	}
	plan := decodeFoldPlan(t, runFoldOK(t, "--card", "t9002", "--dir", dir, "--json"))
	if plan.Archive != fixtureArchive {
		t.Errorf("plan archive = %q, want the greater linked name %s", plan.Archive, fixtureArchive)
	}
	if len(plan.Unlinked) != 0 {
		t.Errorf("unlinked list = %v, want empty when every archive is linked", plan.Unlinked)
	}

	// Month rollover: the older archive is the linked one and receives the
	// fold; the newer is present but unlinked and is listed as skipped
	// information.
	rollover := fixtureRolloverVariant(t)
	beforeR := foldSnapshot(t, rollover)
	out := runFoldOK(t, "--card", "t9001", "--yes", "--dir", rollover)
	afterR := foldSnapshot(t, rollover)
	want09 := string(beforeR[fixtureOlderIndex]) + line9001 + "\n"
	if got := string(afterR[fixtureOlderIndex]); got != want09 {
		t.Fatalf("rollover fold did not file into the linked %s", fixtureOlderIndex)
	}
	if string(afterR[fixtureArchive]) != string(beforeR[fixtureArchive]) {
		t.Errorf("the unlinked %s changed", fixtureArchive)
	}
	if !strings.Contains(out, fixtureArchive) || !strings.Contains(out, "unlinked") {
		t.Errorf("rollover run does not list the unlinked archive %s as information:\n%s", fixtureArchive, out)
	}
	planR := decodeFoldPlan(t, runFoldOK(t, "--card", "t9002", "--dir", rollover, "--json"))
	if planR.Archive != fixtureOlderIndex {
		t.Errorf("rollover plan archive = %q, want the linked %s", planR.Archive, fixtureOlderIndex)
	}
	if len(planR.Unlinked) != 1 || planR.Unlinked[0] != fixtureArchive {
		t.Errorf("rollover unlinked list = %v, want [%s]", planR.Unlinked, fixtureArchive)
	}
}

// TestMemoryFold_Idempotent is AC-MFB-005: the fold is idempotent, recovers
// from a partial apply, and keeps a differing line with its reason.
func TestMemoryFold_Idempotent(t *testing.T) {
	t.Cleanup(func() { memoryFoldSeam = foldTestSeam{} })

	// The second run exits 0, reports nothing remains, and changes nothing.
	dir := specFixtureCopy(t)
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	afterFirst := storeHashes(t, dir)
	out := runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	if !strings.Contains(out, "no line") || !strings.Contains(out, "nothing remains") {
		t.Errorf("the second run does not report that nothing remains:\n%s", out)
	}
	requireSameStore(t, afterFirst, storeHashes(t, dir))
	requireNoTempFiles(t, dir)
	if n := countArchiveLines(t, dir, line9001); n != 1 {
		t.Errorf("the archive carries %d lines with the t9001 target set, want exactly 1", n)
	}

	// Recovery from a partial apply: the seam stops the apply after the
	// archive append and before the MEMORY.md rewrite; the line is then in
	// both files, and the re-run completes without a second append.
	dir2 := specFixtureCopy(t)
	before2 := foldSnapshot(t, dir2)
	memoryFoldSeam = foldTestSeam{failAt: "rewrite"}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir2)
	memoryFoldSeam = foldTestSeam{}
	if !strings.Contains(foldRead(t, dir2, "MEMORY.md"), line9001) {
		t.Errorf("MEMORY.md lost the line despite the interrupted apply")
	}
	if !strings.Contains(foldRead(t, dir2, fixtureArchive), line9001) {
		t.Errorf("the archive did not receive the line before the interruption")
	}
	foldSetsHold(t, before2, foldSnapshot(t, dir2))
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir2)
	done := foldSnapshot(t, dir2)
	assertExactFold(t, before2, done, taxonomy.FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	if n := countArchiveLines(t, dir2, line9001); n != 1 {
		t.Errorf("after the recovery re-run the archive carries %d lines with the t9001 target set, want exactly 1", n)
	}
	requireNoTempFiles(t, dir2)

	// A differing line with the same link-target set is kept in MEMORY.md,
	// reported with its reason, and the archive stays byte-identical.
	dir3 := fixtureDifferingText(t)
	before3 := storeHashes(t, dir3)
	out3 := runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir3)
	if !strings.Contains(out3, "archive carries a different line for the same targets") {
		t.Errorf("the differing-text line is not reported with its reason:\n%s", out3)
	}
	if !strings.Contains(foldRead(t, dir3, "MEMORY.md"), line9001) {
		t.Errorf("the differing-text line did not stay in MEMORY.md")
	}
	requireSameStore(t, before3, storeHashes(t, dir3))

	// A retry whose before-snapshot still sees the archive line, while a
	// concurrent author strips that line from the archive before the apply,
	// must refuse: the archive's copy is verified in THIS run regardless of
	// whether this attempt appends, and the last surviving copy is never
	// deleted from MEMORY.md.
	dir4 := specFixtureCopy(t)
	memoryFoldSeam = foldTestSeam{failAt: "rewrite"}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir4) // partial apply: the line is in both files
	memoryFoldSeam = foldTestSeam{mutateDisk: func(storeDir string) {
		before := foldRead(t, storeDir, fixtureArchive)
		stripped := strings.Replace(before, line9001+"\n", "", 1)
		if stripped == before {
			panic("the archive line to strip was not found")
		}
		if err := os.WriteFile(filepath.Join(storeDir, fixtureArchive), []byte(stripped), 0o644); err != nil {
			panic(err)
		}
	}}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir4)
	memoryFoldSeam = foldTestSeam{}
	if !strings.Contains(foldRead(t, dir4, "MEMORY.md"), line9001) {
		t.Errorf("the retried fold deleted the last surviving copy of the line from MEMORY.md")
	}
	requireNoTempFiles(t, dir4)

	// The gate overlay: a retry whose before-snapshot sees a qualified
	// archive, while a concurrent author strips the archive's OTHER links
	// before the apply. The moved line survives on disk, but the archive
	// loses index qualification — any archive drift must abort the
	// original-line deletion, or the card becomes unreachable.
	dir5 := specFixtureCopy(t)
	memoryFoldSeam = foldTestSeam{failAt: "rewrite"}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir5) // the line is now in both files
	memoryFoldSeam = foldTestSeam{mutateDisk: func(storeDir string) {
		archiveBefore := foldRead(t, storeDir, fixtureArchive)
		crippled := strings.ReplaceAll(archiveBefore,
			"- [t8999 prior card](project_card_t8999_prior.md) — closed earlier in the month\n", "")
		crippled = strings.ReplaceAll(crippled,
			"- [t8998 older card](project_card_t8998_older.md) — closed earlier in the month\n", "")
		if crippled == archiveBefore {
			panic("the archive links to strip were not found")
		}
		if err := os.WriteFile(filepath.Join(storeDir, fixtureArchive), []byte(crippled), 0o644); err != nil {
			panic(err)
		}
	}}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir5)
	memoryFoldSeam = foldTestSeam{}
	if !strings.Contains(foldRead(t, dir5, "MEMORY.md"), line9001) {
		t.Errorf("the retry deleted the moved line although the archive had drifted since the plan")
	}
	if !foldSnapshot(t, dir5).ReachableSet()["project_card_t9001_alpha.md"] {
		t.Errorf("the card's topic file became unreachable across the retried fold")
	}
	requireNoTempFiles(t, dir5)

	// (iii-b, gate overlay) the content re-check is the LAST step before
	// each rename: a concurrent update that lands after the temp file is
	// prepared must abort the rename, and the concurrent update must
	// survive — the rename may never overwrite it.
	dir6 := specFixtureCopy(t)
	memoryFoldSeam = foldTestSeam{mutateDuringWrite: func(storeDir, name string) {
		if name != "MEMORY.md" {
			return
		}
		data, err := os.ReadFile(filepath.Join(storeDir, name))
		if err != nil {
			panic(err)
		}
		concurrent := append([]byte(nil), data...)
		concurrent = append(concurrent, "- [concurrent writer](feedback_alpha.md) — landed mid-apply\n"...)
		if err := os.WriteFile(filepath.Join(storeDir, name), concurrent, 0o600); err != nil {
			panic(err)
		}
	}}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir6)
	memoryFoldSeam = foldTestSeam{}
	after6 := foldRead(t, dir6, "MEMORY.md")
	if !strings.Contains(after6, "- [concurrent writer](feedback_alpha.md) — landed mid-apply") {
		t.Errorf("(iii-b) the concurrent update did not survive the fold's rename")
	}
	if !strings.Contains(after6, line9001) {
		t.Errorf("(iii-b) the fold removed the planned line despite aborting")
	}
	requireNoTempFiles(t, dir6)
}

// countArchiveLines counts the archive lines carrying the same link-target
// set as line (the identity REQ-MFB-005 dedupes by).
func countArchiveLines(t *testing.T, dir, line string) int {
	t.Helper()
	key := taxonomy.LineTargetSetKey(line)
	n := 0
	for _, seg := range strings.Split(foldRead(t, dir, fixtureArchive), "\n") {
		if taxonomy.LineTargetSetKey(seg) == key {
			n++
		}
	}
	return n
}

// TestMemoryFold_EdgeInputs is AC-MFB-006: no line, only-AMBIGUOUS, an
// unlinked archive, no archive, an archive under the threshold, an archive
// exactly at it, invalid ids, and a symbolic-link MEMORY.md.
func TestMemoryFold_EdgeInputs(t *testing.T) {
	// No line, and only AMBIGUOUS lines: exit 0, nothing written.
	for _, card := range []string{"t9999", "t9004"} {
		dir := specFixtureCopy(t)
		before := storeHashes(t, dir)
		out := runFoldOK(t, "--card", card, "--yes", "--dir", dir)
		requireSameStore(t, before, storeHashes(t, dir))
		if !strings.Contains(out, "no line") {
			t.Errorf("card %s: output does not report no line:\n%s", card, out)
		}
	}

	// Present but unlinked archive: refuse, name the file to link, write
	// nothing.
	unlinked := specFixtureCopy(t)
	mem := foldRead(t, unlinked, "MEMORY.md")
	edited := strings.Replace(mem, lineArchive10+"\n", "", 1)
	if edited == mem {
		t.Fatal("fixture edit failed: the archive link was not found")
	}
	if err := os.WriteFile(filepath.Join(unlinked, "MEMORY.md"), []byte(edited), 0o644); err != nil {
		t.Fatalf("write MEMORY.md: %v", err)
	}
	beforeU := storeHashes(t, unlinked)
	rU := runFoldRefused(t, "--card", "t9001", "--yes", "--dir", unlinked)
	requireSameStore(t, beforeU, storeHashes(t, unlinked))
	if msg := rU.stderr + rU.stdout; !strings.Contains(msg, fixtureArchive) || !strings.Contains(msg, "link") {
		t.Errorf("the unlinked-archive refusal does not name %s as a file to link: %s", fixtureArchive, rU.stderr)
	}

	// No archive-pattern file at all: refuse, name the file to create.
	noArch := specFixtureCopy(t)
	if err := os.Remove(filepath.Join(noArch, fixtureArchive)); err != nil {
		t.Fatalf("remove archive: %v", err)
	}
	mem = foldRead(t, noArch, "MEMORY.md")
	edited = strings.Replace(mem, lineArchive10+"\n", "", 1)
	if err := os.WriteFile(filepath.Join(noArch, "MEMORY.md"), []byte(edited), 0o644); err != nil {
		t.Fatalf("write MEMORY.md: %v", err)
	}
	beforeN := storeHashes(t, noArch)
	rN := runFoldRefused(t, "--card", "t9001", "--yes", "--dir", noArch)
	requireSameStore(t, beforeN, storeHashes(t, noArch))
	if msg := rN.stderr + rN.stdout; !strings.Contains(msg, "project_card_archive_<YYYY>_<MM>.md") || !strings.Contains(msg, "create") {
		t.Errorf("the no-archive refusal does not name the file to create: %s", rN.stderr)
	}

	// One resolved link: the fold would leave the archive at two, under the
	// threshold of three — refuse, naming the archive, the count and the
	// threshold (read through the accessor, never a literal).
	oneLink := fixtureNLinkArchive(t, 1)
	beforeO := storeHashes(t, oneLink)
	rO := runFoldRefused(t, "--card", "t9001", "--yes", "--dir", oneLink)
	requireSameStore(t, beforeO, storeHashes(t, oneLink))
	th := taxonomy.SecondaryIndexLinkThreshold()
	msg := rO.stderr + rO.stdout
	if !strings.Contains(msg, fixtureArchive) ||
		!strings.Contains(msg, strconv.Itoa(th-1)) ||
		!strings.Contains(msg, strconv.Itoa(th)) {
		t.Errorf("the under-threshold refusal does not name the archive, the count %d and the threshold %d: %s", th-1, th, rO.stderr)
	}

	// Two resolved links: the fold leaves exactly the threshold — it folds,
	// and the doctor reports no new orphan.
	twoLink := fixtureNLinkArchive(t, 2)
	beforeT := foldSnapshot(t, twoLink)
	baseOrphans := orphanPaths(t, twoLink)
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", twoLink)
	assertExactFold(t, beforeT, foldSnapshot(t, twoLink),
		taxonomy.FoldPlan{Removed: []string{line9001}, Appended: []string{line9001}})
	// The fold brings the archive to exactly the threshold: it becomes an
	// index member, so its previously orphaned targets become reachable —
	// the AC requires no NEW orphan, not an unchanged set.
	requireNoNewOrphans(t, twoLink, baseOrphans)

	// Invalid ids exit non-zero before any read.
	dir := specFixtureCopy(t)
	beforeI := storeHashes(t, dir)
	for _, bad := range []string{"nope", "t9001 ", "T9001", "t12a"} {
		r := runFoldRefused(t, "--card", bad, "--dir", dir)
		if r.stdout != "" {
			t.Errorf("--card %q refused but stdout = %q, want empty", bad, r.stdout)
		}
	}
	requireSameStore(t, beforeI, storeHashes(t, dir))

	// A MEMORY.md that is a symbolic link is refused, nothing written.
	link := seedFoldStore(t, minimalMemory(line9001), minimalFiles())
	if err := os.Remove(filepath.Join(link, "MEMORY.md")); err != nil {
		t.Fatalf("remove MEMORY.md: %v", err)
	}
	if err := os.Symlink(filepath.Join(link, "feedback_a.md"), filepath.Join(link, "MEMORY.md")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	beforeL := storeHashes(t, link)
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", link)
	requireSameStore(t, beforeL, storeHashes(t, link))
}

// TestMemoryFold_ApplyOrderAndAbort is AC-MFB-007: apply order (append,
// verify, rewrite), abort-on-change, the pre-write exactness check, and no
// temporary file left behind on any path.
func TestMemoryFold_ApplyOrderAndAbort(t *testing.T) {
	t.Cleanup(func() { memoryFoldSeam = foldTestSeam{} })

	// (i) a failure at the archive append leaves the store byte-identical.
	dir := specFixtureCopy(t)
	before := storeHashes(t, dir)
	memoryFoldSeam = foldTestSeam{failAt: "append"}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir)
	memoryFoldSeam = foldTestSeam{}
	requireSameStore(t, before, storeHashes(t, dir))
	requireNoTempFiles(t, dir)

	// (ii) a failure at the MEMORY.md rewrite leaves the line in both
	// files, invariants (a) and (b) holding.
	dir2 := specFixtureCopy(t)
	before2 := foldSnapshot(t, dir2)
	memoryFoldSeam = foldTestSeam{failAt: "rewrite"}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir2)
	memoryFoldSeam = foldTestSeam{}
	interrupted := foldSnapshot(t, dir2)
	foldSetsHold(t, before2, interrupted)
	if !strings.Contains(string(interrupted["MEMORY.md"]), line9001) ||
		!strings.Contains(string(interrupted[fixtureArchive]), line9001) {
		t.Errorf("after the interrupted rewrite the line is not present in both files")
	}
	requireNoTempFiles(t, dir2)

	// (iii) MEMORY.md mutated between plan and apply: abort with a non-zero
	// exit, no archive line, MEMORY.md stays the mutated version.
	dir3 := specFixtureCopy(t)
	mem3 := foldRead(t, dir3, "MEMORY.md")
	mutated := strings.Replace(mem3, line9001+"\n",
		line9001+"\n- [t9004 delta note](feedback_delta_note.md) — title leads with a card id but the target names no card\n", 1)
	if mutated == mem3 {
		t.Fatal("fixture edit failed: the STRONG line was not found")
	}
	archBefore3 := foldRead(t, dir3, fixtureArchive)
	memoryFoldSeam = foldTestSeam{mutateDisk: func(storeDir string) {
		if err := os.WriteFile(filepath.Join(storeDir, "MEMORY.md"), []byte(mutated), 0o644); err != nil {
			panic(err)
		}
	}}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir3)
	memoryFoldSeam = foldTestSeam{}
	if got := foldRead(t, dir3, "MEMORY.md"); got != mutated {
		t.Errorf("(iii) MEMORY.md does not stay the mutated version")
	}
	if got := foldRead(t, dir3, fixtureArchive); got != archBefore3 {
		t.Errorf("(iii) the archive gained a line despite the concurrent MEMORY.md change")
	}
	requireNoTempFiles(t, dir3)

	// (iv) the in-memory result diverges from the plan: abort BEFORE any
	// write, exit non-zero, the violated invariant named — first a stray
	// byte in a non-index file ((d3)), then one extra line removed from
	// MEMORY.md ((d1)).
	dir4 := specFixtureCopy(t)
	before4 := storeHashes(t, dir4)
	memoryFoldSeam = foldTestSeam{mutateResult: func(after taxonomy.StoreSnapshot) taxonomy.StoreSnapshot {
		after["feedback_alpha.md"] = append(after["feedback_alpha.md"], 'x')
		return after
	}}
	r4 := runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir4)
	memoryFoldSeam = foldTestSeam{}
	if !strings.Contains(r4.stderr, "d3") {
		t.Errorf("(iv) abort does not name invariant (d3): %s", r4.stderr)
	}
	requireSameStore(t, before4, storeHashes(t, dir4))
	requireNoTempFiles(t, dir4)

	dir5 := specFixtureCopy(t)
	before5 := storeHashes(t, dir5)
	memoryFoldSeam = foldTestSeam{mutateResult: func(after taxonomy.StoreSnapshot) taxonomy.StoreSnapshot {
		after["MEMORY.md"] = []byte(strings.Replace(string(after["MEMORY.md"]), lineBlock+"\n", "", 1))
		return after
	}}
	r5 := runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir5)
	memoryFoldSeam = foldTestSeam{}
	if !strings.Contains(r5.stderr, "d1") {
		t.Errorf("(iv) abort does not name invariant (d1): %s", r5.stderr)
	}
	requireSameStore(t, before5, storeHashes(t, dir5))
	requireNoTempFiles(t, dir5)
}

// TestMemoryFold_ArchiveEntryLinkKept is the card-review round-1 P2 (card
// t1502): a STRONG line carrying BOTH the card link and the archive link —
// when it is MEMORY.md's only archive link, removing the line unlinks the
// fold's own destination: after the first apply ArchiveIndexName() returns
// empty and every subsequent fold refuses. The fold must KEEP the line with
// that reason instead, so the archive stays linked and later folds keep
// working.
func TestMemoryFold_ArchiveEntryLinkKept(t *testing.T) {
	memory := "# Memory Index\n\n" +
		"- [t9001 alpha card — done, merged](project_card_t9001_alpha.md) — files into the [archive](project_card_archive_2026_10.md)\n"
	files := minimalFiles()
	// The archive index must exist for the selection (§1.5: linked AND
	// present); minimalFiles carries the three threshold links and the
	// card's topic file already.
	files["project_card_archive_2026_10.md"] = minimalArchive()
	dir := seedFoldStore(t, memory, files)

	// First fold: the STRONG line is kept — removing it would unlink the
	// archive index.
	out := runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	if !strings.Contains(out, "kept 1 line") {
		t.Errorf("fold output does not report the kept line: %q", out)
	}
	if !strings.Contains(out, "unlink the archive index") {
		t.Errorf("kept reason does not name the archive-unlink hazard: %q", out)
	}
	if got := foldRead(t, dir, "MEMORY.md"); !strings.Contains(got, "t9001 alpha card") {
		t.Errorf("the entry-link line was removed from MEMORY.md — the archive lost its link")
	}

	// A subsequent fold must still resolve the archive (the reviewer's
	// refusal repro) and keep reporting the same kept line.
	out2 := runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	if strings.Contains(out2, "no archive index") {
		t.Errorf("the follow-up fold refused although the archive entry link was kept: %q", out2)
	}
	requireNoTempFiles(t, dir)
}

// TestMemoryFold_MoveBeforeDeletionKeepsLine is the card-review round-1 P2
// (card t1502): a referenced topic file MOVED before the pre-deletion check
// — the two indexes' bytes still match, but the archive's effective
// resolved-link count dropped and the moved line's target no longer
// resolves. The fold must re-run qualification and reachability against the
// CURRENT store right before the deletion and keep the original line on
// failure.
func TestMemoryFold_MoveBeforeDeletionKeepsLine(t *testing.T) {
	dir := specFixtureCopy(t)
	memoryFoldSeam = foldTestSeam{mutateDisk: func(storeDir string) {
		// A concurrent mover renames the card's topic file between the
		// plan and the apply; the indexes' bytes are untouched.
		if err := os.Rename(
			filepath.Join(storeDir, "project_card_t9001_alpha.md"),
			filepath.Join(storeDir, "project_card_t9001_alpha_moved.md")); err != nil {
			panic(err)
		}
	}}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir)
	memoryFoldSeam = foldTestSeam{}
	// The original line stays in MEMORY.md — deleting it would strand the
	// moved card (its new name is linked from no index).
	if got := foldRead(t, dir, "MEMORY.md"); !strings.Contains(got, line9001) {
		t.Errorf("the fold removed the planned line although the line's target had moved away")
	}
	if _, err := os.Stat(filepath.Join(dir, "project_card_t9001_alpha_moved.md")); err != nil {
		t.Errorf("the concurrent move did not survive: %v", err)
	}
	requireNoTempFiles(t, dir)
}

// TestMemoryFold_NeverExistedTargetRefusedCleanly is the card-review
// round-2 P2 (card t1502): a STRONG line whose target NEVER existed is a
// pre-existing broken link, not a mid-run move — the fold must refuse in
// the PLAN phase, before the archive gains anything, so the store stays
// byte-identical and no retry is blocked by a write the refusal already
// made. The apply-time reachability check alone would append first and
// refuse after, dirtying the store on every attempt.
func TestMemoryFold_NeverExistedTargetRefusedCleanly(t *testing.T) {
	memory := "# Memory Index\n\n" +
		"- [Archive](project_card_archive_2026_10.md) — closed cards\n" +
		"- [t9003 gone card — merged?](project_card_t9003_gone.md) — the topic file never existed\n"
	files := minimalFiles()
	files[fixtureArchive] = minimalArchive()
	dir := seedFoldStore(t, memory, files)

	runFoldRefused(t, "--card", "t9003", "--yes", "--dir", dir)
	// The refusal is clean: the store is byte-identical — no archive line,
	// no temp file, nothing to unwind.
	if got := foldRead(t, dir, fixtureArchive); got != minimalArchive() {
		t.Errorf("the archive was written although the fold refused in the plan phase:\n%q", got)
	}
	if got := foldRead(t, dir, "MEMORY.md"); !strings.Contains(got, "t9003 gone card") {
		t.Errorf("the broken-link line was removed although the fold refused")
	}
	// The retry reads the same clean refusal — nothing is stuck halfway.
	runFoldRefused(t, "--card", "t9003", "--yes", "--dir", dir)
	if got := foldRead(t, dir, fixtureArchive); got != minimalArchive() {
		t.Errorf("the retry dirtied the archive:\n%q", got)
	}
	requireNoTempFiles(t, dir)
}

// TestMemoryFold_MoveDuringMemoryPrepKeepsLine is the codex-review round-2
// P2 (card t1502): the effective-state check ran BEFORE the MEMORY.md temp
// file was prepared, but the window up to the LAST rename is still open —
// a mover renaming referenced topic files during the temp preparation
// invalidates the link count the earlier check measured, while the byte
// guards still pass. The pre-rename position re-runs the effective
// verification; on failure the original line stays in MEMORY.md.
func TestMemoryFold_MoveDuringMemoryPrepKeepsLine(t *testing.T) {
	dir := specFixtureCopy(t)
	// Which topic files the fixture archive references: both movers below
	// must rename files the archive (or the moved line) points at.
	memoryFoldSeam = foldTestSeam{mutateDuringWrite: func(storeDir, name string) {
		if name != "MEMORY.md" {
			return
		}
		for _, move := range [][2]string{
			{"project_card_t9001_alpha.md", "project_card_t9001_alpha_moved.md"},
			{"feedback_alpha.md", "feedback_alpha_moved.md"},
		} {
			from := filepath.Join(storeDir, move[0])
			if _, err := os.Stat(from); err == nil {
				if err := os.Rename(from, filepath.Join(storeDir, move[1])); err != nil {
					panic(err)
				}
			}
		}
	}}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir)
	memoryFoldSeam = foldTestSeam{}
	if got := foldRead(t, dir, "MEMORY.md"); !strings.Contains(got, line9001) {
		t.Errorf("the fold removed the planned line although referenced files moved during the MEMORY.md write")
	}
	for _, moved := range []string{"project_card_t9001_alpha_moved.md", "feedback_alpha_moved.md"} {
		if _, err := os.Stat(filepath.Join(dir, moved)); err != nil {
			t.Errorf("the concurrent move of %s did not survive: %v", moved, err)
		}
	}
	requireNoTempFiles(t, dir)
}

// TestReviewArchiveUpdateDuringEffectiveScan is the codex-review round-2 P1
// ordering regression (card t1502): the effective (full-store) check
// re-reads every topic file and takes TIME — an archive update landing
// DURING it must not sail past an already-passed byte comparison. The
// pre-rename order must therefore be effective-FIRST, bytes-LAST: the
// reviewer's FIFO-synchronized reproduction (an archive update timed inside
// the effective scan; the fold succeeded and the update was overwritten)
// is covered twice here — by the recorded check order itself, and by a
// functional rerun whose archive edit inside the effective window must
// abort the rename and survive.
func TestReviewArchiveUpdateDuringEffectiveScan(t *testing.T) {
	// Part 1 — the order probe: within one MEMORY.md write the effective
	// check completes BEFORE the byte comparison.
	dir := specFixtureCopy(t)
	var order []string
	var orderMu sync.Mutex
	memoryFoldSeam = foldTestSeam{
		orderProbe: func(stage string) {
			orderMu.Lock()
			defer orderMu.Unlock()
			order = append(order, stage)
		},
	}
	runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	memoryFoldSeam = foldTestSeam{}
	orderMu.Lock()
	got := append([]string(nil), order...)
	orderMu.Unlock()
	if len(got) < 2 {
		t.Fatalf("no pre-rename probe events recorded: %v", got)
	}
	// Each write's effective check must complete before its byte check; the
	// LAST bytes-done therefore trails the LAST effective-start.
	lastEffective, lastBytes := -1, -1
	for i, stage := range got {
		switch stage {
		case "effective-start":
			lastEffective = i
		case "bytes-done":
			lastBytes = i
		}
	}
	if lastEffective < 0 || lastBytes < 0 || lastEffective > lastBytes {
		t.Errorf("pre-rename check order = %v, want the effective check BEFORE the byte comparison", got)
	}

	// Part 2 — the functional window, injected deterministically: the
	// orderProbe's "effective-start" fires at the top of the pre-rename
	// effective check, so mutating the archive there IS an update landing
	// inside the effective window. The effective check is blind to index
	// bytes, so only the LAST (byte) check can catch this — the fold must
	// abort and the concurrent update must survive.
	dir2 := specFixtureCopy(t)
	archivePath := filepath.Join(dir2, fixtureArchive)
	memoryFoldSeam = foldTestSeam{
		orderProbe: func(stage string) {
			if stage != "effective-start" {
				return
			}
			data, err := os.ReadFile(archivePath)
			if err != nil {
				panic(err)
			}
			concurrent := append([]byte(nil), data...)
			concurrent = append(concurrent, "- [concurrent archive writer](feedback_alpha.md) — landed mid-scan\n"...)
			if err := os.WriteFile(archivePath, concurrent, 0o600); err != nil {
				panic(err)
			}
		},
	}
	runFoldRefused(t, "--card", "t9001", "--yes", "--dir", dir2)
	memoryFoldSeam = foldTestSeam{}
	// The concurrent update survives — the rename never overwrote it.
	if got := foldRead(t, dir2, fixtureArchive); !strings.Contains(got, "landed mid-scan") {
		t.Errorf("the concurrent archive update did not survive the fold's rename")
	}
	if _, statErr := os.Stat(archivePath); statErr != nil {
		t.Errorf("the archive index vanished: %v", statErr)
	}
	requireNoTempFiles(t, dir2)
}

// TestMemoryFold_DuplicateStrongLineFiledOnce is the codex-review round-2
// P2 (card t1502): the same STRONG line appearing TWICE in MEMORY.md with
// neither copy in the archive — the dedupe set must include the lines THIS
// fold has already planned to append, so the second copy is removed
// without a second append (the archive carries exactly ONE copy; a
// different wording on the same targets is still kept with its reason).
func TestMemoryFold_DuplicateStrongLineFiledOnce(t *testing.T) {
	dir := specFixtureCopy(t)
	// Duplicate the t9001 STRONG line: a second identical copy goes to the
	// end of MEMORY.md, and a same-targets different-wording copy between
	// the archive link and it exercises the kept arm.
	extra := line9001 + "\n" +
		"- [t9001 alpha card — done, merged](project_card_t9001_alpha.md) — different wording, same targets\n"
	memPath := filepath.Join(dir, "MEMORY.md")
	data, err := os.ReadFile(memPath)
	if err != nil {
		t.Fatalf("read MEMORY.md: %v", err)
	}
	if err := os.WriteFile(memPath, append(data, []byte(extra)...), 0o600); err != nil {
		t.Fatalf("duplicate the strong line: %v", err)
	}

	out := runFoldOK(t, "--card", "t9001", "--yes", "--dir", dir)
	if n := countArchiveLines(t, dir, line9001); n != 1 {
		t.Errorf("the archive carries %d copies of the identical line, want exactly 1\noutput: %s", n, out)
	}
	if got := foldRead(t, dir, "MEMORY.md"); strings.Contains(got, line9001) {
		t.Errorf("the identical line stayed in MEMORY.md — both copies should have been folded")
	}
	// The different-wording copy is kept in MEMORY.md with its reason.
	if got := foldRead(t, dir, "MEMORY.md"); !strings.Contains(got, "different wording, same targets") {
		t.Errorf("the different-wording copy was folded although its wording differs")
	}
	requireNoTempFiles(t, dir)
}
