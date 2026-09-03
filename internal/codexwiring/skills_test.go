package codexwiring

import (
	"testing"
)

// TestParseSkillEntriesEmpty verifies a config carrying no [[skills.config]]
// array-of-tables yields no entries (the doctor's silent-skip premise).
func TestParseSkillEntriesEmpty(t *testing.T) {
	for _, in := range []string{
		"",
		"model = \"gpt-5\"\n",
		"[mcp_servers.moai]\ncommand = \"moai\"\n",
		"[skills]\nenabled = true\n",
	} {
		if got := ParseSkillEntries([]byte(in)); len(got) != 0 {
			t.Errorf("ParseSkillEntries(%q) = %+v, want no entries", in, got)
		}
	}
}

// TestParseSkillEntriesCanonicalOrder verifies the emitted shape Codex writes:
// path first, enabled second.
func TestParseSkillEntriesCanonicalOrder(t *testing.T) {
	in := "[[skills.config]]\npath = \"/a/SKILL.md\"\nenabled = false\n\n" +
		"[[skills.config]]\npath = \"/b/SKILL.md\"\nenabled = true\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 2 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 2: %+v", len(got), got)
	}
	if got[0].Path != "/a/SKILL.md" || got[0].Enabled != SkillEnabledFalse {
		t.Errorf("entry 0 = %+v, want {/a/SKILL.md false}", got[0])
	}
	if got[1].Path != "/b/SKILL.md" || got[1].Enabled != SkillEnabledTrue {
		t.Errorf("entry 1 = %+v, want {/b/SKILL.md true}", got[1])
	}
}

// TestParseSkillEntriesReversedKeyOrder verifies key order inside an entry is
// not load-bearing (TOML keys are unordered within a table).
func TestParseSkillEntriesReversedKeyOrder(t *testing.T) {
	in := "[[skills.config]]\nenabled = true\npath = \"/a/SKILL.md\"\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(got))
	}
	if got[0].Path != "/a/SKILL.md" || got[0].Enabled != SkillEnabledTrue {
		t.Errorf("reversed-order entry = %+v, want {/a/SKILL.md true}", got[0])
	}
}

// TestParseSkillEntriesEnabledAbsentIsUnspecified verifies an entry with no
// `enabled` key reads as UNSPECIFIED rather than false. Codex's default for
// an absent key is not observed anywhere in this repository, so collapsing it
// onto false would assert a premise nobody verified.
func TestParseSkillEntriesEnabledAbsentIsUnspecified(t *testing.T) {
	got := ParseSkillEntries([]byte("[[skills.config]]\npath = \"/a/SKILL.md\"\n"))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(got))
	}
	if got[0].Enabled != SkillEnabledUnspecified {
		t.Errorf("entry without an `enabled` key = %v, want SkillEnabledUnspecified", got[0].Enabled)
	}
}

// TestParseSkillEntriesEnabledQuoted verifies a quoted value is read at face
// value. Reading `enabled = "true"` as false demotes a LIVE registration to
// stale bookkeeping — the more damaging of the two misreadings.
func TestParseSkillEntriesEnabledQuoted(t *testing.T) {
	cases := []struct {
		in   string
		want SkillEnabled
	}{
		{"[[skills.config]]\npath = \"/a\"\nenabled = \"true\"\n", SkillEnabledTrue},
		{"[[skills.config]]\npath = \"/a\"\nenabled = \"false\"\n", SkillEnabledFalse},
		{"[[skills.config]]\npath = \"/a\"\nenabled = 'true'\n", SkillEnabledTrue},
		{"[[skills.config]]\npath = \"/a\"\nenabled = 'false'\n", SkillEnabledFalse},
		{"[[skills.config]]\npath = \"/a\"\nenabled = yes\n", SkillEnabledUnspecified},
	}
	for _, c := range cases {
		got := ParseSkillEntries([]byte(c.in))
		if len(got) != 1 {
			t.Fatalf("ParseSkillEntries(%q) returned %d entries, want 1", c.in, len(got))
		}
		if got[0].Enabled != c.want {
			t.Errorf("ParseSkillEntries(%q).Enabled = %v, want %v", c.in, got[0].Enabled, c.want)
		}
	}
}

// TestParseSkillEntriesWhitespaceTolerance verifies indentation and spacing
// variation around the header and the assignments still parse.
func TestParseSkillEntriesWhitespaceTolerance(t *testing.T) {
	in := "  [[skills.config]]  \n\tpath   =    \"/a/SKILL.md\"\n  enabled=true\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1: %+v", len(got), got)
	}
	if got[0].Path != "/a/SKILL.md" || got[0].Enabled != SkillEnabledTrue {
		t.Errorf("whitespace-varied entry = %+v, want {/a/SKILL.md true}", got[0])
	}
}

// TestParseSkillEntriesSectionBoundary verifies assignments under a LATER
// table are not folded into the preceding entry — neither `enabled` nor the
// `path` key, which a later table may legitimately also declare.
func TestParseSkillEntriesSectionBoundary(t *testing.T) {
	in := "[[skills.config]]\npath = \"/a/SKILL.md\"\n\n[tui]\nenabled = true\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(got))
	}
	if got[0].Enabled != SkillEnabledUnspecified {
		t.Errorf("`enabled` under [tui] leaked into the skills entry: %+v", got[0])
	}
}

// TestParseSkillEntriesLaterTablePathDoesNotLeak verifies a `path` key
// belonging to a DIFFERENT table that follows is not folded into the
// preceding entry — the same boundary as above, on the other key.
func TestParseSkillEntriesLaterTablePathDoesNotLeak(t *testing.T) {
	in := "[[skills.config]]\nenabled = true\n\n[history]\npath = \"/var/log/codex\"\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1: %+v", len(got), got)
	}
	if got[0].Path != "" {
		t.Errorf("`path` under [history] leaked into the skills entry: %+v", got[0])
	}
}

// TestParseSkillEntriesComments verifies trailing comments on the header and
// on either assignment do not defeat the matchers.
func TestParseSkillEntriesComments(t *testing.T) {
	in := "# registered skills\n[[skills.config]] # moai\npath = \"/a/SKILL.md\" # the file\nenabled = true # on\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1: %+v", len(got), got)
	}
	if got[0].Path != "/a/SKILL.md" || got[0].Enabled != SkillEnabledTrue {
		t.Errorf("commented entry = %+v, want {/a/SKILL.md true}", got[0])
	}
}

// TestParseSkillEntriesCRLF verifies a config written with Windows line
// endings parses identically — the carriage return must not ride into the
// declared path, where it would guarantee a stat failure and a false finding.
func TestParseSkillEntriesCRLF(t *testing.T) {
	in := "[[skills.config]]\r\npath = \"/a/SKILL.md\"\r\nenabled = true\r\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1: %+v", len(got), got)
	}
	if got[0].Path != "/a/SKILL.md" || got[0].Enabled != SkillEnabledTrue {
		t.Errorf("CRLF entry = %+v, want {/a/SKILL.md true}", got[0])
	}
}

// TestParseSkillEntriesMultilineStringPhantom verifies a [[skills.config]]
// header written INSIDE a multi-line string literal is not read as a
// registration. Without the literal state the parser manufactures a phantom
// entry pointing at a path that was never registered, and the doctor then
// warns about a config that is entirely healthy.
func TestParseSkillEntriesMultilineStringPhantom(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{
			name: "basic",
			in: "[profile]\ndescription = \"\"\"\n" +
				"[[skills.config]]\npath = \"/nonexistent/phantom\"\nenabled = true\n" +
				"\"\"\"\n",
		},
		{
			name: "literal",
			in: "[profile]\ndescription = '''\n" +
				"[[skills.config]]\npath = \"/nonexistent/p2\"\nenabled = false\n" +
				"'''\n",
		},
	}
	for _, c := range cases {
		if got := ParseSkillEntries([]byte(c.in)); len(got) != 0 {
			t.Errorf("%s: ParseSkillEntries returned %+v, want no entries (header is inside a string literal)", c.name, got)
		}
	}
}

// TestParseSkillEntriesMultilineStringCloses verifies the literal state is a
// state, not a trapdoor: a real entry AFTER a closed multi-line string is
// still parsed. A one-way skip would silently swallow every later entry.
func TestParseSkillEntriesMultilineStringCloses(t *testing.T) {
	in := "[profile]\ndescription = \"\"\"\nsome [[skills.config]] prose\n\"\"\"\n\n" +
		"[[skills.config]]\npath = \"/real/SKILL.md\"\nenabled = true\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1: %+v", len(got), got)
	}
	if got[0].Path != "/real/SKILL.md" || got[0].Enabled != SkillEnabledTrue {
		t.Errorf("entry after a closed literal = %+v, want {/real/SKILL.md true}", got[0])
	}
}

// TestParseSkillEntriesInlineMultilineNotOpened verifies a multi-line
// delimiter that opens AND closes on one line does not put the parser into
// the literal state — an even delimiter count leaves nothing open.
func TestParseSkillEntriesInlineMultilineNotOpened(t *testing.T) {
	in := "[profile]\ndescription = \"\"\"inline\"\"\"\n\n[[skills.config]]\npath = \"/a/SKILL.md\"\nenabled = true\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1: %+v", len(got), got)
	}
	if got[0].Path != "/a/SKILL.md" {
		t.Errorf("entry after an inline literal = %+v, want /a/SKILL.md", got[0])
	}
}

// TestParseSkillEntriesSimilarHeaderRejected verifies a neighbouring
// array-of-tables is not mistaken for [[skills.config]].
func TestParseSkillEntriesSimilarHeaderRejected(t *testing.T) {
	for _, in := range []string{
		"[[skills.configs]]\npath = \"/a/SKILL.md\"\n",
		"[skills.config]\npath = \"/a/SKILL.md\"\n",
		"[[skills.config.extra]]\npath = \"/a/SKILL.md\"\n",
	} {
		if got := ParseSkillEntries([]byte(in)); len(got) != 0 {
			t.Errorf("ParseSkillEntries(%q) = %+v, want no entries", in, got)
		}
	}
}
