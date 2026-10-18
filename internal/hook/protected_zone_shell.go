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
	"runtime"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/modu-ai/moai-adk/internal/config"
)

// zoneMutationVerbs are the single-word mutating forms of REQ-SIPZ-007.
var zoneMutationVerbs = map[string]bool{
	"rm": true, "unlink": true, "mv": true, "rename": true, "cp": true, "tee": true, "truncate": true,
}

// zoneUnlinkVerbs are the unlink-family members of zoneMutationVerbs whose
// slash-less final component the shell removes AS ITSELF (card t1712): the
// unlink call takes the link's own directory entry and never the target
// behind it. mv, cp, tee and truncate are excluded on purpose — they follow
// a final-component symlink and write the target, so they keep the
// fail-closed follow.
var zoneUnlinkVerbs = map[string]bool{
	"rm": true, "unlink": true,
}

// zoneUnlinkLinkOwn reports whether an rm/unlink operand names its final
// component AS ITSELF: no trailing separator (a separator spelling names the
// directory THROUGH the link) and no "." or ".." final segment (those spell
// the directory too). A child path after the link keeps the link a PARENT
// component — the parent resolves as today and the final name rejoins onto
// the resolved parent. The separator set is platform-aware (card t1712
// repair round 2): "/" separates everywhere, "\" additionally on windows —
// `zonelink\.` names the directory through the link there — while on POSIX a
// backslash is an ordinary filename character
// (SPEC-HOOK-ZONE-BACKSLASH-001) and the same texts stay link-own finals.
// On WINDOWS a final component ending in "." or a space is NON-CANONICAL
// (card t1712 repair round 5): Win32 normalization strips trailing dots and
// spaces from the last component, so `secret.md.` and `secret.md ` alias the
// entry `secret.md` — the link-own resolver preserves the RAW final name and
// would judge an entry the shell never touches, so such a spelling stays on
// the target-following (full resolver) path, the fail-closed direction. On
// POSIX a trailing dot or space is an ordinary filename character and the
// literal entry is the one removed — link-own unchanged.
//
// WINDOWS DEFERS THE NARROWING ENTIRELY (card t1712 repair round 8): the
// narrowing is disabled outright — this predicate returns false for every
// windows operand, routing the unlink family to the PRE-CARD target-
// following full resolution. Rationale: windows filesystem name aliasing
// (8.3 short names, locale-specific case tables, fold pairs like ı/I that
// EqualFold-style folds miss but NTFS equates) cannot be enumerated by fold
// functions, and the stdlib offers no real-entry-name lookup that does not
// follow the final symlink; the target-following resolver's
// EvalSymlinks/Lstat canonicalization is the exact defense reviews 6-8 kept
// citing, so windows defers to it. The rounds-2..7 guards below (separator
// split, trailing dot/space, ~digit shape) stay as BELT-AND-SUSPENDERS for
// any future windows re-enable — they gate a path the early return makes
// unreachable today, not live routing. Residual: the pre-card over-block
// (a link-own removal through a final-component link is denied on windows
// where POSIX allows it) persists on windows — a separate card.
// windows rides in as an argument, the package's pathSegments idiom, so both
// branches classify identically on every host; the single production caller
// passes runtime.GOOS == "windows".
func zoneUnlinkLinkOwn(cand string, windows bool) bool {
	if cand == "" {
		return false
	}
	if windows {
		// class closure, not a mechanism patch (card t1712 repair round 8):
		// every windows operand keeps the pre-card target-following judgment
		return false
	}
	last := strings.LastIndexByte(cand, '/')
	if windows {
		if idx := strings.LastIndexByte(cand, '\\'); idx > last {
			last = idx
		}
		if strings.HasSuffix(cand, "\\") {
			return false
		}
	}
	if strings.HasSuffix(cand, "/") {
		return false
	}
	seg := cand[last+1:]
	if seg == "." || seg == ".." {
		return false
	}
	if windows && (strings.TrimRight(seg, ". ") != seg || hasShortNameShape(seg)) {
		// Win32 strips trailing dots/spaces from the final component and NTFS
		// 8.3 aliases a long name onto its "~N" short name — both spellings
		// name a DIFFERENT entry than the raw text judges, so they never
		// route to the link-own resolver (which preserves the raw name); the
		// target-following resolver's EvalSymlinks/Lstat work judges the
		// REAL entry (card t1712 repair rounds 5-6). Belt-and-suspenders
		// since round 8: unreachable while the windows early return above
		// defers the narrowing, kept for a future windows re-enable.
		return false
	}
	return true
}

// hasShortNameShape reports the NTFS 8.3 short-name shape — a "~" followed
// by a digit anywhere in the final component (SECRET~1.MD, REPORT~2.TXT):
// with 8.3 name generation enabled that spelling removes the long name it
// aliases, and the Go stdlib offers no short-name expansion, so the raw name
// cannot be canonicalized here. The classifier only detects the shape.
// Ceiling: 8.3 aliases outside the "~digit" shape are not covered.
func hasShortNameShape(seg string) bool {
	for i := 0; i+1 < len(seg); i++ {
		if seg[i] == '~' && seg[i+1] >= '0' && seg[i+1] <= '9' {
			return true
		}
	}
	return false
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

// zoneWordWorldReadings returns the word's text PER BASH GENERATION:
// [0] the modern reading, [1] the pre-4.2 reading — identical when the word
// carries no \u/\U (a generation-identical word contributes its one reading
// to both worlds). false when the word is dynamic. Funnel semantics that
// must not cross the worlds (anchor accumulation) index this pair (gate
// round 17 P2).
func zoneWordWorldReadings(w *syntax.Word) ([2]string, bool) {
	t, literal := zoneWordText(w)
	if !literal {
		return [2]string{}, false
	}
	r := [2]string{t, t}
	if zoneWordDual(w) {
		if old, lit := zoneWordTextPre42(w); lit {
			r[1] = old
		}
	}
	return r, true
}

// zoneDedupStrings drops duplicate readings, keeping the order — hash-set
// based (linear; the per-candidate scan of all prior candidates was O(n²)
// and dominated over-cap inputs, gate rounds 29/30 P2).
func zoneDedupStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// zoneCandidateCap bounds the candidate set: a walk naming more candidates
// than the cap cannot be judged soundly at this scale and is denied
// fail-closed — the bounded-walk philosophy (gate round 17 P2).
const zoneCandidateCap = 4096

// zoneCwd is one possible working directory carrying its GENERATION: gen
// -1 is generation-neutral (a bash-generation-identical reading put the
// walk here — every generation can be in it), 0 is a dir only the modern
// reading reached, 1 a dir only the pre-4.2 reading reached. A candidate
// reading of generation i joins only gen -1 and gen i directories — the
// cross-generation cwd×file join is a path no generation executes (gate
// round 22 P1).
type zoneCwd struct {
	dir string
	gen int
}

// zoneFuncBodies is one function name's registry entry: the bodies the name
// may have, and whether the declaration is CONDITIONAL — seen only inside a
// branch world (an if/&&/|| arm, a loop body, a case arm). A straight-line
// declaration replaces and is certain; a branch join marks a name absent
// from the pre-branch entry as conditional, and a call to a conditional name
// is judged under BOTH interpretations — the possibly-declared bodies AND
// the built-in/external verb resolution the shell would use when the
// declaration did not run (card t1574 K2, REQ-ZSP-002).
type zoneFuncBodies struct {
	bodies      []*syntax.Stmt
	conditional bool
}

// zoneWalker carries one shell-policy walk. cwds is the set of POSSIBLE
// working directories at this point — control flow multiplies them, and a
// candidate is denied when any of them covers it (round 6 P1); each entry
// carries the generation that reached it (gate round 22 P1). funcs is the
// function registry PER BASH GENERATION (gate round 25 P1): a function
// execution in one world registers only in that world's registry — the
// other world's externals stay external. world is the generation currently
// being walked: -1 neutral (generation-identical statement text), 0/1
// inside a generation-scoped function body. unbounded records that a
// loop's directory states outgrew the fixed-point bound: the walk is then
// an over-approximation that cannot be completed, and the command is
// denied fail-closed rather than allowed on an incomplete walk (round 9
// P1).
type zoneWalker struct {
	h        *preToolHandler
	root     string // project root; an in-project absolute cd is tracked under it (K2)
	cwds     []zoneCwd
	mutating bool
	// funcs holds each generation's function registry as zoneFuncBodies
	// values; calling holds the names currently being walked per generation;
	// calls counts the bounded re-entries a recursion may unroll per
	// generation (round 17 P1). unbounded records that a walk outgrew one of
	// its bounds — the command is then denied fail-closed rather than
	// allowed on an incomplete analysis. entries counts walker entries
	// across the whole command — the GLOBAL work bound of card t1574 K5
	// (REQ-ZSP-006), one counter across both bash generations: it bounds
	// total work, not per-generation work.
	unbounded bool
	cands     []string
	// linkCands holds the rm/unlink operands judged by the LINK'S OWN path
	// (card t1712): a slash-less final component is resolved only up to its
	// parent, the final component itself never followed. Every other
	// candidate — target-following spellings, the other verbs, redirection
	// targets — lands in cands as before.
	linkCands []string
	funcs     [2]map[string]*zoneFuncBodies
	calling   [2]map[string]bool
	calls     [2]map[string]int
	world     int
	entries   int
}

// setCwds replaces the possible-directory set, dropping duplicates (by
// directory AND generation — the same dir reached by both generations is
// two entries, each joining only its own generation's file readings).
func (w *zoneWalker) setCwds(dirs []zoneCwd) {
	seen := map[zoneCwd]bool{}
	w.cwds = w.cwds[:0]
	for _, d := range dirs {
		if d.dir == "" || seen[d] {
			continue
		}
		seen[d] = true
		w.cwds = append(w.cwds, d)
	}
}

// zoneMergeGens folds a directory reached in BOTH generations into the one
// generation-neutral entry the type already defines: a directory every
// generation can be in is exactly what neutral means — the neutral entry
// executes both worlds' cds and joins both generations' file readings — so
// the fold keeps the reachable-state set intact while dropping the
// per-generation PAIR whose duplication doubled the count the K8 bound
// reads (card t1574 K8, REQ-ZSP-006). Unfolded, one cd command's world 0
// and world 1 moves over the same accumulated set left the same directory
// tagged once per generation, and a chain of 4 literal cds outgrew
// zoneCwdsBound, denying a legitimate script fail-closed as loop-unbounded.
// The fold is a DEDUPLICATION of reachable states, never a truncation: the
// cd-may-fail conservatism (round 8 P1) keeps the pre-move set in the union
// before this runs, every directory the walk could be in survives, and a
// longer chain still denies fail-closed once it genuinely outgrows the
// bound. Order is stable: callers compose these sets positionally.
func zoneMergeGens(cwds []zoneCwd) []zoneCwd {
	reached := map[string]int{}
	for _, c := range cwds {
		switch c.gen {
		case 0:
			reached[c.dir] |= 1
		case 1:
			reached[c.dir] |= 2
		}
	}
	out := make([]zoneCwd, 0, len(cwds))
	seen := map[zoneCwd]bool{}
	for _, c := range cwds {
		if (c.gen == 0 || c.gen == 1) && reached[c.dir] == 3 {
			c = zoneCwd{dir: c.dir, gen: -1}
		}
		if seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	return out
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
func zoneRelativeToSet(cwds []zoneCwd, cand string, gen int) []string {
	if len(cwds) == 0 {
		return []string{cand}
	}
	out := make([]string, 0, len(cwds))
	for _, dir := range cwds {
		if gen != -1 && dir.gen == 1-gen {
			// the cross-generation cwd×file join is a path no generation
			// executes (gate round 22 P1)
			continue
		}
		if dir.dir == "" || dir.dir == "." || zoneIsAbs(cand) {
			out = append(out, cand)
		} else {
			out = append(out, dir.dir+"/"+cand)
		}
	}
	return out
}

// zoneCands expands every candidate against the walker's possible directories.
func (w *zoneWalker) zoneCands(cands [2][]string) {
	for world := 0; world < 2; world++ {
		if w.world >= 0 && w.world != world {
			// inside a generation-scoped function execution: only that
			// world's redirection candidates judge (gate round 27 P2)
			continue
		}
		for _, cand := range cands[world] {
			w.cands = append(w.cands, zoneRelativeToSet(w.cwds, cand, world)...)
		}
	}
}

// zoneRedirectTargets returns the targets of the write redirections in the
// list. Input redirects (`<`), here-docs, and `<&` read the target instead of
// writing it and are skipped; `>`, `>>`, `<>`, `>|`, `&>` and `&>>` create or
// truncate it.
func zoneRedirectTargets(redirs []*syntax.Redirect) (bool, [2][]string) {
	mutating := false
	var targets [2][]string
	for _, rd := range redirs {
		if rd.Op == syntax.DplOut {
			// `>&` onto a NUMBERED descriptor (`2>&1`) duplicates a file
			// descriptor and writes nothing; onto a word (`>& file`) it is a
			// file write in the dialects that accept the spelling. The
			// descriptor decision is PER WORLD: a reading that is numeric in
			// its world dup's the fd there, while the other world's reading
			// can be a real path that world WRITES (gate round 27 P1 —
			// `>& $'1\u0000/../zone_dir/marker.md'`: descriptor in the
			// modern world, a zone write in the pre-4.2 world).
			readings, literal := zoneWordWorldReadings(rd.Word)
			if !literal {
				continue // dynamic target: under-match
			}
			anyPath := false
			for world := 0; world < 2; world++ {
				if t := readings[world]; t != "" && !isZoneDigits(t) {
					anyPath = true
					mutating = true
					targets[world] = append(targets[world], t)
				}
			}
			if !anyPath {
				continue // numeric in both worlds: the fd dups, nothing writes
			}
			continue
		}
		switch rd.Op {
		case syntax.RdrIn, syntax.Hdoc, syntax.DashHdoc, syntax.WordHdoc, syntax.DplIn:
			continue
		}
		readings, literal := zoneWordWorldReadings(rd.Word)
		if !literal {
			continue // dynamic target: under-match
		}
		// an empty world filters PER WORLD: the modern reading may truncate
		// to "" at a leading code-point NUL while the pre-4.2 reading still
		// names the real write target (gate round 15 P1). The mutating flag
		// flips only when a NON-EMPTY target is added — bash cannot modify
		// a nonexistent empty path, so an all-empty reading must not trip
		// the fail-closed denial (gate round 33 P2)
		for world := 0; world < 2; world++ {
			if t := readings[world]; t != "" {
				mutating = true
				targets[world] = append(targets[world], t)
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
// possible working directory. During a generation-scoped function
// execution only THAT generation's redirect readings flip the mutating
// flag — a never-executed world's reading must not (gate round 31 P2).
func (w *zoneWalker) zoneRedirects(redirs []*syntax.Redirect) {
	hits, targets := zoneRedirectTargets(redirs)
	if w.world >= 0 {
		hits = len(targets[w.world]) > 0
	}
	if hits {
		w.mutating = true
		w.zoneCands(targets)
	}
}

// zoneNextCwd returns the working directory after a cd from cur. Only a plain
// literal relative argument is tracked — a bare cd, several arguments,
// substitution, a glob, an option word, or an absolute path outside the root
// resets to the root, the documented under-match (merge-gate round 1 P1-3).
// Several arguments reset only as a pooled set: bash 3.2.57 moves into the
// first operand and ignores the rest, so zoneCdMove keeps this reset and also
// tracks that first operand (SPEC-ZONE-SHELL-CD-TRACKING-001 F1/F2). An option
// word before it also moves the shell, but that form keeps the reset alone: the
// known limitation of this change, not tracked. An absolute
// path inside the root is tracked under its root-relative form (REQ-ZSCD-002,
// SPEC-ZONE-SHELL-CD-TRACKING-001 K2): the shell stands there whatever cur was,
// so the result never concatenates onto cur. The dots are cleaned here because
// the shell's cd already resolved the directory: this is the logical cwd the
// next segment's relative names concatenate onto, never a pre-clean of a
// target path.
func zoneNextCwd(root, cur string, dirs []string) string {
	if len(dirs) != 1 {
		return "."
	}
	arg := dirs[0]
	if arg == "-" || strings.ContainsAny(arg, "$*?") {
		return "."
	}
	if zoneIsAbs(arg) {
		// the lexical form only (REQ-SIPZ-006 semantics, zoneLexicalRel): an
		// absolute path outside the root, or one whose cleaned form escapes it,
		// keeps the reset (REQ-ZSCD-003)
		rel, inside := zoneLexicalRel(root, arg)
		if !inside {
			return "."
		}
		return rel
	}
	if strings.HasPrefix(arg, "-") {
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
				if entry.Kind == config.ZoneUserRoot {
					// a user-root Pattern is root-relative ("claude-skills/moai-demo/"), so it
					// cannot name a project parent; user-root targets are judged by zoneUserRootCovered
					continue
				}
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
// mutating git subcommand, a declared function, a cd. The statement's
// redirections are judged by the walker before this runs — the shell opens
// them before the command executes, for every statement shape (round 8).
// The two bash worlds are FULLY ISOLATED interpretations: each world
// carries its own function registry, its own name→verb binding, and its
// own argument readings; mutable state never flows between world
// iterations, and the worlds meet only at the deny union (gate rounds
// 23/25). A cd world's directory move composes per world and applies to
// subsequent statements only.
func (w *zoneWalker) zoneCall(cmd *syntax.CallExpr) {
	if _, literal := zoneFirstArgWord(cmd.Args); !literal {
		return // a dynamic command word under-matches
	}
	rawReadings, _ := zoneWordWorldReadings(cmd.Args[0])
	// card t1574 K3: strip the static wrapper set — `command`/`builtin`
	// bypass function lookup and `env`/`nohup` are external binaries that
	// exec their argument, so once a wrapper is present the stripped head
	// resolves as an EXECUTABLE ONLY and never re-enters the funcs table:
	// `rm() { :; }; command rm zone_dir/x` must deny, never walk the no-op
	// body (REQ-ZSP-003). `builtin` over-matches by design — bash rejects
	// `builtin rm` and deletes nothing, so judging it as the verb is the
	// safe direction (spec D3). The strip is STRUCTURAL — generation-
	// agnostic AST surgery, the wrapper heads are plain literals — while
	// the stripped head's NAME judges per world in the loop below.
	args, wclass := zoneStripWrapperPrefix(cmd.Args)
	var pathCands [2][]string
	if len(args) > 0 {
		// the strip may consume EVERY word (`env FOO=1` with no command
		// head): an empty stripped head under-matches, and slicing it here
		// would panic
		pathCands = zonePathCandidates(args[1:])
	}
	for world := 0; world < 2; world++ {
		if w.world >= 0 && w.world != world {
			continue // inside a generation-scoped body walk: only that world
		}
		raw := rawReadings[world]
		if raw == "" {
			continue
		}
		// card t1574 K4 per world: the funcs lookup runs on the ORIGINAL
		// command-word reading — a shell dispatches functions by bare name
		// only, so a reading carrying a path separator (/bin/rm, ./rm) can
		// NEVER resolve to a function in its own world. The old order (path
		// fold first, funcs lookup second) let `rm() { :; }; /bin/rm
		// zone_dir/secret.md` walk the no-op body while real bash ran the
		// external binary and deleted the file (REQ-ZSP-004).
		def, declared := w.funcs[world][raw]
		conditional := declared && def.conditional
		if declared {
			// this world walks the declared function. A CERTAIN declaration
			// walks and this world's verb analysis is skipped; a CONDITIONAL
			// one (card t1574 K2) walks the possibly-declared bodies as one
			// interpretation and FALLS THROUGH to the built-in/external
			// resolution the shell uses when the declaration did not run —
			// the call is denied when either interpretation pairs a mutating
			// form with a zone-covered target (REQ-ZSP-002).
			w.walkFunctionBodies(world, raw)
			if !conditional {
				continue
			}
		}
		if len(args) == 0 {
			continue // an empty stripped head under-matches exactly as today
		}
		headReadings, hlite := zoneWordWorldReadings(args[0])
		if !hlite {
			continue // the stripped head is dynamic: under-match as today
		}
		name := headReadings[world]
		if strings.Contains(name, "/") {
			// a literal executable path (/bin/rm, ./rm) names the verb
			// through its base (round 13 P1) — card t1574 K4: for the VERB
			// judgment only, never the funcs table above; the stripped
			// executable may itself be a path (`env /bin/rm x`), and a path
			// to some OTHER binary folds to a base that matches no verb,
			// exactly as before
			name = path.Base(name)
		}
		switch {
		case name == "cd" && !strings.Contains(headReadings[world], "/") &&
			(wclass == zoneWrapperNone || wclass == zoneWrapperShell):
			// A BARE cd, and a bare-named cd under the shell-builtin-
			// preserving wrappers (card t1685, REQ-ZSCD-004): `command cd`
			// and `builtin cd` execute the builtin IN THE CURRENT SHELL, so
			// the directory argument really moves the parent shell and must
			// be tracked — the move is judged on the STRIPPED operand list,
			// so the wrapper words never read as cd operands. Two cd shapes
			// stay untracked (card t1574; card-review P2-c): a cd under the
			// external wrappers — `env cd` is not a working cd at all and
			// `nohup cd` runs in a child process — and a PATH-FORM cd
			// (`command /usr/bin/cd`, `./cd`): the K4 rule above folds the
			// path to the verb name for the judgment, but a path form never
			// resolves to a shell builtin — it execs as an external child
			// whose cwd cannot move the parent shell.
			w.zoneCdMove(world, args)
		case zoneMutationVerbs[name]:
			// the verb's world binds its own argument readings (gate
			// rounds 23/25)
			w.mutating = true
			if zoneUnlinkVerbs[name] {
				// card t1712: an rm/unlink operand whose link is its last
				// component with nothing after it is judged by the LINK'S
				// OWN path — the unlink removes the link, never the target.
				// Every other spelling keeps the target-following judgment,
				// and so does every reading of a DUAL word: the two bash
				// generations render \u/\U differently, and the preserved
				// t1585 family freezes the full-follow judgment for those
				// candidates (gate rounds 13/14).
				zonePathCandidatesWorld(args[1:], world, func(cand string, dual bool) {
					joined := zoneRelativeToSet(w.cwds, cand, world)
					if !dual && zoneUnlinkLinkOwn(cand, runtime.GOOS == "windows") {
						w.linkCands = append(w.linkCands, joined...)
						return
					}
					w.cands = append(w.cands, joined...)
				})
			} else {
				w.zoneCandsWorld(world, pathCands[world])
			}
		case name == "sed":
			if zoneSedInPlace(world, args) {
				// in-place is decided by THIS command's options alone — an
				// earlier mutating command must not turn a read-only sed
				// into a denial (round 6 P2)
				w.mutating = true
				w.zoneCandsWorld(world, pathCands[world])
			}
		case name == "git":
			w.zoneGitArgs(world, args)
		}
	}
}

// zoneCandsWorld expands one world's candidate readings against the
// directories that world's generation can be in (gate round 22 P1).
func (w *zoneWalker) zoneCandsWorld(world int, cands []string) {
	for _, cand := range cands {
		w.cands = append(w.cands, zoneRelativeToSet(w.cwds, cand, world)...)
	}
}

// zoneCdMove composes one world's cd: a generation-tagged directory moves
// only under its own generation's reading, a neutral directory splits into
// per-generation entries, and the pre-move set survives (the cd may fail —
// round 8 P1). The move applies LAST for this command: subsequent
// statements see it, this command's other worlds never did (gate round 21
// P1). The K8 bound of card t1574 brackets the move (REQ-ZSP-006): a set
// already past the bound stops growing — the verdict is already fail-
// closed, and a loop's remaining passes must not square the set further —
// and a move that pushes the set past the bound flags the walk incomplete.
// args is the cd's OWN operand list — args[0] is the cd word itself, the
// wrapper words already stripped away (card t1685: a wrapped cd is judged on
// the stripped list, never with the wrapper as an operand).
func (w *zoneWalker) zoneCdMove(world int, args []*syntax.Word) {
	if len(w.cwds) > zoneCwdsBound {
		w.unbounded = true
		return
	}
	multi := len(args) != 2
	var pooled []string
	var reading string
	dual := false
	dynamic := multi
	// first is the operand bash moves into when several words reach cd
	// (SPEC-ZONE-SHELL-CD-TRACKING-001 F1). haveFirst records that one was
	// seen. optionFirst records an option word before any operand: the
	// union stays off for that cd, and the reset is the only reading.
	var first string
	haveFirst := false
	optionFirst := false
	if !multi {
		if r, lit := zoneWordWorldReadings(args[1]); !lit {
			dynamic = true
		} else {
			reading = r[world]
			dual = zoneWordDual(args[1])
		}
	} else {
		// SPEC-ZONE-SHELL-CD-TRACKING-001 K1 (REQ-ZSCD-001): the first bare `--`
		// ends cd's option parsing and is the separator, never a directory word.
		// It is dropped here, so `cd -- zone_dir` yields the one operand the
		// shell moves into. A second `--` is a literal operand. Several operands
		// keep the pooled reset; the first operand is tracked too, because bash
		// 3.2.57 moves into the first operand and ignores the rest (F1/F2). An
		// option word before that operand keeps the reset alone (a known limitation, not tracked).
		sepSeen := false
		for _, a := range args[1:] {
			t, lit := zoneWordText(a)
			if !lit {
				pooled = append(pooled, "?dynamic")
				if !haveFirst && !optionFirst {
					first, haveFirst = "?dynamic", true
				}
				continue
			}
			if !sepSeen && t == "--" {
				sepSeen = true
				continue
			}
			if sepSeen && t != "-" && strings.HasPrefix(t, "-") {
				// SPEC-ZONE-SHELL-CD-TRACKING-001 K3 (REQ-ZSCD-001): after the
				// separator a hyphen-leading word is an ordinary operand — the shell
				// moves `cd -- -zone` into ./-zone. The ./ spelling states that to the
				// lexical judgment. The bare `-` (OLDPWD, unknown to the guard) and the
				// words before the separator keep their option reset.
				t = "./" + t
			}
			// an option word before the first operand blocks the union: no first
			// operand is recorded after it. An option after the first operand does not.
			option := !sepSeen && t != "-" && strings.HasPrefix(t, "-")
			if option && !haveFirst {
				optionFirst = true
			}
			if !option && !haveFirst && !optionFirst {
				first, haveFirst = t, true
			}
			pooled = append(pooled, t)
		}
	}
	next := make([]zoneCwd, 0, len(w.cwds)*2)
	for _, cwd := range w.cwds {
		if cwd.gen != -1 && cwd.gen != world {
			continue // this world's cd never executes from another world's directory
		}
		gen := cwd.gen
		if gen == -1 {
			gen = world
		}
		switch {
		case dynamic:
			next = append(next, zoneCwd{dir: zoneNextCwd(w.root, cwd.dir, pooled), gen: gen})
			if multi && haveFirst {
				// SPEC-ZONE-SHELL-CD-TRACKING-001 F1/F2: bash 3.2.57 moves into the
				// first operand and ignores the rest. The reset above stays, and the
				// first operand is added as a second possible directory, so a following
				// deletion is judged under both outcomes. No first operand is recorded
				// when an option word comes before it. A single operand gives two
				// equal entries, which setCwds drops.
				next = append(next, zoneCwd{dir: zoneNextCwd(w.root, cwd.dir, []string{first}), gen: gen})
			}
		case !dual:
			// a generation-identical reading keeps the directory's
			// neutrality: the single reading serves this world too
			next = append(next, zoneCwd{dir: zoneNextCwd(w.root, cwd.dir, []string{reading}), gen: gen})
		default:
			if reading == "" {
				continue // an empty reading moves nothing for this world
			}
			next = append(next, zoneCwd{dir: zoneNextCwd(w.root, cwd.dir, []string{reading}), gen: gen})
		}
	}
	// the cd may fail (a missing directory leaves the caller where it
	// was): the pre-cd set survives into the next statement either way
	// (round 8 P1)
	next = append(next, w.cwds...)
	next = zoneMergeGens(next)
	w.setCwds(next)
	if len(w.cwds) > zoneCwdsBound {
		// card t1574 K8: every cd unions the pre-cd set into the next set,
		// so a chain of n cds squares the possible-directory set toward 2^n
		// states. A set past the bound is a walk that cannot complete: deny
		// fail-closed, never silent truncation (REQ-ZSP-006).
		w.unbounded = true
	}
}

// walkFunctionBodies executes a shadowing world's declared function: each
// candidate body runs from THIS world's entry-state snapshot and the final
// registry merges — a later candidate never erases an earlier
// registration (gate round 23 P1) — entirely inside the world's own state:
// its registry, its bindings, and its argument readings (gate round 25
// P1). A recursive call re-enters bounded and RESTORES the outer frame's
// accounting on frame exit (card t1574 K5, REQ-ZSP-005): the old frame
// exit deleted the in-flight marker, so an inner call's exit erased the
// OUTER frame's marker and the next sibling re-entry reset the bounded
// count to 1 — the recursion bound never held and a command real bash
// terminates instantly crashed the walker with a stack overflow. A
// conditional declaration's call keeps the possibility of absence (card
// t1574 K2): every registry entry the bodies INSTALLED stays conditional
// in this world.
func (w *zoneWalker) walkFunctionBodies(world int, name string) {
	def, ok := w.funcs[world][name]
	if !ok {
		return
	}
	savedWorld := w.world
	w.world = world
	defer func() { w.world = savedWorld }()
	prevCalling := w.calling[world][name]
	prevCalls := w.calls[world][name]
	if prevCalling {
		// a recursive call re-enters bounded: each re-entry walks the body
		// from the walker's CURRENT state (the recursion's directory moves
		// are real), and past the bound the command is fail-closed like any
		// other walk that cannot complete (round 17 P1)
		if prevCalls >= zoneRecursionBound {
			w.unbounded = true
			return
		}
		w.calls[world][name] = prevCalls + 1
	} else {
		w.calling[world][name] = true
		w.calls[world][name] = 1
	}
	entry := cloneZoneFuncs(w.funcs[world])
	var result map[string]*zoneFuncBodies
	for _, body := range def.bodies {
		w.funcs[world] = cloneZoneFuncs(entry)
		w.zoneWalkStmt(body)
		if result == nil {
			result = w.funcs[world]
		} else {
			result = mergeZoneFuncs(result, w.funcs[world])
		}
	}
	w.funcs[world] = result
	if def.conditional {
		// card-review P1 (t1574), the transitive instance of REQ-ZSP-002,
		// judged in THIS world's registry: the call itself may not have
		// executed — the declaring branch may not have run, and real bash
		// answers command-not-found — so every declaration the bodies
		// INSTALLED (absent from the pre-call entry) keeps the possibility
		// of absence. A certain call's bodies keep the straight-line
		// replacement semantics unchanged.
		w.funcs[world] = mergeZoneWorlds(entry, result)
	}
	if prevCalling {
		// restore the OUTER frame's accounting: the inner frame's exit must
		// not erase it (card t1574 K5, REQ-ZSP-005)
		w.calling[world][name] = true
		w.calls[world][name] = prevCalls
	} else {
		delete(w.calling[world], name)
		delete(w.calls[world], name)
	}
}

// zoneWrapperClass reports how the stripped head executes under the wrapper
// chain zoneStripWrapperPrefix consumed. The distinction decides cd tracking
// (card t1685, REQ-ZSCD-004): `command` and `builtin` execute their argument
// as a SHELL BUILTIN in the CURRENT shell — a wrapped `cd` really moves the
// parent shell — while `env` and `nohup` are external binaries that exec
// their argument in a child process whose cd cannot move the parent.
// EXTERNAL DOMINATES (card-review P2-b): once an external wrapper is in the
// chain the head runs in (or under) a child process regardless of any inner
// shell-preserving wrapper — `env command cd zone_dir` still cannot move the
// parent shell.
type zoneWrapperClass int

const (
	// zoneWrapperNone — no wrapper was stripped (a bare command head).
	zoneWrapperNone zoneWrapperClass = iota
	// zoneWrapperShell — the last stripped wrapper was `command` or
	// `builtin`; the stripped head still resolves in this shell.
	zoneWrapperShell
	// zoneWrapperExternal — the last stripped wrapper was `env` or `nohup`;
	// the stripped head execs in a child process.
	zoneWrapperExternal
)

// zoneStripWrapperPrefix strips the static wrapper set from a command's
// argument list — `command` (with its optional `-p` and its `--` separator),
// `env` (with the NAME=VALUE assignment words a real env consumes and its
// `-`-prefixed option words — a variable name can never start with `-`, so
// the two consumers are disjoint), `nohup`, and `builtin` (card t1574 K3,
// REQ-ZSP-003; card t1685 for the flag/separator spellings). The enumerated
// set is what the Claude Code permission layer strips (command/builtin) plus
// the card matrix (env, nohup); wrapper forms outside it (exec, nice,
// timeout, xargs, sudo, ...) stay the documented under-matches (spec D2).
// The loop is bounded by the argument list itself and handles nested forms
// (`env command rm x`); an empty stripped head under-matches exactly as
// today. The second result is the wrapper class governing how the stripped
// head executes, with EXTERNAL DOMINATING: the outermost external wrapper in
// the chain decides, and an inner `command`/`builtin` never demotes it (gate
// round: `env cd zone_dir` runs a child process whose cd cannot move the
// parent shell, while `command cd zone_dir` runs the builtin in this shell).
func zoneStripWrapperPrefix(args []*syntax.Word) ([]*syntax.Word, zoneWrapperClass) {
	class := zoneWrapperNone
	for len(args) > 0 {
		head, literal := zoneWordText(args[0])
		if !literal {
			return args, class
		}
		switch head {
		case "command":
			// card-review P2-b: an external wrapper already consumed stays
			// decisive — the inner shell-preserving wrapper itself runs in
			// that external child process.
			if class != zoneWrapperExternal {
				class = zoneWrapperShell
			}
			args = args[1:]
			if len(args) > 0 {
				if t, lit := zoneWordText(args[0]); lit && t == "-p" {
					args = args[1:]
				}
			}
			// card t1685 bypass 1a: the bare `--` ends `command`'s own option
			// parsing — the next word is the command head, never an option —
			// so it is consumed here and the real verb resolves after it.
			if len(args) > 0 {
				if t, lit := zoneWordText(args[0]); lit && t == "--" {
					args = args[1:]
				}
			}
		case "env":
			class = zoneWrapperExternal
			args = args[1:]
			for len(args) > 0 {
				t, lit := zoneWordText(args[0])
				if !lit {
					break
				}
				// card t1685 bypass 1b: env's option words (`-i`, `--`,
				// long forms) are `-`-prefixed and a NAME=VALUE assignment
				// can never start with `-`, so the two consumers are
				// disjoint and the loop order is irrelevant.
				if strings.HasPrefix(t, "-") {
					args = args[1:]
					// card-review P2-a/P2-d: the SPLIT forms of the unset
					// option (`-u NAME`, `--unset NAME`) consume one
					// following word — their NAME argument — so `env -u rm
					// echo …` resolves the read-only echo, never a phantom
					// rm verb. The glued forms (`-uNAME`, `--unset=NAME`)
					// carry the name inside the token; getopt takes a
					// required option argument from the next argv even when
					// it starts with a hyphen.
					if (t == "-u" || t == "--unset") && len(args) > 0 {
						args = args[1:]
					}
					// card-review P2-e: GNU env's info-only options print
					// and EXIT WITHOUT running the command — the trailing
					// words never execute. The strip returns an empty list,
					// and zoneCall's empty-stripped-head path performs no
					// verb judgment, no cd tracking, and collects no
					// candidates for this call.
					if t == "--help" || t == "--version" {
						return nil, class
					}
					continue
				}
				if !zoneEnvAssignment(t) {
					break
				}
				args = args[1:]
			}
		case "nohup":
			class = zoneWrapperExternal
			args = args[1:]
		case "builtin":
			if class != zoneWrapperExternal {
				class = zoneWrapperShell
			}
			args = args[1:]
		default:
			return args, class
		}
	}
	return args, class
}

// zoneEnvAssignment reports whether t is a NAME=VALUE assignment word — the
// environment-setting form env consumes before its command argument. The
// name part follows the shell's variable-name shape: letters, digits (never
// first), underscores.
func zoneEnvAssignment(t string) bool {
	idx := strings.IndexByte(t, '=')
	if idx <= 0 {
		return false
	}
	for i := 0; i < idx; i++ {
		c := t[i]
		switch {
		case c == '_':
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

// zonePathCandidates collects the candidates an argument list names: plain
// words, plus a long option's ATTACHED value — `cp --target-directory=zone_dir
// source.txt` writes into the option's value (round 13 P1). A value that is
// not a path matches nothing and costs one lookup. Short flags carry no
// extractable path here (a GNU short option with an attached value, `-tDIR`,
// stays an accepted under-match). Empty readings filter PER WORLD — the
// modern reading may truncate to "" at a leading code-point NUL while the
// pre-4.2 reading still names a real path — and the attached value is
// extracted from EVERY world's spelling (gate round 15 P1). Card t1574 K1:
// the bare `--` separator word ends the options — POSIX makes every
// following word an operand — so after it every fully-literal word, hyphen-
// leading included, is a candidate and the separator itself never becomes
// one (REQ-ZSP-001); the separator tracks PER WORLD, and dynamic words keep
// the existing under-match.
func zonePathCandidates(args []*syntax.Word) [2][]string {
	out := [2][]string{}
	for world := 0; world < 2; world++ {
		world := world
		zonePathCandidatesWorld(args, world, func(cand string, _ bool) {
			out[world] = append(out[world], cand)
		})
	}
	return out
}

// zonePathCandidatesWorld runs the shared option-word classification for ONE
// bash generation, handing each literal reading to add together with the
// word's dual-generation flag (whether the word carries \u/\U, the one
// escape family whose rendering differs across bash generations). The
// classification — the bare `--` separator, a long option's attached value,
// skipped short flags — is the per-world half of zonePathCandidates, split
// out so the unlink-family caller can decide each reading's disposition at
// the word it came from (card t1712: a dual word's readings keep the
// fail-closed target-following judgment and are never narrowed).
func zonePathCandidatesWorld(args []*syntax.Word, world int, add func(cand string, dual bool)) {
	var seenDashDash bool
	for _, a := range args {
		readings, literal := zoneWordWorldReadings(a)
		if !literal {
			continue
		}
		t := readings[world]
		if t == "" {
			continue
		}
		dual := zoneWordDual(a)
		if !seenDashDash {
			if t == "--" {
				seenDashDash = true
				continue
			}
			if strings.HasPrefix(t, "--") {
				if idx := strings.Index(t, "="); idx >= 0 && idx+1 < len(t) {
					add(t[idx+1:], dual)
				}
				continue
			}
			if strings.HasPrefix(t, "-") {
				continue
			}
		}
		add(t, dual)
	}
}

// zoneSedInPlace reports whether the argument list makes sed in-place for
// the PASSED generation: each option word is read through that world's
// reading — --in-place, GNU's suffixed --in-place=.bak form, or a short
// cluster carrying i (rounds 6 P2 / 13 P1). A word whose escape text
// decodes to an in-place spelling in the modern world reads as its
// literal escape text in the pre-4.2 world — an invalid option there —
// and must not fire the other world's sed (gate round 28 P2).
func zoneSedInPlace(world int, args []*syntax.Word) bool {
	for _, a := range args[1:] {
		readings, lit := zoneWordWorldReadings(a)
		if !lit {
			continue
		}
		if t := readings[world]; t == "--in-place" || strings.HasPrefix(t, "--in-place=") || (strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "--") && strings.Contains(strings.TrimPrefix(t, "-"), "i")) {
			return true
		}
	}
	return false
}

// zoneGitArgs judges a git command's mutating subcommand against the
// possible anchors. Every funnel semantic applies PER WORLD: -C
// accumulates per world — the modern reading extends the modern chain, the
// pre-4.2 reading the pre-4.2 chain, never crossed (the cartesian product
// blew up exponentially: 262,144 candidates for 2 unique paths at 18
// options, gate round 17 P2) — and --work-tree is OVERWRITE-WINS per
// world: git's LAST --work-tree replaces the anchor, and judging an
// already-overwritten anchor is a false deny (gate round 17 P2). Only the
// FIRST non-option word is the subcommand (round 7 P2); valued global
// options consume their argument (rounds 4-5 P1).
// zoneGitArgs judges a git command for ONE generation: everything here —
// the anchors, the file arguments, the subcommand — binds to the PASSED
// world's readings, because this world's name reached the git dispatch
// (gate round 27 P2: the inner world loops this replaces shadowed the
// passed generation and generated cross-generation candidates no
// generation executes). The subcommand word itself stays an exact-string
// match on the modern reading (the justified single-world exception — git
// dispatches subcommands internally, the shell never path-resolves them).
func (w *zoneWalker) zoneGitArgs(world int, args []*syntax.Word) bool {
	dirOpt := ""
	var wtOpts []string
	sub := ""
	subIdx := -1
	for j := 1; j < len(args); j++ {
		// the global-option classification and argument consumption bind to
		// the passed generation like everything else (gate round 34): a
		// word whose pre-4.2 reading is NOT a recognized global option is
		// not consumed in that world's sequence
		readings, lit := zoneWordWorldReadings(args[j])
		if !lit {
			break // dynamic global argument: under-match
		}
		t := readings[world]
		if strings.HasPrefix(t, "-") && zoneWordDual(args[j]) {
			// escape-origin word: split NAME vs VALUE. Only a NAME-origin
			// escape (the option NAME itself is \u/\U escape text)
			// terminates this world's sequence — the pre-4.2 world's git
			// refuses the unknown option (exit 129) and nothing after
			// executes in that world (gate round 34 P2). A \u/\U riding in
			// an option VALUE is ordinary data: 3.2 passes the raw string
			// as the value and the command RUNS (gate round 35 P1 — the
			// earlier whole-word termination over-fired on value escapes
			// and skipped a git that really deleted the marker).
			namePart := t
			if idx := strings.Index(t, "="); idx >= 0 {
				namePart = t[:idx]
			}
			if strings.Contains(namePart, `\u`) || strings.Contains(namePart, `\U`) {
				if world == 1 {
					return false
				}
			}
		}
		if strings.HasPrefix(t, "--work-tree=") {
			// git's LAST --work-tree wins: the option REPLACES the
			// anchor — judging an already-overwritten anchor is a
			// false deny (gate round 17 P2, preserved per world)
			wtOpts = []string{strings.TrimPrefix(t, "--work-tree=")}
			continue
		}
		if strings.HasPrefix(t, "-") {
			if t == "-C" || t == "-c" || t == "--git-dir" || t == "--work-tree" || t == "--namespace" || t == "--super-prefix" {
				if t == "-C" && j+1 < len(args) {
					if dirReadings, lit2 := zoneWordWorldReadings(args[j+1]); lit2 {
						if dir := dirReadings[world]; dir != "" {
							if dirOpt == "" || zoneIsAbs(dir) {
								dirOpt = dir
							} else {
								dirOpt = dirOpt + "/" + dir
							}
						}
					}
				}
				if t == "--work-tree" && j+1 < len(args) {
					if wtReadings, lit2 := zoneWordWorldReadings(args[j+1]); lit2 {
						if wt := wtReadings[world]; wt != "" {
							// overwrite-wins, as above
							wtOpts = []string{wt}
						}
					}
				}
				j++ // the option's value is consumed
			}
			continue
		}
		// the subcommand word binds its own generation: the sub drives the
		// analysis only for the world whose reading it is (gate round 29
		// P1) — superseding the earlier exact-string exception, whose
		// cross-generation join false-denied no-generation executions
		sub = t
		subIdx = j
		break
	}
	if subIdx == -1 || !zoneGitMutating[sub] {
		return false
	}
	w.mutating = true
	var fileArgs []string
	for _, a := range args[subIdx+1:] {
		readings, flit := zoneWordWorldReadings(a)
		if !flit {
			continue
		}
		// per-candidate emptiness/option checks (gate round 17 P1), bound
		// to the passed generation (gate round 27 P2)
		if t := readings[world]; t != "" && t != "--" && !strings.HasPrefix(t, "-") {
			fileArgs = append(fileArgs, t)
		}
	}
	// every possible directory is a base the subcommand's file arguments
	// can resolve against (round 6 P1); a -C moves that base — an absolute
	// -C replaces it, a relative one accumulates (round 5 P1); a
	// --work-tree is its own anchor (round 12 P1). The bases join only the
	// passed generation's file arguments (gate rounds 19/22/27).
	for _, base := range w.cwds {
		// the generation relation rides the base: a generation-tagged base
		// joins its own generation's file arguments, a neutral base joins
		// the calling generation (gate round 22 P1)
		if base.gen != -1 && base.gen != world {
			continue
		}
		gitDir := base.dir
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
			wtDir := base.dir
			if zoneIsAbs(wtOpt) || wtDir == "" || wtDir == "." {
				wtDir = wtOpt
			} else {
				wtDir = wtDir + "/" + wtOpt
			}
			w.cands = append(w.cands, zoneRelativeTo(wtDir, fileArgs)...)
		}
	}
	return true
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

// zoneWalkerEntryBudget bounds the walker's TOTAL work: one counter of
// zoneWalkStmt entries across the whole command. Restore-on-exit accounting
// bounds recursion DEPTH only — a width-W body over the recursion bound B
// yields Θ(W^(B-1)) body walks (W=10, B=8 ≈ 10^7) — so the total work needs
// its own cap. Past the budget the walk is incomplete: the unbounded flag is
// set and every deeper entry aborts immediately, and the command is denied
// fail-closed rather than allowed on an unfinished analysis (card t1574 K5,
// REQ-ZSP-006).
const zoneWalkerEntryBudget = 10000

// zoneCwdsBound caps how large a chain of cd statements may grow the
// possible-directory set. Every cd unions the pre-cd set into the next set,
// so a chain of n cds squares the set toward 2^n states; past the bound the
// walk cannot complete and is denied fail-closed — never truncated silently
// (card t1574 K8, REQ-ZSP-006).
const zoneCwdsBound = 16

// zoneCwdsEqual compares two possible-directory sets member for member.
func zoneCwdsEqual(a, b []zoneCwd) bool {
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
		before := append([]zoneCwd(nil), w.cwds...)
		beforeFuncs := cloneZoneFuncsState(w.funcs)
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
		if zoneCwdsEqual(before, w.cwds) && zoneFuncsEqualState(beforeFuncs, w.funcs) {
			return
		}
	}
	w.unbounded = true
}

// cloneZoneFuncsState copies BOTH generations' registries.
func cloneZoneFuncsState(s [2]map[string]*zoneFuncBodies) [2]map[string]*zoneFuncBodies {
	var out [2]map[string]*zoneFuncBodies
	for i := range s {
		out[i] = cloneZoneFuncs(s[i])
	}
	return out
}

// mergeZoneFuncsState unions both generations' registries.
func mergeZoneFuncsState(a, b [2]map[string]*zoneFuncBodies) [2]map[string]*zoneFuncBodies {
	var out [2]map[string]*zoneFuncBodies
	for i := range a {
		out[i] = mergeZoneFuncs(a[i], b[i])
	}
	return out
}

// mergeZoneWorldsState joins a branch world's post registry into the
// pre-branch entry PER GENERATION — each generation's registry keeps its
// own possibility-of-absence marking (card t1574 K2).
func mergeZoneWorldsState(entry, post [2]map[string]*zoneFuncBodies) [2]map[string]*zoneFuncBodies {
	var out [2]map[string]*zoneFuncBodies
	for i := range entry {
		out[i] = mergeZoneWorlds(entry[i], post[i])
	}
	return out
}

// mergeBranchJoinState joins an if-chain's EXHAUSTIVE branch worlds over
// the pre-branch entry PER GENERATION (gate round 4 of card t1574 on each
// generation's own registry).
func mergeBranchJoinState(entry [2]map[string]*zoneFuncBodies, worlds ...[2]map[string]*zoneFuncBodies) [2]map[string]*zoneFuncBodies {
	var out [2]map[string]*zoneFuncBodies
	for i := range entry {
		worldsSingle := make([]map[string]*zoneFuncBodies, 0, len(worlds))
		for _, world := range worlds {
			worldsSingle = append(worldsSingle, world[i])
		}
		out[i] = mergeBranchJoin(entry[i], worldsSingle...)
	}
	return out
}

// zoneFuncsEqualState compares both generations' registries.
func zoneFuncsEqualState(a, b [2]map[string]*zoneFuncBodies) bool {
	for i := range a {
		if !zoneFuncsEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// zoneFuncsEqual compares two function registries as body-pointer sets per
// name; the conditional dimension joins the comparison so the loop fixed
// point converges on it too (card t1574 K2).
func zoneFuncsEqual(a, b map[string]*zoneFuncBodies) bool {

	if len(a) != len(b) {
		return false
	}
	for name, def := range a {
		other, ok := b[name]
		if !ok || other.conditional != def.conditional || len(other.bodies) != len(def.bodies) {
			return false
		}
		seen := make(map[*syntax.Stmt]bool, len(other.bodies))
		for _, s := range other.bodies {
			seen[s] = true
		}
		for _, s := range def.bodies {
			if !seen[s] {
				return false
			}
		}
	}
	return true
}

// cloneZoneFuncs copies the function registry — a subshell's redefinitions
// die with the subshell, so every subshell-shaped walk runs on its own copy
// (round 11 P1). Entries are shared by pointer: a definition is replaced
// whole, never mutated in place.
func cloneZoneFuncs(m map[string]*zoneFuncBodies) map[string]*zoneFuncBodies {
	out := make(map[string]*zoneFuncBodies, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// mergeZoneFuncs unions the possible bodies of every registry — a branch
// join leaves the name holding every definition any world gave it (round 14
// P1). Bodies are deduplicated by node identity, and the conditional
// dimension ORs: a name conditional in any contributing world stays
// conditional in the union (card t1574 K2).
func mergeZoneFuncs(maps ...map[string]*zoneFuncBodies) map[string]*zoneFuncBodies {
	out := make(map[string]*zoneFuncBodies)
	for _, m := range maps {
		for name, def := range m {
			have, ok := out[name]
			if !ok {
				copied := *def
				copied.bodies = append([]*syntax.Stmt(nil), def.bodies...)
				out[name] = &copied
				continue
			}
			for _, body := range def.bodies {
				dup := false
				for _, candidate := range have.bodies {
					if candidate == body {
						dup = true
						break
					}
				}
				if !dup {
					have.bodies = append(have.bodies, body)
				}
			}
			if def.conditional {
				have.conditional = true
			}
		}
	}
	return out
}

// mergeZoneWorlds joins a branch world's post registry into the pre-branch
// entry: every name the post world holds but the entry lacks was declared
// only inside the branch — the branch may not have run — so the name keeps
// the POSSIBILITY OF ABSENCE and is marked conditional (card t1574 K2,
// REQ-ZSP-002). A name present in both stays certain: the shell defines it
// either way.
func mergeZoneWorlds(entry, post map[string]*zoneFuncBodies) map[string]*zoneFuncBodies {
	out := mergeZoneFuncs(entry, post)
	for name, def := range post {
		if _, was := entry[name]; !was {
			marked := *def
			marked.bodies = append([]*syntax.Stmt(nil), def.bodies...)
			marked.conditional = true
			out[name] = &marked
		}
	}
	return out
}

// mergeBranchJoin joins an if-chain's EXHAUSTIVE branch worlds over the
// pre-branch entry with the certain-intersection rule (gate round 4, card
// t1574): real bash executes exactly one branch, so a name declared
// certainly in EVERY contributing world is certain post-join — the
// then/else chain is the only such join in this walker. The entry world
// (the post-condition registry) holds the names defined before the branch:
// present in every reachable world by definition. A name missing from any
// reachable world, or conditional in any contributing world (an inner join
// already found a world lacking it), keeps the possibility of absence
// (REQ-ZSP-002, spec D1). The sibling joins must NOT use this: &&/||'s
// entry is the always-ran left side (mergeZoneWorlds is already exact
// there), and case plus the loops keep a may-skip world where the
// possibility of absence stays sound.
func mergeBranchJoin(entry map[string]*zoneFuncBodies, worlds ...map[string]*zoneFuncBodies) map[string]*zoneFuncBodies {
	out := mergeZoneFuncs(append([]map[string]*zoneFuncBodies{entry}, worlds...)...)
	for name, def := range out {
		if _, was := entry[name]; was {
			// defined before the branch: certain whatever the bodies did —
			// they only redefine it
			def.conditional = false
			continue
		}
		allCertain := true
		for _, world := range worlds {
			wdef, ok := world[name]
			if !ok || wdef.conditional {
				allCertain = false
				break
			}
		}
		def.conditional = !allCertain
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
	// card t1574 K5: the GLOBAL work bound and the walk abort. Past the
	// entry budget every deeper entry returns without descending, so an
	// exhausted walk stops immediately instead of letting the remaining
	// branches run (REQ-ZSP-006). The entry counter bounds TOTAL WORK across
	// the whole command — restore-on-exit bounds recursion DEPTH only, and a
	// width-W body over the recursion bound B yields Θ(W^(B-1)) body walks
	// the depth bound alone cannot contain. The recursion and cwds bounds
	// FLAG incompleteness without halting the walk: halting on them would
	// leave a covered statement following a bound-tripping construct
	// unjudged and flip the recorded category of commands the preserved
	// family freezes (e.g. `while false; do cd docs; done; rm zone_dir/
	// secret`), while their work is already bounded by this same counter.
	// The counter is ONE per analysis, shared across both bash generations —
	// it bounds total work, not per-generation work. Past the budget the
	// walk is incomplete → the unbounded flag → the fail-closed denial.
	w.entries++
	if w.entries > zoneWalkerEntryBudget {
		w.unbounded = true
		return
	}
	pre := append([]zoneCwd(nil), w.cwds...)
	var preFuncs [2]map[string]*zoneFuncBodies

	if stmt.Background {
		preFuncs = cloneZoneFuncsState(w.funcs)
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
			side := append([]zoneCwd(nil), w.cwds...)
			funcs := cloneZoneFuncsState(w.funcs)
			w.zoneWalkStmt(cmd.X)
			w.setCwds(side)
			w.funcs = cloneZoneFuncsState(funcs)
			w.zoneWalkStmt(cmd.Y)
			w.funcs = funcs
			w.setCwds(side)
		case syntax.AndStmt, syntax.OrStmt: // && and ||
			w.zoneWalkStmt(cmd.X)
			afterX := append([]zoneCwd(nil), w.cwds...)
			xFuncs := cloneZoneFuncsState(w.funcs)
			w.zoneWalkStmt(cmd.Y)
			// the right side may be skipped (the left failed under && or
			// succeeded under ||): the post-left world survives — directories
			// and function definitions alike (round 15 P1); a name the right
			// world declares but the left lacked keeps the possibility of
			// absence (card t1574 K2)
			w.funcs = mergeZoneWorldsState(xFuncs, w.funcs)

			w.cwds = append(w.cwds, afterX...)
			w.setCwds(w.cwds)
		default:
			w.zoneWalkStmt(cmd.X)
			w.zoneWalkStmt(cmd.Y)
		}
	case *syntax.Subshell:
		side := append([]zoneCwd(nil), w.cwds...)
		funcs := cloneZoneFuncsState(w.funcs) // a subshell's redefinitions die with it (round 11 P1)
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
		preFuncs := cloneZoneFuncsState(w.funcs)
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
		// the zero-iteration world may skip the body: a name declared only
		// inside it keeps the possibility of absence (card t1574 K2)
		w.funcs = mergeZoneWorldsState(preFuncs, w.funcs)

	case *syntax.WhileClause:
		// the iteration count is unknown: condition and body walk together
		// to the fixed point — the condition runs every iteration (round 11
		// P1), and the zero-iteration world (the first condition evaluation
		// failing) is the first pass's post-condition set, which the walk
		// already unions (round 8 P1; WhileClause.Until folds `until` into
		// the same shape). The zero-iteration world keeps the pre-loop
		// function registry too (round 14 P1).
		preFuncs := cloneZoneFuncsState(w.funcs)
		// gate round 3 (card t1574): the condition list ALWAYS executes at
		// least once — walk it once up front so the names it declares are
		// snapshotted and stay CERTAIN post-loop. The fixed point below
		// restarts from the pre-loop registry and re-walks condition and
		// body together, which only unions more worlds. Only the BODY may
		// run zero times: the join downgrades body-only declarations to
		// conditional, and the pass below un-downgrades every name the
		// condition snapshot holds.
		for _, s := range cmd.Cond {
			w.zoneWalkStmt(s)
		}
		condDeclared := cloneZoneFuncsState(w.funcs)
		// a FRESH copy, never preFuncs itself: the fixed point's live map
		// must stay a distinct object from the entry world the join below
		// treats as immutable, or the fixed point's declarations would
		// masquerade as pre-loop certainties and the absent-world marking
		// would never fire
		w.funcs = cloneZoneFuncsState(preFuncs)
		w.walkBodyFixedPoint(cmd.Cond, cmd.Do)
		// the zero-iteration world may skip the body: a name declared only
		// inside it keeps the possibility of absence (card t1574 K2)
		w.funcs = mergeZoneWorldsState(preFuncs, w.funcs)
		for i := range w.funcs {
			for name, def := range w.funcs[i] {
				if def.conditional {
					if _, was := condDeclared[i][name]; was {
						// declared by the condition list: every condition
						// evaluation defines it — certain whatever the body
						// did
						def.conditional = false
					}
				}
			}
		}

	case *syntax.CaseClause:
		entry := append([]zoneCwd(nil), w.cwds...)
		entryFuncs := cloneZoneFuncsState(w.funcs)
		worlds := append([]zoneCwd(nil), entry...) // no arm may match: entry survives
		worldsFuncs := cloneZoneFuncsState(entryFuncs)
		for _, item := range cmd.Items {
			// every arm starts from the case's entry set — and, sound over
			// `;&` and `;;&` fall-through, from every earlier arm's exit
			// state too (round 9 P1); the registries union the same way
			// (round 14 P1)
			w.setCwds(append(append([]zoneCwd(nil), entry...), worlds...))
			w.funcs = mergeZoneFuncsState(entryFuncs, worldsFuncs)
			for _, s := range item.Stmts {
				w.zoneWalkStmt(s)
			}
			worlds = append(worlds, w.cwds...)
			worldsFuncs = mergeZoneFuncsState(worldsFuncs, w.funcs)
		}
		w.setCwds(worlds)
		// no arm may match, and every arm is its own world: a name declared
		// only inside an arm keeps the possibility of absence (card t1574 K2)
		w.funcs = mergeZoneWorldsState(entryFuncs, worldsFuncs)

	case *syntax.TimeClause:
		// `time cmd` runs cmd, timed (round 10 P2)
		w.zoneWalkStmt(cmd.Stmt)
	case *syntax.FuncDecl:
		// a declaration alone runs nothing; the name registers PER WORLD —
		// generation-identical in both, per-reading when the name word is
		// dual (gate round 25 P1) — so a later call in the same command
		// walks the body in its own generation's registry (round 10 P2). A
		// straight-line redefinition REPLACES within its world and is
		// CERTAIN — only a branch join unions and marks the conditional
		// possibility of absence (round 14 P1; card t1574 K2).
		if cmd.Name != nil {
			worlds := []int{0, 1}
			if w.world >= 0 {
				worlds = []int{w.world}
			}
			name := cmd.Name.Value
			readings := [2]string{name, name}
			for _, world := range worlds {
				if readings[world] == "" {
					continue
				}
				w.funcs[world][readings[world]] = &zoneFuncBodies{bodies: []*syntax.Stmt{cmd.Body}}
			}
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
	pre := append([]zoneCwd(nil), w.cwds...)
	for _, s := range clause.Cond {
		w.zoneWalkStmt(s) // a condition executes (round 6 P1)
	}
	afterCond := append([]zoneCwd(nil), w.cwds...)
	branchFuncs := cloneZoneFuncsState(w.funcs) // the post-condition registry every branch starts from
	w.setCwds(afterCond)
	for _, s := range clause.Then {
		w.zoneWalkStmt(s)
	}
	afterThen := append([]zoneCwd(nil), w.cwds...)
	thenFuncs := w.funcs
	w.setCwds(afterCond)
	w.funcs = cloneZoneFuncsState(branchFuncs)
	var elseFuncs [2]map[string]*zoneFuncBodies
	var afterElse []zoneCwd
	if clause.Else != nil && len(clause.Else.Cond) == 0 {
		// the TERMINAL else list (gate round 4, card t1574): mvdan folds a
		// plain `else` into an IfClause carrying no condition — its
		// statements are EXHAUSTIVE, exactly one of then/else runs and
		// there is no skip world at this level. Walk it as a plain world.
		for _, s := range clause.Else.Then {
			w.zoneWalkStmt(s)
		}
		elseFuncs = w.funcs
		afterElse = append([]zoneCwd(nil), w.cwds...)
	} else {
		// an elif level (or no else at all): another conditional world whose
		// condition may fail — the skip world is the pre-branch registry
		// the recursion starts from (a nil Else returns immediately,
		// leaving that clone in place)
		w.walkIfChain(clause.Else)
		elseFuncs = w.funcs
		afterElse = append([]zoneCwd(nil), w.cwds...)
	}
	// union: the then world, the else world, the condition-false world —
	// directories and function definitions alike per bash generation (round
	// 14 P1); a name declared only inside a branch keeps the possibility of
	// absence (card t1574 K2), while a name declared certainly in EVERY
	// branch world is certain — the then/else pair is exhaustive (gate
	// round 4, card t1574). The recursive elif side folds its own chain
	// first, so its certainty flags already encode the deeper intersection.
	w.funcs = mergeBranchJoinState(branchFuncs, thenFuncs, elseFuncs)

	w.cwds = append(pre, afterThen...)
	w.cwds = append(w.cwds, afterElse...)
	w.setCwds(w.cwds)
}

// zoneDenyIfCovered judges one candidate's resolved forms against the zone
// and, when covered, records the audit row and returns the deny reason; ""
// continues the walk. Shared by the target-following and link-own candidate
// lists (card t1712).
func (h *preToolHandler) zoneDenyIfCovered(agentID, root string, load config.ProtectedZoneLoad, cand string, forms []zoneForm) string {
	category, covered := zoneShellCovered(forms, load, root)
	display := cand
	if len(forms) > 0 {
		display = forms[0].Display
	}
	// M4 (SPEC-USERASSET-DEPLOY-GUARD-001): the user-root arm — its OWN
	// branch (gate round 26 #3): user-root forms match only ZoneUserRoot
	// entries, filtered by the manifest-tracked containment check (the
	// manifest records files; a directory mutation is covered when any
	// tracked key lives under it). The project/baseline rules above
	// never see the user-root forms. Absorbed into this seam at the
	// t1712 absorb of a9a4d6ea0521 so both candidate lists inherit it.
	if !covered && load.State == config.ZoneStateOK {
		if uCat, uDisplay, uCovered := zoneUserRootCovered(cand, load); uCovered {
			category, covered, display = uCat, true, uDisplay
		}
	}
	if !covered {
		return ""
	}
	reason := zoneDenyReason(agentID, "category", category, display)
	h.recordZoneAudit(root, zoneAuditRow{
		Identity: agentID, Tool: "Bash", Path: display,
		Category: category, Decision: "deny", ManifestState: load.State,
	})
	return reason
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
	w := &zoneWalker{h: h, root: root, cwds: []zoneCwd{{dir: ".", gen: -1}}, funcs: [2]map[string]*zoneFuncBodies{{}, {}}, calling: [2]map[string]bool{{}, {}}, calls: [2]map[string]int{{}, {}}, world: -1}

	for _, stmt := range file.Stmts {
		w.zoneWalkStmt(stmt)
	}
	if !w.mutating && !w.unbounded {
		// the read-only fast path (gate round, card t1574): no mutating form
		// and a COMPLETED walk never reaches the manifest (REQ-SIPZ-008) —
		// the common case keeps its zero-file-access baseline. The unbounded
		// denial below stays ahead of every allow answer.
		return ""
	}

	load := h.loadZone(root)
	if load.State == config.ZoneStateInvalid {
		// Fail closed: a mutating command cannot be checked against a zone of
		// unknown extent (REQ-SIPZ-009). The reason precedence is the parent
		// SPEC's error contract (gate round 2, card t1574): the operator must
		// learn WHICH manifest file to repair, so this branch precedes the
		// per-candidate denies and the unbounded deny — every such walk still
		// answers deny, and the surviving reason is the one that names the
		// repair. The read-only fast path above already returned for commands
		// with no mutating form and a completed walk.
		reason := zoneDenyReason(agentID, "manifest", "invalid", load.InvalidFile)
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: load.InvalidFile,
			Decision: "deny", ManifestState: config.ZoneStateInvalid,
		})
		return reason
	}

	// dedup BEFORE the cap: a unicode-free command contributes the same
	// path from both worlds, and capping before dedup would false-deny a
	// plain command whose deduped set fits (gate round 19 P2). The link-own
	// list dedups the same way, and the cap counts the DEDUPLICATED UNION of
	// both lists by EXACT candidate string (card t1712 repair round 3): one
	// command naming the same path under rm (link-own) and tee
	// (target-following) lands the SAME STRING in BOTH lists, and counting
	// the memberships additively denied a set whose unique population is
	// within the cap. The key stays exact — never folded: on a
	// case-sensitive filesystem paths differing only by case are DISTINCT
	// paths (card t1712 repair round 4), so a folded key would collapse an
	// over-cap case-variant flood under the cap and let it reach
	// per-candidate resolution. Over-counting case variants on a
	// case-insensitive host is the fail-closed direction. Each membership
	// still judges under its own semantics below — a path present in both
	// lists stays deny when either reading hits the zone.
	w.cands = zoneDedupStrings(w.cands)
	w.linkCands = zoneDedupStrings(w.linkCands)
	unique := make(map[string]bool, len(w.cands)+len(w.linkCands))
	for _, list := range [2][]string{w.cands, w.linkCands} {
		for _, cand := range list {
			unique[cand] = true
		}
	}
	overCap := len(unique) > zoneCandidateCap
	if overCap {
		// a candidate set beyond the cap cannot be judged soundly at this
		// scale — denied fail-closed REGARDLESS of manifest state: a set
		// too large to verify is unverifiable whether or not a manifest
		// exists, and the compiled baseline floor alone (frozenZonePrefixes
		// + frozenInstructionFiles) justifies protection (gate round 32
		// P1). Fail-closed fires IMMEDIATELY: no per-candidate resolution
		// of an over-cap set (gate rounds 29/30 P2)
		reason := zoneDenyReason(agentID, "category", "loop-unbounded", "over-cap")
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: "over-cap",
			Category: "loop-unbounded", Decision: "deny", ManifestState: load.State,
		})
		return reason
	}
	// one LISTING CACHE per guard call (repair round 12 P2): N candidates
	// sharing a resolved parent read that parent's directory listing once —
	// never carried across guard calls.
	listings := make(map[string]*zoneNameIndex)
	// the identity backstop (card t1804): one per guard call, like the listings
	identity := newZoneIdentity(root, load.Zone)
	for _, cand := range w.cands {
		if reason := h.zoneDenyIfCovered(agentID, root, load, cand, identity.add(resolveZoneTarget(root, cand, listings), false)); reason != "" {
			return reason
		}
	}
	for _, cand := range w.linkCands {
		// card t1712: the link-own candidates resolve with the final
		// component unfollowed — the link's own path is all the judgment
		// sees, so a link outside the zone pointing at a zone file passes.
		if reason := h.zoneDenyIfCovered(agentID, root, load, cand, identity.add(resolveZoneLinkTarget(root, cand, listings), true)); reason != "" {
			return reason
		}
	}

	if w.unbounded {
		// a loop whose directory states outgrew the fixed-point bound, a
		// walk past the work budget, or a cd chain past the cwds bound
		// cannot be verified against the zone: the mutating command is
		// denied fail-closed rather than allowed on an incomplete walk
		// (round 9 P1). card t1574 K7: the unbounded denial also covers the
		// work-budget abort, which can fire BEFORE the mutating statement is
		// ever reached, leaving `mutating` false — the fast path above would
		// answer allow while real bash deletes (REQ-ZSP-006, audit D5). A
		// walk that flagged a bound AFTER collecting a covered candidate
		// answered at the loop above, whose more specific category stays the
		// recorded verdict (the preserved family freezes those shapes). The
		// denial applies REGARDLESS of manifest state — an unverifiable walk
		// in a manifest-less project still skips the compiled baseline floor
		// that protects AGENTS.md and the frozen instruction files (the same
		// Absent carve-out hole as the over-cap deny, gate round 32, sibling
		// audit), and an incomplete walk may not answer allow, so the
		// ZoneStateAbsent degrade below keeps its COMPLETED-walk meaning
		// only (card t1574, audit D9a). Past this branch a mutating command
		// is guaranteed: the fast path returned early on `!mutating &&
		// !unbounded`.

		reason := zoneDenyReason(agentID, "category", "loop-unbounded", "loop")
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: "loop",
			Category: "loop-unbounded", Decision: "deny", ManifestState: load.State,
		})
		return reason
	}

	if load.State == config.ZoneStateAbsent {
		// Degrade visibly, mirroring the file-tool path (REQ-SIPZ-010). This
		// is the COMPLETED-walk meaning only (card t1574, audit D9a): the
		// walk above finished and collected its candidates — an ABORTED walk
		// never reaches here, the unbounded denial having answered first.
		target := ""
		if len(w.cands) > 0 {
			target = w.cands[0]
		} else if len(w.linkCands) > 0 {
			target = w.linkCands[0]
		}
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: target,
			Decision: "allow", ManifestState: config.ZoneStateAbsent,
		})
	}
	return ""
}
