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
		isREADME := strings.HasPrefix(filepath.Base(rel), "README")

		var got []string
		var prefixes, localeOnly map[int]string
		if isREADME {
			got = extractREADMENames(t, rel, content, enLabels)
		} else {
			got, prefixes, localeOnly = extractConsoleNames(t, rel, content)
		}

		t.Logf("%s: extracted %d tab names", rel, len(got))
		if len(got) != len(tabs) {
			t.Errorf("%s: extracted %d tab names, want %d — extracting nothing is not agreement, it is a guard that read no list", rel, len(got), len(tabs))
			continue
		}

		locale := localeOfPath(rel)
		for i := range tabs {
			// t677: the localized PREFIX of a glossed entry is now compared
			// against the locale's rendered label. t530 left this axis
			// unguarded ("one prefix already diverges"); all eight divergent
			// positions were doc-side drift and are aligned to the console in
			// this card, so the axis is assertable from here on.
			if pfx, isGlossed := prefixes[i]; isGlossed {
				want := i18n[locale][tabs[i].LabelKey]
				if want == "" {
					t.Errorf("%s: tab %d carries a localized name and i18n.js has no %s label for %q", rel, i+1, locale, tabs[i].LabelKey)
					continue
				}
				if pfx != want {
					t.Errorf("%s: tab %d localized name = %q, console renders %q", rel, i+1, pfx, want)
				}
				continue
			}
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
			if isREADME {
				if got[i] != enLabels[i] {
					t.Errorf("%s: tab %d name = %q, console renders %q", rel, i+1, got[i], enLabels[i])
				}
				continue
			}
			// Console page, bare ASCII entry (no gloss, no localized name):
			// it must equal what the console renders in THIS locale. A bare
			// "Codex" in the ko/ja/zh lists read as the English baseline and
			// passed while those locales render "Codex 설정/設定/设置" (t677);
			// entries whose label is the same in every locale (LLM, MCP)
			// keep passing.
			want := i18n[locale][tabs[i].LabelKey]
			if want == "" {
				t.Errorf("%s: tab %d carries no localized name and i18n.js has no %s label for %q", rel, i+1, locale, tabs[i].LabelKey)
				continue
			}
			if got[i] != want {
				t.Errorf("%s: tab %d name = %q, console renders %q", rel, i+1, got[i], want)
			}
		}
	}

	docsTabProseMentions(t, i18n)
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
// A localized page writes most entries as "localized name(English name)".
// Both parts are compared (t677): the English gloss against the console's
// English baseline, and the localized prefix against that locale's rendered
// i18n.js label — the prefix axis t530 left open because eight positions
// already diverged. An entry written only in the locale's own language (the
// ones with no English gloss) is reported through localeOnly so the caller
// compares it against that locale's rendered label instead.
func extractConsoleNames(t *testing.T, rel, content string) (names []string, prefixes map[int]string, localeOnly map[int]string) {
	t.Helper()
	prefixes = map[int]string{}
	localeOnly = map[int]string{}

	for _, line := range strings.Split(content, "\n") {
		m := numberedBoldRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		bold := strings.TrimSpace(m[2])
		idx := len(names)

		if p := parentheticalRe.FindStringSubmatch(bold); p != nil {
			prefixes[idx] = strings.TrimSpace(p[1])
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
	return names, prefixes, localeOnly
}

// ---------------------------------------------------------------------------
// N2b — prose mentions (t677)
// ---------------------------------------------------------------------------

// proseNounRe finds each tab-noun occurrence in a line. The CJK nouns match
// as literals (agglutinated suffixes like 탭에는 / タブには stay attached and
// are harmless — the scan looks BACKWARD from the noun). The English noun is
// word-bounded so "acceptable" does not read as a tab.
var proseNounRe = regexp.MustCompile(`탭|タブ|标签页|\btabs?\b`)

// proseConsoleSuffix narrows the prose scan to the four console pages.
const proseConsoleSuffix = "advanced/moai-web-console.md"

// proseNoise lists the function words that legally sit immediately before a
// tab noun without naming a tab, measured at base 16f3b8a81 across the four
// console pages (e.g. "다음 탭들이", "the tabs below", "each tab"). The list
// carries the same standing cost as the wordNumeralRe class: a new prose
// phrasing that is not listed here turns the names layer red, and adding to
// this list is a deliberate, reviewable edit. Only the SPACE-DELIMITED
// locales are enforced — see proseAdvisoryLocales.
var proseNoise = map[string]map[string]bool{
	"ko": {
		"다음": true, "각": true, "그": true, "있는": true,
		"어느": true, "알려면": true, "자기": true, "이": true,
	},
	"en": {
		"the": true, "each": true, "that": true, "a": true,
		"which": true, "two": true, "owning": true,
	},
}

// proseAdvisoryLocales are the locales where the prose axis is ADVISORY only.
// Japanese and Chinese write no spaces between words, so the run immediately
// before a tab noun blends function words into content ("存在错误的标签页")
// and no mechanical name/noise boundary exists. Unanchored mentions there are
// counted and logged, never failed: the authoritative name fact for every
// locale stays the numbered list, which axis A pins against i18n.js.
var proseAdvisoryLocales = map[string]bool{"ja": true, "zh": true}

// docsTabProseMentions guards the prose axis (card t677, D9-D12): the GLM
// honesty-badge sentence in each console page names the GLM tab in that
// locale's own language ("GLM 설정 탭", "The GLM Settings tab", "GLM設定タブ",
// "GLM设置标签页"). Two failure shapes closed here:
//
//	paragraph deletion — a page must carry at least ONE prose mention of the
//	                    rendered GLM label followed by the tab noun;
//	silent rename      — the mention is matched against the label i18n.js
//	                    renders TODAY, so a rename without a docs edit fails.
//
// In the enforced locales (en, ko) an unanchored candidate — text before a
// tab noun that is neither a rendered label nor listed noise — is a failure:
// that is a stale tab name in prose, the exact shape this card repaired
// (the old "교차 세션 탭" wording). In the advisory locales it is logged.
func docsTabProseMentions(t *testing.T, i18n map[string]map[string]string) {
	t.Helper()
	tabs := consoleTabs()

	var llmTab consoleTab
	for _, tb := range tabs {
		if tb.ID == "llm" {
			llmTab = tb
		}
	}

	for _, rel := range docsTabNameListFiles {
		if !strings.HasSuffix(rel, proseConsoleSuffix) {
			continue
		}
		locale := localeOfPath(rel)

		// A prose mention is anchored when the text before the noun ends
		// with a label the console renders in this locale — or with its
		// English baseline, which every locale's prose uses as the codename
		// shorthand ("Codex 탭", "Codex タブ"). Longest suffix wins so that
		// multi-word labels ("GLM Settings") anchor ahead of their parts.
		valid := map[string]bool{}
		for _, tb := range tabs {
			valid[tb.Baseline] = true
			if l := i18n[locale][tb.LabelKey]; l != "" {
				valid[l] = true
			}
		}
		llmLabel := i18n[locale][llmTab.LabelKey]
		if llmLabel == "" {
			llmLabel = llmTab.Baseline
		}

		content := readRepoFile(t, rel)
		llmMentions := 0
		unanchored := 0
		for _, line := range strings.Split(content, "\n") {
			// The numbered list is axis A's surface; the prose axis reads
			// everything else (headings included).
			if numberedBoldRe.MatchString(line) {
				continue
			}
			for _, loc := range proseNounRe.FindAllStringIndex(line, -1) {
				pre := strings.TrimSuffix(line[:loc[0]], " ")
				anchored := ""
				for l := range valid {
					if len(l) > len(anchored) && strings.HasSuffix(pre, l) {
						anchored = l
					}
				}
				if anchored == llmLabel {
					llmMentions++
				}
				if anchored != "" {
					continue
				}

				run := pre
				if k := strings.LastIndexAny(pre, " "); k >= 0 {
					run = pre[k+1:]
				}
				if proseAdvisoryLocales[locale] {
					unanchored++
					continue
				}
				noise := proseNoise[locale]
				if run != "" && noise[strings.ToLower(run)] {
					continue
				}
				t.Errorf("%s: prose tab mention %q before the tab noun is neither a rendered label nor listed noise — a stale or unknown tab name", rel, run)
			}
		}

		if llmMentions == 0 {
			t.Errorf("%s: no prose mention of the %q tab followed by the tab noun — the GLM honesty-badge paragraph (D9-D12) was deleted or its tab name no longer matches what the console renders", rel, llmLabel)
		}
		t.Logf("%s: prose axis — %d anchored GLM mention(s), %d advisory unanchored", rel, llmMentions, unanchored)
	}
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
