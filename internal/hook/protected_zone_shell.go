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
// source text (`\ ` -> ` `, `\\` -> `\`, `\"` -> `"`). Every two-character
// escape resolves to its second character — a slight over-approximation
// inside double quotes, where `\n` is not an escape and the shell keeps the
// backslash: the guard prefers matching a file the command cannot touch over
// missing one it can (round 8 P1).
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
			b.WriteString(p.Value) // single quotes carry no escapes
		case *syntax.DblQuoted:
			for _, dp := range p.Parts {
				lit, ok := dp.(*syntax.Lit)
				if !ok {
					return "", false
				}
				b.WriteString(zoneUnescapeLit(lit.Value))
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

// zoneWalker carries one shell-policy walk. cwds is the set of POSSIBLE
// working directories at this point — control flow multiplies them, and a
// candidate is denied when any of them covers it (round 6 P1). unbounded
// records that a loop's directory states outgrew the fixed-point bound: the
// walk is then an over-approximation that cannot be completed, and the
// command is denied fail-closed rather than allowed on an incomplete walk
// (round 9 P1).
type zoneWalker struct {
	h         *preToolHandler
	cwds      []string
	mutating  bool
	unbounded bool
	cands     []string
	// funcs maps a function name declared in THIS command to its body — a
	// call runs the body in the caller's state (round 10 P2). calling holds
	// the names currently being walked, breaking recursive declarations.
	funcs   map[string]*syntax.Stmt
	calling map[string]bool
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
		t, literal := zoneWordText(rd.Word)
		if !literal || t == "" {
			continue // dynamic target: under-match
		}
		mutating = true
		targets = append(targets, t)
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
					// removes every file the glob matches there
					if zoneDirContainsGlobMatch(filepath.Join(root, filepath.FromSlash(form.Folded)), entry) {
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
	if body, declared := w.funcs[name]; declared {
		// a call to a function declared in this command runs its body in the
		// caller's directory state (round 10 P2); a recursive declaration
		// breaks the walk here, the documented under-match
		if w.calling[name] {
			return
		}
		w.calling[name] = true
		w.zoneWalkStmt(body)
		delete(w.calling, name)
		return
	}
	if name == "cd" {
		var dirs []string
		for _, a := range cmd.Args[1:] {
			if t, lit := zoneWordText(a); lit {
				dirs = append(dirs, t)
			} else {
				dirs = append(dirs, "?dynamic")
			}
		}
		next := make([]string, 0, len(w.cwds)*2)
		for _, cwd := range w.cwds {
			next = append(next, zoneNextCwd(cwd, dirs))
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
				if t == "--in-place" || (strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "--") && strings.Contains(strings.TrimPrefix(t, "-"), "i")) {
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
		dirOpt := ""
		sub := ""
		subIdx := -1
		for j := 1; j < len(cmd.Args); j++ {
			t, lit := zoneWordText(cmd.Args[j])
			if !lit {
				break // dynamic global argument: under-match
			}
			if strings.HasPrefix(t, "-") {
				if t == "-C" || t == "-c" || t == "--git-dir" || t == "--work-tree" || t == "--namespace" || t == "--super-prefix" {
					if t == "-C" && j+1 < len(cmd.Args) {
						if dir, lit2 := zoneWordText(cmd.Args[j+1]); lit2 {
							if dirOpt == "" || zoneIsAbs(dir) {
								dirOpt = dir
							} else {
								dirOpt = dirOpt + "/" + dir
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
			if ft, flit := zoneWordText(a); flit && ft != "--" && !strings.HasPrefix(ft, "-") {
				fileArgs = append(fileArgs, ft)
			}
		}
		// every possible directory is a base the subcommand's file arguments
		// can resolve against (round 6 P1); a -C moves that base — an absolute
		// -C replaces it, a relative one accumulates (round 5 P1)
		for _, base := range w.cwds {
			gitDir := base
			if dirOpt != "" {
				if zoneIsAbs(dirOpt) || gitDir == "" || gitDir == "." {
					gitDir = dirOpt
				} else {
					gitDir = gitDir + "/" + dirOpt
				}
			}
			w.cands = append(w.cands, zoneRelativeTo(gitDir, fileArgs)...)
		}
		return
	}
	if !zoneMutationVerbs[name] {
		return
	}
	w.mutating = true
	w.zoneCands(zonePathCandidates(cmd.Args[1:]))
}

// zonePathCandidates drops flags and empty words (an empty quoted word is
// not a path) from an argument list; the rest are the candidates the coverage
// check judges.
func zonePathCandidates(args []*syntax.Word) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		t, literal := zoneWordText(a)
		if !literal || t == "" || strings.HasPrefix(t, "-") {
			continue
		}
		out = append(out, t)
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
		// the condition runs EVERY iteration, so it walks with the body —
		// each pass is one loop round (round 11 P1)
		for _, s := range cond {
			w.zoneWalkStmt(s)
		}
		for _, s := range stmts {
			w.zoneWalkStmt(s)
		}
		if zoneCwdsEqual(before, w.cwds) {
			return
		}
	}
	w.unbounded = true
}

// cloneZoneFuncs copies the function registry — a subshell's redefinitions
// die with the subshell, so every subshell-shaped walk runs on its own copy
// (round 11 P1).
func cloneZoneFuncs(m map[string]*syntax.Stmt) map[string]*syntax.Stmt {
	out := make(map[string]*syntax.Stmt, len(m))
	for k, v := range m {
		out[k] = v
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
	var preFuncs map[string]*syntax.Stmt
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
			w.zoneWalkStmt(cmd.Y)
			// the right side may be skipped (the left failed under && or
			// succeeded under ||): the post-left set survives (round 8 P1)
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
	case *syntax.WhileClause:
		// the iteration count is unknown: condition and body walk together
		// to the fixed point — the condition runs every iteration (round 11
		// P1), and the zero-iteration world (the first condition evaluation
		// failing) is the first pass's post-condition set, which the walk
		// already unions (round 8 P1; WhileClause.Until folds `until` into
		// the same shape)
		w.walkBodyFixedPoint(cmd.Cond, cmd.Do)
	case *syntax.CaseClause:
		entry := append([]string(nil), w.cwds...)
		worlds := append([]string(nil), entry...) // no arm may match: entry survives
		for _, item := range cmd.Items {
			// every arm starts from the case's entry set — and, sound over
			// `;&` and `;;&` fall-through, from every earlier arm's exit
			// state too (round 9 P1)
			w.setCwds(append(append([]string(nil), entry...), worlds...))
			for _, s := range item.Stmts {
				w.zoneWalkStmt(s)
			}
			worlds = append(worlds, w.cwds...)
		}
		w.setCwds(worlds)
	case *syntax.TimeClause:
		// `time cmd` runs cmd, timed (round 10 P2)
		w.zoneWalkStmt(cmd.Stmt)
	case *syntax.FuncDecl:
		// a declaration alone runs nothing; the name registers so a later
		// call in the same command walks the body (round 10 P2)
		if cmd.Name != nil {
			w.funcs[cmd.Name.Value] = cmd.Body
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
	w.setCwds(afterCond)
	for _, s := range clause.Then {
		w.zoneWalkStmt(s)
	}
	afterThen := append([]string(nil), w.cwds...)
	w.setCwds(afterCond)
	w.walkIfChain(clause.Else)
	afterElse := append([]string(nil), w.cwds...)
	// union: the then world, the else world, the condition-false world
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
	w := &zoneWalker{h: h, cwds: []string{"."}, funcs: map[string]*syntax.Stmt{}, calling: map[string]bool{}}
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
