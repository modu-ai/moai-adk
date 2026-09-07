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

// SkillEnabled is the four-state reading of an entry's `enabled` key.
//
// The states are kept distinct deliberately. Collapsing "no key" onto false
// asserts a default Codex does NOT apply — measured on codex-cli 0.153.4, an
// entry omitting the key makes codex exit 1 with ``missing field `enabled` in
// `skills.config` ``, so there is no default to collapse onto. The reading is
// reported as DECLARED; what codex then does with it is the doctor's finding to
// state, not this parser's.
//
// The fourth state (SkillEnabledNonBoolean) was added by t508. It separates
// "nothing was said" from "something was said that codex rejects" — two
// observations codex itself reports with different errors, and which the
// previous tri-state collapsed onto one.
type SkillEnabled int

const (
	// SkillEnabledUnspecified is an entry declaring no `enabled` key. It
	// asserts nothing about what the author intended.
	//
	// It no longer covers an unrecognised VALUE: that is
	// SkillEnabledNonBoolean. Folding the two together made the integer case
	// exactly as silent as the absent case, which is the defect t508 closes.
	SkillEnabledUnspecified SkillEnabled = iota
	// SkillEnabledTrue is an entry declaring enabled as the bare TOML boolean
	// true.
	SkillEnabledTrue
	// SkillEnabledFalse is an entry declaring enabled as the bare TOML boolean
	// false.
	SkillEnabledFalse
	// SkillEnabledNonBoolean is an entry that DECLARES `enabled` with a value
	// that is not a bare TOML boolean — an integer, a quoted string, a
	// bareword. Codex requires a bare boolean and refuses to load the config
	// otherwise, so this state asserts something quite specific: the key is
	// present AND codex cannot read it.
	SkillEnabledNonBoolean
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
	// StartLine is the 0-based index of the entry's header line, and EndLine
	// the exclusive end of its extent, both indexing the slice
	// SplitConfigLines returns. Trailing blank lines are excluded: swallowing
	// them would silently change the spacing between entries when one is
	// removed.
	StartLine int
	EndLine   int
	// FirstUnrecognizedLine is the index of the first line in [StartLine,
	// EndLine) that the parser did NOT consume through a recognised branch,
	// or -1 when every line was recognised.
	//
	// The five recognised branches are: this entry's own header, a `path`
	// assignment, an `enabled` assignment, a blank line, and a whole-line
	// comment. Everything else is unrecognised — including every line the
	// multi-line-literal branch swallowed, whatever that line looks like.
	// That distinction is the whole point of reporting the value from here:
	// a comment carrying an odd number of `"""` opens a literal that can
	// swallow an entire healthy registration, and each swallowed line still
	// LOOKS like a header, a key, or a comment. A consumer re-reading the
	// extent's text cannot tell the two apart; the parser can, because it
	// knows which branch spent the line.
	//
	// An index rather than a boolean, because a consumer that skips an entry
	// owes the user the reason — which line stopped it — and one value serves
	// both the judgment and the report.
	FirstUnrecognizedLine int
}

// LineTerm carries the line-ending state splitLines discards.
//
// splitLines trims ONE trailing newline before splitting, so "a\n" and "a"
// produce the same slice: whether the body ended with a terminator cannot be
// recovered from the lines alone and must ride alongside them.
//
// CRLF needs no field of its own. splitLines splits on "\n" only, so each
// line keeps its own trailing "\r" and a plain "\n" join restores it — the
// only byte at risk is the final terminator, which TrailingNewline covers.
type LineTerm struct {
	TrailingNewline bool
}

// SplitConfigLines splits a config body into lines plus the line-ending state
// needed to rebuild it. It is the read half of the lossless pair; the parser's
// line indices address exactly this slice.
func SplitConfigLines(content []byte) ([]string, LineTerm) {
	body := string(content)
	return splitLines(body), LineTerm{TrailingNewline: strings.HasSuffix(body, "\n")}
}

// JoinConfigLines rebuilds a config body from lines and their line-ending
// state. SplitConfigLines followed by JoinConfigLines is byte-identical to the
// input for every body, with or without a trailing terminator, LF or CRLF.
//
// This is a pure string operation. It writes no file — the read/write boundary
// stays where skills.go's header comment puts it.
func JoinConfigLines(lines []string, term LineTerm) []byte {
	if len(lines) == 0 {
		return nil
	}
	out := strings.Join(lines, "\n")
	if term.TrailingNewline {
		out += "\n"
	}
	return []byte(out)
}

// Entry-shape detectors. The header is anchored so [[skills.configs]],
// [skills.config] (a plain table, not array-of-tables) and
// [[skills.config.extra]] do not satisfy the match — each is a distinct TOML
// surface.
//
// The enabled matcher is split in two: skillEnabledKeyRe recognises a BARE TOML
// boolean, and skillEnabledDeclRe recognises the key being declared at all. A
// line matching the second but not the first is a declared non-boolean value.
//
// This REVERSES a deliberate earlier leniency, so the argument being reversed is
// answered here rather than deleted — a future reader who finds only a narrowed
// regex and no reasoning would read this as a regression. The previous comment
// read, in full:
//
//	The enabled matcher accepts a quoted value as well as a bare one. A quoted
//	"true" is a TOML string rather than a boolean, so it is arguably malformed;
//	reading it as false, however, silently DEMOTES a live registration to stale
//	bookkeeping, which is the more damaging misreading. The declared intent is
//	unambiguous, so it is taken at face value and reported as declared.
//
// It makes two arguments. Both are answered:
//
//   - "reading it as false DEMOTES a live registration" — FALSIFIED by
//     measurement. On codex-cli 0.153.4, `enabled = "true"` makes `codex mcp
//     list` exit 1 with ``invalid type: string "true", expected a boolean``.
//     There is no live registration to demote, because codex refuses to load
//     the config at all. The argument weighed two readings of a working config;
//     the config does not work. And the reversal does not do what the argument
//     feared either: the value reads as NonBoolean, never as false, so it is
//     not demoted to stale bookkeeping — it is reported as the fatal shape it is.
//   - "the declared intent is unambiguous, so it is taken at face value" — TRUE,
//     and beside the point. Codex never reads the intent. It reads the bytes,
//     finds a string where a boolean is required, and stops. Face-value
//     reporting of a declared intent is right only where the consuming tool acts
//     on that intent; here it does not, and reporting a registration codex
//     refuses to load is confident and wrong.
var (
	skillsEntryHeaderRe = regexp.MustCompile(`^\[\[skills\.config\]\]\s*(#.*)?$`)
	skillPathKeyRe      = regexp.MustCompile(`^path\s*=\s*"([^"]*)"\s*(#.*)?$`)
	skillEnabledKeyRe   = regexp.MustCompile(`^enabled\s*=\s*(true|false)\s*(#.*)?$`)
	skillEnabledDeclRe  = regexp.MustCompile(`^enabled\s*=`)
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
	lines, _ := SplitConfigLines(content)

	var entries []SkillEntry
	openIdx := -1         // index in entries of the entry whose extent is open
	var recognized []bool // one flag per line of the open extent, from StartLine
	openDelim := ""       // non-empty while inside a multi-line string literal

	// mark records how the current line was spent. Only lines inside an open
	// extent are tracked; outside one there is nothing to disqualify.
	mark := func(ok bool) {
		if openIdx >= 0 {
			recognized = append(recognized, ok)
		}
	}

	// closeEntry finalises the open extent at the exclusive bound end.
	closeEntry := func(end int) {
		if openIdx < 0 {
			return
		}
		e := &entries[openIdx]
		// Trailing blank lines belong to the gap between entries, not to the
		// entry — but only when they were recognised as blank. A blank line
		// swallowed by a literal is unrecognised and must stay inside the
		// extent, or the disqualifying signal is trimmed away with it.
		for end > e.StartLine+1 &&
			strings.TrimSpace(lines[end-1]) == "" &&
			recognized[end-1-e.StartLine] {
			end--
		}
		e.EndLine = end
		e.FirstUnrecognizedLine = -1
		for i := e.StartLine; i < end; i++ {
			if !recognized[i-e.StartLine] {
				e.FirstUnrecognizedLine = i
				break
			}
		}
		openIdx = -1
		recognized = nil
	}

	for idx, raw := range lines {
		line := strings.TrimSpace(raw)

		if openDelim != "" {
			// Inside a literal. The closing delimiter ends it; the remainder
			// of that line cannot legally begin a new assignment, so the
			// whole line is spent either way. Spent HERE means spent by this
			// branch: the line is unrecognised no matter what it looks like.
			if strings.Contains(line, openDelim) {
				openDelim = ""
			}
			mark(false)
			continue
		}
		if d := multilineOpener(line); d != "" {
			openDelim = d
			mark(false)
			continue
		}

		switch {
		case skillsEntryHeaderRe.MatchString(line):
			closeEntry(idx)
			entries = append(entries, SkillEntry{
				StartLine:             idx,
				EndLine:               idx + 1,
				FirstUnrecognizedLine: -1,
			})
			openIdx = len(entries) - 1
			recognized = []bool{true} // the entry's own header
		case anyTableRe.MatchString(line):
			// Any other table header ends the current entry's extent.
			closeEntry(idx)
		default:
			if openIdx < 0 {
				continue
			}
			if m := skillPathKeyRe.FindStringSubmatch(line); m != nil {
				entries[openIdx].Path = m[1]
				mark(true)
			} else if m := skillEnabledKeyRe.FindStringSubmatch(line); m != nil {
				entries[openIdx].Enabled = skillEnabledFrom(m)
				mark(true)
			} else if skillEnabledDeclRe.MatchString(line) {
				// The key is declared with something codex will not accept.
				// The line is marked UNRECOGNISED, which is the same
				// disposition it already had when the old matcher failed on
				// it (`enabled = 1`): a line codex chokes on is not a line
				// this parser consumed. The consequence is deliberate and
				// safe — a consumer keyed on FirstUnrecognizedLine (the prune
				// verb's deletion guard) preserves such an entry rather than
				// removing it.
				entries[openIdx].Enabled = SkillEnabledNonBoolean
				mark(false)
			} else {
				// A blank line and a whole-line comment are recognised; an
				// unknown key is not. The original parser folded all three
				// onto one no-op path, so this classification is NEW
				// computation, not a readback of state already held.
				mark(line == "" || strings.HasPrefix(line, "#"))
			}
		}
	}
	closeEntry(len(lines)) // EOF closes the last extent
	return entries
}

// skillEnabledFrom maps the bare-boolean matcher's single capture group onto a
// reading. The regex admits only `true` and `false`, so the fallback is
// unreachable in practice and exists so a later widening of the pattern cannot
// silently produce a wrong reading.
func skillEnabledFrom(m []string) SkillEnabled {
	switch m[1] {
	case "true":
		return SkillEnabledTrue
	case "false":
		return SkillEnabledFalse
	}
	return SkillEnabledNonBoolean
}
