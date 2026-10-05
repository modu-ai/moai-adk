package hook

// protected_zone_shell.go — the shell half of the protected-zone guard
// (SPEC-SELF-IMPROVE-PROTECTED-ZONE-001, plan.md M3).
//
// An identity Bash command pairing one of the thirteen mutating forms (REQ-
// SIPZ-007: rm, unlink, mv, cp, tee, truncate, sed -i, > , >>, git rm,
// git checkout, git restore, git apply) with a zone-covered argument or
// redirection target is denied. Segments come from the existing shell splitter
// (the one the commit-identity guard uses); within a segment the words are
// tokenized with quote awareness. Under-match and pass on anything
// unclassifiable — variables, command substitution, interpreters, and PowerShell
// stay outside (spec §C.6) — and no environment variable, tool-input field, or
// command-text token suppresses the rule (REQ-SIPZ-013): leading VAR=value
// assignments are skipped, and the splitter already excludes comment text.
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
// zoneMutationVerbs are the single-word mutating forms of REQ-SIPZ-007.
var zoneMutationVerbs = map[string]bool{
	"rm": true, "unlink": true, "mv": true, "cp": true, "tee": true, "truncate": true,
}

// zoneGitMutating are the git subcommands of REQ-SIPZ-007.
var zoneGitMutating = map[string]bool{
	"rm": true, "checkout": true, "restore": true, "apply": true,
}

// zoneShellWords tokenizes one shell segment into words, honoring single and
// double quotes. Substitutions and globs stay as literal words; a word they
// produce normalizes to a path that matches no entry, which is the under-match
// the shell rule accepts.
func zoneShellWords(seg string) []string {
	var words []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	flush := func() {
		if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
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
		case c == '"':
			inDouble = true
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
func zonePathCandidates(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		out = append(out, a)
	}
	return out
}

// zoneMutatingWords classifies one tokenized segment: whether it pairs one of
// the thirteen mutating forms or a redirection, and the path-like candidates
// paired with them. Redirections are collected wherever they sit and however
// they are spaced: every substring between ">" characters of a word is a
// target ("x>file", "2>file", "a>b>c" all contribute), and a word ending in
// ">" hands its target to the next word (merge-gate round 1 P1-2). The
// thirteen forms keep their verb pairing; a verb's segment still contributes
// its redirection targets alongside the verb's arguments.
func zoneMutatingWords(words []string) (bool, []string) {
	i := 0
	for i < len(words) && isZoneAssignment(words[i]) {
		i++
	}
	if i >= len(words) {
		return false, nil
	}
	rest := words[i:]
	mutating := false
	var cands []string
	switch {
	case zoneMutationVerbs[rest[0]]:
		mutating = true
		cands = append(cands, zonePathCandidates(rest[1:])...)
	case rest[0] == "sed" && len(rest) > 1 && strings.HasPrefix(rest[1], "-i"):
		mutating = true
		cands = append(cands, zonePathCandidates(rest[2:])...)
	case rest[0] == "git" && len(rest) > 1 && zoneGitMutating[rest[1]]:
		mutating = true
		cands = append(cands, zonePathCandidates(rest[2:])...)
	}
	nextIsTarget := false
	for _, w := range rest {
		parts := strings.Split(w, ">")
		if len(parts) > 1 {
			mutating = true
			for _, t := range parts[1:] {
				if t != "" {
					cands = append(cands, t)
				}
			}
			nextIsTarget = strings.HasSuffix(w, ">")
			continue
		}
		if nextIsTarget {
			mutating = true
			cands = append(cands, w)
			nextIsTarget = false
		}
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
func zoneNextCwd(cur string, args []string) string {
	if len(args) != 1 {
		return "."
	}
	arg := args[0]
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
// zone: the compiled baseline floor first, then the manifest entries. A bare
// directory operand names the directory itself, so every form is also tried
// with a trailing slash — without it, "rm -rf .claude/hooks" would walk past
// the ".claude/hooks/" entry (merge-gate round 1 P1-1). The returned category
// names the source that covered the candidate.
func zoneShellCovered(forms []zoneForm, load config.ProtectedZoneLoad) (string, bool) {
	for _, form := range forms {
		for _, spelling := range []string{form.Display, form.Display + "/", form.Folded, form.Folded + "/"} {
			if zoneBaselineCovers(spelling) {
				return zoneBaselineCategory, true
			}
		}
	}
	if load.State != config.ZoneStateOK {
		return "", false
	}
	for _, form := range forms {
		for _, folded := range []string{form.Folded, form.Folded + "/"} {
			for i := range load.Zone.Entries {
				if load.Zone.Entries[i].Match(folded) {
					return load.Zone.Entries[i].Category, true
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

	// Segment pass: a plain literal `cd` moves the working directory the later
	// segments' relative names resolve against (merge-gate round 1 P1-3); a
	// mutating form collects its verb arguments and every redirection target.
	// No disk access yet — the cost seam closes before the first manifest read.
	var cands []string
	mutating := false
	cur := "."
	for _, seg := range splitShellSegments(command) {
		words := zoneShellWords(seg.text)
		if len(words) > 0 && words[0] == "cd" {
			cur = zoneNextCwd(cur, words[1:])
			continue
		}
		hit, args := zoneMutatingWords(words)
		if !hit {
			continue
		}
		mutating = true
		for _, a := range args {
			if cur != "." && !zoneIsAbs(a) {
				cands = append(cands, cur+"/"+a)
			} else {
				cands = append(cands, a)
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
