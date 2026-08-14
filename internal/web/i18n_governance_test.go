package web

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// i18n_governance_test.go — the value-level governance suite for
// internal/web/assets/i18n.js (SPEC-I18N-GOVERNANCE-001).
//
// This file owns the parser (REQ-I18NGOV-007), the pure-function detector
// (REQ-I18NGOV-008/009/010/011), the orphan check (REQ-I18NGOV-015), the
// endonym invariants (REQ-I18NGOV-012/013/014), the key-coverage checks
// (REQ-I18NGOV-018/019), the shape guard (REQ-I18NGOV-002..005/016), the
// governance-contract presence check (REQ-I18NGOV-006/021/023), the
// special-character-key parse test (REQ-I18NGOV-007), and the negative control
// (REQ-I18NGOV-017).
//
// All symbols here are test-binary only; the shipped console binary is
// byte-unaffected (C4).

// i18nLocaleOrder is the fixed four-locale order the catalogue carries.
var i18nLocaleOrder = []string{"en", "ko", "ja", "zh"}

// i18nNonEnLocales is every locale except en.
var i18nNonEnLocales = []string{"ko", "ja", "zh"}

// i18nLangOptPrefix is the endonym family delegated to the endonym invariants
// rather than the generic identity comparison (REQ-I18NGOV-011).
const i18nLangOptPrefix = "lang.opt."

// i18nKVRegexp matches a quoted catalogue key, its separating colon, and a
// quoted value that may contain backslash-escaped characters. The key class is
// "any character except a double quote" so keys containing '+', '[', ']', and
// '.' all parse (REQ-I18NGOV-007, plan §B4).
var i18nKVRegexp = regexp.MustCompile(`"([^"]+)"\s*:\s*"((?:[^"\\]|\\.)*)"`)

// parseI18nCatalogue parses the window.MOAI_I18N object literal into a
// per-locale key→value map. It splits the four locale blocks by their top-level
// markers, then within each block extracts "key": "value" pairs, tolerating
// backslash-escaped characters in values (REQ-I18NGOV-007). It is a pure
// function of its input so callers may pass a synthetic catalogue without
// touching the shipped file (REQ-I18NGOV-008).
func parseI18nCatalogue(dict string) (map[string]map[string]string, error) {
	// Locate each top-level locale block by its "\n  <loc>: {" marker.
	idx := make(map[string]int, len(i18nLocaleOrder))
	for _, loc := range i18nLocaleOrder {
		marker := "\n  " + loc + ": {"
		i := strings.Index(dict, marker)
		if i < 0 {
			return nil, fmt.Errorf("i18n catalogue has no %q locale block", loc)
		}
		idx[loc] = i
	}
	out := make(map[string]map[string]string, len(i18nLocaleOrder))
	for n, loc := range i18nLocaleOrder {
		start := idx[loc] + len("\n  "+loc+": {")
		end := len(dict)
		if n+1 < len(i18nLocaleOrder) {
			end = idx[i18nLocaleOrder[n+1]]
		}
		block := dict[start:end]
		out[loc] = parseI18nBlock(block)
	}
	return out, nil
}

// parseI18nBlock extracts the key→value pairs from one locale block's body.
func parseI18nBlock(block string) map[string]string {
	m := make(map[string]string)
	for _, mt := range i18nKVRegexp.FindAllStringSubmatch(block, -1) {
		m[mt[1]] = i18nUnescapeValue(mt[2])
	}
	return m
}

// i18nUnescapeValue expands the common JS string escape sequences so a value
// carrying a quote / backslash / newline is compared against its true string
// form. No value in the shipped catalogue currently contains an escape, but the
// parser must tolerate a future translation that introduces one (plan §B4).
func i18nUnescapeValue(raw string) string {
	if !strings.ContainsRune(raw, '\\') {
		return raw
	}
	var b strings.Builder
	b.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			b.WriteByte(raw[i])
			continue
		}
		if i+1 >= len(raw) {
			b.WriteByte('\\')
			break
		}
		i++
		switch raw[i] {
		case '"':
			b.WriteByte('"')
		case '\\':
			b.WriteByte('\\')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '/':
			b.WriteByte('/')
		default:
			b.WriteByte('\\')
			b.WriteByte(raw[i])
		}
	}
	return b.String()
}

// i18nNormalize applies the REQ-I18NGOV-009 normalization: Unicode NFC
// composition, surrounding-whitespace trimming, and Unicode case folding.
func i18nNormalize(s string) string {
	folder := cases.Fold()
	return folder.String(norm.NFC.String(strings.TrimSpace(s)))
}

// i18nViolation is one untranslated-value finding.
type i18nViolation struct {
	Key         string
	Locale      string
	EnValue     string
	LocaleValue string
}

// detectI18nUntranslated is the pure-function detector (REQ-I18NGOV-008). For
// every non-en locale it compares each value (that also exists in en) to its en
// counterpart under NFC + trim + casefold normalization, excluding the
// lang.opt.* family (REQ-I18NGOV-011) and every allowlisted key. It returns one
// violation per (key, locale) that is normalized-identical yet neither excluded
// nor allowlisted (REQ-I18NGOV-010).
func detectI18nUntranslated(cat map[string]map[string]string, allowlist []i18nAllowEntry) []i18nViolation {
	allowed := make(map[string]struct{}, len(allowlist))
	for _, e := range allowlist {
		allowed[e.Key] = struct{}{}
	}
	en := cat["en"]
	var out []i18nViolation
	for _, loc := range i18nNonEnLocales {
		block, ok := cat[loc]
		if !ok {
			continue
		}
		for k, v := range block {
			env, present := en[k]
			if !present {
				continue // reverse-coverage handles en-absent keys separately.
			}
			if strings.HasPrefix(k, i18nLangOptPrefix) {
				continue // endonym family — structural invariant, not identity.
			}
			if _, ok := allowed[k]; ok {
				continue
			}
			if i18nNormalize(v) == i18nNormalize(env) {
				out = append(out, i18nViolation{Key: k, Locale: loc, EnValue: env, LocaleValue: v})
			}
		}
	}
	return out
}

// detectI18nOrphans returns allowlist entries that no longer earn their place:
// their key is absent from the catalogue, or its value is not identical to en
// in ANY non-en locale (the key was deleted or later translated)
// (REQ-I18NGOV-015).
func detectI18nOrphans(cat map[string]map[string]string, allowlist []i18nAllowEntry) []i18nAllowEntry {
	en := cat["en"]
	var out []i18nAllowEntry
	for _, e := range allowlist {
		if _, ok := en[e.Key]; !ok {
			out = append(out, e)
			continue
		}
		env := en[e.Key]
		matched := false
		for _, loc := range i18nNonEnLocales {
			if v, ok := cat[loc][e.Key]; ok && i18nNormalize(v) == i18nNormalize(env) {
				matched = true
				break
			}
		}
		if !matched {
			out = append(out, e)
		}
	}
	return out
}

// shippedI18nCatalogue loads and parses the embedded i18n.js, failing the test
// if the catalogue cannot be parsed. Every real-catalogue test calls this once.
func shippedI18nCatalogue(t *testing.T) map[string]map[string]string {
	t.Helper()
	dict := readEmbeddedAsset(t, "i18n.js")
	cat, err := parseI18nCatalogue(dict)
	if err != nil {
		t.Fatalf("parse shipped i18n.js: %v", err)
	}
	return cat
}

// --- AC-I18NGOV-001: detector is green on the real catalogue (007..011) ---

func TestI18nUntranslatedValues(t *testing.T) {
	cat := shippedI18nCatalogue(t)
	viol := detectI18nUntranslated(cat, i18nUntranslatedAllowlist)
	for _, v := range viol {
		t.Errorf("untranslated value: locale=%s key=%s en=%q locale=%q", v.Locale, v.Key, v.EnValue, v.LocaleValue)
	}
}

// --- AC-I18NGOV-002: negative control — a non-allowlisted untranslated key
// fails, and the identical input passes once allowlisted (008, 017) ---

func TestI18nUntranslatedDetectorNegativeControl(t *testing.T) {
	// Synthetic catalogue with an inert placeholder key + a short interface
	// word. Contains no secret-shaped or key-shaped literal (C2).
	synthetic := strings.Join([]string{
		"window.MOAI_I18N = {",
		"  en: {",
		`    "syn.placeholder.key": "Refresh",`,
		`    "syn.word.save": "Save",`,
		"  },",
		"  ko: {",
		`    "syn.placeholder.key": "Refresh",`, // identical to en — not allowlisted → violation
		`    "syn.word.save": "저장",`,            // genuinely translated → no violation
		"  },",
		"  ja: {",
		`    "syn.placeholder.key": "Refresh",`,
		`    "syn.word.save": "保存",`,
		"  },",
		"  zh: {",
		`    "syn.placeholder.key": "Refresh",`,
		`    "syn.word.save": "保存",`,
		"  },",
		"}",
		"",
	}, "\n")
	cat, err := parseI18nCatalogue(synthetic)
	if err != nil {
		t.Fatalf("parse synthetic catalogue: %v", err)
	}

	// With an EMPTY allowlist the placeholder key must produce a violation in
	// every non-en locale (proves the detector is not a rubber stamp).
	violEmpty := detectI18nUntranslated(cat, nil)
	if len(violEmpty) == 0 {
		t.Fatal("detector reported zero violations on a synthetic untranslated key — the allowlist gate is not falsifiable")
	}
	sawPlaceholder := false
	for _, v := range violEmpty {
		if v.Key == "syn.placeholder.key" {
			sawPlaceholder = true
			if v.Locale != "ko" && v.Locale != "ja" && v.Locale != "zh" {
				t.Errorf("violation named a non-en locale %q for the placeholder key", v.Locale)
			}
		}
	}
	if !sawPlaceholder {
		t.Errorf("no violation named the placeholder key syn.placeholder.key; got %v", violEmpty)
	}

	// With the placeholder key allowlisted, the identical catalogue must report
	// zero violations.
	allowWithPlaceholder := []i18nAllowEntry{
		{Key: "syn.placeholder.key", Reason: reasonTechnicalIdentifier, Justification: "inert synthetic test key"},
	}
	if v := detectI18nUntranslated(cat, allowWithPlaceholder); len(v) != 0 {
		t.Errorf("allowlisting the key did not suppress the violation; got %v", v)
	}
}

// --- AC-I18NGOV-003: allowlist has no orphan entries (015) ---

func TestI18nAllowlistNoOrphans(t *testing.T) {
	cat := shippedI18nCatalogue(t)
	orphans := detectI18nOrphans(cat, i18nUntranslatedAllowlist)
	for _, o := range orphans {
		t.Errorf("orphan allowlist entry: key=%s reason=%s — key is absent or no longer identical to en in any locale", o.Key, o.Reason)
	}
}

// --- AC-I18NGOV-004: allowlist shape — exact keys, closed taxonomy, bounded
// size (002..005, 016) ---

func TestI18nAllowlistShape(t *testing.T) {
	seen := make(map[string]struct{}, len(i18nUntranslatedAllowlist))
	for _, e := range i18nUntranslatedAllowlist {
		if _, dup := seen[e.Key]; dup {
			t.Errorf("duplicate allowlist entry for key %q", e.Key)
		}
		seen[e.Key] = struct{}{}

		// No wildcard / glob pattern in the key (REQ-I18NGOV-004). Only the
		// unambiguous glob wildcards '*' and '?' are rejected: '.', '+', '[',
		// ']' appear in legitimate exact catalogue keys (plan §B4 —
		// f.report.format.opt.html+md, f.model.opt.*[1m]) and are not pattern
		// indicators in this dotted-key namespace.
		for _, bad := range []string{"*", "?"} {
			if strings.Contains(e.Key, bad) {
				t.Errorf("allowlist key %q contains forbidden wildcard %q", e.Key, bad)
			}
		}

		// No lang.opt.* key in the allowlist (REQ-I18NGOV-014).
		if strings.HasPrefix(e.Key, i18nLangOptPrefix) {
			t.Errorf("allowlist carries a lang.opt.* key %q — the endonym family is handled structurally", e.Key)
		}

		// Non-empty justification (REQ-I18NGOV-002).
		if strings.TrimSpace(e.Justification) == "" {
			t.Errorf("allowlist entry for %q has an empty justification", e.Key)
		}

		// Reason is a member of the closed taxonomy (REQ-I18NGOV-003).
		if _, ok := i18nAllowReasons[e.Reason]; !ok {
			t.Errorf("allowlist entry for %q has reason %q outside the closed taxonomy %v", e.Key, e.Reason, i18nAllowReasons)
		}
	}

	// Entry cap (REQ-I18NGOV-016).
	if n := len(i18nUntranslatedAllowlist); n > i18nMaxAllowlistEntries {
		t.Errorf("allowlist has %d entries, exceeding the cap of %d", n, i18nMaxAllowlistEntries)
	}

	// The closed taxonomy is exactly the three documented reasons.
	wantReasons := map[i18nAllowReason]struct{}{
		reasonTechnicalIdentifier: {},
		reasonProperNoun:          {},
		reasonAcronym:             {},
	}
	if len(i18nAllowReasons) != len(wantReasons) {
		t.Errorf("closed taxonomy has %d reasons, expected %d", len(i18nAllowReasons), len(wantReasons))
	}
	for r := range wantReasons {
		if _, ok := i18nAllowReasons[r]; !ok {
			t.Errorf("closed taxonomy is missing reason %q", r)
		}
	}
}

// --- AC-I18NGOV-005: endonym family is correct by construction (012..014) ---

func TestI18nEndonymInvariants(t *testing.T) {
	cat := shippedI18nCatalogue(t)
	en := cat["en"]

	// REQ-I18NGOV-012: self-consistency — lang.opt.<L> in locale L equals its
	// en value (both render the endonym).
	for _, loc := range i18nLocaleOrder {
		key := i18nLangOptPrefix + loc
		enV, ok := en[key]
		if !ok {
			t.Errorf("en locale has no %q key — endonym baseline missing", key)
			continue
		}
		locV, ok := cat[loc][key]
		if !ok {
			t.Errorf("locale %q has no own %q key", loc, key)
			continue
		}
		if i18nNormalize(locV) != i18nNormalize(enV) {
			t.Errorf("endonym self-consistency failed: lang.opt.%s in %s = %q, en = %q", loc, loc, locV, enV)
		}
	}

	// REQ-I18NGOV-013: exonym distinctness — for a NON-en locale L and language
	// X != L, lang.opt.<X> in L differs from its en value (locale L renders the
	// exonym; en is the endonym baseline and is trivially equal to itself, so
	// it is excluded from this direction).
	for _, loc := range i18nNonEnLocales {
		for _, x := range i18nLocaleOrder {
			if x == loc {
				continue
			}
			key := i18nLangOptPrefix + x
			enV, ok := en[key]
			if !ok {
				continue
			}
			locV, ok := cat[loc][key]
			if !ok {
				t.Errorf("locale %q is missing %q (forward coverage is tested separately)", loc, key)
				continue
			}
			if i18nNormalize(locV) == i18nNormalize(enV) {
				t.Errorf("exonym distinctness failed: lang.opt.%s in %s = %q equals its en value %q", x, loc, locV, enV)
			}
		}
	}
}

// --- AC-I18NGOV-006: forward key coverage (018) ---

func TestI18nKeyCoverageForward(t *testing.T) {
	cat := shippedI18nCatalogue(t)
	en := cat["en"]
	for _, loc := range i18nNonEnLocales {
		for k := range en {
			if _, ok := cat[loc][k]; !ok {
				t.Errorf("locale %q is missing en key %q", loc, k)
			}
		}
	}
}

// --- AC-I18NGOV-007: reverse key coverage modulo the exempt registry
// (019, 020) ---

func TestI18nKeyCoverageReverse(t *testing.T) {
	cat := shippedI18nCatalogue(t)
	en := cat["en"]
	for _, loc := range i18nNonEnLocales {
		for k := range cat[loc] {
			if _, ok := en[k]; ok {
				continue
			}
			if !i18nMatchesExemptPrefix(k) {
				t.Errorf("locale %q defines key %q absent from en and matching no exempt prefix", loc, k)
			}
		}
	}
	// The registry must be non-empty and every member justified.
	if len(i18nEnExemptPrefixes) == 0 {
		t.Error("en-exempt prefix registry is empty")
	}
	for _, p := range i18nEnExemptPrefixes {
		if strings.TrimSpace(p.Prefix) == "" || strings.TrimSpace(p.Justification) == "" {
			t.Errorf("exempt-prefix registry entry %+v has an empty prefix or justification", p)
		}
	}
}

// i18nMatchesExemptPrefix reports whether key begins with any registered
// en-exempt prefix (REQ-I18NGOV-019/020).
func i18nMatchesExemptPrefix(key string) bool {
	for _, p := range i18nEnExemptPrefixes {
		if strings.HasPrefix(key, p.Prefix) {
			return true
		}
	}
	return false
}

// --- AC-I18NGOV-008: parser reads special-character keys (007) ---

func TestI18nParserSpecialKeys(t *testing.T) {
	cat := shippedI18nCatalogue(t)
	specials := []string{
		"f.report.format.opt.html+md",
		"f.model.opt.opus[1m]",
		"f.model.opt.sonnet[1m]",
		"f.model.opt.fable[1m]",
	}
	for _, k := range specials {
		v, ok := cat["en"][k]
		if !ok {
			t.Errorf("parser dropped special key %q from the en locale", k)
			continue
		}
		if strings.TrimSpace(v) == "" {
			t.Errorf("special key %q parsed to an empty value", k)
		}
	}
}

// --- AC-I18NGOV-009: governance contract and named owner are present
// (001, 006, 021, 023) ---

func TestI18nGovernanceContractPresent(t *testing.T) {
	// The allowlist artifact carries an inline governance contract naming its
	// location, entry format, who adds an entry, the reviewer assertion, and
	// the ruling procedure (REQ-I18NGOV-006/023). Those headings live in this
	// file's doc comment — assert their tokens are present so the contract
	// cannot be silently deleted.
	allowSrc := readTestSource(t, "i18n_untranslated_allowlist_test.go")
	for _, token := range []string{
		"Governance contract",
		"Entry form",
		"Who adds",
		"Reviewer assertion",
		"Ruling procedure",
	} {
		if !strings.Contains(allowSrc, token) {
			t.Errorf("allowlist governance contract is missing the %q section", token)
		}
	}

	// The i18n.js header names the allowlist artifact path as the owning
	// governance surface (REQ-I18NGOV-021).
	dict := readEmbeddedAsset(t, "i18n.js")
	if !strings.Contains(dict, "i18n_untranslated_allowlist") {
		t.Error("i18n.js header does not name the i18n_untranslated_allowlist governance surface")
	}
}

// readTestSource reads a test-file's own source so a test can assert on its
// doc-comment contract without duplicating the contract text.
func readTestSource(t *testing.T, name string) string {
	t.Helper()
	src, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read test source %q: %v", name, err)
	}
	return string(src)
}

// --- AC-I18NGOV-011 is structurally asserted by `go build ./...` and
// `GOOS=windows GOARCH=amd64 go build ./...` both succeeding (no production
// symbol added); the run-phase report quotes their verbatim output. ---
