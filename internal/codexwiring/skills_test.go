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
	if got[0].Path != "/a/SKILL.md" || got[0].Enabled {
		t.Errorf("entry 0 = %+v, want {/a/SKILL.md false}", got[0])
	}
	if got[1].Path != "/b/SKILL.md" || !got[1].Enabled {
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
	if got[0].Path != "/a/SKILL.md" || !got[0].Enabled {
		t.Errorf("reversed-order entry = %+v, want {/a/SKILL.md true}", got[0])
	}
}

// TestParseSkillEntriesEnabledAbsent verifies an entry with no `enabled` key
// reads as disabled — TOML's zero value for a missing boolean.
func TestParseSkillEntriesEnabledAbsent(t *testing.T) {
	got := ParseSkillEntries([]byte("[[skills.config]]\npath = \"/a/SKILL.md\"\n"))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(got))
	}
	if got[0].Enabled {
		t.Errorf("entry without an `enabled` key reported enabled: %+v", got[0])
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
	if got[0].Path != "/a/SKILL.md" || !got[0].Enabled {
		t.Errorf("whitespace-varied entry = %+v, want {/a/SKILL.md true}", got[0])
	}
}

// TestParseSkillEntriesSectionBoundary verifies assignments under a LATER
// table are not folded into the preceding entry.
func TestParseSkillEntriesSectionBoundary(t *testing.T) {
	in := "[[skills.config]]\npath = \"/a/SKILL.md\"\n\n[tui]\nenabled = true\n"
	got := ParseSkillEntries([]byte(in))
	if len(got) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(got))
	}
	if got[0].Enabled {
		t.Errorf("`enabled` under [tui] leaked into the skills entry: %+v", got[0])
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
