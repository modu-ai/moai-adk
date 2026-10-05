package hook

// protected_zone_shell.go — the shell half of the protected-zone guard
// (SPEC-SELF-IMPROVE-PROTECTED-ZONE-001, plan.md M3; hardened in merge-gate
// repair rounds 1–3).
//
// An identity Bash command pairing one of the thirteen mutating forms (REQ-
// SIPZ-007: rm, unlink, mv, cp, tee, truncate, sed -i, > , >>, git rm,
// git checkout, git restore, git apply) with a zone-covered argument or
// redirection target is denied. Segments come from the existing shell splitter
// (the one the commit-identity guard uses); within a segment the words are
// tokenized with a per-character quote mask, a single "&" splits async groups,
// and a cd in a piped segment never moves the main shell.
//
// Under-match and pass on anything unclassifiable — variables, command
// substitution, interpreters, and PowerShell stay outside (spec §C.6) — and
// no environment variable, tool-input field, or command-text token suppresses
// the rule (REQ-SIPZ-013): leading VAR=value assignments are skipped whether
// or not their value was quoted, and the splitter already excludes comment
// text. The quote mask decides exactly two things, because shell syntax lives
// outside quotes: which ">" characters are redirection operators (round 2 P2,
// round 3 P1) and where async boundaries sit. Everything else — the command
// word, options, targets, assignments — reads the stripped text.
//
// Unlike a baseline Write/Edit denial, every shell-mutation denial carries the
// protected-zone sentinel and the routing fields — there is no legacy shell
// reason to keep (spec §C.7).

import (
	"encoding/json"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

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

// zoneWord is one shell word: its text with quotes stripped, plus a mask that
// records, per character of Text, whether it came from inside quotes. The
// mask is read only where shell syntax is decided — redirection operators and
// async boundaries live outside quotes; everything else (the command word,
// options, targets, assignments) is the stripped text.
type zoneWord struct {
	Text string
	Mask []bool
}

// zoneShellWords tokenizes one shell segment into words with their quote
// masks. Substitutions and globs stay as literal words; a word they produce
// normalizes to a path that matches no entry, which is the under-match the
// shell rule accepts.
func zoneShellWords(seg string) []zoneWord {
	var words []zoneWord
	var cur strings.Builder
	var mask []bool
	inSingle, inDouble := false, false
	flush := func() {
		if cur.Len() > 0 {
			words = append(words, zoneWord{Text: cur.String(), Mask: mask})
			cur.Reset()
			mask = nil
		}
	}
	for i := 0; i < len(seg); i++ {
		c := seg[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			} else {
				cur.WriteByte(c)
				mask = append(mask, true)
			}
		case inDouble:
			if c == '\\' && i+1 < len(seg) {
				i++
				cur.WriteByte(seg[i])
				mask = append(mask, true)
			} else if c == '"' {
				inDouble = false
			} else {
				cur.WriteByte(c)
				mask = append(mask, true)
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == ' ' || c == '\t':
			flush()
		default:
			cur.WriteByte(c)
			mask = append(mask, false)
		}
	}
	flush()
	return words
}

// hasUnquoted reports whether the word carries an unquoted occurrence of c.
func hasUnquoted(w zoneWord, c byte) bool {
	for i := range w.Text {
		if w.Text[i] == c && !w.Mask[i] {
			return true
		}
	}
	return false
}

// isZoneAssignment reports whether a word is a leading VAR=value assignment —
// skipped so a command-text prefix cannot suppress the rule (REQ-SIPZ-013).
// The value may have been quoted; that changes nothing (round 3 P1).
func isZoneAssignment(w string) bool {
	eq := strings.Index(w, "=")
	if eq <= 0 {
		return false
	}
	name := w[:eq]
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c == '_', c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

// zonePathCandidates drops flags from an argument list; the rest are the
// candidates the coverage check judges.
func zonePathCandidates(args []zoneWord) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if strings.HasPrefix(a.Text, "-") {
			continue
		}
		out = append(out, a.Text)
	}
	return out
}

// zoneAsyncGroups splits a segment's words at an unquoted "&" — an async
// boundary the shared splitter does not segment on (round 2 P1: "true & rm x"
// is two commands). A "&" inside a word splits it the same way; the shell
// runs "a&b" as "a & b". Quoted "&" characters are data. Each group runs in
// the working directory the MAIN shell has when the group starts.
func zoneAsyncGroups(words []zoneWord) [][]zoneWord {
	var groups [][]zoneWord
	cur := make([]zoneWord, 0, len(words))
	for _, w := range words {
		if !hasUnquoted(w, '&') {
			cur = append(cur, w)
			continue
		}
		// split the word at its unquoted "&" characters
		frag := zoneWord{Text: "", Mask: nil}
		for i := range w.Text {
			if w.Text[i] == '&' && !w.Mask[i] {
				if frag.Text != "" {
					cur = append(cur, frag)
					groups = append(groups, cur)
					cur = make([]zoneWord, 0, len(words))
				} else if len(cur) > 0 {
					groups = append(groups, cur)
					cur = make([]zoneWord, 0, len(words))
				}
				frag = zoneWord{Text: "", Mask: nil}
				continue
			}
			frag.Text += w.Text[i : i+1]
			frag.Mask = append(frag.Mask, w.Mask[i])
		}
		if frag.Text != "" {
			cur = append(cur, frag)
		}
	}
	groups = append(groups, cur)
	return groups
}

// zoneRedirectTargets walks a segment's words and returns every redirection
// target. Only an UNQUOTED ">" is an operator (round 2 P2: a quoted ">" is
// string data; round 3 P1: a partially quoted word like `>".moai/logs/x"`
// still carries the operator) — every fragment after the first unquoted ">"
// of a word is a target, and a word ending at an unquoted ">" hands its
// target to the next word, quoted or not.
func zoneRedirectTargets(words []zoneWord) (bool, []string) {
	mutating := false
	var targets []string
	nextIsTarget := false
	for _, w := range words {
		frag := ""
		sawOp := false
		for i := range w.Text {
			if w.Text[i] == '>' && !w.Mask[i] {
				if !sawOp {
					// characters before the first operator are the command's
					// own text, not a target
					sawOp = true
				}
				frag = ""
				continue
			}
			frag += w.Text[i : i+1]
		}
		if sawOp {
			mutating = true
			if frag != "" {
				targets = append(targets, frag)
				nextIsTarget = false
			} else {
				nextIsTarget = true
			}
			continue
		}
		if nextIsTarget {
			mutating = true
			targets = append(targets, w.Text)
			nextIsTarget = false
		}
	}
	return mutating, targets
}

// zoneStripRedirects returns the words left after every unquoted redirection
// operator and its target are removed — the arguments a cd actually takes
// (round 3 P1: "cd .moai > docs/out" is a cd to .moai whose redirect writes
// docs/out, not a multi-argument cd).
func zoneStripRedirects(words []zoneWord) []zoneWord {
	var out []zoneWord
	skipTarget := false
	for _, w := range words {
		if skipTarget {
			skipTarget = false
			continue
		}
		if !w.QuotedAny() && strings.Contains(w.Text, ">") {
			head := w.Text[:strings.Index(w.Text, ">")]
			if head != "" {
				out = append(out, zoneWord{Text: head})
			}
			skipTarget = strings.HasSuffix(w.Text, ">")
			continue
		}
		out = append(out, w)
	}
	return out
}

// QuotedAny reports whether any character of the word came from inside
// quotes — the word cannot carry shell operators.
func (w zoneWord) QuotedAny() bool {
	for _, q := range w.Mask {
		if q {
			return true
		}
	}
	return false
}

// zoneMutatingWords classifies one tokenized command: whether it pairs one of
// the thirteen mutating forms or a redirection, and the path-like candidates
// paired with them.
//
// sed is in-place when any of its arguments carries an option cluster with
// "i" or the long form — combined "-Ei", quoted "'-i'", and plain "-i" all
// count (round 3 P1). The git subcommand is looked up past git's global
// options, and a "-C <dir>" option moves the directory the file arguments
// resolve against (round 3 P1).
func zoneMutatingWords(words []zoneWord) (bool, []string) {
	i := 0
	for i < len(words) && isZoneAssignment(words[i].Text) {
		i++
	}
	if i >= len(words) {
		return false, nil
	}
	rest := words[i:]
	mutating := false
	var cands []string
	switch first := rest[0].Text; {
	case zoneMutationVerbs[first]:
		mutating = true
		cands = append(cands, zonePathCandidates(rest[1:])...)
	case first == "sed":
		for _, w := range rest[1:] {
			t := w.Text
			if t == "--in-place" || (strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "--") && strings.Contains(strings.TrimPrefix(t, "-"), "i")) {
				mutating = true
			}
		}
		if mutating {
			cands = append(cands, zonePathCandidates(rest[1:])...)
		}
	case first == "git":
		gitDir := ""
		for j := 1; j < len(rest); j++ {
			w := rest[j]
			if !w.QuotedAny() && w.Text == "-C" && j+1 < len(rest) {
				gitDir = rest[j+1].Text
				continue
			}
			if zoneGitMutating[w.Text] {
				mutating = true
				cands = append(cands, zoneRelativeTo(gitDir, zonePathCandidates(rest[j+1:]))...)
				break
			}
		}
	}
	if hits, targets := zoneRedirectTargets(rest); hits {
		mutating = true
		cands = append(cands, targets...)
	}
	return mutating, cands
}

// zoneNextCwd returns the segment working directory after a cd. Only a plain
// literal relative argument is tracked — a bare cd, several arguments,
// substitution, a glob, an option, or an absolute path leaves it untracked
// (reset to the root), which is the documented under-match (merge-gate round 1
// P1-3). The dots are cleaned here because the shell's cd already resolved the
// directory: this is the logical cwd the next segment's relative names
// concatenate onto, never a pre-clean of a target path.
func zoneNextCwd(cur string, args []zoneWord) string {
	if len(args) != 1 {
		return "."
	}
	arg := args[0].Text
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

// checkProtectedZoneShell decides one identity Bash call against the zone and
// returns a deny reason, or "" to let the call through. A command with no
// mutating form never reaches the manifest (REQ-SIPZ-008); a mutating command
// under an invalid manifest is denied unread (REQ-SIPZ-009).
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

	// Segment pass. A plain literal `cd` moves the working directory the later
	// segments' relative names resolve against (round 1 P1-3) — except inside
	// a pipeline, where every element (the cd included) runs in a subshell and
	// the main shell's directory never moves (round 3 P1). The cd segment's
	// own redirection targets are judged against the directory the shell
	// evaluates them in, before the cd takes effect, and its arguments are the
	// words left once the redirection is stripped (round 3 P1). Async groups
	// split on "&" (round 2 P1); a boundary restores the directory the main
	// shell had before the async group — the group's own cd ran in a subshell,
	// and the main shell never moved (round 3 P1). No disk access yet beyond
	// the glob-parent walks — the cost seam closes before the first manifest
	// read.
	var cands []string
	mutating := false
	cur := "."
	segs := splitShellSegments(command)
	for i, seg := range segs {
		inPipeline := i+1 < len(segs) && segs[i+1].conn == "|"
		groups := zoneAsyncGroups(zoneShellWords(seg.text))
		for gi, group := range groups {
			if len(group) == 0 {
				continue
			}
			pre := cur
			if !inPipeline && group[0].Text == "cd" {
				if hits, targets := zoneRedirectTargets(group); hits {
					mutating = true
					cands = append(cands, zoneRelativeTo(cur, targets)...)
				}
				cur = zoneNextCwd(cur, zoneStripRedirects(group[1:]))
			} else if hit, args := zoneMutatingWords(group); hit {
				mutating = true
				cands = append(cands, zoneRelativeTo(cur, args)...)
			}
			if gi < len(groups)-1 {
				// async boundary: the group's own directory changes ran in a
				// subshell — the main shell keeps the directory it had before
				// the group (round 3 P1)
				cur = pre
			}
		}
	}
	if !mutating {
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

	for _, cand := range cands {
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
		if len(cands) > 0 {
			target = cands[0]
		}
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: "Bash", Path: target,
			Decision: "allow", ManifestState: config.ZoneStateAbsent,
		})
	}
	return ""
}

// zoneRelativeTo prefixes a tracked working directory (or git's -C directory)
// onto relative candidates, keeping the raw segments for the normalization and
// symlink layers; absolute candidates are judged as they land. An empty dir
// leaves the candidates unchanged.
func zoneRelativeTo(dir string, cands []string) []string {
	out := make([]string, 0, len(cands))
	for _, cand := range cands {
		switch {
		case dir == "" || dir == "." || zoneIsAbs(cand):
			out = append(out, cand)
		case zoneIsAbs(dir):
			out = append(out, dir+"/"+cand)
		default:
			out = append(out, dir+"/"+cand)
		}
	}
	return out
}
