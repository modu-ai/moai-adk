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

func TestE2Sections_ParsesHeadingOrdinalAndBody(t *testing.T) {
	got := e2Sections(e2Doc)
	if len(got) != 1 || got[0].ordinal != 2 || got[0].heading != "## §E.2 Run-phase Evidence" || got[0].body != "\nrow\n\n" {
		t.Fatalf("e2Sections = %+v, want one section at ordinal 2 with its heading and body", got)
	}
	if got := e2Sections("# P\n\n## §E.1 One\n\na\n"); len(got) != 0 {
		t.Errorf("a document without E.2: %+v, want none", got)
	}
	if got := e2Sections(""); len(got) != 0 {
		t.Errorf("an empty document: %+v, want none", got)
	}
}

// A heading above E.2 changes its ordinal; content after E.2 is outside it.
func TestE2Sections_OrdinalAndBoundary(t *testing.T) {
	moved := strings.Replace(e2Doc, "## §E.2", "## §E.0 New\n\nz\n\n## §E.2", 1)
	a, b := e2Sections(e2Doc), e2Sections(moved)
	if len(b) != 1 || b[0].ordinal != 3 || a[0].ordinal == b[0].ordinal {
		t.Errorf("a heading inserted above E.2 must move its ordinal: %+v vs %+v", a, b)
	}
	after := strings.Replace(e2Doc, "\nc\n", "\nc rewritten\n", 1)
	if x, y := e2Sections(e2Doc), e2Sections(after); len(y) != 1 || x[0] != y[0] {
		t.Errorf("a rewrite of E.3 must leave E.2 equal: %+v vs %+v", x, y)
	}
}

// A "## " line inside a fenced block is not a heading, so text after it stays in E.2.
func TestE2Sections_FencedHeadingDoesNotEndTheSection(t *testing.T) {
	doc := "## §E.2 Run\n\n```\n## not a heading\n```\n\nstill E.2\n\n## §E.3 Three\n"
	got := e2Sections(doc)
	if len(got) != 1 || !strings.Contains(got[0].body, "still E.2") {
		t.Fatalf("e2Sections = %+v, want the fenced line kept inside E.2 with the text after it", got)
	}
	changed := strings.Replace(doc, "still E.2", "still E.2, edited", 1)
	if other := e2Sections(changed); len(other) != 1 || other[0] == got[0] {
		t.Errorf("an edit after the fenced line is an edit of E.2: %+v", other)
	}
}

// Two E.2 headings are two sections; losing one is a change.
func TestE2Sections_DuplicateHeadings(t *testing.T) {
	doc := "## §E.2 A\n\none\n\n## §E.2 B\n\ntwo\n"
	if got := e2Sections(doc); len(got) != 2 {
		t.Fatalf("e2Sections = %+v, want both sections", got)
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
