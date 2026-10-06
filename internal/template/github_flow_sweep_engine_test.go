package template

// The classification engine of the github-flow sweep guard (SPEC-GITHUB-FLOW-DEFAULT-001
// design D-9, as amended by D-27; card t1453 M4 step 1).
//
// sweepScan is a pure function of (file path, content): it never touches the real tree,
// so the fixture self-test and the tree run decide through the same code. It classifies
// every line that names the `develop` branch as live text:
//
//	strong (P-A / P-D / P-B)  — a violation
//	weak                      — a violation too (D-27) unless a whole-line allow entry covers it
//	exempt                    — not a violation, with the reason on the finding
//	ratchet (P-C)             — a token-less live sentence, counted, never a violation
//
// RE2 has no lookbehind and its \b is ASCII-only, so the core pattern of design D-9 is
// applied as a candidate regexp plus an explicit look at the neighbouring runes: that is
// what keeps `develop에서` (CJK neighbour) a hit, `manager-develop` and
// `.claude/worktrees/develop` (identifier / path neighbour) a non-hit, and a bare
// `origin/develop` a hit (the `/` exclusion applies to the left neighbour of the whole
// match, and `origin/` is inside the match).

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// sweepClass is the verdict class of one finding.
type sweepClass string

const (
	sweepStrong  sweepClass = "strong"
	sweepWeak    sweepClass = "weak"
	sweepExempt  sweepClass = "exempt"
	sweepRatchet sweepClass = "ratchet"
)

// Exempt-rule names that are marker exemptions. They are the channel D18 bounds with a
// ratchet: a marker is the one way to silence a live line without an allow entry.
const (
	sweepRuleMarkerLine    = "marker-line"
	sweepRuleMarkerQuote   = "marker-quote"
	sweepRuleMarkerSection = "marker-section"
	sweepRuleMarkerH1      = "marker-section-h1"
)

// sweepFinding is one classified line.
type sweepFinding struct {
	File     string
	Line     int // 1-based
	Text     string
	Class    sweepClass
	Rule     string // P-A P-D P-B weak | stamp verb marker-* allow | P-C
	Why      string // the allow entry's reviewed reason, when Rule is "allow"
	AllowIdx int    // index of the allow entry used; -1 otherwise
}

func (f sweepFinding) violation() bool { return f.Class == sweepStrong || f.Class == sweepWeak }

func (f sweepFinding) String() string {
	return fmt.Sprintf("%s:%d: %s/%s: %s", f.File, f.Line, f.Class, f.Rule, strings.TrimSpace(f.Text))
}

// sweepAllow is one allow-list entry. Literal must be the WHOLE line (trimmed): an entry
// carrying a fragment would silence every future line containing it.
type sweepAllow struct {
	File    string
	Literal string
	Why     string
}

// sweepSubtree is one scoped surface subtree and its visited-count floor.
type sweepSubtree struct {
	Name         string
	Files        []string // explicit files, relative to the repo root (instead of Dir)
	Dir          string   // directory, relative to the repo root
	Exts         []string // suffixes that select a file under Dir
	NonRecursive bool
	Floor        int
}

// sweepCounts are the numbers the D18 / P-C ratchets watch.
type sweepCounts struct {
	MarkerLines   int // lines a line / quote / section marker exempted
	MarkerH1Lines int // lines a document-wide H1 marker exempted
	PCLines       int // P-C token-less live sentences
}

// sweepCeilings are the ratchet ceilings: they may fall, they never rise without a
// decision row in design.md.
type sweepCeilings struct {
	MarkerLines   int
	MarkerH1Lines int
	PCLines       int
}

var (
	// The core candidate (design D-9). The `origin/` and `refs/...` prefixes sit INSIDE the
	// match so the boundary check below sees the left neighbour of the whole thing.
	sweepCandidateRE    = regexp.MustCompile(`(?:refs/(?:remotes|heads)/)?(?:origin/)?develop`)
	sweepHyphenSuffixRE = regexp.MustCompile(`^-(?:based|branch|tip|line)\b`)

	sweepStampRE      = regexp.MustCompile(`^\s*@\s*[0-9a-f]{7,40}\b`)
	sweepVerbAfterRE  = regexp.MustCompile(`^\s+(?:a|an|the|your|our|their|this|that|new|more|it|them|and|software|features?|code)\b`)
	sweepVerbBeforeRE = regexp.MustCompile(`(?i)\bto\s+$`)

	// P-B: a branch-word token anywhere on the line. The ASCII tokens carry a LEFT word
	// boundary so `database` is not `base`; the suffix stays open so `worktrees`, `merged`
	// and `pushed` count.
	sweepBranchWordRE = regexp.MustCompile(`(?i)\b(?:branch|base|worktree|integration|merge|push|fork|check ?out|rebase|pull|rev-list)|브랜치|워크트리|기준|통합|병합|분기|체크아웃|리베이스`)

	// A marker needs a retirement word AND a date or a card / SPEC id on the same line.
	// `historical` and `legacy` are deliberately not marker words.
	sweepMarkerWordRE = regexp.MustCompile(`RETIRED|SUPERSEDED|폐기|은퇴`)
	sweepMarkerIDRE   = regexp.MustCompile(`\b20\d\d-\d\d-\d\d\b|\bt\d{3,5}\b|SPEC-[A-Z][A-Z0-9-]*`)

	// P-C: token-less live sentences, counted for the ratchet.
	sweepPCRE = regexp.MustCompile(`(?i)integration (?:worktree|window|branch)|통합 (?:워크트리|브랜치)|일괄 push|batch.?push|commit-dead|lead_push_threshold`)

	sweepHeadingRE = regexp.MustCompile(`^(#{1,6})\s+\S`)
)

func sweepIsCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r)
}

func sweepASCIIWord(r rune) bool {
	return r == '_' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// sweepLeftBlocks: a left neighbour that makes the match part of an identifier or a path
// (`manager-develop`, `push_develop`, `x.develop`, `.claude/worktrees/develop`). Anything
// else — a space, a quote, a CJK particle — is a boundary.
func sweepLeftBlocks(r rune) bool { return sweepASCIIWord(r) || r == '.' || r == '/' || r == '-' }

// sweepRightBlocks: the right neighbour that makes the match an identifier (`developer`,
// `develop-worktree`). `-based`/`-branch`/`-tip`/`-line` are handled separately: they
// are prose.
func sweepRightBlocks(r rune) bool { return sweepASCIIWord(r) || r == '-' }

func sweepHasMarker(s string) bool {
	return sweepMarkerWordRE.MatchString(s) && sweepMarkerIDRE.MatchString(s)
}

// sweepHit is one candidate that passed the boundary check.
type sweepHit struct {
	start, end int
	prefixed   bool
	cjk        bool
	code       bool // inside an inline code span
}

func sweepHits(line string) []sweepHit {
	var hits []sweepHit
	for _, loc := range sweepCandidateRE.FindAllStringIndex(line, -1) {
		start, end := loc[0], loc[1]
		if start > 0 {
			r, _ := utf8.DecodeLastRuneInString(line[:start])
			if sweepLeftBlocks(r) {
				continue
			}
		}
		if end < len(line) {
			r, _ := utf8.DecodeRuneInString(line[end:])
			if sweepRightBlocks(r) && !sweepHyphenSuffixRE.MatchString(line[end:]) {
				continue
			}
		}
		h := sweepHit{start: start, end: end}
		cand := line[start:end]
		h.prefixed = strings.HasPrefix(cand, "origin/") || strings.HasPrefix(cand, "refs/")
		if start > 0 {
			r, _ := utf8.DecodeLastRuneInString(line[:start])
			h.cjk = sweepIsCJK(r)
		}
		if end < len(line) {
			r, _ := utf8.DecodeRuneInString(line[end:])
			h.cjk = h.cjk || sweepIsCJK(r)
		}
		h.code = strings.Count(line[:start], "`")%2 == 1
		hits = append(hits, h)
	}
	return hits
}

// sweepClassify decides one hit, top to bottom, first match wins (design D-9 with the D15
// amendment): the verb exclusion applies only to a hit without an `origin/`/`refs/`
// prefix AND on a line with no branch-word token, so `Lanes merge to develop and push.`
// stays a violation. The returned rule is "" for a non-violation that needs no reason.
func sweepClassify(line string, h sweepHit, inFence, branchWord bool) (sweepClass, string) {
	if sweepStampRE.MatchString(line[h.end:]) {
		return sweepExempt, "stamp"
	}
	if !h.prefixed && !branchWord &&
		sweepVerbBeforeRE.MatchString(line[:h.start]) && sweepVerbAfterRE.MatchString(line[h.end:]) {
		return sweepExempt, "verb"
	}
	switch {
	case h.prefixed || h.code || inFence:
		return sweepStrong, "P-A"
	case h.cjk:
		return sweepStrong, "P-D"
	case branchWord:
		return sweepStrong, "P-B"
	}
	return sweepWeak, "weak"
}

// sweepScan classifies every line of one file. It returns at most one core finding per
// line (the strongest unresolved class, or the exemption that resolved it) plus a
// ratchet finding for a P-C line.
func sweepScan(file, content string, allow []sweepAllow) []sweepFinding {
	lines := strings.Split(content, "\n")
	n := len(lines)

	// Pass 1: fences. A delimiter line carries no content; a fenced line is P-A.
	inFence := make([]bool, n)
	delim := make([]bool, n)
	fence := ""
	for i, l := range lines {
		t := strings.TrimLeft(l, " \t")
		if fence == "" {
			for _, m := range []string{"```", "~~~"} {
				if strings.HasPrefix(t, m) {
					fence = m
					delim[i] = true
					break
				}
			}
			continue
		}
		if strings.HasPrefix(t, fence) {
			fence = ""
			delim[i] = true
			continue
		}
		inFence[i] = true
	}

	// Pass 2: marker scopes. A heading marker covers its own section — up to the next
	// heading of the same or a higher level — and a quote marker covers its quote block
	// only. Neither reaches past that, which is what F-red-10 pins.
	sectionRule := make([]string, n)
	activeLevel := 0
	for i, l := range lines {
		if delim[i] || inFence[i] {
			if activeLevel > 0 {
				sectionRule[i] = sweepRuleMarkerSection
				if activeLevel == 1 {
					sectionRule[i] = sweepRuleMarkerH1
				}
			}
			continue
		}
		if m := sweepHeadingRE.FindStringSubmatch(l); m != nil {
			level := len(m[1])
			if activeLevel > 0 && level <= activeLevel {
				activeLevel = 0
			}
			if activeLevel == 0 && sweepHasMarker(l) {
				activeLevel = level
			}
		}
		if activeLevel > 0 {
			sectionRule[i] = sweepRuleMarkerSection
			if activeLevel == 1 {
				sectionRule[i] = sweepRuleMarkerH1
			}
		}
	}
	quoteMarked := make([]bool, n)
	for i := 0; i < n; {
		if delim[i] || inFence[i] || !strings.HasPrefix(strings.TrimLeft(lines[i], " \t"), ">") {
			i++
			continue
		}
		j, marked := i, false
		for j < n && !delim[j] && !inFence[j] && strings.HasPrefix(strings.TrimLeft(lines[j], " \t"), ">") {
			marked = marked || sweepHasMarker(lines[j])
			j++
		}
		for k := i; k < j; k++ {
			quoteMarked[k] = marked
		}
		i = j
	}
	markerRule := func(i int) string {
		switch {
		case sweepHasMarker(lines[i]):
			return sweepRuleMarkerLine
		case quoteMarked[i]:
			return sweepRuleMarkerQuote
		}
		return sectionRule[i]
	}

	// Pass 3: classify.
	var out []sweepFinding
	for i, l := range lines {
		if delim[i] {
			continue
		}
		branchWord := sweepBranchWordRE.MatchString(l)
		var viol *sweepFinding
		var exempt *sweepFinding
		for _, h := range sweepHits(l) {
			class, rule := sweepClassify(l, h, inFence[i], branchWord)
			f := sweepFinding{File: file, Line: i + 1, Text: l, Class: class, Rule: rule, AllowIdx: -1}
			if class == sweepExempt {
				if exempt == nil {
					exempt = &f
				}
				continue
			}
			// A would-be violation: a marker resolves it first, then a whole-line allow entry.
			if mr := markerRule(i); mr != "" {
				f.Class, f.Rule = sweepExempt, mr
				exempt = &f
				continue
			}
			if idx := sweepAllowIndex(file, l, allow); idx >= 0 {
				f.Class, f.Rule, f.Why, f.AllowIdx = sweepExempt, "allow", allow[idx].Why, idx
				exempt = &f
				continue
			}
			if viol == nil || (viol.Class == sweepWeak && class == sweepStrong) {
				viol = &f
			}
		}
		switch {
		case viol != nil:
			out = append(out, *viol)
		case exempt != nil:
			out = append(out, *exempt)
		}
		if sweepPCRE.MatchString(l) && markerRule(i) == "" {
			out = append(out, sweepFinding{File: file, Line: i + 1, Text: l, Class: sweepRatchet, Rule: "P-C", AllowIdx: -1})
		}
	}
	return out
}

// sweepAllowIndex returns the index of the entry whose file and WHOLE line match, or -1.
func sweepAllowIndex(file, line string, allow []sweepAllow) int {
	t := strings.TrimSpace(line)
	for i, e := range allow {
		if e.File == file && strings.TrimSpace(e.Literal) == t {
			return i
		}
	}
	return -1
}

// sweepValidateAllow rejects the shapes that would hollow the list out and enforces the
// cap. The cap is the price of an entry: adding a line to the list is the cheapest way
// to make a violation disappear, so growth past the cap needs a decision row.
func sweepValidateAllow(entries []sweepAllow, limit int) error {
	if len(entries) > limit {
		return fmt.Errorf("allow list has %d entries, cap is %d — raise the cap only with a decision row", len(entries), limit)
	}
	for i, e := range entries {
		switch {
		case strings.TrimSpace(e.File) == "":
			return fmt.Errorf("allow[%d]: empty file path", i)
		case strings.TrimSpace(e.Literal) == "":
			return fmt.Errorf("allow[%d] %s: no literal — a file-only entry would exempt the whole file", i, e.File)
		case !strings.ContainsAny(strings.TrimSpace(e.Literal), " \t"):
			return fmt.Errorf("allow[%d] %s: literal %q is a bare token, not a whole line", i, e.File, e.Literal)
		case strings.TrimSpace(e.Why) == "":
			return fmt.Errorf("allow[%d] %s: no reviewed reason", i, e.File)
		}
	}
	return nil
}

func sweepCountsOf(findings []sweepFinding) sweepCounts {
	var c sweepCounts
	for _, f := range findings {
		switch {
		case f.Class == sweepRatchet:
			c.PCLines++
		case f.Class == sweepExempt && f.Rule == sweepRuleMarkerH1:
			c.MarkerH1Lines++
		case f.Class == sweepExempt && (f.Rule == sweepRuleMarkerLine || f.Rule == sweepRuleMarkerQuote || f.Rule == sweepRuleMarkerSection):
			c.MarkerLines++
		}
	}
	return c
}

func sweepRatchetProblems(c sweepCounts, ceil sweepCeilings) []string {
	var p []string
	if c.MarkerLines > ceil.MarkerLines {
		p = append(p, fmt.Sprintf("marker-exempted lines = %d, ceiling %d — a marker silenced more live text than the recorded baseline", c.MarkerLines, ceil.MarkerLines))
	}
	if c.MarkerH1Lines > ceil.MarkerH1Lines {
		p = append(p, fmt.Sprintf("document-wide H1-marker-exempted lines = %d, ceiling %d", c.MarkerH1Lines, ceil.MarkerH1Lines))
	}
	if c.PCLines > ceil.PCLines {
		p = append(p, fmt.Sprintf("P-C token-less live sentences = %d, ceiling %d", c.PCLines, ceil.PCLines))
	}
	return p
}

// sweepVisitProblems judges a sweep by what it visited. A sweep that visited nothing is a
// failing guard, never a pass (verification-completeness §1.1), and the total floor and the
// per-subtree floors are separate checks: dropping one small subtree fits inside the total
// floor's slack.
func sweepVisitProblems(visited map[string]int, subtrees []sweepSubtree, totalFloor int) []string {
	var p []string
	total := 0
	for _, v := range visited {
		total += v
	}
	if total == 0 {
		p = append(p, "empty sweep: visited=0 files — a green over nothing asserts nothing")
	}
	for _, s := range subtrees {
		if got := visited[s.Name]; got < s.Floor {
			p = append(p, fmt.Sprintf("subtree %s: visited %d, floor %d", s.Name, got, s.Floor))
		}
	}
	if total < totalFloor {
		p = append(p, fmt.Sprintf("total visited %d, floor %d", total, totalFloor))
	}
	return p
}
