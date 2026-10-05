// loader.go — command-source parsing.
//
// The command source structure contract (verified against the 16 template
// files): YAML frontmatter between two "---" lines — one line each for
// description, argument-hint, allowed-tools (the latter two are Claude-only
// keys and are NOT carried into the published skill) — then the body.
//
// The description of a .md.tmpl source is a locale-conditional Go template
// on a single line:
//
//	{{if eq .ConversationLanguage "ko"}}…{{else if …}}…{{else}}ENGLISH{{end}}
//
// The emitter publishes the unconditional {{else}} branch (the English
// variant). Extraction is fail-closed: a conditional with no {{else}}
// branch, an unbalanced conditional, or a branch that still carries
// template syntax is an error naming the file.
package commandemit

import (
	"fmt"
	"path"
	"strings"
)

// skillNamePrefix is prepended to every command name to derive the
// published-skill identity.
const skillNamePrefix = "moai-"

// ParseCommandDoc parses one command source into its neutral form. It is
// fail-closed: a missing opening/closing delimiter, a missing description,
// or an unextractable conditional description is an error naming the file
// and the offending value.
func ParseCommandDoc(filename string, data []byte) (CommandDoc, error) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return CommandDoc{}, fmt.Errorf("%s: missing opening \"---\" frontmatter delimiter", filename)
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
		return CommandDoc{}, fmt.Errorf("%s: missing closing \"---\" frontmatter delimiter", filename)
	}

	stem := commandStem(filename)
	if stem == "" {
		return CommandDoc{}, fmt.Errorf("%s: cannot derive command name from file name", filename)
	}

	descLine, err := frontmatterDescription(filename, s[len("---\n"):closeStart])
	if err != nil {
		return CommandDoc{}, err
	}
	desc, err := englishDescription(filename, descLine)
	if err != nil {
		return CommandDoc{}, err
	}

	return CommandDoc{
		File:        path.Base(filename),
		Command:     stem,
		SkillName:   skillNamePrefix + stem,
		Description: desc,
		Body:        []byte(s[closeEnd:]),
	}, nil
}

// commandStem derives the command name from the source file name: the stem
// of "<command>.md.tmpl" or "<command>.md".
func commandStem(filename string) string {
	base := path.Base(filename)
	base = strings.TrimSuffix(base, ".md.tmpl")
	base = strings.TrimSuffix(base, ".md")
	return base
}

// frontmatterDescription returns the raw value of the frontmatter
// "description:" line. The sources carry single-line descriptions, so a
// line scan is exact; a missing description is fail-closed.
func frontmatterDescription(filename, fm string) (string, error) {
	for _, line := range strings.Split(fm, "\n") {
		if rest, ok := strings.CutPrefix(line, "description:"); ok {
			desc := strings.TrimSpace(rest)
			if desc == "" {
				return "", fmt.Errorf("%s: description is empty", filename)
			}
			return desc, nil
		}
	}
	return "", fmt.Errorf("%s: frontmatter has no description line", filename)
}

// englishDescription reduces a raw description value to the language-neutral
// English variant. A plain description passes through (unquoted when
// quoted); a locale-conditional Go template must expose an unconditional
// {{else}} branch, whose content is returned — fail-closed otherwise.
func englishDescription(filename, raw string) (string, error) {
	if !strings.Contains(raw, "{{") {
		return unquoteYAMLScalar(filename, raw)
	}

	branch, err := elseBranch(filename, raw)
	if err != nil {
		return "", err
	}
	if strings.Contains(branch, "{{") || strings.Contains(branch, "}}") {
		return "", fmt.Errorf("%s: description {{else}} branch still carries template syntax (%q) — not extractable", filename, branch)
	}
	return unquoteYAMLScalar(filename, branch)
}

// elseBranch extracts the content of the top-level {{else}} branch of the
// locale conditional: the text between the {{else}} at depth 1 and the
// {{end}} that closes the initial {{if}}.
func elseBranch(filename, raw string) (string, error) {
	start := -1 // content start, just past the depth-1 {{else}}
	depth := 0
	for i := 0; i < len(raw); {
		open := strings.Index(raw[i:], "{{")
		if open < 0 {
			break
		}
		tokStart := i + open
		close := strings.Index(raw[tokStart+2:], "}}")
		if close < 0 {
			return "", fmt.Errorf("%s: unterminated template action in description %q", filename, raw)
		}
		tokEnd := tokStart + 2 + close + 2
		action := raw[tokStart+2 : tokStart+2+close]
		action = strings.TrimSpace(action)

		switch {
		case strings.HasPrefix(action, "if "):
			depth++
		case action == "else":
			if depth == 1 && start < 0 {
				start = tokEnd
			}
		case action == "end":
			depth--
			if depth == 0 {
				if start < 0 {
					return "", fmt.Errorf("%s: locale-conditional description has no {{else}} branch — English variant not extractable", filename)
				}
				return raw[start:tokStart], nil
			}
			if depth < 0 {
				return "", fmt.Errorf("%s: unbalanced template actions in description %q", filename, raw)
			}
		}
		i = tokEnd
	}
	return "", fmt.Errorf("%s: unbalanced template actions in description %q", filename, raw)
}

// unquoteYAMLScalar decodes a YAML plain or double-quoted scalar as carried
// on the description line. Only the quoting forms the sources use are
// supported; anything else fails closed.
func unquoteYAMLScalar(filename, v string) (string, error) {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		inner := v[1 : len(v)-1]
		var b strings.Builder
		for i := 0; i < len(inner); i++ {
			switch inner[i] {
			case '\\':
				if i+1 >= len(inner) {
					return "", fmt.Errorf("%s: dangling escape in description %q", filename, v)
				}
				i++
				switch inner[i] {
				case '"', '\\':
					b.WriteByte(inner[i])
				default:
					return "", fmt.Errorf("%s: unsupported escape \\%c in description %q", filename, inner[i], v)
				}
			case '\n', '\r':
				return "", fmt.Errorf("%s: control character in description %q", filename, v)
			default:
				b.WriteByte(inner[i])
			}
		}
		return b.String(), nil
	}
	if strings.ContainsAny(v, "\n\r\x00") {
		return "", fmt.Errorf("%s: control character in description %q", filename, v)
	}
	return v, nil
}

// quoteYAMLScalar renders a description as a YAML double-quoted scalar so
// colons, quotes, and non-ASCII survive the round trip.
func quoteYAMLScalar(v string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(v); i++ {
		switch v[i] {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		default:
			b.WriteByte(v[i])
		}
	}
	b.WriteByte('"')
	return b.String()
}
