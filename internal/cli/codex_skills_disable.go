package cli

// codex_skills_disable.go — the hand that DISABLES one skill in the
// user-layer ~/.codex/config.toml (SPEC-CODEX-SKILL-DISABLE-001).
//
// Claude Code and Codex see the same skill set. This verb turns one of them
// off on the Codex side only, by publishing a single [[skills.config]] entry
// carrying enabled = false.
//
// Two measured facts fix the shape of what is published, and neither is
// negotiable from inside this file:
//
//   - `enabled` is a REQUIRED field of the entry. An entry lacking it is not
//     a degraded feature — it is a hard start failure for the user's whole
//     codex (`missing field 'enabled' in skills.config`, rc=1). Emission is
//     therefore narrowed to one function whose output cannot omit the key.
//   - The gate compares REALPATH-normalised files, not path strings, and the
//     only notation that binds under BOTH mirror shapes (symlink, and the
//     copy fallback) is the absolute literal mirror path
//     <projectRoot>/.agents/skills/<skill>/SKILL.md. The resolved `.claude/…`
//     twin binds under a symlink mirror and is SILENTLY inert under a copy
//     one; directory-shaped and relative notations are inert under both. All
//     three failures are silent — the user believes the skill is off and it
//     is still exposed.
//
// Both facts come from .moai/reports/t502/gate-path-shape.md (19 cells,
// codex-cli 0.153.4) and .moai/reports/t504/skills-config-path-shape.md,
// re-measured at .moai/reports/t502/regate-0.153.4.md before this code landed.
//
// The parser is consumed, never modified: internal/codexwiring/skills.go stays
// read-only by construction. The twin write path
// (internal/cli/codex_skills_prune.go) is adjacent but not shared — the two
// verbs judge in opposite directions, prune proving absence before it deletes
// and this one proving existence before it writes.

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// mirrorSkillsRel is the skill-mirror root, relative to a project root. It is
// the path Codex scans (`<cwd>/.agents/skills`), which is also what makes the
// gate cwd-bound — see the help text this verb prints.
var mirrorSkillsRel = filepath.Join(".agents", "skills")

// codexSkillWriteFileFn is the file-write seam. The backup and the staged
// write both go through it so a test can fail the backup alone and observe
// that the target stayed byte-invariant.
var codexSkillWriteFileFn = os.WriteFile

// codexSkillDisableAction is what the merge did, or declined to do.
type codexSkillDisableAction int

const (
	// codexSkillDisableAppended: no entry declared the path; one was added.
	codexSkillDisableAppended codexSkillDisableAction = iota
	// codexSkillDisableUpdated: an entry declared the path; its enabled value
	// became false.
	codexSkillDisableUpdated
	// codexSkillDisableUnchanged: the entry already reads false. The INPUT
	// bytes are returned, never a re-joined copy — a round-trip defect must
	// not be able to masquerade as a no-op run.
	codexSkillDisableUnchanged
	// codexSkillDisableSkipped: the entry was left alone on purpose.
	codexSkillDisableSkipped
)

// codexSkillDisableVerdict is the merge's disposition. Reason is non-empty
// exactly when the action is Unchanged or Skipped: an entry silently left
// behind teaches the user nothing about what to fix.
type codexSkillDisableVerdict struct {
	Action codexSkillDisableAction
	Path   string
	Reason string
}

// codexSkillResolveOutcome is how a skill NAME resolved to a file.
type codexSkillResolveOutcome int

const (
	// codexSkillResolved: exactly one candidate root carries the skill, and
	// it is the project mirror.
	codexSkillResolved codexSkillResolveOutcome = iota
	// codexSkillMirrorAbsent: the project has no skill mirror at all. The
	// mirror is produced by a deployment run, not by a checkout, so its
	// absence is an ORDINARY state and exits 0.
	codexSkillMirrorAbsent
	// codexSkillUnresolved: no candidate root carries the name — typically a
	// typo, which must be detectable from a script, so it exits non-zero.
	codexSkillUnresolved
	// codexSkillAmbiguous: more than one candidate root carries the name.
	codexSkillAmbiguous
)

// codexSkillResolution is one name-resolution outcome. Path is set only when
// the outcome is codexSkillResolved.
type codexSkillResolution struct {
	Outcome codexSkillResolveOutcome
	Path    string
	Reason  string
}

// codexSkillDisableOptions are the runner's inputs. The roots are parameters
// rather than ambient state so a test can point them at a temp tree without
// touching the process working directory or the real home.
type codexSkillDisableOptions struct {
	Skill       string
	ProjectRoot string
	HomeDir     string
	Force       bool
}

// resolveCodexSkillMirrorPath turns a skill NAME into the absolute literal
// mirror path to publish.
//
// Candidate roots are named by what they ARE, never by the r0/r1/r2 labels
// Codex prints: those are relative labels that get renumbered per
// environment, and two prior measurements already observed different
// orderings. Two roots participate:
//
//   - the project mirror <projectRoot>/.agents/skills — the publication target
//   - the user home mirror ~/.agents/skills — which reaches Codex's root list
//     even under CODEX_HOME isolation (measured), and is therefore the real
//     source of ambiguity rather than a hypothetical one
//
// A name carried ONLY by the home mirror is rejected rather than published:
// this verb publishes the project mirror path, and an entry naming a file the
// project does not carry is a dead registration from birth — the exact debt
// class `moai clean --codex-skills` exists to collect.
func resolveCodexSkillMirrorPath(projectRoot, homeDir, skill string) codexSkillResolution {
	if skill == "" || strings.ContainsAny(skill, `/\`) || skill == "." || skill == ".." {
		return codexSkillResolution{
			Outcome: codexSkillUnresolved,
			Reason:  fmt.Sprintf("%q is not a bare skill name — pass the directory name as it appears under %s", skill, mirrorSkillsRel),
		}
	}

	projMirror := filepath.Join(projectRoot, mirrorSkillsRel)
	if st, err := osStatFn(projMirror); err != nil || !st.IsDir() {
		return codexSkillResolution{
			Outcome: codexSkillMirrorAbsent,
			Reason:  fmt.Sprintf("the project skill mirror %s does not exist, so nothing reaches Codex through it (it is created by a deployment run, not by a checkout)", projMirror),
		}
	}

	type candidate struct{ root, file string }
	cands := []candidate{{projMirror, filepath.Join(projMirror, skill, "SKILL.md")}}
	if homeDir != "" {
		if homeMirror := filepath.Join(homeDir, mirrorSkillsRel); homeMirror != projMirror {
			cands = append(cands, candidate{homeMirror, filepath.Join(homeMirror, skill, "SKILL.md")})
		}
	}

	var found []candidate
	for _, c := range cands {
		// A directory-shaped entry does not gate (cells E4/E5), so a
		// SKILL.md that is not a regular file is not a resolution.
		if st, err := osStatFn(c.file); err == nil && st.Mode().IsRegular() {
			found = append(found, c)
		}
	}

	switch len(found) {
	case 0:
		return codexSkillResolution{
			Outcome: codexSkillUnresolved,
			Reason:  fmt.Sprintf("no SKILL.md for %q under %s — check the name, or run a deployment to refresh the mirror", skill, projMirror),
		}
	case 1:
		if found[0].root != projMirror {
			return codexSkillResolution{
				Outcome: codexSkillUnresolved,
				Reason: fmt.Sprintf("%q exists only under the user home mirror %s, not in this project; this verb publishes the project mirror path, and an entry naming a file this project does not carry would be dead on arrival",
					skill, found[0].root),
			}
		}
		return codexSkillResolution{Outcome: codexSkillResolved, Path: found[0].file}
	default:
		roots := make([]string, 0, len(found))
		for _, c := range found {
			roots = append(roots, c.root)
		}
		return codexSkillResolution{
			Outcome: codexSkillAmbiguous,
			Reason: fmt.Sprintf("%q resolves in more than one skill root (%s); the gate binds one file, so remove or rename the duplicate before disabling",
				skill, strings.Join(roots, ", ")),
		}
	}
}

// upsertCodexSkillDisable is the pure merge: content in, content out, plus a
// verdict. It is the ONLY place an entry is composed, which is what makes the
// `enabled` key impossible to omit.
func upsertCodexSkillDisable(content []byte, skillPath string) ([]byte, codexSkillDisableVerdict) {
	skip := func(format string, a ...any) ([]byte, codexSkillDisableVerdict) {
		return content, codexSkillDisableVerdict{
			Action: codexSkillDisableSkipped, Path: skillPath, Reason: fmt.Sprintf(format, a...),
		}
	}
	if skillPath == "" {
		return skip("no path to write")
	}
	// A TOML basic string cannot carry these verbatim, and this repository's
	// parser reads the value verbatim rather than decoding escapes. Emitting
	// an escaped form would produce an entry Codex reads correctly and every
	// moai reader — including the prune verb, which deletes what it cannot
	// find — reads as a different, absent path.
	if strings.ContainsAny(skillPath, "\"\\\n\r") {
		return skip("the path contains a character this config format cannot carry verbatim (%q)", skillPath)
	}

	lines, term := codexwiring.SplitConfigLines(content)
	var matches []codexwiring.SkillEntry
	for _, e := range codexwiring.ParseSkillEntries(content) {
		if e.Path == skillPath {
			matches = append(matches, e)
		}
	}

	if len(matches) > 1 {
		return skip("%d entries declare this path; collapsing duplicates is `moai clean --codex-skills`'s job, not this verb's", len(matches))
	}

	if len(matches) == 1 {
		e := matches[0]
		// The parser's judgment, consumed as reported. Re-reading the range's
		// text here would resurrect the defect the field exists to close: a
		// comment carrying an odd number of `"""` opens a literal that can
		// swallow a whole healthy registration, and every swallowed line
		// still LOOKS recognisable.
		if e.FirstUnrecognizedLine >= 0 {
			return skip("line %d in its range was not recognised, so the entry is left untouched", e.FirstUnrecognizedLine+1)
		}
		if e.Enabled == codexwiring.SkillEnabledFalse {
			return content, codexSkillDisableVerdict{
				Action: codexSkillDisableUnchanged, Path: skillPath, Reason: "it is already disabled",
			}
		}
		updated, ok := setEnabledFalseInExtent(append([]string(nil), lines...), e)
		if !ok {
			return skip("the entry declares no key this verb can rewrite")
		}
		return codexwiring.JoinConfigLines(updated, term), codexSkillDisableVerdict{
			Action: codexSkillDisableUpdated, Path: skillPath,
		}
	}

	return codexwiring.JoinConfigLines(appendDisableEntry(lines, skillPath), term),
		codexSkillDisableVerdict{Action: codexSkillDisableAppended, Path: skillPath}
}

// setEnabledFalseInExtent rewrites the entry's enabled line in place, or
// inserts one right after its path assignment when the key is absent,
// preserving that line's own indentation and CR. It returns the new line
// slice and whether it found somewhere to put the key.
//
// Only entries whose every line the parser recognised reach here, so a line
// beginning with "enabled" inside the extent IS the enabled assignment.
//
// An insert shifts every later index by one. That is safe here and only here:
// this function is the last consumer of the parser's indices, and the result
// goes straight to the join.
func setEnabledFalseInExtent(lines []string, e codexwiring.SkillEntry) ([]string, bool) {
	pathLine := -1
	for i := e.StartLine; i < e.EndLine && i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "enabled") {
			lines[i] = reshapeLike(lines[i], "enabled = false")
			return lines, true
		}
		if strings.HasPrefix(trimmed, "path") {
			pathLine = i
		}
	}
	if pathLine < 0 {
		return lines, false
	}
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:pathLine+1]...)
	out = append(out, reshapeLike(lines[pathLine], "enabled = false"))
	return append(out, lines[pathLine+1:]...), true
}

// reshapeLike renders body with the same leading whitespace and trailing CR
// as model, so a hand-indented CRLF config keeps its shape.
func reshapeLike(model, body string) string {
	cr := ""
	if strings.HasSuffix(model, "\r") {
		cr = "\r"
		model = strings.TrimSuffix(model, "\r")
	}
	indent := model[:len(model)-len(strings.TrimLeft(model, " \t"))]
	return indent + body + cr
}

// appendDisableEntry appends one complete entry — header, path, enabled —
// using the file's own line ending. The three lines are emitted together and
// nowhere else, so no code path can produce an entry missing `enabled`.
func appendDisableEntry(lines []string, skillPath string) []string {
	cr := ""
	for _, l := range lines {
		if strings.HasSuffix(l, "\r") {
			cr = "\r"
			break
		}
	}
	out := append([]string(nil), lines...)
	if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
		out = append(out, cr)
	}
	return append(out,
		"[[skills.config]]"+cr,
		`path = "`+skillPath+`"`+cr,
		"enabled = false"+cr,
	)
}

// runCodexSkillDisable is the `moai skills disable <name> --codex` runner.
//
// Fail-open on absent inputs, exactly as the prune verb is: an unresolvable
// Codex home or an absent config says so and returns nil. A missing input is
// not an error — there is simply nothing to disable. A name that does not
// resolve is different: that is a typo, and it exits non-zero so a script can
// see it.
func runCodexSkillDisable(p printer.Printer, opts codexSkillDisableOptions) error {
	res := resolveCodexSkillMirrorPath(opts.ProjectRoot, opts.HomeDir, opts.Skill)
	switch res.Outcome {
	case codexSkillMirrorAbsent:
		p.Info("Nothing to disable: %s", res.Reason)
		return nil
	case codexSkillUnresolved, codexSkillAmbiguous:
		p.Info("Refusing to write: %s", res.Reason)
		return fmt.Errorf("cannot disable %q for Codex", opts.Skill)
	}

	codexHome, _ := resolveCodexHomeDir()
	if codexHome == "" {
		p.Info("Codex home does not resolve; nothing to disable")
		return nil
	}
	cfgPath := filepath.Join(codexHome, path.Base(codexwiring.ConfigRelPath))
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		p.Info("%s is absent or unreadable; nothing to disable", cfgPath)
		return nil
	}

	merged, v := upsertCodexSkillDisable(raw, res.Path)
	switch v.Action {
	case codexSkillDisableUnchanged:
		p.Info("Unchanged: %s (%s)", res.Path, v.Reason)
		return nil
	case codexSkillDisableSkipped:
		p.Info("Skipped: %s (%s)", res.Path, v.Reason)
		return nil
	}

	verb := "add an entry for"
	if v.Action == codexSkillDisableUpdated {
		verb = "set enabled = false on the existing entry for"
	}
	if !opts.Force {
		p.Info("[dry-run] Would %s %s in %s", verb, res.Path, cfgPath)
		p.Info("Run again with --force to actually write.")
		return nil
	}

	// The backup lands BEFORE the write. This is a user-authored file in the
	// user's home; a write that outruns its backup is unrecoverable.
	backupPath := fmt.Sprintf("%s.bak-%s", cfgPath, time.Now().UTC().Format("20060102T150405Z"))
	if err := codexSkillWriteFileFn(backupPath, raw, 0o600); err != nil {
		return fmt.Errorf("back up %s: %w", cfgPath, err)
	}
	p.Info("Backup: %s", backupPath)
	p.Info("Backup sha256: %s", sha256Hex(raw))

	mode := fs.FileMode(0o600)
	if st, statErr := os.Stat(cfgPath); statErr == nil {
		mode = st.Mode().Perm()
	}
	if err := writeCodexConfigPreservingMode(cfgPath, merged, mode); err != nil {
		return fmt.Errorf("write %s: %w", cfgPath, err)
	}
	p.Info("Disabled %s for Codex: %s (%s)", opts.Skill, res.Path, cfgPath)
	return nil
}

// writeCodexConfigPreservingMode replaces the config through a staged file so
// a crash mid-write cannot leave a truncated config in the user's home, and
// carries the ORIGINAL permission mode onto the replacement.
//
// The mode carry is the whole point of staging: a verb whose headline
// property is "non-destructive" must not quietly tighten a user's 0644 config
// to 0600 as a side effect of rewriting it. (The prune verb writes in place
// and so never faces this; that difference is deliberate and belongs to that
// verb's own card.)
func writeCodexConfigPreservingMode(cfgPath string, data []byte, mode fs.FileMode) error {
	staged := cfgPath + ".moai-staged"
	if err := codexSkillWriteFileFn(staged, data, mode); err != nil {
		return err
	}
	if err := os.Chmod(staged, mode); err != nil {
		_ = os.Remove(staged)
		return err
	}
	if err := os.Rename(staged, cfgPath); err != nil {
		_ = os.Remove(staged)
		return err
	}
	return nil
}
