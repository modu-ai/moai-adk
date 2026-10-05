package cli

// codex_skills_prune.go — the hand that removes ghost [[skills.config]]
// registrations from the user-layer ~/.codex/config.toml
// (SPEC-CODEX-GHOST-SKILLS-PRUNE-001).
//
// Codex records every registered skill as an array-of-tables entry. When the
// file that entry points at is deleted, Codex neither prunes the entry nor
// complains, so the registration lives on as a ghost. The doctor already
// COUNTS them (codexStaleSkillFinding, which reads only); this is the verb
// that removes them.
//
// It removes only what it can PROVE is absent. Seven classes are never
// removed — a relative path, an oddly-formed one, a home-relative one whose
// home will not resolve, an entry with no path key, a path that resolves (a
// directory resolves too), a stat failure that is not fs.ErrNotExist, and an
// entry whose line range holds anything the parser did not recognise. Absence
// of evidence is not evidence of absence, and the cost of guessing wrong here
// is a destroyed registration the user wrote by hand.
//
// Shape borrowed wholesale from clean_home.go: an allowlist-shaped judgment,
// a single guard consulted in the same pass that produced the candidate,
// dry-run by default, and an explicit --force for the write.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// codexSkillPruneVerdict is one entry's disposition. A skipped entry carries
// the reason, because an entry silently left behind teaches the user nothing:
// they asked for the ghosts to go and must be told which ones stayed and why.
type codexSkillPruneVerdict struct {
	Entry codexwiring.SkillEntry
	// Eligible is true only when all three REQ-CGP-004 conditions hold AND no
	// disqualifying class applies. Disqualification wins over eligibility
	// (REQ-CGP-022).
	Eligible bool
	// SkipReason is empty exactly when Eligible is true.
	SkipReason string
}

// judgeCodexSkillEntry decides one entry's disposition.
//
// The disqualifying checks run FIRST and unconditionally: REQ-CGP-022 makes
// disqualification beat eligibility, so an entry that both looks absent and
// carries an unrecognised line is preserved.
//
// @MX:WARN: [AUTO] deletion guard — every removal site MUST take its decision from this predicate, in the same pass that produced the candidate
// @MX:REASON: [AUTO] a wrong "eligible" here deletes a registration the user wrote by hand, and the parser's FirstUnrecognizedLine is the only signal separating a ghost from a healthy entry a literal happened to swallow
func judgeCodexSkillEntry(e codexwiring.SkillEntry) codexSkillPruneVerdict {
	skip := func(format string, a ...any) codexSkillPruneVerdict {
		return codexSkillPruneVerdict{Entry: e, SkipReason: fmt.Sprintf(format, a...)}
	}

	// The parser's judgment, consumed as reported. Re-reading the range's text
	// here would resurrect the defect the field exists to close: a comment
	// carrying an odd number of `"""` opens a literal that can swallow a whole
	// healthy registration, and every swallowed line still LOOKS recognisable.
	if e.FirstUnrecognizedLine >= 0 {
		return skip("line %d in its range was not recognised", e.FirstUnrecognizedLine+1)
	}
	if e.Path == "" {
		return skip("declares no path")
	}

	var statPath string
	switch classifyCodexSkillPath(e.Path) {
	case codexPathAbsolute:
		// SPEC-CODEX-SKILL-PATH-READBACK-001: the publisher writes the config's
		// forward-slash form; stat reads the path back in the HOST's own form
		// (REQ-CSRB-001). Classification above ran on the DECLARED form, and the
		// home-relative branch below is left untouched — its filepath.Join
		// product is already native and must never be converted (REQ-CSRB-002).
		statPath = fromConfigPath(e.Path, configPathSeparator)
	case codexPathHomeRelative:
		expanded, ok := expandCodexHomeRelativePath(e.Path)
		if !ok {
			// The home itself is unresolvable, so existence is indeterminate
			// — never absent.
			return skip("the user home does not resolve, so the path is indeterminate")
		}
		statPath = expanded
	case codexPathRelative:
		// No resolution base for a relative path is observed in this
		// repository; stat'ing against the process cwd would decide a
		// deletion on a base nobody chose.
		return skip("relative path — no observed resolution base")
	default:
		return skip("oddly-formed path — not resolvable here")
	}

	_, err := osStatFn(statPath)
	switch {
	case err == nil:
		// The path resolves. A DIRECTORY resolves too, and how Codex treats
		// one here is not observed.
		return skip("the path resolves")
	case errors.Is(err, fs.ErrNotExist):
		return codexSkillPruneVerdict{Entry: e, Eligible: true}
	default:
		// Permission denied, a symlink loop, an I/O error: INDETERMINATE, not
		// absent. Deleting here would destroy a healthy registration on the
		// strength of a stat this process was not allowed to finish.
		return skip("stat did not complete (%v) — existence is indeterminate", err)
	}
}

// pruneCodexSkillEntries returns content with every eligible entry's line
// range removed, plus one verdict per declared entry.
//
// The removal is computed once over the whole line slice. Re-reading and
// rewriting the file per entry would shift the line numbers under the
// remaining ranges.
func pruneCodexSkillEntries(content []byte) ([]byte, []codexSkillPruneVerdict) {
	lines, term := codexwiring.SplitConfigLines(content)
	entries := codexwiring.ParseSkillEntries(content)

	verdicts := make([]codexSkillPruneVerdict, 0, len(entries))
	drop := make(map[int]bool)
	removed := false
	for _, e := range entries {
		v := judgeCodexSkillEntry(e)
		verdicts = append(verdicts, v)
		if !v.Eligible {
			continue
		}
		removed = true
		for i := e.StartLine; i < e.EndLine; i++ {
			drop[i] = true
		}
	}
	if !removed {
		// Nothing eligible: return the input bytes untouched rather than a
		// re-joined copy, so a round-trip defect can never masquerade as a
		// no-op run.
		return content, verdicts
	}

	kept := make([]string, 0, len(lines))
	for i, l := range lines {
		if !drop[i] {
			kept = append(kept, l)
		}
	}
	return codexwiring.JoinConfigLines(kept, term), verdicts
}

// runCleanCodexSkills is the `moai clean --codex-skills` scope.
//
// Fail-open in every direction, exactly as codexStaleSkillFinding is: an
// unresolvable home, an absent or unreadable config, or a config declaring no
// entries all return nil after saying so. A missing input is not an error —
// there is simply nothing to prune.
func runCleanCodexSkills(p printer.Printer, force bool) error {
	codexHome, _ := resolveCodexHomeDir()
	if codexHome == "" {
		p.Info("Codex home does not resolve; nothing to prune")
		return nil
	}
	cfgPath := filepath.Join(codexHome, path.Base(codexwiring.ConfigRelPath))
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		p.Info("%s is absent or unreadable; nothing to prune", cfgPath)
		return nil
	}

	pruned, verdicts := pruneCodexSkillEntries(raw)
	if len(verdicts) == 0 {
		p.Info("%s declares no [[skills.config]] entries; nothing to prune", cfgPath)
		return nil
	}

	var eligible int
	for _, v := range verdicts {
		if v.Eligible {
			eligible++
			if force {
				p.Info("Removing: %s", v.Entry.Path)
			} else {
				p.Info("[dry-run] Would remove: %s", v.Entry.Path)
			}
			continue
		}
		// A skipped entry is reported, never swallowed: silence would leave
		// the user believing every ghost is gone.
		p.Info("Kept: %s (%s)", codexSkillDisplayPath(v.Entry), v.SkipReason)
	}

	if eligible == 0 {
		p.Info("No removable entries in %s (%d declared)", cfgPath, len(verdicts))
		return nil
	}
	if !force {
		p.Info("%d of %d entries eligible for removal. Run with --force to actually remove.", eligible, len(verdicts))
		return nil
	}

	// The backup lands BEFORE the write. This is a user-authored file in the
	// user's home; a write that outruns its backup is unrecoverable.
	backupPath := fmt.Sprintf("%s.bak-%s", cfgPath, time.Now().UTC().Format("20060102T150405Z"))
	if err := os.WriteFile(backupPath, raw, 0o600); err != nil {
		return fmt.Errorf("back up %s: %w", cfgPath, err)
	}
	p.Info("Backup: %s", backupPath)
	p.Info("Backup sha256: %s", sha256Hex(raw))

	if err := os.WriteFile(cfgPath, pruned, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", cfgPath, err)
	}
	p.Info("Removed %d of %d entries from %s", eligible, len(verdicts), cfgPath)
	return nil
}

// codexSkillDisplayPath names an entry in a report. An entry declaring no path
// has nothing to name, so it is identified by its header line instead.
func codexSkillDisplayPath(e codexwiring.SkillEntry) string {
	if e.Path == "" {
		return fmt.Sprintf("entry at line %d", e.StartLine+1)
	}
	return e.Path
}
