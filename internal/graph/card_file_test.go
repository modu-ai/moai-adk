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
