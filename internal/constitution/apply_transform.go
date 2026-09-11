package constitution

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// In-memory transforms of the apply step (SPEC-CON-AMEND-APPLY-001
// REQ-CAA-001 … REQ-CAA-004, REQ-CAA-016). Both are pure functions from the
// pre-apply bytes to the new bytes or an error, so a dry-run and a real apply
// share one validation path.

// countOccurrences counts the occurrences of needle in content as an exact
// byte sequence, overlapping occurrences included, with no whitespace
// normalization. An empty needle occurs nowhere.
func countOccurrences(content []byte, needle string) int {
	if needle == "" {
		return 0
	}
	n := 0
	for i := 0; ; {
		j := bytes.Index(content[i:], []byte(needle))
		if j < 0 {
			return n
		}
		n++
		i += j + 1
	}
}

// replaceSourceClause replaces the one occurrence of current with next in the
// source rule file content. The new clause must occur zero times beforehand
// (REQ-CAA-016), then the current clause exactly once in the whole file
// (REQ-CAA-001, REQ-CAA-002); every byte outside that occurrence is kept.
func replaceSourceClause(path string, content []byte, current, next string) ([]byte, error) {
	if n := countOccurrences(content, next); n != 0 {
		return nil, fmt.Errorf("rule file %s: the new clause already occurs %d time(s); want none before the apply", path, n)
	}
	if n := countOccurrences(content, current); n != 1 {
		return nil, fmt.Errorf("rule file %s: the current clause occurs %d time(s); want exactly one", path, n)
	}
	return bytes.Replace(content, []byte(current), []byte(next), 1), nil
}

// registryClauseLine matches an entry's clause: line and captures its indent.
var registryClauseLine = regexp.MustCompile(`^(\s+)clause:\s`)

// quoteYAMLScalar encodes s as a single-line double-quoted yaml scalar.
func quoteYAMLScalar(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}

// decodeRegistryEntries parses registry content with the loader's own fence
// extraction and yaml decoding.
func decodeRegistryEntries(content string) ([]rawEntry, error) {
	fence, err := extractYAMLFence(content)
	if err != nil {
		return nil, err
	}
	var raw []rawEntry
	if err := yaml.Unmarshal([]byte(fence), &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// rewriteRegistryClause rewrites only the clause: line of entry ruleID inside
// the registry's yaml fence (REQ-CAA-003), then re-parses the candidate
// content the way the loader does and verifies the entry count and the
// target clause before anything is written (REQ-CAA-004).
func rewriteRegistryClause(path string, content []byte, ruleID, next string) ([]byte, error) {
	text := string(content)
	before, err := decodeRegistryEntries(text)
	if err != nil {
		return nil, fmt.Errorf("registry %s: %w", path, err)
	}
	if strings.ContainsAny(next, "\r\n") {
		return nil, fmt.Errorf("registry %s: the new clause for %s contains a line break; a registry clause is one line", path, ruleID)
	}

	fence, _ := extractYAMLFence(text)
	start := strings.Index(text, "```yaml") + len("```yaml")
	lines := strings.SplitAfter(fence, "\n")
	idLine := -1
	for i, l := range lines {
		if strings.TrimRight(l, " \t\r\n") == "- id: "+ruleID {
			idLine = i
			break
		}
	}
	clauseLine := -1
	if idLine >= 0 {
		matches := 0
		for i := idLine + 1; i < len(lines) && !strings.HasPrefix(lines[i], "- "); i++ {
			if registryClauseLine.MatchString(lines[i]) {
				clauseLine = i
				matches++
			}
		}
		if matches != 1 {
			clauseLine = -1
		}
	}
	if clauseLine < 0 {
		return nil, fmt.Errorf("registry %s: entry %s has no single clause: line to rewrite", path, ruleID)
	}
	old := lines[clauseLine]
	ending := "\n"
	if strings.HasSuffix(old, "\r\n") {
		ending = "\r\n"
	} else if !strings.HasSuffix(old, "\n") {
		ending = ""
	}
	indent := registryClauseLine.FindStringSubmatch(old)[1]
	lines[clauseLine] = indent + "clause: " + quoteYAMLScalar(next) + ending
	rewritten := text[:start] + strings.Join(lines, "") + text[start+len(fence):]

	after, err := decodeRegistryEntries(rewritten)
	if err != nil {
		return nil, fmt.Errorf("registry %s: the rewritten registry does not parse: %w", path, err)
	}
	if len(after) != len(before) {
		return nil, fmt.Errorf("registry %s: the rewritten registry holds a different number of entries", path)
	}
	for _, e := range after {
		if e.ID == ruleID {
			if e.Clause != next {
				return nil, fmt.Errorf("registry %s: entry %s decodes to a different clause after the rewrite", path, ruleID)
			}
			return []byte(rewritten), nil
		}
	}
	return nil, fmt.Errorf("registry %s: entry %s is missing after the rewrite", path, ruleID)
}
