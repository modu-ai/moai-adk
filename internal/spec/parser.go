package spec

import (
	"fmt"
	"regexp"
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

// ParseAcceptanceCriteriaTyped returns parsing results with Errors and Warnings separated
func ParseAcceptanceCriteriaTyped(markdown string, isFlatFormat bool) *ParseResult {
	return parseAcceptanceCriteriaInternal(markdown, isFlatFormat)
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
	id      string
	given   string
	when    string
	then    string
	reqIDs  []string
	indent  int
}

// findACSectionStart finds the start index of Acceptance Criteria section in markdown
func findACSectionStart(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "##") && strings.Contains(strings.ToLower(trimmed), "acceptance") {
			return i + 1
		}
	}
	return -1
}

// extractACLines extracts parsed line list from AC section
func extractACLines(lines []string, startIdx int, isFlatFormat bool) []acParsedLine {
	var acLines []acParsedLine

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "##") {
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
		})
	}

	return acLines
}

// buildTree converts parsed line list to hierarchical structure
func buildTree(acLines []acParsedLine, _ bool, result *ParseResult) []Acceptance {
	// stack: [root_level_node, level_1_node, ...]
	type stackEntry struct {
		node     *Acceptance
		indent   int
		lineIdx  int
	}

	var roots []Acceptance
	var stack []stackEntry
	seenIDs := make(map[string]bool)

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
			})
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

	for i := range roots {
		if len(roots[i].Children) == 0 && !hasIDSuffix(roots[i].ID) {
			roots[i] = autoWrapSingle(roots[i])
		}
	}

	return roots
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

	// AC ID pattern: AC-XXX-NNN-NN or AC-XXX-NNN-NN.a or AC-XXX-NNN-NN.a.i
	acIDPattern := regexp.MustCompile(`^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\s*:\s*`)
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
