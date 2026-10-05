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
// The tracked directory is a SET of possible directories (round 6 P1): a
// candidate is denied when ANY possible directory covers it — sound over-
// approximation of bash's control flow:
//   - a statement with Background (trailing "&"), a pipeline element, and an
//     explicit Subshell run in a subshell: their cd never moves the main
//     shell, while their mutations and redirections are real;
//   - `&&` carries the updated directory to the right side; `||` restores the
//     pre-left set AND keeps the post-left set (the right side runs only when
//     the left failed, and a successful cd inside a failed chain persists);
//   - an if unions the condition-false, then, and else worlds; a condition
//     EXECUTES and is judged like any other statement list (round 6 P1);
//   - a cd's own redirections evaluate before the cd takes effect, and the
//     cd's directory arguments are its non-redirection words (the AST
//     separates them by construction);
//   - git's `-C <dir>` moves the directory its file arguments resolve against,
//     accumulated per possible start directory, consecutive and quoted
//     options included.
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
			b.WriteString(p.Value)
		case *syntax.SglQuoted:
			b.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, dp := range p.Parts {
				lit, ok := dp.(*syntax.Lit)
				if !ok {
					return "", false
				}
				b.WriteString(lit.Value)
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
// candidate is denied when any of them covers it (round 6 P1).
type zoneWalker struct {
	h        *preToolHandler
	cwds     []string
	mutating bool
	cands    []string
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
		switch rd.Op {
		case syntax.RdrIn, syntax.Hdoc, syntax.DashHdoc, syntax.WordHdoc, syntax.DplIn:
			continue
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
// mutating git subcommand — plus whatever the statement's redirections write.
// A leading assignment prefix is skipped, and `X=1 cd dir` is a cd (round 4).
// The statement's redirections evaluate in the walker's CURRENT directory —
// for a cd, before the cd takes effect (round 4 P1).
func (w *zoneWalker) zoneCall(stmt *syntax.Stmt, cmd *syntax.CallExpr) {
	// every redirection of the call judges here, in the walker's current
	// directory — for a cd, before the cd takes effect (round 4 P1)
	redirectsDone := false
	defer func() {
		if !redirectsDone {
			w.zoneRedirects(stmt.Redirs)
		}
	}()
	name, literal := zoneFirstArgWord(cmd.Args)
	if !literal {
		redirectsDone = true
		return // a dynamic command word under-matches
	}
	if name == "cd" {
		redirectsDone = true
		w.zoneRedirects(stmt.Redirs)
		var dirs []string
		for _, a := range cmd.Args[1:] {
			if t, lit := zoneWordText(a); lit {
				dirs = append(dirs, t)
			} else {
				dirs = append(dirs, "?dynamic")
			}
		}
		next := make([]string, 0, len(w.cwds))
		for _, cwd := range w.cwds {
			next = append(next, zoneNextCwd(cwd, dirs))
		}
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
		// every possible directory is a base the subcommand's file arguments
		// can resolve against; a -C <dir> moves that base, relative -C values
		// accumulating onto it (rounds 4–5 P1)
		for _, base := range w.cwds {
			gitDir := base
			for j := 0; j < len(cmd.Args); j++ {
				t, lit := zoneWordText(cmd.Args[j])
				if !lit {
					continue
				}
				if t == "-C" && j+1 < len(cmd.Args) {
					dir, lit2 := zoneWordText(cmd.Args[j+1])
					if !lit2 {
						break // under-match
					}
					if gitDir == "" || zoneIsAbs(dir) {
						gitDir = dir
					} else {
						gitDir = gitDir + "/" + dir
					}
					j++
					continue
				}
				if zoneGitMutating[t] {
					w.mutating = true
					var fileArgs []string
					for _, a := range cmd.Args[j+1:] {
						if ft, flit := zoneWordText(a); flit && ft != "--" {
							fileArgs = append(fileArgs, ft)
						}
					}
					// the candidates anchor to git's -C directory, not to the
					// shell's working directory (round 5 P1)
					w.cands = append(w.cands, zoneRelativeTo(gitDir, fileArgs)...)
					break
				}
			}
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

// zoneWalkStmt walks one statement. The mutations and redirections of every
// branch are real; the working directory becomes a SET of possible values —
// control flow multiplies it, and subshells (background statements, pipeline
// elements, explicit subshells) never leak their directory back.
func (w *zoneWalker) zoneWalkStmt(stmt *syntax.Stmt) {
	if stmt == nil {
		return
	}
	pre := append([]string(nil), w.cwds...)
	defer func() {
		if stmt.Background {
			// the statement ran asynchronously in a subshell: its cd never
			// moved the main shell (round 4 P1 — the restore point is the
			// directory set before the whole backgrounded list)
			w.setCwds(pre)
		}
	}()
	switch cmd := stmt.Cmd.(type) {
	case nil:
		// a statement of redirections only (round 6 P1)
		w.zoneRedirects(stmt.Redirs)
	case *syntax.BinaryCmd:
		switch cmd.Op {
		case syntax.Pipe, syntax.PipeAll:
			// a pipeline runs every element in a subshell: a cd inside it
			// never moves the main shell, and each element starts from the
			// same pre-pipe directory (round 3–4)
			side := append([]string(nil), w.cwds...)
			w.zoneWalkStmt(cmd.X)
			w.setCwds(side)
			w.zoneWalkStmt(cmd.Y)
			w.setCwds(side)
		case syntax.AndStmt: // &&
			w.zoneWalkStmt(cmd.X)
			w.zoneWalkStmt(cmd.Y)
		case syntax.OrStmt: // ||
			pre := append([]string(nil), w.cwds...)
			w.zoneWalkStmt(cmd.X)
			// the right side runs only if the left failed: the directory is
			// either where the left left it or where it started — keep both
			// (round 6 P1: a successful cd inside a failed chain persists)
			w.cwds = append(w.cwds, pre...)
			w.setCwds(w.cwds)
			w.zoneWalkStmt(cmd.Y)
		default:
			w.zoneWalkStmt(cmd.X)
			w.zoneWalkStmt(cmd.Y)
		}
	case *syntax.Subshell:
		side := append([]string(nil), w.cwds...)
		for _, s := range cmd.Stmts {
			w.zoneWalkStmt(s)
		}
		w.setCwds(side)
	case *syntax.Block:
		for _, s := range cmd.Stmts {
			w.zoneWalkStmt(s)
		}
		w.zoneRedirects(stmt.Redirs) // a block's own redirect (round 6 P1)
	case *syntax.IfClause:
		pre := append([]string(nil), w.cwds...)
		for _, s := range cmd.Cond {
			w.zoneWalkStmt(s) // a condition executes (round 6 P1)
		}
		afterCond := append([]string(nil), w.cwds...)
		for _, s := range cmd.Then {
			w.zoneWalkStmt(s)
		}
		afterThen := append([]string(nil), w.cwds...)
		w.setCwds(pre)
		w.zoneWalkIf(cmd.Else)
		// the condition-false world is the pre-if set; union every world
		w.cwds = append(w.cwds, afterCond...)
		w.cwds = append(w.cwds, afterThen...)
		w.setCwds(w.cwds)
		w.zoneRedirects(stmt.Redirs)
	case *syntax.ForClause:
		for _, s := range cmd.Do {
			w.zoneWalkStmt(s)
		}
	case *syntax.WhileClause:
		// the condition executes too (round 6 P1)
		for _, s := range cmd.Cond {
			w.zoneWalkStmt(s)
		}
		for _, s := range cmd.Do {
			w.zoneWalkStmt(s)
		}
	case *syntax.CaseClause:
		for _, item := range cmd.Items {
			for _, s := range item.Stmts {
				w.zoneWalkStmt(s)
			}
		}
	case *syntax.CallExpr:
		w.zoneCall(stmt, cmd)
	default:
		// function declarations, coprocesses, arithmetic, extended tests:
		// under-match
		return
	}
}

// zoneWalkIf walks an if/elif/else chain — the else member is itself an
// IfClause (an "elif") or carries no command; each branch's statements share
// the walker's directory set.
func (w *zoneWalker) zoneWalkIf(clause *syntax.IfClause) {
	if clause == nil {
		return
	}
	for _, s := range clause.Then {
		w.zoneWalkStmt(s)
	}
	w.zoneWalkIf(clause.Else)
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
	w := &zoneWalker{h: h, cwds: []string{"."}}
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
