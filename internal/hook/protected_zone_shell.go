package hook

// protected_zone_shell.go — the shell half of the protected-zone guard
// (SPEC-SELF-IMPROVE-PROTECTED-ZONE-001, plan.md M3; hardened in merge-gate
// repair rounds 1–3, re-platformed on the mvdan/sh parser by the operator's
// option-A decision after rounds 4–5, and reworked in round 6 onto a
// possible-directory-set walk).
//
// The lexical analysis is mvdan/sh's parser (already a direct dependency,
// v3.14.0): quoting, expansions, async boundaries, pipelines, and redirections
// come from the AST. What remains here is the POLICY layer: which command
// words are mutating, which arguments and redirection targets are zone-
// covered, and how the tracked working directory moves.
//
// The tracked directory is a SET of possible directories (round 8, the
// operator-approved sound set semantics): a candidate is denied when ANY
// possible directory covers it — a sound over-approximation of bash's control
// flow:
//   - a statement's redirections open BEFORE the statement runs, in the
//     walker's current directory — every statement shape, compound included;
//   - a statement with Background (trailing "&"), a pipeline element, and an
//     explicit Subshell run in a subshell: their cd never moves the main
//     shell, while their mutations and redirections are real;
//   - a cd may fail, so its pre-cd set survives; either side of a && or ||
//     may be skipped, so the post-left set survives past the operator;
//   - an if's condition executes and its effects survive into every branch;
//     the then and else worlds union (an elif is an IfClause as the Else
//     member, whose condition executes too); a loop keeps its zero-iteration
//     world and unrolls its body once; every case arm is judged from the
//     same entry set, independently;
//   - a cd's directory arguments are its non-redirection words (the AST
//     separates them by construction);
//   - git's `-C <dir>` moves the directory its file arguments resolve against,
//     accumulated per possible start directory, consecutive and quoted
//     options included; only the first non-option word is the subcommand
//     (round 7 P2).
//
// Dynamic words (expansions, globs) under-match: they normalize to paths no
// entry matches. A parse failure under-matches like unclassifiable text.
// Unlike a baseline Write/Edit denial, every shell-mutation denial carries
// the protected-zone sentinel and the routing fields — there is no legacy
// shell reason to keep (spec §C.7).

import (
	"encoding/json"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/modu-ai/moai-adk/internal/config"
)

// zoneMutationVerbs are the single-word mutating forms of REQ-SIPZ-007.
var zoneMutationVerbs = map[string]bool{
	"rm": true, "unlink": true, "mv": true, "cp": true, "tee": true, "truncate": true,
}

// zoneGitMutating are the git subcommands of REQ-SIPZ-007.
var zoneGitMutating = map[string]bool{
	"rm": true, "checkout": true, "restore": true, "apply": true,
}

// zoneShellParser parses one command per guard invocation. Parsing failures
// under-match: the command is allowed, exactly like the splitter-based
// analyzer treated text it could not classify.
var zoneShellParser = syntax.NewParser()

// zoneParse parses one command string with bash semantics.
func zoneParse(command string) (*syntax.File, bool) {
	file, err := zoneShellParser.Parse(strings.NewReader(command), "")
	if err != nil {
		return nil, false
	}
	return file, true
}

// zoneUnescapeLit resolves the backslash escapes a literal keeps in its
// source text (`\ ` -> ` `, `\\` -> `\`, `\"` -> `"`). It is the decode for
// BARE (unquoted) words, where bash removes the backslash before every
// character. Quoted words decode differently — see zoneUnescapeDbl and
// zoneUnescapeAnsiC (card t1570).
func zoneUnescapeLit(v string) string {
	if !strings.Contains(v, "\\") {
		return v
	}
	var b strings.Builder
	b.Grow(len(v))
	for i := 0; i < len(v); i++ {
		if v[i] == '\\' && i+1 < len(v) {
			i++
		}
		b.WriteByte(v[i])
	}
	return b.String()
}

// zoneUnescapeDbl decodes one literal part of a double-quoted word: bash
// keeps the backslash an escape only before $ ` " \ and newline, and before
// any other character the backslash is literal ("lnk\dir" names a directory
// whose name carries the backslash — collapsing it to lnkdir missed the
// zone, gate round 8 / card t1570). A backslash-newline is a line
// continuation and drops both characters.
func zoneUnescapeDbl(v string) string {
	if !strings.Contains(v, "\\") {
		return v
	}
	var b strings.Builder
	b.Grow(len(v))
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == '\\' && i+1 < len(v) {
			switch v[i+1] {
			case '$', '`', '"', '\\':
				b.WriteByte(v[i+1])
				i++
				continue
			case '\n':
				i++
				continue
			}
			b.WriteByte('\\')
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// zoneUnescapeAnsiC decodes one ANSI-C ($'...') part for a MODERN bash: the
// single-character escapes, octal \nnn, hex \xH.., and \u/\U code points as
// UTF-8. The part's contribution ends at the first NUL of ANY origin —
// modern bash truncates the argument at the first NUL however spelled. An
// escape with no defined meaning keeps the backslash and the character —
// bash renders `$'a\qb'` as `a\qb` — so the guard checks the same text bash
// writes. Words carrying \u/\U escapes are judged in BOTH this world and
// the pre-4.2 reading (zoneUnescapeAnsiCPre42): the rendering is
// version-variant, so the guard's candidate set carries one text per bash
// generation and denies when EITHER lands in the zone — the possible-worlds
// union, the same soundness as the possible-directory set (gate rounds
// 13-14, card t1585). Before this decoder ANSI-C words were matched on
// their raw source text, so any defined escape (`\\`, `\x2e`, ...) hid the
// real path (card t1570).
func zoneUnescapeAnsiC(v string) string {
	return zoneUnescapeAnsiCWorld(v, false)
}

// zoneUnescapeAnsiCPre42 decodes one ANSI-C part the way a PRE-4.2 bash
// renders it: \xHH, octal, and the single-character escapes decode exactly
// like the modern decoder — raw bytes, the part ending at the first
// \x/octal-origin NUL — while \u/\U are UNKNOWN escapes: the backslash and
// the letter stay and the digits that follow are ordinary characters
// (measured on bash 3.2.57: the ⊇ escape text renders as
// 5c 75 32 32 38 37). The old-bash candidate of the words pair (gate round
// 14 P1: the true pre-4.2 path for `link\u0000\x00/...` truncates at the
// \x00 NUL — `link\u0000` — which a literally-named symlink resolves into
// the zone).
func zoneUnescapeAnsiCPre42(v string) string {
	return zoneUnescapeAnsiCWorld(v, true)
}

// zoneUnescapeAnsiCWorld is the shared decode loop behind both worlds; the
// flag selects only the \u/\U arms and, through them, which NULs can end
// the part (any origin in the modern world; \x/octal origins only in the
// pre-4.2 world, where a \u/\U is text and renders no NUL at all).
func zoneUnescapeAnsiCWorld(v string, pre42 bool) string {
	if !strings.Contains(v, "\\") {
		return v
	}
	var b strings.Builder
	b.Grow(len(v))
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c != '\\' || i+1 >= len(v) {
			b.WriteByte(c)
			continue
		}
		i++
		switch e := v[i]; e {
		case 'a':
			b.WriteByte(7)
		case 'b':
			b.WriteByte(8)
		case 'e', 'E':
			b.WriteByte(27)
		case 'f':
			b.WriteByte(12)
		case 'n':
			b.WriteByte(10)
		case 'r':
			b.WriteByte(13)
		case 't':
			b.WriteByte(9)
		case 'v':
			b.WriteByte(11)
		case '\\', '\'', '"', '?':
			b.WriteByte(e)
		case 'x':
			r := zoneHexEscape(v, &i, 2, true)
			if len(r) == 1 && r[0] == 0 {
				// bash terminates the argument at the NUL: the rest of
				// this part is dropped, later word parts still append.
				return b.String()
			}
			b.WriteString(r)
		case 'u':
			if pre42 {
				// unknown to pre-4.2 bash: the backslash and the letter
				// stay, the digits that follow are ordinary characters.
				b.WriteByte('\\')
				b.WriteByte(e)
				break
			}
			r := zoneHexEscape(v, &i, 4, false)
			if len(r) == 1 && r[0] == 0 {
				return b.String()
			}
			b.WriteString(r)
		case 'U':
			if pre42 {
				b.WriteByte('\\')
				b.WriteByte(e)
				break
			}
			r := zoneHexEscape(v, &i, 8, false)
			if len(r) == 1 && r[0] == 0 {
				return b.String()
			}
			b.WriteString(r)
		default:
			if e >= '0' && e <= '7' {
				o := zoneOctalEscape(v, &i, e)
				if o == 0 {
					// the octal NUL terminates exactly like \x00.
					return b.String()
				}
				b.WriteByte(o)
				continue
			}
			b.WriteByte('\\')
			b.WriteByte(e)
		}
	}
	return b.String()
}

// zoneHexEscape reads up to maxDigits hex digits after the \x/\u/\U prefix
// letter at v[*i] and advances *i over the digits consumed. The render
// splits by prefix (card t1585): \x emits ONE RAW BYTE — bash \xHH places
// that byte in the argument (measured $'\xec\xa1\xb4' -> ec a1 b4), while
// \u/\U render the code point as UTF-8 (string(rune(val))). With no digit
// the escape is not defined: the backslash and the prefix letter stay
// literal, the way bash renders them — returned as the bounded two-byte
// slice at the prefix, so the escape can end the string without a panic
// and a non-digit follower survives exactly once (the caller's loop
// advances past the letter only).
func zoneHexEscape(v string, i *int, maxDigits int, rawByte bool) string {
	j := *i + 1
	val := rune(0)
	digits := 0
	for digits < maxDigits && j < len(v) {
		c := v[j]
		var d rune
		switch {
		case c >= '0' && c <= '9':
			d = rune(c - '0')
		case c >= 'a' && c <= 'f':
			d = rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			d = rune(c-'A') + 10
		default:
			d = -1
		}
		if d < 0 {
			break
		}
		val = val*16 + d
		digits++
		j++
	}
	if digits == 0 {
		return v[*i-1 : *i+1]
	}
	*i = j - 1
	if rawByte {
		return string([]byte{byte(val)})
	}
	return string(val)
}

// zoneOctalEscape reads one to two further octal digits after v[*i] (v[*i]
// is the first digit) and renders the byte.
func zoneOctalEscape(v string, i *int, first byte) byte {
	val := rune(first - '0')
	j := *i + 1
	for k := 0; k < 2 && j < len(v); k++ {
		c := v[j]
		if c < '0' || c > '7' {
			break
		}
		val = val*8 + rune(c-'0')
		j++
	}
	*i = j - 1
	return byte(val)
}

// zoneWordText returns the literal text of a word and whether the word is
// fully literal. Words carrying expansions or globs are dynamic: their text
// normalizes to a path no entry matches, so callers drop them (under-match).
func zoneWordText(w *syntax.Word) (string, bool) {
	if w == nil {
		return "", false
	}
	var b strings.Builder
	for _, part := range w.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			b.WriteString(zoneUnescapeLit(p.Value))
		case *syntax.SglQuoted:
			// Plain single quotes carry no escapes. $'...' is ANSI-C
			// quoting (mvdan marks it with Dollar and keeps the raw source
			// text), whose escape set bash decodes — the guard must check
			// the decoded path, not the source text (card t1570). The NUL
			// termination lives inside the decoder, scoped to the
			// \x00/octal origins (gate round 10 P1, card t1585).
			if p.Dollar {
				b.WriteString(zoneUnescapeAnsiC(p.Value))
			} else {
				b.WriteString(p.Value)
			}
		case *syntax.DblQuoted:
			for _, dp := range p.Parts {
				lit, ok := dp.(*syntax.Lit)
				if !ok {
					return "", false
				}
				b.WriteString(zoneUnescapeDbl(lit.Value))
			}
		default:
			return "", false
		}
	}
	return b.String(), true
}

// zoneFirstArgWord returns the literal text of the first word and whether it
// is fully literal.
func zoneFirstArgWord(args []*syntax.Word) (string, bool) {
	if len(args) == 0 {
		return "", false
	}
	return zoneWordText(args[0])
}

// zoneWordTextPre42 returns the word's text as a PRE-4.2 shell reads it:
// identical to zoneWordText except that ANSI-C parts decode through the
// pre-4.2 rendering — \x/octal and the simple escapes decode exactly the
// same (raw bytes, NUL termination included), while \u/\U stay verbatim
// literal text. A word WITHOUT \u/\U decodes identically in both worlds and
// never needs this second candidate (zoneWordDual gates it).
func zoneWordTextPre42(w *syntax.Word) (string, bool) {
	if w == nil {
		return "", false
	}
	var b strings.Builder
	for _, part := range w.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			b.WriteString(zoneUnescapeLit(p.Value))
		case *syntax.SglQuoted:
			if p.Dollar {
				b.WriteString(zoneUnescapeAnsiCPre42(p.Value))
			} else {
				b.WriteString(p.Value)
			}
		case *syntax.DblQuoted:
			for _, dp := range p.Parts {
				lit, ok := dp.(*syntax.Lit)
				if !ok {
					return "", false
				}
				b.WriteString(zoneUnescapeDbl(lit.Value))
			}
		default:
			return "", false
		}
	}
	return b.String(), true
}

// zoneWordDual reports whether the word needs BOTH candidates judged: an
// ANSI-C part carrying \u or \U is the one escape family whose rendering
// differs across bash generations.
func zoneWordDual(w *syntax.Word) bool {
	if w == nil {
		return false
	}
	for _, part := range w.Parts {
		if p, ok := part.(*syntax.SglQuoted); ok && p.Dollar {
			if strings.Contains(p.Value, `\u`) || strings.Contains(p.Value, `\U`) {
				return true
			}
		}
	}
	return false
}

// zoneWordCandidates returns the candidate texts one word contributes to
// judgment: the modern-decoded text, plus the pre-4.2 reading when the word
// carries \u/\U escapes (gate rounds 13-14) — deny when EITHER lands in the
// zone. The bool is false when the word is dynamic.
func zoneWordCandidates(w *syntax.Word) ([]string, bool) {
	t, literal := zoneWordText(w)
	if !literal {
		return nil, false
	}
	if !zoneWordDual(w) {
		return []string{t}, true
	}
	old, lit := zoneWordTextPre42(w)
	if !lit || old == t {
		return []string{t}, true
	}
	return []string{t, old}, true
}

// zoneWalker carries one shell-policy walk. cwds is the set of POSSIBLE
// working directories at this point — control flow multiplies them, and a
// candidate is denied when any of them covers it (round 6 P1). unbounded
// records that a loop's directory states outgrew the fixed-point bound: the
// walk is then an over-approximation that cannot be completed, and the
// command is denied fail-closed rather than allowed on an incomplete walk
// (round 9 P1).
type zoneWalker struct {
	h        *preToolHandler
	cwds     []string
	mutating bool
	// funcs maps a function name declared in THIS command to the bodies it
	// may have: a straight-line redefinition replaces, a branch join unions
	// — the skipped branch's definition must not win (round 14 P1). calling
	// holds the names currently being walked; calls counts the bounded
	// re-entries a recursion may unroll (round 17 P1). unbounded records
	// that a walk outgrew one of its bounds — the command is then denied
	// fail-closed rather than allowed on an incomplete analysis.
	unbounded bool
	cands     []string
	funcs     map[string][]*syntax.Stmt
	calling   map[string]bool
	calls     map[string]int
}

// setCwds replaces the possible-directory set, dropping duplicates.
func (w *zoneWalker) setCwds(dirs []string) {
	seen := map[string]bool{}
	w.cwds = w.cwds[:0]
	for _, d := range dirs {
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		w.cwds = append(w.cwds, d)
	}
}

// zoneRelativeTo prefixes one directory onto relative candidates, keeping the
// raw segments for the normalization and symlink layers; absolute candidates
// are judged as they land. An empty dir leaves the candidates unchanged.
func zoneRelativeTo(dir string, cands []string) []string {
	out := make([]string, 0, len(cands))
	for _, cand := range cands {
		if dir == "" || dir == "." || zoneIsAbs(cand) {
			out = append(out, cand)
		} else {
			out = append(out, dir+"/"+cand)
		}
	}
	return out
}

// zoneRelativeToSet expands one candidate against every possible working
// directory; absolute candidates and an unset root pass through as they land.
func zoneRelativeToSet(cwds []string, cand string) []string {
	if len(cwds) == 0 {
		return []string{cand}
	}
	out := make([]string, 0, len(cwds))
	for _, dir := range cwds {
		if dir == "" || dir == "." || zoneIsAbs(cand) {
			out = append(out, cand)
		} else {
			out = append(out, dir+"/"+cand)
		}
	}
	return out
}

// zoneCands expands every candidate against the walker's possible directories.
func (w *zoneWalker) zoneCands(cands []string) {
	for _, cand := range cands {
		w.cands = append(w.cands, zoneRelativeToSet(w.cwds, cand)...)
	}
}

// zoneRedirectTargets returns the targets of the write redirections in the
// list. Input redirects (`<`), here-docs, and `<&` read the target instead of
// writing it and are skipped; `>`, `>>`, `<>`, `>|`, `&>` and `&>>` create or
// truncate it.
func zoneRedirectTargets(redirs []*syntax.Redirect) (bool, []string) {
	mutating := false
	var targets []string
	for _, rd := range redirs {
		if rd.Op == syntax.DplOut {
			// `>&` onto a NUMBERED descriptor (`2>&1`) duplicates a file
			// descriptor and writes nothing; onto a word (`>& file`) it is a
			// file write in the dialects that accept the spelling — only the
			// numeric form skips (round 11 P1, correcting round 10 P3's
			// blanket skip)
			if t, literal := zoneWordText(rd.Word); literal && isZoneDigits(t) {
				continue
			}
		} else {
			switch rd.Op {
			case syntax.RdrIn, syntax.Hdoc, syntax.DashHdoc, syntax.WordHdoc, syntax.DplIn:
				continue
			}
		}
		worlds, literal := zoneWordCandidates(rd.Word)
		if !literal {
			continue // dynamic target: under-match
		}
		// an empty world filters PER WORLD: the modern reading may truncate
		// to "" at a leading code-point NUL while the pre-4.2 reading still
		// names the real write target (gate round 15 P1)
		mutating = true
		for _, wText := range worlds {
			if wText != "" {
				targets = append(targets, wText)
			}
		}
	}
	return mutating, targets
}

// isZoneDigits reports whether s is a non-empty run of ASCII digits — a file
// descriptor number, not a path.
func isZoneDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// zoneRedirects judges a statement's write redirections against every
// possible working directory.
func (w *zoneWalker) zoneRedirects(redirs []*syntax.Redirect) {
	if hits, targets := zoneRedirectTargets(redirs); hits {
		w.mutating = true
		w.zoneCands(targets)
	}
}

// zoneNextCwd returns the working directory after a cd from cur. Only a plain
// literal relative argument is tracked — a bare cd, several arguments (bash
// rejects them and the cd fails), substitution, a glob, an option, or an
// absolute path leaves it untracked (reset to the root), the documented
// under-match (merge-gate round 1 P1-3). The dots are cleaned here because
// the shell's cd already resolved the directory: this is the logical cwd the
// next segment's relative names concatenate onto, never a pre-clean of a
// target path.
func zoneNextCwd(cur string, dirs []string) string {
	if len(dirs) != 1 {
		return "."
	}
	arg := dirs[0]
	if arg == "-" || strings.HasPrefix(arg, "-") || strings.ContainsAny(arg, "$*?") || zoneIsAbs(arg) {
		return "."
	}
	next := arg
	if cur != "." {
		next = cur + "/" + arg
	}
	next = path.Clean(next)
	if next == ".." || strings.HasPrefix(next, "../") {
		return "."
	}
	return next
}

// zoneBaselineCovers reports whether a project-relative path is covered by the
// compiled baseline floor. A shell denial carries the protected-zone sentinel
// even here, so this is a predicate and not the legacy formatter.
func zoneBaselineCovers(rel string) bool {
	base := path.Base(rel)
	for _, inst := range frozenInstructionFiles {
		if base == inst {
			return true
		}
	}
	for _, fz := range frozenZonePrefixes {
		if strings.HasPrefix(rel, fz.prefix) {
			return true
		}
	}
	return false
}

// zoneDirContainsGlobMatch walks dir and reports whether any file under it
// matches the basename-glob entry — removing such a directory removes every
// file the glob protects there (round 3 P1: an overlay "**/*_test.go" must
// make "rm -rf tests" a denial). The walk is bounded; beyond the bound it
// under-matches, the shell rule's accepted direction.
func zoneDirContainsGlobMatch(dir string, entry config.ZoneEntry) bool {
	seen := 0
	found := false
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return fs.SkipAll
		}
		seen++
		if seen > 5000 {
			return fs.SkipAll
		}
		if d.IsDir() {
			return nil
		}
		if entry.Match(config.FoldZoneText(filepath.ToSlash(p))) {
			found = true
		}
		return nil
	})
	return found
}

// zoneShellCovered judges the normalized forms of one candidate against the
// zone, in two passes. Direct: the candidate — tried with and without a
// trailing slash, so a bare directory operand names the directory itself
// (round 1 P1-1) — is a zone member. Ancestor: the candidate is a parent
// directory of protected entries, so moving or removing it would remove them
// (round 2 P1: "mv .moai moved"; round 3 P1: the same holds for a basename-
// glob entry, judged by walking the candidate); the project root itself is
// the parent of everything. The returned category names the source that
// covered the candidate.
func zoneShellCovered(forms []zoneForm, load config.ProtectedZoneLoad, root string) (string, bool) {
	for _, form := range forms {
		for _, spelling := range []string{form.Display, form.Display + "/", form.Folded, form.Folded + "/"} {
			if zoneBaselineCovers(spelling) {
				return zoneBaselineCategory, true
			}
		}
	}
	if load.State == config.ZoneStateOK {
		for _, form := range forms {
			for _, folded := range []string{form.Folded, form.Folded + "/"} {
				for i := range load.Zone.Entries {
					if load.Zone.Entries[i].Match(folded) {
						return load.Zone.Entries[i].Category, true
					}
				}
			}
		}
	}
	for _, form := range forms {
		folded := form.Folded
		if folded == "" {
			continue
		}
		prefix := folded
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if folded == "." {
			prefix = "" // the project root itself is the parent of everything
		}
		for _, fz := range frozenZonePrefixes {
			if strings.HasPrefix(config.FoldZoneText(fz.prefix), prefix) {
				return zoneBaselineCategory, true
			}
		}
		for _, inst := range frozenInstructionFiles {
			if strings.HasPrefix(config.FoldZoneText(inst), prefix) {
				return zoneBaselineCategory, true
			}
		}
		if load.State == config.ZoneStateOK {
			for i := range load.Zone.Entries {
				entry := load.Zone.Entries[i]
				if entry.Kind == config.ZoneBaseGlob {
					// no path prefix to contain, but removing the directory
					// removes every file the glob matches there. The walk
					// reaches the directory under its ORIGINAL case — the
					// fold is for comparisons only, and a folded path finds
					// nothing on a case-sensitive filesystem (round 13 P1)
					if zoneDirContainsGlobMatch(filepath.Join(root, filepath.FromSlash(form.Display)), entry) {
						return entry.Category, true
					}
					continue
				}
				if strings.HasPrefix(entry.Pattern, prefix) {
					return entry.Category, true
				}
			}
		}
	}
	return "", false
}

// zoneCall judges one simple command: a mutating verb, an in-place sed, a
// mutating git subcommand. The statement's redirections are judged by the
// walker before this runs — the shell opens them before the command executes,
// for every statement shape (round 8).
func (w *zoneWalker) zoneCall(cmd *syntax.CallExpr) {
	name, literal := zoneFirstArgWord(cmd.Args)
	if !literal {
		return // a dynamic command word under-matches
	}
	if strings.Contains(name, "/") {
		// a literal executable path (/bin/rm, ./rm) names the verb through
		// its base (round 13 P1); a path to some OTHER binary folds to a
		// base that matches no verb, exactly as before
		name = path.Base(name)
	}
	if bodies, declared := w.funcs[name]; declared {
		// a call to a function declared in this command runs every body the
		// name may have (round 10 P2). Each body is a separate WORLD, not a
		// sequence: one body's inner declarations must not replace another's
		// (round 16 P1), and the call's RESULT is the union of the bodies'
		// own outcomes — a body that certainly redefines a name retires the
		// pre-call definition instead of keeping it alive alongside (round
		// 17 P2). The directory set needs no such isolation — a candidate
		// is judged against the whole accumulated set anyway.
		if w.calling[name] {
			// a recursive call re-enters bounded: each re-entry walks the
			// body from the walker's CURRENT state (the recursion's
			// directory moves are real), and past the bound the command is
			// fail-closed like any other walk that cannot complete (round
			// 17 P1)
			if w.calls[name] >= zoneRecursionBound {
				w.unbounded = true
				return
			}
			w.calls[name]++
		} else {
			w.calling[name] = true
			w.calls[name] = 1
		}
		entry := cloneZoneFuncs(w.funcs)
		var result map[string][]*syntax.Stmt
		for _, body := range bodies {
			w.funcs = cloneZoneFuncs(entry)
			w.zoneWalkStmt(body)
			if result == nil {
				result = w.funcs
			} else {
				result = mergeZoneFuncs(result, w.funcs)
			}
		}
		w.funcs = result
		delete(w.calling, name)
		return
	}
	if name == "cd" {
		// The cd's ONE argument may carry multiple readings (\u/\U words:
		// gate round 15 P1) — each non-empty reading is a POSSIBLE
		// destination and the possible-directory set unions them, the same
		// structure as the control-flow set. Several arguments still fail
		// the cd (bash rejects them) and a dynamic argument under-matches,
		// exactly as before.
		multi := len(cmd.Args) != 2
		var readings []string
		if multi {
			for _, a := range cmd.Args[1:] {
				if t, lit := zoneWordText(a); lit {
					readings = append(readings, t)
				} else {
					readings = append(readings, "?dynamic")
				}
			}
		} else if worlds, lit := zoneWordCandidates(cmd.Args[1]); lit {
			for _, dir := range worlds {
				if dir != "" {
					readings = append(readings, dir)
				}
			}
		} else {
			readings = []string{"?dynamic"}
		}
		next := make([]string, 0, len(w.cwds)*2)
		for _, cwd := range w.cwds {
			if multi {
				next = append(next, zoneNextCwd(cwd, readings))
				continue
			}
			for _, dir := range readings {
				next = append(next, zoneNextCwd(cwd, []string{dir}))
			}
		}
		// the cd may fail (a missing directory leaves the caller where it
		// was): the pre-cd set survives into the next statement either way
		// (round 8 P1)
		next = append(next, w.cwds...)
		w.setCwds(next)
		return
	}
	switch name {
	case "sed":
		// in-place is decided by THIS command's options alone — an earlier
		// mutating command must not turn a read-only sed into a denial
		// (round 6 P2)
		inPlace := false
		for _, a := range cmd.Args[1:] {
			if t, lit := zoneWordText(a); lit {
				// GNU sed's suffixed form (--in-place=.bak) is in-place too
				// (round 13 P1)
				if t == "--in-place" || strings.HasPrefix(t, "--in-place=") || (strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "--") && strings.Contains(strings.TrimPrefix(t, "-"), "i")) {
					inPlace = true
				}
			}
		}
		if !inPlace {
			return
		}
		w.mutating = true
		w.zoneCands(zonePathCandidates(cmd.Args[1:]))
		return
	case "git":
		// only the FIRST non-option word is the subcommand — a later argument
		// that merely names a mutating verb (a grep pattern, a path) must not
		// trip the guard (round 7 P2). Valued global options consume their
		// argument; a -C <dir> records the directory the subcommand's file
		// arguments resolve against, relative values accumulating across
		// consecutive -C options, quoted spellings included (rounds 4–5 P1).
		// The anchor VALUES may carry multiple readings (\u/\U words: gate
		// round 15 P1) — each reading is a possible anchor and the
		// anchoring loop unions them.
		dirOpts := []string{""}
		wtOpts := []string{""}
		sub := ""
		subIdx := -1
		for j := 1; j < len(cmd.Args); j++ {
			t, lit := zoneWordText(cmd.Args[j])
			if !lit {
				break // dynamic global argument: under-match
			}
			if strings.HasPrefix(t, "--work-tree=") {
				// the attached form's value has readings too: every
				// world's spelling carrying the prefix anchors (gate
				// round 15 P1)
				if worlds, wl := zoneWordCandidates(cmd.Args[j]); wl {
					for _, wText := range worlds {
						if len(wText) > len("--work-tree=") && strings.HasPrefix(wText, "--work-tree=") {
							wtOpts = append(wtOpts, strings.TrimPrefix(wText, "--work-tree="))
						}
					}
				}
				continue
			}
			if strings.HasPrefix(t, "-") {
				if t == "-C" || t == "-c" || t == "--git-dir" || t == "--work-tree" || t == "--namespace" || t == "--super-prefix" {
					if t == "-C" && j+1 < len(cmd.Args) {
						if worlds, lit2 := zoneWordCandidates(cmd.Args[j+1]); lit2 {
							next := make([]string, 0, len(dirOpts)*len(worlds))
							for _, d := range dirOpts {
								for _, dir := range worlds {
									if dir == "" {
										continue
									}
									if d == "" || zoneIsAbs(dir) {
										next = append(next, dir)
									} else {
										next = append(next, d+"/"+dir)
									}
								}
							}
							if len(next) > 0 {
								dirOpts = next
							}
						}
					}
					if t == "--work-tree" && j+1 < len(cmd.Args) {
						if worlds, lit2 := zoneWordCandidates(cmd.Args[j+1]); lit2 {
							for _, wt := range worlds {
								if wt != "" {
									wtOpts = append(wtOpts, wt)
								}
							}
						}
					}
					j++ // the option's value is consumed
				}
				continue
			}
			sub = t
			subIdx = j
			break
		}
		if subIdx == -1 || !zoneGitMutating[sub] {
			return
		}
		w.mutating = true
		var fileArgs []string
		for _, a := range cmd.Args[subIdx+1:] {
			worlds, flit := zoneWordCandidates(a)
			if !flit || len(worlds) == 0 || worlds[0] == "" || worlds[0] == "--" || strings.HasPrefix(worlds[0], "-") {
				continue
			}
			fileArgs = append(fileArgs, worlds...)
		}
		// every possible directory is a base the subcommand's file arguments
		// can resolve against (round 6 P1); a -C moves that base — an absolute
		// -C replaces it, a relative one accumulates (round 5 P1); a
		// --work-tree is its own anchor, so both anchorings are judged when
		// both are present (round 12 P1). Every anchor READING is judged
		// (gate round 15 P1).
		for _, base := range w.cwds {
			for _, dirOpt := range dirOpts {
				gitDir := base
				if dirOpt != "" {
					if zoneIsAbs(dirOpt) || gitDir == "" || gitDir == "." {
						gitDir = dirOpt
					} else {
						gitDir = gitDir + "/" + dirOpt
					}
				}
				w.cands = append(w.cands, zoneRelativeTo(gitDir, fileArgs)...)
				for _, wtOpt := range wtOpts {
					if wtOpt == "" {
						continue
					}
					wtDir := base
					if zoneIsAbs(wtOpt) || wtDir == "" || wtDir == "." {
						wtDir = wtOpt
					} else {
						wtDir = wtDir + "/" + wtOpt
					}
					w.cands = append(w.cands, zoneRelativeTo(wtDir, fileArgs)...)
				}
			}
		}
		return
	}
	if !zoneMutationVerbs[name] && zoneMutationVerbName(cmd.Args) == "" {
		return
	}
	w.mutating = true
	w.zoneCands(zonePathCandidates(cmd.Args[1:]))
}

// zoneMutationVerbName returns the mutation verb the command's EXECUTABLE
// word names in ANY world, or "" when none does. The shell path-resolves
// the executable, so the pre-4.2 reading's base name is what an old shell
// executes — a word whose modern reading truncates away the verb
// (`$'docs\u0000/../rm'`) is still judged as that verb (gate round 15 P1).
func zoneMutationVerbName(args []*syntax.Word) string {
	if len(args) == 0 {
		return ""
	}
	worlds, literal := zoneWordCandidates(args[0])
	if !literal {
		return ""
	}
	for _, t := range worlds {
		if t == "" {
			continue
		}
		name := t
		if strings.Contains(name, "/") {
			name = path.Base(name)
		}
		if zoneMutationVerbs[name] {
			return name
		}
	}
	return ""
}

// zonePathCandidates collects the candidates an argument list names: plain
// words, plus a long option's ATTACHED value — `cp --target-directory=zone_dir
// source.txt` writes into the option's value (round 13 P1). A value that is
// not a path matches nothing and costs one lookup. Short flags carry no
// extractable path here (a GNU short option with an attached value, `-tDIR`,
// stays an accepted under-match). Empty readings filter PER WORLD — the
// modern reading may truncate to "" at a leading code-point NUL while the
// pre-4.2 reading still names a real path — and the attached value is
// extracted from EVERY world's spelling (gate round 15 P1).
func zonePathCandidates(args []*syntax.Word) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		worlds, literal := zoneWordCandidates(a)
		if !literal {
			continue
		}
		kept := make([]string, 0, len(worlds))
		for _, wText := range worlds {
			if wText != "" {
				kept = append(kept, wText)
			}
		}
		if len(kept) == 0 {
			continue
		}
		t := kept[0]
		if strings.HasPrefix(t, "--") {
			for _, wText := range kept {
				if idx := strings.Index(wText, "="); idx >= 0 && idx+1 < len(wText) {
					out = append(out, wText[idx+1:])
				}
			}
			continue
		}
		if strings.HasPrefix(t, "-") {
			continue
		}
		out = append(out, kept...)
	}
	return out
}

// zoneLoopCount returns the exact iteration count of a for loop whose item
// list is fully literal — every item a plain unquoted word — and -1 when the
// count is unknown (a C-style loop, or any quoted or expanding item). A
// literal count above the cap is treated as unknown too: the fixed point
// below is the sound path for lists that long.
func zoneLoopCount(loop syntax.Loop) int {
	it, ok := loop.(*syntax.WordIter)
	if !ok {
		return -1 // C-style `for ((;;))`: unknown
	}
	n := 0
	for _, item := range it.Items {
		for _, p := range item.Parts {
			if _, isLit := p.(*syntax.Lit); !isLit {
				return -1
			}
		}
		n++
	}
	if n > zoneLoopLiteralCap {
		return -1
	}
	return n
}

const (
	// zoneLoopLiteralCap bounds the exact-iteration path; zoneLoopFixedPoint
	// bounds the fixed-point walk. Neither bound is reached by realistic
	// guard-relevant commands, and a walk that runs out of rounds sets the
	// walker's unbounded flag — the fail-closed denial — instead of
	// truncating silently (round 9 P1).
	zoneLoopLiteralCap = 64
	zoneLoopFixedPoint = 16
)

// zoneSymlinkDepthBound bounds how many symlink hops zoneResolve follows
// through a chain of dangling destinations; deeper chains read as outside
// the project (fail closed), standing in for the kernel's ELOOP.
const zoneSymlinkDepthBound = 32

// zoneRecursionBound bounds how many times a recursive call may re-enter
// its own body within one walk; past the bound the walk sets the unbounded
// flag — the fail-closed denial — instead of truncating the recursion
// silently (round 17 P1).
const zoneRecursionBound = 8

// zoneCwdsEqual compares two possible-directory sets member for member.
func zoneCwdsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// walkBodyFixedPoint walks a loop body whose iteration count is unknown.
// Each pass starts from the accumulated possible-directory set — exactly how
// the next iteration runs at the previous one's exit state (round 9 P1) —
// and the walk stops when a pass adds no new directory. A pass cap bounds
// the accumulation; exhausting it sets the unbounded flag (fail closed)
// rather than truncating the loop's states silently.
func (w *zoneWalker) walkBodyFixedPoint(cond []*syntax.Stmt, stmts []*syntax.Stmt) {
	for i := 0; i < zoneLoopFixedPoint; i++ {
		before := append([]string(nil), w.cwds...)
		beforeFuncs := cloneZoneFuncs(w.funcs)
		// the condition runs EVERY iteration, so it walks with the body —
		// each pass is one loop round (round 11 P1)
		for _, s := range cond {
			w.zoneWalkStmt(s)
		}
		for _, s := range stmts {
			w.zoneWalkStmt(s)
		}
		// convergence is reached only when the directory set AND the
		// function registry both stopped changing — a body that redefines a
		// function every round makes iteration 2 call a different body than
		// iteration 1 (round 16 P1)
		if zoneCwdsEqual(before, w.cwds) && zoneFuncsEqual(beforeFuncs, w.funcs) {
			return
		}
	}
	w.unbounded = true
}

// zoneFuncsEqual compares two function registries as body-pointer sets per
// name.
func zoneFuncsEqual(a, b map[string][]*syntax.Stmt) bool {
	if len(a) != len(b) {
		return false
	}
	for name, bodies := range a {
		other, ok := b[name]
		if !ok || len(other) != len(bodies) {
			return false
		}
		seen := make(map[*syntax.Stmt]bool, len(other))
		for _, s := range other {
			seen[s] = true
		}
		for _, s := range bodies {
			if !seen[s] {
				return false
			}
		}
	}
	return true
}

// cloneZoneFuncs copies the function registry — a subshell's redefinitions
// die with the subshell, so every subshell-shaped walk runs on its own copy
// (round 11 P1).
func cloneZoneFuncs(m map[string][]*syntax.Stmt) map[string][]*syntax.Stmt {
	out := make(map[string][]*syntax.Stmt, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// mergeZoneFuncs unions the possible bodies of every registry — a branch
// join leaves the name holding every definition any world gave it (round 14
// P1). Bodies are deduplicated by node identity.
func mergeZoneFuncs(maps ...map[string][]*syntax.Stmt) map[string][]*syntax.Stmt {
	out := make(map[string][]*syntax.Stmt)
	for _, m := range maps {
		for name, bodies := range m {
			for _, body := range bodies {
				dup := false
				for _, have := range out[name] {
					if have == body {
						dup = true
						break
					}
				}
				if !dup {
					out[name] = append(out[name], body)
				}
			}
		}
	}
	return out
}

// zoneWalkStmt walks one statement. Every branch's mutations and redirections
// are real; the working directory is a SET of possible values, and the walk
// unions the worlds control flow can produce (round 8, the operator-approved
// sound set semantics):
//   - a statement's redirections open BEFORE the statement runs, in the
//     walker's current directory — true for every statement shape, compound
//     ones included;
//   - a cd may fail, so its pre-cd set survives into the next statement;
//   - either side of a && or || may be skipped, so the post-left set survives
//     past the operator;
//   - an if's condition executes and its effects survive into every branch;
//     the then and else worlds union (an elif is an IfClause as the Else
//     member, whose condition executes too);
//   - a loop keeps its zero-iteration world and unrolls its body once — the
//     post-body set carries every candidate the repeated body can name;
//   - every case arm is judged from the same entry set, independently;
//   - a statement with Background (trailing "&"), a pipeline element, and an
//     explicit Subshell run in a subshell: their cd never leaks.
func (w *zoneWalker) zoneWalkStmt(stmt *syntax.Stmt) {
	if stmt == nil {
		return
	}
	pre := append([]string(nil), w.cwds...)
	var preFuncs map[string][]*syntax.Stmt
	if stmt.Background {
		preFuncs = cloneZoneFuncs(w.funcs)
	}
	defer func() {
		if stmt.Background {
			// the statement ran asynchronously in a subshell: its cd never
			// moved the main shell (round 4 P1 — the restore point is the
			// directory set before the whole backgrounded list), and its
			// function redefinitions die with it (round 11 P1)
			w.setCwds(pre)
			w.funcs = preFuncs
		}
	}()
	// redirections open before the statement runs, whatever its shape — a
	// nil command ("> f"), a block, a subshell, a compound clause (round 8)
	w.zoneRedirects(stmt.Redirs)
	switch cmd := stmt.Cmd.(type) {
	case nil:
		// a statement of redirections only (round 6 P1)
	case *syntax.BinaryCmd:
		switch cmd.Op {
		case syntax.Pipe, syntax.PipeAll:
			// a pipeline runs every element in a subshell: a cd inside it
			// never moves the main shell, and each element starts from the
			// same pre-pipe directory (round 3–4); redefinitions die with
			// their element (round 11 P1)
			side := append([]string(nil), w.cwds...)
			funcs := cloneZoneFuncs(w.funcs)
			w.zoneWalkStmt(cmd.X)
			w.setCwds(side)
			w.funcs = cloneZoneFuncs(funcs)
			w.zoneWalkStmt(cmd.Y)
			w.funcs = funcs
			w.setCwds(side)
		case syntax.AndStmt, syntax.OrStmt: // && and ||
			w.zoneWalkStmt(cmd.X)
			afterX := append([]string(nil), w.cwds...)
			xFuncs := cloneZoneFuncs(w.funcs)
			w.zoneWalkStmt(cmd.Y)
			// the right side may be skipped (the left failed under && or
			// succeeded under ||): the post-left world survives — directories
			// and function definitions alike (round 15 P1)
			w.funcs = mergeZoneFuncs(xFuncs, w.funcs)
			w.cwds = append(w.cwds, afterX...)
			w.setCwds(w.cwds)
		default:
			w.zoneWalkStmt(cmd.X)
			w.zoneWalkStmt(cmd.Y)
		}
	case *syntax.Subshell:
		side := append([]string(nil), w.cwds...)
		funcs := cloneZoneFuncs(w.funcs) // a subshell's redefinitions die with it (round 11 P1)
		for _, s := range cmd.Stmts {
			w.zoneWalkStmt(s)
		}
		w.funcs = funcs
		w.setCwds(side) // a subshell's cd never leaks
	case *syntax.Block:
		for _, s := range cmd.Stmts {
			w.zoneWalkStmt(s)
		}
	case *syntax.IfClause:
		w.walkIfChain(cmd)
	case *syntax.ForClause:
		// the zero-iteration world keeps the pre-loop function registry
		// (round 14 P1)
		preFuncs := cloneZoneFuncs(w.funcs)
		if n := zoneLoopCount(cmd.Loop); n >= 0 {
			// the item list is fully literal: exactly n iterations run, each
			// from the accumulated set (the next iteration starts where the
			// previous one left off — round 9 P1). A zero-item list runs the
			// zero-iteration world only (round 8 P1).
			for i := 0; i < n; i++ {
				for _, s := range cmd.Do {
					w.zoneWalkStmt(s)
				}
			}
		} else {
			w.walkBodyFixedPoint(nil, cmd.Do)
		}
		w.funcs = mergeZoneFuncs(preFuncs, w.funcs)
	case *syntax.WhileClause:
		// the iteration count is unknown: condition and body walk together
		// to the fixed point — the condition runs every iteration (round 11
		// P1), and the zero-iteration world (the first condition evaluation
		// failing) is the first pass's post-condition set, which the walk
		// already unions (round 8 P1; WhileClause.Until folds `until` into
		// the same shape). The zero-iteration world keeps the pre-loop
		// function registry too (round 14 P1).
		preFuncs := cloneZoneFuncs(w.funcs)
		w.walkBodyFixedPoint(cmd.Cond, cmd.Do)
		w.funcs = mergeZoneFuncs(preFuncs, w.funcs)
	case *syntax.CaseClause:
		entry := append([]string(nil), w.cwds...)
		entryFuncs := cloneZoneFuncs(w.funcs)
		worlds := append([]string(nil), entry...) // no arm may match: entry survives
		worldsFuncs := cloneZoneFuncs(entryFuncs)
		for _, item := range cmd.Items {
			// every arm starts from the case's entry set — and, sound over
			// `;&` and `;;&` fall-through, from every earlier arm's exit
			// state too (round 9 P1); the registries union the same way
			// (round 14 P1)
			w.setCwds(append(append([]string(nil), entry...), worlds...))
			w.funcs = mergeZoneFuncs(entryFuncs, worldsFuncs)
			for _, s := range item.Stmts {
				w.zoneWalkStmt(s)
			}
			worlds = append(worlds, w.cwds...)
			worldsFuncs = mergeZoneFuncs(worldsFuncs, w.funcs)
		}
		w.setCwds(worlds)
		w.funcs = mergeZoneFuncs(entryFuncs, worldsFuncs)
	case *syntax.TimeClause:
		// `time cmd` runs cmd, timed (round 10 P2)
		w.zoneWalkStmt(cmd.Stmt)
	case *syntax.FuncDecl:
		// a declaration alone runs nothing; the name registers so a later
		// call in the same command walks the body (round 10 P2). A
		// straight-line redefinition REPLACES — only a branch join unions
		// (round 14 P1).
		if cmd.Name != nil {
			w.funcs[cmd.Name.Value] = []*syntax.Stmt{cmd.Body}
		}
	case *syntax.CallExpr:
		w.zoneCall(cmd)
	default:
		// coprocesses, arithmetic, extended tests, zsh anonymous functions:
		// under-match
	}
}

// walkIfChain walks one if/elif/else clause. The condition executes and its
// effects survive into every branch (a successful cd inside a condition is
// where the branches run); the then branch runs from the post-condition set,
// the else branch from the post-condition set as well — and the Else member
// of an "elif" is itself an IfClause, whose condition executes too. The
// resulting set unions every world (rounds 6/8).
func (w *zoneWalker) walkIfChain(clause *syntax.IfClause) {
	if clause == nil {
		return
	}
	pre := append([]string(nil), w.cwds...)
	for _, s := range clause.Cond {
		w.zoneWalkStmt(s) // a condition executes (round 6 P1)
	}
	afterCond := append([]string(nil), w.cwds...)
	branchFuncs := cloneZoneFuncs(w.funcs) // the post-condition registry every branch starts from
	w.setCwds(afterCond)
	for _, s := range clause.Then {
		w.zoneWalkStmt(s)
	}
	afterThen := append([]string(nil), w.cwds...)
	thenFuncs := w.funcs
	w.setCwds(afterCond)
	w.funcs = cloneZoneFuncs(branchFuncs)
	w.walkIfChain(clause.Else)
	afterElse := append([]string(nil), w.cwds...)
	elseFuncs := w.funcs
	// union: the then world, the else world, the condition-false world —
	// directories and function definitions alike (round 14 P1)
	w.funcs = mergeZoneFuncs(branchFuncs, thenFuncs, elseFuncs)
	w.cwds = append(pre, afterThen...)
	w.cwds = append(w.cwds, afterElse...)
	w.setCwds(w.cwds)
}

// checkProtectedZoneShell decides one identity Bash call against the zone and
// returns a deny reason, or "" to let the call through. A command with no
// mutating form never reaches the manifest (REQ-SIPZ-008); a mutating command
// under an invalid manifest is denied unread (REQ-SIPZ-009). A command the
// parser rejects under-matches and passes — the accepted direction (spec
// §C.6), now bounded by a real parser instead of a hand-rolled splitter.
//
// @MX:SPEC:SPEC-SELF-IMPROVE-PROTECTED-ZONE-001
func (h *preToolHandler) checkProtectedZoneShell(agentID string, toolInput json.RawMessage) string {
	if !isZoneIdentity(agentID) {
		return ""
	}
	root := h.projectRoot()
	if root == "" {
		return ""
	}
	command := h.extractBashCommand(toolInput)
	if command == "" {
		return ""
	}
	file, ok := zoneParse(command)
	if !ok {
		return ""
	}
	w := &zoneWalker{h: h, cwds: []string{"."}, funcs: map[string][]*syntax.Stmt{}, calling: map[string]bool{}, calls: map[string]int{}}
	for _, stmt := range file.Stmts {
		w.zoneWalkStmt(stmt)
	}
	if !w.mutating {
		return ""
	}

	load := h.loadZone(root)
	if load.State == config.ZoneStateInvalid {
		// Fail closed: a mutating command cannot be checked against a zone of
		// unknown extent (REQ-SIPZ-009).
		reason := zoneDenyReason(agentID, "manifest", "invalid", load.InvalidFile)
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: load.InvalidFile,
			Decision: "deny", ManifestState: config.ZoneStateInvalid,
		})
		return reason
	}

	for _, cand := range w.cands {
		forms := resolveZoneTarget(root, cand)
		category, covered := zoneShellCovered(forms, load, root)
		if !covered {
			continue
		}
		display := cand
		if len(forms) > 0 {
			display = forms[0].Display
		}
		reason := zoneDenyReason(agentID, "category", category, display)
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: display,
			Category: category, Decision: "deny", ManifestState: load.State,
		})
		return reason
	}

	if w.unbounded && load.State != config.ZoneStateAbsent {
		// a loop whose directory states outgrew the fixed-point bound cannot
		// be verified against the zone: the mutating command is denied fail-
		// closed rather than allowed on an incomplete walk (round 9 P1)
		reason := zoneDenyReason(agentID, "category", "loop-unbounded", "loop")
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: "loop",
			Category: "loop-unbounded", Decision: "deny", ManifestState: load.State,
		})
		return reason
	}

	if load.State == config.ZoneStateAbsent {
		// Degrade visibly, mirroring the file-tool path (REQ-SIPZ-010).
		target := ""
		if len(w.cands) > 0 {
			target = w.cands[0]
		}
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: target,
			Decision: "allow", ManifestState: config.ZoneStateAbsent,
		})
	}
	return ""
}
