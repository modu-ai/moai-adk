// Package harness — release-update 구조 검사 (SPEC-RELUP-STRUCT-AUDIT-001).
//
// Structure-audit test family for the three dev-only release-update harness
// surfaces, read directly from the working tree:
//
//	S1  specialist Phase 3 URL list      (M1, M8)   — AC-RSA-002
//	S2  Phase 3 / Phase 4 boundaries     (M2)       — AC-RSA-003
//	S3  codex paragraph block            (M3)       — AC-RSA-004
//	S4  run.js watch-norm sentence       (M4)       — AC-RSA-005
//	S5  best-practices faces             (M5, M6)   — AC-RSA-006
//	S6  manifest top-level domain        (M7)       — AC-RSA-007
//
// Design constraints (spec.md §1.2 decisions D1-D8, REQ-RSA-001..012):
//   - TEST-ONLY file: every helper here is unexported and consumed by no
//     production code (D2).
//   - Probes are pure in-memory string mutations of the real working-tree
//     content — no temp files, no subprocess, no VCS invocation of any form
//     (D3, REQ-RSA-008). The file carries no exec channel at all.
//   - Anchors are content anchors, never line-number constants (D6).
//   - Duplicate handling is two-class (D8): class E exactly-once for the named
//     boundary markers (### Phase 3 / ### Phase 4 — 0 or 2+ occurrences FAIL
//     the check), class F first-match selection for the BP heading (the first
//     match selects the block; duplicates neither fail nor re-select).
//   - S1 clears HTML comment spans FIRST and preserves fence content (P1
//     repair); S5 strips comments and fences before matching.
//
// Literals pinned by plan.md §D1 are copied verbatim from the SPEC
// (REQ-RSA-005 sentence, S6 clause literals). The file deliberately avoids the
// forbidden internal/ literal guarded by AC-RSA-010 and any version-control
// command spelling guarded by AC-RSA-009.
package harness

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// relupWatchNormSentence is the whole watch-norm literal pinned by
// REQ-RSA-005 (plan §D1 S4) — matched whitespace-normalized, exactly once.
const relupWatchNormSentence = `Alpha-window themes remain WATCH-LIST observations — adoption is judged only when the theme lands in a stable release; an alpha-window theme is never reported as adopted drift (REQ-RDX-009).`

// relupDocFacets are the six Phase 3 documentation facets (REQ-RSA-002).
var relupDocFacets = []string{"hooks", "sub-agents", "skills", "plugins", "mcp", "settings"}

// relupCodexTokens are the six in-boundary tokens of S3 (REQ-RSA-004):
// the default-sentence pair first, then the four key-family tokens.
var relupCodexTokens = []string{
	"rust-v0.161.0",
	"emit warning",
	"last_analyzed_version",
	"last_analyzed_date",
	"last_master_research",
	"analysis_history[]",
}

// Working-tree paths of the three audit targets (REQ-RSA-008 — direct reads).
const (
	relupSpecialistPath = ".claude/agents/harness/hns-release-update-specialist.md"
	relupRunJSPath      = ".claude/workflows/hns-release-update-run.js"
	relupManifestPath   = ".claude/commands/harness/release-update/manifest.json"
)

// relupModuleRoot resolves the module root from this test file's own source
// path (runtime.Caller — CWD-independent, so `go test ./internal/harness/`
// works from any working directory).
func relupModuleRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot resolve module root")
	}
	// thisFile = <root>/internal/harness/relup_structure_audit_test.go
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root %q has no go.mod: %v", root, err)
	}
	return root
}

// relupReadTarget reads one audit target from the working tree (read-only).
func relupReadTarget(t *testing.T, relPath string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(relupModuleRoot(t), filepath.FromSlash(relPath)))
	if err != nil {
		t.Fatalf("read %s: %v", relPath, err)
	}
	return string(data)
}

// relupMaskScope selects which markdown spans the mask primitive processes
// (REQ-RSA-001(a) — comments and fences independently selectable).
type relupMaskScope struct {
	comments bool
	fences   bool
}

// relupMaskLines returns the content split into lines with the selected span
// kinds masked. Line positions are always preserved (masked lines become
// empty) so boundary detection never sees fence- or comment-carried text.
// Semantics:
//   - An HTML comment span is recognized only OUTSIDE a code fence (P1:
//     fence content is preserved verbatim when fences are not masked).
//   - A comment span opened outside a fence swallows everything up to its
//     closing marker, fence markers included (M8 sealing).
//   - Fence delimiter lines and fence interior lines are masked iff
//     scope.fences.
func relupMaskLines(content string, scope relupMaskScope) []string {
	lines := strings.Split(content, "\n")
	out := make([]string, len(lines))
	inComment := false
	inFence := false
	for i, line := range lines {
		switch {
		case inComment:
			// Comment spans are cleared FIRST wherever they are recognized;
			// text on the closing line after the marker survives.
			end := strings.Index(line, "-->")
			if end < 0 {
				out[i] = strings.Repeat(" ", len(line))
			} else {
				out[i] = strings.Repeat(" ", end+3) + line[end+3:]
				inComment = false
			}
		case inFence:
			if scope.fences {
				out[i] = ""
			} else {
				out[i] = line
			}
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inFence = false
			}
		default:
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inFence = true
				if scope.fences {
					out[i] = ""
				} else {
					out[i] = line
				}
				continue
			}
			if !scope.comments {
				out[i] = line
				continue
			}
			// Span-precise comment masking on a fresh line.
			var b strings.Builder
			rest := line
			for {
				start := strings.Index(rest, "<!--")
				if start < 0 {
					b.WriteString(rest)
					break
				}
				b.WriteString(rest[:start])
				rest = rest[start:]
				end := strings.Index(rest, "-->")
				if end < 0 {
					b.WriteString(strings.Repeat(" ", len(rest)))
					inComment = true
					break
				}
				b.WriteString(strings.Repeat(" ", end+3))
				rest = rest[end+3:]
			}
			out[i] = b.String()
		}
	}
	return out
}

// relupExactlyOnceLineStart implements the class E contract (D8): the named
// line-start boundary marker must occur EXACTLY once; zero or 2+ occurrences
// are an error (never silently resolved to a first occurrence).
func relupExactlyOnceLineStart(lines []string, prefix string) (int, error) {
	hits := 0
	at := -1
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			hits++
			at = i
		}
	}
	if hits != 1 {
		return -1, fmt.Errorf("class E exactly-once violated for %q: %d line-start occurrences (want exactly 1)", prefix, hits)
	}
	return at, nil
}

// relupFirstLineStart implements the class F contract (D8): the selection
// anchor resolves to its FIRST occurrence; duplicates neither fail nor
// re-select. Returns -1 when absent.
func relupFirstLineStart(lines []string, prefix string) int {
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return i
		}
	}
	return -1
}

// relupHeadingLevel counts the leading '#' of a line-start heading.
func relupHeadingLevel(line string) int {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	return n
}

// relupCountWholeToken counts whole-token occurrences: an occurrence counts
// only when the byte immediately before and immediately after it is not an
// identifier character (REQ-RSA-001(c) — whole-token matching).
func relupCountWholeToken(text, token string) int {
	count := 0
	for i := 0; i+len(token) <= len(text); {
		j := strings.Index(text[i:], token)
		if j < 0 {
			break
		}
		at := i + j
		before := byte(' ')
		if at > 0 {
			before = text[at-1]
		}
		after := byte(' ')
		if at+len(token) < len(text) {
			after = text[at+len(token)]
		}
		if !relupIsWordByte(before) && !relupIsWordByte(after) {
			count++
		}
		i = at + len(token)
	}
	return count
}

func relupIsWordByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	case b == '_':
		return true
	}
	return false
}

// relupPhase3Boundaries is the S2 check (REQ-RSA-003): the Phase 3 block
// boundaries are the line-start headings `### Phase 3` and `### Phase 4`,
// each required exactly once OUTSIDE fences and comments (class E — D8).
// Returns the [start, end) line indices of the block.
func relupPhase3Boundaries(content string) (int, int, error) {
	masked := relupMaskLines(content, relupMaskScope{comments: true, fences: true})
	p3, err := relupExactlyOnceLineStart(masked, "### Phase 3")
	if err != nil {
		return -1, -1, err
	}
	p4, err := relupExactlyOnceLineStart(masked, "### Phase 4")
	if err != nil {
		return -1, -1, err
	}
	if p4 <= p3 {
		return -1, -1, fmt.Errorf("Phase 4 heading (line %d) does not follow Phase 3 heading (line %d)", p4, p3)
	}
	return p3, p4, nil
}

// relupCanonicalDocFacet parses one list item as a structural URL and reports
// the facet it names, ONLY when every structural assertion of REQ-RSA-002
// holds: scheme https, host code.claude.com, path exactly /docs/en/<facet>,
// empty query, empty fragment, termination at end-of-item.
func relupCanonicalDocFacet(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	if u.Scheme != "https" || u.Host != "code.claude.com" {
		return "", false
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", false
	}
	const prefix = "/docs/en/"
	if !strings.HasPrefix(u.Path, prefix) {
		return "", false
	}
	switch facet := strings.TrimPrefix(u.Path, prefix); facet {
	case "hooks", "sub-agents", "skills", "plugins", "mcp", "settings":
		return facet, true
	}
	return "", false
}

// relupPhase3URLFacetCounts is the S1 check (REQ-RSA-002): it clears HTML
// comment spans FIRST (fence content preserved — P1), bounds the Phase 3
// block on class E line-start headings (S2), then parses EACH list item of
// the block as a structural URL. Facet counts come from comment-cleared
// items only, so a comment-wrapped list (M8) and a query-wrapped item (M1)
// both yield 0.
func relupPhase3URLFacetCounts(content string) (map[string]int, error) {
	counts := make(map[string]int, len(relupDocFacets))
	p3, p4, err := relupPhase3Boundaries(content)
	if err != nil {
		return nil, err
	}
	// Content mode: comments cleared, fence content preserved (P1).
	contentLines := relupMaskLines(content, relupMaskScope{comments: true})
	for _, line := range contentLines[p3+1 : p4] {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if facet, ok := relupCanonicalDocFacet(raw); ok {
			counts[facet]++
		}
	}
	return counts, nil
}

// relupCodexBlockBounds bounds the codex paragraph (REQ-RSA-004): from the
// line-start anchor `**Codex axis state` to its terminating blank line, or
// the next line-start heading, whichever comes first. Returns [start, end)
// line indices.
func relupCodexBlockBounds(lines []string) (int, int, error) {
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "**Codex axis state") {
			start = i
			break
		}
	}
	if start < 0 {
		return -1, -1, fmt.Errorf("codex paragraph anchor %q not found at line start", "**Codex axis state")
	}
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" || strings.HasPrefix(lines[i], "#") {
			return start, i, nil
		}
	}
	return start, len(lines), nil
}

// relupCodexTokenCounts is the S3 check (REQ-RSA-004): the six tokens are
// counted with whole-token matching, INSIDE the bounded block only.
func relupCodexTokenCounts(content string) (map[string]int, error) {
	lines := strings.Split(content, "\n")
	s, e, err := relupCodexBlockBounds(lines)
	if err != nil {
		return nil, err
	}
	block := strings.Join(lines[s:e], "\n")
	counts := make(map[string]int, len(relupCodexTokens))
	for _, tok := range relupCodexTokens {
		counts[tok] = relupCountWholeToken(block, tok)
	}
	return counts, nil
}

// relupWatchNormCount is the S4 check (REQ-RSA-005): inside the theme
// checklist block (from the `// Standing 6-theme` line through the first
// `];` line at or after the const declaration), comment prefixes are
// stripped, lines are joined and whitespace-normalized, and the ENTIRE
// watch-norm sentence must match exactly once. Token-only presence counts
// nothing (M4).
func relupWatchNormCount(content string) (int, error) {
	lines := strings.Split(content, "\n")
	start := relupFirstLineStart(lines, "// Standing 6-theme")
	if start < 0 {
		return 0, fmt.Errorf("theme checklist opener %q not found", "// Standing 6-theme")
	}
	constLine := -1
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "const CODEX_THEME_CHECKLIST") {
			constLine = i
			break
		}
	}
	if constLine < 0 {
		return 0, fmt.Errorf("const CODEX_THEME_CHECKLIST declaration not found after line %d", start)
	}
	closeIdx := -1
	for i := constLine; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "];") {
			closeIdx = i
			break
		}
	}
	if closeIdx < 0 {
		return 0, fmt.Errorf("theme checklist closing `];` not found after line %d", constLine)
	}
	parts := make([]string, 0, closeIdx-start+1)
	for _, line := range lines[start : closeIdx+1] {
		t := strings.TrimSpace(line)
		t = strings.TrimPrefix(t, "//")
		parts = append(parts, strings.TrimSpace(t))
	}
	normalized := strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
	return strings.Count(normalized, relupWatchNormSentence), nil
}

// relupBPFaces is the S5 check (REQ-RSA-006) on comment- and fence-stripped
// text (matching mode). Face (a): case-insensitive line-start heading
// matching `best-practice`, >= 1. Face (b): the deliverable literal inside
// the BP block — the FIRST matching heading selects the block (class F
// first-match, D8), which ends at the next same-or-higher heading. Duplicate
// BP headings neither fail nor re-select. Comment- or fence-carried text
// counts zero in both faces.
func relupBPFaces(content string) (headings, deliverables int) {
	lines := relupMaskLines(content, relupMaskScope{comments: true, fences: true})
	first := -1
	firstLevel := 0
	for i, line := range lines {
		if !strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(strings.ToLower(line), "best-practice") {
			continue
		}
		headings++
		if first < 0 {
			first = i
			firstLevel = relupHeadingLevel(line)
		}
	}
	if first < 0 {
		return headings, 0 // class F: nothing selected — deliverable face is 0
	}
	for i := first; i < len(lines); i++ {
		if i > first && strings.HasPrefix(lines[i], "#") && relupHeadingLevel(lines[i]) <= firstLevel {
			break
		}
		deliverables += strings.Count(lines[i], "HTML proposal report")
	}
	return headings, deliverables
}

// relupManifestDomainClauses is the S6 check (REQ-RSA-007): manifest.json is
// DECODED as JSON and only the top-level `domain` string is asserted against
// the three clause literals. Other keys carry no weight (M7).
func relupManifestDomainClauses(content string) (int, error) {
	var manifest struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return 0, fmt.Errorf("manifest is not decodable JSON: %w", err)
	}
	clauses := 0
	for _, c := range []string{"Claude Code", "Codex CLI upstream change tracking", "best-practices axis"} {
		if strings.Contains(manifest.Domain, c) {
			clauses++
		}
	}
	return clauses, nil
}

// relupFacetCount reads one facet out of the S1 count map (zero default).
func relupFacetCount(counts map[string]int, facet string) int {
	return counts[facet]
}

// TestRelupStructurePhase3URLs covers S1 (AC-RSA-002 — M1, M8) and S2
// (AC-RSA-003 — M2) against the working-tree specialist file.
func TestRelupStructurePhase3URLs(t *testing.T) {
	t.Parallel()
	real := relupReadTarget(t, relupSpecialistPath)

	t.Run("positive_six_facets_each_exactly_1", func(t *testing.T) {
		t.Parallel()
		counts, err := relupPhase3URLFacetCounts(real)
		if err != nil {
			t.Fatalf("S1 check failed on the working-tree file: %v", err)
		}
		for _, facet := range relupDocFacets {
			if got := relupFacetCount(counts, facet); got != 1 {
				t.Errorf("facet %q: got %d, want exactly 1", facet, got)
			}
		}
	})

	t.Run("positive_phase3_phase4_exactly_once", func(t *testing.T) {
		t.Parallel()
		p3, p4, err := relupPhase3Boundaries(real)
		if err != nil {
			t.Fatalf("S2 check failed on the working-tree file: %v", err)
		}
		if p3 < 0 || p4 <= p3 {
			t.Fatalf("degenerate boundaries: p3=%d p4=%d", p3, p4)
		}
	})

	t.Run("M1_next_param_wrap_yields_0", func(t *testing.T) {
		t.Parallel()
		const needle = "https://code.claude.com/docs/en/hooks\n"
		if !strings.Contains(real, needle) {
			t.Fatalf("fixture drift: Phase 3 hooks URL not found verbatim")
		}
		variant := strings.Replace(real, needle,
			"https://code.claude.com/docs/en/hooks?next=https://evil.example.com\n", 1)
		counts, err := relupPhase3URLFacetCounts(variant)
		if err != nil {
			t.Fatalf("S1 check returned an unexpected error: %v", err)
		}
		if got := relupFacetCount(counts, "hooks"); got != 0 {
			t.Errorf("query-wrapped hooks item counted %d times, want 0", got)
		}
		for _, facet := range relupDocFacets {
			if facet == "hooks" {
				continue
			}
			if got := relupFacetCount(counts, facet); got != 1 {
				t.Errorf("untouched facet %q: got %d, want 1", facet, got)
			}
		}
	})

	t.Run("M8_comment_wrapped_url_list_yields_0", func(t *testing.T) {
		t.Parallel()
		fenceOpen := "```\nParallel fetch (single message):"
		fenceClose := "https://code.claude.com/docs/en/settings\n```"
		if !strings.Contains(real, fenceOpen) || !strings.Contains(real, fenceClose) {
			t.Fatalf("fixture drift: Phase 3 URL fence not found verbatim")
		}
		variant := strings.Replace(real, fenceOpen, "<!--\n"+fenceOpen, 1)
		variant = strings.Replace(variant, fenceClose, fenceClose+"\n-->", 1)
		counts, err := relupPhase3URLFacetCounts(variant)
		if err != nil {
			t.Fatalf("S1 check returned an unexpected error: %v", err)
		}
		for _, facet := range relupDocFacets {
			if got := relupFacetCount(counts, facet); got != 0 {
				t.Errorf("comment-wrapped facet %q counted %d, want 0", facet, got)
			}
		}
	})

	t.Run("M2_cross_reference_variants_do_not_move_boundary", func(t *testing.T) {
		t.Parallel()
		if _, _, err := relupPhase3Boundaries(real); err != nil {
			t.Fatalf("baseline S2 failed: %v", err)
		}
		variants := map[string]string{
			"fence-internal": "\n```\n### Phase 3 (cross-reference trap)\n```\n",
			"indented":       "\n   ### Phase 3 (indented prose mention)\n",
			"backticked":     "\n`### Phase 3` quoted in prose\n",
		}
		for name, insertion := range variants {
			variant := relupInsertAfterFirstLineStart(real, "### Phase 3", insertion)
			p3, p4, err := relupPhase3Boundaries(variant)
			if err != nil {
				t.Fatalf("%s: boundary check errored: %v", name, err)
			}
			lines := strings.Split(variant, "\n")
			realHeading := -1
			realHeading4 := -1
			for i, line := range lines {
				switch {
				case realHeading < 0 && strings.HasPrefix(line, "### Phase 3 —"):
					realHeading = i
				case realHeading4 < 0 && strings.HasPrefix(line, "### Phase 4"):
					realHeading4 = i
				}
			}
			if p3 != realHeading {
				t.Errorf("%s: start boundary moved — p3=%d, real heading at %d", name, p3, realHeading)
			}
			if p4 != realHeading4 {
				t.Errorf("%s: end boundary moved — p4=%d, real Phase 4 heading at %d", name, p4, realHeading4)
			}
			counts, err := relupPhase3URLFacetCounts(variant)
			if err != nil {
				t.Fatalf("%s: S1 check errored: %v", name, err)
			}
			for _, facet := range relupDocFacets {
				if got := relupFacetCount(counts, facet); got != 1 {
					t.Errorf("%s: facet %q: got %d, want 1 (real URL must stay inside the block)", name, facet, got)
				}
			}
		}
	})

	t.Run("M2_duplicate_phase3_heading_fails_check", func(t *testing.T) {
		t.Parallel()
		variant := relupInsertAfterFirstLineStart(real, "### Phase 3",
			"### Phase 3 — Cross-Reference Official Docs (duplicate boundary marker)\n")
		if _, _, err := relupPhase3Boundaries(variant); err == nil {
			t.Fatal("duplicate line-start Phase 3 heading did not fail the check — class E exactly-once violated")
		}
		if counts, err := relupPhase3URLFacetCounts(variant); err == nil {
			t.Fatalf("S1 check must propagate the class E failure, got counts %v", counts)
		}
	})
}

// relupInsertAfterFirstLineStart inserts `insertion` immediately after the
// first line starting with `prefix` (fixture builder for M-probe variants).
func relupInsertAfterFirstLineStart(content, prefix, insertion string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines)+8)
	inserted := false
	for _, line := range lines {
		out = append(out, line)
		if !inserted && strings.HasPrefix(line, prefix) {
			out = append(out, strings.Split(insertion, "\n")...)
			inserted = true
		}
	}
	if !inserted {
		panic("relupInsertAfterFirstLineStart: anchor " + prefix + " not found")
	}
	return strings.Join(out, "\n")
}

// TestRelupStructureCodexParagraph covers S3 (AC-RSA-004 — M3) against the
// working-tree specialist file.
func TestRelupStructureCodexParagraph(t *testing.T) {
	t.Parallel()
	real := relupReadTarget(t, relupSpecialistPath)

	t.Run("positive_six_tokens_inside_boundary", func(t *testing.T) {
		t.Parallel()
		counts, err := relupCodexTokenCounts(real)
		if err != nil {
			t.Fatalf("S3 check failed on the working-tree file: %v", err)
		}
		for _, tok := range relupCodexTokens {
			if got := counts[tok]; got < 1 {
				t.Errorf("token %q: got %d, want >= 1 inside the bounded block", tok, got)
			}
		}
		// BLK-014G measured vector (plan §F M2 — recorded from this run).
		t.Logf("measured in-boundary vector %v", counts)
	})

	t.Run("M3_appended_paragraph_after_block_not_counted", func(t *testing.T) {
		t.Parallel()
		realCounts, err := relupCodexTokenCounts(real)
		if err != nil {
			t.Fatalf("baseline S3 failed: %v", err)
		}
		lines := strings.Split(real, "\n")
		_, end, err := relupCodexBlockBounds(lines)
		if err != nil {
			t.Fatalf("baseline bounds failed: %v", err)
		}
		var out []string
		out = append(out, lines[:end+1]...)
		out = append(out, "Trailing paragraph outside the block: default `since_codex = \"rust-v0.161.0\"`, emit warning (appended).")
		out = append(out, lines[end+1:]...)
		variant := strings.Join(out, "\n")

		variantCounts, err := relupCodexTokenCounts(variant)
		if err != nil {
			t.Fatalf("S3 check errored on the appended variant: %v", err)
		}
		for _, tok := range relupCodexTokens {
			if variantCounts[tok] != realCounts[tok] {
				t.Errorf("token %q: appended copy leaked into the block (%d -> %d) — terminating blank-line boundary breached",
					tok, realCounts[tok], variantCounts[tok])
			}
		}
	})

	t.Run("M3_default_sentence_moved_out_yields_0", func(t *testing.T) {
		t.Parallel()
		lines := strings.Split(real, "\n")
		start, end, err := relupCodexBlockBounds(lines)
		if err != nil {
			t.Fatalf("baseline bounds failed: %v", err)
		}
		moveIdx := -1
		for i := start; i < end; i++ {
			if strings.Contains(lines[i], `since_codex = "rust-v0.161.0"`) {
				moveIdx = i
				break
			}
		}
		if moveIdx < 0 {
			t.Fatal("fixture drift: default-sentence line not found inside the codex block")
		}
		var out []string
		out = append(out, lines[:moveIdx]...)
		out = append(out, lines[moveIdx+1:end+1]...) // block minus the sentence, through the blank terminator
		out = append(out, lines[moveIdx])            // the sentence, now OUTSIDE the boundary
		out = append(out, lines[end+1:]...)
		variant := strings.Join(out, "\n")

		counts, err := relupCodexTokenCounts(variant)
		if err != nil {
			t.Fatalf("S3 check errored on the moved variant: %v", err)
		}
		for _, tok := range []string{"rust-v0.161.0", "emit warning"} {
			if got := counts[tok]; got != 0 {
				t.Errorf("token %q moved outside the boundary still counted %d, want 0", tok, got)
			}
		}
	})
}

// TestRelupStructureWatchNormSentence covers S4 (AC-RSA-005 — M4) against
// the working-tree Runner file.
func TestRelupStructureWatchNormSentence(t *testing.T) {
	t.Parallel()
	real := relupReadTarget(t, relupRunJSPath)

	t.Run("positive_sentence_exactly_1", func(t *testing.T) {
		t.Parallel()
		got, err := relupWatchNormCount(real)
		if err != nil {
			t.Fatalf("S4 check failed on the working-tree file: %v", err)
		}
		if got != 1 {
			t.Errorf("watch-norm sentence matched %d times, want exactly 1", got)
		}
	})

	t.Run("M4_tokens_only_without_sentence_yields_0", func(t *testing.T) {
		t.Parallel()
		sentenceSpan := "// observed PR numbers and the MoAI exposure surface. Alpha-window themes remain\n" +
			"// WATCH-LIST observations — adoption is judged only when the theme lands in a\n" +
			"// stable release; an alpha-window theme is never reported as adopted drift\n" +
			"// (REQ-RDX-009)."
		tokensOnly := "// observed PR numbers and the MoAI exposure surface. Themes stay on the WATCH-LIST\n" +
			"// until a stable release lands; only then is adoption judged (REQ-RDX-009)."
		if !strings.Contains(real, sentenceSpan) {
			t.Fatal("fixture drift: watch-norm comment span not found verbatim")
		}
		variant := strings.Replace(real, sentenceSpan, tokensOnly, 1)
		got, err := relupWatchNormCount(variant)
		if err != nil {
			t.Fatalf("S4 check errored on the token-only variant: %v", err)
		}
		if got != 0 {
			t.Errorf("token-only variant matched the whole sentence %d times, want 0", got)
		}
	})
}

// TestRelupStructureBPFaces covers S5 (AC-RSA-006 — M5, M6, class F) against
// the working-tree specialist file.
func TestRelupStructureBPFaces(t *testing.T) {
	t.Parallel()
	real := relupReadTarget(t, relupSpecialistPath)

	t.Run("positive_heading_ge1_and_deliverable_ge1", func(t *testing.T) {
		t.Parallel()
		headings, deliverables := relupBPFaces(real)
		if headings < 1 {
			t.Errorf("BP heading face: got %d, want >= 1", headings)
		}
		if deliverables < 1 {
			t.Errorf("BP deliverable face: got %d, want >= 1", deliverables)
		}
	})

	t.Run("M5_comment_carried_deliverable_yields_0", func(t *testing.T) {
		t.Parallel()
		const deliverableSpan = "the BP axis produces an HTML proposal report"
		if !strings.Contains(real, deliverableSpan) {
			t.Fatal("fixture drift: deliverable span not found verbatim")
		}
		variant := strings.Replace(real, deliverableSpan,
			"<!-- "+deliverableSpan+" -->", 1)
		headings, deliverables := relupBPFaces(variant)
		if deliverables != 0 {
			t.Errorf("comment-carried deliverable counted %d, want 0", deliverables)
		}
		if headings < 1 {
			t.Errorf("BP heading face must stay >= 1 on the deliverable-only mutant, got %d", headings)
		}
	})

	t.Run("M6_comment_carried_fake_heading_yields_0", func(t *testing.T) {
		t.Parallel()
		const realHeading = "## Best-Practices Axis (standing procedure — every sweep)"
		if !strings.Contains(real, realHeading) {
			t.Fatal("fixture drift: BP heading line not found verbatim")
		}
		commentWrapped := "<!-- " + realHeading + "\n## Best-Practices Axis (fake heading, comment-carried) -->"
		variant := strings.Replace(real, realHeading, commentWrapped, 1)
		headings, deliverables := relupBPFaces(variant)
		if headings != 0 {
			t.Errorf("comment-carried headings counted %d, want 0", headings)
		}
		if deliverables != 0 {
			t.Errorf("deliverable face with no selectable BP block counted %d, want 0", deliverables)
		}
	})

	t.Run("duplicate_bp_heading_first_match_no_reselect_no_fail", func(t *testing.T) {
		t.Parallel()
		// Class F (D8): a duplicate BP heading neither fails the check nor
		// re-selects — the FIRST heading still selects the block that holds
		// the deliverable, so the deliverable count stays exactly its
		// real-tree value (a last-match re-selection would drop it to 0).
		variant := relupInsertAfterFirstLineStart(real, "## Delegation Map",
			"## Best-Practices Axis (duplicate sweep heading)\n")
		headings, deliverables := relupBPFaces(variant)
		if headings != 2 {
			t.Errorf("duplicate variant heading count = %d, want 2 (duplicates must not fail)", headings)
		}
		_, realDeliverables := relupBPFaces(real)
		if deliverables != realDeliverables {
			t.Errorf("first-match selection broke: deliverable %d, want the real-tree %d", deliverables, realDeliverables)
		}
	})
}

// TestRelupStructureManifestDomain covers S6 (AC-RSA-007 — M7) against the
// working-tree manifest file.
func TestRelupStructureManifestDomain(t *testing.T) {
	t.Parallel()
	real := relupReadTarget(t, relupManifestPath)

	t.Run("positive_top_level_domain_three_clauses", func(t *testing.T) {
		t.Parallel()
		clauses, err := relupManifestDomainClauses(real)
		if err != nil {
			t.Fatalf("S6 check failed on the working-tree manifest: %v", err)
		}
		if clauses != 3 {
			t.Errorf("top-level domain carries %d/3 clause literals, want 3", clauses)
		}
	})

	t.Run("M7_same_line_trailing_key_yields_0", func(t *testing.T) {
		t.Parallel()
		const domainLine = `"domain": "moai-adk-go dev-only maintainer tooling — Claude Code + Codex CLI upstream change tracking and the best-practices axis",`
		if !strings.Contains(real, domainLine) {
			t.Fatal("fixture drift: manifest domain line not found verbatim")
		}
		// The historical broken check was a LINE pattern ("domain" .* clauses).
		// The mutant strips the clauses from the parsed domain value and moves
		// them into a same-line sibling key so that line pattern still passes.
		variantLine := `"domain": "moai-adk-go dev-only maintainer tooling", "note": "Claude Code + Codex CLI upstream change tracking and the best-practices axis",`
		variant := strings.Replace(real, domainLine, variantLine, 1)
		clauses, err := relupManifestDomainClauses(variant)
		if err != nil {
			t.Fatalf("variant manifest failed to decode: %v", err)
		}
		if clauses != 0 {
			t.Errorf("same-line trailing-key variant carries %d clauses on the parsed domain, want 0", clauses)
		}
	})
}
