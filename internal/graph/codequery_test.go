package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// codeQueryFixture: A→B→C chain plus an exported helper with args.
func codeQueryFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "chain"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package chain

import "fmt"

func A() {
	B()
}

func B() {
	C()
}

func C() {
	fmt.Println("done")
}

func Helper(s string, n int) (string, error) {
	return s, nil
}
`
	if err := os.WriteFile(filepath.Join(root, "internal", "chain", "chain.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	// Build the artifact so the query tools have their substrate.
	edges, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteJSONL(filepath.Join(root, ".moai", "project", "graph", "edges.jsonl"), edges); err != nil {
		t.Fatal(err)
	}
	return root
}

// F1 regression (sync-audit MUST-FIX): a `..` escape attempt in the
// LLM-facing file parameter is REJECTED — never resolved outside the root.
func TestFileAPI_RejectsPathEscape(t *testing.T) {
	root := codeQueryFixture(t)
	for _, escape := range []string{
		"../../../etc/secrets/creds.go",
		"internal/chain/../../../../etc/passwd",
		"..",
	} {
		if _, err := FileAPI(root, escape); err == nil {
			t.Errorf("FileAPI(%q) escaped the project root without error", escape)
		}
	}
}

// AC-GF-020 — file_api answers exported signatures only: Helper's full
// signature with parameter names, no function bodies anywhere.
func TestFileAPI_SignaturesOnly(t *testing.T) {
	root := codeQueryFixture(t)

	res, err := FileAPI(root, "internal/chain/chain.go")
	if err != nil {
		t.Fatalf("FileAPI: %v", err)
	}
	if res.File != "internal/chain/chain.go" {
		t.Errorf("File = %q", res.File)
	}
	foundHelper := false
	for _, s := range res.Symbols {
		if strings.Contains(s.Signature, "func Helper(s string, n int) (string, error)") {
			foundHelper = true
		}
		if strings.Contains(s.Signature, "return s, nil") {
			t.Errorf("file_api leaked a function body: %q", s.Signature)
		}
	}
	if !foundHelper {
		t.Errorf("Helper signature missing; symbols=%+v", res.Symbols)
	}
	// Provenance names the tree (REQ-GF-019).
	if !strings.Contains(res.Provenance, root) {
		t.Errorf("provenance must name the tree root, got: %q", res.Provenance)
	}
}

// AC-GF-021 — find_code + trace_calls answer from the code layer.
func TestFindCodeAndTraceCalls(t *testing.T) {
	root := codeQueryFixture(t)

	found, prov, err := FindCode(root, "B")
	if err != nil {
		t.Fatalf("FindCode: %v", err)
	}
	hitB := false
	for _, m := range found {
		if m.Symbol == "B" {
			hitB = true
			if m.Grade == "" {
				t.Errorf("match %+v lacks grade", m)
			}
		}
	}
	if !hitB {
		t.Errorf("find B returned nothing: %+v", found)
	}
	if !strings.Contains(prov, root) {
		t.Errorf("find provenance must name the tree: %q", prov)
	}

	// trace from B: caller {A}, callee {C} via code-call edges.
	callers, callees, err := TraceCalls(root, "B", 1)
	if err != nil {
		t.Fatalf("TraceCalls: %v", err)
	}
	callerA, calleeC := false, false
	for _, e := range callers {
		if strings.HasSuffix(e.From, ":A") && e.To == "B" {
			callerA = true
		}
	}
	for _, e := range callees {
		if strings.HasSuffix(e.From, ":B") && e.To == "C" {
			calleeC = true
		}
	}
	if !callerA {
		t.Errorf("A missing from B's callers: %+v", callers)
	}
	if !calleeC {
		t.Errorf("C missing from B's callees: %+v", callees)
	}
}

// Per-tree anchoring (REQ-GF-019 / the t246 defect family): two trees with
// different content answer from their OWN artifacts — no cross-tree reuse.
func TestCodeQueries_PerTreeAnswers(t *testing.T) {
	treeA := codeQueryFixture(t)
	treeB := codeQueryFixture(t)

	// treeB gains a new caller of C: D() { C() }.
	if err := os.WriteFile(filepath.Join(treeB, "internal", "chain", "extra.go"),
		[]byte("package chain\n\nfunc D() {\n\tC()\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edges, _, err := BuildWithCodeLayers(treeB)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteJSONL(filepath.Join(treeB, ".moai", "project", "graph", "edges.jsonl"), edges); err != nil {
		t.Fatal(err)
	}

	callersA, _, err := TraceCalls(treeA, "C", 1)
	if err != nil {
		t.Fatal(err)
	}
	callersB, _, err := TraceCalls(treeB, "C", 1)
	if err != nil {
		t.Fatal(err)
	}
	inB := false
	for _, e := range callersB {
		if strings.HasSuffix(e.From, ":D") {
			inB = true
		}
	}
	for _, e := range callersA {
		if strings.HasSuffix(e.From, ":D") {
			t.Error("tree A answered with tree B's content — cross-tree leak (t246 family)")
		}
	}
	if !inB {
		t.Errorf("tree B's own answer must include its D caller: %+v", callersB)
	}
}
