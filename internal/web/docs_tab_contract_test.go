package web

// docs_tab_contract_test.go — SPEC-DOCS-TABCOUNT-DRIFT-001 M1 (card t530).
//
// The settings tab count and the tab name list are ONE fact, owned by
// consoleTabs(). Twelve documents currently write that fact by hand in twenty
// places, and four of them are already wrong. This guard is the machine that
// keeps the remaining sites honest once the run phase removes what can be
// removed.
//
// Three layers, each a subtest, because no single regex can decide "is this a
// settings-tab count?" (plan.md §A.3):
//
//	literals  (N0) — the exact strings enumerated in the plan artifact must be
//	                 gone from the twelve target files.
//	allowlist (N1) — a numeral adjacent to a word-bounded tab noun, on BOTH the
//	                 digit and the spelled-out-word axis, must hit nothing
//	                 outside one explicitly-scoped allowlist rule.
//	names     (N2) — the documented tab name lists must equal the labels the
//	                 console actually renders, in consoleTabs() order.
//
// The guard reads the repository from the package directory; go test runs with
// the package dir as the working directory, so the repo root is "../../"
// (same convention as restyle_test.go's osReadFile).

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

// repoRootRel is the repository root relative to internal/web.
const repoRootRel = "../../"

// docsTabTargetFiles is the explicit twelve-file scan scope (plan.md §A.2).
// The list lives in the code, not in a glob, so a reader can see WHAT is being
// guarded. Widening it is a deliberate edit, not an accident of a wildcard.
var docsTabTargetFiles = []string{
	"README.md",
	"README.ko.md",
	"README.ja.md",
	"README.zh.md",
	"docs-site/content/ko/cli-reference/web.md",
	"docs-site/content/en/cli-reference/web.md",
	"docs-site/content/ja/cli-reference/web.md",
	"docs-site/content/zh/cli-reference/web.md",
	"docs-site/content/ko/advanced/moai-web-console.md",
	"docs-site/content/en/advanced/moai-web-console.md",
	"docs-site/content/ja/advanced/moai-web-console.md",
	"docs-site/content/zh/advanced/moai-web-console.md",
}

// docsTabNameListFiles are the eight sites that carry a tab NAME list (N2).
// The four cli-reference pages count tabs but never name them.
var docsTabNameListFiles = []string{
	"README.md",
	"README.ko.md",
	"README.ja.md",
	"README.zh.md",
	"docs-site/content/ko/advanced/moai-web-console.md",
	"docs-site/content/en/advanced/moai-web-console.md",
	"docs-site/content/ja/advanced/moai-web-console.md",
	"docs-site/content/zh/advanced/moai-web-console.md",
}

// countLiteralsPath is the plan artifact holding the exact strings measured at
// the enumerated sites. Reading them from disk rather than restating them here
// keeps the enumeration the single source of truth.
const countLiteralsPath = ".moai/reports/t530/count-literals.txt"

// allowedLinesBaseline is the number of lines the single allowlist rule
// exempts, measured at base 1d150a27d: grep -ci codex over the four
// advanced/moai-web-console.md copies returned 4 each.
//
// Rule count and exemption surface are DIFFERENT numbers. One rule exempting
// sixteen lines is the residual risk spec.md §7 names, so the guard reports
// both: a rule set that stays at one while the exempted surface grows is
// exactly the change this baseline catches.
const allowedLinesBaseline = 16

// readRepoFile reads a repository-relative file. A read failure is fatal, never
// a skip: a skipped file leaves the guard green while guarding nothing
// (acceptance.md §D.1).
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRootRel, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("docs tab guard: cannot read %s: %v (a guard that cannot read its target is not passing, it is blind)", rel, err)
	}
	return string(b)
}

// TestDocsTabContract is the three-layer guard. Its subtest names are asserted
// literally by acceptance.md AC-TCD-004 — the parent test alone passing proves
// nothing, because a parent with no subtests also passes.
func TestDocsTabContract(t *testing.T) {
	t.Run("literals", func(t *testing.T) { docsTabLiterals(t) })
	t.Run("allowlist", func(t *testing.T) { docsTabAllowlist(t) })
	t.Run("names", func(t *testing.T) { docsTabNames(t) })
}

// ---------------------------------------------------------------------------
// N0 — literals
// ---------------------------------------------------------------------------

// docsTabLiterals asserts that none of the enumerated exact strings survives in
// the twelve target files, and reports how many files it actually READ.
//
// The swept counter increments after a successful read, never from len() of the
// declared slice: printing the length of a list proves a list was declared, not
// that a file was opened (acceptance.md AC-TCD-005).
func docsTabLiterals(t *testing.T) {
	literals := loadCountLiterals(t)

	type hit struct {
		file    string
		literal string
	}
	var hits []hit

	swept := 0
	for _, rel := range docsTabTargetFiles {
		content := readRepoFile(t, rel)
		swept++
		for _, lit := range literals {
			if strings.Contains(content, lit) {
				hits = append(hits, hit{file: rel, literal: lit})
			}
		}
	}

	t.Logf("swept %d files against %d enumerated literals", swept, len(literals))
	if swept != len(docsTabTargetFiles) {
		t.Fatalf("swept %d files, want %d", swept, len(docsTabTargetFiles))
	}

	for _, h := range hits {
		t.Errorf("hand-written tab count survives: %s contains %q", h.file, h.literal)
	}
	if len(hits) > 0 {
		t.Errorf("%d enumerated literal(s) still present; the run phase removes the count, it does not rewrite it", len(hits))
	}
}

// loadCountLiterals reads the enumerated literal set. An empty pattern set
// matches nothing and would let the literals layer pass vacuously, so an empty
// file is fatal.
func loadCountLiterals(t *testing.T) []string {
	t.Helper()
	raw := readRepoFile(t, countLiteralsPath)
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		t.Fatalf("docs tab guard: %s is empty; an empty literal set matches nothing and would pass vacuously", countLiteralsPath)
	}
	return out
}

// ---------------------------------------------------------------------------
// N1 — numeral sweep + allowlist
// ---------------------------------------------------------------------------

// tabNoun is the tab noun in the four locales. The English form carries an
// ASCII word boundary: without it, "stays selectable" matches on the "tab"
// inside "selectable" (README.md, measured both ways).
const tabNoun = `(?:\btabs?\b|탭|タブ|标签页)`

// adjacency is the gap allowed between the numeral and the tab noun. Ten
// characters separate "nine" from "tabs" in "The nine settings tabs", so a
// window narrow enough to be tidy is structurally blind to the English sites.
const adjacency = `[^.。]{0,12}`

// digitNumeralRe is the digit axis.
var digitNumeralRe = regexp.MustCompile(`[0-9]+` + adjacency + tabNoun)

// wordNumeralRe is the spelled-out-word axis — the axis whose absence let the
// previous cleanup pass while six sites survived. Alternatives are ordered
// longest-first because Go's regexp is leftmost-FIRST, so "열" placed ahead of
// "열네" would report a truncated phrase.
//
// Maintaining this class per locale is the standing cost of opening the word
// axis (spec.md §7): a new locale or a new spelling that is not listed here
// drops out silently.
var wordNumeralRe = regexp.MustCompile(`(?i)(?:` +
	`seventeen|thirteen|fourteen|eighteen|nineteen|` +
	`fifteen|sixteen|nineteen|twelve|twenty|eleven|` +
	`three|seven|eight|four|five|nine|one|two|six|ten|` +
	`열여섯|열다섯|열네|열세|열두|열한|여덟|다섯|여섯|일곱|아홉|하나|` +
	`열|둘|셋|넷|한|두|세|네|` +
	`[一二三四五六七八九十]` +
	`)` + adjacency + tabNoun)

// ordinalPrefixRe neutralizes an ordinal prefix before the sweep runs. 第三方
// ("third-party") is not a count of anything; its 三 would otherwise read as a
// numeral adjacent to 标签页 in the zh console prose.
var ordinalPrefixRe = regexp.MustCompile(`第[一二三四五六七八九十]`)

// allowlistRule is the ONE exemption: the codex-panel rationale lines, which
// count the Audit and MCP tabs — two tabs, not the settings tab total — and are
// out of scope per spec.md §4.
//
// Identification is by CONTENT, not line number, because this card's edits move
// line numbers. The file scope is load-bearing: an unscoped "codex" rule also
// exempts the four README tab-name-list lines (they list the Codex tab and use
// the lowercase token in the following sentence), which silently removes four
// enumerated sites from the guard — measured, 6 word-axis hits collapse to 3.
type allowRule struct {
	name       string
	fileSuffix string
	token      string
	reason     string
}

var docsTabAllowRules = []allowRule{{
	name:       "codex-panel-rationale",
	fileSuffix: "advanced/moai-web-console.md",
	token:      "codex",
	reason:     "counts the Audit and MCP tabs (codex panel rationale), not the settings tab total — spec.md §4",
}}

// docsTabAllowlist runs the numeral sweep on both axes and asserts the hit set
// equals the allowlist exactly.
//
// The word-axis hit set is PRINTED, one line per hit, so the set itself can be
// compared against the enumeration rather than merely counted. Without that
// output nothing constrains the BREADTH of the word class: a guard implementing
// only {fourteen} passes every shell-level criterion while missing five of the
// six enumerated word-form sites (acceptance.md AC-TCD-012).
//
// The hit lines go to stdout via fmt.Println rather than t.Logf on purpose:
// AC-TCD-012 greps them anchored at column 0 ('^word-axis hit '), and t.Logf
// indents its output and prefixes it with file:line.
func docsTabAllowlist(t *testing.T) {
	type sweepHit struct {
		file   string
		line   int
		phrase string
		axis   string
	}
	var hits []sweepHit
	var wordHits []sweepHit
	allowedLines := 0

	for _, rel := range docsTabTargetFiles {
		content := readRepoFile(t, rel)
		for i, line := range strings.Split(content, "\n") {
			lineNo := i + 1

			if rule, ok := matchAllowRule(rel, line); ok {
				allowedLines++
				_ = rule
				continue
			}

			// Neutralize ordinal prefixes before matching.
			scanned := ordinalPrefixRe.ReplaceAllString(line, "第X")

			if m := digitNumeralRe.FindString(scanned); m != "" {
				hits = append(hits, sweepHit{file: rel, line: lineNo, phrase: strings.TrimSpace(m), axis: "digit"})
			}
			if m := wordNumeralRe.FindString(scanned); m != "" {
				h := sweepHit{file: rel, line: lineNo, phrase: strings.TrimSpace(m), axis: "word"}
				hits = append(hits, h)
				wordHits = append(wordHits, h)
			}
		}
	}

	for _, h := range wordHits {
		fmt.Printf("word-axis hit %s: %s\n", h.file, h.phrase)
	}

	t.Logf("allowed rules %d", len(docsTabAllowRules))
	t.Logf("allowed lines %d", allowedLines)

	if len(docsTabAllowRules) != 1 {
		t.Errorf("allowlist carries %d rules, want exactly 1; a quietly growing allowlist disarms the guard", len(docsTabAllowRules))
	}
	for _, r := range docsTabAllowRules {
		if r.fileSuffix == "" {
			t.Errorf("allowlist rule %q has no file scope; an unscoped rule exempts the README tab-name lists and removes four enumerated sites from the guard", r.name)
		}
		if r.reason == "" {
			t.Errorf("allowlist rule %q carries no stated reason", r.name)
		}
	}
	if allowedLines != allowedLinesBaseline {
		t.Errorf("allowlist exempts %d lines, want %d; the rule count is unchanged but the exempted surface moved", allowedLines, allowedLinesBaseline)
	}

	for _, h := range hits {
		t.Errorf("%s axis: %s:%d writes a settings tab count by hand: %q", h.axis, h.file, h.line, h.phrase)
	}
	if len(hits) > 0 {
		t.Errorf("numeral sweep found %d unallowed hit(s) (%d on the word axis); every one is a count that can drift", len(hits), len(wordHits))
	}
}

func matchAllowRule(rel, line string) (allowRule, bool) {
	lower := strings.ToLower(line)
	for _, r := range docsTabAllowRules {
		if strings.HasSuffix(rel, r.fileSuffix) && strings.Contains(lower, r.token) {
			return r, true
		}
	}
	return allowRule{}, false
}

// ---------------------------------------------------------------------------
// N2 — names
// ---------------------------------------------------------------------------

// numberedBoldRe matches a numbered-list item whose name is bold — the shape
// the console pages use. Depending on richer structure would make the extractor
// fragile; depending on less would let it extract nothing and pass silently,
// which is why the extracted count is asserted rather than merely used.
var numberedBoldRe = regexp.MustCompile(`^\s*([0-9]+)\.\s+\*\*(.+?)\*\*`)

// parentheticalRe splits "localized name(English name)" into its two parts.
// Both ASCII and fullwidth parentheses appear across the locales.
var parentheticalRe = regexp.MustCompile(`^(.*?)\s*[(（]([^)）]+)[)）]\s*$`)

// localeSeparatorRe splits a README tab-name run. Each locale punctuates the
// run differently: ", " (en), "·" (ko), "・" (ja), "、" (zh).
var localeSeparatorRe = regexp.MustCompile(`\s*[,、，·・]\s*`)

// docsTabNames asserts the documented name lists equal the rendered labels, in
// consoleTabs() order.
func docsTabNames(t *testing.T) {
	tabs := consoleTabs()
	enLabels := make([]string, 0, len(tabs))
	for _, tb := range tabs {
		enLabels = append(enLabels, tb.Baseline)
	}

	i18n := loadI18nLabels(t)

	for _, rel := range docsTabNameListFiles {
		content := readRepoFile(t, rel)

		var got []string
		var localeOnly map[int]string
		if strings.HasPrefix(filepath.Base(rel), "README") {
			got = extractREADMENames(t, rel, content, enLabels)
		} else {
			got, localeOnly = extractConsoleNames(t, rel, content)
		}

		t.Logf("%s: extracted %d tab names", rel, len(got))
		if len(got) != len(tabs) {
			t.Errorf("%s: extracted %d tab names, want %d — extracting nothing is not agreement, it is a guard that read no list", rel, len(got), len(tabs))
			continue
		}

		locale := localeOfPath(rel)
		for i := range tabs {
			if _, isLocaleOnly := localeOnly[i]; isLocaleOnly {
				want := i18n[locale][tabs[i].LabelKey]
				if want == "" {
					t.Errorf("%s: tab %d is written only in %s and i18n.js has no %s label for %q", rel, i+1, locale, locale, tabs[i].LabelKey)
					continue
				}
				if got[i] != want {
					t.Errorf("%s: tab %d name = %q, console renders %q", rel, i+1, got[i], want)
				}
				continue
			}
			if got[i] != enLabels[i] {
				t.Errorf("%s: tab %d name = %q, console renders %q", rel, i+1, got[i], enLabels[i])
			}
		}
	}
}

// extractREADMENames pulls the tab-name run out of a README. The run is
// delimited by the first and last rendered labels; everything between them is
// split on the locale's own separator.
func extractREADMENames(t *testing.T, rel, content string, enLabels []string) []string {
	t.Helper()
	first, last := enLabels[0], enLabels[len(enLabels)-1]

	for _, line := range strings.Split(content, "\n") {
		start := strings.Index(line, first)
		if start < 0 {
			continue
		}
		end := strings.Index(line, last)
		if end < start {
			continue
		}
		run := line[start : end+len(last)]
		parts := localeSeparatorRe.Split(run, -1)
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		return out
	}

	t.Errorf("%s: no tab-name run found (expected a line carrying both %q and %q)", rel, first, last)
	return nil
}

// extractConsoleNames pulls the numbered tab list out of a console page and
// splits each bold entry into the name the document actually asserts.
//
// A localized page writes most entries as "localized name(English name)". The
// English part is what the guard compares, because that is the axis the name
// drift lives on. An entry written only in the locale's own language (the ones
// with no English gloss) is reported through localeOnly so the caller compares
// it against that locale's rendered label instead.
//
// The localized PREFIX of a glossed entry is deliberately not compared: at
// least one of those prefixes already diverges from i18n.js for reasons this
// card does not touch, and asserting it would make the criterion red forever.
func extractConsoleNames(t *testing.T, rel, content string) (names []string, localeOnly map[int]string) {
	t.Helper()
	localeOnly = map[int]string{}

	for _, line := range strings.Split(content, "\n") {
		m := numberedBoldRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		bold := strings.TrimSpace(m[2])
		idx := len(names)

		if p := parentheticalRe.FindStringSubmatch(bold); p != nil {
			names = append(names, strings.TrimSpace(p[2]))
			continue
		}
		if isASCII(bold) {
			names = append(names, bold)
			continue
		}
		localeOnly[idx] = bold
		names = append(names, bold)
	}

	if len(names) == 0 {
		t.Errorf("%s: no numbered tab list found", rel)
	}
	return names, localeOnly
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// localeOfPath returns the docs-site locale segment, or "en" for a README
// (every README writes the tab names in English).
func localeOfPath(rel string) string {
	const prefix = "docs-site/content/"
	if !strings.HasPrefix(rel, prefix) {
		return "en"
	}
	rest := rel[len(prefix):]
	if i := strings.Index(rest, "/"); i > 0 {
		return rest[:i]
	}
	return "en"
}

// i18nLocaleRe marks the start of a locale block in i18n.js.
var i18nLocaleRe = regexp.MustCompile(`^\s*([a-z]{2}):\s*\{`)

// i18nEntryRe matches one "key": "value" entry.
var i18nEntryRe = regexp.MustCompile(`^\s*"([^"]+)":\s*"(.*)",?\s*$`)

// loadI18nLabels reads the per-locale rendered labels. i18n.js is the label
// source of truth; schemaform.go's Baseline is its English fallback.
func loadI18nLabels(t *testing.T) map[string]map[string]string {
	t.Helper()
	content := readRepoFile(t, "internal/web/assets/i18n.js")

	out := map[string]map[string]string{}
	locale := ""
	for _, line := range strings.Split(content, "\n") {
		if m := i18nLocaleRe.FindStringSubmatch(line); m != nil {
			locale = m[1]
			out[locale] = map[string]string{}
			continue
		}
		if locale == "" {
			continue
		}
		if m := i18nEntryRe.FindStringSubmatch(line); m != nil {
			out[locale][m[1]] = m[2]
		}
	}

	if len(out) == 0 {
		t.Fatalf("docs tab guard: parsed no locales out of internal/web/assets/i18n.js")
	}
	return out
}
