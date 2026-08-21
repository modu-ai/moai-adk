// loader.go — neutral-layer .md parsing.
//
// The .md structure contract (verified against the 11 template files):
// YAML frontmatter between two "---" lines, in field order name,
// description (block scalar |), tools (CSV string), model, effort, color,
// permissionMode, memory, skills (optional YAML array), hooks (optional) —
// then the body (the agent prompt), carried byte-exact.
//
// The splitter anchors on the FIRST closing "---" only: bodies may contain
// bare "---" horizontal rules (e.g. plan-auditor) that must never be
// mistaken for a delimiter.
package agentemit

import (
	"fmt"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// agentFrontmatter is the YAML-decoded frontmatter subset the emitter
// consumes. Fields the Codex side documents as drops (color, permissionMode,
// memory, hooks) are intentionally not modeled — yaml.Unmarshal ignores
// unknown keys, which is the acceptance path for them.
type agentFrontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       string   `yaml:"tools"`
	Model       string   `yaml:"model"`
	Effort      string   `yaml:"effort"`
	Skills      []string `yaml:"skills"`
}

// ParseAgentDoc parses one agent .md source into its neutral form. It is
// fail-closed: a missing opening/closing delimiter, a missing or
// stem-mismatched name, a missing tools line, or an empty CSV token is an
// error naming the file and the offending value.
func ParseAgentDoc(filename string, data []byte) (AgentDoc, error) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return AgentDoc{}, fmt.Errorf("%s: missing opening \"---\" frontmatter delimiter", filename)
	}

	// Locate the first closing "---" line; the body is everything after it.
	offset := len("---\n")
	closeStart, closeEnd := -1, -1
	for offset <= len(s) {
		lineEnd := strings.IndexByte(s[offset:], '\n')
		var line string
		next := len(s)
		if lineEnd >= 0 {
			line = s[offset : offset+lineEnd]
			next = offset + lineEnd + 1
		} else {
			line = s[offset:]
		}
		if line == "---" {
			closeStart, closeEnd = offset, next
			break
		}
		if lineEnd < 0 {
			break
		}
		offset = next
	}
	if closeStart < 0 {
		return AgentDoc{}, fmt.Errorf("%s: missing closing \"---\" frontmatter delimiter", filename)
	}

	var fm agentFrontmatter
	if err := yaml.Unmarshal([]byte(s[len("---\n"):closeStart]), &fm); err != nil {
		return AgentDoc{}, fmt.Errorf("%s: frontmatter parse: %w", filename, err)
	}
	if strings.TrimSpace(fm.Name) == "" {
		return AgentDoc{}, fmt.Errorf("%s: frontmatter has no name", filename)
	}
	stem := strings.TrimSuffix(path.Base(filename), ".md")
	if fm.Name != stem {
		return AgentDoc{}, fmt.Errorf("%s: frontmatter name %q does not match file stem %q (ambiguous agent identity)", filename, fm.Name, stem)
	}
	if strings.TrimSpace(fm.Tools) == "" {
		return AgentDoc{}, fmt.Errorf("%s: frontmatter has no tools CSV", filename)
	}
	tools, err := splitToolsCSV(filename, fm.Tools)
	if err != nil {
		return AgentDoc{}, err
	}

	return AgentDoc{
		File:        path.Base(filename),
		Name:        fm.Name,
		Description: fm.Description,
		Tools:       tools,
		Model:       fm.Model,
		Effort:      fm.Effort,
		Skills:      fm.Skills,
		Body:        []byte(s[closeEnd:]),
	}, nil
}

// splitToolsCSV splits the tools CSV, trimming whitespace around each token.
// An empty token (e.g. from a trailing comma) is a fail-closed error — an
// empty class lookup must never be emitted.
func splitToolsCSV(filename, csv string) ([]string, error) {
	var tools []string
	for _, tok := range strings.Split(csv, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return nil, fmt.Errorf("%s: empty tools token in CSV %q (trailing comma?)", filename, csv)
		}
		tools = append(tools, tok)
	}
	return tools, nil
}
