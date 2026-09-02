package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// SPEC-SEC-SCAN-SURFACE-001 M2 — the derivation itself, observed in the package
// that owns it. The end-to-end scan-count consequences of the same derivation
// are observed in internal/hook (see pre_tool_scan_prefilter_test.go).

// shippedPrefilters derives the pre-filter from the ruleset the template
// distributes, copied into a temp project root.
func shippedPrefilters(t *testing.T) PrefilterSet {
	t.Helper()
	root := coverageShippedRoot(t)
	configPath := NewRuleManager().FindRulesConfig(root)
	if configPath == "" {
		t.Fatal("expected the shipped ruleset's sgconfig.yml to resolve")
	}
	set := DerivePrefilters(configPath)
	if !set.Known {
		t.Fatal("expected the shipped ruleset to derive a known pre-filter")
	}
	return set
}

// shippedRuleDocs loads every rule document of the shipped ruleset.
func shippedRuleDocs(t *testing.T) []ruleDoc {
	t.Helper()
	root := coverageShippedRoot(t)
	docs, err := loadRuleDocs(NewRuleManager().FindRulesConfig(root))
	if err != nil {
		t.Fatalf("load rule docs: %v", err)
	}
	return docs
}

// parseRuleNode parses a `rule:` body written inline in a test.
func parseRuleNode(t *testing.T, body string) *yaml.Node {
	t.Helper()
	var doc ruleDoc
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatalf("parse rule: %v", err)
	}
	return &doc.Rule
}

func hasToken(tokens []string, want string) bool {
	for _, t := range tokens {
		if t == want {
			return true
		}
	}
	return false
}

// TestPrefilterDerivationSeverityScope closes AC-SSS-008: the pre-filter is
// derived only from error-severity rules.
//
// The two halves are asserted separately: no token contributed exclusively by a
// non-error rule appears, and at least one token from every derivable
// error-severity rule does.
func TestPrefilterDerivationSeverityScope(t *testing.T) {
	set := shippedPrefilters(t)
	docs := shippedRuleDocs(t)

	// Half 1 — nothing a non-error rule alone contributes may appear. The
	// exclusion is computed from the ruleset rather than hardcoded: a token
	// carried by BOTH an error and a warning rule is legitimately present.
	errorTokens := map[string]map[string]struct{}{}
	warningOnly := map[string]map[string]struct{}{}
	for i := range docs {
		doc := &docs[i]
		lang := strings.TrimSpace(doc.Language)
		if lang == "" {
			continue
		}
		toks, ok := deriveRuleTokens(&doc.Rule)
		if !ok {
			continue
		}
		bucket := warningOnly
		if strings.TrimSpace(doc.Severity) == string(SeverityError) {
			bucket = errorTokens
		}
		if _, seen := bucket[lang]; !seen {
			bucket[lang] = map[string]struct{}{}
		}
		for _, tok := range toks {
			bucket[lang][tok] = struct{}{}
		}
	}

	for lang, toks := range warningOnly {
		for tok := range toks {
			if _, alsoError := errorTokens[lang][tok]; alsoError {
				continue
			}
			if hasToken(set.ByLanguage[lang].Tokens, tok) {
				t.Errorf("%s: token %q is contributed only by a non-error rule but appears in the pre-filter", lang, tok)
			}
		}
	}

	// The exclusion means nothing unless some non-error rule actually
	// contributes a token the error rules do not.
	distinct := 0
	for lang, toks := range warningOnly {
		for tok := range toks {
			if _, alsoError := errorTokens[lang][tok]; !alsoError {
				distinct++
			}
		}
	}
	if distinct == 0 {
		t.Fatal("no non-error rule contributes a token of its own; the exclusion above observes nothing")
	}
	t.Logf("non-error-only tokens excluded: %d", distinct)

	// Half 2 — every derivable error-severity rule contributes at least one of
	// its tokens to its language's set.
	for i := range docs {
		doc := &docs[i]
		if strings.TrimSpace(doc.Severity) != string(SeverityError) {
			continue
		}
		lang := strings.TrimSpace(doc.Language)
		toks, ok := deriveRuleTokens(&doc.Rule)
		if !ok || len(toks) == 0 {
			continue
		}
		present := false
		for _, tok := range toks {
			if hasToken(set.ByLanguage[lang].Tokens, tok) {
				present = true
				break
			}
		}
		if !present {
			t.Errorf("%s/%s: no token of this error-severity rule reached the pre-filter (rule tokens %q, set %q)",
				lang, doc.ID, toks, set.ByLanguage[lang].Tokens)
		}
	}

	// A specific, named exclusion, so a future ruleset that stops carrying a
	// warning rule cannot silently empty the computed exclusion above:
	// sec-log-injection-unsanitized is warning-severity and its token is
	// log.Printf.
	if hasToken(set.ByLanguage["go"].Tokens, "log.Printf") {
		t.Error("go: log.Printf comes from a warning-severity rule and must not be a pre-filter token")
	}
}

// TestPrefilterKindPlusRegexAlternation closes AC-SSS-009: the shipped
// ruleset's dominant shape — `kind:` + `regex:` with no `pattern:` — is
// derivable in all four covered languages, with the mandatory literal prefix of
// each alternation branch as its tokens.
func TestPrefilterKindPlusRegexAlternation(t *testing.T) {
	docs := shippedRuleDocs(t)
	want := []string{"AIza", "AKIA", "ghp_", "sk-", "xox"}

	seen := map[string]bool{}
	for i := range docs {
		doc := &docs[i]
		if doc.ID != "sec-hardcoded-credential" {
			continue
		}
		lang := strings.TrimSpace(doc.Language)
		seen[lang] = true

		toks, ok := deriveRuleTokens(&doc.Rule)
		if !ok {
			t.Errorf("%s: sec-hardcoded-credential marked underivable", lang)
			continue
		}
		got := append([]string(nil), toks...)
		sortStrings(got)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%s: tokens = %q, want %q", lang, got, want)
		}
	}

	for _, lang := range []string{"go", "python", "javascript", "typescript"} {
		if !seen[lang] {
			t.Errorf("%s: no sec-hardcoded-credential document found; the assertion above observed nothing for it", lang)
		}
	}
}

// TestPrefilterUnderivableShapes closes the derivation half of AC-SSS-010: an
// unrecognized rule shape marks its language underivable rather than
// contributing a partial token set.
func TestPrefilterUnderivableShapes(t *testing.T) {
	cases := []struct {
		name string
		rule string
	}{
		{
			name: "regex with no literal anchor",
			rule: "rule:\n  regex: '[0-9]+'\n",
		},
		{
			name: "any with one tokenless branch",
			rule: "rule:\n  any:\n    - pattern: os.system($CMD)\n    - pattern: $A($B)\n",
		},
		{
			name: "inside composite",
			rule: "rule:\n  pattern: os.system($CMD)\n  inside:\n    kind: function_definition\n",
		},
		{
			name: "kind alone",
			rule: "rule:\n  kind: string\n",
		},
		{
			name: "regex with an inline case-insensitive flag",
			rule: "rule:\n  regex: '(?i)^secret'\n",
		},
		{
			name: "regex whose leading literal is optional",
			rule: "rule:\n  regex: '^a?bc'\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if toks, ok := deriveRuleTokens(parseRuleNode(t, tc.rule)); ok {
				t.Fatalf("expected underivable, got tokens %q", toks)
			}

			// And the language it belongs to is marked underivable, so the
			// caller escalates rather than skipping.
			set := derivePrefilters([]ruleDoc{
				{ID: "probe", Language: "python", Severity: "error", Rule: *parseRuleNode(t, tc.rule)},
			})
			if !set.Known {
				// A wholly unknown set escalates too; either is fail-open.
				return
			}
			if set.ByLanguage["python"].Derivable {
				t.Errorf("python: expected underivable, got tokens %q", set.ByLanguage["python"].Tokens)
			}
			if set.CanSkip("python", "print('nothing dangerous here')\n") {
				t.Error("an underivable language must never permit a skip")
			}
		})
	}

	// Control: a recognized shape in the same position IS derivable, so the
	// failures above are the shape being rejected rather than the harness
	// rejecting everything.
	t.Run("control: recognized shape derives", func(t *testing.T) {
		set := derivePrefilters([]ruleDoc{
			{ID: "probe", Language: "python", Severity: "error",
				Rule: *parseRuleNode(t, "rule:\n  pattern: os.system($CMD)\n")},
		})
		pf := set.ByLanguage["python"]
		if !set.Known || !pf.Derivable || !hasToken(pf.Tokens, "os.system") {
			t.Fatalf("expected a derivable python pre-filter carrying os.system, got known=%v %+v", set.Known, pf)
		}
		if set.CanSkip("python", "os.system(cmd)\n") {
			t.Error("a payload carrying the token must not be skipped")
		}
		if !set.CanSkip("python", "print('hello')\n") {
			t.Error("a payload carrying no token should be skippable")
		}
	})
}

// TestPrefilterExtractionTable exercises the two extractors row by row against
// spec.md §C.2. These are the soundness-critical units: every token they admit
// is one the pre-filter will later treat as mandatory, so a token that is not
// actually mandatory would let a payload be skipped that a rule could match.
//
// The `want` column of the regex cases is read as: any string the expression
// matches contains this token as a contiguous substring.
func TestPrefilterExtractionTable(t *testing.T) {
	t.Run("regex", func(t *testing.T) {
		cases := []struct {
			expr string
			want []string // nil means underivable
		}{
			// The shipped credential shape, go and js/python spellings.
			{`^"(sk-|AKIA[0-9A-Z]{16}|ghp_[0-9A-Za-z]{36}|xox[baprs]-|AIza[0-9A-Za-z_-]{35})`,
				[]string{"sk-", "AKIA", "ghp_", "xox", "AIza"}},
			{`^["'](sk-|AKIA[0-9A-Z]{16})`, []string{"sk-", "AKIA"}},
			// A bare depth-0 alternation, no group.
			{`abc|def`, []string{"abc", "def"}},
			// No alternation: the mandatory literal prefix.
			{`^abc`, []string{"abc"}},
			{`^"https://api\.`, []string{`"https://api.`}},
			// A single-width non-literal element before any literal is skipped:
			// what follows it is still contiguous.
			{`^\d+abc`, []string{"abc"}},
			{`^[a-z]{3}-x`, []string{"-x"}},
			{`^sk-.*end`, []string{"sk-"}},
			// A quantifier ends the run; a mandatory one keeps the character it
			// applies to, an optional one drops it.
			{`a+b`, []string{"a"}},
			{`foo{2,3}bar`, []string{"foo"}},
			{`foo{0,3}bar`, []string{"fo"}},
			// Underivable: no literal anchor at all.
			{`[0-9]+`, nil},
			{`^$`, nil},
			// Underivable: an inline flag makes a case-sensitive token
			// non-mandatory.
			{`(?i)^secret`, nil},
			{`(?s)^secret`, nil},
			{`(?m)^secret`, nil},
			// Underivable: an optional or quantified leading literal can be
			// absent from a match entirely.
			{`^a?bc`, nil},
			{`^a*bc`, nil},
			{`^(foo)?bar`, nil},
			// Underivable: one alternation branch without a literal prefix.
			{`^(sk-|[0-9]+)`, nil},
		}

		for _, tc := range cases {
			got, ok := regexTokens(tc.expr)
			if tc.want == nil {
				if ok {
					t.Errorf("regexTokens(%q) = %q, want underivable", tc.expr, got)
				}
				continue
			}
			if !ok {
				t.Errorf("regexTokens(%q) underivable, want %q", tc.expr, tc.want)
				continue
			}
			gotSorted := append([]string(nil), got...)
			wantSorted := append([]string(nil), tc.want...)
			sortStrings(gotSorted)
			sortStrings(wantSorted)
			if strings.Join(gotSorted, ",") != strings.Join(wantSorted, ",") {
				t.Errorf("regexTokens(%q) = %q, want %q", tc.expr, got, tc.want)
			}
		}
	})

	t.Run("pattern", func(t *testing.T) {
		cases := []struct {
			pattern string
			want    string // "" means underivable
		}{
			{`exec.Command("sh", "-c", $CMD)`, "exec.Command"},
			{`md5.New()`, "md5.New"},
			{`template.HTML($USER_INPUT)`, "template.HTML"},
			{`SignedString([]byte("$HARDCODED"))`, "SignedString"},
			{`const $NAME = "sk-$$$REST"`, "const"},
			{`subprocess.call($CMD, shell=True)`, "subprocess.call"},
			{`os.system($CMD)`, "os.system"},
			// Entirely metavariables and punctuation: no literal token.
			{`$_, $ERR = $FUNC($$$ARGS)`, ""},
			{`$A($B)`, ""},
			{`$$$ITEMS`, ""},
		}

		for _, tc := range cases {
			got, ok := patternToken(tc.pattern)
			if tc.want == "" {
				if ok {
					t.Errorf("patternToken(%q) = %q, want underivable", tc.pattern, got)
				}
				continue
			}
			if !ok || got != tc.want {
				t.Errorf("patternToken(%q) = %q (ok=%v), want %q", tc.pattern, got, ok, tc.want)
			}
		}
	})

	// The composite rows of the table. The shipped ruleset carries an `any:`
	// but no `all:`, so the `all:` row is exercised synthetically rather than
	// left unobserved.
	t.Run("composite", func(t *testing.T) {
		composites := []struct {
			name string
			rule string
			want []string // nil means underivable
		}{
			{
				name: "all: every child must match, so a tokenless child is harmless",
				rule: "rule:\n  all:\n    - pattern: os.system($CMD)\n    - kind: call\n",
				want: []string{"os.system"},
			},
			{
				name: "all: union of the children's tokens",
				rule: "rule:\n  all:\n    - pattern: os.system($CMD)\n    - pattern: subprocess.run($CMD)\n",
				want: []string{"os.system", "subprocess.run"},
			},
			{
				name: "any: union when every branch yields a token",
				rule: "rule:\n  any:\n    - pattern: child_process.exec($CMD)\n    - pattern: cp.exec($CMD)\n",
				want: []string{"child_process.exec", "cp.exec"},
			},
			{
				name: "kind: + regex: is a conjunction, tokens from the regex",
				rule: "rule:\n  kind: string\n  regex: '^\"(sk-|AKIA[0-9A-Z]{16})'\n",
				want: []string{"sk-", "AKIA"},
			},
			{
				name: "empty all is underivable",
				rule: "rule:\n  all: []\n",
			},
			{
				name: "empty any is underivable",
				rule: "rule:\n  any: []\n",
			},
		}

		for _, tc := range composites {
			t.Run(tc.name, func(t *testing.T) {
				got, ok := deriveRuleTokens(parseRuleNode(t, tc.rule))
				if tc.want == nil {
					if ok {
						t.Fatalf("expected underivable, got %q", got)
					}
					return
				}
				if !ok {
					t.Fatalf("expected derivable with %q, got underivable", tc.want)
				}
				gotSorted := append([]string(nil), got...)
				wantSorted := append([]string(nil), tc.want...)
				sortStrings(gotSorted)
				sortStrings(wantSorted)
				if strings.Join(gotSorted, ",") != strings.Join(wantSorted, ",") {
					t.Errorf("tokens = %q, want %q", got, tc.want)
				}
			})
		}
	})

	t.Run("rule node that is not a mapping", func(t *testing.T) {
		if toks, ok := deriveRuleTokens(parseRuleNode(t, "rule: just-a-scalar\n")); ok {
			t.Errorf("expected underivable for a scalar rule body, got %q", toks)
		}
	})

	// Every token the regex table admits is checked against a string the
	// expression actually matches: the property the soundness invariant rests
	// on, asserted rather than reasoned about. The matching strings are built
	// rather than written out, so no fixture reads as a real credential.
	t.Run("admitted tokens are substrings of real matches", func(t *testing.T) {
		cases := []struct {
			expr    string
			matches []string
		}{
			{`^"(sk-|AKIA[0-9A-Z]{16})`, []string{
				`"` + "sk-" + strings.Repeat("Q", 6),
				`"` + "AKI" + "A" + strings.Repeat("Z", 16),
			}},
			{`^\d+abc`, []string{"1abc", "1234abc"}},
			{`^[a-z]{3}-x`, []string{"foo-x"}},
			{`a+b`, []string{"ab", "aaab"}},
			{`foo{2,3}bar`, []string{"foobar", "fooobar"}},
			{`foo{0,3}bar`, []string{"fobar", "fooobar"}},
		}

		for _, tc := range cases {
			tokens, ok := regexTokens(tc.expr)
			if !ok {
				t.Fatalf("regexTokens(%q) unexpectedly underivable", tc.expr)
			}
			for _, m := range tc.matches {
				hit := false
				for _, tok := range tokens {
					if strings.Contains(m, tok) {
						hit = true
						break
					}
				}
				if !hit {
					t.Errorf("%q: none of the tokens %q occurs in the matching string %q — the token set is not mandatory",
						tc.expr, tokens, m)
				}
			}
		}
	})
}

// TestPrefilterFailOpenOnUnreadableRuleset asserts the derivation's fail-open
// paths: every failure to read, parse, or attribute the ruleset yields an
// unknown set, which never permits a skip (REQ-SSS-009).
func TestPrefilterFailOpenOnUnreadableRuleset(t *testing.T) {
	t.Run("no config path", func(t *testing.T) {
		if set := DerivePrefilters(""); set.Known {
			t.Error("expected unknown for an empty config path")
		}
	})

	t.Run("malformed sgconfig", func(t *testing.T) {
		root := coverageShippedRoot(t)
		writeSGConfig(t, root, "ruleDirs: [go\n  : : not yaml\n")
		if set := DerivePrefilters(NewRuleManager().FindRulesConfig(root)); set.Known {
			t.Error("expected unknown for malformed YAML")
		}
	})

	t.Run("ruleDirs names a missing directory", func(t *testing.T) {
		root := coverageShippedRoot(t)
		writeSGConfig(t, root, "ruleDirs:\n  - does-not-exist\n")
		if set := DerivePrefilters(NewRuleManager().FindRulesConfig(root)); set.Known {
			t.Error("expected unknown for an unwalkable ruleDir")
		}
	})

	t.Run("error rule with no language", func(t *testing.T) {
		root := coverageShippedRoot(t)
		dir := filepath.Join(root, ".moai", "config", "astgrep-rules", "nolang")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		rule := "id: no-language-probe\nseverity: error\nmessage: probe\nrule:\n  pattern: probe($X)\n"
		if err := os.WriteFile(filepath.Join(dir, "probe.yml"), []byte(rule), 0o644); err != nil {
			t.Fatalf("write rule: %v", err)
		}
		writeSGConfig(t, root, "ruleDirs:\n  - nolang\n")
		if set := DerivePrefilters(NewRuleManager().FindRulesConfig(root)); set.Known {
			t.Error("expected unknown when an error-severity rule cannot be attributed to a language")
		}
	})

	t.Run("unknown set never skips", func(t *testing.T) {
		if (PrefilterSet{}).CanSkip("go", "package main\n") {
			t.Error("an unknown pre-filter set must never permit a skip")
		}
	})
}

// sortStrings sorts in place; a local helper so the test file needs no extra
// import for one call site.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
