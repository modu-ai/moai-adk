package web

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"
)

// i18n_governance_test.go — the detector, parser, and governance checks for
// the web console i18n catalogue (SPEC-I18N-GOVERNANCE-001).
//
// All symbols in this file are test-binary-only (package web, _test.go). No
// production symbol depends on them; the shipped console binary is unaffected
// (C4). The detector is a PURE function of (catalogue, allowlist) so the
// negative control can drive it with synthetic inputs (REQ-I18NGOV-008).

// i18nLocales is the locale set the catalogue carries. Adding a locale is out
// of scope for this SPEC (§D Exclusions); the list is fixed.
var i18nLocales = []string{"en", "ko", "ja", "zh"}

// i18nNonEnLocales is the set the untranslated-value detector iterates.
func i18nNonEnLocales() []string { return []string{"ko", "ja", "zh"} }

// i18nEntryRe matches one catalogue entry line: `"key": "value",`. The key
// class is "any char except a double quote" so it admits the four special
// keys carrying `+`, `[`, `]` (B4): f.report.format.opt.html+md and the three
// f.model.opt.*[1m] variants. The value class admits backslash escapes so a
// future translation containing a quotation mark survives the parse.
var i18nEntryRe = regexp.MustCompile(`(?m)^\s*"([^"]+)":\s*"((?:\\.|[^"\\])*)",?\s*$`)

// parseI18nCatalogue splits the i18n.js object literal into its four locale
// blocks and parses each into a key→value map (REQ-I18NGOV-007). It reads keys
// containing `+`, `[`, `]` correctly (B4). It does NOT touch the existing
// localeBlocks() splitter in webux_followup_test.go; that helper stays
// byte-identical so the agentdesc.* exemption tests (C1) are undisturbed.
func parseI18nCatalogue(t *testing.T, dict string) map[string]map[string]string {
	t.Helper()
	out := map[string]map[string]string{}
	for _, loc := range i18nLocales {
		block := extractLocaleBlock(t, dict, loc)
		m := map[string]string{}
		for _, mt := range i18nEntryRe.FindAllStringSubmatch(block, -1) {
			m[mt[1]] = mt[2]
		}
		out[loc] = m
	}
	return out
}

// extractLocaleBlock returns the substring of dict spanning one top-level
// locale block (`<loc>: { ... }`), brace-depth matched. Mirrors the splitting
// heuristic of localeBlocks() but returns the matched span.
func extractLocaleBlock(t *testing.T, dict, loc string) string {
	t.Helper()
	startMark := "\n  " + loc + ": {"
	i := strings.Index(dict, startMark)
	if i < 0 {
		t.Fatalf("i18n.js has no %q locale block", loc)
	}
	rest := dict[i:]
	depth, end := 0, -1
	for j := 0; j < len(rest); j++ {
		switch rest[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = j
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		t.Fatalf("i18n.js %q locale block has no closing brace", loc)
	}
	return rest[:end+1]
}

// i18nNormalize applies the REQ-I18NGOV-009 normalization: Unicode NFC
// composition, surrounding-whitespace trim, and case folding. For the ASCII
// and CJK values in this catalogue NFC and case-fold are the identity for the
// non-ASCII code points; the case-fold step catches the one mixed-case value
// (mp.col.effort: en "Effort" vs non-en "effort").
func i18nNormalize(s string) string {
	return strings.ToLower(strings.TrimSpace(norm.NFC.String(s)))
}

// i18nViolation is one detector finding.
type i18nViolation struct {
	Key    string
	Locale string
	Detail string
}

func (v i18nViolation) String() string {
	return fmt.Sprintf("key=%q locale=%q %s", v.Key, v.Locale, v.Detail)
}

// i18nAllowSet turns the allowlist slice into a constant-time membership set.
func i18nAllowSet() map[string]struct{} {
	set := make(map[string]struct{}, len(i18nUntranslatedAllowlist))
	for _, e := range i18nUntranslatedAllowlist {
		set[e.Key] = struct{}{}
	}
	return set
}

// detectUntranslatedValues is the pure untranslated-value detector
// (REQ-I18NGOV-008, REQ-I18NGOV-009, REQ-I18NGOV-010, REQ-I18NGOV-011). Given
// a parsed catalogue and an allow set, it returns one violation per non-`en`
// value whose normalized form equals its `en` counterpart, excluding the
// lang.opt.* family (delegated to the endonym invariant) and every allowlisted
// key. It does not read the shipped files; callers may pass synthetic inputs.
func detectUntranslatedValues(catalogue map[string]map[string]string, allowed map[string]struct{}) []i18nViolation {
	en := catalogue["en"]
	var out []i18nViolation
	for _, loc := range i18nNonEnLocales() {
		bloc := catalogue[loc]
		for k, v := range bloc {
			ev, ok := en[k]
			if !ok {
				continue // reverse-coverage gap; handled by coverage checks
			}
			if strings.HasPrefix(k, "lang.opt.") {
				continue // endonym family; handled by endonym invariants
			}
			if _, ok := allowed[k]; ok {
				continue
			}
			if i18nNormalize(v) == i18nNormalize(ev) {
				out = append(out, i18nViolation{
					Key:    k,
					Locale: loc,
					Detail: fmt.Sprintf("non-en value %q equals en %q after NFC+trim+casefold", v, ev),
				})
			}
		}
	}
	return out
}

// detectAllowlistOrphans is the pure orphan detector (REQ-I18NGOV-015). An
// allowlist entry is an orphan if its key is absent from the catalogue, OR its
// value is not identical to `en` in ANY non-`en` locale (under normalization).
// The second clause forces an entry's removal in the same change that
// translates or deletes its key.
func detectAllowlistOrphans(catalogue map[string]map[string]string, entries []i18nUntranslatedAllowEntry) []i18nViolation {
	en := catalogue["en"]
	var out []i18nViolation
	for _, e := range entries {
		ev, enHas := en[e.Key]
		if !enHas {
			out = append(out, i18nViolation{Key: e.Key, Detail: "allowlisted key is absent from the en locale"})
			continue
		}
		seenIdentical := false
		for _, loc := range i18nNonEnLocales() {
			v, ok := catalogue[loc][e.Key]
			if !ok {
				continue
			}
			if i18nNormalize(v) == i18nNormalize(ev) {
				seenIdentical = true
				break
			}
		}
		if !seenIdentical {
			out = append(out, i18nViolation{Key: e.Key, Detail: "allowlisted key is not identical to en in any non-en locale (stale exemption)"})
		}
	}
	return out
}

// ── AC-I18NGOV-001 : the real catalogue is green ────────────────────────────

// TestI18nUntranslatedValues runs the detector on the shipped catalogue against
// the shipped allowlist and asserts zero violations. Green here is reachable
// only by ruling on every identity-set member (allowlist entry or translation
// fix); widening the comparison instead fails AC-I18NGOV-004 and the negative
// control AC-I18NGOV-002.
func TestI18nUntranslatedValues(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	cat := parseI18nCatalogue(t, dict)
	violations := detectUntranslatedValues(cat, i18nAllowSet())
	for _, v := range violations {
		t.Errorf("untranslated-value violation: %s", v)
	}
}

// ── AC-I18NGOV-002 : the negative control (anti-rubber-stamp) ──────────────

// TestI18nUntranslatedDetectorNegativeControl proves the allowlist is a gate,
// not a rubber stamp. A synthetic catalogue carrying an inert placeholder key
// whose non-en value equals its en value (a) produces a violation when the key
// is NOT in the allowlist, and (b) produces none once the key IS allowlisted.
// This is only satisfiable because the detector is a pure function (M2); a
// detector that reads the shipped files cannot be driven this way. The fixture
// uses an inert placeholder and a short interface word — no secret-shaped or
// key-shaped literal (C2).
func TestI18nUntranslatedDetectorNegativeControl(t *testing.T) {
	// Arrange: a minimal synthetic catalogue. `cpl.placeholder` is the inert
	// key; its ko value is byte-identical to en, so the detector MUST flag it.
	synthetic := map[string]map[string]string{
		"en": {"cpl.placeholder": "stub", "lang.opt.en": "English"},
		"ko": {"cpl.placeholder": "stub", "lang.opt.en": "영어", "lang.opt.ko": "한국어"},
		"ja": {"cpl.placeholder": "stub", "lang.opt.en": "英語", "lang.opt.ja": "日本語"},
		"zh": {"cpl.placeholder": "stub", "lang.opt.en": "英语", "lang.opt.zh": "中文"},
	}

	// (a) without the key in the allowlist, a violation naming the key+locale
	// must be reported.
	emptyAllow := map[string]struct{}{}
	got := detectUntranslatedValues(synthetic, emptyAllow)
	if !violationNames(got, "cpl.placeholder", []string{"ko", "ja", "zh"}) {
		t.Fatalf("detector without allowlist did not report cpl.placeholder in all three non-en locales; got %v", got)
	}

	// (b) with the key allowlisted, the identical input produces no violation.
	allowWithKey := map[string]struct{}{"cpl.placeholder": {}}
	got2 := detectUntranslatedValues(synthetic, allowWithKey)
	if len(got2) != 0 {
		t.Fatalf("detector with cpl.placeholder allowlisted still reported violations: %v", got2)
	}
}

// violationNames checks that vs contains, for key k, every locale in locs.
func violationNames(vs []i18nViolation, k string, locs []string) bool {
	for _, l := range locs {
		hit := false
		for _, v := range vs {
			if v.Key == k && v.Locale == l {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	return true
}

// ── AC-I18NGOV-003 : the allowlist has no orphans ───────────────────────────

// TestI18nAllowlistNoOrphans asserts every allowlist entry resolves to a real
// catalogue key whose value is identical to en in at least one non-en locale.
// An entry surviving the deletion or translation of its key is a silent blanket
// exemption; this test forces its removal in the same change.
func TestI18nAllowlistNoOrphans(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	cat := parseI18nCatalogue(t, dict)
	orphans := detectAllowlistOrphans(cat, i18nUntranslatedAllowlist)
	for _, o := range orphans {
		t.Errorf("allowlist orphan: %s", o)
	}
}

// ── AC-I18NGOV-004 : allowlist shape (exact keys, closed taxonomy, bound) ───

// TestI18nAllowlistShape asserts: no wildcard/regex metachar in a key; every
// entry carries a non-empty justification; every reason is one of the closed
// three; no lang.opt.* key; entry count ≤ 30. The closed taxonomy is
// additionally compile-time-enforced by the i18nUntranslatedReason type.
func TestI18nAllowlistShape(t *testing.T) {
	valid := map[i18nUntranslatedReason]struct{}{
		reasonTechnicalIdentifier: {},
		reasonProperNoun:          {},
		reasonAcronym:             {},
	}
	const maxEntries = 30
	if len(i18nUntranslatedAllowlist) > maxEntries {
		t.Errorf("allowlist has %d entries; cap is %d (REQ-I18NGOV-016)", len(i18nUntranslatedAllowlist), maxEntries)
	}
	for _, e := range i18nUntranslatedAllowlist {
		if e.Key == "" {
			t.Error("allowlist entry has an empty key")
		}
		// Wildcard / pattern indicators only. `.`, `+`, `[`, `]` are NOT
		// forbidden because they appear as LITERAL characters in real
		// catalogue keys (f.model.opt.opus[1m], f.report.format.opt.html+md).
		// Exact-match semantics are additionally enforced by the orphan
		// check, which requires every allowlisted key to exist verbatim in
		// the catalogue.
		if strings.ContainsAny(e.Key, "*?^$|(){}\\") {
			t.Errorf("allowlist key %q contains a wildcard or regex pattern char (REQ-I18NGOV-004)", e.Key)
		}
		if strings.HasPrefix(e.Key, "lang.opt.") {
			t.Errorf("allowlist key %q is in the lang.opt.* family (REQ-I18NGOV-014)", e.Key)
		}
		if strings.TrimSpace(e.Justification) == "" {
			t.Errorf("allowlist entry %q has an empty justification", e.Key)
		}
		if _, ok := valid[e.Reason]; !ok {
			t.Errorf("allowlist entry %q has reason %q outside the closed taxonomy", e.Key, e.Reason)
		}
	}
}

// ── AC-I18NGOV-005 : the endonym family is correct by construction ──────────

// TestI18nEndonymInvariants asserts the bidirectional lang.opt.* invariant:
// self-equality (REQ-I18NGOV-012) and cross-locale distinctness
// (REQ-I18NGOV-013). Together with the allowlist's refusal of lang.opt.* keys
// (AC-I18NGOV-004), an English endonym copied into the wrong locale is caught
// rather than exempted.
func TestI18nEndonymInvariants(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	cat := parseI18nCatalogue(t, dict)
	// (1) self-equality: lang.opt.<L> in L equals lang.opt.<L> in en.
	for _, self := range i18nLocales {
		key := "lang.opt." + self
		enVal, ok := cat["en"][key]
		if !ok {
			t.Errorf("en locale has no %q key", key)
			continue
		}
		selfVal, ok := cat[self][key]
		if !ok {
			t.Errorf("locale %q has no own endonym %q", self, key)
			continue
		}
		if selfVal != enVal {
			t.Errorf("endonym self-equality broken: lang.opt.%s in %s = %q, en = %q", self, self, selfVal, enVal)
		}
	}
	// (2) exonym distinctness: for each NON-en locale L, lang.opt.<X> where
	// X != L differs from en's lang.opt.<X>. (en is the reference locale and
	// carries all endonyms as the endonym baseline; it is excluded from the
	// exonym loop because comparing en to itself is trivially equal and not
	// the invariant REQ-I18NGOV-013 protects.)
	for _, loc := range i18nNonEnLocales() {
		for _, x := range i18nLocales {
			if x == loc {
				continue
			}
			key := "lang.opt." + x
			enVal, ok := cat["en"][key]
			if !ok {
				continue
			}
			locVal, ok := cat[loc][key]
			if !ok {
				t.Errorf("locale %q is missing exonym %q", loc, key)
				continue
			}
			if locVal == enVal {
				t.Errorf("exonym distinctness broken: lang.opt.%s in %s = %q equals en value (an English endonym leaked into the wrong locale)", x, loc, locVal)
			}
		}
	}
}

// ── AC-I18NGOV-006 : forward key coverage ───────────────────────────────────

// TestI18nKeyCoverageForward asserts every en key exists in ko, ja, zh.
func TestI18nKeyCoverageForward(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	cat := parseI18nCatalogue(t, dict)
	for _, k := range sortedKeys(cat["en"]) {
		for _, loc := range i18nNonEnLocales() {
			if _, ok := cat[loc][k]; !ok {
				t.Errorf("forward coverage: en key %q missing from locale %q", k, loc)
			}
		}
	}
}

// ── AC-I18NGOV-007 : reverse key coverage modulo the exempt registry ────────

// TestI18nKeyCoverageReverse asserts every non-en key absent from en matches a
// declared en-exempt prefix, and every registry member carries a justification.
func TestI18nKeyCoverageReverse(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	cat := parseI18nCatalogue(t, dict)
	// Registry hygiene: each member must carry a non-empty justification.
	for _, r := range i18nEnExemptPrefixRegistry {
		if strings.TrimSpace(r.Justification) == "" {
			t.Errorf("en-exempt prefix %q has no justification", r.Prefix)
		}
	}
	for _, loc := range i18nNonEnLocales() {
		for _, k := range sortedKeys(cat[loc]) {
			if _, ok := cat["en"][k]; ok {
				continue
			}
			if !matchesExemptPrefix(k) {
				t.Errorf("reverse coverage: locale %q key %q absent from en and matches no declared exempt prefix", loc, k)
			}
		}
	}
}

// matchesExemptPrefix reports whether k begins with any registered en-exempt
// prefix.
func matchesExemptPrefix(k string) bool {
	for _, r := range i18nEnExemptPrefixRegistry {
		if strings.HasPrefix(k, r.Prefix) {
			return true
		}
	}
	return false
}

// ── AC-I18NGOV-008 : the parser reads special-character keys ────────────────

// TestI18nParserSpecialKeys asserts the parser admits the four keys outside
// [A-Za-z0-9._-]: f.report.format.opt.html+md and the three f.model.opt.*[1m]
// variants. A key regex assuming that character class would drop them silently.
func TestI18nParserSpecialKeys(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	cat := parseI18nCatalogue(t, dict)
	specials := []string{
		"f.report.format.opt.html+md",
		"f.model.opt.opus[1m]",
		"f.model.opt.sonnet[1m]",
		"f.model.opt.fable[1m]",
	}
	for _, sp := range specials {
		for _, loc := range i18nLocales {
			v, ok := cat[loc][sp]
			if !ok {
				t.Errorf("parser dropped special key %q in locale %q", sp, loc)
				continue
			}
			if strings.TrimSpace(v) == "" {
				t.Errorf("special key %q in locale %q parsed to an empty value", sp, loc)
			}
		}
	}
}

// ── AC-I18NGOV-009 : the governance contract and named owner are present ────

// TestI18nGovernanceContractPresent asserts (a) the allowlist artifact carries
// its inline governance contract, and (b) the i18n.js header names the
// allowlist artifact path as the owning surface.
func TestI18nGovernanceContractPresent(t *testing.T) {
	// (a) the allowlist source carries the contract markers. These literals
	// are load-bearing: a reader (human or grep) locates the contract by them.
	// The Go test CWD is the package directory, so the sibling _test.go source
	// is readable directly.
	allowSrc, err := os.ReadFile("i18n_untranslated_allowlist_test.go")
	if err != nil {
		t.Fatalf("could not read allowlist source: %v", err)
	}
	allowStr := string(allowSrc)
	for _, want := range []string{
		"Governance contract",
		"Location :",
		"Entry    :",
		"Who adds",
		"Ruling   :",
		"Bound    :",
	} {
		if !strings.Contains(allowStr, want) {
			t.Errorf("allowlist artifact missing governance-contract marker %q", want)
		}
	}

	// (b) the i18n.js header names the allowlist artifact path.
	dict := readEmbeddedAsset(t, "i18n.js")
	header := dict
	if i := strings.Index(header, "window.MOAI_I18N"); i >= 0 {
		header = header[:i]
	}
	if !strings.Contains(header, "i18n_untranslated_allowlist") {
		t.Error("i18n.js header does not name the allowlist artifact i18n_untranslated_allowlist_test.go as the owning governance surface")
	}
}

// sortedKeys returns m's keys in sorted order for deterministic iteration.
func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sortStrings(out)
	return out
}

func sortStrings(s []string) {
	// insertion sort keeps this file dependency-free; the catalogues are small.
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
