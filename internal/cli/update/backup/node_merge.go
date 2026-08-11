package backup

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// This file implements the node-tree merge primitives used by merge.go's
// entry points. Decoding YAML into a *yaml.Node (rather than map[string]any)
// preserves every comment (HeadComment/LineComment/FootComment), the source
// key order, the scalar quoting style, and the anchor/alias nodes — none of
// which survive a map[string]any round-trip. This is the fix for issue #1243,
// where `moai update` silently stripped comments and reordered keys.
//
// yaml.v3 mapping nodes store their Content as a flat [k1,v1,k2,v2,...] slice
// (scalar key node, value node, repeated). The mapping helpers below (mappingGet
// / mappingHasKey / appendPairKeyed) operate on that pairing convention.

// systemFields are keys whose value is always taken from the NEW template
// (never preserved from the user's old file), on both the 3-way and 2-way
// paths. This MUST stay in parity between DeepMerge3Way and DeepMergeMaps: a
// restore that degrades from 3-way to 2-way (missing template defaults) must
// not resurrect a stale version over the deployed one.
var systemFields = map[string]bool{
	"template_version": true,
	"version":          true,
}

// retainedKeySink receives advisory notices for keys that DeepMerge3Way retains
// because they are absent from the new template (REQ-UYP-007). It defaults to
// os.Stderr (the production path through MergeYAML3Way); tests in this package
// swap it for a bytes.Buffer to assert on the advisory text.
var retainedKeySink io.Writer = os.Stderr

// SetRetainedKeySinkForTest sets the writer that receives retained-key
// advisories, returning a restore function that resets it to the prior value.
// It is intended for cross-package tests (e.g. SPEC-CONFIG-KEY-HONESTY-001
// AC-CKH-014) that assert over the advisory from outside this package; the
// default sink is os.Stderr. Calling the returned restore function is required
// so concurrent tests do not observe the swapped sink. This is test
// infrastructure only — it changes no merge semantics.
func SetRetainedKeySinkForTest(w io.Writer) (restore func()) {
	prev := retainedKeySink
	retainedKeySink = w
	return func() { retainedKeySink = prev }
}

// encodeNode serializes a node tree with the project's canonical 2-space block
// indentation (REQ-UYP-004). The encoder is closed explicitly so the trailing
// document marker is written.
func encodeNode(n *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(n); err != nil {
		return nil, fmt.Errorf("encode merged YAML: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close YAML encoder: %w", err)
	}
	return buf.Bytes(), nil
}

// emptyMappingNode returns a zero-entry mapping node, used as the merge result
// for empty/null input so the output renders as "{}" rather than "null".
func emptyMappingNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

// decodeDoc decodes a YAML byte slice into a root node, unwrapping the
// DocumentNode wrapper. Empty/whitespace-only input yields an empty mapping
// node. Multi-document input (more than one document separated by "---")
// returns a wrapped error so the caller can fall back (REQ-UYP-015).
func decodeDoc(data []byte, label string) (*yaml.Node, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return emptyMappingNode(), nil
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var doc yaml.Node
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("unmarshal %s YAML: %w", label, err)
	}
	// Multi-document detection (REQ-UYP-015): a second successful decode, or any
	// error other than io.EOF, means trailing documents exist. We refuse rather
	// than silently dropping them.
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("unmarshal %s YAML: multi-document input is not supported", label)
	}
	root := &doc
	var docHead, docFoot string
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		// Capture document-level comments before unwrapping. yaml.v3 attaches a
		// leading file comment and a trailing file comment to the DocumentNode,
		// not to the root mapping; if we do not carry them onto the root, the
		// encoder (which emits a bare mapping, not a document) drops them.
		docHead = root.HeadComment
		docFoot = root.FootComment
		root = root.Content[0]
	}
	if root == nil || root.Kind == 0 {
		return emptyMappingNode(), nil
	}
	// Splice document-level comments onto the root mapping so they survive the
	// unwrap. A root that already carries its own head/foot comment keeps both
	// (document comment first, then the root's own).
	if docHead != "" {
		if root.HeadComment != "" {
			root.HeadComment = docHead + "\n" + root.HeadComment
		} else {
			root.HeadComment = docHead
		}
	}
	if docFoot != "" {
		if root.FootComment != "" {
			root.FootComment = root.FootComment + "\n" + docFoot
		} else {
			root.FootComment = docFoot
		}
	}
	// A config section must be a mapping. The node decoder parses some inputs
	// the map decoder rejects (e.g. a bare scalar like "invalid[yaml" lands as
	// a ScalarNode rather than erroring), so validate the root kind here: a
	// non-mapping root (that is not an explicit null/empty document) is treated
	// as unparseable and returns a wrapped error (REQ-UYP-011), matching the
	// prior map-decoder behavior that rejected non-map roots.
	if root.Kind != yaml.MappingNode {
		if isNullNode(root) {
			return emptyMappingNode(), nil
		}
		return nil, fmt.Errorf("unmarshal %s YAML: root is not a mapping", label)
	}
	return root, nil
}

// mappingGet returns the value node paired with the given scalar key in a
// mapping node, or (nil, false) if absent.
func mappingGet(n *yaml.Node, key string) (*yaml.Node, bool) {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k := n.Content[i]
		if k != nil && k.Kind == yaml.ScalarNode && k.Value == key {
			return n.Content[i+1], true
		}
	}
	return nil, false
}

// mappingHasKey reports whether a scalar key exists in a mapping node.
func mappingHasKey(n *yaml.Node, key string) bool {
	_, ok := mappingGet(n, key)
	return ok
}

// appendPairKeyed appends a (key, value) pair, copying the supplied source key
// node so its HeadComment/LineComment/FootComment survive the merge. This is
// what makes comment preservation work: a comment that yaml.v3 attaches to a
// key node (e.g. a comment immediately above a key) is carried with the key.
func appendPairKeyed(m, srcKey, value *yaml.Node) {
	m.Content = append(m.Content, copyNode(srcKey), value)
}

// nodeValuesEqual reports whether two nodes decode to equal values. It decodes
// each node into a generic any and delegates to ValuesEqual, preserving the
// cross-type numeric equality (int 1 == float 1.0) of the original map-based
// comparison. An explicit null node and an empty mapping node decode to nil
// and map[string]any{} respectively, so they compare unequal (REQ-UYP-016).
func nodeValuesEqual(a, b *yaml.Node) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	var av, bv any
	if err := a.Decode(&av); err != nil {
		return false
	}
	if err := b.Decode(&bv); err != nil {
		return false
	}
	return ValuesEqual(av, bv)
}

// isMappingNode reports whether n is a non-null mapping node.
func isMappingNode(n *yaml.Node) bool {
	return n != nil && n.Kind == yaml.MappingNode && !isNullNode(n)
}

// isNullNode reports whether n is an explicit null (Kind Scalar, Tag !!null).
func isNullNode(n *yaml.Node) bool {
	if n == nil {
		return true
	}
	return n.Kind == yaml.ScalarNode && (n.Tag == "!!null" || n.Tag == "")
}

// deepMerge3WayTo is the testable core of DeepMerge3Way. It writes retained-key
// advisories (keys absent from the new template, REQ-UYP-007) to w. The public
// DeepMerge3Way passes os.Stderr (via retainedKeySink); tests pass a buffer.
//
// Config roots are mappings, so the value-level decision (old==base → new,
// old!=base → user, no-base → user) is applied inside mergeMappingNode3Way for
// each value. This top-level dispatch only handles the defensive non-mapping
// root (e.g. a scalar-only document): a top-level scalar carries no structure
// to merge, so the user value is preserved when present.
func deepMerge3WayTo(newNode, oldNode, baseNode *yaml.Node, w io.Writer) (*yaml.Node, error) {
	if isMappingNode(newNode) && isMappingNode(oldNode) {
		return mergeMappingNode3Way(newNode, oldNode, baseNode, "", w)
	}
	// Defensive: non-mapping root. Preserve the user value if present.
	if oldNode != nil && !isNullNode(oldNode) {
		return copyNode(oldNode), nil
	}
	return copyNode(newNode), nil
}

// mergeMappingNode3Way walks the new mapping's keys in source order, resolving
// each against the old (user) and base (previous template) mappings, then
// appends old-only keys at the end (REQ-UYP-006). Comments and key order come
// from whichever document supplied each surviving node.
func mergeMappingNode3Way(newNode, oldNode, baseNode *yaml.Node, path string, w io.Writer) (*yaml.Node, error) {
	result := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	// Inherit the template's (newNode's) comment and style metadata so nested
	// mappings — including flow mappings like `{ effort: max }` carrying a
	// trailing LineComment — survive with their comments and style intact.
	if newNode != nil {
		result.HeadComment = newNode.HeadComment
		result.LineComment = newNode.LineComment
		result.FootComment = newNode.FootComment
		result.Style = newNode.Style
	}

	// First pass: resolve every key present in the new template.
	for i := 0; i+1 < len(newNode.Content); i += 2 {
		keyNode := newNode.Content[i]
		newV := newNode.Content[i+1]
		if keyNode == nil || keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value
		childPath := joinPath(path, key)

		// System fields always take the new template value (REQ-UYP-010).
		if systemFields[key] {
			appendPairKeyed(result, keyNode, copyNode(newV))
			continue
		}

		oldV, oldExists := mappingGet(oldNode, key)
		if !oldExists {
			// New field added by the template → take it.
			appendPairKeyed(result, keyNode, copyNode(newV))
			continue
		}

		// Both new and old present.
		var baseV *yaml.Node
		baseExists := false
		if baseNode != nil {
			baseV, baseExists = mappingGet(baseNode, key)
		}

		if isMappingNode(newV) && isMappingNode(oldV) {
			// Both mappings → recurse.
			var recBase *yaml.Node
			if baseExists {
				recBase = baseV
			}
			merged, err := mergeMappingNode3Way(newV, oldV, recBase, childPath, w)
			if err != nil {
				return nil, err
			}
			appendPairKeyed(result, keyNode, merged)
			continue
		}

		// Scalar / sequence value: apply the decision.
		if !baseExists {
			// No base entry → user added this value → preserve old.
			appendPairKeyed(result, keyNode, copyNode(oldV))
		} else if nodeValuesEqual(oldV, baseV) {
			// User unchanged from base → take new template value.
			appendPairKeyed(result, keyNode, copyNode(newV))
		} else {
			// User changed from base → preserve user value.
			appendPairKeyed(result, keyNode, copyNode(oldV))
		}
	}

	// Second pass: old-only keys (absent from the new template). REQ-UYP-006
	// retains ALL of them (the prior distinction between user-added and
	// template-removed is reversed), and REQ-UYP-007 reports each retained key.
	for i := 0; i+1 < len(oldNode.Content); i += 2 {
		keyNode := oldNode.Content[i]
		oldV := oldNode.Content[i+1]
		if keyNode == nil || keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value
		if systemFields[key] {
			continue
		}
		if mappingHasKey(newNode, key) {
			continue // resolved in the first pass
		}
		childPath := joinPath(path, key)
		appendPairKeyed(result, keyNode, copyNode(oldV))
		_, _ = fmt.Fprintf(w, "advisory: retained key %q absent from new template (preserved from user config)\n", childPath)
	}

	return result, nil
}

// joinPath builds a dotted key path for advisory messages.
func joinPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

// (parentOf removed — the value-level decision in mergeMappingNode3Way reads
// baseNode directly via mappingGet, so no parent lookup is needed.)

// copyNode returns a shallow copy of n. A shallow copy is sufficient because
// the merged tree is read once by encodeNode and then discarded; the original
// nodes' Content slices are never mutated, so sharing child pointers is safe.
func copyNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	c := *n
	return &c
}

// deepMergeMapsTo is the testable core of DeepMergeMaps (2-way fallback). It
// preserves old values where keys exist in both, adds old-only keys, and
// always takes the new value for system fields. No advisory is emitted on the
// 2-way path (REQ-UYP-007 binds only DeepMerge3Way).
func deepMergeMapsTo(newNode, oldNode *yaml.Node) (*yaml.Node, error) {
	if !isMappingNode(newNode) || !isMappingNode(oldNode) {
		// Non-mapping top level: keep old if present, else new.
		if oldNode != nil {
			return copyNode(oldNode), nil
		}
		return copyNode(newNode), nil
	}
	result := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if newNode != nil {
		result.HeadComment = newNode.HeadComment
		result.LineComment = newNode.LineComment
		result.FootComment = newNode.FootComment
		result.Style = newNode.Style
	}

	// Copy all new values in source order.
	for i := 0; i+1 < len(newNode.Content); i += 2 {
		keyNode := newNode.Content[i]
		newV := newNode.Content[i+1]
		if keyNode == nil || keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value
		oldV, oldExists := mappingGet(oldNode, key)
		if !oldExists {
			appendPairKeyed(result, keyNode, copyNode(newV))
			continue
		}
		// System fields always use the new value.
		if systemFields[key] {
			appendPairKeyed(result, keyNode, copyNode(newV))
			continue
		}
		if isMappingNode(newV) && isMappingNode(oldV) {
			merged, err := deepMergeMapsTo(newV, oldV)
			if err != nil {
				return nil, err
			}
			appendPairKeyed(result, keyNode, merged)
		} else {
			// Keep old value (preserve user setting).
			appendPairKeyed(result, keyNode, copyNode(oldV))
		}
	}

	// Add old-only keys.
	for i := 0; i+1 < len(oldNode.Content); i += 2 {
		keyNode := oldNode.Content[i]
		oldV := oldNode.Content[i+1]
		if keyNode == nil || keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value
		if systemFields[key] {
			continue
		}
		if mappingHasKey(newNode, key) {
			continue
		}
		appendPairKeyed(result, keyNode, copyNode(oldV))
	}

	return result, nil
}

