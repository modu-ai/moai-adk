package spec

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type ParseResult struct {
	Criteria []Acceptance
	Errors   []error
	Warnings []error
}

// ParseAcceptanceCriteria parses Acceptance Criteria section in SPEC markdown
//
//   - Automatic Given inheritance
//   - Maximum depth validation (MaxDepth)
//   - REQ mapping extraction (maps REQ-XXX pattern)
func ParseAcceptanceCriteria(markdown string, isFlatFormat bool) ([]Acceptance, []error) {
	result := parseAcceptanceCriteriaInternal(markdown, isFlatFormat)
	return result.Criteria, append(result.Errors, result.Warnings...)
}

func parseAcceptanceCriteriaInternal(markdown string, isFlatFormat bool) *ParseResult {
	result := &ParseResult{}

	lines := strings.Split(markdown, "\n")

	// Find Acceptance Criteria section
	startIdx := findACSectionStart(lines)
	if startIdx < 0 {
		result.Errors = append(result.Errors, fmt.Errorf("acceptance criteria section not found"))
		return result
	}

	acLines := extractACLines(lines, startIdx, isFlatFormat)

	if len(acLines) == 0 {
		return result
	}

	result.Criteria = buildTree(acLines, isFlatFormat, result)

	for i := range result.Criteria {
		if err := result.Criteria[i].ValidateDepth(); err != nil {
			result.Errors = append(result.Errors, err)
		}
	}

	return result
}

type acParsedLine struct {
	id     string
	given  string
	when   string
	then   string
	reqIDs []string
	indent int
	line   int
}

// acSectionVocabulary is the explicit list of phrases that name an acceptance
// criteria section (card t565). Matching is case-insensitive and runs after the
// file name acceptance.md is removed, so a heading that only points at the
// sibling file does not name the section.
var acSectionVocabulary = []string{
	"acceptance",
	"success criteria",
	"ac matrix",
	"수락 기준",
	"인수 기준",
	"검수 기준",
}

// acNegativeSectionMarkers mark a heading about what the SPEC does not cover;
// such a heading never anchors, even when it mentions acceptance criteria.
var acNegativeSectionMarkers = []string{"out of scope", "out-of-scope", "non-goal"}

// markdownHeadingLevel returns the ATX heading level of a trimmed line, or 0
// when the line is not a heading.
func markdownHeadingLevel(trimmed string) int {
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return 0
	}
	if level < len(trimmed) && trimmed[level] != ' ' && trimmed[level] != '\t' {
		return 0
	}
	return level
}

// isACSectionHeading reports whether a heading names the acceptance criteria
// section.
func isACSectionHeading(trimmed string) bool {
	text := strings.ReplaceAll(strings.ToLower(trimmed), "acceptance.md", "")
	for _, marker := range acNegativeSectionMarkers {
		if strings.Contains(text, marker) {
			return false
		}
	}
	for _, phrase := range acSectionVocabulary {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

// findACSectionStart finds the start index of Acceptance Criteria section in markdown:
// the line after the first heading of level 2 or deeper that names the section.
func findACSectionStart(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if markdownHeadingLevel(trimmed) >= 2 && isACSectionHeading(trimmed) {
			return i + 1
		}
	}
	return -1
}

// extractACLines extracts parsed line list from AC section. The section ends at
// the next heading of the same or a higher level than its anchor, so its own
// deeper subheadings are read.
func extractACLines(lines []string, startIdx int, isFlatFormat bool) []acParsedLine {
	var acLines []acParsedLine
	anchorLevel := markdownHeadingLevel(strings.TrimSpace(lines[startIdx-1]))

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if level := markdownHeadingLevel(trimmed); level > 0 && level <= anchorLevel {
			break
		}

		parsed := parseSingleACLine(trimmed)
		if parsed == nil {
			continue
		}

		indent := calculateIndentationDepth(line)

		// Process all lines with indent 0 in flat format
		if isFlatFormat {
			indent = 0
		}

		acLines = append(acLines, acParsedLine{
			id:     parsed.id,
			given:  parsed.given,
			when:   parsed.when,
			then:   parsed.then,
			reqIDs: parsed.reqIDs,
			indent: indent,
			line:   i + 1,
		})
	}

	return acLines
}

// buildTree converts parsed line list to hierarchical structure
func buildTree(acLines []acParsedLine, _ bool, result *ParseResult) []Acceptance {
	// stack: [root_level_node, level_1_node, ...]
	type stackEntry struct {
		node    *Acceptance
		indent  int
		lineIdx int
	}

	var roots []Acceptance
	var stack []stackEntry
	seenIDs := make(map[string]bool)
	// Duplicate ids (card t564): the grammar cannot tell a bullet that cites an
	// id from the bullet that declares it, so dropping either line loses its REQ
	// mapping silently. The first line's text is kept, every later line's REQ
	// mappings are collected here and merged below, and the duplicate is still
	// reported so the author sees it.
	duplicateReqIDs := make(map[string][]string)

	for i, acLine := range acLines {
		node := Acceptance{
			ID:             acLine.id,
			Given:          acLine.given,
			When:           acLine.when,
			Then:           acLine.then,
			RequirementIDs: acLine.reqIDs,
		}

		if seenIDs[acLine.id] {
			result.Errors = append(result.Errors, &DuplicateAcceptanceID{
				ID:    acLine.id,
				Depth: acLine.indent,
				Line:  acLine.line,
			})
			duplicateReqIDs[acLine.id] = append(duplicateReqIDs[acLine.id], acLine.reqIDs...)
			continue
		}
		seenIDs[acLine.id] = true

		// Remove items deeper than current indent from stack
		for len(stack) > 0 && stack[len(stack)-1].indent >= acLine.indent {
			stack = stack[:len(stack)-1]
		}

		if acLine.indent == 0 || len(stack) == 0 {
			roots = append(roots, node)
			stack = append(stack, stackEntry{
				node:    &roots[len(roots)-1],
				indent:  acLine.indent,
				lineIdx: i,
			})
		} else {
			parent := stack[len(stack)-1].node
			parent.Children = append(parent.Children, node)

			// Given inheritance: child inherits parent's Given if child's Given is empty
			child := &parent.Children[len(parent.Children)-1]
			if child.Given == "" && parent.Given != "" {
				child.Given = parent.Given
			}

			stack = append(stack, stackEntry{
				node:    child,
				indent:  acLine.indent,
				lineIdx: i,
			})
		}
	}

	// Merge before auto-wrapping, which copies RequirementIDs into the wrapper's
	// child and would otherwise leave the merged ids on the empty wrapper.
	mergeDuplicateReqIDs(roots, duplicateReqIDs)

	for i := range roots {
		if len(roots[i].Children) == 0 && !hasIDSuffix(roots[i].ID) {
			roots[i] = autoWrapSingle(roots[i])
		}
	}

	return roots
}

// mergeDuplicateReqIDs appends to each node the REQ ids mapped by later lines
// carrying the same AC id, skipping ids the node already holds.
func mergeDuplicateReqIDs(nodes []Acceptance, extra map[string][]string) {
	if len(extra) == 0 {
		return
	}
	for i := range nodes {
		for _, id := range extra[nodes[i].ID] {
			if !slices.Contains(nodes[i].RequirementIDs, id) {
				nodes[i].RequirementIDs = append(nodes[i].RequirementIDs, id)
			}
		}
		mergeDuplicateReqIDs(nodes[i].Children, extra)
	}
}

func hasIDSuffix(id string) bool {
	return strings.Contains(id, ".")
}

func autoWrapSingle(ac Acceptance) Acceptance {
	childID := ac.ID + ".a"
	child := Acceptance{
		ID:             childID,
		Given:          ac.Given,
		When:           ac.When,
		Then:           ac.Then,
		RequirementIDs: ac.RequirementIDs,
	}
	wrapped := Acceptance{
		ID:       ac.ID,
		Children: []Acceptance{child},
	}
	return wrapped
}

// acIDPattern anchors an AC declaration line, applied after the leading
// "- *" run has been trimmed. Its four widening axes were each derived from the
// corpus (SPEC-AC-COLLECTOR-ANCHOR-001 §B); the anchor is compiled once because
// it is evaluated per line over every spec.md.
//
//	AC-(?:[A-Za-z0-9]+-)*[0-9]+   axis 1 — variable segment count with
//	                              alphanumeric middle segments. The LAST segment
//	                              stays numeric: it is the only measured property
//	                              separating a declaration id from an arbitrary
//	                              AC-prefixed token, so dropping it would admit
//	                              any such token with nothing left to narrow it.
//	(?:\.[a-z](?:\.[a-z]+)?)?     axis 2 — sub-id suffix, UNCHANGED from the
//	                              pre-widening anchor. hasIDSuffix/autoWrapSingle
//	                              branch on the dot, so altering this changes the
//	                              tree shape rather than only recognition.
//	\*{0,2}                       axis 3 — the CLOSING bold marker. The opening
//	                              one is already removed by the TrimLeft below,
//	                              which is deliberately left alone: widening the
//	                              preprocessing would move the risk surface from
//	                              this anchor to every line in the section.
//	(?:\([^()]*\)\s*)?            axis 4a — one parenthesised qualifier between
//	                              the id and the separator, e.g. "(A1)",
//	                              "(REQ-001)". Skipping it exposes the real
//	                              separator behind it. Axes 3 and 4a compose in
//	                              EITHER order, which is why \*{0,2} appears on
//	                              both sides: the corpus writes the qualifier both
//	                              inside the bold span ("**AC-CSS-001-01
//	                              (isolation-validity)** —") and after it
//	                              ("**AC-HFC-001a** (REQ-HFC-001):"), and fixing
//	                              one order rejects 53 lines for a property of
//	                              this anchor rather than of the corpus.
//	                              BRACKET qualifiers ("**AC-1 [REQ-002]**:", 17
//	                              lines) stay out: that is a new axis, not the
//	                              composition of two declared ones.
//	[:—–]                         axis 4b — separator set: colon, em dash
//	                              (U+2014, observed), en dash (U+2013, a design
//	                              decision with no corpus observation). The
//	                              separator stays REQUIRED; widening the set is
//	                              not the same as dropping the requirement, and
//	                              the requirement is what keeps prose bullets out.
var acIDPattern = regexp.MustCompile(`^(AC-(?:[A-Za-z0-9]+-)*[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\*{0,2}\s*(?:\([^()]*\)\s*)?\*{0,2}\s*[:—–]\s*`)

func parseSingleACLine(line string) *struct {
	id     string
	given  string
	when   string
	then   string
	reqIDs []string
} {
	trimmed := strings.TrimSpace(line)

	trimmed = strings.TrimLeft(trimmed, "- *")
	trimmed = strings.TrimSpace(trimmed)

	idMatch := acIDPattern.FindStringSubmatch(trimmed)

	if len(idMatch) < 2 {
		return nil
	}

	id := idMatch[1]
	content := strings.TrimSpace(trimmed[len(idMatch[0]):])

	// Extract REQ mapping
	reqIDs := ExtractRequirementMappings(content)

	// Remove REQ mapping part
	reqRemover := regexp.MustCompile(`\(?\s*(?:maps|MAPS)\s+REQ-[A-Z0-9-]+\s*\)?`)
	cleanContent := strings.TrimSpace(reqRemover.ReplaceAllString(content, ""))

	// EARS pattern parsing: Given ... When ... Then ...
	var given, when, then string

	// Extract Given
	givenRe := regexp.MustCompile(`(?i)^Given\s+(.+?)(?:,\s*(?:When|then)|$)`)
	if match := givenRe.FindStringSubmatch(cleanContent); len(match) > 1 {
		given = "Given " + strings.TrimSpace(match[1])
		cleanContent = strings.TrimSpace(cleanContent[len(match[0]):])
	}

	// Extract When
	whenRe := regexp.MustCompile(`(?i)^When\s+(.+?)(?:,\s*(?:Then|then)|$)`)
	if match := whenRe.FindStringSubmatch(cleanContent); len(match) > 1 {
		when = "When " + strings.TrimSpace(match[1])
		cleanContent = strings.TrimSpace(cleanContent[len(match[0]):])
	}

	// Extract Then
	thenRe := regexp.MustCompile(`(?i)^Then\s+(.+)`)
	if match := thenRe.FindStringSubmatch(cleanContent); len(match) > 1 {
		then = "Then " + strings.TrimSpace(match[1])
	}

	// If pattern does not match, treat entire content as Then
	if when == "" && then == "" && given == "" {
		then = cleanContent
	}

	return &struct {
		id     string
		given  string
		when   string
		then   string
		reqIDs []string
	}{
		id:     id,
		given:  given,
		when:   when,
		then:   then,
		reqIDs: reqIDs,
	}
}

func calculateIndentationDepth(line string) int {
	spaces := 0
	for _, r := range line {
		switch r {
		case ' ':
			spaces++
		case '\t':
			spaces += 2
		default:
			// goto out would be cleaner but switch-break is fine
		}
		if r != ' ' && r != '\t' {
			break
		}
	}
	return spaces / 2
}

// CheckDanglingReferences checks for references to non-existent REQs
func CheckDanglingReferences(criteria []Acceptance, existingREQs map[string]bool) []error {
	var errors []error

	var check func(ac *Acceptance)
	check = func(ac *Acceptance) {
		for _, reqID := range ac.RequirementIDs {
			if !existingREQs[reqID] {
				errors = append(errors, &DanglingRequirementReference{
					ACID:     ac.ID,
					ReqID:    reqID,
					Location: ac.ID,
				})
			}
		}
		for i := range ac.Children {
			check(&ac.Children[i])
		}
	}

	for i := range criteria {
		check(&criteria[i])
	}

	return errors
}
