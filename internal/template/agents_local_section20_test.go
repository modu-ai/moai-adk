// agents_local_section20_test.go — SPEC-AGENT-CORE4-001 REQ-AC4-012(a): the
// CLAUDE.local.md §2.0 text no longer describes the agent copies as hand-edited
// or C2 as the source layer, and names the agent source directory instead.
//
// The check reads the section and nothing else, so it survives later edits of
// the file (piece 2 rewrites the transitional clauses about the eight agents
// not yet generated). The whole-line comparison against the exact new text is a
// one-time M2 cell, not part of this permanent test.
package template

import (
	"os"
	"strings"
	"testing"
)

// agentsLocalPath is the repository-root CLAUDE.local.md, relative to this package.
const agentsLocalPath = "../../CLAUDE.local.md"

// section20Heading prefixes the §2.0 heading line; the section runs to the next "### " heading.
const section20Heading = "### [HARD] §2.0"

// agentSourceDirPath is the directory §2.0 must name as the edit surface.
const agentSourceDirPath = "internal/template/agentemit/source/"

// oldSection20Phrases are whole phrases that occurred only in the eight lines
// the rewrite replaces. They are phrases rather than words on purpose: the bare
// word "hand-edited" appears in the new table rows too.
var oldSection20Phrases = []string{
	"Local dogfooding, hand-edited",
	"Distribution mirror, hand-edited",
	"the third is not edited by hand",
	"only one of them is not a hand-edit target",
	"not** a byte-identical",
	"fix C2 and regenerate",
	"Machine-emitted from C2",
	"emitted from `.md`",
}

// section20Of returns the §2.0 section of the CLAUDE.local.md text.
func section20Of(text string) (string, bool) {
	lines := strings.Split(text, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, section20Heading) {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "### ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n"), true
}

// staleSection20Phrases returns the old phrases still present in section.
func staleSection20Phrases(section string) []string {
	var stale []string
	for _, p := range oldSection20Phrases {
		if strings.Contains(section, p) {
			stale = append(stale, p)
		}
	}
	return stale
}

func TestAgentsLocalSection20Text(t *testing.T) {
	raw, err := os.ReadFile(agentsLocalPath)
	if err != nil {
		t.Fatalf("read CLAUDE.local.md: %v", err)
	}
	section, found := section20Of(string(raw))

	t.Run("section_2_0_exists", func(t *testing.T) {
		if !found {
			t.Fatalf("CLAUDE.local.md has no heading starting %q", section20Heading)
		}
		if strings.Count(section, "\n") < 3 {
			t.Errorf("section is only %d line(s) long; the heading's body was not located", strings.Count(section, "\n")+1)
		}
	})
	t.Run("old_phrases_absent_from_section_2_0", func(t *testing.T) {
		if !found {
			t.Fatal("no §2.0 section to inspect")
		}
		for _, p := range staleSection20Phrases(section) {
			t.Errorf("§2.0 still carries the superseded phrase %q", p)
		}
	})
	t.Run("source_path_named", func(t *testing.T) {
		if !found {
			t.Fatal("no §2.0 section to inspect")
		}
		if !strings.Contains(section, agentSourceDirPath) {
			t.Errorf("§2.0 does not name the agent source directory %q", agentSourceDirPath)
		}
	})
	// The detector must be able to fail: fed the superseded table rows it
	// reports exactly the phrases they carry.
	t.Run("detector_flags_superseded_rows", func(t *testing.T) {
		old := "| C1 | `.claude/agents/moai/*.md` | Local dogfooding, hand-edited |\n" +
			"| C2 | `internal/template/templates/.claude/agents/moai/*.md` | Distribution mirror, hand-edited — **the neutral source layer** |\n"
		got := staleSection20Phrases(old)
		if len(got) != 2 || got[0] != "Local dogfooding, hand-edited" || got[1] != "Distribution mirror, hand-edited" {
			t.Errorf("detector reported %v for the superseded rows", got)
		}
		if got := staleSection20Phrases("| Source | `" + agentSourceDirPath + "*.md` | **The only hand-edited layer** |"); len(got) != 0 {
			t.Errorf("detector flagged the new Source row: %v", got)
		}
	})
}
