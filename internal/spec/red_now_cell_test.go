// red_now_cell_test.go — repository-local verifier for the RED-now cell
// adoption gate (card t343).
//
// The MP-8 form contract itself lives inside a sentinel pair in
// .claude/agents/moai/plan-auditor.md and its template mirror. This test
// EXTRACTS that contract rather than restating it: a copy kept here would keep
// passing after the clause changed, which is the twin-drift failure this
// repository has already met in its .sh / .sh.tmpl pairs. The pattern is the
// one ac_count_clause_test.go established — sentinel extraction with
// exactly-one-and-non-empty asserted BEFORE any comparison, so an anchor
// matching zero or two spans cannot become a vacuous pass.
//
// The sentinel is an HTML comment rather than a `# `-prefixed line. A `# ` line
// in markdown prose is an H1 heading: it renders visibly and it terminates the
// enclosing section, which would break the `### M5` containment assertion this
// file makes. The token inside the comment is unchanged.
//
// Nothing here ships. The test is repository-local and its fixtures live under
// internal/spec/testdata/red_now, outside the distributed template tree.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	redNowBeginSentinel = "<!-- MOAI-REDNOW-BEGIN -->"
	redNowEndSentinel   = "<!-- MOAI-REDNOW-END -->"

	redNowAuditorPath       = ".claude/agents/moai/plan-auditor.md"
	redNowAuditorMirrorPath = "internal/template/templates/.claude/agents/moai/plan-auditor.md"
	redNowRulePath          = ".claude/rules/moai/development/verification-completeness.md"
	redNowRuleMirrorPath    = "internal/template/templates/.claude/rules/moai/development/verification-completeness.md"

	redNowFixtureDir = "testdata/red_now"

	redNowM5Heading         = "### M5: Must-Pass Firewall"
	redNowGroup4Heading     = "### Group 4: Acceptance Criteria Quality"
	redNowMustPassHeading   = "## Must-Pass Results"
	redNowRuleSectionHead   = "## 2. Two-cell adoption discipline"
	redNowMustPassMP8Prefix = "- [PASS/FAIL/N/A] MP-8"
)

// ---------------------------------------------------------------------------
// span extraction
// ---------------------------------------------------------------------------

func redNowRead(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}

// redNowExtractSentinelSpan returns the body between the sentinel pair.
//
// It returns an error rather than calling t.Fatalf so that the zero-pair and
// two-pair mutants (M-2) can be OBSERVED failing instead of aborting the run.
func redNowExtractSentinelSpan(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var begins, ends []int
	for i, ln := range lines {
		switch strings.TrimSpace(ln) {
		case redNowBeginSentinel:
			begins = append(begins, i)
		case redNowEndSentinel:
			ends = append(ends, i)
		}
	}
	if len(begins) != 1 || len(ends) != 1 {
		return "", fmt.Errorf("expected exactly one %q / %q sentinel pair, got begin=%d end=%d",
			redNowBeginSentinel, redNowEndSentinel, len(begins), len(ends))
	}
	if ends[0] <= begins[0] {
		return "", fmt.Errorf("END sentinel at line %d precedes BEGIN sentinel at line %d", ends[0]+1, begins[0]+1)
	}
	body := strings.Join(lines[begins[0]+1:ends[0]], "\n")
	if strings.TrimSpace(body) == "" {
		return "", fmt.Errorf("sentinel pair delimits an empty span")
	}
	return body, nil
}

func redNowMustExtractSpan(t *testing.T, rel string) string {
	t.Helper()
	span, err := redNowExtractSentinelSpan(redNowRead(t, rel))
	if err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	return span
}

// redNowHeadingLevel returns the markdown heading level of a line, or 0.
func redNowHeadingLevel(line string) int {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n == 0 || n >= len(line) || line[n] != ' ' {
		return 0
	}
	return n
}

// redNowExtractSectionSpan returns the body of the section introduced by the
// given heading line, up to the next heading of the same or a higher level.
//
// Exactly-one is asserted before the body is returned: a heading matching zero
// or two lines would otherwise let every containment assertion pass vacuously.
func redNowExtractSectionSpan(content, heading string) (string, error) {
	lines := strings.Split(content, "\n")
	var starts []int
	for i, ln := range lines {
		if strings.TrimSpace(ln) == heading {
			starts = append(starts, i)
		}
	}
	if len(starts) != 1 {
		return "", fmt.Errorf("expected exactly one %q heading, got %d", heading, len(starts))
	}
	level := redNowHeadingLevel(heading)
	if level == 0 {
		return "", fmt.Errorf("%q is not a markdown heading", heading)
	}
	end := len(lines)
	for i := starts[0] + 1; i < len(lines); i++ {
		if l := redNowHeadingLevel(lines[i]); l != 0 && l <= level {
			end = i
			break
		}
	}
	body := strings.Join(lines[starts[0]:end], "\n")
	if strings.TrimSpace(strings.TrimPrefix(body, heading)) == "" {
		return "", fmt.Errorf("section %q is empty", heading)
	}
	return body, nil
}

func redNowMustExtractSection(t *testing.T, content, heading string) string {
	t.Helper()
	body, err := redNowExtractSectionSpan(content, heading)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return body
}

// ---------------------------------------------------------------------------
// the extracted form contract
// ---------------------------------------------------------------------------

// redNowForbiddenMetachars reads the machine-readable half of the MP-8 span.
// The tokens are EXTRACTED, never restated here: changing the contract in the
// agent file changes what this test enforces.
func redNowForbiddenMetachars(t *testing.T, span string) []string {
	t.Helper()
	var out []string
	for _, ln := range strings.Split(span, "\n") {
		ln = strings.TrimSpace(ln)
		const key = "forbidden-metacharacter:"
		if !strings.HasPrefix(ln, key) {
			continue
		}
		tok := strings.TrimSpace(strings.TrimPrefix(ln, key))
		tok = strings.Trim(tok, "`")
		if tok != "" {
			out = append(out, tok)
		}
	}
	if len(out) < 5 {
		t.Fatalf("MP-8 span yielded %d forbidden metacharacters (%v); a near-empty contract "+
			"would make every form check vacuous", len(out), out)
	}
	return out
}

// redNowUnquotedMask marks the byte positions of a command that sit outside any
// quoted span. A GFM table cell and a fenced ledger entry both routinely carry
// a literal `|` inside a quoted regex; treating that as a shell pipe would
// refuse commands that are in fact single invocations.
func redNowUnquotedMask(cmd string) []bool {
	mask := make([]bool, len(cmd))
	var inSingle, inDouble, escaped bool
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		switch {
		case escaped:
			escaped = false
		case c == '\\' && !inSingle:
			escaped = true
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		}
		mask[i] = !inSingle && !inDouble
	}
	return mask
}

// redNowFormViolations returns the forbidden metacharacters that appear
// UNQUOTED in the command, i.e. the ones that would actually make it more than
// one shell invocation.
func redNowFormViolations(cmd string, forbidden []string) []string {
	mask := redNowUnquotedMask(cmd)
	var hits []string
	for _, tok := range forbidden {
		for i := 0; i+len(tok) <= len(cmd); i++ {
			if !strings.HasPrefix(cmd[i:], tok) {
				continue
			}
			allBare := true
			for j := i; j < i+len(tok); j++ {
				if !mask[j] {
					allBare = false
					break
				}
			}
			if allBare {
				hits = append(hits, tok)
				break
			}
		}
	}
	return hits
}

// ---------------------------------------------------------------------------
// carrier-independent command collection
// ---------------------------------------------------------------------------

type redNowCommand struct {
	Carrier string // "table-cell" | "ledger-entry" | "fenced-block"
	Command string
	Line    int
}

var (
	redNowLedgerRe    = regexp.MustCompile(`^E-(\d+)\s\s*(\S.*)$`)
	redNowDollarRe    = regexp.MustCompile(`^\s*\$ (\S.*)$`)
	redNowInlineRe    = regexp.MustCompile("`([^`]+)`")
	redNowLedgerRefRe = regexp.MustCompile(`E-(\d+)`)
	redNowShaRe       = regexp.MustCompile("`[0-9a-f]{7,40}`")
)

// redNowCollectCommands walks an acceptance.md and returns every command it
// carries, whatever the carrier. Three carriers are recognised, and the scan
// is the same scan for all three — which is what makes the predicate
// carrier-independent by construction rather than by promise.
func redNowCollectCommands(content string) []redNowCommand {
	var out []redNowCommand
	inFence := false
	for i, ln := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			if m := redNowLedgerRe.FindStringSubmatch(ln); m != nil {
				out = append(out, redNowCommand{Carrier: "ledger-entry", Command: m[2], Line: i + 1})
				continue
			}
			if m := redNowDollarRe.FindStringSubmatch(ln); m != nil {
				out = append(out, redNowCommand{Carrier: "fenced-block", Command: m[1], Line: i + 1})
			}
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(ln), "|") {
			for _, m := range redNowInlineRe.FindAllStringSubmatch(ln, -1) {
				body := m[1]
				if strings.HasPrefix(body, "$ ") {
					out = append(out, redNowCommand{Carrier: "table-cell", Command: strings.TrimPrefix(body, "$ "), Line: i + 1})
				}
			}
			continue
		}
		if m := redNowDollarRe.FindStringSubmatch(ln); m != nil {
			out = append(out, redNowCommand{Carrier: "fenced-block", Command: m[1], Line: i + 1})
		}
	}
	return out
}

// redNowCollectCellCommandsOnly is the DELIBERATELY narrowed predicate the
// carrier-relocation mutant (M-5) walks through. It exists so the mutant can be
// observed surviving it while the real predicate catches it — rule §5: observe
// the two forms diverge, do not grep for the fixed form.
func redNowCollectCellCommandsOnly(content string) []redNowCommand {
	var out []redNowCommand
	for _, c := range redNowCollectCommands(content) {
		if c.Carrier == "table-cell" {
			out = append(out, c)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// the four-element check
// ---------------------------------------------------------------------------

type redNowLedgerEntry struct {
	Command   string
	HasStdout bool
	HasExit   bool
}

func redNowParseLedger(content string) map[string]redNowLedgerEntry {
	out := map[string]redNowLedgerEntry{}
	inFence := false
	current := ""
	for _, ln := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
			current = ""
			continue
		}
		if !inFence {
			continue
		}
		if m := redNowLedgerRe.FindStringSubmatch(ln); m != nil {
			current = "E-" + m[1]
			out[current] = redNowLedgerEntry{Command: m[2]}
			continue
		}
		if current == "" {
			continue
		}
		e := out[current]
		switch {
		case strings.HasPrefix(strings.TrimSpace(ln), "stdout:"):
			e.HasStdout = true
		case strings.HasPrefix(strings.TrimSpace(ln), "exit:"):
			e.HasExit = true
		}
		out[current] = e
	}
	return out
}

type redNowRow struct {
	ID      string
	Class   string
	RedCell string
	Line    int
}

var redNowRowRe = regexp.MustCompile(`^\|\s*\*\*(AC-[A-Z0-9-]+)\*\*\s*\|`)

// redNowParseRows returns the acceptance matrix rows. The RED-now proof column
// is the second-from-last cell of the six-column matrix.
func redNowParseRows(content string) []redNowRow {
	var out []redNowRow
	for i, ln := range strings.Split(content, "\n") {
		m := redNowRowRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		cells := strings.Split(strings.Trim(strings.TrimSpace(ln), "|"), "|")
		if len(cells) < 6 {
			continue
		}
		out = append(out, redNowRow{
			ID:      m[1],
			Class:   strings.TrimSpace(cells[1]),
			RedCell: strings.TrimSpace(cells[len(cells)-2]),
			Line:    i + 1,
		})
	}
	return out
}

func redNowIsReleaseBlocking(class string) bool {
	return strings.Contains(class, "release-blocking") && !strings.Contains(class, "regression-guard")
}

// redNowElementFindings reports every release-blocking row whose RED-now cell
// does not carry the four elements — the command, its verbatim stdout, its exit
// code, and a tree SHA. The elements may live in the cell or in a fenced
// evidence-ledger entry the cell cites by id; the check follows them to
// whichever carrier holds them.
func redNowElementFindings(content string) []string {
	ledger := redNowParseLedger(content)
	docPinned := redNowShaRe.MatchString(content)
	var out []string
	for _, row := range redNowParseRows(content) {
		if !redNowIsReleaseBlocking(row.Class) {
			continue
		}
		refs := redNowLedgerRefRe.FindAllString(row.RedCell, -1)
		resolved := false
		for _, ref := range refs {
			e, ok := ledger[ref]
			if !ok {
				out = append(out, fmt.Sprintf("%s (line %d): cites %s, which no ledger entry defines", row.ID, row.Line, ref))
				continue
			}
			if e.Command == "" || !e.HasStdout || !e.HasExit {
				out = append(out, fmt.Sprintf("%s (line %d): ledger entry %s is missing the command, its stdout, or its exit code", row.ID, row.Line, ref))
				continue
			}
			resolved = true
		}
		if !resolved {
			out = append(out, fmt.Sprintf("%s (line %d): release-blocking RED-now cell carries no command, stdout and exit code", row.ID, row.Line))
			continue
		}
		if !docPinned && !redNowShaRe.MatchString(row.RedCell) {
			out = append(out, fmt.Sprintf("%s (line %d): RED-now cell has no tree SHA and the document carries no pin to inherit", row.ID, row.Line))
		}
	}
	return out
}

// redNowFormFindings reports every command in the file — any carrier, any
// class — that is not a single shell invocation.
func redNowFormFindings(content string, forbidden []string) []string {
	var out []string
	for _, c := range redNowCollectCommands(content) {
		if hits := redNowFormViolations(c.Command, forbidden); len(hits) > 0 {
			out = append(out, fmt.Sprintf("line %d (%s): %q carries unquoted %v", c.Line, c.Carrier, c.Command, hits))
		}
	}
	return out
}

func redNowFixture(t *testing.T, name string) string {
	t.Helper()
	rel := filepath.Join("internal/spec", redNowFixtureDir, name, "acceptance.md")
	return redNowRead(t, rel)
}

// ===========================================================================
// L1 — the rule clause (AC-RNT-001, AC-RNT-003)
// ===========================================================================

// TestRuleClauseEnumeratesFourElements covers AC-RNT-001. Every predicate is
// asserted INSIDE the extracted §2 span, so a token pasted elsewhere in the
// file does not satisfy it. It does not defeat a token pasted inside the span
// — that residual is mutant M-3, recorded in acceptance.md §D.2.
func TestRuleClauseEnumeratesFourElements(t *testing.T) {
	span := redNowMustExtractSection(t, redNowRead(t, redNowRulePath), redNowRuleSectionHead)
	for _, want := range []string{
		"RED-now cell content",
		"command", "stdout", "exit", "SHA",
		"read-only", "single invocation", "raw file",
	} {
		if !strings.Contains(span, want) {
			t.Errorf("§2 span does not name %q", want)
		}
	}
	for _, carrier := range []string{"table cell", "evidence-ledger"} {
		if !strings.Contains(span, carrier) {
			t.Errorf("§2 span does not admit the carrier %q", carrier)
		}
	}
}

// TestRuleClauseStatesDemotionNotPass covers AC-RNT-003.
func TestRuleClauseStatesDemotionNotPass(t *testing.T) {
	span := redNowMustExtractSection(t, redNowRead(t, redNowRulePath), redNowRuleSectionHead)
	if !strings.Contains(span, "regression-guard") {
		t.Errorf("§2 span does not name the regression-guard disposition")
	}
	notPass := regexp.MustCompile(`not\s+(?:be\s+)?record(?:ed)?\s+as\s+a\s+pass`)
	if !notPass.MatchString(span) {
		t.Errorf("§2 span does not state that an undecidable RED is not recorded as a pass")
	}
	if !strings.Contains(span, "loses release-blocking eligibility") {
		t.Errorf("§2 span does not state the demotion as the disposition")
	}
}

// TestRuleClauseIsStructuralNotLexical covers REQ-RNT-002 on the rule surface:
// the clause must not key on the tense, mood, or a word list of the prose.
func TestRuleClauseIsStructuralNotLexical(t *testing.T) {
	content := redNowRead(t, redNowRulePath)
	for _, banned := range []string{"tense", "mood", "counterfactual"} {
		if strings.Contains(content, banned) {
			t.Errorf("%s carries the lexical discriminator %q", redNowRulePath, banned)
		}
	}
}

// ===========================================================================
// L2 / L3 — the MP-8 clause (AC-RNT-004..007, -013, -014, -015)
// ===========================================================================

func redNowMP8Span(t *testing.T) string {
	t.Helper()
	return redNowMustExtractSpan(t, redNowAuditorPath)
}

// TestMP8SpanNamesReexecution covers AC-RNT-004.
func TestMP8SpanNamesReexecution(t *testing.T) {
	content := redNowRead(t, redNowAuditorPath)
	span := redNowMP8Span(t)
	for _, want := range []string{"re-execute", "current tree", "RED reproduces"} {
		if !strings.Contains(span, want) {
			t.Errorf("MP-8 span does not name %q", want)
		}
	}
	m5 := redNowMustExtractSection(t, content, redNowM5Heading)
	if !strings.Contains(m5, span) {
		t.Errorf("MP-8 span is not reachable from %q — it sits outside the must-pass firewall", redNowM5Heading)
	}
}

// TestMP8SpanIsScoreIndependent covers AC-RNT-005.
func TestMP8SpanIsScoreIndependent(t *testing.T) {
	span := redNowMP8Span(t)
	for _, want := range []string{"severity=critical", "regardless of the aggregate score", "Verdict: FAIL"} {
		if !strings.Contains(span, want) {
			t.Errorf("MP-8 span does not carry %q", want)
		}
	}
}

// TestMP8SpanCarriesNABranch covers AC-RNT-006.
func TestMP8SpanCarriesNABranch(t *testing.T) {
	span := redNowMP8Span(t)
	for _, want := range []string{"N/A", "state the reason", "MP-4 precedent"} {
		if !strings.Contains(span, want) {
			t.Errorf("MP-8 span does not carry %q", want)
		}
	}
	report := redNowMustExtractSection(t, redNowRead(t, redNowAuditorPath), redNowMustPassHeading)
	if !strings.Contains(report, redNowMustPassMP8Prefix) {
		t.Errorf("the report template's MP-8 row does not admit N/A")
	}
}

// TestGroup4AndReportRowExist covers AC-RNT-007. Both assertions are scoped to
// a section span, so the six characters "AC-6:" pasted anywhere else in the
// file satisfy neither.
func TestGroup4AndReportRowExist(t *testing.T) {
	content := redNowRead(t, redNowAuditorPath)
	group4 := redNowMustExtractSection(t, content, redNowGroup4Heading)
	if !strings.Contains(group4, "AC-6:") {
		t.Errorf("%q carries no AC-6 checklist item", redNowGroup4Heading)
	}
	if !strings.Contains(group4, "MP-8") {
		t.Errorf("the Group 4 AC-6 item does not feed MP-8")
	}
	report := redNowMustExtractSection(t, content, redNowMustPassHeading)
	if !strings.Contains(report, redNowMustPassMP8Prefix) {
		t.Errorf("the report template carries no %q row", redNowMustPassMP8Prefix)
	}
}

// TestMP8SpanCarriesExecutionDiscipline covers AC-RNT-013.
func TestMP8SpanCarriesExecutionDiscipline(t *testing.T) {
	span := redNowMP8Span(t)
	for _, want := range []string{
		"shall not execute it further",
		"shall not record the criterion as a pass",
		"Repository execution discipline takes precedence",
		"timeout",
	} {
		if !strings.Contains(span, want) {
			t.Errorf("MP-8 span does not carry the execution-discipline branch %q", want)
		}
	}
}

// TestMP8SpanKeysOnExecutedCount covers AC-RNT-015.
func TestMP8SpanKeysOnExecutedCount(t *testing.T) {
	span := redNowMP8Span(t)
	for _, want := range []string{
		"count of tests actually executed",
		"no tests to run",
	} {
		if !strings.Contains(span, want) {
			t.Errorf("MP-8 span does not key the verdict on the executed-test count: missing %q", want)
		}
	}
	reject := regexp.MustCompile("not treat the presence of an `ok` token")
	if !reject.MatchString(span) {
		t.Errorf("MP-8 span does not reject an `ok`-token-only verdict")
	}
}

// TestMP8LivenessAnchors covers AC-RNT-014 and the §1.3 continued-firing axis:
// MP-8 disappearing from the agent file must be distinguishable from MP-8
// passing. The deletion is OBSERVED, on a mutated copy, not argued.
func TestMP8LivenessAnchors(t *testing.T) {
	content := redNowRead(t, redNowAuditorPath)

	live := func(c string) error {
		if _, err := redNowExtractSentinelSpan(c); err != nil {
			return fmt.Errorf("sentinel span: %w", err)
		}
		report, err := redNowExtractSectionSpan(c, redNowMustPassHeading)
		if err != nil {
			return err
		}
		if !strings.Contains(report, redNowMustPassMP8Prefix) {
			return fmt.Errorf("report template carries no %q row", redNowMustPassMP8Prefix)
		}
		return nil
	}

	if err := live(content); err != nil {
		t.Fatalf("MP-8 liveness anchors absent on the live file: %v", err)
	}

	// Mutant: delete the MP-8 report row.
	var kept []string
	deleted := 0
	for _, ln := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), redNowMustPassMP8Prefix) {
			deleted++
			continue
		}
		kept = append(kept, ln)
	}
	if deleted == 0 {
		t.Fatalf("mutation was a no-op — no %q row to delete", redNowMustPassMP8Prefix)
	}
	if err := live(strings.Join(kept, "\n")); err == nil {
		t.Errorf("mutant survived: MP-8's report row was deleted and the liveness anchors still passed")
	}
}

// TestMP8SentinelMutantsAreDetected covers mutant M-2 in both directions: a
// zero-pair copy and a two-pair copy must each be REJECTED before any
// comparison, so an anchor matching nothing cannot become a vacuous pass.
func TestMP8SentinelMutantsAreDetected(t *testing.T) {
	content := redNowRead(t, redNowAuditorPath)
	if _, err := redNowExtractSentinelSpan(content); err != nil {
		t.Fatalf("baseline extraction broken: %v", err)
	}

	zero := strings.ReplaceAll(strings.ReplaceAll(content, redNowBeginSentinel, ""), redNowEndSentinel, "")
	if _, err := redNowExtractSentinelSpan(zero); err == nil {
		t.Errorf("zero-pair mutant survived: extraction must fail when the sentinels are absent")
	}

	two := content + "\n" + redNowBeginSentinel + "\nplanted\n" + redNowEndSentinel + "\n"
	if _, err := redNowExtractSentinelSpan(two); err == nil {
		t.Errorf("two-pair mutant survived: extraction must fail when the sentinels match twice")
	}

	empty := strings.Replace(content, redNowBeginSentinel, redNowBeginSentinel+"\n"+redNowEndSentinel, 1)
	empty = strings.Replace(empty, "\n"+redNowEndSentinel+"\n"+redNowEndSentinel, "\n"+redNowEndSentinel, 1)
	if _, err := redNowExtractSentinelSpan(empty); err == nil {
		t.Errorf("empty-span mutant survived: extraction must fail on an empty span")
	}
}

// ===========================================================================
// L1 — the form check over fixtures (AC-RNT-008, -009a, -009b)
// ===========================================================================

// TestRedNowViolatingFixtureIsReported covers AC-RNT-009a.
func TestRedNowViolatingFixtureIsReported(t *testing.T) {
	forbidden := redNowForbiddenMetachars(t, redNowMP8Span(t))
	content := redNowFixture(t, "violating")

	elements := redNowElementFindings(content)
	if len(elements) == 0 {
		t.Errorf("violating fixture: the prose-only release-blocking RED cell was not reported")
	}
	form := redNowFormFindings(content, forbidden)
	if len(form) == 0 {
		t.Errorf("violating fixture: the piped command was not reported")
	}
	t.Logf("violating fixture findings: elements=%v form=%v", elements, form)
}

// TestRedNowLegitimateFixtureIsClean covers AC-RNT-009b. Confirming only the
// fail direction is indistinguishable from a check that reports everything.
func TestRedNowLegitimateFixtureIsClean(t *testing.T) {
	forbidden := redNowForbiddenMetachars(t, redNowMP8Span(t))
	content := redNowFixture(t, "legitimate")

	if got := redNowElementFindings(content); len(got) != 0 {
		t.Errorf("legitimate fixture reported element findings: %v", got)
	}
	if got := redNowFormFindings(content, forbidden); len(got) != 0 {
		t.Errorf("legitimate fixture reported form findings: %v", got)
	}
	if n := len(redNowCollectCommands(content)); n == 0 {
		t.Errorf("legitimate fixture yielded zero commands — a clean report over an empty set asserts nothing")
	}
}

// TestCommandScopeIsCarrierIndependent covers AC-RNT-008 and observes mutant
// M-5. The ledger fixture carries its malformed command in NO table cell, so
// the deliberately narrowed cell-scoped predicate finds nothing to check. The
// two predicates are run against the same input and observed to diverge.
func TestCommandScopeIsCarrierIndependent(t *testing.T) {
	forbidden := redNowForbiddenMetachars(t, redNowMP8Span(t))
	content := redNowFixture(t, "ledger")

	carrierIndependent := redNowFormFindings(content, forbidden)
	if len(carrierIndependent) == 0 {
		t.Fatalf("carrier-independent scope missed the malformed ledger command")
	}

	var cellScoped []string
	for _, c := range redNowCollectCellCommandsOnly(content) {
		if hits := redNowFormViolations(c.Command, forbidden); len(hits) > 0 {
			cellScoped = append(cellScoped, c.Command)
		}
	}
	if len(cellScoped) != 0 {
		t.Fatalf("fixture no longer isolates the mutant: the cell-scoped predicate found %v", cellScoped)
	}
	t.Logf("M-5 observed: cell-scoped=0 findings, carrier-independent=%d findings %v",
		len(carrierIndependent), carrierIndependent)

	// The carriers themselves are asserted so a scan that silently stopped
	// recognising one of them would surface here.
	seen := map[string]int{}
	for _, c := range redNowCollectCommands(content) {
		seen[c.Carrier]++
	}
	if seen["ledger-entry"] == 0 {
		t.Errorf("ledger carrier not recognised at all: %v", seen)
	}
}

// TestClassLaunderingMutantIsDetected observes mutant M-4: a malformed command
// moved into a regression-guard criterion. The class-scoped predicate misses it
// by construction; the class-independent one — the one this SPEC adopts —
// reports it. Both are run on the same input.
func TestClassLaunderingMutantIsDetected(t *testing.T) {
	forbidden := redNowForbiddenMetachars(t, redNowMP8Span(t))
	content := redNowFixture(t, "violating")

	classIndependent := redNowFormFindings(content, forbidden)
	if len(classIndependent) == 0 {
		t.Fatalf("class-independent scope missed the laundered command")
	}

	// The narrowed predicate: only commands cited by a release-blocking row.
	ledger := redNowParseLedger(content)
	var classScoped []string
	for _, row := range redNowParseRows(content) {
		if !redNowIsReleaseBlocking(row.Class) {
			continue
		}
		for _, ref := range redNowLedgerRefRe.FindAllString(row.RedCell, -1) {
			e, ok := ledger[ref]
			if !ok {
				continue
			}
			if hits := redNowFormViolations(e.Command, forbidden); len(hits) > 0 {
				classScoped = append(classScoped, e.Command)
			}
		}
	}
	if len(classScoped) != 0 {
		t.Fatalf("fixture no longer isolates the mutant: the class-scoped predicate found %v", classScoped)
	}
	t.Logf("M-4 observed: class-scoped=0 findings, class-independent=%d findings %v",
		len(classIndependent), classIndependent)
}

// TestRedNowFormCheckDivergesOnQuoting is the rule §5 audit-verification form:
// a quoted `|` is a literal and must NOT be refused, while an unquoted one must
// be. Without observing both directions, a checker that refuses everything and
// a checker that refuses nothing are indistinguishable.
func TestRedNowFormCheckDivergesOnQuoting(t *testing.T) {
	forbidden := redNowForbiddenMetachars(t, redNowMP8Span(t))
	cases := []struct {
		cmd  string
		want bool
	}{
		{`grep -c "^| \*\*AC-RNT-" acceptance.md`, false},
		{`grep -c 'a|b' file.md`, false},
		{`grep -c alpha file.md | wc -l`, true},
		{`ls target.txt && echo ok`, true},
		{`ls target.txt ; echo done`, true},
		{`grep -c alpha file.md > out.txt`, true},
		{`ls internal/spec/red_now_cell_test.go`, false},
	}
	for _, tc := range cases {
		got := len(redNowFormViolations(tc.cmd, forbidden)) > 0
		if got != tc.want {
			t.Errorf("form check on %q = %v, want %v (forbidden=%v)", tc.cmd, got, tc.want, forbidden)
		}
	}
}

// ===========================================================================
// mirrors and neutrality (AC-RNT-010, -011, -012)
// ===========================================================================

// TestMP8MirrorSpanIsByteEqual covers AC-RNT-010. The pair legitimately differs
// elsewhere (expected neutralization), so the assertion is span-scoped rather
// than whole-file.
func TestMP8MirrorSpanIsByteEqual(t *testing.T) {
	local := redNowMustExtractSpan(t, redNowAuditorPath)
	mirror := redNowMustExtractSpan(t, redNowAuditorMirrorPath)
	if local != mirror {
		t.Errorf("MP-8 span drifted between carriers\nlocal:\n%s\n\nmirror:\n%s", local, mirror)
	}
	mirrorContent := redNowRead(t, redNowAuditorMirrorPath)
	report := redNowMustExtractSection(t, mirrorContent, redNowMustPassHeading)
	if !strings.Contains(report, redNowMustPassMP8Prefix) {
		t.Errorf("mirror report template carries no %q row", redNowMustPassMP8Prefix)
	}
}

// TestRuleMirrorIsByteIdentical covers REQ-RNT-010 for the rule pair, which was
// byte-identical before this work and must stay so.
func TestRuleMirrorIsByteIdentical(t *testing.T) {
	if redNowRead(t, redNowRulePath) != redNowRead(t, redNowRuleMirrorPath) {
		t.Errorf("%s and %s are not byte-identical", redNowRulePath, redNowRuleMirrorPath)
	}
}

// TestMP8MirrorSpanIsNeutral covers AC-RNT-011. The scope is the ADDED clause,
// not the whole file: the mirror already carries illustrative SPEC-AUTH-001
// placeholders, which are neutral by construction.
func TestMP8MirrorSpanIsNeutral(t *testing.T) {
	span := redNowMustExtractSpan(t, redNowAuditorMirrorPath)
	specID := regexp.MustCompile(`SPEC-([A-Z][A-Z0-9]+-)+[0-9]+`)
	if hits := specID.FindAllString(span, -1); len(hits) != 0 {
		t.Errorf("mirrored MP-8 span carries SPEC identifiers: %v", hits)
	}
	if strings.Contains(span, "t343") {
		t.Errorf("mirrored MP-8 span carries a card id")
	}
	if hits := regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`).FindAllString(span, -1); len(hits) != 0 {
		t.Errorf("mirrored MP-8 span carries an internal date: %v", hits)
	}
}

// TestRedNowArtifactsDoNotShip covers AC-RNT-012.
func TestRedNowArtifactsDoNotShip(t *testing.T) {
	root := filepath.Join(repoRoot(t), "internal/template/templates")
	var hits []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.Contains(filepath.ToSlash(path), "red_now") {
			hits = append(hits, path)
			return nil
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		if strings.Contains(string(raw), "red_now") {
			hits = append(hits, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk template tree: %v", err)
	}
	sort.Strings(hits)
	if len(hits) != 0 {
		t.Errorf("the repository-local test or its fixtures leaked into the template tree: %v", hits)
	}
}
