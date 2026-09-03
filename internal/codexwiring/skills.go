package codexwiring

// skills.go — the read-only [[skills.config]] inspector (t451).
//
// Codex records every registered skill in the user-layer config.toml as an
// array-of-tables entry carrying a path and an enabled flag. A registration
// whose path no longer exists is invisible: Codex neither prunes it nor
// complains, so the doctor is the only surface that can report it. This
// parser is READ-ONLY by construction — nothing here writes, and the
// user-layer config stays byte-invariant (the REQ-CW-005 posture the MCP
// table inspector already takes).
//
// Hand-rolled on regexp/strings like configtoml.go: the repository carries no
// TOML dependency, and the surface parsed here is one fixed array-of-tables
// with two scalar keys.

import (
	"regexp"
	"strings"
)

// SkillEntry is one [[skills.config]] entry as declared on disk.
type SkillEntry struct {
	// Path is the declared SKILL.md location (absolute, as Codex writes it).
	// An entry that declares no path key yields the empty string.
	Path string
	// Enabled is the declared enabled flag. TOML has no implicit default for
	// an absent key, so a missing `enabled` reads as false — the zero value,
	// and the same verdict Codex reaches for an unregistered skill.
	Enabled bool
}

// Entry-shape detectors. The header is anchored so [[skills.configs]],
// [skills.config] (a plain table, not array-of-tables) and
// [[skills.config.extra]] do not satisfy the match — each is a distinct TOML
// surface.
var (
	skillsEntryHeaderRe = regexp.MustCompile(`^\[\[skills\.config\]\]\s*(#.*)?$`)
	skillPathKeyRe      = regexp.MustCompile(`^path\s*=\s*"([^"]*)"\s*(#.*)?$`)
	skillEnabledKeyRe   = regexp.MustCompile(`^enabled\s*=\s*(true|false)\s*(#.*)?$`)
)

// ParseSkillEntries reads the [[skills.config]] entries a Codex config.toml
// declares, in file order. Keys may appear in either order within an entry;
// an entry's extent ends at the next table header of any kind, so
// assignments belonging to a later table never fold into it.
//
// Malformed input yields fewer entries, never an error: this feeds an
// advisory diagnostic, and a parse gap must degrade to silence rather than
// to a false finding.
func ParseSkillEntries(content []byte) []SkillEntry {
	var entries []SkillEntry
	inEntry := false
	for _, raw := range splitLines(string(content)) {
		line := strings.TrimSpace(raw)
		switch {
		case skillsEntryHeaderRe.MatchString(line):
			entries = append(entries, SkillEntry{})
			inEntry = true
		case anyTableRe.MatchString(line):
			// Any other table header ends the current entry's extent.
			inEntry = false
		case inEntry:
			if m := skillPathKeyRe.FindStringSubmatch(line); m != nil {
				entries[len(entries)-1].Path = m[1]
			} else if m := skillEnabledKeyRe.FindStringSubmatch(line); m != nil {
				entries[len(entries)-1].Enabled = m[1] == "true"
			}
		}
	}
	return entries
}
