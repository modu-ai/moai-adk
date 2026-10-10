package backup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

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
// swap it for a bytes.Buffer to assert on the advisory text. The update
// restore path no longer routes through this sink — it uses
// MergeYAML3WayRetained, which collects the keys for the TUI render layer
// instead (t63: the raw per-key stderr append interleaved with the
// cursor-controlled progress redraw on stdout).
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

// retainedKeyAdvisoryLine is the legacy per-key advisory text (REQ-UYP-007).
// Extracted as a helper so the legacy RestoreMoaiConfig wrapper can re-emit
// the identical line for callers not yet routed through the TUI renderer.
func retainedKeyAdvisoryLine(path string) string {
	return fmt.Sprintf("advisory: retained key %q absent from new template (preserved from user config)\n", path)
}

// retainedKeyNotes carries the retained-key advisory state through a 3-way
// merge walk. emit records each retained key path and, when a text sink is
// configured, writes the legacy advisory line to it (the REQ-UYP-007 contract
// the MergeYAML3Way entry keeps for its existing callers and contract tests).
// The production restore path constructs notes WITHOUT a sink
// (MergeYAML3WayRetained) so the keys come back as data and the caller renders
// them through its own output channel — one TUI summary line by default, the
// full list under --verbose (t63 stream-merge fix).
type retainedKeyNotes struct {
	sink io.Writer // optional legacy text sink; nil = collect only
	keys []string  // dotted key paths of retained keys, in walk order
}

// newRetainedKeyNotes builds advisory notes that mirror advisory text to sink
// (pass nil to collect silently).
func newRetainedKeyNotes(sink io.Writer) *retainedKeyNotes {
	return &retainedKeyNotes{sink: sink}
}

// emit records one retained key path, writing the legacy advisory line to the
// configured sink when one is present.
func (n *retainedKeyNotes) emit(path string) {
	n.keys = append(n.keys, path)
	if n.sink != nil {
		_, _ = fmt.Fprint(n.sink, retainedKeyAdvisoryLine(path))
	}
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

// mappingPair returns the key node and the value node paired with the given
// scalar key in a mapping node, or (nil, nil, false) if absent. The key node is
// needed by the comment merge: a comment written above a key is carried on the
// key node, not on the value.
func mappingPair(n *yaml.Node, key string) (*yaml.Node, *yaml.Node, bool) {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil, nil, false
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k := n.Content[i]
		if k != nil && k.Kind == yaml.ScalarNode && k.Value == key {
			return k, n.Content[i+1], true
		}
	}
	return nil, nil, false
}

// mappingGet returns the value node paired with the given scalar key in a
// mapping node, or (nil, false) if absent.
func mappingGet(n *yaml.Node, key string) (*yaml.Node, bool) {
	_, v, ok := mappingPair(n, key)
	return v, ok
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

// nodeComments is the comment trio a yaml.v3 node carries: the block above it,
// the comment on its own line, and the block below it.
type nodeComments struct {
	head, line, foot string
}

// commentsOf returns the comment trio of n, or the zero trio for a nil node — a
// position the template or the base does not have.
func commentsOf(n *yaml.Node) nodeComments {
	if n == nil {
		return nodeComments{}
	}
	return nodeComments{head: n.HeadComment, line: n.LineComment, foot: n.FootComment}
}

// mergeComments sets dst's comment trio to the 3-way merge of the same position
// in the template (newN), the user's file (oldN), and the base (baseN). Each of
// the three slots is merged on its own by mergeCommentBlock.
func mergeComments(dst, newN, oldN, baseN *yaml.Node) {
	n, o, b := commentsOf(newN), commentsOf(oldN), commentsOf(baseN)
	dst.HeadComment = mergeCommentBlock(n.head, o.head, b.head)
	dst.LineComment = mergeCommentBlock(n.line, o.line, b.line)
	dst.FootComment = mergeCommentBlock(n.foot, o.foot, b.foot)
}

// mergedCopy returns a copy of chosen, the node the merge keeps at one position,
// carrying the merged comment trio of that position. The copy leaves the
// caller's nodes untouched: they are shared with the trees being merged.
func mergedCopy(chosen, newN, oldN, baseN *yaml.Node) *yaml.Node {
	c := copyNode(chosen)
	mergeComments(c, newN, oldN, baseN)
	return c
}

// mergeCommentBlock applies the 3-way rule to one comment slot, line by line.
// A block the user left as it stood in the base takes the template's block.
// Once the user has changed it, the user's lines keep their place, a line the
// template removed is dropped, a line the template added is appended, and a
// template line the user deleted stays deleted.
func mergeCommentBlock(newC, oldC, baseC string) string {
	if oldC == baseC {
		return newC
	}
	newLines, oldLines, baseLines := splitCommentLines(newC), splitCommentLines(oldC), splitCommentLines(baseC)
	inNew, inOld, inBase := lineSet(newLines), lineSet(oldLines), lineSet(baseLines)
	var out []string
	for _, line := range oldLines {
		if inNew[line] || !inBase[line] {
			out = append(out, line)
		}
	}
	for _, line := range newLines {
		if !inOld[line] && !inBase[line] {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// splitCommentLines returns the lines of a comment block, or nil for an empty
// block.
func splitCommentLines(block string) []string {
	if block == "" {
		return nil
	}
	return strings.Split(block, "\n")
}

// lineSet returns the set of lines in a comment block.
func lineSet(lines []string) map[string]bool {
	set := make(map[string]bool, len(lines))
	for _, line := range lines {
		set[line] = true
	}
	return set
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

// keptOverTemplateKeys returns the dotted paths of the keys the new template
// carries whose value in the merged document differs from the template's —
// the keys where the merge kept the user's (old) value. System fields are
// skipped: the merge always takes them from the new template (card t1216).
func keptOverTemplateKeys(newData, mergedData []byte) ([]string, error) {
	newRoot, err := decodeDoc(newData, "new")
	if err != nil {
		return nil, err
	}
	mergedRoot, err := decodeDoc(mergedData, "merged")
	if err != nil {
		return nil, err
	}
	var keys []string
	var walk func(n, m *yaml.Node, path string)
	walk = func(n, m *yaml.Node, path string) {
		for i := 0; i+1 < len(n.Content); i += 2 {
			k := n.Content[i]
			if k == nil || k.Kind != yaml.ScalarNode || systemFields[k.Value] {
				continue
			}
			mv, ok := mappingGet(m, k.Value)
			if !ok {
				continue
			}
			nv, p := n.Content[i+1], joinPath(path, k.Value)
			if isMappingNode(nv) && isMappingNode(mv) {
				walk(nv, mv, p)
			} else if !nodeValuesEqual(nv, mv) {
				keys = append(keys, p)
			}
		}
	}
	walk(newRoot, mergedRoot, "")
	return keys, nil
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

// deepMerge3WayTo is the testable core of DeepMerge3Way. It records
// retained-key advisories (keys absent from the new template, REQ-UYP-007) on
// notes. The public DeepMerge3Way passes notes mirroring retainedKeySink
// (os.Stderr in production); tests pass notes over a buffer.
//
// Config roots are mappings, so the value-level decision (old==base → new,
// old!=base → user, no-base → user) is applied inside mergeMappingNode3Way for
// each value. This top-level dispatch only handles the defensive non-mapping
// root (e.g. a scalar-only document): a top-level scalar carries no structure
// to merge, so the user value is preserved when present.
func deepMerge3WayTo(newNode, oldNode, baseNode *yaml.Node, notes *retainedKeyNotes) (*yaml.Node, error) {
	if isMappingNode(newNode) && isMappingNode(oldNode) {
		return mergeMappingNode3Way(newNode, oldNode, baseNode, "", notes)
	}
	// Defensive: non-mapping root. Preserve the user value if present.
	if oldNode != nil && !isNullNode(oldNode) {
		return copyNode(oldNode), nil
	}
	return copyNode(newNode), nil
}

// mergeMappingNode3Way walks the new mapping's keys in source order, resolving
// each against the old (user) and base (previous template) mappings, then
// appends old-only keys at the end (REQ-UYP-006). Key order comes from the
// template. Comments are merged position by position (mergeComments), so a
// comment line the user wrote survives on whichever key, value, or mapping node
// yaml.v3 attached it to. Sequence items are taken whole and carry no merge.
func mergeMappingNode3Way(newNode, oldNode, baseNode *yaml.Node, path string, notes *retainedKeyNotes) (*yaml.Node, error) {
	result := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	// Inherit the template's (newNode's) style so nested flow mappings such as
	// `{ effort: max }` keep their style; the comment trio is merged below.
	if newNode != nil {
		result.Style = newNode.Style
	}
	mergeComments(result, newNode, oldNode, baseNode)

	// First pass: resolve every key present in the new template.
	for i := 0; i+1 < len(newNode.Content); i += 2 {
		keyNode := newNode.Content[i]
		newV := newNode.Content[i+1]
		if keyNode == nil || keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value
		childPath := joinPath(path, key)

		oldKey, oldV, oldExists := mappingPair(oldNode, key)
		if !oldExists {
			// New field added by the template → take it.
			appendPairKeyed(result, keyNode, copyNode(newV))
			continue
		}

		// Both new and old present. The key keeps the comments the user wrote
		// above, beside, or below it, merged over the template's and the base's.
		baseKey, baseV, baseExists := mappingPair(baseNode, key)
		pairKey := mergedCopy(keyNode, keyNode, oldKey, baseKey)

		// System fields always take the new template value (REQ-UYP-010).
		if systemFields[key] {
			appendPairKeyed(result, pairKey, mergedCopy(newV, newV, oldV, baseV))
			continue
		}

		if isMappingNode(newV) && isMappingNode(oldV) {
			// Both mappings → recurse.
			var recBase *yaml.Node
			if baseExists {
				recBase = baseV
			}
			merged, err := mergeMappingNode3Way(newV, oldV, recBase, childPath, notes)
			if err != nil {
				return nil, err
			}
			appendPairKeyed(result, pairKey, merged)
			continue
		}

		// Scalar / sequence value: apply the decision. A user-added value (no
		// base entry) and a user-changed value keep the user's; a value the user
		// left as the base carried takes the template's.
		chosen := oldV
		if baseExists && nodeValuesEqual(oldV, baseV) {
			chosen = newV
		}
		appendPairKeyed(result, pairKey, mergedCopy(chosen, newV, oldV, baseV))
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
		notes.emit(childPath)
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
// 2-way path (REQ-UYP-007 binds only DeepMerge3Way). Comments follow the old
// file per slot (oldWinsComments): the user's block wins, and the template's
// block fills only a slot the user left empty. With no base, a user's comment
// cannot be told apart from a template comment, so the user's block always wins.
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
		result.Style = newNode.Style
	}
	oldWinsComments(result, newNode, oldNode)

	// Copy all new values in source order.
	for i := 0; i+1 < len(newNode.Content); i += 2 {
		keyNode := newNode.Content[i]
		newV := newNode.Content[i+1]
		if keyNode == nil || keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value
		oldKey, oldV, oldExists := mappingPair(oldNode, key)
		if !oldExists {
			appendPairKeyed(result, keyNode, copyNode(newV))
			continue
		}
		pairKey := oldWinsCopy(keyNode, keyNode, oldKey)
		// System fields always use the new value; the user's comments still win.
		if systemFields[key] {
			appendPairKeyed(result, pairKey, oldWinsCopy(newV, newV, oldV))
			continue
		}
		if isMappingNode(newV) && isMappingNode(oldV) {
			merged, err := deepMergeMapsTo(newV, oldV)
			if err != nil {
				return nil, err
			}
			appendPairKeyed(result, pairKey, merged)
		} else {
			// Keep old value (preserve user setting), with the user's comments.
			appendPairKeyed(result, pairKey, oldWinsCopy(oldV, newV, oldV))
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

// oldWinsComments sets dst's comment trio on the 2-way path: per slot, the
// user's block (oldN) wins, and the template's block (newN) fills a slot the
// user left empty. It writes dst in place, so pass a copy when dst is a node
// shared with a tree being merged (see oldWinsCopy).
func oldWinsComments(dst, newN, oldN *yaml.Node) {
	n, o := commentsOf(newN), commentsOf(oldN)
	dst.HeadComment = oldBlockOr(o.head, n.head)
	dst.LineComment = oldBlockOr(o.line, n.line)
	dst.FootComment = oldBlockOr(o.foot, n.foot)
}

// oldWinsCopy returns a copy of chosen, the node the 2-way merge keeps at one
// position, carrying the old-wins comment trio of that position. The caller's
// nodes stay untouched.
func oldWinsCopy(chosen, newN, oldN *yaml.Node) *yaml.Node {
	c := copyNode(chosen)
	oldWinsComments(c, newN, oldN)
	return c
}

// oldBlockOr returns the user's comment block, or the template's when the user
// left the slot empty.
func oldBlockOr(oldC, newC string) string {
	if oldC != "" {
		return oldC
	}
	return newC
}
