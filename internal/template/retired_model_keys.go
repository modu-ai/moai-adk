package template

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// retired_model_keys.go — write-time removal of retired configuration keys.
// It holds the moai update strip step for the per-agent model/effort keys,
// which subagents no longer read because they inherit the main session's model
// and effort (SPEC-AGENT-MODEL-INHERIT-001). The strip is line/indent based
// rather than a YAML round-trip so every line it does not remove survives
// byte-identical, comments included. The former stripRetiredLLMKeys (plan_type
// + claude_models removal on the llm.profile write path) left with
// ApplyProfile in M5: its only caller.

// leadingWS returns the count of leading space/tab characters in a line.
func leadingWS(line string) int {
	n := 0
	for _, r := range line {
		if r == ' ' || r == '\t' {
			n++
			continue
		}
		break
	}
	return n
}

// RetiredModelKey names one key the moai update strip step removed.
type RetiredModelKey struct {
	// Section is the file name under .moai/config/sections/ (e.g. "llm.yaml").
	Section string
	// Key is the dotted key path including the document root (e.g.
	// "llm.agent_overrides"), the same form the retained-key advisory uses.
	Key string
	// UserValues marks a key whose removal dropped values the user entered
	// (an agent_overrides map with entries); the backup still holds them.
	UserValues bool
}

// retiredModelKeySets lists, per section file, the document root and the
// root-level child keys that assigned a per-agent model or effort. The order
// is the report order.
var retiredModelKeySets = []struct {
	section string
	root    string
	keys    []string
}{
	{"llm.yaml", "llm", []string{"profile", "performance_tier", "profiles", "harness_agents", "agent_overrides"}},
	{"workflow.yaml", "workflow", []string{"agent_model_guard", "workflow_agents", "model_routing", "model_routing_profiles"}},
}

// userValueKey is the one retired key without a shipped default to compare
// against: any entry in it was typed by the user.
const userValueKey = "agent_overrides"

// keysToStrip returns the retired keys of set that are not in shipped (the
// dotted keys the current template still carries).
func keysToStrip(root string, keys []string, shipped map[string]bool) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if !shipped[root+"."+k] {
			out = append(out, k)
		}
	}
	return out
}

// ShippedRetiredModelKeys returns the retired per-agent model/effort keys the
// embedded template still ships, as dotted keys ("llm.profile"). The update
// strip step leaves those alone: while the template ships a key, code in the
// same build still reads it, and stripping it would only make every update
// redeploy and re-strip it. A key leaves this set when it leaves the template,
// and the strip step then removes it from user files with no further change.
func ShippedRetiredModelKeys() (map[string]bool, error) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		return nil, fmt.Errorf("load embedded templates: %w", err)
	}
	shipped := map[string]bool{}
	for _, set := range retiredModelKeySets {
		content, err := fs.ReadFile(fsys, path.Join(".moai", "config", "sections", set.section))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("read embedded %s: %w", set.section, err)
		}
		_, found := stripRootChildKeys(content, set.root, set.keys)
		for _, r := range found {
			shipped[set.root+"."+r.key] = true
		}
	}
	return shipped, nil
}

// RetiredModelKeysPresent reports, without writing anything, whether the
// project's llm.yaml or workflow.yaml still carries a retired per-agent
// model/effort key outside shipped. Callers use it to take a backup only when
// a strip will actually change a file.
func RetiredModelKeysPresent(projectRoot string, shipped map[string]bool) (bool, error) {
	for _, set := range retiredModelKeySets {
		content, err := readSectionFile(projectRoot, set.section)
		if err != nil {
			return false, err
		}
		if content == nil {
			continue
		}
		if _, removed := stripRootChildKeys(content, set.root, keysToStrip(set.root, set.keys, shipped)); len(removed) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// StripRetiredModelKeys removes the retired per-agent model/effort keys, except
// those in shipped (see ShippedRetiredModelKeys), from the project's llm.yaml
// and workflow.yaml, together with the comment block directly above each key,
// and returns what it removed in report order. A file without any such key is
// left untouched (not rewritten). The caller takes the configuration backup
// first; this function never does.
//
// @MX:ANCHOR: [AUTO] StripRetiredModelKeys — the moai update strip step for retired per-agent model/effort keys
// @MX:REASON: three update hosts call it (template sync restore, version-matched skip, clean reinstall); each must run it after a backup exists
func StripRetiredModelKeys(projectRoot string, shipped map[string]bool) ([]RetiredModelKey, error) {
	var all []RetiredModelKey
	for _, set := range retiredModelKeySets {
		content, err := readSectionFile(projectRoot, set.section)
		if err != nil {
			return all, err
		}
		if content == nil {
			continue
		}
		out, removed := stripRootChildKeys(content, set.root, keysToStrip(set.root, set.keys, shipped))
		if len(removed) == 0 {
			continue
		}
		p := sectionPath(projectRoot, set.section)
		info, err := os.Stat(p)
		if err != nil {
			return all, fmt.Errorf("stat %s: %w", set.section, err)
		}
		if err := os.WriteFile(p, out, info.Mode().Perm()); err != nil {
			return all, fmt.Errorf("write %s: %w", set.section, err)
		}
		for _, r := range removed {
			all = append(all, RetiredModelKey{
				Section:    set.section,
				Key:        set.root + "." + r.key,
				UserValues: r.key == userValueKey && r.hasValues,
			})
		}
	}
	return all, nil
}

func sectionPath(projectRoot, section string) string {
	return filepath.Join(projectRoot, ".moai", "config", "sections", section)
}

// readSectionFile returns the file content, or nil with no error when the file
// does not exist.
func readSectionFile(projectRoot, section string) ([]byte, error) {
	b, err := os.ReadFile(sectionPath(projectRoot, section))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", section, err)
	}
	return b, nil
}

type removedChildKey struct {
	key       string
	hasValues bool
}

// isCommentOrBlank reports whether a line carries no YAML content.
func isCommentOrBlank(line string) bool {
	t := strings.TrimSpace(line)
	return t == "" || strings.HasPrefix(t, "#")
}

// flowDepth returns depth adjusted by the flow-collection brackets ({ } [ ])
// on line, ignoring brackets inside quoted scalars and after a comment.
func flowDepth(line string, depth int) int {
	var quote rune
	prev := ' '
	for _, r := range line {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '#' && (prev == ' ' || prev == '\t'):
			return depth
		case r == '{' || r == '[':
			depth++
		case r == '}' || r == ']':
			depth--
		}
		prev = r
	}
	return depth
}

// childKeyName returns the mapping key of a `key:` line, or "" when the line
// is not a plain mapping entry.
func childKeyName(line string) string {
	t := strings.TrimSpace(line)
	i := strings.Index(t, ":")
	if i <= 0 || strings.ContainsAny(t[:i], " \t\"'#") {
		return ""
	}
	return t[:i]
}

// stripRootChildKeys removes, from the block of the zero-indent `root:` key,
// every direct child whose name is in keys: the key line, its more-indented
// body, and the comment lines directly above it at the same indent. Lines are
// compared with any trailing "\r" ignored and written back unchanged, so CRLF
// files keep their line endings. It returns the new content and the removed
// keys in the order of keys.
func stripRootChildKeys(content []byte, root string, keys []string) ([]byte, []removedChildKey) {
	lines := strings.Split(string(content), "\n")
	bare := make([]string, len(lines))
	for i, l := range lines {
		bare[i] = strings.TrimRight(l, "\r")
	}

	rootIdx := -1
	for i, l := range bare {
		if leadingWS(l) == 0 && childKeyName(l) == root {
			rootIdx = i
			break
		}
	}
	if rootIdx < 0 {
		return content, nil
	}

	want := map[string]bool{}
	for _, k := range keys {
		want[k] = true
	}
	drop := make([]bool, len(lines))
	found := map[string]removedChildKey{}
	childIndent := -1

	for i := rootIdx + 1; i < len(bare); i++ {
		l := bare[i]
		if isCommentOrBlank(l) {
			continue
		}
		ind := leadingWS(l)
		if ind == 0 {
			break // end of the root block
		}
		if childIndent < 0 {
			childIndent = ind
		}
		if ind != childIndent {
			continue
		}
		name := childKeyName(l)
		if !want[name] {
			continue
		}

		// The key line and its more-indented body; trailing blank lines are
		// handed back so the separation before the next key survives. A flow
		// collection left open on the key line (a YAML encoder writes
		// `key: {a: b` … `}`) continues until it closes, whatever the indent
		// of its closing line.
		depth := flowDepth(l, 0)
		end := i + 1
		last := i
		hasValues := false
		for ; end < len(bare); end++ {
			b := bare[end]
			if strings.TrimSpace(b) == "" {
				continue
			}
			if depth <= 0 && leadingWS(b) <= childIndent {
				break
			}
			depth = flowDepth(b, depth)
			last = end
			if !isCommentOrBlank(b) {
				hasValues = true
			}
		}
		if !hasValues {
			v := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), name+":"))
			if c := strings.Index(v, " #"); c >= 0 {
				v = strings.TrimSpace(v[:c])
			}
			switch v {
			case "", "{}", "[]", "null", "~", `""`, "''":
			default:
				hasValues = true
			}
		}
		start := i
		for start-1 > rootIdx && strings.HasPrefix(strings.TrimSpace(bare[start-1]), "#") && leadingWS(bare[start-1]) == childIndent {
			start--
		}
		for k := start; k <= last; k++ {
			drop[k] = true
		}
		// Removing a block that sat between two blank lines would leave them
		// adjacent; drop the one after the block.
		if start-1 >= 0 && strings.TrimSpace(bare[start-1]) == "" && last+1 < len(bare) && strings.TrimSpace(bare[last+1]) == "" {
			drop[last+1] = true
		}
		found[name] = removedChildKey{key: name, hasValues: hasValues}
		i = last
	}

	if len(found) == 0 {
		return content, nil
	}
	out := make([]string, 0, len(lines))
	for i, l := range lines {
		if !drop[i] {
			out = append(out, l)
		}
	}
	var removed []removedChildKey
	for _, k := range keys {
		if r, ok := found[k]; ok {
			removed = append(removed, r)
		}
	}
	return []byte(strings.Join(out, "\n")), removed
}
