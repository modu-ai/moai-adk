package config

// llm_save.go — the llm section's write path (U-05, REQ-RC30-007 of
// SPEC-UPDATE-REPAIR-RC30-001). A Save of a dirty llm section re-marshaled the
// typed LLMConfig over llm.yaml, which dropped deployment_mode, the unmodelled
// keys, and the comments. A dirty section over an existing file is now
// patched scalar by scalar through yamlpatch, so only the values the caller
// changed are rewritten.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// saveLLM writes the llm section. A dirty section over an existing llm.yaml is
// patched in place: only the scalars that differ from the baseline are
// rewritten. An absent file, or a change a scalar patch cannot express (a
// removed key, a sequence, or a value that must be quoted), takes the typed
// rewrite of saveGated, as it did before this path existed.
func (m *ConfigManager) saveLLM(sectionsDir string) error {
	path := filepath.Join(sectionsDir, "llm.yaml")
	if _, statErr := os.Stat(path); m.llmDirty && statErr == nil {
		edits, patchable, err := llmScalarEdits(m.llmBaseline, m.config.LLM)
		if err != nil {
			return err
		}
		if patchable {
			return yamlpatch.PatchFile(path, edits)
		}
	}
	return m.saveGated(sectionsDir, "llm.yaml", m.llmDirty, llmFileWrapper{LLM: m.config.LLM})
}

// llmBaselineOf renders the llm section as the loaded or last saved state.
// The bytes are a snapshot, so a map the caller still holds and mutates in
// place cannot alias into the baseline. A nil result means the section could
// not be rendered, and llmScalarEdits reports that as an error.
func llmBaselineOf(cfg LLMConfig) []byte {
	data, err := yaml.Marshal(llmFileWrapper{LLM: cfg})
	if err != nil {
		return nil
	}
	return data
}

// llmScalarEdits returns the scalar replacements that turn the baseline llm
// section into cur, with paths rooted at the document, as yamlpatch takes
// them. patchable is false when the change is not a set of scalar
// replacements.
func llmScalarEdits(baseline []byte, cur LLMConfig) ([]yamlpatch.KeyEdit, bool, error) {
	base, err := llmMappingOf(baseline)
	if err != nil {
		return nil, false, fmt.Errorf("llm baseline: %w", err)
	}
	now, err := llmMappingOf(llmBaselineOf(cur))
	if err != nil {
		return nil, false, fmt.Errorf("llm section: %w", err)
	}
	var edits []yamlpatch.KeyEdit
	if !diffScalarNodes([]string{"llm"}, base, now, &edits) {
		return nil, false, nil
	}
	return edits, true, nil
}

// llmMappingOf returns the llm mapping of a document rendered from an
// llmFileWrapper.
func llmMappingOf(data []byte) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("document has no top-level mapping")
	}
	section := mappingValue(doc.Content[0], "llm")
	if section == nil || section.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("document has no llm mapping")
	}
	return section, nil
}

// diffScalarNodes appends the scalar replacements that turn base into now at
// path. It reports false for a change no scalar replacement can express: a
// removed key, a container that changed kind, a sequence that changed, or a
// scalar that must be quoted (the patch carries no style for a new value).
func diffScalarNodes(path []string, base, now *yaml.Node, edits *[]yamlpatch.KeyEdit) bool {
	if base.Kind != now.Kind {
		return false
	}
	switch now.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(base.Content); i += 2 {
			if mappingValue(now, base.Content[i].Value) == nil {
				return false
			}
		}
		for i := 0; i+1 < len(now.Content); i += 2 {
			key := now.Content[i].Value
			child := appendPath(path, key)
			if old := mappingValue(base, key); old != nil {
				if !diffScalarNodes(child, old, now.Content[i+1], edits) {
					return false
				}
			} else if !appendScalarLeaves(child, now.Content[i+1], edits) {
				return false
			}
		}
		return true
	case yaml.ScalarNode:
		if base.Value == now.Value {
			return true
		}
		if now.Style != 0 {
			return false
		}
		*edits = append(*edits, yamlpatch.KeyEdit{Path: path, Value: now.Value})
		return true
	default:
		return nodeEqual(base, now)
	}
}

// appendScalarLeaves appends an upsert for every scalar under n. It is the
// form a key absent from the baseline takes, and it reports false when n holds
// a node a scalar replacement cannot express.
func appendScalarLeaves(path []string, n *yaml.Node, edits *[]yamlpatch.KeyEdit) bool {
	switch n.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			if !appendScalarLeaves(appendPath(path, n.Content[i].Value), n.Content[i+1], edits) {
				return false
			}
		}
		return true
	case yaml.ScalarNode:
		if n.Style != 0 {
			return false
		}
		*edits = append(*edits, yamlpatch.KeyEdit{Path: path, Value: n.Value})
		return true
	default:
		return false
	}
}

// mappingValue returns the value node stored under key in a mapping node, or
// nil when the key is absent.
func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// appendPath returns a new path extended by key, so sibling edits never share
// a backing array.
func appendPath(path []string, key string) []string {
	out := make([]string, len(path)+1)
	copy(out, path)
	out[len(path)] = key
	return out
}

// nodeEqual reports whether two nodes render to the same YAML. It is the
// equality check for containers the scalar patch does not model.
func nodeEqual(a, b *yaml.Node) bool {
	left, errA := yaml.Marshal(a)
	right, errB := yaml.Marshal(b)
	return errA == nil && errB == nil && bytes.Equal(left, right)
}
