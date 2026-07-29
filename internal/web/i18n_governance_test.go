package web

import (
	"os"
	"sort"
	"strings"
	"testing"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// SPEC-I18N-GOVERNANCE-001 — governance detector + checks (M2/M3/M4).
//
// All functions here are pure (REQ-I18NGOV-008): the detector and orphan
// check take a parsed catalogue and an allowlist as inputs and return a
// violation slice, so a caller can supply a synthetic catalogue + allowlist
// without touching the shipped files (the negative control, AC-I18NGOV-002).

// i18nLocales is the fixed locale set, `en` first so detectors can use it as
// the reference baseline.
var i18nLocales = []string{"en", "ko", "ja", "zh"}

// nonEnLocales is the detector's iteration set (every locale compared to en).
var nonEnLocales = []string{"ko", "ja", "zh"}

// parseLocaleKV parses one locale block (the raw slice produced by
// localeBlocks) into a key-to-value map. It reads keys containing `+`, `[`,
// and `]` (B4 special-character keys) and tolerates backslash-escaped quotes
// in values (a future translation may introduce a quotation mark).
func parseLocaleKV(block string) map[string]string {
	out := map[string]string{}
	for _, raw := range strings.Split(block, "\n") {
		s := strings.TrimLeft(raw, " \t")
		if len(s) == 0 || s[0] != '"' {
			continue
		}
		key, rest, ok := parseQuoted(s)
		if !ok {
			continue
		}
		colon := strings.Index(rest, ":")
		if colon < 0 {
			continue
		}
		valTail := strings.TrimLeft(rest[colon+1:], " \t")
		// strip an optional trailing comma so the parser is robust to the
		// object-literal's last entry form.
		valTail = strings.TrimRight(valTail, ", \t")
		val, _, ok := parseQuoted(valTail)
		if !ok {
			continue
		}
		out[key] = val
	}
	return out
}

// parseQuoted reads a double-quoted string starting at s[0], honoring
// backslash escapes, and returns the decoded value, the remainder after the
// closing quote, and ok.
func parseQuoted(s string) (val, rest string, ok bool) {
	if len(s) == 0 || s[0] != '"' {
		return "", "", false
	}
	var b strings.Builder
	i := 1
	for i < len(s) {
		c := s[i]
		if c == '\\' && i+1 < len(s) {
			// keep the escaped character literally (handles \" \\ \n etc.)
			b.WriteByte(s[i+1])
			i += 2
			continue
		}
		if c == '"' {
			return b.String(), s[i+1:], true
		}
		b.WriteByte(c)
		i++
	}
	return "", "", false
}

// parseI18nCatalogue parses the full i18n.js dictionary into a per-locale
// key-value map.
func parseI18nCatalogue(t *testing.T, dict string) map[string]map[string]string {
	t.Helper()
	blocks := localeBlocks(t, dict)
	out := make(map[string]map[string]string, len(i18nLocales))
	for _, loc := range i18nLocales {
		out[loc] = parseLocaleKV(blocks[loc])
	}
	return out
}

// i18nViolation is one detector finding: a non-en value whose normalized form
// equals its en counterpart, for a key that is neither allowlisted nor a
// lang.opt.* member; OR an allowlist orphan (key absent or never identical).
type i18nViolation struct {
	key    string
	locale string
	reason string
}

// foldCaser is the Unicode case-folding caser used by normalizeValue.
var foldCaser = cases.Fold()

// normalizeValue applies the REQ-I18NGOV-009 normalization: Unicode NFC
// composition, surrounding-whitespace trim, and case folding.
func normalizeValue(s string) string {
	nfc := norm.NFC.String(s)
	trimmed := strings.TrimSpace(nfc)
	return foldCaser.String(trimmed)
}

// isLangOpt reports whether key is a member of the lang.opt.* family, which
// is excluded from the generic identity comparison (REQ-I18NGOV-011) and
// governed instead by the endonym invariants.
func isLangOpt(key string) bool {
	return strings.HasPrefix(key, "lang.opt.")
}

// detectUntranslated is the pure detector (REQ-I18NGOV-008/009/010/011). It
// compares every non-en value to its en counterpart under normalization,
// excluding lang.opt.* and every allowlisted key, and returns one violation
// per (key, locale) that remains identical without justification.
func detectUntranslated(cat map[string]map[string]string, allow []untranslatedAllowEntry) []i18nViolation {
	allowed := make(map[string]bool, len(allow))
	for _, e := range allow {
		allowed[e.key] = true
	}
	en := cat["en"]
	var out []i18nViolation
	for _, loc := range nonEnLocales {
		block, ok := cat[loc]
		if !ok {
			continue
		}
		keys := make([]string, 0, len(block))
		for k := range block {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if isLangOpt(k) || allowed[k] {
				continue
			}
			ev, ok := en[k]
			if !ok {
				continue // reverse-coverage is a separate check
			}
			if normalizeValue(block[k]) == normalizeValue(ev) {
				out = append(out, i18nViolation{key: k, locale: loc, reason: "untranslated"})
			}
		}
	}
	return out
}

// detectAllowlistOrphans is the pure orphan check (REQ-I18NGOV-015). An entry
// is an orphan when its key is absent from the catalogue OR its value is not
// byte-identical to en in ANY non-en locale (so a stale exemption cannot
// survive the deletion or translation of its key). The reported reason names
// the entry's taxonomy classification via allowlistReasonFor.
func detectAllowlistOrphans(cat map[string]map[string]string, allow []untranslatedAllowEntry) []i18nViolation {
	en := cat["en"]
	var out []i18nViolation
	for _, e := range allow {
		r, _ := allowlistReasonFor(e.key)
		if _, ok := en[e.key]; !ok {
			out = append(out, i18nViolation{key: e.key, locale: "en", reason: "orphan [" + string(r) + "]: key absent from en catalogue"})
			continue
		}
		identicalSomewhere := false
		for _, loc := range nonEnLocales {
			if block, ok := cat[loc]; ok {
				if v, present := block[e.key]; present && v == en[e.key] {
					identicalSomewhere = true
					break
				}
			}
		}
		if !identicalSomewhere {
			out = append(out, i18nViolation{key: e.key, locale: "*", reason: "orphan [" + string(r) + "]: value not identical to en in any locale"})
		}
	}
	return out
}

// --- AC-I18NGOV-001: real catalogue reports zero violations ---

func TestI18nUntranslatedValues(t *testing.T) {
	cat := parseI18nCatalogue(t, readEmbeddedAsset(t, "i18n.js"))
	violations := detectUntranslated(cat, untranslatedAllowlist)
	if len(violations) != 0 {
		for _, v := range violations {
			t.Errorf("untranslated value: key=%q locale=%q en=%q %q", v.key, v.locale, cat["en"][v.key], cat[v.locale][v.key])
		}
	}
}

// --- AC-I18NGOV-002: negative control (the allowlist is a gate, not a stamp) ---

func TestI18nUntranslatedDetectorNegativeControl(t *testing.T) {
	// Synthetic catalogue with an inert placeholder key whose ko value equals
	// its en value (C2: no secret-shaped or key-shaped literal).
	placeholderKey := "synthetic.placeholder.word"
	synthetic := map[string]map[string]string{
		"en": {placeholderKey: "Baseline"},
		"ko": {placeholderKey: "Baseline"},
		"ja": {placeholderKey: "다름"}, // translated in ja, so only ko is under test
		"zh": {placeholderKey: "不同"},
	}

	// Without the key in the allowlist: the detector MUST flag it.
	if got := detectUntranslated(synthetic, nil); len(got) == 0 {
		t.Fatalf("detector reported no violation for a non-allowlisted untranslated key — the allowlist is a rubber stamp")
	} else {
		found := false
		for _, v := range got {
			if v.key == placeholderKey && v.locale == "ko" {
				found = true
			}
		}
		if !found {
			t.Fatalf("detector did not name the placeholder key in ko; got %+v", got)
		}
	}

	// With the key allowlisted: the identical input MUST report zero.
	allow := []untranslatedAllowEntry{{key: placeholderKey, reason: reasonTechnicalIdentifier, justification: "inert fixture"}}
	if got := detectUntranslated(synthetic, allow); len(got) != 0 {
		t.Fatalf("detector reported %d violation(s) for an allowlisted key; expected 0: %+v", len(got), got)
	}
}

// --- AC-I18NGOV-003: allowlist has no orphan entries ---

func TestI18nAllowlistNoOrphans(t *testing.T) {
	cat := parseI18nCatalogue(t, readEmbeddedAsset(t, "i18n.js"))
	if got := detectAllowlistOrphans(cat, untranslatedAllowlist); len(got) != 0 {
		for _, v := range got {
			t.Errorf("orphan allowlist entry: key=%q reason=%s", v.key, v.reason)
		}
	}
}

// --- AC-I18NGOV-004: allowlist shape (exact keys, closed taxonomy, bound) ---

func TestI18nAllowlistShape(t *testing.T) {
	if len(untranslatedAllowlist) > i18nAllowlistMaxEntries {
		t.Errorf("allowlist has %d entries; cap is %d", len(untranslatedAllowlist), i18nAllowlistMaxEntries)
	}
	valid := map[untranslatedReason]bool{reasonTechnicalIdentifier: true, reasonProperNoun: true, reasonAcronym: true}
	seen := map[string]bool{}
	for _, e := range untranslatedAllowlist {
		if e.key == "" {
			t.Error("allowlist entry has an empty key")
		}
		if strings.ContainsAny(e.key, "*?^$()|\\") {
			// Glob (`*`/`?`) or regex-control chars that only make sense in a
			// pattern. `.` is the catalogue's key separator (present in every
			// key) and `+`/`[`/`]` are literal characters in the four special
			// keys (B4), so they are not flagged here.
			t.Errorf("allowlist entry %q looks wildcard/glob/regex-shaped", e.key)
		}
		if isLangOpt(e.key) {
			t.Errorf("allowlist must not carry a lang.opt.* key: %q", e.key)
		}
		if e.justification == "" {
			t.Errorf("allowlist entry %q has an empty justification", e.key)
		}
		if !valid[e.reason] {
			t.Errorf("allowlist entry %q has reason %q outside the closed taxonomy", e.key, e.reason)
		}
		if seen[e.key] {
			t.Errorf("duplicate allowlist entry: %q", e.key)
		}
		seen[e.key] = true
	}
}

// --- AC-I18NGOV-005: endonym family is correct by construction ---

func TestI18nEndonymInvariants(t *testing.T) {
	cat := parseI18nCatalogue(t, readEmbeddedAsset(t, "i18n.js"))
	en := cat["en"]
	// No lang.opt.* key is allowlisted (REQ-I18NGOV-014).
	for _, e := range untranslatedAllowlist {
		if isLangOpt(e.key) {
			t.Fatalf("a lang.opt.* key is in the allowlist: %q", e.key)
		}
	}
	// Self-consistency: lang.opt.<L> in L equals lang.opt.<L> in en.
	for _, loc := range i18nLocales {
		k := "lang.opt." + loc
		enVal, ok := en[k]
		if !ok {
			t.Fatalf("en has no %q key", k)
		}
		locVal, ok := cat[loc][k]
		if !ok {
			t.Fatalf("locale %q has no own %q key", loc, k)
		}
		if locVal != enVal {
			t.Errorf("endonym self-consistency broken: lang.opt.%s in %s = %q, en = %q", loc, loc, locVal, enVal)
		}
	}
	// Exonym distinctness: for a non-en locale L, lang.opt.<X> in L (X != L)
	// differs from en — L renders the exonym. (en renders every language as
	// its endonym, so it is not subject to this direction.)
	for _, loc := range nonEnLocales {
		for _, x := range i18nLocales {
			if x == loc {
				continue
			}
			k := "lang.opt." + x
			enVal := en[k]
			locVal, ok := cat[loc][k]
			if !ok {
				t.Errorf("locale %q is missing %q", loc, k)
				continue
			}
			if locVal == enVal {
				t.Errorf("exonym distinctness broken: lang.opt.%s in %s equals en (%q)", x, loc, enVal)
			}
		}
	}
}

// --- AC-I18NGOV-006: forward key coverage ---

func TestI18nKeyCoverageForward(t *testing.T) {
	cat := parseI18nCatalogue(t, readEmbeddedAsset(t, "i18n.js"))
	en := cat["en"]
	for _, loc := range nonEnLocales {
		block := cat[loc]
		var missing []string
		for k := range en {
			if _, ok := block[k]; !ok {
				missing = append(missing, k)
			}
		}
		sort.Strings(missing)
		if len(missing) != 0 {
			t.Errorf("locale %q is missing %d en key(s): %v", loc, len(missing), missing)
		}
	}
}

// --- AC-I18NGOV-007: reverse key coverage modulo the exempt registry ---

func TestI18nKeyCoverageReverse(t *testing.T) {
	cat := parseI18nCatalogue(t, readEmbeddedAsset(t, "i18n.js"))
	en := cat["en"]
	for _, loc := range nonEnLocales {
		block := cat[loc]
		var extra []string
		for k := range block {
			if _, ok := en[k]; ok {
				continue
			}
			if isEnExempt(k) {
				continue
			}
			extra = append(extra, k)
		}
		sort.Strings(extra)
		if len(extra) != 0 {
			t.Errorf("locale %q has %d reverse-extra key(s) not in en and not exempt: %v", loc, len(extra), extra)
		}
	}
	// Every registry member must carry a justification.
	for _, e := range enExemptPrefixRegistry {
		if e.justification == "" {
			t.Errorf("en-exempt prefix %q has no justification", e.prefix)
		}
	}
}

// --- AC-I18NGOV-008: parser reads special-character keys ---

func TestI18nParserSpecialKeys(t *testing.T) {
	cat := parseI18nCatalogue(t, readEmbeddedAsset(t, "i18n.js"))
	specials := []string{
		"f.report.format.opt.html+md",
		"f.model.opt.opus[1m]",
		"f.model.opt.sonnet[1m]",
		"f.model.opt.fable[1m]",
	}
	for _, loc := range i18nLocales {
		for _, k := range specials {
			v, ok := cat[loc][k]
			if !ok {
				t.Errorf("locale %q is missing special-character key %q (parser dropped it)", loc, k)
				continue
			}
			if v == "" {
				t.Errorf("locale %q special key %q parsed to an empty value", loc, k)
			}
		}
	}
}

// --- AC-I18NGOV-009: governance contract + owning surface pointer ---

func TestI18nGovernanceContractPresent(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	// The i18n.js header names the allowlist artifact path as the owning
	// governance surface (REQ-I18NGOV-021).
	if !strings.Contains(dict, "i18n_untranslated_allowlist") {
		t.Error("i18n.js header does not name the i18n_untranslated_allowlist governance surface")
	}
	// The allowlist artifact carries the inline governance contract (this
	// test's CWD is the package dir, so the filename resolves directly).
	allowSrc, err := os.ReadFile("i18n_untranslated_allowlist_test.go")
	if err != nil {
		t.Fatalf("cannot read allowlist artifact: %v", err)
	}
	allowStr := string(allowSrc)
	for _, tok := range []string{
		"Governance contract",
		"Location",
		"Entry format",
		"Who adds one",
		"Reviewer assertion",
		"Ruling procedure",
	} {
		if !strings.Contains(allowStr, tok) {
			t.Errorf("allowlist artifact governance contract is missing token %q", tok)
		}
	}
}
