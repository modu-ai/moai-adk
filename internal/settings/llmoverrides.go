package settings

// llmoverrides.go — the neutral write seam for the restored console's
// agent-overrides surface (SPEC-WEB-AGENTFM-RESTORE-001 M3): llm.profile and
// llm.agent_overrides in .moai/config/sections/llm.yaml.
//
// The two keys are deliberately NOT generic schema fields (plan §B-6 — the
// same "schema-external live keys" disposition the pre-deletion surface used),
// so they persist through dedicated entry points here instead of
// ApplySchemaEdits:
//
//   - WriteLLMProfile: a scalar splice through the shared WriteSectionViaSeam
//     (byte-precise on the existing key the re-shipped template carries),
//     gated on the current disk value so an equal submission writes nothing
//     (the REQ-WSL-001 no-op lineage, mtime included).
//   - WriteLLMAgentOverrides: a line/indent block splice. yamlpatch cannot
//     express this write — it writes scalars only and supports no deletion
//     (yamlpatch.go package header), while the override map needs entry
//     upserts AND clears. The splice rewrites exactly the `agent_overrides:`
//     block (its key line plus more-indented body) and preserves every other
//     byte of the file — the lossless contract
//     (SPEC-WEB-SAVE-LOSSLESS-001 lineage) applied at block granularity.
//     The old SetSection("llm")→Save() full re-marshal is NOT revived: it was
//     the GitHub issue #1731 defect mechanism (comments and unmodeled keys
//     destroyed).
//
// Both writers are web/TUI-neutral: the console calls them through the app
// seams; nothing here imports internal/web.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// WriteLLMProfile persists the active profile column to llm.profile
// (REQ-AFR-003). An equal persisted value short-circuits — no write, mtime
// untouched.
func WriteLLMProfile(projectRoot, profile string) error {
	path := []string{"llm", "profile"}
	if disk, ok := readSeamScalar(projectRoot, "llm", path); ok && disk == profile {
		return nil
	}
	return WriteSectionViaSeam(projectRoot, "llm", []yamlpatch.KeyEdit{
		{Path: path, Value: profile},
	})
}

// WriteLLMAgentOverrides persists the FULL desired override map to
// llm.agent_overrides (REQ-AFR-004): the caller resolves pins/clears against
// the current state first and passes the final map; an entry absent from the
// map is absent from the block after the write. When the spliced output is
// byte-identical to the file the write is skipped entirely.
func WriteLLMAgentOverrides(projectRoot string, overrides map[string]config.ModelEffort) error {
	path := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("settings: read llm.yaml: %w", err)
		}
		// Greenfield tolerance (the seam's absent-file contract): a project
		// with no llm.yaml yet gets a minimal root carrying the block.
		data = []byte("llm:\n")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("settings: create sections directory: %w", err)
		}
	}

	out, err := spliceAgentOverridesBlock(string(data), overrides)
	if err != nil {
		return fmt.Errorf("settings: llm.agent_overrides splice: %w", err)
	}
	if out == string(data) {
		return nil // byte-identical — no write, mtime untouched
	}
	return atomicWriteSection(path, []byte(out))
}

// atomicWriteSection writes via temp file + rename in the target directory,
// preserving the original file mode (the same durability shape the strip step
// and yamlpatch's atomic write use).
func atomicWriteSection(path string, data []byte) error {
	info, err := os.Stat(path)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".llmoverrides-*.tmp")
	if err != nil {
		return fmt.Errorf("settings: create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("settings: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("settings: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("settings: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("settings: rename: %w", err)
	}
	return nil
}

// spliceAgentOverridesBlock rewrites the `agent_overrides:` block inside an
// llm.yaml document and returns the new full content. Everything outside the
// block — comments, key order, unknown keys, blank lines, CRLF line endings —
// survives byte-identical. The byte-identity of the RESULT is the caller's
// no-op gate.
func spliceAgentOverridesBlock(doc string, overrides map[string]config.ModelEffort) (string, error) {
	lines := strings.Split(doc, "\n")

	// Locate the zero-indent `llm:` root.
	rootIdx := -1
	for i, l := range lines {
		if leadingWSCount(l) == 0 && childKeyNameOf(l) == "llm" {
			rootIdx = i
			break
		}
	}
	if rootIdx < 0 {
		return "", fmt.Errorf("no zero-indent llm: root found")
	}

	// Child indent: the first indented content line after the root (the
	// document's own step; the template ships 2, hand-edited files may use 4).
	childIndent := -1
	for i := rootIdx + 1; i < len(lines); i++ {
		l := lines[i]
		if isBlankOrComment(l) {
			continue
		}
		ind := leadingWSCount(l)
		if ind == 0 {
			break
		}
		childIndent = ind
		break
	}
	if childIndent <= 0 {
		childIndent = 2
	}

	// Locate the existing agent_overrides key at child indent and the extent
	// of its body (key line + every more-indented line, flow brackets
	// tracked so a multi-line flow mapping stays inside the region).
	keyIdx := -1
	lastIdx := -1
	depth := 0
	for i := rootIdx + 1; i < len(lines); i++ {
		l := lines[i]
		ind := leadingWSCount(l)
		if ind == 0 {
			break // end of the llm block
		}
		if keyIdx < 0 {
			if ind == childIndent && childKeyNameOf(l) == "agent_overrides" {
				keyIdx = i
				lastIdx = i
				depth = flowDepthOf(l, 0)
			}
			continue
		}
		// Inside the block body.
		if depth <= 0 && ind <= childIndent {
			break
		}
		depth = flowDepthOf(l, depth)
		lastIdx = i
	}

	block := renderAgentOverridesBlock(childIndent, overrides)

	var out []string
	if keyIdx < 0 {
		// Absent key: insert the block directly under the llm: root, ahead of
		// the existing children (order inside the block is the console's; the
		// key's absence means no user arrangement to preserve).
		out = append(out, lines[:rootIdx+1]...)
		out = append(out, block...)
		out = append(out, lines[rootIdx+1:]...)
	} else {
		out = append(out, lines[:keyIdx]...)
		out = append(out, block...)
		out = append(out, lines[lastIdx+1:]...)
	}
	return strings.Join(out, "\n"), nil
}

// renderAgentOverridesBlock renders the block at the given child indent: the
// key line carrying `{}` when the map is empty, else one sub-mapping per
// agent in sorted-name order (deterministic output — map iteration must not
// reach the file).
func renderAgentOverridesBlock(childIndent int, overrides map[string]config.ModelEffort) []string {
	ind := strings.Repeat(" ", childIndent)
	names := make([]string, 0, len(overrides))
	for name := range overrides {
		names = append(names, name)
	}
	sort.Strings(names)

	var lines []string
	if len(names) == 0 {
		return []string{ind + "agent_overrides: {}"}
	}
	lines = append(lines, ind+"agent_overrides:")
	for _, name := range names {
		me := overrides[name]
		lines = append(lines,
			ind+"    "+name+":",
			ind+"        model: "+me.Model,
			ind+"        effort: "+me.Effort,
		)
	}
	return lines
}

// leadingWSCount returns the count of leading space/tab characters in a line.
func leadingWSCount(line string) int {
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

// isBlankOrComment reports whether a line carries no YAML content.
func isBlankOrComment(line string) bool {
	t := strings.TrimSpace(line)
	return t == "" || strings.HasPrefix(t, "#")
}

// childKeyNameOf returns the mapping key of a `key:` line, or "" when the
// line is not a plain mapping entry.
func childKeyNameOf(line string) string {
	t := strings.TrimSpace(line)
	i := strings.Index(t, ":")
	if i <= 0 || strings.ContainsAny(t[:i], " \t\"'#") {
		return ""
	}
	return t[:i]
}

// flowDepthOf returns depth adjusted by the flow-collection brackets on line,
// ignoring brackets inside quoted scalars and after a comment (the strip
// step's rule, mirrored here so a flow-shaped body stays inside its block).
func flowDepthOf(line string, depth int) int {
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
