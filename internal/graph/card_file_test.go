package graph

// card_file_test.go — SPEC-TODO-CARD-ISSUANCE-001 M3 (REQ-TCI-016): the five
// acceptance tests of the card→file edge layer. Every fixture is a real git
// repo under t.TempDir(); attribution goes through the engine's single
// attribution point (factory.AttributeSubject) — the layer never grows a
// second matcher (MU-87), so the tests pin the SHIPPED wiring, not a stub.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// cardFileFixture creates a repo whose history carries:
//   - `merge(t5): …` — a landing merge on the first-parent path;
//   - `merge(t7): …` — a landing merge reachable ONLY through the absorb
//     merge's second parent (the first-parent-only walk misses it — MU-86);
//   - `Merge branch 'wt-x' into feature-other (card t8)` — an absorb-direction
//     merge the non-attribution rule mechanically refuses (MU-87's shape the
//     engine's own rule catches).
func cardFileFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitFix(t, root, "init", "-q", "-b", "main")
	gitFix(t, root, "config", "user.email", "fixture@example.com")
	gitFix(t, root, "config", "user.name", "Fixture")
	// Cleanup guarantee (AGENTS.md §4), same discipline as newCheckFixture:
	// refuse the detached `git maintenance` writer outright.
	gitFix(t, root, "config", "gc.auto", "0")
	gitFix(t, root, "config", "gc.autoDetach", "false")

	commit := func(msg string, files map[string]string) {
		t.Helper()
		for name, body := range files {
			path := filepath.Join(root, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			gitFix(t, root, "add", name)
		}
		gitFix(t, root, "commit", "-q", "-m", msg)
	}

	commit("root", map[string]string{"f.txt": "0\n"})
	gitFix(t, root, "checkout", "-q", "-b", "t5br")
	commit("t5 work", map[string]string{"t5.txt": "t5\n"})
	gitFix(t, root, "checkout", "-q", "main")
	gitFix(t, root, "merge", "--no-ff", "-q", "-m", "merge(t5): WT-t5 into develop - real landing", "t5br")
	gitFix(t, root, "checkout", "-q", "-b", "wt-x")
	commit("side work", map[string]string{"s1.txt": "s1\n"})
	gitFix(t, root, "checkout", "-q", "-b", "wt2")
	commit("side2 work", map[string]string{"s2.txt": "s2\n"})
	gitFix(t, root, "checkout", "-q", "wt-x")
	// The t7 landing merge lives on the card branch; main never walks it
	// directly — only the absorb merge's second parent reaches it.
	gitFix(t, root, "merge", "--no-ff", "-q", "-m", "merge(t7): WT-x into develop - real landing", "wt2")
	gitFix(t, root, "checkout", "-q", "main")
	commit("main work", map[string]string{"f.txt": "1\n"})
	gitFix(t, root, "merge", "--no-ff", "-q", "-m", "Merge branch 'wt-x' into feature-other (card t8)", "wt-x")
	return root
}

func cardFileRun(t *testing.T, root string) []CardFileEdge {
	t.Helper()
	edges, err := CardFileEdges(root, cardFileLandedBranch(root), factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardFileEdges: %v", err)
	}
	return edges
}

// TestGraphCardFileEdgesDeterministic — AC-TCI-016 (b): two runs over the
// same tree and reachable history return byte-identical output.
func TestGraphCardFileEdgesDeterministic(t *testing.T) {
	root := cardFileFixture(t)
	first, err := json.Marshal(cardFileRun(t, root))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(cardFileRun(t, root))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("two runs disagree:\nfirst:  %s\nsecond: %s", first, second)
	}
	branch := cardFileLandedBranch(root)
	fp1, err := CardMergeFingerprint(root, branch, factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardMergeFingerprint: %v", err)
	}
	fp2, err := CardMergeFingerprint(root, branch, factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardMergeFingerprint: %v", err)
	}
	if fp1 != fp2 {
		t.Fatalf("fingerprint not deterministic: %s vs %s", fp1, fp2)
	}
}

// TestGraphCardFileEdgesCarryNoQueueState — AC-TCI-016 (c): a queue file in
// the tree changes no output, and no unlanded card id, expected file, or
// finding appears in the edges.
func TestGraphCardFileEdgesCarryNoQueueState(t *testing.T) {
	root := cardFileFixture(t)
	before := cardFileRun(t, root)

	queueDir := filepath.Join(root, ".moai", "state", "todo")
	if err := os.MkdirAll(queueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A queue-shaped file carrying exactly what the layer must NOT read:
	// an unlanded card, its expected files, recorded findings.
	queue := `{"items":[{"id":"t999","state":"queued","title":"unlanded","issuance":{"files":["s1.txt","secret.txt"]},"spec_id":"SPEC-UNLANDED-001"}],"findings":[{"subject_id":"t999","related_id":"t5","relation":"blocks"}]}`
	if err := os.WriteFile(filepath.Join(queueDir, "backlog.json"), []byte(queue), 0o600); err != nil {
		t.Fatal(err)
	}

	after := cardFileRun(t, root)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("queue presence changed the edges:\nbefore: %+v\nafter:  %+v", before, after)
	}
	rendered, err := json.Marshal(after)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"t999", "secret.txt", "SPEC-UNLANDED-001", "blocks"} {
		if strings.Contains(string(rendered), forbidden) {
			t.Fatalf("edges carry queue state %q: %s", forbidden, rendered)
		}
	}
}

// TestGraphCheckNoticesCardFileSource — AC-TCI-016 (d): after a build stamps
// its fingerprints, a NEW card-attributed merge landing makes `graph check`
// read the edges layer stale, naming the card-merges source.
func TestGraphCheckNoticesCardFileSource(t *testing.T) {
	root := cardFileFixture(t)

	edges, err := Build(root)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var cardFileCount int
	for _, e := range edges {
		if e.Kind == KindCardFile {
			cardFileCount++
		}
	}
	if cardFileCount == 0 {
		t.Fatalf("build produced no card-file edges: %+v", edges)
	}

	edgesPath := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	if err := WriteJSONL(edgesPath, edges); err != nil {
		t.Fatal(err)
	}
	metaPath := filepath.Join(filepath.Dir(edgesPath), MetaFileName)
	if err := WriteEdgesMeta(metaPath, root, SourceFingerprintsForEdges(root), len(edges)); err != nil {
		t.Fatal(err)
	}
	if rep := checkEdges(root); rep.Verdict != VerdictFresh {
		t.Fatalf("edges layer not fresh right after a build: %+v", rep)
	}

	// A new card-attributed merge lands on main.
	gitFix(t, root, "checkout", "-q", "-b", "t9br")
	if err := os.WriteFile(filepath.Join(root, "t9.txt"), []byte("t9\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitFix(t, root, "add", "t9.txt")
	gitFix(t, root, "commit", "-q", "-m", "t9 work")
	gitFix(t, root, "checkout", "-q", "main")
	gitFix(t, root, "merge", "--no-ff", "-q", "-m", "merge(t9): WT-y into develop - real landing", "t9br")

	rep := checkEdges(root)
	if rep.Verdict != VerdictStale {
		t.Fatalf("edges layer missed the new card merge: %+v", rep)
	}
	if !strings.Contains(rep.Reason, srcCardMerges) {
		t.Fatalf("stale reason does not name the card-merges source: %s", rep.Reason)
	}
}

// TestGraphCardFileEdgesSeeAbsorbedMerge — AC-TCI-016 (e): the t7 landing
// merge is reachable only through the absorb merge's second parent, and its
// edges still appear; the absorb merge itself yields none (f).
func TestGraphCardFileEdgesSeeAbsorbedMerge(t *testing.T) {
	root := cardFileFixture(t)
	edges := cardFileRun(t, root)

	if len(edges) != 2 {
		t.Fatalf("want exactly the t5 and t7 edges, got %+v", edges)
	}
	want := []CardFileEdge{
		{Card: "t5", File: "t5.txt", SHA: edges[0].SHA},
		{Card: "t7", File: "s2.txt", SHA: edges[1].SHA},
	}
	if !reflect.DeepEqual(edges, want) {
		t.Fatalf("edges mismatch:\n got: %+v\nwant: %+v", edges, want)
	}
	// The evidence pointer is the merge commit's abbreviated SHA.
	if len(edges[1].SHA) != 9 {
		t.Fatalf("edge SHA %q is not the 9-char evidence pointer", edges[1].SHA)
	}
}

// TestGraphCardFileEdgesIgnoreAbsorbMerge — AC-TCI-016 (f): a repo whose only
// merge is absorb-direction produces NO edges; the layer adds no second
// attribution rule, so the subject the engine's own rule refuses stays
// unattributed.
func TestGraphCardFileEdgesIgnoreAbsorbMerge(t *testing.T) {
	root := t.TempDir()
	gitFix(t, root, "init", "-q", "-b", "main")
	gitFix(t, root, "config", "user.email", "fixture@example.com")
	gitFix(t, root, "config", "user.name", "Fixture")
	gitFix(t, root, "config", "gc.auto", "0")
	gitFix(t, root, "config", "gc.autoDetach", "false")

	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		gitFix(t, root, "add", name)
	}
	write("f.txt", "0\n")
	gitFix(t, root, "commit", "-q", "-m", "root")
	gitFix(t, root, "checkout", "-q", "-b", "side")
	write("s1.txt", "s1\n")
	gitFix(t, root, "commit", "-q", "-m", "side work")
	gitFix(t, root, "checkout", "-q", "main")
	gitFix(t, root, "merge", "--no-ff", "-q", "-m", "Merge branch 'side' into feature-other (card t8)", "side")

	edges := cardFileRun(t, root)
	if len(edges) != 0 {
		t.Fatalf("absorb-direction merge produced edges: %+v", edges)
	}
	shas, err := CardAttributedMergeSHAs(root, cardFileLandedBranch(root), factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardAttributedMergeSHAs: %v", err)
	}
	if len(shas) != 0 {
		t.Fatalf("absorb-direction merge entered the fingerprint input: %v", shas)
	}
}

// TestGraphCardFileEdgesKeepNonASCIIPaths — card t1454 card-review r2
// finding 16: the first-parent diff lists files NUL-separated, so a
// non-ASCII path survives byte-for-byte. git's core.quotepath DEFAULT
// C-escapes non-ASCII in its line output, and splitting that output on
// newlines stored the octal escape sequence as the edge's file. The test
// pins the default, so it isolates git from the developer's global config
// — a host setting core.quotepath=false would mask the escape.
func TestGraphCardFileEdgesKeepNonASCIIPaths(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(cfg, []byte("[user]\n\tname = Fixture\n\temail = fixture@example.invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := t.TempDir()
	gitFix(t, root, "init", "-q", "-b", "main")
	gitFix(t, root, "config", "user.email", "fixture@example.com")
	gitFix(t, root, "config", "user.name", "Fixture")
	gitFix(t, root, "config", "gc.auto", "0")
	gitFix(t, root, "config", "gc.autoDetach", "false")
	commit := func(msg string, files map[string]string) {
		t.Helper()
		for name, body := range files {
			path := filepath.Join(root, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			gitFix(t, root, "add", name)
		}
		gitFix(t, root, "commit", "-q", "-m", msg)
	}
	commit("root", map[string]string{"f.txt": "0\n"})
	gitFix(t, root, "checkout", "-q", "-b", "t16br")
	commit("t16 work", map[string]string{filepath.Join("문서", "설계노트.txt"): "설계\n"})
	gitFix(t, root, "checkout", "-q", "main")
	gitFix(t, root, "merge", "--no-ff", "-q", "-m", "merge(t16): WT-t16 into develop - real landing", "t16br")

	edges := cardFileRun(t, root)
	var got string
	for _, e := range edges {
		if e.Card == "t16" {
			got = e.File
		}
	}
	if got != filepath.Join("문서", "설계노트.txt") {
		t.Fatalf("t16's edge file = %q, want the raw path %q", got, filepath.Join("문서", "설계노트.txt"))
	}
}

// TestGraphCardFileEdgesSeeSquashLanding — AC-GCSE-001..003: a card-attributed
// SQUASH landing (a single-parent commit) produces its edge, enters the
// fingerprint input, and flips the fingerprint. A GitHub Flow squash landing
// never creates a merge commit, so a walk restricted to `--merges` commits is
// blind to it (SPEC-GRAPH-CARD-SQUASH-EDGE-001). Arms:
//   - base: the token-free reachable history attributes nothing (AC-GCSE-002
//     control — the broadened walk earns nothing unearned);
//   - landing: exactly one edge {t1560, squash.txt, 9-char SHA}, the landing
//     SHA in the fingerprint input, and a fingerprint difference against the
//     base arm;
//   - root contrast (AC-GCSE-002): a parentless ROOT commit whose subject
//     carries a card token (form 1 scope) makes the attribution gate
//     non-empty for a parentless commit, so the first-parent-diff failure
//     guard (the per-commit continue) is exercised for real — the base arm's
//     token-free subjects skip at the attribution gate before the diff ever
//     runs. Nil error, zero edges.
func TestGraphCardFileEdgesSeeSquashLanding(t *testing.T) {
	initRepo := func(dir string) {
		t.Helper()
		gitFix(t, dir, "init", "-q", "-b", "main")
		gitFix(t, dir, "config", "user.email", "fixture@example.com")
		gitFix(t, dir, "config", "user.name", "Fixture")
		// Same detached-`git maintenance` refusal as cardFileFixture.
		gitFix(t, dir, "config", "gc.auto", "0")
		gitFix(t, dir, "config", "gc.autoDetach", "false")
	}
	commitIn := func(dir, msg string, files map[string]string) {
		t.Helper()
		for name, body := range files {
			path := filepath.Join(dir, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			gitFix(t, dir, "add", name)
		}
		gitFix(t, dir, "commit", "-q", "-m", msg)
	}

	// Landing fixture: token-free base history; squash.txt lives on the card
	// branch, unreachable from main until the squash landing.
	root := t.TempDir()
	initRepo(root)
	commitIn(root, "base", map[string]string{"f.txt": "0\n"})
	gitFix(t, root, "checkout", "-q", "-b", "t1560br")
	commitIn(root, "branch work", map[string]string{"squash.txt": "squash\n"})
	gitFix(t, root, "checkout", "-q", "main")
	branch := cardFileLandedBranch(root)

	// Base arm: nothing attributed before the landing.
	if baseEdges := cardFileRun(t, root); len(baseEdges) != 0 {
		t.Fatalf("token-free base history produced edges: %+v", baseEdges)
	}
	baseSHAs, err := CardAttributedMergeSHAs(root, branch, factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardAttributedMergeSHAs (base): %v", err)
	}
	if len(baseSHAs) != 0 {
		t.Fatalf("token-free base history entered the fingerprint input: %v", baseSHAs)
	}
	baseFP, err := CardMergeFingerprint(root, branch, factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardMergeFingerprint (base): %v", err)
	}

	// The squash landing: a SINGLE-PARENT commit on main whose subject the
	// attributor maps to exactly one card (form 2b — card group + PR group).
	gitFix(t, root, "merge", "--squash", "-q", "t1560br")
	gitFix(t, root, "commit", "-q", "-m", "fix(graph): repair the card-file squash blind spot (card t1560) (#1999)")
	landing := gitFix(t, root, "rev-parse", "HEAD")

	edges := cardFileRun(t, root)
	if len(edges) != 1 {
		t.Fatalf("want exactly one edge for the squash landing, got %+v", edges)
	}
	want := []CardFileEdge{{Card: "t1560", File: "squash.txt", SHA: edges[0].SHA}}
	if !reflect.DeepEqual(edges, want) {
		t.Fatalf("edges mismatch:\n got: %+v\nwant: %+v", edges, want)
	}
	if len(edges[0].SHA) != 9 {
		t.Fatalf("edge SHA %q is not the 9-char evidence pointer", edges[0].SHA)
	}

	// The landing enters the fingerprint input and flips the fingerprint.
	shas, err := CardAttributedMergeSHAs(root, branch, factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardAttributedMergeSHAs: %v", err)
	}
	if len(shas) != 1 || shas[0] != landing {
		t.Fatalf("fingerprint input = %v, want exactly the landing %s", shas, landing)
	}
	fp, err := CardMergeFingerprint(root, branch, factory.AttributeSubject)
	if err != nil {
		t.Fatalf("CardMergeFingerprint: %v", err)
	}
	if fp == baseFP {
		t.Fatalf("fingerprint did not flip after the landing: %s", fp)
	}

	// Root contrast fixture: the ROOT itself carries the card token, so the
	// attribution gate is non-empty for a parentless commit and the
	// first-parent-diff guard decides the outcome.
	rootRoot := t.TempDir()
	initRepo(rootRoot)
	commitIn(rootRoot, "fix(t1561): seeded root (card t1561)", map[string]string{"r.txt": "r\n"})

	rootEdges, err := CardFileEdges(rootRoot, cardFileLandedBranch(rootRoot), factory.AttributeSubject)
	if err != nil {
		t.Fatalf("root fixture errored: %v", err)
	}
	if len(rootEdges) != 0 {
		t.Fatalf("card-attributed root commit produced edges: %+v", rootEdges)
	}
}
