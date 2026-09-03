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

// SkillEnabled is the tri-state reading of an entry's `enabled` key.
//
// The three states are kept distinct deliberately. Collapsing "no key" onto
// false asserts a default this repository has NOT observed Codex to apply,
// and an unverified default is an unobserved premise, not a fact
// (verification-claim-integrity §1). The reading is reported as DECLARED; how
// Codex coerces or defaults a value is left unclaimed.
type SkillEnabled int

const (
	// SkillEnabledUnspecified is an entry declaring no `enabled` key, or one
	// whose value this parser does not recognise. It asserts nothing.
	SkillEnabledUnspecified SkillEnabled = iota
	// SkillEnabledTrue is an entry declaring enabled true.
	SkillEnabledTrue
	// SkillEnabledFalse is an entry declaring enabled false.
	SkillEnabledFalse
)

// SkillEntry is one [[skills.config]] entry as declared on disk.
type SkillEntry struct {
	// Path is the declared SKILL.md location, verbatim as written in the
	// TOML basic string (no escape-sequence decoding). Codex writes an
	// absolute path, but a hand-edited config may declare a ~-relative,
	// relative, or oddly-formed one — consumers must classify the shape,
	// not assume absoluteness. An entry that declares no path key yields
	// the empty string.
	Path string
	// Enabled is the declared enabled flag, tri-state (see SkillEnabled).
	Enabled SkillEnabled
}

// Entry-shape detectors. The header is anchored so [[skills.configs]],
// [skills.config] (a plain table, not array-of-tables) and
// [[skills.config.extra]] do not satisfy the match — each is a distinct TOML
// surface.
//
// The enabled matcher accepts a quoted value as well as a bare one. A quoted
// "true" is a TOML string rather than a boolean, so it is arguably malformed;
// reading it as false, however, silently DEMOTES a live registration to stale
// bookkeeping, which is the more damaging misreading. The declared intent is
// unambiguous, so it is taken at face value and reported as declared.
var (
	skillsEntryHeaderRe = regexp.MustCompile(`^\[\[skills\.config\]\]\s*(#.*)?$`)
	skillPathKeyRe      = regexp.MustCompile(`^path\s*=\s*"([^"]*)"\s*(#.*)?$`)
	skillEnabledKeyRe   = regexp.MustCompile(`^enabled\s*=\s*(?:(true|false)|"(true|false)"|'(true|false)')\s*(#.*)?$`)
)

// multilineDelims are the TOML multi-line string delimiters, longest-first so
// a `"""` is never mistaken for a run of `"`.
var multilineDelims = []string{`"""`, `'''`}

// multilineOpener reports the delimiter a line leaves OPEN, or "" when the
// line closes everything it opened. A delimiter appearing an odd number of
// times on one line leaves a multi-line literal open; an even number (the
// `x = """inline"""` form, or a comment quoting the delimiter twice) does not.
//
// The bias is deliberately conservative: an ambiguous line is treated as
// entering a literal, so the parser under-reports rather than inventing an
// entry. Silence is the safe direction for an advisory diagnostic; a phantom
// entry becomes a false finding against a healthy config.
func multilineOpener(line string) string {
	for _, d := range multilineDelims {
		if n := strings.Count(line, d); n > 0 && n%2 == 1 {
			return d
		}
	}
	return ""
}

// ParseSkillEntries reads the [[skills.config]] entries a Codex config.toml
// declares, in file order. Keys may appear in either order within an entry;
// an entry's extent ends at the next table header of any kind, so
// assignments belonging to a later table never fold into it.
//
// Text inside a multi-line string literal (either delimiter) is skipped
// entirely: a
// `[[skills.config]]` header written inside a documentation string is NOT a
// registration, and treating it as one manufactures a phantom entry pointing
// at a path that was never registered — a false finding against valid TOML.
//
// Malformed input yields fewer entries, never an error: this feeds an
// advisory diagnostic, and a parse gap must degrade to silence rather than
// to a false finding.
func ParseSkillEntries(content []byte) []SkillEntry {
	var entries []SkillEntry
	inEntry := false
	openDelim := "" // non-empty while inside a multi-line string literal
	for _, raw := range splitLines(string(content)) {
		line := strings.TrimSpace(raw)

		if openDelim != "" {
			// Inside a literal. The closing delimiter ends it; the remainder
			// of that line cannot legally begin a new assignment, so the
			// whole line is spent either way.
			if strings.Contains(line, openDelim) {
				openDelim = ""
			}
			continue
		}
		if d := multilineOpener(line); d != "" {
			openDelim = d
			continue
		}

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
				entries[len(entries)-1].Enabled = skillEnabledFrom(m)
			}
		}
	}
	return entries
}

// skillEnabledFrom folds the enabled matcher's three alternation groups (bare,
// double-quoted, single-quoted) onto one tri-state value.
func skillEnabledFrom(m []string) SkillEnabled {
	for _, g := range m[1:4] {
		switch g {
		case "true":
			return SkillEnabledTrue
		case "false":
			return SkillEnabledFalse
		}
	}
	return SkillEnabledUnspecified
}
