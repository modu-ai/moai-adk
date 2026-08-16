package backup

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// mustDecodeDoc decodes a YAML string into a root mapping node for tests.
func mustDecodeDoc(t *testing.T, s string) *yaml.Node {
	t.Helper()
	n, err := decodeDoc([]byte(s), "test")
	if err != nil {
		t.Fatalf("decodeDoc(%q): %v", s, err)
	}
	return n
}

func TestEncodeNode_EmptyMapping(t *testing.T) {
	out, err := encodeNode(emptyMappingNode())
	if err != nil {
		t.Fatalf("encodeNode: %v", err)
	}
	if string(out) != "{}\n" {
		t.Errorf("empty mapping encode = %q, want %q", string(out), "{}\n")
	}
}

func TestEncodeNode_TwoSpaceIndent(t *testing.T) {
	in := []byte("a:\n  b: 1\n")
	n, err := decodeDoc(in, "test")
	if err != nil {
		t.Fatal(err)
	}
	out, err := encodeNode(n)
	if err != nil {
		t.Fatalf("encodeNode: %v", err)
	}
	// AC-UYP-004: each nesting level adds exactly 2 spaces.
	if !strings.Contains(string(out), "  b: 1") {
		t.Errorf("expected 2-space indent, got:\n%s", string(out))
	}
	// And NOT 4-space.
	if strings.Contains(string(out), "    b: 1") {
		t.Errorf("found 4-space indent (should be 2):\n%s", string(out))
	}
}

func TestNodeValuesEqual_Scalars(t *testing.T) {
	a := mustDecodeDoc(t, "k: 1\n")
	b := mustDecodeDoc(t, "k: 1\n")
	av, _ := mappingGet(a, "k")
	bv, _ := mappingGet(b, "k")
	if !nodeValuesEqual(av, bv) {
		t.Errorf("equal scalars should be equal")
	}

	c := mustDecodeDoc(t, "k: 2\n")
	cv, _ := mappingGet(c, "k")
	if nodeValuesEqual(av, cv) {
		t.Errorf("1 vs 2 should be unequal")
	}
}

// REQ-UYP-016: explicit null is distinct from empty mapping.
func TestNodeValuesEqual_NullVsEmptyMap(t *testing.T) {
	docNull := mustDecodeDoc(t, "k:\n")
	docMap := mustDecodeDoc(t, "k: {}\n")
	nullV, _ := mappingGet(docNull, "k")
	mapV, _ := mappingGet(docMap, "k")
	if nodeValuesEqual(nullV, mapV) {
		t.Errorf("explicit null must NOT equal empty mapping (REQ-UYP-016)")
	}
}

func TestNodeValuesEqual_FloatInt(t *testing.T) {
	// Cross-type numeric parity with ValuesEqual: 1 == 1.0.
	a := mustDecodeDoc(t, "k: 1\n")
	b := mustDecodeDoc(t, "k: 1.0\n")
	av, _ := mappingGet(a, "k")
	bv, _ := mappingGet(b, "k")
	if !nodeValuesEqual(av, bv) {
		t.Errorf("int 1 vs float 1.0 should be equal (parity with ValuesEqual)")
	}
}

// mergeMappingNode3Way: user unchanged from base takes new template value.
func TestMergeMappingNode_UserUnchangedTakesNew(t *testing.T) {
	newN := mustDecodeDoc(t, "k: new_template\n")
	oldN := mustDecodeDoc(t, "k: base_value\n")
	baseN := mustDecodeDoc(t, "k: base_value\n")
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, oldN, baseN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	v, _ := mappingGet(merged, "k")
	if v.Value != "new_template" {
		t.Errorf("expected new_template, got %s", v.Value)
	}
}

// mergeMappingNode3Way: user changed from base preserves user value.
func TestMergeMappingNode_UserChangedPreservesUser(t *testing.T) {
	newN := mustDecodeDoc(t, "k: new_template\n")
	oldN := mustDecodeDoc(t, "k: user_custom\n")
	baseN := mustDecodeDoc(t, "k: base_value\n")
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, oldN, baseN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	v, _ := mappingGet(merged, "k")
	if v.Value != "user_custom" {
		t.Errorf("expected user_custom, got %s", v.Value)
	}
}

// mergeMappingNode3Way: old-only key absent from base is preserved + reported.
func TestMergeMappingNode_OldOnlyAbsentFromBasePreserved(t *testing.T) {
	newN := mustDecodeDoc(t, "shared: v\n")
	oldN := mustDecodeDoc(t, "shared: v\nuser_added: custom\n")
	baseN := mustDecodeDoc(t, "shared: base\n")
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, oldN, baseN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	v, ok := mappingGet(merged, "user_added")
	if !ok {
		t.Fatal("user_added should be preserved (REQ-UYP-006)")
	}
	if v.Value != "custom" {
		t.Errorf("user_added value = %s, want custom", v.Value)
	}
	if !strings.Contains(buf.String(), "user_added") {
		t.Errorf("advisory should name user_added (REQ-UYP-007), got: %s", buf.String())
	}
}

// mergeMappingNode3Way: old-only key present in base (template removed) is ALSO
// retained under REQ-UYP-006 (policy reversal) and reported.
func TestMergeMappingNode_TemplateRemovedKeyStillRetained(t *testing.T) {
	newN := mustDecodeDoc(t, "kept: v\n")
	oldN := mustDecodeDoc(t, "kept: v\nretired: gone\n")
	baseN := mustDecodeDoc(t, "kept: v\nretired: gone\n")
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, oldN, baseN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := mappingGet(merged, "retired"); !ok {
		t.Error("retired key must be retained per REQ-UYP-006 (policy reversal)")
	}
	if !strings.Contains(buf.String(), "retired") {
		t.Errorf("retained retired key should be reported (REQ-UYP-007): %s", buf.String())
	}
}

// mergeMappingNode3Way: key order follows the new template, old-only keys appended.
func TestMergeMappingNode_KeyOrderPreserved(t *testing.T) {
	newN := mustDecodeDoc(t, "a: 1\nb: 2\n")
	oldN := mustDecodeDoc(t, "a: 1\nb: 2\nextra: 9\n")
	baseN := mustDecodeDoc(t, "a: 1\nb: 2\n")
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, oldN, baseN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	keys := mappingKeys(merged)
	want := []string{"a", "b", "extra"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	for i, k := range want {
		if keys[i] != k {
			t.Errorf("keys[%d] = %s, want %s", i, keys[i], k)
		}
	}
}

// mappingKeys returns the ordered scalar keys of a mapping node (test helper).
func mappingKeys(n *yaml.Node) []string {
	var keys []string
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i] != nil && n.Content[i].Kind == yaml.ScalarNode {
			keys = append(keys, n.Content[i].Value)
		}
	}
	return keys
}

// REQ-UYP-012: anchors/aliases are carried through, not expanded.
func TestNodeMerge_AliasNotExpanded(t *testing.T) {
	src := []byte("base: &a {k: v}\nderived: *a\n")
	newN, err := decodeDoc(src, "test")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, newN, newN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	out, err := encodeNode(merged)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "*a") {
		t.Errorf("alias *a should be preserved verbatim, got:\n%s", s)
	}
	// Should not contain a second inlined {k: v} from expansion.
	if strings.Count(s, "k: v") > 1 {
		t.Errorf("alias appears expanded (k: v inlined twice), got:\n%s", s)
	}
}

// REQ-UYP-013: merge key `<<` is treated as an ordinary key, not resolved.
func TestNodeMerge_MergeKeyNotResolved(t *testing.T) {
	src := []byte("defaults: &d\n  x: 1\nchild:\n  <<: *d\n  y: 2\n")
	newN, err := decodeDoc(src, "test")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, newN, newN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	out, err := encodeNode(merged)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "<<:") {
		t.Errorf("merge key `<<:` should appear as an ordinary key, got:\n%s", s)
	}
}

// REQ-UYP-014: sequences are replaced wholesale, never appended.
func TestNodeMerge_SequenceReplaced(t *testing.T) {
	newN := mustDecodeDoc(t, "items: [a, b]\n")
	oldN := mustDecodeDoc(t, "items: [c]\n")
	baseN := mustDecodeDoc(t, "items: [a, b]\n")
	var buf bytes.Buffer
	merged, err := mergeMappingNode3Way(newN, oldN, baseN, "", newRetainedKeyNotes(&buf))
	if err != nil {
		t.Fatal(err)
	}
	v, _ := mappingGet(merged, "items")
	out, err := encodeNode(v)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	// old [c] differs from base [a,b] → user wins → exactly [c].
	if !strings.Contains(s, "c") {
		t.Errorf("expected [c], got %s", s)
	}
	if strings.Contains(s, "a") || strings.Contains(s, "b") {
		t.Errorf("sequence was appended instead of replaced, got %s", s)
	}
}

// multi-document input is rejected (REQ-UYP-015) at decode time.
func TestDecodeDoc_MultiDocumentErrors(t *testing.T) {
	multi := []byte("a: 1\n---\nb: 2\n")
	_, err := decodeDoc(multi, "test")
	if err == nil {
		t.Fatal("expected error for multi-document input")
	}
	if !strings.Contains(err.Error(), "multi-document") {
		t.Errorf("error should name multi-document, got: %v", err)
	}
}
