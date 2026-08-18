package backup

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// MergeYAML3Way performs a 3-way merge of YAML documents.
//
// It uses baseData (old template defaults) to detect user changes:
//   - If user value == base value: user didn't change it → use new template value
//   - If user value != base value: user customized it → preserve user value
//
// System fields (version, template_version) always use new values regardless.
//
// The merge decodes each document into a *yaml.Node tree so that comments,
// key order, scalar quoting style, and anchor/alias nodes survive the round
// trip — the fix for issue #1243, where a map[string]any round trip stripped
// every comment and alphabetized keys. Old-only keys (absent from the new
// template) are retained and reported on stderr per REQ-UYP-006/007.
//
// The wrapper signature ([]byte, ...) ([]byte, error) is unchanged so the
// production call site in restore.go needs no edit (Decision D2 mitigation).
func MergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error) {
	merged, _, err := mergeYAML3WayNotes(newData, oldData, baseData, retainedKeySink)
	return merged, err
}

// MergeYAML3WayRetained performs the same 3-way merge as MergeYAML3Way but
// COLLECTS the retained-key paths instead of writing the legacy per-key
// advisory text to the sink. The update restore path uses this entry (t63)
// so the advisory renders through the caller's output channel — one TUI
// summary line by default, the full key list under --verbose — instead of
// appending raw lines to stderr while the update progress line is mid-redraw
// on stdout. The merge semantics are identical to MergeYAML3Way; only the
// advisory's output routing differs.
func MergeYAML3WayRetained(newData, oldData, baseData []byte) (merged []byte, retainedKeys []string, err error) {
	return mergeYAML3WayNotes(newData, oldData, baseData, nil)
}

// mergeYAML3WayNotes is the shared decode-walk-encode core of the two
// 3-way merge entries. textSink configures the legacy advisory text mirror:
// non-nil keeps the REQ-UYP-007 advisory-text-on-sink contract (MergeYAML3Way);
// nil collects silently (MergeYAML3WayRetained).
func mergeYAML3WayNotes(newData, oldData, baseData []byte, textSink io.Writer) ([]byte, []string, error) {
	newRoot, err := decodeDoc(newData, "new")
	if err != nil {
		return nil, nil, err
	}
	oldRoot, err := decodeDoc(oldData, "old")
	if err != nil {
		return nil, nil, err
	}
	baseRoot, err := decodeDoc(baseData, "base")
	if err != nil {
		return nil, nil, err
	}
	notes := newRetainedKeyNotes(textSink)
	merged, err := deepMerge3WayTo(newRoot, oldRoot, baseRoot, notes)
	if err != nil {
		return nil, nil, err
	}
	out, err := encodeNode(merged)
	if err != nil {
		return nil, nil, err
	}
	return out, notes.keys, nil
}

// DeepMerge3Way recursively performs a 3-way merge of node trees.
//
// Decision logic for each key:
//   - old == base → user didn't change → use new value
//   - old != base → user changed → preserve old value
//   - key only in new → new field added by template → use new value
//   - key only in old → absent from new template → retain + report (REQ-UYP-006/007)
//
// The signature is node-typed (Decision D2): the map[string]any representation
// has nowhere to store comments or key order, so format preservation is only
// possible when the merge operates on the node tree directly. Retained-key
// advisories are recorded on advisory notes mirroring retainedKeySink
// (os.Stderr in production).
func DeepMerge3Way(newNode, oldNode, baseNode *yaml.Node) (*yaml.Node, error) {
	return deepMerge3WayTo(newNode, oldNode, baseNode, newRetainedKeyNotes(retainedKeySink))
}

// ValuesEqual compares two interface{} values for equality.
// Handles string, int, float, bool, and nil comparisons.
//
// Retained (Decision D10) even though it has no production caller after the
// node-tree rewrite: it remains the equality primitive over the any (map /
// scalar) domain and is reused by nodeValuesEqual for node-level comparison.
// Removing an exported symbol is a separate breaking-API decision this SPEC
// does not need to make.
func ValuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// MergeYAMLDeep performs a deep merge of two YAML documents (2-way fallback).
//
// The newData takes precedence for structure, but values from oldData are
// preserved when the key exists in both. System fields always use the new
// value. The merge decodes into node trees so comments, key order, and scalar
// quoting survive (same mechanism as MergeYAML3Way). The wrapper signature is
// unchanged so restore.go needs no edit.
func MergeYAMLDeep(newData, oldData []byte) ([]byte, error) {
	newRoot, err := decodeDoc(newData, "new")
	if err != nil {
		return nil, err
	}
	oldRoot, err := decodeDoc(oldData, "old")
	if err != nil {
		return nil, err
	}
	merged, err := DeepMergeMaps(newRoot, oldRoot)
	if err != nil {
		return nil, err
	}
	return encodeNode(merged)
}

// DeepMergeMaps recursively merges oldNode into newNode, preserving old values.
// System fields (version, template_version) always use new values. The
// signature is node-typed per Decision D2. No advisory is emitted on the 2-way
// path (REQ-UYP-007 binds only DeepMerge3Way); old-only keys are preserved
// silently so the 3-way path is never more destructive than this fallback
// (REQ-UYP-008).
func DeepMergeMaps(newNode, oldNode *yaml.Node) (*yaml.Node, error) {
	return deepMergeMapsTo(newNode, oldNode)
}
