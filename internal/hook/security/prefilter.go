package security

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// SPEC-SEC-SCAN-SURFACE-001 M2 — the literal pre-filter.
//
// The pre-write gate's only observable output is the deny, and a deny requires
// an `error`-severity finding. So a payload that cannot possibly produce one
// need not be scanned at all. This file derives, from the resolved ruleset, a
// per-language set of literal substrings that are MANDATORY for at least one
// error-severity rule to match. When none of a language's tokens occurs in the
// payload, no error rule can match it and the scan is skipped.
//
// SOUNDNESS INVARIANT: a token is admitted only when the rule cannot match
// without it. `underivable` therefore always means "scan anyway", never "skip":
// an unrecognized rule shape, an unreadable ruleset, or a rule whose language
// cannot be attributed marks the whole language (or the whole set) underivable
// and escalates. A pre-filter that suppresses a deny is a correctness failure,
// not an optimization.
//
// The extraction rules implemented here are spec.md §C.2's table, row for row.

// Prefilter is one language's pre-filter.
type Prefilter struct {
	// Tokens is the sorted, de-duplicated set of mandatory literal substrings.
	// Meaningful only when Derivable. An empty Tokens with Derivable=true means
	// the language carries no error-severity rule at all, so nothing about it
	// can deny.
	Tokens []string
	// Derivable reports whether every error-severity rule for this language
	// yielded at least one mandatory literal token. False means escalate.
	Derivable bool
}

// PrefilterSet is the derivation result for a whole ruleset.
//
// Known=false is the fail-open state: the configuration could not be read,
// parsed, walked, or attributed to languages. No skip is ever taken from it.
type PrefilterSet struct {
	Known      bool
	ByLanguage map[string]Prefilter
}

// CanSkip reports whether a payload in the given language can be skipped
// because no error-severity rule could match it.
//
// Every uncertainty answers false: an unknown derivation, an empty language, a
// language absent from the derivation, and an underivable language all
// escalate.
func (s PrefilterSet) CanSkip(language, content string) bool {
	if !s.Known || language == "" {
		return false
	}
	pf, ok := s.ByLanguage[language]
	if !ok || !pf.Derivable {
		return false
	}
	for _, token := range pf.Tokens {
		if strings.Contains(content, token) {
			return false
		}
	}
	return true
}

// ruleDoc is one parsed rule document: the fields the derivation reads.
type ruleDoc struct {
	ID       string    `yaml:"id"`
	Language string    `yaml:"language"`
	Severity string    `yaml:"severity"`
	Rule     yaml.Node `yaml:"rule"`
}

// DerivePrefilters resolves the rule documents reachable from an
// ALREADY-RESOLVED sgconfig path and derives the per-language pre-filter.
//
// configPath is the value the caller already holds (LanguageCoverage.ConfigPath),
// so this performs no configuration discovery of its own. Every failure path
// returns Known=false.
func DerivePrefilters(configPath string) PrefilterSet {
	docs, err := loadRuleDocs(configPath)
	if err != nil {
		return PrefilterSet{}
	}
	return derivePrefilters(docs)
}

// loadRuleDocs reads every rule document under the configuration's ruleDirs.
func loadRuleDocs(configPath string) ([]ruleDoc, error) {
	if configPath == "" {
		return nil, errors.New("no rules configuration resolved")
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config sgConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	// An inline `rules:` map carries a shape this derivation does not read.
	// Rather than derive a partial token set from it, escalate.
	if len(config.Rules) > 0 {
		return nil, errors.New("inline rules are not derivable")
	}

	base := filepath.Dir(configPath)
	var docs []ruleDoc
	for _, dir := range config.RuleDirs {
		ruleDir := dir
		if !filepath.IsAbs(ruleDir) {
			ruleDir = filepath.Join(base, ruleDir)
		}
		collected, err := collectRuleDocs(ruleDir)
		if err != nil {
			return nil, err
		}
		docs = append(docs, collected...)
	}
	if len(docs) == 0 {
		return nil, errors.New("no rule documents found")
	}
	return docs, nil
}

// collectRuleDocs walks one rule directory, decoding every YAML document.
func collectRuleDocs(ruleDir string) ([]ruleDoc, error) {
	info, err := os.Stat(ruleDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("ruleDir is not a directory: " + ruleDir)
	}

	var docs []ruleDoc
	walkErr := filepath.WalkDir(ruleDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".yml", ".yaml":
		default:
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		// Rule files are multi-document YAML: the shipped ruleset packs one
		// document per language into a single file.
		dec := yaml.NewDecoder(file)
		for {
			var doc ruleDoc
			decErr := dec.Decode(&doc)
			if errors.Is(decErr, io.EOF) {
				return nil
			}
			if decErr != nil {
				return decErr
			}
			docs = append(docs, doc)
		}
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return docs, nil
}

// derivePrefilters is the pure derivation: parsed rule documents in, the
// per-language pre-filter out. It performs no I/O, which is what makes the
// extraction table directly testable against synthetic rule shapes.
func derivePrefilters(docs []ruleDoc) PrefilterSet {
	if len(docs) == 0 {
		return PrefilterSet{}
	}

	tokens := map[string]map[string]struct{}{}
	underivable := map[string]bool{}

	for i := range docs {
		doc := &docs[i]
		language := strings.TrimSpace(doc.Language)
		isError := strings.TrimSpace(doc.Severity) == string(SeverityError)
		if language == "" {
			// A rule that cannot be attributed to a language could deny in any
			// of them, so nothing can be skipped safely.
			if isError {
				return PrefilterSet{}
			}
			continue
		}
		if _, ok := tokens[language]; !ok {
			tokens[language] = map[string]struct{}{}
		}
		if !isError {
			// Only error-severity rules can produce a deny, so only they
			// contribute tokens (AC-SSS-008). The language is still registered
			// above: a language carrying only warning rules can never deny.
			continue
		}
		ruleTokens, ok := deriveRuleTokens(&doc.Rule)
		if !ok {
			underivable[language] = true
			continue
		}
		for _, t := range ruleTokens {
			tokens[language][t] = struct{}{}
		}
	}

	set := PrefilterSet{Known: true, ByLanguage: make(map[string]Prefilter, len(tokens))}
	for language, seen := range tokens {
		pf := Prefilter{Derivable: !underivable[language]}
		if pf.Derivable {
			pf.Tokens = make([]string, 0, len(seen))
			for t := range seen {
				pf.Tokens = append(pf.Tokens, t)
			}
			sort.Strings(pf.Tokens)
		}
		set.ByLanguage[language] = pf
	}
	return set
}

// derivableRuleKeys are the rule keys the extraction table covers. Any other
// key — inside:, has:, follows:, precedes:, not:, matches: — marks the rule
// underivable, because its effect on what the rule requires is not modelled
// here and unknown is never treated as absent.
var derivableRuleKeys = map[string]bool{
	"pattern": true,
	"all":     true,
	"any":     true,
	"kind":    true,
	"regex":   true,
}

// deriveRuleTokens applies spec.md §C.2's table to one rule node, returning the
// mandatory literal tokens and whether the rule is derivable at all.
//
// The keys of a rule node are a CONJUNCTION: every one must hold for the rule
// to match, so tokens from any single derivable conjunct are mandatory, and
// unioning several conjuncts' tokens only makes a skip harder to reach.
func deriveRuleTokens(node *yaml.Node) ([]string, bool) {
	fields, ok := mappingFields(node)
	if !ok || len(fields) == 0 {
		return nil, false
	}
	for key := range fields {
		if !derivableRuleKeys[key] {
			return nil, false
		}
	}

	var tokens []string
	derivable := false

	// `pattern:` — the literal runs outside any metavariable.
	if n, ok := fields["pattern"]; ok {
		if n.Kind != yaml.ScalarNode {
			return nil, false // the object form of pattern is not modelled
		}
		if token, found := patternToken(n.Value); found {
			tokens = append(tokens, token)
			derivable = true
		}
	}

	// `regex:` — including the `kind:` + `regex:` conjunction, where `kind`
	// narrows and never widens the match set, so the regex's tokens stay
	// mandatory.
	if n, ok := fields["regex"]; ok {
		if n.Kind != yaml.ScalarNode {
			return nil, false
		}
		if regexToks, found := regexTokens(n.Value); found {
			tokens = append(tokens, regexToks...)
			derivable = true
		}
	}

	// `all:` — every child must match, so any child's mandatory token is
	// mandatory for the composite.
	if n, ok := fields["all"]; ok {
		if n.Kind != yaml.SequenceNode || len(n.Content) == 0 {
			return nil, false
		}
		for _, child := range n.Content {
			if childToks, found := deriveRuleTokens(child); found {
				tokens = append(tokens, childToks...)
				derivable = true
			}
		}
	}

	// `any:` — a disjunction. A single tokenless branch can match with none of
	// the other branches' tokens present, so one such branch makes the whole
	// composite underivable.
	if n, ok := fields["any"]; ok {
		if n.Kind != yaml.SequenceNode || len(n.Content) == 0 {
			return nil, false
		}
		for _, child := range n.Content {
			childToks, found := deriveRuleTokens(child)
			if !found {
				return nil, false
			}
			tokens = append(tokens, childToks...)
		}
		derivable = true
	}

	// `kind:` alone contributes no token: a node kind is not a literal. When it
	// is the only conjunct the rule is underivable, which is what a false
	// `derivable` here reports.
	return tokens, derivable
}

// mappingFields flattens a YAML mapping node into key → value.
func mappingFields(node *yaml.Node) (map[string]*yaml.Node, bool) {
	if node == nil {
		return nil, false
	}
	n := node
	if n.Kind == yaml.DocumentNode && len(n.Content) == 1 {
		n = n.Content[0]
	}
	if n.Kind != yaml.MappingNode {
		return nil, false
	}
	fields := make(map[string]*yaml.Node, len(n.Content)/2)
	for i := 0; i+1 < len(n.Content); i += 2 {
		fields[n.Content[i].Value] = n.Content[i+1]
	}
	return fields, true
}

var (
	// metavarRe matches ast-grep metavariables: $NAME, $$$NAME, $_, and a bare
	// $$$.
	metavarRe = regexp.MustCompile(`\$\$\$[A-Za-z_][A-Za-z0-9_]*|\$\$\$|\$[A-Za-z_][A-Za-z0-9_]*`)
	// identChainRe matches an identifier, optionally dotted (exec.Command).
	identChainRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*`)
	// inlineFlagRe matches a regex group opening with a flag or a construct
	// this extractor does not model, e.g. (?i), (?s), (?m), (?P<name>).
	inlineFlagRe = regexp.MustCompile(`\(\?[a-zA-Z]`)
)

// patternToken extracts one mandatory literal token from an ast-grep pattern:
// the longest identifier chain in the pattern's literal runs, i.e. the runs
// left once every metavariable is removed.
//
// Exactly one token is taken rather than every literal run. Soundness needs
// only that each admitted token be mandatory — a pattern match reproduces its
// literal text — and the longest chain is both mandatory and the most selective
// of the candidates. Admitting more tokens would only make a skip rarer.
//
// A pattern with no identifier chain at all (one that is entirely
// metavariables and punctuation) is not derivable.
func patternToken(pattern string) (string, bool) {
	stripped := metavarRe.ReplaceAllString(pattern, " ")
	best := ""
	for _, candidate := range identChainRe.FindAllString(stripped, -1) {
		if len(candidate) > len(best) {
			best = candidate
		}
	}
	if best == "" {
		return "", false
	}
	return best, true
}

// regexTokens extracts the mandatory literal tokens from a rule's `regex:`.
//
//   - an inline flag makes a case-sensitive literal token non-mandatory, so it
//     is underivable
//   - a top-level alternation contributes the union of its branches' mandatory
//     literal prefixes, and is underivable if any branch has none
//   - otherwise the regex contributes its own mandatory literal prefix
func regexTokens(expr string) ([]string, bool) {
	if inlineFlagRe.MatchString(expr) {
		return nil, false
	}
	if branches, ok := topLevelAlternation(expr); ok {
		tokens := make([]string, 0, len(branches))
		for _, branch := range branches {
			prefix := literalPrefix(branch)
			if prefix == "" {
				return nil, false
			}
			tokens = append(tokens, prefix)
		}
		return tokens, true
	}
	prefix := literalPrefix(expr)
	if prefix == "" {
		return nil, false
	}
	return []string{prefix}, true
}

// topLevelAlternation returns the branches of the expression's outermost
// alternation: either a bare depth-0 `a|b`, or the first depth-0 group whose
// content alternates, which is the shape the shipped credential rule uses.
func topLevelAlternation(expr string) ([]string, bool) {
	if parts := splitAlternation(expr); len(parts) > 1 {
		return parts, true
	}

	inClass := false
	depth := 0
	start := -1
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		if c == '\\' {
			i++
			continue
		}
		if inClass {
			if c == ']' {
				inClass = false
			}
			continue
		}
		switch c {
		case '[':
			inClass = true
		case '(':
			if depth == 0 {
				start = i
			}
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
			if depth == 0 && start >= 0 {
				inner := strings.TrimPrefix(expr[start+1:i], "?:")
				if parts := splitAlternation(inner); len(parts) > 1 {
					return parts, true
				}
				start = -1
			}
		}
	}
	return nil, false
}

// splitAlternation splits on `|` at nesting depth 0 and outside a character
// class.
func splitAlternation(expr string) []string {
	var parts []string
	depth := 0
	inClass := false
	last := 0
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		if c == '\\' {
			i++
			continue
		}
		if inClass {
			if c == ']' {
				inClass = false
			}
			continue
		}
		switch c {
		case '[':
			inClass = true
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '|':
			if depth == 0 {
				parts = append(parts, expr[last:i])
				last = i + 1
			}
		}
	}
	if len(parts) == 0 {
		return nil
	}
	return append(parts, expr[last:])
}

// quantifier classifies what follows an element in a regex.
type quantifier int

const (
	quantNone quantifier = iota
	// quantOptional means the element can be absent from a match.
	quantOptional
	// quantMandatory means the element occurs at least once, but what follows
	// it is no longer at a fixed offset from it.
	quantMandatory
)

// quantAt classifies the quantifier at index i, if any.
func quantAt(expr string, i int) quantifier {
	if i >= len(expr) {
		return quantNone
	}
	switch expr[i] {
	case '?', '*':
		return quantOptional
	case '+':
		return quantMandatory
	case '{':
		// {0...} makes the element optional; any other lower bound does not.
		if i+1 < len(expr) && expr[i+1] == '0' {
			return quantOptional
		}
		return quantMandatory
	}
	return quantNone
}

// skipQuantifier advances past a quantifier at i, brace form included.
func skipQuantifier(expr string, i int) int {
	if i >= len(expr) {
		return i
	}
	switch expr[i] {
	case '?', '*', '+':
		return i + 1
	case '{':
		for j := i; j < len(expr); j++ {
			if expr[j] == '}' {
				return j + 1
			}
		}
	}
	return i
}

// skipClass returns the index just past a character class opened at i, or -1
// when the class is unterminated.
func skipClass(expr string, i int) int {
	for j := i + 1; j < len(expr); j++ {
		if expr[j] == '\\' {
			j++
			continue
		}
		if expr[j] == ']' {
			return j + 1
		}
	}
	return -1
}

// literalPrefix returns the leading run of literal characters a match of expr
// must contain contiguously, or "" when there is none.
//
// Two properties keep the result mandatory. A single-width non-literal element
// (a character class, a `.`, a class escape) is skipped only while no literal
// has been collected yet, because a literal run interrupted by one is no longer
// contiguous. And an element the quantifier makes optional terminates the run
// without contributing, so an optional or quantified leading literal yields ""
// — the underivable case of spec.md §C.2.
func literalPrefix(expr string) string {
	var b strings.Builder
	i := 0
	for i < len(expr) {
		c := expr[i]
		switch c {
		case '^':
			// An anchor, never a literal.
			if b.Len() > 0 {
				return b.String()
			}
			i++
		case '$', '|', ')', '(', '*', '+', '?', '{', '}', ']':
			return b.String()
		case '[':
			if b.Len() > 0 {
				return b.String()
			}
			next := skipClass(expr, i)
			if next < 0 {
				return b.String()
			}
			i = skipQuantifier(expr, next)
		case '.':
			if b.Len() > 0 {
				return b.String()
			}
			i = skipQuantifier(expr, i+1)
		case '\\':
			if i+1 >= len(expr) {
				return b.String()
			}
			escaped := expr[i+1]
			if isLetter(escaped) || isDigit(escaped) {
				// A class escape (\d, \w, \s, \b) or a back-reference: not a
				// literal character.
				if b.Len() > 0 {
					return b.String()
				}
				i = skipQuantifier(expr, i+2)
				continue
			}
			switch quantAt(expr, i+2) {
			case quantOptional:
				return b.String()
			case quantMandatory:
				b.WriteByte(escaped)
				return b.String()
			}
			b.WriteByte(escaped)
			i += 2
		default:
			switch quantAt(expr, i+1) {
			case quantOptional:
				return b.String()
			case quantMandatory:
				b.WriteByte(c)
				return b.String()
			}
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
