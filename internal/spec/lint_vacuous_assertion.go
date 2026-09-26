package spec

// lint_vacuous_assertion.go — VacuousTestAssertionRule
// (SPEC-SPEC-LINT-VACUOUS-ASSERT-001, card t1269).
//
// WHAT IT CATCHES. Acceptance criteria verify behavior by running `go test`
// with a `-run` pattern and asserting on the verbose output. Two shapes of that
// verification pass without the behavior existing:
//
//   - run-pattern: `-run` is an unanchored regular expression, so a pattern not
//     closed with ^…$ at every alternation branch and every subtest level
//     selects every test whose name merely contains it.
//   - outcome-assertion: `--- PASS: TestX` written without a whitespace
//     delimiter after the name is also satisfied by any longer name sharing the
//     prefix. Go prints the name followed by a space, so the space is the
//     delimiter; the regex word boundary is NOT one (it holds before the `/` of
//     a subtest printed under a failing parent).
//
// WHY A RULE AND NOT PROSE. The same judgment written as grep commands inside a
// SPEC recurred three times (t1243): prose has no execution record, no positive
// control, and is its own input — its blockquote exclusion opened a blind spot.
// This rule judges every line raw, blockquotes included, and never exempts a
// line because it looks like documentation.
//
// SEVERITY. Warning. Non-advisory only for SPECs whose `created` is on or after
// vacuousGateCutoff, which was chosen strictly later than every `created` in
// the registering tree — so the corpus lands with zero gated findings and the
// unchanged baseline gate turns red on any new violation. A missing `created`
// is advisory (FrontmatterSchemaRule already errors on it); a present but
// unparseable one fails closed, so a malformed date is not an evasion path.
//
// KNOWN DETECTION LIMITS (spec.md §E):
//   - A pattern containing a shell expansion ($VAR, ${…}, $(…)) is not judged:
//     its value is unknown until the shell runs.
//   - A `-run` flag on a backslash-continued line without the `go test` token
//     is not judged; lines are not joined.
//   - Whether a pattern selects any existing test is not judged (axis c).
//   - Outcome spellings other than the four prefixes are not recognized.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// vacuousGateCutoff is the first `created` date whose findings gate. It is the
// day after the newest `created` in the tree that registered the rule; the
// pinning test (TestVacuousAssertionRule_GateCutoff) must move with it.
const vacuousGateCutoff = "2026-09-27"

// vacuousScannedArtifacts are the artifacts that carry decision rules.
// progress.md records commands that already ran; spec-compact.md duplicates
// spec.md; research.md and design.md describe existing state.
var vacuousScannedArtifacts = []string{"spec.md", "plan.md", "acceptance.md"}

var (
	// runFlagPattern locates a -run flag that is not part of a longer flag.
	runFlagPattern = regexp.MustCompile(`(^|[\s` + "`" + `])-run(=|\s+)`)
	// shellExpansionPattern marks a pattern whose value the shell decides.
	shellExpansionPattern = regexp.MustCompile(`(^|[^\\])\$[A-Za-z_{(]`)
	// leadingFlagGroup is an RE2 flag-only group such as (?i).
	leadingFlagGroup = regexp.MustCompile(`^\(\?[A-Za-z]+\)`)
	outcomePrefixes  = []string{"--- PASS: ", "--- FAIL: ", "--- (PASS|FAIL): ", "--- (FAIL|PASS): "}
)

// VacuousTestAssertionRule reports unanchored `-run` patterns and undelimited
// outcome assertions in a SPEC's decision-rule artifacts.
//
// @MX:NOTE: [AUTO] Gate is keyed on frontmatter `created` vs vacuousGateCutoff; warnings on older SPECs stay advisory so the corpus lands green
type VacuousTestAssertionRule struct{}

func (r *VacuousTestAssertionRule) Code() string { return "VacuousTestAssertion" }

// Check reads the scanned artifacts beside doc.Path and emits one warning per
// unanchored run pattern and per undelimited outcome assertion.
func (r *VacuousTestAssertionRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	advisory := vacuousAdvisory(doc.Frontmatter.Created)
	dir := filepath.Dir(doc.Path)
	var findings []Finding
	for _, name := range vacuousScannedArtifacts {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			for _, msg := range vacuousLineMessages(line) {
				findings = append(findings, Finding{
					File:     path,
					Line:     i + 1,
					Severity: SeverityWarning,
					Code:     r.Code(),
					Message:  msg,
					Advisory: advisory,
				})
			}
		}
	}
	return findings
}

// vacuousAdvisory reports whether findings for a SPEC created on `created` are
// advisory: missing → advisory; unparseable → gated (fail closed); otherwise
// advisory exactly when the date precedes the cutoff.
func vacuousAdvisory(created string) bool {
	created = strings.Trim(strings.TrimSpace(created), `"'`)
	if created == "" {
		return true
	}
	if _, err := time.Parse("2006-01-02", created); err != nil {
		return false
	}
	return created < vacuousGateCutoff
}

// vacuousLineMessages returns one message per violation on a single line.
func vacuousLineMessages(line string) []string {
	// A markdown table row reads `\|` as `|` (markdown escaping, not regex).
	if strings.HasPrefix(strings.TrimLeft(line, " \t>"), "|") {
		line = strings.ReplaceAll(line, `\|`, "|")
	}
	var msgs []string
	if strings.Contains(line, "go test") {
		for _, pat := range runPatterns(line) {
			if shellExpansionPattern.MatchString(pat.raw) || runPatternAnchored(pat.value) {
				continue
			}
			msgs = append(msgs, fmt.Sprintf(
				"run-pattern: -run %s is not anchored at every alternation branch and subtest level, "+
					"so it also selects longer test names; use -run '%s'", pat.raw, anchoredForm(pat.value)))
		}
	}
	for _, name := range undelimitedAssertions(line) {
		msgs = append(msgs, fmt.Sprintf(
			"outcome-assertion: '%s' has no whitespace delimiter after the test name, "+
				"so a longer name sharing the prefix also satisfies it; use '%s ' (the name followed by a space)",
			name, name))
	}
	return msgs
}

type runPattern struct {
	raw   string // as written, quotes included
	value string // the regular expression go test receives
}

// runPatterns extracts every -run argument on a line.
func runPatterns(line string) []runPattern {
	var out []runPattern
	for _, loc := range runFlagPattern.FindAllStringIndex(line, -1) {
		rest := line[loc[1]:]
		if rest == "" {
			continue
		}
		switch rest[0] {
		case '\'':
			end := strings.IndexByte(rest[1:], '\'')
			if end < 0 {
				end = len(rest) - 1
			}
			out = append(out, runPattern{raw: rest[:min(end+2, len(rest))], value: rest[1 : end+1]})
		case '"':
			end := 1
			for end < len(rest) && (rest[end] != '"' || rest[end-1] == '\\') {
				end++
			}
			value := strings.ReplaceAll(rest[1:end], `\$`, "$")
			out = append(out, runPattern{raw: rest[:min(end+1, len(rest))], value: value})
		default:
			end := strings.IndexAny(rest, " \t`")
			if end < 0 {
				end = len(rest)
			}
			out = append(out, runPattern{raw: rest[:end], value: rest[:end]})
		}
	}
	return out
}

// escaped reports whether s[i] is preceded by an odd number of backslashes.
func escaped(s string, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}

// classEnd returns the index of the `]` closing the character class opened at
// s[i], honoring a leading `]` / `^]` and POSIX `[:name:]` items.
func classEnd(s string, i int) int {
	j := i + 1
	if j < len(s) && s[j] == '^' {
		j++
	}
	if j < len(s) && s[j] == ']' {
		j++
	}
	for ; j < len(s); j++ {
		switch {
		case s[j] == '\\':
			j++
		case strings.HasPrefix(s[j:], "[:"):
			if k := strings.Index(s[j+2:], ":]"); k >= 0 {
				j += k + 3
			}
		case s[j] == ']':
			return j
		}
	}
	return len(s) - 1
}

// splitTopLevel splits s on sep where sep is outside groups and classes and not
// escaped. It also reports, for a string starting with `(`, the index of the
// paren that closes it (-1 otherwise).
func splitTopLevel(s string, sep byte) (parts []string, firstClose int) {
	depth, start := 0, 0
	firstClose = -1
	for i := 0; i < len(s); i++ {
		if escaped(s, i) {
			continue
		}
		switch s[i] {
		case '[':
			i = classEnd(s, i)
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 && firstClose < 0 && len(s) > 0 && s[0] == '(' {
				firstClose = i
			}
		case sep:
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, s[start:]), firstClose
}

// runPatternAnchored applies the spec.md §B.1 "Anchored" definition.
func runPatternAnchored(pattern string) bool {
	levels, _ := splitTopLevel(pattern, '/')
	for _, level := range levels {
		level = leadingFlagGroup.ReplaceAllString(level, "")
		if levelGrouped(level) {
			continue
		}
		branches, _ := splitTopLevel(level, '|')
		for _, b := range branches {
			if !strings.HasPrefix(b, "^") || !strings.HasSuffix(b, "$") || escaped(b, len(b)-1) || len(b) < 2 {
				return false
			}
		}
	}
	return true
}

// levelGrouped reports the grouped form ^( inner )$ where the opening paren
// closes at the level's second-to-last character — paren matching, not a
// prefix-and-suffix check, so ^(A)|(B)$ is not grouped.
func levelGrouped(level string) bool {
	if len(level) < 4 || !strings.HasPrefix(level, "^(") || !strings.HasSuffix(level, ")$") {
		return false
	}
	_, closeAt := splitTopLevel(level[1:], '|')
	return closeAt+1 == len(level)-2
}

// anchoredForm suggests the conformant spelling of a pattern.
func anchoredForm(pattern string) string {
	levels, _ := splitTopLevel(pattern, '/')
	for i, level := range levels {
		branches, _ := splitTopLevel(level, '|')
		for j, b := range branches {
			b = strings.TrimPrefix(b, "^")
			if strings.HasSuffix(b, "$") && !escaped(b, len(b)-1) {
				b = strings.TrimSuffix(b, "$")
			}
			branches[j] = b
		}
		if len(branches) == 1 {
			levels[i] = "^" + branches[0] + "$"
		} else {
			levels[i] = "^(" + strings.Join(branches, "|") + ")$"
		}
	}
	return strings.Join(levels, "/")
}

// undelimitedAssertions returns prefix+name for each outcome assertion on the
// line whose name is not followed by a whitespace delimiter.
func undelimitedAssertions(line string) []string {
	var out []string
	for _, prefix := range outcomePrefixes {
		for from := 0; ; {
			k := strings.Index(line[from:], prefix)
			if k < 0 {
				break
			}
			start := from + k + len(prefix)
			from = start
			end := testNameEnd(line, start)
			if end == start {
				continue // no name: a prefix listing, not an assertion
			}
			if !delimiterAt(line, end) {
				out = append(out, prefix+line[start:end])
			}
		}
	}
	return out
}

// delimiterAt reports a whitespace delimiter at s[i]: space, tab, \s, or
// [[:space:]].
func delimiterAt(s string, i int) bool {
	rest := s[i:]
	return strings.HasPrefix(rest, " ") || strings.HasPrefix(rest, "\t") ||
		strings.HasPrefix(rest, `\s`) || strings.HasPrefix(rest, "[[:space:]]")
}

// testNameEnd returns the end of the test name starting at s[i]: a Go
// identifier, then `/`-separated subtest components. A component stops at
// whitespace, a quote, a backtick, `(`, or a delimiter; a backslash escape
// other than \s (e.g. `\.`) belongs to the component.
func testNameEnd(s string, i int) int {
	rs := []rune(s[i:])
	n := 0
	if n >= len(rs) || (!unicode.IsLetter(rs[0]) && rs[0] != '_') {
		return i
	}
	for n < len(rs) && (unicode.IsLetter(rs[n]) || unicode.IsDigit(rs[n]) || rs[n] == '_') {
		n++
	}
	for n+1 < len(rs) && rs[n] == '/' && componentRune(rs, n+1) > 0 {
		n++
		for {
			w := componentRune(rs, n)
			if w == 0 {
				break
			}
			n += w
		}
	}
	return i + len(string(rs[:n]))
}

// componentRune returns how many runes at rs[n] belong to a subtest component
// (0 when the component ends there).
func componentRune(rs []rune, n int) int {
	if n >= len(rs) {
		return 0
	}
	switch r := rs[n]; {
	case unicode.IsSpace(r), r == '\'', r == '"', r == '`', r == '(':
		return 0
	case r == '\\':
		if n+1 >= len(rs) || rs[n+1] == 's' {
			return 0
		}
		return 2
	case strings.HasPrefix(string(rs[n:]), "[[:space:]]"):
		return 0
	}
	return 1
}
