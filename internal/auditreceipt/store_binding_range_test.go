package auditreceipt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Edge coverage for the receipt binding across evidence-only commits
// (SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-009/010, AC-AEC-003a/003b): the section
// parser, the digest split, what WriteReceipt records, and each branch of the range
// check that the hook-driven rows do not reach. The main rows live in
// internal/hook/audit_receipt_range_test.go.

const e2Doc = "# P\n\n## §E.1 One\n\na\n\n## §E.2 Run-phase Evidence\n\nrow\n\n## §E.3 Three\n\nc\n"

func TestE2RawRegion_ParsesHeadingOrdinalAndText(t *testing.T) {
	got, present, ok := e2RawRegion(e2Doc)
	if !ok || !present || got.ordinal != 2 || got.text != "## §E.2 Run-phase Evidence\n\nrow\n" {
		t.Fatalf("e2RawRegion = %+v present=%v ok=%v, want one region at ordinal 2 from its heading up to the E.3 line", got, present, ok)
	}
	if got, present, ok := e2RawRegion("# P\n\n## §E.1 One\n\na\n"); !ok || present {
		t.Errorf("a document without E.2: %+v present=%v ok=%v, want none and unambiguous", got, present, ok)
	}
	if got, present, ok := e2RawRegion(""); !ok || present {
		t.Errorf("an empty document: %+v present=%v ok=%v, want none and unambiguous", got, present, ok)
	}
}

// A heading above E.2 changes its ordinal; content after E.2 is outside it.
func TestE2RawRegion_OrdinalAndBoundary(t *testing.T) {
	moved := strings.Replace(e2Doc, "## §E.2", "## §E.0 New\n\nz\n\n## §E.2", 1)
	a, _, _ := e2RawRegion(e2Doc)
	b, present, ok := e2RawRegion(moved)
	if !ok || !present || b.ordinal != 3 || a.ordinal == b.ordinal {
		t.Errorf("a heading inserted above E.2 must move its ordinal: %+v vs %+v", a, b)
	}
	if e2DocsUnchanged(e2Doc, moved) {
		t.Error("a heading inserted above E.2 compared unchanged")
	}
	after := strings.Replace(e2Doc, "\nc\n", "\nc rewritten\n", 1)
	if x, _, _ := e2RawRegion(e2Doc); !e2DocsUnchanged(e2Doc, after) {
		t.Errorf("a rewrite of E.3 must leave E.2 equal: %+v", x)
	}
}

// A "## " line inside a fenced block is part of the raw text of E.2, so the text after
// it stays in E.2 and an edit there is an edit of E.2.
func TestE2RawRegion_FencedHeadingDoesNotEndTheSection(t *testing.T) {
	doc := "## §E.2 Run\n\n```\n## not a heading\n```\n\nstill E.2\n\n## §E.3 Three\n"
	got, present, ok := e2RawRegion(doc)
	if !ok || !present || !strings.Contains(got.text, "still E.2") {
		t.Fatalf("e2RawRegion = %+v present=%v ok=%v, want the fenced line kept inside E.2 with the text after it", got, present, ok)
	}
	changed := strings.Replace(doc, "still E.2", "still E.2, edited", 1)
	if e2DocsUnchanged(doc, changed) {
		t.Error("an edit after the fenced line compared unchanged; it is an edit of E.2")
	}
}

// Whatever fence shape precedes or sits inside E.2 (AC-AEC-003b, sync-audit D1), the
// raw region still finds the heading and keeps the text under it, so an edit under a
// later E.2 heading still reads as a change. Same inputs as the former fence-tracking
// table; the outcome is now independent of the fence rules.
func TestE2RawRegion_FenceShapesNeverCutTheSection(t *testing.T) {
	const e2Old = "## §E.2 Run-phase Evidence\nAC-1 PASS"
	const e2New = "## §E.2 Run-phase Evidence\nAC-1 FAIL (rewritten)"

	t.Run("a three-backtick line inside a four-backtick fence", func(t *testing.T) {
		const head = "## §E.1 Plan\n````text\nA fence opens with a line like:\n```\n````\n"
		_, oldHas, oldOK := e2RawRegion(head + e2Old)
		_, newHas, newOK := e2RawRegion(head + e2New)
		if !oldHas || !newHas || !oldOK || !newOK {
			t.Fatalf("present old=%v new=%v ok old=%v new=%v, want the E.2 heading after the closed fence seen on both sides", oldHas, newHas, oldOK, newOK)
		}
		if e2DocsUnchanged(head+e2Old, head+e2New) {
			t.Error("a rewrite of the E.2 body compared unchanged; the receipt would stay fresh")
		}
	})

	t.Run("control: no nested fence", func(t *testing.T) {
		const head = "## §E.1 Plan\n```text\nplain\n```\n"
		if e2DocsUnchanged(head+e2Old, head+e2New) {
			t.Error("a rewrite of the E.2 body compared unchanged")
		}
		if !e2DocsUnchanged(head+e2Old, head+e2Old) {
			t.Error("identical documents compared changed")
		}
	})

	t.Run("a backtick run inside a tilde fence", func(t *testing.T) {
		doc := "## §E.2 Run\n\n~~~\n```\n## not a heading\n~~~\n\nstill E.2\n\n## §E.3 Three\n"
		got, present, ok := e2RawRegion(doc)
		if !ok || !present || !strings.Contains(got.text, "still E.2") {
			t.Fatalf("e2RawRegion = %+v present=%v ok=%v, want the text after the tilde fence kept inside E.2", got, present, ok)
		}
	})

	t.Run("a longer closing run", func(t *testing.T) {
		doc := "## §E.1 Plan\n```\ncode\n`````\n## §E.2 Run\nrow\n"
		if _, present, ok := e2RawRegion(doc); !ok || !present {
			t.Fatalf("present=%v ok=%v, want the E.2 heading after the closing run", present, ok)
		}
	})

	t.Run("a closing line carrying text", func(t *testing.T) {
		doc := "## §E.2 Run\n```\n```text\n## not a heading\n```\nstill E.2\n"
		got, present, ok := e2RawRegion(doc)
		if !ok || !present || !strings.Contains(got.text, "still E.2") {
			t.Fatalf("e2RawRegion = %+v present=%v ok=%v, want the info-string line kept as content", got, present, ok)
		}
	})

	t.Run("control: a heading line inside a plain fence", func(t *testing.T) {
		doc := "## §E.2 Run\n```\n## not a heading\n```\nstill E.2\n"
		got, present, ok := e2RawRegion(doc)
		if !ok || !present || !strings.Contains(got.text, "still E.2") {
			t.Fatalf("e2RawRegion = %+v present=%v ok=%v, want the fenced heading kept as text", got, present, ok)
		}
	})
}

// Two E.2 headings are ambiguous; losing one is a change, and so is keeping both.
func TestE2RawRegion_DuplicateHeadings(t *testing.T) {
	doc := "## §E.2 A\n\none\n\n## §E.2 B\n\ntwo\n"
	if _, _, ok := e2RawRegion(doc); ok {
		t.Fatal("two E.2 heading lines read as unambiguous, want ambiguous")
	}
	if e2DocsUnchanged(doc, "## §E.2 A\n\none\n") {
		t.Error("losing one of two E.2 headings compared unchanged")
	}
	if e2DocsUnchanged(doc, doc) {
		t.Error("two E.2 heading lines compared unchanged; an ambiguous document must read as changed")
	}
}

// The head-bound digest keeps the format it always had: the stream starts with the
// head line. Recomputed here independently from the same git reads.
func TestTreeState_DigestFormatIsUnchanged(t *testing.T) {
	root := initContentRepo(t)
	head, digest, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState: %v", err)
	}
	headOut, _ := runGit(t, root, "rev-parse", "HEAD")
	statusOut, _ := gitInRaw(root, "status", "--porcelain", "-z")
	diffOut, _ := runGit(t, root, "diff", "HEAD")
	sum := sha256.Sum256([]byte(fmt.Sprintf("head:%s\x00status\x00%s\x00diff\x00%s", headOut, statusOut, diffOut)))
	if want := hex.EncodeToString(sum[:]); digest != want || head != headOut {
		t.Errorf("TreeState = (%s, %s), want (%s, %s)", head, digest, headOut, want)
	}
}

// The uncommitted-state digest does not depend on the head, and moves with the
// uncommitted state.
func TestUncommittedState_IsHeadIndependent(t *testing.T) {
	root := initContentRepo(t)
	before, err := UncommittedState(root)
	if err != nil {
		t.Fatalf("UncommittedState: %v", err)
	}
	if _, err := runGit(t, root, "-c", "user.email=fx@example.com", "-c", "user.name=fx", "commit", "-q", "--allow-empty", "-m", "empty"); err != nil {
		t.Fatal(err)
	}
	after, _ := UncommittedState(root)
	if before != after {
		t.Errorf("an empty commit moved the uncommitted-state digest: %s -> %s", before, after)
	}
	_, treeBefore, _ := TreeState(root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package a\n\n// edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edited, _ := UncommittedState(root)
	_, treeAfter, _ := TreeState(root)
	if edited == after || treeAfter == treeBefore {
		t.Errorf("an uncommitted edit must move both digests: uncommitted %s -> %s, tree %s -> %s", after, edited, treeBefore, treeAfter)
	}
	if after == treeBefore {
		t.Errorf("the head-free digest must differ from the head-bound one")
	}
}

// WriteReceipt records the uncommitted state only when the state it reads now is the
// state the binding names; a binding captured earlier for a tree that has since moved
// records none.
func TestWriteReceipt_RecordsUncommittedStateOnlyWhenTheBindingHolds(t *testing.T) {
	root := initContentRepo(t)
	ignoreState(t, root)

	t.Run("fallback bind records it", func(t *testing.T) {
		r := Receipt{Tool: ToolAuditMulti, TreeRoot: root}
		id, err := WriteReceipt(root, &r)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := ReadReceipt(root, id)
		want, _ := UncommittedState(root)
		if got.UncommittedDigest == "" || got.UncommittedDigest != want {
			t.Errorf("UncommittedDigest = %q, want %q", got.UncommittedDigest, want)
		}
	})

	t.Run("a start-of-audit binding that still holds records it", func(t *testing.T) {
		head, digest, _ := TreeState(root)
		r := Receipt{Tool: ToolAuditMulti, TreeRoot: root, HeadCommit: head, TreeDigest: digest}
		id, _ := WriteReceipt(root, &r)
		got, _ := ReadReceipt(root, id)
		if got.UncommittedDigest == "" {
			t.Errorf("a binding equal to the current state must record the uncommitted state")
		}
	})

	t.Run("a binding the tree has moved past records none", func(t *testing.T) {
		head, digest, _ := TreeState(root)
		if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package a\n\n// moved during the audit\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := Receipt{Tool: ToolAuditMulti, TreeRoot: root, HeadCommit: head, TreeDigest: digest}
		id, _ := WriteReceipt(root, &r)
		got, _ := ReadReceipt(root, id)
		if got.UncommittedDigest != "" {
			t.Errorf("UncommittedDigest = %q after the tree moved past the binding; the newer state must not be recorded as the judged one", got.UncommittedDigest)
		}
	})

	t.Run("a capture-failed receipt records none", func(t *testing.T) {
		r := Receipt{Tool: ToolAuditMulti, TreeRoot: root, BindCaptureFailed: true}
		id, _ := WriteReceipt(root, &r)
		got, _ := ReadReceipt(root, id)
		if got.UncommittedDigest != "" {
			t.Errorf("a capture-failed receipt recorded %q", got.UncommittedDigest)
		}
	})
}

// ignoreState commits an ignore rule for the runtime state directory, as a project
// carries one: the receipt written under it must not show up as uncommitted state.
func ignoreState(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".moai/state/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, root, "ignore runtime state")
}

// rangeFixture: a repo with an audited SPEC progress.md, and a receipt judged at its
// head.
func rangeFixture(t *testing.T) (root string, r Receipt) {
	t.Helper()
	root = initContentRepo(t)
	ignoreState(t, root)
	write := func(rel, body string) {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".moai/specs/SPEC-X-001/progress.md", e2Doc)
	commitAll(t, root, "base")
	id, err := WriteReceipt(root, &Receipt{Tool: ToolAuditMulti, TreeRoot: root, CreatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	r, _ = ReadReceipt(root, id)
	return root, r
}

func commitAll(t *testing.T, root, msg string) {
	t.Helper()
	if _, err := runGit(t, root, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(t, root, "-c", "user.email=fx@example.com", "-c", "user.name=fx", "commit", "-q", "--allow-empty", "-m", msg); err != nil {
		t.Fatal(err)
	}
}

func currentHead(t *testing.T, root string) string {
	t.Helper()
	h, err := runGit(t, root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestEvidenceOnlyRange_Edges(t *testing.T) {
	const spec = "SPEC-X-001"
	progress := ".moai/specs/" + spec + "/progress.md"

	t.Run("an empty commit nets to no change and keeps the receipt", func(t *testing.T) {
		root, r := rangeFixture(t)
		commitAll(t, root, "empty")
		if !evidenceOnlyRange(r, currentHead(t, root), spec) {
			t.Error("an empty commit is evidence-only")
		}
	})
	t.Run("a recorded head that is not a commit id is refused before git reads it", func(t *testing.T) {
		root, r := rangeFixture(t)
		commitAll(t, root, "empty")
		r.HeadCommit = "--output=/tmp/x"
		if evidenceOnlyRange(r, currentHead(t, root), spec) {
			t.Error("a malformed recorded head was accepted")
		}
		r.HeadCommit = currentHead(t, root)
		if evidenceOnlyRange(r, "-x", spec) {
			t.Error("a malformed current head was accepted")
		}
	})
	t.Run("a head git does not know reads as stale", func(t *testing.T) {
		root, r := rangeFixture(t)
		r.HeadCommit = strings.Repeat("a", 40)
		if evidenceOnlyRange(r, currentHead(t, root), spec) {
			t.Error("an unknown recorded head was accepted")
		}
	})
	t.Run("a progress.md added in the range without E.2 is allowed, with E.2 it is not", func(t *testing.T) {
		root, r := rangeFixture(t)
		other := ".moai/specs/SPEC-Z-001/progress.md"
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, other)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, other), []byte("# P\n\n## §E.1 One\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		commitAll(t, root, "add")
		if !evidenceOnlyRange(r, currentHead(t, root), "SPEC-Z-001") {
			t.Error("a new progress.md without E.2 under the cited SPEC is evidence-only")
		}
		if err := os.WriteFile(filepath.Join(root, other), []byte("# P\n\n## §E.1 One\n\n## §E.2 Run\n\nx\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		commitAll(t, root, "grow")
		if evidenceOnlyRange(r, currentHead(t, root), "SPEC-Z-001") {
			t.Error("a progress.md that gained an E.2 section after the receipt is a change of E.2")
		}
	})
	t.Run("a deleted progress.md that had E.2 is stale", func(t *testing.T) {
		root, r := rangeFixture(t)
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(progress))); err != nil {
			t.Fatal(err)
		}
		commitAll(t, root, "delete")
		if evidenceOnlyRange(r, currentHead(t, root), spec) {
			t.Error("deleting the progress.md that carried E.2 was accepted")
		}
	})
	t.Run("a type change of an allowed file is stale", func(t *testing.T) {
		root, r := rangeFixture(t)
		rel := ".moai/reports/t1/verdict.md"
		if err := os.MkdirAll(filepath.Join(root, ".moai/reports/t1"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte("v\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		commitAll(t, root, "file")
		r2 := r
		r2.HeadCommit = currentHead(t, root)
		r2.UncommittedDigest, _ = UncommittedState(root)
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("/etc/hostname", filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		commitAll(t, root, "symlink")
		if evidenceOnlyRange(r2, currentHead(t, root), spec) {
			t.Error("replacing an evidence file with a symlink was accepted")
		}
	})
	t.Run("a card id, unknown-spec or empty citation allows no progress.md", func(t *testing.T) {
		root, r := rangeFixture(t)
		body := strings.Replace(e2Doc, "\na\n", "\na edited\n", 1)
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(progress)), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		commitAll(t, root, "E.1 only")
		head := currentHead(t, root)
		if !evidenceOnlyRange(r, head, spec) {
			t.Fatal("control: an E.1-only edit under the cited SPEC is evidence-only")
		}
		for _, cited := range []string{"t1717", "card-t1", UnknownSpec, "", "SPEC-X-001/../SPEC-Y-001", "../SPEC-X-001"} {
			if evidenceOnlyRange(r, head, cited) {
				t.Errorf("citation %q allowed a progress.md change", cited)
			}
		}
	})
}

// The three-argument entry point never allows a progress.md: the spec= value is what
// admits one.
func TestCheckCitedReceiptsSince_AllowsNoProgressFile(t *testing.T) {
	root, r := rangeFixture(t)
	body := strings.Replace(e2Doc, "\na\n", "\na edited\n", 1)
	if err := os.WriteFile(filepath.Join(root, ".moai/specs/SPEC-X-001/progress.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, root, "E.1 only")
	start := &StartMarker{AgentID: "a", TreeRoot: root, StartedAt: time.Now().UTC().Add(-time.Hour)}
	if ok, cause := CheckCitedReceiptsSince(root, start, time.Time{}, []string{r.ReceiptID}); ok || cause != CauseReceiptTreeStateChanged {
		t.Errorf("CheckCitedReceiptsSince = (%v, %q), want a stale receipt", ok, cause)
	}
	if ok, cause := CheckCitedReceiptsForSpec(root, start, time.Time{}, []string{r.ReceiptID}, "SPEC-X-001"); !ok {
		t.Errorf("CheckCitedReceiptsForSpec = (%v, %q), want the receipt fresh", ok, cause)
	}
}
