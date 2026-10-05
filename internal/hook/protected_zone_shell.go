package hook

// protected_zone_shell.go — the shell half of the protected-zone guard
// (SPEC-SELF-IMPROVE-PROTECTED-ZONE-001, plan.md M3; hardened in merge-gate
// repair rounds 1 and 2).
//
// An identity Bash command pairing one of the thirteen mutating forms (REQ-
// SIPZ-007: rm, unlink, mv, cp, tee, truncate, sed -i, > , >>, git rm,
// git checkout, git restore, git apply) with a zone-covered argument or
// redirection target is denied. Segments come from the existing shell splitter
// (the one the commit-identity guard uses); within a segment the words are
// tokenized with quote awareness, and a single "&" — which the splitter does
// not treat as a separator — splits the segment into async groups, each
// judged in the working directory it inherits.
//
// Under-match and pass on anything unclassifiable — variables, command
// substitution, interpreters, and PowerShell stay outside (spec §C.6) — and
// no environment variable, tool-input field, or command-text token suppresses
// the rule (REQ-SIPZ-013): leading VAR=value assignments are skipped, and the
// splitter already excludes comment text. Quoted text is data: a ">" inside
// quotes is never a redirection operator, while a quoted word after a bare
// ">" is still the target.
//
// Unlike a baseline Write/Edit denial, every shell-mutation denial carries the
// protected-zone sentinel and the routing fields — there is no legacy shell
// reason to keep (spec §C.7).

import (
	"encoding/json"
	"path"
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

// zoneWord is one shell word: its text with quotes stripped, and whether any
// part of it came from inside quotes. Quoting decides what the word can be —
// a quoted word is data for the operator scan (a quoted ">" is string text,
// round 2 P2) while still being usable as a redirection target.
type zoneWord struct {
	Text   string
	Quoted bool
}

// zoneShellWords tokenizes one shell segment into words, honoring single and
// double quotes and recording per word whether quoting contributed to it.
// Substitutions and globs stay as literal words; a word they produce
// normalizes to a path that matches no entry, which is the under-match the
// shell rule accepts.
func zoneShellWords(seg string) []zoneWord {
	var words []zoneWord
	var cur strings.Builder
	curQuoted := false
	inSingle, inDouble := false, false
	flush := func() {
		if cur.Len() > 0 {
			words = append(words, zoneWord{Text: cur.String(), Quoted: curQuoted})
			cur.Reset()
			curQuoted = false
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
			}
		case inDouble:
			if c == '\\' && i+1 < len(seg) {
				i++
				cur.WriteByte(seg[i])
			} else if c == '"' {
				inDouble = false
			} else {
				cur.WriteByte(c)
			}
		case c == '\'':
			inSingle = true
			curQuoted = true
		case c == '"':
			inDouble = true
			curQuoted = true
		case c == ' ' || c == '\t':
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return words
}

// isZoneAssignment reports whether a word is a leading VAR=value assignment —
// skipped so a command-text prefix cannot suppress the rule (REQ-SIPZ-013).
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

// zoneAsyncGroups splits a segment's words at single "&" — an async boundary
// the shared splitter does not segment on (round 2 P1: "true & rm x" is two
// commands). A "&" inside a word splits it the same way; the shell runs "a&b"
// as "a & b". Each group runs in the working directory the main shell has at
// that point — a cd inside an async group does not move the later groups.
func zoneAsyncGroups(words []zoneWord) [][]zoneWord {
	var groups [][]zoneWord
	cur := make([]zoneWord, 0, len(words))
	for _, w := range words {
		if !w.Quoted && strings.Contains(w.Text, "&") {
			fragments := strings.Split(w.Text, "&")
			if fragments[0] != "" {
				cur = append(cur, zoneWord{Text: fragments[0]})
			}
			if len(cur) > 0 {
				groups = append(groups, cur)
				cur = make([]zoneWord, 0, len(words))
			}
			for _, f := range fragments[1 : len(fragments)-1] {
				if f != "" {
					groups = append(groups, []zoneWord{{Text: f}})
				}
			}
			if last := fragments[len(fragments)-1]; last != "" {
				cur = append(cur, zoneWord{Text: last})
			}
			continue
		}
		cur = append(cur, w)
	}
	groups = append(groups, cur)
	return groups
}

// zoneRedirectTargets walks a segment's words and returns every redirection
// target: the word after a bare ">" or ">>", the remainder of a word that
// starts with one, and every substring between ">" characters of a word that
// carries one inline ("x>file", "2>file", "a>b>c" all contribute). Only
// unquoted words carry operators — a quoted ">" is string data (round 2 P2) —
// while a quoted word after a bare ">" is still the target.
func zoneRedirectTargets(words []zoneWord) (bool, []string) {
	mutating := false
	var targets []string
	nextIsTarget := false
	for _, w := range words {
		if w.Quoted {
			if nextIsTarget {
				mutating = true
				targets = append(targets, w.Text)
				nextIsTarget = false
			}
			continue
		}
		if strings.Contains(w.Text, ">") {
			mutating = true
			for _, t := range strings.Split(w.Text, ">")[1:] {
				if t != "" {
					targets = append(targets, t)
				}
			}
			nextIsTarget = strings.HasSuffix(w.Text, ">")
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

// zoneMutatingWords classifies one tokenized command: whether it pairs one of
// the thirteen mutating forms or a redirection, and the path-like candidates
// paired with them.
//
// sed is in-place when any of its arguments carries "-i", wherever the options
// sit (round 2 P1: "sed -e s/a/b/ -i ” f"). The git subcommand is looked up
// past git's global options (round 2 P1: "git -c k=v checkout -- f").
func zoneMutatingWords(words []zoneWord) (bool, []string) {
	i := 0
	for i < len(words) && !words[i].Quoted && isZoneAssignment(words[i].Text) {
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
			if !w.Quoted && strings.HasPrefix(w.Text, "-i") {
				mutating = true
			}
		}
		if mutating {
			cands = append(cands, zonePathCandidates(rest[1:])...)
		}
	case first == "git":
		for j := 1; j < len(rest); j++ {
			if !rest[j].Quoted && zoneGitMutating[rest[j].Text] {
				mutating = true
				cands = append(cands, zonePathCandidates(rest[j+1:])...)
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

// zoneShellCovered judges the normalized forms of one candidate against the
// zone, in two passes. Direct: the candidate — tried with and without a
// trailing slash, so a bare directory operand names the directory itself
// (round 1 P1-1) — is a zone member. Ancestor: the candidate is a parent
// directory of protected entries, so moving or removing it would remove them
// (round 2 P1: "mv .moai moved" removed both manifests while every individual
// entry still matched); the project root itself is the parent of everything.
// The returned category names the source that covered the candidate.
func zoneShellCovered(forms []zoneForm, load config.ProtectedZoneLoad) (string, bool) {
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
					continue // a basename glob carries no path prefix to contain
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
	// segments' relative names resolve against (round 1 P1-3) — but the cd
	// segment's own redirection targets are judged against the directory the
	// shell evaluates them in, the one before the cd takes effect (round 2
	// P1). Async groups split on "&" (round 2 P1) and each inherits the main
	// shell's directory, so the tracking resets at every boundary. No disk
	// access yet — the cost seam closes before the first manifest read.
	var cands []string
	mutating := false
	cur := "."
	for _, seg := range splitShellSegments(command) {
		groups := zoneAsyncGroups(zoneShellWords(seg.text))
		for gi, group := range groups {
			if len(group) == 0 {
				continue
			}
			if group[0].Text == "cd" {
				if hits, targets := zoneRedirectTargets(group); hits {
					mutating = true
					cands = append(cands, zoneRelativeTo(cur, targets)...)
				}
				cur = zoneNextCwd(cur, group[1:])
			} else if hit, args := zoneMutatingWords(group); hit {
				mutating = true
				cands = append(cands, zoneRelativeTo(cur, args)...)
			}
			if gi < len(groups)-1 {
				// async boundary: the group before it ran in a subshell, so
				// the main shell's working directory is unchanged from here
				cur = "."
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
		category, covered := zoneShellCovered(forms, load)
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

// zoneRelativeTo prefixes a tracked working directory onto relative
// candidates, keeping the raw segments for the normalization and symlink
// layers; absolute candidates are judged as they land.
func zoneRelativeTo(cur string, cands []string) []string {
	out := make([]string, 0, len(cands))
	for _, cand := range cands {
		if cur != "." && !zoneIsAbs(cand) {
			out = append(out, cur+"/"+cand)
		} else {
			out = append(out, cand)
		}
	}
	return out
}
