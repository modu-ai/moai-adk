// Package mirrornotice turns a template deployment's skill-mirror outcome into
// the lines a user sees.
//
// The deployer reports its outcome by return value and never prints — the
// template package has no output surface, by design — so the three call sites
// that deploy templates (moai init, moai update template sync, and clean
// reinstall) are what decide whether a fallback reaches the user at all. This
// package is the single place that decides WHAT they say, so all three say the
// same thing.
//
// Lines returns strings rather than writing to a writer because one of the
// three call sites (moai init) needs a []string: it appends to
// project.InitResult.Warnings, which an existing summary panel renders.
package mirrornotice

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/template"
)

const (
	// Token is the stable substring every notice summary line carries. Code
	// asserting that a notice is present — or absent — matches on this rather
	// than on wording, which stays free to change.
	Token = ".agents/skills mirror"

	// maxFailedExamples caps how many failure warnings the notice quotes. The
	// notice must not grow with the number of skills, so the count carries the
	// scale and the examples carry the diagnosis. A deployment mirrors dozens
	// of skills; a per-skill listing would bury the summary that matters.
	maxFailedExamples = 3
)

// Lines returns the notice for a deploy result, or nothing when the run has
// nothing the user needs to act on.
//
// Two outcomes are reported: skills that fell back to a directory copy, and
// skills whose mirror could not be created at all.
//
// MirrorModeSkipped is deliberately NOT reported. Its warning ("a non-symlink
// entry already exists … left untouched") attributes the deployer's own earlier
// copy to the user, so on a fallback platform it would send the user looking
// for a file of theirs that does not exist — worse than silence. Reporting it
// waits for the deployer to gain a discriminator between its own output and a
// real user directory.
func Lines(res *template.DeployResult) []string {
	if res == nil {
		return nil
	}

	var copied, failed int
	var examples []string
	for _, e := range res.SkillMirrors {
		switch e.Mode {
		case template.MirrorModeCopy:
			copied++
		case template.MirrorModeFailed:
			failed++
			if e.Warning != "" && len(examples) < maxFailedExamples {
				examples = append(examples, e.Warning)
			}
		}
	}

	if copied == 0 && failed == 0 {
		return nil
	}

	var lines []string
	if copied > 0 {
		lines = append(lines, fmt.Sprintf(
			"%s: %d skill(s) copied instead of linked (this system could not create symlinks). "+
				"The copies do not follow later updates to .claude/skills.",
			Token, copied))
	}
	if failed > 0 {
		lines = append(lines, fmt.Sprintf(
			"%s: %d skill(s) could not be mirrored; they remain reachable via .claude/skills.",
			Token, failed))
		for _, w := range examples {
			lines = append(lines, "  "+w)
		}
	}
	return lines
}

// WriteTo writes the notice to w, one line each. A run with nothing to report
// writes nothing at all, so a quiet deployment stays byte-identical.
func WriteTo(w io.Writer, res *template.DeployResult) {
	if w == nil {
		return
	}
	for _, line := range Lines(res) {
		_, _ = fmt.Fprintln(w, line)
	}
}
