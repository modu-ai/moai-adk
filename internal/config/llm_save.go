package config

// llm_save.go — the llm section's write path (U-05, REQ-RC30-007 of
// SPEC-UPDATE-REPAIR-RC30-001). A Save of a dirty llm section re-marshaled the
// typed LLMConfig over llm.yaml, which dropped deployment_mode, the unmodelled
// keys, and the comments. A dirty section over an existing file is now
// patched through yamlpatch: only the keys the caller changed are rewritten,
// and the keys and comments the section does not own survive.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// saveLLM writes the llm section. A dirty section over an existing llm.yaml is
// patched through yamlpatch: each changed scalar is replaced, a value that must
// be quoted is written quoted, and a removed key is deleted, so the keys and
// comments the section does not own survive (REQ-RC30-007). An absent file takes
// the rooted typed write, which has no content to keep, and an existing section
// that is not dirty keeps saveGated's no-op.
func (m *ConfigManager) saveLLM(sectionsDir string) error {
	path := filepath.Join(sectionsDir, "llm.yaml")
	if info, statErr := os.Stat(path); m.llmDirty && statErr == nil {
		base, err := llmMappingOf(m.llmBaseline)
		if err != nil {
			return fmt.Errorf("llm baseline: %w", err)
		}
		edits, err := llmSectionEditsFrom(base, m.config.LLM)
		if err != nil {
			return err
		}
		// The file keeps its mode: the current one is passed through. Unlike
		// PatchFile, PatchFileWithMode also accepts a blank llm.yaml.
		return yamlpatch.PatchFileAtRoot(m.root, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, "llm.yaml"), edits,
			yamlpatch.RootOptions{Mode: info.Mode().Perm(), AllowEmpty: true})
	}
	return m.saveGated(sectionsDir, "llm.yaml", m.llmDirty, llmFileWrapper{LLM: m.config.LLM})
}

// saveSectionAtRoot preserves typed section serialization without using a
// path-based parent creation or publishing operation outside the project root.
func (m *ConfigManager) saveSectionAtRoot(filename string, data any) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filename, err)
	}
	return yamlpatch.WriteFileAtRoot(m.root, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, filename), yamlData,
		yamlpatch.RootOptions{ParentPerm: defs.DirPerm})
}

// LLMSectionEdits returns the yamlpatch edits that turn the llm section held in
// persisted into desired, with paths rooted at the document. A nil persisted
// section is a file that does not exist yet, so every scalar of desired becomes
// an edit. A change no edit can express is an error naming the field; a caller
// returns it rather than falling back to a typed rewrite.
func LLMSectionEdits(persisted *LLMConfig, desired LLMConfig) ([]yamlpatch.KeyEdit, error) {
	base := &yaml.Node{Kind: yaml.MappingNode}
	if persisted != nil {
		var err error
		if base, err = llmMappingOf(llmBaselineOf(*persisted)); err != nil {
			return nil, fmt.Errorf("llm baseline: %w", err)
		}
	}
	return llmSectionEditsFrom(base, desired)
}

// llmBaselineOf renders the llm section as the loaded or last saved state.
// The bytes are a snapshot, so a map the caller still holds and mutates in
// place cannot alias into the baseline. A nil result means the section could
// not be rendered, and llmMappingOf reports that as an error.
func llmBaselineOf(cfg LLMConfig) []byte {
	data, err := yaml.Marshal(llmFileWrapper{LLM: cfg})
	if err != nil {
		return nil
	}
	return data
}

// llmSectionEditsFrom returns the yamlpatch edits that turn the base llm mapping
// into cur, with paths rooted at the document. The base is the mapping the
// caller loaded or last saved, so a key the typed struct does not model is never
// part of the diff and stays in the file untouched.
func llmSectionEditsFrom(base *yaml.Node, cur LLMConfig) ([]yamlpatch.KeyEdit, error) {
	now, err := llmMappingOf(llmBaselineOf(cur))
	if err != nil {
		return nil, fmt.Errorf("llm section: %w", err)
	}
	var edits []yamlpatch.KeyEdit
	if err := diffLLMNodes([]string{"llm"}, base, now, &edits); err != nil {
		return nil, err
	}
	return edits, nil
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

// diffLLMNodes appends the yamlpatch edits that turn base into now at path. A
// key base has and now lacks is removed. A scalar that changed is replaced,
// double-quoted when now is quoted. A key only now has is upserted scalar by
// scalar. A change no edit can express (a container that changed kind, a
// sequence that changed, or a literal or folded block scalar) is an error that
// names the path.
func diffLLMNodes(path []string, base, now *yaml.Node, edits *[]yamlpatch.KeyEdit) error {
	if base.Kind != now.Kind {
		return unexpressedLLMChange(path)
	}
	switch now.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(base.Content); i += 2 {
			if mappingValue(now, base.Content[i].Value) == nil {
				*edits = append(*edits, yamlpatch.KeyEdit{Path: appendPath(path, base.Content[i].Value), Remove: true})
			}
		}
		for i := 0; i+1 < len(now.Content); i += 2 {
			key := now.Content[i].Value
			child := appendPath(path, key)
			if old := mappingValue(base, key); old != nil {
				if err := diffLLMNodes(child, old, now.Content[i+1], edits); err != nil {
					return err
				}
			} else if err := appendLLMLeaves(child, now.Content[i+1], edits); err != nil {
				return err
			}
		}
		return nil
	case yaml.ScalarNode:
		if base.Value == now.Value {
			return nil
		}
		quoted, err := llmQuoting(path, now)
		if err != nil {
			return err
		}
		*edits = append(*edits, yamlpatch.KeyEdit{Path: path, Value: now.Value, Quoted: quoted})
		return nil
	default:
		if nodeEqual(base, now) {
			return nil
		}
		return unexpressedLLMChange(path)
	}
}

// appendLLMLeaves appends an upsert for every scalar under n. It is the form a
// key absent from the baseline takes, and it reports an error for a node a
// scalar upsert cannot express.
func appendLLMLeaves(path []string, n *yaml.Node, edits *[]yamlpatch.KeyEdit) error {
	switch n.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			if err := appendLLMLeaves(appendPath(path, n.Content[i].Value), n.Content[i+1], edits); err != nil {
				return err
			}
		}
		return nil
	case yaml.ScalarNode:
		quoted, err := llmQuoting(path, n)
		if err != nil {
			return err
		}
		*edits = append(*edits, yamlpatch.KeyEdit{Path: path, Value: n.Value, Quoted: quoted})
		return nil
	default:
		return unexpressedLLMChange(path)
	}
}

// llmQuoting reports whether a scalar must be written double-quoted. yaml.Marshal
// quotes exactly the values that plain style would read back as another type,
// so the node's style is the answer. A block scalar has no yamlpatch form.
func llmQuoting(path []string, n *yaml.Node) (bool, error) {
	switch n.Style {
	case 0:
		return false, nil
	case yaml.DoubleQuotedStyle, yaml.SingleQuotedStyle:
		return true, nil
	default:
		return false, unexpressedLLMChange(path)
	}
}

// unexpressedLLMChange names a change the yamlpatch seam cannot write.
func unexpressedLLMChange(path []string) error {
	return fmt.Errorf("%s: change cannot be written through the yamlpatch seam", strings.Join(path, "."))
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
