package codexwiring

import "testing"

// skills_extent_test.go — the line-extent and line-recognition axis of the
// [[skills.config]] parser (SPEC-CODEX-GHOST-SKILLS-PRUNE-001 M1).
//
// The pruner deletes an entry by line range, so the parser must report which
// lines an entry occupies AND whether every line in that range was actually
// consumed by a recognised branch. Recognition is a PARSER-STATE judgment,
// never a re-reading of the line's text: a line swallowed by the multi-line
// literal branch can look exactly like a header, a key, or a comment, and a
// text re-scan would wave it through.

func TestParseSkillEntriesExtentBasic(t *testing.T) {
	t.Parallel()

	content := []byte("[[skills.config]]\n" + // 0
		"path = \"/a\"\n" + // 1
		"enabled = false\n" + // 2
		"\n" + // 3 (trailing blank — excluded from the extent)
		"[other]\n" + // 4
		"k = 1\n") // 5

	entries := ParseSkillEntries(content)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.StartLine != 0 || e.EndLine != 3 {
		t.Errorf("extent = [%d,%d), want [0,3)", e.StartLine, e.EndLine)
	}
	if e.FirstUnrecognizedLine != -1 {
		t.Errorf("FirstUnrecognizedLine = %d, want -1", e.FirstUnrecognizedLine)
	}
}

func TestParseSkillEntriesExtentClosesAtEOF(t *testing.T) {
	t.Parallel()

	content := []byte("[other]\n" + // 0
		"[[skills.config]]\n" + // 1
		"path = \"/a\"\n") // 2

	entries := ParseSkillEntries(content)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].StartLine != 1 || entries[0].EndLine != 3 {
		t.Errorf("extent = [%d,%d), want [1,3)", entries[0].StartLine, entries[0].EndLine)
	}
}

func TestParseSkillEntriesRecognisesBlankAndCommentLines(t *testing.T) {
	t.Parallel()

	content := []byte("[[skills.config]]\n" +
		"# a whole-line comment\n" +
		"\n" +
		"path = \"/a\" # trailing comment\n" +
		"enabled = true\n" +
		"[other]\n")

	entries := ParseSkillEntries(content)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].FirstUnrecognizedLine != -1 {
		t.Errorf("FirstUnrecognizedLine = %d, want -1", entries[0].FirstUnrecognizedLine)
	}
}

func TestParseSkillEntriesUnrecognizedLines(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
		want    int // FirstUnrecognizedLine
	}{
		{
			name: "unknown key",
			content: "[[skills.config]]\n" + // 0
				"path = \"/a\"\n" + // 1
				"notes = \"hi\"\n" + // 2 — neither path nor enabled
				"[other]\n",
			want: 2,
		},
		{
			name: "multi-line literal inside the entry",
			content: "[[skills.config]]\n" + // 0
				"path = \"/a\"\n" + // 1
				"notes = \"\"\"\n" + // 2 — opens a literal
				"prose\n" + // 3 — swallowed
				"\"\"\"\n" + // 4 — closes
				"[other]\n",
			want: 2,
		},
		{
			name: "inline multi-line delimiters on one line",
			content: "[[skills.config]]\n" +
				"path = \"/a\"\n" +
				"notes = \"\"\"inline\"\"\"\n" + // 2 — even count, falls to the switch
				"[other]\n",
			want: 2,
		},
		{
			name: "array continuation read as a table header",
			content: "[[skills.config]]\n" + // 0
				"path = \"/a\"\n" + // 1
				"matrix = [\n" + // 2 — unknown key
				"[\"x\", \"y\"]\n" + // 3 — anyTableRe matches: closes the extent
				"]\n" +
				"[other]\n",
			want: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			entries := ParseSkillEntries([]byte(tc.content))
			if len(entries) != 1 {
				t.Fatalf("entries = %d, want 1", len(entries))
			}
			if got := entries[0].FirstUnrecognizedLine; got != tc.want {
				t.Errorf("FirstUnrecognizedLine = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestParseSkillEntriesSwallowedRegistration is the 7-d counter-example
// (acceptance.md AC-CGP-003 row 7-d). A comment carrying an ODD number of
// `"""` opens a literal that swallows a whole healthy registration. The
// parser reports ONE entry whose extent contains that registration, and every
// line in the swallowed span LOOKS recognisable — which is why recognition
// must be decided by parser state.
func TestParseSkillEntriesSwallowedRegistration(t *testing.T) {
	t.Parallel()

	content := []byte("[[skills.config]]\n" + // 0
		"path = \"/gone\"\n" + // 1
		"# uses \"\"\" in prose\n" + // 2 — opens the literal
		"[[skills.config]]\n" + // 3 — swallowed
		"path = \"/exists\"\n" + // 4 — swallowed
		"# and \"\"\" again\n" + // 5 — closes the literal
		"[other]\n") // 6

	entries := ParseSkillEntries(content)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1 (the swallowed registration must be invisible)", len(entries))
	}
	e := entries[0]
	if e.Path != "/gone" {
		t.Fatalf("Path = %q, want /gone", e.Path)
	}
	if e.StartLine != 0 || e.EndLine != 6 {
		t.Errorf("extent = [%d,%d), want [0,6)", e.StartLine, e.EndLine)
	}
	if e.FirstUnrecognizedLine != 2 {
		t.Errorf("FirstUnrecognizedLine = %d, want 2 (the literal-opening comment)", e.FirstUnrecognizedLine)
	}
}

// TestParseSkillEntriesSwallowedRegistrationNarrow is variant 7-d': the
// swallowed span carries no header at all, so a "one header per extent" guard
// would not fire. Recognition must still reject the extent.
func TestParseSkillEntriesSwallowedRegistrationNarrow(t *testing.T) {
	t.Parallel()

	content := []byte("[[skills.config]]\n" + // 0
		"path = \"/gone\"\n" + // 1
		"# uses \"\"\" in prose\n" + // 2 — opens
		"path = \"/exists\"\n" + // 3 — swallowed
		"# and \"\"\" again\n" + // 4 — closes
		"[other]\n") // 5

	entries := ParseSkillEntries(content)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].Path != "/gone" {
		t.Fatalf("Path = %q, want /gone", entries[0].Path)
	}
	if entries[0].FirstUnrecognizedLine != 2 {
		t.Errorf("FirstUnrecognizedLine = %d, want 2", entries[0].FirstUnrecognizedLine)
	}
}

// TestConfigLinesRoundTrip is AC-CGP-013: the reassembly function is exercised
// DIRECTLY, not through a pruner run that writes nothing. splitLines trims one
// trailing newline, so "a\n" and "a" yield the same slice — the line-ending
// state must ride outside the slice or variant b breaks.
func TestConfigLinesRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
	}{
		{"a: trailing newline", "[[skills.config]]\npath = \"/a\"\n[other]\n"},
		{"b: no trailing newline", "[[skills.config]]\npath = \"/a\"\n[other]"},
		{"c: CRLF", "[[skills.config]]\r\npath = \"/a\"\r\n[other]\r\n"},
		{"c': CRLF without a trailing terminator", "[[skills.config]]\r\npath = \"/a\"\r\n[other]"},
		{"empty", ""},
		{"lone newline", "\n"},
		{"two trailing newlines", "a\n\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			lines, term := SplitConfigLines([]byte(tc.content))
			got := string(JoinConfigLines(lines, term))
			if got != tc.content {
				t.Errorf("round trip = %q, want %q", got, tc.content)
			}
		})
	}
}
